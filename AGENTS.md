# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

DJ is a Go TUI that visualizes and controls OpenAI Codex CLI agent sessions via the Codex App Server's JSON-RPC 2.0 stdio protocol. It uses an event-driven Bubble Tea architecture where a reactive `ThreadStore` is the single source of truth, populated by protocol events and driving canvas card rendering.

## Commands

```bash
# Build
go build -o dj ./cmd/dj

# Run
./dj

# Test
go test ./...                                  # all tests
go test ./internal/appserver -v                # single package, verbose
go test ./internal/appserver -run TestClientCall -v  # single test

# Integration test (requires codex CLI installed)
go test ./internal/appserver -v -tags=integration

# Lint (runs in CI via golangci-lint action)
golangci-lint run

# CI runs tests with race detector
go test ./... -v -race
```

## Architecture

Entry point: `cmd/dj/main.go`

### Core packages under `internal/`:

- **appserver/**: IPC layer for the Codex App Server. Spawns `codex app-server --listen stdio://` as a child process. `Client` manages bidirectional JSON-RPC 2.0 over stdio (JSONL). `ReadLoop` reads newline-delimited JSON from stdout. `Dispatch` routes messages: pending-request responses, server-to-client requests, and notifications. `Call` provides synchronous request/response with a `sync.Map`-based pending tracker. `Initialize` performs the required handshake.

- **state/** (planned): Reactive state store. `ThreadStore` maps thread IDs to `ThreadState`. Single source of truth — no direct app-server queries for UI.

- **tui/** (planned): Bubble Tea TUI layer. Canvas grid of agent cards, session panes with terminal output, agent tree navigation, overlay screens. tmux-style `Ctrl+B` prefix key for pane operations.

- **config/** (planned): Viper-based config loader with TOML format.

### Key patterns

- Single app-server process per TUI instance; multiple threads managed via JSON-RPC
- External goroutine (JSON-RPC reader) injects events via `program.Send(msg)` (Bubble Tea message passing)
- PTY I/O routes through `command/exec` with `tty: true` over the existing JSON-RPC connection
- All APIs accept `context.Context` for cancellation
- Goroutine-based concurrency with channels and `sync.Mutex` for shared state

## Configuration

Project config: `dj.toml` (TOML, planned). Viper-based with cobra CLI integration.

## Dependencies

Core: Bubble Tea (TUI), Lipgloss (styling), Bubbles (components), cobra+viper (CLI+config), JSON-RPC 2.0 over stdio. No external test frameworks — standard `testing` package only.

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

## CI Enforcement

- **golangci-lint**: govet, staticcheck, funlen (60 lines max), cyclop (complexity 15 max)
- **File length**: 300 lines max for non-test, non-generated `.go` files
- **Race detector**: `go test -race` on all packages
