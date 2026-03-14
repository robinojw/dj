package tui

import (
	tui "github.com/grindlemire/go-tui"
	"github.com/robinojw/dj/internal/tui/theme"
)

// SkillInfo holds display info for a skill.
type SkillInfo struct {
	Name        string
	Description string
	Source      string // "builtin", "project", "user"
	Implicit    bool
}

type skillBrowser struct {
	app      *tui.App
	skills   *tui.State[[]SkillInfo]
	selected *tui.State[int]
	onClose  func()
	t        *theme.Theme
}

func NewSkillBrowser(t *theme.Theme, onClose func()) *skillBrowser {
	return &skillBrowser{
		skills:   tui.NewState([]SkillInfo{}),
		selected: tui.NewState(0),
		onClose:  onClose,
		t:        t,
	}
}

func (s *skillBrowser) SetSkills(skills []SkillInfo) {
	s.skills.Set(skills)
}

func (s *skillBrowser) KeyMap() tui.KeyMap {
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
			skills := s.skills.Get()
			s.selected.Update(func(v int) int {
				if v < len(skills)-1 {
					return v + 1
				}
				return v
			})
		}),
		tui.OnRuneStop('j', func(ke tui.KeyEvent) {
			skills := s.skills.Get()
			s.selected.Update(func(v int) int {
				if v < len(skills)-1 {
					return v + 1
				}
				return v
			})
		}),
	}
}

func implicitIcon(implicit bool) string {
	if implicit {
		return "⚡"
	}
	return " "
}

templ (s *skillBrowser) Render() {
	<div class="flex-col h-full border-rounded p-1">
		<span class="text-cyan font-bold">{"  Skills Library                           Ctrl+K  "}</span>
		<hr />
		if len(s.skills.Get()) == 0 {
			<span class="text-dim">{"  No skills loaded."}</span>
		} else {
			for i, skill := range s.skills.Get() {
				if s.selected.Get() == i {
					<span class="text-cyan font-bold">{implicitIcon(skill.Implicit) + " /" + skill.Name + " [" + skill.Source + "] " + skill.Description}</span>
				} else {
					<span class="text-dim">{implicitIcon(skill.Implicit) + " /" + skill.Name + " [" + skill.Source + "] " + skill.Description}</span>
				}
			}
		}
		<hr />
		<span class="text-dim">{"  [/name] invoke in chat  [↑/↓] navigate  [Esc] back"}</span>
	</div>
}
