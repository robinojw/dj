package tui

import (
	"strings"
	"testing"
)

func TestHelpShowsScrollKeybinding(testing *testing.T) {
	help := NewHelpModel()
	view := help.View()
	if !strings.Contains(view, "Scroll") {
		testing.Error("expected Scroll keybinding in help")
	}
}
