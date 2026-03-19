package tui

import (
	"strings"
	"testing"
)

const (
	statusBarTestWidth     = 80
	statusBarTestThreads   = 3
	statusBarTestAgents    = 3
	statusBarTestCompleted = 1
	statusBarTestSelected  = "Build web app"
	statusBarTestError     = "connection lost"
)

func TestStatusBarConnected(testing *testing.T) {
	bar := NewStatusBar()
	bar.SetConnected(true)
	bar.SetThreadCount(statusBarTestThreads)
	bar.SetSelectedThread(statusBarTestSelected)

	output := bar.View()

	if !strings.Contains(output, "Connected") {
		testing.Errorf("expected Connected in output:\n%s", output)
	}
	if !strings.Contains(output, "3 threads") {
		testing.Errorf("expected thread count in output:\n%s", output)
	}
	if !strings.Contains(output, statusBarTestSelected) {
		testing.Errorf("expected selected thread in output:\n%s", output)
	}
}

func TestStatusBarDisconnected(testing *testing.T) {
	bar := NewStatusBar()
	bar.SetConnected(false)

	output := bar.View()

	if !strings.Contains(output, "Disconnected") {
		testing.Errorf("expected Disconnected in output:\n%s", output)
	}
}

func TestStatusBarError(testing *testing.T) {
	bar := NewStatusBar()
	bar.SetError(statusBarTestError)

	output := bar.View()

	if !strings.Contains(output, statusBarTestError) {
		testing.Errorf("expected error in output:\n%s", output)
	}
}

func TestStatusBarAgentCount(testing *testing.T) {
	bar := NewStatusBar()
	bar.SetWidth(statusBarTestWidth)
	bar.SetAgentCount(statusBarTestAgents, statusBarTestCompleted)
	view := bar.View()
	if !strings.Contains(view, "3 agents") {
		testing.Error("expected agent count in status bar")
	}
}
