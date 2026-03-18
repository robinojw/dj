package tui

import "github.com/charmbracelet/lipgloss"

var (
	headerTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39"))
	headerHintStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))
)

const headerTitle = "DJ — Codex TUI Visualizer"

var headerHints = []string{
	"n: new",
	"s: select",
	"Enter: open",
	"?: help",
	"t: tree",
	"Ctrl+B: prefix",
}

const headerHintSeparator = "  "

type HeaderBar struct {
	width int
}

func NewHeaderBar(width int) HeaderBar {
	return HeaderBar{width: width}
}

func (header *HeaderBar) SetWidth(width int) {
	header.width = width
}

func (header HeaderBar) View() string {
	title := headerTitleStyle.Render(headerTitle)

	hints := ""
	for index, hint := range headerHints {
		if index > 0 {
			hints += headerHintSeparator
		}
		hints += hint
	}
	renderedHints := headerHintStyle.Render(hints)

	gap := header.width - lipgloss.Width(title) - lipgloss.Width(renderedHints)
	if gap < 1 {
		gap = 1
	}
	padding := lipgloss.NewStyle().Width(gap).Render("")
	return title + padding + renderedHints
}
