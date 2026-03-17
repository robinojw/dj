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

func TestItemStartedUnmarshal(t *testing.T) {
	raw := `{"threadId":"t-1","itemId":"item-1","role":"assistant","type":"message"}`
	var params ItemStarted
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatal(err)
	}
	if params.Role != "assistant" {
		t.Errorf("expected assistant, got %s", params.Role)
	}
	if params.Type != "message" {
		t.Errorf("expected message, got %s", params.Type)
	}
	if params.ItemID != "item-1" {
		t.Errorf("expected item-1, got %s", params.ItemID)
	}
}

func TestItemCompletedUnmarshal(t *testing.T) {
	raw := `{"threadId":"t-1","itemId":"item-1","content":"Hello world"}`
	var params ItemCompleted
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatal(err)
	}
	if params.Content != "Hello world" {
		t.Errorf("expected Hello world, got %s", params.Content)
	}
}

func TestItemMessageDeltaUnmarshal(t *testing.T) {
	raw := `{"threadId":"t-1","itemId":"item-1","delta":"more text"}`
	var params ItemMessageDelta
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatal(err)
	}
	if params.Delta != "more text" {
		t.Errorf("expected 'more text', got %s", params.Delta)
	}
}

func TestTurnStartedUnmarshal(t *testing.T) {
	raw := `{"threadId":"t-1","turnId":"turn-1"}`
	var params TurnStarted
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatal(err)
	}
	if params.TurnID != "turn-1" {
		t.Errorf("expected turn-1, got %s", params.TurnID)
	}
}

func TestTurnCompletedUnmarshal(t *testing.T) {
	raw := `{"threadId":"t-1","turnId":"turn-1"}`
	var params TurnCompleted
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatal(err)
	}
	if params.TurnID != "turn-1" {
		t.Errorf("expected turn-1, got %s", params.TurnID)
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
