package tui

import "testing"

func TestSessionPanelPinAddsThread(t *testing.T) {
	panel := NewSessionPanelModel()
	panel.Pin("t-1")

	if len(panel.PinnedSessions()) != 1 {
		t.Fatalf("expected 1 pinned session, got %d", len(panel.PinnedSessions()))
	}
	if panel.PinnedSessions()[0] != "t-1" {
		t.Errorf("expected t-1, got %s", panel.PinnedSessions()[0])
	}
}

func TestSessionPanelPinIgnoresDuplicate(t *testing.T) {
	panel := NewSessionPanelModel()
	panel.Pin("t-1")
	panel.Pin("t-1")

	if len(panel.PinnedSessions()) != 1 {
		t.Errorf("expected 1 pinned session, got %d", len(panel.PinnedSessions()))
	}
}

func TestSessionPanelUnpin(t *testing.T) {
	panel := NewSessionPanelModel()
	panel.Pin("t-1")
	panel.Pin("t-2")
	panel.Unpin("t-1")

	if len(panel.PinnedSessions()) != 1 {
		t.Fatalf("expected 1, got %d", len(panel.PinnedSessions()))
	}
	if panel.PinnedSessions()[0] != "t-2" {
		t.Errorf("expected t-2, got %s", panel.PinnedSessions()[0])
	}
}

func TestSessionPanelUnpinClampsFocus(t *testing.T) {
	panel := NewSessionPanelModel()
	panel.Pin("t-1")
	panel.Pin("t-2")
	panel.SetActivePaneIdx(1)
	panel.Unpin("t-2")

	if panel.ActivePaneIdx() != 0 {
		t.Errorf("expected clamped to 0, got %d", panel.ActivePaneIdx())
	}
}

func TestSessionPanelCycleRight(t *testing.T) {
	panel := NewSessionPanelModel()
	panel.Pin("t-1")
	panel.Pin("t-2")
	panel.Pin("t-3")

	panel.CycleRight()
	if panel.ActivePaneIdx() != 1 {
		t.Errorf("expected 1, got %d", panel.ActivePaneIdx())
	}

	panel.CycleRight()
	panel.CycleRight()
	if panel.ActivePaneIdx() != 2 {
		t.Errorf("expected clamped to 2, got %d", panel.ActivePaneIdx())
	}
}

func TestSessionPanelCycleLeft(t *testing.T) {
	panel := NewSessionPanelModel()
	panel.Pin("t-1")
	panel.Pin("t-2")
	panel.SetActivePaneIdx(1)

	panel.CycleLeft()
	if panel.ActivePaneIdx() != 0 {
		t.Errorf("expected 0, got %d", panel.ActivePaneIdx())
	}

	panel.CycleLeft()
	if panel.ActivePaneIdx() != 0 {
		t.Errorf("expected clamped to 0, got %d", panel.ActivePaneIdx())
	}
}

func TestSessionPanelActiveThreadID(t *testing.T) {
	panel := NewSessionPanelModel()
	if panel.ActiveThreadID() != "" {
		t.Errorf("expected empty, got %s", panel.ActiveThreadID())
	}

	panel.Pin("t-1")
	panel.Pin("t-2")
	panel.SetActivePaneIdx(1)

	if panel.ActiveThreadID() != "t-2" {
		t.Errorf("expected t-2, got %s", panel.ActiveThreadID())
	}
}

func TestSessionPanelIsPinned(t *testing.T) {
	panel := NewSessionPanelModel()
	panel.Pin("t-1")

	if !panel.IsPinned("t-1") {
		t.Error("expected t-1 to be pinned")
	}
	if panel.IsPinned("t-2") {
		t.Error("expected t-2 to not be pinned")
	}
}

func TestSessionPanelSplitRatio(t *testing.T) {
	panel := NewSessionPanelModel()
	if panel.SplitRatio() != defaultSplitRatio {
		t.Errorf("expected %f, got %f", defaultSplitRatio, panel.SplitRatio())
	}
}

func TestSessionPanelSessionDimensions(t *testing.T) {
	panel := NewSessionPanelModel()
	panel.Pin("t-1")
	panel.Pin("t-2")

	width, height := panel.SessionDimensions(120, 40)
	expectedWidth := 120 / 2
	expectedHeight := 40 - dividerHeight
	if width != expectedWidth {
		t.Errorf("expected width %d, got %d", expectedWidth, width)
	}
	if height != expectedHeight {
		t.Errorf("expected height %d, got %d", expectedHeight, height)
	}
}

func TestSessionPanelSessionDimensionsEmpty(t *testing.T) {
	panel := NewSessionPanelModel()
	width, height := panel.SessionDimensions(120, 40)
	if width != 0 || height != 0 {
		t.Errorf("expected 0,0 for empty panel, got %d,%d", width, height)
	}
}

func TestSessionPanelZoomToggle(t *testing.T) {
	panel := NewSessionPanelModel()
	panel.Pin("t-1")
	panel.Pin("t-2")

	if panel.Zoomed() {
		t.Error("expected not zoomed initially")
	}

	panel.ToggleZoom()
	if !panel.Zoomed() {
		t.Error("expected zoomed after toggle")
	}

	panel.ToggleZoom()
	if panel.Zoomed() {
		t.Error("expected not zoomed after second toggle")
	}
}
