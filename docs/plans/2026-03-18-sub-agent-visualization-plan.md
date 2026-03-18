# Sub-Agent Visualization Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Visualize Codex sub-agent hierarchies on DJ's canvas grid with parent-child connectors by migrating to the v2 app-server protocol.

**Architecture:** Replace the legacy `{id, msg: {type}}` protocol with JSON-RPC 2.0 `{method, params}` in the appserver package. Wire collaboration events (`CollabAgentSpawn*`) into ThreadStore parent-child relationships. Render tree-ordered cards with box-drawing connectors on the canvas.

**Tech Stack:** Go 1.25, Bubble Tea, Lipgloss, JSON-RPC 2.0 over stdio (JSONL)

**Design doc:** `docs/plans/2026-03-18-sub-agent-visualization-design.md`

**CI constraints:** funlen 60 lines, cyclop 15, file max 300 lines (non-test), `go test -race`

---

## Phase 1: Protocol Types

### Task 1: JSON-RPC Envelope Types

Replace `ProtoEvent` / `EventHeader` / `ProtoSubmission` with JSON-RPC 2.0 types.

**Files:**
- Modify: `internal/appserver/protocol.go`
- Test: `internal/appserver/protocol_test.go`

**Step 1: Write failing tests for JSON-RPC parsing**

```go
// protocol_test.go
package appserver

import (
	"encoding/json"
	"testing"
)

func TestParseNotification(t *testing.T) {
	raw := `{"jsonrpc":"2.0","method":"thread/started","params":{"thread_id":"t-1"}}`
	var message JsonRpcMessage
	if err := json.Unmarshal([]byte(raw), &message); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if message.Method != "thread/started" {
		t.Errorf("expected thread/started, got %s", message.Method)
	}
	if message.IsRequest() {
		t.Error("notification should not be a request")
	}
}

func TestParseRequest(t *testing.T) {
	raw := `{"jsonrpc":"2.0","id":"req-1","method":"item/commandExecution/requestApproval","params":{"command":"ls"}}`
	var message JsonRpcMessage
	if err := json.Unmarshal([]byte(raw), &message); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if message.ID != "req-1" {
		t.Errorf("expected req-1, got %s", message.ID)
	}
	if !message.IsRequest() {
		t.Error("should be a request")
	}
}

func TestParseResponse(t *testing.T) {
	raw := `{"jsonrpc":"2.0","id":"dj-1","result":{"ok":true}}`
	var message JsonRpcMessage
	if err := json.Unmarshal([]byte(raw), &message); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !message.IsResponse() {
		t.Error("should be a response")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/appserver -run TestParse -v`
Expected: Compile error — `JsonRpcMessage` undefined

**Step 3: Implement JSON-RPC types**

Replace `protocol.go` contents:

```go
package appserver

import "encoding/json"

// JsonRpcMessage represents a JSON-RPC 2.0 message (notification, request, or response).
type JsonRpcMessage struct {
	JsonRpc string          `json:"jsonrpc"`
	ID      string          `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   json.RawMessage `json:"error,omitempty"`
}

// IsRequest returns true if this message is a server-to-client request.
func (message JsonRpcMessage) IsRequest() bool {
	return message.ID != "" && message.Method != ""
}

// IsResponse returns true if this message is a response to a client request.
func (message JsonRpcMessage) IsResponse() bool {
	return message.ID != "" && message.Method == ""
}

// IsNotification returns true if this message is a server notification.
func (message JsonRpcMessage) IsNotification() bool {
	return message.ID == "" && message.Method != ""
}

