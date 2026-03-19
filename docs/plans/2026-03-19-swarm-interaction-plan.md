# Swarm Interaction UI Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Wire up swarm keybindings (p, m, s) with persona picker, text input bar, and canvas filtering so users can spawn agents, message them, and view the swarm.

**Architecture:** Reuse the existing `MenuModel` for persona/agent pickers with a new `menuIntent` field to distinguish picker types from the thread context menu. Add a lightweight `InputBarModel` for single-line text entry that renders in place of the status bar. Canvas gains a `swarmFilter` mode that hides non-agent threads.

**Tech Stack:** Go, Bubble Tea, Lipgloss. No new dependencies.

---

### Task 1: InputBarModel — Component

**Files:**
- Create: `internal/tui/inputbar.go`
- Test: `internal/tui/inputbar_test.go`

**Step 1: Write the failing test**

Create `internal/tui/inputbar_test.go`:

```go
package tui

import (
	"strings"
	"testing"
)

const (
	testInputPrompt = "Task: "
	testInputValue  = "Design the API"
)

func TestInputBarView(testing *testing.T) {
	bar := NewInputBarModel(testInputPrompt)
	bar.InsertRune('H')
	bar.InsertRune('i')
	view := bar.View()
	if !strings.Contains(view, testInputPrompt) {
		testing.Error("expected prompt in view")
	}
	if !strings.Contains(view, "Hi") {
		testing.Error("expected typed value in view")
	}
}

func TestInputBarDeleteRune(testing *testing.T) {
	bar := NewInputBarModel(testInputPrompt)
	bar.InsertRune('A')
	bar.InsertRune('B')
	bar.DeleteRune()
	value := bar.Value()
	if value != "A" {
		testing.Errorf("expected 'A', got %q", value)
	}
}

func TestInputBarDeleteRuneEmpty(testing *testing.T) {
	bar := NewInputBarModel(testInputPrompt)
	bar.DeleteRune()
	value := bar.Value()
	if value != "" {
		testing.Errorf("expected empty, got %q", value)
	}
}

func TestInputBarValue(testing *testing.T) {
	bar := NewInputBarModel(testInputPrompt)
	bar.InsertRune('G')
	bar.InsertRune('o')
	value := bar.Value()
	if value != "Go" {
		testing.Errorf("expected 'Go', got %q", value)
	}
}

func TestInputBarReset(testing *testing.T) {
	bar := NewInputBarModel(testInputPrompt)
	bar.InsertRune('X')
	bar.Reset()
	value := bar.Value()
	if value != "" {
		testing.Errorf("expected empty after reset, got %q", value)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run "TestInputBar" -v`
Expected: FAIL — NewInputBarModel not defined

**Step 3: Write minimal implementation**

Create `internal/tui/inputbar.go`:

```go
package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	inputBarCursor   = "█"
	inputBarColorBg  = "236"
	inputBarColorFg  = "252"
	inputBarColorAcc = "39"
)

var (
	inputBarStyle = lipgloss.NewStyle().
		Background(lipgloss.Color(inputBarColorBg)).
		Foreground(lipgloss.Color(inputBarColorFg)).
		Padding(0, 1)
	inputBarPromptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(inputBarColorAcc)).
		Bold(true)
)

type InputBarModel struct {
	prompt string
	value  strings.Builder
}

func NewInputBarModel(prompt string) InputBarModel {
	return InputBarModel{prompt: prompt}
}

func (bar *InputBarModel) InsertRune(r rune) {
	bar.value.WriteRune(r)
}

func (bar *InputBarModel) DeleteRune() {
	current := bar.value.String()
	if len(current) == 0 {
		return
	}
	runes := []rune(current)
	bar.value.Reset()
	bar.value.WriteString(string(runes[:len(runes)-1]))
}

func (bar *InputBarModel) Value() string {
	return bar.value.String()
}

func (bar *InputBarModel) Reset() {
	bar.value.Reset()
}

func (bar InputBarModel) View() string {
	prompt := inputBarPromptStyle.Render(bar.prompt)
	text := bar.value.String() + inputBarCursor
	return inputBarStyle.Render(prompt + text)
}

func (bar InputBarModel) ViewWithWidth(width int) string {
	prompt := inputBarPromptStyle.Render(bar.prompt)
	text := bar.value.String() + inputBarCursor
	style := inputBarStyle.Width(width)
	return style.Render(prompt + text)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/ -run "TestInputBar" -v -race`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/inputbar.go internal/tui/inputbar_test.go
