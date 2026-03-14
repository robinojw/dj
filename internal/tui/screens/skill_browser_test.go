package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/tui/theme"
)

func newSkillBrowserWithSkills(t *testing.T) SkillBrowserModel {
	t.Helper()
	m := NewSkillBrowserModel(theme.DefaultTheme())
	m.SetSkills([]SkillInfo{
		{Name: "commit", Description: "Create a git commit", Source: "builtin", Implicit: false},
		{Name: "review-pr", Description: "Review a pull request", Source: "builtin", Implicit: true},
		{Name: "deploy", Description: "Deploy to production", Source: "project", Implicit: false},
		{Name: "tdd", Description: "Test-driven development", Source: "user", Implicit: false},
	})
	return m
}

func TestSkillBrowser_InitialState(t *testing.T) {
	m := NewSkillBrowserModel(theme.DefaultTheme())

	if len(m.skills) != 0 {
		t.Errorf("expected no skills initially, got %d", len(m.skills))
	}
	if m.selected != 0 {
		t.Errorf("selected = %d, want 0", m.selected)
	}
}

func TestSkillBrowser_SetSkills(t *testing.T) {
	m := newSkillBrowserWithSkills(t)

	if len(m.skills) != 4 {
		t.Fatalf("expected 4 skills, got %d", len(m.skills))
	}
}

func TestSkillBrowser_NavigateDown(t *testing.T) {
	m := newSkillBrowserWithSkills(t)

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.selected != 1 {
		t.Errorf("selected = %d, want 1", m.selected)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.selected != 2 {
		t.Errorf("selected = %d, want 2", m.selected)
	}
}

func TestSkillBrowser_NavigateUp(t *testing.T) {
	m := newSkillBrowserWithSkills(t)
	m.selected = 3

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.selected != 2 {
		t.Errorf("selected = %d, want 2", m.selected)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.selected != 1 {
		t.Errorf("selected = %d, want 1", m.selected)
	}
}

func TestSkillBrowser_NavigationBounds(t *testing.T) {
	m := newSkillBrowserWithSkills(t)

	// Can't go above 0.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.selected != 0 {
		t.Errorf("selected = %d, want 0 (clamped at top)", m.selected)
	}

	// Can't go past last.
	m.selected = 3
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.selected != 3 {
		t.Errorf("selected = %d, want 3 (clamped at bottom)", m.selected)
	}
}

func TestSkillBrowser_NavigationEmptyList(t *testing.T) {
	m := NewSkillBrowserModel(theme.DefaultTheme())

	// Should not panic on empty list.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})

	if m.selected != 0 {
		t.Errorf("selected = %d, want 0", m.selected)
	}
}

func TestSkillBrowser_WindowResize(t *testing.T) {
	m := NewSkillBrowserModel(theme.DefaultTheme())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if m.width != 120 || m.height != 40 {
		t.Errorf("dimensions = %dx%d, want 120x40", m.width, m.height)
	}
}

func TestSkillBrowser_View_ZeroDimensions(t *testing.T) {
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
			m := NewSkillBrowserModel(theme.DefaultTheme())
			if sz.w > 0 || sz.h > 0 {
				m, _ = m.Update(tea.WindowSizeMsg{Width: sz.w, Height: sz.h})
			}
			_ = m.View()
		})

		t.Run(sz.name+"_with_skills", func(t *testing.T) {
			m := newSkillBrowserWithSkills(t)
			if sz.w > 0 || sz.h > 0 {
				m, _ = m.Update(tea.WindowSizeMsg{Width: sz.w, Height: sz.h})
			}
			_ = m.View()
		})
	}
}

func TestSkillBrowser_View_DoesNotPanic(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		m := NewSkillBrowserModel(theme.DefaultTheme())
		m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		v := m.View()
		if v == "" {
			t.Error("View() returned empty string")
		}
	})

	t.Run("with_skills", func(t *testing.T) {
		m := newSkillBrowserWithSkills(t)
		m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		v := m.View()
		if v == "" {
			t.Error("View() returned empty string")
		}
	})
}
