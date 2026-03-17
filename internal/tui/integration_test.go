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
	client := appserver.NewClient("codex", "app-server", "--listen", "stdio://")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := client.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer client.Stop()

	router := appserver.NewNotificationRouter()
	client.Router = router
	go client.ReadLoop(client.Dispatch)

	caps, err := client.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	t.Logf("Connected: %s %s", caps.ServerInfo.Name, caps.ServerInfo.Version)

	store := state.NewThreadStore()

	result, err := client.CreateThread(ctx, "Say hello")
	if err != nil {
		t.Fatalf("CreateThread failed: %v", err)
	}
	store.Add(result.ThreadID, "Say hello")
	t.Logf("Created thread: %s", result.ThreadID)

	threads := store.All()
	if len(threads) != 1 {
		t.Fatalf("expected 1 thread, got %d", len(threads))
	}
}
