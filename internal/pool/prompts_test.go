package pool

import (
	"strings"
	"testing"

	"github.com/robinojw/dj/internal/roster"
)

const (
	testPromptPersonaName    = "Architect"
	testPromptPersonaContent = "You design systems."
	testPromptTask           = "Design the API"
	testPromptPersonaDesc    = "System design"
	testPromptLanguage       = "Go"
	testPromptCI             = "GitHub Actions"
	testPromptLint           = "golangci-lint"
)

func TestBuildWorkerPrompt(testing *testing.T) {
	persona := &roster.PersonaDefinition{
		Name:    testPromptPersonaName,
		Content: testPromptPersonaContent,
	}
	prompt := BuildWorkerPrompt(persona, testPromptTask)

	hasName := strings.Contains(prompt, testPromptPersonaName)
	if !hasName {
		testing.Error("expected prompt to contain persona name")
	}
	hasContent := strings.Contains(prompt, testPromptPersonaContent)
	if !hasContent {
		testing.Error("expected prompt to contain persona content")
	}
	hasTask := strings.Contains(prompt, testPromptTask)
	if !hasTask {
		testing.Error("expected prompt to contain task")
	}
}

func TestBuildOrchestratorPrompt(testing *testing.T) {
	personas := map[string]roster.PersonaDefinition{
		testPersonaArchID: {
			ID:          testPersonaArchID,
			Name:        testPromptPersonaName,
			Description: testPromptPersonaDesc,
		},
	}
	signals := &roster.RepoSignals{
		Languages:  []string{testPromptLanguage},
		CIProvider: testPromptCI,
		LintConfig: testPromptLint,
	}
	prompt := BuildOrchestratorPrompt(personas, signals)

	hasPreamble := strings.Contains(prompt, orchestratorPreamble)
	if !hasPreamble {
		testing.Error("expected prompt to contain preamble")
	}
	hasPersona := strings.Contains(prompt, testPromptPersonaDesc)
	if !hasPersona {
		testing.Error("expected prompt to contain persona description")
	}
	hasLanguage := strings.Contains(prompt, testPromptLanguage)
	if !hasLanguage {
		testing.Error("expected prompt to contain language")
	}
	hasInstructions := strings.Contains(prompt, "dj-command")
	if !hasInstructions {
		testing.Error("expected prompt to contain dj-command instructions")
	}
}

func TestBuildOrchestratorPromptNilSignals(testing *testing.T) {
	personas := map[string]roster.PersonaDefinition{
		testPersonaTestID: {
			ID:          testPersonaTestID,
			Name:        testPersonaTestName,
			Description: "Testing",
		},
	}
	prompt := BuildOrchestratorPrompt(personas, nil)

	hasNoRepoContext := !strings.Contains(prompt, "Repo context")
	if !hasNoRepoContext {
		testing.Error("expected no repo context when signals are nil")
	}
}
