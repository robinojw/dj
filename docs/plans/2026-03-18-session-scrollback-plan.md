# Session Scrollback Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Enable mouse-wheel scrollback on PTY session panels so users can scroll up through codex CLI output history.

**Architecture:** Add scroll offset tracking to `PTYSession`, build a custom viewport renderer that combines scrollback buffer lines with visible screen lines, intercept mouse wheel events in Bubble Tea's `Update()` to adjust the offset, and overlay a scroll indicator when scrolled up.

**Tech Stack:** Go, Bubble Tea (mouse events), charmbracelet/x/vt (scrollback buffer, `uv.Line.Render()`), Lipgloss (indicator styling)

---

### Task 1: Add scroll state to PTYSession

**Files:**
- Modify: `internal/tui/pty_session.go`
- Test: `internal/tui/pty_session_test.go`

**Step 1: Write failing tests for scroll state methods**

Add to `pty_session_test.go`:

```go
func TestPTYSessionScrollOffset(t *testing.T) {
	session := NewPTYSession(PTYSessionConfig{
		ThreadID: "t-1",
		Command:  "echo",
		Args:     []string{"test"},
		SendMsg:  func(msg PTYOutputMsg) {},
	})

	if session.ScrollOffset() != 0 {
		t.Errorf("expected initial offset 0, got %d", session.ScrollOffset())
	}

	if session.IsScrolledUp() {
		t.Error("expected not scrolled up initially")
	}
}

func TestPTYSessionScrollUpDown(t *testing.T) {
	session := NewPTYSession(PTYSessionConfig{
		ThreadID: "t-1",
		Command:  "echo",
		Args:     []string{"test"},
		SendMsg:  func(msg PTYOutputMsg) {},
	})

	session.ScrollUp(5)
	if session.ScrollOffset() != 0 {
		t.Errorf("expected offset 0 with no scrollback, got %d", session.ScrollOffset())
	}

	session.ScrollDown(3)
	if session.ScrollOffset() != 0 {
		t.Errorf("expected offset 0 after scroll down, got %d", session.ScrollOffset())
	}
}

func TestPTYSessionScrollToBottom(t *testing.T) {
	session := NewPTYSession(PTYSessionConfig{
		ThreadID: "t-1",
		Command:  "echo",
		Args:     []string{"test"},
		SendMsg:  func(msg PTYOutputMsg) {},
	})

	session.ScrollToBottom()
	if session.ScrollOffset() != 0 {
		t.Errorf("expected offset 0 after scroll to bottom, got %d", session.ScrollOffset())
	}
	if session.IsScrolledUp() {
		t.Error("expected not scrolled up after scroll to bottom")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/tui -run TestPTYSessionScroll -v`
Expected: FAIL — `ScrollOffset`, `IsScrolledUp`, `ScrollUp`, `ScrollDown`, `ScrollToBottom` undefined

**Step 3: Implement scroll state on PTYSession**

Add to `pty_session.go`:

1. Add `scrollOffset int` field to the `PTYSession` struct.

2. Add these methods:

```go
const scrollStep = 3

func (ps *PTYSession) ScrollUp(lines int) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	maxOffset := ps.emulator.ScrollbackLen()
	ps.scrollOffset += lines
	if ps.scrollOffset > maxOffset {
		ps.scrollOffset = maxOffset
	}
}

func (ps *PTYSession) ScrollDown(lines int) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.scrollOffset -= lines
	if ps.scrollOffset < 0 {
		ps.scrollOffset = 0
	}
}

func (ps *PTYSession) ScrollToBottom() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.scrollOffset = 0
}

func (ps *PTYSession) ScrollOffset() int {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	return ps.scrollOffset
}

func (ps *PTYSession) IsScrolledUp() bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	return ps.scrollOffset > 0
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/tui -run TestPTYSessionScroll -v`
Expected: PASS

**Step 5: Commit**

```
git add internal/tui/pty_session.go internal/tui/pty_session_test.go
git commit -m "feat: add scroll state tracking to PTYSession"
```

---

