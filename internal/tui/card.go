package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/state"
)

const (
	minCardWidth      = 20
	minCardHeight     = 4
	cardBorderPadding = 4
	ellipsisLen       = 3
	colorGray         = lipgloss.Color("245")
)

var (
	statusColors = map[string]lipgloss.Color{
		state.StatusActive:    lipgloss.Color("42"),
		state.StatusIdle:      colorGray,
		state.StatusCompleted: lipgloss.Color("34"),
		state.StatusError:     lipgloss.Color("196"),
	}

	defaultStatusColor = colorGray
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

	secondLine := card.thread.Status
	hasActivity := card.thread.Activity != ""
	if hasActivity {
		secondLine = card.thread.Activity
	}

	styledSecondLine := lipgloss.NewStyle().
		Foreground(statusColor).
		Render(truncate(secondLine, card.width-cardBorderPadding))

	titleMaxLen := card.width - cardBorderPadding
	title := truncate(card.thread.Title, titleMaxLen)
	content := fmt.Sprintf("%s\n%s", title, styledSecondLine)

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
	return text[:maxLen-ellipsisLen] + "..."
}
