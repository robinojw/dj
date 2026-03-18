package tui

import (
	"encoding/json"
	"testing"

	"github.com/robinojw/dj/internal/appserver"
)

func TestBridgeSessionConfigured(t *testing.T) {
	event := appserver.ProtoEvent{
		ID:  "",
		Msg: json.RawMessage(`{"type":"session_configured","session_id":"s-1","model":"o4-mini"}`),
	}
	msg := ProtoEventToMsg(event)
	configured, ok := msg.(SessionConfiguredMsg)
	if !ok {
		t.Fatalf("expected SessionConfiguredMsg, got %T", msg)
	}
	if configured.SessionID != "s-1" {
		t.Errorf("expected s-1, got %s", configured.SessionID)
	}
	if configured.Model != "o4-mini" {
		t.Errorf("expected o4-mini, got %s", configured.Model)
	}
}

func TestBridgeTaskStarted(t *testing.T) {
	event := appserver.ProtoEvent{
		Msg: json.RawMessage(`{"type":"task_started","model_context_window":200000}`),
	}
	msg := ProtoEventToMsg(event)
	_, ok := msg.(TaskStartedMsg)
	if !ok {
		t.Fatalf("expected TaskStartedMsg, got %T", msg)
	}
}

func TestBridgeAgentDelta(t *testing.T) {
	event := appserver.ProtoEvent{
		Msg: json.RawMessage(`{"type":"agent_message_delta","delta":"Hello"}`),
	}
	msg := ProtoEventToMsg(event)
	delta, ok := msg.(AgentDeltaMsg)
	if !ok {
		t.Fatalf("expected AgentDeltaMsg, got %T", msg)
	}
	if delta.Delta != "Hello" {
		t.Errorf("expected Hello, got %s", delta.Delta)
	}
}

func TestBridgeAgentMessage(t *testing.T) {
	event := appserver.ProtoEvent{
		Msg: json.RawMessage(`{"type":"agent_message","message":"Hello world"}`),
	}
	msg := ProtoEventToMsg(event)
	completed, ok := msg.(AgentMessageCompletedMsg)
	if !ok {
		t.Fatalf("expected AgentMessageCompletedMsg, got %T", msg)
	}
	if completed.Message != "Hello world" {
		t.Errorf("expected Hello world, got %s", completed.Message)
	}
}

func TestBridgeTaskComplete(t *testing.T) {
	event := appserver.ProtoEvent{
		Msg: json.RawMessage(`{"type":"task_complete","last_agent_message":"Done"}`),
	}
	msg := ProtoEventToMsg(event)
	complete, ok := msg.(TaskCompleteMsg)
	if !ok {
		t.Fatalf("expected TaskCompleteMsg, got %T", msg)
	}
	if complete.LastMessage != "Done" {
		t.Errorf("expected Done, got %s", complete.LastMessage)
	}
}

func TestBridgeExecApproval(t *testing.T) {
	event := appserver.ProtoEvent{
		ID:  "req-1",
		Msg: json.RawMessage(`{"type":"exec_command_request","command":"ls","cwd":"/tmp"}`),
	}
	msg := ProtoEventToMsg(event)
	approval, ok := msg.(ExecApprovalRequestMsg)
	if !ok {
		t.Fatalf("expected ExecApprovalRequestMsg, got %T", msg)
	}
	if approval.EventID != "req-1" {
		t.Errorf("expected req-1, got %s", approval.EventID)
	}
	if approval.Command != "ls" {
		t.Errorf("expected ls, got %s", approval.Command)
	}
}

func TestBridgeUnknownEventReturnsNil(t *testing.T) {
	event := appserver.ProtoEvent{
		Msg: json.RawMessage(`{"type":"unknown_event"}`),
	}
	msg := ProtoEventToMsg(event)
	if msg != nil {
		t.Errorf("expected nil for unknown event, got %T", msg)
	}
}

func TestBridgeInvalidJSONReturnsNil(t *testing.T) {
	event := appserver.ProtoEvent{
		Msg: json.RawMessage(`not json`),
	}
	msg := ProtoEventToMsg(event)
	if msg != nil {
		t.Errorf("expected nil for invalid JSON, got %T", msg)
	}
}
