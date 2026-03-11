package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/robinojw/dj/internal/api"
	"github.com/robinojw/dj/internal/memory"
	"github.com/robinojw/dj/internal/modes"
	"github.com/robinojw/dj/internal/skills"
	"github.com/robinojw/dj/internal/tools"
)

// Worker is a sub-agent goroutine that processes a single subtask.
type Worker struct {
	ID           string
	Task         Subtask
	Status       string // "pending", "running", "completed", "error"
	Output       string
	Mode         AgentMode
	client       *api.ResponsesClient
	skills       *skills.Registry
	memory       *memory.Manager
	model        string
	parentID     string
	registry     *tools.ToolRegistry
	gate         *modes.Gate
	permReqCh    chan<- modes.PermissionRequest
	lastToolName string
	lastToolArgs map[string]any
}

func NewWorker(
	task Subtask,
	client *api.ResponsesClient,
	skillsRegistry *skills.Registry,
	model string,
	parentID string,
	mode AgentMode,
	mem *memory.Manager,
	gate *modes.Gate,
	permReqCh chan<- modes.PermissionRequest,
) *Worker {
	return &Worker{
		ID:        task.ID,
		Task:      task,
		Status:    "pending",
		Mode:      mode,
		client:    client,
		skills:    skillsRegistry,
		memory:    mem,
		model:     model,
		parentID:  parentID,
		gate:      gate,
		permReqCh: permReqCh,
	}
}

// Run executes the worker's task and sends updates to the updates channel.
func (w *Worker) Run(ctx context.Context, updates chan<- WorkerUpdate) {
	w.Status = "running"

	// Build system instructions including relevant skills
	instructions := w.buildInstructions()

	req := api.CreateResponseRequest{
		Model:        w.model,
		Input:        api.MakeStringInput(w.Task.Description),
		Instructions: instructions,
		Reasoning: &api.Reasoning{
			Effort: Modes[w.Mode].ReasoningEffort,
		},
		Stream: true,
	}

	chunks, errs := w.client.Stream(ctx, req)

	for chunk := range chunks {
		select {
		case <-ctx.Done():
			updates <- WorkerUpdate{
				WorkerID: w.ID,
				Type:     UpdateError,
				Error:    ctx.Err(),
			}
			w.Status = "error"
			return
		default:
		}

		switch chunk.Type {
		case "response.output_text.delta":
			w.Output += chunk.Delta
			updates <- WorkerUpdate{
				WorkerID: w.ID,
				Type:     UpdateDelta,
				Content:  chunk.Delta,
			}

		case "response.output_item.added":
			if chunk.Item != nil && chunk.Item.Type == "function_call" {
				// Track tool name and args
				w.lastToolName = chunk.Item.Name
				if chunk.Item.Arguments != "" {
					_ = json.Unmarshal([]byte(chunk.Item.Arguments), &w.lastToolArgs)
				}

				updates <- WorkerUpdate{
					WorkerID: w.ID,
					Type:     UpdateToolCall,
					Content:  fmt.Sprintf("Calling %s", chunk.Item.Name),
				}
			}

		case "response.function_call_result":
			// Generate diff if this tool mutates files
			if w.isMutatingTool(w.lastToolName) {
				if filePath, ok := w.extractToolFilePath(w.lastToolName, w.lastToolArgs); ok {
					if diff, diffErr := generateGitDiff(filePath); diffErr == nil {
						updates <- WorkerUpdate{
							WorkerID: w.ID,
							Type:     UpdateDiffResult,
							Content:  diff.DiffText,
							DiffInfo: &diff,
						}
					}
				}
			}

		case "response.completed":
			if chunk.Response != nil {
				updates <- WorkerUpdate{
					WorkerID: w.ID,
					Type:     UpdateCompleted,
					Content:  w.Output,
					Usage: UsageInfo{
						InputTokens:  chunk.Response.Usage.InputTokens,
						OutputTokens: chunk.Response.Usage.OutputTokens,
					},
				}
			}
		}
	}

	// Check for stream errors
	for err := range errs {
		updates <- WorkerUpdate{
			WorkerID: w.ID,
			Type:     UpdateError,
			Error:    err,
		}
		w.Status = "error"
		return
	}

	w.Status = "completed"
}

