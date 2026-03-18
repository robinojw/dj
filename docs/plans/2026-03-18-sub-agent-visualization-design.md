# Sub-Agent Visualization Design

## Goal

Visualize Codex sub-agent hierarchies on DJ's canvas grid by migrating to the v2 app-server protocol and rendering parent-child connectors between cards.

## Context

Codex CLI (v0.42+) spawns sub-agents that appear as separate threads linked by `parent_thread_id`. The v2 app-server protocol emits 10 collaboration events covering the full sub-agent lifecycle. DJ currently uses the legacy protocol format and silently drops unknown events, missing all collaboration data.

## Design

### 1. Protocol Layer Migration

Replace the legacy `{id, msg: {type}}` envelope with JSON-RPC 2.0 `{method, params}`.

**Envelope change in `appserver/protocol.go`:**

The current `ProtoEvent` and `EventHeader` types are replaced with a `JsonRpcMessage` that handles notifications (no `id`), requests (`id` + `method`), and responses (`id` + `result`/`error`).

**Method constants in `appserver/methods.go`:**

| Legacy                    | V2                                         |
|---------------------------|--------------------------------------------|
| `session_configured`      | `thread/started`                           |
| `task_started`            | `turn/started`                             |
| `task_complete`           | `turn/completed`                           |
| `agent_message_delta`     | `item/agentMessage/delta`                  |
| `agent_message`           | `item/completed` (agent message item)      |
| `exec_command_request`    | `item/commandExecution/requestApproval`    |
| `patch_apply_request`     | `item/fileChange/requestApproval`          |

New methods for collaboration:
- `thread/started` (with `SubAgent` source)
- `thread/status/changed`
- `item/started` / `item/completed` (for `CollabAgentToolCall` items)

**Client ReadLoop (`client.go`):**

Parse the JSON-RPC envelope and route by `method` field.

**New types in `appserver/types_collab.go`:**

Types for 10 collaboration events:
- `CollabAgentSpawnBeginEvent` / `CollabAgentSpawnEndEvent` — carries `sender_thread_id`, `new_thread_id`, `agent_nickname`, `agent_role`, `depth`
- `CollabAgentInteractionBeginEvent` / `CollabAgentInteractionEndEvent` — carries `sender_thread_id`, `receiver_thread_id`, `prompt`
- `CollabWaitingBeginEvent` / `CollabWaitingEndEvent` — carries `sender_thread_id`, `receiver_thread_ids`
- `CollabCloseBeginEvent` / `CollabCloseEndEvent` — carries `sender_thread_id`, `receiver_thread_id`
- `CollabResumeBeginEvent` / `CollabResumeEndEvent` — carries `sender_thread_id`, `receiver_thread_id`

Supporting types: `SubAgentSource`, `SessionSource`, `AgentStatus`, `CollabAgentTool`, `CollabAgentToolCallStatus`.

### 2. State Layer Extensions

**`ThreadState` new fields:**

```
AgentNickname string   // from thread.agent_nickname
AgentRole     string   // from thread.agent_role
Depth         int      // nesting level (0 = root)
Model         string   // model used by this thread
```

**Parent-child wiring:**

When `thread/started` arrives with `source: SubAgent(ThreadSpawn{parent_thread_id, depth, ...})`, call `store.AddWithParent()` and populate the new fields.

**Status mapping:**

| AgentStatus   | DJ Status   |
|---------------|-------------|
| PendingInit   | idle        |
| Running       | active      |
| Interrupted   | idle        |
| Completed     | completed   |
| Errored       | error       |
| Shutdown      | completed   |

**Tree ordering:**

New `store.TreeOrder()` method returns threads in depth-first order (roots first, then children recursively). Same traversal as `TreeModel.rebuild()`.

### 3. Canvas Edge Rendering

**Layout with connectors:**

The grid renders threads in tree order. Between grid rows that have parent-child relationships, a connector row is inserted using box-drawing characters.

```
+------------+  +------------+  +------------+
| Main       |  | Other      |  | Another    |
+-----+------+  +------------+  +------------+
      |
      +---------------------------+
      |                           |
+-----+------+  +------------+  +-+----------+
| Sub-1      |  |            |  | Sub-2      |
+------------+  +------------+  +------------+
```

**Implementation in `canvas_edges.go`:**

`renderConnectorRow(parentPositions, childPositions, cardWidth, columnGap) string`:
1. Find horizontal center of each card by column
2. For each parent-child pair, draw `|` down from parent center
3. Draw horizontal `─` to each child center
4. Use `┬` at parent, `├`/`┤` at branches, `┴` at children

**Edge styling:**
- Color by parent status (green if active, gray if idle)
- Dim edges to completed/errored children

### 4. Card Enhancements

Sub-agent cards display:
- `↳` prefix on title to indicate child status
- Agent role as subtitle line
- Same status color coding as root cards

```
+----------------+
| ↳ Sub-Agent    |
|   researcher   |
|   active       |
+----------------+
```

### 5. Multi-Thread Protocol Routing

Eliminate global `sessionID`. Every Bubble Tea message carries `ThreadID` from the v2 protocol's per-notification `thread_id` field.

App handlers look up the correct `ThreadState` by ID:

```
handleTurnStarted(msg):
    store.UpdateStatus(msg.ThreadID, active, "")

handleAgentDelta(msg):
    thread = store.Get(msg.ThreadID)
    thread.AppendDelta(msg.MessageID, msg.Delta)
```

### 6. Bridge Routing

`tui/bridge.go` switches on v2 `method` strings instead of legacy `type` strings. Decode functions extract `thread_id` and event-specific data into typed Bubble Tea messages.

## Out of Scope

- DAG/pipeline visualization (parent-child tree only)
- Drag-and-drop card rearrangement
- Custom edge styling beyond status colors
- Collaboration event replay/history
- Manual thread linking UI

## Files Touched

- `internal/appserver/protocol.go` — JSON-RPC envelope types
- `internal/appserver/methods.go` — v2 method constants
- `internal/appserver/types_thread.go` — v2 thread/turn types
- `internal/appserver/types_collab.go` — new collaboration types
- `internal/appserver/client.go` — ReadLoop v2 parsing
- `internal/state/thread.go` — new ThreadState fields
- `internal/state/store.go` — TreeOrder() method
- `internal/tui/bridge.go` — v2 method routing
- `internal/tui/messages.go` — new Bubble Tea messages
- `internal/tui/canvas.go` — tree-ordered rendering
- `internal/tui/canvas_edges.go` — new connector rendering
- `internal/tui/card.go` — sub-agent display enhancements
- `internal/tui/app.go` — multi-thread routing
- `internal/tui/app_proto.go` — new event handlers
- Tests for all above
