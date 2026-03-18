package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/state"
)

func TestAppHandlesArrowKeys(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "First")
	store.Add("t-2", "Second")

	app := NewAppModel(store)

	rightKey := tea.KeyMsg{Type: tea.KeyRight}
	updated, _ := app.Update(rightKey)
	appModel := updated.(AppModel)

	if appModel.canvas.SelectedIndex() != 1 {
		t.Errorf("expected index 1 after right, got %d", appModel.canvas.SelectedIndex())
	}
}

func TestAppHandlesSessionConfigured(t *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	msg := SessionConfiguredMsg{
		SessionID: "s-1",
		Model:     "o4-mini",
	}
	updated, _ := app.Update(msg)
	appModel := updated.(AppModel)

	if appModel.sessionID != "s-1" {
		t.Errorf("expected sessionID s-1, got %s", appModel.sessionID)
	}

	threads := store.All()
	if len(threads) != 1 {
		t.Fatalf("expected 1 thread, got %d", len(threads))
	}
	if threads[0].ID != "s-1" {
		t.Errorf("expected thread s-1, got %s", threads[0].ID)
	}
}

func TestAppHandlesTaskStarted(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("s-1", "Test")

	app := NewAppModel(store)
	app.sessionID = "s-1"

	updated, _ := app.Update(TaskStartedMsg{})
	appModel := updated.(AppModel)

	thread, _ := store.Get("s-1")
	if thread.Status != state.StatusActive {
		t.Errorf("expected active, got %s", thread.Status)
	}
	if appModel.currentMessageID == "" {
		t.Error("expected currentMessageID to be set")
	}
	if len(thread.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(thread.Messages))
	}
}

func TestAppHandlesAgentDelta(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("s-1", "Test")

	app := NewAppModel(store)
	app.sessionID = "s-1"
	app.currentMessageID = "msg-1"

	thread, _ := store.Get("s-1")
	thread.AppendMessage(state.ChatMessage{ID: "msg-1", Role: "assistant"})

	updated, _ := app.Update(AgentDeltaMsg{Delta: "Hello"})
	_ = updated.(AppModel)

	if thread.Messages[0].Content != "Hello" {
		t.Errorf("expected Hello, got %s", thread.Messages[0].Content)
	}
}

func TestAppHandlesTaskComplete(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("s-1", "Test")

	app := NewAppModel(store)
	app.sessionID = "s-1"
	app.currentMessageID = "msg-1"

	updated, _ := app.Update(TaskCompleteMsg{LastMessage: "Done"})
	appModel := updated.(AppModel)

	thread, _ := store.Get("s-1")
	if thread.Status != state.StatusCompleted {
		t.Errorf("expected completed, got %s", thread.Status)
	}
	if appModel.currentMessageID != "" {
		t.Error("expected currentMessageID to be cleared")
	}
}

func TestAppToggleCanvasMode(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Test")

	app := NewAppModel(store)

	if app.CanvasMode() != CanvasModeGrid {
		t.Errorf("expected grid mode, got %d", app.CanvasMode())
	}

	tKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}
	updated, _ := app.Update(tKey)
	appModel := updated.(AppModel)

	if appModel.CanvasMode() != CanvasModeTree {
		t.Errorf("expected tree mode, got %d", appModel.CanvasMode())
	}
}

func TestAppTreeNavigationWhenFocused(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "First")
	store.Add("t-2", "Second")

	app := NewAppModel(store)

	toggleKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}
	updated, _ := app.Update(toggleKey)
	app = updated.(AppModel)

	downKey := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ = app.Update(downKey)
	app = updated.(AppModel)

	if app.tree.SelectedID() != "t-2" {
		t.Errorf("expected tree at t-2, got %s", app.tree.SelectedID())
	}
}

func TestAppHandlesQuit(t *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	quitKey := tea.KeyMsg{Type: tea.KeyCtrlC}
	_, cmd := app.Update(quitKey)

	if cmd == nil {
		t.Fatal("expected quit command")
	}
}

func TestAppEnterOpensSession(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Test Task")

	app := NewAppModel(store, WithInteractiveCommand("cat"))

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	appModel := updated.(AppModel)
	defer appModel.StopAllPTYSessions()

	if appModel.FocusPane() != FocusPaneSession {
		t.Errorf("expected session focus, got %d", appModel.FocusPane())
	}

	if len(appModel.sessionPanel.PinnedSessions()) != 1 {
		t.Fatalf("expected 1 pinned session, got %d", len(appModel.sessionPanel.PinnedSessions()))
	}

	_, hasPTY := appModel.ptySessions["t-1"]
	if !hasPTY {
		t.Error("expected PTY session to be stored for thread t-1")
	}
}

