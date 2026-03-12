package screens

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/tui/theme"
)

// ---------------------------------------------------------------------------
// Streaming lifecycle
// ---------------------------------------------------------------------------

// TestChatModel_StreamDelta_NoCopyPanic verifies that streaming deltas through
// the bubbletea Update loop (value-copy semantics) does not trigger a
// "strings: illegal use of non-zero Builder copied by value" panic.
//
// Regression test for: https://github.com/robinojw/dj/issues/8
func TestChatModel_StreamDelta_NoCopyPanic(t *testing.T) {
	m := NewChatModel(theme.DefaultTheme())

	// Simulate bubbletea's update loop. Each call to Update copies
	// ChatModel by value (value receiver). If the buffer were a
	// strings.Builder value (not pointer), the second WriteString
	// would panic because the non-zero builder was copied.
	m1, _ := m.Update(StreamDeltaMsg{Delta: "Hello"})

	// This second update is the one that panicked with value-type buffer:
	// m1's buffer has been written to, and calling Update copies it again.
	m2, _ := m1.Update(StreamDeltaMsg{Delta: " world"})

	if m2.buffer.String() != "Hello world" {
		t.Errorf("buffer = %q, want %q", m2.buffer.String(), "Hello world")
	}
}

// TestChatModel_FullStreamCycle exercises the complete streaming lifecycle:
// submit → multiple deltas → done. Verifies no panic and correct state
// transitions across value-copy boundaries.
func TestChatModel_FullStreamCycle(t *testing.T) {
	m := NewChatModel(theme.DefaultTheme())

	// Simulate receiving a window size first (required for viewport).
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// User submits a message — sets streaming=true and resets buffer.
	m.messages = append(m.messages, chatMessage{Role: "user", Content: "run cd ."})
	m.streaming = true
	m.buffer.Reset()

	// Stream multiple deltas through the value-copy Update loop.
	deltas := []string{"Changed", " directory", " to", " ."}
	for _, d := range deltas {
		m, _ = m.Update(StreamDeltaMsg{Delta: d})
	}

	want := strings.Join(deltas, "")
	if got := m.buffer.String(); got != want {
		t.Errorf("buffer after deltas = %q, want %q", got, want)
	}

	// StreamDone finalises the assistant message.
	m, _ = m.Update(StreamDoneMsg{})

	if m.streaming {
		t.Error("streaming should be false after StreamDoneMsg")
	}
	if m.buffer.Len() != 0 {
		t.Errorf("buffer should be empty after StreamDoneMsg, got %q", m.buffer.String())
	}
	if len(m.messages) < 2 {
		t.Fatalf("expected at least 2 messages, got %d", len(m.messages))
	}
	last := m.messages[len(m.messages)-1]
	if last.Role != "assistant" {
		t.Errorf("last message role = %q, want %q", last.Role, "assistant")
	}
	if last.Content != want {
		t.Errorf("last message content = %q, want %q", last.Content, want)
	}
}

// TestChatModel_BufferResetBetweenStreams ensures that starting a new stream
// after completing one does not carry stale buffer state through the
// value-copy Update loop.
func TestChatModel_BufferResetBetweenStreams(t *testing.T) {
	m := NewChatModel(theme.DefaultTheme())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// First stream cycle.
	m.streaming = true
	m.buffer.Reset()
	m, _ = m.Update(StreamDeltaMsg{Delta: "first response"})
	m, _ = m.Update(StreamDoneMsg{})

	// Second stream cycle — buffer must be independent of the first.
	m.streaming = true
	m.buffer.Reset()
	m, _ = m.Update(StreamDeltaMsg{Delta: "second"})
	m, _ = m.Update(StreamDeltaMsg{Delta: " response"})

	if got := m.buffer.String(); got != "second response" {
		t.Errorf("buffer = %q, want %q", got, "second response")
	}
}

// TestChatModel_StreamError verifies that a StreamErrorMsg appends an error
// message and resets the streaming state.
func TestChatModel_StreamError(t *testing.T) {
	m := NewChatModel(theme.DefaultTheme())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	m.streaming = true
	m.buffer.Reset()
	m, _ = m.Update(StreamDeltaMsg{Delta: "partial"})

	// Error arrives mid-stream.
	m, _ = m.Update(StreamErrorMsg{Err: &testError{msg: "connection reset"}})

	if m.streaming {
		t.Error("streaming should be false after error")
	}
	if m.buffer.Len() != 0 {
		t.Error("buffer should be reset after error")
	}
	if len(m.messages) == 0 {
		t.Fatal("expected an error message to be appended")
	}
	last := m.messages[len(m.messages)-1]
	if last.Role != "assistant" {
		t.Errorf("error message role = %q, want %q", last.Role, "assistant")
	}
	if !strings.Contains(last.Content, "connection reset") {
		t.Errorf("error message = %q, want it to contain %q", last.Content, "connection reset")
	}
}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }

// ---------------------------------------------------------------------------
// Window resize
// ---------------------------------------------------------------------------

func TestChatModel_WindowResize(t *testing.T) {
	m := NewChatModel(theme.DefaultTheme())

	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if m.width != 120 {
		t.Errorf("width = %d, want 120", m.width)
	}
	if m.height != 40 {
		t.Errorf("height = %d, want 40", m.height)
	}
	if m.viewport.Width != 118 { // 120 - 2 for horizontal padding
		t.Errorf("viewport.Width = %d, want 118", m.viewport.Width)
	}
	// viewport height = total height - 4 (room for input + status)
	if m.viewport.Height != 36 {
		t.Errorf("viewport.Height = %d, want 36", m.viewport.Height)
	}
}

