# Split Layout & Pinned Sessions Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the current single-session fullscreen view with a persistent top/bottom split layout where multiple sessions can be pinned side-by-side in a bottom panel, with a focus state machine routing keys between canvas and session panes.

**Architecture:** The `AppModel` gains an ordered `pinnedSessions []string` slice and a `FocusPane` enum replacing the current `focus int`. The `View()` always renders canvas on top, and when sessions are pinned, renders a divider + horizontal session panel below. A `SessionPanelModel` sub-model owns the panel state to keep `AppModel` lean. PTY resize is triggered whenever the pinned set or terminal size changes.

**Tech Stack:** Go, Bubble Tea, Lipgloss, creack/pty, charmbracelet/x/vt

---

## Task 1: Add new message types for pin/unpin/focus

New messages that the rest of the system will produce and `Update()` will consume.

**Files:**
- Modify: `internal/tui/msgs.go`
- Test: `internal/tui/msgs_test.go`

**Step 1: Write the failing test**

Add to `internal/tui/msgs_test.go`:

```go
func TestPinUnpinMessages(t *testing.T) {
	pinMsg := PinSessionMsg{ThreadID: "t-1"}
	if pinMsg.ThreadID != "t-1" {
		t.Errorf("expected t-1, got %s", pinMsg.ThreadID)
	}

	unpinMsg := UnpinSessionMsg{ThreadID: "t-2"}
	if unpinMsg.ThreadID != "t-2" {
		t.Errorf("expected t-2, got %s", unpinMsg.ThreadID)
	}

	focusMsg := FocusSessionPaneMsg{Index: 2}
	if focusMsg.Index != 2 {
		t.Errorf("expected 2, got %d", focusMsg.Index)
	}

	switchMsg := SwitchPaneFocusMsg{Pane: FocusPaneSession}
	if switchMsg.Pane != FocusPaneSession {
		t.Errorf("expected FocusPaneSession, got %d", switchMsg.Pane)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui -run TestPinUnpinMessages -v`
Expected: FAIL — `PinSessionMsg`, `UnpinSessionMsg`, `FocusSessionPaneMsg`, `SwitchPaneFocusMsg`, `FocusPaneSession` undefined

**Step 3: Write minimal implementation**

Add to `internal/tui/msgs.go` (after the existing `PTYOutputMsg`):

```go
type FocusPane int

const (
	FocusPaneCanvas  FocusPane = iota
	FocusPaneSession
)

type PinSessionMsg struct {
	ThreadID string
}

type UnpinSessionMsg struct {
	ThreadID string
}

type FocusSessionPaneMsg struct {
	Index int
}

type SwitchPaneFocusMsg struct {
	Pane FocusPane
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui -run TestPinUnpinMessages -v`
Expected: PASS

**Step 5: Run full test suite**

Run: `go test ./... -v -race`
Expected: All 135 tests pass

**Step 6: Commit**

```bash
git add internal/tui/msgs.go internal/tui/msgs_test.go
git commit -m "feat(tui): add pin/unpin/focus message types for session panel"
```

---

## Task 2: Create SessionPanelModel sub-model

Extracts panel state from AppModel into its own sub-model. This is where `pinnedSessions`, `activePaneIdx`, and `splitRatio` live.

**Files:**
- Create: `internal/tui/session_panel.go`
- Create: `internal/tui/session_panel_test.go`

**Step 1: Write the failing test**

Create `internal/tui/session_panel_test.go`:

```go
package tui

import "testing"

func TestSessionPanelPinAddsThread(t *testing.T) {
	panel := NewSessionPanelModel()
	panel.Pin("t-1")

	if len(panel.PinnedSessions()) != 1 {
		t.Fatalf("expected 1 pinned session, got %d", len(panel.PinnedSessions()))
	}
	if panel.PinnedSessions()[0] != "t-1" {
		t.Errorf("expected t-1, got %s", panel.PinnedSessions()[0])
	}
}

func TestSessionPanelPinIgnoresDuplicate(t *testing.T) {
	panel := NewSessionPanelModel()
	panel.Pin("t-1")
	panel.Pin("t-1")

	if len(panel.PinnedSessions()) != 1 {
		t.Errorf("expected 1 pinned session, got %d", len(panel.PinnedSessions()))
	}
}

func TestSessionPanelUnpin(t *testing.T) {
	panel := NewSessionPanelModel()
	panel.Pin("t-1")
	panel.Pin("t-2")
	panel.Unpin("t-1")

	if len(panel.PinnedSessions()) != 1 {
		t.Fatalf("expected 1, got %d", len(panel.PinnedSessions()))
	}
	if panel.PinnedSessions()[0] != "t-2" {
		t.Errorf("expected t-2, got %s", panel.PinnedSessions()[0])
	}
}

func TestSessionPanelUnpinClampsFocus(t *testing.T) {
	panel := NewSessionPanelModel()
	panel.Pin("t-1")
	panel.Pin("t-2")
	panel.SetActivePaneIdx(1)
	panel.Unpin("t-2")

	if panel.ActivePaneIdx() != 0 {
		t.Errorf("expected clamped to 0, got %d", panel.ActivePaneIdx())
	}
}

func TestSessionPanelCycleRight(t *testing.T) {
	panel := NewSessionPanelModel()
	panel.Pin("t-1")
	panel.Pin("t-2")
	panel.Pin("t-3")

	panel.CycleRight()
	if panel.ActivePaneIdx() != 1 {
		t.Errorf("expected 1, got %d", panel.ActivePaneIdx())
	}

	panel.CycleRight()
	panel.CycleRight()
	if panel.ActivePaneIdx() != 2 {
		t.Errorf("expected clamped to 2, got %d", panel.ActivePaneIdx())
	}
}

func TestSessionPanelCycleLeft(t *testing.T) {
	panel := NewSessionPanelModel()
	panel.Pin("t-1")
	panel.Pin("t-2")
	panel.SetActivePaneIdx(1)

	panel.CycleLeft()
	if panel.ActivePaneIdx() != 0 {
		t.Errorf("expected 0, got %d", panel.ActivePaneIdx())
	}

	panel.CycleLeft()
	if panel.ActivePaneIdx() != 0 {
		t.Errorf("expected clamped to 0, got %d", panel.ActivePaneIdx())
	}
}

func TestSessionPanelActiveThreadID(t *testing.T) {
	panel := NewSessionPanelModel()
	if panel.ActiveThreadID() != "" {
		t.Errorf("expected empty, got %s", panel.ActiveThreadID())
	}

	panel.Pin("t-1")
	panel.Pin("t-2")
	panel.SetActivePaneIdx(1)

	if panel.ActiveThreadID() != "t-2" {
		t.Errorf("expected t-2, got %s", panel.ActiveThreadID())
	}
}

func TestSessionPanelIsPinned(t *testing.T) {
	panel := NewSessionPanelModel()
	panel.Pin("t-1")

	if !panel.IsPinned("t-1") {
		t.Error("expected t-1 to be pinned")
	}
	if panel.IsPinned("t-2") {
		t.Error("expected t-2 to not be pinned")
	}
}

func TestSessionPanelSplitRatio(t *testing.T) {
	panel := NewSessionPanelModel()
	if panel.SplitRatio() != defaultSplitRatio {
		t.Errorf("expected %f, got %f", defaultSplitRatio, panel.SplitRatio())
	}
}

func TestSessionPanelSessionDimensions(t *testing.T) {
	panel := NewSessionPanelModel()
	panel.Pin("t-1")
	panel.Pin("t-2")

	width, height := panel.SessionDimensions(120, 40)
	expectedWidth := 120 / 2
	expectedHeight := 40 - dividerHeight
	if width != expectedWidth {
		t.Errorf("expected width %d, got %d", expectedWidth, width)
	}
	if height != expectedHeight {
		t.Errorf("expected height %d, got %d", expectedHeight, height)
	}
}

func TestSessionPanelSessionDimensionsEmpty(t *testing.T) {
	panel := NewSessionPanelModel()
	width, height := panel.SessionDimensions(120, 40)
	if width != 0 || height != 0 {
		t.Errorf("expected 0,0 for empty panel, got %d,%d", width, height)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui -run TestSessionPanel -v`
