package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/tui/theme"
)

type InputMode int

const (
	ModeNormal InputMode = iota
	ModeSkillComplete
	ModeMentionComplete // @ trigger
)

// SkillSuggestion is a single autocomplete candidate.
type SkillSuggestion struct {
	Name        string
	Description string
}

// ChatInput is a text input with $skill autocomplete support.
type ChatInput struct {
	textInput   textinput.Model
	mode        InputMode
	suggestions []SkillSuggestion
	selected    int
	allSkills   []SkillSuggestion
	width       int
	theme       *theme.Theme
}

func NewChatInput(t *theme.Theme) ChatInput {
	ti := textinput.New()
	ti.Placeholder = "Send a message... ($ for skills)"
	ti.Focus()
	ti.CharLimit = 4096

	return ChatInput{
		textInput: ti,
		theme:     t,
	}
}

func (c *ChatInput) SetWidth(w int) {
	c.width = w
	c.textInput.Width = w - 4
}

func (c *ChatInput) SetSkills(skills []SkillSuggestion) {
	c.allSkills = skills
}

func (c *ChatInput) Value() string {
	return c.textInput.Value()
}

func (c *ChatInput) Reset() {
	c.textInput.Reset()
	c.mode = ModeNormal
	c.suggestions = nil
	c.selected = 0
}

func (c *ChatInput) Focus() tea.Cmd {
	return c.textInput.Focus()
}

func (c ChatInput) Update(msg tea.Msg) (ChatInput, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case c.mode == ModeSkillComplete:
			switch msg.String() {
			case "tab", "down":
				c.selected = (c.selected + 1) % max(len(c.suggestions), 1)
				return c, nil
			case "shift+tab", "up":
				c.selected--
				if c.selected < 0 {
					c.selected = max(len(c.suggestions)-1, 0)
				}
				return c, nil
			case "enter":
				if len(c.suggestions) > 0 {
					c.insertSkill(c.suggestions[c.selected].Name)
				}
				c.mode = ModeNormal
				c.suggestions = nil
				return c, nil
			case "esc":
				c.mode = ModeNormal
				c.suggestions = nil
				return c, nil
			}
		}
	}

	var cmd tea.Cmd
	c.textInput, cmd = c.textInput.Update(msg)

	// Check for $ trigger
	val := c.textInput.Value()
	if idx := strings.LastIndex(val, "$"); idx >= 0 {
		query := val[idx+1:]
		c.mode = ModeSkillComplete
		c.suggestions = c.filterSkills(query)
		c.selected = 0
	} else if c.mode == ModeSkillComplete {
		c.mode = ModeNormal
		c.suggestions = nil
	}

	// Check for @ trigger (mention mode)
	if strings.HasSuffix(val, "@") && c.mode == ModeNormal {
		c.mode = ModeMentionComplete
	} else if c.mode == ModeMentionComplete && !strings.Contains(val, "@") {
		c.mode = ModeNormal
	}

	return c, cmd
}

func (c ChatInput) View() string {
	input := c.textInput.View()

	if c.mode == ModeSkillComplete && len(c.suggestions) > 0 {
		popup := c.renderSuggestions()
		return lipgloss.JoinVertical(lipgloss.Left, popup, input)
	}

	return input
}

func (c *ChatInput) filterSkills(query string) []SkillSuggestion {
	query = strings.ToLower(query)
	var matches []SkillSuggestion
	for _, s := range c.allSkills {
		if query == "" || strings.Contains(strings.ToLower(s.Name), query) {
			matches = append(matches, s)
		}
		if len(matches) >= 5 {
			break
		}
	}
	return matches
}

func (c *ChatInput) insertSkill(name string) {
	val := c.textInput.Value()
	if idx := strings.LastIndex(val, "$"); idx >= 0 {
		c.textInput.SetValue(val[:idx] + "$" + name + " ")
		c.textInput.CursorEnd()
	}
}

func (c ChatInput) renderSuggestions() string {
	var rows []string
	for i, s := range c.suggestions {
		line := "  " + s.Name + "  " + c.theme.MutedStyle().Render(s.Description)
		if i == c.selected {
			line = c.theme.SelectedStyle().Render("> " + s.Name + "  " + s.Description)
		}
		rows = append(rows, line)
	}
	content := strings.Join(rows, "\n")
	return c.theme.PanelStyle().Render(content)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
