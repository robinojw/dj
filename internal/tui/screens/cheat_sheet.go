package screens

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/tui/theme"
)

// CheatSheetModel displays keyboard shortcuts and mode explanations on startup.
type CheatSheetModel struct {
	width  int
	height int
	theme  *theme.Theme
}

func NewCheatSheetModel(t *theme.Theme) CheatSheetModel {
	return CheatSheetModel{theme: t}
}

func (m CheatSheetModel) Init() tea.Cmd { return nil }

func (m CheatSheetModel) Update(msg tea.Msg) (CheatSheetModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m CheatSheetModel) View() string {
	w := max(min(m.width-4, 60), 0)
	border := strings.Repeat("═", w)

	title := m.theme.AccentStyle().Render("  dj — Keyboard Shortcuts                 Ctrl+H  ")

	accent := m.theme.AccentStyle()
	muted := m.theme.MutedStyle()
	primary := m.theme.PrimaryStyle()

	shortcuts := []struct{ key, desc string }{
		{"Enter", "Send message"},
		{"Ctrl+E", "Enhance prompt"},
		{"Ctrl+T", "Team view"},
		{"Ctrl+K", "Skills browser"},
		{"Ctrl+M", "MCP manager"},
		{"Ctrl+H", "This cheat sheet"},
		{"Tab", "Cycle mode (Confirm/Plan/Turbo)"},
		{"Ctrl+/", "Cycle model"},
		{"Ctrl+Z", "Undo (checkpoint)"},
		{"Ctrl+D", "Toggle debug overlay"},
		{"Esc", "Back / dismiss"},
		{"Ctrl+Q", "Quit"},
	}

	var rows []string
	for _, s := range shortcuts {
		key := accent.Render(padRight(s.key, 12))
		rows = append(rows, "  "+key+muted.Render(s.desc))
	}

	shortcutBlock := strings.Join(rows, "\n")

	modesTitle := primary.Render("  Execution Modes")
	modes := []struct{ name, desc string }{
		{"Confirm", "Prompts before each tool execution"},
		{"Plan", "Plans actions before executing"},
		{"Turbo", "Executes without confirmation"},
	}

	var modeRows []string
	for _, mode := range modes {
		name := accent.Render(padRight(mode.name, 12))
		modeRows = append(modeRows, "  "+name+muted.Render(mode.desc))
	}
	modeBlock := strings.Join(modeRows, "\n")

	footer := muted.Render("  [Esc] dismiss")

	return lipgloss.JoinVertical(lipgloss.Left,
		"╔"+border+"╗",
		title,
		"╠"+border+"╣",
		shortcutBlock,
		"╠"+border+"╣",
		modesTitle,
		modeBlock,
		"╠"+border+"╣",
		footer,
		"╚"+border+"╝",
	)
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}
