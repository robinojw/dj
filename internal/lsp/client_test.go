package lsp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectGo(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)

	binDir := t.TempDir()
	os.WriteFile(filepath.Join(binDir, "gopls"), []byte("#!/bin/sh\n"), 0755)
	t.Setenv("PATH", binDir+":"+os.Getenv("PATH"))

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

func TestDetectSkipsWhenCommandMissing(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)

	t.Setenv("PATH", t.TempDir())

	result := Detect(dir)
	if result != nil {
		t.Error("Expected nil when LSP server command is not on PATH")
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
