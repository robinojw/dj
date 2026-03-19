# Agent Pool Design — Dynamic Persona Swarm

## Overview

DJ evolves from a single-process Codex visualizer into a multi-process agent swarm orchestrator. Roster generates persona definitions from repo analysis. DJ reads those personas, spawns multiple Codex processes (each persona-typed), and visualizes the full swarm on a canvas grid. A designated orchestrator Codex session routes tasks to personas, while any agent can request specialist spawns mid-session.

## Decisions

- **Control model**: Hybrid — DJ manages process lifecycle, an orchestrator Codex session handles task routing
- **Spawn model**: Tiered — a few top-level Codex processes for parallelism, each can spawn internal sub-agents
- **Roster integration**: Read output files (`.roster/personas/*.md`, `.roster/signals.json`) — no compile-time dependency
- **Orchestrator communication**: Output-stream parsing — orchestrator emits `dj-command` fenced code blocks that DJ parses from delta events
- **Scope**: Full dynamic swarm — orchestrator routing, multi-process pool, dynamic mid-session spawning, cross-agent communication

## Section 1: Core Data Model

### PersonaDefinition

Loaded from `.roster/personas/*.md` at startup. Parsed from YAML frontmatter.

```go
type PersonaDefinition struct {
    ID          string
    Name        string
    Description string
    Triggers    []string
    Content     string   // full markdown body — becomes the system prompt
}
```

### AgentProcess

Represents a running Codex process with a persona identity.

```go
type AgentProcess struct {
    ID        string              // unique agent ID (e.g. "architect-1")
    PersonaID string              // links to PersonaDefinition.ID
    ThreadID  string              // Codex thread ID from thread/started
    Client    *appserver.Client   // the Codex process
    Role      AgentRole           // orchestrator | worker
    Task      string              // the task assigned to this agent
    Status    string              // spawning | active | completed | error
    ParentID  string              // which agent requested this spawn
}
```

### AgentRole

Two types of agents in the swarm:

- **Orchestrator**: Has all persona definitions in its system prompt. Analyzes tasks, emits spawn commands. Exactly one per DJ instance.
- **Worker**: Has a single persona's system prompt. Does the actual work. Can request specialist spawns via `dj-command` blocks.

### ThreadState changes

The existing `ThreadState` gains a new field linking it to the agent pool:

```go
AgentProcessID string  // links to AgentProcess.ID
```

This lets the canvas show persona-specific styling per card — color, icon, role label all driven by the persona definition.

## Section 2: AgentPool — Process Manager

New `internal/pool/` package. Manages the lifecycle of multiple Codex processes.

### Responsibilities

1. **Spawn** — Start a new Codex process with a persona-specific system prompt
2. **Route events** — Each process has its own `ReadLoop`. The pool multiplexes all event streams into a single channel that the TUI consumes
3. **Stop** — Gracefully shut down individual agents or the entire pool
4. **Lookup** — Find agents by ID, persona, or thread ID

### Core structure

```go
type AgentPool struct {
    agents    map[string]*AgentProcess   // keyed by agent ID
    mu        sync.RWMutex
    events    chan PoolEvent              // multiplexed events from all processes
    command   string                     // base codex command
    args      []string                   // base codex args
    personas  map[string]PersonaDefinition
    idCounter atomic.Int64
}
```

### PoolEvent

Wraps a JSON-RPC message with the source agent ID so the TUI knows which process it came from:

```go
type PoolEvent struct {
    AgentID string
    Message appserver.JSONRPCMessage
}
```

### Spawn flow

```
pool.Spawn(personaID, task, parentAgentID)
    -> create AgentProcess
    -> appserver.NewClient(command, args...)
    -> client.Start(ctx)
    -> client.Initialize() with persona system prompt injected
    -> client.SendUserInput(task)
    -> goroutine: client.ReadLoop -> wrap events as PoolEvent -> pool.events channel
```

The persona's `Content` (full markdown body) is sent as the system prompt during the `initialize` handshake. This is how the Codex process becomes persona-typed.

### Integration with AppModel

Today `AppModel` has a single `client *appserver.Client`. This becomes:

```go
pool        *pool.AgentPool     // replaces client
poolEvents  chan pool.PoolEvent  // replaces events chan
```

The `listenForEvents()` method reads from `pool.Events()` instead of a single client's event channel. The `PoolEvent.AgentID` maps each event to the right agent process, which maps to the right `ThreadState` via `AgentProcessID`.

### Orchestrator bootstrap

On startup, the pool auto-spawns one orchestrator agent:

```go
pool.SpawnOrchestrator(allPersonas, repoSignals)
```

This creates an agent whose system prompt contains all persona definitions, their triggers, and the repo signals. No task yet — it waits for the user's first input.

## Section 3: Orchestrator Bridge — Command Parsing

The orchestrator emits structured commands in its response stream. DJ detects and parses them from chunked `item/agentMessage/delta` events.

