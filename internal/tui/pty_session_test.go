package tui

import (
	"strings"
	"sync"
	"testing"
	"time"
)

const ptyTestTimeout = 5 * time.Second

func TestPTYSessionStartAndRender(t *testing.T) {
	var mu sync.Mutex
	var messages []PTYOutputMsg

	session := NewPTYSession(PTYSessionConfig{
		ThreadID: "t-1",
		Command:  "echo",
		Args:     []string{"hello pty"},
		SendMsg: func(msg PTYOutputMsg) {
			mu.Lock()
			messages = append(messages, msg)
			mu.Unlock()
		},
	})

	if err := session.Start(); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	deadline := time.After(ptyTestTimeout)
	for {
		mu.Lock()
		hasExited := false
		for _, msg := range messages {
			if msg.Exited {
				hasExited = true
				break
			}
		}
		mu.Unlock()

		if hasExited {
			break
		}

		select {
		case <-deadline:
			t.Fatal("timed out waiting for process exit")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	rendered := session.Render()
	if !strings.Contains(rendered, "hello pty") {
		t.Errorf("expected 'hello pty' in render output, got %q", rendered)
	}

	if session.Running() {
		t.Error("expected session to not be running after exit")
	}
}

func TestPTYSessionWriteBytes(t *testing.T) {
	var mu sync.Mutex
	var gotOutput bool

	session := NewPTYSession(PTYSessionConfig{
		ThreadID: "t-1",
		Command:  "cat",
		SendMsg: func(msg PTYOutputMsg) {
			mu.Lock()
			if !msg.Exited {
				gotOutput = true
			}
			mu.Unlock()
		},
	})

	if err := session.Start(); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	defer session.Stop()

	err := session.WriteBytes([]byte("test input\n"))
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	deadline := time.After(ptyTestTimeout)
	for {
		mu.Lock()
		ready := gotOutput
		mu.Unlock()

		if ready {
			break
		}

		select {
		case <-deadline:
			t.Fatal("timed out waiting for cat output")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	rendered := session.Render()
	if !strings.Contains(rendered, "test input") {
		t.Errorf("expected 'test input' in render output, got %q", rendered)
	}
}

func TestPTYSessionResize(t *testing.T) {
	session := NewPTYSession(PTYSessionConfig{
		ThreadID: "t-1",
		Command:  "echo",
		Args:     []string{"resize test"},
		SendMsg:  func(msg PTYOutputMsg) {},
	})

	session.Resize(120, 40)

	if session.emulator.Width() != 120 {
		t.Errorf("expected width 120, got %d", session.emulator.Width())
	}
	if session.emulator.Height() != 40 {
		t.Errorf("expected height 40, got %d", session.emulator.Height())
	}
}

func TestPTYSessionExitCallback(t *testing.T) {
	exitCh := make(chan PTYOutputMsg, 10)

	session := NewPTYSession(PTYSessionConfig{
		ThreadID: "t-1",
		Command:  "true",
		SendMsg: func(msg PTYOutputMsg) {
			exitCh <- msg
		},
	})

	if err := session.Start(); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	deadline := time.After(ptyTestTimeout)
	for {
		select {
		case msg := <-exitCh:
			if msg.Exited {
				if msg.ThreadID != "t-1" {
					t.Errorf("expected thread ID t-1, got %s", msg.ThreadID)
				}
				return
			}
		case <-deadline:
			t.Fatal("timed out waiting for exit callback")
		}
	}
}

func TestPTYSessionWriteAfterStop(t *testing.T) {
	session := NewPTYSession(PTYSessionConfig{
		ThreadID: "t-1",
		Command:  "cat",
		SendMsg:  func(msg PTYOutputMsg) {},
	})

	if err := session.Start(); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	session.Stop()

	err := session.WriteBytes([]byte("should fail"))
	if err == nil {
		t.Error("expected error writing after stop")
	}
}

func TestPTYSessionStop(t *testing.T) {
	session := NewPTYSession(PTYSessionConfig{
		ThreadID: "t-1",
		Command:  "sleep",
		Args:     []string{"60"},
		SendMsg:  func(msg PTYOutputMsg) {},
	})

	if err := session.Start(); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	if !session.Running() {
		t.Error("expected session to be running")
	}

	session.Stop()

	if session.Running() {
		t.Error("expected session to be stopped")
	}
}
