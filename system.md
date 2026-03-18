# DJ System Snapshot — 2026-03-18

## What DJ Is

DJ is a Go TUI multiplexer for OpenAI Codex CLI agent sessions. Think tmux for AI agents — a canvas grid of agent cards with the ability to open any card into a full interactive codex session. It runs on Bubble Tea with an event-driven architecture.

**Branch:** `robin/phase-9` (all changes uncommitted)
**Status:** 135 tests passing, 0 failures, builds clean with `-race`

## Architecture: Two-Process Hybrid

DJ manages **two kinds of codex processes** per thread:

1. **`codex proto`** (background, JSON-RPC) — spawned once at startup. Delivers structured events (session configured, task started, agent deltas, token counts) used for canvas card metadata. Managed by `internal/appserver/Client`.

2. **`codex` interactive** (PTY, per-session) — spawned lazily when user opens a session card. The real codex CLI TUI rendered through a VT terminal emulator. Managed by `internal/tui/PTYSession`.

```
┌──────────────────────────────────────────────┐
│  DJ TUI (Bubble Tea)                         │
│                                              │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐       │
│  │ Card    │ │ Card    │ │ Card    │  Canvas │
│  │ (idle)  │ │ (active)│ │ (done)  │        │
│  └─────────┘ └─────────┘ └─────────┘       │
│                                              │
│  ┌──────────────────────────────────┐       │
│  │ PTY Session (real codex CLI)     │  Open  │
│  │ ← vt.SafeEmulator.Render()      │  Card  │
│  └──────────────────────────────────┘       │
│                                              │
│  Status: ● Connected | 3 threads | model    │
└──────────────────────────────────────────────┘
         │                        │
    JSON-RPC stdio            PTY stdin/stdout
         │                        │
   codex proto              codex (interactive)
   (1 process)              (N processes, lazy)
```

## Package Structure

```
dj/
├── cmd/dj/main.go                 (59 lines)   Entry point, Cobra CLI
├── internal/
│   ├── appserver/                               JSON-RPC 2.0 IPC layer
│   │   ├── client.go              (177 lines)   Process spawn, bidirectional JSONL
│   │   ├── protocol.go            (22 lines)    ProtoEvent, ProtoSubmission types
│   │   ├── methods.go             (23 lines)    Event/op string constants
│   │   └── types_thread.go        (44 lines)    SessionConfigured, TaskComplete, etc.
│   │
│   ├── state/                                   Reactive state store
│   │   ├── store.go               (130 lines)   ThreadStore (RWMutex, insertion order)
│   │   └── thread.go              (50 lines)    ThreadState, ChatMessage, status enums
│   │
│   ├── config/
│   │   └── config.go              (63 lines)    Viper TOML: appserver + interactive + UI
│   │
│   └── tui/                                     Bubble Tea UI layer
│       ├── app.go                 (240 lines)   AppModel: orchestrator, Update/View
│       ├── app_view.go            (35 lines)    View() layout rendering
│       ├── app_proto.go           (118 lines)   JSON-RPC event handlers
│       ├── app_pty.go             (115 lines)   PTY session lifecycle
│       ├── app_menu.go            (72 lines)    Ctrl+B menu dispatch
│       ├── msgs.go                (54 lines)    All TUI message types
│       ├── bridge.go              (92 lines)    ProtoEvent → tea.Msg conversion
│       ├── session.go             (67 lines)    SessionModel: thin PTY wrapper
│       ├── pty_session.go         (176 lines)   PTYSession: creack/pty + vt emulator
│       ├── pty_keys.go            (107 lines)   KeyMsg → ANSI byte conversion
│       ├── canvas.go              (85 lines)    3-column card grid
│       ├── tree.go                (110 lines)   Hierarchical thread tree
│       ├── card.go                (77 lines)    Individual thread card (30×6)
│       ├── statusbar.go           (91 lines)    Connection state + thread count
│       ├── prefix.go              (51 lines)    Ctrl+B tmux-style prefix handler
│       ├── menu.go                (82 lines)    Context menu (fork/delete/rename)
│       ├── help.go                (64 lines)    Keybinding overlay
│       └── actions.go             (15 lines)    Fork/Delete/Rename msg types
```

**27 source files, 24 test files, 51 total Go files.**
**All non-test files under 300 lines (CI enforced).**

