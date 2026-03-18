# TUI Improvements Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a header shortcuts bar, make `n` spawn+open sessions, and make the TUI full-height with centered, scaled cards.

**Architecture:** Three independent UI improvements to the Bubble Tea TUI. The header bar is a new render function. The `n` key change modifies `createThread` and `ThreadCreatedMsg` handling to chain into the existing `openSession` flow. The full-height layout changes card sizing from constants to dynamic functions and wraps the canvas grid in `lipgloss.Place` for centering.

**Tech Stack:** Go, Bubble Tea, Lipgloss

---

### Task 1: Header Bar — Test

**Files:**
- Create: `internal/tui/header_test.go`

**Step 1: Write the failing test**

```go
package tui

import (
	"strings"
	"testing"
)

func TestHeaderBarRendersTitle(t *testing.T) {
	header := NewHeaderBar(80)
	output := header.View()

	if !strings.Contains(output, "DJ") {
		t.Errorf("expected title in header, got:\n%s", output)
	}
}

func TestHeaderBarRendersShortcuts(t *testing.T) {
	header := NewHeaderBar(80)
	output := header.View()

	if !strings.Contains(output, "n: new") {
		t.Errorf("expected shortcut hints in header, got:\n%s", output)
	}
}

func TestHeaderBarFitsWidth(t *testing.T) {
	header := NewHeaderBar(120)
	output := header.View()

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if len(line) > 120 {
			t.Errorf("header exceeds width 120: len=%d", len(line))
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui -run TestHeaderBar -v`
Expected: FAIL — `NewHeaderBar` undefined

---

### Task 2: Header Bar — Implementation

**Files:**
- Create: `internal/tui/header.go`
- Modify: `internal/tui/app_view.go:10-16` (replace titleStyle and title rendering)

**Step 1: Create header.go**

```go
package tui

import "github.com/charmbracelet/lipgloss"

var (
	headerTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39"))
	headerHintStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))
)

const headerTitle = "DJ — Codex TUI Visualizer"

var headerHints = []string{
	"n: new",
	"Enter: open",
	"?: help",
	"t: tree",
	"Ctrl+B: prefix",
}

type HeaderBar struct {
	width int
}

func NewHeaderBar(width int) HeaderBar {
	return HeaderBar{width: width}
}

func (header *HeaderBar) SetWidth(width int) {
	header.width = width
}

func (header HeaderBar) View() string {
	title := headerTitleStyle.Render(headerTitle)

	hints := ""
	for index, hint := range headerHints {
		if index > 0 {
			hints += "  "
		}
		hints += hint
	}
	renderedHints := headerHintStyle.Render(hints)

	return lipgloss.NewStyle().Width(header.width).Render(
		lipgloss.JoinHorizontal(lipgloss.Top, title, "  ", renderedHints),
	)
}
```

Wait — `lipgloss.Place` is better here for left/right alignment. Revised:

```go
func (header HeaderBar) View() string {
	title := headerTitleStyle.Render(headerTitle)

	hints := ""
	for index, hint := range headerHints {
		if index > 0 {
			hints += "  "
		}
		hints += hint
	}
	renderedHints := headerHintStyle.Render(hints)

	leftRight := title + renderedHints
	return lipgloss.PlaceHorizontal(header.width, lipgloss.Left, title,
		lipgloss.WithWhitespaceChars(" ")) + "\r" +
		lipgloss.PlaceHorizontal(header.width, lipgloss.Right, renderedHints,
			lipgloss.WithWhitespaceChars(" "))
}
```

Actually, the simplest approach: render title left-aligned, render hints right-aligned, pad the gap with spaces.

```go
func (header HeaderBar) View() string {
	title := headerTitleStyle.Render(headerTitle)

	hints := ""
	for index, hint := range headerHints {
		if index > 0 {
			hints += "  "
		}
		hints += hint
	}
	renderedHints := headerHintStyle.Render(hints)

	gap := header.width - lipgloss.Width(title) - lipgloss.Width(renderedHints)
	if gap < 1 {
		gap = 1
	}
	padding := lipgloss.NewStyle().Width(gap).Render("")
	return title + padding + renderedHints
}
```

**Step 2: Update app_view.go**

Remove the `titleStyle` variable (lines 10-13). Replace `title := titleStyle.Render(...)` with usage of the new HeaderBar.

