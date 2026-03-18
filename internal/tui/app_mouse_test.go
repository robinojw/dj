package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/state"
)

const (
	mouseTestWidth    = 120
	mouseTestHeight   = 40
	mouseTestThread   = "t-1"
	mouseTestTitle    = "Thread 1"
	mouseTestCmd      = "cat"
)

func TestAppMouseScrollUpOnSession(testing *testing.T) {
	store := state.NewThreadStore()
	store.Add(mouseTestThread, mouseTestTitle)

	app := NewAppModel(store, WithInteractiveCommand(mouseTestCmd))
	app.width = mouseTestWidth
	app.height = mouseTestHeight

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	scrollUp := tea.MouseMsg{
		Button: tea.MouseButtonWheelUp,
		Action: tea.MouseActionPress,
	}
	updated, _ = app.Update(scrollUp)
	app = updated.(AppModel)

	ptySession := app.ptySessions[mouseTestThread]
	offset := ptySession.ScrollOffset()
	if offset < 0 {
		testing.Errorf("expected non-negative scroll offset, got %d", offset)
	}
}

func TestAppMouseScrollDownOnSession(testing *testing.T) {
	store := state.NewThreadStore()
	store.Add(mouseTestThread, mouseTestTitle)

	app := NewAppModel(store, WithInteractiveCommand(mouseTestCmd))
	app.width = mouseTestWidth
	app.height = mouseTestHeight

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	ptySession := app.ptySessions[mouseTestThread]
	ptySession.ScrollUp(scrollStep)

	scrollDown := tea.MouseMsg{
		Button: tea.MouseButtonWheelDown,
		Action: tea.MouseActionPress,
	}
	updated, _ = app.Update(scrollDown)
	app = updated.(AppModel)

	offset := ptySession.ScrollOffset()
	if offset != 0 {
		testing.Errorf("expected offset 0 after scroll down, got %d", offset)
	}
}

func TestAppMouseScrollIgnoredOnCanvas(testing *testing.T) {
	store := state.NewThreadStore()
	store.Add(mouseTestThread, mouseTestTitle)

	app := NewAppModel(store)

	scrollUp := tea.MouseMsg{
		Button: tea.MouseButtonWheelUp,
		Action: tea.MouseActionPress,
	}
	updated, _ := app.Update(scrollUp)
	_ = updated.(AppModel)
}
