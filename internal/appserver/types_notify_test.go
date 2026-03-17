package appserver

import (
	"encoding/json"
	"testing"
)

func TestThreadStatusChangedUnmarshal(t *testing.T) {
	raw := `{"threadId":"t-1","status":"completed","title":"Done"}`
	var params ThreadStatusChanged
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatal(err)
	}
	if params.ThreadID != "t-1" {
		t.Errorf("expected t-1, got %s", params.ThreadID)
	}
	if params.Status != ThreadStatusCompleted {
		t.Errorf("expected completed, got %s", params.Status)
	}
}

func TestThreadMessageCreatedUnmarshal(t *testing.T) {
	raw := `{"threadId":"t-1","messageId":"m-1","role":"assistant","content":"Hello"}`
	var params ThreadMessageCreated
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatal(err)
	}
	if params.Role != "assistant" {
		t.Errorf("expected assistant, got %s", params.Role)
	}
	if params.Content != "Hello" {
		t.Errorf("expected Hello, got %s", params.Content)
	}
}

func TestThreadMessageDeltaUnmarshal(t *testing.T) {
	raw := `{"threadId":"t-1","messageId":"m-1","delta":"more text"}`
	var params ThreadMessageDelta
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatal(err)
	}
	if params.Delta != "more text" {
		t.Errorf("expected 'more text', got %s", params.Delta)
	}
}

func TestCommandOutputUnmarshal(t *testing.T) {
	raw := `{"threadId":"t-1","execId":"e-1","data":"line of output\n"}`
	var params CommandOutput
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatal(err)
	}
	if params.ExecID != "e-1" {
		t.Errorf("expected e-1, got %s", params.ExecID)
	}
}

func TestCommandFinishedUnmarshal(t *testing.T) {
	raw := `{"threadId":"t-1","execId":"e-1","exitCode":0}`
	var params CommandFinished
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatal(err)
	}
	if params.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", params.ExitCode)
	}
}
