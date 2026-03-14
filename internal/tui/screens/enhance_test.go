package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/tui/theme"
)

func TestEnhanceModel_InitialState(t *testing.T) {
	m := NewEnhanceModel(theme.DefaultTheme())

	if m.loading {
		t.Error("should not be loading initially")
	}
	if m.ready {
		t.Error("should not be ready initially")
	}
}

func TestEnhanceModel_WindowResize(t *testing.T) {
	m := NewEnhanceModel(theme.DefaultTheme())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func TestEnhanceModel_SetOriginal(t *testing.T) {
	m := NewEnhanceModel(theme.DefaultTheme())
	m.SetOriginal("fix the bug")

	if m.original != "fix the bug" {
		t.Errorf("original = %q, want %q", m.original, "fix the bug")
	}
	if !m.loading {
		t.Error("should be loading after SetOriginal")
	}
	if m.ready {
		t.Error("should not be ready after SetOriginal")
	}
}

func TestEnhanceModel_ReceiveResult(t *testing.T) {
	m := NewEnhanceModel(theme.DefaultTheme())
	m.SetOriginal("fix the bug")

	m, _ = m.Update(enhanceResultMsg{text: "Fix the authentication bug in login handler"})

	if m.loading {
		t.Error("should not be loading after result")
	}
	if !m.ready {
		t.Error("should be ready after result")
	}
	if m.enhanced != "Fix the authentication bug in login handler" {
		t.Errorf("enhanced = %q, want enhanced text", m.enhanced)
	}
}

func TestEnhanceModel_EnterWhenNotReady(t *testing.T) {
	m := NewEnhanceModel(theme.DefaultTheme())

	// Enter before result arrives should produce no command.
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("Enter when not ready should not produce a command")
	}
}

func TestEnhanceModel_EnterWhenReady(t *testing.T) {
	m := NewEnhanceModel(theme.DefaultTheme())
	m.SetOriginal("fix bug")
	m, _ = m.Update(enhanceResultMsg{text: "Fix the authentication bug"})

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter when ready should produce a command")
	}

	// Execute the command and check the message type.
	msg := cmd()
	enhanced, ok := msg.(EnhancedPromptMsg)
	if !ok {
		t.Fatalf("expected EnhancedPromptMsg, got %T", msg)
	}
	if enhanced.Text != "Fix the authentication bug" {
		t.Errorf("enhanced text = %q, want %q", enhanced.Text, "Fix the authentication bug")
	}
}

func TestEnhanceModel_View_ZeroDimensions(t *testing.T) {
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
			m := NewEnhanceModel(theme.DefaultTheme())
			if sz.w > 0 || sz.h > 0 {
				m, _ = m.Update(tea.WindowSizeMsg{Width: sz.w, Height: sz.h})
			}
			// Must not panic in any state.
			_ = m.View()

			// Also test with content loaded.
			m.SetOriginal("test")
			_ = m.View()

			m, _ = m.Update(enhanceResultMsg{text: "enhanced"})
			_ = m.View()
		})
	}
}

func TestEnhanceModel_View_States(t *testing.T) {
	th := theme.DefaultTheme()

	t.Run("idle", func(t *testing.T) {
		m := NewEnhanceModel(th)
		m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		v := m.View()
		if v == "" {
			t.Error("View() returned empty string in idle state")
		}
	})

	t.Run("loading", func(t *testing.T) {
		m := NewEnhanceModel(th)
		m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		m.SetOriginal("test prompt")
		v := m.View()
		if v == "" {
			t.Error("View() returned empty string in loading state")
		}
	})

	t.Run("ready", func(t *testing.T) {
		m := NewEnhanceModel(th)
		m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		m.SetOriginal("test prompt")
		m, _ = m.Update(enhanceResultMsg{text: "enhanced prompt"})
		v := m.View()
		if v == "" {
			t.Error("View() returned empty string in ready state")
		}
	})
}
