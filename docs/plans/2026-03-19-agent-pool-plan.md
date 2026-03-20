# Agent Pool Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Transform DJ from a single-process Codex visualizer into a multi-process agent swarm orchestrator that reads roster personas and dynamically spawns, routes, and visualizes persona-typed Codex agents.

**Architecture:** DJ reads `.roster/personas/*.md` and `.roster/signals.json` at startup. An AgentPool manages multiple Codex processes. An orchestrator Codex session analyzes tasks and emits structured `dj-command` blocks that DJ parses from delta streams to spawn workers, route messages, and coordinate completions.

**Tech Stack:** Go 1.25, Bubble Tea, Lipgloss, JSON-RPC 2.0 over JSONL, YAML v3 for frontmatter parsing.

**Design doc:** `docs/plans/2026-03-19-agent-pool-design.md`

---

### Task 1: Roster Persona Loader — Types

**Files:**
- Create: `internal/roster/types.go`
- Test: `internal/roster/types_test.go`

**Step 1: Write the failing test**

```go
package roster

import "testing"

func TestPersonaDefinitionFields(t *testing.T) {
	persona := PersonaDefinition{
		ID:          "architect",
		Name:        "Architect",
		Description: "System architecture",
		Triggers:    []string{"new service", "API boundary"},
		Content:     "## Principles\n\nFavour simplicity.",
	}

	if persona.ID != "architect" {
		t.Errorf("expected ID architect, got %s", persona.ID)
	}
	if len(persona.Triggers) != 2 {
		t.Errorf("expected 2 triggers, got %d", len(persona.Triggers))
	}
}

func TestRepoSignalsFields(t *testing.T) {
	signals := RepoSignals{
		RepoName:  "myapp",
		Languages: []string{"Go", "TypeScript"},
	}

	if signals.RepoName != "myapp" {
		t.Errorf("expected myapp, got %s", signals.RepoName)
	}
	if len(signals.Languages) != 2 {
		t.Errorf("expected 2 languages, got %d", len(signals.Languages))
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/roster/ -run TestPersonaDefinition -v`
Expected: FAIL — package does not exist

**Step 3: Write minimal implementation**

