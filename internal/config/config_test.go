package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.AppServer.Command != DefaultAppServerCommand {
		t.Errorf("expected default command %s, got %s", DefaultAppServerCommand, cfg.AppServer.Command)
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dj.toml")

	content := `
[appserver]
command = "/usr/local/bin/codex"

[ui]
theme = "dark"
`
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.AppServer.Command != "/usr/local/bin/codex" {
		t.Errorf("expected custom command, got %s", cfg.AppServer.Command)
	}
	if cfg.UI.Theme != "dark" {
		t.Errorf("expected dark theme, got %s", cfg.UI.Theme)
	}
}

func TestLoadMissingFileUsesDefaults(t *testing.T) {
	cfg, err := Load("/nonexistent/dj.toml")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.AppServer.Command != DefaultAppServerCommand {
		t.Errorf("expected default command, got %s", cfg.AppServer.Command)
	}
}
