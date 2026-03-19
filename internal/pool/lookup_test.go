package pool

import (
	"testing"

	"github.com/robinojw/dj/internal/roster"
)

const (
	testThreadABC      = "thread-abc"
	testPersonaArchID  = "architect"
	testPersonaTestID  = "test"
	testPersonaArchName = "Architect"
	testPersonaTestName = "Tester"
	testTask           = "task"
	expectedTwoPersonas = 2
)

func TestGetByThreadID(testing *testing.T) {
	personas := []roster.PersonaDefinition{{ID: testPersonaArchID, Name: testPersonaArchName}}
	agentPool := NewAgentPool(testCommand, []string{testArg}, personas, DefaultMaxAgents)

	agentID, _ := agentPool.Spawn(testPersonaArchID, testTask, "")
	agent, _ := agentPool.Get(agentID)
	agent.ThreadID = testThreadABC

	found, exists := agentPool.GetByThreadID(testThreadABC)
	if !exists {
		testing.Fatal("expected to find agent by thread ID")
	}
	if found.ID != agentID {
		testing.Errorf("expected %s, got %s", agentID, found.ID)
	}
}

func TestGetByThreadIDNotFound(testing *testing.T) {
	agentPool := NewAgentPool(testCommand, []string{testArg}, nil, DefaultMaxAgents)
	_, exists := agentPool.GetByThreadID(nonexistentID)
	if exists {
		testing.Error("expected agent to not exist")
	}
}

func TestGetOrchestrator(testing *testing.T) {
	agentPool := NewAgentPool(testCommand, []string{testArg}, nil, DefaultMaxAgents)

	_, exists := agentPool.GetOrchestrator()
	if exists {
		testing.Error("expected no orchestrator initially")
	}
}

func TestPersonas(testing *testing.T) {
	personas := []roster.PersonaDefinition{
		{ID: testPersonaArchID, Name: testPersonaArchName},
		{ID: testPersonaTestID, Name: testPersonaTestName},
	}
	agentPool := NewAgentPool(testCommand, []string{testArg}, personas, DefaultMaxAgents)

	result := agentPool.Personas()
	if len(result) != expectedTwoPersonas {
		testing.Errorf("expected %d personas, got %d", expectedTwoPersonas, len(result))
	}
}
