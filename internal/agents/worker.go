package agents

import (
	"context"
	"fmt"

	"github.com/robinojw/dj/internal/api"
	"github.com/robinojw/dj/internal/memory"
	"github.com/robinojw/dj/internal/skills"
)

// Worker is a sub-agent goroutine that processes a single subtask.
type Worker struct {
	ID       string
	Task     Subtask
	Status   string // "pending", "running", "completed", "error"
	Output   string
	Mode     AgentMode
	client   *api.ResponsesClient
	skills   *skills.Registry
	memory   *memory.Manager
	model    string
	parentID string
}

func NewWorker(
	task Subtask,
	client *api.ResponsesClient,
	skillsRegistry *skills.Registry,
	model string,
	parentID string,
	mode AgentMode,
	mem *memory.Manager,
) *Worker {
	return &Worker{
		ID:       task.ID,
		Task:     task,
		Status:   "pending",
		Mode:     mode,
		client:   client,
		skills:   skillsRegistry,
		memory:   mem,
		model:    model,
		parentID: parentID,
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
				updates <- WorkerUpdate{
					WorkerID: w.ID,
					Type:     UpdateToolCall,
					Content:  fmt.Sprintf("Calling %s", chunk.Item.Name),
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
