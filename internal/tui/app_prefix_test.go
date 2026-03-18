package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/state"
)

func TestAppCtrlBMOpensMenu(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testTitleTest)

	app := NewAppModel(store)

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	updated, _ := app.Update(ctrlB)
	app = updated.(AppModel)

	mKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	updated, _ = app.Update(mKey)
	app = updated.(AppModel)

	if !app.MenuVisible() {
		test.Error("expected menu to be visible")
	}
}

func TestAppMenuEscCloses(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testTitleTest)

	app := NewAppModel(store)

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	updated, _ := app.Update(ctrlB)
	app = updated.(AppModel)

	mKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	updated, _ = app.Update(mKey)
	app = updated.(AppModel)

	escKey := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ = app.Update(escKey)
	app = updated.(AppModel)

	if app.MenuVisible() {
		test.Error("expected menu hidden after Esc")
	}
}

func TestAppCtrlBEscCancelsPrefix(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testTitleTest)

	app := NewAppModel(store)

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	updated, _ := app.Update(ctrlB)
	app = updated.(AppModel)

	escKey := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ = app.Update(escKey)
	app = updated.(AppModel)

	if app.MenuVisible() {
		test.Error("expected menu not visible after prefix cancel")
	}
}

func TestAppMenuNavigation(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testTitleTest)

	app := NewAppModel(store)

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	updated, _ := app.Update(ctrlB)
	app = updated.(AppModel)

	mKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	updated, _ = app.Update(mKey)
	app = updated.(AppModel)

	downKey := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ = app.Update(downKey)
	app = updated.(AppModel)

	if app.menu.SelectedIndex() != 1 {
		test.Errorf("expected menu index 1, got %d", app.menu.SelectedIndex())
	}
}

func TestAppCtrlBXUnpinsSession(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testTitleThread1)
	app := NewAppModel(store, WithInteractiveCommand(testCommandEcho, testArgHello))

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
		test.Errorf("expected 0 pinned after unpin, got %d", len(app.sessionPanel.PinnedSessions()))
	}
	if app.FocusPane() != FocusPaneCanvas {
		test.Errorf("expected focus back to canvas, got %d", app.FocusPane())
	}
}

func TestAppCtrlBZTogglesZoom(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testTitleThread1)
	app := NewAppModel(store, WithInteractiveCommand(testCommandEcho, testArgHello))

	spaceKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := app.Update(spaceKey)
	app = updated.(AppModel)

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	updated, _ = app.Update(ctrlB)
	app = updated.(AppModel)

	zKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}}
	updated, _ = app.Update(zKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	if !app.sessionPanel.Zoomed() {
		test.Error("expected zoomed after Ctrl+B z")
	}
}

func TestAppCtrlBRightCyclesPaneRight(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testTitleThread1)
	store.Add(testThreadID2, testTitleThread2)
	app := NewAppModel(store, WithInteractiveCommand(testCommandEcho, testArgHello))

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
		test.Fatalf("expected active pane 0, got %d", app.sessionPanel.ActivePaneIdx())
	}

	ctrlB := tea.KeyMsg{Type: tea.KeyCtrlB}
	updated, _ = app.Update(ctrlB)
	app = updated.(AppModel)

	rightKey := tea.KeyMsg{Type: tea.KeyRight}
	updated, _ = app.Update(rightKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	if app.sessionPanel.ActivePaneIdx() != 1 {
		test.Errorf("expected active pane 1, got %d", app.sessionPanel.ActivePaneIdx())
	}
}

func TestAppViewShowsDividerWhenPinned(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testTitleThread1)
	app := NewAppModel(store, WithInteractiveCommand(testCommandEcho, testArgHello))
	app.width = testAppWidth
	app.height = testAppHeight

	spaceKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := app.Update(spaceKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	view := app.View()
	if !strings.Contains(view, "─") {
		test.Error("expected divider line in view when sessions pinned")
	}
}
