package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/state"
)

func TestAppKillSessionRemovesThread(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testTitleThread1)
	store.Add(testThreadID2, testTitleThread2)
	app := NewAppModel(store)

	kKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	updated, _ := app.Update(kKey)
	appModel := updated.(AppModel)

	threads := store.All()
	if len(threads) != 1 {
		test.Fatalf("expected 1 thread after kill, got %d", len(threads))
	}
	if threads[0].ID != testThreadID2 {
		test.Errorf("expected %s remaining, got %s", testThreadID2, threads[0].ID)
	}
	_ = appModel
}

func TestAppKillSessionStopsPTY(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testTitleThread1)
	app := NewAppModel(store, WithInteractiveCommand(testCommandCat))

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	app = updated.(AppModel)

	escKey := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ = app.Update(escKey)
	app = updated.(AppModel)

	kKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	updated, _ = app.Update(kKey)
	appModel := updated.(AppModel)

	if len(appModel.ptySessions) != 0 {
		test.Errorf("expected 0 PTY sessions after kill, got %d", len(appModel.ptySessions))
	}
}

func TestAppKillSessionUnpins(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testTitleThread1)
	app := NewAppModel(store, WithInteractiveCommand(testCommandCat))

	spaceKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := app.Update(spaceKey)
	app = updated.(AppModel)

	escKey := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ = app.Update(escKey)
	app = updated.(AppModel)

	kKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	updated, _ = app.Update(kKey)
	appModel := updated.(AppModel)

	if len(appModel.sessionPanel.PinnedSessions()) != 0 {
		test.Errorf("expected 0 pinned after kill, got %d", len(appModel.sessionPanel.PinnedSessions()))
	}
}

func TestAppKillSessionClampsSelection(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testTitleThread1)
	store.Add(testThreadID2, testTitleThread2)
	app := NewAppModel(store)

	rightKey := tea.KeyMsg{Type: tea.KeyRight}
	updated, _ := app.Update(rightKey)
	app = updated.(AppModel)

	kKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	updated, _ = app.Update(kKey)
	appModel := updated.(AppModel)

	if appModel.canvas.SelectedIndex() != 0 {
		test.Errorf("expected selection clamped to 0, got %d", appModel.canvas.SelectedIndex())
	}
}

func TestAppKillSessionWithNoThreadsDoesNothing(test *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	kKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	updated, _ := app.Update(kKey)
	appModel := updated.(AppModel)

	if len(store.All()) != 0 {
		test.Errorf("expected 0 threads, got %d", len(store.All()))
	}
	_ = appModel
}

func TestAppKillSessionReturnsFocusToCanvas(test *testing.T) {
	store := state.NewThreadStore()
	store.Add(testThreadID1, testTitleThread1)
	app := NewAppModel(store, WithInteractiveCommand(testCommandCat))

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	app = updated.(AppModel)

	escKey := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ = app.Update(escKey)
	app = updated.(AppModel)

	kKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	updated, _ = app.Update(kKey)
	appModel := updated.(AppModel)

	if appModel.FocusPane() != FocusPaneCanvas {
		test.Errorf("expected canvas focus after killing last pinned, got %d", appModel.FocusPane())
	}
}
