package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/modes"
	"github.com/robinojw/dj/internal/tui/theme"
)

func TestPermissionModal_ShowHide(t *testing.T) {
	modal := NewPermissionModal(theme.DefaultTheme())

	if modal.Visible() {
		t.Error("Modal should be hidden initially")
	}

	req := modes.PermissionRequest{
		ID:     "test-1",
		Tool:   "write_file",
		Args:   map[string]any{"path": "test.go"},
		RespCh: make(chan modes.PermissionResp, 1),
	}

	modal, _ = modal.Update(req)

	if !modal.Visible() {
		t.Error("Modal should be visible after request")
	}
}

func TestPermissionModal_ScopeCycle(t *testing.T) {
	modal := NewPermissionModal(theme.DefaultTheme())

	req := modes.PermissionRequest{
		ID:     "test-2",
		Tool:   "bash",
		RespCh: make(chan modes.PermissionResp, 1),
	}
	modal, _ = modal.Update(req)

	// Initial scope is Once
	if modal.scope != modes.RememberOnce {
		t.Errorf("Expected RememberOnce, got %v", modal.scope)
	}

	// Tab cycles to Session
	modal, _ = modal.Update(tea.KeyMsg{Type: tea.KeyTab})
	if modal.scope != modes.RememberSession {
		t.Errorf("Expected RememberSession, got %v", modal.scope)
	}

	// Tab cycles to Always
	modal, _ = modal.Update(tea.KeyMsg{Type: tea.KeyTab})
	if modal.scope != modes.RememberAlways {
		t.Errorf("Expected RememberAlways, got %v", modal.scope)
	}

	// Tab cycles back to Once
	modal, _ = modal.Update(tea.KeyMsg{Type: tea.KeyTab})
	if modal.scope != modes.RememberOnce {
		t.Errorf("Expected RememberOnce, got %v", modal.scope)
	}
}
