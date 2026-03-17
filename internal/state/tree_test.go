package state

import "testing"

func TestThreadStateParentID(t *testing.T) {
	thread := NewThreadState("t-child", "Child Task")
	thread.ParentID = "t-parent"

	if thread.ParentID != "t-parent" {
		t.Errorf("expected t-parent, got %s", thread.ParentID)
	}
}

func TestStoreChildren(t *testing.T) {
	store := NewThreadStore()
	store.Add("t-root", "Root")
	store.AddWithParent("t-child-1", "Child 1", "t-root")
	store.AddWithParent("t-child-2", "Child 2", "t-root")

	children := store.Children("t-root")
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}
}

func TestStoreRoots(t *testing.T) {
	store := NewThreadStore()
	store.Add("t-root-1", "Root 1")
	store.Add("t-root-2", "Root 2")
	store.AddWithParent("t-child", "Child", "t-root-1")

	roots := store.Roots()
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots, got %d", len(roots))
	}
}
