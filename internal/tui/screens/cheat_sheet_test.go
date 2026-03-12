package screens

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/tui/theme"
)

func TestCheatSheet_InitialState(t *testing.T) {
	m := NewCheatSheetModel(theme.DefaultTheme())

	if m.width != 0 || m.height != 0 {
		t.Errorf("dimensions = %dx%d, want 0x0", m.width, m.height)
	}
}

func TestCheatSheet_WindowResize(t *testing.T) {
	m := NewCheatSheetModel(theme.DefaultTheme())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if m.width != 120 || m.height != 40 {
		t.Errorf("dimensions = %dx%d, want 120x40", m.width, m.height)
	}
}

func TestCheatSheet_View_DoesNotPanic(t *testing.T) {
	m := NewCheatSheetModel(theme.DefaultTheme())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	v := m.View()
	if v == "" {
		t.Error("View() returned empty string")
	}
}

func TestCheatSheet_View_ContainsShortcuts(t *testing.T) {
	m := NewCheatSheetModel(theme.DefaultTheme())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	v := m.View()

	expected := []string{
		"Ctrl+E", "Ctrl+T", "Ctrl+K", "Ctrl+M", "Ctrl+H",
		"Ctrl+Z", "Ctrl+D", "Ctrl+Q", "Tab", "Esc", "Enter",
	}
	for _, key := range expected {
		if !strings.Contains(v, key) {
			t.Errorf("View() missing shortcut %q", key)
		}
	}
}

func TestCheatSheet_View_ContainsModes(t *testing.T) {
	m := NewCheatSheetModel(theme.DefaultTheme())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	v := m.View()

	modes := []string{"Confirm", "Plan", "Turbo"}
	for _, mode := range modes {
		if !strings.Contains(v, mode) {
			t.Errorf("View() missing mode %q", mode)
		}
	}
}

func TestCheatSheet_View_ContainsFooter(t *testing.T) {
	m := NewCheatSheetModel(theme.DefaultTheme())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	v := m.View()

	if !strings.Contains(v, "dismiss") {
		t.Error("View() missing dismiss footer")
	}
}

func TestCheatSheet_Init_ReturnsNil(t *testing.T) {
	m := NewCheatSheetModel(theme.DefaultTheme())
	if cmd := m.Init(); cmd != nil {
		t.Error("Init() should return nil")
	}
}
