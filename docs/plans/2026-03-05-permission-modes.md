# Three-Mode Permission System Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace dual-mode system (Plan/Build) with three-mode permission architecture (Confirm/Plan/Turbo) featuring two-layer gate system and permission modals.

**Architecture:** Two-layer gate: Layer 1 filters tools before API requests, Layer 2 intercepts tool execution at runtime. Permission requests suspend worker goroutines and show TUI modals. Allow/deny lists with glob matching provide defense-in-depth.

**Tech Stack:** Go 1.25.4, bubbletea TUI, BurntSushi/toml config parsing

---

## Task 1: Create Modes Package Foundation

**Files:**
- Create: `internal/modes/types.go`
- Create: `internal/modes/types_test.go`

**Step 1: Write test for ExecutionMode enum**

Create `internal/modes/types_test.go`:

```go
package modes

import "testing"

func TestExecutionMode_String(t *testing.T) {
	tests := []struct {
		mode ExecutionMode
		want string
	}{
		{ModeConfirm, "Confirm"},
		{ModePlan, "Plan"},
		{ModeTurbo, "Turbo"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecutionMode_StatusLabel(t *testing.T) {
	tests := []struct {
		mode ExecutionMode
		want string
	}{
		{ModeConfirm, "⏸ CONFIRM"},
		{ModePlan, "◎ PLAN"},
		{ModeTurbo, "⚡ TURBO"},
	}

	for _, tt := range tests {
		t.Run(tt.mode.String(), func(t *testing.T) {
			if got := tt.mode.StatusLabel(); got != tt.want {
				t.Errorf("StatusLabel() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/modes -v`
Expected: FAIL with "no such file or directory"

**Step 3: Implement ExecutionMode types**

Create `internal/modes/types.go`:

```go
package modes

// ExecutionMode determines tool access and permission behavior.
type ExecutionMode int

const (
	ModeConfirm ExecutionMode = iota // ask before write/exec/MCP tools
	ModePlan                          // read-only, high reasoning
	ModeTurbo                         // bypass all permissions
)

func (m ExecutionMode) String() string {
	switch m {
	case ModeConfirm:
		return "Confirm"
	case ModePlan:
		return "Plan"
	case ModeTurbo:
		return "Turbo"
	default:
		return "Unknown"
	}
}

// StatusLabel returns the badge text for the status bar.
func (m ExecutionMode) StatusLabel() string {
	switch m {
	case ModeConfirm:
		return "⏸ CONFIRM"
	case ModePlan:
		return "◎ PLAN"
	case ModeTurbo:
		return "⚡ TURBO"
	default:
		return "UNKNOWN"
	}
}

// ToolClass categorizes tools for permission decisions.
type ToolClass int

const (
	ToolRead      ToolClass = iota // read_file, list_dir, search_code
	ToolWrite                      // write_file, create_file, delete_file
	ToolExec                       // bash, run_script, run_tests
	ToolMCPMutate                  // MCP tools that modify state
	ToolMCPRead                    // MCP tools flagged read-only
	ToolNetwork                    // web_fetch, http_request
)

// GateDecision is the result of gate evaluation.
type GateDecision int

const (
	GateAllow   GateDecision = iota // execute immediately
	GateDeny                         // block execution
	GateAskUser                      // show permission modal
)

// ModeConfig holds the system prompt and reasoning effort for a mode.
type ModeConfig struct {
	Mode            ExecutionMode
	AllowedTools    []string // nil = all tools (Turbo/Confirm), specific list for Plan
	SystemPrompt    string
	ReasoningEffort string
}

var planSystemPrompt = `You are a senior software architect in Plan mode.
Your job is to analyze, reason, and produce a detailed implementation plan.
You may ONLY read files, search code, and list directories.
You may NOT write files, execute commands, or invoke MCP tools that mutate state.
Think deeply. Output a numbered, step-by-step implementation plan with exact file paths.`

var confirmSystemPrompt = `You are a skilled software engineer in Confirm mode.
You have access to all tools, but destructive operations require user permission.
Make focused, incremental changes. Explain your intent before executing risky operations.
Run tests after edits to verify correctness.`

var turboSystemPrompt = `You are a skilled software engineer in Turbo mode.
You have full autonomy - all tools are available without permission checks.
Work efficiently. Make minimal, focused changes. Run tests after each edit.`

// Modes maps each ExecutionMode to its configuration.
var Modes = map[ExecutionMode]ModeConfig{
	ModePlan: {
		Mode:            ModePlan,
		AllowedTools:    []string{"read_file", "search_code", "list_dir"},
		SystemPrompt:    planSystemPrompt,
		ReasoningEffort: "high",
	},
	ModeConfirm: {
		Mode:            ModeConfirm,
		AllowedTools:    nil, // all tools available
		SystemPrompt:    confirmSystemPrompt,
		ReasoningEffort: "medium",
	},
	ModeTurbo: {
		Mode:            ModeTurbo,
		AllowedTools:    nil, // all tools available
		SystemPrompt:    turboSystemPrompt,
		ReasoningEffort: "medium",
	},
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/modes -v`
Expected: PASS (2 tests)

**Step 5: Commit**

```bash
git add internal/modes/
git commit -m "feat(modes): add ExecutionMode enum with Confirm/Plan/Turbo"
```

---

## Task 2: Add Tool Classification

**Files:**
- Modify: `internal/modes/types.go`
- Modify: `internal/modes/types_test.go`

**Step 1: Write test for tool classification**

Add to `internal/modes/types_test.go`:

```go
func TestClassifyTool(t *testing.T) {
	tests := []struct {
		tool string
		want ToolClass
	}{
		{"read_file", ToolRead},
		{"list_dir", ToolRead},
		{"search_code", ToolRead},
		{"write_file", ToolWrite},
		{"create_file", ToolWrite},
		{"delete_file", ToolWrite},
		{"bash", ToolExec},
		{"run_script", ToolExec},
		{"run_tests", ToolExec},
		{"web_fetch", ToolNetwork},
		{"http_request", ToolNetwork},
		{"unknown_tool", ToolWrite}, // default to conservative
	}

	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			if got := ClassifyTool(tt.tool); got != tt.want {
				t.Errorf("ClassifyTool(%q) = %v, want %v", tt.tool, got, tt.want)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/modes -v`
Expected: FAIL with "undefined: ClassifyTool"

**Step 3: Implement tool classification**

Add to `internal/modes/types.go`:

