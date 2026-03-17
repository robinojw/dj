package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/state"
)

func TestForkActionMsg(t *testing.T) {
	msg := ForkThreadMsg{
		ParentID:     "t-1",
		Instructions: "Continue from here",
	}
	if msg.ParentID != "t-1" {
		t.Errorf("expected t-1, got %s", msg.ParentID)
	}
}

func TestDeleteActionMsg(t *testing.T) {
	msg := DeleteThreadMsg{ThreadID: "t-1"}
	if msg.ThreadID != "t-1" {
		t.Errorf("expected t-1, got %s", msg.ThreadID)
	}
}

func TestMenuEnterDispatchesFork(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Test")

	app := NewAppModel(store, nil)

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	updated, _ := app.Update(ctrlB)
	app = updated.(AppModel)

	mKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	updated, _ = app.Update(mKey)
	app = updated.(AppModel)

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := app.Update(enterKey)
	app = updated.(AppModel)

	if app.MenuVisible() {
		t.Error("expected menu closed after Enter")
	}

	if cmd == nil {
		t.Fatal("expected a command from fork action")
	}

	msg := cmd()
	forkMsg, ok := msg.(ForkThreadMsg)
	if !ok {
		t.Fatalf("expected ForkThreadMsg, got %T", msg)
	}
	if forkMsg.ParentID != "t-1" {
		t.Errorf("expected t-1, got %s", forkMsg.ParentID)
	}
}

func TestMenuEnterDispatchesDelete(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Test")

	app := NewAppModel(store, nil)

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	updated, _ := app.Update(ctrlB)
	app = updated.(AppModel)

	mKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	updated, _ = app.Update(mKey)
	app = updated.(AppModel)

	downKey := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ = app.Update(downKey)
	app = updated.(AppModel)

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := app.Update(enterKey)
	app = updated.(AppModel)

	if app.MenuVisible() {
		t.Error("expected menu closed after Enter")
	}

	if cmd == nil {
		t.Fatal("expected a command from delete action")
	}

	msg := cmd()
	deleteMsg, ok := msg.(DeleteThreadMsg)
	if !ok {
		t.Fatalf("expected DeleteThreadMsg, got %T", msg)
	}
	if deleteMsg.ThreadID != "t-1" {
		t.Errorf("expected t-1, got %s", deleteMsg.ThreadID)
	}
}
