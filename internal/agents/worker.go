package agents

import (
	"context"
	"encoding/json"
	"fmt"

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