```go
// toolClasses maps tool names to their security classification.
var toolClasses = map[string]ToolClass{
	"read_file":    ToolRead,
	"list_dir":     ToolRead,
	"search_code":  ToolRead,
	"write_file":   ToolWrite,
	"create_file":  ToolWrite,
	"delete_file":  ToolWrite,
	"bash":         ToolExec,
	"run_script":   ToolExec,
	"run_tests":    ToolExec,
	"web_fetch":    ToolNetwork,
	"http_request": ToolNetwork,
}

// ClassifyTool returns the security class of a tool.
// Unknown tools default to ToolWrite (conservative).
func ClassifyTool(toolName string) ToolClass {
	if class, ok := toolClasses[toolName]; ok {
		return class
	}
	return ToolWrite // default to conservative
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/modes -v`
Expected: PASS (3 tests)

**Step 5: Commit**

```bash
git add internal/modes/
git commit -m "feat(modes): add tool classification with conservative defaults"
```

---

## Task 3: Implement Glob Pattern Matching

**Files:**
- Create: `internal/modes/glob.go`
- Create: `internal/modes/glob_test.go`

**Step 1: Write test for glob matching**

Create `internal/modes/glob_test.go`:

```go
package modes

import "testing"

func TestMatchGlob(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		want    bool
	}{
		{"bash(git status*)", "bash(git status)", true},
		{"bash(git status*)", "bash(git status --short)", true},
		{"bash(git status*)", "bash(git diff)", false},
		{"read_file(.env*)", "read_file(.env)", true},
		{"read_file(.env*)", "read_file(.env.local)", true},
		{"read_file(.env*)", "read_file(config.toml)", false},
		{"exact_match", "exact_match", true},
		{"exact_match", "exact_match_not", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"~"+tt.input, func(t *testing.T) {
			if got := MatchGlob(tt.pattern, tt.input); got != tt.want {
				t.Errorf("MatchGlob(%q, %q) = %v, want %v", tt.pattern, tt.input, got, tt.want)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/modes -v`
Expected: FAIL with "undefined: MatchGlob"

**Step 3: Implement glob matching**

Create `internal/modes/glob.go`:

```go
package modes

import "strings"

// MatchGlob checks if input matches the glob pattern.
// Supports * wildcard only (not full regex).
func MatchGlob(pattern, input string) bool {
	// Exact match if no wildcards
	if !strings.Contains(pattern, "*") {
		return pattern == input
	}

	// Split on * and check parts in order
	parts := strings.Split(pattern, "*")

	// Must start with first part
	if !strings.HasPrefix(input, parts[0]) {
		return false
	}
	input = strings.TrimPrefix(input, parts[0])

	// Check middle parts in order
	for i := 1; i < len(parts)-1; i++ {
		idx := strings.Index(input, parts[i])
		if idx == -1 {
			return false
		}
		input = input[idx+len(parts[i]):]
	}

	// Must end with last part (unless it's empty from trailing *)
	if len(parts) > 1 && parts[len(parts)-1] != "" {
		if !strings.HasSuffix(input, parts[len(parts)-1]) {
			return false
		}
	}

	return true
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/modes -v`
Expected: PASS (4 tests)

**Step 5: Commit**

```bash
git add internal/modes/
git commit -m "feat(modes): add glob pattern matching for allow/deny lists"
```

---

## Task 4: Implement Gate with Allow/Deny Lists

**Files:**
- Create: `internal/modes/gate.go`
- Create: `internal/modes/gate_test.go`

**Step 1: Write test for gate evaluation**

Create `internal/modes/gate_test.go`:

```go
package modes

import "testing"

func TestGate_Evaluate_DenyList(t *testing.T) {
	gate := NewGate(ModeConfirm, []string{}, []string{"bash(rm -rf*)"})

	// Deny list always wins
	if got := gate.Evaluate("bash(rm -rf /)", nil); got != GateDeny {
		t.Errorf("Expected GateDeny for deny list match, got %v", got)
	}
}

func TestGate_Evaluate_AllowList(t *testing.T) {
	gate := NewGate(ModeConfirm, []string{"write_file"}, []string{})

	// Allow list passes even for write tools in Confirm
	if got := gate.Evaluate("write_file", nil); got != GateAllow {
		t.Errorf("Expected GateAllow for allow list match, got %v", got)
	}
}

func TestGate_Evaluate_ConfirmMode(t *testing.T) {
	gate := NewGate(ModeConfirm, []string{}, []string{})

	tests := []struct {
		tool string
		want GateDecision
	}{
		{"read_file", GateAllow},
		{"write_file", GateAskUser},
		{"bash", GateAskUser},
	}

	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			if got := gate.Evaluate(tt.tool, nil); got != tt.want {
				t.Errorf("Evaluate(%q) = %v, want %v", tt.tool, got, tt.want)
			}
		})
	}
}

func TestGate_Evaluate_PlanMode(t *testing.T) {
	gate := NewGate(ModePlan, []string{}, []string{})

	tests := []struct {
		tool string
		want GateDecision
	}{
		{"read_file", GateAllow},
		{"write_file", GateDeny},
		{"bash", GateDeny},
	}

	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			if got := gate.Evaluate(tt.tool, nil); got != tt.want {
				t.Errorf("Evaluate(%q) = %v, want %v", tt.tool, got, tt.want)
			}
		})
	}
}

func TestGate_Evaluate_TurboMode(t *testing.T) {
	gate := NewGate(ModeTurbo, []string{}, []string{})

	// Turbo allows everything (except deny list)
	if got := gate.Evaluate("write_file", nil); got != GateAllow {
		t.Errorf("Expected GateAllow in Turbo mode, got %v", got)
	}
	if got := gate.Evaluate("bash", nil); got != GateAllow {
		t.Errorf("Expected GateAllow in Turbo mode, got %v", got)
	}
}

func TestGate_AllowForSession(t *testing.T) {
	gate := NewGate(ModeConfirm, []string{}, []string{})

	// Initially asks
	if got := gate.Evaluate("write_file", nil); got != GateAskUser {
		t.Errorf("Expected GateAskUser before session allow, got %v", got)
	}

	// Add to session allow list
	gate.AllowForSession("write_file")

	// Now allows
	if got := gate.Evaluate("write_file", nil); got != GateAllow {
		t.Errorf("Expected GateAllow after session allow, got %v", got)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/modes -v`
Expected: FAIL with "undefined: NewGate"

**Step 3: Implement Gate**

Create `internal/modes/gate.go`:

