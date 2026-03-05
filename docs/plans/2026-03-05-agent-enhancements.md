# Agent Enhancements (1–7) Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add 7 major capabilities — dual-mode agents, LSP integration, checkpoint/undo, session persistence, event hooks, context compaction, and @mention injection — turning DJ from a streaming chat shell into a production-grade agentic coding assistant.

**Architecture:** Each enhancement is a self-contained package under `internal/` with a thin integration layer wired through `tui/app.go` and `cmd/harness/main.go`. The features build on each other: memory files (4) feeds into context compaction (6), and the checkpoint system (3) stores `previous_response_id` for API continuation. LSP (2) and @mentions (7) both inject context into the prompt, so they share a `context.Injector` interface.

**Tech Stack:** Go 1.22+, bubbletea/lipgloss (existing), `go.lsp.dev/protocol` + `go.lsp.dev/jsonrpc2` for LSP, existing `encoding/json` + `os/exec` for hooks.

---

## Task 1: Dual-Mode Agents (Plan vs Build)

**Files:**
- Create: `internal/agents/modes.go`
- Create: `internal/agents/modes_test.go`
- Modify: `internal/agents/worker.go`
- Modify: `internal/tui/app.go`
- Modify: `internal/tui/screens/chat.go`
- Modify: `internal/tui/components/statusbar.go`

### Step 1: Write the failing test for mode configuration

```go
// internal/agents/modes_test.go
package agents

import "testing"

func TestModeConfigPlan(t *testing.T) {
	cfg := Modes[ModePlan]
	if cfg.ReasoningEffort != "high" {
		t.Errorf("Plan mode should use high reasoning, got %s", cfg.ReasoningEffort)
	}
	if len(cfg.AllowedTools) == 0 {
		t.Error("Plan mode should have an explicit allowlist")
	}
	for _, tool := range cfg.AllowedTools {
		if tool == "write_file" || tool == "run_command" {
			t.Errorf("Plan mode should not allow %s", tool)
		}
	}
}

func TestModeConfigBuild(t *testing.T) {
	cfg := Modes[ModeBuild]
	if cfg.ReasoningEffort != "medium" {
		t.Errorf("Build mode should use medium reasoning, got %s", cfg.ReasoningEffort)
	}
	if cfg.AllowedTools != nil {
		t.Error("Build mode should allow all tools (nil allowlist)")
	}
}

func TestFilterToolsByMode(t *testing.T) {
	allTools := []string{"read_file", "write_file", "run_command", "search_code", "list_dir"}
	cfg := Modes[ModePlan]
	filtered := FilterTools(allTools, cfg)

	for _, tool := range filtered {
		if tool == "write_file" || tool == "run_command" {
			t.Errorf("Plan mode filtered list should not contain %s", tool)
		}
	}
	if len(filtered) != 3 {
		t.Errorf("Expected 3 tools in plan mode, got %d", len(filtered))
	}
}
```

### Step 2: Run test to verify it fails

Run: `cd /Users/robin.white/dev/dj && go test ./internal/agents/ -run TestMode -v`
Expected: FAIL — `Modes` undefined, `FilterTools` undefined

### Step 3: Write the implementation

```go
// internal/agents/modes.go
package agents

import "fmt"

// AgentMode determines the agent's permission level and persona.
type AgentMode int

const (
	ModeBuild AgentMode = iota // full tools: read, write, run, MCP
	ModePlan                   // read-only: no writes, no exec, no MCP mutations
)

func (m AgentMode) String() string {
	switch m {
	case ModeBuild:
		return "Build"
	case ModePlan:
		return "Plan"
	default:
		return fmt.Sprintf("Unknown(%d)", int(m))
	}
}

// ModeConfig holds the constraints and persona for a mode.
type ModeConfig struct {
	Mode            AgentMode
	AllowedTools    []string // nil = all tools allowed
	SystemPrompt    string
	ReasoningEffort string // "low", "medium", "high"
}

var planSystemPrompt = `You are a senior software architect in Plan mode.
Your job is to analyze, reason, and produce a detailed implementation plan.
You may ONLY read files, search code, and list directories.
You may NOT write files, execute commands, or invoke MCP tools that mutate state.
Think deeply. Output a numbered, step-by-step implementation plan with exact file paths.`

var buildSystemPrompt = `You are a skilled software engineer in Build mode.
You have full access to all tools: read, write, search, execute, and MCP.
Follow the plan precisely. Make minimal, focused changes. Run tests after each edit.`

// Modes maps each AgentMode to its configuration.
var Modes = map[AgentMode]ModeConfig{
	ModePlan: {
		Mode:            ModePlan,
		AllowedTools:    []string{"read_file", "search_code", "list_dir"},
		SystemPrompt:    planSystemPrompt,
		ReasoningEffort: "high",
	},
	ModeBuild: {
		Mode:            ModeBuild,
		AllowedTools:    nil, // all tools enabled
		SystemPrompt:    buildSystemPrompt,
		ReasoningEffort: "medium",
	},
}

// FilterTools returns only the tools allowed by the given mode config.
// If cfg.AllowedTools is nil, all tools are returned.
func FilterTools(allTools []string, cfg ModeConfig) []string {
	if cfg.AllowedTools == nil {
		return allTools
	}
	allowed := make(map[string]bool, len(cfg.AllowedTools))
	for _, t := range cfg.AllowedTools {
		allowed[t] = true
	}
	var filtered []string
	for _, t := range allTools {
		if allowed[t] {
			filtered = append(filtered, t)
		}
	}
	return filtered
}
```

### Step 4: Run tests to verify they pass

Run: `cd /Users/robin.white/dev/dj && go test ./internal/agents/ -run TestMode -v`
Expected: PASS

### Step 5: Wire mode into Worker

Modify `internal/agents/worker.go` — add a `Mode` field to `Worker` and use it in `buildInstructions()`:

```go
// In Worker struct, add:
Mode     AgentMode

// In NewWorker(), add parameter:
func NewWorker(
	task Subtask,
	client *api.ResponsesClient,
	skillsRegistry *skills.Registry,
	model string,
	parentID string,
	mode AgentMode,     // NEW
) *Worker {
	return &Worker{
		// ... existing fields ...
		Mode: mode,
	}
}

// In buildInstructions(), replace base string:
func (w *Worker) buildInstructions() string {
	modeCfg := Modes[w.Mode]
	base := modeCfg.SystemPrompt + "\n\n"
	base += fmt.Sprintf("Subtask: %s\n", w.Task.Description)
	// ... rest unchanged ...
}

// In Run(), use mode's reasoning effort:
req := api.CreateResponseRequest{
	// ...
	Reasoning: &api.Reasoning{
		Effort: Modes[w.Mode].ReasoningEffort,
	},
	// ...
}
```

