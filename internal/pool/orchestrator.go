package pool

import (
	"fmt"

	"github.com/robinojw/dj/internal/orchestrator"
	"github.com/robinojw/dj/internal/roster"
)

const orchestratorAgentID = "orchestrator"

func (agentPool *AgentPool) SpawnOrchestrator(signals *roster.RepoSignals) (string, error) {
	agent, err := agentPool.registerOrchestrator()
	if err != nil {
		return "", err
	}

	isLive := agentPool.ctx != nil
	if !isLive {
		return orchestratorAgentID, nil
	}

	prompt := BuildOrchestratorPrompt(agentPool.personas, signals)
	if err := startAgentProcess(agentPool.ctx, agent, agentPool.command, agentPool.args, agentPool.events, prompt); err != nil {
		agentPool.removeAgent(orchestratorAgentID)
		return "", fmt.Errorf("start orchestrator: %w", err)
	}

	return orchestratorAgentID, nil
}

func (agentPool *AgentPool) registerOrchestrator() (*AgentProcess, error) {
	agentPool.mu.Lock()
	defer agentPool.mu.Unlock()

	isAtCapacity := len(agentPool.agents) >= agentPool.maxAgents
	if isAtCapacity {
		return nil, fmt.Errorf("agent pool at capacity (%d)", agentPool.maxAgents)
	}

	agent := &AgentProcess{
		ID:     orchestratorAgentID,
		Role:   RoleOrchestrator,
		Status: AgentStatusSpawning,
		Parser: orchestrator.NewCommandParser(),
	}
	agentPool.agents[orchestratorAgentID] = agent

	return agent, nil
}
