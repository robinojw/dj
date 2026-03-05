package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/tui/theme"
)

func TestTurboModal_ShowHide(t *testing.T) {
	modal := NewTurboModal(theme.DefaultTheme())

	if modal.Visible() {
		t.Error("Modal should be hidden initially")
	}

	modal.Show()

	if !modal.Visible() {
		t.Error("Modal should be visible after Show()")
	}
}

func TestTurboModal_Confirm(t *testing.T) {
	modal := NewTurboModal(theme.DefaultTheme())
	modal.Show()

	// Press 'y' to confirm
	modal, _ = modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	if modal.Visible() {
		t.Error("Modal should be hidden after confirmation")
	}
	if !modal.Confirmed() {
		t.Error("Expected confirmation")
	}
}

func TestTurboModal_Cancel(t *testing.T) {
	modal := NewTurboModal(theme.DefaultTheme())
	modal.Show()

	// Press 'n' to cancel
	modal, _ = modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	if modal.Visible() {
		t.Error("Modal should be hidden after cancellation")
	}
	if modal.Confirmed() {
		t.Error("Expected cancellation")
	}
}
