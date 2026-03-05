# DJ

**Orchestrate AI agents in your terminal.**

DJ is a TUI that coordinates multiple AI agents to tackle complex coding tasks. When your prompt needs more than one agent, DJ spawns a team view where specialized workers collaborate—each with their own context, reasoning, and tool access.

## Features

### 🤝 Multi-Agent Orchestration

DJ spawns worker agents when tasks require parallel effort. Watch them collaborate in real-time through the Team View (Ctrl+T), with live updates showing each agent's progress, tool calls, and reasoning.

### 🎯 Three Permission Modes

Control agent autonomy with Tab-switchable modes:

- **Confirm** (⏸): Review destructive operations before execution
- **Plan** (◎): Read-only mode with high reasoning effort for architecture planning
- **Turbo** (⚡): Full autonomy for safe/isolated environments

Customize which tools auto-approve or block via `harness.toml`.

### 🔧 Built-in Skills

Pre-packaged workflows guide agents through common tasks:

- **enhance-prompt**: Clarify and expand vague prompts
- **explain-code**: Generate detailed code explanations
- **refactor**: Restructure code while preserving behavior
- **write-tests**: Create comprehensive test suites

Browse and invoke skills via Ctrl+K.

### 🔌 MCP Server Integration

Connect Model Context Protocol servers for extended capabilities:

- **GitHub**: Issues, PRs, code search
- **Filesystem**: Read/write operations with permission gates
- **Custom servers**: Wire your own via stdio or HTTP

Manage servers through the MCP Manager (Ctrl+M).

### 🎨 Adaptive Themes

Five built-in themes adapt to your terminal:

- Tokyo Night (default)
- Catppuccin
- Gruvbox
- Nord
- One Dark

Themes support custom JSON overrides.

### ⚡ Advanced Features

- **LSP Integration**: Auto-detect and connect language servers for diagnostics and completion
- **Event Hooks**: Trigger shell commands on tool calls, errors, or session events
- **Checkpointing**: Revert state with Ctrl+Z
- **Cost Tracking**: Monitor token usage and API costs
- **Streaming**: Real-time response rendering

## Installation

### Homebrew (Recommended)

```bash
brew install robinojw/dj
```

### Go Install

```bash
go install github.com/robinojw/dj/cmd/harness@latest
```

### Pre-built Binaries

