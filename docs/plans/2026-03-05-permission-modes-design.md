# Three-Mode Permission System Design

**Date:** 2026-03-05
**Author:** DJ Agent Enhancement Team
**Status:** Approved

## Overview

This design replaces the current dual-mode agent system (Plan/Build) with a three-mode permission architecture that controls tool access at two layers: filtering tools before API requests and intercepting tool execution at runtime.

## Goals

- Give users fine-grained control over agent autonomy
- Prevent accidental destructive operations in default mode
- Enable full autonomy in isolated environments
- Maintain read-only architectural planning mode
- Provide defense-in-depth through allow/deny lists

## Non-Goals

- Real-time policy updates without restart
- Per-tool custom permission rules beyond allow/deny
- Undo functionality for approved operations (checkpoint system handles this)

## Execution Modes

The system provides three modes, each with distinct tool access and behavior:

### Confirm Mode (Default)

The agent asks permission before executing write, exec, or MCP mutation tools. Read operations proceed automatically. The system prompts the user with tool details and intent before execution. The status bar displays an amber badge: `⏸ CONFIRM`.

Users can approve operations with three remember scopes:
- **Once**: Allow this single call
- **Session**: Allow this tool for the current session
- **Always**: Persist to configuration for all future sessions

### Plan Mode

The agent operates read-only with high reasoning effort. The system filters available tools to `read_file`, `search_code`, and `list_dir` before sending the API request. The agent receives an architect persona system prompt that emphasizes deliberate planning over action. The status bar displays a blue badge: `◎ PLAN`.

### Turbo Mode

The agent bypasses all permission checks. All tools pass through both filter and runtime gates automatically. The system requires explicit user confirmation on first activation per session, warning that the agent can write files, execute commands, and make network requests. The status bar displays a red badge: `⚡ TURBO`.

## Architecture

### Two-Layer Gate System

The permission system intercepts tool calls at two points:

**Layer 1: Filter Phase**

At API request time, the system filters available tools based on mode:
- **Plan mode**: Sends only `["read_file", "search_code", "list_dir"]` to API
- **Confirm/Turbo modes**: Send all available tools to API

This prevents the model from suggesting disallowed tools in Plan mode and reduces wasted API calls.

**Layer 2: Runtime Gate**

When the model calls a tool in the response stream, the gate evaluates the call:
- **Plan mode**: Auto-allow (pre-filtered set is safe)
- **Confirm mode**: Suspend worker, show permission dialog, wait for user decision
- **Turbo mode**: Auto-allow (bypass all checks)

Both layers respect allow/deny lists for defense-in-depth.

### Tool Classification

Tools are classified into security-relevant categories:

| Class | Tools | Behavior |
|-------|-------|----------|
| **Read** | `read_file`, `list_dir`, `search_code` | Auto-allowed in all modes |
| **Write** | `write_file`, `create_file`, `delete_file` | Requires permission in Confirm, blocked in Plan |
| **Exec** | `bash`, `run_script`, `run_tests` | Requires permission in Confirm, blocked in Plan |
| **MCP Mutate** | MCP tools that modify state | Requires permission in Confirm, blocked in Plan |
| **MCP Read** | MCP tools flagged read-only | Auto-allowed in all modes |
| **Network** | `web_fetch`, `http_request` | Requires permission in Confirm, blocked in Plan |

Unknown tools default to Write classification (conservative - requires permission).

### Permission Gate Component

The gate evaluates each tool call using priority-ordered rules:

1. **Deny list always wins** - Checked first regardless of mode
2. **Allow list passes** - Checked second (session or persisted)
3. **Mode-specific logic** - Determines behavior for unlisted tools

```go
type Gate struct {
    mode      ExecutionMode
    allowList []string   // session or persisted
    denyList  []string   // from config
}

func (g *Gate) Evaluate(tool string, args map[string]any) GateDecision {
    if g.isDenied(tool) {
        return GateDeny
    }
    if g.isAllowed(tool) {
        return GateAllow
    }

    class := classifyTool(tool)

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
    }
}
```

## Data Flow

### Worker Tool Execution

When a worker needs to call a tool:

1. **Filter phase**: Worker filters available tools based on mode before building API request
2. **API call**: Model receives filtered tool list and generates response
3. **Tool call arrives**: Response stream emits `response.output_item.added` with `type=function_call`
4. **Runtime gate**: Worker calls `gate.Evaluate(toolName, args)`
5. **Decision handling**:
   - `GateDeny`: Return error, log blocking
   - `GateAllow`: Execute tool immediately
   - `GateAskUser`: Suspend worker, send `PermissionRequest` to TUI

### Permission Request Flow

When `GateAskUser` is returned:

1. Worker creates response channel: `respCh := make(chan PermissionResp, 1)`
2. Worker sends request to TUI: `permRequestCh <- PermissionRequest{ID, WorkerID, Tool, Args, RespCh}`
3. **Worker blocks**: `resp := <-respCh`
4. TUI shows permission modal with tool details
5. User decides: Allow (with scope) or Deny
6. TUI sends response: `respCh <- PermissionResp{Allowed, RememberFor}`
7. Worker unblocks and proceeds or errors
8. If approved with `RememberSession`, adds tool to `gate.allowList`
9. If approved with `RememberAlways`, persists tool to `harness.toml`

