# TUI Improvements Design

## 1. Header Bar with Keyboard Shortcuts

Replace the plain title with a full-width header bar. Title left-aligned (cyan, bold), keyboard shortcuts right-aligned (dimmed gray) on the same line.

```
DJ — Codex TUI Visualizer          n: new  Enter: open  ?: help  t: tree  Ctrl+B: prefix
```

Use `lipgloss.Place` with left/right alignment to compose the two halves into one row.

## 2. `n` Key Spawns and Opens a New Session

Pressing `n` creates a new thread, adds it to the store, pins it, spawns a blank PTY, and focuses the session pane automatically.

Flow:
1. Generate a unique thread ID and incrementing title ("Session 1", "Session 2", etc.)
2. Add to `ThreadStore`
3. Move canvas selection to the new thread
4. Pin it, spawn PTY, focus session pane (reuse `openSession` logic)

The `ThreadCreatedMsg` handler chains into the open-session sequence rather than just updating the store.

## 3. Full-Height Layout with Centered, Scaled Cards

The TUI fills the terminal. The status bar anchors to the bottom. Cards scale to fill and center within the canvas area.

Height budget:
- Header: 1 line
- Canvas: terminal height - header - status bar (or split with session panel when pinned)
- Status bar: 1 line

Card scaling:
- `cardWidth = (canvasWidth - columnGaps) / canvasColumns`
- `cardHeight = (canvasHeight - rowGaps) / numRows`
- Minimum clamp: 20 wide, 4 tall

Centering:
- `lipgloss.Place(width, canvasHeight, lipgloss.Center, lipgloss.Center, grid)` centers the card grid both horizontally and vertically.

Canvas receives width/height so it can compute dynamic card sizes. Card styles become functions rather than constants.