Expected: FAIL — `NewSessionPanelModel` undefined

**Step 3: Write minimal implementation**

Create `internal/tui/session_panel.go`:

```go
package tui

const (
	defaultSplitRatio = 0.5
	dividerHeight     = 1
)

type SessionPanelModel struct {
	pinnedSessions []string
	activePaneIdx  int
	splitRatio     float64
}

func NewSessionPanelModel() SessionPanelModel {
	return SessionPanelModel{
		splitRatio: defaultSplitRatio,
	}
}

func (panel *SessionPanelModel) Pin(threadID string) {
	if panel.IsPinned(threadID) {
		return
	}
	panel.pinnedSessions = append(panel.pinnedSessions, threadID)
}

func (panel *SessionPanelModel) Unpin(threadID string) {
	filtered := make([]string, 0, len(panel.pinnedSessions))
	for _, pinned := range panel.pinnedSessions {
		if pinned != threadID {
			filtered = append(filtered, pinned)
		}
	}
	panel.pinnedSessions = filtered
	panel.clampActivePaneIdx()
}

func (panel *SessionPanelModel) IsPinned(threadID string) bool {
	for _, pinned := range panel.pinnedSessions {
		if pinned == threadID {
			return true
		}
	}
	return false
}

func (panel *SessionPanelModel) PinnedSessions() []string {
	return panel.pinnedSessions
}

func (panel *SessionPanelModel) ActivePaneIdx() int {
	return panel.activePaneIdx
}

func (panel *SessionPanelModel) SetActivePaneIdx(index int) {
	panel.activePaneIdx = index
	panel.clampActivePaneIdx()
}

func (panel *SessionPanelModel) ActiveThreadID() string {
	if len(panel.pinnedSessions) == 0 {
		return ""
	}
	return panel.pinnedSessions[panel.activePaneIdx]
}

func (panel *SessionPanelModel) CycleRight() {
	maxIdx := len(panel.pinnedSessions) - 1
	if panel.activePaneIdx < maxIdx {
		panel.activePaneIdx++
	}
}

func (panel *SessionPanelModel) CycleLeft() {
	if panel.activePaneIdx > 0 {
		panel.activePaneIdx--
	}
}

func (panel SessionPanelModel) SplitRatio() float64 {
	return panel.splitRatio
}

func (panel SessionPanelModel) SessionDimensions(panelWidth int, panelHeight int) (int, int) {
	count := len(panel.pinnedSessions)
	if count == 0 {
		return 0, 0
	}
	sessionWidth := panelWidth / count
	sessionHeight := panelHeight - dividerHeight
	return sessionWidth, sessionHeight
}

func (panel *SessionPanelModel) clampActivePaneIdx() {
	maxIdx := len(panel.pinnedSessions) - 1
	if maxIdx < 0 {
		panel.activePaneIdx = 0
		return
	}
	if panel.activePaneIdx > maxIdx {
		panel.activePaneIdx = maxIdx
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui -run TestSessionPanel -v`
Expected: All SessionPanel tests pass

**Step 5: Run full test suite**

Run: `go test ./... -v -race`
Expected: All tests pass (no regressions)

**Step 6: Commit**

```bash
git add internal/tui/session_panel.go internal/tui/session_panel_test.go
git commit -m "feat(tui): add SessionPanelModel for pinned session management"
```

---

## Task 3: Integrate SessionPanelModel into AppModel

Replace the single `session *SessionModel` field with `sessionPanel SessionPanelModel` and update the `focusPane` field. Migrate existing single-session behavior to work through the panel. Preserve all existing tests.

**Files:**
- Modify: `internal/tui/app.go`
- Modify: `internal/tui/app_pty.go`
- Modify: `internal/tui/app_view.go`
- Modify: `internal/tui/app_test.go`

**Step 1: Write the failing test**

Add to `internal/tui/app_test.go`:

```go
func TestAppHasPinnedSessions(t *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	if len(app.sessionPanel.PinnedSessions()) != 0 {
		t.Errorf("expected 0 pinned sessions, got %d", len(app.sessionPanel.PinnedSessions()))
	}
}

func TestAppFocusPaneDefaultsToCanvas(t *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	if app.focusPane != FocusPaneCanvas {
		t.Errorf("expected FocusPaneCanvas, got %d", app.focusPane)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui -run "TestAppHasPinnedSessions|TestAppFocusPaneDefaultsToCanvas" -v`
Expected: FAIL — `app.sessionPanel` and `app.focusPane` undefined

**Step 3: Refactor AppModel struct**

In `internal/tui/app.go`, replace these fields:

```go
// REMOVE these fields from AppModel:
// session          *SessionModel
// focus            int

// ADD these fields to AppModel:
sessionPanel    SessionPanelModel
focusPane       FocusPane
```

The full updated `AppModel` struct:

```go
type AppModel struct {
	store            *state.ThreadStore
	client           *appserver.Client
	statusBar        *StatusBar
	canvas           CanvasModel
	tree             TreeModel
	prefix           *PrefixHandler
	menu             MenuModel
	help             HelpModel
	menuVisible      bool
	helpVisible      bool
	focusPane        FocusPane
	canvasMode       int
	width            int
	height           int
	sessionID        string
	currentMessageID string
	events           chan appserver.ProtoEvent
	ptySessions      map[string]*PTYSession
	ptyEvents        chan PTYOutputMsg
	interactiveCmd   string
	interactiveArgs  []string
	sessionPanel     SessionPanelModel
}
```

Keep the focus constants `FocusCanvas`, `FocusTree` as `canvasMode` values (rename the field for clarity):

```go
const (
	CanvasModeGrid = iota
	CanvasModeTree
)
```

