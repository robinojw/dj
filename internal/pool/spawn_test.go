package pool

import "testing"

const (
	testPersonaArch  = "architect"
	testTaskSome     = "some task"
	zeroMaxAgents    = 0
	nonexistentID    = "nonexistent"
)

func TestSpawnRejectsUnknownPersona(testing *testing.T) {
	agentPool := NewAgentPool(testCommand, []string{testArg}, nil, DefaultMaxAgents)
	_, err := agentPool.Spawn(nonexistentID, testTaskSome, "")
	if err == nil {
		testing.Error("expected error for unknown persona")
	}
}

func TestSpawnRejectsAtCapacity(testing *testing.T) {
	agentPool := NewAgentPool(testCommand, []string{testArg}, nil, zeroMaxAgents)
	_, err := agentPool.Spawn(testPersonaArch, testTaskSome, "")
	if err == nil {
		testing.Error("expected error when at capacity")
	}
}

func TestNextAgentID(testing *testing.T) {
	agentPool := NewAgentPool(testCommand, []string{testArg}, nil, DefaultMaxAgents)
	id1 := agentPool.nextAgentID(testPersonaArch)
	id2 := agentPool.nextAgentID(testPersonaArch)
	if id1 == id2 {
		testing.Errorf("expected unique IDs, got %s and %s", id1, id2)
	}
}

func TestStopAgentNotFound(testing *testing.T) {
	agentPool := NewAgentPool(testCommand, []string{testArg}, nil, DefaultMaxAgents)
	err := agentPool.StopAgent(nonexistentID)
	if err == nil {
		testing.Error("expected error for nonexistent agent")
	}
}
