package config

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	errLoadFailed              = "Load failed: %v"
	errUnexpected              = "unexpected error: %v"
	expectedDefaultRosterPath  = ".roster"
	expectedDefaultMaxAgents   = 10
	permFile                   = 0o644
)

func TestLoadDefaults(testing *testing.T) {
	cfg, err := Load("")
	if err != nil {
		testing.Fatalf(errLoadFailed, err)
	}
	if cfg.AppServer.Command != DefaultAppServerCommand {
		testing.Errorf("expected default command %s, got %s", DefaultAppServerCommand, cfg.AppServer.Command)
	}
}

func TestLoadFromFile(testing *testing.T) {
	dir := testing.TempDir()
	path := filepath.Join(dir, "dj.toml")

	content := `
[appserver]
command = "/usr/local/bin/codex"

[ui]
theme = "dark"
`
	os.WriteFile(path, []byte(content), permFile)

	cfg, err := Load(path)
	if err != nil {
		testing.Fatalf(errLoadFailed, err)
	}
	if cfg.AppServer.Command != "/usr/local/bin/codex" {
		testing.Errorf("expected custom command, got %s", cfg.AppServer.Command)
	}
	if cfg.UI.Theme != "dark" {
		testing.Errorf("expected dark theme, got %s", cfg.UI.Theme)
	}
}

func TestLoadMissingFileUsesDefaults(testing *testing.T) {
	cfg, err := Load("/nonexistent/dj.toml")
	if err != nil {
		testing.Fatalf(errLoadFailed, err)
	}
	if cfg.AppServer.Command != DefaultAppServerCommand {
		testing.Errorf("expected default command, got %s", cfg.AppServer.Command)
	}
}

func TestDefaultRosterConfig(testing *testing.T) {
	cfg, err := Load("")
	if err != nil {
		testing.Fatalf(errUnexpected, err)
	}
	if cfg.Roster.Path != expectedDefaultRosterPath {
		testing.Errorf("expected %s, got %s", expectedDefaultRosterPath, cfg.Roster.Path)
	}
	if !cfg.Roster.AutoOrchestrate {
		testing.Error("expected auto_orchestrate to be true by default")
	}
}

func TestDefaultPoolConfig(testing *testing.T) {
	cfg, err := Load("")
	if err != nil {
		testing.Fatalf(errUnexpected, err)
	}
	if cfg.Pool.MaxAgents != expectedDefaultMaxAgents {
		testing.Errorf("expected max_agents %d, got %d", expectedDefaultMaxAgents, cfg.Pool.MaxAgents)
	}
}
