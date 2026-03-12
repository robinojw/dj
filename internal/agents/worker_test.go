package agents

import "testing"

func TestWorkerNilHooksDoesNotPanic(t *testing.T) {
	// Constructing a worker with nil hooks should not panic
	task := Subtask{ID: "test-1", Description: "test task"}
	w := NewWorker(task, nil, nil, "test-model", "parent", ModeConfirm, nil, nil, nil, nil)
	if w.hooks != nil {
		t.Error("Expected nil hooks on worker")
	}
}
