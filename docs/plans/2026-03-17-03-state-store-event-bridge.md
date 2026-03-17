# Phase 3: State Store & Event Bridge

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build the reactive ThreadStore that holds all thread state, and the Bubble Tea event bridge that converts JSON-RPC notifications into `tea.Msg` types. The store is the single source of truth — the TUI never queries the app-server directly for display data.

**Architecture:** `ThreadStore` is a mutex-protected map of `threadID → ThreadState`. Each `ThreadState` holds metadata, messages, and command output. The event bridge is a set of `tea.Msg` types and a function that wires the `NotificationRouter` to call `program.Send(msg)` for each event. Bubble Tea's `Update` function then calls store mutations.

**Tech Stack:** Go, `sync.Mutex`, Bubble Tea `tea.Msg`

---

### Task 1: Define ThreadState Type

**Files:**
- Create: `internal/state/thread.go`
- Create: `internal/state/thread_test.go`

**Step 1: Write tests for ThreadState**

```go
// internal/state/thread_test.go
package state

import "testing"

func TestNewThreadState(t *testing.T) {
	thread := NewThreadState("t-1", "Build a web app")
	if thread.ID != "t-1" {
		t.Errorf("expected t-1, got %s", thread.ID)
	}
	if thread.Status != StatusIdle {
		t.Errorf("expected idle, got %s", thread.Status)
	}
	if len(thread.Messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(thread.Messages))
	}
}

func TestThreadStateAppendMessage(t *testing.T) {
	thread := NewThreadState("t-1", "Test")
	thread.AppendMessage(ChatMessage{
		ID:      "m-1",
		Role:    "user",
		Content: "Hello",
	})
	if len(thread.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(thread.Messages))
	}
	if thread.Messages[0].Content != "Hello" {
		t.Errorf("expected Hello, got %s", thread.Messages[0].Content)
	}
}

func TestThreadStateAppendDelta(t *testing.T) {
	thread := NewThreadState("t-1", "Test")
	thread.AppendMessage(ChatMessage{ID: "m-1", Role: "assistant", Content: "He"})
	thread.AppendDelta("m-1", "llo")

	if thread.Messages[0].Content != "Hello" {
		t.Errorf("expected Hello, got %s", thread.Messages[0].Content)
	}
}

func TestThreadStateAppendOutput(t *testing.T) {
	thread := NewThreadState("t-1", "Test")
	thread.AppendOutput("e-1", "line 1\n")
	thread.AppendOutput("e-1", "line 2\n")

	output := thread.CommandOutput["e-1"]
	if output != "line 1\nline 2\n" {
		t.Errorf("expected combined output, got %q", output)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/state/ -v`
Expected: FAIL — package not found

**Step 3: Implement ThreadState**

```go
// internal/state/thread.go
package state

// Thread status constants.
const (
	StatusIdle      = "idle"
	StatusActive    = "active"
	StatusCompleted = "completed"
	StatusError     = "error"
)

// ChatMessage is a single message in a thread's conversation.
type ChatMessage struct {
	ID      string
	Role    string
	Content string
}

// ThreadState holds all state for a single agent thread.
type ThreadState struct {
	ID            string
	Title         string
	Status        string
	Messages      []ChatMessage
	CommandOutput map[string]string // execId → accumulated output
}

// NewThreadState creates an idle thread with the given ID and title.
func NewThreadState(id string, title string) *ThreadState {
	return &ThreadState{
		ID:            id,
		Title:         title,
		Status:        StatusIdle,
		Messages:      make([]ChatMessage, 0),
		CommandOutput: make(map[string]string),
	}
}

// AppendMessage adds a new message to the thread.
func (ts *ThreadState) AppendMessage(msg ChatMessage) {
	ts.Messages = append(ts.Messages, msg)
}

// AppendDelta appends streaming text to an existing message.
func (ts *ThreadState) AppendDelta(messageID string, delta string) {
	for i := range ts.Messages {
		if ts.Messages[i].ID == messageID {
			ts.Messages[i].Content += delta
			return
		}
	}
}

// AppendOutput appends command output for the given exec ID.
func (ts *ThreadState) AppendOutput(execID string, data string) {
	ts.CommandOutput[execID] += data
}
```

