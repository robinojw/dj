package tui

import (
	tui "github.com/grindlemire/go-tui"
	"github.com/robinojw/dj/internal/tui/theme"
)

type cheatSheet struct {
	app     *tui.App
	onClose func()
	t       *theme.Theme
}

func NewCheatSheet(t *theme.Theme, onClose func()) *cheatSheet {
	return &cheatSheet{t: t, onClose: onClose}
}

func (c *cheatSheet) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKeyStop(tui.KeyEscape, func(ke tui.KeyEvent) {
			c.onClose()
		}),
	}
}

type shortcut struct {
	Key  string
	Desc string
}

var shortcuts = []shortcut{
	{"Enter", "Send message"},
	{"Ctrl+E", "Enhance prompt"},
	{"Ctrl+T", "Team view"},
	{"Ctrl+K", "Skills browser"},
	{"Ctrl+M", "MCP manager"},
	{"Ctrl+H", "This cheat sheet"},
	{"Ctrl+F", "Diff pager"},
	{"Tab", "Cycle mode (Confirm/Plan/Turbo)"},
	{"Ctrl+N", "Cycle model"},
	{"Ctrl+Z", "Undo (checkpoint)"},
	{"Ctrl+D", "Toggle debug overlay"},
	{"Esc", "Back / dismiss"},
	{"Ctrl+Q", "Quit"},
}

var modeDescs = []shortcut{
	{"Confirm", "Prompts before each tool execution"},
	{"Plan", "Plans actions before executing"},
	{"Turbo", "Executes without confirmation"},
}

templ (c *cheatSheet) Render() {
	<div class="flex-col border-double p-1 items-center justify-center h-full">
		<span class="text-cyan font-bold">{"  dj — Keyboard Shortcuts                 Ctrl+H  "}</span>
		<hr />
		<div class="flex-col p-1">
			for _, s := range shortcuts {
				<div class="flex-row">
					<span class="text-cyan font-bold w-14">{s.Key}</span>
					<span class="text-dim">{s.Desc}</span>
				</div>
			}
		</div>
		<hr />
		<span class="font-bold text-cyan">{"  Execution Modes"}</span>
		<div class="flex-col p-1">
			for _, m := range modeDescs {
				<div class="flex-row">
					<span class="text-cyan font-bold w-14">{m.Key}</span>
					<span class="text-dim">{m.Desc}</span>
				</div>
			}
		</div>
		<hr />
		<span class="text-dim">{"  [Esc] dismiss"}</span>
	</div>
}