## Data Flow

### Protocol Events (background metadata)
```
codex proto stdout → Client.ReadLoop() goroutine
  → app.events channel
  → listenForEvents() tea.Cmd
  → protoEventMsg → Update()
  → ProtoEventToMsg() → SessionConfiguredMsg | TaskStartedMsg | AgentDeltaMsg | ...
  → ThreadStore updated (status, messages, titles)
  → Canvas cards re-render
```

### PTY Session (interactive display)
```
codex interactive stdout → PTYSession.readLoop() goroutine
  → vt.SafeEmulator.Write(bytes)
  → PTYOutputMsg → app.ptyEvents channel
  → listenForPTYEvents() tea.Cmd
  → View() calls emulator.Render() → ANSI string → terminal

User keypress → tea.KeyMsg → handleSessionKey()
  → KeyMsgToBytes(msg) → ANSI escape bytes
  → PTYSession.WriteBytes() → PTY stdin → codex receives input
```

### Session Lifecycle
```
1. Enter on canvas card  → openSession()
2. First time           → NewPTYSession() + Start() → stored in ptySessions[threadID]
3. Esc                  → closeSession() — PTY stays alive in background
4. Re-enter same card   → reconnect to existing PTYSession (no new process)
5. Process exits        → PTYOutputMsg{Exited: true} → session.MarkExited()
6. App quit             → StopAllPTYSessions() kills all
```

## Key Routing

```
Any focus:
  Ctrl+B         → prefix mode (next key = action)
  Ctrl+B m       → thread menu (fork/delete/rename)

Canvas/Tree focus:
  ←/→/↑/↓        → navigate cards or tree
  Enter           → open session (spawn or reconnect PTY)
  t               → toggle canvas ↔ tree
  n               → create new thread
  ?               → help overlay
  Esc / Ctrl+C    → quit DJ

Session focus:
  Esc             → close session view, return to canvas
  Ctrl+C          → quit DJ
  Everything else → forwarded to PTY stdin (codex handles it)
```

## Concurrency Model

| Goroutine | Purpose | Sync Mechanism |
|-----------|---------|----------------|
| `Client.ReadLoop()` | Read JSON-RPC from codex proto stdout | `app.events` channel (buffered 64) |
| `Client.drainStderr()` | Consume codex proto stderr | Fire-and-forget |
| `PTYSession.readLoop()` | Read PTY output, feed VT emulator | `vt.SafeEmulator` (internal RWMutex) + `app.ptyEvents` channel |
| `listenForEvents()` | Bridge protocol channel → Bubble Tea | Blocking channel read in tea.Cmd |
| `listenForPTYEvents()` | Bridge PTY channel → Bubble Tea | Blocking channel read in tea.Cmd |

**Shared state protection:**
- `ThreadStore`: `sync.RWMutex`
- `Client.stdin`: `sync.Mutex`
- `Client.running`: `atomic.Bool`
- `PTYSession.{running,exitCode}`: `sync.Mutex`
- `vt.SafeEmulator`: Thread-safe wrapper (internal RWMutex for concurrent Write/Render)

## Configuration

`dj.toml` (TOML, optional):

```toml
[appserver]
command = "codex"       # Background JSON-RPC process
args = ["proto"]

[interactive]
command = "codex"       # Interactive PTY process
args = []

[ui]
theme = "default"
```

All values have sensible defaults. Config file is optional.

## Dependencies

| Dependency | Version | Purpose |
|-----------|---------|---------|
| `charmbracelet/bubbletea` | v1.3.10 | TUI framework |
| `charmbracelet/lipgloss` | v1.1.0 | Terminal styling |
| `charmbracelet/bubbles` | v1.0.0 | Viewport, textinput (used by other components) |
| `charmbracelet/x/vt` | latest | VT100 terminal emulator |
| `creack/pty` | v1.1.24 | PTY spawning (Unix) |
| `spf13/cobra` | v1.10.2 | CLI framework |
| `spf13/viper` | v1.21.0 | TOML config |

## Test Coverage

**135 tests total**, all passing with `-race`:

| Package | Tests | Key Coverage |
|---------|-------|-------------|
| `appserver` | 17 | Client start/stop, send/receive, protocol types, methods |
| `config` | 3 | Default loading, file loading, missing file |
| `state` | 13 | Store CRUD, thread hierarchy, deltas, status |
| `tui` | 102 | App routing, canvas/tree navigation, session lifecycle, PTY start/stop/resize/write, key encoding, bridge decoding, menu/help/prefix, card rendering |

