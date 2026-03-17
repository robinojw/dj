# Phase 2: Protocol Types & Enhanced Dispatch

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Define the Codex App Server-specific protocol types for thread management, message streaming, and command execution. Enhance dispatch to route typed notifications and server requests.

**Architecture:** Typed request params and result structs wrap `json.RawMessage` for each RPC method. A `NotificationRouter` maps method names to typed handler functions, replacing the generic `OnNotification` callback. Notification types are defined as constants to prevent string repetition.

**Tech Stack:** Go, `encoding/json`

---

### Task 1: Define Protocol Method Constants

**Files:**
- Create: `internal/appserver/methods.go`
- Create: `internal/appserver/methods_test.go`

**Step 1: Write test that method constants match expected strings**

```go
// internal/appserver/methods_test.go
package appserver

import "testing"

func TestMethodConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"ThreadCreate", MethodThreadCreate, "thread/create"},
		{"ThreadList", MethodThreadList, "thread/list"},
		{"ThreadDelete", MethodThreadDelete, "thread/delete"},
		{"ThreadSendMessage", MethodThreadSendMessage, "thread/sendMessage"},
		{"CommandExec", MethodCommandExec, "command/exec"},
		{"NotifyThreadStatus", NotifyThreadStatusChanged, "thread/status/changed"},
		{"NotifyThreadMessage", NotifyThreadMessageCreated, "thread/message/created"},
		{"NotifyMessageDelta", NotifyThreadMessageDelta, "thread/message/delta"},
		{"NotifyCommandOutput", NotifyCommandOutput, "command/output"},
		{"NotifyCommandFinished", NotifyCommandFinished, "command/finished"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.constant)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/appserver/ -v -run TestMethodConstants`
Expected: FAIL — constants not defined

**Step 3: Implement method constants**

```go
// internal/appserver/methods.go
package appserver

// Client-to-server request methods.
const (
	MethodThreadCreate      = "thread/create"
	MethodThreadList        = "thread/list"
	MethodThreadDelete      = "thread/delete"
	MethodThreadSendMessage = "thread/sendMessage"
	MethodCommandExec       = "command/exec"
)

// Server-to-client notification methods.
const (
	NotifyThreadStatusChanged  = "thread/status/changed"
	NotifyThreadMessageCreated = "thread/message/created"
	NotifyThreadMessageDelta   = "thread/message/delta"
	NotifyCommandOutput        = "command/output"
	NotifyCommandFinished      = "command/finished"
)
```

**Step 4: Run tests**

Run: `go test ./internal/appserver/ -v -run TestMethodConstants`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/appserver/methods.go internal/appserver/methods_test.go
git commit -m "feat(appserver): define protocol method constants"
```

---

### Task 2: Define Thread Management Types

**Files:**
- Create: `internal/appserver/types_thread.go`
- Create: `internal/appserver/types_thread_test.go`

**Step 1: Write tests for thread type marshaling**

```go
// internal/appserver/types_thread_test.go
package appserver

import (
	"encoding/json"
	"testing"
)

func TestThreadCreateParamsMarshal(t *testing.T) {
	params := ThreadCreateParams{
		Instructions: "Build a web server",
	}
	data, err := json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	json.Unmarshal(data, &parsed)
	if parsed["instructions"] != "Build a web server" {
		t.Errorf("expected instructions, got %v", parsed["instructions"])
	}
}

func TestThreadCreateResultUnmarshal(t *testing.T) {
	raw := `{"threadId":"t-abc123"}`
	var result ThreadCreateResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		t.Fatal(err)
	}
	if result.ThreadID != "t-abc123" {
		t.Errorf("expected t-abc123, got %s", result.ThreadID)
	}
}

func TestThreadListResultUnmarshal(t *testing.T) {
	raw := `{"threads":[{"id":"t-1","status":"active","title":"Test"}]}`
	var result ThreadListResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Threads) != 1 {
		t.Fatalf("expected 1 thread, got %d", len(result.Threads))
	}
	if result.Threads[0].ID != "t-1" {
		t.Errorf("expected id t-1, got %s", result.Threads[0].ID)
	}
	if result.Threads[0].Status != "active" {
		t.Errorf("expected status active, got %s", result.Threads[0].Status)
	}
}

