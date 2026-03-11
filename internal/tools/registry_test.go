package tools

import (
	"context"
	"testing"
)

func TestRegistry_RegisterAndDispatch(t *testing.T) {
	r := NewRegistry()

	called := false
	r.Register("my_tool", func(ctx context.Context, args map[string]any) (string, error) {
		called = true
		return "ok", nil
	}, ToolAnnotations{ReadOnly: true})

	result, err := r.Dispatch(context.Background(), "my_tool", nil)
	if err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}
	if result != "ok" {
		t.Errorf("Dispatch() = %q, want %q", result, "ok")
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestRegistry_DispatchUnknown(t *testing.T) {
	r := NewRegistry()

	_, err := r.Dispatch(context.Background(), "nonexistent", nil)
	if err == nil {
		t.Error("expected error for unknown tool, got nil")
	}
}

func TestRegistry_Has(t *testing.T) {
	r := NewRegistry()
	r.Register("tool_a", func(ctx context.Context, args map[string]any) (string, error) {
		return "", nil
	}, ToolAnnotations{})

	if !r.Has("tool_a") {
		t.Error("Has(tool_a) = false, want true")
	}
	if r.Has("tool_b") {
		t.Error("Has(tool_b) = true, want false")
	}
}

func TestRegistry_Annotations(t *testing.T) {
	r := NewRegistry()
	r.Register("reader", func(ctx context.Context, args map[string]any) (string, error) {
		return "", nil
	}, ToolAnnotations{ReadOnly: true, Idempotent: true})

	r.Register("writer", func(ctx context.Context, args map[string]any) (string, error) {
		return "", nil
	}, ToolAnnotations{Destructive: true})

	ann := r.Annotations("reader")
	if !ann.ReadOnly {
		t.Error("reader annotation ReadOnly = false, want true")
	}
	if !ann.Idempotent {
		t.Error("reader annotation Idempotent = false, want true")
	}
	if ann.Destructive {
		t.Error("reader annotation Destructive = true, want false")
	}

	if !r.IsDestructive("writer") {
		t.Error("IsDestructive(writer) = false, want true")
	}
	if r.IsDestructive("reader") {
		t.Error("IsDestructive(reader) = true, want false")
	}
}

func TestRegistry_Names(t *testing.T) {
	r := NewRegistry()
	noop := func(ctx context.Context, args map[string]any) (string, error) { return "", nil }

	r.Register("a", noop, ToolAnnotations{})
	r.Register("b", noop, ToolAnnotations{})
	r.Register("c", noop, ToolAnnotations{})

	names := r.Names()
	if len(names) != 3 {
		t.Fatalf("Names() returned %d names, want 3", len(names))
	}

	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}
	for _, want := range []string{"a", "b", "c"} {
		if !nameSet[want] {
			t.Errorf("Names() missing %q", want)
		}
	}
}

func TestDefaultRegistry(t *testing.T) {
	r := NewDefaultRegistry(t.TempDir())

	expected := []string{"read_file", "write_file", "edit_file", "str_replace", "list_dir", "run_tests"}
	for _, name := range expected {
		if !r.Has(name) {
			t.Errorf("default registry missing tool %q", name)
		}
	}
}

func TestDefaultRegistry_Annotations(t *testing.T) {
	r := NewDefaultRegistry(t.TempDir())

	tests := []struct {
		name        string
		readOnly    bool
		destructive bool
	}{
		{"read_file", true, false},
		{"write_file", false, true},
		{"edit_file", false, true},
		{"str_replace", false, true},
		{"list_dir", true, false},
		{"run_tests", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ann := r.Annotations(tt.name)
			if ann.ReadOnly != tt.readOnly {
				t.Errorf("ReadOnly = %v, want %v", ann.ReadOnly, tt.readOnly)
			}
			if ann.Destructive != tt.destructive {
				t.Errorf("Destructive = %v, want %v", ann.Destructive, tt.destructive)
			}
		})
	}
}
