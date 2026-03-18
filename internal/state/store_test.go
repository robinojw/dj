package state

import "testing"

const (
	storeTestThreadID    = "t-1"
	storeTestSecondID    = "t-2"
	storeTestMissingID   = "missing"
	storeTestMyThread    = "My Thread"
	storeTestTitle       = "Test"
	storeTestFirstTitle  = "First"
	storeTestSecondTitle = "Second"
	storeTestRunning     = "Running"
	storeTestActivity    = "Running: git status"
	storeTestExpectedTwo = 2
)

func TestStoreAddAndGet(testing *testing.T) {
	store := NewThreadStore()
	store.Add(storeTestThreadID, storeTestMyThread)

	thread, exists := store.Get(storeTestThreadID)
	if !exists {
		testing.Fatal("expected thread to exist")
	}
	if thread.Title != storeTestMyThread {
		testing.Errorf("expected My Thread, got %s", thread.Title)
	}
}

func TestStoreGetMissing(testing *testing.T) {
	store := NewThreadStore()
	_, exists := store.Get(storeTestMissingID)
	if exists {
		testing.Error("expected thread to not exist")
	}
}

func TestStoreDelete(testing *testing.T) {
	store := NewThreadStore()
	store.Add(storeTestThreadID, storeTestTitle)
	store.Delete(storeTestThreadID)

	_, exists := store.Get(storeTestThreadID)
	if exists {
		testing.Error("expected thread to be deleted")
	}
}

func TestStoreAll(testing *testing.T) {
	store := NewThreadStore()
	store.Add(storeTestThreadID, storeTestFirstTitle)
	store.Add(storeTestSecondID, storeTestSecondTitle)

	all := store.All()
	if len(all) != storeTestExpectedTwo {
		testing.Fatalf("expected 2 threads, got %d", len(all))
	}
}

func TestStoreUpdateStatus(testing *testing.T) {
	store := NewThreadStore()
	store.Add(storeTestThreadID, storeTestTitle)
	store.UpdateStatus(storeTestThreadID, StatusActive, storeTestRunning)

	thread, _ := store.Get(storeTestThreadID)
	if thread.Status != StatusActive {
		testing.Errorf("expected active, got %s", thread.Status)
	}
	if thread.Title != storeTestRunning {
		testing.Errorf("expected Running, got %s", thread.Title)
	}
}

func TestStoreUpdateStatusMissing(testing *testing.T) {
	store := NewThreadStore()
	store.UpdateStatus(storeTestMissingID, StatusActive, storeTestTitle)
}

func TestStoreIDs(testing *testing.T) {
	store := NewThreadStore()
	store.Add(storeTestThreadID, storeTestFirstTitle)
	store.Add(storeTestSecondID, storeTestSecondTitle)

	ids := store.IDs()
	if len(ids) != storeTestExpectedTwo {
		testing.Fatalf("expected 2 ids, got %d", len(ids))
	}
}

func TestStoreUpdateActivity(testing *testing.T) {
	store := NewThreadStore()
	store.Add(storeTestThreadID, storeTestTitle)
	store.UpdateActivity(storeTestThreadID, storeTestActivity)

	thread, _ := store.Get(storeTestThreadID)
	if thread.Activity != storeTestActivity {
		testing.Errorf("expected %s, got %s", storeTestActivity, thread.Activity)
	}
}

func TestStoreUpdateActivityMissing(testing *testing.T) {
	store := NewThreadStore()
	store.UpdateActivity(storeTestMissingID, storeTestActivity)
}
