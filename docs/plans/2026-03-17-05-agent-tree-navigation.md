# Phase 5: Agent Tree Navigation

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a hierarchical tree view sidebar for navigating parent/child agent threads. The tree shows thread relationships and allows keyboard navigation. Selecting a tree node updates the canvas selection.

**Architecture:** `TreeModel` is a Bubble Tea component that renders a tree from `ThreadStore` data. Threads can have parent IDs, forming a hierarchy. The tree supports expand/collapse with Enter, and Up/Down navigation. A `parentID` field is added to `ThreadState`. The App model gains a focus system toggling between canvas and tree views.

**Tech Stack:** Go, Bubble Tea, Lipgloss

**Prerequisites:** Phase 4 (canvas UI, app model)

---

### Task 1: Add Parent Tracking to ThreadState

**Files:**
- Modify: `internal/state/thread.go`
- Modify: `internal/state/store.go`
- Create: `internal/state/tree_test.go`

**Step 1: Write tests for parent-child relationships**

```go
// internal/state/tree_test.go
package state

import "testing"

func TestThreadStateParentID(t *testing.T) {
	thread := NewThreadState("t-child", "Child Task")
	thread.ParentID = "t-parent"

	if thread.ParentID != "t-parent" {
		t.Errorf("expected t-parent, got %s", thread.ParentID)
	}
}

func TestStoreChildren(t *testing.T) {
	store := NewThreadStore()
	store.Add("t-root", "Root")
	store.AddWithParent("t-child-1", "Child 1", "t-root")
	store.AddWithParent("t-child-2", "Child 2", "t-root")

	children := store.Children("t-root")
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}
}

func TestStoreRoots(t *testing.T) {
	store := NewThreadStore()
	store.Add("t-root-1", "Root 1")
	store.Add("t-root-2", "Root 2")
	store.AddWithParent("t-child", "Child", "t-root-1")

	roots := store.Roots()
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots, got %d", len(roots))
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/state/ -v -run "TestThreadStateParent|TestStoreChildren|TestStoreRoots"`
Expected: FAIL — `ParentID` field and methods not defined

**Step 3: Add ParentID to ThreadState and new store methods**

Add `ParentID string` field to `ThreadState` in `thread.go`.

Add to `store.go`:

```go
// AddWithParent creates a new thread with a parent relationship.
func (s *ThreadStore) AddWithParent(id string, title string, parentID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	thread := NewThreadState(id, title)
	thread.ParentID = parentID
	s.threads[id] = thread
	s.order = append(s.order, id)
}

// Children returns all threads whose parent is the given ID.
func (s *ThreadStore) Children(parentID string) []*ThreadState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var children []*ThreadState
	for _, id := range s.order {
		thread := s.threads[id]
		if thread.ParentID == parentID {
			children = append(children, thread)
		}
	}
	return children
}

// Roots returns all threads with no parent.
func (s *ThreadStore) Roots() []*ThreadState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var roots []*ThreadState
	for _, id := range s.order {
		thread := s.threads[id]
		if thread.ParentID == "" {
			roots = append(roots, thread)
		}
	}
	return roots
}
```

**Step 4: Run tests**

Run: `go test ./internal/state/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/state/thread.go internal/state/store.go internal/state/tree_test.go
git commit -m "feat(state): add parent-child thread relationships"
```

---

### Task 2: Build Tree Component

**Files:**
- Create: `internal/tui/tree.go`
- Create: `internal/tui/tree_test.go`

**Step 1: Write tests for tree rendering and navigation**

```go
// internal/tui/tree_test.go
package tui

import (
	"strings"
	"testing"

	"github.com/robinojw/dj/internal/state"
)

func TestTreeRender(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Root Task")
	store.AddWithParent("t-2", "Subtask A", "t-1")

	tree := NewTreeModel(store)
	output := tree.View()

	if !strings.Contains(output, "Root Task") {
		t.Errorf("expected Root Task in output:\n%s", output)
	}
	if !strings.Contains(output, "Subtask A") {
		t.Errorf("expected Subtask A in output:\n%s", output)
	}
}

func TestTreeNavigation(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "First")
	store.Add("t-2", "Second")

	tree := NewTreeModel(store)

	if tree.SelectedID() != "t-1" {
		t.Errorf("expected t-1, got %s", tree.SelectedID())
	}

	tree.MoveDown()
	if tree.SelectedID() != "t-2" {
		t.Errorf("expected t-2, got %s", tree.SelectedID())
	}

	tree.MoveUp()
	if tree.SelectedID() != "t-1" {
		t.Errorf("expected t-1, got %s", tree.SelectedID())
	}
}

func TestTreeEmpty(t *testing.T) {
	store := state.NewThreadStore()
	tree := NewTreeModel(store)

	if tree.SelectedID() != "" {
		t.Errorf("expected empty ID, got %s", tree.SelectedID())
	}

	output := tree.View()
	if !strings.Contains(output, "No threads") {
		t.Errorf("expected empty message:\n%s", output)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -v -run TestTree`