Update `NewAppModel`:

```go
func NewAppModel(store *state.ThreadStore, opts ...AppOption) AppModel {
	app := AppModel{
		store:        store,
		statusBar:    NewStatusBar(),
		canvas:       NewCanvasModel(store),
		tree:         NewTreeModel(store),
		prefix:       NewPrefixHandler(),
		help:         NewHelpModel(),
		events:       make(chan appserver.ProtoEvent, eventChannelSize),
		ptySessions:  make(map[string]*PTYSession),
		ptyEvents:    make(chan PTYOutputMsg, eventChannelSize),
		sessionPanel: NewSessionPanelModel(),
	}
	for _, opt := range opts {
		opt(&app)
	}
	return app
}
```

**Step 4: Update Focus() accessor**

Replace `Focus()` with `FocusPane()`:

```go
func (app AppModel) FocusPane() FocusPane {
	return app.focusPane
}

func (app AppModel) CanvasMode() int {
	return app.canvasMode
}
```

**Step 5: Update handleKey routing**

In `internal/tui/app.go`, update `handleKey`:

```go
func (app AppModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if app.helpVisible {
		return app.handleHelpKey(msg)
	}

	if app.menuVisible {
		return app.handleMenuKey(msg)
	}

	prefixResult := app.prefix.HandleKey(msg)
	switch prefixResult {
	case PrefixWaiting:
		return app, nil
	case PrefixComplete:
		return app.handlePrefixAction()
	case PrefixCancelled:
		return app, nil
	}

	if app.focusPane == FocusPaneSession {
		return app.handleSessionKey(msg)
	}

	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		return app, tea.Quit
	case tea.KeyEnter:
		return app.openSession()
	case tea.KeyRunes:
		return app.handleRune(msg)
	default:
		return app.handleArrow(msg)
	}
}
```

**Step 6: Update toggleFocus to use canvasMode**

```go
func (app *AppModel) toggleCanvasMode() {
	if app.canvasMode == CanvasModeGrid {
		app.canvasMode = CanvasModeTree
		return
	}
	app.canvasMode = CanvasModeGrid
}
```

Update `handleRune` to call `toggleCanvasMode()` instead of `toggleFocus()`.

Update `handleArrow` to check `app.canvasMode == CanvasModeTree` instead of `app.focus == FocusTree`.

**Step 7: Update app_pty.go**

In `internal/tui/app_pty.go`:

- `openSession()`: Instead of setting `app.session` and `app.focus`, pin the session and set focus:

```go
func (app AppModel) openSession() (tea.Model, tea.Cmd) {
	threadID := app.canvas.SelectedThreadID()
	if threadID == "" {
		return app, nil
	}

	thread, exists := app.store.Get(threadID)
	if !exists {
		return app, nil
	}

	if !app.sessionPanel.IsPinned(threadID) {
		existingPTY, hasExisting := app.ptySessions[threadID]
		if !hasExisting {
			ptySession := NewPTYSession(PTYSessionConfig{
				ThreadID: threadID,
				Command:  app.resolveInteractiveCmd(),
				Args:     app.interactiveArgs,
				SendMsg:  app.ptyEventCallback(),
			})
			if err := ptySession.Start(); err != nil {
				app.statusBar.SetError(err.Error())
				return app, nil
			}
			app.ptySessions[threadID] = ptySession
			existingPTY = ptySession
		}
		_ = thread
		app.sessionPanel.Pin(threadID)
	}

	app.focusPane = FocusPaneSession
	app.sessionPanel.SetActivePaneIdx(app.pinnedIndex(threadID))
	return app, app.rebalancePTYSizes()
}
```

Add helper:

```go
func (app AppModel) pinnedIndex(threadID string) int {
	for index, pinned := range app.sessionPanel.PinnedSessions() {
		if pinned == threadID {
			return index
		}
	}
	return 0
}
```

- `closeSession()`: Return focus to canvas without unpinning:

```go
func (app *AppModel) closeSession() {
	app.focusPane = FocusPaneCanvas
}
```

- `handleSessionKey()`: Forward keys to the active pinned PTY:

```go
func (app AppModel) handleSessionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return app, tea.Quit
	case tea.KeyEsc:
		app.closeSession()
		return app, nil
	default:
		return app.forwardKeyToPTY(msg)
	}
}

func (app AppModel) forwardKeyToPTY(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	activeID := app.sessionPanel.ActiveThreadID()
	if activeID == "" {
		return app, nil
	}

	ptySession, exists := app.ptySessions[activeID]
	if !exists {
		return app, nil
	}

	data := KeyMsgToBytes(msg)
	if data == nil {
		return app, nil
	}

	ptySession.WriteBytes(data)
	return app, nil
}
```

- `handlePTYOutput()`: Check if the exited session is pinned:

```go
func (app AppModel) handlePTYOutput(msg PTYOutputMsg) (tea.Model, tea.Cmd) {
	return app, app.listenForPTYEvents()
}
```

- Add `rebalancePTYSizes()`:

```go
func (app AppModel) rebalancePTYSizes() tea.Cmd {
	pinned := app.sessionPanel.PinnedSessions()
	if len(pinned) == 0 {
		return nil
	}

	canvasHeight := int(float64(app.height) * app.sessionPanel.SplitRatio())
	panelHeight := app.height - canvasHeight - dividerHeight
	sessionWidth, _ := app.sessionPanel.SessionDimensions(app.width, panelHeight)

	for _, threadID := range pinned {
		ptySession, exists := app.ptySessions[threadID]
		if exists {
			ptySession.Resize(sessionWidth, panelHeight)
		}
	}
	return nil
}
```

**Step 8: Update app_view.go**

In `internal/tui/app_view.go`, update `View()`:

```go
func (app AppModel) View() string {
	title := titleStyle.Render("DJ — Codex TUI Visualizer")
	status := app.statusBar.View()

	if app.helpVisible {
		return title + "\n" + app.help.View() + "\n" + status
	}

	if app.menuVisible {
		return title + "\n" + app.menu.View() + "\n" + status
	}

	canvas := app.renderCanvas()
	hasPinned := len(app.sessionPanel.PinnedSessions()) > 0

	if !hasPinned {
		return title + "\n" + canvas + "\n" + status
	}

	divider := app.renderDivider()
	panel := app.renderSessionPanel()
	return title + "\n" + canvas + "\n" + divider + "\n" + panel + "\n" + status
}

func (app AppModel) renderCanvas() string {
	canvas := app.canvas.View()
	if app.canvasMode == CanvasModeTree {
		treeView := app.tree.View()
		return lipgloss.JoinHorizontal(lipgloss.Top, treeView+"  ", canvas)
	}
	return canvas
}

func (app AppModel) renderDivider() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Width(app.width).
		Render(strings.Repeat("─", app.width))
}

func (app AppModel) renderSessionPanel() string {
	pinned := app.sessionPanel.PinnedSessions()
	count := len(pinned)
	if count == 0 {
		return ""
	}

	canvasHeight := int(float64(app.height) * app.sessionPanel.SplitRatio())
	panelHeight := app.height - canvasHeight - dividerHeight
	sessionWidth := app.width / count

	panes := make([]string, count)
	for index, threadID := range pinned {
		content := ""
		ptySession, exists := app.ptySessions[threadID]
		if exists {
			content = ptySession.Render()
		}

		isActive := index == app.sessionPanel.ActivePaneIdx() && app.focusPane == FocusPaneSession
		style := app.sessionPaneStyle(sessionWidth, panelHeight, isActive)
		panes[index] = style.Render(content)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, panes...)
}

func (app AppModel) sessionPaneStyle(width int, height int, active bool) lipgloss.Style {
	borderColor := lipgloss.Color("240")
	if active {
		borderColor = lipgloss.Color("39")
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width - 2).
		Height(height - 2)
}
```

