package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/state"
)

const (
	indicatorTestWidth       = 80
	indicatorTestHeight      = 30
	indicatorScrollUp        = 5
	indicatorTestThread      = "t-1"
	indicatorScrollbackLines = 30
)

func TestAppViewShowsScrollIndicator(testing *testing.T) {
	store := state.NewThreadStore()
	store.Add(indicatorTestThread, "Thread 1")

	app := NewAppModel(store, WithInteractiveCommand("cat"))
	app.width = indicatorTestWidth
	app.height = indicatorTestHeight

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := app.Update(enterKey)
	app = updated.(AppModel)
	defer app.StopAllPTYSessions()

	ptySession := app.ptySessions[indicatorTestThread]
	fillScrollback(ptySession)

	ptySession.ScrollUp(indicatorScrollUp)

	view := app.View()
	hasDownArrow := strings.Contains(view, "↓")
	hasLinesBelow := strings.Contains(view, "lines below")
	hasIndicator := hasDownArrow || hasLinesBelow
	if !hasIndicator {
		testing.Error("expected scroll indicator when scrolled up")
	}
}

func fillScrollback(session *PTYSession) {
	for index := 0; index < indicatorScrollbackLines; index++ {
		line := fmt.Sprintf("scrollback line %d\r\n", index)
		session.emulator.Write([]byte(line))
	}
}
