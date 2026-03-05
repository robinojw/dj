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

	respCh := make(chan bool, 1)
	modal.SetResponseChannel(respCh)

	// Press 'y' to confirm
	modal, _ = modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	select {
	case confirmed := <-respCh:
		if !confirmed {
			t.Error("Expected confirmation")
		}
	default:
		t.Error("No response received")
	}

	if modal.Visible() {
		t.Error("Modal should be hidden after confirmation")
	}
}

func TestTurboModal_Cancel(t *testing.T) {
	modal := NewTurboModal(theme.DefaultTheme())
	modal.Show()

	respCh := make(chan bool, 1)
	modal.SetResponseChannel(respCh)

	// Press 'n' to cancel
	modal, _ = modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	select {
	case confirmed := <-respCh:
		if confirmed {
			t.Error("Expected cancellation")
		}
	default:
		t.Error("No response received")
	}

	if modal.Visible() {
		t.Error("Modal should be hidden after cancellation")
	}
}
