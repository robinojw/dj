package agents

import (
	"context"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/api"
	"github.com/robinojw/dj/internal/memory"
	"github.com/robinojw/dj/internal/skills"
)

// Orchestrator manages task decomposition and multi-agent dispatch.
type Orchestrator struct {
	RootID    string
	Workers   map[string]*Worker
	UpdatesCh chan WorkerUpdate
	Mode      AgentMode
	Memory    *memory.Manager
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

// Dispatch spawns workers for each subtask.
func (o *Orchestrator) Dispatch(subtasks []Subtask) tea.Cmd {
	o.mu.Lock()
	for _, task := range subtasks {
		w := NewWorker(task, o.client, o.skills, o.model, o.RootID, o.Mode, o.Memory)
		o.Workers[w.ID] = w
	}
	o.mu.Unlock()

	// Launch workers, respecting dependencies
	return func() tea.Msg {
		var wg sync.WaitGroup
		completed := make(map[string]bool)
		var completedMu sync.Mutex

		// Start independent tasks first
		for _, w := range o.Workers {
			if len(w.Task.DependsOn) == 0 {
				wg.Add(1)
				go func(worker *Worker) {
					defer wg.Done()
					worker.Run(context.Background(), o.UpdatesCh)
					completedMu.Lock()
					completed[worker.ID] = true
					completedMu.Unlock()
				}(w)
			}
		}

		// Wait for independent tasks, then start dependent ones
		wg.Wait()

		// Start tasks whose dependencies are met
		for _, w := range o.Workers {
			if len(w.Task.DependsOn) > 0 && w.Status == "pending" {
				allDone := true
				completedMu.Lock()
				for _, dep := range w.Task.DependsOn {
					if !completed[dep] {
						allDone = false
						break
					}
				}
				completedMu.Unlock()

				if allDone {
					wg.Add(1)
					go func(worker *Worker) {
						defer wg.Done()
						worker.Run(context.Background(), o.UpdatesCh)
					}(w)
				}
			}
		}

		wg.Wait()
		return nil
	}
}

// ListenForUpdates returns a tea.Cmd that listens for worker updates.
func (o *Orchestrator) ListenForUpdates() tea.Cmd {
	return func() tea.Msg {
		update := <-o.UpdatesCh
		if update.Type == UpdateCompleted || update.Type == UpdateError {
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
		if w.Status != "completed" && w.Status != "error" {
			return false
		}
	}
	return len(o.Workers) > 0
}
