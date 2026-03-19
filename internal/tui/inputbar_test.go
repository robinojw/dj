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