**Step 9: Update WindowSizeMsg handler**

In `app.go` `Update()`, update the `WindowSizeMsg` case:

```go
case tea.WindowSizeMsg:
	app.width = msg.Width
	app.height = msg.Height
	app.statusBar.SetWidth(msg.Width)
	return app, app.rebalancePTYSizes()
```

**Step 10: Update existing tests**

All existing tests that reference `app.Focus()` must be updated to use `app.FocusPane()` or `app.CanvasMode()`. Tests that check `FocusSession` must check `FocusPaneSession`. Tests that check `FocusCanvas` must check `FocusPaneCanvas`. Tests that check `FocusTree` must check `app.CanvasMode() == CanvasModeTree`.

Key test updates in `app_test.go`:
- `TestAppToggleFocus`: Change `app.Focus() == FocusTree` → `appModel.CanvasMode() == CanvasModeTree`
- `TestAppTreeNavigationWhenFocused`: Change the focus setup and assertions
- `TestAppEnterOpensSession`: Check `app.FocusPane() == FocusPaneSession` and `len(app.sessionPanel.PinnedSessions()) == 1`
- `TestAppEscClosesSession`: Check `app.FocusPane() == FocusPaneCanvas` (pinned sessions remain)
- `TestAppReconnectsExistingPTY`: Check pinned session count stays 1
- `TestAppForwardKeyToPTY`: Update to work through pinned panel

**Step 11: Run all tests**

Run: `go test ./... -v -race`
Expected: All tests pass

**Step 12: Run linter**

Run: `golangci-lint run`
Expected: No issues (all files under 300 lines, functions under 60 lines)

**Step 13: Commit**

```bash
git add internal/tui/app.go internal/tui/app_view.go internal/tui/app_pty.go internal/tui/app_test.go
git commit -m "refactor(tui): integrate SessionPanelModel, replace single-session with pinned panel"
```

---

## Task 4: Add Space key to toggle pin from canvas

When the user presses Space on a canvas card, it pins or unpins that thread's session in the bottom panel.

**Files:**
- Modify: `internal/tui/app.go` (handleRune)
- Modify: `internal/tui/app_pty.go` (togglePin)
- Modify: `internal/tui/app_test.go`

**Step 1: Write the failing test**

Add to `internal/tui/app_test.go`:

```go
func TestAppSpacePinsSession(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Thread 1")
	app := NewAppModel(store, WithInteractiveCommand("echo", "hello"))

	spaceKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := app.Update(spaceKey)
	appModel := updated.(AppModel)
	defer appModel.StopAllPTYSessions()

	if len(appModel.sessionPanel.PinnedSessions()) != 1 {
		t.Fatalf("expected 1 pinned, got %d", len(appModel.sessionPanel.PinnedSessions()))
	}
	if appModel.sessionPanel.PinnedSessions()[0] != "t-1" {
		t.Errorf("expected t-1, got %s", appModel.sessionPanel.PinnedSessions()[0])
	}
}

func TestAppSpaceUnpinsSession(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Thread 1")
	app := NewAppModel(store, WithInteractiveCommand("echo", "hello"))

	spaceKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := app.Update(spaceKey)
	appModel := updated.(AppModel)

	updated2, _ := appModel.Update(spaceKey)
	appModel2 := updated2.(AppModel)
	defer appModel2.StopAllPTYSessions()

	if len(appModel2.sessionPanel.PinnedSessions()) != 0 {
		t.Errorf("expected 0 pinned after unpin, got %d", len(appModel2.sessionPanel.PinnedSessions()))
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui -run "TestAppSpacePinsSession|TestAppSpaceUnpinsSession" -v`
Expected: FAIL — Space key does nothing

**Step 3: Implement toggle pin**

In `internal/tui/app.go` `handleRune`, add the space case:

```go
func (app AppModel) handleRune(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "t":
		app.toggleCanvasMode()
	case "n":
		return app, app.createThread()
	case "?":
		app.helpVisible = !app.helpVisible
	case " ":
		return app.togglePin()
	}
	return app, nil
}
```

In `internal/tui/app_pty.go`, add `togglePin`:

```go
func (app AppModel) togglePin() (tea.Model, tea.Cmd) {
	threadID := app.canvas.SelectedThreadID()
	if threadID == "" {
		return app, nil
	}

	if app.sessionPanel.IsPinned(threadID) {
		app.sessionPanel.Unpin(threadID)
		return app, app.rebalancePTYSizes()
	}

	return app.pinSession(threadID)
}

func (app AppModel) pinSession(threadID string) (tea.Model, tea.Cmd) {
	_, exists := app.store.Get(threadID)
	if !exists {
		return app, nil
	}

	_, hasPTY := app.ptySessions[threadID]
	if !hasPTY {
		ptySession := NewPTYSession(PTYSessionConfig{
			ThreadID: threadID,
			Command:  app.resolveInteractiveCmd(),
			Args:     app.interactiveArgs,
			SendMsg:  app.ptyEventCallback(),
		})
		if err := ptySession.Start(); err != nil {
			app.statusBar.SetError(err.Error())
			return app, nil
		}
		app.ptySessions[threadID] = ptySession
	}

	app.sessionPanel.Pin(threadID)
	return app, app.rebalancePTYSizes()
}
```

Refactor `openSession` to reuse `pinSession`:

```go
func (app AppModel) openSession() (tea.Model, tea.Cmd) {
	threadID := app.canvas.SelectedThreadID()
	if threadID == "" {
		return app, nil
	}

	if !app.sessionPanel.IsPinned(threadID) {
		pinned, cmd := app.pinSession(threadID)
		app = pinned.(AppModel)
		if cmd != nil {
			_ = cmd
		}
	}

	app.focusPane = FocusPaneSession
	app.sessionPanel.SetActivePaneIdx(app.pinnedIndex(threadID))
	return app, app.rebalancePTYSizes()
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui -run "TestAppSpacePinsSession|TestAppSpaceUnpinsSession" -v`
Expected: PASS

**Step 5: Run full test suite**

Run: `go test ./... -v -race`
Expected: All tests pass

