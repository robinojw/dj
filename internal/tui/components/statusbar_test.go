package components

import (
	"strings"
	"testing"

	"github.com/robinojw/dj/internal/modes"
	"github.com/robinojw/dj/internal/tui/theme"
)

func TestStatusBar_View_BasicRendering(t *testing.T) {
	sb := NewStatusBar(theme.DefaultTheme())
	sb.Width = 120
	sb.InputTokens = 5000
	sb.OutputTokens = 1200
	sb.CumulativeCost = 0.0345
	sb.Model = "gpt-4o"

	v := sb.View()
	if v == "" {
		t.Error("View() returned empty string")
	}
}

func TestStatusBar_View_WithMCPBadges(t *testing.T) {
	sb := NewStatusBar(theme.DefaultTheme())
	sb.Width = 120
	sb.ActiveMCPs = []string{"filesystem", "github"}

	v := sb.View()
	if v == "" {
		t.Fatal("View() returned empty string")
	}
	if !strings.Contains(v, "filesystem") {
		t.Error("expected MCP badge for filesystem")
	}
	if !strings.Contains(v, "github") {
		t.Error("expected MCP badge for github")
	}
}

func TestStatusBar_View_WithLSP(t *testing.T) {
	sb := NewStatusBar(theme.DefaultTheme())
	sb.Width = 120
	sb.LSPServer = "gopls"

	v := sb.View()
	if !strings.Contains(v, "gopls") {
		t.Error("expected LSP server name in view")
	}
}

func TestStatusBar_View_Compacting(t *testing.T) {
	sb := NewStatusBar(theme.DefaultTheme())
	sb.Width = 120
	sb.Compacting = true

	v := sb.View()
	if !strings.Contains(v, "Compacting") {
		t.Error("expected compacting indicator")
	}
}

func TestStatusBar_ContextPercentageCapped(t *testing.T) {
	sb := NewStatusBar(theme.DefaultTheme())
	sb.Width = 120
	// Set tokens way above 400k context window.
	sb.InputTokens = 800_000

	// The view should still render without issue (capped at 100%).
	v := sb.View()
	if v == "" {
		t.Error("View() returned empty string with excessive tokens")
	}
	// Should show 100.0% not 200.0%.
	if strings.Contains(v, "200.0") {
		t.Error("context percentage should be capped at 100%")
	}
}

func TestStatusBar_ModeBadges(t *testing.T) {
	tests := []struct {
		mode modes.ExecutionMode
		want string
	}{
		{modes.ModeConfirm, "CONFIRM"},
		{modes.ModePlan, "PLAN"},
		{modes.ModeTurbo, "TURBO"},
	}

	for _, tt := range tests {
		t.Run(tt.mode.String(), func(t *testing.T) {
			sb := NewStatusBar(theme.DefaultTheme())
			sb.Width = 120
			sb.Mode = tt.mode

			v := sb.View()
			if !strings.Contains(v, tt.want) {
				t.Errorf("expected mode label %q in view", tt.want)
			}
		})
	}
}

func TestStatusBar_EmptyState(t *testing.T) {
	sb := NewStatusBar(theme.DefaultTheme())
	sb.Width = 80

	// Should render without panic even with zero values.
	v := sb.View()
	if v == "" {
		t.Error("View() returned empty string in empty state")
	}
}

func TestStatusBar_NoModel(t *testing.T) {
	sb := NewStatusBar(theme.DefaultTheme())
	sb.Width = 120
	sb.Model = ""

	// Should not include a model badge.
	v := sb.View()
	if v == "" {
		t.Error("View() returned empty string without model")
	}
}