// JsonRpcRequest is an outgoing client-to-server request.
type JsonRpcRequest struct {
	JsonRpc string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JsonRpcResponse is an outgoing client response to a server request.
type JsonRpcResponse struct {
	JsonRpc string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Result  interface{} `json:"result"`
}
```

Note: Keep `ProtoEvent`, `EventHeader`, and `ProtoSubmission` temporarily as deprecated aliases (removed after bridge migration in Task 7). This prevents breaking all existing code at once.

```go
// Deprecated: Legacy protocol types kept during migration.
type ProtoEvent = JsonRpcMessage
type EventHeader struct {
	Type string `json:"type"`
}
type ProtoSubmission = JsonRpcRequest
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/appserver -run TestParse -v`
Expected: PASS

**Step 5: Run full test suite to verify no regressions**

Run: `go test ./... -v -race`
Expected: All pass (aliases maintain backward compat)

**Step 6: Commit**

```
git add internal/appserver/protocol.go internal/appserver/protocol_test.go
git commit -m "feat: add JSON-RPC 2.0 envelope types for v2 protocol"
```

---

### Task 2: V2 Method Constants

Replace legacy event type strings with v2 method strings.

**Files:**
- Modify: `internal/appserver/methods.go`

**Step 1: Add v2 method constants alongside legacy ones**

```go
package appserver

// Legacy event types (deprecated — remove after bridge migration).
const (
	EventSessionConfigured = "session_configured"
	EventTaskStarted       = "task_started"
	EventTaskComplete      = "task_complete"
	EventAgentMessage      = "agent_message"
	EventAgentMessageDelta = "agent_message_delta"
	EventAgentReasoning    = "agent_reasoning"
	EventAgentReasonDelta  = "agent_reasoning_delta"
	EventTokenCount        = "token_count"
	EventExecApproval      = "exec_command_request"
	EventPatchApproval     = "patch_apply_request"
	EventAgentReasonBreak  = "agent_reasoning_section_break"
)

// V2 server notification methods.
const (
	MethodThreadStarted       = "thread/started"
	MethodThreadStatusChanged = "thread/status/changed"
	MethodThreadClosed        = "thread/closed"
	MethodTurnStarted         = "turn/started"
	MethodTurnCompleted       = "turn/completed"
	MethodItemStarted         = "item/started"
	MethodItemCompleted       = "item/completed"
	MethodAgentMessageDelta   = "item/agentMessage/delta"
	MethodTokenUsageUpdated   = "thread/tokenUsage/updated"
	MethodExecOutputDelta     = "item/commandExecution/outputDelta"
	MethodErrorNotification   = "error"
)

// V2 server request methods (require response).
const (
	MethodExecApproval = "item/commandExecution/requestApproval"
	MethodFileApproval = "item/fileChange/requestApproval"
)

// V2 client request methods (outgoing).
const (
	MethodInitialize  = "initialize"
	MethodThreadStart = "thread/start"
	MethodTurnStart   = "turn/start"
	MethodTurnInterrupt = "turn/interrupt"
)

// Legacy operation types (deprecated — remove after client migration).
const (
	OpUserInput     = "user_input"
	OpInterrupt     = "interrupt"
	OpExecApproval  = "exec_approval"
	OpPatchApproval = "patch_approval"
	OpShutdown      = "shutdown"
)
```

**Step 2: Run full test suite**

Run: `go test ./... -v -race`
Expected: All pass (additive change)

**Step 3: Commit**

```
git add internal/appserver/methods.go
git commit -m "feat: add v2 JSON-RPC method constants"
```

---

### Task 3: V2 Thread & Turn Types

Replace `SessionConfigured`, `TaskStarted`, `TaskComplete` with v2 types.

**Files:**
- Modify: `internal/appserver/types_thread.go`
- Test: `internal/appserver/types_thread_test.go`

**Step 1: Write failing tests**

```go
package appserver

import (
	"encoding/json"
	"testing"
)

func TestUnmarshalThreadStarted(t *testing.T) {
	raw := `{
		"thread": {
			"id": "t-1",
			"status": "idle",
			"source": {"type": "sub_agent", "parent_thread_id": "t-0", "depth": 1, "agent_nickname": "scout", "agent_role": "researcher"}
		}
	}`
	var notification ThreadStartedNotification
	if err := json.Unmarshal([]byte(raw), &notification); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if notification.Thread.ID != "t-1" {
		t.Errorf("expected t-1, got %s", notification.Thread.ID)
	}
	if notification.Thread.Source.Type != SourceTypeSubAgent {
		t.Errorf("expected sub_agent source, got %s", notification.Thread.Source.Type)
	}
	if notification.Thread.Source.ParentThreadID != "t-0" {
		t.Errorf("expected parent t-0, got %s", notification.Thread.Source.ParentThreadID)
	}
}

func TestUnmarshalThreadStartedCLISource(t *testing.T) {
	raw := `{"thread": {"id": "t-1", "status": "idle", "source": {"type": "cli"}}}`
	var notification ThreadStartedNotification
	if err := json.Unmarshal([]byte(raw), &notification); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if notification.Thread.Source.Type != SourceTypeCLI {
		t.Errorf("expected cli source, got %s", notification.Thread.Source.Type)
	}
}

func TestUnmarshalTurnStarted(t *testing.T) {
	raw := `{"thread_id": "t-1", "turn": {"id": "turn-1", "status": "in_progress"}}`
	var notification TurnStartedNotification
	if err := json.Unmarshal([]byte(raw), &notification); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if notification.ThreadID != "t-1" {
		t.Errorf("expected t-1, got %s", notification.ThreadID)
	}
}

func TestUnmarshalTurnCompleted(t *testing.T) {
	raw := `{"thread_id": "t-1", "turn": {"id": "turn-1", "status": "completed"}}`
	var notification TurnCompletedNotification
	if err := json.Unmarshal([]byte(raw), &notification); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if notification.ThreadID != "t-1" {
		t.Errorf("expected t-1, got %s", notification.ThreadID)
	}
}
```

**Step 2: Run to verify failure**

Run: `go test ./internal/appserver -run TestUnmarshal -v`
Expected: Compile error — types undefined

**Step 3: Implement v2 types**

Rewrite `types_thread.go`:

```go
package appserver

const (
	SourceTypeCLI      = "cli"
	SourceTypeSubAgent = "sub_agent"
	SourceTypeExec     = "exec"
)

const (
	ThreadStatusIdle   = "idle"
	ThreadStatusActive = "active"
)

type SessionSource struct {
	Type           string `json:"type"`
	ParentThreadID string `json:"parent_thread_id,omitempty"`
	Depth          int    `json:"depth,omitempty"`
	AgentNickname  string `json:"agent_nickname,omitempty"`
	AgentRole      string `json:"agent_role,omitempty"`
}

type Thread struct {
	ID     string        `json:"id"`
	Status string        `json:"status"`
	Source SessionSource `json:"source"`
}

type ThreadStartedNotification struct {
	Thread Thread `json:"thread"`
}

type ThreadStatusChangedNotification struct {
	ThreadID string `json:"thread_id"`
	Status   string `json:"status"`
}

type TurnStartedNotification struct {
	ThreadID string `json:"thread_id"`
	Turn     Turn   `json:"turn"`
}

type TurnCompletedNotification struct {
	ThreadID string `json:"thread_id"`
	Turn     Turn   `json:"turn"`
}

type Turn struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type AgentMessageDeltaNotification struct {
	ThreadID string `json:"thread_id"`
	Delta    string `json:"delta"`
}

type ItemCompletedNotification struct {
	ThreadID string `json:"thread_id"`
	Item     Item   `json:"item"`
}

type Item struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}
```

Keep legacy types at bottom with deprecation note:

```go
// Deprecated: Legacy types kept during migration. Remove after bridge migration.
type SessionConfigured struct {
	SessionID       string `json:"session_id"`
	Model           string `json:"model"`
	ReasoningEffort string `json:"reasoning_effort"`
	HistoryLogID    int64  `json:"history_log_id"`
	RolloutPath     string `json:"rollout_path"`
}

type TaskStarted struct {
	ModelContextWindow int `json:"model_context_window"`
}

type TaskComplete struct {
	LastAgentMessage string `json:"last_agent_message"`
}

type AgentMessage struct {
	Message string `json:"message"`
}

type AgentDelta struct {
	Delta string `json:"delta"`
}
```

Note: This file may exceed 300 lines with legacy types. Split the legacy types into `types_legacy.go` if needed for CI.

**Step 4: Run tests**

Run: `go test ./internal/appserver -run TestUnmarshal -v`
Expected: PASS

**Step 5: Run full suite**

Run: `go test ./... -v -race`
Expected: All pass

**Step 6: Commit**

```
git add internal/appserver/types_thread.go internal/appserver/types_thread_test.go
git commit -m "feat: add v2 thread and turn notification types"
```

---

### Task 4: V2 Collaboration Types

New types for the 10 collaboration events.

**Files:**
- Create: `internal/appserver/types_collab.go`
- Test: `internal/appserver/types_collab_test.go`

**Step 1: Write failing tests**

```go
package appserver

import (
	"encoding/json"
	"testing"
)

func TestUnmarshalCollabSpawnEnd(t *testing.T) {
	raw := `{
		"call_id": "call-1",
		"sender_thread_id": "t-0",
		"new_thread_id": "t-1",
		"new_agent_nickname": "scout",
		"new_agent_role": "researcher",
		"status": "running"
	}`
	var event CollabSpawnEndEvent
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if event.SenderThreadID != "t-0" {
		t.Errorf("expected t-0, got %s", event.SenderThreadID)
	}
	if event.NewThreadID != "t-1" {
		t.Errorf("expected t-1, got %s", event.NewThreadID)
	}
	if event.Status != AgentStatusRunning {
		t.Errorf("expected running, got %s", event.Status)
	}
}