### Step 6: Wire mode into Orchestrator

Modify `internal/agents/orchestrator.go` — add `Mode` field, pass it to workers:

```go
// In Orchestrator struct, add:
Mode AgentMode

// In Dispatch(), when creating workers:
w := NewWorker(task, o.client, o.skills, o.model, o.RootID, o.Mode)
```

### Step 7: Add mode toggle to App and Chat screen

Modify `internal/tui/app.go`:
```go
// In App struct, add:
mode agents.AgentMode

// In Update(), add Tab key handler:
case "tab":
	if a.mode == agents.ModeBuild {
		a.mode = agents.ModePlan
	} else {
		a.mode = agents.ModeBuild
	}
	return a, nil
```

Modify `internal/tui/screens/chat.go` — add a `Mode` field and display it:
```go
// In ChatModel struct, add:
Mode agents.AgentMode

// Add setter:
func (m *ChatModel) SetMode(mode agents.AgentMode) {
	m.Mode = mode
}
```

Modify `internal/tui/components/statusbar.go` — add mode badge:
```go
// In StatusBar struct, add:
Mode string

// In View(), prepend mode badge:
modeBadge := s.Theme.BadgeStyle().Render(s.Mode)
content := fmt.Sprintf("%s  CTX %s %.1f%%  ...", modeBadge, ctxBar, ctxPct, ...)
```

### Step 8: Run full build

Run: `cd /Users/robin.white/dev/dj && go build ./...`
Expected: Clean build

### Step 9: Commit

```bash
git add internal/agents/modes.go internal/agents/modes_test.go internal/agents/worker.go internal/agents/orchestrator.go internal/tui/app.go internal/tui/screens/chat.go internal/tui/components/statusbar.go
git commit -m "feat: add Plan/Build dual-mode agents with Tab toggle"
```

---

## Task 2: LSP Integration

**Files:**
- Create: `internal/lsp/client.go`
- Create: `internal/lsp/client_test.go`
- Create: `internal/lsp/detect.go`
- Create: `internal/lsp/types.go`
- Modify: `internal/tui/components/statusbar.go`
- Modify: `internal/tui/app.go`
- Modify: `cmd/harness/main.go`
- Modify: `config/config.go`
- Modify: `go.mod`

### Step 1: Add LSP dependency

Run: `cd /Users/robin.white/dev/dj && go get go.lsp.dev/protocol@latest go.lsp.dev/jsonrpc2@latest`

### Step 2: Write LSP types

```go
// internal/lsp/types.go
package lsp

// Diagnostic represents an LSP diagnostic (error, warning, etc.).
type Diagnostic struct {
	File     string
	Line     int
	Column   int
	Severity string // "error", "warning", "info", "hint"
	Message  string
	Source   string // "gopls", "typescript-language-server", etc.
}

// ServerConfig holds the configuration to launch an LSP server.
type ServerConfig struct {
	Language string // "go", "typescript", "python"
	Command  string // "gopls", "typescript-language-server --stdio"
	Args     []string
}

// DetectedServer is the result of auto-detection.
type DetectedServer struct {
	Config   ServerConfig
	RootPath string
}
```

### Step 3: Write language detection

```go
// internal/lsp/detect.go
package lsp

import "os"

// knownServers maps language indicators to LSP server commands.
var knownServers = map[string]ServerConfig{
	"go.mod": {
		Language: "go",
		Command:  "gopls",
		Args:     []string{"serve"},
	},
	"tsconfig.json": {
		Language: "typescript",
		Command:  "typescript-language-server",
		Args:     []string{"--stdio"},
	},
	"package.json": {
		Language: "typescript",
		Command:  "typescript-language-server",
		Args:     []string{"--stdio"},
	},
	"pyproject.toml": {
		Language: "python",
		Command:  "pylsp",
		Args:     nil,
	},
	"setup.py": {
		Language: "python",
		Command:  "pylsp",
		Args:     nil,
	},
}

// Detect scans the project root for known language indicators
// and returns the first matching LSP server config.
func Detect(root string) *DetectedServer {
	for marker, cfg := range knownServers {
		if _, err := os.Stat(root + "/" + marker); err == nil {
			return &DetectedServer{Config: cfg, RootPath: root}
		}
	}
	return nil
}
```

### Step 4: Write the failing test for the LSP client

```go
// internal/lsp/client_test.go
package lsp

import "testing"

func TestDetectGo(t *testing.T) {
	// Use the DJ project itself (has go.mod at root)
	result := Detect("/Users/robin.white/dev/dj")
	if result == nil {
		t.Fatal("Expected to detect Go LSP server")
	}
	if result.Config.Language != "go" {
		t.Errorf("Expected language 'go', got %s", result.Config.Language)
	}
	if result.Config.Command != "gopls" {
		t.Errorf("Expected command 'gopls', got %s", result.Config.Command)
	}
}

func TestDetectNoMatch(t *testing.T) {
	result := Detect("/tmp")
	if result != nil {
		t.Error("Expected nil for directory with no known language markers")
	}
}

func TestDiagnosticsFormat(t *testing.T) {
	d := Diagnostic{
		File:     "main.go",
		Line:     10,
		Column:   5,
		Severity: "error",
		Message:  "undefined: foo",
		Source:   "gopls",
	}
	expected := "main.go:10:5: error: undefined: foo (gopls)"
	got := FormatDiagnostic(d)
	if got != expected {
		t.Errorf("Expected %q, got %q", expected, got)
	}
}
```

### Step 5: Run tests to verify they fail

Run: `cd /Users/robin.white/dev/dj && go test ./internal/lsp/ -v`
Expected: FAIL — `FormatDiagnostic` undefined

### Step 6: Write the LSP client implementation

