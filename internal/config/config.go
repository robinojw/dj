package config

import (
	"github.com/spf13/viper"
)

const (
	DefaultAppServerCommand    = "codex"
	DefaultInteractiveCommand  = "codex"
	DefaultTheme               = "default"
)

type Config struct {
	AppServer   AppServerConfig
	Interactive InteractiveConfig
	UI          UIConfig
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

func Load(path string) (*Config, error) {
	viperInstance := viper.New()
	viperInstance.SetConfigType("toml")

	viperInstance.SetDefault("appserver.command", DefaultAppServerCommand)
	viperInstance.SetDefault("appserver.args", []string{"proto"})
	viperInstance.SetDefault("interactive.command", DefaultInteractiveCommand)
	viperInstance.SetDefault("interactive.args", []string{})
	viperInstance.SetDefault("ui.theme", DefaultTheme)

	if path != "" {
		viperInstance.SetConfigFile(path)
		_ = viperInstance.ReadInConfig()
	}

	cfg := &Config{
		AppServer: AppServerConfig{
			Command: viperInstance.GetString("appserver.command"),
			Args:    viperInstance.GetStringSlice("appserver.args"),
		},
		Interactive: InteractiveConfig{
			Command: viperInstance.GetString("interactive.command"),
			Args:    viperInstance.GetStringSlice("interactive.args"),
		},
		UI: UIConfig{
			Theme: viperInstance.GetString("ui.theme"),
		},
	}

	return cfg, nil
}
