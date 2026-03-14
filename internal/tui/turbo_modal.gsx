package tui

import (
	tui "github.com/grindlemire/go-tui"
	"github.com/robinojw/dj/internal/tui/theme"
)

type turboModal struct {
	app       *tui.App
	visible   *tui.State[bool]
	confirmed *tui.State[bool]
	onResult  func(confirmed bool)
	t         *theme.Theme
}

func NewTurboModal(t *theme.Theme, onResult func(bool)) *turboModal {
	return &turboModal{
		visible:   tui.NewState(false),
		confirmed: tui.NewState(false),
		onResult:  onResult,
		t:         t,
	}
}

func (m *turboModal) Show() {
	m.visible.Set(true)
	if m.app != nil {
		m.app.SetInlineHeight(10)
	}
}

func (m *turboModal) IsVisible() bool {
	return m.visible.Get()
}

func (m *turboModal) IsConfirmed() bool {
	return m.confirmed.Get()
}

func (m *turboModal) KeyMap() tui.KeyMap {
	if !m.visible.Get() {
		return nil
	}
	return tui.KeyMap{
		tui.OnRuneStop('y', func(ke tui.KeyEvent) {
			m.confirmed.Set(true)
			m.visible.Set(false)
			if m.app != nil {
				m.app.SetInlineHeight(3)
			}
			m.onResult(true)
		}),
		tui.OnRuneStop('n', func(ke tui.KeyEvent) {
			m.visible.Set(false)
			if m.app != nil {
				m.app.SetInlineHeight(3)
			}
			m.onResult(false)
		}),
		tui.OnKeyStop(tui.KeyEscape, func(ke tui.KeyEvent) {
			m.visible.Set(false)
			if m.app != nil {
				m.app.SetInlineHeight(3)
			}
			m.onResult(false)
		}),
	}
}

templ (m *turboModal) Render() {
	if m.visible.Get() {
		<div class="flex-col border-rounded border-red p-1">
			<span class="text-red font-bold">{"⚠ Enable Turbo Mode?"}</span>
			<span class="text-dim">{"Turbo mode executes ALL tools without confirmation."}</span>
			<span class="text-dim">{"This includes:"}</span>
			<span class="text-red">{"  • File creation, modification, and deletion"}</span>
			<span class="text-red">{"  • Shell command execution"}</span>
			<span class="text-red">{"  • Network requests"}</span>
			<span class="text-dim">{""}</span>
			<span class="text-dim">{"[y] confirm  [n/Esc] cancel"}</span>
		</div>
	}
}