### Task 2: Implement custom viewport rendering when scrolled

**Files:**
- Create: `internal/tui/pty_scroll.go`
- Test: `internal/tui/pty_scroll_test.go`
- Modify: `internal/tui/pty_session.go` (update `Render()`)

**Step 1: Write failing test for scrolled rendering**

Create `internal/tui/pty_scroll_test.go`:

```go
package tui

import (
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

func TestRenderScrolledViewport(t *testing.T) {
	scrollbackLines := []uv.Line{
		uv.NewLine(10),
		uv.NewLine(10),
	}
	screenLines := []string{"visible-1", "visible-2", "visible-3"}

	result := renderScrolledViewport(scrollbackLines, screenLines, 3, 1)

	if len(result) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(result))
	}

	if !strings.Contains(result[2], "visible-2") {
		t.Errorf("expected visible-2 at bottom, got %q", result[2])
	}
}

func TestRenderScrolledViewportAtMaxOffset(t *testing.T) {
	scrollbackLines := []uv.Line{
		uv.NewLine(10),
		uv.NewLine(10),
	}
	screenLines := []string{"vis-1", "vis-2"}

	result := renderScrolledViewport(scrollbackLines, screenLines, 2, 4)

	if len(result) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(result))
	}
}

func TestRenderScrolledViewportZeroOffset(t *testing.T) {
	scrollbackLines := []uv.Line{
		uv.NewLine(10),
	}
	screenLines := []string{"vis-1", "vis-2"}

	result := renderScrolledViewport(scrollbackLines, screenLines, 2, 0)

	if len(result) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(result))
	}
	if result[0] != "vis-1" {
		t.Errorf("expected vis-1, got %q", result[0])
	}
	if result[1] != "vis-2" {
		t.Errorf("expected vis-2, got %q", result[1])
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/tui -run TestRenderScrolledViewport -v`
Expected: FAIL — `renderScrolledViewport` undefined

**Step 3: Implement the scrolled viewport renderer**

Create `internal/tui/pty_scroll.go`:

```go
package tui

import (
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
)

func renderScrolledViewport(
	scrollbackLines []uv.Line,
	screenLines []string,
	viewportHeight int,
	scrollOffset int,
) []string {
	allLines := make([]string, 0, len(scrollbackLines)+len(screenLines))

	for _, line := range scrollbackLines {
		allLines = append(allLines, line.Render())
	}
	allLines = append(allLines, screenLines...)

	totalLines := len(allLines)
	end := totalLines - scrollOffset
	if end < 0 {
		end = 0
	}
	start := end - viewportHeight
	if start < 0 {
		start = 0
	}
	if end > totalLines {
		end = totalLines
	}

	visible := allLines[start:end]

	for len(visible) < viewportHeight {
		visible = append([]string{""}, visible...)
	}

	return visible
}

func renderScrolledOutput(
	scrollbackLines []uv.Line,
	screenLines []string,
	viewportHeight int,
	scrollOffset int,
) string {
	lines := renderScrolledViewport(scrollbackLines, screenLines, viewportHeight, scrollOffset)
	return strings.Join(lines, "\n")
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/tui -run TestRenderScrolledViewport -v`
Expected: PASS

**Step 5: Update PTYSession.Render() to use scrolled viewport**

Modify `Render()` in `pty_session.go`:

```go
func (ps *PTYSession) Render() string {
	ps.mu.Lock()
	offset := ps.scrollOffset
	ps.mu.Unlock()

	if offset == 0 {
		return ps.emulator.Render()
	}

	scrollback := ps.emulator.Scrollback()
	scrollbackLen := scrollback.Len()
	scrollbackLines := make([]uv.Line, scrollbackLen)
	for i := 0; i < scrollbackLen; i++ {
		scrollbackLines[i] = scrollback.Line(i)
	}

	screenContent := ps.emulator.Render()
	screenLines := strings.Split(screenContent, "\n")

	return renderScrolledOutput(
		scrollbackLines,
		screenLines,
		ps.emulator.Height(),
		offset,
	)
}
```

