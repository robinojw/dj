# Phase 4: Canvas UI & Agent Cards

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build the Bubble Tea application shell with a canvas grid of agent cards. Each card renders a thread's status, title, and recent activity. Arrow keys navigate between cards. The selected card is highlighted.

**Architecture:** The `App` model is the root Bubble Tea model. It owns the `ThreadStore`, the `appserver.Client`, and the canvas layout. The canvas is a grid of `CardModel` components rendered via Lipgloss. The `App.Update` method handles keyboard input and event bridge messages, mutating the store and triggering re-renders.

**Tech Stack:** Go, Bubble Tea, Lipgloss

**Prerequisites:** Phase 3 (state store, message types, event bridge)

---

### Task 1: Add Bubble Tea and Lipgloss Dependencies

**Files:**
- Modify: `go.mod`

**Step 1: Add dependencies**

```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go mod tidy
```

**Step 2: Verify**

Run: `go build ./...`
Expected: builds successfully

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "deps: add Bubble Tea and Lipgloss"
```

---

### Task 2: Build Agent Card Component

**Files:**
- Create: `internal/tui/card.go`
- Create: `internal/tui/card_test.go`

**Step 1: Write tests for card rendering**

```go
// internal/tui/card_test.go
package tui

import (
	"strings"
	"testing"

	"github.com/robinojw/dj/internal/state"
)

func TestCardRenderShowsTitle(t *testing.T) {
	thread := state.NewThreadState("t-1", "Build web app")
	thread.Status = state.StatusActive

	card := NewCardModel(thread, false)
	output := card.View()

	if !strings.Contains(output, "Build web app") {
		t.Errorf("expected title in output, got:\n%s", output)
	}
}

func TestCardRenderShowsStatus(t *testing.T) {
	thread := state.NewThreadState("t-1", "Test")
	thread.Status = state.StatusActive

	card := NewCardModel(thread, false)
	output := card.View()

	if !strings.Contains(output, "active") {
		t.Errorf("expected status in output, got:\n%s", output)
	}
}

func TestCardRenderSelectedHighlight(t *testing.T) {
	thread := state.NewThreadState("t-1", "Test")
	card := NewCardModel(thread, true)
	selected := card.View()

	card2 := NewCardModel(thread, false)
	unselected := card2.View()

	if selected == unselected {
		t.Error("selected and unselected cards should differ")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -v -run TestCard`
Expected: FAIL — `NewCardModel` not defined

**Step 3: Implement card component**

```go
// internal/tui/card.go
package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/state"
)

const (
	cardWidth  = 30
	cardHeight = 6
)

var (
	cardStyle = lipgloss.NewStyle().
			Width(cardWidth).
			Height(cardHeight).
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1)

	selectedCardStyle = cardStyle.
				BorderForeground(lipgloss.Color("39"))

	statusColors = map[string]lipgloss.Color{
		state.StatusActive:    lipgloss.Color("42"),
		state.StatusIdle:      lipgloss.Color("245"),
		state.StatusCompleted: lipgloss.Color("34"),
		state.StatusError:     lipgloss.Color("196"),
	}
)

// CardModel renders a single agent thread as a card.
type CardModel struct {
	thread   *state.ThreadState
	selected bool
}

// NewCardModel creates a card for the given thread.
func NewCardModel(thread *state.ThreadState, selected bool) CardModel {
	return CardModel{
		thread:   thread,
		selected: selected,
	}
}

// View renders the card.
func (c CardModel) View() string {
	statusColor, exists := statusColors[c.thread.Status]
	if !exists {
		statusColor = lipgloss.Color("245")
	}

	statusLine := lipgloss.NewStyle().
		Foreground(statusColor).
		Render(c.thread.Status)

	title := truncate(c.thread.Title, cardWidth-4)
	content := fmt.Sprintf("%s\n%s", title, statusLine)

	style := cardStyle
	if c.selected {
		style = selectedCardStyle
	}

	return style.Render(content)
}

func truncate(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}
```

**Step 4: Run tests**

Run: `go test ./internal/tui/ -v -run TestCard`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/card.go internal/tui/card_test.go
git commit -m "feat(tui): agent card component with status colors"
```

---

### Task 3: Build Canvas Grid Layout

**Files:**
- Create: `internal/tui/canvas.go`
- Create: `internal/tui/canvas_test.go`

**Step 1: Write tests for canvas grid**

