package pool

import (
	"fmt"
	"strings"

	"github.com/robinojw/dj/internal/roster"
)

const orchestratorPreamble = "You are DJ's orchestrator. You coordinate a team of specialist agents to accomplish tasks."

const orchestratorFooter = "Analyze the user's request, decide which specialists to spawn, and coordinate their work."

const workerPromptFmt = "You are acting as the %s specialist.\n\n%s\n\nYour task: %s"

const fenceOpen = "```dj-command\n"

const fenceClose = "```\n"

func BuildWorkerPrompt(persona *roster.PersonaDefinition, task string) string {
	return fmt.Sprintf(workerPromptFmt, persona.Name, persona.Content, task)
}

func BuildOrchestratorPrompt(personas map[string]roster.PersonaDefinition, signals *roster.RepoSignals) string {
	var builder strings.Builder
	builder.WriteString(orchestratorPreamble)
	builder.WriteString("\n\nAvailable personas:\n")
	for id, persona := range personas {
		fmt.Fprintf(&builder, "- %s: %s\n", id, persona.Description)
	}
	if signals != nil {
		appendRepoContext(&builder, signals)
	}
	appendInstructions(&builder)
	builder.WriteString("\n")
	builder.WriteString(orchestratorFooter)
	return builder.String()
}

func appendRepoContext(builder *strings.Builder, signals *roster.RepoSignals) {
	builder.WriteString("\nRepo context:\n")
	fmt.Fprintf(builder, "Languages: %s\n", strings.Join(signals.Languages, ", "))
	if signals.CIProvider != "" {
		fmt.Fprintf(builder, "CI: %s\n", signals.CIProvider)
	}
	if signals.LintConfig != "" {
		fmt.Fprintf(builder, "Lint: %s\n", signals.LintConfig)
	}
}

func appendInstructions(builder *strings.Builder) {
	builder.WriteString("\nTo spawn an agent, emit a fenced code block:\n")
	builder.WriteString(fenceOpen)
	builder.WriteString("{\"action\":\"spawn\",\"persona\":\"<id>\",\"task\":\"<description>\"}\n")
	builder.WriteString(fenceClose)
	builder.WriteString("\nTo message an existing agent:\n")
	builder.WriteString(fenceOpen)
	builder.WriteString("{\"action\":\"message\",\"target\":\"<agent-id>\",\"content\":\"<message>\"}\n")
	builder.WriteString(fenceClose)
	builder.WriteString("\nWhen done coordinating, emit:\n")
	builder.WriteString(fenceOpen)
	builder.WriteString("{\"action\":\"complete\",\"content\":\"<summary>\"}\n")
	builder.WriteString(fenceClose)
}
