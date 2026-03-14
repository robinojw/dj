package tui

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/robinojw/dj/internal/api"
)

func TestBridgeStreamToChannel_TextAndDone(t *testing.T) {
	chunks := make(chan api.ResponseChunk, 3)
	errs := make(chan error)
	eventCh := make(chan streamEvent, 10)

	chunks <- api.ResponseChunk{Type: "response.output_text.delta", Delta: "Hello"}
	chunks <- api.ResponseChunk{Type: "response.output_text.delta", Delta: " world"}
	chunks <- api.ResponseChunk{Type: "response.completed", Response: &api.ResponseObject{Usage: api.Usage{InputTokens: 10, OutputTokens: 5}}}
	close(chunks)

	ctx := context.Background()
	bridgeStreamToChannel(ctx, chunks, errs, eventCh)

	ev1 := <-eventCh
	if ev1.Type != eventText || ev1.Delta != "Hello" {
		t.Fatalf("expected text 'Hello', got %+v", ev1)
	}
	ev2 := <-eventCh
	if ev2.Type != eventText || ev2.Delta != " world" {
		t.Fatalf("expected text ' world', got %+v", ev2)
	}
	ev3 := <-eventCh
	if ev3.Type != eventDone || ev3.Usage.InputTokens != 10 {
		t.Fatalf("expected done with 10 input tokens, got %+v", ev3)
	}
}

func TestBridgeStreamToChannel_Error(t *testing.T) {
	chunks := make(chan api.ResponseChunk)
	errs := make(chan error, 1)
	eventCh := make(chan streamEvent, 10)

	errs <- fmt.Errorf("connection refused")

	ctx := context.Background()
	bridgeStreamToChannel(ctx, chunks, errs, eventCh)

	ev := <-eventCh
	if ev.Type != eventError {
		t.Fatalf("expected error event, got %+v", ev)
	}
}

func TestBridgeStreamToChannel_ChunksClosedNoError(t *testing.T) {
	chunks := make(chan api.ResponseChunk)
	errs := make(chan error)
	eventCh := make(chan streamEvent, 10)

	close(chunks)

	ctx := context.Background()
	bridgeStreamToChannel(ctx, chunks, errs, eventCh)

	ev := <-eventCh
	if ev.Type != eventDone {
		t.Fatalf("expected done event, got %+v", ev)
	}
}

func TestBridgeStreamToChannel_CancelStopsGoroutine(t *testing.T) {
	chunks := make(chan api.ResponseChunk) // unbuffered, will block
	errs := make(chan error)
	eventCh := make(chan streamEvent, 10)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		bridgeStreamToChannel(ctx, chunks, errs, eventCh)
	}()

	// Cancel immediately — goroutine should exit
	cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// goroutine exited, good
	case <-time.After(2 * time.Second):
		t.Fatal("bridge goroutine did not exit after context cancellation")
	}
}

func TestBridgeStreamToChannel_CancelDuringSend(t *testing.T) {
	chunks := make(chan api.ResponseChunk, 1)
	errs := make(chan error)
	eventCh := make(chan streamEvent) // unbuffered — send will block

	ctx, cancel := context.WithCancel(context.Background())

	chunks <- api.ResponseChunk{Type: "response.output_text.delta", Delta: "blocked"}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		bridgeStreamToChannel(ctx, chunks, errs, eventCh)
	}()

	// Give goroutine time to try sending, then cancel
	time.Sleep(50 * time.Millisecond)
	cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// goroutine exited via trySend abort
	case <-time.After(2 * time.Second):
		t.Fatal("bridge goroutine did not exit after context cancellation during send")
	}
}

func TestBridgeStreamToChannel_ChannelNotClosed(t *testing.T) {
	chunks := make(chan api.ResponseChunk, 1)
	errs := make(chan error)
	eventCh := make(chan streamEvent, 10)

	chunks <- api.ResponseChunk{Type: "response.completed", Response: &api.ResponseObject{Usage: api.Usage{InputTokens: 1}}}
	close(chunks)

	ctx := context.Background()
	bridgeStreamToChannel(ctx, chunks, errs, eventCh)

	// Read the done event
	ev := <-eventCh
	if ev.Type != eventDone {
		t.Fatalf("expected done, got %+v", ev)
	}

	// Channel should still be open (not closed by bridge)
	// Sending should not panic
	eventCh <- streamEvent{Type: eventText, Delta: "after"}
	ev2 := <-eventCh
	if ev2.Delta != "after" {
		t.Fatalf("expected 'after', channel was closed by bridge")
	}
}

func TestBridgeStreamToChannel_ReuseChannelAcrossStreams(t *testing.T) {
	eventCh := make(chan streamEvent, 20)

	// First stream
	chunks1 := make(chan api.ResponseChunk, 2)
	errs1 := make(chan error)
	chunks1 <- api.ResponseChunk{Type: "response.output_text.delta", Delta: "stream1"}
	chunks1 <- api.ResponseChunk{Type: "response.completed", Response: &api.ResponseObject{}}
	close(chunks1)

	ctx1 := context.Background()
	bridgeStreamToChannel(ctx1, chunks1, errs1, eventCh)

	ev1 := <-eventCh
	if ev1.Delta != "stream1" {
		t.Fatalf("expected 'stream1', got %q", ev1.Delta)
	}
	<-eventCh // done

	// Second stream on same channel
	chunks2 := make(chan api.ResponseChunk, 2)
	errs2 := make(chan error)
	chunks2 <- api.ResponseChunk{Type: "response.output_text.delta", Delta: "stream2"}
	chunks2 <- api.ResponseChunk{Type: "response.completed", Response: &api.ResponseObject{}}
	close(chunks2)

	ctx2 := context.Background()
	bridgeStreamToChannel(ctx2, chunks2, errs2, eventCh)

	ev2 := <-eventCh
	if ev2.Delta != "stream2" {
		t.Fatalf("expected 'stream2', got %q", ev2.Delta)
	}
}

func TestFormatDiffSummary(t *testing.T) {
	diff := "+added line\n-removed line\n context\n+another add"
	result := formatDiffSummary("src/main.go", diff)
	expected := "▶ src/main.go Modified: +2 -1"
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestCalculateDiffStats_ExcludesMarkers(t *testing.T) {
	lines := []string{"--- a/file.go", "+++ b/file.go", "+added", "-removed", " context"}
	stats := calculateDiffStats(lines)
	if stats.additions != 1 || stats.deletions != 1 {
		t.Fatalf("expected 1 add, 1 del, got %+v", stats)
	}
}

func TestClassifyError(t *testing.T) {
	tests := []struct {
		input    string
		contains string
	}{
		{"401 Unauthorized", "authentication failed"},
		{"404 Not Found", "model not found"},
		{"connection refused", "cannot connect"},
		{"deadline exceeded", "request timeout"},
		{"something else", "stream error"},
	}
	for _, tt := range tests {
		err := classifyError(fmt.Errorf("%s", tt.input))
		if err == nil || !contains(err.Error(), tt.contains) {
			t.Errorf("classifyError(%q) = %v, want containing %q", tt.input, err, tt.contains)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
