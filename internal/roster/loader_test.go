package roster

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	testSignalsRepoName    = "myapp"
	testSignalsCIProvider  = "GitHub Actions"
	expectedPersonaCount   = 1
	expectedSignalLangCount = 1

	errUnexpected    = "unexpected error: %v"
	errExpectedSGotS = "expected %s, got %s"

	permDir  = 0o755
	permFile = 0o644
)

func TestLoadPersonas(testing *testing.T) {
	dir := testing.TempDir()
	personaDir := filepath.Join(dir, "personas")
	os.MkdirAll(personaDir, permDir)

	content := "---\nid: architect\nname: Architect\ndescription: System architecture\ntriggers:\n  - new service\n  - API boundary\n---\n\n## Principles\n\nFavour simplicity."
	os.WriteFile(filepath.Join(personaDir, "architect.md"), []byte(content), permFile)

	personas, err := LoadPersonas(personaDir)
	if err != nil {
		testing.Fatalf(errUnexpected, err)
	}
	if len(personas) != expectedPersonaCount {
		testing.Fatalf("expected %d persona, got %d", expectedPersonaCount, len(personas))
	}
	if personas[0].ID != testPersonaID {
		testing.Errorf(errExpectedSGotS, testPersonaID, personas[0].ID)
	}
	if personas[0].Name != testPersonaName {
		testing.Errorf(errExpectedSGotS, testPersonaName, personas[0].Name)
	}
	if len(personas[0].Triggers) != expectedTriggerCount {
		testing.Errorf("expected %d triggers, got %d", expectedTriggerCount, len(personas[0].Triggers))
	}
	hasContent := personas[0].Content != ""
	if !hasContent {
		testing.Error("expected non-empty content")
	}
}

func TestLoadPersonasEmptyDir(testing *testing.T) {
	dir := testing.TempDir()
	personas, err := LoadPersonas(dir)
	if err != nil {
		testing.Fatalf(errUnexpected, err)
	}
	if len(personas) != 0 {
		testing.Errorf("expected 0 personas, got %d", len(personas))
	}
}

func TestLoadPersonasMissingDir(testing *testing.T) {
	_, err := LoadPersonas("/nonexistent/path")
	if err == nil {
		testing.Error("expected error for missing directory")
	}
}

func TestLoadSignals(testing *testing.T) {
	dir := testing.TempDir()
	signalsJSON := `{"repo_name":"myapp","languages":["Go"],"frameworks":[],"ci_provider":"GitHub Actions","file_count":50}`
	path := filepath.Join(dir, "signals.json")
	os.WriteFile(path, []byte(signalsJSON), permFile)

	signals, err := LoadSignals(path)
	if err != nil {
		testing.Fatalf(errUnexpected, err)
	}
	if signals.RepoName != testSignalsRepoName {
		testing.Errorf(errExpectedSGotS, testSignalsRepoName, signals.RepoName)
	}
	if len(signals.Languages) != expectedSignalLangCount {
		testing.Errorf("expected %d language, got %d", expectedSignalLangCount, len(signals.Languages))
	}
	if signals.CIProvider != testSignalsCIProvider {
		testing.Errorf(errExpectedSGotS, testSignalsCIProvider, signals.CIProvider)
	}
}

func TestLoadSignalsMissingFile(testing *testing.T) {
	_, err := LoadSignals("/nonexistent/signals.json")
	if err == nil {
		testing.Error("expected error for missing file")
	}
}
