package tui

import (
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

func TestAppHandlesThreadStatusMsg(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Initial")

	app := NewAppModel(store)

	msg := ThreadStatusMsg{
		ThreadID: "t-1",
		Status:   "active",
		Title:    "Running",
	}
	app.Update(msg)

	thread, _ := store.Get("t-1")
	if thread.Status != "active" {
		t.Errorf("expected active, got %s", thread.Status)
	}
	if thread.Title != "Running" {
		t.Errorf("expected Running, got %s", thread.Title)
	}
}

func TestAppToggleFocus(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Test")

	app := NewAppModel(store)

	if app.Focus() != FocusCanvas {
		t.Errorf("expected canvas focus, got %d", app.Focus())
	}

	tabKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}
	updated, _ := app.Update(tabKey)
	appModel := updated.(AppModel)

	if appModel.Focus() != FocusTree {
		t.Errorf("expected tree focus, got %d", appModel.Focus())
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

	app := NewAppModel(store)

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	appModel := updated.(AppModel)

	if appModel.Focus() != FocusSession {
		t.Errorf("expected session focus, got %d", appModel.Focus())
	}
}

func TestAppEscClosesSession(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Test Task")

	app := NewAppModel(store)

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	appModel := updated.(AppModel)

	escKey := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ = appModel.Update(escKey)
	appModel = updated.(AppModel)

	if appModel.Focus() != FocusCanvas {
		t.Errorf("expected canvas focus after Esc, got %d", appModel.Focus())
	}
}

func TestAppEnterWithNoThreadsDoesNothing(t *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	appModel := updated.(AppModel)

	if appModel.Focus() != FocusCanvas {
		t.Errorf("expected canvas focus when no threads, got %d", appModel.Focus())
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

func TestAppSessionRefreshesOnMessage(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Test")

	app := NewAppModel(store)

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	app = updated.(AppModel)

	msg := ThreadMessageMsg{
		ThreadID:  "t-1",
		MessageID: "m-1",
		Role:      "user",
		Content:   "Hello",
	}
	updated, _ = app.Update(msg)
	app = updated.(AppModel)

	if app.Focus() != FocusSession {
		t.Errorf("expected session focus maintained, got %d", app.Focus())
	}
}