```go
// internal/lsp/client.go
package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const diagnosticCollectTimeout = 500 * time.Millisecond

// Client manages a running LSP server process.
type Client struct {
	config   ServerConfig
	rootPath string
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	stdout   *bufio.Reader
	mu       sync.Mutex
	nextID   atomic.Int64
	diagMu   sync.Mutex
	diags    map[string][]Diagnostic // file → diagnostics
}

// NewClient creates a new LSP client but does not start the server.
func NewClient(cfg ServerConfig, rootPath string) *Client {
	return &Client{
		config:   cfg,
		rootPath: rootPath,
		diags:    make(map[string][]Diagnostic),
	}
}

// Start launches the LSP server and performs the initialize handshake.
func (c *Client) Start(ctx context.Context) error {
	args := c.config.Args
	c.cmd = exec.CommandContext(ctx, c.config.Command, args...)

	var err error
	c.stdin, err = c.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe: %w", err)
	}

	stdoutPipe, err := c.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	c.stdout = bufio.NewReader(stdoutPipe)

	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("start LSP server %s: %w", c.config.Command, err)
	}

	// Send initialize request
	return c.initialize(ctx)
}

func (c *Client) initialize(ctx context.Context) error {
	params := map[string]interface{}{
		"processId": nil,
		"rootUri":   "file://" + c.rootPath,
		"capabilities": map[string]interface{}{
			"textDocument": map[string]interface{}{
				"publishDiagnostics": map[string]interface{}{},
			},
		},
	}

	var result json.RawMessage
	if err := c.call("initialize", params, &result); err != nil {
		return fmt.Errorf("LSP initialize: %w", err)
	}

	// Send initialized notification
	return c.notify("initialized", struct{}{})
}

// NotifyChange sends textDocument/didChange and collects diagnostics.
func (c *Client) NotifyChange(file string, content string) ([]Diagnostic, error) {
	uri := "file://" + file

	// Send didOpen (simplified — full impl tracks open files)
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri":        uri,
			"languageId": c.config.Language,
			"version":    1,
			"text":       content,
		},
	}
	if err := c.notify("textDocument/didOpen", params); err != nil {
		return nil, err
	}

	// Collect diagnostics that arrive within timeout
	return c.collectDiagnostics(file, diagnosticCollectTimeout), nil
}

func (c *Client) collectDiagnostics(file string, timeout time.Duration) []Diagnostic {
	time.Sleep(timeout)
	c.diagMu.Lock()
	defer c.diagMu.Unlock()
	return c.diags[file]
}

// Language returns the language this client handles.
func (c *Client) Language() string {
	return c.config.Language
}

// Command returns the LSP server command name.
func (c *Client) Command() string {
	return c.config.Command
}

// Close shuts down the LSP server.
func (c *Client) Close() error {
	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.cmd != nil && c.cmd.Process != nil {
		return c.cmd.Process.Kill()
	}
	return nil
}

// FormatDiagnostic returns a human-readable string for a diagnostic.
func FormatDiagnostic(d Diagnostic) string {
	return fmt.Sprintf("%s:%d:%d: %s: %s (%s)",
		d.File, d.Line, d.Column, d.Severity, d.Message, d.Source)
}

// FormatDiagnostics returns a newline-separated string of all diagnostics.
func FormatDiagnostics(diags []Diagnostic) string {
	lines := make([]string, len(diags))
	for i, d := range diags {
		lines[i] = FormatDiagnostic(d)
	}
	return strings.Join(lines, "\n")
}

// --- JSON-RPC 2.0 transport (LSP uses Content-Length framing) ---

func (c *Client) call(method string, params interface{}, result interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := int(c.nextID.Add(1))

	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	}

	if err := c.writeMessage(req); err != nil {
		return err
	}

	return c.readResponse(result)
}

func (c *Client) notify(method string, params interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}

	return c.writeMessage(req)
}

func (c *Client) writeMessage(msg interface{}) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	_, err = fmt.Fprintf(c.stdin, "%s%s", header, body)
	return err
}

func (c *Client) readResponse(result interface{}) error {
	// Read Content-Length header
	var contentLength int
	for {
		line, err := c.stdout.ReadString('\n')
		if err != nil {
			return fmt.Errorf("read header: %w", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break // empty line = end of headers
		}
		if strings.HasPrefix(line, "Content-Length:") {
			fmt.Sscanf(line, "Content-Length: %d", &contentLength)
		}
	}

	if contentLength == 0 {
		return fmt.Errorf("no Content-Length in response")
	}

	body := make([]byte, contentLength)
	if _, err := io.ReadFull(c.stdout, body); err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	var resp struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("LSP error %d: %s", resp.Error.Code, resp.Error.Message)
	}

	if result != nil && resp.Result != nil {
		return json.Unmarshal(resp.Result, result)
	}

	return nil
}
```

### Step 7: Run tests

Run: `cd /Users/robin.white/dev/dj && go test ./internal/lsp/ -v`
Expected: PASS

### Step 8: Add LSP config to `config/config.go`

```go
// Add to Config struct:
LSP LSPConfig `toml:"lsp"`

// New struct:
type LSPConfig struct {
	Enabled  bool   `toml:"enabled"`
	Language string `toml:"language"` // override auto-detect
	Command  string `toml:"command"`  // override default server
}
```

### Step 9: Wire LSP into `cmd/harness/main.go`

After MCP setup, add:
```go
// Auto-detect and start LSP server
var lspClient *lsp.Client
if cfg.LSP.Enabled || cfg.LSP.Language == "" {
	cwd, _ := os.Getwd()
	if detected := lsp.Detect(cwd); detected != nil {
		lspClient = lsp.NewClient(detected.Config, detected.RootPath)
		if err := lspClient.Start(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: LSP server failed to start: %v\n", err)
			lspClient = nil
		} else {
			defer lspClient.Close()
		}
	}
}
```

### Step 10: Add LSP badge to status bar

In `internal/tui/components/statusbar.go`, add:
```go
// In StatusBar struct:
LSPServer string // e.g. "gopls"

// In View(), after MCP badges:
var lspBadge string
if s.LSPServer != "" {
	lspBadge = " " + s.Theme.BadgeStyle().Render("⚡ LSP: " + s.LSPServer)
}
```

### Step 11: Run full build

Run: `cd /Users/robin.white/dev/dj && go build ./...`
Expected: Clean build

### Step 12: Commit

```bash
git add internal/lsp/ config/config.go cmd/harness/main.go internal/tui/components/statusbar.go go.mod go.sum
git commit -m "feat: add LSP integration with auto-detection and diagnostic feedback"
```

---

## Task 3: Checkpoint / Undo System

**Files:**
- Create: `internal/checkpoint/manager.go`
- Create: `internal/checkpoint/manager_test.go`
- Modify: `internal/tui/app.go`
- Modify: `internal/tui/screens/chat.go`

### Step 1: Write the failing test

