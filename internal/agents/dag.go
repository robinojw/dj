package agents

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// dagState holds the mutable scheduling state for topological execution.
type dagState struct {
	inDegree   map[string]int      // remaining unmet deps per task
	dependents map[string][]string // task ID → IDs that depend on it
	mu         sync.Mutex
}

// buildDAG validates the dependency graph and returns the scheduling state.
// Returns an error if a cycle is detected or a dependency references a missing task.
func buildDAG(subtasks []Subtask) (*dagState, error) {
	ids := make(map[string]bool, len(subtasks))
	for _, t := range subtasks {
		ids[t.ID] = true
	}

	// Check for missing dependencies
	for _, t := range subtasks {
		for _, dep := range t.DependsOn {
			if !ids[dep] {
				return nil, fmt.Errorf("task %q depends on unknown task %q", t.ID, dep)
			}
		}
	}

	inDegree := make(map[string]int, len(subtasks))
	dependents := make(map[string][]string, len(subtasks))

	for _, t := range subtasks {
		inDegree[t.ID] = len(t.DependsOn)
		for _, dep := range t.DependsOn {
			dependents[dep] = append(dependents[dep], t.ID)
		}
	}

	// Kahn's algorithm for cycle detection
	queue := make([]string, 0)
	for _, t := range subtasks {
		if inDegree[t.ID] == 0 {
			queue = append(queue, t.ID)
		}
	}

	visited := 0
	tempInDegree := make(map[string]int, len(inDegree))
	for k, v := range inDegree {
		tempInDegree[k] = v
	}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		visited++
		for _, dep := range dependents[node] {
			tempInDegree[dep]--
			if tempInDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	if visited != len(subtasks) {
		var cycleNodes []string
		for id, deg := range tempInDegree {
			if deg > 0 {
				cycleNodes = append(cycleNodes, id)
			}
		}
		sort.Strings(cycleNodes)
		return nil, fmt.Errorf("dependency cycle detected among tasks: [%s]", strings.Join(cycleNodes, ", "))
	}

	return &dagState{
		inDegree:   inDegree,
		dependents: dependents,
	}, nil
}

// readySet returns all task IDs with zero in-degree (ready to run).
func (d *dagState) readySet() []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	var ready []string
	for id, deg := range d.inDegree {
		if deg == 0 {
			ready = append(ready, id)
		}
	}
	sort.Strings(ready)
	return ready
}

// markCompleted decrements in-degrees for dependents of the given task
// and returns any newly unblocked task IDs.
func (d *dagState) markCompleted(taskID string) []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	var ready []string
	for _, dep := range d.dependents[taskID] {
		d.inDegree[dep]--
		if d.inDegree[dep] == 0 {
			ready = append(ready, dep)
		}
	}
	delete(d.inDegree, taskID)
	return ready
}

// skipDependents marks all transitive dependents of failedID as "skipped"
// and sends UpdateSkipped for each.
func skipDependents(dag *dagState, failedID string, workers map[string]*Worker, updates chan<- WorkerUpdate) {
	dag.mu.Lock()

	visited := make(map[string]bool)
	queue := []string{failedID}
	var pending []WorkerUpdate

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, dep := range dag.dependents[current] {
			if visited[dep] {
				continue
			}
			visited[dep] = true

			if w, ok := workers[dep]; ok && w.Status == "pending" {
				w.Status = "skipped"
				delete(dag.inDegree, dep)
				pending = append(pending, WorkerUpdate{
					WorkerID: dep,
					Type:     UpdateSkipped,
				})
			}

			queue = append(queue, dep)
		}
	}

	dag.mu.Unlock()

	for _, u := range pending {
		updates <- u
	}
}
