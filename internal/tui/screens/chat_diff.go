package screens

import (
	"fmt"
	"strings"
)

// renderDiff renders a single collapsible diff with focus indicator.
func (m *ChatModel) renderDiff(diff CollapsibleDiff, focused bool) []string {
	var lines []string

	// Summary line
	icon := "▶" // collapsed
	if diff.Expanded {
		icon = "▼" // expanded
	}

	focusMarker := " "
	if focused {
		focusMarker = "●"
	}

	stats := m.calculateDiffStats(diff.DiffLines)
	summaryStyle := m.theme.MutedStyle()
	if focused {
		summaryStyle = m.theme.AccentStyle().Bold(true)
	}

	summary := fmt.Sprintf("%s %s Modified: %s  (+%d -%d)",
		focusMarker, icon, diff.FilePath, stats.additions, stats.deletions)
	lines = append(lines, summaryStyle.Render(summary))

	// Expanded content
	if diff.Expanded {
		for _, line := range diff.DiffLines {
			if line == "" {
				continue
			}
			styledLine := m.styleDiffLine(line)
			lines = append(lines, "  "+styledLine)
		}
	}

	return lines
}

// diffStats holds addition and deletion counts.
type diffStats struct {
	additions int
	deletions int
}

// calculateDiffStats counts + and - lines in the diff.
func (m *ChatModel) calculateDiffStats(lines []string) diffStats {
	stats := diffStats{}
	for _, line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			stats.additions++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			stats.deletions++
		}
	}
	return stats
}

// styleDiffLine applies color styling based on diff line type.
func (m *ChatModel) styleDiffLine(line string) string {
	if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
		return m.theme.SuccessStyle().Render(line)
	} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
		return m.theme.ErrorStyle().Render(line)
	} else if strings.HasPrefix(line, "@@") {
		return m.theme.PrimaryStyle().Render(line)
	}
	return m.theme.MutedStyle().Render(line)
}
