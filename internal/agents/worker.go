package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/robinojw/dj/internal/api"
	"github.com/robinojw/dj/internal/hooks"
	"github.com/robinojw/dj/internal/memory"
	"github.com/robinojw/dj/internal/modes"
	"github.com/robinojw/dj/internal/skills"
	"github.com/robinojw/dj/internal/tools"
)

// maxToolTurns limits the number of tool-call round-trips to prevent infinite loops.
const maxToolTurns = 25

// Worker is a sub-agent goroutine that processes a single subtask.
type Worker struct {
	ID           string
	Task         Subtask
	Status       string // "pending", "running", "completed", "error", "skipped"
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
	hooks        *hooks.Runner
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
	registry *tools.ToolRegistry,
	permReqCh chan<- modes.PermissionRequest,
	hooks *hooks.Runner,
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
		registry:  registry,
		permReqCh: permReqCh,
		hooks:     hooks,
	}
}

// Run executes the worker's task and sends updates to the updates channel.
// Supports multi-turn tool execution: when the model emits function calls,
// the worker executes them through the ToolRegistry and feeds results back.
func (w *Worker) Run(ctx context.Context, updates chan<- WorkerUpdate) {
	w.Status = "running"

	instructions := w.buildInstructions()
	modeCfg := Modes[w.Mode]

	req := api.CreateResponseRequest{
		Model:        w.model,
		Input:        api.MakeStringInput(w.Task.Description),
		Instructions: instructions,
		Tools:        w.buildToolDefs(modeCfg),
		Reasoning: &api.Reasoning{
			Effort: modeCfg.ReasoningEffort,
		},
		Stream: true,
	}

	for turn := 0; turn < maxToolTurns; turn++ {
		completedResponse, streamErr := w.streamResponse(ctx, req, updates)
		if streamErr != nil {
			updates <- WorkerUpdate{
				WorkerID: w.ID,
				Type:     UpdateError,
				Error:    streamErr,
			}
			w.Status = "error"
			return
		}

		if completedResponse == nil {
			break
		}

		// Collect function calls from the completed response
		var functionCalls []api.OutputItem
		for _, item := range completedResponse.Output {
			if item.Type == "function_call" {
				functionCalls = append(functionCalls, item)
			}
		}

		if len(functionCalls) == 0 {
			// No tool calls — emit completion and finish
			updates <- WorkerUpdate{
				WorkerID: w.ID,
				Type:     UpdateCompleted,
				Content:  w.Output,
				Usage: UsageInfo{
					InputTokens:  completedResponse.Usage.InputTokens,
					OutputTokens: completedResponse.Usage.OutputTokens,
				},
			}
			break
		}

		// Execute each function call and collect results
		results := w.executeToolCalls(ctx, functionCalls, updates)

		// Build follow-up request with tool results
		resultsJSON, _ := json.Marshal(results)
		req = api.CreateResponseRequest{
			Model:              w.model,
			Input:              resultsJSON,
			PreviousResponseID: completedResponse.ID,
			Reasoning: &api.Reasoning{
				Effort: Modes[w.Mode].ReasoningEffort,
			},
			Stream: true,
		}
	}

	w.Status = "completed"
}

// streamResponse streams a single API response, forwarding text deltas and
// tool call notifications to the updates channel. Returns the completed
// response object (nil if the stream ended without one) and any stream error.
func (w *Worker) streamResponse(
	ctx context.Context,
	req api.CreateResponseRequest,
	updates chan<- WorkerUpdate,
) (*api.ResponseObject, error) {
	chunks, errs := w.client.Stream(ctx, req)

	var completedResponse *api.ResponseObject

	for chunk := range chunks {
		select {
		case <-ctx.Done():
			w.fireHook(hooks.HookOnError, map[string]string{
				"error_msg": ctx.Err().Error(),
				"worker_id": w.ID,
				"tool_name": w.lastToolName,
			}, updates)
			return nil, ctx.Err()
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
				w.lastToolName = chunk.Item.Name
				if chunk.Item.Arguments != "" {
					_ = json.Unmarshal([]byte(chunk.Item.Arguments), &w.lastToolArgs)
				}

				argsJSON, _ := json.Marshal(w.lastToolArgs)
				w.fireHook(hooks.HookPreToolCall, map[string]string{
					"tool_name": chunk.Item.Name,
					"tool_args": string(argsJSON),
					"worker_id": w.ID,
				}, updates)

				updates <- WorkerUpdate{
					WorkerID: w.ID,
					Type:     UpdateToolCall,
					Content:  fmt.Sprintf("Calling %s", chunk.Item.Name),
				}
			}

		case "response.completed":
			if chunk.Response != nil {
				completedResponse = chunk.Response
			}
		}
	}

	// Check for stream errors
	for err := range errs {
		w.fireHook(hooks.HookOnError, map[string]string{
			"error_msg": err.Error(),
			"worker_id": w.ID,
			"tool_name": w.lastToolName,
		}, updates)
		return nil, err
	}

	return completedResponse, nil
}

