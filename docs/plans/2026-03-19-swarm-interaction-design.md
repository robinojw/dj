# Swarm Interaction UI Design

## Goal

Wire up the stub keybindings (p, m, s, K) so users can spawn persona agents, message them, filter the canvas to swarm-only view, and see swarm hints in the header bar.

## Architecture

Reuse the existing `MenuModel` for persona and agent pickers. Add a new lightweight `InputBar` component for task/message text entry. The flow is two-step: pick from a list, then type in the input bar. No new dependencies.

## Components

### InputBarModel

A single-line text input that renders in place of the status bar.

- `prompt string` — label like "Task for Architect: "
- `value string` — user-typed text
- `View()` renders prompt + value + cursor
- No `Update()` — the app feeds keystrokes directly

### New AppModel Fields

- `inputBar InputBarModel`
- `inputBarVisible bool`
- `inputBarIntent InputIntent` — enum: `IntentSpawnTask`, `IntentSendMessage`
- `menuIntent MenuIntent` — enum: `MenuIntentThread`, `MenuIntentPersonaPicker`, `MenuIntentAgentPicker`
- `pendingPersonaID string`
- `pendingTargetAgentID string`
- `swarmFilter bool`

## Interaction Flows

### Spawn Agent (p)

1. Press `p` → `showPersonaPicker()` builds MenuModel from `pool.Personas()`
2. Sets `menuVisible = true`, `menuIntent = MenuIntentPersonaPicker`
3. Arrow/Enter/Esc handled by existing `handleMenuKey`
4. Enter → `dispatchMenuAction` routes via `menuIntent` → stores `pendingPersonaID`, shows input bar with "Task for <Name>: "
5. Enter on input bar → `pool.Spawn(pendingPersonaID, value, "")`, creates thread in store, clears state

### Message Agent (m)

1. Press `m` → `sendMessageToAgent()` builds MenuModel from `pool.All()`
2. Sets `menuVisible = true`, `menuIntent = MenuIntentAgentPicker`
3. Enter → stores `pendingTargetAgentID`, shows input bar with "Message to <Name>: "
4. Enter on input bar → `client.SendUserInput(message)` on target agent, clears state

### Swarm Filter (s)

- Toggles `swarmFilter` bool
- Canvas skips threads where `AgentProcessID == ""` when filter is active
- Selection index clamps to filtered set

### Header Wiring

- In `NewAppModel`, after applying opts: if `pool != nil` → `header.SetSwarmActive(true)`

## Key Handling Priority

```
handleKey:
  1. helpVisible     → handleHelpKey
  2. inputBarVisible → handleInputBarKey  (NEW)
  3. menuVisible     → handleMenuKey
  4. prefix          → handlePrefix
  5. session focus   → handleSessionKey
  6. canvas          → handleCanvasKey
```

### handleInputBarKey

- Printable runes → append to value
- Backspace → delete last char
- Enter → dispatch based on intent, clear state
- Esc → dismiss, clear pending state

## View Changes

When `inputBarVisible` is true, the input bar renders in place of the status bar at the bottom of the screen. The canvas remains visible above.

## Error Handling

- No pool: `p` and `m` are no-ops
- No personas loaded: status bar error "No personas available"
- No active agents: status bar error "No active agents"
- Spawn at capacity: `pool.Spawn` error → status bar
- Empty input on Enter: dismiss without action
