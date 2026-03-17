package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPrefixKeyCapture(t *testing.T) {
	prefix := NewPrefixHandler()

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	result := prefix.HandleKey(ctrlB)

	if result != PrefixWaiting {
		t.Errorf("expected waiting, got %d", result)
	}
	if !prefix.Active() {
		t.Error("expected prefix to be active")
	}
}

func TestPrefixKeyFollowUp(t *testing.T) {
	prefix := NewPrefixHandler()

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	prefix.HandleKey(ctrlB)

	mKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	result := prefix.HandleKey(mKey)

	if result != PrefixComplete {
		t.Errorf("expected complete, got %d", result)
	}
	if prefix.Action() != 'm' {
		t.Errorf("expected 'm', got %c", prefix.Action())
	}
	if prefix.Active() {
		t.Error("expected prefix to be inactive after completion")
	}
}

func TestPrefixKeyTimeout(t *testing.T) {
	prefix := NewPrefixHandler()

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	prefix.HandleKey(ctrlB)

	escKey := tea.KeyMsg{Type: tea.KeyEsc}
	result := prefix.HandleKey(escKey)

	if result != PrefixCancelled {
		t.Errorf("expected cancelled, got %d", result)
	}
	if prefix.Active() {
		t.Error("expected prefix to be inactive after cancel")
	}
}

func TestPrefixKeyInactivePassthrough(t *testing.T) {
	prefix := NewPrefixHandler()

	normalKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	result := prefix.HandleKey(normalKey)

	if result != PrefixNone {
		t.Errorf("expected none, got %d", result)
	}
}
