package appserver

import (
	"encoding/json"
	"testing"
)

func TestThreadCreateParamsMarshal(t *testing.T) {
	params := ThreadCreateParams{
		Instructions: "Build a web server",
	}
	data, err := json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	json.Unmarshal(data, &parsed)
	if parsed["instructions"] != "Build a web server" {
		t.Errorf("expected instructions, got %v", parsed["instructions"])
	}
}

func TestThreadCreateResultUnmarshal(t *testing.T) {
	raw := `{"threadId":"t-abc123"}`
	var result ThreadCreateResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		t.Fatal(err)
	}
	if result.ThreadID != "t-abc123" {
		t.Errorf("expected t-abc123, got %s", result.ThreadID)
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
