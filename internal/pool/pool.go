package pool

import (
	"sync"
	"sync/atomic"

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