```go
package modes

// Gate controls tool access through allow/deny lists and mode rules.
type Gate struct {
	mode      ExecutionMode
	allowList []string // session or persisted
	denyList  []string // from config
}

// NewGate creates a gate with the given mode and lists.
func NewGate(mode ExecutionMode, allowList, denyList []string) *Gate {
	return &Gate{
		mode:      mode,
		allowList: allowList,
		denyList:  denyList,
	}
}

// SetMode updates the gate's execution mode.
func (g *Gate) SetMode(mode ExecutionMode) {
	g.mode = mode
}

// Evaluate determines whether a tool call should be allowed.
func (g *Gate) Evaluate(toolName string, args map[string]any) GateDecision {
	// 1. Deny list always wins
	if g.isDenied(toolName) {
		return GateDeny
	}

	// 2. Allow list passes
	if g.isAllowed(toolName) {
		return GateAllow
	}

	// 3. Mode-specific logic
	class := ClassifyTool(toolName)

	switch g.mode {
	case ModeTurbo:
		return GateAllow
	case ModePlan:
		if class == ToolRead || class == ToolMCPRead {
			return GateAllow
		}
		return GateDeny
	case ModeConfirm:
		if class == ToolRead || class == ToolMCPRead {
			return GateAllow
		}
		return GateAskUser
	default:
		return GateDeny
	}
}

// isDenied checks if tool matches deny list (with glob).
func (g *Gate) isDenied(toolName string) bool {
	for _, pattern := range g.denyList {
		if MatchGlob(pattern, toolName) {
			return true
		}
	}
	return false
}

// isAllowed checks if tool matches allow list (with glob).
func (g *Gate) isAllowed(toolName string) bool {
	for _, pattern := range g.allowList {
		if MatchGlob(pattern, toolName) {
			return true
		}
	}
	return false
}

// AllowForSession adds a tool to the session allow list.
func (g *Gate) AllowForSession(toolName string) {
	g.allowList = append(g.allowList, toolName)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/modes -v`
Expected: PASS (9 tests)

**Step 5: Commit**

```bash
git add internal/modes/
git commit -m "feat(modes): implement gate with allow/deny lists and mode logic"
```

---

## Task 5: Add Permission Request Types

**Files:**
- Create: `internal/modes/permission.go`
- Create: `internal/modes/permission_test.go`

**Step 1: Write test for permission flow**

Create `internal/modes/permission_test.go`:

```go
package modes

import (
	"testing"
	"time"
)

func TestPermissionRequest_ResponseFlow(t *testing.T) {
	req := PermissionRequest{
		ID:       "test-123",
		WorkerID: "worker-1",
		Tool:     "write_file",
		Args:     map[string]any{"path": "test.go"},
		RespCh:   make(chan PermissionResp, 1),
	}

	// Simulate user approval
	go func() {
		req.RespCh <- PermissionResp{
			Allowed:     true,
			RememberFor: RememberSession,
		}
	}()

	// Worker blocks waiting for response
	resp := <-req.RespCh

	if !resp.Allowed {
		t.Error("Expected approval")
	}
	if resp.RememberFor != RememberSession {
		t.Errorf("Expected RememberSession, got %v", resp.RememberFor)
	}
}

func TestPermissionRequest_Timeout(t *testing.T) {
	req := PermissionRequest{
		ID:     "test-456",
		RespCh: make(chan PermissionResp, 1),
	}

	// Simulate timeout
	select {
	case <-req.RespCh:
		t.Error("Should not receive response")
	case <-time.After(10 * time.Millisecond):
		// Expected timeout
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/modes -v`
Expected: FAIL with "undefined: PermissionRequest"

**Step 3: Implement permission types**

Create `internal/modes/permission.go`:

```go
package modes

// PermissionRequest is sent from worker to TUI when permission is needed.
type PermissionRequest struct {
	ID       string              // unique request ID
	WorkerID string              // worker making the request
	Tool     string              // tool name (e.g. "write_file")
	Args     map[string]any      // tool arguments
	RespCh   chan PermissionResp // response channel (worker blocks on this)
}

// PermissionResp is sent from TUI to worker after user decision.
type PermissionResp struct {
	Allowed     bool          // true if user approved
	RememberFor RememberScope // how to remember this decision
}

// RememberScope determines how long a permission decision lasts.
type RememberScope int

const (
	RememberOnce    RememberScope = iota // allow this single call
	RememberSession                      // allow for current session
	RememberAlways                       // persist to config
)

func (r RememberScope) String() string {
	switch r {
	case RememberOnce:
		return "Once"
	case RememberSession:
		return "Session"
	case RememberAlways:
		return "Always"
	default:
		return "Unknown"
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/modes -v`
Expected: PASS (11 tests)

**Step 5: Commit**

```bash
git add internal/modes/
git commit -m "feat(modes): add permission request/response types"
```

---

## Task 6: Add Execution Config to Config Package

**Files:**
- Modify: `config/config.go`
- Modify: `config/config_test.go`

**Step 1: Write test for execution config parsing**

Add to `config/config_test.go`:

```go
func TestExecutionConfig_Parse(t *testing.T) {
	tomlData := `
[execution]
default_mode = "confirm"

[execution.allow]
tools = ["read_file", "bash(git status*)"]

[execution.deny]
tools = ["bash(rm -rf*)", "write_file(.env*)"]
`
	var cfg Config
	if _, err := toml.Decode(tomlData, &cfg); err != nil {
		t.Fatalf("Failed to parse TOML: %v", err)
	}

	if cfg.Execution.DefaultMode != "confirm" {
		t.Errorf("Expected default_mode=confirm, got %q", cfg.Execution.DefaultMode)
	}
	if len(cfg.Execution.Allow.Tools) != 2 {
		t.Errorf("Expected 2 allow tools, got %d", len(cfg.Execution.Allow.Tools))
	}
	if len(cfg.Execution.Deny.Tools) != 2 {
		t.Errorf("Expected 2 deny tools, got %d", len(cfg.Execution.Deny.Tools))
	}
}

func TestExecutionConfig_DefaultMode(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Execution.DefaultMode != "confirm" {
		t.Errorf("Expected default mode=confirm, got %q", cfg.Execution.DefaultMode)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./config -v`
Expected: FAIL with "undefined: cfg.Execution"

**Step 3: Add execution config to config.go**

Modify `config/config.go`:

```go
type Config struct {
	Model     ModelConfig      `toml:"model"`
	Theme     ThemeConfig      `toml:"theme"`
	MCP       MCPConfig        `toml:"mcp"`
	Skills    SkillsConfig     `toml:"skills"`
	LSP       LSPConfig        `toml:"lsp"`
	Hooks     HooksConfig      `toml:"hooks"`
	Execution ExecutionConfig  `toml:"execution"`  // ADD THIS
}

// ADD THESE TYPES:

type ExecutionConfig struct {
	DefaultMode string      `toml:"default_mode"`
	Allow       AllowConfig `toml:"allow"`
	Deny        DenyConfig  `toml:"deny"`
}

type AllowConfig struct {
	Tools []string `toml:"tools"`
}

type DenyConfig struct {
	Tools []string `toml:"tools"`
}
```

Update `DefaultConfig()`:

