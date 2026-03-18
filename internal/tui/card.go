package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/state"
)

const (
	minCardWidth  = 20
	minCardHeight = 4
)

var (
	statusColors = map[string]lipgloss.Color{
		state.StatusActive:    lipgloss.Color("42"),
		state.StatusIdle:      lipgloss.Color("245"),
		state.StatusCompleted: lipgloss.Color("34"),
		state.StatusError:     lipgloss.Color("196"),
	}

	defaultStatusColor = lipgloss.Color("245")
)

type CardModel struct {
	thread   *state.ThreadState
	selected bool
	width    int
	height   int
}

func NewCardModel(thread *state.ThreadState, selected bool) CardModel {
	return CardModel{
		thread:   thread,
		selected: selected,
		width:    minCardWidth,
		height:   minCardHeight,
	}
}

func (card *CardModel) SetSize(width int, height int) {
	if width < minCardWidth {
		width = minCardWidth
	}
	if height < minCardHeight {
		height = minCardHeight
	}
	card.width = width
	card.height = height
}

func (card CardModel) View() string {
	statusColor, exists := statusColors[card.thread.Status]
	if !exists {
		statusColor = defaultStatusColor
	}

	statusLine := lipgloss.NewStyle().
		Foreground(statusColor).
		Render(card.thread.Status)

	title := truncate(card.thread.Title, card.width-4)
	content := fmt.Sprintf("%s\n%s", title, statusLine)

	style := lipgloss.NewStyle().
		Width(card.width).
		Height(card.height).
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)

	if card.selected {
		style = style.
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("39"))
	}

	return style.Render(content)
}

func truncate(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}
