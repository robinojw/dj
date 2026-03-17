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

	received := make(chan appserver.SessionConfigured, 1)
	router := appserver.NewEventRouter()
	router.OnSessionConfigured(func(event appserver.SessionConfigured) {
		received <- event
	})
	client.Router = router

	go client.ReadLoop()

	select {
	case event := <-received:
		t.Logf("Connected: session=%s model=%s", event.SessionID, event.Model)

		store := state.NewThreadStore()
		store.Add(event.SessionID, event.Model)

		threads := store.All()
		if len(threads) != 1 {
			t.Fatalf("expected 1 thread, got %d", len(threads))
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for session_configured")
	}
}
