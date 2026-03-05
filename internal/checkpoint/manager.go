package checkpoint

import (
	"fmt"
	"os"
	"time"
)

// Checkpoint records the state of files before a destructive action.
type Checkpoint struct {
	ID            string
	ResponseID    string            // Responses API previous_response_id
	Timestamp     time.Time
	FileSnapshots map[string][]byte // path → content before mutation (nil = didn't exist)
	Description   string            // e.g. "Before: write auth/handler.go"
}

// Manager maintains a bounded stack of checkpoints.
type Manager struct {
	stack   []Checkpoint
	maxSize int
	counter int
}

func NewManager(maxSize int) *Manager {
	if maxSize <= 0 {
		maxSize = 20
	}
	return &Manager{maxSize: maxSize}
}

// Before snapshots the given files and returns a checkpoint.
// Files that don't exist are recorded with nil content (so Restore deletes them).
func (m *Manager) Before(files []string, responseID, description string) (Checkpoint, error) {
	m.counter++
	snap := make(map[string][]byte, len(files))

	for _, f := range files {
		content, err := os.ReadFile(f)
		if os.IsNotExist(err) {
			snap[f] = nil // file didn't exist — restore will delete it
		} else if err != nil {
			return Checkpoint{}, fmt.Errorf("snapshot %s: %w", f, err)
		} else {
			snap[f] = content
		}
	}

	cp := Checkpoint{
		ID:            fmt.Sprintf("cp-%d", m.counter),
		ResponseID:    responseID,
		Timestamp:     time.Now(),
		FileSnapshots: snap,
		Description:   description,
	}

	m.Push(cp)
	return cp, nil
}

// Restore reverts all files in the checkpoint to their snapshotted state.
func (m *Manager) Restore(cp Checkpoint) error {
	for path, content := range cp.FileSnapshots {
		if content == nil {
			// File didn't exist before — remove it
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove %s: %w", path, err)
			}
		} else {
			if err := os.WriteFile(path, content, 0644); err != nil {
				return fmt.Errorf("restore %s: %w", path, err)
			}
		}
	}
	return nil
}

// Push adds a checkpoint to the stack, evicting the oldest if at capacity.
func (m *Manager) Push(cp Checkpoint) {
	if len(m.stack) >= m.maxSize {
		m.stack = m.stack[1:]
	}
	m.stack = append(m.stack, cp)
}

// Pop removes and returns the most recent checkpoint.
func (m *Manager) Pop() *Checkpoint {
	if len(m.stack) == 0 {
		return nil
	}
	cp := m.stack[len(m.stack)-1]
	m.stack = m.stack[:len(m.stack)-1]
	return &cp
}

// Peek returns the most recent checkpoint without removing it.
func (m *Manager) Peek() *Checkpoint {
	if len(m.stack) == 0 {
		return nil
	}
	return &m.stack[len(m.stack)-1]
}

// Len returns the number of stored checkpoints.
func (m *Manager) Len() int {
	return len(m.stack)
}

// List returns all checkpoints (oldest first).
func (m *Manager) List() []Checkpoint {
	return m.stack
}
