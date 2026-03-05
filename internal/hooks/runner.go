package hooks

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const hookTimeout = 10 * time.Second

// HookEvent identifies a lifecycle point.
type HookEvent string

const (
	HookPreToolCall  HookEvent = "pre_tool_call"
	HookPostToolCall HookEvent = "post_tool_call"
	HookOnError      HookEvent = "on_error"
	HookSessionEnd   HookEvent = "on_session_end"
)

// Config holds hook shell command templates.
type Config struct {
	Hooks map[string]string
}

// Runner executes configured hooks at lifecycle points.
type Runner struct {
	config Config
}

func NewRunner(cfg Config) *Runner {
	return &Runner{config: cfg}
}

// Fire executes the hook for the given event with variable substitution.
// Returns nil if no hook is configured for the event.
func (r *Runner) Fire(event HookEvent, vars map[string]string) error {
	cmdTemplate, ok := r.config.Hooks[string(event)]
	if !ok || cmdTemplate == "" {
		return nil
	}

	expanded := expandVars(cmdTemplate, vars)

	cmd := exec.Command("sh", "-c", expanded)
	cmd.WaitDelay = hookTimeout

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("hook %s failed: %w\noutput: %s", event, err, string(output))
	}

	return nil
}

// expandVars replaces {{key}} placeholders with values from vars.
func expandVars(template string, vars map[string]string) string {
	result := template
	for k, v := range vars {
		result = strings.ReplaceAll(result, "{{"+k+"}}", v)
	}
	return result
}
