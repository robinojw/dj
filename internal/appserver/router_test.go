package appserver

import (
	"encoding/json"
	"sync/atomic"
	"testing"
)

func TestRouterDispatchesAgentMessageDelta(t *testing.T) {
	router := NewEventRouter()

	var receivedDelta string
	router.OnAgentMessageDelta(func(event AgentMessageDelta) {
		receivedDelta = event.Delta
	})

	event := Event{
		ID:  "sub-1",
		Msg: json.RawMessage(`{"type":"agent_message_delta","delta":"hello"}`),
	}
	router.HandleEvent(event)

	if receivedDelta != "hello" {
		t.Errorf("expected hello, got %s", receivedDelta)
	}
}

func TestRouterDispatchesExecCommandBegin(t *testing.T) {
	router := NewEventRouter()

	var receivedCommand string
	router.OnExecCommandBegin(func(event ExecCommandBegin) {
		receivedCommand = event.Command
	})

	event := Event{
		ID:  "sub-1",
		Msg: json.RawMessage(`{"type":"exec_command_begin","call_id":"cmd-1","command":"ls"}`),
	}
	router.HandleEvent(event)

	if receivedCommand != "ls" {
		t.Errorf("expected ls, got %s", receivedCommand)
	}
}

func TestRouterDispatchesExecCommandOutputDelta(t *testing.T) {
	router := NewEventRouter()

	var receivedData string
	router.OnExecCommandOutputDelta(func(event ExecCommandOutputDelta) {
		receivedData = event.Delta
	})

	event := Event{
		ID:  "sub-1",
		Msg: json.RawMessage(`{"type":"exec_command_output_delta","call_id":"cmd-1","delta":"output\n"}`),
	}
	router.HandleEvent(event)

	if receivedData != "output\n" {
		t.Errorf("expected output, got %s", receivedData)
	}
}

func TestRouterDispatchesExecApprovalRequest(t *testing.T) {
	router := NewEventRouter()

	var called atomic.Bool
	router.OnExecApprovalRequest(func(event ExecApprovalRequest) {
		called.Store(true)
	})

	event := Event{
		ID:  "",
		Msg: json.RawMessage(`{"type":"exec_approval_request","call_id":"cmd-1","command":"rm file"}`),
	}
	router.HandleEvent(event)

	if !called.Load() {
		t.Error("handler was not called")
	}
}

func TestRouterDispatchesTaskComplete(t *testing.T) {
	router := NewEventRouter()

	var called atomic.Bool
	router.OnTaskComplete(func(event TaskComplete) {
		called.Store(true)
	})

	event := Event{
		ID:  "sub-1",
		Msg: json.RawMessage(`{"type":"task_complete"}`),
	}
	router.HandleEvent(event)

	if !called.Load() {
		t.Error("handler was not called")
	}
}