git commit -m "feat(tui): add InputBarModel component for text input"
```

---

### Task 2: Wire Header Swarm Hints

**Files:**
- Modify: `internal/tui/app.go:43-62`
- Test: `internal/tui/app_test.go`

**Step 1: Write the failing test**

Add to `internal/tui/app_test.go`:

```go
func TestNewAppModelPoolSetsSwarmActive(testing *testing.T) {
	store := state.NewThreadStore()
	agentPool := pool.NewAgentPool("codex", []string{"proto"}, nil, 10)
	app := NewAppModel(store, WithPool(agentPool))
	view := app.header.View()
	if !strings.Contains(view, "p: persona") {
		testing.Error("expected swarm hints in header when pool is set")
	}
}
```

Add `"strings"` to imports if not present. The constant `10` should use the existing test constant (add `testPoolMaxAgentsApp = 10` if needed, or reuse from existing tests).

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run TestNewAppModelPoolSetsSwarmActive -v`
Expected: FAIL — header does not contain swarm hints

**Step 3: Wire pool presence to header**

In `internal/tui/app.go`, in `NewAppModel`, after the `for _, opt := range opts` loop, add:

```go
	hasPool := app.pool != nil
	if hasPool {
		app.header.SetSwarmActive(true)
	}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/ -run TestNewAppModelPoolSetsSwarmActive -v -race`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/app.go internal/tui/app_test.go
