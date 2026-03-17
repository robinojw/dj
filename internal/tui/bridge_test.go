package tui

import (
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

func TestBridgeThreadStatusChanged(t *testing.T) {
	sender := &mockSender{}
	router := appserver.NewNotificationRouter()
	WireEventBridge(router, sender)

	router.Handle(appserver.NotifyThreadStatusChanged,
		[]byte(`{"threadId":"t-1","status":"active","title":"Running"}`))

	if len(sender.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(sender.messages))
	}
	msg, ok := sender.messages[0].(ThreadStatusMsg)
	if !ok {
		t.Fatalf("expected ThreadStatusMsg, got %T", sender.messages[0])
	}
	if msg.ThreadID != "t-1" {
		t.Errorf("expected t-1, got %s", msg.ThreadID)
	}
}

func TestBridgeCommandOutput(t *testing.T) {
	sender := &mockSender{}
	router := appserver.NewNotificationRouter()
	WireEventBridge(router, sender)

	router.Handle(appserver.NotifyCommandOutput,
		[]byte(`{"threadId":"t-1","execId":"e-1","data":"hello\n"}`))

	if len(sender.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(sender.messages))
	}
	msg, ok := sender.messages[0].(CommandOutputMsg)
	if !ok {
		t.Fatalf("expected CommandOutputMsg, got %T", sender.messages[0])
	}
	if msg.Data != "hello\n" {
		t.Errorf("expected hello, got %s", msg.Data)
	}
}
