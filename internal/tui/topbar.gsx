package tui

import (
	tui "github.com/grindlemire/go-tui"
	"github.com/robinojw/dj/internal/tui/theme"
)

type topBar struct {
	branch *tui.State[string]
	cwd    *tui.State[string]
	title  *tui.State[string]
	t      *theme.Theme
}

func NewTopBar(t *theme.Theme, branch, cwd, title *tui.State[string]) *topBar {
	return &topBar{branch: branch, cwd: cwd, title: title, t: t}
}

templ (b *topBar) Render() {
	<div class="flex justify-between px-1 shrink-0">
		<span class="text-cyan font-bold">{b.title.Get()}</span>
		<span class="text-dim">{b.branch.Get() + "  " + b.cwd.Get()}</span>
	</div>
}
