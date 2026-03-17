package tui

import (
	"encoding/json"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/appserver"
)

type mockSender struct {
	messages []tea.Msg
}

func (mock *mockSender) Send(msg tea.Msg) {
	mock.messages = append(mock.messages, msg)
}

func TestBridgeSessionConfigured(t *testing.T) {
	sender := &mockSender{}
	router := appserver.NewEventRouter()
	WireEventBridge(router, sender)

	event := appserver.Event{
		ID:  "",
		Msg: json.RawMessage(`{"type":"session_configured","session_id":"sess-1","model":"gpt-4o"}`),
	}
	router.HandleEvent(event)

	if len(sender.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(sender.messages))
	}
	msg, ok := sender.messages[0].(AppServerConnectedMsg)
	if !ok {
		t.Fatalf("expected AppServerConnectedMsg, got %T", sender.messages[0])
	}
	if msg.SessionID != "sess-1" {
		t.Errorf("expected sess-1, got %s", msg.SessionID)
	}
	if msg.Model != "gpt-4o" {
		t.Errorf("expected gpt-4o, got %s", msg.Model)
	}
}

func TestBridgeAgentMessageDelta(t *testing.T) {
	sender := &mockSender{}
	router := appserver.NewEventRouter()
	WireEventBridge(router, sender)

	event := appserver.Event{
		ID:  "sub-1",
		Msg: json.RawMessage(`{"type":"agent_message_delta","delta":"hello"}`),
	}
	router.HandleEvent(event)

	if len(sender.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(sender.messages))
	}
	msg, ok := sender.messages[0].(ThreadDeltaMsg)
	if !ok {
		t.Fatalf("expected ThreadDeltaMsg, got %T", sender.messages[0])
	}
	if msg.Delta != "hello" {
		t.Errorf("expected hello, got %s", msg.Delta)
	}
}

func TestBridgeExecCommandBegin(t *testing.T) {
	sender := &mockSender{}
	router := appserver.NewEventRouter()
	WireEventBridge(router, sender)

	event := appserver.Event{
		ID:  "sub-1",
		Msg: json.RawMessage(`{"type":"exec_command_begin","call_id":"cmd-1","command":"ls -la"}`),
	}
	router.HandleEvent(event)

	if len(sender.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(sender.messages))
	}
	msg, ok := sender.messages[0].(CommandOutputMsg)
	if !ok {
		t.Fatalf("expected CommandOutputMsg, got %T", sender.messages[0])
	}
	if msg.ExecID != "cmd-1" {
		t.Errorf("expected cmd-1, got %s", msg.ExecID)
	}
}

func TestBridgeServerError(t *testing.T) {
	sender := &mockSender{}
	router := appserver.NewEventRouter()
	WireEventBridge(router, sender)

	event := appserver.Event{
		ID:  "",
		Msg: json.RawMessage(`{"type":"error","message":"something broke"}`),
	}
	router.HandleEvent(event)

	if len(sender.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(sender.messages))
	}
	msg, ok := sender.messages[0].(AppServerErrorMsg)
	if !ok {
		t.Fatalf("expected AppServerErrorMsg, got %T", sender.messages[0])
	}
	if msg.Error() != "something broke" {
		t.Errorf("expected 'something broke', got %s", msg.Error())
	}
}
