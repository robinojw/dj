package appserver

import (
	"context"
	"testing"
	"time"
)

func TestClientStartStop(t *testing.T) {
	// Use 'cat' as a mock app-server: it reads stdin and echoes to stdout.
	// This verifies process lifecycle without a real codex binary.
	client := NewClient("cat")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if !client.Running() {
		t.Fatal("expected client to be running")
	}

	if err := client.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	if client.Running() {
		t.Fatal("expected client to be stopped")
	}
}
