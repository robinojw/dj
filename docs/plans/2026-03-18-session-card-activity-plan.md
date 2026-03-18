# Session Card Activity Indicators — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Show real-time activity on session cards (e.g. "Running: git status", "Thinking...", streaming text snippets) by adding an `Activity` field to `ThreadState` and rendering it on cards in place of the status word.

**Architecture:** Protocol event handlers set an `Activity` string on the thread via the store. The card's `View()` method checks `Activity` first — if non-empty, renders it with status color instead of the status word. Activity is cleared on `agent_message` (completed) and `task_complete`.

**Tech Stack:** Go, Bubble Tea, Lipgloss, standard `testing` package.

---

### Task 1: Add Activity field and methods to ThreadState

**Files:**
- Modify: `internal/state/thread.go`
- Modify: `internal/state/thread_test.go`

**Step 1: Write the failing tests**

Add to `internal/state/thread_test.go`:

```go
func TestThreadStateSetActivity(t *testing.T) {
	thread := NewThreadState("t-1", "Test")
	thread.SetActivity("Running: git status")

	if thread.Activity != "Running: git status" {
		t.Errorf("expected Running: git status, got %s", thread.Activity)
	}
}

func TestThreadStateClearActivity(t *testing.T) {
	thread := NewThreadState("t-1", "Test")
	thread.SetActivity("Thinking...")
	thread.ClearActivity()

	if thread.Activity != "" {
		t.Errorf("expected empty activity, got %s", thread.Activity)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/state -run TestThreadState(Set|Clear)Activity -v`
Expected: FAIL — `thread.Activity undefined`, `thread.SetActivity undefined`

**Step 3: Write minimal implementation**

Add `Activity` field to `ThreadState` struct and two methods in `internal/state/thread.go`:

```go
type ThreadState struct {
	ID            string
	Title         string
	Status        string
	Activity      string
	ParentID      string
	Messages      []ChatMessage
	CommandOutput map[string]string
}

func (threadState *ThreadState) SetActivity(activity string) {
	threadState.Activity = activity
}

func (threadState *ThreadState) ClearActivity() {
	threadState.Activity = ""
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/state -run TestThreadState(Set|Clear)Activity -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/state/thread.go internal/state/thread_test.go
git commit -m "feat: add Activity field to ThreadState"
```

---

### Task 2: Add UpdateActivity to ThreadStore

**Files:**
- Modify: `internal/state/store.go`
- Modify: `internal/state/store_test.go`

**Step 1: Write the failing tests**

Add to `internal/state/store_test.go`:

```go
func TestStoreUpdateActivity(t *testing.T) {
	store := NewThreadStore()
	store.Add("t-1", "Test")
	store.UpdateActivity("t-1", "Thinking...")

	thread, _ := store.Get("t-1")
	if thread.Activity != "Thinking..." {
		t.Errorf("expected Thinking..., got %s", thread.Activity)
	}
}

func TestStoreUpdateActivityMissing(t *testing.T) {
	store := NewThreadStore()
	store.UpdateActivity("missing", "Thinking...")
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/state -run TestStoreUpdateActivity -v`
Expected: FAIL — `store.UpdateActivity undefined`

**Step 3: Write minimal implementation**

Add to `internal/state/store.go`:

```go
func (store *ThreadStore) UpdateActivity(id string, activity string) {
	store.mu.Lock()
	defer store.mu.Unlock()

	thread, exists := store.threads[id]
	if !exists {
		return
	}
	thread.Activity = activity
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/state -run TestStoreUpdateActivity -v`
Expected: PASS

**Step 5: Run full state package tests**

Run: `go test ./internal/state -v`
Expected: All PASS

**Step 6: Commit**

```bash
git add internal/state/store.go internal/state/store_test.go
git commit -m "feat: add UpdateActivity to ThreadStore"
```

---

### Task 3: Render activity on session cards

**Files:**
- Modify: `internal/tui/card.go`
- Modify: `internal/tui/card_test.go`

**Step 1: Write the failing tests**

Add to `internal/tui/card_test.go`:

```go
func TestCardRenderShowsActivity(t *testing.T) {
	thread := state.NewThreadState("t-1", "o3-mini")
	thread.Status = state.StatusActive
	thread.Activity = "Running: git status"

	card := NewCardModel(thread, false)
	output := card.View()

	if !strings.Contains(output, "Running: git status") {
		t.Errorf("expected activity in output, got:\n%s", output)
	}
}

func TestCardRenderFallsBackToStatus(t *testing.T) {
	thread := state.NewThreadState("t-1", "o3-mini")
	thread.Status = state.StatusIdle

	card := NewCardModel(thread, false)
	output := card.View()

	hasActivity := strings.Contains(output, "Running")
	if hasActivity {
		t.Errorf("expected no activity when idle, got:\n%s", output)
	}
	if !strings.Contains(output, "idle") {
		t.Errorf("expected status fallback, got:\n%s", output)
	}
}

func TestCardRenderActivityTruncated(t *testing.T) {
	thread := state.NewThreadState("t-1", "o3-mini")
	thread.Status = state.StatusActive
	thread.Activity = "Running: npm install --save-dev @types/react @types/react-dom"

	card := NewCardModel(thread, false)
	card.SetSize(minCardWidth, minCardHeight)
	output := card.View()

	if !strings.Contains(output, "...") {
		t.Errorf("expected truncated activity, got:\n%s", output)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/tui -run TestCardRender(ShowsActivity|FallsBack|ActivityTruncated) -v`
