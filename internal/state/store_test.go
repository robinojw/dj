package state

import "testing"

const (
	storeTestThreadID1   = "t-1"
	storeTestThreadID2   = "t-2"
	storeTestParentID    = "t-0"
	storeTestMissingID   = "missing"
	storeTestTitleThread = "My Thread"
	storeTestTitleTest   = "Test"
	storeTestTitleFirst  = "First"
	storeTestTitleSecond = "Second"
	storeTestTitleRoot   = "Root"
	storeTestTitleChild  = "Child"
	storeTestTitleScout  = "Scout"
	storeTestRunning     = "Running"
	storeTestActivity    = "Running: git status"
	storeTestNickname    = "scout"
	storeTestRole        = "researcher"
	storeTestChildNotFnd = "child not found"
	storeTestExpectedTwo = 2
)

func TestStoreAddAndGet(test *testing.T) {
	store := NewThreadStore()
	store.Add(storeTestThreadID1, storeTestTitleThread)

	thread, exists := store.Get(storeTestThreadID1)
	if !exists {
		test.Fatal("expected thread to exist")
	}
	if thread.Title != storeTestTitleThread {
		test.Errorf("expected My Thread, got %s", thread.Title)
	}
}

func TestStoreGetMissing(test *testing.T) {
	store := NewThreadStore()
	_, exists := store.Get(storeTestMissingID)
	if exists {
		test.Error("expected thread to not exist")
	}
}

func TestStoreDelete(test *testing.T) {
	store := NewThreadStore()
	store.Add(storeTestThreadID1, storeTestTitleTest)
	store.Delete(storeTestThreadID1)

	_, exists := store.Get(storeTestThreadID1)
	if exists {
		test.Error("expected thread to be deleted")
	}
}

func TestStoreAll(test *testing.T) {
	store := NewThreadStore()
	store.Add(storeTestThreadID1, storeTestTitleFirst)
	store.Add(storeTestThreadID2, storeTestTitleSecond)

	all := store.All()
	if len(all) != storeTestExpectedTwo {
		test.Fatalf("expected 2 threads, got %d", len(all))
	}
}

func TestStoreUpdateStatus(test *testing.T) {
	store := NewThreadStore()
	store.Add(storeTestThreadID1, storeTestTitleTest)
	store.UpdateStatus(storeTestThreadID1, StatusActive, storeTestRunning)

	thread, _ := store.Get(storeTestThreadID1)
	if thread.Status != StatusActive {
		test.Errorf("expected active, got %s", thread.Status)
	}
	if thread.Title != storeTestRunning {
		test.Errorf("expected Running, got %s", thread.Title)
	}
}

func TestStoreUpdateStatusMissing(test *testing.T) {
	store := NewThreadStore()
	store.UpdateStatus(storeTestMissingID, StatusActive, storeTestTitleTest)
}

func TestStoreIDs(test *testing.T) {
	store := NewThreadStore()
	store.Add(storeTestThreadID1, storeTestTitleFirst)
	store.Add(storeTestThreadID2, storeTestTitleSecond)

	ids := store.IDs()
	if len(ids) != storeTestExpectedTwo {
		test.Fatalf("expected 2 ids, got %d", len(ids))
	}
}

func TestAddWithParentFields(test *testing.T) {
	store := NewThreadStore()
	store.Add(storeTestParentID, storeTestTitleRoot)
	store.AddWithParent(storeTestThreadID1, storeTestTitleChild, storeTestParentID)

	child, exists := store.Get(storeTestThreadID1)
	if !exists {
		test.Fatal(storeTestChildNotFnd)
	}
	if child.ParentID != storeTestParentID {
		test.Errorf("expected parent t-0, got %s", child.ParentID)
	}
}

func TestAddSubAgent(test *testing.T) {
	store := NewThreadStore()
	store.Add(storeTestParentID, storeTestTitleRoot)
	store.AddSubAgent(storeTestThreadID1, storeTestTitleScout, storeTestParentID, storeTestNickname, storeTestRole, 1)

	child, exists := store.Get(storeTestThreadID1)
	if !exists {
		test.Fatal(storeTestChildNotFnd)
	}
	if child.AgentNickname != storeTestNickname {
		test.Errorf("expected scout, got %s", child.AgentNickname)
	}
	if child.AgentRole != storeTestRole {
		test.Errorf("expected researcher, got %s", child.AgentRole)
	}
	if child.Depth != 1 {
		test.Errorf("expected depth 1, got %d", child.Depth)
	}
}

func TestStoreUpdateActivity(test *testing.T) {
	store := NewThreadStore()
	store.Add(storeTestThreadID1, storeTestTitleTest)
	store.UpdateActivity(storeTestThreadID1, storeTestActivity)

	thread, _ := store.Get(storeTestThreadID1)
	if thread.Activity != storeTestActivity {
		test.Errorf("expected %s, got %s", storeTestActivity, thread.Activity)
	}
}

func TestStoreUpdateActivityMissing(test *testing.T) {
	store := NewThreadStore()
	store.UpdateActivity(storeTestMissingID, storeTestActivity)
}