func TestAppEscClosesSession(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Test Task")

	app := NewAppModel(store, WithInteractiveCommand("cat"))

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	appModel := updated.(AppModel)
	defer appModel.StopAllPTYSessions()

	escKey := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ = appModel.Update(escKey)
	appModel = updated.(AppModel)

	if appModel.FocusPane() != FocusPaneCanvas {
		t.Errorf("expected canvas focus after Esc, got %d", appModel.FocusPane())
	}

	_, hasPTY := appModel.ptySessions["t-1"]
	if !hasPTY {
		t.Error("expected PTY session to stay alive after Esc")
	}

	if len(appModel.sessionPanel.PinnedSessions()) != 1 {
		t.Error("expected session to remain pinned after Esc")
	}
}

func TestAppEnterWithNoThreadsDoesNothing(t *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	appModel := updated.(AppModel)

	if appModel.FocusPane() != FocusPaneCanvas {
		t.Errorf("expected canvas focus when no threads, got %d", appModel.FocusPane())
	}
}

func TestAppCtrlBMOpensMenu(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Test")

	app := NewAppModel(store)

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	updated, _ := app.Update(ctrlB)
	app = updated.(AppModel)

	mKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	updated, _ = app.Update(mKey)
	app = updated.(AppModel)

	if !app.MenuVisible() {
		t.Error("expected menu to be visible")
	}
}

func TestAppMenuEscCloses(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Test")

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
		t.Error("expected menu hidden after Esc")
	}
}

func TestAppCtrlBEscCancelsPrefix(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Test")

	app := NewAppModel(store)

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	updated, _ := app.Update(ctrlB)
	app = updated.(AppModel)

	escKey := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ = app.Update(escKey)
	app = updated.(AppModel)

	if app.MenuVisible() {
		t.Error("expected menu not visible after prefix cancel")
	}
}

func TestAppMenuNavigation(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Test")

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

	if app.menu.SelectedIndex() != 1 {
		t.Errorf("expected menu index 1, got %d", app.menu.SelectedIndex())
	}
}

func TestAppHelpToggle(t *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	helpKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	updated, _ := app.Update(helpKey)
	appModel := updated.(AppModel)

	if !appModel.HelpVisible() {
		t.Error("expected help to be visible")
	}

	updated, _ = appModel.Update(helpKey)
	appModel = updated.(AppModel)

	if appModel.HelpVisible() {
		t.Error("expected help to be hidden")
	}
}

func TestAppHelpEscCloses(t *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	helpKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	updated, _ := app.Update(helpKey)
	appModel := updated.(AppModel)

	escKey := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ = appModel.Update(escKey)
	appModel = updated.(AppModel)

	if appModel.HelpVisible() {
		t.Error("expected help hidden after Esc")
	}
}

func TestAppNewThread(t *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	nKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	_, cmd := app.Update(nKey)

	if cmd == nil {
		t.Error("expected command for thread creation")
	}
}

func TestAppHandlesThreadCreatedMsg(t *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	msg := ThreadCreatedMsg{ThreadID: "t-new", Title: "New Thread"}
	updated, _ := app.Update(msg)
	_ = updated.(AppModel)

	threads := store.All()
	if len(threads) != 1 {
		t.Fatalf("expected 1 thread, got %d", len(threads))
	}
	if threads[0].ID != "t-new" {
		t.Errorf("expected thread t-new, got %s", threads[0].ID)
	}
}

func TestAppHandlesAgentDeltaWithoutSession(t *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	updated, _ := app.Update(AgentDeltaMsg{Delta: "test"})
	_ = updated.(AppModel)
}

func TestAppAutoApprovesExecRequest(t *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	msg := ExecApprovalRequestMsg{EventID: "req-1", Command: "ls"}
	updated, _ := app.Update(msg)
	_ = updated.(AppModel)
}

func TestAppForwardKeyToPTY(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Test")

	app := NewAppModel(store, WithInteractiveCommand("cat"))

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	if app.FocusPane() != FocusPaneSession {
		t.Fatal("expected session focus after Enter")
	}

	letterKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	updated, _ = app.Update(letterKey)
	app = updated.(AppModel)

	if app.FocusPane() != FocusPaneSession {
		t.Errorf("expected session focus maintained, got %d", app.FocusPane())
	}
}

func TestAppReconnectsExistingPTY(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Test")

	app := NewAppModel(store, WithInteractiveCommand("cat"))

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	escKey := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ = app.Update(escKey)
	app = updated.(AppModel)

	if app.FocusPane() != FocusPaneCanvas {
		t.Errorf("expected canvas focus, got %d", app.FocusPane())
	}

	updated, _ = app.Update(enterKey)
	app = updated.(AppModel)

	if app.FocusPane() != FocusPaneSession {
		t.Errorf("expected session focus on reconnect, got %d", app.FocusPane())
	}

	if len(app.ptySessions) != 1 {
		t.Errorf("expected 1 PTY session (reused), got %d", len(app.ptySessions))
	}
}

func TestAppHandlesPTYOutput(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Test")

	app := NewAppModel(store, WithInteractiveCommand("cat"))

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	exitMsg := PTYOutputMsg{ThreadID: "t-1", Exited: true}
	updated, _ = app.Update(exitMsg)
	_ = updated.(AppModel)
}

