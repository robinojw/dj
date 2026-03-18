package tui

import (
	"strings"
	"testing"
)

func TestHelpRender(t *testing.T) {
	help := NewHelpModel()
	output := help.View()

	expectedBindings := []string{"←/→", "↑/↓", "Enter", "Esc", "Ctrl+B", "?", "Ctrl+C"}
	for _, binding := range expectedBindings {
		if !strings.Contains(output, binding) {
			t.Errorf("expected %q in help output:\n%s", binding, output)
		}
	}
}

func TestHelpContainsActions(t *testing.T) {
	help := NewHelpModel()
	output := help.View()

	expectedActions := []string{"Navigate", "Open", "session", "Back", "context menu", "help", "Quit"}
	for _, action := range expectedActions {
		if !strings.Contains(output, action) {
			t.Errorf("expected %q in help output:\n%s", action, output)
		}
	}
}