git commit -m "feat(tui): wire header swarm hints when pool is present"
```

---

### Task 3: Menu Intent and Swarm Fields on AppModel

**Files:**
- Modify: `internal/tui/app.go:10-41`
- Modify: `internal/tui/msgs.go`
- Test: `internal/tui/app_test.go`

**Step 1: Write the failing test**

Add to `internal/tui/app_test.go`:

```go
func TestAppModelSwarmFieldsDefault(testing *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)
	if app.menuIntent != MenuIntentThread {
		testing.Error("expected default menu intent to be thread")
	}
	if app.inputBarVisible {
		testing.Error("expected input bar hidden by default")
	}
	if app.swarmFilter {
		testing.Error("expected swarm filter off by default")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run TestAppModelSwarmFieldsDefault -v`
Expected: FAIL — menuIntent, inputBarVisible, swarmFilter not defined

**Step 3: Add new types and fields**

In `internal/tui/msgs.go`, add the intent enums:

```go
type MenuIntent int

const (
	MenuIntentThread MenuIntent = iota
	MenuIntentPersonaPicker
	MenuIntentAgentPicker
)

type InputIntent int

const (
	IntentSpawnTask InputIntent = iota
	IntentSendMessage
)
```

In `internal/tui/app.go`, add fields to `AppModel`:

```go
	inputBar             InputBarModel
	inputBarVisible      bool
	inputBarIntent       InputIntent
	menuIntent           MenuIntent
	pendingPersonaID     string
	pendingTargetAgentID string
	swarmFilter          bool
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/ -run TestAppModelSwarmFieldsDefault -v -race`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/app.go internal/tui/msgs.go internal/tui/app_test.go
git commit -m "feat(tui): add menu intent, input intent, and swarm fields to AppModel"
```

---

### Task 4: showPersonaPicker Implementation

**Files:**
- Modify: `internal/tui/app_swarm.go:8-10`
- Test: `internal/tui/app_swarm_test.go`

**Step 1: Write the failing test**

Create `internal/tui/app_swarm_test.go`:

```go
package tui

import (
	"testing"

	"github.com/robinojw/dj/internal/pool"
	"github.com/robinojw/dj/internal/roster"
	"github.com/robinojw/dj/internal/state"
)

const testSwarmMaxAgents = 10

func TestShowPersonaPickerShowsMenu(testing *testing.T) {
	store := state.NewThreadStore()
	personas := []roster.PersonaDefinition{
		{ID: "architect", Name: "Architect"},
		{ID: "test", Name: "Test"},
	}
	agentPool := pool.NewAgentPool("echo", []string{}, personas, testSwarmMaxAgents)
	app := NewAppModel(store, WithPool(agentPool))

	updated, _ := app.showPersonaPicker()
	resultApp := updated.(AppModel)

	if !resultApp.menuVisible {
		testing.Error("expected menu to be visible after showPersonaPicker")
	}
	if resultApp.menuIntent != MenuIntentPersonaPicker {
		testing.Error("expected menu intent to be persona picker")
	}
}

func TestShowPersonaPickerNoPool(testing *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	updated, _ := app.showPersonaPicker()
	resultApp := updated.(AppModel)

	if resultApp.menuVisible {
		testing.Error("expected menu hidden when no pool")
	}
}

func TestShowPersonaPickerNoPersonas(testing *testing.T) {
	store := state.NewThreadStore()
	agentPool := pool.NewAgentPool("echo", []string{}, nil, testSwarmMaxAgents)
	app := NewAppModel(store, WithPool(agentPool))

	updated, _ := app.showPersonaPicker()
	resultApp := updated.(AppModel)

	if resultApp.menuVisible {
		testing.Error("expected menu hidden when no personas")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run "TestShowPersonaPicker" -v`
Expected: FAIL — showPersonaPicker is a no-op stub

**Step 3: Implement showPersonaPicker**

Replace the stub in `internal/tui/app_swarm.go`:

```go
func (app AppModel) showPersonaPicker() (tea.Model, tea.Cmd) {
	if app.pool == nil {
		return app, nil
	}

	personas := app.pool.Personas()
	if len(personas) == 0 {
		app.statusBar.SetError("No personas available")
		return app, nil
	}

	items := buildPersonaMenuItems(personas)
	app.menu = NewMenuModel("Spawn Persona Agent", items)
	app.menuVisible = true
	app.menuIntent = MenuIntentPersonaPicker
	return app, nil
}

func buildPersonaMenuItems(personas map[string]roster.PersonaDefinition) []MenuItem {
	var items []MenuItem
	for _, persona := range personas {
		items = append(items, MenuItem{
			Label: persona.Name,
			Key:   rune(persona.ID[0]),
		})
	}
	return items
}
```

Add import for `"github.com/robinojw/dj/internal/roster"`.

Note: The `MenuItem.Key` is the first rune of the persona ID — it's used for display only in the menu, not for selection logic.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/ -run "TestShowPersonaPicker" -v -race`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/app_swarm.go internal/tui/app_swarm_test.go
git commit -m "feat(tui): implement showPersonaPicker with persona menu"
```

---

### Task 5: handleInputBarKey and Key Priority

**Files:**
- Modify: `internal/tui/app_keys.go:9-27`
- Test: `internal/tui/app_keys_test.go` (or new `internal/tui/app_inputbar_test.go`)

**Step 1: Write the failing test**

Create `internal/tui/app_inputbar_test.go`:

```go
package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/state"
)

func TestHandleInputBarKeyTyping(testing *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)
	app.inputBarVisible = true
	app.inputBar = NewInputBarModel("Task: ")

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'H'}}
	updated, _ := app.Update(msg)
	resultApp := updated.(AppModel)

	value := resultApp.inputBar.Value()
	if value != "H" {
		testing.Errorf("expected 'H', got %q", value)
	}
}

func TestHandleInputBarKeyEscDismisses(testing *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)
	app.inputBarVisible = true
	app.inputBar = NewInputBarModel("Task: ")

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ := app.Update(msg)
	resultApp := updated.(AppModel)

	if resultApp.inputBarVisible {
		testing.Error("expected input bar dismissed on Esc")
	}
}

func TestHandleInputBarKeyBackspace(testing *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)
	app.inputBarVisible = true
	app.inputBar = NewInputBarModel("Task: ")
	app.inputBar.InsertRune('A')
	app.inputBar.InsertRune('B')

	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	updated, _ := app.Update(msg)
	resultApp := updated.(AppModel)

	value := resultApp.inputBar.Value()
	if value != "A" {
		testing.Errorf("expected 'A', got %q", value)
	}
}

