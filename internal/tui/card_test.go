package tui

import (
	"strings"
	"testing"

	"github.com/robinojw/dj/internal/state"
)

const (
	testThreadID     = "t-1"
	testTitleBuild   = "Build web app"
	testTitleGeneric = "Test"
	testActivity     = "Running: git status"
	testLongActivity = "This is a very long activity string that should definitely be truncated when rendered on a small card"
	testLargeWidth   = 50
	testLargeHeight  = 10
)

func TestCardRenderShowsTitle(testing *testing.T) {
	thread := state.NewThreadState(testThreadID, testTitleBuild)
	thread.Status = state.StatusActive

	card := NewCardModel(thread, false)
	output := card.View()

	if !strings.Contains(output, testTitleBuild) {
		testing.Errorf("expected title in output, got:\n%s", output)
	}
}

func TestCardRenderShowsStatus(testing *testing.T) {
	thread := state.NewThreadState(testThreadID, testTitleGeneric)
	thread.Status = state.StatusActive

	card := NewCardModel(thread, false)
	output := card.View()

	if !strings.Contains(output, "active") {
		testing.Errorf("expected status in output, got:\n%s", output)
	}
}

func TestCardRenderSelectedHighlight(testing *testing.T) {
	thread := state.NewThreadState(testThreadID, testTitleGeneric)
	card := NewCardModel(thread, true)
	selected := card.View()

	card2 := NewCardModel(thread, false)
	unselected := card2.View()

	if selected == unselected {
		testing.Error("selected and unselected cards should differ")
	}
}

func TestCardDynamicSize(testing *testing.T) {
	thread := state.NewThreadState(testThreadID, testTitleGeneric)
	thread.Status = state.StatusActive

	card := NewCardModel(thread, false)
	card.SetSize(testLargeWidth, testLargeHeight)
	output := card.View()

	if !strings.Contains(output, testTitleGeneric) {
		testing.Errorf("expected title in dynamic card, got:\n%s", output)
	}
}

func TestCardRenderShowsActivity(testing *testing.T) {
	thread := state.NewThreadState(testThreadID, testTitleGeneric)
	thread.Status = state.StatusActive
	thread.Activity = testActivity

	card := NewCardModel(thread, false)
	card.SetSize(testLargeWidth, testLargeHeight)
	output := card.View()

	if !strings.Contains(output, testActivity) {
		testing.Errorf("expected activity in output, got:\n%s", output)
	}
}

func TestCardRenderFallsBackToStatus(testing *testing.T) {
	thread := state.NewThreadState(testThreadID, testTitleGeneric)
	thread.Status = state.StatusIdle

	card := NewCardModel(thread, false)
	output := card.View()

	if !strings.Contains(output, "idle") {
		testing.Errorf("expected status fallback in output, got:\n%s", output)
	}
}

func TestCardRenderActivityTruncated(testing *testing.T) {
	thread := state.NewThreadState(testThreadID, testTitleGeneric)
	thread.Status = state.StatusActive
	thread.Activity = testLongActivity

	card := NewCardModel(thread, false)
	card.SetSize(minCardWidth, minCardHeight)
	output := card.View()

	if !strings.Contains(output, "...") {
		testing.Errorf("expected truncated activity with ellipsis, got:\n%s", output)
	}
}