In `app_view.go`, change:
```go
// Remove:
var titleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("39")).
	MarginBottom(1)

// In View():
// Replace: title := titleStyle.Render("DJ — Codex TUI Visualizer")
// With:    title := app.header.View()
```

Add `header HeaderBar` field to `AppModel` in `app.go:16-39`. Initialize in `NewAppModel`. Update width in the `tea.WindowSizeMsg` handler alongside `statusBar.SetWidth`.

**Step 3: Run tests**

Run: `go test ./internal/tui -run TestHeaderBar -v`
Expected: PASS

**Step 4: Run all tests to check for regressions**

Run: `go test ./internal/tui -v`
Expected: All pass (some tests check for "DJ" in view output — the title is still present)

**Step 5: Commit**

```bash
git add internal/tui/header.go internal/tui/header_test.go internal/tui/app.go internal/tui/app_view.go
git commit -m "feat: add header bar with keyboard shortcut hints"
```

---

### Task 3: `n` Key Spawns Session — Test

**Files:**
- Modify: `internal/tui/app_test.go`

**Step 1: Write the failing test**

Add to `app_test.go`:

```go
func TestAppNewThreadCreatesAndOpensSession(t *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store, WithInteractiveCommand("cat"))
	app.width = 120
	app.height = 40

	nKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updated, cmd := app.Update(nKey)
	app = updated.(AppModel)

	if cmd == nil {
		t.Fatal("expected command from n key")
	}

	msg := cmd()
	updated, _ = app.Update(msg)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	threads := store.All()
	if len(threads) != 1 {
		t.Fatalf("expected 1 thread, got %d", len(threads))
	}

	if app.FocusPane() != FocusPaneSession {
		t.Errorf("expected session focus after new thread, got %d", app.FocusPane())
	}

	if len(app.sessionPanel.PinnedSessions()) != 1 {
		t.Errorf("expected 1 pinned session, got %d", len(app.sessionPanel.PinnedSessions()))
	}
}

func TestAppNewThreadIncrementsTitle(t *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store, WithInteractiveCommand("cat"))
	app.width = 120
	app.height = 40

	nKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}

	updated, cmd := app.Update(nKey)
	app = updated.(AppModel)
	msg := cmd()
	updated, _ = app.Update(msg)
	app = updated.(AppModel)

	updated, cmd = app.Update(nKey)
	app = updated.(AppModel)
	msg = cmd()
	updated, _ = app.Update(msg)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	threads := store.All()
	if len(threads) != 2 {
		t.Fatalf("expected 2 threads, got %d", len(threads))
	}
	if threads[0].Title != "Session 1" {
		t.Errorf("expected 'Session 1', got %s", threads[0].Title)
	}
	if threads[1].Title != "Session 2" {
		t.Errorf("expected 'Session 2', got %s", threads[1].Title)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui -run TestAppNewThread -v`
Expected: FAIL — new thread doesn't open session or increment titles

---

### Task 4: `n` Key Spawns Session — Implementation

**Files:**
- Modify: `internal/tui/app.go:16-39` (add `sessionCounter` field)
- Modify: `internal/tui/app.go:128-132` (`ThreadCreatedMsg` handler)
- Modify: `internal/tui/app.go:187-197` (`createThread` function)

**Step 1: Add sessionCounter to AppModel**

In `app.go`, add `sessionCounter int` field to `AppModel` struct.

**Step 2: Update createThread to generate proper IDs**

Replace `createThread` (lines 187-197):

```go
func (app *AppModel) createThread() tea.Cmd {
	app.sessionCounter++
	counter := app.sessionCounter
	return func() tea.Msg {
		return ThreadCreatedMsg{
			ThreadID: fmt.Sprintf("session-%d", counter),
			Title:    fmt.Sprintf("Session %d", counter),
		}
	}
}
```

Note: this removes the `app.client == nil` guard — `n` always creates a local session now. Add `"fmt"` import if not present.

**Step 3: Update ThreadCreatedMsg handler to chain into openSession**

Replace the `ThreadCreatedMsg` case in `Update()` (lines 128-132):

```go
case ThreadCreatedMsg:
	app.store.Add(msg.ThreadID, msg.Title)
	app.statusBar.SetThreadCount(len(app.store.All()))
	app.canvas.SetSelected(len(app.store.All()) - 1)
	return app.openSession()
```