```go
func DefaultConfig() Config {
	return Config{
		Model: ModelConfig{
			Default:         "gpt-5.1-codex-mini",
			ReasoningEffort: "medium",
			TeamThreshold:   3,
		},
		Theme: ThemeConfig{
			Name: "tokyonight",
		},
		Skills: SkillsConfig{
			Paths: []string{
				"./.codex/skills",
				"~/.config/codex-harness/skills",
			},
		},
		Execution: ExecutionConfig{  // ADD THIS
			DefaultMode: "confirm",
			Allow:       AllowConfig{Tools: []string{}},
			Deny:        DenyConfig{Tools: []string{}},
		},
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./config -v`
Expected: PASS

**Step 5: Commit**

```bash
git add config/
git commit -m "feat(config): add execution config with allow/deny lists"
```

---

## Task 7: Create Permission Modal Component

**Files:**
- Create: `internal/tui/components/permission_modal.go`
- Create: `internal/tui/components/permission_modal_test.go`

**Step 1: Write test for permission modal**

Create `internal/tui/components/permission_modal_test.go`:

```go
package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/modes"
	"github.com/robinojw/dj/internal/tui/theme"
)

func TestPermissionModal_ShowHide(t *testing.T) {
	modal := NewPermissionModal(theme.NewTheme("tokyonight"))

	if modal.Visible() {
		t.Error("Modal should be hidden initially")
	}

	req := modes.PermissionRequest{
		ID:   "test-1",
		Tool: "write_file",
		Args: map[string]any{"path": "test.go"},
		RespCh: make(chan modes.PermissionResp, 1),
	}

	modal, _ = modal.Update(req)

	if !modal.Visible() {
		t.Error("Modal should be visible after request")
	}
}

func TestPermissionModal_ScopeCycle(t *testing.T) {
	modal := NewPermissionModal(theme.NewTheme("tokyonight"))

	req := modes.PermissionRequest{
		ID:     "test-2",
		Tool:   "bash",
		RespCh: make(chan modes.PermissionResp, 1),
	}
	modal, _ = modal.Update(req)

	// Initial scope is Once
	if modal.scope != modes.RememberOnce {
		t.Errorf("Expected RememberOnce, got %v", modal.scope)
	}

	// Tab cycles to Session
	modal, _ = modal.Update(tea.KeyMsg{Type: tea.KeyTab})
	if modal.scope != modes.RememberSession {
		t.Errorf("Expected RememberSession, got %v", modal.scope)
	}

	// Tab cycles to Always
	modal, _ = modal.Update(tea.KeyMsg{Type: tea.KeyTab})
	if modal.scope != modes.RememberAlways {
		t.Errorf("Expected RememberAlways, got %v", modal.scope)
	}

	// Tab cycles back to Once
	modal, _ = modal.Update(tea.KeyMsg{Type: tea.KeyTab})
	if modal.scope != modes.RememberOnce {
		t.Errorf("Expected RememberOnce, got %v", modal.scope)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/components -v`
Expected: FAIL with "undefined: NewPermissionModal"

**Step 3: Implement permission modal**

Create `internal/tui/components/permission_modal.go`:

```go
package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/modes"
	"github.com/robinojw/dj/internal/tui/theme"
)

type PermissionModal struct {
	request *modes.PermissionRequest
	scope   modes.RememberScope
	visible bool
	theme   *theme.Theme
}

func NewPermissionModal(t *theme.Theme) PermissionModal {
	return PermissionModal{
		scope: modes.RememberOnce,
		theme: t,
	}
}

func (m PermissionModal) Visible() bool {
	return m.visible
}

func (m PermissionModal) Update(msg tea.Msg) (PermissionModal, tea.Cmd) {
	switch msg := msg.(type) {
	case modes.PermissionRequest:
		m.request = &msg
		m.visible = true
		m.scope = modes.RememberOnce
		return m, nil

	case tea.KeyMsg:
		if !m.visible || m.request == nil {
			return m, nil
		}

		switch msg.String() {
		case "y":
			// Approve
			m.request.RespCh <- modes.PermissionResp{
				Allowed:     true,
				RememberFor: m.scope,
			}
			m.visible = false
			m.request = nil
			return m, nil

		case "n", "esc":
			// Deny
			m.request.RespCh <- modes.PermissionResp{
				Allowed: false,
			}
			m.visible = false
			m.request = nil
			return m, nil

		case "tab":
			// Cycle scope: Once → Session → Always → Once
			m.scope = (m.scope + 1) % 3
			return m, nil
		}
	}

	return m, nil
}

func (m PermissionModal) View() string {
	if !m.visible || m.request == nil {
		return ""
	}

	// Tool name and args
	toolLine := fmt.Sprintf("🔧 %s", m.request.Tool)
	argsLines := formatArgs(m.request.Args)

	// Scope indicators
	scopeOptions := []string{
		renderScope(modes.RememberOnce, m.scope == modes.RememberOnce),
		renderScope(modes.RememberSession, m.scope == modes.RememberSession),
		renderScope(modes.RememberAlways, m.scope == modes.RememberAlways),
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		"",
		fmt.Sprintf("Worker %s wants to run:", m.request.WorkerID),
		"",
		toolLine,
		strings.Repeat("─", 60),
		argsLines,
		"",
		strings.Repeat("═", 60),
		"Remember this decision?",
		strings.Join(scopeOptions, "   "),
		strings.Repeat("═", 60),
		"[y] Allow    [n] Deny    [Tab] cycle scope    [Esc] Deny",
	)

	return m.theme.PanelStyle().
		Width(64).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("214")).
		Render(content)
}

func renderScope(scope modes.RememberScope, selected bool) string {
	label := scope.String()
	if selected {
		return fmt.Sprintf("● %s", label)
	}
	return fmt.Sprintf("○ %s", label)
}

func formatArgs(args map[string]any) string {
	if len(args) == 0 {
		return "(no arguments)"
	}

	var lines []string
	for k, v := range args {
		lines = append(lines, fmt.Sprintf("%s: %v", k, v))
	}
	return strings.Join(lines, "\n")
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/components -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/components/
git commit -m "feat(tui): add permission modal component"
```

---

## Task 8: Create Turbo Warning Modal Component

**Files:**
- Create: `internal/tui/components/turbo_modal.go`
- Create: `internal/tui/components/turbo_modal_test.go`

**Step 1: Write test for turbo modal**

Create `internal/tui/components/turbo_modal_test.go`:

