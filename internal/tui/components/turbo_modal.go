package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/tui/theme"
)

type TurboModal struct {
	visible bool
	respCh  chan bool
	theme   *theme.Theme
}

func NewTurboModal(t *theme.Theme) TurboModal {
	return TurboModal{
		theme: t,
	}
}

func (m *TurboModal) Show() {
	m.visible = true
}

func (m TurboModal) Visible() bool {
	return m.visible
}

func (m *TurboModal) SetResponseChannel(ch chan bool) {
	m.respCh = ch
}

func (m TurboModal) Update(msg tea.Msg) (TurboModal, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y":
			if m.respCh != nil {
				m.respCh <- true
			}
			m.visible = false
			return m, nil

		case "n", "esc":
			if m.respCh != nil {
				m.respCh <- false
			}
			m.visible = false
			return m, nil
		}
	}

	return m, nil
}

func (m TurboModal) View() string {
	if !m.visible {
		return ""
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		"",
		"TURBO bypasses ALL permission prompts.",
		"",
		"The agent can:",
		"  • Write/delete any files",
		"  • Execute any shell commands",
		"  • Make network requests",
		"",
		"Only use in isolated/safe environments.",
		"",
		strings.Repeat("═", 60),
		"[y] Activate Turbo    [n] Cancel",
	)

	return m.theme.PanelStyle().
		Width(64).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")). // red
		Render("⚡ TURBO MODE WARNING\n\n" + content)
}
