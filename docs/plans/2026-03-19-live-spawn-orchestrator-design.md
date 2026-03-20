# Live Process Spawning & Orchestrator Bootstrap Design

## Overview

Close the gap between DJ's agent pool bookkeeping and actual Codex process management. `pool.Spawn()` becomes a live operation that starts a real `codex proto` child process, wires its JSON-RPC event stream into the pool, and auto-approves execution requests. A new `SpawnOrchestrator()` method bootstraps a coordinator agent on startup that receives user tasks and emits `dj-command` spawn blocks to fan out work across persona-typed workers.

## Decisions

- **Protocol**: JSON-RPC via `codex proto` (app-server mode) — matches existing architecture
- **Persona injection**: First turn message contains persona instructions + task
- **Approval policy**: Auto-approve all exec/file requests — agents run fully autonomously
- **Orchestrator**: Auto-spawns on startup when `auto_orchestrate = true` in config

## Section 1: Client Initialize Method

Add `Initialize()` to `appserver.Client`. Sends the JSON-RPC handshake Codex expects on connection:

```json
{"jsonrpc":"2.0","id":"dj-1","method":"initialize","params":{"clientInfo":{"name":"dj","version":"0.1.0"}}}
```

Fire-and-forget — the ReadLoop handles the response. No blocking wait needed.

**File**: `internal/appserver/client_send.go`

## Section 2: Live Process Spawning

`pool.Spawn()` gains a real process lifecycle. After creating the `AgentProcess` struct:

1. `client := appserver.NewClient(pool.command, pool.args...)`
2. `client.Start(ctx)` — spawns `codex proto` as a child process
3. `client.Initialize()` — sends the handshake
4. Start a goroutine running `client.ReadLoop(handler)` where the handler:
   - Wraps each message as `PoolEvent{AgentID, Message}`
   - Pushes to `pool.events` channel
   - Detects `MethodExecApproval` / `MethodFileApproval` and auto-responds with `client.SendApproval(requestID, true)` before forwarding
5. `client.SendUserInput(buildWorkerPrompt(persona, task))` — first turn
6. Set `agent.Status = AgentStatusActive`, store the client on the agent

The pool stores a `context.Context` (created during `NewAgentPool`) so `StopAll` can cancel all child processes.

**Files**: `internal/pool/pool.go`, `internal/pool/spawn.go` (new, extracted from pool.go for clarity)

## Section 3: Orchestrator Bootstrap

New method `SpawnOrchestrator(signals *roster.RepoSignals)` on `AgentPool`:

- Creates an agent with `Role=RoleOrchestrator`, no persona ID
- First turn message is the orchestrator prompt (Section 4)
- No task — orchestrator idles until the user submits one
- Called from `main.go` during startup when `auto_orchestrate = true`

When the user presses `n` and types a task, the TUI sends it to the orchestrator via `client.SendUserInput()`. The orchestrator analyzes and emits `dj-command` spawn blocks. The existing `CommandParser` + `handlePoolEvent` pipeline processes those.

**Files**: `internal/pool/pool.go`, `internal/pool/orchestrator.go` (new)

## Section 4: Prompt Templates

New `internal/pool/prompts.go` file with prompt construction functions.

**Worker prompt** (first turn message):

```
You are acting as the {Name} specialist.

{persona.Content}

Your task: {task}
```

**Orchestrator prompt** (first turn message):

```
You are DJ's orchestrator. You coordinate a team of specialist agents to accomplish tasks.

Available personas:
- {id}: {description}
[...one line per persona]

Repo context:
Languages: {signals.Languages}
CI: {signals.CIProvider}
Lint: {signals.LintConfig}

To spawn an agent, emit a fenced code block:
```dj-command
{"action":"spawn","persona":"architect","task":"Design the auth module"}
```

To message an existing agent:
```dj-command
{"action":"message","target":"architect-1","content":"Please add rate limiting"}
```

When done coordinating, emit:
```dj-command
{"action":"complete","content":"Summary of results"}
```

Analyze the user's request, decide which specialists to spawn, and coordinate their work.
```

## Section 5: Auto-Approval

In the ReadLoop handler for each agent, before forwarding events:

```go
if isApprovalRequest(message) {
    client.SendApproval(message.ID, true)
}
pool.events <- PoolEvent{AgentID: agentID, Message: message}
```

The TUI still receives approval events for display (showing what commands were run, what files were changed). Agents never block waiting for human approval.

## Section 6: Process Lifecycle and Cleanup

- ReadLoop goroutine detects process exit when `scanner.Scan()` returns false
- Sends a synthetic completion event so the TUI updates the card status
- Pool marks agent as completed
- `StopAgent()` calls `client.Stop()` on the specific agent (already partially implemented)
- `StopAll()` iterates and stops all agents (already implemented, just needs client.Stop() calls)
- Graceful shutdown in `main.go` calls `pool.StopAll()` before exit

## Section 7: Startup Flow

Updated `main.go` flow:

```
1. config.Load()
2. roster.LoadPersonas() + LoadSignals()
3. pool.NewAgentPool(ctx, command, args, personas, maxAgents)
4. if auto_orchestrate && len(personas) > 0:
       pool.SpawnOrchestrator(signals)
5. tui.NewAppModel(store, WithPool(pool))
6. program.Run()
7. pool.StopAll()  // graceful shutdown
```

## Section 8: TUI Integration Changes

- `handleThreadCreated` for pool mode: when orchestrator is spawned, add it to the store so it appears on the canvas
- User presses `n` in pool mode → task sent to orchestrator (not creating a local thread)
- Manual `p` key still works for direct persona spawning (bypasses orchestrator)

## New/Modified Files

| File | Change |
|------|--------|
| `internal/appserver/client_send.go` | Add `Initialize()` method |
| `internal/pool/pool.go` | Add context, update `Spawn()` with live process, add `SpawnOrchestrator()` |
| `internal/pool/spawn.go` | New — extracted spawn logic with ReadLoop wiring |
| `internal/pool/orchestrator.go` | New — orchestrator bootstrap and prompt |
| `internal/pool/prompts.go` | New — worker and orchestrator prompt templates |
| `cmd/dj/main.go` | Call `SpawnOrchestrator()` on startup, pass context |
