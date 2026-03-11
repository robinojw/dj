package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadFile_Basic(t *testing.T) {
	dir := t.TempDir()
	content := "line1\nline2\nline3\n"
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte(content), 0644)

	handler := ReadFileHandler(dir)
	result, err := handler(context.Background(), map[string]any{
		"file_path": "test.txt",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "1\tline1") {
		t.Errorf("expected line-numbered output, got: %s", result)
	}
	if !strings.Contains(result, "2\tline2") {
		t.Errorf("expected line 2, got: %s", result)
	}
}

func TestReadFile_OffsetAndLimit(t *testing.T) {
	dir := t.TempDir()
	content := "a\nb\nc\nd\ne\n"
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte(content), 0644)

	handler := ReadFileHandler(dir)
	result, err := handler(context.Background(), map[string]any{
		"file_path": "test.txt",
		"offset":    float64(1), // JSON numbers come as float64
		"limit":     float64(2),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should show lines 2-3 (offset=1, limit=2)
	if !strings.Contains(result, "2\tb") {
		t.Errorf("expected line 2 'b', got: %s", result)
	}
	if !strings.Contains(result, "3\tc") {
		t.Errorf("expected line 3 'c', got: %s", result)
	}
	if strings.Contains(result, "1\ta") {
		t.Errorf("should not contain line 1")
	}
}

func TestReadFile_PathTraversal(t *testing.T) {
	dir := t.TempDir()

	handler := ReadFileHandler(dir)

	tests := []struct {
		name string
		path string
	}{
		{"dot-dot relative", "../../../etc/passwd"},
		{"absolute outside", "/etc/passwd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler(context.Background(), map[string]any{
				"file_path": tt.path,
			})
			if err == nil {
				t.Error("expected error for path traversal, got nil")
			}
		})
	}
}

func TestReadFile_MissingArg(t *testing.T) {
	handler := ReadFileHandler(t.TempDir())
	_, err := handler(context.Background(), map[string]any{})
	if err == nil {
		t.Error("expected error for missing file_path")
	}
}

func TestReadFile_Directory(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "subdir"), 0755)

	handler := ReadFileHandler(dir)
	_, err := handler(context.Background(), map[string]any{
		"file_path": "subdir",
	})
	if err == nil {
		t.Error("expected error for directory, got nil")
	}
}

func TestSafePath(t *testing.T) {
	root := "/workspace"

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"relative ok", "src/main.go", false},
		{"relative subdir", "src/../src/main.go", false},
		{"traversal", "../etc/passwd", true},
		{"absolute inside", "/workspace/src/main.go", false},
		{"absolute outside", "/etc/passwd", true},
		{"root itself", ".", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := safePath(root, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("safePath(%q, %q) error = %v, wantErr %v", root, tt.path, err, tt.wantErr)
			}
		})
	}
}