```go
// internal/checkpoint/manager_test.go
package checkpoint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateAndRestore(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")
	os.WriteFile(file, []byte("original"), 0644)

	mgr := NewManager(20)

	// Create checkpoint
	cp, err := mgr.Before([]string{file}, "test-response-id", "Before: write test.txt")
	if err != nil {
		t.Fatalf("Before: %v", err)
	}

	// Modify the file
	os.WriteFile(file, []byte("modified"), 0644)

	// Verify file was changed
	data, _ := os.ReadFile(file)
	if string(data) != "modified" {
		t.Fatal("File should be modified")
	}

	// Restore checkpoint
	if err := mgr.Restore(cp); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	// Verify file restored
	data, _ = os.ReadFile(file)
	if string(data) != "original" {
		t.Errorf("Expected 'original' after restore, got %q", string(data))
	}
}

func TestUndoStack(t *testing.T) {
	mgr := NewManager(3)

	for i := 0; i < 5; i++ {
		mgr.Push(Checkpoint{Description: "cp"})
	}

	if mgr.Len() != 3 {
		t.Errorf("Expected max 3 checkpoints, got %d", mgr.Len())
	}
}

func TestUndoPop(t *testing.T) {
	mgr := NewManager(10)
	mgr.Push(Checkpoint{ID: "a", Description: "first"})
	mgr.Push(Checkpoint{ID: "b", Description: "second"})

	cp := mgr.Pop()
	if cp == nil {
		t.Fatal("Expected a checkpoint")
	}
	if cp.ID != "b" {
		t.Errorf("Expected 'b', got %q", cp.ID)
	}
	if mgr.Len() != 1 {
		t.Errorf("Expected 1 remaining, got %d", mgr.Len())
	}
}

func TestUndoPopEmpty(t *testing.T) {
	mgr := NewManager(10)
	cp := mgr.Pop()
	if cp != nil {
		t.Error("Expected nil from empty stack")
	}
}

func TestNewFileRestore(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "new.txt")

	// File doesn't exist yet — checkpoint records absence
	mgr := NewManager(20)
	cp, err := mgr.Before([]string{file}, "resp-1", "Before: create new.txt")
	if err != nil {
		t.Fatalf("Before: %v", err)
	}

	// Create the file
	os.WriteFile(file, []byte("new content"), 0644)

	// Restore should delete the file
	if err := mgr.Restore(cp); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Error("File should not exist after restoring to pre-creation state")
	}
}
```

### Step 2: Run tests to verify they fail

Run: `cd /Users/robin.white/dev/dj && go test ./internal/checkpoint/ -v`
Expected: FAIL — package doesn't exist

### Step 3: Write the implementation

```go
// internal/checkpoint/manager.go
package checkpoint

import (
	"fmt"
	"os"
	"time"
)

// Checkpoint records the state of files before a destructive action.
type Checkpoint struct {
	ID             string
	ResponseID     string            // Responses API previous_response_id
	Timestamp      time.Time
	FileSnapshots  map[string][]byte // path → content before mutation (nil = didn't exist)
	Description    string            // e.g. "Before: write auth/handler.go"
}

// Manager maintains a bounded stack of checkpoints.
type Manager struct {
	stack   []Checkpoint
	maxSize int
	counter int
}

func NewManager(maxSize int) *Manager {
	if maxSize <= 0 {
		maxSize = 20
	}
	return &Manager{maxSize: maxSize}
}

// Before snapshots the given files and returns a checkpoint.
// Files that don't exist are recorded with nil content (so Restore deletes them).
func (m *Manager) Before(files []string, responseID, description string) (Checkpoint, error) {
	m.counter++
	snap := make(map[string][]byte, len(files))

	for _, f := range files {
		content, err := os.ReadFile(f)
		if os.IsNotExist(err) {
			snap[f] = nil // file didn't exist — restore will delete it
		} else if err != nil {
			return Checkpoint{}, fmt.Errorf("snapshot %s: %w", f, err)
		} else {
			snap[f] = content
		}
	}

	cp := Checkpoint{
		ID:            fmt.Sprintf("cp-%d", m.counter),
		ResponseID:    responseID,
		Timestamp:     time.Now(),
		FileSnapshots: snap,
		Description:   description,
	}

	m.Push(cp)
	return cp, nil
}

// Restore reverts all files in the checkpoint to their snapshotted state.
func (m *Manager) Restore(cp Checkpoint) error {
	for path, content := range cp.FileSnapshots {
		if content == nil {
			// File didn't exist before — remove it
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove %s: %w", path, err)
			}
		} else {
			if err := os.WriteFile(path, content, 0644); err != nil {
				return fmt.Errorf("restore %s: %w", path, err)
			}
		}
	}
	return nil
}

// Push adds a checkpoint to the stack, evicting the oldest if at capacity.
func (m *Manager) Push(cp Checkpoint) {
	if len(m.stack) >= m.maxSize {
		m.stack = m.stack[1:]
	}
	m.stack = append(m.stack, cp)
}

// Pop removes and returns the most recent checkpoint.
func (m *Manager) Pop() *Checkpoint {
	if len(m.stack) == 0 {
		return nil
	}
	cp := m.stack[len(m.stack)-1]
	m.stack = m.stack[:len(m.stack)-1]
	return &cp
}

// Peek returns the most recent checkpoint without removing it.
func (m *Manager) Peek() *Checkpoint {
	if len(m.stack) == 0 {
		return nil
	}
	return &m.stack[len(m.stack)-1]
}

// Len returns the number of stored checkpoints.
func (m *Manager) Len() int {
	return len(m.stack)
}

// List returns all checkpoints (oldest first).
func (m *Manager) List() []Checkpoint {
	return m.stack
}
```

### Step 4: Run tests

Run: `cd /Users/robin.white/dev/dj && go test ./internal/checkpoint/ -v`
Expected: PASS

### Step 5: Wire Ctrl+Z into app.go

In `internal/tui/app.go`, add:
```go
// In App struct:
checkpoints *checkpoint.Manager

// In NewApp():
checkpoints: checkpoint.NewManager(20),

// In Update(), add key handler:
case "ctrl+z":
	cp := a.checkpoints.Pop()
	if cp != nil {
		if err := a.checkpoints.Restore(*cp); err == nil {
			// TODO: Show diff in chat output
			return a, func() tea.Msg {
				return screens.StreamDeltaMsg{Delta: fmt.Sprintf("\n[Restored: %s]\n", cp.Description)}
			}
		}
	}
```

### Step 6: Run full build

Run: `cd /Users/robin.white/dev/dj && go build ./...`
Expected: Clean build

### Step 7: Commit

```bash
git add internal/checkpoint/ internal/tui/app.go
git commit -m "feat: add checkpoint/undo system with Ctrl+Z restore"
```

---

## Task 4: Session Persistence & Memory Files

