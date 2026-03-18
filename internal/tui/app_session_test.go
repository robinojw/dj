package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/state"
)

func TestAppEnterOpensSession(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(appTestThreadID1, appTestTitleTestTask)

	app := NewAppModel(store, WithInteractiveCommand(appTestCmdCat))

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	appModel := updated.(AppModel)
	defer appModel.StopAllPTYSessions()

	if appModel.FocusPane() != FocusPaneSession {
		test.Errorf(appTestExpectSessionFocus, appModel.FocusPane())
	}

	if len(appModel.sessionPanel.PinnedSessions()) != 1 {
		test.Fatalf(appTestExpected1Pinned, len(appModel.sessionPanel.PinnedSessions()))
	}

	_, hasPTY := appModel.ptySessions[appTestThreadID1]
	if !hasPTY {
		test.Error("expected PTY session to be stored for thread t-1")
	}
}

func TestAppEscClosesSession(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(appTestThreadID1, appTestTitleTestTask)

	app := NewAppModel(store, WithInteractiveCommand(appTestCmdCat))

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

	_, hasPTY := appModel.ptySessions[appTestThreadID1]
	if !hasPTY {
		test.Error("expected PTY session to stay alive after Esc")
	}

	if len(appModel.sessionPanel.PinnedSessions()) != 1 {
		test.Error("expected session to remain pinned after Esc")
	}
}

func TestAppForwardKeyToPTY(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(appTestThreadID1, appTestTitleTest)

	app := NewAppModel(store, WithInteractiveCommand(appTestCmdCat))

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
	store.Add(appTestThreadID1, appTestTitleTest)

	app := NewAppModel(store, WithInteractiveCommand(appTestCmdCat))

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
	store.Add(appTestThreadID1, appTestTitleTest)

	app := NewAppModel(store, WithInteractiveCommand(appTestCmdCat))

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	exitMsg := PTYOutputMsg{ThreadID: appTestThreadID1, Exited: true}
	updated, _ = app.Update(exitMsg)
	_ = updated.(AppModel)
}

func TestAppSpacePinsSession(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(appTestThreadID1, appTestTitleThread1)
	app := NewAppModel(store, WithInteractiveCommand(appTestCmdEcho, appTestArgHello))

	spaceKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := app.Update(spaceKey)
	appModel := updated.(AppModel)
	defer appModel.StopAllPTYSessions()

	if len(appModel.sessionPanel.PinnedSessions()) != 1 {
		test.Fatalf(appTestExpected1Pinned, len(appModel.sessionPanel.PinnedSessions()))
	}
	if appModel.sessionPanel.PinnedSessions()[0] != appTestThreadID1 {
		test.Errorf("expected t-1, got %s", appModel.sessionPanel.PinnedSessions()[0])
	}
}

func TestAppSpaceUnpinsSession(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(appTestThreadID1, appTestTitleThread1)
	app := NewAppModel(store, WithInteractiveCommand(appTestCmdEcho, appTestArgHello))

	spaceKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := app.Update(spaceKey)
	appModel := updated.(AppModel)

	updated, _ = appModel.Update(spaceKey)
	appModel = updated.(AppModel)
	defer appModel.StopAllPTYSessions()

	if len(appModel.sessionPanel.PinnedSessions()) != 0 {
		test.Errorf(appTestExpected0Pinned, len(appModel.sessionPanel.PinnedSessions()))
	}
}

func TestAppTabSwitchesToSessionPanel(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(appTestThreadID1, appTestTitleThread1)
	app := NewAppModel(store, WithInteractiveCommand(appTestCmdEcho, appTestArgHello))

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
		test.Errorf(appTestExpectCanvasFocus, app.FocusPane())
	}
}
