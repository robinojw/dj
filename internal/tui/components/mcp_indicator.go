package components

import (
	"strings"

	"github.com/robinojw/dj/internal/tui/theme"
)

// MCPIndicator renders active MCP server badges.
type MCPIndicator struct {
	Servers []MCPServerStatus
	Theme   *theme.Theme
}

type MCPServerStatus struct {
	Name   string
	Type   string // "stdio", "http", "sse"
	Active bool
	Tools  int
}

func NewMCPIndicator(t *theme.Theme) MCPIndicator {
	return MCPIndicator{Theme: t}
}

func (m MCPIndicator) View() string {
	if len(m.Servers) == 0 {
		return m.Theme.MutedStyle().Render("No MCP servers")
	}

	var badges []string
	for _, s := range m.Servers {
		if s.Active {
			badges = append(badges, m.Theme.BadgeStyle().Render("⚡ "+s.Name))
		}
	}

	if len(badges) == 0 {
		return m.Theme.MutedStyle().Render("No active MCP servers")
	}

	return strings.Join(badges, " ")
}
