package tui

import (
	"strings"
	"testing"
)

func TestDividerRenderShowsLabels(t *testing.T) {
	sessions := []string{"agent-a", "agent-b"}
	result := renderDividerBar(sessions, 1, 80)

	if !strings.Contains(result, "agent-a") {
		t.Error("expected agent-a in divider")
	}
	if !strings.Contains(result, "agent-b") {
		t.Error("expected agent-b in divider")
	}
}

func TestDividerRenderHighlightsActive(t *testing.T) {
	sessions := []string{"agent-a", "agent-b"}
	result := renderDividerBar(sessions, 0, 80)

	if !strings.Contains(result, "agent-a") {
		t.Error("expected agent-a label present")
	}
}

func TestDividerRenderEmpty(t *testing.T) {
	result := renderDividerBar(nil, 0, 80)
	if result != "" {
		t.Errorf("expected empty string for no sessions, got %q", result)
	}
}

func TestDividerRenderNumbersLabels(t *testing.T) {
	sessions := []string{"a", "b", "c"}
	result := renderDividerBar(sessions, 0, 120)

	if !strings.Contains(result, "1:") {
		t.Error("expected numbered label starting at 1")
	}
	if !strings.Contains(result, "3:") {
		t.Error("expected label 3 for third session")
	}
}

func TestDividerTruncatesLongLabels(t *testing.T) {
	longName := "this-is-a-very-long-session-name-that-exceeds-max"
	sessions := []string{longName}
	result := renderDividerBar(sessions, 0, 120)

	if strings.Contains(result, longName) {
		t.Error("expected long label to be truncated")
	}
	if !strings.Contains(result, "...") {
		t.Error("expected truncated label to contain ellipsis")
	}
}

func TestTruncateLabel(t *testing.T) {
	short := "hello"
	if truncateLabel(short, 20) != "hello" {
		t.Errorf("expected hello, got %s", truncateLabel(short, 20))
	}

	long := "abcdefghijklmnopqrstuvwxyz"
	truncated := truncateLabel(long, 10)
	if len(truncated) != 10 {
		t.Errorf("expected length 10, got %d", len(truncated))
	}
	if !strings.HasSuffix(truncated, "...") {
		t.Error("expected truncated string to end with ...")
	}
}
