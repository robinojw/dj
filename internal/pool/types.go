package pool

import (
	"github.com/robinojw/dj/internal/appserver"
	"github.com/robinojw/dj/internal/orchestrator"
	"github.com/robinojw/dj/internal/roster"
)

const (
	RoleOrchestrator = "orchestrator"
	RoleWorker       = "worker"
)

const (
	AgentStatusSpawning  = "spawning"
	AgentStatusActive    = "active"
	AgentStatusCompleted = "completed"
	AgentStatusError     = "error"
)

type AgentProcess struct {
	ID        string
	PersonaID string
	ThreadID  string
	Client    *appserver.Client
	Role      string
	Task      string
	Status    string
	ParentID  string
	Persona   *roster.PersonaDefinition
	Parser    *orchestrator.CommandParser
}

type PoolEvent struct {
	AgentID string
	Message appserver.JSONRPCMessage
}
