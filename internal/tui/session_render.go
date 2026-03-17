package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/state"
)

var (
	userStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)
	assistantStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)
	commandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))
	outputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true)
)

func RenderMessages(thread *state.ThreadState) string {
	hasMessages := len(thread.Messages) > 0
	hasOutput := len(thread.CommandOutput) > 0

	if !hasMessages && !hasOutput {
		return emptyStyle.Render("No messages yet. Waiting for activity...")
	}

	var sections []string

	for _, msg := range thread.Messages {
		label := formatRole(msg.Role)
		sections = append(sections, fmt.Sprintf("%s\n%s", label, msg.Content))
	}

	for execID, output := range thread.CommandOutput {
		header := commandStyle.Render(fmt.Sprintf("Command [%s]:", execID))
		body := outputStyle.Render(output)
		sections = append(sections, fmt.Sprintf("%s\n%s", header, body))
	}

	return strings.Join(sections, "\n\n")
}

func formatRole(role string) string {
	switch role {
	case "user":
		return userStyle.Render("You:")
	case "assistant":
		return assistantStyle.Render("Agent:")
	default:
		return lipgloss.NewStyle().Bold(true).Render(role + ":")
	}
}
