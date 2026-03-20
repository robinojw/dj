# Implementation Plan: Claude Code Agent Teams Support

**Branch:** `nick/claude-code-agent-teams`
**Author:** Nick Hess
**Date:** 2026-03-20
**Status:** Proposal

---

## Goal

Extend DJ to support Claude Code agent sessions alongside the existing Codex integration, with first-class support for multi-agent team workflows — where a lead agent orchestrates sub-agents with distinct roles (researcher, coder, reviewer), shared context, and coordinated approvals.

## Relationship to Existing Development

The existing roadmap (documented in `system.md`) focuses on completing the Codex v2 protocol integration. This plan is designed to **run in parallel** without blocking that work, with one structural dependency noted below.

### Can Be Done Independently (No Conflicts)

- **Phases 1-3** (Provider interface, Claude Code adapter, config extensions) touch new files and packages only. They don't modify any existing Codex-specific code — they wrap it behind a new abstraction.
- **Phase 5** (approval queue) is additive UI — a new overlay that consumes existing message types.
- **Phase 6** (token/cost dashboard) builds on the already-stubbed `MethodTokenUsageUpdated` event.

### Requires Coordination (Soft Dependencies)

- **Phase 4** (inter-agent messaging and team orchestration) extends `ThreadStore` and `ThreadState`. If the repo author is refactoring `AppModel` to reduce struct weight (item #8 in system.md), coordinate on the store API shape first. However, Phase 4 only *adds* fields and methods — it doesn't modify existing ones — so merge conflicts would be minimal.

### Should Wait (Hard Dependency)

- **Phase 4's thread-to-session mapping** benefits from the existing roadmap item #3 ("Thread-to-interactive-session mapping"). That work establishes the pattern for linking PTY sessions to thread IDs in the proto channel. Starting Phase 4 before that lands is possible but would likely require rework. **Recommendation:** Begin Phases 1-3 immediately; start Phase 4 after roadmap item #3 merges.

---

## Architecture Overview

The core idea is a **Provider interface** that abstracts the IPC layer, letting the same TUI, state store, and canvas render agents from any runtime.

```
┌──────────────────────────────────────────────────┐
│                   TUI Layer                       │
│   Canvas  │  Tree  │  Session Pane  │  Approvals  │
├──────────────────────────────────────────────────┤
│                 State Store                       │
│   ThreadStore + TeamStore + ContextBus            │
├──────────────────────────────────────────────────┤
│               Provider Interface                  │
├────────────────────┬─────────────────────────────┤
│  CodexProvider     │  ClaudeCodeProvider          │
│  (existing client) │  (new adapter)               │
└────────────────────┴─────────────────────────────┘
```

---

## Phase 1: Provider Interface

**Goal:** Abstract the IPC layer so multiple agent runtimes can coexist.
**New files only — no modifications to existing code.**

### 1.1 Define the Provider interface

Create `internal/provider/provider.go`:

```go
type Provider interface {
    Start(ctx context.Context) error
    Stop() error
    Running() bool
    ReadLoop(handler func(Event))
    SendApproval(requestID string, approved bool)
    CreateSession(opts SessionOpts) (string, error)
}

type Event struct {
    Type      string
    ThreadID  string
    ParentID  string
    Raw       json.RawMessage
}

type SessionOpts struct {
    ThreadID string
    Role     string
    Prompt   string
    Tools    []string
    Worktree bool
}
```

### 1.2 Wrap existing Client as CodexProvider

Create `internal/provider/codex.go` — a thin adapter that wraps `appserver.Client` and satisfies the `Provider` interface. The existing `Client` stays untouched; the adapter translates `JSONRPCMessage` into `provider.Event`.

### 1.3 Update AppModel to accept Provider

Add a `WithProvider(p provider.Provider)` option alongside the existing `WithClient`. The TUI checks for Provider first, falls back to direct Client access. This keeps backward compatibility — existing `main.go` continues working without changes until Phase 2.

**Estimated scope:** ~200 lines across 3 new files. 0 lines modified in existing code.

---

## Phase 2: Claude Code Adapter

**Goal:** Implement a ClaudeCodeProvider that spawns and communicates with Claude Code.

### 2.1 Research Claude Code's IPC surface

Claude Code can be invoked as a subprocess. Key integration points:

- `claude --print` for non-interactive single-shot tasks
- `claude` CLI with `--output-format stream-json` for streaming structured events
- The Claude Agent SDK for programmatic Go integration (if Go bindings exist)
- Sub-agent spawning via `--agent` or Agent tool invocations

The adapter needs to:
- Spawn `claude` as a child process (similar to how `appserver.Client` spawns `codex proto`)
- Parse streaming JSON output into `provider.Event` structs
- Forward user input / approval responses back via stdin

### 2.2 Implement ClaudeCodeProvider

Create `internal/provider/claudecode.go`:

- `Start()` — spawn `claude` with appropriate flags, set up stdio pipes
- `ReadLoop()` — parse streaming JSON events, map to `provider.Event`
- `SendApproval()` — write approval response to stdin
- `CreateSession()` — spawn a new `claude` subprocess (one per agent), or use the Agent SDK's task delegation

### 2.3 Event mapping

Map Claude Code's event stream to DJ's existing message types:

| Claude Code Event        | DJ Event / Message        |
|--------------------------|---------------------------|
| Session started          | ThreadStartedMsg          |
| Tool use requested       | V2ExecApprovalMsg         |
| File edit requested      | V2FileApprovalMsg         |
| Text delta               | V2AgentDeltaMsg           |
| Sub-agent spawned        | CollabSpawnMsg            |
| Sub-agent completed      | CollabCloseMsg            |
| Session completed        | TurnCompletedMsg          |
| Error                    | AppServerErrorMsg         |

**Estimated scope:** ~300 lines. New files only.

---

## Phase 3: Config Extensions

**Goal:** Let users configure Claude Code as their provider and define agent teams in `dj.toml`.

### 3.1 Provider selection

Extend `config.go` with a new top-level `provider` field:

```toml
# dj.toml
provider = "claude-code"  # or "codex" (default)

[claude-code]
command = "claude"
model = "claude-sonnet-4-6"
max_tokens = 100000

[codex]
command = "codex"
args = ["proto"]
```

### 3.2 Team topology config

Add a `[[team.agents]]` array for defining multi-agent teams:

```toml
[[team.agents]]
name = "lead"
role = "orchestrator"
model = "claude-sonnet-4-6"
prompt = "You coordinate the team. Delegate research to @researcher and code to @coder."
tools = ["Agent", "Read", "Write", "Bash"]

[[team.agents]]
name = "researcher"
role = "researcher"
model = "claude-sonnet-4-6"
prompt = "You research codebases and provide analysis."
tools = ["Read", "Grep", "Glob", "WebSearch"]

[[team.agents]]
name = "coder"
role = "coder"
model = "claude-sonnet-4-6"
prompt = "You write and edit code."
tools = ["Read", "Write", "Edit", "Bash"]
worktree = true
```

### 3.3 Config types

Add to `internal/config/config.go`:

```go
type ClaudeCodeConfig struct {
    Command   string
    Model     string
    MaxTokens int
}

type TeamAgentConfig struct {
    Name     string
    Role     string
    Model    string
    Prompt   string
    Tools    []string
    Worktree bool
}

type TeamConfig struct {
    Agents []TeamAgentConfig
}
```

**Estimated scope:** ~80 lines added to config.go, new config_test.go cases.

---

## Phase 4: Team Orchestration & Inter-Agent Messaging

**Goal:** Enable agents to communicate, share context, and coordinate work.

> **Dependency note:** This phase benefits from roadmap item #3 (thread-to-session mapping). Begin after that lands, or accept potential rework.

### 4.1 Extend ThreadState for team metadata

Add fields to `internal/state/thread.go`:

```go
type ThreadState struct {
    // ... existing fields ...
    Provider    string            // "codex" or "claude-code"
    TeamRole    string            // from config: "orchestrator", "researcher", "coder"
    Tools       []string          // permitted tools for this agent
    WorktreeDir string            // git worktree path, if isolated
    TokensUsed  int64             // running token count
    CostCents   int64             // running cost estimate
}
```

### 4.2 TeamStore

Create `internal/state/team.go`:

```go
type TeamStore struct {
    mu       sync.RWMutex
    agents   map[string]*TeamAgent  // name -> agent config + state
    messages []TeamMessage          // inter-agent message log
}

type TeamAgent struct {
    Config   TeamAgentConfig
    ThreadID string  // linked thread in ThreadStore
    Status   string
}

type TeamMessage struct {
    From      string
    To        string
    Type      string  // "task", "result", "context", "review"
    Content   string
    Timestamp time.Time
}
```

### 4.3 Context bus

Create `internal/state/context_bus.go` — a pub/sub channel where agents broadcast file changes, decisions, and task completions. The TUI subscribes to render inter-agent communication on the canvas (connection lines between cards, message badges).

### 4.4 Team startup sequence

When `provider = "claude-code"` and `[[team.agents]]` is configured:

1. For each agent in config, call `provider.CreateSession()` with role/prompt/tools
2. Register each session in `TeamStore` and `ThreadStore`
3. Start the lead agent's turn first; sub-agents idle until delegated to
4. When lead agent uses the `Agent` tool (spawning a sub-agent), match it to a pre-configured team agent by role name

**Estimated scope:** ~400 lines across 3 new files + minor additions to ThreadState.

---

## Phase 5: Unified Approval Queue

**Goal:** Aggregate tool-use approval requests from all running agents into one UI.

### 5.1 Approval queue model

Create `internal/tui/approval_queue.go`:

- Collects `V2ExecApprovalMsg` and `V2FileApprovalMsg` from all providers
- Renders as a bottom panel or overlay (toggled with `a` key)
- Shows: agent name, tool being used, command/file, approve/deny controls
- Keyboard: `y` approve, `n` deny, `j/k` navigate queue items

### 5.2 Auto-approve policies

Extend team config with approval rules:

```toml
[[team.agents]]
name = "coder"
auto_approve = ["Read", "Grep", "Glob"]  # auto-approve these tools
# all other tools require manual approval
```

### 5.3 Wire into AppModel

Add `approvalQueue ApprovalQueueModel` to `AppModel`. Route approval messages through the queue instead of auto-approving (current behavior in `handleV2ExecApproval`).

**Estimated scope:** ~250 lines. New files + small changes to app.go message routing.

---

## Phase 6: Token & Cost Dashboard

**Goal:** Real-time visibility into per-agent and aggregate token usage.

### 6.1 Handle MethodTokenUsageUpdated

The event constant already exists in `methods_v2.go`. Add:
- A `TokenUsageMsg` Bubble Tea message type
- A decoder in `bridge_v2.go`
- A handler in `app_proto_v2.go` that updates `ThreadState.TokensUsed`

### 6.2 Cost estimation

Add model pricing to config:

```toml
[pricing]
"claude-sonnet-4-6" = { input = 3.0, output = 15.0 }  # per 1M tokens
"claude-opus-4-6" = { input = 15.0, output = 75.0 }
```

### 6.3 Dashboard widget

Create `internal/tui/cost_bar.go` — renders in the status bar area:
- Per-agent: `lead: $0.12 (8.2k tok) | coder: $0.03 (1.1k tok)`
- Aggregate: `Total: $0.15`
- Optional budget alert when threshold exceeded

**Estimated scope:** ~200 lines across new files + small additions to existing handlers.

---

## Phase 7: Git Worktree Integration

**Goal:** Agents working on code get isolated worktrees; DJ manages lifecycle and reconciliation.

### 7.1 Worktree manager

Create `internal/worktree/manager.go`:

- `Create(branchName string) (path string, err error)` — `git worktree add`
- `Remove(path string) error` — cleanup on agent completion
- `Diff(path string) (string, error)` — show changes vs main
- `Merge(path string) error` — merge agent's worktree back

### 7.2 Agent worktree lifecycle

When an agent with `worktree = true` starts:
1. Create a worktree: `git worktree add /tmp/dj-<agent>-<id> -b dj/<agent>-<id>`
2. Pass the worktree path to the Claude Code subprocess as working directory
3. On completion, show a diff overlay before merging or discarding

### 7.3 Diff overlay in TUI

Create `internal/tui/diff_overlay.go`:
- Triggered when a worktree agent completes
- Shows unified diff with syntax highlighting (Lipgloss)
- Keys: `m` merge, `d` discard, `e` edit (open in PTY session)

**Estimated scope:** ~300 lines across 2 new packages + TUI overlay.

---

## Implementation Order & Timeline

```
Phase 1: Provider Interface         ← Start immediately (no dependencies)
Phase 2: Claude Code Adapter        ← Start after Phase 1
Phase 3: Config Extensions          ← Parallel with Phase 2
Phase 4: Team Orchestration         ← After roadmap item #3 lands
Phase 5: Approval Queue             ← After Phase 2 (needs real events to test)
Phase 6: Token/Cost Dashboard       ← Anytime after Phase 1
Phase 7: Git Worktree Integration   ← After Phase 4 (needs team context)
```

Phases 1-3 are the foundation and can proceed immediately without touching existing code. Phase 4 is the critical integration point where this work intersects with the existing roadmap.

## File Impact Summary

### New Files (no conflict risk)

```
internal/provider/provider.go        # Provider interface
internal/provider/codex.go           # Codex adapter (wraps existing Client)
internal/provider/claudecode.go      # Claude Code adapter
internal/provider/claudecode_test.go
internal/state/team.go               # TeamStore
internal/state/team_test.go
internal/state/context_bus.go        # Inter-agent pub/sub
internal/state/context_bus_test.go
internal/tui/approval_queue.go       # Approval queue UI
internal/tui/approval_queue_test.go
internal/tui/cost_bar.go             # Token/cost widget
internal/tui/cost_bar_test.go
internal/tui/diff_overlay.go         # Worktree diff view
internal/tui/diff_overlay_test.go
internal/worktree/manager.go         # Git worktree lifecycle
internal/worktree/manager_test.go
docs/claude-code-agent-teams.md      # This document
```

### Modified Files (minimal, additive only)

```
internal/config/config.go            # Add ClaudeCodeConfig, TeamConfig types
internal/state/thread.go             # Add Provider, TeamRole, TokensUsed fields
internal/tui/app.go                  # Add WithProvider option, approval queue field
internal/tui/bridge_v2.go            # Add TokenUsageMsg case
internal/tui/app_proto_v2.go         # Add token usage handler
cmd/dj/main.go                       # Provider selection from config
```

### Untouched (existing Codex work proceeds freely)

```
internal/appserver/*                 # All existing IPC code
internal/tui/canvas.go               # Canvas rendering
internal/tui/card.go                 # Card rendering
internal/tui/tree.go                 # Tree view
internal/tui/pty_session.go          # PTY management
internal/tui/pty_keys.go             # Key encoding
internal/tui/menu.go                 # Context menu
internal/tui/help.go                 # Help overlay
internal/tui/prefix.go               # Ctrl+B handler
internal/tui/statusbar.go            # Status bar
```

---

## Open Questions

1. **Claude Code's subprocess protocol** — Does `claude --output-format stream-json` emit structured events granular enough for real-time TUI rendering (deltas, tool calls, sub-agent lifecycle)? If not, the Agent SDK may be required.

2. **Agent SDK Go bindings** — The Claude Agent SDK is primarily TypeScript/Python. If no Go bindings exist, the adapter would need to either shell out to a thin Node/Python bridge or parse CLI output directly.

3. **Multi-agent context window** — How should shared context (file contents, decisions) be passed between agents? Options: shared filesystem, explicit context injection via prompts, or a coordination protocol.

4. **Worktree merge conflicts** — When two agents edit overlapping files in separate worktrees, how should conflicts be surfaced? Consider a dedicated "conflict resolution" agent role.
