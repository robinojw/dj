package tui

import "testing"

func TestMsgTypes(t *testing.T) {
	statusMsg := ThreadStatusMsg{
		ThreadID: "t-1",
		Status:   "active",
		Title:    "Running",
	}
	if statusMsg.ThreadID != "t-1" {
		t.Errorf("expected t-1, got %s", statusMsg.ThreadID)
	}

	messageMsg := ThreadMessageMsg{
		ThreadID:  "t-1",
		MessageID: "m-1",
		Role:      "assistant",
		Content:   "Hello",
	}
	if messageMsg.Role != "assistant" {
		t.Errorf("expected assistant, got %s", messageMsg.Role)
	}

	deltaMsg := ThreadDeltaMsg{
		ThreadID:  "t-1",
		MessageID: "m-1",
		Delta:     "world",
	}
	if deltaMsg.Delta != "world" {
		t.Errorf("expected world, got %s", deltaMsg.Delta)
	}

	outputMsg := CommandOutputMsg{
		ThreadID: "t-1",
		ExecID:   "e-1",
		Data:     "output\n",
	}
	if outputMsg.Data != "output\n" {
		t.Errorf("expected output, got %s", outputMsg.Data)
	}

	finishedMsg := CommandFinishedMsg{
		ThreadID: "t-1",
		ExecID:   "e-1",
		ExitCode: 0,
	}
	if finishedMsg.ExitCode != 0 {
		t.Errorf("expected 0, got %d", finishedMsg.ExitCode)
	}
}