```go
package roster

type PersonaDefinition struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Triggers    []string `yaml:"triggers"`
	Content     string   `yaml:"-"`
}

type RepoSignals struct {
	RepoName    string   `json:"repo_name"`
	Languages   []string `json:"languages"`
	Frameworks  []string `json:"frameworks"`
	CIProvider  string   `json:"ci_provider,omitempty"`
	LintConfig  string   `json:"lint_config,omitempty"`
	IsMonorepo  bool     `json:"is_monorepo"`
	HasDocker   bool     `json:"has_docker"`
	HasE2E      bool     `json:"has_e2e"`
	FileCount   int      `json:"file_count"`
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/roster/ -v -race`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/roster/types.go internal/roster/types_test.go
git commit -m "feat(roster): add PersonaDefinition and RepoSignals types"
```

---

### Task 2: Roster Persona Loader — Parse Frontmatter

**Files:**
- Create: `internal/roster/loader.go`
- Test: `internal/roster/loader_test.go`

**Step 1: Write the failing test**

```go
package roster

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPersonas(t *testing.T) {
	dir := t.TempDir()
	personaDir := filepath.Join(dir, "personas")
	os.MkdirAll(personaDir, 0o755)

	content := "---\nid: architect\nname: Architect\ndescription: System architecture\ntriggers:\n  - new service\n  - API boundary\n---\n\n## Principles\n\nFavour simplicity."
	os.WriteFile(filepath.Join(personaDir, "architect.md"), []byte(content), 0o644)

	personas, err := LoadPersonas(personaDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(personas) != 1 {
		t.Fatalf("expected 1 persona, got %d", len(personas))
	}
	if personas[0].ID != "architect" {
		t.Errorf("expected ID architect, got %s", personas[0].ID)
	}
	if personas[0].Name != "Architect" {
		t.Errorf("expected name Architect, got %s", personas[0].Name)
	}
	if len(personas[0].Triggers) != 2 {
		t.Errorf("expected 2 triggers, got %d", len(personas[0].Triggers))
	}
	hasContent := personas[0].Content != ""
	if !hasContent {
		t.Error("expected non-empty content")
	}
}

func TestLoadPersonasEmptyDir(t *testing.T) {
	dir := t.TempDir()
	personas, err := LoadPersonas(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(personas) != 0 {
		t.Errorf("expected 0 personas, got %d", len(personas))
	}
}

func TestLoadPersonasMissingDir(t *testing.T) {
	_, err := LoadPersonas("/nonexistent/path")
	if err == nil {
		t.Error("expected error for missing directory")
	}
}

func TestLoadSignals(t *testing.T) {
	dir := t.TempDir()
	signalsJSON := `{"repo_name":"myapp","languages":["Go"],"frameworks":[],"ci_provider":"GitHub Actions","file_count":50}`
	path := filepath.Join(dir, "signals.json")
	os.WriteFile(path, []byte(signalsJSON), 0o644)

	signals, err := LoadSignals(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if signals.RepoName != "myapp" {
		t.Errorf("expected myapp, got %s", signals.RepoName)
	}
	if len(signals.Languages) != 1 {
		t.Errorf("expected 1 language, got %d", len(signals.Languages))
	}
	if signals.CIProvider != "GitHub Actions" {
		t.Errorf("expected GitHub Actions, got %s", signals.CIProvider)
	}
}

func TestLoadSignalsMissingFile(t *testing.T) {
	_, err := LoadSignals("/nonexistent/signals.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/roster/ -run TestLoad -v`
Expected: FAIL — LoadPersonas and LoadSignals not defined

**Step 3: Write minimal implementation**

```go
package roster

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

const frontmatterDelimiter = "---"

func LoadPersonas(dir string) ([]PersonaDefinition, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read persona dir: %w", err)
	}

	var personas []PersonaDefinition
	for _, entry := range entries {
		isMarkdown := !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md")
		if !isMarkdown {
			continue
		}
		persona, err := loadPersonaFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("load persona %s: %w", entry.Name(), err)
		}
		personas = append(personas, persona)
	}
	return personas, nil
}

func loadPersonaFile(path string) (PersonaDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return PersonaDefinition{}, fmt.Errorf("read file: %w", err)
	}

	frontmatter, body, err := splitFrontmatter(string(data))
	if err != nil {
		return PersonaDefinition{}, fmt.Errorf("parse frontmatter: %w", err)
	}

	var persona PersonaDefinition
	if err := yaml.Unmarshal([]byte(frontmatter), &persona); err != nil {
		return PersonaDefinition{}, fmt.Errorf("unmarshal frontmatter: %w", err)
	}
	persona.Content = strings.TrimSpace(body)
	return persona, nil
}

func splitFrontmatter(content string) (string, string, error) {
	trimmed := strings.TrimSpace(content)
	if !strings.HasPrefix(trimmed, frontmatterDelimiter) {
		return "", "", fmt.Errorf("missing opening frontmatter delimiter")
	}
	rest := trimmed[len(frontmatterDelimiter):]
	endIndex := strings.Index(rest, "\n"+frontmatterDelimiter)
	if endIndex == -1 {
		return "", "", fmt.Errorf("missing closing frontmatter delimiter")
	}
	frontmatter := rest[:endIndex]
	body := rest[endIndex+len("\n"+frontmatterDelimiter):]
	return frontmatter, body, nil
}

func LoadSignals(path string) (*RepoSignals, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read signals file: %w", err)
	}

	var signals RepoSignals
	if err := json.Unmarshal(data, &signals); err != nil {
		return nil, fmt.Errorf("unmarshal signals: %w", err)
	}
	return &signals, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/roster/ -v -race`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/roster/loader.go internal/roster/loader_test.go
git commit -m "feat(roster): add persona and signals file loaders"
```

---

### Task 3: Orchestrator Command Parser — Types and Parsing

**Files:**
- Create: `internal/orchestrator/command.go`
- Test: `internal/orchestrator/command_test.go`

**Step 1: Write the failing test**

```go
package orchestrator

import "testing"

func TestCommandParserSingleBlock(t *testing.T) {
	parser := NewCommandParser()
	parser.Feed("Some text before\n```dj-command\n")
	parser.Feed(`{"action":"spawn","persona":"architect","task":"Design API"}`)
	parser.Feed("\n```\nSome text after")

	commands := parser.Flush()
	if len(commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(commands))
	}
	if commands[0].Action != "spawn" {
		t.Errorf("expected spawn, got %s", commands[0].Action)
	}
	if commands[0].Persona != "architect" {
		t.Errorf("expected architect, got %s", commands[0].Persona)
	}
	if commands[0].Task != "Design API" {
		t.Errorf("expected Design API, got %s", commands[0].Task)
	}
}

func TestCommandParserMultipleBlocks(t *testing.T) {
	parser := NewCommandParser()
	parser.Feed("```dj-command\n{\"action\":\"spawn\",\"persona\":\"architect\",\"task\":\"A\"}\n```\n")
	parser.Feed("```dj-command\n{\"action\":\"spawn\",\"persona\":\"test\",\"task\":\"B\"}\n```\n")

	commands := parser.Flush()
	if len(commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(commands))
	}
	if commands[0].Persona != "architect" {
		t.Errorf("expected architect, got %s", commands[0].Persona)
	}
	if commands[1].Persona != "test" {
		t.Errorf("expected test, got %s", commands[1].Persona)
	}
}

func TestCommandParserChunkedDelta(t *testing.T) {
	parser := NewCommandParser()
	parser.Feed("```dj-")
	parser.Feed("command\n{\"action\":")
	parser.Feed("\"message\",\"target\":\"arch-1\"")
	parser.Feed(",\"content\":\"hello\"}\n`")
	parser.Feed("``\n")

	commands := parser.Flush()
	if len(commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(commands))
	}
	if commands[0].Action != "message" {
		t.Errorf("expected message, got %s", commands[0].Action)
	}
	if commands[0].Target != "arch-1" {
		t.Errorf("expected arch-1, got %s", commands[0].Target)
	}
}

func TestCommandParserNoCommands(t *testing.T) {
	parser := NewCommandParser()
	parser.Feed("Just regular text with no commands at all.")

	commands := parser.Flush()
	if len(commands) != 0 {
		t.Errorf("expected 0 commands, got %d", len(commands))
	}
}

func TestCommandParserMalformedJSON(t *testing.T) {
	parser := NewCommandParser()
	parser.Feed("```dj-command\n{invalid json}\n```\n")

	commands := parser.Flush()
	if len(commands) != 0 {
		t.Errorf("expected 0 commands for malformed JSON, got %d", len(commands))
	}
}

func TestCommandParserStripsCommands(t *testing.T) {
	parser := NewCommandParser()
	parser.Feed("Before\n```dj-command\n{\"action\":\"complete\",\"content\":\"done\"}\n```\nAfter")

	_ = parser.Flush()
	cleaned := parser.CleanedText()
	if cleaned != "Before\nAfter" {
		t.Errorf("expected 'Before\\nAfter', got %q", cleaned)
	}
}

func TestCommandParserCompleteAction(t *testing.T) {
	parser := NewCommandParser()
	parser.Feed("```dj-command\n{\"action\":\"complete\",\"content\":\"Task finished with 2 findings\"}\n```\n")

	commands := parser.Flush()
	if len(commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(commands))
	}
	if commands[0].Action != "complete" {
		t.Errorf("expected complete, got %s", commands[0].Action)
	}
	if commands[0].Content != "Task finished with 2 findings" {
		t.Errorf("unexpected content: %s", commands[0].Content)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/orchestrator/ -v`
Expected: FAIL — package does not exist

**Step 3: Write minimal implementation**

```go
package orchestrator

import (
	"encoding/json"
	"strings"
)

const (
	fenceOpen  = "```dj-command\n"
	fenceClose = "\n```"
)

type DJCommand struct {
	Action  string `json:"action"`
	Persona string `json:"persona,omitempty"`
	Task    string `json:"task,omitempty"`
	Target  string `json:"target,omitempty"`
	Content string `json:"content,omitempty"`
}

type CommandParser struct {
	buffer      strings.Builder
	commands    []DJCommand
	cleanedText strings.Builder
}

func NewCommandParser() *CommandParser {
	return &CommandParser{}
}

func (parser *CommandParser) Feed(delta string) {
	parser.buffer.WriteString(delta)
}

func (parser *CommandParser) Flush() []DJCommand {
	parser.commands = nil
	parser.cleanedText.Reset()

	text := parser.buffer.String()
	parser.buffer.Reset()

	for {
		openIndex := strings.Index(text, fenceOpen)
		if openIndex == -1 {
			parser.cleanedText.WriteString(text)
			break
		}

		parser.cleanedText.WriteString(text[:openIndex])
		rest := text[openIndex+len(fenceOpen):]

		closeIndex := strings.Index(rest, fenceClose)
		if closeIndex == -1 {
			parser.buffer.WriteString(text[openIndex:])
			break
		}

		jsonBlock := strings.TrimSpace(rest[:closeIndex])
		var command DJCommand
		if err := json.Unmarshal([]byte(jsonBlock), &command); err == nil {
			parser.commands = append(parser.commands, command)
		}

		remaining := rest[closeIndex+len(fenceClose):]
		trimmedRemaining := strings.TrimPrefix(remaining, "\n")
		text = trimmedRemaining
	}

	return parser.commands
}

func (parser *CommandParser) CleanedText() string {
	return parser.cleanedText.String()
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/orchestrator/ -v -race`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/orchestrator/command.go internal/orchestrator/command_test.go
git commit -m "feat(orchestrator): add dj-command delta stream parser"
```

---

### Task 4: AgentPool — Types and Constructor

**Files:**
- Create: `internal/pool/types.go`
- Create: `internal/pool/pool.go`
- Test: `internal/pool/pool_test.go`

**Step 1: Write the failing test**

```go
package pool

import "testing"

func TestNewAgentPool(t *testing.T) {
	pool := NewAgentPool("codex", []string{"proto"}, nil, DefaultMaxAgents)

	if pool == nil {
		t.Fatal("expected non-nil pool")
	}

	agents := pool.All()
	if len(agents) != 0 {
		t.Errorf("expected 0 agents, got %d", len(agents))
	}
}

func TestAgentPoolGet(t *testing.T) {
	pool := NewAgentPool("codex", []string{"proto"}, nil, DefaultMaxAgents)

	_, exists := pool.Get("nonexistent")
	if exists {
		t.Error("expected agent to not exist")
	}
}

func TestAgentRoleConstants(t *testing.T) {
	if RoleOrchestrator != "orchestrator" {
		t.Errorf("expected orchestrator, got %s", RoleOrchestrator)
	}
	if RoleWorker != "worker" {
		t.Errorf("expected worker, got %s", RoleWorker)
	}
}

func TestAgentStatusConstants(t *testing.T) {
	if AgentStatusSpawning != "spawning" {
		t.Errorf("expected spawning, got %s", AgentStatusSpawning)
	}
	if AgentStatusActive != "active" {
		t.Errorf("expected active, got %s", AgentStatusActive)
	}
	if AgentStatusCompleted != "completed" {
		t.Errorf("expected completed, got %s", AgentStatusCompleted)
	}
	if AgentStatusError != "error" {
		t.Errorf("expected error, got %s", AgentStatusError)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/pool/ -v`
Expected: FAIL — package does not exist

**Step 3: Write types**

`internal/pool/types.go`:

```go
package pool

import (
	"github.com/robinojw/dj/internal/appserver"
	"github.com/robinojw/dj/internal/orchestrator"
	"github.com/robinojw/dj/internal/roster"
)

const (
	RoleOrchestrator = "orchestrator"
	RoleWorker       = "worker"
)

const (
	AgentStatusSpawning  = "spawning"
	AgentStatusActive    = "active"
	AgentStatusCompleted = "completed"
	AgentStatusError     = "error"
)

type AgentProcess struct {
	ID        string
	PersonaID string
	ThreadID  string
	Client    *appserver.Client
	Role      string
	Task      string
	Status    string
	ParentID  string
	Persona   *roster.PersonaDefinition
	Parser    *orchestrator.CommandParser
}

type PoolEvent struct {
	AgentID string
	Message appserver.JSONRPCMessage
}
```

`internal/pool/pool.go`:

```go
package pool

import (
	"sync"
	"sync/atomic"

	"github.com/robinojw/dj/internal/roster"
)

const DefaultMaxAgents = 10

const poolEventChannelSize = 128

type AgentPool struct {
	agents    map[string]*AgentProcess
	mu        sync.RWMutex
	events    chan PoolEvent
	command   string
	args      []string
	personas  map[string]roster.PersonaDefinition
	maxAgents int
	idCounter atomic.Int64
}

func NewAgentPool(command string, args []string, personas []roster.PersonaDefinition, maxAgents int) *AgentPool {
	personaMap := make(map[string]roster.PersonaDefinition, len(personas))
	for _, persona := range personas {
		personaMap[persona.ID] = persona
	}

	return &AgentPool{
		agents:    make(map[string]*AgentProcess),
		events:    make(chan PoolEvent, poolEventChannelSize),
		command:   command,
		args:      args,
		personas:  personaMap,
		maxAgents: maxAgents,
	}
}

func (pool *AgentPool) Events() <-chan PoolEvent {
	return pool.events
}

func (pool *AgentPool) Get(agentID string) (*AgentProcess, bool) {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	agent, exists := pool.agents[agentID]
	return agent, exists
}

func (pool *AgentPool) All() []*AgentProcess {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	result := make([]*AgentProcess, 0, len(pool.agents))
	for _, agent := range pool.agents {
		result = append(result, agent)
	}
	return result
}

func (pool *AgentPool) Count() int {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	return len(pool.agents)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/pool/ -v -race`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/pool/types.go internal/pool/pool.go internal/pool/pool_test.go
git commit -m "feat(pool): add AgentPool types and constructor"
```

---

### Task 5: AgentPool — Spawn and Stop

**Files:**
- Modify: `internal/pool/pool.go`
- Test: `internal/pool/spawn_test.go`

**Step 1: Write the failing test**

```go
package pool

import (
	"testing"
)

func TestSpawnRejectsUnknownPersona(t *testing.T) {
	pool := NewAgentPool("codex", []string{"proto"}, nil, DefaultMaxAgents)
	_, err := pool.Spawn("nonexistent", "some task", "")
	if err == nil {
		t.Error("expected error for unknown persona")
	}
}

func TestSpawnRejectsAtCapacity(t *testing.T) {
	pool := NewAgentPool("codex", []string{"proto"}, nil, 0)
	_, err := pool.Spawn("architect", "some task", "")
	if err == nil {
		t.Error("expected error when at capacity")
	}
}

func TestNextAgentID(t *testing.T) {
	pool := NewAgentPool("codex", []string{"proto"}, nil, DefaultMaxAgents)
	id1 := pool.nextAgentID("architect")
	id2 := pool.nextAgentID("architect")
	if id1 == id2 {
		t.Errorf("expected unique IDs, got %s and %s", id1, id2)
	}
}

func TestStopAgentNotFound(t *testing.T) {
	pool := NewAgentPool("codex", []string{"proto"}, nil, DefaultMaxAgents)
	err := pool.StopAgent("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent agent")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/pool/ -run TestSpawn -v`
Expected: FAIL — Spawn not defined

**Step 3: Add Spawn and Stop methods to pool.go**

Append to `internal/pool/pool.go`:

```go
func (pool *AgentPool) Spawn(personaID string, task string, parentAgentID string) (string, error) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	isAtCapacity := len(pool.agents) >= pool.maxAgents
	if isAtCapacity {
		return "", fmt.Errorf("agent pool at capacity (%d)", pool.maxAgents)
	}

	persona, exists := pool.personas[personaID]
	if !exists {
		return "", fmt.Errorf("unknown persona: %s", personaID)
	}

	agentID := pool.nextAgentID(personaID)
	agent := &AgentProcess{
		ID:        agentID,
		PersonaID: personaID,
		Role:      RoleWorker,
		Task:      task,
		Status:    AgentStatusSpawning,
		ParentID:  parentAgentID,
		Persona:   &persona,
		Parser:    orchestrator.NewCommandParser(),
	}
	pool.agents[agentID] = agent

	return agentID, nil
}

func (pool *AgentPool) StopAgent(agentID string) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	agent, exists := pool.agents[agentID]
	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	if agent.Client != nil {
		agent.Client.Stop()
	}
	agent.Status = AgentStatusCompleted
	delete(pool.agents, agentID)
	return nil
}

func (pool *AgentPool) StopAll() {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	for _, agent := range pool.agents {
		if agent.Client != nil {
			agent.Client.Stop()
		}
	}
	pool.agents = make(map[string]*AgentProcess)
}

func (pool *AgentPool) nextAgentID(personaID string) string {
	counter := pool.idCounter.Add(1)
	return fmt.Sprintf("%s-%d", personaID, counter)
}
```

Add `fmt` and `"github.com/robinojw/dj/internal/orchestrator"` to the import block.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/pool/ -v -race`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/pool/pool.go internal/pool/spawn_test.go
git commit -m "feat(pool): add Spawn, StopAgent, and StopAll methods"
```

---

### Task 6: AgentPool — GetByThreadID and Orchestrator Lookup

**Files:**
- Modify: `internal/pool/pool.go`
- Test: `internal/pool/lookup_test.go`

**Step 1: Write the failing test**

```go
package pool

import (
	"testing"

	"github.com/robinojw/dj/internal/roster"
)

func TestGetByThreadID(t *testing.T) {
	personas := []roster.PersonaDefinition{{ID: "architect", Name: "Architect"}}
	pool := NewAgentPool("codex", []string{"proto"}, personas, DefaultMaxAgents)

	agentID, _ := pool.Spawn("architect", "task", "")
	agent, _ := pool.Get(agentID)
	agent.ThreadID = "thread-abc"

	found, exists := pool.GetByThreadID("thread-abc")
	if !exists {
		t.Fatal("expected to find agent by thread ID")
	}
	if found.ID != agentID {
		t.Errorf("expected %s, got %s", agentID, found.ID)
	}
}

func TestGetByThreadIDNotFound(t *testing.T) {
	pool := NewAgentPool("codex", []string{"proto"}, nil, DefaultMaxAgents)
	_, exists := pool.GetByThreadID("nonexistent")
	if exists {
		t.Error("expected agent to not exist")
	}
}

func TestGetOrchestrator(t *testing.T) {
	pool := NewAgentPool("codex", []string{"proto"}, nil, DefaultMaxAgents)

	_, exists := pool.GetOrchestrator()
	if exists {
		t.Error("expected no orchestrator initially")
	}
}

func TestPersonas(t *testing.T) {
	personas := []roster.PersonaDefinition{
		{ID: "architect", Name: "Architect"},
		{ID: "test", Name: "Tester"},
	}
	pool := NewAgentPool("codex", []string{"proto"}, personas, DefaultMaxAgents)

	result := pool.Personas()
	if len(result) != 2 {
		t.Errorf("expected 2 personas, got %d", len(result))
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/pool/ -run TestGetByThreadID -v`
Expected: FAIL — GetByThreadID not defined

**Step 3: Add lookup methods to pool.go**

```go
func (pool *AgentPool) GetByThreadID(threadID string) (*AgentProcess, bool) {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	for _, agent := range pool.agents {
		if agent.ThreadID == threadID {
			return agent, true
		}
	}
	return nil, false
}

func (pool *AgentPool) GetOrchestrator() (*AgentProcess, bool) {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	for _, agent := range pool.agents {
		if agent.Role == RoleOrchestrator {
			return agent, true
		}
	}
	return nil, false
}

func (pool *AgentPool) Personas() map[string]roster.PersonaDefinition {
	return pool.personas
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/pool/ -v -race`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/pool/pool.go internal/pool/lookup_test.go
git commit -m "feat(pool): add thread ID and orchestrator lookup methods"
```

---

### Task 7: Config — Add Roster and Pool Sections

**Files:**
- Modify: `internal/config/config.go:13-63`
- Test: `internal/config/config_test.go`

**Step 1: Write the failing test**

Add to `internal/config/config_test.go`:

```go
func TestDefaultRosterConfig(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Roster.Path != ".roster" {
		t.Errorf("expected .roster, got %s", cfg.Roster.Path)
	}
	if !cfg.Roster.AutoOrchestrate {
		t.Error("expected auto_orchestrate to be true by default")
	}
}

func TestDefaultPoolConfig(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Pool.MaxAgents != 10 {
		t.Errorf("expected max_agents 10, got %d", cfg.Pool.MaxAgents)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/config/ -run TestDefaultRoster -v`
Expected: FAIL — cfg.Roster undefined

**Step 3: Add RosterConfig and PoolConfig to config.go**

Add the new types and fields:

```go
type RosterConfig struct {
	Path            string
	AutoOrchestrate bool
}

type PoolConfig struct {
	MaxAgents int
}
```

Add to `Config` struct:

```go
Roster RosterConfig
Pool   PoolConfig
```

Add defaults in `Load()`:

```go
viperInstance.SetDefault("roster.path", ".roster")
viperInstance.SetDefault("roster.auto_orchestrate", true)
viperInstance.SetDefault("pool.max_agents", 10)
```

Add to config construction:

```go
Roster: RosterConfig{
	Path:            viperInstance.GetString("roster.path"),
	AutoOrchestrate: viperInstance.GetBool("roster.auto_orchestrate"),
},
Pool: PoolConfig{
	MaxAgents: viperInstance.GetInt("pool.max_agents"),
},
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/config/ -v -race`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat(config): add roster and pool configuration sections"
```

---

### Task 8: ThreadState — Add AgentProcessID Field

**Files:**
- Modify: `internal/state/thread.go:16-27`
- Test: `internal/state/thread_test.go`

**Step 1: Write the failing test**

Add to `internal/state/thread_test.go`:

```go
func TestThreadStateAgentProcessID(t *testing.T) {
	thread := NewThreadState("t1", "Test")
	if thread.AgentProcessID != "" {
		t.Error("expected empty AgentProcessID for new thread")
	}
	thread.AgentProcessID = "architect-1"
	if thread.AgentProcessID != "architect-1" {
		t.Errorf("expected architect-1, got %s", thread.AgentProcessID)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/state/ -run TestThreadStateAgentProcessID -v`
Expected: FAIL — AgentProcessID undefined

**Step 3: Add field to ThreadState**

In `internal/state/thread.go`, add to the `ThreadState` struct:

```go
AgentProcessID string
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/state/ -v -race`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/state/thread.go internal/state/thread_test.go
git commit -m "feat(state): add AgentProcessID field to ThreadState"
```

---

### Task 9: TUI — New Pool-Aware Messages

**Files:**
- Create: `internal/tui/msgs_pool.go`
- Test: `internal/tui/msgs_pool_test.go`

**Step 1: Write the failing test**

```go
package tui

import "testing"

func TestSpawnRequestMsgFields(t *testing.T) {
	msg := SpawnRequestMsg{
		SourceAgentID: "orchestrator-1",
		Persona:       "architect",
		Task:          "Design API",
	}
	if msg.SourceAgentID != "orchestrator-1" {
		t.Errorf("expected orchestrator-1, got %s", msg.SourceAgentID)
	}
}

func TestAgentMessageMsgFields(t *testing.T) {
	msg := AgentMessageMsg{
		SourceAgentID: "test-1",
		TargetAgentID: "architect-1",
		Content:       "Need rate limiter",
	}
	if msg.TargetAgentID != "architect-1" {
		t.Errorf("expected architect-1, got %s", msg.TargetAgentID)
	}
}

func TestAgentCompleteMsgFields(t *testing.T) {
	msg := AgentCompleteMsg{
		AgentID: "security-1",
		Content: "Found 2 issues",
	}
	if msg.AgentID != "security-1" {
		t.Errorf("expected security-1, got %s", msg.AgentID)
	}
}

func TestPoolEventMsgFields(t *testing.T) {
	msg := PoolEventMsg{AgentID: "test-1"}
	if msg.AgentID != "test-1" {
		t.Errorf("expected test-1, got %s", msg.AgentID)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run TestSpawnRequestMsg -v`
Expected: FAIL — SpawnRequestMsg not defined

**Step 3: Write the message types**

```go
package tui

import "github.com/robinojw/dj/internal/appserver"

type SpawnRequestMsg struct {
	SourceAgentID string
	Persona       string
	Task          string
}

type AgentMessageMsg struct {
	SourceAgentID string
	TargetAgentID string
	Content       string
}

type AgentCompleteMsg struct {
	AgentID string
	Content string
}

type PoolEventMsg struct {
	AgentID string
	Message appserver.JSONRPCMessage
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/ -run "TestSpawnRequestMsg|TestAgentMessageMsg|TestAgentCompleteMsg|TestPoolEventMsg" -v -race`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/msgs_pool.go internal/tui/msgs_pool_test.go
git commit -m "feat(tui): add pool-aware message types"
```

---

### Task 10: TUI — Persona-Aware Card Colors

**Files:**
- Modify: `internal/tui/card.go:10-34`
- Test: `internal/tui/card_test.go`

**Step 1: Write the failing test**

Add to `internal/tui/card_test.go`:

```go
func TestPersonaColorMapping(t *testing.T) {
	tests := []struct {
		personaID string
		expected  lipgloss.Color
	}{
		{"architect", PersonaColorArchitect},
		{"test", PersonaColorTest},
		{"security", PersonaColorSecurity},
		{"reviewer", PersonaColorReviewer},
		{"performance", PersonaColorPerformance},
		{"design", PersonaColorDesign},
		{"devops", PersonaColorDevOps},
		{"unknown", defaultPersonaColor},
	}

	for _, tc := range tests {
		color := PersonaColor(tc.personaID)
		if color != tc.expected {
			t.Errorf("PersonaColor(%s) = %s, want %s", tc.personaID, color, tc.expected)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run TestPersonaColorMapping -v`
Expected: FAIL — PersonaColor not defined

**Step 3: Add persona color palette to card.go**

Add constants and lookup function:

```go
var (
	PersonaColorArchitect   = lipgloss.Color("33")
	PersonaColorTest        = lipgloss.Color("42")
	PersonaColorSecurity    = lipgloss.Color("196")
	PersonaColorReviewer    = lipgloss.Color("226")
	PersonaColorPerformance = lipgloss.Color("44")
	PersonaColorDesign      = lipgloss.Color("201")
	PersonaColorDevOps      = lipgloss.Color("208")
	PersonaColorDocs        = lipgloss.Color("252")
	PersonaColorAPI         = lipgloss.Color("75")
	PersonaColorData        = lipgloss.Color("178")
	PersonaColorAccessibility = lipgloss.Color("141")
	defaultPersonaColor     = lipgloss.Color("245")
)

var personaColors = map[string]lipgloss.Color{
	"architect":     PersonaColorArchitect,
	"test":          PersonaColorTest,
	"security":      PersonaColorSecurity,
	"reviewer":      PersonaColorReviewer,
	"performance":   PersonaColorPerformance,
	"design":        PersonaColorDesign,
	"devops":        PersonaColorDevOps,
	"docs":          PersonaColorDocs,
	"api":           PersonaColorAPI,
	"data":          PersonaColorData,
	"accessibility": PersonaColorAccessibility,
}

func PersonaColor(personaID string) lipgloss.Color {
	color, exists := personaColors[personaID]
	if !exists {
		return defaultPersonaColor
	}
	return color
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/ -run TestPersonaColorMapping -v -race`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/card.go internal/tui/card_test.go
git commit -m "feat(tui): add persona color palette for agent cards"
```

---

### Task 11: TUI — Card Persona Badge and Orchestrator Border

**Files:**
- Modify: `internal/tui/card.go:36-134`
- Test: `internal/tui/card_test.go`

**Step 1: Write the failing test**

Add to `internal/tui/card_test.go`:

```go
func TestCardPersonaBadge(t *testing.T) {
	thread := &state.ThreadState{
		ID:             "t1",
		Title:          "Design API",
		Status:         state.StatusActive,
		AgentProcessID: "architect-1",
	}
	card := NewCardModel(thread, false, false)
	card.SetPersonaBadge("Architect")
	view := card.View()
	if !strings.Contains(view, "Architect") {
		t.Error("expected persona badge in card view")
	}
}

func TestCardOrchestratorBorder(t *testing.T) {
	thread := &state.ThreadState{
		ID:     "t1",
		Title:  "Orchestrator",
		Status: state.StatusIdle,
	}
	card := NewCardModel(thread, false, false)
	card.SetOrchestrator(true)
	view := card.View()
	if view == "" {
		t.Error("expected non-empty card view")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run TestCardPersonaBadge -v`
Expected: FAIL — SetPersonaBadge not defined

**Step 3: Add persona badge and orchestrator flag to CardModel**

Add fields to `CardModel`:

```go
personaBadge  string
orchestrator  bool
```

Add setter methods:

```go
func (card *CardModel) SetPersonaBadge(badge string) {
	card.personaBadge = badge
}

func (card *CardModel) SetOrchestrator(isOrchestrator bool) {
	card.orchestrator = isOrchestrator
}
```

Modify `buildContent` to include the badge:

```go
func (card CardModel) buildContent(title string, statusLine string) string {
	hasBadge := card.personaBadge != ""
	isSubAgent := card.thread.ParentID != ""
	hasRole := isSubAgent && card.thread.AgentRole != ""

	badgeLine := ""
	if hasBadge {
		badgeColor := PersonaColor(strings.ToLower(card.personaBadge))
		badgeLine = lipgloss.NewStyle().
			Foreground(badgeColor).
			Bold(true).
			Render(card.personaBadge)
	}

	roleLine := ""
	if hasRole {
		roleLine = lipgloss.NewStyle().
			Foreground(colorIdle).
			Render(roleIndent + card.thread.AgentRole)
	}

	lines := []string{title}
	if badgeLine != "" {
		lines = append(lines, badgeLine)
	}
	if roleLine != "" {
		lines = append(lines, roleLine)
	}
	lines = append(lines, statusLine)
	return strings.Join(lines, "\n")
}
```

Modify `buildBorderStyle` to handle orchestrator:

```go
func (card CardModel) buildBorderStyle() lipgloss.Style {
	style := lipgloss.NewStyle().
		Width(card.width).
		Height(card.height).
		Padding(0, 1)

	if card.orchestrator {
		style = style.
			Border(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("214"))
	} else if card.selected {
		style = style.
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("39"))
	} else {
		style = style.Border(lipgloss.RoundedBorder())
	}
	return style
}
```

Add `"strings"` to imports.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/ -run "TestCardPersonaBadge|TestCardOrchestratorBorder" -v -race`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/card.go internal/tui/card_test.go
git commit -m "feat(tui): add persona badge and orchestrator border to cards"
```

---

### Task 12: TUI — AppModel Pool Integration

**Files:**
- Modify: `internal/tui/app.go:1-60`
- Modify: `internal/tui/app_proto.go:1-53`
- Test: `internal/tui/app_test.go`

**Step 1: Write the failing test**

Add to `internal/tui/app_test.go`:

```go
func TestNewAppModelWithPool(t *testing.T) {
	store := state.NewThreadStore()
	pool := pool.NewAgentPool("codex", []string{"proto"}, nil, 10)
	app := NewAppModel(store, WithPool(pool))
	if app.pool == nil {
		t.Error("expected pool to be set")
	}
}

func TestNewAppModelWithoutPool(t *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)
	if app.pool != nil {
		t.Error("expected pool to be nil for backward compatibility")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run TestNewAppModelWithPool -v`
Expected: FAIL — WithPool not defined

**Step 3: Add pool to AppModel**

In `internal/tui/app.go`, add field:

```go
pool *pool.AgentPool
```

Add option:

```go
func WithPool(agentPool *pool.AgentPool) AppOption {
	return func(app *AppModel) {
		app.pool = agentPool
	}
}
```

Add import for `"github.com/robinojw/dj/internal/pool"`.

In `internal/tui/app_proto.go`, modify `connectClient` and `listenForEvents` to work with either a single client or a pool. When pool is set, listen on pool events instead.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/ -run "TestNewAppModelWithPool|TestNewAppModelWithoutPool" -v -race`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/app.go internal/tui/app_proto.go internal/tui/app_test.go
git commit -m "feat(tui): integrate AgentPool into AppModel"
```

---

### Task 13: TUI — Handle SpawnRequest, AgentMessage, AgentComplete

**Files:**
- Create: `internal/tui/app_pool.go`
- Test: `internal/tui/app_pool_test.go`

**Step 1: Write the failing test**

```go
package tui

import (
	"testing"

	"github.com/robinojw/dj/internal/pool"
	"github.com/robinojw/dj/internal/roster"
	"github.com/robinojw/dj/internal/state"
)

func TestHandleSpawnRequestCreatesThread(t *testing.T) {
	store := state.NewThreadStore()
	personas := []roster.PersonaDefinition{{ID: "architect", Name: "Architect"}}
	agentPool := pool.NewAgentPool("echo", []string{"hello"}, personas, 10)

	app := NewAppModel(store, WithPool(agentPool))
	msg := SpawnRequestMsg{
		SourceAgentID: "orchestrator-1",
		Persona:       "architect",
		Task:          "Design API",
	}

	updated, _ := app.handleSpawnRequest(msg)
	resultApp := updated.(AppModel)
	threads := resultApp.store.All()

	hasThread := len(threads) > 0
	if !hasThread {
		t.Error("expected at least one thread after spawn request")
	}
}

func TestHandleAgentCompleteUpdatesStatus(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t1", "Test Agent")

	agentPool := pool.NewAgentPool("echo", []string{}, nil, 10)
	app := NewAppModel(store, WithPool(agentPool))

	msg := AgentCompleteMsg{
		AgentID: "test-1",
		Content: "Done",
	}

	updated, _ := app.handleAgentComplete(msg)
	_ = updated.(AppModel)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run TestHandleSpawnRequest -v`
Expected: FAIL — handleSpawnRequest not defined

**Step 3: Implement pool event handlers**

```go
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/state"
)

func (app AppModel) handleSpawnRequest(msg SpawnRequestMsg) (tea.Model, tea.Cmd) {
	if app.pool == nil {
		return app, nil
	}

	agentID, err := app.pool.Spawn(msg.Persona, msg.Task, msg.SourceAgentID)
	if err != nil {
		app.statusBar.SetError(err.Error())
		return app, nil
	}

	app.store.Add(agentID, msg.Task)
	agent, exists := app.pool.Get(agentID)
	if exists {
		thread, threadExists := app.store.Get(agentID)
		if threadExists {
			thread.AgentProcessID = agentID
			thread.AgentRole = msg.Persona
			thread.ParentID = msg.SourceAgentID
		}
		_ = agent
	}

	app.statusBar.SetThreadCount(len(app.store.All()))
	app.tree.Refresh()
	return app, nil
}

func (app AppModel) handleAgentMessage(msg AgentMessageMsg) (tea.Model, tea.Cmd) {
	if app.pool == nil {
		return app, nil
	}

	targetAgent, exists := app.pool.Get(msg.TargetAgentID)
	if !exists {
		return app, nil
	}

	if targetAgent.Client == nil {
		return app, nil
	}

	sourceAgent, sourceExists := app.pool.Get(msg.SourceAgentID)
	senderLabel := msg.SourceAgentID
	if sourceExists && sourceAgent.Persona != nil {
		senderLabel = sourceAgent.Persona.Name
	}

	wrappedMessage := "[From: " + msg.SourceAgentID + " (" + senderLabel + ")] " + msg.Content
	targetAgent.Client.SendUserInput(wrappedMessage)
	return app, nil
}

func (app AppModel) handleAgentComplete(msg AgentCompleteMsg) (tea.Model, tea.Cmd) {
	app.store.UpdateStatus(msg.AgentID, state.StatusCompleted, "")
	app.store.UpdateActivity(msg.AgentID, "")
	return app, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/ -run "TestHandleSpawnRequest|TestHandleAgentComplete" -v -race`
Expected: PASS

**Step 5: Wire handlers into Update()**

In `internal/tui/app.go`, add cases to `handleAgentMsg`:

```go
case SpawnRequestMsg:
	return app.handleSpawnRequest(msg)
case AgentMessageMsg:
	return app.handleAgentMessage(msg)
case AgentCompleteMsg:
	return app.handleAgentComplete(msg)
```

**Step 6: Commit**

```bash
git add internal/tui/app_pool.go internal/tui/app_pool_test.go internal/tui/app.go
git commit -m "feat(tui): handle spawn request, agent message, and agent complete events"
```

---

### Task 14: TUI — Header and Status Bar Swarm Indicators

**Files:**
- Modify: `internal/tui/header.go:15-22`
- Modify: `internal/tui/statusbar.go:24-91`
- Test: `internal/tui/header_test.go`
- Test: `internal/tui/statusbar_test.go`

**Step 1: Write the failing test**

Add to `internal/tui/header_test.go`:

```go
func TestHeaderSwarmHints(t *testing.T) {
	header := NewHeaderBar(80)
	header.SetSwarmActive(true)
	view := header.View()
	if !strings.Contains(view, "p: persona") {
		t.Error("expected persona hint when swarm is active")
	}
}
```

Add to `internal/tui/statusbar_test.go`:

```go
func TestStatusBarAgentCount(t *testing.T) {
	bar := NewStatusBar()
	bar.SetWidth(80)
	bar.SetAgentCount(3, 1)
	view := bar.View()
	if !strings.Contains(view, "3 agents") {
		t.Error("expected agent count in status bar")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run "TestHeaderSwarmHints|TestStatusBarAgentCount" -v`
Expected: FAIL — SetSwarmActive and SetAgentCount not defined

**Step 3: Add swarm hints to header**

In `internal/tui/header.go`, add `swarmActive bool` field to `HeaderBar`. Add `SetSwarmActive(bool)` method. In `View()`, append swarm-specific hints (`p: persona`, `m: message`, `s: swarm`) when active.

In `internal/tui/statusbar.go`, add `agentCount int` and `completedCount int` fields. Add `SetAgentCount(total, completed int)` method. In `View()`, add agent count section when > 0.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/ -run "TestHeaderSwarmHints|TestStatusBarAgentCount" -v -race`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/header.go internal/tui/header_test.go internal/tui/statusbar.go internal/tui/statusbar_test.go
git commit -m "feat(tui): add swarm status indicators to header and status bar"
```

---

### Task 15: TUI — New Keybindings (p, m, s, K)

**Files:**
- Modify: `internal/tui/app_keys.go:58-71`
- Test: `internal/tui/app_keys_test.go`

**Step 1: Write the failing test**

```go
func TestPersonaPickerKeybinding(t *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	updated, _ := app.handleRune(msg)
	_ = updated
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run TestPersonaPickerKeybinding -v`
Expected: May pass (no-op) or fail depending on existing handling

**Step 3: Add keybindings to handleRune**

In `internal/tui/app_keys.go`, update `handleRune`:

```go
case "p":
	return app.showPersonaPicker()
case "m":
	return app.sendMessageToAgent()
case "K":
	return app.killAgent()
```

Remap existing `s` (select/pin) to avoid conflict with swarm view. Existing `s` is used for pin toggle alongside `" "` — keep space for pin, use `s` for swarm toggle:

```go
case "s":
	return app.toggleSwarmView()
case " ":
	return app.togglePin()
```

Implement stub methods in a new `internal/tui/app_swarm.go`:

```go
func (app AppModel) showPersonaPicker() (tea.Model, tea.Cmd) {
	return app, nil
}

func (app AppModel) sendMessageToAgent() (tea.Model, tea.Cmd) {
	return app, nil
}

func (app AppModel) killAgent() (tea.Model, tea.Cmd) {
	if app.pool == nil {
		return app, nil
	}
	threadID := app.canvas.SelectedThreadID()
	agent, exists := app.pool.GetByThreadID(threadID)
	if !exists {
		return app, nil
	}
	app.pool.StopAgent(agent.ID)
	app.store.UpdateStatus(threadID, state.StatusCompleted, "")
	return app, nil
}

func (app AppModel) toggleSwarmView() (tea.Model, tea.Cmd) {
	return app, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/tui/ -v -race`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/app_keys.go internal/tui/app_swarm.go internal/tui/app_keys_test.go
git commit -m "feat(tui): add p/m/s/K keybindings for swarm control"
```

---

### Task 16: Main — Startup Flow with Roster and Pool

**Files:**
- Modify: `cmd/dj/main.go:35-59`
- Test: manual verification (requires codex CLI)

**Step 1: Update runApp**

```go
func runApp(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	store := state.NewThreadStore()
	var opts []tui.AppOption

	personas, signals := loadRoster(cfg)
	hasPersonas := len(personas) > 0

	if hasPersonas && cfg.Roster.AutoOrchestrate {
		agentPool := pool.NewAgentPool(
			cfg.AppServer.Command,
			cfg.AppServer.Args,
			personas,
			cfg.Pool.MaxAgents,
		)
		opts = append(opts, tui.WithPool(agentPool))
		_ = signals
	} else {
		client := appserver.NewClient(cfg.AppServer.Command, cfg.AppServer.Args...)
		defer client.Stop()
		opts = append(opts, tui.WithClient(client))
	}

	opts = append(opts, tui.WithInteractiveCommand(cfg.Interactive.Command, cfg.Interactive.Args...))
	app := tui.NewAppModel(store, opts...)

	program := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())
	finalModel, err := program.Run()

	if finalApp, ok := finalModel.(tui.AppModel); ok {
		finalApp.StopAllPTYSessions()
	}

	return err
}

func loadRoster(cfg *config.Config) ([]roster.PersonaDefinition, *roster.RepoSignals) {
	personaDir := filepath.Join(cfg.Roster.Path, "personas")
	personas, err := roster.LoadPersonas(personaDir)
	if err != nil {
		return nil, nil
	}

	signalsPath := filepath.Join(cfg.Roster.Path, "signals.json")
	signals, err := roster.LoadSignals(signalsPath)
	if err != nil {
		return personas, nil
	}

	return personas, signals
}
```

Add imports: `"path/filepath"`, `"github.com/robinojw/dj/internal/pool"`, `"github.com/robinojw/dj/internal/roster"`.

**Step 2: Run build to verify compilation**

Run: `go build -o dj ./cmd/dj`
Expected: Build succeeds

**Step 3: Run all tests**

Run: `go test ./... -v -race`
Expected: All PASS

**Step 4: Commit**

```bash
git add cmd/dj/main.go
git commit -m "feat(main): integrate roster loading and agent pool into startup"
```

---

### Task 17: Integration — Pool Event Multiplexing

**Files:**
- Modify: `internal/tui/app_proto.go`
- Create: `internal/tui/app_pool_events.go`
- Test: `internal/tui/app_pool_events_test.go`

**Step 1: Write the failing test**

```go
package tui

import (
	"testing"

	"github.com/robinojw/dj/internal/appserver"
	poolpkg "github.com/robinojw/dj/internal/pool"
	"github.com/robinojw/dj/internal/state"
)

func TestListenForPoolEvents(t *testing.T) {
	store := state.NewThreadStore()
	agentPool := poolpkg.NewAgentPool("echo", []string{}, nil, 10)
	app := NewAppModel(store, WithPool(agentPool))

	cmd := app.listenForPoolEvents()
	if cmd == nil {
		t.Error("expected non-nil command")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run TestListenForPoolEvents -v`
Expected: FAIL — listenForPoolEvents not defined

**Step 3: Implement pool event listener**

```go
package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/orchestrator"
)

func (app AppModel) listenForPoolEvents() tea.Cmd {
	if app.pool == nil {
		return nil
	}
	return func() tea.Msg {
		event, ok := <-app.pool.Events()
		if !ok {
			return AppServerErrorMsg{Err: fmt.Errorf("pool events closed")}
		}
		return PoolEventMsg{
			AgentID: event.AgentID,
			Message: event.Message,
		}
	}
}

func (app AppModel) handlePoolEvent(msg PoolEventMsg) (tea.Model, tea.Cmd) {
	agent, exists := app.pool.Get(msg.AgentID)
	if !exists {
		return app, app.listenForPoolEvents()
	}

	tuiMsg := V2MessageToMsg(msg.Message)
	if tuiMsg == nil {
		return app, app.listenForPoolEvents()
	}

	if deltaMsg, ok := tuiMsg.(V2AgentDeltaMsg); ok {
		agent.Parser.Feed(deltaMsg.Delta)
		commands := agent.Parser.Flush()
		return app.processCommands(msg.AgentID, commands, tuiMsg)
	}

	updated, innerCmd := app.Update(tuiMsg)
	resultApp := updated.(AppModel)
	return resultApp, tea.Batch(innerCmd, resultApp.listenForPoolEvents())
}

func (app AppModel) processCommands(agentID string, commands []orchestrator.DJCommand, originalMsg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	updated, innerCmd := app.Update(originalMsg)
	resultApp := updated.(AppModel)
	cmds = append(cmds, innerCmd)

	for _, command := range commands {
		switch command.Action {
		case "spawn":
			spawnMsg := SpawnRequestMsg{
				SourceAgentID: agentID,
				Persona:       command.Persona,
				Task:          command.Task,
			}
			spawnUpdated, spawnCmd := resultApp.handleSpawnRequest(spawnMsg)
			resultApp = spawnUpdated.(AppModel)
			cmds = append(cmds, spawnCmd)
		case "message":
			msgCmd := AgentMessageMsg{
				SourceAgentID: agentID,
				TargetAgentID: command.Target,
				Content:       command.Content,
			}
			msgUpdated, msgInnerCmd := resultApp.handleAgentMessage(msgCmd)
			resultApp = msgUpdated.(AppModel)
			cmds = append(cmds, msgInnerCmd)
		case "complete":
			completeMsg := AgentCompleteMsg{
				AgentID: agentID,
				Content: command.Content,
			}
			completeUpdated, completeCmd := resultApp.handleAgentComplete(completeMsg)
			resultApp = completeUpdated.(AppModel)
			cmds = append(cmds, completeCmd)
		}
	}

	cmds = append(cmds, resultApp.listenForPoolEvents())
	return resultApp, tea.Batch(cmds...)
}
```

**Step 4: Wire into Init() and Update()**

In `app.go` `Init()`, add `app.listenForPoolEvents()` to the batch when pool is present.

In `Update()`, add:

```go
case PoolEventMsg:
	return app.handlePoolEvent(msg)
```

**Step 5: Run tests**

Run: `go test ./... -v -race`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/tui/app_pool_events.go internal/tui/app_pool_events_test.go internal/tui/app.go
git commit -m "feat(tui): pool event multiplexing with command parsing"
```

---

### Task 18: Full Build Verification and Lint

**Step 1: Run full test suite**

Run: `go test ./... -v -race`
Expected: All PASS

**Step 2: Run linter**

Run: `golangci-lint run`
Expected: No errors (may need to fix funlen/cyclop violations by extracting helpers)

**Step 3: Run build**

Run: `go build -o dj ./cmd/dj`
Expected: Build succeeds

**Step 4: Fix any lint violations**

If `funlen` or `cyclop` flags functions as too long/complex, extract helper functions to stay within the 60-line / 15-complexity limits from `.golangci.yml`.

**Step 5: Final commit**

```bash
git add -A
git commit -m "chore: fix lint violations and verify full build"
```

---

## Summary

| Task | Package | Description |
|------|---------|-------------|
| 1 | `internal/roster/` | PersonaDefinition and RepoSignals types |
| 2 | `internal/roster/` | Persona and signals file loaders |
| 3 | `internal/orchestrator/` | dj-command delta stream parser |
| 4 | `internal/pool/` | AgentPool types and constructor |
| 5 | `internal/pool/` | Spawn, StopAgent, StopAll methods |
| 6 | `internal/pool/` | Thread ID and orchestrator lookup |
| 7 | `internal/config/` | Roster and pool config sections |
| 8 | `internal/state/` | AgentProcessID field on ThreadState |
| 9 | `internal/tui/` | Pool-aware message types |
| 10 | `internal/tui/` | Persona color palette |
| 11 | `internal/tui/` | Card persona badge and orchestrator border |
| 12 | `internal/tui/` | AppModel pool integration |
| 13 | `internal/tui/` | Spawn request, agent message, agent complete handlers |
| 14 | `internal/tui/` | Header and status bar swarm indicators |
| 15 | `internal/tui/` | p/m/s/K keybindings |
| 16 | `cmd/dj/` | Startup flow with roster and pool |
| 17 | `internal/tui/` | Pool event multiplexing with command parsing |
| 18 | — | Full build verification and lint |
