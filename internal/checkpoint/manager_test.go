package checkpoint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateAndRestore(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")
	os.WriteFile(file, []byte("original"), 0644)

	mgr := NewManager(20)

	// Create checkpoint
	cp, err := mgr.Before([]string{file}, "test-response-id", "Before: write test.txt")
	if err != nil {
		t.Fatalf("Before: %v", err)
	}

	// Modify the file
	os.WriteFile(file, []byte("modified"), 0644)

	// Verify file was changed
	data, _ := os.ReadFile(file)
	if string(data) != "modified" {
		t.Fatal("File should be modified")
	}

	// Restore checkpoint
	if err := mgr.Restore(cp); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	// Verify file restored
	data, _ = os.ReadFile(file)
	if string(data) != "original" {
		t.Errorf("Expected 'original' after restore, got %q", string(data))
	}
}

func TestUndoStack(t *testing.T) {
	mgr := NewManager(3)

	for i := 0; i < 5; i++ {
		mgr.Push(Checkpoint{Description: "cp"})
	}

	if mgr.Len() != 3 {
		t.Errorf("Expected max 3 checkpoints, got %d", mgr.Len())
	}
}

func TestUndoPop(t *testing.T) {
	mgr := NewManager(10)
	mgr.Push(Checkpoint{ID: "a", Description: "first"})
	mgr.Push(Checkpoint{ID: "b", Description: "second"})

	cp := mgr.Pop()
	if cp == nil {
		t.Fatal("Expected a checkpoint")
	}
	if cp.ID != "b" {
		t.Errorf("Expected 'b', got %q", cp.ID)
	}
	if mgr.Len() != 1 {
		t.Errorf("Expected 1 remaining, got %d", mgr.Len())
	}
}

func TestUndoPopEmpty(t *testing.T) {
	mgr := NewManager(10)
	cp := mgr.Pop()
	if cp != nil {
		t.Error("Expected nil from empty stack")
	}
}

func TestNewFileRestore(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "new.txt")

	// File doesn't exist yet — checkpoint records absence
	mgr := NewManager(20)
	cp, err := mgr.Before([]string{file}, "resp-1", "Before: create new.txt")
	if err != nil {
		t.Fatalf("Before: %v", err)
	}

	// Create the file
	os.WriteFile(file, []byte("new content"), 0644)

	// Restore should delete the file
	if err := mgr.Restore(cp); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Error("File should not exist after restoring to pre-creation state")
	}
}