**Step 6: Commit**

```bash
git add internal/tui/app.go internal/tui/app_pty.go internal/tui/app_test.go
git commit -m "feat(tui): space key toggles pin/unpin session from canvas"
```

---

## Task 5: Add Tab key to switch between canvas and session panel

Tab moves focus down to the session panel (if sessions are pinned). Esc from session panel returns focus to canvas.

**Files:**
- Modify: `internal/tui/app.go`
- Modify: `internal/tui/app_test.go`

**Step 1: Write the failing test**

Add to `internal/tui/app_test.go`:

```go
func TestAppTabSwitchesToSessionPanel(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Thread 1")
	app := NewAppModel(store, WithInteractiveCommand("echo", "hello"))

	spaceKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := app.Update(spaceKey)
	app = updated.(AppModel)

	tabKey := tea.KeyMsg{Type: tea.KeyTab}
	updated, _ = app.Update(tabKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	if app.FocusPane() != FocusPaneSession {
		t.Errorf("expected FocusPaneSession, got %d", app.FocusPane())
	}
}

func TestAppTabDoesNothingWithNoPinnedSessions(t *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	tabKey := tea.KeyMsg{Type: tea.KeyTab}
	updated, _ := app.Update(tabKey)
	app = updated.(AppModel)

	if app.FocusPane() != FocusPaneCanvas {
		t.Errorf("expected FocusPaneCanvas, got %d", app.FocusPane())
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui -run "TestAppTabSwitches|TestAppTabDoesNothing" -v`
Expected: FAIL — Tab currently moves canvas selection

**Step 3: Implement Tab focus switch**

In `internal/tui/app.go`, update `handleKey` to intercept Tab before arrow handling:

```go
func (app AppModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if app.helpVisible {
		return app.handleHelpKey(msg)
	}

	if app.menuVisible {
		return app.handleMenuKey(msg)
	}

	prefixResult := app.prefix.HandleKey(msg)
	switch prefixResult {
	case PrefixWaiting:
		return app, nil
	case PrefixComplete:
		return app.handlePrefixAction()
	case PrefixCancelled:
		return app, nil
	}

	if app.focusPane == FocusPaneSession {
		return app.handleSessionKey(msg)
	}

	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		return app, tea.Quit
	case tea.KeyEnter:
		return app.openSession()
	case tea.KeyTab:
		return app.switchToSessionPanel()
	case tea.KeyRunes:
		return app.handleRune(msg)
	default:
		return app.handleArrow(msg)
	}
}

func (app AppModel) switchToSessionPanel() (tea.Model, tea.Cmd) {
	hasPinned := len(app.sessionPanel.PinnedSessions()) > 0
	if !hasPinned {
		return app, nil
	}
	app.focusPane = FocusPaneSession
	return app, nil
}
```

Also remove `tea.KeyTab` from `handleCanvasArrow` (it was previously mapped to `MoveRight`):

```go
func (app *AppModel) handleCanvasArrow(msg tea.KeyMsg) {
	switch msg.Type {
	case tea.KeyRight:
		app.canvas.MoveRight()
	case tea.KeyLeft, tea.KeyShiftTab:
		app.canvas.MoveLeft()
	case tea.KeyDown:
		app.canvas.MoveDown()
	case tea.KeyUp:
		app.canvas.MoveUp()
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui -run "TestAppTabSwitches|TestAppTabDoesNothing" -v`
Expected: PASS

**Step 5: Run full test suite**

Run: `go test ./... -v -race`
Expected: All tests pass

**Step 6: Commit**

```bash
git add internal/tui/app.go internal/tui/app_test.go
git commit -m "feat(tui): Tab key switches focus between canvas and session panel"
```

---

## Task 6: Add Ctrl+B prefix actions for session panel

Add panel navigation bindings: `Ctrl+B ←/→` cycles panes, `Ctrl+B x` unpins focused session, `Ctrl+B z` toggles zoom (full-height single session).

**Files:**
- Modify: `internal/tui/app_menu.go` (handlePrefixAction)
- Modify: `internal/tui/session_panel.go` (zoom state)
- Modify: `internal/tui/session_panel_test.go`
- Modify: `internal/tui/app_test.go`

**Step 1: Write the failing test for zoom toggle**

Add to `internal/tui/session_panel_test.go`:

```go
func TestSessionPanelZoomToggle(t *testing.T) {
	panel := NewSessionPanelModel()
	panel.Pin("t-1")
	panel.Pin("t-2")

	if panel.Zoomed() {
		t.Error("expected not zoomed initially")
	}

	panel.ToggleZoom()
	if !panel.Zoomed() {
		t.Error("expected zoomed after toggle")
	}

	panel.ToggleZoom()
	if panel.Zoomed() {
		t.Error("expected not zoomed after second toggle")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui -run TestSessionPanelZoomToggle -v`
Expected: FAIL — `Zoomed()` and `ToggleZoom()` undefined

**Step 3: Add zoom to SessionPanelModel**

In `internal/tui/session_panel.go`, add `zoomed bool` field and methods:

```go
type SessionPanelModel struct {
	pinnedSessions []string
	activePaneIdx  int
	splitRatio     float64
	zoomed         bool
}

func (panel *SessionPanelModel) Zoomed() bool {
	return panel.zoomed
}

func (panel *SessionPanelModel) ToggleZoom() {
	panel.zoomed = !panel.zoomed
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui -run TestSessionPanelZoomToggle -v`
Expected: PASS

**Step 5: Write failing tests for prefix actions**

Add to `internal/tui/app_test.go`:

```go
func TestAppCtrlBXUnpinsSession(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Thread 1")
	app := NewAppModel(store, WithInteractiveCommand("echo", "hello"))

	spaceKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := app.Update(spaceKey)
	app = updated.(AppModel)

	tabKey := tea.KeyMsg{Type: tea.KeyTab}
	updated, _ = app.Update(tabKey)
	app = updated.(AppModel)

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	updated, _ = app.Update(ctrlB)
	app = updated.(AppModel)

	xKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	updated, _ = app.Update(xKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	if len(app.sessionPanel.PinnedSessions()) != 0 {
		t.Errorf("expected 0 pinned after unpin, got %d", len(app.sessionPanel.PinnedSessions()))
	}
	if app.FocusPane() != FocusPaneCanvas {
		t.Errorf("expected focus back to canvas, got %d", app.FocusPane())
	}
}

func TestAppCtrlBZTogglesZoom(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Thread 1")
	app := NewAppModel(store, WithInteractiveCommand("echo", "hello"))

	spaceKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := app.Update(spaceKey)
	app = updated.(AppModel)

	tabKey := tea.KeyMsg{Type: tea.KeyTab}
	updated, _ = app.Update(tabKey)
	app = updated.(AppModel)

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	updated, _ = app.Update(ctrlB)
	app = updated.(AppModel)

	zKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}}
	updated, _ = app.Update(zKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	if !app.sessionPanel.Zoomed() {
		t.Error("expected zoomed after Ctrl+B z")
	}
}
```

