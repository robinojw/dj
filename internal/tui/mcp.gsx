package tui

import (
	"fmt"

	tui "github.com/grindlemire/go-tui"
	"github.com/robinojw/dj/internal/tui/theme"
)

// MCPServerInfo holds display info for an MCP server.
type MCPServerInfo struct {
	Name      string
	Type      string // "stdio", "http", "sse"
	ToolCount int
	Active    bool
}

type mcpManager struct {
	app      *tui.App
	servers  *tui.State[[]MCPServerInfo]
	selected *tui.State[int]
	onClose  func()
	onToggle func(name string, active bool)
	t        *theme.Theme
}

func NewMCPManager(t *theme.Theme, onClose func(), onToggle func(string, bool)) *mcpManager {
	return &mcpManager{
		servers:  tui.NewState([]MCPServerInfo{}),
		selected: tui.NewState(0),
		onClose:  onClose,
		onToggle: onToggle,
		t:        t,
	}
}

func (m *mcpManager) SetServers(servers []MCPServerInfo) {
	m.servers.Set(servers)
}

func (m *mcpManager) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKeyStop(tui.KeyEscape, func(ke tui.KeyEvent) {
			m.onClose()
		}),
		tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) {
			m.selected.Update(func(v int) int {
				if v > 0 {
					return v - 1
				}
				return 0
			})
		}),
		tui.OnRuneStop('k', func(ke tui.KeyEvent) {
			m.selected.Update(func(v int) int {
				if v > 0 {
					return v - 1
				}
				return 0
			})
		}),
		tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) {
			servers := m.servers.Get()
			m.selected.Update(func(v int) int {
				if v < len(servers)-1 {
					return v + 1
				}
				return v
			})
		}),
		tui.OnRuneStop('j', func(ke tui.KeyEvent) {
			servers := m.servers.Get()
			m.selected.Update(func(v int) int {
				if v < len(servers)-1 {
					return v + 1
				}
				return v
			})
		}),
		tui.OnKeyStop(tui.KeyEnter, func(ke tui.KeyEvent) {
			servers := m.servers.Get()
			idx := m.selected.Get()
			if idx < len(servers) {
				srv := servers[idx]
				if m.onToggle != nil {
					m.onToggle(srv.Name, !srv.Active)
				}
			}
		}),
	}
}

func mcpStatusIcon(active bool) string {
	if active {
		return "● ✓"
	}
	return "○ ✗"
}

templ (m *mcpManager) Render() {
	<div class="flex-col h-full border-rounded p-1">
		<span class="text-cyan font-bold">{"  MCP Servers                              Ctrl+M  "}</span>
		<hr />
		if len(m.servers.Get()) == 0 {
			<span class="text-dim">{"  No MCP servers configured."}</span>
		} else {
			for i, srv := range m.servers.Get() {
				if m.selected.Get() == i {
					<span class="text-cyan font-bold">{fmt.Sprintf("  %s %s [%s] %d tools", mcpStatusIcon(srv.Active), srv.Name, srv.Type, srv.ToolCount)}</span>
				} else {
					<span class="text-dim">{fmt.Sprintf("  %s %s [%s] %d tools", mcpStatusIcon(srv.Active), srv.Name, srv.Type, srv.ToolCount)}</span>
				}
			}
		}
		<hr />
		<span class="text-dim">{"  [Enter] toggle  [↑/↓] navigate  [Esc] back"}</span>
	</div>
}
