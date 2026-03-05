package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/tui/theme"
)

const maxDebugEntries = 50

// DebugEntry represents a single debug log entry.
type DebugEntry struct {
	Time    time.Time
	Level   string // "ERROR", "WARN", "INFO"
	Message string
}

// DebugOverlay displays error and debug messages in a floating panel.
type DebugOverlay struct {
	entries []DebugEntry
	visible bool
	width   int
	height  int
	theme   *theme.Theme
}

func NewDebugOverlay(t *theme.Theme) DebugOverlay {
	return DebugOverlay{theme: t}
}

func (d *DebugOverlay) SetSize(w, h int) {
	d.width = w
	d.height = h
}

func (d *DebugOverlay) Toggle() {
	d.visible = !d.visible
}

func (d *DebugOverlay) IsVisible() bool {
	return d.visible
}

func (d *DebugOverlay) AddError(msg string) {
	d.add("ERROR", msg)
}

func (d *DebugOverlay) AddWarn(msg string) {
	d.add("WARN", msg)
}

func (d *DebugOverlay) AddInfo(msg string) {
	d.add("INFO", msg)
}

func (d *DebugOverlay) add(level, msg string) {
	d.entries = append(d.entries, DebugEntry{
		Time:    time.Now(),
		Level:   level,
		Message: msg,
	})
	if len(d.entries) > maxDebugEntries {
		d.entries = d.entries[len(d.entries)-maxDebugEntries:]
	}
}

func (d *DebugOverlay) Clear() {
	d.entries = nil
}

func (d DebugOverlay) View() string {
	if !d.visible || d.width == 0 || d.height == 0 {
		return ""
	}

	panelW := d.width / 3
	if panelW < 30 {
		panelW = 30
	}
	panelH := d.height / 3
	if panelH < 5 {
		panelH = 5
	}

	// Build log lines (most recent at bottom)
	maxLines := panelH - 3 // border + title + padding
	if maxLines < 1 {
		maxLines = 1
	}
	start := 0
	if len(d.entries) > maxLines {
		start = len(d.entries) - maxLines
	}

	var lines []string
	for _, e := range d.entries[start:] {
		ts := e.Time.Format("15:04:05")
		var levelStyle lipgloss.Style
		switch e.Level {
		case "ERROR":
			levelStyle = d.theme.ErrorStyle()
		case "WARN":
			levelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(d.theme.Colors.Warning)).
				Bold(true)
		default:
			levelStyle = d.theme.MutedStyle()
		}
		line := fmt.Sprintf("%s %s %s",
			d.theme.MutedStyle().Render(ts),
			levelStyle.Render(e.Level),
			e.Message,
		)
		// Truncate to fit panel width
		if lipgloss.Width(line) > panelW-4 {
			line = line[:panelW-7] + "..."
		}
		lines = append(lines, line)
	}

	if len(lines) == 0 {
		lines = append(lines, d.theme.MutedStyle().Render("No debug messages"))
	}

	content := strings.Join(lines, "\n")

	panel := lipgloss.NewStyle().
		Width(panelW).
		Height(panelH).
		Background(lipgloss.Color(d.theme.Colors.PanelBg)).
		Foreground(lipgloss.Color(d.theme.Colors.Foreground)).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(d.theme.Colors.Error)).
		Padding(0, 1).
		Render(d.theme.ErrorStyle().Render("DEBUG") + "\n" + content)

	// Position top-right
	leftPad := d.width - lipgloss.Width(panel)
	if leftPad < 0 {
		leftPad = 0
	}

	return lipgloss.NewStyle().
		PaddingLeft(leftPad).
		Render(panel)
}
