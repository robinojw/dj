# Phase 8: Config, Help & Keybindings

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add Viper-based TOML configuration, a help overlay displaying keybindings, and the cobra CLI command structure. Config controls theme, default model, and app-server binary path.

**Architecture:** Cobra root command with `--config` flag. Viper loads `dj.toml` from the working directory, with user-level fallback at `~/.config/dj/config.toml`. Config struct is passed to AppModel on startup. Help overlay is a Bubble Tea component toggled with `?`. Keybinding display is a styled list rendered in the overlay.

**Tech Stack:** Go, cobra, viper, TOML

**Prerequisites:** Phase 7 (prefix keys, context menu, all TUI components)

---

### Task 1: Add Cobra and Viper Dependencies

**Files:**
- Modify: `go.mod`

**Step 1: Add dependencies**

```bash
go get github.com/spf13/cobra
go get github.com/spf13/viper
go mod tidy
```

**Step 2: Verify**

Run: `go build ./...`
Expected: builds successfully

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "deps: add cobra and viper for CLI and config"
```

---

### Task 2: Define Config Types

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Step 1: Write tests for config loading**

```go
// internal/config/config_test.go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/config/ -v`
Expected: FAIL — package not found

**Step 3: Implement config types and loader**

```go
// internal/config/config.go
package config

import (
	"github.com/spf13/viper"
)

// Default configuration values.
const (
	DefaultAppServerCommand = "codex"
	DefaultTheme            = "default"
)

// Config is the top-level DJ configuration.
type Config struct {
	AppServer AppServerConfig
	UI        UIConfig
}

// AppServerConfig controls the app-server connection.
type AppServerConfig struct {
	Command string
	Args    []string
}

// UIConfig controls UI appearance.
type UIConfig struct {
	Theme string
}

// Load reads configuration from the given path, falling back to defaults.
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("toml")

	v.SetDefault("appserver.command", DefaultAppServerCommand)
	v.SetDefault("appserver.args", []string{"app-server", "--listen", "stdio://"})
	v.SetDefault("ui.theme", DefaultTheme)

	if path != "" {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				// File exists but can't be read — use defaults
			}
		}
	}

	cfg := &Config{
		AppServer: AppServerConfig{
			Command: v.GetString("appserver.command"),
			Args:    v.GetStringSlice("appserver.args"),
		},
		UI: UIConfig{
			Theme: v.GetString("ui.theme"),
		},
	}

	return cfg, nil
}
```

**Step 4: Run tests**

Run: `go test ./internal/config/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat(config): Viper-based TOML config with defaults"
```

---

### Task 3: Build Help Overlay

**Files:**
- Create: `internal/tui/help.go`
- Create: `internal/tui/help_test.go`

**Step 1: Write tests for help overlay**

```go
// internal/tui/help_test.go
package tui

import (
	"strings"
	"testing"
)

func TestHelpRender(t *testing.T) {
	help := NewHelpModel()
	output := help.View()

	expectedBindings := []string{"←/→", "↑/↓", "Enter", "Esc", "Ctrl+B", "?", "Ctrl+C"}
	for _, binding := range expectedBindings {
		if !strings.Contains(output, binding) {
			t.Errorf("expected %q in help output:\n%s", binding, output)
		}
	}
}

func TestHelpContainsActions(t *testing.T) {
	help := NewHelpModel()
	output := help.View()

	expectedActions := []string{"Navigate", "Open session", "Back", "Menu", "Help", "Quit"}
	for _, action := range expectedActions {
		if !strings.Contains(output, action) {
			t.Errorf("expected %q in help output:\n%s", action, output)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -v -run TestHelp`
Expected: FAIL — `NewHelpModel` not defined

**Step 3: Implement help overlay**

```go
// internal/tui/help.go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	helpBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39")).
			Padding(1, 2)
	helpTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			MarginBottom(1)
	helpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Width(12)
	helpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))
)

type keybinding struct {
	key         string
	description string
}

var keybindings = []keybinding{
	{"←/→", "Navigate cards horizontally"},
	{"↑/↓", "Navigate cards vertically"},
	{"Enter", "Open session pane"},
	{"Esc", "Back / close overlay"},
	{"t", "Toggle tree view"},
	{"n", "New thread"},
	{"Ctrl+B", "Prefix key (tmux-style)"},
	{"Ctrl+B m", "Open context menu (Menu)"},
	{"?", "Toggle help (Help)"},
	{"Ctrl+C", "Quit"},
}

// HelpModel displays the keybinding reference overlay.
type HelpModel struct{}

// NewHelpModel creates a help overlay.
func NewHelpModel() HelpModel {
	return HelpModel{}
}

// View renders the help overlay.
func (h HelpModel) View() string {
	title := helpTitleStyle.Render("Keybindings")

	var lines []string
	for _, kb := range keybindings {
		key := helpKeyStyle.Render(kb.key)
		desc := helpDescStyle.Render(kb.description)
		lines = append(lines, fmt.Sprintf("%s %s", key, desc))
	}

	content := title + "\n" + strings.Join(lines, "\n")
	return helpBorderStyle.Render(content)
}
```

**Step 4: Run tests**

Run: `go test ./internal/tui/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/help.go internal/tui/help_test.go
git commit -m "feat(tui): help overlay with keybinding reference"
```

---

### Task 4: Build Cobra Root Command

**Files:**
- Modify: `cmd/dj/main.go`

**Step 1: Rewrite main.go with cobra command**

```go
// cmd/dj/main.go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/robinojw/dj/internal/config"
	"github.com/robinojw/dj/internal/state"
	"github.com/robinojw/dj/internal/tui"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:   "dj",
	Short: "DJ — Codex TUI Visualizer",
	RunE:  runApp,
}

func init() {
	rootCmd.Flags().StringVar(&configPath, "config", "", "path to dj.toml config file")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runApp(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	_ = cfg // used in later phases for app-server connection

	store := state.NewThreadStore()
	app := tui.NewAppModel(store)

	program := tea.NewProgram(app, tea.WithAltScreen())
	_, err = program.Run()
	return err
}
```

**Step 2: Verify it builds**

Run: `go build ./cmd/dj && ./dj --help`
Expected: Shows usage with `--config` flag

**Step 3: Commit**

```bash
git add cmd/dj/main.go
git commit -m "feat: cobra root command with --config flag"
```

---

### Task 5: Wire Help Toggle into App

**Files:**
- Modify: `internal/tui/app.go`
- Modify: `internal/tui/app_test.go`

**Step 1: Write test for ? toggling help**

Add to `app_test.go`:

```go
func TestAppHelpToggle(t *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	helpKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	updated, _ := app.Update(helpKey)
	appModel := updated.(AppModel)

	if !appModel.HelpVisible() {
		t.Error("expected help to be visible")
	}

	updated, _ = appModel.Update(helpKey)
	appModel = updated.(AppModel)

	if appModel.HelpVisible() {
		t.Error("expected help to be hidden")
	}
}
```

**Step 2: Add help toggle to App**

Add `help HelpModel`, `helpVisible bool` fields. Add `HelpVisible() bool` method. In `handleKey`, `?` toggles `helpVisible`. In `View()`, render help overlay on top when visible.

**Step 3: Run tests**

Run: `go test ./internal/tui/ -v`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/tui/app.go internal/tui/app_test.go
git commit -m "feat(tui): ? toggles help overlay"
```