// ---------------------------------------------------------------------------
// Diff handling
// ---------------------------------------------------------------------------

func TestChatModel_StreamDiffMsg(t *testing.T) {
	m := NewChatModel(theme.DefaultTheme())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	diff := StreamDiffMsg{
		FilePath:  "main.go",
		DiffText:  "+added line\n-removed line",
		Timestamp: time.Now(),
	}
	m, _ = m.Update(diff)

	if len(m.diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(m.diffs))
	}
	if m.diffs[0].FilePath != "main.go" {
		t.Errorf("diff FilePath = %q, want %q", m.diffs[0].FilePath, "main.go")
	}
	if m.focusedDiffIndex != 0 {
		t.Errorf("focusedDiffIndex = %d, want 0", m.focusedDiffIndex)
	}
	if m.viewportMode != "diff_nav" {
		t.Errorf("viewportMode = %q, want %q", m.viewportMode, "diff_nav")
	}
}

func TestChatModel_MultipleDiffs_AutoFocusLatest(t *testing.T) {
	m := NewChatModel(theme.DefaultTheme())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	for i, path := range []string{"a.go", "b.go", "c.go"} {
		m, _ = m.Update(StreamDiffMsg{
			FilePath:  path,
			DiffText:  "+line",
			Timestamp: time.Now(),
		})
		if m.focusedDiffIndex != i {
			t.Errorf("after diff %d: focusedDiffIndex = %d, want %d", i, m.focusedDiffIndex, i)
		}
	}

	if len(m.diffs) != 3 {
		t.Errorf("expected 3 diffs, got %d", len(m.diffs))
	}
}

// ---------------------------------------------------------------------------
// Diff navigation (viewportMode == "diff_nav")
// ---------------------------------------------------------------------------

func TestChatModel_DiffNavigation_TabCycles(t *testing.T) {
	m := newChatWithDiffs(t, 3)

	// focusedDiffIndex starts at 2 (last added). Tab should cycle to 0.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focusedDiffIndex != 0 {
		t.Errorf("after tab: focusedDiffIndex = %d, want 0", m.focusedDiffIndex)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focusedDiffIndex != 1 {
		t.Errorf("after second tab: focusedDiffIndex = %d, want 1", m.focusedDiffIndex)
	}
}

func TestChatModel_DiffNavigation_ShiftTabCyclesBackward(t *testing.T) {
	m := newChatWithDiffs(t, 3)

	// Start at 2. Shift+Tab should go to 1.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.focusedDiffIndex != 1 {
		t.Errorf("focusedDiffIndex = %d, want 1", m.focusedDiffIndex)
	}

	// Wrap around: 1 → 0 → 2
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.focusedDiffIndex != 2 {
		t.Errorf("focusedDiffIndex = %d, want 2 (wrapped)", m.focusedDiffIndex)
	}
}

func TestChatModel_DiffNavigation_EnterTogglesExpand(t *testing.T) {
	m := newChatWithDiffs(t, 2)
	m.focusedDiffIndex = 0

	if m.diffs[0].Expanded {
		t.Fatal("diff should start collapsed")
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !m.diffs[0].Expanded {
		t.Error("diff should be expanded after Enter")
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.diffs[0].Expanded {
		t.Error("diff should be collapsed after second Enter")
	}
}

func TestChatModel_DiffNavigation_SpaceTogglesExpand(t *testing.T) {
	m := newChatWithDiffs(t, 1)
	m.focusedDiffIndex = 0

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !m.diffs[0].Expanded {
		t.Error("space should toggle diff expansion")
	}
}

func TestChatModel_DiffNavigation_EscExits(t *testing.T) {
	m := newChatWithDiffs(t, 2)

	if m.viewportMode != "diff_nav" {
		t.Fatal("should start in diff_nav mode")
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if m.viewportMode != "chat" {
		t.Errorf("viewportMode = %q, want %q", m.viewportMode, "chat")
	}
	if m.focusedDiffIndex != -1 {
		t.Errorf("focusedDiffIndex = %d, want -1", m.focusedDiffIndex)
	}
}

// ---------------------------------------------------------------------------
// View rendering smoke test
// ---------------------------------------------------------------------------

func TestChatModel_View_DoesNotPanic(t *testing.T) {
	m := NewChatModel(theme.DefaultTheme())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Empty state
	v := m.View()
	if v == "" {
		t.Error("View() returned empty string")
	}

	// With messages and streaming
	m.messages = append(m.messages, chatMessage{Role: "user", Content: "hello"})
	m.streaming = true
	m.buffer.WriteString("thinking...")
	v = m.View()
	if v == "" {
		t.Error("View() returned empty string with content")
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newChatWithDiffs creates a ChatModel pre-populated with n diffs in diff_nav mode.
func newChatWithDiffs(t *testing.T, n int) ChatModel {
	t.Helper()
	m := NewChatModel(theme.DefaultTheme())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	for i := 0; i < n; i++ {
		m, _ = m.Update(StreamDiffMsg{
			FilePath:  strings.Repeat("file", 1) + string(rune('a'+i)) + ".go",
			DiffText:  "+added\n-removed",
			Timestamp: time.Now(),
		})
	}
	return m
}
