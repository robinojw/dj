//go:build integration

package appserver

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestIntegrationProtoConnect(t *testing.T) {
	client := NewClient("codex", "proto")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := client.Start(ctx); err != nil {
		t.Fatalf("Failed to start codex proto: %v", err)
	}
	defer client.Stop()

	events := make(chan ProtoEvent, 10)
	go client.ReadLoop(func(event ProtoEvent) {
		events <- event
	})

	select {
	case event := <-events:
		var header EventHeader
		if err := json.Unmarshal(event.Msg, &header); err != nil {
			t.Fatalf("unmarshal header: %v", err)
		}
		if header.Type != EventSessionConfigured {
			t.Errorf("expected session_configured, got %s", header.Type)
		}
		t.Logf("Connected: received %s event", header.Type)
	case <-ctx.Done():
		t.Fatal("timeout waiting for session_configured")
	}
}
