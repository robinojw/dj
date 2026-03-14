package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/tui/theme"
)

// MCPServerInfo holds display info for a configured MCP server.
type MCPServerInfo struct {
	Name      string
	Type      string // "stdio", "http", "sse"
	ToolCount int
	Active    bool
}

// MCPToggleMsg is sent when a server is toggled on/off.
type MCPToggleMsg struct {
	Name   string
	Active bool
}

// MCPManagerModel is the MCP server browser screen.
type MCPManagerModel struct {
	servers  []MCPServerInfo
	selected int
	width    int
	height   int
	theme    *theme.Theme
}

func NewMCPManagerModel(t *theme.Theme) MCPManagerModel {
	return MCPManagerModel{theme: t}
}

func (m *MCPManagerModel) SetServers(servers []MCPServerInfo) {
	m.servers = servers
}

func (m MCPManagerModel) Init() tea.Cmd { return nil }

func (m MCPManagerModel) Update(msg tea.Msg) (MCPManagerModel, tea.Cmd) {
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
			if m.selected < len(m.servers)-1 {
				m.selected++
			}
		case "enter":
			if len(m.servers) > 0 {
				s := &m.servers[m.selected]
				s.Active = !s.Active
				return m, func() tea.Msg {
					return MCPToggleMsg{Name: s.Name, Active: s.Active}
				}
			}
		}
	}

	return m, nil
}

func (m MCPManagerModel) View() string {
	w := max(min(m.width-4, 56), 0)
	border := strings.Repeat("═", w)

	title := m.theme.AccentStyle().Render("  MCP Servers                              Ctrl+M  ")

	var rows []string
	for i, s := range m.servers {
		marker := "○"
		if s.Active {
			marker = "●"
		}
		status := m.theme.ErrorStyle().Render("✗ off")
		if s.Active {
			status = m.theme.SuccessStyle().Render("✓ active")
		}

		line := fmt.Sprintf("  %s %-20s [%-5s] %2d tools  %s",
			marker, s.Name, s.Type, s.ToolCount, status)

		if i == m.selected {
			line = m.theme.SelectedStyle().Render(line)
		}
		rows = append(rows, line)
	}

	if len(rows) == 0 {
		rows = append(rows, m.theme.MutedStyle().Render("  No MCP servers configured"))
	}

	body := strings.Join(rows, "\n")
	footer := m.theme.MutedStyle().Render("  [Enter] toggle  [i] inspect tools  [+] add server")

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