**Step 4: Add SetSelected to CanvasModel**

In `canvas.go`, add:

```go
func (canvas *CanvasModel) SetSelected(index int) {
	threads := canvas.store.All()
	if index >= 0 && index < len(threads) {
		canvas.selected = index
	}
}
```

**Step 5: Run tests**

Run: `go test ./internal/tui -run TestAppNewThread -v`
Expected: PASS

**Step 6: Update existing TestAppNewThread test**

The existing `TestAppNewThread` (line 351) and `TestAppHandlesThreadCreatedMsg` (line 363) need updating since behavior changed. `TestAppNewThread` should verify the cmd produces a `ThreadCreatedMsg`. `TestAppHandlesThreadCreatedMsg` now expects a pinned session — update or split the test.

**Step 7: Run all tests**

Run: `go test ./internal/tui -v`
Expected: All pass

**Step 8: Commit**

```bash
git add internal/tui/app.go internal/tui/app_test.go internal/tui/canvas.go
git commit -m "feat: n key spawns new session and auto-opens it"
```

---

### Task 5: Full-Height Layout — Test

**Files:**
- Modify: `internal/tui/canvas_test.go`
- Modify: `internal/tui/card_test.go`

**Step 1: Write failing tests for dynamic card sizing**

Add to `card_test.go`:

```go
func TestCardDynamicSize(t *testing.T) {
	thread := state.NewThreadState("t-1", "Test")
	thread.Status = state.StatusActive

	card := NewCardModel(thread, false)
	card.SetSize(50, 10)
	output := card.View()

	if !strings.Contains(output, "Test") {
		t.Errorf("expected title in dynamic card, got:\n%s", output)
	}
}
```

Add to `canvas_test.go`:

```go
func TestCanvasViewWithDimensions(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "First")
	store.Add("t-2", "Second")
	store.Add("t-3", "Third")

	canvas := NewCanvasModel(store)
	canvas.SetDimensions(120, 30)
	output := canvas.View()

	if !strings.Contains(output, "First") {
		t.Errorf("expected First in output:\n%s", output)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/tui -run "TestCardDynamic|TestCanvasViewWithDimensions" -v`
Expected: FAIL — `SetSize`/`SetDimensions` undefined

---

### Task 6: Full-Height Layout — Card Scaling

**Files:**
- Modify: `internal/tui/card.go`

**Step 1: Make card sizes dynamic**

Replace fixed `cardWidth`/`cardHeight` constants with fields on `CardModel`. Keep the constants as defaults/minimums.

```go
const (
	minCardWidth  = 20
	minCardHeight = 4
)

type CardModel struct {
	thread   *state.ThreadState
	selected bool
	width    int
	height   int
}

func NewCardModel(thread *state.ThreadState, selected bool) CardModel {
	return CardModel{
		thread:   thread,
		selected: selected,
		width:    minCardWidth,
		height:   minCardHeight,
	}
}

func (card *CardModel) SetSize(width int, height int) {
	if width < minCardWidth {
		width = minCardWidth
	}
	if height < minCardHeight {
		height = minCardHeight
	}
	card.width = width
	card.height = height
}

func (card CardModel) View() string {
	statusColor, exists := statusColors[card.thread.Status]
	if !exists {
		statusColor = defaultStatusColor
	}

	statusLine := lipgloss.NewStyle().
		Foreground(statusColor).
		Render(card.thread.Status)

	title := truncate(card.thread.Title, card.width-4)
	content := fmt.Sprintf("%s\n%s", title, statusLine)

	style := lipgloss.NewStyle().
		Width(card.width).
		Height(card.height).
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)

	if card.selected {
		style = style.
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("39"))
	}

	return style.Render(content)
}
```

Remove the old `cardStyle` and `selectedCardStyle` package-level vars.

**Step 2: Run card tests**

Run: `go test ./internal/tui -run TestCard -v`
Expected: PASS

---

### Task 7: Full-Height Layout — Canvas Centering

**Files:**
- Modify: `internal/tui/canvas.go`

**Step 1: Add dimensions and centering to CanvasModel**

```go
type CanvasModel struct {
	store    *state.ThreadStore
	selected int
	width    int
	height   int
}

func (canvas *CanvasModel) SetDimensions(width int, height int) {
	canvas.width = width
	canvas.height = height
}
```

