package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/tui/components"
	"github.com/robinojw/dj/internal/tui/theme"
)

// TeamSpawnedMsg triggers a switch to the team screen.
type TeamSpawnedMsg struct{}

// AgentSelectedMsg is emitted when the user navigates to a different agent in the graph.
type AgentSelectedMsg struct {
	ID string
}

// AgentStatus represents the state of a single agent.
type AgentStatus struct {
	ID       string
	Name     string
	Status   string // "running", "idle", "completed", "error", "skipped"
	ParentID string
	Output   string
}

// TeamModel is the split-pane team orchestration screen.
type TeamModel struct {
	agents      map[string]*AgentStatus
	topology    []Edge
	selected    string
	graph       AgentGraphWidget
	output      AgentOutputWidget
	statusBar   components.StatusBar
	graphHeight int
	outputHeight int
	width       int
	height      int
	theme       *theme.Theme
}

// Edge represents a parent → child relationship in the agent graph.
type Edge struct {
	From string
	To   string
}

func NewTeamModel(t *theme.Theme) TeamModel {
	return TeamModel{
		agents:    make(map[string]*AgentStatus),
		theme:     t,
		statusBar: components.NewStatusBar(t),
	}
}

func (m TeamModel) Init() tea.Cmd { return nil }

func (m TeamModel) Update(msg tea.Msg) (TeamModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.graphHeight = msg.Height * 40 / 100
		m.outputHeight = msg.Height - m.graphHeight - 1
		m.statusBar.Width = msg.Width

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "down", "left", "right":
			m.graph, _ = m.graph.Update(msg)
		}

	case AgentSelectedMsg:
		m.selected = msg.ID
	}

	return m, nil
}

func (m TeamModel) View() string {
	graphView := m.renderGraph()
	outputView := m.renderOutput()

	graphPanel := m.theme.PanelStyle().
		Width(m.width - 2).
		Height(m.graphHeight).
		Render(graphView)

	outputPanel := m.theme.PanelStyle().
		Width(m.width - 2).
		Height(m.outputHeight).
		Render(outputView)

	return lipgloss.JoinVertical(lipgloss.Left,
		graphPanel,
		outputPanel,
		m.statusBar.View(),
	)
}

func (m TeamModel) renderGraph() string {
	if len(m.agents) == 0 {
		return m.theme.MutedStyle().Render("No agents spawned. Press Ctrl+T to force team mode.")
	}

	var lines []string
	// Find root agents (no parent)
	var roots []string
	for id, a := range m.agents {
		if a.ParentID == "" {
			roots = append(roots, id)
		}
	}

	for _, rootID := range roots {
		agent := m.agents[rootID]
		marker := "●"
		if rootID == m.selected {
			marker = "▶"
		}
		lines = append(lines, fmt.Sprintf("  [%s %s] %s", marker, agent.Name, statusIcon(agent.Status)))

		// Children
		for id, a := range m.agents {
			if a.ParentID == rootID {
				childMarker := "○"
				if id == m.selected {
					childMarker = "▶"
				}
				lines = append(lines, fmt.Sprintf("    ├── [%s %s] %s", childMarker, a.Name, statusIcon(a.Status)))
			}
		}
	}

	return strings.Join(lines, "\n")
}

func (m TeamModel) renderOutput() string {
	if m.selected == "" {
		return m.theme.MutedStyle().Render("Select an agent to view output")
	}
	agent, ok := m.agents[m.selected]
	if !ok {
		return m.theme.MutedStyle().Render("Agent not found")
	}
	if agent.Output == "" {
		return m.theme.MutedStyle().Render("No output yet...")
	}
	return agent.Output
}

func statusIcon(status string) string {
	switch status {
	case "running":
		return "⏳"
	case "completed":
		return "✅"
	case "error":
		return "❌"
	case "skipped":
		return "⏭️"
	default:
		return "💤"
	}
}

// --- Agent Graph Widget (embedded) ---

type AgentGraphWidget struct {
	selected string
}

func (g AgentGraphWidget) Update(msg tea.Msg) (AgentGraphWidget, tea.Cmd) {
	// Navigation will be wired when agents are populated
	return g, nil
}

// --- Agent Output Widget (embedded) ---

type AgentOutputWidget struct {
	content string
}

func (o AgentOutputWidget) View() string {
	return o.content
}