### Mode Switching

User presses **Tab** to cycle: Confirm → Plan → Turbo → Confirm

On Turbo activation (first time per session):
1. Check `app.turboConfirmed` flag
2. If false, show warning modal
3. User confirms or cancels
4. Set `app.turboConfirmed = true` on confirmation
5. Update `app.mode` and `app.gate.SetMode()`
6. Update status bar badge

## User Interface

### Permission Modal

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

**Key bindings:**
- `y`: Approve with current scope
- `n` or `Esc`: Deny
- `Tab`: Cycle remember scope (Once → Session → Always)

### Turbo Warning Modal

```
╔══════════════════════════════════════════════════════════════╗
║  ⚡ TURBO MODE WARNING                                       ║
╠══════════════════════════════════════════════════════════════╣
║                                                              ║
║  TURBO bypasses ALL permission prompts.                      ║
║                                                              ║
║  The agent can:                                              ║
║  • Write/delete any files                                    ║
║  • Execute any shell commands                                ║
║  • Make network requests                                     ║
║                                                              ║
║  Only use in isolated/safe environments.                     ║
║                                                              ║
╠══════════════════════════════════════════════════════════════╣
║  [y] Activate Turbo    [n] Cancel                           ║
╚══════════════════════════════════════════════════════════════╝
```

Appears once per session on first Turbo activation.

### Status Bar Integration

Each mode displays a color-coded badge:

```
[⚡ TURBO]  CTX ████░░░░ 34.2%  OUT 1,204  💰$0.0043  ⚡ github-mcp
[◎ PLAN  ]  CTX ██░░░░░░ 12.1%  OUT 847    💰$0.0011
[⏸ CONFIRM]  CTX ███░░░░░ 21.8%  OUT 2,103  💰$0.0029
```

Colors:
- **Turbo**: Red (danger signal)
- **Plan**: Blue (read-only safe)
- **Confirm**: Amber (cautious default)

## Configuration

### harness.toml

```toml
[execution]
default_mode = "confirm"  # confirm | plan | turbo

[execution.allow]
# Auto-approved in all modes (overrides mode restrictions)
tools = [
    "read_file",
    "list_dir",
    "search_code",
    "run_tests",
    "bash(git status*)",   # glob pattern matching
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

Glob patterns support wildcards:
- `bash(git status*)` matches `bash(git status)` and `bash(git status --short)`
- `read_file(.env*)` matches `read_file(.env)` and `read_file(.env.local)`

### Config Struct

```go
type Config struct {
    // ... existing fields ...
    Execution ExecutionConfig `toml:"execution"`
}

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

## Error Handling

### Worker Timeout

If user does not respond to permission request within 5 minutes:
- Worker times out with error: `permission request timed out`
- Display in chat: `[Worker A: permission request timed out after 5m]`
- Request auto-denied

### Multiple Concurrent Requests

If multiple workers request permissions simultaneously:
- Queue requests in App (FIFO)
- Show one modal at a time
- Display queue indicator: `⚠ 3 pending permission requests`
- Process in order received

### Mode Switch During Pending Request

If user presses Tab while permission modal is visible:
- Auto-deny the pending request
- Close modal
- Display message: `[Permission denied - mode changed]`
- New mode takes effect immediately

### Invalid Tool Classification

If tool name not found in `toolClasses` map:
- Default to `ToolWrite` classification (conservative)
- Log warning: `Unknown tool classification: custom_tool`
- Proceed with write-level permission requirements

### Allow/Deny List Conflicts

If a tool appears in both allow and deny lists:
- Deny always wins (security floor)
- Log warning at startup: `Conflict: bash(git push*) in both allow and deny - denying`
- Remove from allow list in memory

### Plan Mode Escape Attempts

If model tries to call write tools in Plan mode (rare due to filtering):
- Block at gate with `GateDeny`
- Do not show error to model (prevents confusion/retries)
- Log to debug: `Plan mode blocked attempt to call write_file`

## Testing Strategy

### Unit Tests

**Gate logic** (`internal/modes/gate_test.go`):
- Deny list blocks in all modes
- Allow list passes in all modes
- Confirm mode returns `GateAskUser` for writes
- Plan mode returns `GateDeny` for writes
- Turbo mode returns `GateAllow` for everything
- Glob pattern matching works correctly

**Permission flow** (`internal/modes/permission_test.go`):
- Request/response channel communication
- Remember scope handling (Once, Session, Always)
- Timeout behavior after 5 minutes

**Config parsing** (`config/config_test.go`):
- Execution config loads from TOML
- Default mode validation
- Allow/deny list parsing
- Conflict detection

### Integration Tests

**TUI behavior**:
- Modal shows on `GateAskUser`
- Modal hides on decision
- Key bindings work (y/n/Tab/Esc)
- Scope cycles correctly
- Turbo warning appears once per session

