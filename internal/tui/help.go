package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const helpKeyColumnWidth = 12

var (
	helpBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39")).
			Padding(1, 2)
	helpTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			MarginBottom(1)
	helpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Width(helpKeyColumnWidth)
	helpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))
)

type keybinding struct {
	key         string
	description string
}

var keybindings = []keybinding{
	{"←/→", "Navigate cards horizontally"},
	{"↑/↓", "Navigate cards vertically"},
	{"Enter", "Open + focus session"},
	{"Space", "Toggle pin/unpin session"},
	{"Tab", "Switch to session panel"},
	{"Esc", "Back / close overlay"},
	{"t", "Toggle tree view"},
	{"n", "New thread"},
	{"Ctrl+B", "Prefix key (tmux-style)"},
	{"Ctrl+B m", "Open context menu"},
	{"Ctrl+B ←/→", "Cycle session panes"},
	{"Ctrl+B 1-9", "Jump to session pane"},
	{"Ctrl+B x", "Unpin focused session"},
	{"Ctrl+B z", "Toggle zoom session"},
	{"?", "Toggle help"},
	{"Ctrl+C", "Quit"},
}

type HelpModel struct{}

func NewHelpModel() HelpModel {
	return HelpModel{}
}

func (help HelpModel) View() string {
	title := helpTitleStyle.Render("Keybindings")

	var lines []string
	for _, binding := range keybindings {
		key := helpKeyStyle.Render(binding.key)
		desc := helpDescStyle.Render(binding.description)
		lines = append(lines, fmt.Sprintf("%s %s", key, desc))
	}

	content := title + "\n" + strings.Join(lines, "\n")
	return helpBorderStyle.Render(content)
}
