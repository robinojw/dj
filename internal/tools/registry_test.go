package tools

import (
	"context"
	"testing"
)

func TestRegistry_MutatesFiles(t *testing.T) {
	r := NewRegistry()
	noop := func(ctx context.Context, args map[string]any) (string, error) { return "", nil }

	r.Register("writer", noop, ToolAnnotations{
		Destructive:   true,
		MutatesFiles:  true,
		FilePathParam: "file_path",
	})
	r.Register("reader", noop, ToolAnnotations{ReadOnly: true})

	ann := r.Annotations("writer")
	if !ann.MutatesFiles {
		t.Error("MutatesFiles = false, want true")
	}
	if ann.FilePathParam != "file_path" {
		t.Errorf("FilePathParam = %q, want %q", ann.FilePathParam, "file_path")
	}

	ann2 := r.Annotations("reader")
	if ann2.MutatesFiles {
		t.Error("reader MutatesFiles = true, want false")
	}
	if ann2.FilePathParam != "" {
		t.Errorf("reader FilePathParam = %q, want empty", ann2.FilePathParam)
	}
}

func TestRegistry_RegisterAnnotationsOnly(t *testing.T) {
	r := NewRegistry()

	r.RegisterAnnotationsOnly("mcp_write", ToolAnnotations{
		MutatesFiles:  true,
		FilePathParam: "path",
	})

	if !r.HasAnnotations("mcp_write") {
		t.Error("HasAnnotations(mcp_write) = false, want true")
	}
	// Has() returns false — no handler registered
	if r.Has("mcp_write") {
		t.Error("Has(mcp_write) = true, want false (no handler)")
	}
	ann := r.Annotations("mcp_write")
	if !ann.MutatesFiles {
		t.Error("MutatesFiles = false, want true")
	}
	if ann.FilePathParam != "path" {
		t.Errorf("FilePathParam = %q, want %q", ann.FilePathParam, "path")
	}
}
