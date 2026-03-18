package tui

import (
	"fmt"
	"testing"
)

const (
	msgsTestError      = "test error"
	msgsTestThreadID1  = "t-1"
	msgsTestThreadID2  = "t-2"
	msgsTestExpectedFmt = "expected t-1, got %s"
	msgsTestFocusIndex = 2
)

func TestMsgTypes(test *testing.T) {
	errorMsg := AppServerErrorMsg{Err: fmt.Errorf("%s", msgsTestError)}
	if errorMsg.Error() != msgsTestError {
		test.Errorf("expected test error, got %s", errorMsg.Error())
	}

	createdMsg := ThreadCreatedMsg{ThreadID: msgsTestThreadID1, Title: "Test"}
	if createdMsg.ThreadID != msgsTestThreadID1 {
		test.Errorf(msgsTestExpectedFmt, createdMsg.ThreadID)
	}
}

func TestPinUnpinMessages(test *testing.T) {
	pinMsg := PinSessionMsg{ThreadID: msgsTestThreadID1}
	if pinMsg.ThreadID != msgsTestThreadID1 {
		test.Errorf(msgsTestExpectedFmt, pinMsg.ThreadID)
	}

	unpinMsg := UnpinSessionMsg{ThreadID: msgsTestThreadID2}
	if unpinMsg.ThreadID != msgsTestThreadID2 {
		test.Errorf("expected t-2, got %s", unpinMsg.ThreadID)
	}

	focusMsg := FocusSessionPaneMsg{Index: msgsTestFocusIndex}
	if focusMsg.Index != msgsTestFocusIndex {
		test.Errorf("expected 2, got %d", focusMsg.Index)
	}

	switchMsg := SwitchPaneFocusMsg{Pane: FocusPaneSession}
	if switchMsg.Pane != FocusPaneSession {
		test.Errorf("expected FocusPaneSession, got %d", switchMsg.Pane)
	}
}