Download from [releases](https://github.com/robinojw/dj/releases) for Linux, macOS, or Windows.

## Quick Start

1. **Set your OpenAI API key:**

```bash
export OPENAI_API_KEY="sk-..."
```

2. **Launch DJ:**

```bash
dj
```

3. **Try a multi-step task:**

```
Refactor the authentication module to use JWT tokens, update tests, and document the changes
```

Watch DJ spawn worker agents for refactoring, testing, and documentation.

## Key Bindings

| Key | Action |
|-----|--------|
| **Tab** | Cycle permission modes (Confirm → Plan → Turbo) |
| **Ctrl+T** | Open Team View (multi-agent orchestration) |
| **Ctrl+K** | Browse and invoke Skills |
| **Ctrl+M** | Manage MCP servers |
| **Ctrl+E** | Enhance prompt with AI suggestions |
| **Ctrl+Z** | Undo last checkpoint |
| **Ctrl+Q** | Quit |
| **Esc** | Return to main chat |

## Configuration

DJ reads from two locations:

1. **Project config**: `harness.toml` (version-controlled settings)
2. **User config**: `~/.config/codex-harness/config.toml` (personal overrides)

### Example Configuration

```toml
[model]
default = "gpt-5.1-codex-mini"
reasoning_effort = "medium"
team_threshold = 3  # Spawn team view after 3 subtasks

[theme]
name = "tokyonight"  # or path to custom .json

[execution]
default_mode = "confirm"  # confirm | plan | turbo

[execution.allow]
# Auto-approved tools (glob patterns supported)
tools = [
    "read_file",
    "bash(git status*)",
    "bash(git diff*)",
]

[execution.deny]
# Blocked tools (security floor, applies to all modes)
tools = [
    "bash(rm -rf*)",
    "read_file(.env*)",
]

[mcp.servers]
  # Example: Connect to external APIs via MCP
  # Note: For filesystem and GitHub, use native Go operations and gh CLI
  # MCP is best for remote services and specialized integrations

[skills]
paths = [
  "./skills",
  "~/.config/codex-harness/skills",
]

[hooks]
pre_tool_call = "echo 'Tool: $TOOL_NAME'"
post_tool_call = "notify-send 'Tool completed'"
on_error = "logger 'DJ error: $ERROR_MSG'"
on_session_end = "cleanup.sh"
```

### Permission System

The three-mode system balances speed and safety:

**Confirm Mode** (default): Prompts before write/execute operations. Shows permission modal with "Remember this decision" options:

- **Once**: Allow single invocation
- **Session**: Allow for current session
- **Always**: Persist to `harness.toml`

**Plan Mode**: Read-only with high reasoning. Useful for architectural planning without execution risk.

**Turbo Mode**: Bypasses all prompts. Requires confirmation on first activation. Deny list still applies.

See [docs/user-guide-permissions.md](docs/user-guide-permissions.md) for details.

### MCP Servers

Connect external tools via Model Context Protocol.

**Note:** For filesystem operations and GitHub access, `dj` uses native Go operations and the `gh` CLI directly—no MCP servers needed. MCP is best for remote services and specialized integrations.

**Example - HTTP server** (remote API):
```toml
[mcp.servers.stripe]
type = "http"
url = "https://mcp.stripe.com"
headers = { Authorization = "Bearer ${STRIPE_MCP_KEY}" }
auto_start = false
```

DJ auto-discovers servers from `~/.config/claude/mcp.json` and `~/.config/codex-harness/mcp.json`.

### Custom Skills

Create skills in `skills/<skill-name>/SKILL.md`:

```markdown
---
name: my-skill
description: Brief description shown in browser
trigger: /myskill
---

# Skill prompt content here

Instructions for the agent...
```

Skills appear in the browser (Ctrl+K) and can be invoked via `/myskill` in chat.

### Themes

Customize colors in `themes/<name>.json`:

```json
{
  "name": "custom",
  "background": "#1a1b26",
  "foreground": "#c0caf5",
  "primary": "#7aa2f7",
  "secondary": "#9d7cd8",
  "accent": "#bb9af7",
  "muted": "#565f89",
  "border": "#3b4261"
}
```

Load via `theme.name = "custom"` in config.

## How It Works

1. **Single-agent mode**: Simple queries execute in the main chat with streaming responses
2. **Multi-agent mode**: Complex tasks trigger orchestrator logic:
   - Orchestrator analyzes prompt and plans subtasks
   - Spawns worker agents for parallel execution
   - Aggregates results and coordinates dependencies
   - Compacts conversation history to manage context
3. **Permission gates**: All tool calls pass through two-layer filtering (allow/deny lists + mode-based rules)
4. **Checkpointing**: State snapshots enable undo via Ctrl+Z (20-checkpoint ring buffer)

## Architecture

```
┌─────────────────────────────────────────┐
│  TUI (Bubble Tea)                       │
│  ┌────────┬────────┬────────┬────────┐  │
│  │ Chat   │ Team   │ Skills │  MCP   │  │
│  └────────┴────────┴────────┴────────┘  │
└─────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────┐
│  Orchestrator                           │
│  - Task routing                         │
│  - Worker spawning                      │
│  - Result aggregation                   │
└─────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────┐
│  Permission Gate                        │
│  - Mode enforcement                     │
│  - Allow/deny lists                     │
│  - Glob pattern matching                │
└─────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────┐
│  Workers (parallel agents)              │
│  - Independent context                  │
│  - Tool execution                       │
│  - Streaming responses                  │
└─────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────┐
│  Integrations                           │
│  - MCP servers (stdio/HTTP)             │
│  - LSP clients (diagnostics)            │
│  - Skills (loaded from disk)            │
│  - Hooks (shell commands)               │
└─────────────────────────────────────────┘
```

## Troubleshooting

**"OPENAI_API_KEY environment variable is required"**

Export your API key:
```bash
export OPENAI_API_KEY="sk-..."
```

Add to `~/.bashrc` or `~/.zshrc` for persistence.

**MCP server fails to start**

Check server installation:
```bash
# Example for Node.js-based MCP servers
npx @modelcontextprotocol/server-example --version
```

Set `auto_start = false` and start manually via MCP Manager (Ctrl+M).

**Note:** If you're seeing slow startup times, avoid using MCP servers for capabilities that `dj` can handle natively (filesystem, GitHub).

**Theme not loading**

Verify theme file exists:
```bash
ls themes/tokyonight.json
```

Or use absolute path in config:
```toml
[theme]
name = "/path/to/custom.json"
```

**Permission modal not showing**

Ensure you're in Confirm mode (status bar shows `⏸ CONFIRM`). Press Tab to cycle modes.

## Development

```bash
# Clone
git clone https://github.com/robinojw/dj.git
cd dj

# Build
go build -o dj ./cmd/harness

# Run
./dj

# Test
go test ./...

# Install locally
go install ./cmd/harness
```

## Contributing

Contributions welcome! Areas of interest:

- Additional built-in skills
- MCP server integrations
- Theme designs
- LSP server support for more languages
- Documentation improvements

Open an issue or PR on [GitHub](https://github.com/robinojw/dj).

## License

[MIT](LICENSE)

## Acknowledgments

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [MCP](https://modelcontextprotocol.io) - Tool integration protocol
- OpenAI Codex - Language models