**Step 4: Run tests**

Run: `go test ./internal/state/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/state/thread.go internal/state/thread_test.go
git commit -m "feat(state): define ThreadState with messages and command output"
```

---

### Task 2: Build ThreadStore

**Files:**
- Create: `internal/state/store.go`
- Create: `internal/state/store_test.go`

**Step 1: Write tests for ThreadStore**

```go
// internal/state/store_test.go
package state

import "testing"

func TestStoreAddAndGet(t *testing.T) {
	store := NewThreadStore()
	store.Add("t-1", "My Thread")

	thread, exists := store.Get("t-1")
	if !exists {
		t.Fatal("expected thread to exist")
	}
	if thread.Title != "My Thread" {
		t.Errorf("expected My Thread, got %s", thread.Title)
	}
}

func TestStoreGetMissing(t *testing.T) {
	store := NewThreadStore()
	_, exists := store.Get("missing")
	if exists {
		t.Error("expected thread to not exist")
	}
}

func TestStoreDelete(t *testing.T) {
	store := NewThreadStore()
	store.Add("t-1", "Test")
	store.Delete("t-1")

	_, exists := store.Get("t-1")
	if exists {
		t.Error("expected thread to be deleted")
	}
}

func TestStoreAll(t *testing.T) {
	store := NewThreadStore()
	store.Add("t-1", "First")
	store.Add("t-2", "Second")

	all := store.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 threads, got %d", len(all))
	}
}

func TestStoreUpdateStatus(t *testing.T) {
	store := NewThreadStore()
	store.Add("t-1", "Test")
	store.UpdateStatus("t-1", StatusActive, "Running")

	thread, _ := store.Get("t-1")
	if thread.Status != StatusActive {
		t.Errorf("expected active, got %s", thread.Status)
	}
	if thread.Title != "Running" {
		t.Errorf("expected Running, got %s", thread.Title)
	}
}

func TestStoreUpdateStatusMissing(t *testing.T) {
	store := NewThreadStore()
	store.UpdateStatus("missing", StatusActive, "Test")
}

func TestStoreIDs(t *testing.T) {
	store := NewThreadStore()
	store.Add("t-1", "First")
	store.Add("t-2", "Second")

	ids := store.IDs()
	if len(ids) != 2 {
		t.Fatalf("expected 2 ids, got %d", len(ids))
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/state/ -v -run TestStore`
Expected: FAIL — `NewThreadStore` not defined

**Step 3: Implement ThreadStore**

```go
// internal/state/store.go
package state

import (
	"sort"
	"sync"
)

// ThreadStore is the single source of truth for all thread state.
// All methods are safe for concurrent use.
type ThreadStore struct {
	mu      sync.RWMutex
	threads map[string]*ThreadState
	order   []string // insertion order for stable iteration
}

// NewThreadStore creates an empty store.
func NewThreadStore() *ThreadStore {
	return &ThreadStore{
		threads: make(map[string]*ThreadState),
		order:   make([]string, 0),
	}
}

// Add creates a new thread in the store.
func (s *ThreadStore) Add(id string, title string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.threads[id] = NewThreadState(id, title)
	s.order = append(s.order, id)
}

// Get returns a snapshot of the thread state. Returns false if not found.
func (s *ThreadStore) Get(id string) (*ThreadState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	thread, exists := s.threads[id]
	return thread, exists
}

// Delete removes a thread from the store.
func (s *ThreadStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.threads, id)
	s.order = removeFromSlice(s.order, id)
}

// All returns all threads in insertion order.
func (s *ThreadStore) All() []*ThreadState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*ThreadState, 0, len(s.order))
	for _, id := range s.order {
		if thread, exists := s.threads[id]; exists {
			result = append(result, thread)
		}
	}
	return result
}

// IDs returns all thread IDs in sorted order.
func (s *ThreadStore) IDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.threads))
	for id := range s.threads {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// UpdateStatus updates a thread's status and title.
func (s *ThreadStore) UpdateStatus(id string, status string, title string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	thread, exists := s.threads[id]
	if !exists {
		return
	}
	thread.Status = status
	if title != "" {
		thread.Title = title
	}
}

func removeFromSlice(slice []string, target string) []string {
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		if item != target {
			result = append(result, item)
		}
	}
	return result
}
```