**Files:**
- Create: `internal/memory/manager.go`
- Create: `internal/memory/manager_test.go`
- Modify: `internal/agents/worker.go`
- Modify: `internal/tui/app.go`
- Modify: `cmd/harness/main.go`

### Step 1: Write the failing test

```go
// internal/memory/manager_test.go
package memory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadProjectMemory(t *testing.T) {
	dir := t.TempDir()
	agentsFile := filepath.Join(dir, "AGENTS.md")
	os.WriteFile(agentsFile, []byte("# Project Rules\n- No ORM"), 0644)

	mgr := NewManager(agentsFile, filepath.Join(dir, "memory.md"))
	ctx := mgr.LoadContext()

	if !strings.Contains(ctx, "No ORM") {
		t.Error("Expected project memory to contain 'No ORM'")
	}
	if !strings.Contains(ctx, "<project_memory>") {
		t.Error("Expected project memory wrapper tag")
	}
}

func TestLoadUserMemory(t *testing.T) {
	dir := t.TempDir()
	userFile := filepath.Join(dir, "memory.md")
	os.WriteFile(userFile, []byte("Prefer tabs over spaces"), 0644)

	mgr := NewManager(filepath.Join(dir, "AGENTS.md"), userFile)
	ctx := mgr.LoadContext()

	if !strings.Contains(ctx, "Prefer tabs over spaces") {
		t.Error("Expected user memory content")
	}
	if !strings.Contains(ctx, "<user_memory>") {
		t.Error("Expected user memory wrapper tag")
	}
}

func TestMissingFilesAreEmpty(t *testing.T) {
	mgr := NewManager("/nonexistent/AGENTS.md", "/nonexistent/memory.md")
	ctx := mgr.LoadContext()

	if !strings.Contains(ctx, "<project_memory>") {
		t.Error("Should still contain tags even with missing files")
	}
}

func TestAppendUserMemory(t *testing.T) {
	dir := t.TempDir()
	userFile := filepath.Join(dir, "memory.md")
	os.WriteFile(userFile, []byte("line1"), 0644)

	mgr := NewManager(filepath.Join(dir, "AGENTS.md"), userFile)
	if err := mgr.AppendUserMemory("line2"); err != nil {
		t.Fatalf("AppendUserMemory: %v", err)
	}

	data, _ := os.ReadFile(userFile)
	if !strings.Contains(string(data), "line2") {
		t.Error("Expected appended content")
	}
	if !strings.Contains(string(data), "line1") {
		t.Error("Expected original content preserved")
	}
}
```

### Step 2: Run tests to verify they fail

Run: `cd /Users/robin.white/dev/dj && go test ./internal/memory/ -v`
Expected: FAIL — package doesn't exist

### Step 3: Write the implementation

```go
// internal/memory/manager.go
package memory

import (
	"fmt"
	"os"
	"path/filepath"
)

// Manager handles project-level (AGENTS.md) and user-level (memory.md) persistence.
type Manager struct {
	projectPath string // ./AGENTS.md
	userPath    string // ~/.config/codex-harness/memory.md
}

func NewManager(projectPath, userPath string) *Manager {
	return &Manager{
		projectPath: projectPath,
		userPath:    userPath,
	}
}

// DefaultManager creates a manager with standard paths.
func DefaultManager() *Manager {
	home, _ := os.UserHomeDir()
	return &Manager{
		projectPath: "AGENTS.md",
		userPath:    filepath.Join(home, ".config", "codex-harness", "memory.md"),
	}
}

// LoadContext reads both memory files and returns formatted context for injection.
func (m *Manager) LoadContext() string {
	project := readFileOrEmpty(m.projectPath)
	user := readFileOrEmpty(m.userPath)

	return fmt.Sprintf("<project_memory>\n%s\n</project_memory>\n\n<user_memory>\n%s\n</user_memory>",
		project, user)
}

// AppendUserMemory appends a line to the user memory file, creating it if needed.
func (m *Manager) AppendUserMemory(content string) error {
	dir := filepath.Dir(m.userPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create memory dir: %w", err)
	}

	f, err := os.OpenFile(m.userPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open memory file: %w", err)
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "\n%s", content)
	return err
}

// ProjectPath returns the project memory file path.
func (m *Manager) ProjectPath() string {
	return m.projectPath
}

// UserPath returns the user memory file path.
func (m *Manager) UserPath() string {
	return m.userPath
}

func readFileOrEmpty(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}
```

### Step 4: Run tests

Run: `cd /Users/robin.white/dev/dj && go test ./internal/memory/ -v`
Expected: PASS

### Step 5: Wire memory into App and Worker

In `internal/tui/app.go`, add:
```go
// In App struct:
memory *memory.Manager

// In NewApp(), add parameter and store it.
```

In `internal/agents/worker.go`, modify `buildInstructions()`:
```go
// At the top of buildInstructions(), prepend memory context:
func (w *Worker) buildInstructions() string {
	modeCfg := Modes[w.Mode]
	base := modeCfg.SystemPrompt + "\n\n"

	// Inject memory context if available
	if w.memory != nil {
		base += w.memory.LoadContext() + "\n\n"
	}

	base += fmt.Sprintf("Subtask: %s\n", w.Task.Description)
	// ... rest unchanged
}
```

Add `memory *memory.Manager` to the `Worker` struct and `NewWorker` parameters.

### Step 6: Wire into main.go

In `cmd/harness/main.go`, after skills setup:
```go
memMgr := memory.DefaultManager()
```

Pass `memMgr` to `tui.NewApp()`.

### Step 7: Run full build

Run: `cd /Users/robin.white/dev/dj && go build ./...`
Expected: Clean build

### Step 8: Commit

```bash
git add internal/memory/ internal/agents/worker.go internal/tui/app.go cmd/harness/main.go
git commit -m "feat: add AGENTS.md and user memory persistence with context injection"
```

---

## Task 5: Event Hooks

**Files:**
- Create: `internal/hooks/runner.go`
- Create: `internal/hooks/runner_test.go`
- Modify: `config/config.go`
- Modify: `internal/tui/app.go`

### Step 1: Write the failing test

