package agents

// WorkerUpdate is a message sent from a worker goroutine to the orchestrator.
type WorkerUpdate struct {
	WorkerID string
	Type     UpdateType
	Content  string
	Error    error
	Usage    UsageInfo
}

type UpdateType int

const (
	UpdateDelta     UpdateType = iota // streaming text delta
	UpdateToolCall                    // tool call started
	UpdateToolResult                  // tool call completed
	UpdateCompleted                   // worker finished
	UpdateError                       // worker encountered an error
)

type UsageInfo struct {
	InputTokens  int
	OutputTokens int
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