func TestUnmarshalCollabWaitingEnd(t *testing.T) {
	raw := `{
		"sender_thread_id": "t-0",
		"call_id": "call-2",
		"statuses": {"t-1": "completed", "t-2": "running"}
	}`
	var event CollabWaitingEndEvent
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(event.Statuses) != 2 {
		t.Errorf("expected 2 statuses, got %d", len(event.Statuses))
	}
}

func TestUnmarshalCollabCloseEnd(t *testing.T) {
	raw := `{
		"call_id": "call-3",
		"sender_thread_id": "t-0",
		"receiver_thread_id": "t-1",
		"receiver_agent_nickname": "scout",
		"receiver_agent_role": "researcher",
		"status": "shutdown"
	}`
	var event CollabCloseEndEvent
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if event.ReceiverThreadID != "t-1" {
		t.Errorf("expected t-1, got %s", event.ReceiverThreadID)
	}
	if event.Status != AgentStatusShutdown {
		t.Errorf("expected shutdown, got %s", event.Status)
	}
}
```

**Step 2: Run to verify failure**

Run: `go test ./internal/appserver -run TestUnmarshalCollab -v`
Expected: Compile error

**Step 3: Implement collaboration types**

Create `types_collab.go`:

```go
package appserver

const (
	AgentStatusPendingInit = "pending_init"
	AgentStatusRunning     = "running"
	AgentStatusInterrupted = "interrupted"
	AgentStatusCompleted   = "completed"
	AgentStatusErrored     = "errored"
	AgentStatusShutdown    = "shutdown"
)

type CollabSpawnBeginEvent struct {
	CallID         string `json:"call_id"`
	SenderThreadID string `json:"sender_thread_id"`
	Prompt         string `json:"prompt,omitempty"`
	Model          string `json:"model,omitempty"`
}

type CollabSpawnEndEvent struct {
	CallID           string `json:"call_id"`
	SenderThreadID   string `json:"sender_thread_id"`
	NewThreadID      string `json:"new_thread_id"`
	NewAgentNickname string `json:"new_agent_nickname,omitempty"`
	NewAgentRole     string `json:"new_agent_role,omitempty"`
	Status           string `json:"status"`
}

type CollabInteractionBeginEvent struct {
	CallID           string `json:"call_id"`
	SenderThreadID   string `json:"sender_thread_id"`
	ReceiverThreadID string `json:"receiver_thread_id"`
	Prompt           string `json:"prompt,omitempty"`
}

type CollabInteractionEndEvent struct {
	CallID                string `json:"call_id"`
	SenderThreadID        string `json:"sender_thread_id"`
	ReceiverThreadID      string `json:"receiver_thread_id"`
	ReceiverAgentNickname string `json:"receiver_agent_nickname,omitempty"`
	ReceiverAgentRole     string `json:"receiver_agent_role,omitempty"`
	Status                string `json:"status"`
}

type CollabWaitingBeginEvent struct {
	CallID           string   `json:"call_id"`
	SenderThreadID   string   `json:"sender_thread_id"`
	ReceiverThreadIDs []string `json:"receiver_thread_ids"`
}

type CollabWaitingEndEvent struct {
	CallID         string            `json:"call_id"`
	SenderThreadID string            `json:"sender_thread_id"`
	Statuses       map[string]string `json:"statuses"`
}

type CollabCloseBeginEvent struct {
	CallID           string `json:"call_id"`
	SenderThreadID   string `json:"sender_thread_id"`
	ReceiverThreadID string `json:"receiver_thread_id"`
}

type CollabCloseEndEvent struct {
	CallID                string `json:"call_id"`
	SenderThreadID        string `json:"sender_thread_id"`
	ReceiverThreadID      string `json:"receiver_thread_id"`
	ReceiverAgentNickname string `json:"receiver_agent_nickname,omitempty"`
	ReceiverAgentRole     string `json:"receiver_agent_role,omitempty"`
	Status                string `json:"status"`
}

type CollabResumeBeginEvent struct {
	CallID                string `json:"call_id"`
	SenderThreadID        string `json:"sender_thread_id"`
	ReceiverThreadID      string `json:"receiver_thread_id"`
	ReceiverAgentNickname string `json:"receiver_agent_nickname,omitempty"`
	ReceiverAgentRole     string `json:"receiver_agent_role,omitempty"`
}

type CollabResumeEndEvent struct {
	CallID                string `json:"call_id"`
	SenderThreadID        string `json:"sender_thread_id"`
	ReceiverThreadID      string `json:"receiver_thread_id"`
	ReceiverAgentNickname string `json:"receiver_agent_nickname,omitempty"`
	ReceiverAgentRole     string `json:"receiver_agent_role,omitempty"`
	Status                string `json:"status"`
}
```

**Step 4: Run tests**

Run: `go test ./internal/appserver -run TestUnmarshalCollab -v`
Expected: PASS

**Step 5: Commit**

```
git add internal/appserver/types_collab.go internal/appserver/types_collab_test.go
git commit -m "feat: add v2 collaboration event types"
```

---

### Task 5: V2 Approval Types

Add types for v2 server requests (command/file approval).

**Files:**
- Create: `internal/appserver/types_approval.go`
- Test: `internal/appserver/types_approval_test.go`

**Step 1: Write failing test**

```go
package appserver

import (
	"encoding/json"
	"testing"
)

func TestUnmarshalCommandApproval(t *testing.T) {
	raw := `{"thread_id":"t-1","command":{"command":"ls -la","cwd":"/tmp"}}`
	var request CommandApprovalRequest
	if err := json.Unmarshal([]byte(raw), &request); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if request.ThreadID != "t-1" {
		t.Errorf("expected t-1, got %s", request.ThreadID)
	}
	if request.Command.Command != "ls -la" {
		t.Errorf("expected ls -la, got %s", request.Command.Command)
	}
}
```

**Step 2: Run to verify failure**

Run: `go test ./internal/appserver -run TestUnmarshalCommand -v`
Expected: Compile error

**Step 3: Implement**

```go
package appserver

type CommandApprovalRequest struct {
	ThreadID string         `json:"thread_id"`
	Command  CommandDetails `json:"command"`
}

type CommandDetails struct {
	Command string `json:"command"`
	Cwd     string `json:"cwd,omitempty"`
}

type FileChangeApprovalRequest struct {
	ThreadID string `json:"thread_id"`
	Patch    string `json:"patch"`
}

// Deprecated: Legacy approval types kept during migration.
type ExecCommandRequest struct {
	Command string `json:"command"`
	Cwd     string `json:"cwd,omitempty"`
}

type PatchApplyRequest struct {
	Patch string `json:"patch"`
}

type UserInputOp struct {
	Type  string      `json:"type"`
	Items []InputItem `json:"items"`
}

