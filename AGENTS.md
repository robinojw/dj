# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

DJ is a Go TUI application that orchestrates AI agents in the terminal. It uses the OpenAI Responses API with SSE streaming, spawns parallel worker agents for complex tasks, and provides a permission-gated tool execution system. Built on Charmbracelet's Bubble Tea framework.

## Commands

```bash
# Build
go build -o dj ./cmd/harness

# Run
./dj

# Test
go test ./...                        # all tests
go test ./internal/agents -v         # single package, verbose
go test ./internal/tools -run TestEditFile -v  # single test

# Lint (runs in CI via golangci-lint action)
golangci-lint run

# CI runs tests with race detector
go test ./... -v -race
```

## Architecture

Entry point: `cmd/harness/main.go`

### Core packages under `internal/`:

- **agents/**: Multi-agent orchestration. `Orchestrator` decomposes tasks into subtasks, `DAG` schedules them via topological sort (Kahn's algorithm), `Worker` goroutines execute subtasks with independent context and multi-turn tool call loops (up to 25 turns). `TaskRouter` analyzes complexity and routes appropriately.

- **api/**: OpenAI Responses API client. `ResponsesClient` streams SSE from `/v1/responses`. `Tracker` records token usage and costs. Request/response types in `models.go`. Tool schemas and instructions are sent with chat requests.

- **tools/**: Native Go tool implementations with a `ToolRegistry`. Each tool has `ToolAnnotations` (ReadOnly, Destructive, Idempotent, MutatesFiles) used by the permission system. `edit_file` uses 3-tier whitespace-tolerant string matching (exact → trimmed → normalized).

- **modes/**: Permission system with three execution modes (Confirm/Plan/Turbo). `Gate` evaluates tool calls against deny list → allow list → mode rules. Glob patterns supported for tool matching (e.g., `bash(git status*)`).

- **tui/**: Bubble Tea TUI with screen-based navigation. Main screens: Chat, Team (multi-agent view), Skill Browser, MCP Manager, Enhance, Cheat Sheet. Components: chat input, status bar, permission modal, debug overlay.

- **mcp/**: Model Context Protocol client supporting stdio and HTTP transports. Auto-discovers servers from `~/.config/claude/mcp.json` and `~/.config/codex-harness/mcp.json`. JSON-RPC 2.0.

- **skills/**: Loads SKILL.md files (YAML frontmatter + instructions) from configured paths. Supports trigger keywords and implicit invocation.

- **lsp/**: Auto-detects language servers by marker files (go.mod → gopls, tsconfig.json → typescript-language-server, pyproject.toml → pylsp).

- **checkpoint/**: Ring buffer (20 entries) for undo state snapshots (Ctrl+Z).

- **config/**: Two-level TOML config loading — project `harness.toml` + user `~/.config/codex-harness/config.toml`.

### Key patterns

- Goroutine-based concurrency with channels for worker communication and `sync.Mutex` for shared state
- All APIs accept `context.Context` for cancellation
- Interface-based design: `ToolClassifier`, `ToolHandler` for pluggable tools
- Annotation-driven tool metadata feeds into permission gate decisions
- Screen stacking in TUI for modal navigation

## Configuration

Project config: `harness.toml` (TOML). Sections: `[model]`, `[theme]`, `[execution]`, `[execution.allow]`, `[execution.deny]`, `[mcp.servers]`, `[skills]`, `[hooks]`. User overrides in `~/.config/codex-harness/config.toml`.

## Dependencies

Core: Bubble Tea (TUI), Lipgloss (styling), Bubbles (UI components), BurntSushi/toml, go-humanize, yaml.v3. LSP via go.lsp.dev packages. No external test frameworks — standard `testing` package only.

## Code Style

### 1. No shortened variable names
Use descriptive names. `err` not `e`, `registry` not `r`, `skill` not `s`. Single-letter loop counters (`i`, `j`, `k`) and standard Go receivers are acceptable.

### 2. No nested if statements
Use early returns and guard clauses. Extract nested logic into helper functions.

### 3. No complex boolean conditions inline
Assign compound conditions to a descriptively named variable before using in an if statement.

```go
// Bad
if w.Status != "completed" && w.Status != "error" && w.Status != "skipped" {

// Good
isStillRunning := w.Status != "completed" && w.Status != "error" && w.Status != "skipped"
if isStillRunning {
```

### 4. No inline comments
Write readable code instead of commenting it. Only use godoc comments on exported types and complex functions where absolutely necessary.

### 5. Intentional directory and file naming
Structure directories and filenames so they are easy to search and grep through.

### 6. No repeated raw strings
Assign repeated string literals to named constants so they can be reused and found in one place.

### 7. Named constants for magic numbers
Extract numeric literals (buffer sizes, timeouts, weights, permissions) into named constants with descriptive names.

### 8. Use fmt.Errorf for errors
Always use `fmt.Errorf("context: %w", err)` for error wrapping. Never use `fmt.Sprintf("Error: %v", err)`.
