package pool

import "testing"

const (
	testCommand     = "codex"
	testArg         = "proto"
	expectedZero    = 0
	errExpectedNil  = "expected non-nil pool"
	errExpectedDGot = "expected %d agents, got %d"
)

func TestNewAgentPool(testing *testing.T) {
	agentPool := NewAgentPool(testCommand, []string{testArg}, nil, DefaultMaxAgents)

	if agentPool == nil {
		testing.Fatal(errExpectedNil)
	}

	agents := agentPool.All()
	if len(agents) != expectedZero {
		testing.Errorf(errExpectedDGot, expectedZero, len(agents))
	}
}

func TestAgentPoolGet(testing *testing.T) {
	agentPool := NewAgentPool(testCommand, []string{testArg}, nil, DefaultMaxAgents)

	_, exists := agentPool.Get("nonexistent")
	if exists {
		testing.Error("expected agent to not exist")
	}
}

func TestAgentRoleConstants(testing *testing.T) {
	if RoleOrchestrator != "orchestrator" {
		testing.Errorf("expected orchestrator, got %s", RoleOrchestrator)
	}
	if RoleWorker != "worker" {
		testing.Errorf("expected worker, got %s", RoleWorker)
	}
}

func TestAgentStatusConstants(testing *testing.T) {
	if AgentStatusSpawning != "spawning" {
		testing.Errorf("expected spawning, got %s", AgentStatusSpawning)
	}
	if AgentStatusActive != "active" {
		testing.Errorf("expected active, got %s", AgentStatusActive)
	}
	if AgentStatusCompleted != "completed" {
		testing.Errorf("expected completed, got %s", AgentStatusCompleted)
	}
	if AgentStatusError != "error" {
		testing.Errorf("expected error, got %s", AgentStatusError)
	}
}
