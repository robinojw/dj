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
	events := make(chan ThreadStartedMsg, 1)

	go client.ReadLoop(func(message appserver.JSONRPCMessage) {
		msg := V2MessageToMsg(message)
		if started, ok := msg.(ThreadStartedMsg); ok {
			store.Add(started.ThreadID, started.ThreadID)
			events <- started
		}
	})

	select {
	case started := <-events:
		test.Logf("Connected: thread %s started", started.ThreadID)
	case <-ctx.Done():
		test.Fatal("timeout waiting for thread_started")
	}

	threads := store.All()
	if len(threads) != 1 {
		test.Fatalf("expected 1 thread, got %d", len(threads))
	}
}
