package skills

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const scriptTimeout = 30 * time.Second

// Executor runs skill scripts and captures output.
type Executor struct{}

func NewExecutor() *Executor {
	return &Executor{}
}

// RunScript executes a skill script and returns its stdout.
func (e *Executor) RunScript(ctx context.Context, script SkillScript) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, scriptTimeout)
	defer cancel()

	var cmd *exec.Cmd
	switch script.Language {
	case "bash":
		cmd = exec.CommandContext(ctx, "bash", script.Path)
	case "python":
		cmd = exec.CommandContext(ctx, "python3", script.Path)
	case "node":
		cmd = exec.CommandContext(ctx, "node", script.Path)
	default:
		return "", fmt.Errorf("unsupported script language: %s", script.Language)
	}

	// Set working directory to the skill's directory
	cmd.Dir = filepath.Dir(script.Path)
	cmd.Env = append(os.Environ(),
		"SKILL_DIR="+filepath.Dir(script.Path),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("script %s failed: %w\nstderr: %s",
			script.Filename, err, stderr.String())
	}

	return stdout.String(), nil
}

// RunAllScripts runs all scripts in a skill sequentially.
func (e *Executor) RunAllScripts(ctx context.Context, skill Skill) ([]ScriptResult, error) {
	var results []ScriptResult

	for _, script := range skill.Scripts {
		output, err := e.RunScript(ctx, script)
		results = append(results, ScriptResult{
			Script: script,
			Output: output,
			Err:    err,
		})
		if err != nil {
			return results, err
		}
	}

	return results, nil
}

// ScriptResult holds the output of a single script execution.
type ScriptResult struct {
	Script SkillScript
	Output string
	Err    error
}
