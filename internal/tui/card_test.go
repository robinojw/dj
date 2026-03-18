package tui

import (
	"strings"
	"testing"

	"github.com/robinojw/dj/internal/state"
)

const (
	testThreadID       = "t-1"
	testThreadTitle    = "Test"
	testBuildTitle     = "Build web app"
	testActivity       = "Running: git status"
	testLongActivity   = "This is a very long activity string that should definitely be truncated when rendered on a small card"
	testCardWidth      = 50
	testCardHeight     = 10
	testParentThreadID = "t-0"
	testScoutTitle     = "Scout"
	testResearcherRole = "researcher"
	testSubCardWidth   = 30
	testSubCardHeight  = 6
	testDepthArrow     = "\u21b3"
)

func TestCardRenderShowsTitle(testing *testing.T) {
	thread := state.NewThreadState(testThreadID, testBuildTitle)
	thread.Status = state.StatusActive

	card := NewCardModel(thread, false, false)
	output := card.View()

	if !strings.Contains(output, testBuildTitle) {
		testing.Errorf("expected title in output, got:\n%s", output)
	}
}

func TestCardRenderShowsStatus(testing *testing.T) {
	thread := state.NewThreadState(testThreadID, testThreadTitle)
	thread.Status = state.StatusActive

	card := NewCardModel(thread, false, false)
	output := card.View()

	if !strings.Contains(output, "active") {
		testing.Errorf("expected status in output, got:\n%s", output)
	}
}

func TestCardRenderSelectedHighlight(testing *testing.T) {
	thread := state.NewThreadState(testThreadID, testThreadTitle)
	card := NewCardModel(thread, true, false)
	selected := card.View()

	card2 := NewCardModel(thread, false, false)
	unselected := card2.View()

	if selected == unselected {
		testing.Error("selected and unselected cards should differ")
	}
}

func TestCardDynamicSize(testing *testing.T) {
	thread := state.NewThreadState(testThreadID, testThreadTitle)
	thread.Status = state.StatusActive

	card := NewCardModel(thread, false, false)
	card.SetSize(testCardWidth, testCardHeight)
	output := card.View()

	if !strings.Contains(output, testThreadTitle) {
		testing.Errorf("expected title in dynamic card, got:\n%s", output)
	}
}

func TestCardRenderShowsActivity(testing *testing.T) {
	thread := state.NewThreadState(testThreadID, testThreadTitle)
	thread.Status = state.StatusActive
	thread.Activity = testActivity

	card := NewCardModel(thread, false, false)
	card.SetSize(testCardWidth, testCardHeight)
	output := card.View()

	if !strings.Contains(output, testActivity) {
		testing.Errorf("expected activity in output, got:\n%s", output)
	}
}

func TestCardRenderFallsBackToStatus(testing *testing.T) {
	thread := state.NewThreadState(testThreadID, testThreadTitle)
	thread.Status = state.StatusIdle

	card := NewCardModel(thread, false, false)
	output := card.View()

	if !strings.Contains(output, "idle") {
		testing.Errorf("expected status fallback in output, got:\n%s", output)
	}
}

func TestCardRenderActivityTruncated(testing *testing.T) {
	thread := state.NewThreadState(testThreadID, testThreadTitle)
	thread.Status = state.StatusActive
	thread.Activity = testLongActivity

	card := NewCardModel(thread, false, false)
	card.SetSize(minCardWidth, minCardHeight)
	output := card.View()

	if !strings.Contains(output, "...") {
		testing.Errorf("expected truncated activity with ellipsis, got:\n%s", output)
	}
}

func TestCardPinnedShowsIndicator(testing *testing.T) {
	thread := state.NewThreadState(testThreadID, testBuildTitle)
	thread.Status = state.StatusActive

	card := NewCardModel(thread, false, true)
	card.SetSize(testCardWidth, testCardHeight)
	output := card.View()

	if !strings.Contains(output, "✓") {
		testing.Errorf("expected pinned indicator in output, got:\n%s", output)
	}
}

func TestSubAgentCardShowsRole(test *testing.T) {
	thread := state.NewThreadState(testThreadID, testScoutTitle)
	thread.ParentID = testParentThreadID
	thread.AgentRole = testResearcherRole

	card := NewCardModel(thread, false, false)
	card.SetSize(testSubCardWidth, testSubCardHeight)
	view := card.View()

	if !strings.Contains(view, testResearcherRole) {
		test.Error("expected agent role in card view")
	}
}

func TestSubAgentCardShowsDepthPrefix(test *testing.T) {
	thread := state.NewThreadState(testThreadID, testScoutTitle)
	thread.ParentID = testParentThreadID
	thread.Depth = 1

	card := NewCardModel(thread, false, false)
	card.SetSize(testSubCardWidth, testSubCardHeight)
	view := card.View()

	if !strings.Contains(view, testDepthArrow) {
		test.Error("expected depth prefix in sub-agent card")
	}
}

func TestRootCardNoDepthPrefix(test *testing.T) {
	thread := state.NewThreadState(testParentThreadID, "Root Session")

	card := NewCardModel(thread, false, false)
	card.SetSize(testSubCardWidth, testSubCardHeight)
	view := card.View()

	if strings.Contains(view, testDepthArrow) {
		test.Error("root card should not have depth prefix")
	}
}