```go
// internal/hooks/runner_test.go
package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandVars(t *testing.T) {
	template := "echo '{{tool}} {{args}}'"
	vars := map[string]string{"tool": "write_file", "args": "main.go"}
	result := expandVars(template, vars)

	if !strings.Contains(result, "write_file") {
		t.Errorf("Expected 'write_file' in result, got %q", result)
	}
	if !strings.Contains(result, "main.go") {
		t.Errorf("Expected 'main.go' in result, got %q", result)
	}
}

func TestFireHook(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "hook_output.txt")

	cfg := Config{
		Hooks: map[string]string{
			string(HookPreToolCall): "echo 'pre-tool' > " + outFile,
		},
	}

	runner := NewRunner(cfg)
	err := runner.Fire(HookPreToolCall, nil)
	if err != nil {
		t.Fatalf("Fire: %v", err)
	}

	data, _ := os.ReadFile(outFile)
	if !strings.Contains(string(data), "pre-tool") {
		t.Errorf("Expected 'pre-tool' in output, got %q", string(data))
	}
}

func TestFireUnconfiguredHook(t *testing.T) {
	cfg := Config{Hooks: map[string]string{}}
	runner := NewRunner(cfg)

	// Should be a no-op, not an error
	err := runner.Fire(HookOnError, nil)
	if err != nil {
		t.Errorf("Expected nil for unconfigured hook, got %v", err)
	}
}
```

### Step 2: Run tests to verify they fail

Run: `cd /Users/robin.white/dev/dj && go test ./internal/hooks/ -v`
Expected: FAIL — package doesn't exist

### Step 3: Write the implementation

```go
// internal/hooks/runner.go
package hooks

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const hookTimeout = 10 * time.Second

// HookEvent identifies a lifecycle point.
type HookEvent string

const (
	HookPreToolCall  HookEvent = "pre_tool_call"
	HookPostToolCall HookEvent = "post_tool_call"
	HookOnError      HookEvent = "on_error"
	HookSessionEnd   HookEvent = "on_session_end"
)

// Config holds hook shell command templates.
type Config struct {
	Hooks map[string]string
}

// Runner executes configured hooks at lifecycle points.
type Runner struct {
	config Config
}

func NewRunner(cfg Config) *Runner {
	return &Runner{config: cfg}
}

// Fire executes the hook for the given event with variable substitution.
// Returns nil if no hook is configured for the event.
func (r *Runner) Fire(event HookEvent, vars map[string]string) error {
	cmdTemplate, ok := r.config.Hooks[string(event)]
	if !ok || cmdTemplate == "" {
		return nil
	}

	expanded := expandVars(cmdTemplate, vars)

	cmd := exec.Command("sh", "-c", expanded)
	cmd.WaitDelay = hookTimeout

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("hook %s failed: %w\noutput: %s", event, err, string(output))
	}

	return nil
}

// expandVars replaces {{key}} placeholders with values from vars.
func expandVars(template string, vars map[string]string) string {
	result := template
	for k, v := range vars {
		result = strings.ReplaceAll(result, "{{"+k+"}}", v)
	}
	return result
}
```

### Step 4: Run tests

Run: `cd /Users/robin.white/dev/dj && go test ./internal/hooks/ -v`
Expected: PASS

### Step 5: Add hooks to config

In `config/config.go`, add to `Config` struct:
```go
Hooks HooksConfig `toml:"hooks"`

type HooksConfig struct {
	PreToolCall  string `toml:"pre_tool_call"`
	PostToolCall string `toml:"post_tool_call"`
	OnError      string `toml:"on_error"`
	OnSessionEnd string `toml:"on_session_end"`
}
```

### Step 6: Wire hooks into app.go

In `internal/tui/app.go`:
```go
// In App struct:
hooks *hooks.Runner

// In NewApp(), add hooks parameter.

// Defer session-end hook in main.go:
defer hookRunner.Fire(hooks.HookSessionEnd, map[string]string{"summary": "session ended"})
```

### Step 7: Run full build

Run: `cd /Users/robin.white/dev/dj && go build ./...`
Expected: Clean build

### Step 8: Commit

```bash
git add internal/hooks/ config/config.go internal/tui/app.go cmd/harness/main.go
git commit -m "feat: add event hooks system with pre/post tool call and session lifecycle"
```

---

## Task 6: Context Compaction

**Files:**
- Create: `internal/agents/compactor.go`
- Create: `internal/agents/compactor_test.go`
- Modify: `internal/tui/screens/chat.go`
- Modify: `internal/tui/components/statusbar.go`

### Step 1: Write the failing test

```go
// internal/agents/compactor_test.go
package agents

import (
	"testing"

	"github.com/robinojw/dj/internal/api"
)

func TestShouldCompactBelowThreshold(t *testing.T) {
	c := NewCompactor(nil, 0.60)
	usage := api.Usage{InputTokens: 100_000} // 25% of 400K
	if c.ShouldCompact(usage) {
		t.Error("Should NOT compact at 25%")
	}
}

func TestShouldCompactAboveThreshold(t *testing.T) {
	c := NewCompactor(nil, 0.60)
	usage := api.Usage{InputTokens: 280_000} // 70% of 400K
	if !c.ShouldCompact(usage) {
		t.Error("Should compact at 70%")
	}
}

func TestShouldCompactAtExactThreshold(t *testing.T) {
	c := NewCompactor(nil, 0.60)
	usage := api.Usage{InputTokens: 240_000} // exactly 60%
	if c.ShouldCompact(usage) {
		t.Error("Should NOT compact at exactly the threshold (needs to exceed)")
	}
}

func TestBuildCompactionPrompt(t *testing.T) {
	turns := []Turn{
		{Role: "user", Content: "Fix the login bug"},
		{Role: "assistant", Content: "I'll check auth.go..."},
		{Role: "user", Content: "Also update the tests"},
	}

	prompt := buildCompactionPrompt(turns)
	if prompt == "" {
		t.Error("Expected non-empty compaction prompt")
	}
}
```

### Step 2: Run tests to verify they fail

Run: `cd /Users/robin.white/dev/dj && go test ./internal/agents/ -run TestShouldCompact -v`
Expected: FAIL — `NewCompactor`, `Turn`, `buildCompactionPrompt` undefined

### Step 3: Write the implementation

