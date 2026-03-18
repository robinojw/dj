//go:build integration

package tui

import (
	"context"
	"testing"
	"time"

	"github.com/robinojw/dj/internal/appserver"
	"github.com/robinojw/dj/internal/state"
)

func TestIntegrationEndToEnd(t *testing.T) {
	client := appserver.NewClient("codex", "proto")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := client.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer client.Stop()

	store := state.NewThreadStore()
	events := make(chan SessionConfiguredMsg, 1)

	go client.ReadLoop(func(event appserver.ProtoEvent) {
		msg := ProtoEventToMsg(event)
		if configured, ok := msg.(SessionConfiguredMsg); ok {
			store.Add(configured.SessionID, configured.Model)
			events <- configured
		}
	})

	select {
	case configured := <-events:
		t.Logf("Connected: session %s, model %s", configured.SessionID, configured.Model)
	case <-ctx.Done():
		t.Fatal("timeout waiting for session_configured")
	}

	threads := store.All()
	if len(threads) != 1 {
		t.Fatalf("expected 1 thread, got %d", len(threads))
	}
}
