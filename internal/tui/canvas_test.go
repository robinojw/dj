package tui

import (
	"strings"
	"testing"

	"github.com/robinojw/dj/internal/state"
)

const (
	testThreadID1    = "t-1"
	testThreadID2    = "t-2"
	testThreadID3    = "t-3"
	testThreadTitle1 = "First"
	testThreadTitle2 = "Second"
	testThreadTitle3 = "Third"
	testCanvasWidth  = 120
	testCanvasHeight = 30
)

func TestCanvasNavigation(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testThreadTitle1)
	store.Add(testThreadID2, testThreadTitle2)
	store.Add(testThreadID3, testThreadTitle3)

	canvas := NewCanvasModel(store)

	if canvas.SelectedIndex() != 0 {
		test.Errorf("expected initial index 0, got %d", canvas.SelectedIndex())
	}

	canvas.MoveRight()
	if canvas.SelectedIndex() != 1 {
		test.Errorf("expected index 1 after right, got %d", canvas.SelectedIndex())
	}

	canvas.MoveLeft()
	if canvas.SelectedIndex() != 0 {
		test.Errorf("expected index 0 after left, got %d", canvas.SelectedIndex())
	}
}

func TestCanvasNavigationBounds(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testThreadTitle1)
	store.Add(testThreadID2, testThreadTitle2)

	canvas := NewCanvasModel(store)

	canvas.MoveLeft()
	if canvas.SelectedIndex() != 0 {
		test.Errorf("expected clamped at 0, got %d", canvas.SelectedIndex())
	}

	canvas.MoveRight()
	canvas.MoveRight()
	if canvas.SelectedIndex() != 1 {
		test.Errorf("expected clamped at 1, got %d", canvas.SelectedIndex())
	}
}

func TestCanvasSelectedThreadID(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testThreadTitle1)
	store.Add(testThreadID2, testThreadTitle2)

	canvas := NewCanvasModel(store)
	canvas.MoveRight()

	id := canvas.SelectedThreadID()
	if id != testThreadID2 {
		test.Errorf("expected %s, got %s", testThreadID2, id)
	}
}

func TestCanvasEmptyStore(test *testing.T) {
	store := state.NewThreadStore()
	canvas := NewCanvasModel(store)

	if canvas.SelectedThreadID() != "" {
		test.Errorf("expected empty ID for empty canvas")
	}

	canvas.MoveRight()
	if canvas.SelectedIndex() != 0 {
		test.Errorf("expected 0 for empty canvas")
	}
}

func TestCanvasClampSelectedAfterDeletion(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testThreadTitle1)
	store.Add(testThreadID2, testThreadTitle2)

	canvas := NewCanvasModel(store)
	canvas.MoveRight()

	if canvas.SelectedIndex() != 1 {
		test.Fatalf("expected index 1, got %d", canvas.SelectedIndex())
	}

	store.Delete(testThreadID2)
	canvas.ClampSelected()

	if canvas.SelectedIndex() != 0 {
		test.Errorf("expected clamped to 0, got %d", canvas.SelectedIndex())
	}
}

func TestCanvasClampSelectedEmptyStore(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, "Only")

	canvas := NewCanvasModel(store)
	store.Delete(testThreadID1)
	canvas.ClampSelected()

	if canvas.SelectedIndex() != 0 {
		test.Errorf("expected 0 for empty store, got %d", canvas.SelectedIndex())
	}
}

func TestCanvasViewWithDimensions(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testThreadTitle1)
	store.Add(testThreadID2, testThreadTitle2)
	store.Add(testThreadID3, testThreadTitle3)

	canvas := NewCanvasModel(store)
	canvas.SetDimensions(testCanvasWidth, testCanvasHeight)
	output := canvas.View()

	if !strings.Contains(output, testThreadTitle1) {
		test.Errorf("expected %s in output:\n%s", testThreadTitle1, output)
	}
}
