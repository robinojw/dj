package tui

import (
	"fmt"
	"sort"
	"strings"

	tui "github.com/grindlemire/go-tui"
	"github.com/robinojw/dj/internal/modes"
	"github.com/robinojw/dj/internal/tui/theme"
)

type permissionModal struct {
	app     *tui.App
	request *tui.State[*modes.PermissionRequest]
	scope   *tui.State[modes.RememberScope]
	t       *theme.Theme
}

func NewPermissionModal(t *theme.Theme) *permissionModal {
	return &permissionModal{
		request: tui.NewState[*modes.PermissionRequest](nil),
		scope:   tui.NewState(modes.RememberOnce),
		t:       t,
	}
}

func (p *permissionModal) Show(req *modes.PermissionRequest) {
	p.request.Set(req)
	p.scope.Set(modes.RememberOnce)
	if p.app != nil {
		p.app.SetInlineHeight(10)
	}
}

func (p *permissionModal) dismiss() {
	p.request.Set(nil)
	if p.app != nil {
		p.app.SetInlineHeight(3)
	}
}

func (p *permissionModal) Visible() bool {
	return p.request.Get() != nil
}

func (p *permissionModal) KeyMap() tui.KeyMap {
	if p.request.Get() == nil {
		return nil
	}
	return tui.KeyMap{
		tui.OnRuneStop('y', func(ke tui.KeyEvent) {
			req := p.request.Get()
			if req != nil && req.RespCh != nil {
				req.RespCh <- modes.PermissionResp{Allowed: true, RememberFor: p.scope.Get()}
			}
			p.dismiss()
		}),
		tui.OnRuneStop('n', func(ke tui.KeyEvent) {
			req := p.request.Get()
			if req != nil && req.RespCh != nil {
				req.RespCh <- modes.PermissionResp{Allowed: false}
			}
			p.dismiss()
		}),
		tui.OnKeyStop(tui.KeyEscape, func(ke tui.KeyEvent) {
			req := p.request.Get()
			if req != nil && req.RespCh != nil {
				req.RespCh <- modes.PermissionResp{Allowed: false}
			}
			p.dismiss()
		}),
		tui.OnKeyStop(tui.KeyTab, func(ke tui.KeyEvent) {
			p.scope.Update(func(s modes.RememberScope) modes.RememberScope {
				return (s + 1) % 3
			})
		}),
	}
}

templ (p *permissionModal) Render() {
	if p.request.Get() != nil {
		<div class="flex-col border-rounded border-yellow p-1">
			<span class="text-yellow font-bold">{"🔒 Permission Required"}</span>
			<span class="text-cyan">{"Tool: " + p.request.Get().Tool}</span>
			if len(p.request.Get().Args) > 0 {
				<span class="text-dim">{formatPermArgs(p.request.Get().Args)}</span>
			}
			<div class="flex-row">
				<span class="text-dim">{"Scope: "}</span>
				@scopeOption("Once", p.scope.Get() == modes.RememberOnce)
				<span class="text-dim">{" │ "}</span>
				@scopeOption("Session", p.scope.Get() == modes.RememberSession)
				<span class="text-dim">{" │ "}</span>
				@scopeOption("Always", p.scope.Get() == modes.RememberAlways)
			</div>
			<span class="text-dim">{"[y] allow  [n/Esc] deny  [Tab] scope"}</span>
		</div>
	}
}

templ scopeOption(label string, selected bool) {
	if selected {
		<span class="text-yellow font-bold">{"[" + label + "]"}</span>
	} else {
		<span class="text-dim">{label}</span>
	}
}

func formatPermArgs(args map[string]any) string {
	keys := make([]string, 0, len(args))
	for k := range args {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", k, args[k]))
	}
	return strings.Join(parts, ", ")
}