Expected: FAIL — `NewTreeModel` not defined

**Step 3: Implement tree component**

```go
// internal/tui/tree.go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/state"
)

var (
	treeSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")).
				Bold(true)
	treeNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))
	treeDimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
)

// TreeModel renders a hierarchical list of threads.
type TreeModel struct {
	store    *state.ThreadStore
	selected int
	flatList []string // flattened thread IDs in display order
}

// NewTreeModel creates a tree backed by the given store.
func NewTreeModel(store *state.ThreadStore) TreeModel {
	tree := TreeModel{store: store}
	tree.rebuild()
	return tree
}

// SelectedID returns the thread ID of the selected node.
func (t *TreeModel) SelectedID() string {
	if len(t.flatList) == 0 {
		return ""
	}
	return t.flatList[t.selected]
}

// MoveDown moves selection down.
func (t *TreeModel) MoveDown() {
	if t.selected < len(t.flatList)-1 {
		t.selected++
	}
}

// MoveUp moves selection up.
func (t *TreeModel) MoveUp() {
	if t.selected > 0 {
		t.selected--
	}
}

// Refresh rebuilds the flat list from the store.
func (t *TreeModel) Refresh() {
	t.rebuild()
}

func (t *TreeModel) rebuild() {
	t.flatList = nil
	roots := t.store.Roots()
	for _, root := range roots {
		t.flatList = append(t.flatList, root.ID)
		t.addChildren(root.ID, 1)
	}
}

func (t *TreeModel) addChildren(parentID string, depth int) {
	children := t.store.Children(parentID)
	for _, child := range children {
		t.flatList = append(t.flatList, child.ID)
		t.addChildren(child.ID, depth+1)
	}
}

func (t *TreeModel) depthOf(id string) int {
	thread, exists := t.store.Get(id)
	if !exists || thread.ParentID == "" {
		return 0
	}
	return 1 + t.depthOf(thread.ParentID)
}

// View renders the tree.
func (t *TreeModel) View() string {
	if len(t.flatList) == 0 {
		return treeDimStyle.Render("No threads")
	}

	var lines []string
	for i, id := range t.flatList {
		thread, exists := t.store.Get(id)
		if !exists {
			continue
		}

		depth := t.depthOf(id)
		indent := strings.Repeat("  ", depth)
		prefix := "├─"
		if depth == 0 {
			prefix = "●"
		}

		label := fmt.Sprintf("%s%s %s", indent, prefix, thread.Title)

		style := treeNormalStyle
		if i == t.selected {
			style = treeSelectedStyle
		}
		lines = append(lines, style.Render(label))
	}

	return strings.Join(lines, "\n")
}
```

**Step 4: Run tests**

Run: `go test ./internal/tui/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/tree.go internal/tui/tree_test.go
git commit -m "feat(tui): hierarchical agent tree component"
```

---

### Task 3: Add Focus System to App Model

**Files:**
- Modify: `internal/tui/app.go`
- Modify: `internal/tui/app_test.go`

**Step 1: Write tests for focus toggling**

Add to `app_test.go`:

```go
func TestAppToggleFocus(t *testing.T) {
	store := state.NewThreadStore()
	store.Add("t-1", "Test")

	app := NewAppModel(store)

	if app.Focus() != FocusCanvas {
		t.Errorf("expected canvas focus, got %d", app.Focus())
	}

	tabKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}
	updated, _ := app.Update(tabKey)
	appModel := updated.(AppModel)

	if appModel.Focus() != FocusTree {
		t.Errorf("expected tree focus, got %d", appModel.Focus())
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -v -run TestAppToggle`
Expected: FAIL — `Focus`, `FocusCanvas`, `FocusTree` not defined

**Step 3: Implement focus system**

Add focus constants and tree to `app.go`:

```go
// Focus panel constants.
const (
	FocusCanvas = iota
	FocusTree
)
```

Add `tree TreeModel` and `focus int` fields to `AppModel`. Update `NewAppModel` to initialize the tree. Add `Focus() int` method. Update `handleKey` to handle `'t'` for toggling focus, and route arrow keys to the focused panel.

Update `View()` to render tree sidebar alongside canvas when tree is focused.

**Step 4: Run tests**

Run: `go test ./internal/tui/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/app.go internal/tui/app_test.go
git commit -m "feat(tui): focus system toggling between canvas and tree"
```
