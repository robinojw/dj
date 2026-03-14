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
	"github.com/robinojw/dj/internal/modes"
	"github.com/robinojw/dj/internal/skills"
)

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