**Step 6: Run test to verify they fail**

Run: `go test ./internal/tui -run "TestAppCtrlBXUnpins|TestAppCtrlBZToggles" -v`
Expected: FAIL — prefix actions 'x' and 'z' not handled

**Step 7: Implement prefix actions**

In `internal/tui/app_menu.go`, expand `handlePrefixAction`:

```go
func (app AppModel) handlePrefixAction() (tea.Model, tea.Cmd) {
	action := app.prefix.Action()
	switch action {
	case 'm':
		app.showMenu()
	case 'x':
		return app.unpinActiveSession()
	case 'z':
		return app.toggleZoom()
	}
	return app, nil
}

func (app AppModel) unpinActiveSession() (tea.Model, tea.Cmd) {
	activeID := app.sessionPanel.ActiveThreadID()
	if activeID == "" {
		return app, nil
	}
	app.sessionPanel.Unpin(activeID)
	hasPinned := len(app.sessionPanel.PinnedSessions()) > 0
	if !hasPinned {
		app.focusPane = FocusPaneCanvas
	}
	return app, app.rebalancePTYSizes()
}

func (app AppModel) toggleZoom() (tea.Model, tea.Cmd) {
	app.sessionPanel.ToggleZoom()
	return app, app.rebalancePTYSizes()
}
```

**Step 8: Run tests to verify they pass**

Run: `go test ./internal/tui -run "TestAppCtrlBXUnpins|TestAppCtrlBZToggles" -v`
Expected: PASS

**Step 9: Run full test suite**

Run: `go test ./... -v -race`
Expected: All tests pass

**Step 10: Commit**

```bash
git add internal/tui/app_menu.go internal/tui/session_panel.go internal/tui/session_panel_test.go internal/tui/app_test.go
git commit -m "feat(tui): Ctrl+B x/z prefix actions for unpin and zoom"
```

---

## Task 7: Update session panel View() for zoom mode

When zoomed, only the active session renders at full panel width. When not zoomed, all pinned sessions render side-by-side.

**Files:**
- Modify: `internal/tui/app_view.go`
- Modify: `internal/tui/app_test.go`

**Step 1: Write the failing test**

Add to `internal/tui/app_test.go`:

```go
func TestAppViewShowsDividerWhenPinned(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Thread 1")
	app := NewAppModel(store, WithInteractiveCommand("echo", "hello"))
	app.width = 120
	app.height = 40

	spaceKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := app.Update(spaceKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	view := app.View()
	if !strings.Contains(view, "─") {
		t.Error("expected divider line in view when sessions pinned")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui -run TestAppViewShowsDividerWhenPinned -v`
Expected: FAIL — no divider character present

**Step 3: Update renderSessionPanel for zoom**

In `internal/tui/app_view.go`, update `renderSessionPanel`:

```go
func (app AppModel) renderSessionPanel() string {
	pinned := app.sessionPanel.PinnedSessions()
	count := len(pinned)
	if count == 0 {
		return ""
	}

	canvasHeight := int(float64(app.height) * app.sessionPanel.SplitRatio())
	panelHeight := app.height - canvasHeight - dividerHeight

	if app.sessionPanel.Zoomed() {
		return app.renderZoomedSession(panelHeight)
	}
	return app.renderSideBySideSessions(pinned, panelHeight)
}

func (app AppModel) renderZoomedSession(panelHeight int) string {
	activeID := app.sessionPanel.ActiveThreadID()
	if activeID == "" {
		return ""
	}

	content := ""
	ptySession, exists := app.ptySessions[activeID]
	if exists {
		content = ptySession.Render()
	}

	style := app.sessionPaneStyle(app.width, panelHeight, true)
	return style.Render(content)
}

func (app AppModel) renderSideBySideSessions(pinned []string, panelHeight int) string {
	count := len(pinned)
	sessionWidth := app.width / count

	panes := make([]string, count)
	for index, threadID := range pinned {
		content := ""
		ptySession, exists := app.ptySessions[threadID]
		if exists {
			content = ptySession.Render()
		}

		isActive := index == app.sessionPanel.ActivePaneIdx() && app.focusPane == FocusPaneSession
		style := app.sessionPaneStyle(sessionWidth, panelHeight, isActive)
		panes[index] = style.Render(content)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, panes...)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui -run TestAppViewShowsDividerWhenPinned -v`
Expected: PASS

**Step 5: Run full test suite + linter**

Run: `go test ./... -v -race && golangci-lint run`
Expected: All pass, no lint issues

**Step 6: Commit**

```bash
git add internal/tui/app_view.go internal/tui/app_test.go
git commit -m "feat(tui): zoom mode renders single session at full panel width"
```

---

## Task 8: Update help overlay with new keybindings

Add the new pinning and panel keybindings to the help screen.

**Files:**
- Modify: `internal/tui/help.go`
- Modify: `internal/tui/app_test.go` (optional: verify help content)

**Step 1: Write the failing test**

Add to `internal/tui/app_test.go`:

```go
func TestHelpShowsPinKeybinding(t *testing.T) {
	help := NewHelpModel()
	view := help.View()
	if !strings.Contains(view, "Space") {
		t.Error("expected Space keybinding in help")
	}
	if !strings.Contains(view, "Ctrl+B x") {
		t.Error("expected Ctrl+B x keybinding in help")
	}
	if !strings.Contains(view, "Ctrl+B z") {
		t.Error("expected Ctrl+B z keybinding in help")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui -run TestHelpShowsPinKeybinding -v`
Expected: FAIL — Space/Ctrl+B x/Ctrl+B z not in help

**Step 3: Update keybindings list**

In `internal/tui/help.go`, update the keybindings slice:

```go
var keybindings = []keybinding{
	{"←/→", "Navigate cards horizontally"},
	{"↑/↓", "Navigate cards vertically"},
	{"Enter", "Open + focus session"},
	{"Space", "Toggle pin/unpin session"},
	{"Tab", "Switch to session panel"},
	{"Esc", "Back / close overlay"},
	{"t", "Toggle tree view"},
	{"n", "New thread"},
	{"Ctrl+B", "Prefix key (tmux-style)"},
	{"Ctrl+B m", "Open context menu"},
	{"Ctrl+B x", "Unpin focused session"},
	{"Ctrl+B z", "Toggle zoom session"},
	{"?", "Toggle help"},
	{"Ctrl+C", "Quit"},
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui -run TestHelpShowsPinKeybinding -v`
Expected: PASS

**Step 5: Run full test suite**

Run: `go test ./... -v -race`
Expected: All tests pass

**Step 6: Commit**

```bash
git add internal/tui/help.go internal/tui/app_test.go
git commit -m "docs(tui): update help overlay with pin/unpin/zoom keybindings"
```

---

## Task 9: Remove orphaned SessionModel

