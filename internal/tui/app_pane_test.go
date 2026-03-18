package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/state"
)

func TestAppCtrlBXUnpinsSession(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(appTestThreadID1, appTestTitleThread1)
	app := NewAppModel(store, WithInteractiveCommand(appTestCmdEcho, appTestArgHello))

	spaceKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := app.Update(spaceKey)
	app = updated.(AppModel)

	tabKey := tea.KeyMsg{Type: tea.KeyTab}
	updated, _ = app.Update(tabKey)
	app = updated.(AppModel)

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	updated, _ = app.Update(ctrlB)
	app = updated.(AppModel)

	xKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	updated, _ = app.Update(xKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	if len(app.sessionPanel.PinnedSessions()) != 0 {
		test.Errorf(appTestExpected0Pinned, len(app.sessionPanel.PinnedSessions()))
	}
	if app.FocusPane() != FocusPaneCanvas {
		test.Errorf("expected focus back to canvas, got %d", app.FocusPane())
	}
}

func TestAppCtrlBZTogglesZoom(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(appTestThreadID1, appTestTitleThread1)
	app := NewAppModel(store, WithInteractiveCommand(appTestCmdEcho, appTestArgHello))

	spaceKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := app.Update(spaceKey)
	app = updated.(AppModel)

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	updated, _ = app.Update(ctrlB)
	app = updated.(AppModel)

	zKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}}
	updated, _ = app.Update(zKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	if !app.sessionPanel.Zoomed() {
		test.Error("expected zoomed after Ctrl+B z")
	}
}

func TestAppCtrlBRightCyclesPaneRight(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(appTestThreadID1, appTestTitleThread1)
	store.Add(appTestThreadID2, appTestTitleThread2)
	app := NewAppModel(store, WithInteractiveCommand(appTestCmdEcho, appTestArgHello))

	space := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := app.Update(space)
	app = updated.(AppModel)

	app.canvas.MoveRight()
	updated, _ = app.Update(space)
	app = updated.(AppModel)

	tab := tea.KeyMsg{Type: tea.KeyTab}
	updated, _ = app.Update(tab)
	app = updated.(AppModel)

	if app.sessionPanel.ActivePaneIdx() != appTestExpectedPaneIndex0 {
		test.Fatalf("expected active pane 0, got %d", app.sessionPanel.ActivePaneIdx())
	}

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	updated, _ = app.Update(ctrlB)
	app = updated.(AppModel)

	rightKey := tea.KeyMsg{Type: tea.KeyRight}
	updated, _ = app.Update(rightKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	if app.sessionPanel.ActivePaneIdx() != appTestExpectedPaneIndex1 {
		test.Errorf("expected active pane 1, got %d", app.sessionPanel.ActivePaneIdx())
	}
}

func TestAppViewShowsDividerWhenPinned(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(appTestThreadID1, appTestTitleThread1)
	app := NewAppModel(store, WithInteractiveCommand(appTestCmdEcho, appTestArgHello))
	app.width = appTestWidth
	app.height = appTestHeight

	spaceKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := app.Update(spaceKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	view := app.View()
	hasDivider := strings.Contains(view, "\u2500")
	if !hasDivider {
		test.Error("expected divider line in view when sessions pinned")
	}
}

func TestHelpShowsPinKeybinding(test *testing.T) {
	help := NewHelpModel()
	view := help.View()
	if !strings.Contains(view, "Space") {
		test.Error("expected Space keybinding in help")
	}
	if !strings.Contains(view, "Ctrl+B x") {
		test.Error("expected Ctrl+B x keybinding in help")
	}
	if !strings.Contains(view, "Ctrl+B z") {
		test.Error("expected Ctrl+B z keybinding in help")
	}
}

func TestHelpShowsKillKeybinding(test *testing.T) {
	help := NewHelpModel()
	view := help.View()
	if !strings.Contains(view, "Kill") {
		test.Error("expected Kill keybinding in help")
	}
}

func TestAppFocusPaneDefaultsToCanvas(test *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	if app.FocusPane() != FocusPaneCanvas {
		test.Errorf(appTestExpectCanvasFocus, app.FocusPane())
	}
}

func TestAppHasPinnedSessions(test *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	if len(app.sessionPanel.PinnedSessions()) != 0 {
		test.Errorf("expected 0 pinned sessions, got %d", len(app.sessionPanel.PinnedSessions()))
	}
}

func TestAppNewThreadCreatesAndOpensSession(test *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store, WithInteractiveCommand(appTestCmdCat))
	app.width = appTestWidth
	app.height = appTestHeight

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
		test.Fatalf(appTestExpected1Thread, len(threads))
	}

	if app.FocusPane() != FocusPaneSession {
		test.Errorf("expected session focus after new thread, got %d", app.FocusPane())
	}

	if len(app.sessionPanel.PinnedSessions()) != 1 {
		test.Errorf(appTestExpected1Pinned, len(app.sessionPanel.PinnedSessions()))
	}
}

func TestAppNewThreadIncrementsTitle(test *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store, WithInteractiveCommand(appTestCmdCat))
	app.width = appTestWidth
	app.height = appTestHeight

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
	if len(threads) != appTestExpectedThreads {
		test.Fatalf("expected 2 threads, got %d", len(threads))
	}
	if threads[0].Title != appTestTitleSession1 {
		test.Errorf("expected 'Session 1', got %s", threads[0].Title)
	}
	if threads[1].Title != appTestTitleSession2 {
		test.Errorf("expected 'Session 2', got %s", threads[1].Title)
	}
}
