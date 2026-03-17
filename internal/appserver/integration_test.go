//go:build integration

package appserver

import (
	"context"
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

	received := make(chan SessionConfigured, 1)
	client.Router = NewEventRouter()
	client.Router.OnSessionConfigured(func(event SessionConfigured) {
		received <- event
	})

	go client.ReadLoop()

	select {
	case event := <-received:
		t.Logf("Connected: session=%s model=%s", event.SessionID, event.Model)
	case <-ctx.Done():
		t.Fatal("timeout waiting for session_configured")
	}
}
