package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const maxTabLabelLength = 20

var (
	dividerLineStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
	dividerActiveTabStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)
	dividerInactiveTabStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))
)

func renderDividerBar(sessions []string, activeIdx int, width int) string {
	if len(sessions) == 0 {
		return ""
	}

	var tabs []string
	for index, name := range sessions {
		label := fmt.Sprintf(" %d: %s ", index+1, truncateLabel(name, maxTabLabelLength))
		if index == activeIdx {
			tabs = append(tabs, dividerActiveTabStyle.Render(label))
		} else {
			tabs = append(tabs, dividerInactiveTabStyle.Render(label))
		}
	}

	separator := dividerLineStyle.Render("│")
	tabBar := strings.Join(tabs, separator)
	remaining := width - lipgloss.Width(tabBar)
	if remaining > 0 {
		tabBar += dividerLineStyle.Render(strings.Repeat("─", remaining))
	}
	return tabBar
}

func truncateLabel(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}
