package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListDir_Basic(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(dir, "file2.go"), []byte("bb"), 0644)
	os.MkdirAll(filepath.Join(dir, "subdir"), 0755)

	handler := ListDirHandler(dir)
	result, err := handler(context.Background(), map[string]any{
		"path": ".",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "file1.txt") {
		t.Errorf("result missing file1.txt: %s", result)
	}
	if !strings.Contains(result, "file2.go") {
		t.Errorf("result missing file2.go: %s", result)
	}
	if !strings.Contains(result, "subdir/") {
		t.Errorf("result missing subdir/: %s", result)
	}
}

func TestListDir_DefaultPath(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hi"), 0644)

	handler := ListDirHandler(dir)
	// No "path" argument should default to "."
	result, err := handler(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "hello.txt") {
		t.Errorf("expected hello.txt in output: %s", result)
	}
}

func TestListDir_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	handler := ListDirHandler(dir)
	result, err := handler(context.Background(), map[string]any{
		"path": ".",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "(empty directory)" {
		t.Errorf("result = %q, want %q", result, "(empty directory)")
	}
}

func TestListDir_PathTraversal(t *testing.T) {
	dir := t.TempDir()
	handler := ListDirHandler(dir)

	_, err := handler(context.Background(), map[string]any{
		"path": "../../../",
	})
	if err == nil {
		t.Error("expected error for path traversal, got nil")
	}
}

func TestListDir_NotADirectory(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hi"), 0644)

	handler := ListDirHandler(dir)
	_, err := handler(context.Background(), map[string]any{
		"path": "file.txt",
	})
	if err == nil {
		t.Error("expected error for non-directory, got nil")
	}
}

func TestListDir_ShowsFileSizes(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("12345"), 0644)

	handler := ListDirHandler(dir)
	result, err := handler(context.Background(), map[string]any{
		"path": ".",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "5 bytes") {
		t.Errorf("expected file size in output: %s", result)
	}
}
