package tui

import (
	"strings"

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
	colorGreen  = lipgloss.Color("42")
	colorRed    = lipgloss.Color("196")
	colorGray   = lipgloss.Color("245")
	colorIdle   = colorGray
)

var (
	statusColors = map[string]lipgloss.Color{
		state.StatusActive:    colorGreen,
		state.StatusIdle:      colorIdle,
		state.StatusCompleted: lipgloss.Color("34"),
		state.StatusError:     colorRed,
	}

	defaultStatusColor = colorIdle
)

var (
	PersonaColorArchitect     = lipgloss.Color("33")
	PersonaColorTest          = colorGreen
	PersonaColorSecurity      = colorRed
	PersonaColorReviewer      = lipgloss.Color("226")
	PersonaColorPerformance   = lipgloss.Color("44")
	PersonaColorDesign        = lipgloss.Color("201")
	PersonaColorDevOps        = lipgloss.Color("208")
	PersonaColorDocs          = lipgloss.Color("252")
	PersonaColorAPI           = lipgloss.Color("75")
	PersonaColorData          = lipgloss.Color("178")
	PersonaColorAccessibility = lipgloss.Color("141")
	defaultPersonaColor       = colorGray
)

var personaColors = map[string]lipgloss.Color{
	"architect":     PersonaColorArchitect,
	"test":          PersonaColorTest,
	"security":      PersonaColorSecurity,
	"reviewer":      PersonaColorReviewer,
	"performance":   PersonaColorPerformance,
	"design":        PersonaColorDesign,
	"devops":        PersonaColorDevOps,
	"docs":          PersonaColorDocs,
	"api":           PersonaColorAPI,
	"data":          PersonaColorData,
	"accessibility": PersonaColorAccessibility,
}

func PersonaColor(personaID string) lipgloss.Color {
	color, exists := personaColors[personaID]
	if !exists {
		return defaultPersonaColor
	}
	return color
}

const pinnedIndicator = " ✓"
const subAgentPrefix = "↳ "
const roleIndent = "  "

type CardModel struct {
	thread       *state.ThreadState
	selected     bool
	pinned       bool
	orchestrator bool
	personaBadge string
	width        int
	height       int
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

func (card *CardModel) SetPersonaBadge(badge string) {
	card.personaBadge = badge
}

func (card *CardModel) SetOrchestrator(isOrchestrator bool) {
	card.orchestrator = isOrchestrator
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
	lines := []string{title}

	hasBadge := card.personaBadge != ""
	if hasBadge {
		badgeColor := PersonaColor(strings.ToLower(card.personaBadge))
		badgeLine := lipgloss.NewStyle().
			Foreground(badgeColor).
			Bold(true).
			Render(card.personaBadge)
		lines = append(lines, badgeLine)
	}

	isSubAgent := card.thread.ParentID != ""
	hasRole := isSubAgent && card.thread.AgentRole != ""
	if hasRole {
		roleLine := lipgloss.NewStyle().
			Foreground(colorIdle).
			Render(roleIndent + card.thread.AgentRole)
		lines = append(lines, roleLine)
	}

	lines = append(lines, statusLine)
	return strings.Join(lines, "\n")
}

func (card CardModel) buildBorderStyle() lipgloss.Style {
	style := lipgloss.NewStyle().
		Width(card.width).
		Height(card.height).
		Padding(0, 1)

	if card.orchestrator {
		return style.
			Border(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("214"))
	}

	if card.selected {
		return style.
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("39"))
	}

	return style.Border(lipgloss.RoundedBorder())
}

func truncate(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-truncateEllipsisLen] + "..."
}
