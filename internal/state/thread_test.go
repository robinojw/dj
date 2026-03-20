package state

import "testing"

const (
	testThreadID      = "t-1"
	testTitle         = "Test"
	testMessageID     = "m-1"
	testExecID        = "e-1"
	testGreeting      = "Hello"
	testActivity      = "Running: git status"
	testAgentProcess  = "architect-1"

	errExpectedHello  = "expected Hello, got %s"
	errExpectedSGotS  = "expected %s, got %s"
)

func TestNewThreadState(testing *testing.T) {
	thread := NewThreadState(testThreadID, "Build a web app")
	if thread.ID != testThreadID {
		testing.Errorf("expected t-1, got %s", thread.ID)
	}
	if thread.Status != StatusIdle {
		testing.Errorf("expected idle, got %s", thread.Status)
	}
	if len(thread.Messages) != 0 {
		testing.Errorf("expected 0 messages, got %d", len(thread.Messages))
	}
}

func TestThreadStateAppendMessage(testing *testing.T) {
	thread := NewThreadState(testThreadID, testTitle)
	thread.AppendMessage(ChatMessage{
		ID:      testMessageID,
		Role:    "user",
		Content: testGreeting,
	})
	if len(thread.Messages) != 1 {
		testing.Fatalf("expected 1 message, got %d", len(thread.Messages))
	}
	if thread.Messages[0].Content != testGreeting {
		testing.Errorf(errExpectedHello, thread.Messages[0].Content)
	}
}

func TestThreadStateAppendDelta(testing *testing.T) {
	thread := NewThreadState(testThreadID, testTitle)
	thread.AppendMessage(ChatMessage{ID: testMessageID, Role: "assistant", Content: "He"})
	thread.AppendDelta(testMessageID, "llo")

	if thread.Messages[0].Content != testGreeting {
		testing.Errorf(errExpectedHello, thread.Messages[0].Content)
	}
}

func TestThreadStateAppendOutput(testing *testing.T) {
	thread := NewThreadState(testThreadID, testTitle)
	thread.AppendOutput(testExecID, "line 1\n")
	thread.AppendOutput(testExecID, "line 2\n")

	output := thread.CommandOutput[testExecID]
	if output != "line 1\nline 2\n" {
		testing.Errorf("expected combined output, got %q", output)
	}
}

func TestThreadStateSetActivity(testing *testing.T) {
	thread := NewThreadState(testThreadID, testTitle)
	thread.SetActivity(testActivity)

	if thread.Activity != testActivity {
		testing.Errorf(errExpectedSGotS, testActivity, thread.Activity)
	}
}

func TestThreadStateClearActivity(testing *testing.T) {
	thread := NewThreadState(testThreadID, testTitle)
	thread.SetActivity("Thinking...")
	thread.ClearActivity()

	if thread.Activity != "" {
		testing.Errorf("expected empty activity, got %s", thread.Activity)
	}
}

func TestThreadStateAgentProcessID(testing *testing.T) {
	thread := NewThreadState(testThreadID, testTitle)
	if thread.AgentProcessID != "" {
		testing.Error("expected empty AgentProcessID for new thread")
	}
	thread.AgentProcessID = testAgentProcess
	if thread.AgentProcessID != testAgentProcess {
		testing.Errorf(errExpectedSGotS, testAgentProcess, thread.AgentProcessID)
	}
}
