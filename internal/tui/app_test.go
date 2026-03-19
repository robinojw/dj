package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	poolpkg "github.com/robinojw/dj/internal/pool"
	"github.com/robinojw/dj/internal/state"
)

const (
	appTestThreadID1          = "t-1"
	appTestThreadID2          = "t-2"
	appTestThreadNew          = "t-new"
	appTestTitleFirst         = "First"
	appTestTitleSecond        = "Second"
	appTestTitleTest          = "Test"
	appTestTitleTestTask      = "Test Task"
	appTestTitleThread1       = "Thread 1"
	appTestTitleThread2       = "Thread 2"
	appTestTitleNewThread     = "New Thread"
	appTestTitleSession1      = "Session 1"
	appTestTitleSession2      = "Session 2"
	appTestCmdCat             = "cat"
	appTestCmdEcho            = "echo"
	appTestArgHello           = "hello"
	appTestWidth              = 120
	appTestHeight             = 40
	appTestExpectedThreads    = 2
	appTestExpectedIndex1     = 1
	appTestExpectedPaneIndex0 = 0
	appTestExpectedPaneIndex1 = 1
	appTestExpected1Pinned    = "expected 1 pinned session, got %d"
	appTestExpected0Pinned    = "expected 0 pinned after unpin, got %d"
	appTestExpected1Thread    = "expected 1 thread, got %d"
	appTestExpectSessionFocus = "expected session focus, got %d"
	appTestExpectCanvasFocus  = "expected FocusPaneCanvas, got %d"
	appTestExpectedStrFmt     = "expected %s, got %s"
	appTestPoolMaxAgents      = 10
)

func TestAppHandlesArrowKeys(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(appTestThreadID1, appTestTitleFirst)
	store.Add(appTestThreadID2, appTestTitleSecond)

	app := NewAppModel(store)

	rightKey := tea.KeyMsg{Type: tea.KeyRight}
	updated, _ := app.Update(rightKey)
	appModel := updated.(AppModel)

	if appModel.canvas.SelectedIndex() != appTestExpectedIndex1 {
		test.Errorf("expected index 1 after right, got %d", appModel.canvas.SelectedIndex())
	}
}

func TestAppToggleCanvasMode(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(appTestThreadID1, appTestTitleTest)

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
	store.Add(appTestThreadID1, appTestTitleFirst)
	store.Add(appTestThreadID2, appTestTitleSecond)

	app := NewAppModel(store)

	toggleKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}
	updated, _ := app.Update(toggleKey)
	app = updated.(AppModel)

	downKey := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ = app.Update(downKey)
	app = updated.(AppModel)

	if app.tree.SelectedID() != appTestThreadID2 {
		test.Errorf("expected tree at t-2, got %s", app.tree.SelectedID())
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

func TestAppCtrlBMOpensMenu(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(appTestThreadID1, appTestTitleTest)

	app := NewAppModel(store)

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	updated, _ := app.Update(ctrlB)
	app = updated.(AppModel)

	mKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	updated, _ = app.Update(mKey)
	app = updated.(AppModel)

	if !app.MenuVisible() {
		test.Error("expected menu to be visible")
	}
}

func TestAppMenuEscCloses(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(appTestThreadID1, appTestTitleTest)

	app := NewAppModel(store)

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	updated, _ := app.Update(ctrlB)
	app = updated.(AppModel)

	mKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	updated, _ = app.Update(mKey)
	app = updated.(AppModel)

	escKey := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ = app.Update(escKey)
	app = updated.(AppModel)

	if app.MenuVisible() {
		test.Error("expected menu hidden after Esc")
	}
}

func TestAppCtrlBEscCancelsPrefix(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(appTestThreadID1, appTestTitleTest)

	app := NewAppModel(store)

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	updated, _ := app.Update(ctrlB)
	app = updated.(AppModel)

	escKey := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ = app.Update(escKey)
	app = updated.(AppModel)

	if app.MenuVisible() {
		test.Error("expected menu not visible after prefix cancel")
	}
}

func TestAppMenuNavigation(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(appTestThreadID1, appTestTitleTest)

	app := NewAppModel(store)

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	updated, _ := app.Update(ctrlB)
	app = updated.(AppModel)

	mKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	updated, _ = app.Update(mKey)
	app = updated.(AppModel)

	downKey := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ = app.Update(downKey)
	app = updated.(AppModel)

	if app.menu.SelectedIndex() != appTestExpectedIndex1 {
		test.Errorf("expected menu index 1, got %d", app.menu.SelectedIndex())
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

func TestAppHandlesThreadCreatedMsg(test *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store, WithInteractiveCommand(appTestCmdCat))
	app.width = appTestWidth
	app.height = appTestHeight

	msg := ThreadCreatedMsg{ThreadID: appTestThreadNew, Title: appTestTitleNewThread}
	updated, _ := app.Update(msg)
	appModel := updated.(AppModel)
	defer appModel.StopAllPTYSessions()

	threads := store.All()
	if len(threads) != 1 {
		test.Fatalf(appTestExpected1Thread, len(threads))
	}
	if threads[0].ID != appTestThreadNew {
		test.Errorf("expected thread t-new, got %s", threads[0].ID)
	}
	if appModel.FocusPane() != FocusPaneSession {
		test.Errorf(appTestExpectSessionFocus, appModel.FocusPane())
	}
}

func TestNewAppModelWithPool(testing *testing.T) {
	store := state.NewThreadStore()
	agentPool := poolpkg.NewAgentPool(appTestCmdEcho, []string{}, nil, appTestPoolMaxAgents)
	app := NewAppModel(store, WithPool(agentPool))
	if app.pool == nil {
		testing.Error("expected pool to be set")
	}
}

func TestNewAppModelWithoutPool(testing *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)
	if app.pool != nil {
		testing.Error("expected pool to be nil for backward compatibility")
	}
}

