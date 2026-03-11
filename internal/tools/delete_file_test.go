package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDeleteFile_Basic(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "doomed.txt"), []byte("goodbye"), 0644)

	handler := DeleteFileHandler(dir)
	result, err := handler(context.Background(), map[string]any{
		"file_path": "doomed.txt",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == "" {
		t.Error("expected non-empty result")
	}

	if _, err := os.Stat(filepath.Join(dir, "doomed.txt")); !os.IsNotExist(err) {
		t.Error("file should have been deleted")
	}
}

func TestDeleteFile_NonExistent(t *testing.T) {
	dir := t.TempDir()
	handler := DeleteFileHandler(dir)

	_, err := handler(context.Background(), map[string]any{
		"file_path": "nonexistent.txt",
	})
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestDeleteFile_Directory(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "subdir"), 0755)

	handler := DeleteFileHandler(dir)
	_, err := handler(context.Background(), map[string]any{
		"file_path": "subdir",
	})
	if err == nil {
		t.Error("expected error when trying to delete directory")
	}
}

func TestDeleteFile_PathTraversal(t *testing.T) {
	dir := t.TempDir()
	handler := DeleteFileHandler(dir)

	_, err := handler(context.Background(), map[string]any{
		"file_path": "../../../tmp/evil.txt",
	})
	if err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestDeleteFile_MissingArg(t *testing.T) {
	handler := DeleteFileHandler(t.TempDir())
	_, err := handler(context.Background(), map[string]any{})
	if err == nil {
		t.Error("expected error for missing file_path")
	}
}
