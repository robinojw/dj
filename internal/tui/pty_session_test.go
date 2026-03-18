package tui

import (
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	ptyTestTimeout       = 5 * time.Second
	ptyPollInterval      = 10 * time.Millisecond
	ptyExitChannelBuffer = 10
	ptyResizeWidth       = 120
	ptyResizeHeight      = 40
	ptyScrollUpAmount    = 5
	ptyScrollDownAmount  = 3

	ptyTestThreadID = "t-1"
	testCmdEcho    = "echo"
	testCmdCat     = "cat"
	testArgHelloPty = "hello pty"
	testArgTest     = "test"
	testStartFailed = "start failed: %v"
)

func noopSendMsg(msg PTYOutputMsg) {}

func TestPTYSessionStartAndRender(testing *testing.T) {
	var mu sync.Mutex
	var messages []PTYOutputMsg

	session := NewPTYSession(PTYSessionConfig{
		ThreadID: ptyTestThreadID,
		Command:  testCmdEcho,
		Args:     []string{testArgHelloPty},
		SendMsg: func(msg PTYOutputMsg) {
			mu.Lock()
			messages = append(messages, msg)
			mu.Unlock()
		},
	})

	if err := session.Start(); err != nil {
		testing.Fatalf(testStartFailed, err)
	}

	deadline := time.After(ptyTestTimeout)
	for {
		hasExited := checkForExit(&mu, messages)

		if hasExited {
			break
		}

		select {
		case <-deadline:
			testing.Fatal("timed out waiting for process exit")
		default:
			time.Sleep(ptyPollInterval)
		}
	}

	rendered := session.Render()
	if !strings.Contains(rendered, testArgHelloPty) {
		testing.Errorf("expected 'hello pty' in render output, got %q", rendered)
	}

	if session.Running() {
		testing.Error("expected session to not be running after exit")
	}
}

func checkForExit(mu *sync.Mutex, messages []PTYOutputMsg) bool {
	mu.Lock()
	defer mu.Unlock()

	for _, msg := range messages {
		if msg.Exited {
			return true
		}
	}
	return false
}

func TestPTYSessionWriteBytes(testing *testing.T) {
	var mu sync.Mutex
	var gotOutput bool

	session := NewPTYSession(PTYSessionConfig{
		ThreadID: ptyTestThreadID,
		Command:  testCmdCat,
		SendMsg: func(msg PTYOutputMsg) {
			mu.Lock()
			if !msg.Exited {
				gotOutput = true
			}
			mu.Unlock()
		},
	})

	if err := session.Start(); err != nil {
		testing.Fatalf(testStartFailed, err)
	}
	defer session.Stop()

	err := session.WriteBytes([]byte("test input\n"))
	if err != nil {
		testing.Fatalf("write failed: %v", err)
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
			testing.Fatal("timed out waiting for cat output")
		default:
			time.Sleep(ptyPollInterval)
		}
	}

	rendered := session.Render()
	if !strings.Contains(rendered, "test input") {
		testing.Errorf("expected 'test input' in render output, got %q", rendered)
	}
}

func TestPTYSessionResize(testing *testing.T) {
	session := NewPTYSession(PTYSessionConfig{
		ThreadID: ptyTestThreadID,
		Command:  testCmdEcho,
		Args:     []string{"resize test"},
		SendMsg:  noopSendMsg,
	})

	session.Resize(ptyResizeWidth, ptyResizeHeight)

	if session.emulator.Width() != ptyResizeWidth {
		testing.Errorf("expected width %d, got %d", ptyResizeWidth, session.emulator.Width())
	}
	if session.emulator.Height() != ptyResizeHeight {
		testing.Errorf("expected height %d, got %d", ptyResizeHeight, session.emulator.Height())
	}
}

func TestPTYSessionExitCallback(testing *testing.T) {
	exitCh := make(chan PTYOutputMsg, ptyExitChannelBuffer)

	session := NewPTYSession(PTYSessionConfig{
		ThreadID: ptyTestThreadID,
		Command:  "true",
		SendMsg: func(msg PTYOutputMsg) {
			exitCh <- msg
		},
	})

	if err := session.Start(); err != nil {
		testing.Fatalf(testStartFailed, err)
	}

	deadline := time.After(ptyTestTimeout)
	for {
		select {
		case msg := <-exitCh:
			if !msg.Exited {
				continue
			}
			if msg.ThreadID != ptyTestThreadID {
				testing.Errorf("expected thread ID %s, got %s", ptyTestThreadID, msg.ThreadID)
			}
			return
		case <-deadline:
			testing.Fatal("timed out waiting for exit callback")
		}
	}
}

func TestPTYSessionWriteAfterStop(testing *testing.T) {
	session := NewPTYSession(PTYSessionConfig{
		ThreadID: ptyTestThreadID,
		Command:  testCmdCat,
		SendMsg:  noopSendMsg,
	})

	if err := session.Start(); err != nil {
		testing.Fatalf(testStartFailed, err)
	}

	session.Stop()

	err := session.WriteBytes([]byte("should fail"))
	if err == nil {
		testing.Error("expected error writing after stop")
	}
}

func TestPTYSessionStop(testing *testing.T) {
	session := NewPTYSession(PTYSessionConfig{
		ThreadID: ptyTestThreadID,
		Command:  "sleep",
		Args:     []string{"60"},
		SendMsg:  noopSendMsg,
	})

	if err := session.Start(); err != nil {
		testing.Fatalf(testStartFailed, err)
	}

	if !session.Running() {
		testing.Error("expected session to be running")
	}

	session.Stop()

	if session.Running() {
		testing.Error("expected session to be stopped")
	}
}

func TestPTYSessionScrollOffset(testing *testing.T) {
	session := NewPTYSession(PTYSessionConfig{
		ThreadID: ptyTestThreadID,
		Command:  testCmdEcho,
		Args:     []string{testArgTest},
		SendMsg:  noopSendMsg,
	})

	if session.ScrollOffset() != 0 {
		testing.Errorf("expected initial offset 0, got %d", session.ScrollOffset())
	}

	if session.IsScrolledUp() {
		testing.Error("expected not scrolled up initially")
	}
}

func TestPTYSessionScrollUpDown(testing *testing.T) {
	session := NewPTYSession(PTYSessionConfig{
		ThreadID: ptyTestThreadID,
		Command:  testCmdEcho,
		Args:     []string{testArgTest},
		SendMsg:  noopSendMsg,
	})

	session.ScrollUp(ptyScrollUpAmount)
	if session.ScrollOffset() != 0 {
		testing.Errorf("expected offset 0 with no scrollback, got %d", session.ScrollOffset())
	}

	session.ScrollDown(ptyScrollDownAmount)
	if session.ScrollOffset() != 0 {
		testing.Errorf("expected offset 0 after scroll down, got %d", session.ScrollOffset())
	}
}

func TestPTYSessionScrollToBottom(testing *testing.T) {
	session := NewPTYSession(PTYSessionConfig{
		ThreadID: ptyTestThreadID,
		Command:  testCmdEcho,
		Args:     []string{testArgTest},
		SendMsg:  noopSendMsg,
	})

	session.ScrollToBottom()
	if session.ScrollOffset() != 0 {
		testing.Errorf("expected offset 0 after scroll to bottom, got %d", session.ScrollOffset())
	}
	if session.IsScrolledUp() {
		testing.Error("expected not scrolled up after scroll to bottom")
	}
}
