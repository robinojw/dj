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
	err := runner.Fire(HookPreToolCall, nil)
	if err != nil {
		t.Fatalf("Fire: %v", err)
	}

	data, _ := os.ReadFile(outFile)
	if !strings.Contains(string(data), "pre-tool") {
		t.Errorf("Expected 'pre-tool' in output, got %q", string(data))
	}
}

func TestFireUnconfiguredHook(t *testing.T) {
	cfg := Config{Hooks: map[string]string{}}
	runner := NewRunner(cfg)

	// Should be a no-op, not an error
	err := runner.Fire(HookOnError, nil)
	if err != nil {
		t.Errorf("Expected nil for unconfigured hook, got %v", err)
	}
}