```go
// internal/tui/canvas_test.go
package tui

import (
	"testing"

	"github.com/robinojw/dj/internal/state"
)

func TestCanvasNavigation(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "First")
	store.Add("t-2", "Second")
	store.Add("t-3", "Third")

	canvas := NewCanvasModel(store)

	if canvas.SelectedIndex() != 0 {
		t.Errorf("expected initial index 0, got %d", canvas.SelectedIndex())
	}

	canvas.MoveRight()
	if canvas.SelectedIndex() != 1 {
		t.Errorf("expected index 1 after right, got %d", canvas.SelectedIndex())
	}

	canvas.MoveLeft()
	if canvas.SelectedIndex() != 0 {
		t.Errorf("expected index 0 after left, got %d", canvas.SelectedIndex())
	}
}

func TestCanvasNavigationBounds(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "First")
	store.Add("t-2", "Second")

	canvas := NewCanvasModel(store)

	canvas.MoveLeft()
	if canvas.SelectedIndex() != 0 {
		t.Errorf("expected clamped at 0, got %d", canvas.SelectedIndex())
	}

	canvas.MoveRight()
	canvas.MoveRight()
	if canvas.SelectedIndex() != 1 {
		t.Errorf("expected clamped at 1, got %d", canvas.SelectedIndex())
	}
}

func TestCanvasSelectedThreadID(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "First")
	store.Add("t-2", "Second")

	canvas := NewCanvasModel(store)
	canvas.MoveRight()

	id := canvas.SelectedThreadID()
	if id != "t-2" {
		t.Errorf("expected t-2, got %s", id)
	}
}

func TestCanvasEmptyStore(t *testing.T) {
	store := state.NewThreadStore()
	canvas := NewCanvasModel(store)

	if canvas.SelectedThreadID() != "" {
		t.Errorf("expected empty ID for empty canvas")
	}

	canvas.MoveRight()
	if canvas.SelectedIndex() != 0 {
		t.Errorf("expected 0 for empty canvas")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -v -run TestCanvas`
Expected: FAIL — `NewCanvasModel` not defined

**Step 3: Implement canvas grid**

```go
// internal/tui/canvas.go
package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/state"
)

const canvasColumns = 3

// CanvasModel manages a grid of agent cards.
type CanvasModel struct {
	store    *state.ThreadStore
	selected int
}

// NewCanvasModel creates a canvas backed by the given store.
func NewCanvasModel(store *state.ThreadStore) CanvasModel {
	return CanvasModel{store: store}
}

// SelectedIndex returns the current selection index.
func (c *CanvasModel) SelectedIndex() int {
	return c.selected
}

// SelectedThreadID returns the ID of the selected thread.
func (c *CanvasModel) SelectedThreadID() string {
	threads := c.store.All()
	if len(threads) == 0 {
		return ""
	}
	return threads[c.selected].ID
}

// MoveRight advances selection to the next card.
func (c *CanvasModel) MoveRight() {
	threads := c.store.All()
	if c.selected < len(threads)-1 {
		c.selected++
	}
}

// MoveLeft moves selection to the previous card.
func (c *CanvasModel) MoveLeft() {
	if c.selected > 0 {
		c.selected--
	}
}

// MoveDown moves selection down one row.
func (c *CanvasModel) MoveDown() {
	threads := c.store.All()
	next := c.selected + canvasColumns
	if next < len(threads) {
		c.selected = next
	}
}

// MoveUp moves selection up one row.
func (c *CanvasModel) MoveUp() {
	next := c.selected - canvasColumns
	if next >= 0 {
		c.selected = next
	}
}

// View renders the canvas grid.
func (c *CanvasModel) View() string {
	threads := c.store.All()
	if len(threads) == 0 {
		return "No active threads. Press 'n' to create one."
	}

	var rows []string
	for i := 0; i < len(threads); i += canvasColumns {
		end := i + canvasColumns
		if end > len(threads) {
			end = len(threads)
		}

		var cards []string
		for j := i; j < end; j++ {
			isSelected := j == c.selected
			card := NewCardModel(threads[j], isSelected)
			cards = append(cards, card.View())
		}

		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cards...))
	}

	return strings.Join(rows, "\n")
}
```

**Step 4: Run tests**

Run: `go test ./internal/tui/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/canvas.go internal/tui/canvas_test.go
git commit -m "feat(tui): canvas grid layout with arrow key navigation"
```

---

### Task 4: Build Root App Model

**Files:**
- Create: `internal/tui/app.go`
- Create: `internal/tui/app_test.go`

**Step 1: Write tests for app model**

