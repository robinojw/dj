# Phase 7: Context Menu, Fork & Delete

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build an overlay context menu for thread operations (fork, delete, rename). Implement the `Ctrl+B` prefix key system for pane operations (tmux-style). Wire fork/delete operations through the app-server client.

**Architecture:** `ContextMenuModel` is an overlay component rendered on top of the canvas or session pane. It appears when pressing `Ctrl+B` followed by `m` (menu). Menu items trigger client RPC calls and update the store. Fork creates a child thread via the API. Delete removes a thread. The prefix key system buffers `Ctrl+B` and waits for the next keystroke.

**Tech Stack:** Go, Bubble Tea, Lipgloss

**Prerequisites:** Phase 6 (session pane, full focus system)

---

### Task 1: Build Prefix Key Handler

**Files:**
- Create: `internal/tui/prefix.go`
- Create: `internal/tui/prefix_test.go`

**Step 1: Write tests for prefix key detection**

```go
// internal/tui/prefix_test.go
package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPrefixKeyCapture(t *testing.T) {
	prefix := NewPrefixHandler()

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	result := prefix.HandleKey(ctrlB)

	if result != PrefixWaiting {
		t.Errorf("expected waiting, got %d", result)
	}
	if !prefix.Active() {
		t.Error("expected prefix to be active")
	}
}

func TestPrefixKeyFollowUp(t *testing.T) {
	prefix := NewPrefixHandler()

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	prefix.HandleKey(ctrlB)

	mKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	result := prefix.HandleKey(mKey)

	if result != PrefixComplete {
		t.Errorf("expected complete, got %d", result)
	}
	if prefix.Action() != 'm' {
		t.Errorf("expected 'm', got %c", prefix.Action())
	}
	if prefix.Active() {
		t.Error("expected prefix to be inactive after completion")
	}
}

func TestPrefixKeyTimeout(t *testing.T) {
	prefix := NewPrefixHandler()

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	prefix.HandleKey(ctrlB)

	escKey := tea.KeyMsg{Type: tea.KeyEsc}
	result := prefix.HandleKey(escKey)

	if result != PrefixCancelled {
		t.Errorf("expected cancelled, got %d", result)
	}
	if prefix.Active() {
		t.Error("expected prefix to be inactive after cancel")
	}
}

func TestPrefixKeyInactivePassthrough(t *testing.T) {
	prefix := NewPrefixHandler()

	normalKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	result := prefix.HandleKey(normalKey)

	if result != PrefixNone {
		t.Errorf("expected none, got %d", result)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -v -run TestPrefix`
Expected: FAIL — `NewPrefixHandler` not defined

**Step 3: Implement prefix key handler**

```go
// internal/tui/prefix.go
package tui

import tea "github.com/charmbracelet/bubbletea"

// Prefix key result states.
const (
	PrefixNone      = iota
	PrefixWaiting
	PrefixComplete
	PrefixCancelled
)

// PrefixHandler implements tmux-style Ctrl+B prefix key detection.
type PrefixHandler struct {
	active bool
	action rune
}

// NewPrefixHandler creates an inactive prefix handler.
func NewPrefixHandler() *PrefixHandler {
	return &PrefixHandler{}
}

// Active returns whether the prefix is waiting for a follow-up key.
func (p *PrefixHandler) Active() bool {
	return p.active
}

// Action returns the follow-up key that completed the prefix.
func (p *PrefixHandler) Action() rune {
	return p.action
}

// HandleKey processes a key event through the prefix system.
func (p *PrefixHandler) HandleKey(msg tea.KeyMsg) int {
	if !p.active {
		if msg.Type == tea.KeyCtrlB {
			p.active = true
			return PrefixWaiting
		}
		return PrefixNone
	}

	p.active = false

	if msg.Type == tea.KeyEsc {
		return PrefixCancelled
	}

	if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
		p.action = msg.Runes[0]
		return PrefixComplete
	}

	return PrefixCancelled
}
```

**Step 4: Run tests**

Run: `go test ./internal/tui/ -v -run TestPrefix`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/prefix.go internal/tui/prefix_test.go
git commit -m "feat(tui): tmux-style Ctrl+B prefix key handler"
```

---

### Task 2: Build Context Menu Component

**Files:**
- Create: `internal/tui/menu.go`
- Create: `internal/tui/menu_test.go`

**Step 1: Write tests for context menu**

```go
// internal/tui/menu_test.go
package tui

import (
	"strings"
	"testing"
)

func TestMenuRender(t *testing.T) {
	items := []MenuItem{
		{Label: "Fork Thread", Key: 'f'},
		{Label: "Delete Thread", Key: 'd'},
		{Label: "Rename Thread", Key: 'r'},
	}
	menu := NewMenuModel("Thread Actions", items)

	output := menu.View()
	if !strings.Contains(output, "Fork Thread") {
		t.Errorf("expected Fork Thread in output:\n%s", output)
	}
	if !strings.Contains(output, "Delete Thread") {
		t.Errorf("expected Delete Thread in output:\n%s", output)
	}
}

func TestMenuNavigation(t *testing.T) {
	items := []MenuItem{
		{Label: "First", Key: 'a'},
		{Label: "Second", Key: 'b'},
	}
	menu := NewMenuModel("Test", items)

	if menu.SelectedIndex() != 0 {
		t.Errorf("expected 0, got %d", menu.SelectedIndex())
	}

	menu.MoveDown()
	if menu.SelectedIndex() != 1 {
		t.Errorf("expected 1, got %d", menu.SelectedIndex())
	}

	menu.MoveDown()
	if menu.SelectedIndex() != 1 {
		t.Errorf("expected clamped at 1, got %d", menu.SelectedIndex())
	}
}

