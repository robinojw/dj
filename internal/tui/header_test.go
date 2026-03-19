package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

const (
	headerTestWidth    = 80
	headerTestWidthLg  = 120
	headerTestMaxWidth = 120
)

func TestHeaderBarRendersTitle(testing *testing.T) {
	header := NewHeaderBar(headerTestWidth)
	output := header.View()

	if !strings.Contains(output, "DJ") {
		testing.Errorf("expected title in header, got:\n%s", output)
	}
}

func TestHeaderBarRendersShortcuts(testing *testing.T) {
	header := NewHeaderBar(headerTestWidth)
	output := header.View()

	if !strings.Contains(output, "n: new") {
		testing.Errorf("expected shortcut hints in header, got:\n%s", output)
	}
}

func TestHeaderBarFitsWidth(testing *testing.T) {
	header := NewHeaderBar(headerTestWidthLg)
	output := header.View()

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if lipgloss.Width(line) > headerTestMaxWidth {
			testing.Errorf("header exceeds width %d: len=%d", headerTestMaxWidth, lipgloss.Width(line))
		}
	}
}

func TestHeaderSwarmHints(testing *testing.T) {
	header := NewHeaderBar(headerTestWidth)
	header.SetSwarmActive(true)
	view := header.View()
	if !strings.Contains(view, "p: persona") {
		testing.Error("expected persona hint when swarm is active")
	}
}
