package screens

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/tui/theme"
)

// EnhancedPromptMsg is sent when the user accepts the enhanced prompt.
type EnhancedPromptMsg struct {
	Text string
}

// EnhanceModel shows a before/after prompt enhancement modal.
type EnhanceModel struct {
	original string
	enhanced string
	loading  bool
	ready    bool
	width    int
	height   int
	theme    *theme.Theme
}

func NewEnhanceModel(t *theme.Theme) EnhanceModel {
	return EnhanceModel{theme: t}
}

func (m EnhanceModel) Init() tea.Cmd { return nil }

func (m EnhanceModel) Update(msg tea.Msg) (EnhanceModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.ready {
				return m, func() tea.Msg {
					return EnhancedPromptMsg{Text: m.enhanced}
				}
			}
		}

	case enhanceResultMsg:
		m.enhanced = msg.text
		m.loading = false
		m.ready = true
	}

	return m, nil
}

type enhanceResultMsg struct {
	text string
}

func (m EnhanceModel) View() string {
	title := m.theme.AccentStyle().Render("  Enhance Prompt                           Ctrl+E  ")
	border := strings.Repeat("═", min(m.width-4, 56))

	var body string
	if m.loading {
		body = m.theme.MutedStyle().Render("  Enhancing prompt...")
	} else if m.ready {
		beforeLabel := m.theme.MutedStyle().Render("  BEFORE")
		beforeContent := "  " + m.original
		afterLabel := m.theme.SuccessStyle().Render("  AFTER")
		afterContent := "  " + m.enhanced

		body = strings.Join([]string{
			beforeLabel,
			"  " + strings.Repeat("─", min(m.width-8, 48)),
			beforeContent,
			"",
			afterLabel,
			"  " + strings.Repeat("─", min(m.width-8, 48)),
			afterContent,
		}, "\n")
	} else {
		body = m.theme.MutedStyle().Render("  Type a prompt in chat, then press Ctrl+E to enhance it.")
	}

	footer := m.theme.MutedStyle().Render("  [Enter] use enhanced  [Esc] keep original")

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

// SetOriginal sets the prompt to enhance and starts the enhancement.
func (m *EnhanceModel) SetOriginal(text string) tea.Cmd {
	m.original = text
	m.loading = true
	m.ready = false
	m.enhanced = ""
	return nil // The actual API call will be wired by the orchestrator
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
