package tui

import (
	"fmt"
	"time"

	tui "github.com/grindlemire/go-tui"
	"github.com/robinojw/dj/internal/tui/theme"
)

const maxDebugEntries = 50

// DebugEntry represents a single debug log entry.
type DebugEntry struct {
	Time    time.Time
	Level   string // "ERROR", "WARN", "INFO"
	Message string
}

type debugOverlayComponent struct {
	app     *tui.App
	entries *tui.State[[]DebugEntry]
	visible *tui.State[bool]
	t       *theme.Theme
}

func NewDebugOverlay(t *theme.Theme) *debugOverlayComponent {
	return &debugOverlayComponent{
		entries: tui.NewState([]DebugEntry{}),
		visible: tui.NewState(false),
		t:       t,
	}
}

func (d *debugOverlayComponent) Toggle() {
	d.visible.Update(func(v bool) bool { return !v })
}

func (d *debugOverlayComponent) IsVisible() bool {
	return d.visible.Get()
}

func (d *debugOverlayComponent) AddError(msg string) {
	d.add("ERROR", msg)
}

func (d *debugOverlayComponent) AddWarn(msg string) {
	d.add("WARN", msg)
}

func (d *debugOverlayComponent) AddInfo(msg string) {
	d.add("INFO", msg)
}

func (d *debugOverlayComponent) add(level, msg string) {
	d.entries.Update(func(entries []DebugEntry) []DebugEntry {
		entries = append(entries, DebugEntry{
			Time:    time.Now(),
			Level:   level,
			Message: msg,
		})
		if len(entries) > maxDebugEntries {
			entries = entries[len(entries)-maxDebugEntries:]
		}
		return entries
	})
}

func debugLevelClass(level string) string {
	switch level {
	case "ERROR":
		return "text-red font-bold"
	case "WARN":
		return "text-yellow font-bold"
	default:
		return "text-dim"
	}
}

templ (d *debugOverlayComponent) Render() {
	if d.visible.Get() {
		<div class="flex-col border-rounded border-red p-1 w-1/3">
			<span class="text-red font-bold">{"DEBUG"}</span>
			<hr />
			if len(d.entries.Get()) == 0 {
				<span class="text-dim">{"No debug messages"}</span>
			} else {
				for _, entry := range d.entries.Get() {
					<div class="flex-row">
						<span class="text-dim">{entry.Time.Format("15:04:05") + " "}</span>
						<span class={debugLevelClass(entry.Level)}>{entry.Level + " "}</span>
						<span class="text-dim">{truncateMsg(entry.Message, 60)}</span>
					</div>
				}
			}
		</div>
	}
}

func truncateMsg(msg string, maxLen int) string {
	if len(msg) <= maxLen {
		return msg
	}
	return msg[:maxLen-3] + "..."
}

// QueueAdd adds a debug entry from a goroutine (thread-safe via QueueUpdate).
func (d *debugOverlayComponent) QueueAdd(level, msg string) {
	if d.app != nil {
		d.app.QueueUpdate(func() {
			d.add(level, msg)
		})
	}
}

// QueueError adds an error from a goroutine.
func (d *debugOverlayComponent) QueueError(msg string) {
	d.QueueAdd("ERROR", msg)
}

// QueueInfo adds an info message from a goroutine.
func (d *debugOverlayComponent) QueueInfo(format string, args ...any) {
	d.QueueAdd("INFO", fmt.Sprintf(format, args...))
}
