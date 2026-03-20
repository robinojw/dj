package pool

import (
	"context"
	"testing"
	"time"

	"github.com/robinojw/dj/internal/appserver"
	"github.com/robinojw/dj/internal/roster"
)

const (
	testPersonaArch    = "architect"
	testTaskSome       = "some task"
	zeroMaxAgents      = 0
	nonexistentID      = "nonexistent"
	testProcessTimeout = 5 * time.Second
	testJSONRPCVersion = "2.0"
	testAgentID        = "test-agent-1"
	testCatCommand     = "cat"
	testSpawnFailed    = "Spawn failed: %v"
	testExpectedStatus = "expected status %s, got %s"
	testAgentExists    = "expected agent to exist"
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

func TestIsApprovalRequestExec(testing *testing.T) {
	message := appserver.JSONRPCMessage{
		JSONRPC: testJSONRPCVersion,
		ID:      "req-1",
		Method:  appserver.MethodExecApproval,
	}
	if !isApprovalRequest(message) {
		testing.Error("expected exec approval to be detected")
	}
}

func TestIsApprovalRequestFile(testing *testing.T) {
	message := appserver.JSONRPCMessage{
		JSONRPC: testJSONRPCVersion,
		ID:      "req-2",
		Method:  appserver.MethodFileApproval,
	}
	if !isApprovalRequest(message) {
		testing.Error("expected file approval to be detected")
	}
}

func TestIsApprovalRequestNotification(testing *testing.T) {
	message := appserver.JSONRPCMessage{
		JSONRPC: testJSONRPCVersion,
		Method:  appserver.MethodThreadStarted,
	}
	if isApprovalRequest(message) {
		testing.Error("expected notification to not be an approval request")
	}
}

func TestStartAgentProcess(testing *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testProcessTimeout)
	defer cancel()

	events := make(chan PoolEvent, poolEventChannelSize)
	agent := &AgentProcess{
		ID:     testAgentID,
		Status: AgentStatusSpawning,
	}

	err := startAgentProcess(ctx, agent, testCatCommand, []string{}, events, "hello world")
	if err != nil {
		testing.Fatalf("startAgentProcess failed: %v", err)
	}
	defer agent.Client.Stop()

	if agent.Client == nil {
		testing.Fatal("expected client to be set")
	}
	if agent.Status != AgentStatusActive {
		testing.Errorf(testExpectedStatus, AgentStatusActive, agent.Status)
	}

	select {
	case event := <-events:
		if event.AgentID != testAgentID {
			testing.Errorf("expected agent ID %s, got %s", testAgentID, event.AgentID)
		}
	case <-time.After(testProcessTimeout):
		testing.Fatal("timeout waiting for event from agent process")
	}
}

func TestStartAgentProcessBadCommand(testing *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testProcessTimeout)
	defer cancel()

	events := make(chan PoolEvent, poolEventChannelSize)
	agent := &AgentProcess{
		ID:     "test-fail-1",
		Status: AgentStatusSpawning,
	}

	err := startAgentProcess(ctx, agent, "nonexistent-binary-xyz", []string{}, events, "hello")
	if err == nil {
		testing.Error("expected error for nonexistent command")
	}
}

func TestSpawnLiveStartsProcess(testing *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testProcessTimeout)
	defer cancel()

	personas := []roster.PersonaDefinition{
		{ID: testPersonaArchID, Name: testPersonaArchName, Content: "System design specialist."},
	}
	agentPool := NewAgentPool(testCatCommand, []string{}, personas, DefaultMaxAgents)
	agentPool.SetContext(ctx)
	defer agentPool.StopAll()

	agentID, err := agentPool.Spawn(testPersonaArchID, testTask, "")
	if err != nil {
		testing.Fatalf(testSpawnFailed, err)
	}

	agent, exists := agentPool.Get(agentID)
	if !exists {
		testing.Fatal(testAgentExists)
	}
	if agent.Client == nil {
		testing.Fatal("expected client to be set on live spawn")
	}
	if agent.Status != AgentStatusActive {
		testing.Errorf(testExpectedStatus, AgentStatusActive, agent.Status)
	}
}

func TestSpawnWithoutContextBookkeepingOnly(testing *testing.T) {
	personas := []roster.PersonaDefinition{
		{ID: testPersonaArchID, Name: testPersonaArchName},
	}
	agentPool := NewAgentPool("echo", []string{}, personas, DefaultMaxAgents)

	agentID, err := agentPool.Spawn(testPersonaArchID, testTask, "")
	if err != nil {
		testing.Fatalf(testSpawnFailed, err)
	}

	agent, exists := agentPool.Get(agentID)
	if !exists {
		testing.Fatal(testAgentExists)
	}
	if agent.Client != nil {
		testing.Error("expected client to be nil without context")
	}
	if agent.Status != AgentStatusSpawning {
		testing.Errorf(testExpectedStatus, AgentStatusSpawning, agent.Status)
	}
}