The old `SessionModel` in `session.go` is no longer used — the session panel renders directly from `PTYSession`. Remove it and its tests.

**Files:**
- Delete: `internal/tui/session.go`
- Delete: `internal/tui/session_test.go`
- Modify: `internal/tui/app_view.go` (remove any remaining references)

**Step 1: Search for SessionModel references**

Run: `grep -r "SessionModel" internal/tui/`
Verify that no remaining code references `SessionModel`, `NewSessionModel`, or `session.View()`.

**Step 2: Delete the files**

```bash
rm internal/tui/session.go internal/tui/session_test.go
```

**Step 3: Run full test suite**

Run: `go test ./... -v -race`
Expected: All tests pass (no references to deleted code)

**Step 4: Run linter**

Run: `golangci-lint run`
Expected: No issues

**Step 5: Commit**

```bash
git add -u internal/tui/session.go internal/tui/session_test.go
git commit -m "refactor(tui): remove orphaned SessionModel replaced by session panel"
```

---

## Task 10: Divider bar with tab-style session labels

Replace the plain `───` divider with labeled tabs showing which sessions are pinned and which is active.

**Files:**
- Create: `internal/tui/divider.go`
- Create: `internal/tui/divider_test.go`
- Modify: `internal/tui/app_view.go` (use new divider)

**Step 1: Write the failing test**

Create `internal/tui/divider_test.go`:

```go
package tui

import (
	"strings"
	"testing"
)

func TestDividerRenderShowsLabels(t *testing.T) {
	sessions := []string{"agent-a", "agent-b"}
	result := renderDividerBar(sessions, 1, 80)

	if !strings.Contains(result, "agent-a") {
		t.Error("expected agent-a in divider")
	}
	if !strings.Contains(result, "agent-b") {
		t.Error("expected agent-b in divider")
	}
}

func TestDividerRenderHighlightsActive(t *testing.T) {
	sessions := []string{"agent-a", "agent-b"}
	result := renderDividerBar(sessions, 0, 80)

	if !strings.Contains(result, "agent-a") {
		t.Error("expected agent-a label present")
	}
}

func TestDividerRenderEmpty(t *testing.T) {
	result := renderDividerBar(nil, 0, 80)
	if result != "" {
		t.Errorf("expected empty string for no sessions, got %q", result)
	}
}

func TestDividerRenderNumbersLabels(t *testing.T) {
	sessions := []string{"a", "b", "c"}
	result := renderDividerBar(sessions, 0, 120)

	if !strings.Contains(result, "1:") {
		t.Error("expected numbered label starting at 1")
	}
	if !strings.Contains(result, "3:") {
		t.Error("expected label 3 for third session")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui -run TestDividerRender -v`
Expected: FAIL — `renderDividerBar` undefined

**Step 3: Implement divider**

Create `internal/tui/divider.go`:

```go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	dividerLineStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
	dividerActiveTabStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)
	dividerInactiveTabStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))
)

func renderDividerBar(sessions []string, activeIdx int, width int) string {
	if len(sessions) == 0 {
		return ""
	}

	var tabs []string
	for index, name := range sessions {
		label := fmt.Sprintf(" %d: %s ", index+1, truncateLabel(name, 20))
		if index == activeIdx {
			tabs = append(tabs, dividerActiveTabStyle.Render(label))
		} else {
			tabs = append(tabs, dividerInactiveTabStyle.Render(label))
		}
	}

	tabBar := strings.Join(tabs, dividerLineStyle.Render("│"))
	remaining := width - lipgloss.Width(tabBar)
	if remaining > 0 {
		tabBar += dividerLineStyle.Render(strings.Repeat("─", remaining))
	}
	return tabBar
}

func truncateLabel(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui -run TestDividerRender -v`
Expected: PASS

**Step 5: Wire into app_view.go**

In `internal/tui/app_view.go`, replace `renderDivider()`:

```go
func (app AppModel) renderDivider() string {
	pinned := app.sessionPanel.PinnedSessions()
	activeIdx := app.sessionPanel.ActivePaneIdx()

	labels := make([]string, len(pinned))
	for index, threadID := range pinned {
		thread, exists := app.store.Get(threadID)
		if exists {
			labels[index] = thread.Title
		} else {
			labels[index] = threadID
		}
	}
	return renderDividerBar(labels, activeIdx, app.width)
}
```

**Step 6: Run full test suite + linter**

Run: `go test ./... -v -race && golangci-lint run`
Expected: All pass

**Step 7: Commit**

```bash
git add internal/tui/divider.go internal/tui/divider_test.go internal/tui/app_view.go
git commit -m "feat(tui): labeled divider bar with numbered session tabs"
```

---

## Task 11: Ctrl+B arrow keys cycle panes in session panel

When focused on the session panel, `Ctrl+B ←` and `Ctrl+B →` cycle between pinned sessions. `Ctrl+B 1-9` jumps directly to that session index.

**Files:**
- Modify: `internal/tui/prefix.go` (support arrow keys as prefix actions)
- Modify: `internal/tui/app_menu.go` (handle new prefix actions)
- Modify: `internal/tui/app_test.go`

**Step 1: Write the failing test**

Add to `internal/tui/app_test.go`:

```go
func TestAppCtrlBRightCyclesPaneRight(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Thread 1")
	store.Add("t-2", "Thread 2")
	app := NewAppModel(store, WithInteractiveCommand("echo", "hello"))

	space := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := app.Update(space)
	app = updated.(AppModel)

	app.canvas.MoveRight()
	updated, _ = app.Update(space)
	app = updated.(AppModel)

	tab := tea.KeyMsg{Type: tea.KeyTab}
	updated, _ = app.Update(tab)
	app = updated.(AppModel)

	if app.sessionPanel.ActivePaneIdx() != 0 {
		t.Fatalf("expected active pane 0, got %d", app.sessionPanel.ActivePaneIdx())
	}

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	updated, _ = app.Update(ctrlB)
	app = updated.(AppModel)

	rightKey := tea.KeyMsg{Type: tea.KeyRight}
	updated, _ = app.Update(rightKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	if app.sessionPanel.ActivePaneIdx() != 1 {
		t.Errorf("expected active pane 1, got %d", app.sessionPanel.ActivePaneIdx())
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui -run TestAppCtrlBRightCyclesPaneRight -v`
Expected: FAIL — arrow key not handled as prefix action

**Step 3: Update prefix handler to support arrow keys**

In `internal/tui/prefix.go`, add support for non-rune prefix completions. Add a `keyType` field alongside `action`:

