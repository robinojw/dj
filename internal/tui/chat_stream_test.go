package tui

import (
	"fmt"
	"testing"

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

	bridgeStreamToChannel(chunks, errs, eventCh)

	// Should get: text, text, done
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

	bridgeStreamToChannel(chunks, errs, eventCh)

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

	bridgeStreamToChannel(chunks, errs, eventCh)

	ev := <-eventCh
	if ev.Type != eventDone {
		t.Fatalf("expected done event, got %+v", ev)
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