### Command protocol

The orchestrator (and any worker agent requesting a specialist) emits fenced code blocks with a `dj-command` language tag:

````
```dj-command
{"action": "spawn", "persona": "architect", "task": "Design the API boundary for the auth module"}
```
````

### Supported commands

```go
type DJCommand struct {
    Action  string `json:"action"`   // spawn | message | complete
    Persona string `json:"persona"`  // persona ID to spawn (for spawn)
    Task    string `json:"task"`     // task description (for spawn)
    Target  string `json:"target"`   // target agent ID (for message)
    Content string `json:"content"`  // message content (for message)
}
```

Three actions:

- **spawn** — Request DJ to spawn a new persona agent with a task
- **message** — Send a message from one agent to another (cross-agent communication via DJ as hub)
- **complete** — Agent signals it's done with its task. DJ updates status and notifies the orchestrator

### Delta buffer and parser

New `internal/orchestrator/` package:

```go
type CommandParser struct {
    buffer    strings.Builder
    inBlock   bool           // currently inside a dj-command fence
    commands  chan DJCommand  // parsed commands emitted here
}

func (parser *CommandParser) Feed(delta string)
func (parser *CommandParser) Commands() <-chan DJCommand
```

Parsing logic:

1. Append each delta to the buffer
2. Scan for opening fence `` ```dj-command ``
3. When found, set `inBlock = true`, start accumulating command content
4. Scan for closing fence `` ``` ``
5. When found, parse the accumulated JSON, emit `DJCommand` on the channel
6. Strip the command block from the text that gets displayed (internal commands should not pollute the UI)

### Per-agent parsers

Each `AgentProcess` gets its own `CommandParser`. When the pool receives an `item/agentMessage/delta` event, it feeds the delta to the corresponding parser. Commands flow from parser to pool to TUI as Bubble Tea messages:

```go
type SpawnRequestMsg struct {
    SourceAgentID string
    Persona       string
    Task          string
}
```

The TUI handles `SpawnRequestMsg` by calling `pool.Spawn(persona, task, sourceAgentID)`.

## Section 4: Persona Loader — Roster Integration

New `internal/roster/` package. Reads `.roster/` output files at startup — no compile-time dependency on roster.

### What DJ reads

1. **`.roster/signals.json`** — Repo signals (languages, frameworks, CI, etc.). Passed to the orchestrator's system prompt for context-aware routing.
2. **`.roster/personas/*.md`** — Persona templates with YAML frontmatter. Each becomes a `PersonaDefinition`.

### Loader

```go
package roster

func LoadPersonas(dir string) ([]PersonaDefinition, error)
func LoadSignals(path string) (*RepoSignals, error)
```

`LoadPersonas` walks `.roster/personas/`, splits YAML frontmatter from markdown body (`---` delimiters), unmarshals into `PersonaDefinition`. The `Content` field holds the full markdown body that becomes the agent's system prompt.

`LoadSignals` reads and unmarshals the JSON file. DJ defines its own `RepoSignals` struct — mirrors roster's but decoupled.

### Startup flow

```
1. config.Load()
2. roster.LoadPersonas(".roster/personas/")
3. roster.LoadSignals(".roster/signals.json")
4. pool.NewAgentPool(config, personas)
5. pool.SpawnOrchestrator(personas, signals)
6. tui.NewAppModel(store, pool)
```

### Graceful degradation

If `.roster/` does not exist (roster was not run), DJ falls back to current behavior — single Codex process, no personas, no orchestrator. The pool is optional. This keeps DJ usable standalone.

### Config additions

New section in `dj.toml`:

```toml
[roster]
path = ".roster"          # where to find roster output
auto_orchestrate = true   # spawn orchestrator on startup

[pool]
max_agents = 10           # maximum concurrent agent processes
```

## Section 5: Cross-Agent Communication & Dynamic Spawning

### DJ as message bus

Agents cannot talk directly — they are separate Codex processes with isolated stdio pipes. DJ is the router. When an agent emits a `message` command:

````
```dj-command
{"action": "message", "target": "architect-1", "content": "The auth module needs a rate limiter."}
```
````

DJ receives this as a `DJCommand`, looks up the target agent in the pool, and injects the message into that agent's Codex process via `client.SendUserInput()`. The target agent sees it as a new turn with sender context.

### Message injection format

When DJ delivers a cross-agent message, it wraps it with sender context:

```
[From: test-1 (Test Engineer)] The auth module needs a rate limiter. What's your recommended pattern?
```

The receiving agent knows who is talking and can respond via its own `dj-command` message block.

### Dynamic mid-session spawning

Any worker agent can request a specialist. The flow:

