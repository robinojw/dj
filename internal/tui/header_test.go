package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestHeaderBarRendersTitle(t *testing.T) {
	header := NewHeaderBar(80)
	output := header.View()

	if !strings.Contains(output, "DJ") {
		t.Errorf("expected title in header, got:\n%s", output)
	}
}

func TestHeaderBarRendersShortcuts(t *testing.T) {
	header := NewHeaderBar(80)
	output := header.View()

	if !strings.Contains(output, "n: new") {
		t.Errorf("expected shortcut hints in header, got:\n%s", output)
	}
}

func TestHeaderBarFitsWidth(t *testing.T) {
	header := NewHeaderBar(120)
	output := header.View()

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if lipgloss.Width(line) > 120 {
			t.Errorf("header exceeds width 120: len=%d", lipgloss.Width(line))
		}
	}
}
