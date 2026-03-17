# Phase 6: Session Pane

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a scrollable session pane that displays a selected thread's conversation messages and command output. The pane opens when Enter is pressed on a canvas card and shows the full conversation history with streaming delta support.

**Architecture:** `SessionModel` is a Bubble Tea component that renders a thread's messages and command output in a scrollable viewport. It subscribes to the selected thread ID and re-renders when the store changes. The viewport uses Bubble Tea's `viewport` component from the Bubbles library for scrolling. The App model gains a third focus mode (FocusSession) with Esc to return to canvas.

**Tech Stack:** Go, Bubble Tea, Lipgloss, Bubbles viewport

**Prerequisites:** Phase 5 (agent tree, focus system)

---

### Task 1: Add Bubbles Viewport Dependency

**Files:**
- Modify: `go.mod`

**Step 1: Add dependency**

```bash
go get github.com/charmbracelet/bubbles
go mod tidy
```

**Step 2: Verify**

Run: `go build ./...`
Expected: builds successfully

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "deps: add Bubbles component library"
```

---

### Task 2: Build Session Message Renderer

**Files:**
- Create: `internal/tui/session_render.go`
- Create: `internal/tui/session_render_test.go`

**Step 1: Write tests for message rendering**

```go
// internal/tui/session_render_test.go
package tui

import (
	"strings"
	"testing"

	"github.com/robinojw/dj/internal/state"
)

func TestRenderMessages(t *testing.T) {
	thread := state.NewThreadState("t-1", "Test")
	thread.AppendMessage(state.ChatMessage{
		ID: "m-1", Role: "user", Content: "Hello",
	})
	thread.AppendMessage(state.ChatMessage{
		ID: "m-2", Role: "assistant", Content: "Hi there",
	})

	output := RenderMessages(thread)

	if !strings.Contains(output, "Hello") {
		t.Errorf("expected user message in output:\n%s", output)
	}
	if !strings.Contains(output, "Hi there") {
		t.Errorf("expected assistant message in output:\n%s", output)
	}
}

func TestRenderMessagesWithCommand(t *testing.T) {
	thread := state.NewThreadState("t-1", "Test")
	thread.AppendOutput("e-1", "go test ./...\nPASS\n")

	output := RenderMessages(thread)

	if !strings.Contains(output, "PASS") {
		t.Errorf("expected command output in output:\n%s", output)
	}
}

func TestRenderMessagesEmpty(t *testing.T) {
	thread := state.NewThreadState("t-1", "Test")
	output := RenderMessages(thread)

	if !strings.Contains(output, "No messages") {
		t.Errorf("expected empty state message:\n%s", output)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -v -run TestRenderMessages`
Expected: FAIL — `RenderMessages` not defined

**Step 3: Implement message renderer**

```go
// internal/tui/session_render.go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/state"
)

var (
	userStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)
	assistantStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)
	commandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))
	outputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true)
)

// RenderMessages produces a formatted string of all messages and command output.
func RenderMessages(thread *state.ThreadState) string {
	hasMessages := len(thread.Messages) > 0
	hasOutput := len(thread.CommandOutput) > 0

	if !hasMessages && !hasOutput {
		return emptyStyle.Render("No messages yet. Waiting for activity...")
	}

	var sections []string

	for _, msg := range thread.Messages {
		label := formatRole(msg.Role)
		sections = append(sections, fmt.Sprintf("%s\n%s", label, msg.Content))
	}

	for execID, output := range thread.CommandOutput {
		header := commandStyle.Render(fmt.Sprintf("Command [%s]:", execID))
		body := outputStyle.Render(output)
		sections = append(sections, fmt.Sprintf("%s\n%s", header, body))
	}

	return strings.Join(sections, "\n\n")
}

func formatRole(role string) string {
	switch role {
	case "user":
		return userStyle.Render("You:")
	case "assistant":
		return assistantStyle.Render("Agent:")
	default:
		return lipgloss.NewStyle().Bold(true).Render(role + ":")
	}
}
```

**Step 4: Run tests**

Run: `go test ./internal/tui/ -v -run TestRenderMessages`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/session_render.go internal/tui/session_render_test.go
git commit -m "feat(tui): session message renderer with role formatting"
```

---

### Task 3: Build Session Pane with Viewport

**Files:**
- Create: `internal/tui/session.go`
- Create: `internal/tui/session_test.go`

