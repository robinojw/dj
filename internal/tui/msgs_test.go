package tui

import (
	"fmt"
	"testing"
)

func TestMsgTypes(t *testing.T) {
	configuredMsg := SessionConfiguredMsg{
		SessionID: "s-1",
		Model:     "o4-mini",
	}
	if configuredMsg.SessionID != "s-1" {
		t.Errorf("expected s-1, got %s", configuredMsg.SessionID)
	}

	deltaMsg := AgentDeltaMsg{Delta: "hello"}
	if deltaMsg.Delta != "hello" {
		t.Errorf("expected hello, got %s", deltaMsg.Delta)
	}

	completeMsg := TaskCompleteMsg{LastMessage: "done"}
	if completeMsg.LastMessage != "done" {
		t.Errorf("expected done, got %s", completeMsg.LastMessage)
	}

	errorMsg := AppServerErrorMsg{Err: fmt.Errorf("test error")}
	if errorMsg.Error() != "test error" {
		t.Errorf("expected test error, got %s", errorMsg.Error())
	}

	createdMsg := ThreadCreatedMsg{ThreadID: "t-1", Title: "Test"}
	if createdMsg.ThreadID != "t-1" {
		t.Errorf("expected t-1, got %s", createdMsg.ThreadID)
	}
}

func TestPinUnpinMessages(t *testing.T) {
	pinMsg := PinSessionMsg{ThreadID: "t-1"}
	if pinMsg.ThreadID != "t-1" {
		t.Errorf("expected t-1, got %s", pinMsg.ThreadID)
	}

	unpinMsg := UnpinSessionMsg{ThreadID: "t-2"}
	if unpinMsg.ThreadID != "t-2" {
		t.Errorf("expected t-2, got %s", unpinMsg.ThreadID)
	}

	focusMsg := FocusSessionPaneMsg{Index: 2}
	if focusMsg.Index != 2 {
		t.Errorf("expected 2, got %d", focusMsg.Index)
	}

	switchMsg := SwitchPaneFocusMsg{Pane: FocusPaneSession}
	if switchMsg.Pane != FocusPaneSession {
		t.Errorf("expected FocusPaneSession, got %d", switchMsg.Pane)
	}
}