type InputItem struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}
```

**Step 4: Run tests and commit**

Run: `go test ./internal/appserver -v -race`
Expected: PASS

```
git add internal/appserver/types_approval.go internal/appserver/types_approval_test.go
git commit -m "feat: add v2 approval request types"
```

---

## Phase 2: Client & Bridge Migration

### Task 6: Client ReadLoop V2 Parsing

Update ReadLoop to parse JSON-RPC messages and route by type.

**Files:**
- Modify: `internal/appserver/client.go`
- Modify: `internal/appserver/client_test.go`

**Step 1: Write failing tests for v2 ReadLoop**

Add to `client_test.go`:

```go
func TestReadLoopParsesV2Notification(t *testing.T) {
	reader, writer := io.Pipe()
	client := &Client{
		scanner: bufio.NewScanner(reader),
	}
	client.scanner.Buffer(make([]byte, scannerBufferSize), scannerBufferSize)

	var received JsonRpcMessage
	done := make(chan struct{})
	go func() {
		client.ReadLoop(func(message JsonRpcMessage) {
			received = message
			close(done)
		})
	}()

	line := `{"jsonrpc":"2.0","method":"thread/started","params":{"thread":{"id":"t-1"}}}` + "\n"
	writer.Write([]byte(line))
	writer.Close()
	<-done

	if received.Method != "thread/started" {
		t.Errorf("expected thread/started, got %s", received.Method)
	}
}

func TestReadLoopParsesV2Request(t *testing.T) {
	reader, writer := io.Pipe()
	client := &Client{
		scanner: bufio.NewScanner(reader),
	}
	client.scanner.Buffer(make([]byte, scannerBufferSize), scannerBufferSize)

	var received JsonRpcMessage
	done := make(chan struct{})
	go func() {
		client.ReadLoop(func(message JsonRpcMessage) {
			received = message
			close(done)
		})
	}()

	line := `{"jsonrpc":"2.0","id":"req-1","method":"item/commandExecution/requestApproval","params":{"command":"ls"}}` + "\n"
	writer.Write([]byte(line))
	writer.Close()
	<-done

	if received.ID != "req-1" {
		t.Errorf("expected req-1, got %s", received.ID)
	}
	if !received.IsRequest() {
		t.Error("should be a request")
	}
}
```

**Step 2: Run to verify failure**

Run: `go test ./internal/appserver -run TestReadLoopParsesV2 -v`
Expected: Compile error — ReadLoop signature expects `func(ProtoEvent)`, not `func(JsonRpcMessage)`

**Step 3: Update ReadLoop and related methods**

Change `ReadLoop` handler signature from `func(ProtoEvent)` to `func(JsonRpcMessage)`:

```go
func (client *Client) ReadLoop(handler func(JsonRpcMessage)) {
	for client.scanner.Scan() {
		line := client.scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var message JsonRpcMessage
		if err := json.Unmarshal(line, &message); err != nil {
			continue
		}

		handler(message)
	}
}
```

Update `Send` to use `JsonRpcRequest`:

```go
func (client *Client) Send(request *JsonRpcRequest) error {
	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	client.mu.Lock()
	defer client.mu.Unlock()

	data = append(data, '\n')
	_, err = client.stdin.Write(data)
	return err
}
```

Update `SendUserInput`, `SendInterrupt`, `SendApproval` to build `JsonRpcRequest` instead of `ProtoSubmission`.

Also update the `OnEvent` callback type on the Client struct from `func(event ProtoEvent)` to `func(message JsonRpcMessage)`.

**Step 4: Update all call sites**

The `tui/app_proto.go` file references `ProtoEvent` in:
- `events chan appserver.ProtoEvent` → `events chan appserver.JsonRpcMessage`
- `protoEventMsg` wrapper struct
- `connectClient` goroutine
- `handleProtoEvent` method

Update all of these. The bridge call in `handleProtoEvent` stays the same shape but takes `JsonRpcMessage` instead.

**Step 5: Run full test suite**

Run: `go test ./... -v -race`
Expected: All pass after updating all references

**Step 6: Commit**

```
git add internal/appserver/client.go internal/appserver/client_test.go internal/tui/app.go internal/tui/app_proto.go
git commit -m "feat: migrate client ReadLoop to JSON-RPC 2.0 message format"
```

---

### Task 7: New Bubble Tea Messages

Add thread-scoped message types for v2 events.

**Files:**
- Modify: `internal/tui/msgs.go`

**Step 1: Add v2 message types**

Append to `msgs.go`:

```go
type ThreadStartedMsg struct {
	ThreadID      string
	Status        string
	SourceType    string
	ParentID      string
	Depth         int
	AgentNickname string
	AgentRole     string
}

type ThreadStatusChangedMsg struct {
	ThreadID string
	Status   string
}

type TurnStartedMsg struct {
	ThreadID string
	TurnID   string
}

type TurnCompletedMsg struct {
	ThreadID string
	TurnID   string
}

type V2AgentDeltaMsg struct {
	ThreadID string
	Delta    string
}

type V2ExecApprovalMsg struct {
	RequestID string
	ThreadID  string
	Command   string
	Cwd       string
}

type V2FileApprovalMsg struct {
	RequestID string
	ThreadID  string
	Patch     string
}

type CollabSpawnMsg struct {
	SenderThreadID   string
	NewThreadID      string
	NewAgentNickname string
	NewAgentRole     string
	Status           string
}

type CollabCloseMsg struct {
	SenderThreadID   string
	ReceiverThreadID string
	Status           string
}

type CollabStatusUpdateMsg struct {
	ThreadID string
	Status   string
}
```

**Step 2: Run full suite**

Run: `go test ./... -v -race`
Expected: All pass (additive change)

**Step 3: Commit**

```
git add internal/tui/msgs.go
git commit -m "feat: add v2 Bubble Tea message types with thread-scoped IDs"
```

---

### Task 8: Bridge V2 Routing

Rewrite bridge to route on JSON-RPC method strings.

**Files:**
- Modify: `internal/tui/bridge.go`
- Modify: `internal/tui/bridge_test.go`

**Step 1: Write failing tests for v2 bridge**

```go
func TestBridgeV2ThreadStarted(t *testing.T) {
	message := appserver.JsonRpcMessage{
		Method: appserver.MethodThreadStarted,
		Params: json.RawMessage(`{"thread":{"id":"t-1","status":"idle","source":{"type":"cli"}}}`),
	}
	msg := V2MessageToMsg(message)
	started, ok := msg.(ThreadStartedMsg)
	if !ok {
		t.Fatalf("expected ThreadStartedMsg, got %T", msg)
	}
	if started.ThreadID != "t-1" {
		t.Errorf("expected t-1, got %s", started.ThreadID)
	}
}

