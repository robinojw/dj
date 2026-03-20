package tui

import (
	"testing"

	poolpkg "github.com/robinojw/dj/internal/pool"
	"github.com/robinojw/dj/internal/roster"
	"github.com/robinojw/dj/internal/state"
)

const (
	poolTestPersonaID   = "architect"
	poolTestPersonaName = "Architect"
	poolTestTaskLabel   = "Design API"
	poolTestSourceAgent = "orchestrator-1"
	poolTestAgentID     = "test-1"
	poolTestThreadID    = "t1"
	poolTestThreadTitle = "Test Agent"
	poolTestDoneContent = "Done"
)

func TestHandleSpawnRequestCreatesThread(testing *testing.T) {
	store := state.NewThreadStore()
	personas := []roster.PersonaDefinition{{ID: poolTestPersonaID, Name: poolTestPersonaName}}
	agentPool := poolpkg.NewAgentPool(appTestCmdEcho, []string{appTestArgHello}, personas, appTestPoolMaxAgents)

	app := NewAppModel(store, WithPool(agentPool))
	msg := SpawnRequestMsg{
		SourceAgentID: poolTestSourceAgent,
		Persona:       poolTestPersonaID,
		Task:          poolTestTaskLabel,
	}

	updated, _ := app.handleSpawnRequest(msg)
	resultApp := updated.(AppModel)
	threads := resultApp.store.All()

	hasThread := len(threads) > 0
	if !hasThread {
		testing.Error("expected at least one thread after spawn request")
	}
}

func TestHandleAgentCompleteUpdatesStatus(testing *testing.T) {
	store := state.NewThreadStore()
	store.Add(poolTestThreadID, poolTestThreadTitle)

	agentPool := poolpkg.NewAgentPool(appTestCmdEcho, []string{}, nil, appTestPoolMaxAgents)
	app := NewAppModel(store, WithPool(agentPool))

	msg := AgentCompleteMsg{
		AgentID: poolTestThreadID,
		Content: poolTestDoneContent,
	}

	updated, _ := app.handleAgentComplete(msg)
	_ = updated.(AppModel)

	thread, exists := store.Get(poolTestThreadID)
	if !exists {
		testing.Fatal("expected thread to exist")
	}
	if thread.Status != state.StatusCompleted {
		testing.Errorf("expected completed status, got %s", thread.Status)
	}
}
