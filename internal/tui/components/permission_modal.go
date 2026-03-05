package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/modes"
	"github.com/robinojw/dj/internal/tui/theme"
)

type PermissionModal struct {
	request *modes.PermissionRequest
	scope   modes.RememberScope
	visible bool
	theme   *theme.Theme
}

func NewPermissionModal(t *theme.Theme) PermissionModal {
	return PermissionModal{
		scope: modes.RememberOnce,
		theme: t,
	}
}

func (m PermissionModal) Visible() bool {
	return m.visible
}

func (m PermissionModal) Update(msg tea.Msg) (PermissionModal, tea.Cmd) {
	switch msg := msg.(type) {
	case modes.PermissionRequest:
		m.request = &msg
		m.visible = true
		m.scope = modes.RememberOnce
		return m, nil

	case tea.KeyMsg:
		if !m.visible || m.request == nil {
			return m, nil
		}

		switch msg.String() {
		case "y":
			// Approve
			m.request.RespCh <- modes.PermissionResp{
				Allowed:     true,
				RememberFor: m.scope,
			}
			m.visible = false
			m.request = nil
			return m, nil

		case "n", "esc":
			// Deny
			m.request.RespCh <- modes.PermissionResp{
				Allowed: false,
			}
			m.visible = false
			m.request = nil
			return m, nil

		case "tab":
			// Cycle scope: Once → Session → Always → Once
			m.scope = (m.scope + 1) % 3
			return m, nil
		}
	}

	return m, nil
}

func (m PermissionModal) View() string {
	if !m.visible || m.request == nil {
		return ""
	}

	// Tool name and args
	toolLine := fmt.Sprintf("🔧 %s", m.request.Tool)
	argsLines := formatArgs(m.request.Args)

	// Scope indicators
	scopeOptions := []string{
		renderScope(modes.RememberOnce, m.scope == modes.RememberOnce),
		renderScope(modes.RememberSession, m.scope == modes.RememberSession),
		renderScope(modes.RememberAlways, m.scope == modes.RememberAlways),
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		"",
		fmt.Sprintf("Worker %s wants to run:", m.request.WorkerID),
		"",
		toolLine,
		strings.Repeat("─", 60),
		argsLines,
		"",
		strings.Repeat("═", 60),
		"Remember this decision?",
		strings.Join(scopeOptions, "   "),
		strings.Repeat("═", 60),
		"[y] Allow    [n] Deny    [Tab] cycle scope    [Esc] Deny",
	)

	return m.theme.PanelStyle().
		Width(64).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("214")).
		Render(content)
}

func renderScope(scope modes.RememberScope, selected bool) string {
	label := scope.String()
	if selected {
		return fmt.Sprintf("● %s", label)
	}
	return fmt.Sprintf("○ %s", label)
}

func formatArgs(args map[string]any) string {
	if len(args) == 0 {
		return "(no arguments)"
	}

	var lines []string
	for k, v := range args {
		lines = append(lines, fmt.Sprintf("%s: %v", k, v))
	}
	return strings.Join(lines, "\n")
}
