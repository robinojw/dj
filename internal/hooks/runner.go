package hooks

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

const defaultTimeout = 10 * time.Second

// HookEvent identifies a lifecycle point.
type HookEvent string

const (
	HookPreToolCall  HookEvent = "pre_tool_call"
	HookPostToolCall HookEvent = "post_tool_call"
	HookOnError      HookEvent = "on_error"
	HookSessionEnd   HookEvent = "on_session_end"
)

// HookResult captures the outcome of a hook execution.
type HookResult struct {
	Event    HookEvent
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
	Err      error // non-nil only for infrastructure failures (sh not found, timeout)
}

// Config holds hook shell command templates.
type Config struct {
	Hooks   map[string]string
	Timeout time.Duration // 0 means use defaultTimeout
}

func (c Config) timeout() time.Duration {
	if c.Timeout > 0 {
		return c.Timeout
	}
	return defaultTimeout
}

// Runner executes configured hooks at lifecycle points.
type Runner struct {
	config Config
}

func NewRunner(cfg Config) *Runner {
	return &Runner{config: cfg}
}

// Fire executes the hook for the given event with variable substitution.
// Returns (nil, nil) if no hook is configured for the event.
// Returns (*HookResult, nil) on both success and non-zero exit.
// The error return is reserved for infrastructure failures.
func (r *Runner) Fire(event HookEvent, vars map[string]string) (*HookResult, error) {
	cmdTemplate, ok := r.config.Hooks[string(event)]
	if !ok || cmdTemplate == "" {
		return nil, nil
	}

	expanded := expandVars(cmdTemplate, vars)

	cmd := exec.Command("sh", "-c", expanded)
	cmd.WaitDelay = r.config.timeout()

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	result := &HookResult{
		Event:    event,
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		Duration: duration,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				result.ExitCode = status.ExitStatus()
			} else {
				result.ExitCode = 1
			}
		} else {
			// Infrastructure failure (e.g., sh not found, timeout)
			result.Err = fmt.Errorf("hook %s failed: %w", event, err)
			return result, result.Err
		}
	}

	return result, nil
}

// FireAsync launches the hook in a background goroutine (fire-and-forget).
// Used for on_session_end where blocking is undesirable.
func (r *Runner) FireAsync(event HookEvent, vars map[string]string) {
	cmdTemplate, ok := r.config.Hooks[string(event)]
	if !ok || cmdTemplate == "" {
		return
	}

	expanded := expandVars(cmdTemplate, vars)

	go func() {
		cmd := exec.Command("sh", "-c", expanded)
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
		cmd.WaitDelay = r.config.timeout()
		_ = cmd.Run()
	}()
}

// expandVars replaces {{key}} placeholders with values from vars.
func expandVars(template string, vars map[string]string) string {
	result := template
	for k, v := range vars {
		result = strings.ReplaceAll(result, "{{"+k+"}}", v)
	}
	return result
}
