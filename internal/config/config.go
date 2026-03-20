package config

import (
	"github.com/spf13/viper"
)

const defaultCommand = "codex"

const (
	DefaultAppServerCommand   = defaultCommand
	DefaultInteractiveCommand = defaultCommand
	DefaultTheme              = "default"
	DefaultRosterPath         = ".roster"
	DefaultMaxAgents          = 10
)

const (
	keyAppServerCommand    = "appserver.command"
	keyAppServerArgs       = "appserver.args"
	keyInteractiveCommand  = "interactive.command"
	keyInteractiveArgs     = "interactive.args"
	keyUITheme             = "ui.theme"
	keyRosterPath          = "roster.path"
	keyRosterAutoOrch      = "roster.auto_orchestrate"
	keyPoolMaxAgents       = "pool.max_agents"
)

type Config struct {
	AppServer   AppServerConfig
	Interactive InteractiveConfig
	UI          UIConfig
	Roster      RosterConfig
	Pool        PoolConfig
}

type InteractiveConfig struct {
	Command string
	Args    []string
}

type AppServerConfig struct {
	Command string
	Args    []string
}

type UIConfig struct {
	Theme string
}

type RosterConfig struct {
	Path            string
	AutoOrchestrate bool
}

type PoolConfig struct {
	MaxAgents int
}

func Load(path string) (*Config, error) {
	viperInstance := viper.New()
	viperInstance.SetConfigType("toml")

	viperInstance.SetDefault(keyAppServerCommand, DefaultAppServerCommand)
	viperInstance.SetDefault(keyAppServerArgs, []string{"proto"})
	viperInstance.SetDefault(keyInteractiveCommand, DefaultInteractiveCommand)
	viperInstance.SetDefault(keyInteractiveArgs, []string{})
	viperInstance.SetDefault(keyUITheme, DefaultTheme)
	viperInstance.SetDefault(keyRosterPath, DefaultRosterPath)
	viperInstance.SetDefault(keyRosterAutoOrch, true)
	viperInstance.SetDefault(keyPoolMaxAgents, DefaultMaxAgents)

	if path != "" {
		viperInstance.SetConfigFile(path)
		_ = viperInstance.ReadInConfig()
	}

	cfg := &Config{
		AppServer: AppServerConfig{
			Command: viperInstance.GetString(keyAppServerCommand),
			Args:    viperInstance.GetStringSlice(keyAppServerArgs),
		},
		Interactive: InteractiveConfig{
			Command: viperInstance.GetString(keyInteractiveCommand),
			Args:    viperInstance.GetStringSlice(keyInteractiveArgs),
		},
		UI: UIConfig{
			Theme: viperInstance.GetString(keyUITheme),
		},
		Roster: RosterConfig{
			Path:            viperInstance.GetString(keyRosterPath),
			AutoOrchestrate: viperInstance.GetBool(keyRosterAutoOrch),
		},
		Pool: PoolConfig{
			MaxAgents: viperInstance.GetInt(keyPoolMaxAgents),
		},
	}

	return cfg, nil
}
