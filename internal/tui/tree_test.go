package tui

import (
	"strings"
	"testing"

	"github.com/robinojw/dj/internal/state"
)

func TestTreeRender(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Root Task")
	store.AddWithParent("t-2", "Subtask A", "t-1")

	tree := NewTreeModel(store)
	output := tree.View()

	if !strings.Contains(output, "Root Task") {
		t.Errorf("expected Root Task in output:\n%s", output)
	}
	if !strings.Contains(output, "Subtask A") {
		t.Errorf("expected Subtask A in output:\n%s", output)
	}
}

func TestTreeNavigation(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "First")
	store.Add("t-2", "Second")

	tree := NewTreeModel(store)

	if tree.SelectedID() != "t-1" {
		t.Errorf("expected t-1, got %s", tree.SelectedID())
	}

	tree.MoveDown()
	if tree.SelectedID() != "t-2" {
		t.Errorf("expected t-2, got %s", tree.SelectedID())
	}

	tree.MoveUp()
	if tree.SelectedID() != "t-1" {
		t.Errorf("expected t-1, got %s", tree.SelectedID())
	}
}

func TestTreeEmpty(t *testing.T) {
	store := state.NewThreadStore()
	tree := NewTreeModel(store)

	if tree.SelectedID() != "" {
		t.Errorf("expected empty ID, got %s", tree.SelectedID())
	}

	output := tree.View()
	if !strings.Contains(output, "No threads") {
		t.Errorf("expected empty message:\n%s", output)
	}
}
