package state

import "testing"

func TestStoreAddAndGet(t *testing.T) {
	store := NewThreadStore()
	store.Add("t-1", "My Thread")

	thread, exists := store.Get("t-1")
	if !exists {
		t.Fatal("expected thread to exist")
	}
	if thread.Title != "My Thread" {
		t.Errorf("expected My Thread, got %s", thread.Title)
	}
}

func TestStoreGetMissing(t *testing.T) {
	store := NewThreadStore()
	_, exists := store.Get("missing")
	if exists {
		t.Error("expected thread to not exist")
	}
}

func TestStoreDelete(t *testing.T) {
	store := NewThreadStore()
	store.Add("t-1", "Test")
	store.Delete("t-1")

	_, exists := store.Get("t-1")
	if exists {
		t.Error("expected thread to be deleted")
	}
}

func TestStoreAll(t *testing.T) {
	store := NewThreadStore()
	store.Add("t-1", "First")
	store.Add("t-2", "Second")

	all := store.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 threads, got %d", len(all))
	}
}

func TestStoreUpdateStatus(t *testing.T) {
	store := NewThreadStore()
	store.Add("t-1", "Test")
	store.UpdateStatus("t-1", StatusActive, "Running")

	thread, _ := store.Get("t-1")
	if thread.Status != StatusActive {
		t.Errorf("expected active, got %s", thread.Status)
	}
	if thread.Title != "Running" {
		t.Errorf("expected Running, got %s", thread.Title)
	}
}

func TestStoreUpdateStatusMissing(t *testing.T) {
	store := NewThreadStore()
	store.UpdateStatus("missing", StatusActive, "Test")
}

func TestStoreIDs(t *testing.T) {
	store := NewThreadStore()
	store.Add("t-1", "First")
	store.Add("t-2", "Second")

	ids := store.IDs()
	if len(ids) != 2 {
		t.Fatalf("expected 2 ids, got %d", len(ids))
	}
}
