package tui

import (
	"strings"
	"testing"

	"github.com/robinojw/dj/internal/state"
)

func TestCardRenderShowsTitle(t *testing.T) {
	thread := state.NewThreadState("t-1", "Build web app")
	thread.Status = state.StatusActive

	card := NewCardModel(thread, false)
	output := card.View()

	if !strings.Contains(output, "Build web app") {
		t.Errorf("expected title in output, got:\n%s", output)
	}
}

func TestCardRenderShowsStatus(t *testing.T) {
	thread := state.NewThreadState("t-1", "Test")
	thread.Status = state.StatusActive

	card := NewCardModel(thread, false)
	output := card.View()

	if !strings.Contains(output, "active") {
		t.Errorf("expected status in output, got:\n%s", output)
	}
}

func TestCardRenderSelectedHighlight(t *testing.T) {
	thread := state.NewThreadState("t-1", "Test")
	card := NewCardModel(thread, true)
	selected := card.View()

	card2 := NewCardModel(thread, false)
	unselected := card2.View()

	if selected == unselected {
		t.Error("selected and unselected cards should differ")
	}
}

func TestCardDynamicSize(t *testing.T) {
	thread := state.NewThreadState("t-1", "Test")
	thread.Status = state.StatusActive

	card := NewCardModel(thread, false)
	card.SetSize(50, 10)
	output := card.View()

	if !strings.Contains(output, "Test") {
		t.Errorf("expected title in dynamic card, got:\n%s", output)
	}
}
