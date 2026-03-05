package config

import (
	"testing"

	"github.com/BurntSushi/toml"
)

func TestExecutionConfig_Parse(t *testing.T) {
	tomlData := `
[execution]
default_mode = "confirm"

[execution.allow]
tools = ["read_file", "bash(git status*)"]

[execution.deny]
tools = ["bash(rm -rf*)", "write_file(.env*)"]
`
	var cfg Config
	if _, err := toml.Decode(tomlData, &cfg); err != nil {
		t.Fatalf("Failed to parse TOML: %v", err)
	}

	if cfg.Execution.DefaultMode != "confirm" {
		t.Errorf("Expected default_mode=confirm, got %q", cfg.Execution.DefaultMode)
	}
	if len(cfg.Execution.Allow.Tools) != 2 {
		t.Errorf("Expected 2 allow tools, got %d", len(cfg.Execution.Allow.Tools))
	}
	if len(cfg.Execution.Deny.Tools) != 2 {
		t.Errorf("Expected 2 deny tools, got %d", len(cfg.Execution.Deny.Tools))
	}
}

func TestExecutionConfig_DefaultMode(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Execution.DefaultMode != "confirm" {
		t.Errorf("Expected default mode=confirm, got %q", cfg.Execution.DefaultMode)
	}
}
