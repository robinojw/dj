package tui

import (
	"strings"
	"testing"

	"github.com/robinojw/dj/internal/state"
)

func TestRenderMessages(t *testing.T) {
	thread := state.NewThreadState("t-1", "Test")
	thread.AppendMessage(state.ChatMessage{
		ID: "m-1", Role: "user", Content: "Hello",
	})
	thread.AppendMessage(state.ChatMessage{
		ID: "m-2", Role: "assistant", Content: "Hi there",
	})

	output := RenderMessages(thread)

	if !strings.Contains(output, "Hello") {
		t.Errorf("expected user message in output:\n%s", output)
	}
	if !strings.Contains(output, "Hi there") {
		t.Errorf("expected assistant message in output:\n%s", output)
	}
}

func TestRenderMessagesWithCommand(t *testing.T) {
	thread := state.NewThreadState("t-1", "Test")
	thread.AppendOutput("e-1", "go test ./...\nPASS\n")

	output := RenderMessages(thread)

	if !strings.Contains(output, "PASS") {
		t.Errorf("expected command output in output:\n%s", output)
	}
}

func TestRenderMessagesEmpty(t *testing.T) {
	thread := state.NewThreadState("t-1", "Test")
	output := RenderMessages(thread)

	if !strings.Contains(output, "No messages") {
		t.Errorf("expected empty state message:\n%s", output)
	}
}
