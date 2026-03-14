package tui

import (
	"fmt"
	"strings"

	tui "github.com/grindlemire/go-tui"
	"github.com/robinojw/dj/internal/tui/theme"
)

type diffPager struct {
	app      *tui.App
	diffs    *tui.State[[]storedDiff]
	expanded *tui.State[map[int]bool]
	focused  *tui.State[int]
	scrollY  *tui.State[int]
	onClose  func()
	t        *theme.Theme
}

func NewDiffPager(t *theme.Theme, diffs []storedDiff, onClose func()) *diffPager {
	expandedMap := make(map[int]bool)
	return &diffPager{
		diffs:    tui.NewState(diffs),
		expanded: tui.NewState(expandedMap),
		focused:  tui.NewState(0),
		scrollY:  tui.NewState(0),
		onClose:  onClose,
		t:        t,
	}
}

func (d *diffPager) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKeyStop(tui.KeyEscape, func(ke tui.KeyEvent) {
			d.onClose()
		}),
		tui.OnKeyStop(tui.KeyTab, func(ke tui.KeyEvent) {
			diffs := d.diffs.Get()
			if len(diffs) > 0 {
				d.focused.Update(func(v int) int { return (v + 1) % len(diffs) })
			}
		}),
		tui.OnKeyModStop(tui.KeyTab, tui.ModShift, func(ke tui.KeyEvent) {
			diffs := d.diffs.Get()
			if len(diffs) > 0 {
				d.focused.Update(func(v int) int {
					v--
					if v < 0 {
						v = len(diffs) - 1
					}
					return v
				})
			}
		}),
		tui.OnKeyStop(tui.KeyEnter, func(ke tui.KeyEvent) {
			idx := d.focused.Get()
			d.expanded.Update(func(m map[int]bool) map[int]bool {
				cp := make(map[int]bool, len(m))
				for k, v := range m {
					cp[k] = v
				}
				cp[idx] = !cp[idx]
				return cp
			})
		}),
		tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) {
			d.scrollY.Update(func(v int) int {
				if v > 0 {
					return v - 1
				}
				return 0
			})
		}),
		tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) {
			d.scrollY.Update(func(v int) int { return v + 1 })
		}),
	}
}

templ (d *diffPager) Render() {
	<div class="flex-col h-full border-rounded p-1">
		<span class="text-cyan font-bold">{"Diff Pager (Tab/Shift+Tab: navigate, Enter: expand, Esc: close)"}</span>
		<hr />
		for i, diff := range d.diffs.Get() {
			if d.focused.Get() == i {
				<span class="text-cyan font-bold">{diffIcon(d.expanded.Get()[i]) + " ● " + diff.FilePath + " " + diffStatsStr(diff.DiffLines)}</span>
			} else {
				<span class="text-dim">{diffIcon(d.expanded.Get()[i]) + "   " + diff.FilePath + " " + diffStatsStr(diff.DiffLines)}</span>
			}
			if d.expanded.Get()[i] {
				for _, line := range diff.DiffLines {
					<span class={diffLineClass(line)}>{line}</span>
				}
			}
		}
	</div>
}

func diffIcon(expanded bool) string {
	if expanded {
		return "▼"
	}
	return "▶"
}

func diffStatsStr(lines []string) string {
	stats := calculateDiffStats(lines)
	return fmt.Sprintf("+%d -%d", stats.additions, stats.deletions)
}

func diffLineClass(line string) string {
	if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
		return "text-green"
	} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
		return "text-red"
	} else if strings.HasPrefix(line, "@@") {
		return "text-cyan"
	}
	return "text-dim"
}