**Integration test** (behind `//go:build integration` tag): connects to real `codex proto`.

## What Works

- Canvas grid of thread cards with arrow key navigation
- Thread tree view (toggle with `t`)
- Status bar showing connection state and thread count
- Ctrl+B prefix key system (tmux-style)
- Context menu (fork/delete/rename thread — message dispatch only)
- Help overlay with all keybindings
- JSON-RPC protocol bridge: session configured, task lifecycle, agent deltas, exec/patch auto-approval
- **PTY-embedded sessions**: real codex CLI rendered via VT emulator
- **Key forwarding**: all keys in session focus routed to PTY as ANSI bytes
- **Session persistence**: PTY stays alive when Esc closes the view, reconnects on re-enter
- **Graceful shutdown**: all PTY sessions killed on app exit
- Interactive command configurable via `dj.toml` or `WithInteractiveCommand` option

## What's Incomplete / Needs Work

### Not Yet Implemented
1. **Thread creation via JSON-RPC** — `createThread()` returns a local stub when client is present (the `return nil` path). No actual `client.CreateThread()` method exists yet.
2. **Fork/Delete/Rename dispatch** — `dispatchMenuAction()` creates `ForkThreadMsg`, `DeleteThreadMsg`, `RenameThreadMsg` but `Update()` has no handlers for them. The messages are dropped.
3. **Thread-to-interactive-session mapping** — `openSession()` spawns a bare `codex` process with no arguments linking it to a specific thread. The interactive process doesn't know which thread ID it belongs to in the proto channel.
4. **Multiple proto sessions** — The app tracks a single `sessionID` and `currentMessageID`. Canvas metadata only updates for the one active proto session. Multi-thread proto state routing doesn't exist.
5. **Token count display** — `EventTokenCount` constant exists but no handler or display.
6. **Agent reasoning** — `EventAgentReasoning`, `EventAgentReasonDelta`, `EventAgentReasonBreak` constants exist but no bridge decoder or handler.
7. **Scrollback in PTY sessions** — The VT emulator supports scrollback (`DefaultScrollbackSize = 10000`) but there's no UI to scroll back through it (all keys go to PTY).

### Structural Issues
8. **`app.go` struct is getting heavy** — 18 fields on `AppModel`. The value-receiver-with-mutations pattern (Bubble Tea idiom) means the struct is copied on every `Update()` call, including the `ptySessions` map and channels (shared via pointer/reference semantics, but still copied structurally).
9. **No error recovery for PTY spawn failure** — If `codex` isn't installed, `openSession()` sets a status bar error but the user can't retry or see what went wrong beyond the one-line error.
10. **Proto handlers still update thread store** — `handleTaskStarted()`, `handleAgentDelta()`, etc. still append `ChatMessage` to `ThreadState.Messages` and `CommandOutput`. This data is now unused (sessions render via VT emulator), but accumulates memory.

### Not Yet Wired
11. **`ForkThreadMsg`/`DeleteThreadMsg`/`RenameThreadMsg`** — Actions exist, menu dispatches them, but no `case` in `Update()`.
12. **`ThreadDeletedMsg`** — Type exists in msgs.go but no handler in `Update()`.
13. **Interactive args for thread context** — The `InteractiveConfig.Args` from config get passed to every PTY session, but there's no way to pass thread-specific context (like a thread ID or conversation file) to the interactive codex process.

## Uncommitted Changes

All work on branch `robin/phase-9` is **uncommitted**. The diff covers 33 files (+969, -1477 lines) spanning both the earlier appserver simplification and the PTY session implementation. Key new files not yet tracked by git:

- `internal/tui/app_proto.go` (extracted from app.go in earlier work)
- `internal/tui/app_view.go` (extracted from app.go in earlier work)
- `internal/tui/app_pty.go` (new — PTY session management)
- `internal/tui/pty_keys.go` (new — key mapping)
- `internal/tui/pty_keys_test.go` (new)
- `internal/tui/pty_session.go` (new — PTY + VT wrapper)
- `internal/tui/pty_session_test.go` (new)
- `cmd/test_protocol/` (new — protocol testing tool)
