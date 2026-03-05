package modes

// ExecutionMode determines tool access and permission behavior.
type ExecutionMode int

const (
	ModeConfirm ExecutionMode = iota // ask before write/exec/MCP tools
	ModePlan                          // read-only, high reasoning
	ModeTurbo                         // bypass all permissions
)

func (m ExecutionMode) String() string {
	switch m {
	case ModeConfirm:
		return "Confirm"
	case ModePlan:
		return "Plan"
	case ModeTurbo:
		return "Turbo"
	default:
		return "Unknown"
	}
}

// StatusLabel returns the badge text for the status bar.
func (m ExecutionMode) StatusLabel() string {
	switch m {
	case ModeConfirm:
		return "⏸ CONFIRM"
	case ModePlan:
		return "◎ PLAN"
	case ModeTurbo:
		return "⚡ TURBO"
	default:
		return "UNKNOWN"
	}
}

// ToolClass categorizes tools for permission decisions.
type ToolClass int

const (
	ToolRead      ToolClass = iota // read_file, list_dir, search_code
	ToolWrite                      // write_file, create_file, delete_file
	ToolExec                       // bash, run_script, run_tests
	ToolMCPMutate                  // MCP tools that modify state
	ToolMCPRead                    // MCP tools flagged read-only
	ToolNetwork                    // web_fetch, http_request
)

// GateDecision is the result of gate evaluation.
type GateDecision int

const (
	GateAllow   GateDecision = iota // execute immediately
	GateDeny                         // block execution
	GateAskUser                      // show permission modal
)

// ModeConfig holds the system prompt and reasoning effort for a mode.
type ModeConfig struct {
	Mode            ExecutionMode
	AllowedTools    []string // nil = all tools (Turbo/Confirm), specific list for Plan
	SystemPrompt    string
	ReasoningEffort string
}

var planSystemPrompt = `You are a senior software architect in Plan mode.
Your job is to analyze, reason, and produce a detailed implementation plan.
You may ONLY read files, search code, and list directories.
You may NOT write files, execute commands, or invoke MCP tools that mutate state.
Think deeply. Output a numbered, step-by-step implementation plan with exact file paths.`

var confirmSystemPrompt = `You are a skilled software engineer in Confirm mode.
You have access to all tools, but destructive operations require user permission.
Make focused, incremental changes. Explain your intent before executing risky operations.
Run tests after edits to verify correctness.`

var turboSystemPrompt = `You are a skilled software engineer in Turbo mode.
You have full autonomy - all tools are available without permission checks.
Work efficiently. Make minimal, focused changes. Run tests after each edit.`

// Modes maps each ExecutionMode to its configuration.
var Modes = map[ExecutionMode]ModeConfig{
	ModePlan: {
		Mode:            ModePlan,
		AllowedTools:    []string{"read_file", "search_code", "list_dir"},
		SystemPrompt:    planSystemPrompt,
		ReasoningEffort: "high",
	},
	ModeConfirm: {
		Mode:            ModeConfirm,
		AllowedTools:    nil, // all tools available
		SystemPrompt:    confirmSystemPrompt,
		ReasoningEffort: "medium",
	},
	ModeTurbo: {
		Mode:            ModeTurbo,
		AllowedTools:    nil, // all tools available
		SystemPrompt:    turboSystemPrompt,
		ReasoningEffort: "medium",
	},
}
