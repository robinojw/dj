package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/state"
)

const testInputBarPrompt = "Task: "

func TestHandleInputBarKeyTyping(testing *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)
	app.inputBarVisible = true
	app.inputBar = NewInputBarModel(testInputBarPrompt)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'H'}}
	updated, _ := app.Update(msg)
	resultApp := updated.(AppModel)

	value := resultApp.inputBar.Value()
	if value != "H" {
		testing.Errorf("expected 'H', got %q", value)
	}
}

func TestHandleInputBarKeyEscDismisses(testing *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)
	app.inputBarVisible = true
	app.inputBar = NewInputBarModel(testInputBarPrompt)

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ := app.Update(msg)
	resultApp := updated.(AppModel)

	if resultApp.inputBarVisible {
		testing.Error("expected input bar dismissed on Esc")
	}
}

func TestHandleInputBarKeyBackspace(testing *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)
	app.inputBarVisible = true
	app.inputBar = NewInputBarModel(testInputBarPrompt)
	app.inputBar.InsertRune('A')
	app.inputBar.InsertRune('B')

	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	updated, _ := app.Update(msg)
	resultApp := updated.(AppModel)

	value := resultApp.inputBar.Value()
	if value != "A" {
		testing.Errorf("expected 'A', got %q", value)
	}
}

func TestHandleInputBarKeyEnterEmptyDismisses(testing *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)
	app.inputBarVisible = true
	app.inputBar = NewInputBarModel(testInputBarPrompt)

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(msg)
	resultApp := updated.(AppModel)

	if resultApp.inputBarVisible {
		testing.Error("expected input bar dismissed on empty Enter")
	}
}
