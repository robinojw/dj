package agents

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/robinojw/dj/internal/api"
)

const defaultTeamThreshold = 3

// TaskRouter analyzes a prompt and decides whether to spawn a team.
type TaskRouter struct {
	client    *api.ResponsesClient
	model     string
	threshold int
}

func NewTaskRouter(client *api.ResponsesClient, model string, threshold int) *TaskRouter {
	if threshold <= 0 {
		threshold = defaultTeamThreshold
	}
	return &TaskRouter{
		client:    client,
		model:     model,
		threshold: threshold,
	}
}

// Analyze makes a preflight call with low reasoning effort to decompose a task.
func (r *TaskRouter) Analyze(ctx context.Context, prompt string) (TaskAnalysis, error) {
	instructions := `Analyze the following task and decompose it into subtasks.
Return a JSON object with this schema:
{
  "subtasks": [{"id": "1", "description": "...", "depends_on": [], "files": []}],
  "complexity": "low|medium|high",
  "parallelizable": true|false
}
Be concise. Only create subtasks if the work genuinely has multiple independent parts.`

	req := api.CreateResponseRequest{
		Model:        r.model,
		Input:        api.MakeStringInput(prompt),
		Instructions: instructions,
		Reasoning: &api.Reasoning{
			Effort: "low",
		},
	}

	resp, err := r.client.Send(ctx, req)
	if err != nil {
		return TaskAnalysis{}, fmt.Errorf("task router analysis failed: %w", err)
	}

	// Extract text from response output
	var text string
	for _, item := range resp.Output {
		if item.Type == "message" && item.Content != nil {
			var content []struct {
				Text string `json:"text"`
			}
			if err := json.Unmarshal(item.Content, &content); err == nil {
				for _, c := range content {
					text += c.Text
				}
			}
		}
	}

	var analysis TaskAnalysis
	if err := json.Unmarshal([]byte(text), &analysis); err != nil {
		// If parsing fails, treat as a single task
		return TaskAnalysis{
			Subtasks:       []Subtask{{ID: "1", Description: prompt}},
			Complexity:     "low",
			Parallelizable: false,
		}, nil
	}

	return analysis, nil
}

// ShouldSpawnTeam returns true if the analysis suggests using multiple agents.
func (r *TaskRouter) ShouldSpawnTeam(analysis TaskAnalysis) bool {
	return len(analysis.Subtasks) >= r.threshold && analysis.Parallelizable
}
