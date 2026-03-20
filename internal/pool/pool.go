package pool

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/robinojw/dj/internal/orchestrator"
	"github.com/robinojw/dj/internal/roster"
)

const DefaultMaxAgents = 10

const poolEventChannelSize = 128

type AgentPool struct {
	agents    map[string]*AgentProcess
	mu        sync.RWMutex
	events    chan PoolEvent
	command   string
	args      []string
	personas  map[string]roster.PersonaDefinition
	maxAgents int
	idCounter atomic.Int64
	ctx       context.Context
}

func NewAgentPool(command string, args []string, personas []roster.PersonaDefinition, maxAgents int) *AgentPool {
	personaMap := make(map[string]roster.PersonaDefinition, len(personas))
	for _, persona := range personas {
		personaMap[persona.ID] = persona
	}

	return &AgentPool{
		agents:    make(map[string]*AgentProcess),
		events:    make(chan PoolEvent, poolEventChannelSize),
		command:   command,
		args:      args,
		personas:  personaMap,
		maxAgents: maxAgents,
	}
}

func (agentPool *AgentPool) Events() <-chan PoolEvent {
	return agentPool.events
}

func (agentPool *AgentPool) Get(agentID string) (*AgentProcess, bool) {
	agentPool.mu.RLock()
	defer agentPool.mu.RUnlock()

	agent, exists := agentPool.agents[agentID]
	return agent, exists
}

func (agentPool *AgentPool) All() []*AgentProcess {
	agentPool.mu.RLock()
	defer agentPool.mu.RUnlock()

	result := make([]*AgentProcess, 0, len(agentPool.agents))
	for _, agent := range agentPool.agents {
		result = append(result, agent)
	}
	return result
}

func (agentPool *AgentPool) Count() int {
	agentPool.mu.RLock()
	defer agentPool.mu.RUnlock()

	return len(agentPool.agents)
}

func (agentPool *AgentPool) SetContext(ctx context.Context) {
	agentPool.ctx = ctx
}

func (agentPool *AgentPool) Spawn(personaID string, task string, parentAgentID string) (string, error) {
	agent, err := agentPool.registerAgent(personaID, task, parentAgentID)
	if err != nil {
		return "", err
	}

	isLive := agentPool.ctx != nil
	if !isLive {
		return agent.ID, nil
	}

	prompt := BuildWorkerPrompt(agent.Persona, task)
	if err := startAgentProcess(agentPool.ctx, agent, agentPool.command, agentPool.args, agentPool.events, prompt); err != nil {
		agentPool.removeAgent(agent.ID)
		return "", fmt.Errorf("start agent: %w", err)
	}

	return agent.ID, nil
}

func (agentPool *AgentPool) registerAgent(personaID string, task string, parentAgentID string) (*AgentProcess, error) {
	agentPool.mu.Lock()
	defer agentPool.mu.Unlock()

	isAtCapacity := len(agentPool.agents) >= agentPool.maxAgents
	if isAtCapacity {
		return nil, fmt.Errorf("agent pool at capacity (%d)", agentPool.maxAgents)
	}

	persona, exists := agentPool.personas[personaID]
	if !exists {
		return nil, fmt.Errorf("unknown persona: %s", personaID)
	}

	agentID := agentPool.nextAgentID(personaID)
	agent := &AgentProcess{
		ID:        agentID,
		PersonaID: personaID,
		Role:      RoleWorker,
		Task:      task,
		Status:    AgentStatusSpawning,
		ParentID:  parentAgentID,
		Persona:   &persona,
		Parser:    orchestrator.NewCommandParser(),
	}
	agentPool.agents[agentID] = agent

	return agent, nil
}

func (agentPool *AgentPool) removeAgent(agentID string) {
	agentPool.mu.Lock()
	defer agentPool.mu.Unlock()
	delete(agentPool.agents, agentID)
}

func (agentPool *AgentPool) StopAgent(agentID string) error {
	agentPool.mu.Lock()
	defer agentPool.mu.Unlock()

	agent, exists := agentPool.agents[agentID]
	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	if agent.Client != nil {
		agent.Client.Stop()
	}
	agent.Status = AgentStatusCompleted
	delete(agentPool.agents, agentID)
	return nil
}

func (agentPool *AgentPool) StopAll() {
	agentPool.mu.Lock()
	defer agentPool.mu.Unlock()

	for _, agent := range agentPool.agents {
		if agent.Client != nil {
			agent.Client.Stop()
		}
	}
	agentPool.agents = make(map[string]*AgentProcess)
}

func (agentPool *AgentPool) GetByThreadID(threadID string) (*AgentProcess, bool) {
	agentPool.mu.RLock()
	defer agentPool.mu.RUnlock()

	for _, agent := range agentPool.agents {
		if agent.ThreadID == threadID {
			return agent, true
		}
	}
	return nil, false
}

func (agentPool *AgentPool) GetOrchestrator() (*AgentProcess, bool) {
	agentPool.mu.RLock()
	defer agentPool.mu.RUnlock()

	for _, agent := range agentPool.agents {
		if agent.Role == RoleOrchestrator {
			return agent, true
		}
	}
	return nil, false
}

func (agentPool *AgentPool) Personas() map[string]roster.PersonaDefinition {
	return agentPool.personas
}

func (agentPool *AgentPool) nextAgentID(personaID string) string {
	counter := agentPool.idCounter.Add(1)
	return fmt.Sprintf("%s-%d", personaID, counter)
}
