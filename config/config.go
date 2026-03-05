package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Model  ModelConfig  `toml:"model"`
	Theme  ThemeConfig  `toml:"theme"`
	MCP    MCPConfig    `toml:"mcp"`
	Skills SkillsConfig `toml:"skills"`
}

type ModelConfig struct {
	Default         string `toml:"default"`
	ReasoningEffort string `toml:"reasoning_effort"`
	TeamThreshold   int    `toml:"team_threshold"`
}

type ThemeConfig struct {
	Name string `toml:"name"`
}

type MCPConfig struct {
	Servers map[string]MCPServerEntry `toml:"servers"`
}

type MCPServerEntry struct {
	Type      string            `toml:"type"`
	Command   string            `toml:"command"`
	URL       string            `toml:"url"`
	Headers   map[string]string `toml:"headers"`
	AutoStart bool              `toml:"auto_start"`
}

type SkillsConfig struct {
	Paths []string `toml:"paths"`
}

func DefaultConfig() Config {
	return Config{
		Model: ModelConfig{
			Default:         "gpt-5.1-codex-mini",
			ReasoningEffort: "medium",
			TeamThreshold:   3,
		},
		Theme: ThemeConfig{
			Name: "tokyonight",
		},
		Skills: SkillsConfig{
			Paths: []string{
				"./.codex/skills",
				"~/.config/codex-harness/skills",
			},
		},
	}
}

// Load reads config from harness.toml in the project root,
// then overlays user config from ~/.config/codex-harness/config.toml.
func Load() (Config, error) {
	cfg := DefaultConfig()

	// Project-level config
	if _, err := os.Stat("harness.toml"); err == nil {
		if _, err := toml.DecodeFile("harness.toml", &cfg); err != nil {
			return cfg, err
		}
	}

	// User-level override
	home, err := os.UserHomeDir()
	if err == nil {
		userCfg := filepath.Join(home, ".config", "codex-harness", "config.toml")
		if _, err := os.Stat(userCfg); err == nil {
			if _, err := toml.DecodeFile(userCfg, &cfg); err != nil {
				return cfg, err
			}
		}
	}

	return cfg, nil
}

// ExpandPath resolves ~ to the user's home directory.
func ExpandPath(path string) string {
	if len(path) > 1 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
