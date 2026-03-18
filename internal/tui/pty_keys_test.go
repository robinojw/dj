package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestKeyMsgToBytesRune(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	result := KeyMsgToBytes(msg)

	if len(result) != 1 || result[0] != 'a' {
		t.Errorf("expected 'a', got %v", result)
	}
}

func TestKeyMsgToBytesMultiByteRune(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'é'}}
	result := KeyMsgToBytes(msg)

	if string(result) != "é" {
		t.Errorf("expected 'é', got %q", string(result))
	}
}

func TestKeyMsgToBytesEnter(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result := KeyMsgToBytes(msg)

	if len(result) != 1 || result[0] != '\r' {
		t.Errorf("expected carriage return, got %v", result)
	}
}

func TestKeyMsgToBytesTab(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyTab}
	result := KeyMsgToBytes(msg)

	if len(result) != 1 || result[0] != '\t' {
		t.Errorf("expected tab, got %v", result)
	}
}

func TestKeyMsgToBytesBackspace(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	result := KeyMsgToBytes(msg)

	if len(result) != 1 || result[0] != 0x7f {
		t.Errorf("expected 0x7f, got %v", result)
	}
}

func TestKeyMsgToBytesCtrlC(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	result := KeyMsgToBytes(msg)

	if len(result) != 1 || result[0] != 0x03 {
		t.Errorf("expected 0x03, got %v", result)
	}
}

func TestKeyMsgToBytesArrowUp(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyUp}
	result := KeyMsgToBytes(msg)

	expected := []byte{0x1b, '[', 'A'}
	if len(result) != len(expected) {
		t.Fatalf("expected %d bytes, got %d", len(expected), len(result))
	}
	for i, b := range expected {
		if result[i] != b {
			t.Errorf("byte %d: expected 0x%02x, got 0x%02x", i, b, result[i])
		}
	}
}

func TestKeyMsgToBytesArrowDown(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyDown}
	result := KeyMsgToBytes(msg)

	expected := []byte{0x1b, '[', 'B'}
	if len(result) != len(expected) {
		t.Fatalf("expected %d bytes, got %d", len(expected), len(result))
	}
	for i, b := range expected {
		if result[i] != b {
			t.Errorf("byte %d: expected 0x%02x, got 0x%02x", i, b, result[i])
		}
	}
}

func TestKeyMsgToBytesAltModifier(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}, Alt: true}
	result := KeyMsgToBytes(msg)

	if len(result) != 2 || result[0] != 0x1b || result[1] != 'x' {
		t.Errorf("expected ESC + 'x', got %v", result)
	}
}

func TestKeyMsgToBytesAltArrow(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyUp, Alt: true}
	result := KeyMsgToBytes(msg)

	expected := []byte{0x1b, 0x1b, '[', 'A'}
	if len(result) != len(expected) {
		t.Fatalf("expected %d bytes, got %d", len(expected), len(result))
	}
	for i, b := range expected {
		if result[i] != b {
			t.Errorf("byte %d: expected 0x%02x, got 0x%02x", i, b, result[i])
		}
	}
}

func TestKeyMsgToBytesFunctionKeys(t *testing.T) {
	tests := []struct {
		keyType  tea.KeyType
		expected []byte
	}{
		{tea.KeyF1, []byte{0x1b, 'O', 'P'}},
		{tea.KeyF5, []byte{0x1b, '[', '1', '5', '~'}},
		{tea.KeyF12, []byte{0x1b, '[', '2', '4', '~'}},
	}

	for _, test := range tests {
		msg := tea.KeyMsg{Type: test.keyType}
		result := KeyMsgToBytes(msg)

		if len(result) != len(test.expected) {
			t.Errorf("key %v: expected %d bytes, got %d", test.keyType, len(test.expected), len(result))
			continue
		}
		for i, b := range test.expected {
			if result[i] != b {
				t.Errorf("key %v byte %d: expected 0x%02x, got 0x%02x", test.keyType, i, b, result[i])
			}
		}
	}
}

func TestKeyMsgToBytesPageUpDown(t *testing.T) {
	pgUp := tea.KeyMsg{Type: tea.KeyPgUp}
	result := KeyMsgToBytes(pgUp)
	expected := []byte{0x1b, '[', '5', '~'}

	if len(result) != len(expected) {
		t.Fatalf("PgUp: expected %d bytes, got %d", len(expected), len(result))
	}
	for i, b := range expected {
		if result[i] != b {
			t.Errorf("PgUp byte %d: expected 0x%02x, got 0x%02x", i, b, result[i])
		}
	}
}

func TestKeyMsgToBytesSpace(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeySpace}
	result := KeyMsgToBytes(msg)

	if len(result) != 1 || result[0] != ' ' {
		t.Errorf("expected space byte, got %v", result)
	}
}

func TestKeyMsgToBytesUnknown(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyType(9999)}
	result := KeyMsgToBytes(msg)

	if result != nil {
		t.Errorf("expected nil for unknown key, got %v", result)
	}
}
