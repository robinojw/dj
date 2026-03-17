package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/robinojw/dj/internal/appserver"
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

	client := appserver.NewClient(cfg.AppServer.Command, cfg.AppServer.Args...)
	defer client.Stop()

	router := appserver.NewEventRouter()
	client.Router = router

	store := state.NewThreadStore()
	app := tui.NewAppModel(store, client)

	program := tea.NewProgram(app, tea.WithAltScreen())
	tui.WireEventBridge(router, program)

	_, err = program.Run()
	return err
}
