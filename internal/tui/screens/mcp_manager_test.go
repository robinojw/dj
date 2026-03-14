package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/tui/theme"
)

func newMCPWithServers(t *testing.T) MCPManagerModel {
	t.Helper()
	m := NewMCPManagerModel(theme.DefaultTheme())
	m.SetServers([]MCPServerInfo{
		{Name: "filesystem", Type: "stdio", ToolCount: 5, Active: true},
		{Name: "github", Type: "http", ToolCount: 12, Active: false},
		{Name: "slack", Type: "sse", ToolCount: 3, Active: true},
	})
	return m
}

func TestMCPManager_InitialState(t *testing.T) {
	m := NewMCPManagerModel(theme.DefaultTheme())

	if len(m.servers) != 0 {
		t.Errorf("expected no servers initially, got %d", len(m.servers))
	}
	if m.selected != 0 {
		t.Errorf("selected = %d, want 0", m.selected)
	}
}

func TestMCPManager_SetServers(t *testing.T) {
	m := newMCPWithServers(t)

	if len(m.servers) != 3 {
		t.Fatalf("expected 3 servers, got %d", len(m.servers))
	}
	if m.servers[0].Name != "filesystem" {
		t.Errorf("first server = %q, want %q", m.servers[0].Name, "filesystem")
	}
}

func TestMCPManager_NavigateDown(t *testing.T) {
	m := newMCPWithServers(t)

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.selected != 1 {
		t.Errorf("selected = %d, want 1", m.selected)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.selected != 2 {
		t.Errorf("selected = %d, want 2", m.selected)
	}
}

func TestMCPManager_NavigateUp(t *testing.T) {
	m := newMCPWithServers(t)
	m.selected = 2

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.selected != 1 {
		t.Errorf("selected = %d, want 1", m.selected)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.selected != 0 {
		t.Errorf("selected = %d, want 0", m.selected)
	}
}

func TestMCPManager_NavigationBounds(t *testing.T) {
	m := newMCPWithServers(t)

	// Can't go above 0.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.selected != 0 {
		t.Errorf("selected = %d, want 0 (clamped at top)", m.selected)
	}

	// Can't go past last.
	m.selected = 2
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.selected != 2 {
		t.Errorf("selected = %d, want 2 (clamped at bottom)", m.selected)
	}
}

func TestMCPManager_NavigationEmptyList(t *testing.T) {
	m := NewMCPManagerModel(theme.DefaultTheme())

	// Should not panic on empty list.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})

	if m.selected != 0 {
		t.Errorf("selected = %d, want 0", m.selected)
	}
}

func TestMCPManager_ToggleActive(t *testing.T) {
	m := newMCPWithServers(t)

	// "github" is at index 1, initially inactive.
	m.selected = 1
	if m.servers[1].Active {
		t.Fatal("github should start inactive")
	}

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !m.servers[1].Active {
		t.Error("github should be active after toggle")
	}

	// The command should emit an MCPToggleMsg.
	if cmd == nil {
		t.Fatal("expected a command from toggle")
	}
	msg := cmd()
	toggle, ok := msg.(MCPToggleMsg)
	if !ok {
		t.Fatalf("expected MCPToggleMsg, got %T", msg)
	}
	if toggle.Name != "github" || !toggle.Active {
		t.Errorf("toggle = %+v, want Name=github Active=true", toggle)
	}

	// Toggle again to deactivate.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.servers[1].Active {
		t.Error("github should be inactive after second toggle")
	}
}

func TestMCPManager_ToggleEmptyList(t *testing.T) {
	m := NewMCPManagerModel(theme.DefaultTheme())

	// Enter on empty list should not panic and produce no command.
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("toggle on empty list should not produce a command")
	}
}

func TestMCPManager_WindowResize(t *testing.T) {
	m := NewMCPManagerModel(theme.DefaultTheme())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if m.width != 120 || m.height != 40 {
		t.Errorf("dimensions = %dx%d, want 120x40", m.width, m.height)
	}
}

func TestMCPManager_View_ZeroDimensions(t *testing.T) {
	sizes := []struct {
		name string
		w, h int
	}{
		{"no_resize", 0, 0},
		{"1x1", 1, 1},
		{"2x2", 2, 2},
		{"narrow", 3, 100},
		{"short", 100, 1},
	}
	for _, sz := range sizes {
		t.Run(sz.name, func(t *testing.T) {
			m := NewMCPManagerModel(theme.DefaultTheme())
			if sz.w > 0 || sz.h > 0 {
				m, _ = m.Update(tea.WindowSizeMsg{Width: sz.w, Height: sz.h})
			}
			_ = m.View()
		})

		t.Run(sz.name+"_with_servers", func(t *testing.T) {
			m := newMCPWithServers(t)
			if sz.w > 0 || sz.h > 0 {
				m, _ = m.Update(tea.WindowSizeMsg{Width: sz.w, Height: sz.h})
			}
			_ = m.View()
		})
	}
}

func TestMCPManager_View_DoesNotPanic(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		m := NewMCPManagerModel(theme.DefaultTheme())
		m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		v := m.View()
		if v == "" {
			t.Error("View() returned empty string")
		}
	})

	t.Run("with_servers", func(t *testing.T) {
		m := newMCPWithServers(t)
		m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		v := m.View()
		if v == "" {
			t.Error("View() returned empty string")
		}
	})
}
