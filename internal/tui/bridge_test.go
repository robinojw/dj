package tui

import (
	"encoding/json"
	"testing"

	"github.com/robinojw/dj/internal/appserver"
)

const (
	bridgeExpectedTypeFormat = "expected %T, got %T"
	bridgeTestRequestID     = "req-1"
)

func TestBridgeSessionConfigured(test *testing.T) {
	message := appserver.JsonRpcMessage{
		Params: json.RawMessage(`{"type":"session_configured","session_id":"s-1","model":"o4-mini"}`),
	}
	msg := ProtoEventToMsg(message)
	configured, ok := msg.(SessionConfiguredMsg)
	if !ok {
		test.Fatalf(bridgeExpectedTypeFormat, SessionConfiguredMsg{}, msg)
	}
	if configured.SessionID != "s-1" {
		test.Errorf("expected s-1, got %s", configured.SessionID)
	}
	if configured.Model != "o4-mini" {
		test.Errorf("expected o4-mini, got %s", configured.Model)
	}
}

func TestBridgeTaskStarted(test *testing.T) {
	message := appserver.JsonRpcMessage{
		Params: json.RawMessage(`{"type":"task_started","model_context_window":200000}`),
	}
	msg := ProtoEventToMsg(message)
	_, ok := msg.(TaskStartedMsg)
	if !ok {
		test.Fatalf(bridgeExpectedTypeFormat, TaskStartedMsg{}, msg)
	}
}

func TestBridgeAgentDelta(test *testing.T) {
	message := appserver.JsonRpcMessage{
		Params: json.RawMessage(`{"type":"agent_message_delta","delta":"Hello"}`),
	}
	msg := ProtoEventToMsg(message)
	delta, ok := msg.(AgentDeltaMsg)
	if !ok {
		test.Fatalf(bridgeExpectedTypeFormat, AgentDeltaMsg{}, msg)
	}
	if delta.Delta != "Hello" {
		test.Errorf("expected Hello, got %s", delta.Delta)
	}
}

func TestBridgeAgentMessage(test *testing.T) {
	message := appserver.JsonRpcMessage{
		Params: json.RawMessage(`{"type":"agent_message","message":"Hello world"}`),
	}
	msg := ProtoEventToMsg(message)
	completed, ok := msg.(AgentMessageCompletedMsg)
	if !ok {
		test.Fatalf(bridgeExpectedTypeFormat, AgentMessageCompletedMsg{}, msg)
	}
	if completed.Message != "Hello world" {
		test.Errorf("expected Hello world, got %s", completed.Message)
	}
}

func TestBridgeTaskComplete(test *testing.T) {
	message := appserver.JsonRpcMessage{
		Params: json.RawMessage(`{"type":"task_complete","last_agent_message":"Done"}`),
	}
	msg := ProtoEventToMsg(message)
	complete, ok := msg.(TaskCompleteMsg)
	if !ok {
		test.Fatalf(bridgeExpectedTypeFormat, TaskCompleteMsg{}, msg)
	}
	if complete.LastMessage != "Done" {
		test.Errorf("expected Done, got %s", complete.LastMessage)
	}
}

func TestBridgeExecApproval(test *testing.T) {
	message := appserver.JsonRpcMessage{
		ID:     bridgeTestRequestID,
		Params: json.RawMessage(`{"type":"exec_command_request","command":"ls","cwd":"/tmp"}`),
	}
	msg := ProtoEventToMsg(message)
	approval, ok := msg.(ExecApprovalRequestMsg)
	if !ok {
		test.Fatalf(bridgeExpectedTypeFormat, ExecApprovalRequestMsg{}, msg)
	}
	if approval.EventID != bridgeTestRequestID {
		test.Errorf("expected %s, got %s", bridgeTestRequestID, approval.EventID)
	}
	if approval.Command != "ls" {
		test.Errorf("expected ls, got %s", approval.Command)
	}
}

func TestBridgeAgentReasoningDelta(test *testing.T) {
	message := appserver.JsonRpcMessage{
		Params: json.RawMessage(`{"type":"agent_reasoning_delta","delta":"Let me think..."}`),
	}
	msg := ProtoEventToMsg(message)
	reasoning, ok := msg.(AgentReasoningDeltaMsg)
	if !ok {
		test.Fatalf(bridgeExpectedTypeFormat, AgentReasoningDeltaMsg{}, msg)
	}
	if reasoning.Delta != "Let me think..." {
		test.Errorf("expected Let me think..., got %s", reasoning.Delta)
	}
}

func TestBridgeUnknownEventReturnsNil(test *testing.T) {
	message := appserver.JsonRpcMessage{
		Params: json.RawMessage(`{"type":"unknown_event"}`),
	}
	msg := ProtoEventToMsg(message)
	if msg != nil {
		test.Errorf("expected nil for unknown event, got %T", msg)
	}
}

func TestBridgeNilParamsReturnsNil(test *testing.T) {
	message := appserver.JsonRpcMessage{}
	msg := ProtoEventToMsg(message)
	if msg != nil {
		test.Errorf("expected nil for nil params, got %T", msg)
	}
}
