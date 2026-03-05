package modes

import (
	"os"
	"strings"
	"testing"
)

func TestPersistToConfig(t *testing.T) {
	// Create temp config file
	tmpfile, err := os.CreateTemp("", "harness-*.toml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Write initial config
	initial := `[execution]
default_mode = "confirm"

[execution.allow]
tools = ["read_file"]

[execution.deny]
tools = []
`
	if _, err := tmpfile.Write([]byte(initial)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	// Persist new tool
	if err := PersistToolToAllowList(tmpfile.Name(), "write_file"); err != nil {
		t.Fatalf("Failed to persist: %v", err)
	}

	// Read back and verify
	data, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "write_file") {
		t.Error("write_file not found in persisted config")
	}
}

func TestPersistToConfig_NoDuplicates(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "harness-*.toml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	initial := `[execution]
default_mode = "confirm"

[execution.allow]
tools = ["read_file"]

[execution.deny]
tools = []
`
	if _, err := tmpfile.Write([]byte(initial)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	// Persist same tool twice
	if err := PersistToolToAllowList(tmpfile.Name(), "read_file"); err != nil {
		t.Fatalf("Failed to persist: %v", err)
	}

	data, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Should only appear once
	count := strings.Count(string(data), "read_file")
	if count != 1 {
		t.Errorf("Expected read_file to appear once, appeared %d times", count)
	}
}
