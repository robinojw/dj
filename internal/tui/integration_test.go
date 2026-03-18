//go:build integration

package tui

import (
	"context"
	"testing"
	"time"

	"github.com/robinojw/dj/internal/appserver"
	"github.com/robinojw/dj/internal/state"
)

const integrationTimeout = 15 * time.Second

func TestIntegrationEndToEnd(test *testing.T) {
	client := appserver.NewClient("codex", "proto")

	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	defer cancel()

	if err := client.Start(ctx); err != nil {
		test.Fatalf("Start failed: %v", err)
	}
	defer client.Stop()

	store := state.NewThreadStore()
	events := make(chan SessionConfiguredMsg, 1)

	go client.ReadLoop(func(message appserver.JsonRpcMessage) {
		msg := ProtoEventToMsg(message)
		if configured, ok := msg.(SessionConfiguredMsg); ok {
			store.Add(configured.SessionID, configured.Model)
			events <- configured
		}
	})

	select {
	case configured := <-events:
		test.Logf("Connected: session %s, model %s", configured.SessionID, configured.Model)
	case <-ctx.Done():
		test.Fatal("timeout waiting for session_configured")
	}

	threads := store.All()
	if len(threads) != 1 {
		test.Fatalf("expected 1 thread, got %d", len(threads))
	}
}
