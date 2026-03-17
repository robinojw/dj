package tui

import (
	"strings"
	"testing"
)

func TestStatusBarConnected(t *testing.T) {
	bar := NewStatusBar()
	bar.SetConnected(true)
	bar.SetThreadCount(3)
	bar.SetSelectedThread("Build web app")

	output := bar.View()

	if !strings.Contains(output, "Connected") {
		t.Errorf("expected Connected in output:\n%s", output)
	}
	if !strings.Contains(output, "3 threads") {
		t.Errorf("expected thread count in output:\n%s", output)
	}
	if !strings.Contains(output, "Build web app") {
		t.Errorf("expected selected thread in output:\n%s", output)
	}
}

func TestStatusBarDisconnected(t *testing.T) {
	bar := NewStatusBar()
	bar.SetConnected(false)

	output := bar.View()

	if !strings.Contains(output, "Disconnected") {
		t.Errorf("expected Disconnected in output:\n%s", output)
	}
	if !strings.Contains(output, "codex") {
		t.Errorf("expected codex hint in disconnected output:\n%s", output)
	}
}

func TestStatusBarConnecting(t *testing.T) {
	bar := NewStatusBar()
	bar.SetConnecting()

	output := bar.View()

	if !strings.Contains(output, "Connecting") {
		t.Errorf("expected Connecting in output:\n%s", output)
	}
}

func TestStatusBarConnectingClearedOnConnect(t *testing.T) {
	bar := NewStatusBar()
	bar.SetConnecting()
	bar.SetConnected(true)

	output := bar.View()

	if strings.Contains(output, "Connecting") {
		t.Errorf("expected Connecting cleared after connect:\n%s", output)
	}
	if !strings.Contains(output, "Connected") {
		t.Errorf("expected Connected in output:\n%s", output)
	}
}

func TestStatusBarError(t *testing.T) {
	bar := NewStatusBar()
	bar.SetError("connection lost")

	output := bar.View()

	if !strings.Contains(output, "connection lost") {
		t.Errorf("expected error in output:\n%s", output)
	}
}
