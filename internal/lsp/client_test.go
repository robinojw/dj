package lsp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectGo(t *testing.T) {
	// Create a temp dir with a go.mod file
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)

	result := Detect(dir)
	if result == nil {
		t.Fatal("Expected to detect Go LSP server")
	}
	if result.Config.Language != "go" {
		t.Errorf("Expected language 'go', got %s", result.Config.Language)
	}
	if result.Config.Command != "gopls" {
		t.Errorf("Expected command 'gopls', got %s", result.Config.Command)
	}
}

func TestDetectNoMatch(t *testing.T) {
	result := Detect(t.TempDir())
	if result != nil {
		t.Error("Expected nil for directory with no known language markers")
	}
}

func TestDiagnosticsFormat(t *testing.T) {
	d := Diagnostic{
		File:     "main.go",
		Line:     10,
		Column:   5,
		Severity: "error",
		Message:  "undefined: foo",
		Source:   "gopls",
	}
	expected := "main.go:10:5: error: undefined: foo (gopls)"
	got := FormatDiagnostic(d)
	if got != expected {
		t.Errorf("Expected %q, got %q", expected, got)
	}
}