**Step 1: Write tests for session pane**

```go
// internal/tui/session_test.go
package tui

import (
	"strings"
	"testing"

	"github.com/robinojw/dj/internal/state"
)

func TestSessionViewShowsThreadTitle(t *testing.T) {
	thread := state.NewThreadState("t-1", "My Task")
	thread.Status = state.StatusActive

	session := NewSessionModel(thread)
	output := session.View()

	if !strings.Contains(output, "My Task") {
		t.Errorf("expected thread title in output:\n%s", output)
	}
}

func TestSessionViewShowsMessages(t *testing.T) {
	thread := state.NewThreadState("t-1", "Test")
	thread.AppendMessage(state.ChatMessage{
		ID: "m-1", Role: "user", Content: "Hello world",
	})

	session := NewSessionModel(thread)
	session.SetSize(80, 24)
	output := session.View()

	if !strings.Contains(output, "Hello world") {
		t.Errorf("expected message content in output:\n%s", output)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -v -run TestSession`
Expected: FAIL — `NewSessionModel` not defined

**Step 3: Implement session pane**

```go
// internal/tui/session.go
package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/state"
)

const sessionHeaderHeight = 3

var sessionHeaderStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("39")).
	BorderStyle(lipgloss.NormalBorder()).
	BorderBottom(true).
	BorderForeground(lipgloss.Color("240"))

// SessionModel displays a thread's conversation in a scrollable pane.
type SessionModel struct {
	thread   *state.ThreadState
	viewport viewport.Model
	ready    bool
}

// NewSessionModel creates a session pane for the given thread.
func NewSessionModel(thread *state.ThreadState) SessionModel {
	return SessionModel{thread: thread}
}

// SetSize sets the viewport dimensions.
func (s *SessionModel) SetSize(width int, height int) {
	viewHeight := height - sessionHeaderHeight
	if viewHeight < 1 {
		viewHeight = 1
	}
	s.viewport = viewport.New(width, viewHeight)
	s.viewport.SetContent(RenderMessages(s.thread))
	s.ready = true
}

// Refresh re-renders the content from the thread state.
func (s *SessionModel) Refresh() {
	if s.ready {
		s.viewport.SetContent(RenderMessages(s.thread))
		s.viewport.GotoBottom()
	}
}

// View renders the session pane.
func (s SessionModel) View() string {
	header := sessionHeaderStyle.Render(
		fmt.Sprintf("%s [%s]", s.thread.Title, s.thread.Status),
	)

	if !s.ready {
		return header + "\n" + RenderMessages(s.thread)
	}

	return header + "\n" + s.viewport.View()
}
```

**Step 4: Run tests**

Run: `go test ./internal/tui/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/session.go internal/tui/session_test.go
git commit -m "feat(tui): scrollable session pane with viewport"
```

---

### Task 4: Wire Session into App Model

**Files:**
- Modify: `internal/tui/app.go`
- Modify: `internal/tui/app_test.go`

**Step 1: Write tests for Enter to open session and Esc to close**

Add to `app_test.go`:

```go
func TestAppEnterOpensSession(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Test Task")

	app := NewAppModel(store)

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	appModel := updated.(AppModel)

	if appModel.Focus() != FocusSession {
		t.Errorf("expected session focus, got %d", appModel.Focus())
	}
}

func TestAppEscClosesSession(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Test Task")

	app := NewAppModel(store)

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	appModel := updated.(AppModel)

	escKey := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ = appModel.Update(escKey)
	appModel = updated.(AppModel)

	if appModel.Focus() != FocusCanvas {
		t.Errorf("expected canvas focus after Esc, got %d", appModel.Focus())
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -v -run "TestAppEnter|TestAppEsc"`
Expected: FAIL — `FocusSession` not defined

**Step 3: Add FocusSession mode**

Add `FocusSession` to focus constants. Add `session *SessionModel` field to AppModel. In `handleKey`:
- Enter on canvas → create SessionModel for selected thread, set focus to FocusSession
- Esc in session → set focus back to FocusCanvas
- In session focus, route Up/Down to viewport scrolling

Update `View()` to render session pane when FocusSession is active.

**Step 4: Run tests**

Run: `go test ./internal/tui/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/app.go internal/tui/app_test.go
git commit -m "feat(tui): Enter opens session pane, Esc returns to canvas"
```