```go
package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/tui/theme"
)

func TestTurboModal_ShowHide(t *testing.T) {
	modal := NewTurboModal(theme.NewTheme("tokyonight"))

	if modal.Visible() {
		t.Error("Modal should be hidden initially")
	}

	modal.Show()

	if !modal.Visible() {
		t.Error("Modal should be visible after Show()")
	}
}

func TestTurboModal_Confirm(t *testing.T) {
	modal := NewTurboModal(theme.NewTheme("tokyonight"))
	modal.Show()

	respCh := make(chan bool, 1)
	modal.SetResponseChannel(respCh)

	// Press 'y' to confirm
	modal, _ = modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	select {
	case confirmed := <-respCh:
		if !confirmed {
			t.Error("Expected confirmation")
		}
	default:
		t.Error("No response received")
	}

	if modal.Visible() {
		t.Error("Modal should be hidden after confirmation")
	}
}

func TestTurboModal_Cancel(t *testing.T) {
	modal := NewTurboModal(theme.NewTheme("tokyonight"))
	modal.Show()

	respCh := make(chan bool, 1)
	modal.SetResponseChannel(respCh)

	// Press 'n' to cancel
	modal, _ = modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	select {
	case confirmed := <-respCh:
		if confirmed {
			t.Error("Expected cancellation")
		}
	default:
		t.Error("No response received")
	}

	if modal.Visible() {
		t.Error("Modal should be hidden after cancellation")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/components -v`
Expected: FAIL with "undefined: NewTurboModal"

**Step 3: Implement turbo modal**

Create `internal/tui/components/turbo_modal.go`:

```go
package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/tui/theme"
)

type TurboModal struct {
	visible bool
	respCh  chan bool
	theme   *theme.Theme
}

func NewTurboModal(t *theme.Theme) TurboModal {
	return TurboModal{
		theme: t,
	}
}

func (m *TurboModal) Show() {
	m.visible = true
}

func (m TurboModal) Visible() bool {
	return m.visible
}

func (m *TurboModal) SetResponseChannel(ch chan bool) {
	m.respCh = ch
}

func (m TurboModal) Update(msg tea.Msg) (TurboModal, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y":
			if m.respCh != nil {
				m.respCh <- true
			}
			m.visible = false
			return m, nil

		case "n", "esc":
			if m.respCh != nil {
				m.respCh <- false
			}
			m.visible = false
			return m, nil
		}
	}

	return m, nil
}

func (m TurboModal) View() string {
	if !m.visible {
		return ""
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		"",
		"TURBO bypasses ALL permission prompts.",
		"",
		"The agent can:",
		"• Write/delete any files",
		"• Execute any shell commands",
		"• Make network requests",
		"",
		"Only use in isolated/safe environments.",
		"",
		strings.Repeat("═", 60),
		"[y] Activate Turbo    [n] Cancel",
	)

	return m.theme.PanelStyle().
		Width(64).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")). // red
		Render("⚡ TURBO MODE WARNING\n\n" + content)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/components -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/components/
git commit -m "feat(tui): add turbo mode warning modal"
```

---

## Task 9: Update Status Bar for Three Modes

**Files:**
- Modify: `internal/tui/components/statusbar.go`

**Step 1: Read current status bar implementation**

Run: `cat internal/tui/components/statusbar.go | head -50`

**Step 2: Update status bar to use ExecutionMode**

Modify `internal/tui/components/statusbar.go`:

Change the `Mode` field from `string` to use the modes package:

```go
import (
	// ... existing imports ...
	"github.com/robinojw/dj/internal/modes"
)

type StatusBar struct {
	// ... existing fields ...
	Mode            modes.ExecutionMode  // CHANGE from string
	// ... rest of fields ...
}
```

Update the mode badge rendering in `View()`:

```go
func (s StatusBar) View() string {
	// ... existing code ...

	// Mode badge with color
	modeStyle := s.getModeStyle()
	modeBadge := modeStyle.Render(s.Mode.StatusLabel())

	// ... rest of rendering ...
}

func (s StatusBar) getModeStyle() lipgloss.Style {
	switch s.Mode {
	case modes.ModeTurbo:
		return s.theme.DangerStyle().Bold(true)  // red
	case modes.ModePlan:
		return s.theme.InfoStyle().Bold(true)    // blue
	case modes.ModeConfirm:
		return s.theme.WarningStyle()            // amber
	default:
		return lipgloss.NewStyle()
	}
}
```

**Step 3: Build to verify it compiles**

Run: `go build ./...`
Expected: SUCCESS

**Step 4: Commit**

```bash
git add internal/tui/components/statusbar.go
git commit -m "feat(tui): update status bar for three-mode system"
```

---

## Task 10: Migrate AgentMode to ExecutionMode in modes.go

**Files:**
- Modify: `internal/agents/modes.go`

**Step 1: Replace AgentMode with ExecutionMode reference**

Modify `internal/agents/modes.go`:

Remove the old `AgentMode` enum and import from `internal/modes`:

```go
package agents

import (
	"github.com/robinojw/dj/internal/modes"
)

// Re-export for convenience
type AgentMode = modes.ExecutionMode

const (
	ModePlan    = modes.ModePlan
	ModeConfirm = modes.ModeConfirm
	ModeTurbo   = modes.ModeTurbo
)

// Modes is now sourced from the modes package
var Modes = modes.Modes

// FilterTools remains but now uses modes.ModeConfig
func FilterTools(allTools []string, cfg modes.ModeConfig) []string {
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

**Step 2: Update modes_test.go**

Modify `internal/agents/modes_test.go` to use the re-exported types:

```go
package agents

import "testing"

func TestFilterTools(t *testing.T) {
	allTools := []string{"read_file", "write_file", "bash"}

	// Plan mode filters to read-only
	planCfg := Modes[ModePlan]
	filtered := FilterTools(allTools, planCfg)

	if len(filtered) != 1 || filtered[0] != "read_file" {
		t.Errorf("Expected only read_file, got %v", filtered)
	}

	// Confirm/Turbo allow all tools
	confirmCfg := Modes[ModeConfirm]
	filtered = FilterTools(allTools, confirmCfg)

	if len(filtered) != 3 {
		t.Errorf("Expected all 3 tools, got %d", len(filtered))
	}
}
```

**Step 3: Run tests**

Run: `go test ./internal/agents -v`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/agents/modes.go internal/agents/modes_test.go
git commit -m "refactor(agents): migrate AgentMode to use modes.ExecutionMode"
```

---

## Task 11: Wire Gate and Permission Channel in App

**Files:**
- Modify: `internal/tui/app.go`

**Step 1: Add gate and modals to App struct**

Modify `internal/tui/app.go`:

```go
import (
	// ... existing imports ...
	"github.com/robinojw/dj/internal/modes"
	"github.com/robinojw/dj/internal/tui/components"
)

type App struct {
	// ... existing fields ...
	mode            modes.ExecutionMode
	gate            *modes.Gate
	permissionModal components.PermissionModal
	turboModal      components.TurboModal
	turboConfirmed  bool
	permRequestCh   chan modes.PermissionRequest
	// ... rest of fields ...
}
```

**Step 2: Update NewApp to initialize gate**