Expected: FAIL — activity text not present in output

**Step 3: Write minimal implementation**

Modify `card.View()` in `internal/tui/card.go`. Replace the `statusLine` logic with an activity-aware check:

```go
func (card CardModel) View() string {
	statusColor, exists := statusColors[card.thread.Status]
	if !exists {
		statusColor = defaultStatusColor
	}

	secondLine := card.thread.Status
	hasActivity := card.thread.Activity != ""
	if hasActivity {
		secondLine = card.thread.Activity
	}

	styledSecondLine := lipgloss.NewStyle().
		Foreground(statusColor).
		Render(truncate(secondLine, card.width-cardBorderPadding))

	titleMaxLen := card.width - cardBorderPadding
	title := truncate(card.thread.Title, titleMaxLen)
	content := fmt.Sprintf("%s\n%s", title, styledSecondLine)

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

Note: This removes the pinned indicator rendering since the canvas does not use it (the session panel handles pinned display separately via the divider bar).

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/tui -run TestCardRender -v`
Expected: All PASS (including existing `TestCardRenderShowsTitle`, `TestCardRenderShowsStatus`, `TestCardRenderSelectedHighlight`, `TestCardDynamicSize`)

**Step 5: Commit**

```bash
git add internal/tui/card.go internal/tui/card_test.go
git commit -m "feat: render activity on session cards"
```

---

### Task 4: Add AgentReasoningDeltaMsg to bridge

**Files:**
- Modify: `internal/tui/msgs.go`
- Modify: `internal/tui/bridge.go`
- Modify: `internal/tui/bridge_test.go`

**Step 1: Write the failing test**

Add to `internal/tui/bridge_test.go`:

```go
func TestBridgeAgentReasoningDelta(t *testing.T) {
	event := appserver.ProtoEvent{
		Msg: json.RawMessage(`{"type":"agent_reasoning_delta","delta":"Let me think..."}`),
	}
	msg := ProtoEventToMsg(event)
	reasoning, ok := msg.(AgentReasoningDeltaMsg)
	if !ok {
		t.Fatalf("expected AgentReasoningDeltaMsg, got %T", msg)
	}
	if reasoning.Delta != "Let me think..." {
		t.Errorf("expected Let me think..., got %s", reasoning.Delta)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui -run TestBridgeAgentReasoningDelta -v`
Expected: FAIL — `AgentReasoningDeltaMsg` undefined

**Step 3: Write minimal implementation**

Add to `internal/tui/msgs.go`:

```go
type AgentReasoningDeltaMsg struct {
	Delta string
}
```

Add case to switch in `ProtoEventToMsg` in `internal/tui/bridge.go`:

```go
case appserver.EventAgentReasonDelta:
	return decodeReasoningDelta(event.Msg)
```

Add decoder function to `internal/tui/bridge.go`:

```go
func decodeReasoningDelta(raw json.RawMessage) tea.Msg {
	var delta appserver.AgentDelta
	if err := json.Unmarshal(raw, &delta); err != nil {
		return nil
	}
	return AgentReasoningDeltaMsg{Delta: delta.Delta}
}
```

