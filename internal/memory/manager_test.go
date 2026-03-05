package memory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadProjectMemory(t *testing.T) {
	dir := t.TempDir()
	agentsFile := filepath.Join(dir, "AGENTS.md")
	os.WriteFile(agentsFile, []byte("# Project Rules\n- No ORM"), 0644)

	mgr := NewManager(agentsFile, filepath.Join(dir, "memory.md"))
	ctx := mgr.LoadContext()

	if !strings.Contains(ctx, "No ORM") {
		t.Error("Expected project memory to contain 'No ORM'")
	}
	if !strings.Contains(ctx, "<project_memory>") {
		t.Error("Expected project memory wrapper tag")
	}
}

func TestLoadUserMemory(t *testing.T) {
	dir := t.TempDir()
	userFile := filepath.Join(dir, "memory.md")
	os.WriteFile(userFile, []byte("Prefer tabs over spaces"), 0644)

	mgr := NewManager(filepath.Join(dir, "AGENTS.md"), userFile)
	ctx := mgr.LoadContext()

	if !strings.Contains(ctx, "Prefer tabs over spaces") {
		t.Error("Expected user memory content")
	}
	if !strings.Contains(ctx, "<user_memory>") {
		t.Error("Expected user memory wrapper tag")
	}
}

func TestMissingFilesAreEmpty(t *testing.T) {
	mgr := NewManager("/nonexistent/AGENTS.md", "/nonexistent/memory.md")
	ctx := mgr.LoadContext()

	if !strings.Contains(ctx, "<project_memory>") {
		t.Error("Should still contain tags even with missing files")
	}
}

func TestAppendUserMemory(t *testing.T) {
	dir := t.TempDir()
	userFile := filepath.Join(dir, "memory.md")
	os.WriteFile(userFile, []byte("line1"), 0644)

	mgr := NewManager(filepath.Join(dir, "AGENTS.md"), userFile)
	if err := mgr.AppendUserMemory("line2"); err != nil {
		t.Fatalf("AppendUserMemory: %v", err)
	}

	data, _ := os.ReadFile(userFile)
	if !strings.Contains(string(data), "line2") {
		t.Error("Expected appended content")
	}
	if !strings.Contains(string(data), "line1") {
		t.Error("Expected original content preserved")
	}
}
