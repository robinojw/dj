package agents

import (
	"context"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/api"
	"github.com/robinojw/dj/internal/memory"
	"github.com/robinojw/dj/internal/modes"
	"github.com/robinojw/dj/internal/skills"
	"github.com/robinojw/dj/internal/tools"
)

// Orchestrator manages task decomposition and multi-agent dispatch.
type Orchestrator struct {
	RootID    string
	Workers   map[string]*Worker
	UpdatesCh chan WorkerUpdate
	Mode      AgentMode
	Memory    *memory.Manager
	Gate      *modes.Gate
	Registry  *tools.ToolRegistry
	PermReqCh chan modes.PermissionRequest
	client    *api.ResponsesClient
	skills    *skills.Registry
	model     string
	tracker   *api.Tracker
	mu        sync.RWMutex
}

func NewOrchestrator(
	client *api.ResponsesClient,
	skillsRegistry *skills.Registry,
	tracker *api.Tracker,
	model string,
) *Orchestrator {
	return &Orchestrator{
		RootID:    "root",
		Workers:   make(map[string]*Worker),
		UpdatesCh: make(chan WorkerUpdate, 128),
		client:    client,
		skills:    skillsRegistry,
		model:     model,
		tracker:   tracker,
	}
}

// Dispatch spawns workers for each subtask using topological scheduling.
func (o *Orchestrator) Dispatch(subtasks []Subtask) tea.Cmd {
	o.mu.Lock()
	for _, task := range subtasks {
		w := NewWorker(task, o.client, o.skills, o.model, o.RootID, o.Mode, o.Memory, o.Gate, o.Registry, o.PermReqCh)
		o.Workers[w.ID] = w
	}
	o.mu.Unlock()

	return func() tea.Msg {
		dag, err := buildDAG(subtasks)
		if err != nil {
			o.UpdatesCh <- WorkerUpdate{
				Type:    UpdateError,
				Content: err.Error(),
				Error:   err,
			}
			return nil
		}

		var wg sync.WaitGroup
		doneCh := make(chan string, len(subtasks))

		launch := func(id string) {
			w := o.Workers[id]
			if w == nil {
				return
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				w.Run(context.Background(), o.UpdatesCh)
				doneCh <- w.ID
			}()
		}

		// Start all initially ready tasks
		for _, id := range dag.readySet() {
			launch(id)
		}

		// Coordinator: listen for completions, enqueue newly ready tasks
		go func() {
			remaining := len(subtasks)
			for remaining > 0 {
				id := <-doneCh
				remaining--

				w := o.Workers[id]
				if w != nil && w.Status == "error" {
					skipDependents(dag, id, o.Workers, o.UpdatesCh)
					// Count skipped workers
					o.mu.RLock()
					for _, sw := range o.Workers {
						if sw.Status == "skipped" {
							remaining--
						}
					}
					o.mu.RUnlock()
					continue
				}

				for _, readyID := range dag.markCompleted(id) {
					launch(readyID)
				}
			}
		}()

		wg.Wait()
		return nil
	}
}

// ListenForUpdates returns a tea.Cmd that listens for worker updates.
func (o *Orchestrator) ListenForUpdates() tea.Cmd {
	return func() tea.Msg {
		update := <-o.UpdatesCh
		if update.Type == UpdateCompleted || update.Type == UpdateError || update.Type == UpdateSkipped {
			if update.Usage.InputTokens > 0 {
				o.tracker.Record(api.Usage{
					InputTokens:  update.Usage.InputTokens,
					OutputTokens: update.Usage.OutputTokens,
				})
			}
		}
		return update
	}
}

// GetWorker returns a worker by ID.
func (o *Orchestrator) GetWorker(id string) *Worker {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.Workers[id]
}

// AllCompleted returns true if all workers have finished.
func (o *Orchestrator) AllCompleted() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	for _, w := range o.Workers {
		if w.Status != "completed" && w.Status != "error" && w.Status != "skipped" {
			return false
		}
	}
	return len(o.Workers) > 0
}