**Step 4: Run tests**

Run: `go test ./internal/state/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/state/store.go internal/state/store_test.go
git commit -m "feat(state): ThreadStore with concurrent-safe CRUD"
```

---

### Task 3: Define Bubble Tea Message Types

**Files:**
- Create: `internal/tui/msgs.go`
- Create: `internal/tui/msgs_test.go`

**Step 1: Write tests for message types**

```go
// internal/tui/msgs_test.go
package tui

import "testing"

func TestMsgTypes(t *testing.T) {
	statusMsg := ThreadStatusMsg{
		ThreadID: "t-1",
		Status:   "active",
		Title:    "Running",
	}
	if statusMsg.ThreadID != "t-1" {
		t.Errorf("expected t-1, got %s", statusMsg.ThreadID)
	}

	messageMsg := ThreadMessageMsg{
		ThreadID:  "t-1",
		MessageID: "m-1",
		Role:      "assistant",
		Content:   "Hello",
	}
	if messageMsg.Role != "assistant" {
		t.Errorf("expected assistant, got %s", messageMsg.Role)
	}

	deltaMsg := ThreadDeltaMsg{
		ThreadID:  "t-1",
		MessageID: "m-1",
		Delta:     "world",
	}
	if deltaMsg.Delta != "world" {
		t.Errorf("expected world, got %s", deltaMsg.Delta)
	}

	outputMsg := CommandOutputMsg{
		ThreadID: "t-1",
		ExecID:   "e-1",
		Data:     "output\n",
	}
	if outputMsg.Data != "output\n" {
		t.Errorf("expected output, got %s", outputMsg.Data)
	}

	finishedMsg := CommandFinishedMsg{
		ThreadID: "t-1",
		ExecID:   "e-1",
		ExitCode: 0,
	}
	if finishedMsg.ExitCode != 0 {
		t.Errorf("expected 0, got %d", finishedMsg.ExitCode)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -v -run TestMsgTypes`
Expected: FAIL — types not defined

**Step 3: Implement message types**

```go
// internal/tui/msgs.go
package tui

// ThreadStatusMsg is sent when a thread's status changes.
type ThreadStatusMsg struct {
	ThreadID string
	Status   string
	Title    string
}

// ThreadMessageMsg is sent when a new message is created in a thread.
type ThreadMessageMsg struct {
	ThreadID  string
	MessageID string
	Role      string
	Content   string
}

// ThreadDeltaMsg is sent for streaming message deltas.
type ThreadDeltaMsg struct {
	ThreadID  string
	MessageID string
	Delta     string
}

// CommandOutputMsg is sent when command output arrives.
type CommandOutputMsg struct {
	ThreadID string
	ExecID   string
	Data     string
}

// CommandFinishedMsg is sent when a command finishes execution.
type CommandFinishedMsg struct {
	ThreadID string
	ExecID   string
	ExitCode int
}

// ThreadCreatedMsg is sent when a new thread is created.
type ThreadCreatedMsg struct {
	ThreadID string
	Title    string
}

// ThreadDeletedMsg is sent when a thread is deleted.
type ThreadDeletedMsg struct {
	ThreadID string
}

// AppServerErrorMsg is sent when the app-server connection has an error.
type AppServerErrorMsg struct {
	Err error
}

func (m AppServerErrorMsg) Error() string {
	return m.Err.Error()
}
```

**Step 4: Run tests**

Run: `go test ./internal/tui/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/msgs.go internal/tui/msgs_test.go
git commit -m "feat(tui): define Bubble Tea message types for event bridge"
```

---

### Task 4: Build Event Bridge