```go
func NewApp(
	t *theme.Theme,
	client *api.ResponsesClient,
	tracker *api.Tracker,
	model string,
	cfg config.Config,  // ADD config parameter
) App {
	// Create gate from config
	gate := modes.NewGate(
		modes.ModeConfirm, // default mode
		cfg.Execution.Allow.Tools,
		cfg.Execution.Deny.Tools,
	)

	return App{
		screen:          ScreenChat,
		chat:            screens.NewChatModel(t),
		team:            screens.NewTeamModel(t),
		enhance:         screens.NewEnhanceModel(t),
		mcpManager:      screens.NewMCPManagerModel(t),
		skillBrowser:    screens.NewSkillBrowserModel(t),
		theme:           t,
		tracker:         tracker,
		client:          client,
		model:           model,
		checkpoints:     checkpoint.NewManager(20),
		mode:            modes.ModeConfirm,
		gate:            gate,
		permissionModal: components.NewPermissionModal(t),
		turboModal:      components.NewTurboModal(t),
		permRequestCh:   make(chan modes.PermissionRequest, 10),
	}
}
```

**Step 3: Update Tab key to cycle three modes**

Modify the `Update()` method in `app.go`:

```go
case "tab":
	// Cycle: Confirm → Plan → Turbo → Confirm
	newMode := (a.mode + 1) % 3

	// Check if switching to Turbo
	if newMode == modes.ModeTurbo && !a.turboConfirmed {
		a.turboModal.Show()
		respCh := make(chan bool, 1)
		a.turboModal.SetResponseChannel(respCh)

		// Wait for user response
		go func() {
			confirmed := <-respCh
			if confirmed {
				a.turboConfirmed = true
				a.mode = modes.ModeTurbo
				a.gate.SetMode(modes.ModeTurbo)
				a.chat.SetMode(modes.ModeTurbo)
			}
		}()
		return a, nil
	}

	a.mode = newMode
	a.gate.SetMode(newMode)
	a.chat.SetMode(newMode)
	return a, nil
```

**Step 4: Handle permission requests in Update**

Add to `Update()`:

```go
case modes.PermissionRequest:
	a.permissionModal, cmd = a.permissionModal.Update(msg)
	return a, cmd
```

**Step 5: Build to verify**

Run: `go build ./...`
Expected: SUCCESS

**Step 6: Commit**

```bash
git add internal/tui/app.go
git commit -m "feat(tui): wire gate and permission modals in app"
```

---

## Task 12: Add Permission Gate to Worker

**Files:**
- Modify: `internal/agents/worker.go`

**Step 1: Add gate and permission channel to Worker struct**

Modify `internal/agents/worker.go`:

```go
import (
	// ... existing imports ...
	"github.com/robinojw/dj/internal/modes"
	"time"
)

type Worker struct {
	// ... existing fields ...
	Mode        modes.ExecutionMode
	gate        *modes.Gate
	permReqCh   chan<- modes.PermissionRequest
	// ... rest of fields ...
}
```

**Step 2: Update NewWorker to accept gate and channel**

```go
func NewWorker(
	task Subtask,
	client *api.ResponsesClient,
	skillsRegistry *skills.Registry,
	model string,
	parentID string,
	mode modes.ExecutionMode,
	mem *memory.Manager,
	gate *modes.Gate,
	permReqCh chan<- modes.PermissionRequest,
) *Worker {
	return &Worker{
		ID:        task.ID,
		Task:      task,
		Status:    "pending",
		Mode:      mode,
		client:    client,
		skills:    skillsRegistry,
		memory:    mem,
		model:     model,
		parentID:  parentID,
		gate:      gate,
		permReqCh: permReqCh,
	}
}
```

**Step 3: Add tool execution gate method**

Add to `worker.go`:

```go
// executeTool runs a tool call through the permission gate.
func (w *Worker) executeTool(toolName string, args map[string]any) error {
	decision := w.gate.Evaluate(toolName, args)

	switch decision {
	case modes.GateDeny:
		return fmt.Errorf("tool %q blocked by deny list or mode", toolName)

	case modes.GateAllow:
		// Execute immediately (actual tool execution to be implemented)
		return nil

	case modes.GateAskUser:
		// Suspend and wait for user decision
		respCh := make(chan modes.PermissionResp, 1)
		req := modes.PermissionRequest{
			ID:       fmt.Sprintf("%s-%s", w.ID, toolName),
			WorkerID: w.ID,
			Tool:     toolName,
			Args:     args,
			RespCh:   respCh,
		}

		w.permReqCh <- req

		// Block with timeout
		select {
		case resp := <-respCh:
			if !resp.Allowed {
				return fmt.Errorf("user denied tool: %s", toolName)
			}

			// Handle remember scope
			if resp.RememberFor == modes.RememberSession {
				w.gate.AllowForSession(toolName)
			}
			if resp.RememberFor == modes.RememberAlways {
				// TODO: persist to config
			}

			return nil

		case <-time.After(5 * time.Minute):
			return fmt.Errorf("permission request timed out")
		}

	default:
		return fmt.Errorf("unknown gate decision: %v", decision)
	}
}
```

**Step 4: Update buildRequest to filter tools (Layer 1)**

Add method to `worker.go`:

```go
func (w *Worker) filterToolsForMode(allTools []string) []string {
	modeCfg := Modes[w.Mode]
	return FilterTools(allTools, modeCfg)
}
```

**Step 5: Build to verify**

Run: `go build ./...`
Expected: SUCCESS

**Step 6: Commit**

```bash
git add internal/agents/worker.go
git commit -m "feat(agents): add permission gate to worker with two-layer filtering"
```

---

## Task 13: Update Orchestrator to Pass Gate

**Files:**
- Modify: `internal/agents/orchestrator.go`

**Step 1: Add gate to Orchestrator**

Modify `internal/agents/orchestrator.go`:

```go
type Orchestrator struct {
	// ... existing fields ...
	Mode    modes.ExecutionMode
	Memory  *memory.Manager
	Gate    *modes.Gate
	PermReqCh chan modes.PermissionRequest
}
```

**Step 2: Update Dispatch to pass gate to workers**

```go
func (o *Orchestrator) Dispatch(ctx context.Context, task Task) (<-chan WorkerUpdate, error) {
	// ... existing code ...

	for _, subtask := range subtasks {
		worker := NewWorker(
			subtask,
			o.Client,
			o.Skills,
			o.Model,
			task.ID,
			o.Mode,
			o.Memory,
			o.Gate,
			o.PermReqCh,
		)
		// ... rest of dispatch ...
	}
	// ... rest of method ...
}
```

**Step 3: Build to verify**

Run: `go build ./...`
Expected: SUCCESS

**Step 4: Commit**

