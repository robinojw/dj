package tui

import (
	tui "github.com/grindlemire/go-tui"
	"github.com/robinojw/dj/internal/tui/theme"
)

type enhanceScreen struct {
	app      *tui.App
	original *tui.State[string]
	enhanced *tui.State[string]
	loading  *tui.State[bool]
	onClose  func()
	onAccept func(text string)
	t        *theme.Theme
}

func NewEnhanceScreen(t *theme.Theme, onClose func(), onAccept func(string)) *enhanceScreen {
	return &enhanceScreen{
		original: tui.NewState(""),
		enhanced: tui.NewState(""),
		loading:  tui.NewState(false),
		onClose:  onClose,
		onAccept: onAccept,
		t:        t,
	}
}

func (e *enhanceScreen) SetOriginal(text string) {
	e.original.Set(text)
	e.loading.Set(true)
	e.enhanced.Set("")
}

func (e *enhanceScreen) SetEnhanced(text string) {
	e.enhanced.Set(text)
	e.loading.Set(false)
}

func (e *enhanceScreen) KeyMap() tui.KeyMap {
	km := tui.KeyMap{
		tui.OnKeyStop(tui.KeyEscape, func(ke tui.KeyEvent) {
			e.onClose()
		}),
	}
	if e.enhanced.Get() != "" && !e.loading.Get() {
		km = append(km, tui.OnKeyStop(tui.KeyEnter, func(ke tui.KeyEvent) {
			if e.onAccept != nil {
				e.onAccept(e.enhanced.Get())
			}
		}))
	}
	return km
}

templ (e *enhanceScreen) Render() {
	<div class="flex-col h-full border-rounded p-1">
		<span class="text-cyan font-bold">{"  Enhance Prompt                           Ctrl+E  "}</span>
		<hr />
		if e.loading.Get() {
			<span class="text-dim">{"  Enhancing prompt..."}</span>
		} else if e.enhanced.Get() != "" {
			<div class="flex-col p-1">
				<span class="text-dim font-bold">{"BEFORE:"}</span>
				<span class="text-dim">{e.original.Get()}</span>
				<hr />
				<span class="text-cyan font-bold">{"AFTER:"}</span>
				<span class="text-cyan">{e.enhanced.Get()}</span>
			</div>
		} else {
			<span class="text-dim">{"  Type a prompt in chat, then press Ctrl+E to enhance it."}</span>
		}
		<hr />
		<span class="text-dim">{"  [Enter] use enhanced  [Esc] keep original"}</span>
	</div>
}