**Step 2: Update View() to use dynamic sizing and centering**

```go
func (canvas *CanvasModel) View() string {
	threads := canvas.store.All()
	if len(threads) == 0 {
		return lipgloss.Place(canvas.width, canvas.height,
			lipgloss.Center, lipgloss.Center,
			"No active threads. Press 'n' to create one.")
	}

	numRows := (len(threads) + canvasColumns - 1) / canvasColumns
	cardWidth, cardHeight := canvas.cardDimensions(numRows)

	var rows []string
	for rowStart := 0; rowStart < len(threads); rowStart += canvasColumns {
		rowEnd := rowStart + canvasColumns
		if rowEnd > len(threads) {
			rowEnd = len(threads)
		}

		var cards []string
		for index := rowStart; index < rowEnd; index++ {
			isSelected := index == canvas.selected
			card := NewCardModel(threads[index], isSelected)
			card.SetSize(cardWidth, cardHeight)
			cards = append(cards, card.View())
		}

		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cards...))
	}

	grid := strings.Join(rows, "\n")
	return lipgloss.Place(canvas.width, canvas.height,
		lipgloss.Center, lipgloss.Center, grid)
}

func (canvas CanvasModel) cardDimensions(numRows int) (int, int) {
	columnGap := 0
	rowGap := 1

	cardWidth := (canvas.width - columnGap*(canvasColumns-1)) / canvasColumns
	if cardWidth < minCardWidth {
		cardWidth = minCardWidth
	}

	totalRowGaps := rowGap * (numRows - 1)
	cardHeight := (canvas.height - totalRowGaps) / numRows
	if cardHeight < minCardHeight {
		cardHeight = minCardHeight
	}

	return cardWidth, cardHeight
}
```

**Step 3: Run canvas tests**

Run: `go test ./internal/tui -run TestCanvas -v`
Expected: PASS

---

### Task 8: Full-Height Layout — View Composition

**Files:**
- Modify: `internal/tui/app_view.go`

**Step 1: Update View() for full-height layout**

The key change: compute available canvas height, pass dimensions to canvas, use `lipgloss.JoinVertical` to stack header + canvas + status bar across the full terminal height.

```go
const (
	headerHeight    = 1
	statusBarHeight = 1
)

func (app AppModel) View() string {
	header := app.header.View()
	status := app.statusBar.View()

	if app.helpVisible {
		return header + "\n" + app.help.View() + "\n" + status
	}

	if app.menuVisible {
		return header + "\n" + app.menu.View() + "\n" + status
	}

	hasPinned := len(app.sessionPanel.PinnedSessions()) > 0

	if hasPinned {
		canvasHeight := int(float64(app.height) * app.sessionPanel.SplitRatio()) - headerHeight - statusBarHeight
		if canvasHeight < 1 {
			canvasHeight = 1
		}
		app.canvas.SetDimensions(app.width, canvasHeight)
		canvas := app.renderCanvas()
		divider := app.renderDivider()
		panel := app.renderSessionPanel()
		return header + "\n" + canvas + "\n" + divider + "\n" + panel + "\n" + status
	}

	canvasHeight := app.height - headerHeight - statusBarHeight
	if canvasHeight < 1 {
		canvasHeight = 1
	}
	app.canvas.SetDimensions(app.width, canvasHeight)
	canvas := app.renderCanvas()
	return header + "\n" + canvas + "\n" + status
}
```

**Step 2: Run all tests**

Run: `go test ./internal/tui -v`
Expected: All pass

**Step 3: Run linter**

Run: `golangci-lint run ./internal/tui/...`
Expected: No new issues (check funlen, file length)

**Step 4: Commit**

```bash
git add internal/tui/card.go internal/tui/card_test.go internal/tui/canvas.go internal/tui/canvas_test.go internal/tui/app_view.go
git commit -m "feat: full-height layout with centered, scaled cards"
```

---

### Task 9: Integration Test and Final Verification

**Step 1: Build the binary**

Run: `go build -o dj ./cmd/dj`
Expected: Builds cleanly

**Step 2: Run full test suite with race detector**

Run: `go test ./... -v -race`
Expected: All pass

**Step 3: Run linter**

Run: `golangci-lint run`
Expected: Clean

**Step 4: Commit any fixups**

If any fixes were needed, commit them.
