package modes

// ExecutionMode determines tool access and permission behavior.
type ExecutionMode int

const (
	ModeConfirm ExecutionMode = iota
	ModePlan
	ModeTurbo
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
	ToolRead      ToolClass = iota
	ToolWrite
	ToolExec
	ToolMCPMutate
	ToolMCPRead
	ToolNetwork
)

var toolClasses = map[string]ToolClass{
	"read_file":    ToolRead,
	"list_dir":     ToolRead,
	"search_code":  ToolRead,
	"write_file":   ToolWrite,
	"create_file":  ToolWrite,
	"delete_file":  ToolWrite,
	"bash":         ToolExec,
	"run_script":   ToolExec,
	"run_tests":    ToolExec,
	"web_fetch":    ToolNetwork,
	"http_request": ToolNetwork,
}

// ClassifyTool returns the security class of a tool.
// Unknown tools default to ToolWrite (conservative).
func ClassifyTool(toolName string) ToolClass {
	if class, ok := toolClasses[toolName]; ok {
		return class
	}
	return ToolWrite
}

// ToolClassifier provides annotation data for tool classification.
// Implemented by tools.ToolRegistry.
type ToolClassifier interface {
	ToolAnnotations(name string) (readOnly, destructive, mutatesFiles bool, ok bool)
}

// ClassifyToolWithRegistry consults registry annotations before the static map.
func ClassifyToolWithRegistry(toolName string, registry ToolClassifier) ToolClass {
	if registry != nil {
		readOnly, destructive, mutatesFiles, ok := registry.ToolAnnotations(toolName)
		if ok {
			if readOnly {
				return ToolRead
			}
			if destructive || mutatesFiles {
				return ToolWrite
			}
		}
	}
	return ClassifyTool(toolName)
}

// GateDecision is the result of gate evaluation.
type GateDecision int

const (
	GateAllow   GateDecision = iota
	GateDeny
	GateAskUser
)

// ModeConfig holds the system prompt and reasoning effort for a mode.
type ModeConfig struct {
	Mode            ExecutionMode
	AllowedTools    []string
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
		AllowedTools:    nil,
		SystemPrompt:    confirmSystemPrompt,
		ReasoningEffort: "medium",
	},
	ModeTurbo: {
		Mode:            ModeTurbo,
		AllowedTools:    nil,
		SystemPrompt:    turboSystemPrompt,
		ReasoningEffort: "medium",
	},
}
