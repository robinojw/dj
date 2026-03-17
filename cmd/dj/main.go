package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/state"
	"github.com/robinojw/dj/internal/tui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	store := state.NewThreadStore()
	app := tui.NewAppModel(store)

	program := tea.NewProgram(app, tea.WithAltScreen())
	_, err := program.Run()
	return err
}
