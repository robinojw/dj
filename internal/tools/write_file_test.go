package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFile_CreateNew(t *testing.T) {
	dir := t.TempDir()
	handler := WriteFileHandler(dir)

	result, err := handler(context.Background(), map[string]any{
		"file_path": "new.txt",
		"content":   "hello world",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "11 bytes") {
		t.Errorf("result = %q, expected byte count", result)
	}

	data, err := os.ReadFile(filepath.Join(dir, "new.txt"))
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("file content = %q, want %q", data, "hello world")
	}
}

func TestWriteFile_CreateParentDirs(t *testing.T) {
	dir := t.TempDir()
	handler := WriteFileHandler(dir)

	_, err := handler(context.Background(), map[string]any{
		"file_path": "deep/nested/dir/file.txt",
		"content":   "content",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "deep", "nested", "dir", "file.txt")); err != nil {
		t.Errorf("file not created: %v", err)
	}
}

func TestWriteFile_BackupOnOverwrite(t *testing.T) {
	dir := t.TempDir()
	original := "original content"
	os.WriteFile(filepath.Join(dir, "existing.txt"), []byte(original), 0644)

	handler := WriteFileHandler(dir)
	_, err := handler(context.Background(), map[string]any{
		"file_path": "existing.txt",
		"content":   "new content",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that backup was created
	entries, _ := os.ReadDir(dir)
	backupFound := false
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "existing.txt.bak.") {
			backupFound = true
			data, _ := os.ReadFile(filepath.Join(dir, e.Name()))
			if string(data) != original {
				t.Errorf("backup content = %q, want %q", data, original)
			}
		}
	}
	if !backupFound {
		t.Error("no backup file created")
	}

	// Check new content
	data, _ := os.ReadFile(filepath.Join(dir, "existing.txt"))
	if string(data) != "new content" {
		t.Errorf("file content = %q, want %q", data, "new content")
	}
}

func TestWriteFile_PathTraversal(t *testing.T) {
	dir := t.TempDir()
	handler := WriteFileHandler(dir)

	_, err := handler(context.Background(), map[string]any{
		"file_path": "../../../tmp/evil.txt",
		"content":   "bad",
	})
	if err == nil {
		t.Error("expected error for path traversal, got nil")
	}
}

func TestWriteFile_MissingArgs(t *testing.T) {
	handler := WriteFileHandler(t.TempDir())

	_, err := handler(context.Background(), map[string]any{
		"content": "data",
	})
	if err == nil {
		t.Error("expected error for missing file_path")
	}

	_, err = handler(context.Background(), map[string]any{
		"file_path": "test.txt",
	})
	if err == nil {
		t.Error("expected error for missing content")
	}
}
