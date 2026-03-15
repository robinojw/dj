package tui

import (
	"fmt"

	tui "github.com/grindlemire/go-tui"
	"github.com/robinojw/dj/internal/tui/theme"
)

type AgentStatus struct {
	ID       string
	Name     string
	Status   string
	ParentID string
	Output   string
}

type teamScreen struct {
	app         *tui.App
	agents      *tui.State[[]AgentStatus]
	cursor      *tui.State[cursorPos]
	onClose     func()
	onOpenAgent func(agentID string)
	onSplitView func()
	t           *theme.Theme
}

func NewTeamScreen(t *theme.Theme, onClose func(), onOpenAgent func(string), onSplitView func()) *teamScreen {
	return &teamScreen{
		agents:      tui.NewState([]AgentStatus{}),
		cursor:      tui.NewState(cursorPos{}),
		onClose:     onClose,
		onOpenAgent: onOpenAgent,
		onSplitView: onSplitView,
		t:           t,
	}
}

func (s *teamScreen) SetAgents(agents []AgentStatus) {
	s.agents.Set(agents)
}

func (s *teamScreen) SetAgentStatus(workerID, status string) {
	s.agents.Update(func(agents []AgentStatus) []AgentStatus {
		for i, agent := range agents {
			if agent.ID == workerID {
				agents[i].Status = status
			}
		}
		return agents
	})
}

func (s *teamScreen) AppendAgentDelta(workerID, delta string) {
	s.agents.Update(func(agents []AgentStatus) []AgentStatus {
		for i, agent := range agents {
			if agent.ID == workerID {
				agents[i].Output += delta
				if agents[i].Status != "running" {
					agents[i].Status = "running"
				}
			}
		}
		return agents
	})
}

func (s *teamScreen) AppendToolCall(workerID, content string) {
	s.agents.Update(func(agents []AgentStatus) []AgentStatus {
		for i, agent := range agents {
			if agent.ID == workerID {
				agents[i].Output += "\n[Tool] " + content
			}
		}
		return agents
	})
}

func (s *teamScreen) AppendToolResult(workerID, content string) {
	s.agents.Update(func(agents []AgentStatus) []AgentStatus {
		for i, agent := range agents {
			if agent.ID == workerID {
				agents[i].Output += "\n[Result] " + content
			}
		}
		return agents
	})
}

func agentStatusIcon(status string) string {
	switch status {
	case "running":
		return ">"
	case "completed":
		return "+"
	case "error":
		return "!"
	case "skipped":
		return "-"
	default:
		return "."
	}
}

func (s *teamScreen) cursorAgentID() string {
	layers := buildDAGLayers(s.agents.Get())
	pos := s.cursor.Get()
	if pos.Col < len(layers) && pos.Row < len(layers[pos.Col]) {
		return layers[pos.Col][pos.Row].ID
	}
	return ""
}

func (s *teamScreen) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKeyStop(tui.KeyEscape, func(ke tui.KeyEvent) { s.onClose() }),
		tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) {
			s.cursor.Update(func(p cursorPos) cursorPos {
				if p.Row > 0 {
					p.Row--
				}
				return p
			})
		}),
		tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) {
			layers := buildDAGLayers(s.agents.Get())
			s.cursor.Update(func(p cursorPos) cursorPos {
				if p.Col < len(layers) {
					maxRow := len(layers[p.Col]) - 1
					if p.Row < maxRow {
						p.Row++
					}
				}
				return p
			})
		}),
		tui.OnKey(tui.KeyLeft, func(ke tui.KeyEvent) {
			s.cursor.Update(func(p cursorPos) cursorPos {
				if p.Col > 0 {
					p.Col--
					p.Row = 0
				}
				return p
			})
		}),
		tui.OnKey(tui.KeyRight, func(ke tui.KeyEvent) {
			layers := buildDAGLayers(s.agents.Get())
			s.cursor.Update(func(p cursorPos) cursorPos {
				if p.Col < len(layers)-1 {
					p.Col++
					p.Row = 0
				}
				return p
			})
		}),
		tui.OnKeyStop(tui.KeyEnter, func(ke tui.KeyEvent) {
			if agentID := s.cursorAgentID(); agentID != "" && s.onOpenAgent != nil {
				s.onOpenAgent(agentID)
			}
		}),
		tui.OnRuneStop('/', func(ke tui.KeyEvent) {
			if s.onSplitView != nil {
				s.onSplitView()
			}
		}),
		tui.OnRuneStop('k', func(ke tui.KeyEvent) {
			s.cursor.Update(func(p cursorPos) cursorPos {
				if p.Row > 0 {
					p.Row--
				}
				return p
			})
		}),
		tui.OnRuneStop('j', func(ke tui.KeyEvent) {
			layers := buildDAGLayers(s.agents.Get())
			s.cursor.Update(func(p cursorPos) cursorPos {
				if p.Col < len(layers) {
					maxRow := len(layers[p.Col]) - 1
					if p.Row < maxRow {
						p.Row++
					}
				}
				return p
			})
		}),
		tui.OnRuneStop('h', func(ke tui.KeyEvent) {
			s.cursor.Update(func(p cursorPos) cursorPos {
				if p.Col > 0 {
					p.Col--
					p.Row = 0
				}
				return p
			})
		}),
		tui.OnRuneStop('l', func(ke tui.KeyEvent) {
			layers := buildDAGLayers(s.agents.Get())
			s.cursor.Update(func(p cursorPos) cursorPos {
				if p.Col < len(layers)-1 {
					p.Col++
					p.Row = 0
				}
				return p
			})
		}),
	}
}

func agentLabel(agent AgentStatus) string {
	return fmt.Sprintf("[%s] %s (%s)", agentStatusIcon(agent.Status), agent.Name, agent.Status)
}

templ (s *teamScreen) Render() {
	<div class="flex-col h-full border-rounded p-1">
		<span class="text-cyan font-bold">{"  Team View  (Enter: open session  /: split  Esc: back)"}</span>
		<hr />
		if len(s.agents.Get()) == 0 {
			<span class="text-dim">{"  No agents running."}</span>
		} else {
			<div class="flex flex-row gap-4 flex-1">
				for col, layer := range buildDAGLayers(s.agents.Get()) {
					<div class="flex-col gap-1">
						for row, agent := range layer {
							if s.cursor.Get().Col == col && s.cursor.Get().Row == row {
								<div class="border-rounded border-cyan p-1 flex-col" width={24}>
									<span class="text-cyan font-bold">{agent.Name}</span>
									<span class="text-cyan">{agentStatusIcon(agent.Status) + " " + agent.Status}</span>
								</div>
							} else {
								<div class="border-rounded p-1 flex-col" width={24}>
									<span class="font-bold">{agent.Name}</span>
									<span class="text-dim">{agentStatusIcon(agent.Status) + " " + agent.Status}</span>
								</div>
							}
						}
					</div>
				}
			</div>
		}
		<hr />
		<span class="text-dim">{"  [h/l] columns  [j/k] rows  [Enter] open  [/] split"}</span>
	</div>
}
