package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	gotui "github.com/grindlemire/go-tui"
	"github.com/robinojw/dj/config"
	"github.com/robinojw/dj/internal/api"
	"github.com/robinojw/dj/internal/hooks"
	"github.com/robinojw/dj/internal/lsp"
	"github.com/robinojw/dj/internal/mcp"
	"github.com/robinojw/dj/internal/memory"
	"github.com/robinojw/dj/internal/skills"
	"github.com/robinojw/dj/internal/tools"
	tuipkg "github.com/robinojw/dj/internal/tui"
	"github.com/robinojw/dj/internal/tui/theme"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("dj %s (%s) built %s\n", version, shortCommit(commit), buildDate)
		os.Exit(0)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "OPENAI_API_KEY environment variable is required")
		os.Exit(1)
	}

	// Load theme
	t := loadTheme(cfg.Theme.Name)

	// Create API client and tracker
	client := api.NewResponsesClient(apiKey)
	tracker := api.NewTracker(cfg.Model.Default)

	// Load skills
	skillPaths := cfg.Skills.Paths
	// Append built-in skills path
	if exe, err := os.Executable(); err == nil {
		skillPaths = append(skillPaths, filepath.Join(filepath.Dir(exe), "..", "skills"))
	}
	skillPaths = append(skillPaths, "skills") // local skills directory
	skillRegistry := skills.NewRegistry(skillPaths)
	if err := skillRegistry.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load skills: %v\n", err)
	}

	// Set up MCP servers
	mcpConfigs := buildMCPConfigs(cfg)
	mcpRegistry := mcp.NewRegistry(mcpConfigs)
	ctx := context.Background()
	if err := mcpRegistry.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: MCP startup error: %v\n", err)
	}
	defer mcpRegistry.StopAll()

	// Auto-detect and start LSP server
	var lspClient *lsp.Client
	if cfg.LSP.Enabled || cfg.LSP.Language == "" {
		cwd, _ := os.Getwd()
		if detected := lsp.Detect(cwd); detected != nil {
			lspClient = lsp.NewClient(detected.Config, detected.RootPath)
			if err := lspClient.Start(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: LSP server failed to start: %v\n", err)
				lspClient = nil
			} else {
				defer lspClient.Close()
			}
		}
	}
	_ = lspClient // will be wired to app in future steps

	// Set up memory manager
	memMgr := memory.DefaultManager()
	_ = memMgr // will be wired to app in future steps

	// Set up event hooks
	var hookTimeout time.Duration
	if cfg.Hooks.Timeout != "" {
		if parsed, err := time.ParseDuration(cfg.Hooks.Timeout); err == nil {
			hookTimeout = parsed
		} else {
			fmt.Fprintf(os.Stderr, "Warning: invalid hooks timeout %q: %v\n", cfg.Hooks.Timeout, err)
		}
	}
	hookRunner := hooks.NewRunner(hooks.Config{
		Hooks: map[string]string{
			string(hooks.HookPreToolCall):  cfg.Hooks.PreToolCall,
			string(hooks.HookPostToolCall): cfg.Hooks.PostToolCall,
			string(hooks.HookOnError):      cfg.Hooks.OnError,
			string(hooks.HookSessionEnd):   cfg.Hooks.OnSessionEnd,
		},
		Timeout: hookTimeout,
	})
	defer hookRunner.FireAsync(hooks.HookSessionEnd, map[string]string{"summary": "session ended"})

	// Create tool registry
	cwd, _ := os.Getwd()
	toolRegistry := tools.NewDefaultRegistry(cwd)

	rootComponent := tuipkg.NewRootApp(t, client, tracker, cfg.Model.Default, cfg, toolRegistry, hookRunner)

	gotuiApp, err := gotui.NewApp(
		gotui.WithInlineHeight(3),
		gotui.WithRootComponent(rootComponent),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating TUI: %v\n", err)
		os.Exit(1)
	}
	defer gotuiApp.Close()

	if err := gotuiApp.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func shortCommit(s string) string {
	if len(s) > 7 {
		return s[:7]
	}
	return s
}

func loadTheme(name string) *theme.Theme {
	// Try built-in themes directory relative to the binary
	exe, err := os.Executable()
	if err == nil {
		themesDir := filepath.Join(filepath.Dir(exe), "..", "themes")
		if t, err := theme.LoadBuiltin(name, themesDir); err == nil {
			return t
		}
	}

	// Try local themes directory
	if t, err := theme.LoadBuiltin(name, "themes"); err == nil {
		return t
	}

	// Try absolute path (user may have specified a file path)
	if t, err := theme.LoadFromFile(name); err == nil {
		return t
	}

	return theme.DefaultTheme()
}

func buildMCPConfigs(cfg config.Config) []mcp.MCPServerConfig {
	var configs []mcp.MCPServerConfig

	for name, entry := range cfg.MCP.Servers {
		configs = append(configs, mcp.MCPServerConfig{
			Name:      name,
			Type:      entry.Type,
			Command:   entry.Command,
			URL:       entry.URL,
			Headers:   entry.Headers,
			AutoStart: entry.AutoStart,
		})
	}

	// Also discover from common config locations
	discovered := mcp.DiscoverServers()
	configs = append(configs, discovered...)

	return configs
}
