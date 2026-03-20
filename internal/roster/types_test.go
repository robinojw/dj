package roster

import "testing"

const (
	testPersonaID          = "architect"
	testPersonaName        = "Architect"
	testPersonaDescription = "System architecture"
	testRepoName           = "myapp"
	expectedTriggerCount   = 2
	expectedLanguageCount  = 2
)

func TestPersonaDefinitionFields(testing *testing.T) {
	persona := PersonaDefinition{
		ID:          testPersonaID,
		Name:        testPersonaName,
		Description: testPersonaDescription,
		Triggers:    []string{"new service", "API boundary"},
		Content:     "## Principles\n\nFavour simplicity.",
	}

	if persona.ID != testPersonaID {
		testing.Errorf("expected ID %s, got %s", testPersonaID, persona.ID)
	}
	if len(persona.Triggers) != expectedTriggerCount {
		testing.Errorf("expected %d triggers, got %d", expectedTriggerCount, len(persona.Triggers))
	}
}

func TestRepoSignalsFields(testing *testing.T) {
	signals := RepoSignals{
		RepoName:  testRepoName,
		Languages: []string{"Go", "TypeScript"},
	}

	if signals.RepoName != testRepoName {
		testing.Errorf("expected %s, got %s", testRepoName, signals.RepoName)
	}
	if len(signals.Languages) != expectedLanguageCount {
		testing.Errorf("expected %d languages, got %d", expectedLanguageCount, len(signals.Languages))
	}
}
