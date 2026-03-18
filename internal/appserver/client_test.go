package appserver

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"
)

func TestClientStartStop(t *testing.T) {
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

func TestClientSendAndReadLoop(t *testing.T) {
	client := NewClient("cat")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer client.Stop()

	events := make(chan ProtoEvent, 10)
	go client.ReadLoop(func(event ProtoEvent) {
		events <- event
	})

	sub := &ProtoSubmission{
		ID: "test-1",
		Op: json.RawMessage(`{"type":"user_input"}`),
	}
	if err := client.Send(sub); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	select {
	case event := <-events:
		if event.ID != "test-1" {
			t.Errorf("expected id test-1, got %s", event.ID)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestClientNextID(t *testing.T) {
	client := NewClient("echo")
	first := client.NextID()
	second := client.NextID()
	if first == second {
		t.Error("expected unique IDs")
	}
	if first != "dj-1" {
		t.Errorf("expected dj-1, got %s", first)
	}
	if second != "dj-2" {
		t.Errorf("expected dj-2, got %s", second)
	}
}

func TestClientSendUserInput(t *testing.T) {
	client := NewClient("cat")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer client.Stop()

	events := make(chan ProtoEvent, 10)
	go client.ReadLoop(func(event ProtoEvent) {
		events <- event
	})

	id, err := client.SendUserInput("Hello")
	if err != nil {
		t.Fatalf("SendUserInput failed: %v", err)
	}
	if id == "" {
		t.Error("expected non-empty id")
	}

	select {
	case event := <-events:
		if event.ID != id {
			t.Errorf("expected id %s, got %s", id, event.ID)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestReadLoopParsesProtoEvents(t *testing.T) {
	clientRead, serverWrite := io.Pipe()

	client := &Client{}
	client.stdout = clientRead
	client.scanner = bufio.NewScanner(clientRead)
	client.scanner.Buffer(make([]byte, scannerBufferSize), scannerBufferSize)
	client.running.Store(true)

	events := make(chan ProtoEvent, 10)
	go client.ReadLoop(func(event ProtoEvent) {
		events <- event
	})

	eventJSON := `{"id":"","msg":{"type":"session_configured","session_id":"s-1","model":"o4-mini"}}` + "\n"
	serverWrite.Write([]byte(eventJSON))

	select {
	case event := <-events:
		var header EventHeader
		json.Unmarshal(event.Msg, &header)
		if header.Type != EventSessionConfigured {
			t.Errorf("expected session_configured, got %s", header.Type)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for event")
	}

	serverWrite.Close()
}

func TestClientSendApproval(t *testing.T) {
	client := NewClient("cat")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer client.Stop()

	events := make(chan ProtoEvent, 10)
	go client.ReadLoop(func(event ProtoEvent) {
		events <- event
	})

	err := client.SendApproval("req-1", OpExecApproval, true)
	if err != nil {
		t.Fatalf("SendApproval failed: %v", err)
	}

	select {
	case event := <-events:
		if event.ID != "req-1" {
			t.Errorf("expected id req-1, got %s", event.ID)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}