```go
type PrefixHandler struct {
	active  bool
	action  rune
	keyType tea.KeyType
}

func (handler *PrefixHandler) KeyType() tea.KeyType {
	return handler.keyType
}

func (handler *PrefixHandler) HandleKey(msg tea.KeyMsg) int {
	if !handler.active {
		if msg.Type == tea.KeyCtrlB {
			handler.active = true
			return PrefixWaiting
		}
		return PrefixNone
	}

	handler.active = false

	if msg.Type == tea.KeyEsc {
		return PrefixCancelled
	}

	hasRunes := msg.Type == tea.KeyRunes && len(msg.Runes) > 0
	if hasRunes {
		handler.action = msg.Runes[0]
		handler.keyType = msg.Type
		return PrefixComplete
	}

	isArrow := msg.Type == tea.KeyLeft || msg.Type == tea.KeyRight || msg.Type == tea.KeyUp || msg.Type == tea.KeyDown
	if isArrow {
		handler.action = 0
		handler.keyType = msg.Type
		return PrefixComplete
	}

	return PrefixCancelled
}
```

**Step 4: Update handlePrefixAction for arrows and number keys**

In `internal/tui/app_menu.go`:

```go
func (app AppModel) handlePrefixAction() (tea.Model, tea.Cmd) {
	action := app.prefix.Action()
	keyType := app.prefix.KeyType()

	switch {
	case action == 'm':
		app.showMenu()
	case action == 'x':
		return app.unpinActiveSession()
	case action == 'z':
		return app.toggleZoom()
	case keyType == tea.KeyRight:
		app.sessionPanel.CycleRight()
	case keyType == tea.KeyLeft:
		app.sessionPanel.CycleLeft()
	case action >= '1' && action <= '9':
		return app.jumpToPane(action)
	}
	return app, nil
}

func (app AppModel) jumpToPane(digit rune) (tea.Model, tea.Cmd) {
	index := int(digit - '1')
	pinned := app.sessionPanel.PinnedSessions()
	if index >= len(pinned) {
		return app, nil
	}
	app.sessionPanel.SetActivePaneIdx(index)
	app.focusPane = FocusPaneSession
	return app, nil
}
```

**Step 5: Run test to verify it passes**

Run: `go test ./internal/tui -run TestAppCtrlBRightCyclesPaneRight -v`
Expected: PASS

**Step 6: Run full test suite**

Run: `go test ./... -v -race`
Expected: All tests pass

**Step 7: Commit**

```bash
git add internal/tui/prefix.go internal/tui/app_menu.go internal/tui/app_test.go
git commit -m "feat(tui): Ctrl+B arrow keys and 1-9 for session panel navigation"
```

---

## Task 12: Rebalance PTY sizes on all layout changes

Ensure PTY resize is called whenever: sessions are pinned/unpinned, terminal is resized, or zoom is toggled. This prevents codex CLI output corruption.

**Files:**
- Modify: `internal/tui/app_pty.go` (rebalancePTYSizes handles zoom)
- Modify: `internal/tui/pty_session_test.go`

**Step 1: Write the failing test**

Add to `internal/tui/pty_session_test.go`:

```go
func TestPTYSessionResizeUpdatesEmulatorDimensions(t *testing.T) {
	session := NewPTYSession(PTYSessionConfig{
		ThreadID: "t-1",
		Command:  "true",
		Args:     nil,
		SendMsg:  func(msg PTYOutputMsg) {},
	})

	session.Resize(100, 30)
	if session.emulator.Width() != 100 {
		t.Errorf("expected width 100, got %d", session.emulator.Width())
	}
	if session.emulator.Height() != 30 {
		t.Errorf("expected height 30, got %d", session.emulator.Height())
	}
}
```

**Step 2: Run test to verify it passes (existing behavior)**

Run: `go test ./internal/tui -run TestPTYSessionResizeUpdatesEmulatorDimensions -v`
Expected: PASS (this confirms existing resize works)

**Step 3: Update rebalancePTYSizes to handle zoom**

In `internal/tui/app_pty.go`:

```go
func (app AppModel) rebalancePTYSizes() tea.Cmd {
	pinned := app.sessionPanel.PinnedSessions()
	if len(pinned) == 0 {
		return nil
	}

	canvasHeight := int(float64(app.height) * app.sessionPanel.SplitRatio())
	panelHeight := app.height - canvasHeight - dividerHeight

	if app.sessionPanel.Zoomed() {
		activeID := app.sessionPanel.ActiveThreadID()
		ptySession, exists := app.ptySessions[activeID]
		if exists {
			ptySession.Resize(app.width, panelHeight)
		}
		return nil
	}

	count := len(pinned)
	sessionWidth := app.width / count
	for _, threadID := range pinned {
		ptySession, exists := app.ptySessions[threadID]
		if exists {
			ptySession.Resize(sessionWidth, panelHeight)
		}
	}
	return nil
}
```

**Step 4: Run full test suite**

Run: `go test ./... -v -race`
Expected: All tests pass

**Step 5: Commit**

```bash
git add internal/tui/app_pty.go internal/tui/pty_session_test.go
git commit -m "fix(tui): PTY resize handles zoom mode and multi-session rebalancing"
```

---

## Task 13: Final cleanup and file length check

Ensure all files are under 300 lines, functions under 60 lines, and linter passes.

**Files:**
- All modified files

**Step 1: Check file lengths**

Run: `wc -l internal/tui/*.go | sort -rn | head -20`
Expected: All non-test files under 300 lines

**Step 2: Run full test suite with race detector**

Run: `go test ./... -v -race`
Expected: All tests pass

**Step 3: Run linter**

Run: `golangci-lint run`
Expected: No issues

**Step 4: If any file exceeds 300 lines**

Extract functions into new files. For example, if `app.go` is too long, move `handleArrow`, `handleCanvasArrow`, `handleTreeArrow` into `app_nav.go`.

If `app_pty.go` is too long, consider moving `rebalancePTYSizes` and `pinnedIndex` into `app_panel.go`.

**Step 5: Final commit**

```bash
git add -A
git commit -m "chore(tui): file length cleanup, all files under 300 lines"
```

---

## Summary

| Task | Description | New Files | Modified Files |
|------|-------------|-----------|---------------|
| 1 | Pin/unpin/focus message types | — | msgs.go, msgs_test.go |
| 2 | SessionPanelModel sub-model | session_panel.go, session_panel_test.go | — |
| 3 | Integrate panel into AppModel | — | app.go, app_view.go, app_pty.go, app_test.go |
| 4 | Space key toggles pin | — | app.go, app_pty.go, app_test.go |
| 5 | Tab key switches focus | — | app.go, app_test.go |
| 6 | Ctrl+B x/z prefix actions | — | app_menu.go, session_panel.go, session_panel_test.go, app_test.go |
| 7 | Zoom mode rendering | — | app_view.go, app_test.go |
| 8 | Updated help overlay | — | help.go, app_test.go |
| 9 | Remove orphaned SessionModel | — | session.go (delete), session_test.go (delete) |
| 10 | Labeled divider bar | divider.go, divider_test.go | app_view.go |
| 11 | Ctrl+B arrows/1-9 navigation | — | prefix.go, app_menu.go, app_test.go |
| 12 | PTY resize on layout changes | — | app_pty.go, pty_session_test.go |
| 13 | Final cleanup + lint | — | any files over 300 lines |
