package pool

import (
	"context"
	"testing"
	"time"

	"github.com/robinojw/dj/internal/roster"
)

const (
	testOrchestratorTimeout = 5 * time.Second
	testSpawnOrchFailed     = "SpawnOrchestrator failed: %v"
	testExpectedSGotS       = "expected %s, got %s"
	testEchoCommand         = "echo"
	testOrchestratorExists  = "expected orchestrator to exist"
)

func TestSpawnOrchestrator(testing *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testOrchestratorTimeout)
	defer cancel()

	personas := []roster.PersonaDefinition{
		{ID: testPersonaArchID, Name: testPersonaArchName, Description: "System design"},
	}
	agentPool := NewAgentPool(testCatCommand, []string{}, personas, DefaultMaxAgents)
	agentPool.SetContext(ctx)
	defer agentPool.StopAll()

	agentID, err := agentPool.SpawnOrchestrator(nil)
	if err != nil {
		testing.Fatalf(testSpawnOrchFailed, err)
	}
	if agentID != orchestratorAgentID {
		testing.Errorf(testExpectedSGotS, orchestratorAgentID, agentID)
	}

	agent, exists := agentPool.GetOrchestrator()
	if !exists {
		testing.Fatal(testOrchestratorExists)
	}
	if agent.Role != RoleOrchestrator {
		testing.Errorf(testExpectedSGotS, RoleOrchestrator, agent.Role)
	}
	if agent.Status != AgentStatusActive {
		testing.Errorf(testExpectedStatus, AgentStatusActive, agent.Status)
	}
}

func TestSpawnOrchestratorWithoutContext(testing *testing.T) {
	personas := []roster.PersonaDefinition{
		{ID: testPersonaArchID, Name: testPersonaArchName},
	}
	agentPool := NewAgentPool(testEchoCommand, []string{}, personas, DefaultMaxAgents)

	agentID, err := agentPool.SpawnOrchestrator(nil)
	if err != nil {
		testing.Fatalf(testSpawnOrchFailed, err)
	}
	if agentID != orchestratorAgentID {
		testing.Errorf(testExpectedSGotS, orchestratorAgentID, agentID)
	}

	agent, exists := agentPool.GetOrchestrator()
	if !exists {
		testing.Fatal(testOrchestratorExists)
	}
	if agent.Client != nil {
		testing.Error("expected client to be nil without context")
	}
}

func TestSpawnOrchestratorAtCapacity(testing *testing.T) {
	agentPool := NewAgentPool(testEchoCommand, []string{}, nil, zeroMaxAgents)
	_, err := agentPool.SpawnOrchestrator(nil)
	if err == nil {
		testing.Error("expected error when at capacity")
	}
}