func TestAppSpacePinsSession(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Thread 1")
	app := NewAppModel(store, WithInteractiveCommand("echo", "hello"))

	spaceKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := app.Update(spaceKey)
	appModel := updated.(AppModel)
	defer appModel.StopAllPTYSessions()

	if len(appModel.sessionPanel.PinnedSessions()) != 1 {
		t.Fatalf("expected 1 pinned, got %d", len(appModel.sessionPanel.PinnedSessions()))
	}
	if appModel.sessionPanel.PinnedSessions()[0] != "t-1" {
		t.Errorf("expected t-1, got %s", appModel.sessionPanel.PinnedSessions()[0])
	}
}

func TestAppSpaceUnpinsSession(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Thread 1")
	app := NewAppModel(store, WithInteractiveCommand("echo", "hello"))

	spaceKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := app.Update(spaceKey)
	appModel := updated.(AppModel)

	updated2, _ := appModel.Update(spaceKey)
	appModel2 := updated2.(AppModel)
	defer appModel2.StopAllPTYSessions()

	if len(appModel2.sessionPanel.PinnedSessions()) != 0 {
		t.Errorf("expected 0 pinned after unpin, got %d", len(appModel2.sessionPanel.PinnedSessions()))
	}
}

func TestAppTabSwitchesToSessionPanel(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Thread 1")
	app := NewAppModel(store, WithInteractiveCommand("echo", "hello"))

	spaceKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := app.Update(spaceKey)
	app = updated.(AppModel)

	tabKey := tea.KeyMsg{Type: tea.KeyTab}
	updated, _ = app.Update(tabKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	if app.FocusPane() != FocusPaneSession {
		t.Errorf("expected FocusPaneSession, got %d", app.FocusPane())
	}
}

func TestAppTabDoesNothingWithNoPinnedSessions(t *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	tabKey := tea.KeyMsg{Type: tea.KeyTab}
	updated, _ := app.Update(tabKey)
	app = updated.(AppModel)

	if app.FocusPane() != FocusPaneCanvas {
		t.Errorf("expected FocusPaneCanvas, got %d", app.FocusPane())
	}
}

func TestAppCtrlBXUnpinsSession(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Thread 1")
	app := NewAppModel(store, WithInteractiveCommand("echo", "hello"))

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
		t.Errorf("expected 0 pinned after unpin, got %d", len(app.sessionPanel.PinnedSessions()))
	}
	if app.FocusPane() != FocusPaneCanvas {
		t.Errorf("expected focus back to canvas, got %d", app.FocusPane())
	}
}

func TestAppCtrlBZTogglesZoom(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Thread 1")
	app := NewAppModel(store, WithInteractiveCommand("echo", "hello"))

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
		t.Error("expected zoomed after Ctrl+B z")
	}
}

func TestAppCtrlBRightCyclesPaneRight(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Thread 1")
	store.Add("t-2", "Thread 2")
	app := NewAppModel(store, WithInteractiveCommand("echo", "hello"))

	space := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := app.Update(space)
	app = updated.(AppModel)

	app.canvas.MoveRight()
	updated, _ = app.Update(space)
	app = updated.(AppModel)

	tab := tea.KeyMsg{Type: tea.KeyTab}
	updated, _ = app.Update(tab)
	app = updated.(AppModel)

	if app.sessionPanel.ActivePaneIdx() != 0 {
		t.Fatalf("expected active pane 0, got %d", app.sessionPanel.ActivePaneIdx())
	}

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	updated, _ = app.Update(ctrlB)
	app = updated.(AppModel)

	rightKey := tea.KeyMsg{Type: tea.KeyRight}
	updated, _ = app.Update(rightKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	if app.sessionPanel.ActivePaneIdx() != 1 {
		t.Errorf("expected active pane 1, got %d", app.sessionPanel.ActivePaneIdx())
	}
}

func TestAppViewShowsDividerWhenPinned(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Thread 1")
	app := NewAppModel(store, WithInteractiveCommand("echo", "hello"))
	app.width = 120
	app.height = 40

	spaceKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := app.Update(spaceKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	view := app.View()
	if !strings.Contains(view, "─") {
		t.Error("expected divider line in view when sessions pinned")
	}
}

func TestHelpShowsPinKeybinding(t *testing.T) {
	help := NewHelpModel()
	view := help.View()
	if !strings.Contains(view, "Space") {
		t.Error("expected Space keybinding in help")
	}
	if !strings.Contains(view, "Ctrl+B x") {
		t.Error("expected Ctrl+B x keybinding in help")
	}
	if !strings.Contains(view, "Ctrl+B z") {
		t.Error("expected Ctrl+B z keybinding in help")
	}
}

func TestAppFocusPaneDefaultsToCanvas(t *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	if app.FocusPane() != FocusPaneCanvas {
		t.Errorf("expected FocusPaneCanvas, got %d", app.FocusPane())
	}
}

func TestAppHasPinnedSessions(t *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	if len(app.sessionPanel.PinnedSessions()) != 0 {
		t.Errorf("expected 0 pinned sessions, got %d", len(app.sessionPanel.PinnedSessions()))
	}
}
