package tui

import (
	"encoding/json"
	"testing"

	"github.com/robinojw/dj/internal/appserver"
)

const testRequestID = "req-1"

func TestBridgeSessionConfigured(testing *testing.T) {
	event := appserver.ProtoEvent{
		ID:  "",
		Msg: json.RawMessage(`{"type":"session_configured","session_id":"s-1","model":"o4-mini"}`),
	}
	msg := ProtoEventToMsg(event)
	configured, ok := msg.(SessionConfiguredMsg)
	if !ok {
		testing.Fatalf("expected SessionConfiguredMsg, got %T", msg)
	}
	if configured.SessionID != "s-1" {
		testing.Errorf("expected s-1, got %s", configured.SessionID)
	}
	if configured.Model != "o4-mini" {
		testing.Errorf("expected o4-mini, got %s", configured.Model)
	}
}

func TestBridgeTaskStarted(testing *testing.T) {
	event := appserver.ProtoEvent{
		Msg: json.RawMessage(`{"type":"task_started","model_context_window":200000}`),
	}
	msg := ProtoEventToMsg(event)
	_, ok := msg.(TaskStartedMsg)
	if !ok {
		testing.Fatalf("expected TaskStartedMsg, got %T", msg)
	}
}

func TestBridgeAgentDelta(testing *testing.T) {
	event := appserver.ProtoEvent{
		Msg: json.RawMessage(`{"type":"agent_message_delta","delta":"Hello"}`),
	}
	msg := ProtoEventToMsg(event)
	delta, ok := msg.(AgentDeltaMsg)
	if !ok {
		testing.Fatalf("expected AgentDeltaMsg, got %T", msg)
	}
	if delta.Delta != "Hello" {
		testing.Errorf("expected Hello, got %s", delta.Delta)
	}
}

func TestBridgeAgentMessage(testing *testing.T) {
	event := appserver.ProtoEvent{
		Msg: json.RawMessage(`{"type":"agent_message","message":"Hello world"}`),
	}
	msg := ProtoEventToMsg(event)
	completed, ok := msg.(AgentMessageCompletedMsg)
	if !ok {
		testing.Fatalf("expected AgentMessageCompletedMsg, got %T", msg)
	}
	if completed.Message != "Hello world" {
		testing.Errorf("expected Hello world, got %s", completed.Message)
	}
}

func TestBridgeTaskComplete(testing *testing.T) {
	event := appserver.ProtoEvent{
		Msg: json.RawMessage(`{"type":"task_complete","last_agent_message":"Done"}`),
	}
	msg := ProtoEventToMsg(event)
	complete, ok := msg.(TaskCompleteMsg)
	if !ok {
		testing.Fatalf("expected TaskCompleteMsg, got %T", msg)
	}
	if complete.LastMessage != "Done" {
		testing.Errorf("expected Done, got %s", complete.LastMessage)
	}
}

func TestBridgeExecApproval(testing *testing.T) {
	event := appserver.ProtoEvent{
		ID:  testRequestID,
		Msg: json.RawMessage(`{"type":"exec_command_request","command":"ls","cwd":"/tmp"}`),
	}
	msg := ProtoEventToMsg(event)
	approval, ok := msg.(ExecApprovalRequestMsg)
	if !ok {
		testing.Fatalf("expected ExecApprovalRequestMsg, got %T", msg)
	}
	if approval.EventID != testRequestID {
		testing.Errorf("expected %s, got %s", testRequestID, approval.EventID)
	}
	if approval.Command != "ls" {
		testing.Errorf("expected ls, got %s", approval.Command)
	}
}

func TestBridgeAgentReasoningDelta(testing *testing.T) {
	event := appserver.ProtoEvent{
		Msg: json.RawMessage(`{"type":"agent_reasoning_delta","delta":"Let me think..."}`),
	}
	msg := ProtoEventToMsg(event)
	reasoning, ok := msg.(AgentReasoningDeltaMsg)
	if !ok {
		testing.Fatalf("expected AgentReasoningDeltaMsg, got %T", msg)
	}
	if reasoning.Delta != "Let me think..." {
		testing.Errorf("expected Let me think..., got %s", reasoning.Delta)
	}
}

func TestBridgeUnknownEventReturnsNil(testing *testing.T) {
	event := appserver.ProtoEvent{
		Msg: json.RawMessage(`{"type":"unknown_event"}`),
	}
	msg := ProtoEventToMsg(event)
	if msg != nil {
		testing.Errorf("expected nil for unknown event, got %T", msg)
	}
}

func TestBridgeInvalidJSONReturnsNil(testing *testing.T) {
	event := appserver.ProtoEvent{
		Msg: json.RawMessage(`not json`),
	}
	msg := ProtoEventToMsg(event)
	if msg != nil {
		testing.Errorf("expected nil for invalid JSON, got %T", msg)
	}
}