func TestBridgeV2SubAgentThread(t *testing.T) {
	message := appserver.JsonRpcMessage{
		Method: appserver.MethodThreadStarted,
		Params: json.RawMessage(`{"thread":{"id":"t-2","status":"idle","source":{"type":"sub_agent","parent_thread_id":"t-1","depth":1,"agent_nickname":"scout","agent_role":"researcher"}}}`),
	}
	msg := V2MessageToMsg(message)
	started, ok := msg.(ThreadStartedMsg)
	if !ok {
		t.Fatalf("expected ThreadStartedMsg, got %T", msg)
	}
	if started.ParentID != "t-1" {
		t.Errorf("expected parent t-1, got %s", started.ParentID)
	}
	if started.AgentRole != "researcher" {
		t.Errorf("expected researcher, got %s", started.AgentRole)
	}
}

func TestBridgeV2TurnStarted(t *testing.T) {
	message := appserver.JsonRpcMessage{
		Method: appserver.MethodTurnStarted,
		Params: json.RawMessage(`{"thread_id":"t-1","turn":{"id":"turn-1","status":"in_progress"}}`),
	}
	msg := V2MessageToMsg(message)
	turn, ok := msg.(TurnStartedMsg)
	if !ok {
		t.Fatalf("expected TurnStartedMsg, got %T", msg)
	}
	if turn.ThreadID != "t-1" {
		t.Errorf("expected t-1, got %s", turn.ThreadID)
	}
}

func TestBridgeV2AgentDelta(t *testing.T) {
	message := appserver.JsonRpcMessage{
		Method: appserver.MethodAgentMessageDelta,
		Params: json.RawMessage(`{"thread_id":"t-1","delta":"hello"}`),
	}
	msg := V2MessageToMsg(message)
	delta, ok := msg.(V2AgentDeltaMsg)
	if !ok {
		t.Fatalf("expected V2AgentDeltaMsg, got %T", msg)
	}
	if delta.Delta != "hello" {
		t.Errorf("expected hello, got %s", delta.Delta)
	}
}

func TestBridgeV2UnknownMethodReturnsNil(t *testing.T) {
	message := appserver.JsonRpcMessage{
		Method: "some/unknown/method",
	}
	msg := V2MessageToMsg(message)
	if msg != nil {
		t.Errorf("expected nil for unknown method, got %T", msg)
	}
}
```

**Step 2: Run to verify failure**

Run: `go test ./internal/tui -run TestBridgeV2 -v`
Expected: Compile error — `V2MessageToMsg` undefined

**Step 3: Implement V2MessageToMsg**

Add new function to `bridge.go` (keep legacy `ProtoEventToMsg` temporarily):

```go
func V2MessageToMsg(message appserver.JsonRpcMessage) tea.Msg {
	switch message.Method {
	case appserver.MethodThreadStarted:
		return decodeThreadStarted(message.Params)
	case appserver.MethodTurnStarted:
		return decodeTurnStarted(message.Params)
	case appserver.MethodTurnCompleted:
		return decodeTurnCompleted(message.Params)
	case appserver.MethodAgentMessageDelta:
		return decodeV2AgentDelta(message.Params)
	case appserver.MethodThreadStatusChanged:
		return decodeThreadStatusChanged(message.Params)
	case appserver.MethodExecApproval:
		return decodeV2ExecApproval(message)
	case appserver.MethodFileApproval:
		return decodeV2FileApproval(message)
	}
	return nil
}
```

Each decode function is a small helper. Example for `decodeThreadStarted`:

```go
func decodeThreadStarted(raw json.RawMessage) tea.Msg {
	var notification appserver.ThreadStartedNotification
	if err := json.Unmarshal(raw, &notification); err != nil {
		return nil
	}
	thread := notification.Thread
	return ThreadStartedMsg{
		ThreadID:      thread.ID,
		Status:        thread.Status,
		SourceType:    thread.Source.Type,
		ParentID:      thread.Source.ParentThreadID,
		Depth:         thread.Source.Depth,
		AgentNickname: thread.Source.AgentNickname,
		AgentRole:     thread.Source.AgentRole,
	}
}
```

Put each decode helper in its own small function. If `bridge.go` exceeds 300 lines, split v2 decoders into `bridge_v2.go`.

**Step 4: Run tests**

Run: `go test ./internal/tui -run TestBridgeV2 -v`
Expected: PASS

**Step 5: Wire V2MessageToMsg into handleProtoEvent**

In `app_proto.go`, update `handleProtoEvent` to call `V2MessageToMsg` instead of `ProtoEventToMsg`:

```go
func (app AppModel) handleProtoEvent(message appserver.JsonRpcMessage) (tea.Model, tea.Cmd) {
	tuiMsg := V2MessageToMsg(message)
	if tuiMsg == nil {
		return app, app.listenForEvents()
	}
	updatedApp, command := app.Update(tuiMsg)
	nextListen := app.listenForEvents()
	return updatedApp, tea.Batch(command, nextListen)
}
```

**Step 6: Run full suite and commit**

Run: `go test ./... -v -race`
Expected: All pass

```
git add internal/tui/bridge.go internal/tui/bridge_test.go internal/tui/app_proto.go
git commit -m "feat: add v2 bridge routing with V2MessageToMsg"
```

---

## Phase 3: State Layer

### Task 9: ThreadState Extensions

Add sub-agent fields to ThreadState.

**Files:**
- Modify: `internal/state/thread.go`
- Modify: `internal/state/store.go`
- Test: `internal/state/store_test.go`

**Step 1: Write failing test**

```go
func TestAddWithParentFields(t *testing.T) {
	store := NewThreadStore()
	store.Add("t-0", "Root")
	store.AddWithParent("t-1", "Child", "t-0")

	child, exists := store.Get("t-1")
	if !exists {
		t.Fatal("child not found")
	}
	if child.ParentID != "t-0" {
		t.Errorf("expected parent t-0, got %s", child.ParentID)
	}
}

