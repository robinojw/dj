package modes

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/robinojw/dj/config"
)

// PersistToolToAllowList adds a tool to the allow list in harness.toml.
func PersistToolToAllowList(configPath string, toolName string) error {
	// Read current config
	var cfg config.Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	// Check if already in list
	for _, t := range cfg.Execution.Allow.Tools {
		if t == toolName {
			return nil // already present
		}
	}

	// Add tool
	cfg.Execution.Allow.Tools = append(cfg.Execution.Allow.Tools, toolName)

	// Write back
	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("open config for write: %w", err)
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	return nil
}
