package agents

import (
	"time"

	"github.com/robinojw/dj/internal/hooks"
)

// WorkerUpdate is a message sent from a worker goroutine to the orchestrator.
type WorkerUpdate struct {
	WorkerID string
	Type     UpdateType
	Content  string
	Error    error
	Usage    UsageInfo
	DiffInfo   *DiffInfo         // populated when Type == UpdateDiffResult
	HookResult *hooks.HookResult // populated when Type == UpdateHookResult
}

type UpdateType int

const (
	UpdateDelta      UpdateType = iota // streaming text delta
	UpdateToolCall                     // tool call started
	UpdateToolResult                   // tool call completed
	UpdateDiffResult                   // diff result from git
	UpdateCompleted                    // worker finished
	UpdateError                        // worker encountered an error
	UpdateSkipped                      // worker skipped due to failed dependency
	UpdateHookResult                   // hook execution result
)

type UsageInfo struct {
	InputTokens  int
	OutputTokens int
}

// DiffInfo contains git diff output for a file edit operation.
type DiffInfo struct {
	FilePath  string
	DiffText  string
	Timestamp time.Time
}

// TurnKind identifies the type of a session turn.
type TurnKind int

const (
	TurnText       TurnKind = iota
	TurnToolCall
	TurnToolResult
	TurnDiff
	TurnError
)

// SessionTurn is one unit of output in a worker's session.
type SessionTurn struct {
	Kind      TurnKind
	Content   string
	ToolName  string
	Timestamp time.Time
}

// WorkerSession stores the full message history for a single worker agent.
type WorkerSession struct {
	WorkerID string
	Turns    []SessionTurn
}

// Subtask is a unit of work assigned to a worker agent.
type Subtask struct {
	ID          string
	Description string
	DependsOn   []string // IDs of subtasks that must complete first
	Files       []string // scoped files/paths
}

// TaskAnalysis is the result of the task router's complexity analysis.
type TaskAnalysis struct {
	Subtasks       []Subtask `json:"subtasks"`
	Complexity     string    `json:"complexity"`     // "low", "medium", "high"
	Parallelizable bool     `json:"parallelizable"`
}
