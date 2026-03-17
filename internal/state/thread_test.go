package state

import "testing"

func TestNewThreadState(t *testing.T) {
	thread := NewThreadState("t-1", "Build a web app")
	if thread.ID != "t-1" {
		t.Errorf("expected t-1, got %s", thread.ID)
	}
	if thread.Status != StatusIdle {
		t.Errorf("expected idle, got %s", thread.Status)
	}
	if len(thread.Messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(thread.Messages))
	}
}

func TestThreadStateAppendMessage(t *testing.T) {
	thread := NewThreadState("t-1", "Test")
	thread.AppendMessage(ChatMessage{
		ID:      "m-1",
		Role:    "user",
		Content: "Hello",
	})
	if len(thread.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(thread.Messages))
	}
	if thread.Messages[0].Content != "Hello" {
		t.Errorf("expected Hello, got %s", thread.Messages[0].Content)
	}
}

func TestThreadStateAppendDelta(t *testing.T) {
	thread := NewThreadState("t-1", "Test")
	thread.AppendMessage(ChatMessage{ID: "m-1", Role: "assistant", Content: "He"})
	thread.AppendDelta("m-1", "llo")

	if thread.Messages[0].Content != "Hello" {
		t.Errorf("expected Hello, got %s", thread.Messages[0].Content)
	}
}

func TestThreadStateAppendOutput(t *testing.T) {
	thread := NewThreadState("t-1", "Test")
	thread.AppendOutput("e-1", "line 1\n")
	thread.AppendOutput("e-1", "line 2\n")

	output := thread.CommandOutput["e-1"]
	if output != "line 1\nline 2\n" {
		t.Errorf("expected combined output, got %q", output)
	}
}
