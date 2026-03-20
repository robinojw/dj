package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	inputBarCursor   = "█"
	inputBarColorBg  = "236"
	inputBarColorFg  = "252"
	inputBarColorAcc = "39"
)

var (
	inputBarStyle = lipgloss.NewStyle().
		Background(lipgloss.Color(inputBarColorBg)).
		Foreground(lipgloss.Color(inputBarColorFg)).
		Padding(0, 1)
	inputBarPromptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(inputBarColorAcc)).
		Bold(true)
)

type InputBarModel struct {
	prompt string
	value  strings.Builder
}

func NewInputBarModel(prompt string) InputBarModel {
	return InputBarModel{prompt: prompt}
}

func (bar *InputBarModel) InsertRune(r rune) {
	bar.value.WriteRune(r)
}

func (bar *InputBarModel) DeleteRune() {
	current := bar.value.String()
	if len(current) == 0 {
		return
	}
	runes := []rune(current)
	bar.value.Reset()
	bar.value.WriteString(string(runes[:len(runes)-1]))
}

func (bar *InputBarModel) Value() string {
	return bar.value.String()
}

func (bar *InputBarModel) Reset() {
	bar.value.Reset()
}

func (bar InputBarModel) View() string {
	prompt := inputBarPromptStyle.Render(bar.prompt)
	text := bar.value.String() + inputBarCursor
	return inputBarStyle.Render(prompt + text)
}

func (bar InputBarModel) ViewWithWidth(width int) string {
	prompt := inputBarPromptStyle.Render(bar.prompt)
	text := bar.value.String() + inputBarCursor
	style := inputBarStyle.Width(width)
	return style.Render(prompt + text)
}