// executeToolCalls runs each function call through the permission gate and
// ToolRegistry, generates diffs for mutating operations, and returns
// FunctionCallResult items to feed back to the API.
func (w *Worker) executeToolCalls(
	ctx context.Context,
	calls []api.OutputItem,
	updates chan<- WorkerUpdate,
) []api.FunctionCallResult {
	results := make([]api.FunctionCallResult, 0, len(calls))

	for _, call := range calls {
		var args map[string]any
		if call.Arguments != "" {
			_ = json.Unmarshal([]byte(call.Arguments), &args)
		}

		output, err := w.executeTool(ctx, call.Name, args)
		if err != nil {
			output = fmt.Sprintf("Error: %v", err)
		}

		argsJSON, _ := json.Marshal(args)
		w.fireHook(hooks.HookPostToolCall, map[string]string{
			"tool_name": call.Name,
			"tool_args": string(argsJSON),
			"worker_id": w.ID,
		}, updates)

		// Generate diff if this tool mutates files
		if err == nil && w.isMutatingTool(call.Name) {
			if filePath, ok := w.extractToolFilePath(call.Name, args); ok {
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

		updates <- WorkerUpdate{
			WorkerID: w.ID,
			Type:     UpdateToolResult,
			Content:  output,
		}

		results = append(results, api.FunctionCallResult{
			Type:   "function_call_output",
			CallID: call.ID,
			Output: output,
		})
	}

	return results
}

// executeTool runs a tool call through the permission gate and dispatches
// to the native ToolRegistry handler when allowed.
func (w *Worker) executeTool(ctx context.Context, toolName string, args map[string]any) (string, error) {
	decision := w.gate.Evaluate(toolName, args)

	switch decision {
	case modes.GateDeny:
		return "", fmt.Errorf("tool %q blocked by deny list or mode", toolName)

	case modes.GateAllow:
		return w.dispatchTool(ctx, toolName, args)

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
				return "", fmt.Errorf("user denied tool: %s", toolName)
			}

			if resp.RememberFor == modes.RememberSession {
				w.gate.AllowForSession(toolName)
			}
			if resp.RememberFor == modes.RememberAlways {
				if err := modes.PersistToolToAllowList("harness.toml", toolName); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to persist to config: %v\n", err)
				}
			}

			return w.dispatchTool(ctx, toolName, args)

		case <-time.After(5 * time.Minute):
			return "", fmt.Errorf("permission request timed out")
		}

	default:
		return "", fmt.Errorf("unknown gate decision: %v", decision)
	}
}

// dispatchTool executes a tool through the registry if a native handler exists,
// otherwise returns empty output (tool execution deferred to the Responses API).
func (w *Worker) dispatchTool(ctx context.Context, toolName string, args map[string]any) (string, error) {
	if w.registry != nil && w.registry.Has(toolName) {
		return w.registry.Dispatch(ctx, toolName, args)
	}
	return "", nil
}

// fireHook fires a hook and sends the result as an update. Advisory only — never fatal.
func (w *Worker) fireHook(event hooks.HookEvent, vars map[string]string, updates chan<- WorkerUpdate) {
	if w.hooks == nil {
		return
	}
	result, _ := w.hooks.Fire(event, vars)
	if result != nil {
		updates <- WorkerUpdate{
			WorkerID:   w.ID,
			Type:       UpdateHookResult,
			HookResult: result,
		}
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

	if w.skills != nil {
		matcher := skills.NewMatcher(w.skills)
		if skill := matcher.BestMatch(w.Task.Description); skill != nil {
			base += fmt.Sprintf("\n--- Skill: %s ---\n%s\n", skill.Name, skill.Instructions)
		}
	}

	return base
}

// buildToolDefs returns API tool definitions filtered by the current mode.
func (w *Worker) buildToolDefs(modeCfg modes.ModeConfig) []api.Tool {
	if w.registry == nil {
		return nil
	}
	return w.registry.ToolDefinitions(modeCfg.AllowedTools)
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
func extractFilePath(args map[string]any) (string, bool) {
	if args == nil {
		return "", false
	}

	for _, key := range []string{"file_path", "path", "filepath"} {
		if val, ok := args[key]; ok {
			if str, ok := val.(string); ok && str != "" {
				return str, true
			}
		}
	}

	return "", false
}