```bash
git add internal/agents/orchestrator.go
git commit -m "feat(agents): pass gate and permission channel to workers"
```

---

## Task 14: Update Main to Wire Everything

**Files:**
- Modify: `cmd/harness/main.go`

**Step 1: Update main to create gate and wire components**

Modify `cmd/harness/main.go`:

```go
func main() {
	// ... existing config loading ...

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Parse default mode
	var defaultMode modes.ExecutionMode
	switch cfg.Execution.DefaultMode {
	case "plan":
		defaultMode = modes.ModePlan
	case "turbo":
		defaultMode = modes.ModeTurbo
	default:
		defaultMode = modes.ModeConfirm
	}

	// Create gate
	gate := modes.NewGate(
		defaultMode,
		cfg.Execution.Allow.Tools,
		cfg.Execution.Deny.Tools,
	)

	// ... existing client, tracker setup ...

	// Create app with gate
	app := tui.NewApp(theme, client, tracker, cfg.Model.Default, cfg)

	// ... existing tea.NewProgram and Run ...
}
```

**Step 2: Build to verify**

Run: `go build ./cmd/harness`
Expected: SUCCESS

**Step 3: Commit**

```bash
git add cmd/harness/main.go
git commit -m "feat(main): wire gate and three-mode system in main"
```

---

## Task 15: Update Chat Screen to Use ExecutionMode

**Files:**
- Modify: `internal/tui/screens/chat.go`

**Step 1: Update ChatModel to use modes.ExecutionMode**

Modify `internal/tui/screens/chat.go`:

```go
import (
	// ... existing imports ...
	"github.com/robinojw/dj/internal/modes"
)

type ChatModel struct {
	// ... existing fields ...
	Mode modes.ExecutionMode
	// ... rest of fields ...
}

func (m *ChatModel) SetMode(mode modes.ExecutionMode) {
	m.Mode = mode
	m.statusBar.Mode = mode
}
```

**Step 2: Build to verify**

Run: `go build ./...`
Expected: SUCCESS

**Step 3: Commit**

```bash
git add internal/tui/screens/chat.go
git commit -m "refactor(tui): update chat screen to use ExecutionMode"
```

---

## Task 16: Add Config Persist for "Always" Scope

**Files:**
- Create: `internal/modes/persist.go`
- Create: `internal/modes/persist_test.go`

**Step 1: Write test for config persistence**

Create `internal/modes/persist_test.go`:

```go
package modes

import (
	"os"
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

	if !strings.Contains(string(data), `"write_file"`) {
		t.Error("write_file not found in persisted config")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/modes -v`
Expected: FAIL with "undefined: PersistToolToAllowList"

**Step 3: Implement config persistence**

Create `internal/modes/persist.go`:

```go
package modes

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/robinojw/dj/config"
)

// PersistToolToAllowList adds a tool to the allow list in harness.toml.
func PersistToolToAllowList(configPath string, toolName string) error {
	// Read current config
	var cfg config.Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	// Check if already in list
	for _, t := range cfg.Execution.Allow.Tools {
		if t == toolName {
			return nil // already present
		}
	}

	// Add tool
	cfg.Execution.Allow.Tools = append(cfg.Execution.Allow.Tools, toolName)

	// Write back
	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("open config for write: %w", err)
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/modes -v`
Expected: PASS

**Step 5: Wire persistence in worker.go**

Modify `internal/agents/worker.go` in the `executeTool` method:

```go
if resp.RememberFor == modes.RememberAlways {
	if err := modes.PersistToolToAllowList("harness.toml", toolName); err != nil {
		// Log error but don't fail the tool call
		fmt.Fprintf(os.Stderr, "Warning: failed to persist to config: %v\n", err)
	}
}
```

**Step 6: Commit**

```bash
git add internal/modes/persist.go internal/modes/persist_test.go internal/agents/worker.go
git commit -m "feat(modes): add config persistence for 'Always' scope"
```

---

## Task 17: Add Integration Test for Full Flow

**Files:**
- Create: `internal/modes/integration_test.go`

**Step 1: Write integration test**

Create `internal/modes/integration_test.go`:

```go
package modes

import (
	"testing"
	"time"
)

func TestIntegration_PermissionFlow(t *testing.T) {
	gate := NewGate(ModeConfirm, []string{}, []string{})

	// Simulate worker requesting permission
	respCh := make(chan PermissionResp, 1)
	req := PermissionRequest{
		ID:       "int-test-1",
		WorkerID: "worker-1",
		Tool:     "write_file",
		Args:     map[string]any{"path": "test.go"},
		RespCh:   respCh,
	}

	// Simulate worker evaluating
	decision := gate.Evaluate(req.Tool, req.Args)
	if decision != GateAskUser {
		t.Fatalf("Expected GateAskUser, got %v", decision)
	}

	// Simulate TUI approval
	go func() {
		time.Sleep(10 * time.Millisecond)
		respCh <- PermissionResp{
			Allowed:     true,
			RememberFor: RememberSession,
		}
	}()

	// Worker blocks
	resp := <-respCh

	if !resp.Allowed {
		t.Error("Expected approval")
	}
	if resp.RememberFor != RememberSession {
		t.Errorf("Expected RememberSession, got %v", resp.RememberFor)
	}

	// Add to session allow list
	gate.AllowForSession(req.Tool)

	// Second call should auto-allow
	decision = gate.Evaluate(req.Tool, req.Args)
	if decision != GateAllow {
		t.Errorf("Expected GateAllow after session allow, got %v", decision)
	}
}

func TestIntegration_ModeCycle(t *testing.T) {
	tests := []struct {
		mode ExecutionMode
		tool string
		want GateDecision
	}{
		{ModeConfirm, "read_file", GateAllow},
		{ModeConfirm, "write_file", GateAskUser},
		{ModePlan, "read_file", GateAllow},
		{ModePlan, "write_file", GateDeny},
		{ModeTurbo, "read_file", GateAllow},
		{ModeTurbo, "write_file", GateAllow},
	}

	for _, tt := range tests {
		t.Run(tt.mode.String()+"_"+tt.tool, func(t *testing.T) {
			gate := NewGate(tt.mode, []string{}, []string{})
			decision := gate.Evaluate(tt.tool, nil)
			if decision != tt.want {
				t.Errorf("Mode %s, tool %s: got %v, want %v",
					tt.mode, tt.tool, decision, tt.want)
			}
		})
	}
}
```

**Step 2: Run integration test**

Run: `go test ./internal/modes -v -run Integration`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/modes/integration_test.go
git commit -m "test(modes): add integration tests for permission flow"
```

---

## Task 18: Update harness.toml with Execution Config

**Files:**
- Modify: `harness.toml`

**Step 1: Add execution section to harness.toml**

Modify `harness.toml`:

```toml
# ... existing config ...

