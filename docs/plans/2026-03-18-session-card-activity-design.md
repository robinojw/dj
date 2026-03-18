# Session Card Activity Indicators

## Problem

Session cards show only a static status word (idle/active/completed/error). Users cannot tell what the CLI is actually doing — whether it's thinking, running a command, applying a patch, or streaming a response.

## Design

An `Activity` field on `ThreadState` tracks the current action for each session. The card renders activity instead of the status word when present. When there is no activity, the card falls back to the existing status display.

## Activity Mapping

| Protocol Event | Activity Text | Behavior |
|---|---|---|
| `task_started` | `Thinking...` | Set on task start |
| `agent_reasoning_delta` | `Thinking...` | Set during reasoning |
| `agent_message_delta` | Truncated snippet of streaming text | Updated with each delta |
| `exec_command_request` | `Running: <command>` | Truncated to card width |
| `patch_apply_request` | `Applying patch...` | Set on patch request |
| `agent_message` (completed) | Clears activity | Falls back to status |
| `task_complete` | Clears activity | Falls back to status |

## Card Rendering

Activity replaces the status line when present:

```
Active with activity:
┌──────────────────────────────┐
│ o3-mini                      │
│ Running: git status          │
└──────────────────────────────┘

Idle (no activity):
┌──────────────────────────────┐
│ o3-mini                      │
│ idle                         │
└──────────────────────────────┘
```

Activity text uses the same color as the current status.

## Changes

- `state/thread.go`: Add `Activity` field, `SetActivity()`, `ClearActivity()` methods
- `state/store.go`: Add `UpdateActivity(id, activity)` method (thread-safe)
- `tui/card.go`: Prefer `thread.Activity` over `thread.Status` for card second line
- `tui/app_proto.go`: Set activity in event handlers
- `tui/bridge.go`: Handle `agent_reasoning_delta` events
- `tui/msgs.go`: Add `AgentReasoningDeltaMsg` type
