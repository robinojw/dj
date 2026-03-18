# Session Scrollback Design

## Problem

When a codex CLI session produces output that scrolls past the visible area, the user cannot scroll up to review earlier output. The vt emulator stores a scrollback buffer, but the TUI does not expose it.

## Approach

Add scroll state directly to `PTYSession`. The emulator already maintains a scrollback buffer via `Scrollback()`. When the user scrolls up with the mouse wheel, the `Render()` method builds a custom viewport from scrollback lines + visible screen lines instead of calling `emulator.Render()`.

## Design

### Scroll state on PTYSession

Add to `PTYSession`:
- `scrollOffset int` — 0 means at bottom (live), positive means lines scrolled up
- `ScrollUp(lines int)` — increase offset, clamped to max scrollback
- `ScrollDown(lines int)` — decrease offset, clamped to 0
- `ScrollToBottom()` — reset offset to 0
- `IsScrolledUp() bool` — returns `scrollOffset > 0`
- `ScrollOffset() int` — returns current offset

### Custom viewport rendering

When `scrollOffset > 0`, `Render()` builds output by:
1. Collecting all scrollback lines via `Scrollback().Lines()`
2. Collecting visible screen lines via `CellAt(x, y)` for each row
3. Concatenating into one logical buffer (scrollback on top, screen on bottom)
4. Slicing a window of `emulator.Height()` lines, offset from the bottom by `scrollOffset`
5. Converting cells to styled strings for display

When `scrollOffset == 0`, `Render()` calls `emulator.Render()` as before.

### Mouse input

- Enable mouse mode with `tea.WithMouseCellMotion()` in program options
- Handle `tea.MouseMsg` in `Update()`:
  - Scroll wheel up → `ScrollUp` on active PTY session
  - Scroll wheel down → `ScrollDown` on active PTY session
- Do not forward scroll wheel events to the PTY process
- Non-scroll mouse events are not forwarded (PTY apps that need mouse input are out of scope)

### Auto-scroll behavior

When new output arrives while scrolled up, the view stays in place. The user must scroll down manually or the offset resets on keyboard input to the PTY.

### Scroll indicator

When `IsScrolledUp()` is true, render a bottom-line indicator in the session pane:
- Format: `↓ N lines below`
- Styled with a distinct background so it overlays the content visibly
- Disappears when scroll offset returns to 0

## Files changed

- `internal/tui/pty_session.go` — scroll state, modified `Render()`
- `internal/tui/app.go` — mouse message handling in `Update()`
- `internal/tui/app_pty.go` — scroll dispatch for active session
- `internal/tui/app_view.go` — scroll indicator overlay in `renderPTYContent()`
- `cmd/dj/main.go` — add `tea.WithMouseCellMotion()` to program options
