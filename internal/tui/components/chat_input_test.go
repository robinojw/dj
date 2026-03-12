package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/tui/theme"
)

func newInputWithSkills(t *testing.T) ChatInput {
	t.Helper()
	ci := NewChatInput(theme.DefaultTheme())
	ci.SetSkills([]SkillSuggestion{
		{Name: "commit", Description: "Create a git commit"},
		{Name: "review-pr", Description: "Review a pull request"},
		{Name: "compile", Description: "Compile the project"},
		{Name: "deploy", Description: "Deploy to production"},
	})
	return ci
}

func TestChatInput_InitialState(t *testing.T) {
	ci := NewChatInput(theme.DefaultTheme())

	if ci.Value() != "" {
		t.Errorf("value = %q, want empty", ci.Value())
	}
	if ci.mode != ModeNormal {
		t.Errorf("mode = %d, want ModeNormal", ci.mode)
	}
}

func TestChatInput_Reset(t *testing.T) {
	ci := newInputWithSkills(t)

	// Type something by sending key events.
	ci, _ = ci.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	ci, _ = ci.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	if ci.Value() == "" {
		t.Fatal("expected non-empty value after typing")
	}

	ci.Reset()

	if ci.Value() != "" {
		t.Errorf("value after reset = %q, want empty", ci.Value())
	}
	if ci.mode != ModeNormal {
		t.Errorf("mode after reset = %d, want ModeNormal", ci.mode)
	}
}

func TestChatInput_SkillAutocomplete_DollarTrigger(t *testing.T) {
	ci := newInputWithSkills(t)

	// Type '$' to trigger skill autocomplete.
	ci, _ = ci.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'$'}})

	if ci.mode != ModeSkillComplete {
		t.Errorf("mode = %d, want ModeSkillComplete", ci.mode)
	}
	// With empty query after $, all skills should match (up to 5).
	if len(ci.suggestions) != 4 {
		t.Errorf("suggestions = %d, want 4", len(ci.suggestions))
	}
}

func TestChatInput_SkillAutocomplete_Filtering(t *testing.T) {
	ci := newInputWithSkills(t)

	// Type "$com" — should filter to "commit" and "compile" (both contain "com").
	for _, r := range "$com" {
		ci, _ = ci.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	if ci.mode != ModeSkillComplete {
		t.Fatalf("mode = %d, want ModeSkillComplete", ci.mode)
	}
	if len(ci.suggestions) != 2 {
		t.Errorf("filtered suggestions = %d, want 2 (commit, compile)", len(ci.suggestions))
	}
}

func TestChatInput_SkillAutocomplete_TabCycles(t *testing.T) {
	ci := newInputWithSkills(t)
	ci, _ = ci.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'$'}})

	if ci.selected != 0 {
		t.Fatalf("initial selected = %d, want 0", ci.selected)
	}

	// Tab cycles forward.
	ci, _ = ci.Update(tea.KeyMsg{Type: tea.KeyTab})
	if ci.selected != 1 {
		t.Errorf("after tab: selected = %d, want 1", ci.selected)
	}

	// Down also cycles forward.
	ci, _ = ci.Update(tea.KeyMsg{Type: tea.KeyDown})
	if ci.selected != 2 {
		t.Errorf("after down: selected = %d, want 2", ci.selected)
	}
}

func TestChatInput_SkillAutocomplete_ShiftTabCyclesBackward(t *testing.T) {
	ci := newInputWithSkills(t)
	ci, _ = ci.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'$'}})

	// Shift+Tab wraps backward from 0.
	ci, _ = ci.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if ci.selected != len(ci.suggestions)-1 {
		t.Errorf("selected = %d, want %d (wrapped to last)", ci.selected, len(ci.suggestions)-1)
	}
}

func TestChatInput_SkillAutocomplete_EscExits(t *testing.T) {
	ci := newInputWithSkills(t)
	ci, _ = ci.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'$'}})

	if ci.mode != ModeSkillComplete {
		t.Fatal("should be in SkillComplete mode")
	}

	ci, _ = ci.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if ci.mode != ModeNormal {
		t.Errorf("mode = %d, want ModeNormal after Esc", ci.mode)
	}
	if ci.suggestions != nil {
		t.Error("suggestions should be nil after Esc")
	}
}

func TestChatInput_SkillAutocomplete_EnterInserts(t *testing.T) {
	ci := newInputWithSkills(t)
	ci, _ = ci.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'$'}})

	// Select first suggestion and press Enter.
	ci, _ = ci.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if ci.mode != ModeNormal {
		t.Errorf("mode = %d, want ModeNormal after Enter", ci.mode)
	}

	val := ci.Value()
	// Should have inserted "$<skillname> ".
	if val == "" || val == "$" {
		t.Errorf("value after insert = %q, expected skill name to be inserted", val)
	}
}

func TestChatInput_MentionTrigger(t *testing.T) {
	ci := NewChatInput(theme.DefaultTheme())

	// Type '@' to trigger mention mode.
	ci, _ = ci.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'@'}})

	if ci.mode != ModeMentionComplete {
		t.Errorf("mode = %d, want ModeMentionComplete", ci.mode)
	}
}

func TestChatInput_SetWidth(t *testing.T) {
	ci := NewChatInput(theme.DefaultTheme())
	ci.SetWidth(100)

	if ci.width != 100 {
		t.Errorf("width = %d, want 100", ci.width)
	}
}

func TestChatInput_View_DoesNotPanic(t *testing.T) {
	ci := NewChatInput(theme.DefaultTheme())

	v := ci.View()
	if v == "" {
		t.Error("View() returned empty string")
	}
}

func TestChatInput_View_WithSuggestions(t *testing.T) {
	ci := newInputWithSkills(t)
	ci, _ = ci.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'$'}})

	v := ci.View()
	if v == "" {
		t.Error("View() returned empty string with suggestions")
	}
}

func TestChatInput_FilterSkills(t *testing.T) {
	ci := newInputWithSkills(t)

	tests := []struct {
		query string
		want  int
	}{
		{"", 4},          // all skills
		{"com", 2},       // commit, compile
		{"deploy", 1},    // deploy
		{"nonexistent", 0},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got := ci.filterSkills(tt.query)
			if len(got) != tt.want {
				t.Errorf("filterSkills(%q) returned %d results, want %d", tt.query, len(got), tt.want)
			}
		})
	}
}

func TestChatInput_FilterSkills_MaxFive(t *testing.T) {
	ci := NewChatInput(theme.DefaultTheme())
	var skills []SkillSuggestion
	for i := 0; i < 10; i++ {
		skills = append(skills, SkillSuggestion{Name: "skill", Description: "test"})
	}
	ci.SetSkills(skills)

	got := ci.filterSkills("")
	if len(got) != 5 {
		t.Errorf("filterSkills returned %d results, want 5 (max)", len(got))
	}
}
