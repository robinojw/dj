package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestEditFile_ExactMatch(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "test.go"), []byte("func foo() {\n\treturn nil\n}\n"), 0644)

	handler := EditFileHandler(dir)
	result, err := handler(context.Background(), map[string]any{
		"file_path":  "test.go",
		"old_string": "return nil",
		"new_string": "return 42",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == "" {
		t.Error("expected non-empty result")
	}

	data, _ := os.ReadFile(filepath.Join(dir, "test.go"))
	if got := string(data); got != "func foo() {\n\treturn 42\n}\n" {
		t.Errorf("file content = %q", got)
	}
}

func TestEditFile_TrimmedLineMatch(t *testing.T) {
	dir := t.TempDir()
	// File has tabs, but search string uses spaces
	os.WriteFile(filepath.Join(dir, "test.go"), []byte("func foo() {\n\treturn nil\n}\n"), 0644)

	handler := EditFileHandler(dir)
	_, err := handler(context.Background(), map[string]any{
		"file_path":  "test.go",
		"old_string": "  return nil  ", // extra whitespace
		"new_string": "\treturn 42",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "test.go"))
	content := string(data)
	if content != "func foo() {\n\treturn 42\n}\n" {
		t.Errorf("file content after trimmed match = %q", content)
	}
}

func TestEditFile_NormalizedWhitespaceMatch(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hello   world   foo"), 0644)

	handler := EditFileHandler(dir)
	_, err := handler(context.Background(), map[string]any{
		"file_path":  "test.txt",
		"old_string": "hello world foo",
		"new_string": "replaced",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "test.txt"))
	if string(data) != "replaced" {
		t.Errorf("file content = %q, want %q", data, "replaced")
	}
}

func TestEditFile_NotFound(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hello world"), 0644)

	handler := EditFileHandler(dir)
	_, err := handler(context.Background(), map[string]any{
		"file_path":  "test.txt",
		"old_string": "nonexistent string that does not appear",
		"new_string": "replacement",
	})
	if err == nil {
		t.Error("expected error for no match, got nil")
	}
}

func TestEditFile_PathTraversal(t *testing.T) {
	dir := t.TempDir()
	handler := EditFileHandler(dir)

	_, err := handler(context.Background(), map[string]any{
		"file_path":  "../../../etc/passwd",
		"old_string": "root",
		"new_string": "evil",
	})
	if err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestEditFile_DeleteContent(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("keep this\nremove this\nkeep this too"), 0644)

	handler := EditFileHandler(dir)
	_, err := handler(context.Background(), map[string]any{
		"file_path":  "test.txt",
		"old_string": "\nremove this",
		"new_string": "",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "test.txt"))
	if string(data) != "keep this\nkeep this too" {
		t.Errorf("file content = %q", data)
	}
}

func TestReplaceWithWhitespaceTolerance(t *testing.T) {
	tests := []struct {
		name    string
		content string
		old     string
		new     string
		want    string
		count   int
	}{
		{
			name:    "exact match",
			content: "hello world",
			old:     "world",
			new:     "earth",
			want:    "hello earth",
			count:   1,
		},
		{
			name:    "no match",
			content: "hello world",
			old:     "mars",
			new:     "earth",
			want:    "hello world",
			count:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, count := replaceWithWhitespaceTolerance(tt.content, tt.old, tt.new)
			if count != tt.count {
				t.Errorf("count = %d, want %d", count, tt.count)
			}
			if got != tt.want {
				t.Errorf("result = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeWhitespace(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello   world", "hello world"},
		{"  leading", "leading"},
		{"trailing  ", "trailing"},
		{"hello\n\tworld", "hello world"},
		{"a  b  c", "a b c"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := normalizeWhitespace(tt.input); got != tt.want {
				t.Errorf("normalizeWhitespace(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