func TestMenuSelect(t *testing.T) {
	items := []MenuItem{
		{Label: "Fork", Key: 'f'},
		{Label: "Delete", Key: 'd'},
	}
	menu := NewMenuModel("Test", items)

	selected := menu.Selected()
	if selected.Key != 'f' {
		t.Errorf("expected f, got %c", selected.Key)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -v -run TestMenu`
Expected: FAIL — `NewMenuModel` not defined

**Step 3: Implement context menu**

```go
// internal/tui/menu.go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	menuBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39")).
			Padding(1, 2)
	menuTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			MarginBottom(1)
	menuItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))
	menuSelectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")).
				Bold(true)
)

// MenuItem is a single entry in the context menu.
type MenuItem struct {
	Label string
	Key   rune
}

// MenuModel is an overlay context menu.
type MenuModel struct {
	title    string
	items    []MenuItem
	selected int
}

// NewMenuModel creates a context menu.
func NewMenuModel(title string, items []MenuItem) MenuModel {
	return MenuModel{
		title: title,
		items: items,
	}
}

// SelectedIndex returns the current selection.
func (m *MenuModel) SelectedIndex() int {
	return m.selected
}

// Selected returns the currently highlighted menu item.
func (m *MenuModel) Selected() MenuItem {
	return m.items[m.selected]
}

// MoveDown moves selection down.
func (m *MenuModel) MoveDown() {
	if m.selected < len(m.items)-1 {
		m.selected++
	}
}

// MoveUp moves selection up.
func (m *MenuModel) MoveUp() {
	if m.selected > 0 {
		m.selected--
	}
}

// View renders the menu overlay.
func (m MenuModel) View() string {
	title := menuTitleStyle.Render(m.title)

	var lines []string
	for i, item := range m.items {
		style := menuItemStyle
		prefix := "  "
		if i == m.selected {
			style = menuSelectedItemStyle
			prefix = "▸ "
		}
		line := style.Render(fmt.Sprintf("%s[%c] %s", prefix, item.Key, item.Label))
		lines = append(lines, line)
	}

	content := title + "\n" + strings.Join(lines, "\n")
	return menuBorderStyle.Render(content)
}
```

**Step 4: Run tests**

Run: `go test ./internal/tui/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/menu.go internal/tui/menu_test.go
git commit -m "feat(tui): context menu overlay component"
```

---

### Task 3: Wire Prefix + Menu into App

**Files:**
- Modify: `internal/tui/app.go`
- Modify: `internal/tui/app_test.go`

**Step 1: Write tests for Ctrl+B → m opening context menu**

Add to `app_test.go`:

```go
func TestAppCtrlBMOpensMenu(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Test")

	app := NewAppModel(store)

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	updated, _ := app.Update(ctrlB)
	app = updated.(AppModel)

	mKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	updated, _ = app.Update(mKey)
	app = updated.(AppModel)

	if !app.MenuVisible() {
		t.Error("expected menu to be visible")
	}
}
```

**Step 2: Integrate prefix handler and menu into app**

Add `prefix *PrefixHandler`, `menu *MenuModel`, and `menuVisible bool` to AppModel. In `handleKey`, check prefix first. On `PrefixComplete` with action `'m'`, show the thread context menu. On Enter in menu, execute the selected action. On Esc in menu, close it.

**Step 3: Run tests**

Run: `go test ./internal/tui/ -v`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/tui/app.go internal/tui/app_test.go
git commit -m "feat(tui): wire Ctrl+B prefix and context menu into app"
```

---

### Task 4: Implement Fork and Delete Actions

**Files:**
- Create: `internal/tui/actions.go`
- Create: `internal/tui/actions_test.go`

**Step 1: Write tests for action commands**

```go
// internal/tui/actions_test.go
package tui

import "testing"

func TestForkActionMsg(t *testing.T) {
	msg := ForkThreadMsg{
		ParentID:     "t-1",
		Instructions: "Continue from here",
	}
	if msg.ParentID != "t-1" {
		t.Errorf("expected t-1, got %s", msg.ParentID)
	}
}

func TestDeleteActionMsg(t *testing.T) {
	msg := DeleteThreadMsg{ThreadID: "t-1"}
	if msg.ThreadID != "t-1" {
		t.Errorf("expected t-1, got %s", msg.ThreadID)
	}
}
```

**Step 2: Implement action message types**

```go
// internal/tui/actions.go
package tui

// ForkThreadMsg requests forking a thread.
type ForkThreadMsg struct {
	ParentID     string
	Instructions string
}

// DeleteThreadMsg requests deleting a thread.
type DeleteThreadMsg struct {
	ThreadID string
}

// RenameThreadMsg requests renaming a thread.
type RenameThreadMsg struct {
	ThreadID string
	NewTitle string
}
```

**Step 3: Handle action messages in App.Update**

In `app.go`, handle `ForkThreadMsg` by calling `client.CreateThread` with a `tea.Cmd` that returns the result. Handle `DeleteThreadMsg` by calling `client.DeleteThread` and removing from store. These are async commands that return result messages.

**Step 4: Run tests**

Run: `go test ./internal/tui/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/actions.go internal/tui/actions_test.go internal/tui/app.go
git commit -m "feat(tui): fork and delete thread actions via context menu"
```
