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

templ (c *cheatSheet) Render() {
	<div class="flex-col border-double p-1 items-center justify-center h-full">
		<span class="text-cyan font-bold">{"  dj — Keyboard Shortcuts                 Ctrl+H  "}</span>
		<hr />
		<div class="flex-col p-1">
			@shortcutRow("Enter", "Send message")
			@shortcutRow("Ctrl+E", "Enhance prompt")
			@shortcutRow("Ctrl+T", "Team view")
			@shortcutRow("Ctrl+K", "Skills browser")
			@shortcutRow("Ctrl+M", "MCP manager")
			@shortcutRow("Ctrl+H", "This cheat sheet")
			@shortcutRow("Ctrl+F", "Diff pager")
			@shortcutRow("Tab", "Cycle mode (Confirm/Plan/Turbo)")
			@shortcutRow("Ctrl+N", "Cycle model")
			@shortcutRow("Ctrl+Z", "Undo (checkpoint)")
			@shortcutRow("Ctrl+D", "Toggle debug overlay")
			@shortcutRow("Esc", "Back / dismiss")
			@shortcutRow("Ctrl+Q", "Quit")
		</div>
		<hr />
		<span class="font-bold text-cyan">{"  Execution Modes"}</span>
		<div class="flex-col p-1">
			@shortcutRow("Confirm", "Prompts before each tool execution")
			@shortcutRow("Plan", "Plans actions before executing")
			@shortcutRow("Turbo", "Executes without confirmation")
		</div>
		<hr />
		<span class="text-dim">{"  [Esc] dismiss"}</span>
	</div>
}

templ shortcutRow(key string, desc string) {
	<div class="flex-row">
		<span class="text-cyan font-bold w-14">{key}</span>
		<span class="text-dim">{desc}</span>
	</div>
}