Note: Reuses `appserver.AgentDelta` struct since reasoning deltas have the same `{"delta":"..."}` JSON shape.

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/tui -run TestBridge -v`
Expected: All PASS

**Step 5: Commit**

```bash
git add internal/tui/msgs.go internal/tui/bridge.go internal/tui/bridge_test.go
git commit -m "feat: decode agent_reasoning_delta events in bridge"
```

---

### Task 5: Wire event handlers to set activity

**Files:**
- Modify: `internal/tui/app_proto.go`
- Modify: `internal/tui/app.go` (add Update case for new msg type)

**Step 1: Check current Update switch for msg routing**

Read `internal/tui/app.go` to find the `Update` method's switch statement and identify where to add the new `AgentReasoningDeltaMsg` case.

**Step 2: Update handleTaskStarted to set activity**

In `internal/tui/app_proto.go`, modify `handleTaskStarted`:

```go
func (app AppModel) handleTaskStarted() (tea.Model, tea.Cmd) {
	if app.sessionID == "" {
		return app, nil
	}
	app.store.UpdateStatus(app.sessionID, state.StatusActive, "")
	app.store.UpdateActivity(app.sessionID, "Thinking...")
	messageID := fmt.Sprintf("msg-%d", time.Now().UnixNano())
	app.currentMessageID = messageID
	thread, exists := app.store.Get(app.sessionID)
	if exists {
		thread.AppendMessage(state.ChatMessage{
			ID:      messageID,
			Role:    "assistant",
			Content: "",
		})
	}
	return app, nil
}
```

**Step 3: Update handleAgentDelta to set activity snippet**

In `internal/tui/app_proto.go`, modify `handleAgentDelta`. After appending the delta, update activity with a snippet of the latest message content:

```go
func (app AppModel) handleAgentDelta(msg AgentDeltaMsg) (tea.Model, tea.Cmd) {
	noSession := app.sessionID == ""
	noMessage := app.currentMessageID == ""
	if noSession || noMessage {
		return app, nil
	}
	thread, exists := app.store.Get(app.sessionID)
	if !exists {
		return app, nil
	}
	thread.AppendDelta(app.currentMessageID, msg.Delta)
	snippet := latestMessageSnippet(thread, app.currentMessageID)
	app.store.UpdateActivity(app.sessionID, snippet)
	return app, nil
}
```

**Step 4: Add latestMessageSnippet helper**

Add to `internal/tui/app_proto.go`:

```go
const activitySnippetMaxLen = 40

func latestMessageSnippet(thread *state.ThreadState, messageID string) string {
	for index := range thread.Messages {
		if thread.Messages[index].ID != messageID {
			continue
		}
		content := thread.Messages[index].Content
		if len(content) <= activitySnippetMaxLen {
			return content
		}
		return content[len(content)-activitySnippetMaxLen:]
	}
	return ""
}
```

**Step 5: Add handleReasoningDelta handler**

Add to `internal/tui/app_proto.go`:

```go
func (app AppModel) handleReasoningDelta() (tea.Model, tea.Cmd) {
	if app.sessionID != "" {
		app.store.UpdateActivity(app.sessionID, "Thinking...")
	}
	return app, nil
}
```

**Step 6: Update handleExecApproval to set activity**

In `internal/tui/app_proto.go`, modify `handleExecApproval`:

```go
func (app AppModel) handleExecApproval(msg ExecApprovalRequestMsg) (tea.Model, tea.Cmd) {
	if app.sessionID != "" {
		activity := fmt.Sprintf("Running: %s", msg.Command)
		app.store.UpdateActivity(app.sessionID, activity)
	}
	if app.client != nil {
		app.client.SendApproval(msg.EventID, appserver.OpExecApproval, true)
	}
	return app, nil
}
```

**Step 7: Update handlePatchApproval to set activity**

In `internal/tui/app_proto.go`, modify `handlePatchApproval`:

```go
func (app AppModel) handlePatchApproval(msg PatchApprovalRequestMsg) (tea.Model, tea.Cmd) {
	if app.sessionID != "" {
		app.store.UpdateActivity(app.sessionID, "Applying patch...")
	}
	if app.client != nil {
		app.client.SendApproval(msg.EventID, appserver.OpPatchApproval, true)
	}
	return app, nil
}
```

**Step 8: Update handleAgentMessageCompleted to clear activity**

In `internal/tui/app_proto.go`, modify `handleAgentMessageCompleted`:

```go
func (app AppModel) handleAgentMessageCompleted() (tea.Model, tea.Cmd) {
	app.currentMessageID = ""
	if app.sessionID != "" {
		app.store.UpdateActivity(app.sessionID, "")
	}
	return app, nil
}
```

**Step 9: Update handleTaskComplete to clear activity**

In `internal/tui/app_proto.go`, modify `handleTaskComplete`:

```go
func (app AppModel) handleTaskComplete() (tea.Model, tea.Cmd) {
	if app.sessionID != "" {
		app.store.UpdateStatus(app.sessionID, state.StatusCompleted, "")
		app.store.UpdateActivity(app.sessionID, "")
	}
	app.currentMessageID = ""
	return app, nil
}
```

**Step 10: Add AgentReasoningDeltaMsg case to Update in app.go**

Find the switch statement in `app.go`'s `Update` method and add:

```go
case AgentReasoningDeltaMsg:
	return app.handleReasoningDelta()
```

**Step 11: Run all tests**

Run: `go test ./... -v -race`
Expected: All PASS

**Step 12: Run linter**

Run: `golangci-lint run`
Expected: No issues

**Step 13: Commit**

```bash
git add internal/tui/app_proto.go internal/tui/app.go
git commit -m "feat: wire protocol events to session card activity"
```

---

### Task 6: Verify build and full test suite

**Step 1: Build**

Run: `go build -o dj ./cmd/dj`
Expected: Build succeeds

**Step 2: Full test suite with race detector**

Run: `go test ./... -v -race`
Expected: All PASS

**Step 3: Lint**

Run: `golangci-lint run`
Expected: No issues

**Step 4: Final commit if any cleanup needed**

Only if lint or tests revealed issues to fix.