Add `"strings"` and `uv "github.com/charmbracelet/ultraviolet"` to imports in `pty_session.go`.

**Step 6: Run all PTY tests**

Run: `go test ./internal/tui -run TestPTYSession -v`
Expected: PASS

**Step 7: Commit**

```
git add internal/tui/pty_scroll.go internal/tui/pty_scroll_test.go internal/tui/pty_session.go
git commit -m "feat: custom viewport rendering for scrolled PTY sessions"
```

---

### Task 3: Enable mouse events and handle scroll wheel

**Files:**
- Modify: `cmd/dj/main.go`
- Modify: `internal/tui/app.go`
- Modify: `internal/tui/app_pty.go`
- Test: `internal/tui/app_test.go`

**Step 1: Write failing test for mouse scroll handling**

Add to `app_test.go`:

```go
func TestAppMouseScrollUpOnSession(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Thread 1")

	app := NewAppModel(store, WithInteractiveCommand("cat"))
	app.width = 120
	app.height = 40

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	scrollUp := tea.MouseMsg{
		Button: tea.MouseButtonWheelUp,
		Action: tea.MouseActionPress,
	}
	updated, _ = app.Update(scrollUp)
	app = updated.(AppModel)

	ptySession := app.ptySessions["t-1"]
	offset := ptySession.ScrollOffset()
	if offset < 0 {
		t.Errorf("expected non-negative scroll offset, got %d", offset)
	}
}

func TestAppMouseScrollDownOnSession(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Thread 1")

	app := NewAppModel(store, WithInteractiveCommand("cat"))
	app.width = 120
	app.height = 40

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	ptySession := app.ptySessions["t-1"]
	ptySession.ScrollUp(10)

	scrollDown := tea.MouseMsg{
		Button: tea.MouseButtonWheelDown,
		Action: tea.MouseActionPress,
	}
	updated, _ = app.Update(scrollDown)
	app = updated.(AppModel)

	offset := ptySession.ScrollOffset()
	expectedMax := 10 - scrollStep
	if offset > expectedMax {
		t.Errorf("expected offset <= %d after scroll down, got %d", expectedMax, offset)
	}
}

func TestAppMouseScrollIgnoredOnCanvas(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Thread 1")

	app := NewAppModel(store)

	scrollUp := tea.MouseMsg{
		Button: tea.MouseButtonWheelUp,
		Action: tea.MouseActionPress,
	}
	updated, _ := app.Update(scrollUp)
	_ = updated.(AppModel)
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/tui -run TestAppMouseScroll -v`
Expected: FAIL — `tea.MouseMsg` not handled in `Update()`

**Step 3: Add mouse event handling to Update**

In `app.go`, add a `tea.MouseMsg` case to `Update()`:

```go
case tea.MouseMsg:
	return app.handleMouse(msg)
```

Add the handler method to `app_pty.go`:

```go
func (app AppModel) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	isScrollWheel := msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown
	if !isScrollWheel {
		return app, nil
	}

	if app.focusPane != FocusPaneSession {
		return app, nil
	}

	activeID := app.sessionPanel.ActiveThreadID()
	if activeID == "" {
		return app, nil
	}

	ptySession, exists := app.ptySessions[activeID]
	if !exists {
		return app, nil
	}

	if msg.Button == tea.MouseButtonWheelUp {
		ptySession.ScrollUp(scrollStep)
	} else {
		ptySession.ScrollDown(scrollStep)
	}

	return app, nil
}
```

**Step 4: Enable mouse in main.go**

In `cmd/dj/main.go`, change:

```go
program := tea.NewProgram(app, tea.WithAltScreen())
```

to:

```go
program := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())
```

**Step 5: Run tests to verify they pass**

Run: `go test ./internal/tui -run TestAppMouseScroll -v`
Expected: PASS

**Step 6: Run all tests**

Run: `go test ./internal/tui -v`
Expected: PASS

**Step 7: Commit**

```
git add cmd/dj/main.go internal/tui/app.go internal/tui/app_pty.go internal/tui/app_test.go
git commit -m "feat: mouse wheel scroll for PTY session panels"
```

