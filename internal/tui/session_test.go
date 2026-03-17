package tui

import (
	"strings"
	"testing"

	"github.com/robinojw/dj/internal/state"
)

func TestSessionViewShowsThreadTitle(t *testing.T) {
	thread := state.NewThreadState("t-1", "My Task")
	thread.Status = state.StatusActive

	session := NewSessionModel(thread)
	output := session.View()

	if !strings.Contains(output, "My Task") {
		t.Errorf("expected thread title in output:\n%s", output)
	}
}

func TestSessionViewShowsMessages(t *testing.T) {
	thread := state.NewThreadState("t-1", "Test")
	thread.AppendMessage(state.ChatMessage{
		ID: "m-1", Role: "user", Content: "Hello world",
	})

	session := NewSessionModel(thread)
	session.SetSize(80, 24)
	output := session.View()

	if !strings.Contains(output, "Hello world") {
		t.Errorf("expected message content in output:\n%s", output)
	}
}
