package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1)
	statusConnectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("42"))
	statusDisconnectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196"))
	statusErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true)
)

// StatusBar displays connection state and context info.
type StatusBar struct {
	connected      bool
	threadCount    int
	selectedThread string
	errorMessage   string
	width          int
}

// NewStatusBar creates a status bar.
func NewStatusBar() *StatusBar {
	return &StatusBar{}
}

// SetConnected updates the connection state.
func (s *StatusBar) SetConnected(connected bool) {
	s.connected = connected
	if connected {
		s.errorMessage = ""
	}
}

// SetThreadCount updates the thread count display.
func (s *StatusBar) SetThreadCount(count int) {
	s.threadCount = count
}

// SetSelectedThread updates the selected thread name.
func (s *StatusBar) SetSelectedThread(name string) {
	s.selectedThread = name
}

// SetError sets an error message.
func (s *StatusBar) SetError(msg string) {
	s.errorMessage = msg
}

// SetWidth sets the status bar width.
func (s *StatusBar) SetWidth(width int) {
	s.width = width
}

// View renders the status bar.
func (s StatusBar) View() string {
	var left string
	if s.connected {
		left = statusConnectedStyle.Render("● Connected")
	} else {
		left = statusDisconnectedStyle.Render("○ Disconnected")
	}

	if s.errorMessage != "" {
		left += " " + statusErrorStyle.Render(s.errorMessage)
	}

	middle := ""
	if s.threadCount > 0 {
		middle = fmt.Sprintf(" | %d threads", s.threadCount)
	}

	right := ""
	if s.selectedThread != "" {
		right = fmt.Sprintf(" | %s", s.selectedThread)
	}

	content := left + middle + right
	style := statusBarStyle.Width(s.width)
	return style.Render(content)
}