func TestThreadStatusValues(t *testing.T) {
	if ThreadStatusActive != "active" {
		t.Errorf("expected active, got %s", ThreadStatusActive)
	}
	if ThreadStatusCompleted != "completed" {
		t.Errorf("expected completed, got %s", ThreadStatusCompleted)
	}
	if ThreadStatusError != "error" {
		t.Errorf("expected error, got %s", ThreadStatusError)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/appserver/ -v -run TestThread`
Expected: FAIL — types not defined

**Step 3: Implement thread types**

```go
// internal/appserver/types_thread.go
package appserver

// Thread status constants.
const (
	ThreadStatusActive    = "active"
	ThreadStatusIdle      = "idle"
	ThreadStatusCompleted = "completed"
	ThreadStatusError     = "error"
)

// ThreadCreateParams is the params for thread/create.
type ThreadCreateParams struct {
	Instructions string `json:"instructions"`
}

// ThreadCreateResult is the result of thread/create.
type ThreadCreateResult struct {
	ThreadID string `json:"threadId"`
}

// ThreadDeleteParams is the params for thread/delete.
type ThreadDeleteParams struct {
	ThreadID string `json:"threadId"`
}

// ThreadListResult is the result of thread/list.
type ThreadListResult struct {
	Threads []ThreadSummary `json:"threads"`
}

// ThreadSummary is a thread entry in the thread/list result.
type ThreadSummary struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Title  string `json:"title"`
}

// ThreadSendMessageParams is the params for thread/sendMessage.
type ThreadSendMessageParams struct {
	ThreadID string `json:"threadId"`
	Content  string `json:"content"`
}

// ThreadSendMessageResult is the result of thread/sendMessage.
type ThreadSendMessageResult struct {
	MessageID string `json:"messageId"`
}
```

**Step 4: Run tests**

Run: `go test ./internal/appserver/ -v -run TestThread`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/appserver/types_thread.go internal/appserver/types_thread_test.go
git commit -m "feat(appserver): define thread management protocol types"
```

---

### Task 3: Define Notification Types

**Files:**
- Create: `internal/appserver/types_notify.go`
- Create: `internal/appserver/types_notify_test.go`

**Step 1: Write tests for notification type unmarshaling**

```go
// internal/appserver/types_notify_test.go
package appserver

import (
	"encoding/json"
	"testing"
)

func TestThreadStatusChangedUnmarshal(t *testing.T) {
	raw := `{"threadId":"t-1","status":"completed","title":"Done"}`
	var params ThreadStatusChanged
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatal(err)
	}
	if params.ThreadID != "t-1" {
		t.Errorf("expected t-1, got %s", params.ThreadID)
	}
	if params.Status != ThreadStatusCompleted {
		t.Errorf("expected completed, got %s", params.Status)
	}
}

func TestThreadMessageCreatedUnmarshal(t *testing.T) {
	raw := `{"threadId":"t-1","messageId":"m-1","role":"assistant","content":"Hello"}`
	var params ThreadMessageCreated
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatal(err)
	}
	if params.Role != "assistant" {
		t.Errorf("expected assistant, got %s", params.Role)
	}
	if params.Content != "Hello" {
		t.Errorf("expected Hello, got %s", params.Content)
	}
}

func TestThreadMessageDeltaUnmarshal(t *testing.T) {
	raw := `{"threadId":"t-1","messageId":"m-1","delta":"more text"}`
	var params ThreadMessageDelta
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatal(err)
	}
	if params.Delta != "more text" {
		t.Errorf("expected 'more text', got %s", params.Delta)
	}
}

func TestCommandOutputUnmarshal(t *testing.T) {
	raw := `{"threadId":"t-1","execId":"e-1","data":"line of output\n"}`
	var params CommandOutput
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatal(err)
	}
	if params.ExecID != "e-1" {
		t.Errorf("expected e-1, got %s", params.ExecID)
	}
}

func TestCommandFinishedUnmarshal(t *testing.T) {
	raw := `{"threadId":"t-1","execId":"e-1","exitCode":0}`
	var params CommandFinished
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatal(err)
	}
	if params.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", params.ExitCode)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/appserver/ -v -run "TestThread(Status|Message)|TestCommand"`
Expected: FAIL — types not defined

**Step 3: Implement notification types**

```go
// internal/appserver/types_notify.go
package appserver

// ThreadStatusChanged is the params for thread/status/changed notifications.
type ThreadStatusChanged struct {
	ThreadID string `json:"threadId"`
	Status   string `json:"status"`
	Title    string `json:"title"`
}

// ThreadMessageCreated is the params for thread/message/created notifications.
type ThreadMessageCreated struct {
	ThreadID  string `json:"threadId"`
	MessageID string `json:"messageId"`
	Role      string `json:"role"`
	Content   string `json:"content"`
}

// ThreadMessageDelta is the params for thread/message/delta notifications.
type ThreadMessageDelta struct {
	ThreadID  string `json:"threadId"`
	MessageID string `json:"messageId"`
	Delta     string `json:"delta"`
}

// CommandOutput is the params for command/output notifications.
type CommandOutput struct {
	ThreadID string `json:"threadId"`
	ExecID   string `json:"execId"`
	Data     string `json:"data"`
}

// CommandFinished is the params for command/finished notifications.
type CommandFinished struct {
	ThreadID string `json:"threadId"`
	ExecID   string `json:"execId"`
	ExitCode int    `json:"exitCode"`
}
```

**Step 4: Run tests**

Run: `go test ./internal/appserver/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/appserver/types_notify.go internal/appserver/types_notify_test.go
git commit -m "feat(appserver): define notification protocol types"
```

---

### Task 4: Define Command Execution Types

**Files:**
- Create: `internal/appserver/types_command.go`
- Create: `internal/appserver/types_command_test.go`

**Step 1: Write tests for command types**

```go
// internal/appserver/types_command_test.go
package appserver

import (
	"encoding/json"
	"testing"
)

func TestCommandExecParamsMarshal(t *testing.T) {
	params := CommandExecParams{
		ThreadID: "t-1",
		Command:  "go test ./...",
		TTY:      true,
	}
	data, err := json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	json.Unmarshal(data, &parsed)
	if parsed["tty"] != true {
		t.Errorf("expected tty true, got %v", parsed["tty"])
	}
}

func TestCommandExecResultUnmarshal(t *testing.T) {
	raw := `{"execId":"e-abc123"}`
	var result CommandExecResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		t.Fatal(err)
	}
	if result.ExecID != "e-abc123" {
		t.Errorf("expected e-abc123, got %s", result.ExecID)
	}
}

func TestConfirmExecParamsUnmarshal(t *testing.T) {
	raw := `{"threadId":"t-1","command":"rm -rf /tmp/test"}`
	var params ConfirmExecParams
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatal(err)
	}
	if params.ThreadID != "t-1" {
		t.Errorf("expected t-1, got %s", params.ThreadID)
	}
	if params.Command != "rm -rf /tmp/test" {
		t.Errorf("expected command, got %s", params.Command)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/appserver/ -v -run "TestCommandExec|TestConfirmExec"`
Expected: FAIL — types not defined

**Step 3: Implement command types**

```go
// internal/appserver/types_command.go
package appserver

// CommandExecParams is the params for command/exec.
type CommandExecParams struct {
	ThreadID string `json:"threadId"`
	Command  string `json:"command"`
	TTY      bool   `json:"tty"`
}

// CommandExecResult is the result of command/exec.
type CommandExecResult struct {
	ExecID string `json:"execId"`
}

// ConfirmExecParams is a server-to-client request asking the user
// to confirm a command before execution.
type ConfirmExecParams struct {
	ThreadID string `json:"threadId"`
	Command  string `json:"command"`
}

// ConfirmExecResult is the client's response to a confirm exec request.
type ConfirmExecResult struct {
	Approved bool `json:"approved"`
}
```

**Step 4: Run tests**

Run: `go test ./internal/appserver/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/appserver/types_command.go internal/appserver/types_command_test.go
git commit -m "feat(appserver): define command execution protocol types"
```

---

### Task 5: Build NotificationRouter

**Files:**
- Create: `internal/appserver/router.go`
- Create: `internal/appserver/router_test.go`

**Step 1: Write tests for notification routing**

```go
// internal/appserver/router_test.go
package appserver

import (
	"encoding/json"
	"sync/atomic"
	"testing"
)

func TestRouterDispatchesNotification(t *testing.T) {
	router := NewNotificationRouter()

	var called atomic.Bool
	router.OnThreadStatusChanged(func(params ThreadStatusChanged) {
		called.Store(true)
		if params.ThreadID != "t-1" {
			t.Errorf("expected t-1, got %s", params.ThreadID)
		}
	})

	raw := json.RawMessage(`{"threadId":"t-1","status":"active","title":"Test"}`)
	router.Handle(NotifyThreadStatusChanged, raw)

	if !called.Load() {
		t.Error("handler was not called")
	}
}

func TestRouterIgnoresUnregisteredMethod(t *testing.T) {
	router := NewNotificationRouter()
	router.Handle("unknown/method", json.RawMessage(`{}`))
}

func TestRouterDispatchesMessageDelta(t *testing.T) {
	router := NewNotificationRouter()

	var receivedDelta string
	router.OnThreadMessageDelta(func(params ThreadMessageDelta) {
		receivedDelta = params.Delta
	})

	raw := json.RawMessage(`{"threadId":"t-1","messageId":"m-1","delta":"hello"}`)
	router.Handle(NotifyThreadMessageDelta, raw)

	if receivedDelta != "hello" {
		t.Errorf("expected hello, got %s", receivedDelta)
	}
}

func TestRouterDispatchesCommandOutput(t *testing.T) {
	router := NewNotificationRouter()

	var receivedData string
	router.OnCommandOutput(func(params CommandOutput) {
		receivedData = params.Data
	})

	raw := json.RawMessage(`{"threadId":"t-1","execId":"e-1","data":"output line\n"}`)
	router.Handle(NotifyCommandOutput, raw)

	if receivedData != "output line\n" {
		t.Errorf("expected output, got %s", receivedData)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/appserver/ -v -run TestRouter`
Expected: FAIL — `NewNotificationRouter` not defined

**Step 3: Implement NotificationRouter**

```go
// internal/appserver/router.go
package appserver

import "encoding/json"

// NotificationRouter dispatches typed notifications by method name.
type NotificationRouter struct {
	handlers map[string]func(json.RawMessage)
}

// NewNotificationRouter creates an empty router.
func NewNotificationRouter() *NotificationRouter {
	return &NotificationRouter{
		handlers: make(map[string]func(json.RawMessage)),
	}
}

// Handle dispatches a notification to its registered handler.
func (r *NotificationRouter) Handle(method string, params json.RawMessage) {
	handler, exists := r.handlers[method]
	if !exists {
		return
	}
	handler(params)
}

// OnThreadStatusChanged registers a handler for thread/status/changed.
func (r *NotificationRouter) OnThreadStatusChanged(fn func(ThreadStatusChanged)) {
	r.handlers[NotifyThreadStatusChanged] = func(raw json.RawMessage) {
		var params ThreadStatusChanged
		if err := json.Unmarshal(raw, &params); err != nil {
			return
		}
		fn(params)
	}
}

// OnThreadMessageCreated registers a handler for thread/message/created.
func (r *NotificationRouter) OnThreadMessageCreated(fn func(ThreadMessageCreated)) {
	r.handlers[NotifyThreadMessageCreated] = func(raw json.RawMessage) {
		var params ThreadMessageCreated
		if err := json.Unmarshal(raw, &params); err != nil {
			return
		}
		fn(params)
	}
}

// OnThreadMessageDelta registers a handler for thread/message/delta.
func (r *NotificationRouter) OnThreadMessageDelta(fn func(ThreadMessageDelta)) {
	r.handlers[NotifyThreadMessageDelta] = func(raw json.RawMessage) {
		var params ThreadMessageDelta
		if err := json.Unmarshal(raw, &params); err != nil {
			return
		}
		fn(params)
	}
}

// OnCommandOutput registers a handler for command/output.
func (r *NotificationRouter) OnCommandOutput(fn func(CommandOutput)) {
	r.handlers[NotifyCommandOutput] = func(raw json.RawMessage) {
		var params CommandOutput
		if err := json.Unmarshal(raw, &params); err != nil {
			return
		}
		fn(params)
	}
}

// OnCommandFinished registers a handler for command/finished.
func (r *NotificationRouter) OnCommandFinished(fn func(CommandFinished)) {
	r.handlers[NotifyCommandFinished] = func(raw json.RawMessage) {
		var params CommandFinished
		if err := json.Unmarshal(raw, &params); err != nil {
			return
		}
		fn(params)
	}
}
```

**Step 4: Run tests**

Run: `go test ./internal/appserver/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/appserver/router.go internal/appserver/router_test.go
git commit -m "feat(appserver): typed notification router"
```

---

### Task 6: Typed Client Helper Methods

**Files:**
- Create: `internal/appserver/client_thread.go`
- Create: `internal/appserver/client_thread_test.go`

**Step 1: Write tests for typed thread helpers**

```go
// internal/appserver/client_thread_test.go
package appserver

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"
)

func TestClientCreateThread(t *testing.T) {
	clientRead, serverWrite := io.Pipe()
	serverRead, clientWrite := io.Pipe()

	go mockThreadCreateServer(t, serverRead, serverWrite)

	client := &Client{}
	client.stdin = clientWrite
	client.stdout = clientRead
	client.scanner = bufio.NewScanner(clientRead)
	client.scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	client.running.Store(true)

	go client.ReadLoop(client.Dispatch)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := client.CreateThread(ctx, "Build a web app")
	if err != nil {
		t.Fatalf("CreateThread failed: %v", err)
	}
	if result.ThreadID != "t-new-123" {
		t.Errorf("expected t-new-123, got %s", result.ThreadID)
	}
}

func mockThreadCreateServer(t *testing.T, reader *io.PipeReader, writer *io.PipeWriter) {
	t.Helper()
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	if !scanner.Scan() {
		t.Error("mock: failed to read request")
		return
	}
	var req Message
	if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
		t.Errorf("mock: unmarshal: %v", err)
		return
	}

	resp := Message{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  json.RawMessage(`{"threadId":"t-new-123"}`),
	}
	data, _ := json.Marshal(resp)
	data = append(data, '\n')
	writer.Write(data)
}

func TestClientListThreads(t *testing.T) {
	clientRead, serverWrite := io.Pipe()
	serverRead, clientWrite := io.Pipe()

	go mockThreadListServer(t, serverRead, serverWrite)

	client := &Client{}
	client.stdin = clientWrite
	client.stdout = clientRead
	client.scanner = bufio.NewScanner(clientRead)
	client.scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	client.running.Store(true)

	go client.ReadLoop(client.Dispatch)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := client.ListThreads(ctx)
	if err != nil {
		t.Fatalf("ListThreads failed: %v", err)
	}
	if len(result.Threads) != 2 {
		t.Fatalf("expected 2 threads, got %d", len(result.Threads))
	}
}

func mockThreadListServer(t *testing.T, reader *io.PipeReader, writer *io.PipeWriter) {
	t.Helper()
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	if !scanner.Scan() {
		t.Error("mock: failed to read request")
		return
	}
	var req Message
	if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
		t.Errorf("mock: unmarshal: %v", err)
		return
	}

	resp := Message{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  json.RawMessage(`{"threads":[{"id":"t-1","status":"active","title":"A"},{"id":"t-2","status":"idle","title":"B"}]}`),
	}
	data, _ := json.Marshal(resp)
	data = append(data, '\n')
	writer.Write(data)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/appserver/ -v -run "TestClientCreate|TestClientList"`
Expected: FAIL — methods not defined

**Step 3: Implement typed client helpers**

```go
// internal/appserver/client_thread.go
package appserver

import (
	"context"
	"encoding/json"
	"fmt"
)

// CreateThread sends a thread/create request and returns the new thread ID.
func (c *Client) CreateThread(ctx context.Context, instructions string) (*ThreadCreateResult, error) {
	params, _ := json.Marshal(ThreadCreateParams{
		Instructions: instructions,
	})

	resp, err := c.Call(ctx, MethodThreadCreate, params)
	if err != nil {
		return nil, fmt.Errorf("thread/create: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("thread/create: %w", resp.Error)
	}

	var result ThreadCreateResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("unmarshal thread/create result: %w", err)
	}
	return &result, nil
}

// ListThreads sends a thread/list request and returns all threads.
func (c *Client) ListThreads(ctx context.Context) (*ThreadListResult, error) {
	resp, err := c.Call(ctx, MethodThreadList, json.RawMessage(`{}`))
	if err != nil {
		return nil, fmt.Errorf("thread/list: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("thread/list: %w", resp.Error)
	}

	var result ThreadListResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("unmarshal thread/list result: %w", err)
	}
	return &result, nil
}

// DeleteThread sends a thread/delete request.
func (c *Client) DeleteThread(ctx context.Context, threadID string) error {
	params, _ := json.Marshal(ThreadDeleteParams{
		ThreadID: threadID,
	})

	resp, err := c.Call(ctx, MethodThreadDelete, params)
	if err != nil {
		return fmt.Errorf("thread/delete: %w", err)
	}
	if resp.Error != nil {
		return fmt.Errorf("thread/delete: %w", resp.Error)
	}
	return nil
}
```

**Step 4: Run tests**

Run: `go test ./internal/appserver/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/appserver/client_thread.go internal/appserver/client_thread_test.go
git commit -m "feat(appserver): typed thread management client methods"
```

---

### Task 7: Wire Router into Client Dispatch

**Files:**
- Modify: `internal/appserver/client.go`
- Create: `internal/appserver/dispatch_test.go`

**Step 1: Write test for router-based dispatch**

```go
// internal/appserver/dispatch_test.go
package appserver

import (
	"encoding/json"
	"sync/atomic"
	"testing"
)

func TestDispatchRoutesNotificationToRouter(t *testing.T) {
	client := &Client{}
	router := NewNotificationRouter()

	var called atomic.Bool
	router.OnThreadStatusChanged(func(params ThreadStatusChanged) {
		called.Store(true)
		if params.ThreadID != "t-1" {
			t.Errorf("expected t-1, got %s", params.ThreadID)
		}
	})

	client.Router = router

	msg := Message{
		JSONRPC: "2.0",
		Method:  NotifyThreadStatusChanged,
		Params:  json.RawMessage(`{"threadId":"t-1","status":"active","title":"Test"}`),
	}
	client.Dispatch(msg)

	if !called.Load() {
		t.Error("router handler was not called")
	}
}

func TestDispatchFallsBackToOnNotification(t *testing.T) {
	client := &Client{}

	var called atomic.Bool
	client.OnNotification = func(method string, params json.RawMessage) {
		called.Store(true)
	}

	msg := Message{
		JSONRPC: "2.0",
		Method:  "custom/notification",
		Params:  json.RawMessage(`{}`),
	}
	client.Dispatch(msg)

	if !called.Load() {
		t.Error("OnNotification was not called")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/appserver/ -v -run TestDispatchRoutes`
Expected: FAIL — `Router` field not defined on Client

**Step 3: Add Router field to Client and update Dispatch**

In `internal/appserver/client.go`, add the Router field to the Client struct:

```go
// Add to Client struct, after OnServerRequest:
	// Router dispatches typed notifications by method name.
	// Falls back to OnNotification for unregistered methods.
	Router *NotificationRouter
```

Update the notification section of Dispatch:

```go
// Replace the notification section in Dispatch:
	// Notification (no ID)
	if msg.Method == "" {
		return
	}

	routerHandled := r.Router != nil
	if routerHandled {
		r.Router.Handle(msg.Method, msg.Params)
	}

	if c.OnNotification != nil {
		c.OnNotification(msg.Method, msg.Params)
	}
```

> **Note:** Both Router and OnNotification fire — Router for typed handling, OnNotification as a generic hook (useful for logging or the event bridge in Phase 3).

**Step 4: Run tests**

Run: `go test ./internal/appserver/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/appserver/client.go internal/appserver/dispatch_test.go
git commit -m "feat(appserver): wire notification router into dispatch"
```
