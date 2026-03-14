package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/robinojw/dj/internal/api"
)

// streamEventType identifies the kind of stream event.
type streamEventType int

const (
	eventText streamEventType = iota
	eventDone
	eventError
	eventDiff
)

// streamEvent carries a streaming update from the API to the UI.
type streamEvent struct {
	Type      streamEventType
	Delta     string
	Usage     api.Usage
	Err       error
	FilePath  string
	DiffText  string
	Timestamp time.Time
}

// diffStats holds addition/deletion counts for a diff.
type diffStats struct {
	additions int
	deletions int
}

// calculateDiffStats counts +/- lines in a diff, excluding +++/--- markers.
func calculateDiffStats(lines []string) diffStats {
	var s diffStats
	for _, line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			s.additions++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			s.deletions++
		}
	}
	return s
}

// formatDiffSummary returns a collapsed diff summary line, e.g. "▶ src/main.go Modified: +3 -1"
func formatDiffSummary(filePath, diffText string) string {
	lines := strings.Split(diffText, "\n")
	stats := calculateDiffStats(lines)
	return fmt.Sprintf("▶ %s Modified: +%d -%d", filePath, stats.additions, stats.deletions)
}

// storedDiff holds a diff for the diff pager.
type storedDiff struct {
	FilePath  string
	DiffLines []string
	Timestamp time.Time
}

// bridgeStreamToChannel reads from api.Stream() channels and sends streamEvents
// to eventCh. It does NOT close eventCh (the channel is reused across streams).
// The goroutine exits when ctx is cancelled or the stream completes.
func bridgeStreamToChannel(ctx context.Context, chunks <-chan api.ResponseChunk, errs <-chan error, eventCh chan<- streamEvent) {
	for {
		select {
		case <-ctx.Done():
			return

		case chunk, ok := <-chunks:
			if !ok {
				select {
				case err := <-errs:
					trySend(ctx, eventCh, streamEvent{Type: eventError, Err: classifyError(err)})
				default:
					trySend(ctx, eventCh, streamEvent{Type: eventDone, Usage: api.Usage{}})
				}
				return
			}

			switch chunk.Type {
			case "response.output_text.delta":
				if chunk.Delta != "" {
					trySend(ctx, eventCh, streamEvent{Type: eventText, Delta: chunk.Delta})
				}
			case "response.completed":
				usage := api.Usage{}
				if chunk.Response != nil {
					usage = chunk.Response.Usage
				}
				trySend(ctx, eventCh, streamEvent{Type: eventDone, Usage: usage})
				return
			}

		case err, ok := <-errs:
			if ok {
				trySend(ctx, eventCh, streamEvent{Type: eventError, Err: classifyError(err)})
			} else {
				trySend(ctx, eventCh, streamEvent{Type: eventDone, Usage: api.Usage{}})
			}
			return
		}
	}
}

// trySend sends an event to the channel, aborting if the context is cancelled.
func trySend(ctx context.Context, ch chan<- streamEvent, ev streamEvent) {
	select {
	case ch <- ev:
	case <-ctx.Done():
	}
}

// classifyError wraps errors with user-friendly context.
func classifyError(err error) error {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "401") || strings.Contains(msg, "authentication"):
		return fmt.Errorf("authentication failed: check OPENAI_API_KEY environment variable")
	case strings.Contains(msg, "404"):
		return fmt.Errorf("model not found: check model name in config")
	case strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline"):
		return fmt.Errorf("request timeout: API server took too long to respond")
	case strings.Contains(msg, "connection refused"):
		return fmt.Errorf("cannot connect to API: check network and base URL")
	default:
		return fmt.Errorf("stream error: %w", err)
	}
}
