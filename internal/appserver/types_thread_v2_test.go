package appserver

import (
	"encoding/json"
	"testing"
)

const (
	v2UnmarshalFailFormat  = "v2 unmarshal: %v"
	expectedThreadID       = "t-1"
	expectedThreadIDFormat = "expected t-1, got %s"
)

func TestUnmarshalThreadStarted(test *testing.T) {
	raw := `{
		"thread": {
			"id": "t-1",
			"status": "idle",
			"source": {"type": "sub_agent", "parent_thread_id": "t-0", "depth": 1, "agent_nickname": "scout", "agent_role": "researcher"}
		}
	}`
	var notification ThreadStartedNotification
	if err := json.Unmarshal([]byte(raw), &notification); err != nil {
		test.Fatalf(v2UnmarshalFailFormat, err)
	}
	if notification.Thread.ID != expectedThreadID {
		test.Errorf(expectedThreadIDFormat, notification.Thread.ID)
	}
	if notification.Thread.Source.Type != SourceTypeSubAgent {
		test.Errorf("expected sub_agent source, got %s", notification.Thread.Source.Type)
	}
	if notification.Thread.Source.ParentThreadID != "t-0" {
		test.Errorf("expected parent t-0, got %s", notification.Thread.Source.ParentThreadID)
	}
}

func TestUnmarshalThreadStartedCLISource(test *testing.T) {
	raw := `{"thread": {"id": "t-1", "status": "idle", "source": {"type": "cli"}}}`
	var notification ThreadStartedNotification
	if err := json.Unmarshal([]byte(raw), &notification); err != nil {
		test.Fatalf(v2UnmarshalFailFormat, err)
	}
	if notification.Thread.Source.Type != SourceTypeCLI {
		test.Errorf("expected cli source, got %s", notification.Thread.Source.Type)
	}
}

func TestUnmarshalTurnStarted(test *testing.T) {
	raw := `{"thread_id": "t-1", "turn": {"id": "turn-1", "status": "in_progress"}}`
	var notification TurnStartedNotification
	if err := json.Unmarshal([]byte(raw), &notification); err != nil {
		test.Fatalf(v2UnmarshalFailFormat, err)
	}
	if notification.ThreadID != expectedThreadID {
		test.Errorf(expectedThreadIDFormat, notification.ThreadID)
	}
}

func TestUnmarshalTurnCompleted(test *testing.T) {
	raw := `{"thread_id": "t-1", "turn": {"id": "turn-1", "status": "completed"}}`
	var notification TurnCompletedNotification
	if err := json.Unmarshal([]byte(raw), &notification); err != nil {
		test.Fatalf(v2UnmarshalFailFormat, err)
	}
	if notification.ThreadID != expectedThreadID {
		test.Errorf(expectedThreadIDFormat, notification.ThreadID)
	}
}