// executeTool runs a tool call through the permission gate.
func (w *Worker) executeTool(toolName string, args map[string]any) error {
	decision := w.gate.Evaluate(toolName, args)

	switch decision {
	case modes.GateDeny:
		return fmt.Errorf("tool %q blocked by deny list or mode", toolName)

	case modes.GateAllow:
		return nil

	case modes.GateAskUser:
		respCh := make(chan modes.PermissionResp, 1)
		req := modes.PermissionRequest{
			ID:       fmt.Sprintf("%s-%s", w.ID, toolName),
			WorkerID: w.ID,
			Tool:     toolName,
			Args:     args,
			RespCh:   respCh,
		}

		w.permReqCh <- req

		select {
		case resp := <-respCh:
			if !resp.Allowed {
				return fmt.Errorf("user denied tool: %s", toolName)
			}

			if resp.RememberFor == modes.RememberSession {
				w.gate.AllowForSession(toolName)
			}
			if resp.RememberFor == modes.RememberAlways {
				if err := modes.PersistToolToAllowList("harness.toml", toolName); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to persist to config: %v\n", err)
				}
			}

			return nil

		case <-time.After(5 * time.Minute):
			return fmt.Errorf("permission request timed out")
		}

	default:
		return fmt.Errorf("unknown gate decision: %v", decision)
	}
}

// filterToolsForMode returns tools available in the worker's current mode.
func (w *Worker) filterToolsForMode(allTools []string) []string {
	modeCfg := Modes[w.Mode]
	return FilterTools(allTools, modeCfg)
}

func (w *Worker) buildInstructions() string {
	modeCfg := Modes[w.Mode]
	base := modeCfg.SystemPrompt + "\n\n"

	// Inject memory context if available
	if w.memory != nil {
		base += w.memory.LoadContext() + "\n\n"
	}

	base += fmt.Sprintf("Subtask: %s\n", w.Task.Description)

	if len(w.Task.Files) > 0 {
		base += "\nScoped files:\n"
		for _, f := range w.Task.Files {
			base += fmt.Sprintf("- %s\n", f)
		}
	}

	// Check for implicitly matched skills
	if w.skills != nil {
		matcher := skills.NewMatcher(w.skills)
		if skill := matcher.BestMatch(w.Task.Description); skill != nil {
			base += fmt.Sprintf("\n--- Skill: %s ---\n%s\n", skill.Name, skill.Instructions)
		}
	}

	return base
}

// isMutatingTool returns true if the tool's annotations indicate it writes to the filesystem.
// Returns false for unregistered tools — no hardcoded fallback.
func (w *Worker) isMutatingTool(toolName string) bool {
	if w.registry == nil {
		return false
	}
	if !w.registry.HasAnnotations(toolName) {
		return false
	}
	return w.registry.Annotations(toolName).MutatesFiles
}

// extractToolFilePath uses registry annotations to find the file path arg.
// Falls back to multi-key scan for unregistered tools.
func (w *Worker) extractToolFilePath(toolName string, args map[string]any) (string, bool) {
	if w.registry != nil && w.registry.HasAnnotations(toolName) {
		if key := w.registry.Annotations(toolName).FilePathParam; key != "" {
			if args == nil {
				return "", false
			}
			if val, ok := args[key]; ok {
				if str, ok := val.(string); ok && str != "" {
					return str, true
				}
			}
			return "", false
		}
	}
	return extractFilePath(args)
}

// generateGitDiff runs git diff for the given file and returns the output.
// Returns error if not in a git repo or git command fails.
func generateGitDiff(filePath string) (DiffInfo, error) {
	cmd := exec.Command("git", "diff", "HEAD", filePath)
	output, err := cmd.Output()
	if err != nil {
		return DiffInfo{}, fmt.Errorf("git diff failed: %w", err)
	}

	return DiffInfo{
		FilePath:  filePath,
		DiffText:  string(output),
		Timestamp: time.Now(),
	}, nil
}

// extractFilePath extracts the file_path argument from tool call args.
// Returns the path and true if found, empty string and false otherwise.
func extractFilePath(args map[string]any) (string, bool) {
	if args == nil {
		return "", false
	}

	// Try common parameter names
	for _, key := range []string{"file_path", "path", "filepath"} {
		if val, ok := args[key]; ok {
			if str, ok := val.(string); ok && str != "" {
				return str, true
			}
		}
	}

	return "", false
}
