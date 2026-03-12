package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandVars(t *testing.T) {
	template := "echo '{{tool}} {{args}}'"
	vars := map[string]string{"tool": "write_file", "args": "main.go"}
	result := expandVars(template, vars)

	if !strings.Contains(result, "write_file") {
		t.Errorf("Expected 'write_file' in result, got %q", result)
	}
	if !strings.Contains(result, "main.go") {
		t.Errorf("Expected 'main.go' in result, got %q", result)
	}
}

func TestFireHook(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "hook_output.txt")

	cfg := Config{
		Hooks: map[string]string{
			string(HookPreToolCall): "echo 'pre-tool' > " + outFile,
		},
	}

	runner := NewRunner(cfg)
	_, err := runner.Fire(HookPreToolCall, nil)
	if err != nil {
		t.Fatalf("Fire: %v", err)
	}

	data, _ := os.ReadFile(outFile)
	if !strings.Contains(string(data), "pre-tool") {
		t.Errorf("Expected 'pre-tool' in output, got %q", string(data))
	}
}

func TestFireUnconfiguredReturnsNil(t *testing.T) {
	cfg := Config{Hooks: map[string]string{}}
	runner := NewRunner(cfg)

	result, err := runner.Fire(HookOnError, nil)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if result != nil {
		t.Errorf("Expected nil result for unconfigured hook, got %+v", result)
	}
}

func TestFireReturnsResult(t *testing.T) {
	cfg := Config{
		Hooks: map[string]string{
			string(HookPreToolCall): "echo hello",
		},
	}
	runner := NewRunner(cfg)
	result, err := runner.Fire(HookPreToolCall, nil)
	if err != nil {
		t.Fatalf("Fire() error = %v", err)
	}
	if result == nil {
		t.Fatal("Fire() returned nil result for configured hook")
	}
	if result.Stdout != "hello\n" {
		t.Errorf("Stdout = %q, want %q", result.Stdout, "hello\n")
	}
	if result.Stderr != "" {
		t.Errorf("Stderr = %q, want empty", result.Stderr)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
	if result.Duration <= 0 {
		t.Error("Duration should be positive")
	}
	if result.Event != HookPreToolCall {
		t.Errorf("Event = %q, want %q", result.Event, HookPreToolCall)
	}
}

func TestFireCapturesStderr(t *testing.T) {
	cfg := Config{
		Hooks: map[string]string{
			string(HookOnError): "echo err >&2",
		},
	}
	runner := NewRunner(cfg)
	result, err := runner.Fire(HookOnError, nil)
	if err != nil {
		t.Fatalf("Fire() error = %v", err)
	}
	if result.Stdout != "" {
		t.Errorf("Stdout = %q, want empty", result.Stdout)
	}
	if result.Stderr != "err\n" {
		t.Errorf("Stderr = %q, want %q", result.Stderr, "err\n")
	}
}

func TestFireNonZeroExit(t *testing.T) {
	cfg := Config{
		Hooks: map[string]string{
			string(HookPreToolCall): "exit 42",
		},
	}
	runner := NewRunner(cfg)
	result, err := runner.Fire(HookPreToolCall, nil)
	if err != nil {
		t.Fatalf("Fire() infrastructure error = %v", err)
	}
	if result == nil {
		t.Fatal("Fire() returned nil result")
	}
	if result.ExitCode != 42 {
		t.Errorf("ExitCode = %d, want 42", result.ExitCode)
	}
}

