package screens

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/tui/theme"
)

func TestTeamModel_InitialState(t *testing.T) {
	m := NewTeamModel(theme.DefaultTheme())

	if len(m.agents) != 0 {
		t.Errorf("expected no agents initially, got %d", len(m.agents))
	}
	if m.selected != "" {
		t.Errorf("expected empty selection, got %q", m.selected)
	}
}

func TestTeamModel_WindowResize(t *testing.T) {
	m := NewTeamModel(theme.DefaultTheme())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
	// Graph takes 40% of height.
	wantGraphH := 50 * 40 / 100
	if m.graphHeight != wantGraphH {
		t.Errorf("graphHeight = %d, want %d", m.graphHeight, wantGraphH)
	}
	// Output takes remaining minus 1 for status bar.
	wantOutputH := 50 - wantGraphH - 1
	if m.outputHeight != wantOutputH {
		t.Errorf("outputHeight = %d, want %d", m.outputHeight, wantOutputH)
	}
}

func TestTeamModel_AgentSelection(t *testing.T) {
	m := NewTeamModel(theme.DefaultTheme())

	m, _ = m.Update(AgentSelectedMsg{ID: "agent-1"})
	if m.selected != "agent-1" {
		t.Errorf("selected = %q, want %q", m.selected, "agent-1")
	}

	m, _ = m.Update(AgentSelectedMsg{ID: "agent-2"})
	if m.selected != "agent-2" {
		t.Errorf("selected = %q, want %q", m.selected, "agent-2")
	}
}

func TestTeamModel_RenderGraph_EmptyShowsPlaceholder(t *testing.T) {
	m := NewTeamModel(theme.DefaultTheme())

	graph := m.renderGraph()
	if !strings.Contains(graph, "No agents spawned") {
		t.Errorf("empty graph should show placeholder, got %q", graph)
	}
}

func TestTeamModel_RenderGraph_WithAgents(t *testing.T) {
	m := NewTeamModel(theme.DefaultTheme())
	m.agents["root"] = &AgentStatus{
		ID:     "root",
		Name:   "Orchestrator",
		Status: "running",
	}
	m.agents["child-1"] = &AgentStatus{
		ID:       "child-1",
		Name:     "Worker-1",
		Status:   "completed",
		ParentID: "root",
	}

	graph := m.renderGraph()
	if !strings.Contains(graph, "Orchestrator") {
		t.Error("graph should contain root agent name")
	}
	if !strings.Contains(graph, "Worker-1") {
		t.Error("graph should contain child agent name")
	}
}

func TestTeamModel_RenderOutput_NoSelection(t *testing.T) {
	m := NewTeamModel(theme.DefaultTheme())

	out := m.renderOutput()
	if !strings.Contains(out, "Select an agent") {
		t.Errorf("expected placeholder, got %q", out)
	}
}

func TestTeamModel_RenderOutput_SelectedAgentWithOutput(t *testing.T) {
	m := NewTeamModel(theme.DefaultTheme())
	m.agents["a1"] = &AgentStatus{
		ID:     "a1",
		Name:   "Worker",
		Output: "task completed successfully",
	}
	m.selected = "a1"

	out := m.renderOutput()
	if out != "task completed successfully" {
		t.Errorf("output = %q, want %q", out, "task completed successfully")
	}
}

func TestTeamModel_RenderOutput_SelectedAgentNoOutput(t *testing.T) {
	m := NewTeamModel(theme.DefaultTheme())
	m.agents["a1"] = &AgentStatus{ID: "a1", Name: "Worker"}
	m.selected = "a1"

	out := m.renderOutput()
	if !strings.Contains(out, "No output yet") {
		t.Errorf("expected no-output placeholder, got %q", out)
	}
}

func TestTeamModel_RenderOutput_MissingAgent(t *testing.T) {
	m := NewTeamModel(theme.DefaultTheme())
	m.selected = "nonexistent"

	out := m.renderOutput()
	if !strings.Contains(out, "Agent not found") {
		t.Errorf("expected not-found message, got %q", out)
	}
}

func TestTeamModel_StatusIcons(t *testing.T) {
	tests := []struct {
		status string
		icon   string
	}{
		{"running", "⏳"},
		{"completed", "✅"},
		{"error", "❌"},
		{"skipped", "⏭️"},
		{"idle", "💤"},
		{"unknown", "💤"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			if got := statusIcon(tt.status); got != tt.icon {
				t.Errorf("statusIcon(%q) = %q, want %q", tt.status, got, tt.icon)
			}
		})
	}
}

func TestTeamModel_View_DoesNotPanic(t *testing.T) {
	m := NewTeamModel(theme.DefaultTheme())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	v := m.View()
	if v == "" {
		t.Error("View() returned empty string")
	}
}
