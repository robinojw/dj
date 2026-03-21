# Implementation Plan: Claude Code Agent Teams Support

**Branch:** `nick/claude-code-agent-teams`
**Author:** Nick Hess
**Date:** 2026-03-20
**Status:** Proposal (rev 2 — addresses PR #50 review)

---

## Goal

Extend DJ to support Claude Code agent sessions alongside the existing Codex integration, with first-class support for multi-agent team workflows — where a lead agent orchestrates sub-agents with distinct roles (researcher, coder, reviewer), shared context, and coordinated approvals.

## Relationship to Existing Development

The existing roadmap (documented in `system.md`) focuses on completing the Codex v2 protocol integration. This plan is designed to **run in parallel** without blocking that work, with one structural dependency noted below.

### Can Be Done Independently (Low Conflict Risk)

- **Phases 0-2** (IPC spike, Provider interface, approval safety gate) introduce new packages and files. Phase 1 adds a `WithProvider` option to `app.go` and Phase 2 adds approval routing changes to `app_proto_v2.go` — both are small, additive modifications.
- **Phase 6** (token/cost dashboard) builds on the already-stubbed `MethodTokenUsageUpdated` event, adding a decoder to `bridge_v2.go` and a handler to `app_proto_v2.go`.

### Requires Coordination (Soft Dependencies)

- **Phase 5** (team orchestration) extends `ThreadStore` and `ThreadState`. If the repo author is refactoring `AppModel` to reduce struct weight (item #8 in system.md), coordinate on the store API shape first. However, Phase 5 only *adds* fields and methods — it doesn't modify existing ones — so merge conflicts would be minimal.

### Should Wait (Hard Dependency)

- **Phase 5's thread-to-session mapping** benefits from the existing roadmap item #3 ("Thread-to-interactive-session mapping"). That work establishes the pattern for linking PTY sessions to thread IDs in the proto channel. Starting Phase 5 before that lands is possible but would likely require rework. **Recommendation:** Begin Phases 0-4 immediately; start Phase 5 after roadmap item #3 merges.

---

## Architecture Overview

The core idea is a **Provider interface** that abstracts the IPC layer, letting the same TUI, state store, and canvas render agents from any runtime.

```
┌──────────────────────────────────────────────────┐
│                   TUI Layer                       │
│   Canvas  │  Tree  │  Session Pane  │  Approvals  │
├──────────────────────────────────────────────────┤
│                 State Store                       │
│   ThreadStore + TeamStore (tea.Msg-driven)        │
├──────────────────────────────────────────────────┤
│               Provider Interface                  │
├────────────────────┬─────────────────────────────┤
│  CodexProvider     │  ClaudeCodeProvider          │
│  (existing client) │  (new adapter)               │
└────────────────────┴─────────────────────────────┘
```

---

## Phase 0: Claude IPC Feasibility Spike

**Goal:** Validate that Claude Code's subprocess interface provides sufficient event granularity for real-time TUI rendering before committing to an adapter implementation.

### 0.1 Acceptance criteria

This phase gates all subsequent Claude-specific work. It is complete when:

1. **Streaming output captured:** Run `claude --output-format stream-json` against a non-trivial prompt and capture the full event stream. Document the event taxonomy (event types, field shapes, timing).
2. **Approval granularity confirmed:** Verify that tool-use requests (exec, file edit) arrive as discrete events with enough metadata to render an approval prompt (tool name, arguments, file paths).
3. **Sub-agent lifecycle observable:** Confirm whether Agent tool invocations produce start/complete events for sub-agents, or whether DJ must infer lifecycle from output heuristics.
4. **Input channel validated:** Confirm how approval responses and user input are sent back to the Claude process (stdin JSON, signals, or other mechanism).
5. **Minimal mapping produced:** Map at least one Claude event to one DJ `tea.Msg` type and render it in a throwaway test harness.

### 0.2 Deliverables

- `docs/claude-ipc-spike.md` — event taxonomy, mapping feasibility, identified gaps
- Decision: CLI streaming vs Agent SDK vs hybrid approach
- Go/no-go for Phase 3 (Claude adapter)

**Done when:** Spike document exists with captured event samples and a clear recommendation for the adapter strategy.

---

## Phase 1: Provider Interface

**Goal:** Abstract the IPC layer so multiple agent runtimes can coexist.

### 1.1 Define the Provider interface

Create `internal/provider/provider.go`:

```go
type Provider interface {
    Start(ctx context.Context) error
    Stop() error
    Running() bool
    ReadLoop(ctx context.Context, handler func(Event)) error
    SendApproval(ctx context.Context, approval ApprovalResponse) error
    SendInput(ctx context.Context, threadID string, input []byte) error
    CreateSession(ctx context.Context, opts SessionOpts) (correlationID string, err error)
}
```

Design decisions addressing review feedback:

- **`ReadLoop` accepts `context.Context`** — fixes the cancellation gap in the current `Client.ReadLoop` which has no stop signal. The provider must check ctx on each iteration and return when cancelled.
- **`ReadLoop` returns `error`** — callers can distinguish clean shutdown from I/O failures.
- **`SendApproval` takes a typed `ApprovalResponse`** — a single `bool` doesn't carry enough info for Claude's tool approval model (which may need tool-specific metadata). See below.
- **`SendInput` added** — DJ's core interaction model is PTY-first. This method routes user keystrokes to the agent session.
- **`CreateSession` returns a correlation ID, not a thread ID** — Codex thread creation is notification-driven (`thread/started` event). The actual thread ID arrives asynchronously via a `ThreadStartedMsg`. The correlation ID lets the caller match the response to the request.

### 1.2 Typed events

The Provider emits **typed, provider-neutral event structs** rather than raw envelopes. This avoids pushing per-provider decoding into the TUI layer.

```go
type Event struct {
    Kind     EventKind
    ThreadID string
    ParentID string
    Payload  any
}

type EventKind int

const (
    EventThreadStarted EventKind = iota
    EventThreadClosed
    EventTurnStarted
    EventTurnCompleted
    EventAgentDelta
    EventExecApproval
    EventFileApproval
    EventTokenUsage
    EventSubAgentSpawned
    EventSubAgentCompleted
    EventError
)

type ApprovalResponse struct {
    RequestID string
    Approved  bool
    Tool      string
    Metadata  map[string]any
}
```

Each provider is responsible for mapping its wire format into these typed events. The TUI bridge then converts `provider.Event` → `tea.Msg` in a single, provider-agnostic switch.

### 1.3 Wrap existing Client as CodexProvider

Create `internal/provider/codex.go` — a thin adapter that wraps `appserver.Client` and satisfies the `Provider` interface. The existing `Client` stays untouched; the adapter translates `JSONRPCMessage` into `provider.Event`.

`CreateSession` for CodexProvider: returns a correlation ID immediately. The actual thread ID arrives via a `ThreadStarted` event that includes the correlation ID for matching.

### 1.4 Update AppModel to accept Provider

Add a `WithProvider(p provider.Provider)` option to `app.go` alongside the existing `WithClient`. The TUI checks for Provider first, falls back to direct Client access. This keeps backward compatibility — existing `main.go` continues working without changes until Phase 3.

**Files modified:** `internal/tui/app.go` (add `WithProvider` option).
**Estimated scope:** ~250 lines across 3 new files + 1 modified file.

**Done when:** CodexProvider passes all existing integration tests with no behavior change. `ReadLoop` respects context cancellation.

---

## Phase 2: Approval Safety Gate

**Goal:** Replace auto-approve with a manual-by-default approval model before any multi-agent work begins.

> **Why this comes before the Claude adapter:** Today `handleV2ExecApproval` and `handleV2FileApproval` in `app_proto_v2.go` unconditionally call `SendApproval(..., true)`. Introducing Claude sessions without an approval gate means all tool calls would auto-approve — a serious safety gap, especially in multi-agent mode.

### 2.1 Approval queue model

Create `internal/tui/approval_queue.go`:

- Collects `ExecApproval` and `FileApproval` events from all providers
- Renders as a bottom panel or overlay (toggled with `a` key)
- Shows: agent name, tool being used, command/file, approve/deny controls
- Keyboard: `y` approve, `n` deny, `j/k` navigate queue items

### 2.2 Default approval policy

- **Non-Codex providers:** Manual approval for all tools by default
- **Codex provider:** Preserves current auto-approve behavior (backward compatible)
- **Team mode (any provider):** Manual approval for all tools by default
- Auto-approve is opt-in per tool via allowlist in config:

```toml
[[team.agents]]
name = "coder"
auto_approve = ["Read", "Grep", "Glob"]
```

### 2.3 Wire into AppModel

Add `approvalQueue ApprovalQueueModel` to `AppModel`. Route approval messages through the queue instead of auto-approving. The queue checks the policy config and either auto-approves (if allowlisted) or enqueues for manual review.

**Files modified:** `internal/tui/app.go` (queue field), `internal/tui/app_proto_v2.go` (route through queue).
**Estimated scope:** ~250 lines. New files + small changes to existing routing.

**Done when:** Codex sessions still auto-approve (no behavior change). Non-Codex providers require manual approval. Queue renders and responds to keyboard input.

---

## Phase 3: Claude Code Adapter

**Goal:** Implement a ClaudeCodeProvider that spawns and communicates with Claude Code.

> **Gate:** Phase 0 must be complete with a go decision before starting this phase.

### 3.1 Implement ClaudeCodeProvider

Create `internal/provider/claudecode.go`:

- `Start()` — spawn `claude` with appropriate flags, set up stdio pipes
- `ReadLoop()` — parse streaming events, map to typed `provider.Event` structs, respect `context.Context` for cancellation
- `SendApproval()` — write typed approval response back to the Claude process
- `SendInput()` — route user keystrokes to the Claude session
- `CreateSession()` — spawn a new `claude` subprocess (one per agent), return correlation ID

### 3.2 Session pane integration

The MVP session pane for Claude renders **streamed text deltas in a virtual terminal view**, not a full PTY attachment.

- Claude's output arrives as structured `AgentDelta` events containing text content
- The session pane accumulates deltas into a scrollable text buffer (similar to a log viewer)
- User input is forwarded via `SendInput` for approval responses and conversational replies
- Full PTY integration (if needed) is deferred to after roadmap item #3 establishes the pattern

This means Claude sessions initially look more like a streaming log than an interactive terminal. If Phase 0 reveals that Claude supports a PTY-attachable mode, this decision can be revisited.

### 3.3 Event mapping

Map Claude Code's event stream to DJ's typed Provider events. Based on Phase 0 findings, mappings will fall into two categories:

**Confirmed mappings** (validated during Phase 0 spike):

| Claude Code Event        | Provider EventKind         |
|--------------------------|----------------------------|
| *(populated after Phase 0)* | |

**Desired mappings** (pending Phase 0 validation):

| Claude Code Event        | Provider EventKind         | DJ Message Type     |
|--------------------------|----------------------------|---------------------|
| Session started          | EventThreadStarted         | ThreadStartedMsg    |
| Tool use requested       | EventExecApproval          | V2ExecApprovalMsg   |
| File edit requested      | EventFileApproval          | V2FileApprovalMsg   |
| Text delta               | EventAgentDelta            | V2AgentDeltaMsg     |
| Sub-agent spawned        | EventSubAgentSpawned       | CollabSpawnMsg      |
| Sub-agent completed      | EventSubAgentCompleted     | CollabCloseMsg      |
| Session completed        | EventTurnCompleted         | TurnCompletedMsg    |
| Error                    | EventError                 | AppServerErrorMsg   |

**Estimated scope:** ~300 lines. New files only.

**Done when:** A single Claude Code session can start, render streaming output in the session pane, and respond to at least one approval request through the approval queue.

---

## Phase 4: Config Extensions

**Goal:** Let users configure Claude Code as their provider and define agent teams in `dj.toml`.

### 4.1 Provider selection

Extend `config.go` with provider configuration:

```toml
# dj.toml
provider = "claude_code"  # or "codex" (default)

[providers.codex]
command = "codex"
args = ["proto"]

[providers.claude_code]
command = "claude"
model = "claude-sonnet-4-6"
max_tokens = 100000
```

Design decisions:
- **`[providers.X]` namespace** avoids top-level key pollution and works cleanly with Viper's dot-path lookups (no hyphenated keys).
- **`[appserver]` retained as alias** — existing `[appserver]` config maps to `[providers.codex]` internally. Users with existing configs see zero behavior change.
- **`provider = "codex"` is the default** — omitting the field preserves current behavior.

### 4.2 Backward compatibility

Existing users with only `[appserver]`, `[interactive]`, and `[ui]` sections see no behavior change:

1. If `provider` key is absent, default to `"codex"`
2. If `[appserver]` exists but `[providers.codex]` does not, treat `[appserver]` as the codex provider config
3. If both exist, `[providers.codex]` takes precedence (with a warning log)
4. New `[providers.*]` sections are entirely optional

### 4.3 Team topology config

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
auto_approve = ["Read", "Grep", "Glob"]

[[team.agents]]
name = "coder"
role = "coder"
model = "claude-sonnet-4-6"
prompt = "You write and edit code."
tools = ["Read", "Write", "Edit", "Bash"]
worktree = true
```

### 4.4 Config types

Add to `internal/config/config.go`:

```go
type ProviderConfig struct {
    Codex     CodexProviderConfig     `mapstructure:"codex"`
    ClaudeCode ClaudeCodeProviderConfig `mapstructure:"claude_code"`
}

type CodexProviderConfig struct {
    Command string
    Args    []string
}

type ClaudeCodeProviderConfig struct {
    Command   string
    Model     string
    MaxTokens int `mapstructure:"max_tokens"`
}

type TeamAgentConfig struct {
    Name        string
    Role        string
    Model       string
    Prompt      string
    Tools       []string
    AutoApprove []string `mapstructure:"auto_approve"`
    Worktree    bool
}

type TeamConfig struct {
    Agents []TeamAgentConfig
}
```

**Files modified:** `internal/config/config.go` (add types and Viper bindings).
**Estimated scope:** ~100 lines added to config.go, new config_test.go cases.

**Done when:** Existing `dj.toml` files with `[appserver]`/`[interactive]`/`[ui]` continue to work with no changes. New `[providers.*]` and `[[team.agents]]` sections parse correctly.

---

## Phase 5: Team Orchestration & Inter-Agent Messaging

**Goal:** Enable agents to communicate, share context, and coordinate work.

> **Dependency note:** This phase benefits from roadmap item #3 (thread-to-session mapping). Begin after that lands, or accept potential rework.

### 5.1 Extend ThreadState for team metadata

Add fields to `internal/state/thread.go`:

```go
type ThreadState struct {
    // ... existing fields ...
    Provider    string
    TeamRole    string
    Tools       []string
    WorktreeDir string
    InputTokens  int64
    OutputTokens int64
    CostCents    int64
}
```

Token tracking uses separate `InputTokens` and `OutputTokens` fields to support accurate cost calculation (pricing differs per direction). If a provider only reports aggregate tokens, set `InputTokens` to the total and `OutputTokens` to zero, and note degraded cost accuracy in the dashboard.

### 5.2 TeamStore

Create `internal/state/team.go`. TeamStore state updates flow through `tea.Msg` types to match DJ's existing concurrency pattern (single-goroutine `Update()` in Bubble Tea):

```go
type TeamUpdateMsg struct {
    AgentName string
    ThreadID  string
    Status    string
}

type TeamMessageMsg struct {
    From      string
    To        string
    Kind      string
    Content   string
    Timestamp time.Time
}

type TeamStore struct {
    agents   map[string]*TeamAgent
    messages []TeamMessage
}
```

The `TeamStore` has **no mutex** — it is only mutated inside `AppModel.Update()` in response to `TeamUpdateMsg` and `TeamMessageMsg` values, matching how `ThreadStore` updates are triggered from the TUI layer. External goroutines (providers, JSON-RPC readers) inject updates via `program.Send(TeamUpdateMsg{...})`.

### 5.3 Inter-agent communication

Rather than a separate `ContextBus` pub/sub system, inter-agent events are modeled as `tea.Msg` types routed through Bubble Tea's existing `program.Send`. This avoids adding concurrency surface area when DJ already has a message bus.

Team event message types:
- `TeamContextShareMsg` — an agent broadcasts file changes, decisions, or task completions
- `TeamTaskDelegateMsg` — lead agent assigns work to a sub-agent
- `TeamTaskCompleteMsg` — sub-agent reports completion

The `TeamStore` logs these for display (connection lines between canvas cards, message badges). The TUI subscribes by handling them in `Update()`.

### 5.4 Team startup sequence

When `provider = "claude_code"` and `[[team.agents]]` is configured:

1. For each agent in config, call `provider.CreateSession()` with role/prompt/tools — returns correlation ID
2. Match incoming `ThreadStartedMsg` events to correlation IDs and register in `TeamStore`
3. Start the lead agent's turn first; sub-agents idle until delegated to
4. When lead agent uses the `Agent` tool (spawning a sub-agent), match it to a pre-configured team agent by role name

**Estimated scope:** ~350 lines across 2 new files + minor additions to ThreadState.

**Done when:** A team of 2+ agents can start, with the lead agent delegating to a sub-agent. TeamStore reflects agent states. Canvas shows team relationships.

---

## Phase 6: Token & Cost Dashboard

**Goal:** Real-time visibility into per-agent and aggregate token usage.

### 6.1 Handle MethodTokenUsageUpdated

The event constant already exists in `methods_v2.go`. Add:
- A `TokenUsageMsg` Bubble Tea message type with `InputTokens` and `OutputTokens` fields
- A decoder in `bridge_v2.go`
- A handler in `app_proto_v2.go` that updates `ThreadState.InputTokens` and `ThreadState.OutputTokens`

### 6.2 Cost estimation

Add model pricing to config:

```toml
[pricing]
[pricing."claude-sonnet-4-6"]
input = 3.0    # per 1M tokens
output = 15.0

[pricing."claude-opus-4-6"]
input = 15.0
output = 75.0
```

Cost calculation: `(InputTokens * inputRate + OutputTokens * outputRate) / 1_000_000`. When only aggregate tokens are available (OutputTokens = 0), use the input rate as a lower-bound estimate and mark the cost display with `~` to indicate approximation.

### 6.3 Dashboard widget

Create `internal/tui/cost_bar.go` — renders in the status bar area:
- Per-agent: `lead: $0.12 (8.2k tok) | coder: $0.03 (1.1k tok)`
- Aggregate: `Total: $0.15`
- Optional budget alert when threshold exceeded

**Files modified:** `internal/tui/bridge_v2.go` (add decoder), `internal/tui/app_proto_v2.go` (add handler).
**Estimated scope:** ~200 lines across new files + small additions to existing handlers.

**Done when:** Token usage events update per-agent counters. Cost bar renders accurate estimates. Aggregate total is correct across all agents.

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

**Done when:** An agent with `worktree = true` gets an isolated git worktree. On completion, the diff overlay shows changes. Merge and discard both work correctly. Worktree is cleaned up on discard.

---

## Implementation Order & Timeline

```
Phase 0: Claude IPC Spike            ← Start immediately (gates Phase 3)
Phase 1: Provider Interface           ← Start immediately (no dependencies)
Phase 2: Approval Safety Gate         ← After Phase 1 (needs Provider contract)
Phase 3: Claude Code Adapter          ← After Phase 0 go-decision + Phase 1
Phase 4: Config Extensions            ← Parallel with Phase 3
Phase 5: Team Orchestration           ← After Phase 3 + roadmap item #3
Phase 6: Token/Cost Dashboard         ← After Phase 1 (needs Provider events)
Phase 7: Git Worktree Integration     ← After Phase 5 (needs team context)
```

Phases 0 and 1 can proceed in parallel. Phase 2 (approval safety) is sequenced before Phase 3 (Claude adapter) so that the first Claude integration ships with a proper approval gate. Phase 4 is the critical integration point where this work intersects with the existing roadmap.

## File Impact Summary

### New Files (no conflict risk)

```
docs/claude-ipc-spike.md             # Phase 0 spike findings
internal/provider/provider.go        # Provider interface + typed events
internal/provider/codex.go           # Codex adapter (wraps existing Client)
internal/provider/claudecode.go      # Claude Code adapter
internal/provider/claudecode_test.go
internal/state/team.go               # TeamStore (tea.Msg-driven)
internal/state/team_test.go
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

### Modified Files (small, additive changes)

```
internal/config/config.go            # Add ProviderConfig, TeamConfig types, backward compat
internal/state/thread.go             # Add Provider, TeamRole, InputTokens, OutputTokens fields
internal/tui/app.go                  # Add WithProvider option, approval queue field
internal/tui/app_proto_v2.go         # Route approvals through queue, add token handler
internal/tui/bridge_v2.go            # Add TokenUsageMsg decoder
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

## Migration: Existing Users

Users with existing `dj.toml` files containing only `[appserver]`, `[interactive]`, and `[ui]` sections see **zero behavior change**:

- The `provider` key defaults to `"codex"` when absent
- `[appserver]` is treated as an alias for `[providers.codex]` when the new key doesn't exist
- No new sections are required — all `[providers.*]` and `[[team.agents]]` config is optional
- The approval queue is inactive for the Codex provider (preserves current auto-approve behavior)

---

## Open Questions (Blocking — Must Resolve in Phase 0)

1. **Claude Code's subprocess protocol** — Does `claude --output-format stream-json` emit structured events granular enough for real-time TUI rendering (deltas, tool calls, sub-agent lifecycle)? Phase 0 must capture real output and document the event taxonomy.

2. **Agent SDK Go bindings** — The Claude Agent SDK is primarily TypeScript/Python. If no Go bindings exist, the adapter would need to either shell out to a thin Node/Python bridge or parse CLI output directly. Phase 0 must determine the integration path.

3. **Approval input channel** — How does the Claude process receive approval responses? stdin JSON? Signals? This determines the `SendApproval` implementation.

4. **Sub-agent lifecycle observability** — Are Agent tool invocations observable as discrete events, or must DJ infer sub-agent lifecycle from output patterns? This determines whether `EventSubAgentSpawned`/`EventSubAgentCompleted` are achievable.

## Open Questions (Non-Blocking — Resolve During Implementation)

5. **Multi-agent context window** — How should shared context (file contents, decisions) be passed between agents? Options: shared filesystem, explicit context injection via prompts, or a coordination protocol.

6. **Worktree merge conflicts** — When two agents edit overlapping files in separate worktrees, how should conflicts be surfaced? Consider a dedicated "conflict resolution" agent role.
