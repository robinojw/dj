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
	statusConnectingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214"))
	statusDisconnectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196"))
	statusErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true)
)

type StatusBar struct {
	connected      bool
	connecting     bool
	threadCount    int
	selectedThread string
	errorMessage   string
	width          int
}

func NewStatusBar() *StatusBar {
	return &StatusBar{}
}

func (statusBar *StatusBar) SetConnecting() {
	statusBar.connecting = true
	statusBar.connected = false
	statusBar.errorMessage = ""
}

func (statusBar *StatusBar) SetConnected(connected bool) {
	statusBar.connecting = false
	statusBar.connected = connected
	if connected {
		statusBar.errorMessage = ""
	}
}

func (statusBar *StatusBar) SetThreadCount(count int) {
	statusBar.threadCount = count
}

func (statusBar *StatusBar) SetSelectedThread(name string) {
	statusBar.selectedThread = name
}

func (statusBar StatusBar) renderConnectionState() string {
	if statusBar.connected {
		return statusConnectedStyle.Render("● Connected")
	}
	if statusBar.connecting {
		return statusConnectingStyle.Render("◌ Connecting to app-server...")
	}
	return statusDisconnectedStyle.Render("○ Disconnected — requires codex CLI (codex app-server)")
}

func (statusBar *StatusBar) SetError(message string) {
	statusBar.errorMessage = message
}

func (statusBar *StatusBar) SetWidth(width int) {
	statusBar.width = width
}

func (statusBar StatusBar) View() string {
	left := statusBar.renderConnectionState()

	if statusBar.errorMessage != "" {
		left += " " + statusErrorStyle.Render(statusBar.errorMessage)
	}

	middle := ""
	if statusBar.threadCount > 0 {
		middle = fmt.Sprintf(" | %d threads", statusBar.threadCount)
	}

	right := ""
	if statusBar.selectedThread != "" {
		right = fmt.Sprintf(" | %s", statusBar.selectedThread)
	}

	content := left + middle + right
	style := statusBarStyle.Width(statusBar.width)
	return style.Render(content)
}