```go
// internal/agents/compactor.go
package agents

import (
	"context"
	"fmt"
	"strings"

	"github.com/robinojw/dj/internal/api"
)

const contextWindowSize = 400_000

// Turn represents a single conversation turn for compaction.
type Turn struct {
	Role    string
	Content string
}

// Compactor summarises conversation history to free up context window.
type Compactor struct {
	client    *api.ResponsesClient
	threshold float64 // fraction of context window (e.g. 0.60)
}

func NewCompactor(client *api.ResponsesClient, threshold float64) *Compactor {
	if threshold <= 0 {
		threshold = 0.60
	}
	return &Compactor{client: client, threshold: threshold}
}

// ShouldCompact returns true when input tokens exceed the threshold.
func (c *Compactor) ShouldCompact(usage api.Usage) bool {
	return float64(usage.InputTokens)/float64(contextWindowSize) > c.threshold
}

// Compact summarises the conversation history into a compressed memory block.
func (c *Compactor) Compact(ctx context.Context, history []Turn) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("no API client for compaction")
	}

	prompt := buildCompactionPrompt(history)

	req := api.CreateResponseRequest{
		Model:        "gpt-5.1-codex-mini",
		Input:        api.MakeStringInput(prompt),
		Instructions: compactionInstructions,
		Reasoning: &api.Reasoning{
			Effort: "low",
		},
	}

	resp, err := c.client.Send(ctx, req)
	if err != nil {
		return "", fmt.Errorf("compaction call: %w", err)
	}

	// Extract text from response
	var text string
	for _, item := range resp.Output {
		if item.Content != nil {
			text += string(item.Content)
		}
	}

	return text, nil
}

const compactionInstructions = `Summarise this conversation concisely, preserving:
1. Decisions made and their rationale
2. Files that were read, created, or modified
3. Current task state and next steps
4. Any errors encountered and how they were resolved
5. User preferences expressed during the conversation

Output a structured summary, not a transcript. Be concise but complete.`

func buildCompactionPrompt(turns []Turn) string {
	var sb strings.Builder
	sb.WriteString("Conversation history to summarise:\n\n")
	for _, t := range turns {
		sb.WriteString(fmt.Sprintf("[%s]: %s\n\n", t.Role, t.Content))
	}
	return sb.String()
}
```

### Step 4: Run tests

Run: `cd /Users/robin.white/dev/dj && go test ./internal/agents/ -run "TestShouldCompact|TestBuild" -v`
Expected: PASS

### Step 5: Wire compaction into chat screen

In `internal/tui/screens/chat.go`, after handling `StreamDoneMsg`:
```go
// Check if compaction is needed
// This is informational — actual compaction triggers on next submit
```

In `internal/tui/components/statusbar.go`:
```go
// In StatusBar struct:
Compacting bool

// In View(), add compacting indicator:
if s.Compacting {
	content += " " + s.Theme.AccentStyle().Render("⚡ Compacting context...")
}
```

### Step 6: Run full build

Run: `cd /Users/robin.white/dev/dj && go build ./...`
Expected: Clean build

### Step 7: Commit

```bash
git add internal/agents/compactor.go internal/agents/compactor_test.go internal/tui/screens/chat.go internal/tui/components/statusbar.go
git commit -m "feat: add context compaction when token usage exceeds 60% threshold"
```

---

## Task 7: @mention Context Injection

**Files:**
- Create: `internal/mentions/parser.go`
- Create: `internal/mentions/parser_test.go`
- Create: `internal/mentions/resolvers.go`
- Modify: `internal/tui/components/chat_input.go`
- Modify: `internal/tui/screens/chat.go`

### Step 1: Write the failing test for mention parsing

```go
// internal/mentions/parser_test.go
package mentions

import "testing"

func TestParseFileMention(t *testing.T) {
	mentions := Parse("Fix the bug in @src/auth/handler.go please")
	if len(mentions) != 1 {
		t.Fatalf("Expected 1 mention, got %d", len(mentions))
	}
	if mentions[0].Type != MentionFile {
		t.Errorf("Expected file mention, got %s", mentions[0].Type)
	}
	if mentions[0].Value != "src/auth/handler.go" {
		t.Errorf("Expected 'src/auth/handler.go', got %q", mentions[0].Value)
	}
}

func TestParseURLMention(t *testing.T) {
	mentions := Parse("Check @https://docs.stripe.com/api for reference")
	if len(mentions) != 1 {
		t.Fatalf("Expected 1 mention, got %d", len(mentions))
	}
	if mentions[0].Type != MentionURL {
		t.Errorf("Expected URL mention, got %s", mentions[0].Type)
	}
}

func TestParseFunctionMention(t *testing.T) {
	mentions := Parse("Look at @fn:CreateSession")
	if len(mentions) != 1 {
		t.Fatalf("Expected 1 mention, got %d", len(mentions))
	}
	if mentions[0].Type != MentionFunction {
		t.Errorf("Expected function mention, got %s", mentions[0].Type)
	}
	if mentions[0].Value != "CreateSession" {
		t.Errorf("Expected 'CreateSession', got %q", mentions[0].Value)
	}
}

func TestParseGitMention(t *testing.T) {
	mentions := Parse("Show changes from @git:HEAD~3")
	if len(mentions) != 1 {
		t.Fatalf("Expected 1 mention, got %d", len(mentions))
	}
	if mentions[0].Type != MentionGit {
		t.Errorf("Expected git mention, got %s", mentions[0].Type)
	}
	if mentions[0].Value != "HEAD~3" {
		t.Errorf("Expected 'HEAD~3', got %q", mentions[0].Value)
	}
}

func TestParseTestMention(t *testing.T) {
	mentions := Parse("Run @test:TestAuthFlow and show me the output")
	if len(mentions) != 1 {
		t.Fatalf("Expected 1 mention, got %d", len(mentions))
	}
	if mentions[0].Type != MentionTest {
		t.Errorf("Expected test mention, got %s", mentions[0].Type)
	}
}

func TestParseMultipleMentions(t *testing.T) {
	mentions := Parse("Compare @src/old.go with @src/new.go")
	if len(mentions) != 2 {
		t.Fatalf("Expected 2 mentions, got %d", len(mentions))
	}
}

func TestParseNoMentions(t *testing.T) {
	mentions := Parse("Just a normal message with email user@example.com")
	if len(mentions) != 0 {
		t.Errorf("Expected 0 mentions, got %d", len(mentions))
	}
}

func TestStripMentions(t *testing.T) {
	input := "Fix @src/auth.go and run @test:TestAuth"
	stripped := StripMentions(input)
	if stripped != "Fix  and run " {
		t.Errorf("Expected mentions stripped, got %q", stripped)
	}
}
```

### Step 2: Run tests to verify they fail

Run: `cd /Users/robin.white/dev/dj && go test ./internal/mentions/ -v`
Expected: FAIL — package doesn't exist

### Step 3: Write the mention parser

