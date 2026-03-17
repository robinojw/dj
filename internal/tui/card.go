package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/state"
)

const (
	cardWidth  = 30
	cardHeight = 6
)

var (
	cardStyle = lipgloss.NewStyle().
			Width(cardWidth).
			Height(cardHeight).
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1)

	selectedCardStyle = lipgloss.NewStyle().
				Width(cardWidth).
				Height(cardHeight).
				Border(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("39")).
				Padding(0, 1)

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
}

func NewCardModel(thread *state.ThreadState, selected bool) CardModel {
	return CardModel{
		thread:   thread,
		selected: selected,
	}
}

func (card CardModel) View() string {
	statusColor, exists := statusColors[card.thread.Status]
	if !exists {
		statusColor = defaultStatusColor
	}

	statusLine := lipgloss.NewStyle().
		Foreground(statusColor).
		Render(card.thread.Status)

	title := truncate(card.thread.Title, cardWidth-4)
	content := fmt.Sprintf("%s\n%s", title, statusLine)

	style := cardStyle
	if card.selected {
		style = selectedCardStyle
	}

	return style.Render(content)
}

func truncate(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}