func TestAddSubAgent(t *testing.T) {
	store := NewThreadStore()
	store.Add("t-0", "Root")
	store.AddSubAgent("t-1", "Scout", "t-0", "scout", "researcher", 1)

	child, exists := store.Get("t-1")
	if !exists {
		t.Fatal("child not found")
	}
	if child.AgentNickname != "scout" {
		t.Errorf("expected scout, got %s", child.AgentNickname)
	}
	if child.AgentRole != "researcher" {
		t.Errorf("expected researcher, got %s", child.AgentRole)
	}
	if child.Depth != 1 {
		t.Errorf("expected depth 1, got %d", child.Depth)
	}
}
```

**Step 2: Run to verify failure**

Run: `go test ./internal/state -run TestAddSubAgent -v`
Expected: Compile error — `AddSubAgent` undefined

**Step 3: Add fields and method**

In `thread.go`, add fields:

```go
type ThreadState struct {
	ID            string
	Title         string
	Status        string
	ParentID      string
	AgentNickname string
	AgentRole     string
	Depth         int
	Model         string
	Messages      []ChatMessage
	CommandOutput map[string]string
}
```

In `store.go`, add:

```go
func (store *ThreadStore) AddSubAgent(id string, title string, parentID string, nickname string, role string, depth int) {
	store.mu.Lock()
	defer store.mu.Unlock()

	thread := NewThreadState(id, title)
	thread.ParentID = parentID
	thread.AgentNickname = nickname
	thread.AgentRole = role
	thread.Depth = depth
	store.threads[id] = thread
	store.order = append(store.order, id)
}
```

**Step 4: Run tests**

Run: `go test ./internal/state -v -race`
Expected: All PASS

**Step 5: Commit**

```
git add internal/state/thread.go internal/state/store.go internal/state/store_test.go
git commit -m "feat: add sub-agent fields to ThreadState and AddSubAgent method"
```

---

### Task 10: Store TreeOrder Method

Depth-first traversal for tree-ordered grid layout.

**Files:**
- Modify: `internal/state/store.go`
- Test: `internal/state/store_test.go`

**Step 1: Write failing test**

```go
func TestTreeOrder(t *testing.T) {
	store := NewThreadStore()
	store.Add("root-1", "Root 1")
	store.Add("root-2", "Root 2")
	store.AddWithParent("child-1a", "Child 1a", "root-1")
	store.AddWithParent("child-1b", "Child 1b", "root-1")
	store.AddWithParent("child-2a", "Child 2a", "root-2")

	ordered := store.TreeOrder()
	expectedOrder := []string{"root-1", "child-1a", "child-1b", "root-2", "child-2a"}
	if len(ordered) != len(expectedOrder) {
		t.Fatalf("expected %d threads, got %d", len(expectedOrder), len(ordered))
	}
	for index, thread := range ordered {
		if thread.ID != expectedOrder[index] {
			t.Errorf("position %d: expected %s, got %s", index, expectedOrder[index], thread.ID)
		}
	}
}

func TestTreeOrderNestedChildren(t *testing.T) {
	store := NewThreadStore()
	store.Add("root", "Root")
	store.AddWithParent("child", "Child", "root")
	store.AddWithParent("grandchild", "Grandchild", "child")

	ordered := store.TreeOrder()
	expectedOrder := []string{"root", "child", "grandchild"}
	for index, thread := range ordered {
		if thread.ID != expectedOrder[index] {
			t.Errorf("position %d: expected %s, got %s", index, expectedOrder[index], thread.ID)
		}
	}
}
```

**Step 2: Run to verify failure**

Run: `go test ./internal/state -run TestTreeOrder -v`
Expected: Compile error

**Step 3: Implement TreeOrder**

Add to `store.go`:

```go
func (store *ThreadStore) TreeOrder() []*ThreadState {
	store.mu.RLock()
	defer store.mu.RUnlock()

	var result []*ThreadState
	for _, id := range store.order {
		thread := store.threads[id]
		if thread.ParentID == "" {
			result = append(result, thread)
			result = store.appendChildrenRecursive(result, id)
		}
	}
	return result
}

func (store *ThreadStore) appendChildrenRecursive(result []*ThreadState, parentID string) []*ThreadState {
	for _, id := range store.order {
		thread := store.threads[id]
		if thread.ParentID == parentID {
			result = append(result, thread)
			result = store.appendChildrenRecursive(result, id)
		}
	}
	return result
}
```

**Step 4: Run tests**

Run: `go test ./internal/state -run TestTreeOrder -v`
Expected: PASS

**Step 5: Commit**

```
git add internal/state/store.go internal/state/store_test.go
git commit -m "feat: add TreeOrder for depth-first thread traversal"
```

---

## Phase 4: Multi-Thread Event Routing

### Task 11: Thread-Scoped App Handlers

Remove global `sessionID`, route events by ThreadID.

**Files:**
- Modify: `internal/tui/app.go`
- Modify: `internal/tui/app_proto.go`

**Step 1: Add v2 handler functions**

Add handlers for the new message types in `app_proto.go`:

```go
func (app AppModel) handleThreadStarted(msg ThreadStartedMsg) (tea.Model, tea.Cmd) {
	isSubAgent := msg.SourceType == appserver.SourceTypeSubAgent
	if isSubAgent {
		app.store.AddSubAgent(msg.ThreadID, msg.AgentNickname, msg.ParentID, msg.AgentNickname, msg.AgentRole, msg.Depth)
	} else {
		app.store.Add(msg.ThreadID, msg.ThreadID)
	}
	app.statusBar.SetThreadCount(len(app.store.All()))
	app.tree.Refresh()
	return app, nil
}

func (app AppModel) handleTurnStarted(msg TurnStartedMsg) (tea.Model, tea.Cmd) {
	app.store.UpdateStatus(msg.ThreadID, state.StatusActive, "")
	return app, nil
}

func (app AppModel) handleTurnCompleted(msg TurnCompletedMsg) (tea.Model, tea.Cmd) {
	app.store.UpdateStatus(msg.ThreadID, state.StatusCompleted, "")
	return app, nil
}

func (app AppModel) handleV2AgentDelta(msg V2AgentDeltaMsg) (tea.Model, tea.Cmd) {
	thread, exists := app.store.Get(msg.ThreadID)
	if !exists {
		return app, nil
	}
	thread.AppendDelta("", msg.Delta)
	return app, nil
}

func (app AppModel) handleCollabSpawn(msg CollabSpawnMsg) (tea.Model, tea.Cmd) {
	app.tree.Refresh()
	return app, nil
}

func (app AppModel) handleCollabClose(msg CollabCloseMsg) (tea.Model, tea.Cmd) {
	agentStatus := mapAgentStatusToDJ(msg.Status)
	app.store.UpdateStatus(msg.ReceiverThreadID, agentStatus, "")
	return app, nil
}
```

Add status mapping helper:

```go
func mapAgentStatusToDJ(agentStatus string) string {
	statusMap := map[string]string{
		appserver.AgentStatusPendingInit: state.StatusIdle,
		appserver.AgentStatusRunning:     state.StatusActive,
		appserver.AgentStatusInterrupted: state.StatusIdle,
		appserver.AgentStatusCompleted:   state.StatusCompleted,
		appserver.AgentStatusErrored:     state.StatusError,
		appserver.AgentStatusShutdown:    state.StatusCompleted,
	}
	djStatus, exists := statusMap[agentStatus]
	if !exists {
		return state.StatusIdle
	}
	return djStatus
}
```

**Step 2: Wire into Update switch**

In `app.go`'s `Update` method, add cases for new message types:

```go
case ThreadStartedMsg:
	return app.handleThreadStarted(msg)
