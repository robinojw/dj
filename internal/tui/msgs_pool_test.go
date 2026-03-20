package tui

import "testing"

const (
	testOrchestratorID = "orchestrator-1"
	testArchitectID    = "architect-1"
	testSecurityID     = "security-1"
	testTestID         = "test-1"
	testPersonaArch    = "architect"
	testTaskDesignAPI  = "Design API"
	testMsgContent     = "Need rate limiter"
	testFindingsMsg    = "Found 2 issues"

	errPoolExpectedSGotS = "expected %s, got %s"
)

func TestSpawnRequestMsgFields(testing *testing.T) {
	msg := SpawnRequestMsg{
		SourceAgentID: testOrchestratorID,
		Persona:       testPersonaArch,
		Task:          testTaskDesignAPI,
	}
	if msg.SourceAgentID != testOrchestratorID {
		testing.Errorf(errPoolExpectedSGotS, testOrchestratorID, msg.SourceAgentID)
	}
}

func TestAgentMessageMsgFields(testing *testing.T) {
	msg := AgentMessageMsg{
		SourceAgentID: testTestID,
		TargetAgentID: testArchitectID,
		Content:       testMsgContent,
	}
	if msg.TargetAgentID != testArchitectID {
		testing.Errorf(errPoolExpectedSGotS, testArchitectID, msg.TargetAgentID)
	}
}

func TestAgentCompleteMsgFields(testing *testing.T) {
	msg := AgentCompleteMsg{
		AgentID: testSecurityID,
		Content: testFindingsMsg,
	}
	if msg.AgentID != testSecurityID {
		testing.Errorf(errPoolExpectedSGotS, testSecurityID, msg.AgentID)
	}
}

func TestPoolEventMsgFields(testing *testing.T) {
	msg := PoolEventMsg{AgentID: testTestID}
	if msg.AgentID != testTestID {
		testing.Errorf(errPoolExpectedSGotS, testTestID, msg.AgentID)
	}
}