[execution]
default_mode = "confirm"  # confirm | plan | turbo

[execution.allow]
# Tools always auto-approved (overrides mode restrictions)
tools = [
    "read_file",
    "list_dir",
    "search_code",
    "run_tests",
    "bash(git status*)",
    "bash(git diff*)",
    "bash(go build*)",
]

[execution.deny]
# Blocked in all modes (security floor)
tools = [
    "bash(rm -rf*)",
    "bash(curl*)",
    "read_file(.env*)",
    "write_file(.env*)",
]
```

**Step 2: Commit**

```bash
git add harness.toml
git commit -m "config: add execution section with default allow/deny lists"
```

---

## Task 19: Run All Tests

**Step 1: Run full test suite**

Run: `go test ./... -v`
Expected: ALL PASS

**Step 2: Check for missed imports or build errors**

Run: `go build ./...`
Expected: SUCCESS

**Step 3: If any tests fail, fix them**

Fix any compilation errors or test failures.

**Step 4: Commit any fixes**

```bash
git add .
git commit -m "fix: resolve test failures and build errors"
```

---

## Task 20: Manual Testing

**Step 1: Run the application**

Run: `go run ./cmd/harness`

**Step 2: Test Tab cycling**

- Press Tab multiple times
- Verify status bar shows: ⏸ CONFIRM → ◎ PLAN → ⚡ TURBO → ⏸ CONFIRM

**Step 3: Test Turbo warning on first activation**

- Tab to Turbo mode for first time
- Verify warning modal appears
- Press 'y' to confirm
- Verify mode switches to Turbo

**Step 4: Test permission modal in Confirm mode**

- Switch to Confirm mode
- Try a write operation
- Verify permission modal appears
- Test 'y' to allow
- Test 'n' to deny
- Test Tab to cycle scopes

**Step 5: Test Plan mode blocks writes**

- Switch to Plan mode
- Try a write operation
- Verify no modal appears and operation is blocked

**Step 6: Test session allow list**

- In Confirm mode, approve tool with "Session" scope
- Try same tool again
- Verify it auto-allows without prompt

**Step 7: Document any issues**

Create a checklist in the commit message for remaining manual tests.

**Step 8: Commit manual test results**

```bash
git commit --allow-empty -m "test: manual testing of three-mode permission system

Manual test checklist:
- [x] Tab cycles through Confirm/Plan/Turbo
- [x] Status bar shows correct badges
- [x] Turbo warning appears on first activation
- [x] Permission modal shows in Confirm mode
- [x] Plan mode blocks writes without prompt
- [x] Session scope persists within session
- [ ] Always scope persists to config (requires restart test)
- [ ] Deny list blocks in all modes
- [ ] Allow list overrides in all modes
"
```

---

## Task 21: Update Documentation

**Files:**
- Create: `docs/user-guide-permissions.md`

**Step 1: Write user guide for permission modes**

Create `docs/user-guide-permissions.md`:

```markdown
# Permission Modes User Guide

DJ provides three execution modes that control agent autonomy and tool access.

## Modes

### Confirm Mode (Default)

The agent asks permission before executing write, execute, or MCP mutation tools. Read operations proceed automatically.

**Status badge:** `⏸ CONFIRM` (amber)

**Use when:** You want oversight of destructive operations

**Key binding:** Tab to cycle modes

### Plan Mode

The agent operates in read-only mode with high reasoning effort. It can only read files, search code, and list directories.

**Status badge:** `◎ PLAN` (blue)

**Use when:** You want architectural planning without execution risk

### Turbo Mode

The agent bypasses all permission checks. All tools execute immediately without prompts.

**Status badge:** `⚡ TURBO` (red)

**Use when:** Working in isolated/safe environments where speed matters

**Warning:** Requires confirmation on first activation per session

## Permission Modal

When in Confirm mode, the agent will show a permission modal before executing risky tools:

```
╔══════════════════════════════════════════════════════════════╗
║  ⚠  Permission Required                                      ║
╠══════════════════════════════════════════════════════════════╣
║                                                              ║
║  Worker A wants to run:                                      ║
║                                                              ║
║  🔧 bash                                                     ║
║  ─────────────────────────────────────────────────────────  ║
║  $ npm run build && npm test                                 ║
║                                                              ║
╠══════════════════════════════════════════════════════════════╣
║  Remember this decision?                                     ║
║  ○ Just this once   ● This session   ○ Always               ║
╠══════════════════════════════════════════════════════════════╣
║  [y] Allow    [n] Deny    [Tab] cycle scope    [Esc] Deny   ║
╚══════════════════════════════════════════════════════════════╝
```

**Controls:**
- `y`: Approve with current scope
- `n` or `Esc`: Deny
- `Tab`: Cycle remember scope

**Remember scopes:**
- **Once**: Allow this single call
- **Session**: Allow this tool for the current session
- **Always**: Persist to `harness.toml` for all future sessions

## Configuration

Edit `harness.toml` to customize permission behavior:

```toml
[execution]
default_mode = "confirm"  # confirm | plan | turbo

[execution.allow]
# Auto-approved in all modes
tools = [
    "read_file",
    "run_tests",
    "bash(git status*)",   # glob patterns supported
]

[execution.deny]
# Blocked in all modes (security floor)
tools = [
    "bash(rm -rf*)",
    "read_file(.env*)",
]
```

**Glob patterns:** Use `*` wildcards for flexible matching:
- `bash(git status*)` matches `bash(git status)` and `bash(git status --short)`
- `read_file(.env*)` matches `.env`, `.env.local`, `.env.production`

## Security

**Deny list wins:** Tools in the deny list are blocked even in Turbo mode

**Unknown tools:** Default to write classification (require permission)

**Defense-in-depth:** Allow/deny lists apply at both filter and runtime layers
```

**Step 2: Commit**

```bash
git add docs/user-guide-permissions.md
git commit -m "docs: add user guide for three-mode permission system"
```

---

## Summary

This plan implements the three-mode permission system through 21 granular tasks:

1. **Foundation** (Tasks 1-5): Core types, tool classification, glob matching, gate logic, permission types
2. **Configuration** (Task 6): Execution config in TOML
3. **UI Components** (Tasks 7-8): Permission modal and Turbo warning
4. **Status Bar** (Task 9): Three-mode badges
5. **Migration** (Tasks 10-15): Migrate from AgentMode to ExecutionMode across codebase
6. **Persistence** (Task 16): "Always" scope saves to config
7. **Testing** (Tasks 17-20): Integration tests and manual verification
8. **Documentation** (Task 21): User guide

Each task follows TDD: write test → run (fail) → implement → run (pass) → commit.

**Total estimated time:** 3-4 hours for complete implementation

**Verification:** All tests pass, manual testing confirms behavior, documentation complete