---

### Task 4: Add scroll indicator overlay

**Files:**
- Modify: `internal/tui/app_view.go`
- Test: `internal/tui/app_test.go`

**Step 1: Write failing test for scroll indicator**

Add to `app_test.go`:

```go
func TestAppViewShowsScrollIndicator(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Thread 1")

	app := NewAppModel(store, WithInteractiveCommand("cat"))
	app.width = 80
	app.height = 30

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	ptySession := app.ptySessions["t-1"]
	ptySession.ScrollUp(5)

	view := app.View()
	hasIndicator := strings.Contains(view, "↓") || strings.Contains(view, "lines below")
	if !hasIndicator {
		t.Error("expected scroll indicator when scrolled up")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui -run TestAppViewShowsScrollIndicator -v`
Expected: FAIL — no indicator in view output

**Step 3: Add scroll indicator to renderPTYContent**

In `app_view.go`, modify `renderPTYContent`:

```go
const scrollIndicatorStyle = "240"

func (app AppModel) renderPTYContent(threadID string) string {
	ptySession, exists := app.ptySessions[threadID]
	if !exists {
		return ""
	}

	content := ptySession.Render()
	hasVisibleContent := strings.TrimSpace(content) != ""
	if !hasVisibleContent && !ptySession.Running() {
		return fmt.Sprintf("[process exited: %d]", ptySession.ExitCode())
	}

	if ptySession.IsScrolledUp() {
		indicator := renderScrollIndicator(ptySession.ScrollOffset())
		lines := strings.Split(content, "\n")
		if len(lines) > 0 {
			lines[len(lines)-1] = indicator
		}
		content = strings.Join(lines, "\n")
	}

	return content
}

func renderScrollIndicator(linesBelow int) string {
	text := fmt.Sprintf(" ↓ %d lines below ", linesBelow)
	style := lipgloss.NewStyle().
		Background(lipgloss.Color(scrollIndicatorStyle)).
		Foreground(lipgloss.Color("255"))
	return style.Render(text)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui -run TestAppViewShowsScrollIndicator -v`
Expected: PASS

**Step 5: Run all tests**

Run: `go test ./internal/tui -v`
Expected: PASS

**Step 6: Commit**

```
git add internal/tui/app_view.go internal/tui/app_test.go
git commit -m "feat: scroll indicator overlay when session is scrolled up"
```

---

### Task 5: Update help screen with scroll keybinding

**Files:**
- Modify: `internal/tui/help.go`
- Test: `internal/tui/app_test.go`

**Step 1: Write failing test**

Add to `app_test.go`:

```go
func TestHelpShowsScrollKeybinding(t *testing.T) {
	help := NewHelpModel()
	view := help.View()
	if !strings.Contains(view, "Scroll") {
		t.Error("expected Scroll keybinding in help")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui -run TestHelpShowsScrollKeybinding -v`
Expected: FAIL

**Step 3: Add scroll entry to help**

Read `internal/tui/help.go` and add a line for mouse scroll in the keybindings list. The exact format depends on the existing help entries — match their pattern. Add something like:

```
"Mouse Wheel    Scroll session up/down"
```

in the session section of the help text.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui -run TestHelpShowsScrollKeybinding -v`
Expected: PASS

**Step 5: Commit**

```
git add internal/tui/help.go internal/tui/app_test.go
git commit -m "feat: add scroll keybinding to help screen"
```

---

### Task 6: Lint and full test pass

**Files:** All modified files

**Step 1: Run linter**

Run: `golangci-lint run`
Expected: No errors. If there are funlen violations (60 line max), extract helper functions.

**Step 2: Run all tests with race detector**

Run: `go test ./... -v -race`
Expected: PASS

**Step 3: Run build**

Run: `go build -o dj ./cmd/dj`
Expected: Build succeeds

**Step 4: Fix any issues found**

Address lint/race/build errors if any.

**Step 5: Commit fixes if needed**

```
git add -A
git commit -m "fix: lint and race detector issues"
```
