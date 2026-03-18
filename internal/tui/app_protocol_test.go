package tui

import (
	"testing"

	"github.com/robinojw/dj/internal/state"
)

const (
	errExpected1Thread  = "expected 1 thread, got %d"
	errExpectedThreadID = "expected thread %s, got %s"
)

func TestAppHandlesSessionConfigured(test *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	msg := SessionConfiguredMsg{
		SessionID: testSessionID1,
		Model:     testModelName,
	}
	updated, _ := app.Update(msg)
	appModel := updated.(AppModel)

	if appModel.sessionID != testSessionID1 {
		test.Errorf("expected sessionID %s, got %s", testSessionID1, appModel.sessionID)
	}

	threads := store.All()
	if len(threads) != 1 {
		test.Fatalf(errExpected1Thread, len(threads))
	}
	if threads[0].ID != testSessionID1 {
		test.Errorf(errExpectedThreadID, testSessionID1, threads[0].ID)
	}
}

func TestAppHandlesTaskStarted(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testSessionID1, testTitleTest)

	app := NewAppModel(store)
	app.sessionID = testSessionID1

	updated, _ := app.Update(TaskStartedMsg{})
	appModel := updated.(AppModel)

	thread, _ := store.Get(testSessionID1)
	if thread.Status != state.StatusActive {
		test.Errorf("expected active, got %s", thread.Status)
	}
	if appModel.currentMessageID == "" {
		test.Error("expected currentMessageID to be set")
	}
	if len(thread.Messages) != 1 {
		test.Fatalf("expected 1 message, got %d", len(thread.Messages))
	}
}

func TestAppHandlesAgentDelta(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testSessionID1, testTitleTest)

	app := NewAppModel(store)
	app.sessionID = testSessionID1
	app.currentMessageID = testMessageID1

	thread, _ := store.Get(testSessionID1)
	thread.AppendMessage(state.ChatMessage{ID: testMessageID1, Role: testRoleAssistant})

	updated, _ := app.Update(AgentDeltaMsg{Delta: testDeltaHello})
	_ = updated.(AppModel)

	if thread.Messages[0].Content != testDeltaHello {
		test.Errorf("expected %s, got %s", testDeltaHello, thread.Messages[0].Content)
	}
}

func TestAppHandlesTaskComplete(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testSessionID1, testTitleTest)

	app := NewAppModel(store)
	app.sessionID = testSessionID1
	app.currentMessageID = testMessageID1

	updated, _ := app.Update(TaskCompleteMsg{LastMessage: testLastMessageDone})
	appModel := updated.(AppModel)

	thread, _ := store.Get(testSessionID1)
	if thread.Status != state.StatusCompleted {
		test.Errorf("expected completed, got %s", thread.Status)
	}
	if appModel.currentMessageID != "" {
		test.Error("expected currentMessageID to be cleared")
	}
}

func TestAppHandlesAgentDeltaWithoutSession(test *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	updated, _ := app.Update(AgentDeltaMsg{Delta: testDeltaTest})
	_ = updated.(AppModel)
}

func TestAppAutoApprovesExecRequest(test *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	msg := ExecApprovalRequestMsg{EventID: testEventID1, Command: testCommandLS}
	updated, _ := app.Update(msg)
	_ = updated.(AppModel)
}

func TestAppHandlesThreadCreatedMsg(test *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store, WithInteractiveCommand(testCommandCat))
	app.width = testAppWidth
	app.height = testAppHeight

	msg := ThreadCreatedMsg{ThreadID: testNewThreadID, Title: testTitleNewThread}
	updated, _ := app.Update(msg)
	appModel := updated.(AppModel)
	defer appModel.StopAllPTYSessions()

	threads := store.All()
	if len(threads) != 1 {
		test.Fatalf(errExpected1Thread, len(threads))
	}
	if threads[0].ID != testNewThreadID {
		test.Errorf(errExpectedThreadID, testNewThreadID, threads[0].ID)
	}
	if appModel.FocusPane() != FocusPaneSession {
		test.Errorf("expected session focus, got %d", appModel.FocusPane())
	}
}
