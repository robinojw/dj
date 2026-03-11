package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunTests_PassingTests(t *testing.T) {
	dir := t.TempDir()

	// Create a minimal Go module with a passing test
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module testmod\n\ngo 1.21\n"), 0644)
	os.WriteFile(filepath.Join(dir, "main_test.go"), []byte(`package testmod

import "testing"

func TestPass(t *testing.T) {
	if 1+1 != 2 {
		t.Fatal("math is broken")
	}
}
`), 0644)

	handler := RunTestsHandler(dir)
	result, err := handler(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "PASS") {
		t.Errorf("expected PASS in result: %s", result)
	}
}

func TestRunTests_FailingTests(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module testmod\n\ngo 1.21\n"), 0644)
	os.WriteFile(filepath.Join(dir, "main_test.go"), []byte(`package testmod

import "testing"

func TestFail(t *testing.T) {
	t.Fatal("intentional failure")
}
`), 0644)

	handler := RunTestsHandler(dir)
	result, err := handler(context.Background(), nil)

	// run_tests should not return an error — it captures test failures in the output
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "FAIL") {
		t.Errorf("expected FAIL in result: %s", result)
	}
}

func TestRunTests_CustomPackage(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module testmod\n\ngo 1.21\n"), 0644)
	os.WriteFile(filepath.Join(dir, "main_test.go"), []byte(`package testmod

import "testing"

func TestHello(t *testing.T) {}
`), 0644)

	handler := RunTestsHandler(dir)
	result, err := handler(context.Background(), map[string]any{
		"package": ".",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "PASS") {
		t.Errorf("expected PASS in result: %s", result)
	}
}

func TestRunTests_RunFilter(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module testmod\n\ngo 1.21\n"), 0644)
	os.WriteFile(filepath.Join(dir, "main_test.go"), []byte(`package testmod

import "testing"

func TestAlpha(t *testing.T) {}
func TestBeta(t *testing.T) { t.Fatal("fail") }
`), 0644)

	handler := RunTestsHandler(dir)
	result, err := handler(context.Background(), map[string]any{
		"package": ".",
		"run":     "TestAlpha",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "PASS") {
		t.Errorf("expected PASS when filtering to TestAlpha: %s", result)
	}
}
