package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/robinojw/dj/internal/appserver"
	"github.com/robinojw/dj/internal/config"
	"github.com/robinojw/dj/internal/pool"
	"github.com/robinojw/dj/internal/roster"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := state.NewThreadStore()
	var opts []tui.AppOption
	var agentPool *pool.AgentPool

	personas, signals := loadRoster(cfg)
	hasPersonas := len(personas) > 0
	shouldUsePool := hasPersonas && cfg.Roster.AutoOrchestrate

	if shouldUsePool {
		agentPool = pool.NewAgentPool(
			cfg.AppServer.Command,
			cfg.AppServer.Args,
			personas,
			cfg.Pool.MaxAgents,
		)
		agentPool.SetContext(ctx)
		opts = append(opts, tui.WithPool(agentPool))
		bootOrchestrator(agentPool, signals, store)
	} else {
		client := appserver.NewClient(cfg.AppServer.Command, cfg.AppServer.Args...)
		defer client.Stop()
		opts = append(opts, tui.WithClient(client))
	}

	opts = append(opts, tui.WithInteractiveCommand(cfg.Interactive.Command, cfg.Interactive.Args...))
	app := tui.NewAppModel(store, opts...)

	program := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())
	finalModel, err := program.Run()

	if finalApp, ok := finalModel.(tui.AppModel); ok {
		finalApp.StopAllPTYSessions()
	}

	if agentPool != nil {
		agentPool.StopAll()
	}

	return err
}

func bootOrchestrator(agentPool *pool.AgentPool, signals *roster.RepoSignals, store *state.ThreadStore) {
	agentID, err := agentPool.SpawnOrchestrator(signals)
	if err != nil {
		fmt.Fprintf(os.Stderr, "orchestrator: %v\n", err)
		return
	}
	store.Add(agentID, "Orchestrator")
	thread, exists := store.Get(agentID)
	if !exists {
		return
	}
	thread.AgentProcessID = agentID
	thread.AgentRole = pool.RoleOrchestrator
}

func loadRoster(cfg *config.Config) ([]roster.PersonaDefinition, *roster.RepoSignals) {
	personaDir := filepath.Join(cfg.Roster.Path, "personas")
	personas, err := roster.LoadPersonas(personaDir)
	if err != nil {
		return nil, nil
	}

	signalsPath := filepath.Join(cfg.Roster.Path, "signals.json")
	signals, err := roster.LoadSignals(signalsPath)
	if err != nil {
		return personas, nil
	}

	return personas, signals
}
