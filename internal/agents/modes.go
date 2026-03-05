package agents

import (
	"github.com/robinojw/dj/internal/modes"
)

// AgentMode is an alias for modes.ExecutionMode for backward compatibility.
type AgentMode = modes.ExecutionMode

// Re-export mode constants.
const (
	ModePlan    = modes.ModePlan
	ModeConfirm = modes.ModeConfirm
	ModeTurbo   = modes.ModeTurbo
)

// ModeBuild is an alias for ModeConfirm for backward compatibility.
const ModeBuild = ModeConfirm

// ModeConfig is an alias for modes.ModeConfig.
type ModeConfig = modes.ModeConfig

// Modes is now sourced from the modes package.
var Modes = modes.Modes

// FilterTools returns only the tools allowed by the given mode config.
// If cfg.AllowedTools is nil, all tools are returned.
func FilterTools(allTools []string, cfg modes.ModeConfig) []string {
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