case TurnStartedMsg:
	return app.handleTurnStarted(msg)
case TurnCompletedMsg:
	return app.handleTurnCompleted(msg)
case V2AgentDeltaMsg:
	return app.handleV2AgentDelta(msg)
case CollabSpawnMsg:
	return app.handleCollabSpawn(msg)
case CollabCloseMsg:
	return app.handleCollabClose(msg)
case ThreadStatusChangedMsg:
	agentStatus := mapAgentStatusToDJ(msg.Status)
	app.store.UpdateStatus(msg.ThreadID, agentStatus, "")
	return app, nil
```

**Step 3: Run full test suite**

Run: `go test ./... -v -race`
Expected: All pass

**Step 4: Commit**

```
git add internal/tui/app.go internal/tui/app_proto.go
git commit -m "feat: add thread-scoped v2 event handlers with multi-thread routing"
```

---

## Phase 5: Canvas Visualization

### Task 12: Tree-Ordered Canvas Layout

Switch canvas from insertion order to tree order.

**Files:**
- Modify: `internal/tui/canvas.go`
- Test: `internal/tui/canvas_test.go`

**Step 1: Write failing test**

```go
func TestCanvasTreeOrder(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("root", "Root")
	store.AddWithParent("child-1", "Child 1", "root")
	store.AddWithParent("child-2", "Child 2", "root")

	canvas := NewCanvasModel(store)
	canvas.SetDimensions(120, 40)

	view := canvas.View()
	rootIndex := strings.Index(view, "Root")
	child1Index := strings.Index(view, "Child 1")
	child2Index := strings.Index(view, "Child 2")

	if rootIndex == -1 || child1Index == -1 || child2Index == -1 {
		t.Fatal("expected all threads to appear in view")
	}
	if rootIndex > child1Index {
		t.Error("root should appear before child-1")
	}
	if child1Index > child2Index {
		t.Error("child-1 should appear before child-2")
	}
}
```

**Step 2: Run to verify failure**

Run: `go test ./internal/tui -run TestCanvasTreeOrder -v`
Expected: May pass or fail depending on insertion order — the point is to verify tree order is used

**Step 3: Switch to TreeOrder in canvas View**

In `canvas.go`, change `View()`:

```go
func (canvas *CanvasModel) View() string {
	threads := canvas.store.TreeOrder()
	if len(threads) == 0 {
		return canvas.renderEmpty()
	}
	// ...
}
```

Also update `SelectedThreadID`, `MoveRight`, `MoveLeft`, `MoveDown`, `MoveUp` to use `TreeOrder()` instead of `All()`.

**Step 4: Run tests**

Run: `go test ./internal/tui -v -race`
Expected: All pass

**Step 5: Commit**

```
git add internal/tui/canvas.go internal/tui/canvas_test.go
git commit -m "feat: switch canvas to tree-ordered thread layout"
```

---

### Task 13: Canvas Edge Connectors

Draw box-drawing connectors between parent and child cards.

**Files:**
- Create: `internal/tui/canvas_edges.go`
- Test: `internal/tui/canvas_edges_test.go`

**Step 1: Write failing tests**

```go
package tui

import (
	"strings"
	"testing"
)

func TestRenderConnectorSimple(t *testing.T) {
	parentCol := 0
	childCols := []int{0}
	connector := renderConnectorRow(parentCol, childCols, 20, 2)
	if !strings.Contains(connector, "│") {
		t.Error("expected vertical connector")
	}
}

func TestRenderConnectorBranching(t *testing.T) {
	parentCol := 0
	childCols := []int{0, 2}
	connector := renderConnectorRow(parentCol, childCols, 20, 2)
	if !strings.Contains(connector, "├") || !strings.Contains(connector, "─") {
		t.Error("expected branching connector with horizontal lines")
	}
}

func TestRenderConnectorNoChildren(t *testing.T) {
	parentCol := 0
	childCols := []int{}
	connector := renderConnectorRow(parentCol, childCols, 20, 2)
	if connector != "" {
		t.Error("expected empty string for no children")
	}
}
```

**Step 2: Run to verify failure**

Run: `go test ./internal/tui -run TestRenderConnector -v`
Expected: Compile error

**Step 3: Implement connector rendering**

Create `canvas_edges.go`:

```go
package tui

import "strings"

const (
	edgeVertical   = "│"
	edgeHorizontal = "─"
	edgeTeeDown    = "┬"
	edgeTeeRight   = "├"
	edgeCornerRight = "┐"
	edgeElbow       = "└"
)

func renderConnectorRow(parentCol int, childCols []int, cardWidth int, gap int) string {
	if len(childCols) == 0 {
		return ""
	}

	cellWidth := cardWidth + gap
	parentCenter := parentCol*cellWidth + cardWidth/2
	totalWidth := computeConnectorWidth(childCols, cellWidth, cardWidth)

	line := buildConnectorLine(parentCenter, childCols, cellWidth, cardWidth, totalWidth)
	return line
}

func computeConnectorWidth(childCols []int, cellWidth int, cardWidth int) int {
	maxCol := 0
	for _, col := range childCols {
		if col > maxCol {
			maxCol = col
		}
	}
	return maxCol*cellWidth + cardWidth
}

func buildConnectorLine(parentCenter int, childCols []int, cellWidth int, cardWidth int, totalWidth int) string {
	childCenters := make(map[int]bool)
	for _, col := range childCols {
		center := col*cellWidth + cardWidth/2
		childCenters[center] = true
	}

	minCenter := totalWidth
	maxCenter := 0
	for center := range childCenters {
		if center < minCenter {
			minCenter = center
		}
		if center > maxCenter {
			maxCenter = center
		}
	}

	topLine := strings.Repeat(" ", parentCenter) + edgeVertical
	spanStart := minCenter
	spanEnd := maxCenter
	if parentCenter < spanStart {
		spanStart = parentCenter
	}
	if parentCenter > spanEnd {
		spanEnd = parentCenter
	}

	var bottomLine strings.Builder
	for position := 0; position <= spanEnd; position++ {
		isChildCenter := childCenters[position]
		isParentCenter := position == parentCenter
		isInSpan := position >= spanStart && position <= spanEnd

		character := resolveConnectorChar(position, isChildCenter, isParentCenter, isInSpan)
		bottomLine.WriteString(character)
	}

	return topLine + "\n" + bottomLine.String()
}