1. Agent outputs a `spawn` dj-command
2. DJ parses it, creates `SpawnRequestMsg` with `SourceAgentID`
3. Pool spawns a new persona agent with `ParentID` set to the requesting agent
4. Canvas renders it as a child of the requesting agent's card
5. When the spawned agent completes, it emits a `complete` command
6. DJ notifies the parent agent that the specialist finished (via message injection)

### Completion flow

When an agent finishes its task:

````
```dj-command
{"action": "complete", "content": "Security review complete. Found 2 issues: [details]"}
```
````

DJ handles this by:

1. Updating the agent's status to `completed`
2. If the agent has a parent, injecting the completion content into the parent's process
3. Notifying the orchestrator so it can decide next steps

### Orchestrator awareness

The orchestrator stays informed of the full swarm state. When agents spawn, complete, or fail, DJ injects status updates into the orchestrator's process:

```
[DJ System] Agent security-1 (Security) completed task: "Review auth API"
Result: Found 2 issues: [details]
Active agents: architect-1 (active), test-1 (active)
```

### Ordering and concurrency

- Message delivery is async — DJ queues messages per agent and delivers them between turns (not mid-turn)
- Spawn requests are serialized through the pool's mutex
- Completion notifications are delivered after the agent's process stops producing output

## Section 6: Canvas & UI Changes

### Persona-aware cards

Cards gain:

- **Persona badge** — Short label from `PersonaDefinition.Name`
- **Persona color** — Each persona ID maps to a distinct color (architect=blue, test=green, security=red, reviewer=yellow, performance=cyan, design=magenta, devops=orange)
- **Role subtitle** — Shows the assigned task (truncated to card width)
- **Orchestrator indicator** — Double-line or bold border to distinguish from workers

### Canvas layout modes

- **Grid** (existing) — Works for small swarms (< 8 agents)
- **Tree** (existing) — Left sidebar hierarchy
- **Swarm view** (new) — Orchestrator centered at top, workers in a row below, children below that. Org chart layout. Toggled with `s` key.

### Header bar

- **Active persona count** — "4 agents" with breakdown
- **Swarm status** — "Orchestrating..." / "3 active, 1 completed"

### New keybindings

| Key | Action |
|-----|--------|
| `n` | Submit task to orchestrator (prompts for task input) |
| `p` | Manual persona picker — spawn specific persona without orchestrator |
| `m` | Send message to selected agent (prompts for text) |
| `s` | Toggle swarm view layout |
| `K` | Kill selected agent (stop its Codex process) |

### Session panel

When viewing an agent's session output, `dj-command` blocks are stripped. Users see reasoning and work output, not internal plumbing.

## Section 7: End-to-End Flow

### Example: "Add user authentication with JWT"

```
1. User starts DJ
   -> Load config, personas, signals
   -> Spawn orchestrator (system prompt has all personas + repo signals)
   -> Canvas: one card "Orchestrator" (bold border, idle)

2. User presses 'n', types task
   -> DJ sends to orchestrator via SendUserInput()
   -> Orchestrator analyzes, outputs spawn commands:
      spawn architect: "Design JWT auth module structure"
      spawn security: "Define token expiry and secret management"

3. DJ parses commands, spawns two Codex processes
   -> Canvas: orchestrator top, architect + security below with connectors

4. Architect works, needs test coverage
   -> Outputs spawn command for test persona
   -> DJ spawns test agent as child of architect
   -> Canvas: test card under architect

5. Security agent completes
   -> Result injected into orchestrator and architect

6. All agents complete
   -> Orchestrator synthesizes final summary
   -> All cards show completed (blue)

7. User clicks any card to see full session output
```

### Error handling

- **Agent crash**: Pool detects via `client.Running()`, sets status to error, notifies orchestrator. Orchestrator can retry.
- **Orchestrator crash**: DJ detects, shows error in status bar, auto-respawns with fresh context. Workers keep running.
- **Malformed dj-command**: Parser logs error, skips command, continues processing. Does not crash.
- **Unknown persona ID**: Pool rejects, notifies orchestrator with available persona list.
- **Runaway spawning**: `max_agents` config limit. Pool rejects when at capacity.

### Graceful shutdown

When user quits (`q` or `Ctrl+C`):

1. Stop all worker agents (close stdin, wait for exit)
2. Stop orchestrator last
3. Clean up PTY sessions

## New Packages

| Package | Purpose |
|---------|---------|
| `internal/pool/` | AgentPool, AgentProcess, PoolEvent — multi-process management |
| `internal/orchestrator/` | CommandParser, DJCommand — delta stream parsing |
| `internal/roster/` | PersonaDefinition, RepoSignals, LoadPersonas, LoadSignals |

## Modified Packages

| Package | Changes |
|---------|---------|
| `internal/tui/` | AppModel gains pool, swarm view layout, persona-aware cards, new keybindings |
| `internal/state/` | ThreadState gains AgentProcessID field |
| `internal/config/` | New roster and pool config sections |
| `cmd/dj/` | Startup flow adds persona loading and pool initialization |
