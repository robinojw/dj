package tui

import (
	"strings"
	"testing"

	"github.com/robinojw/dj/internal/state"
)

func TestCanvasNavigation(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "First")
	store.Add("t-2", "Second")
	store.Add("t-3", "Third")

	canvas := NewCanvasModel(store)

	if canvas.SelectedIndex() != 0 {
		t.Errorf("expected initial index 0, got %d", canvas.SelectedIndex())
	}

	canvas.MoveRight()
	if canvas.SelectedIndex() != 1 {
		t.Errorf("expected index 1 after right, got %d", canvas.SelectedIndex())
	}

	canvas.MoveLeft()
	if canvas.SelectedIndex() != 0 {
		t.Errorf("expected index 0 after left, got %d", canvas.SelectedIndex())
	}
}

func TestCanvasNavigationBounds(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "First")
	store.Add("t-2", "Second")

	canvas := NewCanvasModel(store)

	canvas.MoveLeft()
	if canvas.SelectedIndex() != 0 {
		t.Errorf("expected clamped at 0, got %d", canvas.SelectedIndex())
	}

	canvas.MoveRight()
	canvas.MoveRight()
	if canvas.SelectedIndex() != 1 {
		t.Errorf("expected clamped at 1, got %d", canvas.SelectedIndex())
	}
}

func TestCanvasSelectedThreadID(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "First")
	store.Add("t-2", "Second")

	canvas := NewCanvasModel(store)
	canvas.MoveRight()

	id := canvas.SelectedThreadID()
	if id != "t-2" {
		t.Errorf("expected t-2, got %s", id)
	}
}

func TestCanvasEmptyStore(t *testing.T) {
	store := state.NewThreadStore()
	canvas := NewCanvasModel(store)

	if canvas.SelectedThreadID() != "" {
		t.Errorf("expected empty ID for empty canvas")
	}

	canvas.MoveRight()
	if canvas.SelectedIndex() != 0 {
		t.Errorf("expected 0 for empty canvas")
	}
}

func TestCanvasViewWithDimensions(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "First")
	store.Add("t-2", "Second")
	store.Add("t-3", "Third")

	canvas := NewCanvasModel(store)
	canvas.SetDimensions(120, 30)
	output := canvas.View()

	if !strings.Contains(output, "First") {
		t.Errorf("expected First in output:\n%s", output)
	}
}