**Files:**
- Create: `internal/tui/bridge.go`
- Create: `internal/tui/bridge_test.go`

**Step 1: Write tests for event bridge**

```go
// internal/tui/bridge_test.go
package tui

import (
	"testing"

	"github.com/robinojw/dj/internal/appserver"
)

type mockSender struct {
	messages []any
}

func (m *mockSender) Send(msg any) {
	m.messages = append(m.messages, msg)
}

func TestBridgeThreadStatusChanged(t *testing.T) {
	sender := &mockSender{}
	router := appserver.NewNotificationRouter()
	WireEventBridge(router, sender)

	router.Handle(appserver.NotifyThreadStatusChanged,
		[]byte(`{"threadId":"t-1","status":"active","title":"Running"}`))

	if len(sender.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(sender.messages))
	}
	msg, ok := sender.messages[0].(ThreadStatusMsg)
	if !ok {
		t.Fatalf("expected ThreadStatusMsg, got %T", sender.messages[0])
	}
	if msg.ThreadID != "t-1" {
		t.Errorf("expected t-1, got %s", msg.ThreadID)
	}
}

func TestBridgeCommandOutput(t *testing.T) {
	sender := &mockSender{}
	router := appserver.NewNotificationRouter()
	WireEventBridge(router, sender)

	router.Handle(appserver.NotifyCommandOutput,
		[]byte(`{"threadId":"t-1","execId":"e-1","data":"hello\n"}`))

	if len(sender.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(sender.messages))
	}
	msg, ok := sender.messages[0].(CommandOutputMsg)
	if !ok {
		t.Fatalf("expected CommandOutputMsg, got %T", sender.messages[0])
	}
	if msg.Data != "hello\n" {
		t.Errorf("expected hello, got %s", msg.Data)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -v -run TestBridge`
Expected: FAIL — `WireEventBridge` not defined

**Step 3: Implement event bridge**

```go
// internal/tui/bridge.go
package tui

import "github.com/robinojw/dj/internal/appserver"

// MessageSender is an interface for sending Bubble Tea messages.
// In production this is *tea.Program; in tests it's a mock.
type MessageSender interface {
	Send(msg any)
}

// WireEventBridge registers notification handlers on the router
// that convert appserver notifications into Bubble Tea messages.
func WireEventBridge(router *appserver.NotificationRouter, sender MessageSender) {
	router.OnThreadStatusChanged(func(params appserver.ThreadStatusChanged) {
		sender.Send(ThreadStatusMsg{
			ThreadID: params.ThreadID,
			Status:   params.Status,
			Title:    params.Title,
		})
	})

	router.OnThreadMessageCreated(func(params appserver.ThreadMessageCreated) {
		sender.Send(ThreadMessageMsg{
			ThreadID:  params.ThreadID,
			MessageID: params.MessageID,
			Role:      params.Role,
			Content:   params.Content,
		})
	})

	router.OnThreadMessageDelta(func(params appserver.ThreadMessageDelta) {
		sender.Send(ThreadDeltaMsg{
			ThreadID:  params.ThreadID,
			MessageID: params.MessageID,
			Delta:     params.Delta,
		})
	})

	router.OnCommandOutput(func(params appserver.CommandOutput) {
		sender.Send(CommandOutputMsg{
			ThreadID: params.ThreadID,
			ExecID:   params.ExecID,
			Data:     params.Data,
		})
	})

	router.OnCommandFinished(func(params appserver.CommandFinished) {
		sender.Send(CommandFinishedMsg{
			ThreadID: params.ThreadID,
			ExecID:   params.ExecID,
			ExitCode: params.ExitCode,
		})
	})
}
```

**Step 4: Run tests**

Run: `go test ./internal/tui/ -v`
Expected: PASS

> **Note:** This task introduces the first cross-package dependency. Run `go mod tidy` after if needed, but since it's all within the module no external deps are added.

**Step 5: Commit**

```bash
git add internal/tui/bridge.go internal/tui/bridge_test.go
git commit -m "feat(tui): event bridge wires appserver notifications to Bubble Tea"
```