```go
// internal/tui/app_test.go
package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/state"
)

func TestAppHandlesArrowKeys(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "First")
	store.Add("t-2", "Second")

	app := NewAppModel(store)

	rightKey := tea.KeyMsg{Type: tea.KeyRight}
	updated, _ := app.Update(rightKey)
	appModel := updated.(AppModel)

	if appModel.canvas.SelectedIndex() != 1 {
		t.Errorf("expected index 1 after right, got %d", appModel.canvas.SelectedIndex())
	}
}

func TestAppHandlesThreadStatusMsg(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Initial")

	app := NewAppModel(store)

	msg := ThreadStatusMsg{
		ThreadID: "t-1",
		Status:   "active",
		Title:    "Running",
	}
	app.Update(msg)

	thread, _ := store.Get("t-1")
	if thread.Status != "active" {
		t.Errorf("expected active, got %s", thread.Status)
	}
	if thread.Title != "Running" {
		t.Errorf("expected Running, got %s", thread.Title)
	}
}

func TestAppHandlesQuit(t *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	quitKey := tea.KeyMsg{Type: tea.KeyCtrlC}
	_, cmd := app.Update(quitKey)

	if cmd == nil {
		t.Fatal("expected quit command")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -v -run TestApp`
Expected: FAIL — `NewAppModel` not defined

**Step 3: Implement app model**

```go
// internal/tui/app.go
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/state"
)

var titleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("39")).
	MarginBottom(1)

// AppModel is the root Bubble Tea model.
type AppModel struct {
	store  *state.ThreadStore
	canvas CanvasModel
	width  int
	height int
}

// NewAppModel creates the root app with the given store.
func NewAppModel(store *state.ThreadStore) AppModel {
	return AppModel{
		store:  store,
		canvas: NewCanvasModel(store),
	}
}

// Init returns the initial command.
func (m AppModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case ThreadStatusMsg:
		m.store.UpdateStatus(msg.ThreadID, msg.Status, msg.Title)
		return m, nil
	case ThreadMessageMsg:
		return m.handleThreadMessage(msg)
	case ThreadDeltaMsg:
		return m.handleThreadDelta(msg)
	case CommandOutputMsg:
		return m.handleCommandOutput(msg)
	}
	return m, nil
}

func (m AppModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		return m, tea.Quit
	case tea.KeyRight, tea.KeyTab:
		m.canvas.MoveRight()
	case tea.KeyLeft, tea.KeyShiftTab:
		m.canvas.MoveLeft()
	case tea.KeyDown:
		m.canvas.MoveDown()
	case tea.KeyUp:
		m.canvas.MoveUp()
	}
	return m, nil
}

func (m AppModel) handleThreadMessage(msg ThreadMessageMsg) (tea.Model, tea.Cmd) {
	thread, exists := m.store.Get(msg.ThreadID)
	if !exists {
		return m, nil
	}
	thread.AppendMessage(state.ChatMessage{
		ID:      msg.MessageID,
		Role:    msg.Role,
		Content: msg.Content,
	})
	return m, nil
}

func (m AppModel) handleThreadDelta(msg ThreadDeltaMsg) (tea.Model, tea.Cmd) {
	thread, exists := m.store.Get(msg.ThreadID)
	if !exists {
		return m, nil
	}
	thread.AppendDelta(msg.MessageID, msg.Delta)
	return m, nil
}

func (m AppModel) handleCommandOutput(msg CommandOutputMsg) (tea.Model, tea.Cmd) {
	thread, exists := m.store.Get(msg.ThreadID)
	if !exists {
		return m, nil
	}
	thread.AppendOutput(msg.ExecID, msg.Data)
	return m, nil
}

// View renders the full UI.
func (m AppModel) View() string {
	title := titleStyle.Render("DJ — Codex TUI Visualizer")
	canvas := m.canvas.View()
	return title + "\n" + canvas + "\n"
}
```

**Step 4: Run tests**

Run: `go test ./internal/tui/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/app.go internal/tui/app_test.go
git commit -m "feat(tui): root App model with canvas and event handling"
```

---

### Task 5: Wire main.go to Launch TUI

**Files:**
- Modify: `cmd/dj/main.go`

**Step 1: Update main.go to start Bubble Tea**

```go
// cmd/dj/main.go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/state"
	"github.com/robinojw/dj/internal/tui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	store := state.NewThreadStore()
	app := tui.NewAppModel(store)

	program := tea.NewProgram(app, tea.WithAltScreen())
	_, err := program.Run()
	return err
}
```

**Step 2: Verify it builds and runs**

Run: `go build ./cmd/dj && echo "Build OK"`
Expected: Build OK

> Note: Running `./dj` will show the TUI with an empty canvas. Press Ctrl+C to exit.

**Step 3: Commit**

```bash
git add cmd/dj/main.go
git commit -m "feat: wire main.go to launch Bubble Tea TUI"
```
