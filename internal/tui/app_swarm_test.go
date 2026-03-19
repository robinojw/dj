package tui

import (
	"strings"
	"testing"

	poolpkg "github.com/robinojw/dj/internal/pool"
	"github.com/robinojw/dj/internal/roster"
	"github.com/robinojw/dj/internal/state"
)

const (
	testSwarmMaxAgents = 10
	testSwarmCommand   = "echo"
)

func TestAppModelSwarmFieldsDefault(testing *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)
	if app.menuIntent != MenuIntentThread {
		testing.Error("expected default menu intent to be thread")
	}
	if app.inputBarVisible {
		testing.Error("expected input bar hidden by default")
	}
	if app.swarmFilter {
		testing.Error("expected swarm filter off by default")
	}
}

func TestNewAppModelPoolSetsSwarmActive(testing *testing.T) {
	store := state.NewThreadStore()
	agentPool := poolpkg.NewAgentPool(testSwarmCommand, []string{}, nil, testSwarmMaxAgents)
	app := NewAppModel(store, WithPool(agentPool))
	view := app.header.View()
	if !strings.Contains(view, "p: persona") {
		testing.Error("expected swarm hints in header when pool is set")
	}
}

func TestShowPersonaPickerShowsMenu(testing *testing.T) {
	store := state.NewThreadStore()
	personas := []roster.PersonaDefinition{
		{ID: "architect", Name: "Architect"},
		{ID: "test", Name: "Test"},
	}
	agentPool := poolpkg.NewAgentPool(testSwarmCommand, []string{}, personas, testSwarmMaxAgents)
	app := NewAppModel(store, WithPool(agentPool))

	updated, _ := app.showPersonaPicker()
	resultApp := updated.(AppModel)

	if !resultApp.menuVisible {
		testing.Error("expected menu to be visible after showPersonaPicker")
	}
	if resultApp.menuIntent != MenuIntentPersonaPicker {
		testing.Error("expected menu intent to be persona picker")
	}
}

func TestShowPersonaPickerNoPool(testing *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	updated, _ := app.showPersonaPicker()
	resultApp := updated.(AppModel)

	if resultApp.menuVisible {
		testing.Error("expected menu hidden when no pool")
	}
}

func TestShowPersonaPickerNoPersonas(testing *testing.T) {
	store := state.NewThreadStore()
	agentPool := poolpkg.NewAgentPool(testSwarmCommand, []string{}, nil, testSwarmMaxAgents)
	app := NewAppModel(store, WithPool(agentPool))

	updated, _ := app.showPersonaPicker()
	resultApp := updated.(AppModel)

	if resultApp.menuVisible {
		testing.Error("expected menu hidden when no personas")
	}
}
