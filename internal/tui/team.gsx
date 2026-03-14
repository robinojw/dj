package tui

import (
	tui "github.com/grindlemire/go-tui"
	"github.com/robinojw/dj/internal/tui/theme"
)

// AgentStatus holds the state of a single agent in the team view.
type AgentStatus struct {
	ID       string
	Name     string
	Status   string // "running", "idle", "completed", "error", "skipped"
	ParentID string
	Output   string
}

type teamScreen struct {
	app      *tui.App
	agents   *tui.State[[]AgentStatus]
	selected *tui.State[int]
	onClose  func()
	t        *theme.Theme
}

func NewTeamScreen(t *theme.Theme, onClose func()) *teamScreen {
	return &teamScreen{
		agents:   tui.NewState([]AgentStatus{}),
		selected: tui.NewState(0),
		onClose:  onClose,
		t:        t,
	}
}

func (s *teamScreen) SetAgents(agents []AgentStatus) {
	s.agents.Set(agents)
}

func agentStatusIcon(status string) string {
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

func (s *teamScreen) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKeyStop(tui.KeyEscape, func(ke tui.KeyEvent) {
			s.onClose()
		}),
		tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) {
			s.selected.Update(func(v int) int {
				if v > 0 {
					return v - 1
				}
				return 0
			})
		}),
		tui.OnRuneStop('k', func(ke tui.KeyEvent) {
			s.selected.Update(func(v int) int {
				if v > 0 {
					return v - 1
				}
				return 0
			})
		}),
		tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) {
			agents := s.agents.Get()
			s.selected.Update(func(v int) int {
				if v < len(agents)-1 {
					return v + 1
				}
				return v
			})
		}),
		tui.OnRuneStop('j', func(ke tui.KeyEvent) {
			agents := s.agents.Get()
			s.selected.Update(func(v int) int {
				if v < len(agents)-1 {
					return v + 1
				}
				return v
			})
		}),
	}
}

templ (s *teamScreen) Render() {
	<div class="flex-col h-full border-rounded p-1">
		<span class="text-cyan font-bold">{"  Team View                                Ctrl+T  "}</span>
		<hr />
		if len(s.agents.Get()) == 0 {
			<span class="text-dim">{"  No agents running. Start a complex task to see the team view."}</span>
		} else {
			<div class="flex-col">
				for i, agent := range s.agents.Get() {
					if s.selected.Get() == i {
						<span class="text-cyan font-bold">{"● " + agentStatusIcon(agent.Status) + " " + agent.Name + " [" + agent.Status + "]"}</span>
					} else {
						<span class="text-dim">{"  " + agentStatusIcon(agent.Status) + " " + agent.Name + " [" + agent.Status + "]"}</span>
					}
				}
			</div>
			<hr />
			<div class="flex-col">
				<span class="text-cyan font-bold">{"Output:"}</span>
				if s.selected.Get() < len(s.agents.Get()) {
					<span class="text-dim">{s.agents.Get()[s.selected.Get()].Output}</span>
				}
			</div>
		}
		<hr />
		<span class="text-dim">{"  [↑/↓] navigate  [Esc] back"}</span>
	</div>
}
