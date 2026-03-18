package appserver

import (
	"encoding/json"
	"testing"
)

const (
	collabUnmarshalFailFormat = "collab unmarshal: %v"
	collabChildThreadID       = "t-1"
	collabExpectedChildFormat = "expected t-1, got %s"
	expectedWaitingStatuses   = 2
)

func TestUnmarshalCollabSpawnEnd(test *testing.T) {
	raw := `{
		"call_id": "call-1",
		"sender_thread_id": "t-0",
		"new_thread_id": "t-1",
		"new_agent_nickname": "scout",
		"new_agent_role": "researcher",
		"status": "running"
	}`
	var event CollabSpawnEndEvent
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		test.Fatalf(collabUnmarshalFailFormat, err)
	}
	if event.SenderThreadID != "t-0" {
		test.Errorf("expected t-0, got %s", event.SenderThreadID)
	}
	if event.NewThreadID != collabChildThreadID {
		test.Errorf(collabExpectedChildFormat, event.NewThreadID)
	}
	if event.Status != AgentStatusRunning {
		test.Errorf("expected running, got %s", event.Status)
	}
}

func TestUnmarshalCollabWaitingEnd(test *testing.T) {
	raw := `{
		"sender_thread_id": "t-0",
		"call_id": "call-2",
		"statuses": {"t-1": "completed", "t-2": "running"}
	}`
	var event CollabWaitingEndEvent
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		test.Fatalf(collabUnmarshalFailFormat, err)
	}
	if len(event.Statuses) != expectedWaitingStatuses {
		test.Errorf("expected %d statuses, got %d", expectedWaitingStatuses, len(event.Statuses))
	}
}

func TestUnmarshalCollabCloseEnd(test *testing.T) {
	raw := `{
		"call_id": "call-3",
		"sender_thread_id": "t-0",
		"receiver_thread_id": "t-1",
		"receiver_agent_nickname": "scout",
		"receiver_agent_role": "researcher",
		"status": "shutdown"
	}`
	var event CollabCloseEndEvent
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		test.Fatalf(collabUnmarshalFailFormat, err)
	}
	if event.ReceiverThreadID != collabChildThreadID {
		test.Errorf(collabExpectedChildFormat, event.ReceiverThreadID)
	}
	if event.Status != AgentStatusShutdown {
		test.Errorf("expected shutdown, got %s", event.Status)
	}
}
