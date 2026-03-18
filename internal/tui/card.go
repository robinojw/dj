package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/state"
)

const (
	minCardWidth        = 20
	maxCardWidth        = 50
	minCardHeight       = 4
	maxCardHeight       = 12
	cardBorderPadding   = 4
	truncateEllipsisLen = 3
)

var (
	colorIdle = lipgloss.Color("245")

	statusColors = map[string]lipgloss.Color{
		state.StatusActive:    lipgloss.Color("42"),
		state.StatusIdle:      colorIdle,
		state.StatusCompleted: lipgloss.Color("34"),
		state.StatusError:     lipgloss.Color("196"),
	}

	defaultStatusColor = colorIdle
)

const pinnedIndicator = " ✓"
const subAgentPrefix = "↳ "
const roleIndent = "  "

type CardModel struct {
	thread   *state.ThreadState
	selected bool
	pinned   bool
	width    int
	height   int
}

func NewCardModel(thread *state.ThreadState, selected bool, pinned bool) CardModel {
	return CardModel{
		thread:   thread,
		selected: selected,
		pinned:   pinned,
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
	title := card.buildTitle()
	statusLine := card.buildStatusLine()
	content := card.buildContent(title, statusLine)
	style := card.buildBorderStyle()
	return style.Render(content)
}

func (card CardModel) buildTitle() string {
	titleMaxLen := card.width - cardBorderPadding
	if card.pinned {
		titleMaxLen -= len(pinnedIndicator)
	}

	title := card.thread.Title
	isSubAgent := card.thread.ParentID != ""
	if isSubAgent {
		title = subAgentPrefix + title
	}

	title = truncate(title, titleMaxLen)
	if card.pinned {
		title += pinnedIndicator
	}
	return title
}

func (card CardModel) buildStatusLine() string {
	statusColor, exists := statusColors[card.thread.Status]
	if !exists {
		statusColor = defaultStatusColor
	}

	secondLine := card.thread.Status
	hasActivity := card.thread.Activity != ""
	if hasActivity {
		secondLine = card.thread.Activity
	}

	return lipgloss.NewStyle().
		Foreground(statusColor).
		Render(truncate(secondLine, card.width-cardBorderPadding))
}

func (card CardModel) buildContent(title string, statusLine string) string {
	isSubAgent := card.thread.ParentID != ""
	hasRole := isSubAgent && card.thread.AgentRole != ""
	if hasRole {
		roleLine := lipgloss.NewStyle().
			Foreground(colorIdle).
			Render(roleIndent + card.thread.AgentRole)
		return fmt.Sprintf("%s\n%s\n%s", title, roleLine, statusLine)
	}
	return fmt.Sprintf("%s\n%s", title, statusLine)
}

func (card CardModel) buildBorderStyle() lipgloss.Style {
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
	return style
}

func truncate(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-truncateEllipsisLen] + "..."
}
