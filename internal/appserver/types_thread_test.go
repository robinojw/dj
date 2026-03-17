package appserver

import (
	"encoding/json"
	"testing"
)

func TestThreadStartParamsMarshal(t *testing.T) {
	params := ThreadStartParams{
		Model: "gpt-4o",
	}
	data, err := json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	json.Unmarshal(data, &parsed)
	if parsed["model"] != "gpt-4o" {
		t.Errorf("expected model gpt-4o, got %v", parsed["model"])
	}
}

func TestThreadStartParamsOmitsEmptyModel(t *testing.T) {
	params := ThreadStartParams{}
	data, err := json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	json.Unmarshal(data, &parsed)
	if _, hasModel := parsed["model"]; hasModel {
		t.Error("expected model to be omitted when empty")
	}
}

func TestThreadStartResultUnmarshal(t *testing.T) {
	raw := `{"thread":{"id":"thr_abc123"}}`
	var result ThreadStartResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		t.Fatal(err)
	}
	if result.Thread.ID != "thr_abc123" {
		t.Errorf("expected thr_abc123, got %s", result.Thread.ID)
	}
}

func TestThreadListResultUnmarshal(t *testing.T) {
	raw := `{"threads":[{"id":"t-1","status":"active","title":"Test"}]}`
	var result ThreadListResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Threads) != 1 {
		t.Fatalf("expected 1 thread, got %d", len(result.Threads))
	}
	if result.Threads[0].ID != "t-1" {
		t.Errorf("expected id t-1, got %s", result.Threads[0].ID)
	}
	if result.Threads[0].Status != "active" {
		t.Errorf("expected status active, got %s", result.Threads[0].Status)
	}
}

func TestThreadStatusValues(t *testing.T) {
	if ThreadStatusActive != "active" {
		t.Errorf("expected active, got %s", ThreadStatusActive)
	}
	if ThreadStatusCompleted != "completed" {
		t.Errorf("expected completed, got %s", ThreadStatusCompleted)
	}
	if ThreadStatusError != "error" {
		t.Errorf("expected error, got %s", ThreadStatusError)
	}
}
