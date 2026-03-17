package appserver

import (
	"bufio"
	"encoding/json"
	"io"
	"testing"
	"time"
)

func TestSendUserTurn(t *testing.T) {
	clientRead, _ := io.Pipe()
	serverRead, clientWrite := io.Pipe()

	client := &Client{}
	client.stdin = clientWrite
	client.stdout = clientRead
	client.scanner = bufio.NewScanner(clientRead)
	client.scanner.Buffer(make([]byte, scannerBufferSize), scannerBufferSize)
	client.running.Store(true)

	received := make(chan map[string]any, 1)
	go func() {
		scanner := bufio.NewScanner(serverRead)
		scanner.Buffer(make([]byte, scannerBufferSize), scannerBufferSize)
		if scanner.Scan() {
			var parsed map[string]any
			json.Unmarshal(scanner.Bytes(), &parsed)
			received <- parsed
		}
	}()

	err := client.SendUserTurn("Hello world", "/tmp", "o4-mini")
	if err != nil {
		t.Fatalf("SendUserTurn failed: %v", err)
	}

	select {
	case msg := <-received:
		if msg["id"] == nil || msg["id"] == "" {
			t.Error("expected non-empty id")
		}
		opRaw, _ := json.Marshal(msg["op"])
		var op map[string]any
		json.Unmarshal(opRaw, &op)
		if op["type"] != OpUserTurn {
			t.Errorf("expected user_turn, got %v", op["type"])
		}
		if op["model"] != "o4-mini" {
			t.Errorf("expected model o4-mini, got %v", op["model"])
		}
		if op["cwd"] != "/tmp" {
			t.Errorf("expected cwd /tmp, got %v", op["cwd"])
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for submission")
	}
}

func TestSendInterrupt(t *testing.T) {
	clientRead, _ := io.Pipe()
	serverRead, clientWrite := io.Pipe()

	client := &Client{}
	client.stdin = clientWrite
	client.stdout = clientRead
	client.scanner = bufio.NewScanner(clientRead)
	client.scanner.Buffer(make([]byte, scannerBufferSize), scannerBufferSize)
	client.running.Store(true)

	received := make(chan map[string]any, 1)
	go func() {
		scanner := bufio.NewScanner(serverRead)
		scanner.Buffer(make([]byte, scannerBufferSize), scannerBufferSize)
		if scanner.Scan() {
			var parsed map[string]any
			json.Unmarshal(scanner.Bytes(), &parsed)
			received <- parsed
		}
	}()

	err := client.SendInterrupt()
	if err != nil {
		t.Fatalf("SendInterrupt failed: %v", err)
	}

	select {
	case msg := <-received:
		opRaw, _ := json.Marshal(msg["op"])
		var op map[string]any
		json.Unmarshal(opRaw, &op)
		if op["type"] != OpInterrupt {
			t.Errorf("expected interrupt, got %v", op["type"])
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for submission")
	}
}