func TestHandleInputBarKeyEnterEmptyDismisses(testing *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)
	app.inputBarVisible = true
	app.inputBar = NewInputBarModel("Task: ")

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(msg)
	resultApp := updated.(AppModel)

	if resultApp.inputBarVisible {
		testing.Error("expected input bar dismissed on empty Enter")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run "TestHandleInputBarKey" -v`
Expected: FAIL — input bar keystrokes are not handled (they fall through to canvas)

**Step 3: Add handleInputBarKey and wire into handleKey**

In `internal/tui/app_keys.go`, add the input bar check after help:

```go
func (app AppModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if app.helpVisible {
		return app.handleHelpKey(msg)
	}

	if app.inputBarVisible {
		return app.handleInputBarKey(msg)
	}

	if app.menuVisible {
		return app.handleMenuKey(msg)
	}

	if result, model, cmd := app.handlePrefix(msg); result {
		return model, cmd
	}

	if app.focusPane == FocusPaneSession {
		return app.handleSessionKey(msg)
	}

	return app.handleCanvasKey(msg)
}
```

Create a new function `handleInputBarKey` — either in `app_keys.go` or a new `app_inputbar.go` file (prefer `app_inputbar.go` to keep files focused):

Create `internal/tui/app_inputbar.go`:

```go
package tui

import tea "github.com/charmbracelet/bubbletea"

func (app AppModel) handleInputBarKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		return app.dismissInputBar()
	case tea.KeyEnter:
		return app.submitInputBar()
	case tea.KeyBackspace:
		app.inputBar.DeleteRune()
		return app, nil
	case tea.KeyRunes:
		for _, r := range msg.Runes {
			app.inputBar.InsertRune(r)
		}
		return app, nil
	}
	return app, nil
}

func (app AppModel) dismissInputBar() (tea.Model, tea.Cmd) {
	app.inputBarVisible = false
	app.pendingPersonaID = ""
	app.pendingTargetAgentID = ""
	app.inputBar.Reset()
	return app, nil
}

func (app AppModel) submitInputBar() (tea.Model, tea.Cmd) {
	value := app.inputBar.Value()
	isEmpty := value == ""
	if isEmpty {
		return app.dismissInputBar()
	}

	switch app.inputBarIntent {
	case IntentSpawnTask:
		return app.executeSpawn(value)
	case IntentSendMessage:
		return app.executeSendMessage(value)
	}
	return app.dismissInputBar()
}

func (app AppModel) executeSpawn(task string) (tea.Model, tea.Cmd) {
	app.inputBarVisible = false
	personaID := app.pendingPersonaID
	app.pendingPersonaID = ""
	app.inputBar.Reset()

	if app.pool == nil {
		return app, nil
	}

	agentID, err := app.pool.Spawn(personaID, task, "")
	if err != nil {
		app.statusBar.SetError(err.Error())
		return app, nil
	}

	app.store.Add(agentID, task)
	app.store.UpdateStatus(agentID, "active", "")
	app.statusBar.SetThreadCount(len(app.store.All()))
	app.tree.Refresh()
	return app, nil
}

