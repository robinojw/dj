package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/state"
)

const (
	errExpected1Pinned     = "expected 1 pinned session, got %d"
	errExpectedSessionFocus = "expected session focus, got %d"
	errExpectedStrFmt      = "expected %s, got %s"
)

func TestAppEnterOpensSession(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testTitleTestTask)

	app := NewAppModel(store, WithInteractiveCommand(testCommandCat))

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	appModel := updated.(AppModel)
	defer appModel.StopAllPTYSessions()

	if appModel.FocusPane() != FocusPaneSession {
		test.Errorf(errExpectedSessionFocus, appModel.FocusPane())
	}

	if len(appModel.sessionPanel.PinnedSessions()) != 1 {
		test.Fatalf(errExpected1Pinned, len(appModel.sessionPanel.PinnedSessions()))
	}

	_, hasPTY := appModel.ptySessions[testThreadID1]
	if !hasPTY {
		test.Errorf("expected PTY session for thread %s", testThreadID1)
	}
}

func TestAppEscClosesSession(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testTitleTestTask)

	app := NewAppModel(store, WithInteractiveCommand(testCommandCat))

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	appModel := updated.(AppModel)
	defer appModel.StopAllPTYSessions()

	escKey := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ = appModel.Update(escKey)
	appModel = updated.(AppModel)

	if appModel.FocusPane() != FocusPaneCanvas {
		test.Errorf("expected canvas focus after Esc, got %d", appModel.FocusPane())
	}

	_, hasPTY := appModel.ptySessions[testThreadID1]
	if !hasPTY {
		test.Error("expected PTY session to stay alive after Esc")
	}

	if len(appModel.sessionPanel.PinnedSessions()) != 1 {
		test.Error("expected session to remain pinned after Esc")
	}
}

func TestAppEnterWithNoThreadsDoesNothing(test *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	appModel := updated.(AppModel)

	if appModel.FocusPane() != FocusPaneCanvas {
		test.Errorf("expected canvas focus when no threads, got %d", appModel.FocusPane())
	}
}

func TestAppForwardKeyToPTY(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testTitleTest)

	app := NewAppModel(store, WithInteractiveCommand(testCommandCat))

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	if app.FocusPane() != FocusPaneSession {
		test.Fatal("expected session focus after Enter")
	}

	letterKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	updated, _ = app.Update(letterKey)
	app = updated.(AppModel)

	if app.FocusPane() != FocusPaneSession {
		test.Errorf("expected session focus maintained, got %d", app.FocusPane())
	}
}

func TestAppReconnectsExistingPTY(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testTitleTest)

	app := NewAppModel(store, WithInteractiveCommand(testCommandCat))

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	escKey := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ = app.Update(escKey)
	app = updated.(AppModel)

	if app.FocusPane() != FocusPaneCanvas {
		test.Errorf("expected canvas focus, got %d", app.FocusPane())
	}

	updated, _ = app.Update(enterKey)
	app = updated.(AppModel)

	if app.FocusPane() != FocusPaneSession {
		test.Errorf("expected session focus on reconnect, got %d", app.FocusPane())
	}

	if len(app.ptySessions) != 1 {
		test.Errorf("expected 1 PTY session (reused), got %d", len(app.ptySessions))
	}
}

func TestAppHandlesPTYOutput(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testTitleTest)

	app := NewAppModel(store, WithInteractiveCommand(testCommandCat))

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	exitMsg := PTYOutputMsg{ThreadID: testThreadID1, Exited: true}
	updated, _ = app.Update(exitMsg)
	_ = updated.(AppModel)
}

func TestAppSpacePinsSession(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testTitleThread1)
	app := NewAppModel(store, WithInteractiveCommand(testCommandEcho, testArgHello))

	spaceKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := app.Update(spaceKey)
	appModel := updated.(AppModel)
	defer appModel.StopAllPTYSessions()

	if len(appModel.sessionPanel.PinnedSessions()) != 1 {
		test.Fatalf("expected 1 pinned, got %d", len(appModel.sessionPanel.PinnedSessions()))
	}
	if appModel.sessionPanel.PinnedSessions()[0] != testThreadID1 {
		test.Errorf(errExpectedStrFmt, testThreadID1, appModel.sessionPanel.PinnedSessions()[0])
	}
}

func TestAppSpaceUnpinsSession(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testTitleThread1)
	app := NewAppModel(store, WithInteractiveCommand(testCommandEcho, testArgHello))

	spaceKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := app.Update(spaceKey)
	appModel := updated.(AppModel)

	updated2, _ := appModel.Update(spaceKey)
	appModel2 := updated2.(AppModel)
	defer appModel2.StopAllPTYSessions()

	if len(appModel2.sessionPanel.PinnedSessions()) != 0 {
		test.Errorf("expected 0 pinned after unpin, got %d", len(appModel2.sessionPanel.PinnedSessions()))
	}
}

func TestAppTabSwitchesToSessionPanel(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testTitleThread1)
	app := NewAppModel(store, WithInteractiveCommand(testCommandEcho, testArgHello))

	spaceKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := app.Update(spaceKey)
	app = updated.(AppModel)

	tabKey := tea.KeyMsg{Type: tea.KeyTab}
	updated, _ = app.Update(tabKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	if app.FocusPane() != FocusPaneSession {
		test.Errorf("expected FocusPaneSession, got %d", app.FocusPane())
	}
}

func TestAppTabDoesNothingWithNoPinnedSessions(test *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	tabKey := tea.KeyMsg{Type: tea.KeyTab}
	updated, _ := app.Update(tabKey)
	app = updated.(AppModel)

	if app.FocusPane() != FocusPaneCanvas {
		test.Errorf("expected FocusPaneCanvas, got %d", app.FocusPane())
	}
}

func TestAppNewThreadCreatesAndOpensSession(test *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store, WithInteractiveCommand(testCommandCat))
	app.width = testAppWidth
	app.height = testAppHeight

	nKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updated, cmd := app.Update(nKey)
	app = updated.(AppModel)

	if cmd == nil {
		test.Fatal("expected command from n key")
	}

	msg := cmd()
	updated, _ = app.Update(msg)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	threads := store.All()
	if len(threads) != 1 {
		test.Fatalf(errExpected1Thread, len(threads))
	}

	if app.FocusPane() != FocusPaneSession {
		test.Errorf(errExpectedSessionFocus, app.FocusPane())
	}

	if len(app.sessionPanel.PinnedSessions()) != 1 {
		test.Errorf(errExpected1Pinned, len(app.sessionPanel.PinnedSessions()))
	}
}

func TestAppNewThreadIncrementsTitle(test *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store, WithInteractiveCommand(testCommandCat))
	app.width = testAppWidth
	app.height = testAppHeight

	nKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	escKey := tea.KeyMsg{Type: tea.KeyEsc}

	updated, cmd := app.Update(nKey)
	app = updated.(AppModel)
	msg := cmd()
	updated, _ = app.Update(msg)
	app = updated.(AppModel)

	updated, _ = app.Update(escKey)
	app = updated.(AppModel)

	updated, cmd = app.Update(nKey)
	app = updated.(AppModel)
	msg = cmd()
	updated, _ = app.Update(msg)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	threads := store.All()
	if len(threads) != testExpectedTwo {
		test.Fatalf("expected %d threads, got %d", testExpectedTwo, len(threads))
	}
	if threads[0].Title != testTitleSession1 {
		test.Errorf(errExpectedStrFmt, testTitleSession1, threads[0].Title)
	}
	if threads[1].Title != testTitleSession2 {
		test.Errorf(errExpectedStrFmt, testTitleSession2, threads[1].Title)
	}
}
