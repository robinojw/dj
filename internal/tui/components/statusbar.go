package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/robinojw/dj/internal/tui/theme"
)

// StatusBar displays token counts, cost, context usage, and active MCP servers.
type StatusBar struct {
	InputTokens    int
	OutputTokens   int
	CumulativeCost float64
	ActiveMCPs     []string
	Mode           string // "Build", "Plan"
	Width          int
	Theme          *theme.Theme
}

func NewStatusBar(t *theme.Theme) StatusBar {
	return StatusBar{Theme: t}
}

func (s StatusBar) View() string {
	ctxPct := float64(s.InputTokens) / 400_000 * 100
	if ctxPct > 100 {
		ctxPct = 100
	}
	ctxBar := progressBar(ctxPct, 20, s.Theme)

	var mcpBadges string
	if len(s.ActiveMCPs) > 0 {
		badges := make([]string, len(s.ActiveMCPs))
		for i, m := range s.ActiveMCPs {
			badges[i] = s.Theme.BadgeStyle().Render("⚡ " + m)
		}
		mcpBadges = " " + strings.Join(badges, " ")
	}

	var modeBadge string
	if s.Mode != "" {
		modeBadge = s.Theme.BadgeStyle().Render(s.Mode) + "  "
	}

	content := fmt.Sprintf("%sCTX %s %.1f%%  OUT %s  $%.4f%s",
		modeBadge,
		ctxBar, ctxPct,
		humanize.Comma(int64(s.OutputTokens)),
		s.CumulativeCost,
		mcpBadges,
	)

	return s.Theme.StatusStyle().
		Width(s.Width).
		Render(content)
}

func progressBar(pct float64, width int, t *theme.Theme) string {
	filled := int(pct / 100 * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	barColor := t.ProgressBarColors(pct)
	filledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(barColor))
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Colors.Muted))

	return filledStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", empty))
}
