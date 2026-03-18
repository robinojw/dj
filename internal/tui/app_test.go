package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/state"
)

const (
	testSessionID1      = "s-1"
	testNewThreadID     = "t-new"
	testTitleTest       = "Test"
	testTitleTestTask   = "Test Task"
	testTitleThread1    = "Thread 1"
	testTitleThread2    = "Thread 2"
	testTitleNewThread  = "New Thread"
	testTitleSession1   = "Session 1"
	testTitleSession2   = "Session 2"
	testCommandCat      = "cat"
	testCommandEcho     = "echo"
	testArgHello        = "hello"
	testMessageID1      = "msg-1"
	testModelName       = "o4-mini"
	testEventID1        = "req-1"
	testCommandLS       = "ls"
	testDeltaHello      = "Hello"
	testDeltaTest       = "test"
	testLastMessageDone = "Done"
	testRoleAssistant   = "assistant"
	testAppWidth        = 120
	testAppHeight       = 40
	testExpectedTwo     = 2
)

func TestAppHandlesArrowKeys(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testThreadTitle1)
	store.Add(testThreadID2, testThreadTitle2)

	app := NewAppModel(store)

	rightKey := tea.KeyMsg{Type: tea.KeyRight}
	updated, _ := app.Update(rightKey)
	appModel := updated.(AppModel)

	if appModel.canvas.SelectedIndex() != 1 {
		test.Errorf("expected index 1 after right, got %d", appModel.canvas.SelectedIndex())
	}
}

func TestAppToggleCanvasMode(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testTitleTest)

	app := NewAppModel(store)

	if app.CanvasMode() != CanvasModeGrid {
		test.Errorf("expected grid mode, got %d", app.CanvasMode())
	}

	tKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}
	updated, _ := app.Update(tKey)
	appModel := updated.(AppModel)

	if appModel.CanvasMode() != CanvasModeTree {
		test.Errorf("expected tree mode, got %d", appModel.CanvasMode())
	}
}

func TestAppTreeNavigationWhenFocused(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testThreadTitle1)
	store.Add(testThreadID2, testThreadTitle2)

	app := NewAppModel(store)

	toggleKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}
	updated, _ := app.Update(toggleKey)
	app = updated.(AppModel)

	downKey := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ = app.Update(downKey)
	app = updated.(AppModel)

	if app.tree.SelectedID() != testThreadID2 {
		test.Errorf("expected tree at %s, got %s", testThreadID2, app.tree.SelectedID())
	}
}

func TestAppHandlesQuit(test *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	quitKey := tea.KeyMsg{Type: tea.KeyCtrlC}
	_, cmd := app.Update(quitKey)

	if cmd == nil {
		test.Fatal("expected quit command")
	}
}

func TestAppHelpToggle(test *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	helpKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	updated, _ := app.Update(helpKey)
	appModel := updated.(AppModel)

	if !appModel.HelpVisible() {
		test.Error("expected help to be visible")
	}

	updated, _ = appModel.Update(helpKey)
	appModel = updated.(AppModel)

	if appModel.HelpVisible() {
		test.Error("expected help to be hidden")
	}
}

func TestAppHelpEscCloses(test *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	helpKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	updated, _ := app.Update(helpKey)
	appModel := updated.(AppModel)

	escKey := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ = appModel.Update(escKey)
	appModel = updated.(AppModel)

	if appModel.HelpVisible() {
		test.Error("expected help hidden after Esc")
	}
}

func TestAppNewThread(test *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	nKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	_, cmd := app.Update(nKey)

	if cmd == nil {
		test.Error("expected command for thread creation")
	}
}

func TestAppFocusPaneDefaultsToCanvas(test *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	if app.FocusPane() != FocusPaneCanvas {
		test.Errorf("expected FocusPaneCanvas, got %d", app.FocusPane())
	}
}

func TestAppHasPinnedSessions(test *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	if len(app.sessionPanel.PinnedSessions()) != 0 {
		test.Errorf("expected 0 pinned sessions, got %d", len(app.sessionPanel.PinnedSessions()))
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