**Worker suspension**:
- Worker blocks on permission request
- Worker unblocks after user decision
- Session-scoped allow list persists across calls
- Deny causes proper error propagation

### Manual Testing Checklist

- [ ] Tab cycles: Confirm → Plan → Turbo → Confirm
- [ ] Status bar shows correct badge for each mode
- [ ] Plan mode blocks writes without prompting
- [ ] Confirm mode shows permission dialog for writes
- [ ] Confirm mode auto-allows reads
- [ ] Turbo mode bypasses all prompts
- [ ] Turbo warning appears on first activation
- [ ] "Always" scope persists to harness.toml
- [ ] "Session" scope persists within session only
- [ ] Deny list blocks tools in all modes
- [ ] Allow list overrides mode restrictions
- [ ] Multiple concurrent requests queue properly
- [ ] Mode switch cancels pending requests
- [ ] Permission timeout works after 5 minutes

## Migration Plan

### Breaking Changes

The current `AgentMode` enum changes from `{ModePlan, ModeBuild}` to `ExecutionMode` with `{ModeConfirm, ModePlan, ModeTurbo}`.

**Affected files:**
- `internal/agents/modes.go`: Rename and restructure
- `internal/agents/worker.go`: Update mode handling
- `internal/agents/orchestrator.go`: Update mode passing
- `internal/tui/app.go`: Replace Tab toggle logic
- `internal/tui/screens/chat.go`: Update mode field
- `internal/tui/components/statusbar.go`: Update badge rendering

**Migration steps:**
1. Add new `ExecutionMode` alongside `AgentMode` temporarily
2. Add gate infrastructure (filter + runtime)
3. Wire permission request channels
4. Add TUI modal component
5. Update status bar badges
6. Remove old `AgentMode` and `ModeBuild`
7. Update all references to use `ExecutionMode`

### Backward Compatibility

No backward compatibility required. This is a breaking change to an unreleased feature (agent modes were added in this worktree branch).

### Default Behavior

On first run after upgrade:
- Mode defaults to `ModeConfirm` (safe default)
- No tools in allow list (only reads auto-allowed)
- No tools in deny list
- User must explicitly opt into Turbo mode

## Implementation Checklist

- [ ] Create `internal/modes/` package
- [ ] Define `ExecutionMode` enum and mode configs
- [ ] Implement `Gate` struct with `Evaluate()` method
- [ ] Add tool classification map
- [ ] Implement glob pattern matching for allow/deny
- [ ] Add `PermissionRequest` and `PermissionResp` types
- [ ] Create `PermissionModal` component
- [ ] Wire permission request channel in `App`
- [ ] Add worker suspension logic
- [ ] Implement remember scope handling
- [ ] Add session-scoped allow list
- [ ] Implement persist-to-config for "Always"
- [ ] Create Turbo warning modal
- [ ] Update status bar badge rendering
- [ ] Add Tab key cycling for three modes
- [ ] Update config struct for execution section
- [ ] Add config parsing for allow/deny lists
- [ ] Write unit tests for gate logic
- [ ] Write unit tests for permission flow
- [ ] Write integration tests for TUI
- [ ] Add manual testing checklist
- [ ] Update documentation

## Security Considerations

### Deny List as Security Floor

The deny list enforces a security floor that no mode can bypass. Even Turbo mode respects deny list rules. This prevents accidental configuration of overly permissive modes.

### Default-Deny for Unknown Tools

Unknown tools default to `ToolWrite` classification, requiring permission in Confirm mode and blocking in Plan mode. This prevents privilege escalation through unclassified tools.

### Glob Pattern Wildcards

Glob patterns in allow/deny lists support `*` wildcards but not arbitrary regex. This prevents overly broad patterns like `.*` from bypassing security controls.

### MCP Tool Classification

MCP tools must be classified at registration time based on server metadata. MCP servers should declare which tools are read-only. Unknown MCP tools default to `ToolMCPMutate` (requires permission).

### Session vs Persistent Allow Lists

Session-scoped allow lists clear on exit, preventing accidental persistence of one-time permissions. Only explicit "Always" selections persist to configuration.

## Future Enhancements

### Per-Tool Custom Rules

Support custom permission rules beyond binary allow/deny:
- Prompt with pre-filled approval for certain args
- Conditional allows based on context
- Time-limited permissions

### Real-Time Policy Updates

Support reloading allow/deny lists without restart:
- Watch `harness.toml` for changes
- Hot-reload gate configuration
- Notify user of policy changes

### Audit Logging

Log all permission decisions for compliance:
- Timestamp of request
- Tool name and arguments
- User decision and scope
- Current mode

### Tool Usage Analytics

Track which tools require most permissions:
- Identify friction points
- Suggest allow list additions
- Optimize default permissions

## Conclusion

The three-mode permission system gives users fine-grained control over agent autonomy while maintaining safety by default. The two-layer gate architecture provides defense-in-depth, and the modal UI makes permission decisions transparent and contextual.
