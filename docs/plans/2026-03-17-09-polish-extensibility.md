# Phase 9: Polish & Extensibility

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add error handling for app-server disconnection, status bar with connection state, graceful shutdown, and wire everything end-to-end so `./dj` spawns the app-server, performs the handshake, and renders live threads.

**Architecture:** The App model gains lifecycle management: spawn app-server on Init, run ReadLoop in a goroutine, handle disconnection gracefully. A status bar component shows connection state and selected thread info. Error messages are displayed inline. The `n` key creates a new thread via the API.

**Tech Stack:** Go, Bubble Tea, Lipgloss

**Prerequisites:** Phase 8 (config, cobra CLI, all TUI components)

---

### Task 1: Build Status Bar Component

**Files:**
- Create: `internal/tui/statusbar.go`
- Create: `internal/tui/statusbar_test.go`

**Step 1: Write tests for status bar**

```go
// internal/tui/statusbar_test.go
package tui

import (
	"strings"
	"testing"
)

func TestStatusBarConnected(t *testing.T) {
	bar := NewStatusBar()
	bar.SetConnected(true)
	bar.SetThreadCount(3)
	bar.SetSelectedThread("Build web app")

	output := bar.View()

	if !strings.Contains(output, "Connected") {
		t.Errorf("expected Connected in output:\n%s", output)
	}
	if !strings.Contains(output, "3 threads") {
		t.Errorf("expected thread count in output:\n%s", output)
	}
	if !strings.Contains(output, "Build web app") {
		t.Errorf("expected selected thread in output:\n%s", output)
	}
}

func TestStatusBarDisconnected(t *testing.T) {
	bar := NewStatusBar()
	bar.SetConnected(false)

	output := bar.View()

	if !strings.Contains(output, "Disconnected") {
		t.Errorf("expected Disconnected in output:\n%s", output)
	}
}

func TestStatusBarError(t *testing.T) {
	bar := NewStatusBar()
	bar.SetError("connection lost")

	output := bar.View()

	if !strings.Contains(output, "connection lost") {
		t.Errorf("expected error in output:\n%s", output)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -v -run TestStatusBar`
Expected: FAIL — `NewStatusBar` not defined

**Step 3: Implement status bar**

```go
// internal/tui/statusbar.go
package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1)
	statusConnectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("42"))
	statusDisconnectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196"))
	statusErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true)
)

// StatusBar displays connection state and context info.
type StatusBar struct {
	connected      bool
	threadCount    int
	selectedThread string
	errorMessage   string
	width          int
}

// NewStatusBar creates a status bar.
func NewStatusBar() *StatusBar {
	return &StatusBar{}
}

// SetConnected updates the connection state.
func (s *StatusBar) SetConnected(connected bool) {
	s.connected = connected
	if connected {
		s.errorMessage = ""
	}
}

// SetThreadCount updates the thread count display.
func (s *StatusBar) SetThreadCount(count int) {
	s.threadCount = count
}

// SetSelectedThread updates the selected thread name.
func (s *StatusBar) SetSelectedThread(name string) {
	s.selectedThread = name
}

// SetError sets an error message.
func (s *StatusBar) SetError(msg string) {
	s.errorMessage = msg
}

// SetWidth sets the status bar width.
func (s *StatusBar) SetWidth(width int) {
	s.width = width
}

// View renders the status bar.
func (s StatusBar) View() string {
	var left string
	if s.connected {
		left = statusConnectedStyle.Render("● Connected")
	} else {
		left = statusDisconnectedStyle.Render("○ Disconnected")
	}

	if s.errorMessage != "" {
		left += " " + statusErrorStyle.Render(s.errorMessage)
	}

	middle := ""
	if s.threadCount > 0 {
		middle = fmt.Sprintf(" | %d threads", s.threadCount)
	}

	right := ""
	if s.selectedThread != "" {
		right = fmt.Sprintf(" | %s", s.selectedThread)
	}

	content := left + middle + right
	style := statusBarStyle.Width(s.width)
	return style.Render(content)
}
```

**Step 4: Run tests**

