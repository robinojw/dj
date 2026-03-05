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
