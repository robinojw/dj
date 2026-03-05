package mentions

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const resolveTimeout = 10 * time.Second

// Resolve fetches the content for each mention.
func Resolve(ctx context.Context, mentions []Mention) []ResolvedMention {
	resolved := make([]ResolvedMention, len(mentions))

	for i, m := range mentions {
		resolved[i] = ResolvedMention{Mention: m}

		timeoutCtx, cancel := context.WithTimeout(ctx, resolveTimeout)
		switch m.Type {
		case MentionFile:
			resolved[i].Content, resolved[i].Error = resolveFile(m.Value)
		case MentionURL:
			resolved[i].Content, resolved[i].Error = resolveURL(timeoutCtx, m.Value)
		case MentionFunction:
			resolved[i].Content, resolved[i].Error = resolveFunction(m.Value)
		case MentionGit:
			resolved[i].Content, resolved[i].Error = resolveGit(timeoutCtx, m.Value)
		case MentionTest:
			resolved[i].Content, resolved[i].Error = resolveTest(timeoutCtx, m.Value)
		}
		cancel()
	}

	return resolved
}

// FormatResolved builds a context string to inject into the prompt.
func FormatResolved(resolved []ResolvedMention) string {
	var sb strings.Builder
	for _, r := range resolved {
		if r.Error != nil {
			sb.WriteString(fmt.Sprintf("\n[%s %s: error: %v]\n", r.Type, r.Value, r.Error))
			continue
		}
		sb.WriteString(fmt.Sprintf("\n--- %s: %s ---\n%s\n", r.Type, r.Value, truncate(r.Content, 8000)))
	}
	return sb.String()
}

func resolveFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	return string(data), nil
}

func resolveURL(ctx context.Context, url string) (string, error) {
	// Shell out to curl for simplicity — avoids importing net/http for one-off fetches
	cmd := exec.CommandContext(ctx, "curl", "-sL", "--max-time", "10", url)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", url, err)
	}
	return string(output), nil
}

func resolveFunction(name string) (string, error) {
	// Grep for function definition in current directory
	cmd := exec.Command("grep", "-rn", fmt.Sprintf("func.*%s", name), ".")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("function %s not found", name)
	}
	return string(output), nil
}

func resolveGit(ctx context.Context, ref string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "diff", ref)
	output, err := cmd.Output()
	if err != nil {
		// Try git show as fallback
		cmd = exec.CommandContext(ctx, "git", "show", ref)
		output, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("git ref %s: %w", ref, err)
		}
	}
	return string(output), nil
}

func resolveTest(ctx context.Context, name string) (string, error) {
	// Try Go test first
	cmd := exec.CommandContext(ctx, "go", "test", "-run", name, "-v", "./...")
	output, err := cmd.CombinedOutput()
	if err == nil {
		return string(output), nil
	}

	// Try npm test as fallback
	cmd = exec.CommandContext(ctx, "npx", "vitest", "run", "-t", name)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("test %s failed: %w\n%s", name, err, string(output))
	}
	return string(output), nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "\n... [truncated]"
}
