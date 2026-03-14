package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/tui/theme"
)

// SkillInfo holds display info for a skill.
type SkillInfo struct {
	Name        string
	Description string
	Source      string // "builtin", "project", "user"
	Implicit    bool
}

// SkillBrowserModel is the skills library screen.
type SkillBrowserModel struct {
	skills   []SkillInfo
	selected int
	width    int
	height   int
	theme    *theme.Theme
}

func NewSkillBrowserModel(t *theme.Theme) SkillBrowserModel {
	return SkillBrowserModel{theme: t}
}

func (m *SkillBrowserModel) SetSkills(skills []SkillInfo) {
	m.skills = skills
}

func (m SkillBrowserModel) Init() tea.Cmd { return nil }

func (m SkillBrowserModel) Update(msg tea.Msg) (SkillBrowserModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.skills)-1 {
				m.selected++
			}
		}
	}

	return m, nil
}

func (m SkillBrowserModel) View() string {
	w := max(min(m.width-4, 60), 0)
	border := strings.Repeat("═", w)

	title := m.theme.AccentStyle().Render("  Skills Library                           Ctrl+K  ")

	var rows []string
	for i, s := range m.skills {
		implicit := " "
		if s.Implicit {
			implicit = "⚡"
		}
		src := m.theme.MutedStyle().Render(fmt.Sprintf("[%s]", s.Source))

		line := fmt.Sprintf("  %s $%-18s %s  %s", implicit, s.Name, src, s.Description)
		if i == m.selected {
			line = m.theme.SelectedStyle().Render(line)
		}
		rows = append(rows, line)
	}

	if len(rows) == 0 {
		rows = append(rows, m.theme.MutedStyle().Render("  No skills found"))
	}

	body := strings.Join(rows, "\n")
	footer := m.theme.MutedStyle().Render("  [Enter] view details  [$name] invoke in chat  [Esc] back")

	return lipgloss.JoinVertical(lipgloss.Left,
		"╔"+border+"╗",
		title,
		"╠"+border+"╣",
		body,
		"╠"+border+"╣",
		footer,
		"╚"+border+"╝",
	)
}