func resolveConnectorChar(position int, isChild bool, isParent bool, inSpan bool) string {
	if isParent && isChild {
		return edgeTeeDown
	}
	if isParent {
		return edgeTeeDown
	}
	if isChild && inSpan {
		return edgeElbow
	}
	if isChild {
		return edgeVertical
	}
	if inSpan {
		return edgeHorizontal
	}
	return " "
}
```

Note: This is a starting implementation. Exact box-drawing logic may need refinement during development — the tests will guide the correct character choices for edge cases.

**Step 4: Run tests**

Run: `go test ./internal/tui -run TestRenderConnector -v`
Expected: PASS (may need tweaking of exact characters in tests to match implementation)

**Step 5: Integrate into canvas.go renderGrid**

In `canvas.go`, after rendering each row, check if any card in the current row is a parent of cards in the next row. If so, insert a connector row:

```go
func (canvas *CanvasModel) renderGrid(threads []*state.ThreadState) string {
	// ... existing card rendering logic ...

	// Between rows, insert connector lines for parent-child relationships
	// Build position map: threadID -> column index
	// For each row boundary, find parents above and children below
	// Call renderConnectorRow for each parent
}
```

**Step 6: Run full suite and commit**

Run: `go test ./... -v -race`
Expected: All pass

```
git add internal/tui/canvas_edges.go internal/tui/canvas_edges_test.go internal/tui/canvas.go
git commit -m "feat: add canvas edge connectors between parent and child cards"
```

---

### Task 14: Card Sub-Agent Display

Show agent role and depth indicator on sub-agent cards.

**Files:**
- Modify: `internal/tui/card.go`
- Test: `internal/tui/card_test.go`

**Step 1: Write failing test**

```go
func TestSubAgentCardShowsRole(t *testing.T) {
	thread := state.NewThreadState("t-1", "Scout")
	thread.ParentID = "t-0"
	thread.AgentRole = "researcher"

	card := NewCardModel(thread, false, false)
	card.SetSize(30, 6)
	view := card.View()

	if !strings.Contains(view, "researcher") {
		t.Error("expected agent role in card view")
	}
}

func TestSubAgentCardShowsDepthPrefix(t *testing.T) {
	thread := state.NewThreadState("t-1", "Scout")
	thread.ParentID = "t-0"
	thread.Depth = 1

	card := NewCardModel(thread, false, false)
	card.SetSize(30, 6)
	view := card.View()

	if !strings.Contains(view, "↳") {
		t.Error("expected depth prefix ↳ in sub-agent card")
	}
}

func TestRootCardNoDepthPrefix(t *testing.T) {
	thread := state.NewThreadState("t-0", "Root Session")

	card := NewCardModel(thread, false, false)
	card.SetSize(30, 6)
	view := card.View()

	if strings.Contains(view, "↳") {
		t.Error("root card should not have depth prefix")
	}
}
```

**Step 2: Run to verify failure**

Run: `go test ./internal/tui -run TestSubAgent -v`
Expected: FAIL

**Step 3: Update card View**

In `card.go`, modify the `View()` method to add sub-agent info:

```go
func (card CardModel) View() string {
	statusColor, exists := statusColors[card.thread.Status]
	if !exists {
		statusColor = defaultStatusColor
	}

	statusLine := lipgloss.NewStyle().
		Foreground(statusColor).
		Render(card.thread.Status)

	title := card.buildTitle()
	content := card.buildContent(title, statusLine)

	style := card.buildBorderStyle()
	return style.Render(content)
}

func (card CardModel) buildTitle() string {
	titleMaxLen := card.width - cardBorderPadding
	if card.pinned {
		titleMaxLen -= len(pinnedIndicator)
	}

	title := card.thread.Title
	isSubAgent := card.thread.ParentID != ""
	if isSubAgent {
		title = subAgentPrefix + title
	}

	title = truncate(title, titleMaxLen)
	if card.pinned {
		title += pinnedIndicator
	}
	return title
}
```

Add the constant:

```go
const subAgentPrefix = "↳ "
```

Add role line for sub-agents:

```go
func (card CardModel) buildContent(title string, statusLine string) string {
	isSubAgent := card.thread.ParentID != ""
	hasRole := isSubAgent && card.thread.AgentRole != ""
	if hasRole {
		roleLine := lipgloss.NewStyle().
			Foreground(colorIdle).
			Render("  " + card.thread.AgentRole)
		return fmt.Sprintf("%s\n%s\n%s", title, roleLine, statusLine)
	}
	return fmt.Sprintf("%s\n%s", title, statusLine)
}
```

**Step 4: Run tests**

Run: `go test ./internal/tui -run TestSubAgent -v && go test ./internal/tui -run TestRootCard -v`
Expected: PASS

**Step 5: Run full suite and commit**

Run: `go test ./... -v -race`
Expected: All pass

```
git add internal/tui/card.go internal/tui/card_test.go
git commit -m "feat: show agent role and depth prefix on sub-agent cards"
```

---

## Phase 6: Cleanup

### Task 15: Remove Legacy Protocol Types

Remove deprecated types and legacy bridge function.

**Files:**
- Modify: `internal/appserver/protocol.go` — remove `ProtoEvent`, `EventHeader`, `ProtoSubmission` aliases
- Modify: `internal/appserver/methods.go` — remove legacy event constants and op constants
- Modify: `internal/appserver/types_thread.go` — remove legacy `SessionConfigured`, `TaskStarted`, etc.
- Modify: `internal/tui/bridge.go` — remove `ProtoEventToMsg` and legacy decode functions
- Update tests: remove legacy bridge tests

**Step 1: Remove all deprecated types and functions**

Delete the legacy aliases from `protocol.go`, legacy constants from `methods.go`, legacy types from `types_thread.go`/`types_approval.go`, and the `ProtoEventToMsg` function from `bridge.go`.

**Step 2: Fix any remaining compile errors**

Grep for any remaining references to removed types and update them.

**Step 3: Run full suite**

Run: `go test ./... -v -race`
Expected: All pass

**Step 4: Run linter**

Run: `golangci-lint run`
Expected: Clean

**Step 5: Commit**

```
git add -A
git commit -m "refactor: remove legacy protocol types after v2 migration"
```

---

### Task 16: File Length & Lint Compliance

Verify all files meet CI constraints.

**Step 1: Check file lengths**

Run: `find internal -name '*.go' ! -name '*_test.go' -exec wc -l {} + | sort -n`
Expected: No non-test file exceeds 300 lines

**Step 2: Check function lengths**

Run: `golangci-lint run`
Expected: No funlen or cyclop violations

**Step 3: Split any oversized files**

If `bridge.go` exceeds 300 lines: split v2 decoders into `bridge_v2.go`.
If `app_proto.go` exceeds 300 lines: split v2 handlers into `app_proto_v2.go`.
If `types_thread.go` exceeds 300 lines: split by category.

**Step 4: Final full suite**

Run: `go test ./... -v -race && golangci-lint run`
Expected: All clean

**Step 5: Commit if changes needed**

```
git add -A
git commit -m "refactor: split oversized files for CI compliance"
```