func (app AppModel) executeSendMessage(content string) (tea.Model, tea.Cmd) {
	app.inputBarVisible = false
	targetID := app.pendingTargetAgentID
	app.pendingTargetAgentID = ""
	app.inputBar.Reset()

	if app.pool == nil {
		return app, nil
	}

	targetAgent, exists := app.pool.Get(targetID)
	if !exists {
		app.statusBar.SetError("Agent not found")
		return app, nil
	}

	if targetAgent.Client != nil {
		targetAgent.Client.SendUserInput(content)
	}
	return app, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/ -run "TestHandleInputBarKey" -v -race`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/app_keys.go internal/tui/app_inputbar.go internal/tui/app_inputbar_test.go
git commit -m "feat(tui): add input bar key handling with spawn and message dispatch"
```

---

### Task 6: Persona Picker Dispatch → Input Bar

**Files:**
- Modify: `internal/tui/app_menu.go:84-105`
- Modify: `internal/tui/app_swarm.go`
- Test: `internal/tui/app_swarm_test.go`

**Step 1: Write the failing test**

Add to `internal/tui/app_swarm_test.go`:

```go
func TestPersonaPickerDispatchShowsInputBar(testing *testing.T) {
	store := state.NewThreadStore()
	personas := []roster.PersonaDefinition{
		{ID: "architect", Name: "Architect"},
	}
	agentPool := pool.NewAgentPool("echo", []string{}, personas, testSwarmMaxAgents)
	app := NewAppModel(store, WithPool(agentPool))

	app.showPersonaPicker()

	selected := app.menu.Selected()
	app.menuVisible = false
	updated, _ := app.dispatchPersonaPick(selected)
	resultApp := updated.(AppModel)

	if !resultApp.inputBarVisible {
		testing.Error("expected input bar visible after persona pick")
	}
	if resultApp.inputBarIntent != IntentSpawnTask {
		testing.Error("expected spawn task intent")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run TestPersonaPickerDispatchShowsInputBar -v`
Expected: FAIL — dispatchPersonaPick not defined

**Step 3: Implement persona dispatch and route menu intent**

In `internal/tui/app_swarm.go`, add:

```go
func (app AppModel) dispatchPersonaPick(item MenuItem) (tea.Model, tea.Cmd) {
	persona := app.findPersonaByName(item.Label)
	if persona == nil {
		return app, nil
	}

	app.pendingPersonaID = persona.ID
	app.inputBar = NewInputBarModel("Task for " + persona.Name + ": ")
	app.inputBarVisible = true
	app.inputBarIntent = IntentSpawnTask
	return app, nil
}

func (app AppModel) findPersonaByName(name string) *roster.PersonaDefinition {
	if app.pool == nil {
		return nil
	}
	for _, persona := range app.pool.Personas() {
		if persona.Name == name {
			return &persona
		}
	}
	return nil
}
```

In `internal/tui/app_menu.go`, modify `handleMenuKey` Enter case to route via `menuIntent`:

Replace the Enter case in `handleMenuKey`:

```go
	case tea.KeyEnter:
		selected := app.menu.Selected()
		intent := app.menuIntent
		app.closeMenu()
		return app.dispatchMenuByIntent(intent, selected)
```

Add the routing function:

```go
func (app AppModel) dispatchMenuByIntent(intent MenuIntent, item MenuItem) (tea.Model, tea.Cmd) {
	switch intent {
	case MenuIntentPersonaPicker:
		return app.dispatchPersonaPick(item)
	case MenuIntentAgentPicker:
		return app.dispatchAgentPick(item)
	default:
		return app.dispatchMenuAction(item)
	}
}
```

Add a stub for `dispatchAgentPick` in `app_swarm.go`:

```go
func (app AppModel) dispatchAgentPick(item MenuItem) (tea.Model, tea.Cmd) {
	return app, nil
}
```

Also reset `menuIntent` in `closeMenu`:

```go
func (app *AppModel) closeMenu() {
	app.menuVisible = false
	app.menuIntent = MenuIntentThread
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/ -run "TestPersonaPickerDispatch" -v -race`
Expected: PASS

**Step 5: Run all tests to check nothing is broken**

Run: `go test ./internal/tui/ -v -race`
Expected: All PASS

**Step 6: Commit**

```bash
git add internal/tui/app_menu.go internal/tui/app_swarm.go internal/tui/app_swarm_test.go
git commit -m "feat(tui): route persona picker to input bar for task entry"
```

---

### Task 7: sendMessageToAgent + Agent Picker Dispatch

**Files:**
- Modify: `internal/tui/app_swarm.go`
- Test: `internal/tui/app_swarm_test.go`

**Step 1: Write the failing test**

Add to `internal/tui/app_swarm_test.go`:

```go
func TestSendMessageToAgentShowsMenu(testing *testing.T) {
	store := state.NewThreadStore()
	personas := []roster.PersonaDefinition{
		{ID: "architect", Name: "Architect"},
	}
	agentPool := pool.NewAgentPool("echo", []string{}, personas, testSwarmMaxAgents)
	agentPool.Spawn("architect", "Design API", "")
	app := NewAppModel(store, WithPool(agentPool))

	updated, _ := app.sendMessageToAgent()
	resultApp := updated.(AppModel)

	if !resultApp.menuVisible {
		testing.Error("expected menu visible for agent picker")
	}
	if resultApp.menuIntent != MenuIntentAgentPicker {
		testing.Error("expected agent picker intent")
	}
}

func TestSendMessageToAgentNoAgents(testing *testing.T) {
	store := state.NewThreadStore()
	agentPool := pool.NewAgentPool("echo", []string{}, nil, testSwarmMaxAgents)
	app := NewAppModel(store, WithPool(agentPool))

	updated, _ := app.sendMessageToAgent()
	resultApp := updated.(AppModel)

	if resultApp.menuVisible {
		testing.Error("expected menu hidden when no agents")
	}
}

func TestDispatchAgentPickShowsInputBar(testing *testing.T) {
	store := state.NewThreadStore()
	personas := []roster.PersonaDefinition{
		{ID: "architect", Name: "Architect"},
	}
	agentPool := pool.NewAgentPool("echo", []string{}, personas, testSwarmMaxAgents)
	agentID, _ := agentPool.Spawn("architect", "Design API", "")
	app := NewAppModel(store, WithPool(agentPool))

	item := MenuItem{Label: agentID, Key: 'a'}
	updated, _ := app.dispatchAgentPick(item)
	resultApp := updated.(AppModel)

	if !resultApp.inputBarVisible {
		testing.Error("expected input bar visible after agent pick")
	}
	if resultApp.inputBarIntent != IntentSendMessage {
		testing.Error("expected send message intent")
	}
	if resultApp.pendingTargetAgentID != agentID {
		testing.Errorf("expected target %s, got %s", agentID, resultApp.pendingTargetAgentID)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run "TestSendMessageToAgent|TestDispatchAgentPick" -v`
Expected: FAIL — sendMessageToAgent is a stub, dispatchAgentPick is a stub

**Step 3: Implement sendMessageToAgent and dispatchAgentPick**

Replace the stubs in `internal/tui/app_swarm.go`:

```go
func (app AppModel) sendMessageToAgent() (tea.Model, tea.Cmd) {
	if app.pool == nil {
		return app, nil
	}

	agents := app.pool.All()
	if len(agents) == 0 {
		app.statusBar.SetError("No active agents")
		return app, nil
	}

	items := buildAgentMenuItems(agents)
	app.menu = NewMenuModel("Message Agent", items)
	app.menuVisible = true
	app.menuIntent = MenuIntentAgentPicker
	return app, nil
}

func buildAgentMenuItems(agents []*pool.AgentProcess) []MenuItem {
	var items []MenuItem
	for _, agent := range agents {
		label := agent.ID
		if agent.Persona != nil {
			label = agent.Persona.Name + " (" + agent.ID + ")"
		}
		items = append(items, MenuItem{
			Label: label,
			Key:   rune(agent.ID[0]),
		})
	}
	return items
}

func (app AppModel) dispatchAgentPick(item MenuItem) (tea.Model, tea.Cmd) {
	agentID := extractAgentID(item.Label)
	app.pendingTargetAgentID = agentID
	app.inputBar = NewInputBarModel("Message to " + item.Label + ": ")
	app.inputBarVisible = true
	app.inputBarIntent = IntentSendMessage
	return app, nil
}
```

Add helper to extract agent ID from menu label:

```go
func extractAgentID(label string) string {
	parenStart := strings.LastIndex(label, "(")
	parenEnd := strings.LastIndex(label, ")")
	hasParen := parenStart != -1 && parenEnd > parenStart
	if hasParen {
		return label[parenStart+1 : parenEnd]
	}
	return label
}
```

Add imports for `"strings"` and `"github.com/robinojw/dj/internal/pool"`.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/ -run "TestSendMessageToAgent|TestDispatchAgentPick" -v -race`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/app_swarm.go internal/tui/app_swarm_test.go
git commit -m "feat(tui): implement agent picker and message dispatch"
```

---

### Task 8: Input Bar in View

**Files:**
- Modify: `internal/tui/app_view.go:22-47`
- Test: `internal/tui/app_test.go`

**Step 1: Write the failing test**

Add to `internal/tui/app_test.go`:

```go
func TestAppViewShowsInputBar(testing *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)
	app.inputBarVisible = true
	app.inputBar = NewInputBarModel("Task: ")
	app.width = 80
	app.height = 24

	view := app.View()
	if !strings.Contains(view, "Task: ") {
		testing.Error("expected input bar prompt in view")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run TestAppViewShowsInputBar -v`
Expected: FAIL — input bar not rendered in view

**Step 3: Add input bar rendering to View()**

In `internal/tui/app_view.go`, in the `View()` method, replace the status bar with the input bar when visible. The simplest approach is: after computing all sections, swap the status line.

Modify the `View()` function. After `status := app.statusBar.View()`, add:

```go
func (app AppModel) View() string {
	title := app.header.View()
	status := app.renderBottomBar()

	if app.helpVisible {
		return joinSections(title, app.help.View(), status)
	}
	// ... rest unchanged
}

func (app AppModel) renderBottomBar() string {
	if app.inputBarVisible {
		return app.inputBar.ViewWithWidth(app.width)
	}
	return app.statusBar.View()
}
```

Replace `app.statusBar.View()` with `app.renderBottomBar()` in the existing View method.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/ -run TestAppViewShowsInputBar -v -race`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/app_view.go internal/tui/app_test.go
git commit -m "feat(tui): render input bar in place of status bar"
```

---

### Task 9: Canvas Swarm Filter

**Files:**
- Modify: `internal/tui/canvas.go:16-22,109-120`
- Test: `internal/tui/canvas_test.go`

**Step 1: Write the failing test**

Add to `internal/tui/canvas_test.go`:

```go
func TestCanvasSwarmFilter(testing *testing.T) {
	store := state.NewThreadStore()
	store.Add("t1", "Regular Thread")
	store.Add("t2", "Agent Thread")
	thread2, _ := store.Get("t2")
	thread2.AgentProcessID = "architect-1"

	canvas := NewCanvasModel(store)
	canvas.SetSwarmFilter(true)

	filtered := canvas.filteredThreads()
	if len(filtered) != 1 {
		testing.Errorf("expected 1 agent thread, got %d", len(filtered))
	}
	if filtered[0].ID != "t2" {
		testing.Errorf("expected t2, got %s", filtered[0].ID)
	}
}

func TestCanvasSwarmFilterOff(testing *testing.T) {
	store := state.NewThreadStore()
	store.Add("t1", "Regular Thread")
	store.Add("t2", "Agent Thread")

	canvas := NewCanvasModel(store)
	canvas.SetSwarmFilter(false)

	filtered := canvas.filteredThreads()
	if len(filtered) != 2 {
		testing.Errorf("expected 2 threads, got %d", len(filtered))
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run "TestCanvasSwarmFilter" -v`
Expected: FAIL — SetSwarmFilter and filteredThreads not defined

**Step 3: Add swarm filter to CanvasModel**

In `internal/tui/canvas.go`, add field and methods:

```go
type CanvasModel struct {
	store       *state.ThreadStore
	selected    int
	pinnedIDs   map[string]bool
	width       int
	height      int
	swarmFilter bool
}

func (canvas *CanvasModel) SetSwarmFilter(enabled bool) {
	canvas.swarmFilter = enabled
}

func (canvas *CanvasModel) filteredThreads() []*state.ThreadState {
	threads := canvas.store.TreeOrder()
	if !canvas.swarmFilter {
		return threads
	}

	var filtered []*state.ThreadState
	for _, thread := range threads {
		isAgent := thread.AgentProcessID != ""
		if isAgent {
			filtered = append(filtered, thread)
		}
	}
	return filtered
}
```

Update `View()` and `SelectedThreadID()` to use `filteredThreads()` instead of `canvas.store.TreeOrder()`:

```go
func (canvas *CanvasModel) View() string {
	threads := canvas.filteredThreads()
	if len(threads) == 0 {
		return canvas.renderEmpty()
	}

	grid := canvas.renderGrid(threads)
	if canvas.hasDimensions() {
		return canvas.centerContent(grid)
	}
	return grid
}

func (canvas *CanvasModel) SelectedThreadID() string {
	threads := canvas.filteredThreads()
	if len(threads) == 0 {
		return ""
	}
	clampedIndex := canvas.selected
	if clampedIndex >= len(threads) {
		clampedIndex = len(threads) - 1
	}
	return threads[clampedIndex].ID
}
```

Also update `MoveRight` and `MoveDown` to use `filteredThreads()`:

```go
func (canvas *CanvasModel) MoveRight() {
	threads := canvas.filteredThreads()
	if canvas.selected < len(threads)-1 {
		canvas.selected++
	}
}

func (canvas *CanvasModel) MoveDown() {
	threads := canvas.filteredThreads()
	next := canvas.selected + canvasColumns
	if next < len(threads) {
		canvas.selected = next
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/ -run "TestCanvasSwarmFilter" -v -race`
Expected: PASS

**Step 5: Run all canvas tests to check nothing is broken**

Run: `go test ./internal/tui/ -run "TestCanvas" -v -race`
Expected: All PASS

**Step 6: Commit**

```bash
git add internal/tui/canvas.go internal/tui/canvas_test.go
git commit -m "feat(tui): add swarm filter to canvas"
```

---

### Task 10: toggleSwarmView Implementation

**Files:**
- Modify: `internal/tui/app_swarm.go`
- Test: `internal/tui/app_swarm_test.go`

**Step 1: Write the failing test**

Add to `internal/tui/app_swarm_test.go`:

```go
func TestToggleSwarmViewFiltersCanvas(testing *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	updated, _ := app.toggleSwarmView()
	resultApp := updated.(AppModel)

	if !resultApp.swarmFilter {
		testing.Error("expected swarm filter enabled after toggle")
	}

	updated2, _ := resultApp.toggleSwarmView()
	resultApp2 := updated2.(AppModel)

	if resultApp2.swarmFilter {
		testing.Error("expected swarm filter disabled after second toggle")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run TestToggleSwarmViewFiltersCanvas -v`
Expected: FAIL — toggleSwarmView is a stub

**Step 3: Implement toggleSwarmView**

Replace the stub in `internal/tui/app_swarm.go`:

```go
func (app AppModel) toggleSwarmView() (tea.Model, tea.Cmd) {
	app.swarmFilter = !app.swarmFilter
	app.canvas.SetSwarmFilter(app.swarmFilter)
	return app, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/ -run TestToggleSwarmViewFiltersCanvas -v -race`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/app_swarm.go internal/tui/app_swarm_test.go
git commit -m "feat(tui): implement toggleSwarmView with canvas filtering"
```

---

### Task 11: Full Build Verification and Lint

**Step 1: Run full test suite**

Run: `go test ./... -v -race`
Expected: All PASS

**Step 2: Run linter**

Run: `golangci-lint run`
Expected: No errors (fix any funlen/cyclop violations by extracting helpers)

**Step 3: Run build**

Run: `go build -o dj ./cmd/dj`
Expected: Build succeeds

**Step 4: Fix any lint violations**

If `funlen` or `cyclop` flags functions as too long/complex, extract helper functions to stay within the 60-line / 15-complexity limits.

**Step 5: Final commit**

```bash
git add -A
git commit -m "chore: fix lint violations and verify full build"
```

---

## Summary

| Task | Package | Description |
|------|---------|-------------|
| 1 | `internal/tui/` | InputBarModel component with insert/delete/reset/view |
| 2 | `internal/tui/` | Wire header swarm hints when pool is present |
| 3 | `internal/tui/` | MenuIntent, InputIntent enums and new AppModel fields |
| 4 | `internal/tui/` | showPersonaPicker builds persona menu from pool |
| 5 | `internal/tui/` | handleInputBarKey + key priority + spawn/message dispatch |
| 6 | `internal/tui/` | Persona picker dispatch → input bar for task entry |
| 7 | `internal/tui/` | sendMessageToAgent + agent picker dispatch |
| 8 | `internal/tui/` | Input bar renders in place of status bar |
| 9 | `internal/tui/` | Canvas swarm filter hides non-agent threads |
| 10 | `internal/tui/` | toggleSwarmView flips canvas filter |
| 11 | — | Full build verification and lint |
