//go:build integration

package appserver

import (
	"context"
	"testing"
	"time"
)

func TestIntegrationAppServerConnect(t *testing.T) {
	client := NewClient("codex", "app-server", "--listen", "stdio://")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := client.Start(ctx); err != nil {
		t.Fatalf("Failed to start app-server: %v", err)
	}
	defer client.Stop()

	go client.ReadLoop(client.Dispatch)

	caps, err := client.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	t.Logf("Connected to: %s %s", caps.ServerInfo.Name, caps.ServerInfo.Version)
}
