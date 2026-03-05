package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// DiscoverServers auto-discovers MCP server configurations from common locations.
func DiscoverServers() []MCPServerConfig {
	var configs []MCPServerConfig

	home, err := os.UserHomeDir()
	if err != nil {
		return configs
	}

	// Check Claude Code MCP config
	claudeConfig := filepath.Join(home, ".claude", "mcp_servers.json")
	if servers, err := loadJSONConfig(claudeConfig); err == nil {
		configs = append(configs, servers...)
	}

	// Check VS Code MCP settings
	vscodeConfig := filepath.Join(home, ".vscode", "mcp.json")
	if servers, err := loadJSONConfig(vscodeConfig); err == nil {
		configs = append(configs, servers...)
	}

	// Check project-local .mcp.json
	if servers, err := loadJSONConfig(".mcp.json"); err == nil {
		configs = append(configs, servers...)
	}

	return configs
}

type mcpJSONConfig struct {
	MCPServers map[string]mcpJSONEntry `json:"mcpServers"`
}

type mcpJSONEntry struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

func loadJSONConfig(path string) ([]MCPServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg mcpJSONConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	var configs []MCPServerConfig
	for name, entry := range cfg.MCPServers {
		serverType := "stdio"
		command := entry.Command
		if len(entry.Args) > 0 {
			command += " " + joinArgs(entry.Args)
		}

		if entry.URL != "" {
			serverType = "http"
			command = ""
		}

		configs = append(configs, MCPServerConfig{
			Name:      name,
			Type:      serverType,
			Command:   command,
			URL:       entry.URL,
			Headers:   entry.Headers,
			AutoStart: false, // discovered servers default to manual start
		})
	}

	return configs, nil
}

func joinArgs(args []string) string {
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += " "
		}
		// Simple quoting for args with spaces
		if containsSpace(arg) {
			result += "\"" + arg + "\""
		} else {
			result += arg
		}
	}
	return result
}

func containsSpace(s string) bool {
	for _, c := range s {
		if c == ' ' {
			return true
		}
	}
	return false
}
