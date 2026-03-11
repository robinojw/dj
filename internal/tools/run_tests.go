package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// RunTestsHandler returns a ToolHandler that wraps `go test ./...` and returns structured output.
// NOTE: This is Go-specific. Other language test runners would need separate handlers.
// workspaceRoot is used as the working directory for the test command.
func RunTestsHandler(workspaceRoot string) ToolHandler {
	return func(ctx context.Context, args map[string]any) (string, error) {
		// Allow optional package pattern override
		pkg := "./..."
		if p, ok := stringArg(args, "package"); ok {
			pkg = p
		}

		cmdArgs := []string{"test", pkg, "-v", "-count=1"}

		// Allow optional -run filter
		if run, ok := stringArg(args, "run"); ok {
			cmdArgs = append(cmdArgs, "-run", run)
		}

		cmd := exec.CommandContext(ctx, "go", cmdArgs...)
		cmd.Dir = workspaceRoot

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()

		var sb strings.Builder
		sb.WriteString("## Test Results\n\n")

		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				fmt.Fprintf(&sb, "**Status:** FAIL (exit code %d)\n\n", exitErr.ExitCode())
			} else {
				fmt.Fprintf(&sb, "**Status:** ERROR (%v)\n\n", err)
			}
		} else {
			sb.WriteString("**Status:** PASS\n\n")
		}

		if stdout.Len() > 0 {
			sb.WriteString("### Output\n```\n")
			sb.WriteString(stdout.String())
			sb.WriteString("```\n")
		}

		if stderr.Len() > 0 {
			sb.WriteString("### Errors\n```\n")
			sb.WriteString(stderr.String())
			sb.WriteString("```\n")
		}

		return sb.String(), nil
	}
}
