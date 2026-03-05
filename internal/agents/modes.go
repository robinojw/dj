package agents

import "fmt"

// AgentMode determines the agent's permission level and persona.
type AgentMode int

const (
	ModeBuild AgentMode = iota // full tools: read, write, run, MCP
	ModePlan                   // read-only: no writes, no exec, no MCP mutations
)

func (m AgentMode) String() string {
	switch m {
	case ModeBuild:
		return "Build"
	case ModePlan:
		return "Plan"
	default:
		return fmt.Sprintf("Unknown(%d)", int(m))
	}
}

// ModeConfig holds the constraints and persona for a mode.
type ModeConfig struct {
	Mode            AgentMode
	AllowedTools    []string // nil = all tools allowed
	SystemPrompt    string
	ReasoningEffort string // "low", "medium", "high"
}

var planSystemPrompt = `You are a senior software architect in Plan mode.
Your job is to analyze, reason, and produce a detailed implementation plan.
You may ONLY read files, search code, and list directories.
You may NOT write files, execute commands, or invoke MCP tools that mutate state.
Think deeply. Output a numbered, step-by-step implementation plan with exact file paths.`

var buildSystemPrompt = `You are a skilled software engineer in Build mode.
You have full access to all tools: read, write, search, execute, and MCP.
Follow the plan precisely. Make minimal, focused changes. Run tests after each edit.`

// Modes maps each AgentMode to its configuration.
var Modes = map[AgentMode]ModeConfig{
	ModePlan: {
		Mode:            ModePlan,
		AllowedTools:    []string{"read_file", "search_code", "list_dir"},
		SystemPrompt:    planSystemPrompt,
		ReasoningEffort: "high",
	},
	ModeBuild: {
		Mode:            ModeBuild,
		AllowedTools:    nil, // all tools enabled
		SystemPrompt:    buildSystemPrompt,
		ReasoningEffort: "medium",
	},
}

// FilterTools returns only the tools allowed by the given mode config.
// If cfg.AllowedTools is nil, all tools are returned.
func FilterTools(allTools []string, cfg ModeConfig) []string {
	if cfg.AllowedTools == nil {
		return allTools
	}
	allowed := make(map[string]bool, len(cfg.AllowedTools))
	for _, t := range cfg.AllowedTools {
		allowed[t] = true
	}
	var filtered []string
	for _, t := range allTools {
		if allowed[t] {
			filtered = append(filtered, t)
		}
	}
	return filtered
}