Run: `go test ./internal/tui/ -v -run TestStatusBar`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/statusbar.go internal/tui/statusbar_test.go
git commit -m "feat(tui): status bar with connection state"
```

---

### Task 2: Add App-Server Lifecycle to App Model

**Files:**
- Modify: `internal/tui/app.go`
- Modify: `cmd/dj/main.go`

**Step 1: Wire app-server spawn into Init**

Update `AppModel` to accept an `*appserver.Client` and `*config.Config`. The `Init()` command should:
1. Spawn the app-server process
2. Start the ReadLoop goroutine
3. Perform the Initialize handshake
4. Return the server capabilities as a message

```go
// AppServerConnectedMsg is sent after successful handshake.
type AppServerConnectedMsg struct {
	ServerName    string
	ServerVersion string
}
```

In `Init()`:

```go
func (m AppModel) Init() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := m.client.Start(ctx); err != nil {
			return AppServerErrorMsg{Err: err}
		}

		router := appserver.NewNotificationRouter()
		m.client.Router = router
		WireEventBridge(router, m.program)

		go m.client.ReadLoop(m.client.Dispatch)

		caps, err := m.client.Initialize(ctx)
		if err != nil {
			return AppServerErrorMsg{Err: err}
		}

		return AppServerConnectedMsg{
			ServerName:    caps.ServerInfo.Name,
			ServerVersion: caps.ServerInfo.Version,
		}
	}
}
```

**Step 2: Handle connection messages in Update**

In `Update`, handle `AppServerConnectedMsg` to update the status bar and `AppServerErrorMsg` to display errors.

**Step 3: Update main.go to pass client and config**

```go
func runApp(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	client := appserver.NewClient(cfg.AppServer.Command, cfg.AppServer.Args...)
	store := state.NewThreadStore()
	app := tui.NewAppModel(store, client)

	program := tea.NewProgram(app, tea.WithAltScreen())
	app.SetProgram(program)

	_, err = program.Run()

	client.Stop()
	return err
}
```

**Step 4: Verify it builds**

Run: `go build ./cmd/dj && echo "Build OK"`
Expected: Build OK

**Step 5: Commit**

```bash
git add internal/tui/app.go cmd/dj/main.go
git commit -m "feat: wire app-server lifecycle into TUI startup"
```

---

### Task 3: Add New Thread Creation

**Files:**
- Modify: `internal/tui/app.go`
- Modify: `internal/tui/app_test.go`

**Step 1: Write test for 'n' key creating thread**

Add to `app_test.go`:

```go
func TestAppNewThread(t *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	nKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	_, cmd := app.Update(nKey)

	if cmd == nil {
		t.Error("expected command for thread creation")
	}
}
```

**Step 2: Implement**

In `handleKey`, when `'n'` is pressed, return a `tea.Cmd` that calls `client.CreateThread()` and returns a `ThreadCreatedMsg`. In `Update`, handle `ThreadCreatedMsg` by adding the thread to the store.

**Step 3: Run tests**

Run: `go test ./internal/tui/ -v`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/tui/app.go internal/tui/app_test.go
git commit -m "feat(tui): 'n' key creates new thread via app-server"
```

---

### Task 4: Graceful Shutdown

**Files:**
- Modify: `internal/tui/app.go`

**Step 1: Add cleanup on quit**

When Ctrl+C or quit is triggered, stop the app-server client before exiting. Use a `tea.Sequence` or handle cleanup in `main.go` after `program.Run()` returns.

The app-server `Stop()` is already called in `main.go`'s defer. Verify the ReadLoop goroutine exits cleanly when stdin is closed (it already does — scanner returns false on EOF).

**Step 2: Verify clean shutdown**

Run: `go build ./cmd/dj && echo "Build OK"`
Expected: Build OK. Manual test: launch, Ctrl+C exits cleanly.

**Step 3: Commit**

```bash
git add internal/tui/app.go
git commit -m "feat: graceful shutdown with app-server cleanup"
```

---

### Task 5: End-to-End Integration Test

**Files:**
- Create: `internal/tui/integration_test.go`

**Step 1: Write build-tagged integration test**

```go
//go:build integration

package tui

import (
	"context"
	"testing"
	"time"

	"github.com/robinojw/dj/internal/appserver"
	"github.com/robinojw/dj/internal/state"
)

func TestIntegrationEndToEnd(t *testing.T) {
	client := appserver.NewClient("codex", "app-server", "--listen", "stdio://")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := client.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer client.Stop()

	router := appserver.NewNotificationRouter()
	client.Router = router
	go client.ReadLoop(client.Dispatch)

	caps, err := client.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	t.Logf("Connected: %s %s", caps.ServerInfo.Name, caps.ServerInfo.Version)

	store := state.NewThreadStore()

	result, err := client.CreateThread(ctx, "Say hello")
	if err != nil {
		t.Fatalf("CreateThread failed: %v", err)
	}
	store.Add(result.ThreadID, "Say hello")
	t.Logf("Created thread: %s", result.ThreadID)

	threads := store.All()
	if len(threads) != 1 {
		t.Fatalf("expected 1 thread, got %d", len(threads))
	}
}
```

**Step 2: Verify it compiles**

Run: `go vet -tags=integration ./internal/tui/`
Expected: no errors

**Step 3: Commit**

```bash
git add internal/tui/integration_test.go
git commit -m "test(tui): end-to-end integration test with real app-server"
```
