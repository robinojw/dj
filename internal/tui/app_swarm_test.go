package tui

import (
	"strings"
	"testing"

	poolpkg "github.com/robinojw/dj/internal/pool"
	"github.com/robinojw/dj/internal/roster"
	"github.com/robinojw/dj/internal/state"
)

const (
	testSwarmMaxAgents   = 10
	testSwarmCommand     = "echo"
	testSwarmPersonaID   = "architect"
	testSwarmPersonaName = "Architect"
	testSwarmTask        = "Design API"
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
		{ID: testSwarmPersonaID, Name: testSwarmPersonaName},
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

func TestPersonaPickerDispatchShowsInputBar(testing *testing.T) {
	store := state.NewThreadStore()
	personas := []roster.PersonaDefinition{
		{ID: testSwarmPersonaID, Name: testSwarmPersonaName},
	}
	agentPool := poolpkg.NewAgentPool(testSwarmCommand, []string{}, personas, testSwarmMaxAgents)
	app := NewAppModel(store, WithPool(agentPool))

	updated, _ := app.showPersonaPicker()
	app = updated.(AppModel)

	selected := app.menu.Selected()
	app.menuVisible = false
	updated, _ = app.dispatchPersonaPick(selected)
	resultApp := updated.(AppModel)

	if !resultApp.inputBarVisible {
		testing.Error("expected input bar visible after persona pick")
	}
	if resultApp.inputBarIntent != IntentSpawnTask {
		testing.Error("expected spawn task intent")
	}
}

func TestSendMessageToAgentShowsMenu(testing *testing.T) {
	store := state.NewThreadStore()
	personas := []roster.PersonaDefinition{
		{ID: testSwarmPersonaID, Name: testSwarmPersonaName},
	}
	agentPool := poolpkg.NewAgentPool(testSwarmCommand, []string{}, personas, testSwarmMaxAgents)
	agentPool.Spawn(testSwarmPersonaID, testSwarmTask, "")
	app := NewAppModel(store, WithPool(agentPool))

	updated, _ := app.sendMessageToAgent()
	resultApp := updated.(AppModel)

	if !resultApp.menuVisible {
		testing.Error("expected menu visible for agent picker")
	}
	if resultApp.menuIntent != MenuIntentAgentPicker {
		testing.Error("expected agent picker intent")
	}
}

func TestSendMessageToAgentNoAgents(testing *testing.T) {
	store := state.NewThreadStore()
	agentPool := poolpkg.NewAgentPool(testSwarmCommand, []string{}, nil, testSwarmMaxAgents)
	app := NewAppModel(store, WithPool(agentPool))

	updated, _ := app.sendMessageToAgent()
	resultApp := updated.(AppModel)

	if resultApp.menuVisible {
		testing.Error("expected menu hidden when no agents")
	}
}

func TestDispatchAgentPickShowsInputBar(testing *testing.T) {
	store := state.NewThreadStore()
	personas := []roster.PersonaDefinition{
		{ID: testSwarmPersonaID, Name: testSwarmPersonaName},
	}
	agentPool := poolpkg.NewAgentPool(testSwarmCommand, []string{}, personas, testSwarmMaxAgents)
	agentID, _ := agentPool.Spawn(testSwarmPersonaID, testSwarmTask, "")
	app := NewAppModel(store, WithPool(agentPool))

	item := MenuItem{Label: agentID, Key: rune(agentID[0])}
	updated, _ := app.dispatchAgentPick(item)
	resultApp := updated.(AppModel)

	if !resultApp.inputBarVisible {
		testing.Error("expected input bar visible after agent pick")
	}
	if resultApp.inputBarIntent != IntentSendMessage {
		testing.Error("expected send message intent")
	}
	if resultApp.pendingTargetAgentID != agentID {
		testing.Errorf("expected target %s, got %s", agentID, resultApp.pendingTargetAgentID)
	}
}

func TestToggleSwarmViewFiltersCanvas(testing *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	updated, _ := app.toggleSwarmView()
	resultApp := updated.(AppModel)

	if !resultApp.swarmFilter {
		testing.Error("expected swarm filter enabled after toggle")
	}

	updated2, _ := resultApp.toggleSwarmView()
	resultApp2 := updated2.(AppModel)

	if resultApp2.swarmFilter {
		testing.Error("expected swarm filter disabled after second toggle")
	}
}
