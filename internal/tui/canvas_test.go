package tui

import (
	"strings"
	"testing"

	"github.com/robinojw/dj/internal/state"
)

const (
	canvasTestID1        = "t-1"
	canvasTestID2        = "t-2"
	canvasTestID3        = "t-3"
	canvasTestFirst      = "First"
	canvasTestSecond     = "Second"
	canvasTestThird      = "Third"
	canvasTestWidth      = 120
	canvasTestHeight     = 30
	canvasTestRootID     = "root"
	canvasTestChild1ID   = "child-1"
	canvasTestChild2ID   = "child-2"
	canvasTestTitleRoot   = "Root"
	canvasTestTitleChild1 = "Child 1"
	canvasTestTitleChild2 = "Child 2"
	canvasTestTitleOnly  = "Only"
)

func TestCanvasNavigation(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(canvasTestID1, canvasTestFirst)
	store.Add(canvasTestID2, canvasTestSecond)
	store.Add(canvasTestID3, canvasTestThird)

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
	store.Add(canvasTestID1, canvasTestFirst)
	store.Add(canvasTestID2, canvasTestSecond)

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
	store.Add(canvasTestID1, canvasTestFirst)
	store.Add(canvasTestID2, canvasTestSecond)

	canvas := NewCanvasModel(store)
	canvas.MoveRight()

	id := canvas.SelectedThreadID()
	if id != canvasTestID2 {
		test.Errorf("expected %s, got %s", canvasTestID2, id)
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
	store.Add(canvasTestID1, canvasTestFirst)
	store.Add(canvasTestID2, canvasTestSecond)

	canvas := NewCanvasModel(store)
	canvas.MoveRight()

	if canvas.SelectedIndex() != 1 {
		test.Fatalf("expected index 1, got %d", canvas.SelectedIndex())
	}

	store.Delete(canvasTestID2)
	canvas.ClampSelected()

	if canvas.SelectedIndex() != 0 {
		test.Errorf("expected clamped to 0, got %d", canvas.SelectedIndex())
	}
}

func TestCanvasClampSelectedEmptyStore(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(canvasTestID1, canvasTestTitleOnly)

	canvas := NewCanvasModel(store)
	store.Delete(canvasTestID1)
	canvas.ClampSelected()

	if canvas.SelectedIndex() != 0 {
		test.Errorf("expected 0 for empty store, got %d", canvas.SelectedIndex())
	}
}

func TestCanvasViewWithDimensions(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(canvasTestID1, canvasTestFirst)
	store.Add(canvasTestID2, canvasTestSecond)
	store.Add(canvasTestID3, canvasTestThird)

	canvas := NewCanvasModel(store)
	canvas.SetDimensions(canvasTestWidth, canvasTestHeight)
	output := canvas.View()

	if !strings.Contains(output, canvasTestFirst) {
		test.Errorf("expected First in output:\n%s", output)
	}
}

func TestCanvasTreeOrder(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(canvasTestRootID, canvasTestTitleRoot)
	store.AddWithParent(canvasTestChild1ID, canvasTestTitleChild1, canvasTestRootID)
	store.AddWithParent(canvasTestChild2ID, canvasTestTitleChild2, canvasTestRootID)

	canvas := NewCanvasModel(store)
	canvas.SetDimensions(canvasTestWidth, canvasTestHeight)

	view := canvas.View()
	rootIndex := strings.Index(view, canvasTestTitleRoot)
	child1Index := strings.Index(view, canvasTestTitleChild1)
	child2Index := strings.Index(view, canvasTestTitleChild2)

	allVisible := rootIndex != -1 && child1Index != -1 && child2Index != -1
	if !allVisible {
		test.Fatal("expected all threads to appear in view")
	}
	if rootIndex > child1Index {
		test.Error("root should appear before child-1")
	}
	if child1Index > child2Index {
		test.Error("child-1 should appear before child-2")
	}
}