```go
// internal/mentions/parser.go
package mentions

import (
	"regexp"
	"strings"
)

// MentionType categorizes the kind of @mention.
type MentionType string

const (
	MentionFile     MentionType = "file"
	MentionURL      MentionType = "url"
	MentionFunction MentionType = "function"
	MentionGit      MentionType = "git"
	MentionTest     MentionType = "test"
)

// Mention represents a parsed @mention from user input.
type Mention struct {
	Type     MentionType
	Value    string // the parsed value (path, URL, symbol name, ref, test name)
	Raw      string // the full raw text matched including @
	StartIdx int
	EndIdx   int
}

// ResolvedMention is a mention with its fetched content.
type ResolvedMention struct {
	Mention
	Content string
	Error   error
}

// parser is a typed prefix handler.
type parser struct {
	prefix  string
	typ     MentionType
	extract func(string) string
}

var parsers = []parser{
	{prefix: "@fn:", typ: MentionFunction, extract: func(s string) string { return s }},
	{prefix: "@git:", typ: MentionGit, extract: func(s string) string { return s }},
	{prefix: "@test:", typ: MentionTest, extract: func(s string) string { return s }},
	{prefix: "@https://", typ: MentionURL, extract: func(s string) string { return "https://" + s }},
	{prefix: "@http://", typ: MentionURL, extract: func(s string) string { return "http://" + s }},
	// File mention is the fallback — any @ followed by a path-like string
}

// mentionRegex matches @-prefixed tokens, avoiding email addresses.
// It requires @ to be at the start of the string or preceded by whitespace.
var mentionRegex = regexp.MustCompile(`(?:^|\s)(@(?:fn:|git:|test:|https?://|)[^\s,;]+)`)

// Parse extracts all @mentions from the input string.
func Parse(input string) []Mention {
	var mentions []Mention

	matches := mentionRegex.FindAllStringSubmatchIndex(input, -1)
	for _, match := range matches {
		// match[2] and match[3] are the submatch (group 1) indices
		raw := input[match[2]:match[3]]

		m := classify(raw)
		if m != nil {
			m.StartIdx = match[2]
			m.EndIdx = match[3]
			mentions = append(mentions, *m)
		}
	}

	return mentions
}

// StripMentions removes all @mention tokens from the input.
func StripMentions(input string) string {
	mentions := Parse(input)
	if len(mentions) == 0 {
		return input
	}

	// Remove from right to left to preserve indices
	result := input
	for i := len(mentions) - 1; i >= 0; i-- {
		m := mentions[i]
		result = result[:m.StartIdx] + result[m.EndIdx:]
	}
	return result
}

func classify(raw string) *Mention {
	// Strip leading @
	body := raw[1:]

	// Try typed prefixes first
	for _, p := range parsers {
		prefix := p.prefix[1:] // strip the @ since we already removed it
		if strings.HasPrefix(body, prefix) {
			value := body[len(prefix):]
			return &Mention{
				Type:  p.typ,
				Value: p.extract(value),
				Raw:   raw,
			}
		}
	}

	// Fallback: file mention if it looks like a path
	if isPathLike(body) {
		return &Mention{
			Type:  MentionFile,
			Value: body,
			Raw:   raw,
		}
	}

	return nil
}

func isPathLike(s string) bool {
	// Must contain a / or a file extension to be a path
	return strings.Contains(s, "/") || strings.Contains(s, ".")
}
```

### Step 4: Run tests

Run: `cd /Users/robin.white/dev/dj && go test ./internal/mentions/ -v`
Expected: PASS (or iterate on regex edge cases)

### Step 5: Write the resolvers

```go
// internal/mentions/resolvers.go
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

		ctx, cancel := context.WithTimeout(ctx, resolveTimeout)
		switch m.Type {
		case MentionFile:
			resolved[i].Content, resolved[i].Error = resolveFile(m.Value)
		case MentionURL:
			resolved[i].Content, resolved[i].Error = resolveURL(ctx, m.Value)
		case MentionFunction:
			resolved[i].Content, resolved[i].Error = resolveFunction(m.Value)
		case MentionGit:
			resolved[i].Content, resolved[i].Error = resolveGit(ctx, m.Value)
		case MentionTest:
			resolved[i].Content, resolved[i].Error = resolveTest(ctx, m.Value)
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
```

### Step 6: Wire @mentions into chat_input.go

In `internal/tui/components/chat_input.go`, add a new mode:

```go
const (
	ModeNormal InputMode = iota
	ModeSkillComplete
	ModeMentionComplete  // NEW: @ trigger
)
```

Add mention detection in the `Update` method alongside the existing `$` skill trigger:
```go
// After the $ skill trigger check, add:
if strings.HasSuffix(val, "@") {
	c.mode = ModeMentionComplete
	return c, nil
}
```

### Step 7: Wire @mention resolution into chat.go

In `internal/tui/screens/chat.go`, modify the submit handling:

```go
// In the SubmitMsg handler, before sending to API:
// 1. Parse mentions from text
// 2. Resolve them
// 3. Strip mentions from display text
// 4. Append resolved content to the API input
```

### Step 8: Write a test for the full parse → resolve → format pipeline

```go
// Add to internal/mentions/parser_test.go:
func TestFormatResolved(t *testing.T) {
	resolved := []ResolvedMention{
		{
			Mention: Mention{Type: MentionFile, Value: "main.go"},
			Content: "package main\n\nfunc main() {}",
		},
		{
			Mention: Mention{Type: MentionGit, Value: "HEAD~1"},
			Error:   fmt.Errorf("not a git repo"),
		},
	}

	output := FormatResolved(resolved)
	if !strings.Contains(output, "package main") {
		t.Error("Expected file content in output")
	}
	if !strings.Contains(output, "not a git repo") {
		t.Error("Expected error message in output")
	}
}
```

### Step 9: Run all tests

Run: `cd /Users/robin.white/dev/dj && go test ./internal/mentions/ -v`
Expected: PASS

### Step 10: Run full build

Run: `cd /Users/robin.white/dev/dj && go build ./...`
Expected: Clean build

### Step 11: Commit

```bash
git add internal/mentions/ internal/tui/components/chat_input.go internal/tui/screens/chat.go
git commit -m "feat: add @mention context injection (file, URL, function, git, test)"
```

---

## Summary

| Task | Feature | New Files | Modified Files |
|------|---------|-----------|----------------|
| 1 | Plan/Build modes | 2 | 4 |
| 2 | LSP integration | 4 | 4 |
| 3 | Checkpoint/undo | 2 | 2 |
| 4 | Memory persistence | 2 | 3 |
| 5 | Event hooks | 2 | 3 |
| 6 | Context compaction | 2 | 2 |
| 7 | @mention injection | 3 | 2 |
| **Total** | | **17 new** | **~12 modified** |

Each task is self-contained with tests first, implementation second, wiring third, commit last. Tasks 1–5 can be parallelized (no dependencies). Task 6 depends on the tracker being accurate (no code dependency, just conceptual). Task 7 is fully independent.
