//go:build integration

package appserver

import (
	"context"
	"testing"
	"time"
)

const integrationTestTimeout = 15 * time.Second
const integrationEventBuffer = 10

func TestIntegrationV2Connect(test *testing.T) {
	client := NewClient("codex", "proto")

	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()

	if err := client.Start(ctx); err != nil {
		test.Fatalf("Failed to start codex proto: %v", err)
	}
	defer client.Stop()

	events := make(chan JSONRPCMessage, integrationEventBuffer)
	go client.ReadLoop(func(message JSONRPCMessage) {
		events <- message
	})

	select {
	case message := <-events:
		if message.Method == "" {
			test.Fatal("expected a notification with a method")
		}
		test.Logf("Connected: received method %s", message.Method)
	case <-ctx.Done():
		test.Fatal("timeout waiting for first event")
	}
}
