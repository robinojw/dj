package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	statusBarColorRed = lipgloss.Color("196")

	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1)
	statusConnectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("42"))
	statusDisconnectedStyle = lipgloss.NewStyle().
				Foreground(statusBarColorRed)
	statusErrorStyle = lipgloss.NewStyle().
				Foreground(statusBarColorRed).
				Bold(true)
)

type StatusBar struct {
	connected      bool
	threadCount    int
	selectedThread string
	errorMessage   string
	width          int
	agentCount     int
	completedCount int
}

func NewStatusBar() *StatusBar {
	return &StatusBar{}
}

func (bar *StatusBar) SetConnected(connected bool) {
	bar.connected = connected
	if connected {
		bar.errorMessage = ""
	}
}

func (bar *StatusBar) SetThreadCount(count int) {
	bar.threadCount = count
}

func (bar *StatusBar) SetSelectedThread(name string) {
	bar.selectedThread = name
}

func (bar *StatusBar) SetError(msg string) {
	bar.errorMessage = msg
}

func (bar *StatusBar) SetWidth(width int) {
	bar.width = width
}

func (bar *StatusBar) SetAgentCount(total int, completed int) {
	bar.agentCount = total
	bar.completedCount = completed
}

func (bar StatusBar) View() string {
	left := bar.renderConnectionStatus()
	middle := bar.renderCounts()
	right := bar.renderSelected()

	content := left + middle + right
	style := statusBarStyle.Width(bar.width)
	return style.Render(content)
}

func (bar StatusBar) renderConnectionStatus() string {
	var left string
	if bar.connected {
		left = statusConnectedStyle.Render("● Connected")
	} else {
		left = statusDisconnectedStyle.Render("○ Disconnected")
	}

	if bar.errorMessage != "" {
		left += " " + statusErrorStyle.Render(bar.errorMessage)
	}
	return left
}

func (bar StatusBar) renderCounts() string {
	middle := ""
	if bar.threadCount > 0 {
		middle = fmt.Sprintf(" | %d threads", bar.threadCount)
	}
	if bar.agentCount > 0 {
		middle += fmt.Sprintf(" | %d agents", bar.agentCount)
	}
	return middle
}

func (bar StatusBar) renderSelected() string {
	if bar.selectedThread != "" {
		return fmt.Sprintf(" | %s", bar.selectedThread)
	}
	return ""
}
