package config

import (
	"github.com/spf13/viper"
)

const (
	DefaultAppServerCommand = "codex"
	DefaultTheme            = "default"
)

type Config struct {
	AppServer AppServerConfig
	UI        UIConfig
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
		UI: UIConfig{
			Theme: viperInstance.GetString("ui.theme"),
		},
	}

	return cfg, nil
}
