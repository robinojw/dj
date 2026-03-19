package appserver

import (
	"bufio"
	"context"
	"io"
	"testing"
	"time"
)

const (
	clientTestTimeout       = 5 * time.Second
	clientTestEventWait     = 3 * time.Second
	clientTestChannelSize   = 10
	clientTestCommand       = "cat"
	clientTestStartFail     = "Start failed: %v"
	clientTestTimeoutMsg    = "timeout waiting for message"
	clientTestRequestID     = "req-1"
	clientTestSendID        = "test-1"
	clientTestNewline       = "\n"
	clientTestExpectedID    = "expected id %s, got %s"
	clientTestExpectedValue = "expected %s, got %s"
)

func TestClientStartStop(test *testing.T) {
	client := NewClient(clientTestCommand)

	ctx, cancel := context.WithTimeout(context.Background(), clientTestTimeout)
	defer cancel()

	if err := client.Start(ctx); err != nil {
		test.Fatalf(clientTestStartFail, err)
	}

	if !client.Running() {
		test.Fatal("expected client to be running")
	}

	if err := client.Stop(); err != nil {
		test.Fatalf("Stop failed: %v", err)
	}

	if client.Running() {
		test.Fatal("expected client to be stopped")
	}
}

func TestClientSendAndReadLoop(test *testing.T) {
	client := NewClient(clientTestCommand)

	ctx, cancel := context.WithTimeout(context.Background(), clientTestTimeout)
	defer cancel()

	if err := client.Start(ctx); err != nil {
		test.Fatalf(clientTestStartFail, err)
	}
	defer client.Stop()

	messages := make(chan JSONRPCMessage, clientTestChannelSize)
	go client.ReadLoop(func(message JSONRPCMessage) {
		messages <- message
	})

	request := &JSONRPCRequest{
		jsonRPCOutgoing: jsonRPCOutgoing{JSONRPC: jsonRPCVersion, ID: clientTestSendID},
		Method:          MethodTurnStart,
	}
	if err := client.Send(request); err != nil {
		test.Fatalf("Send failed: %v", err)
	}

	select {
	case message := <-messages:
		if message.ID != clientTestSendID {
			test.Errorf(clientTestExpectedID, clientTestSendID, message.ID)
		}
	case <-time.After(clientTestEventWait):
		test.Fatal(clientTestTimeoutMsg)
	}
}

func TestClientNextID(test *testing.T) {
	client := NewClient("echo")
	first := client.NextID()
	second := client.NextID()
	if first == second {
		test.Error("expected unique IDs")
	}
	if first != "dj-1" {
		test.Errorf("expected dj-1, got %s", first)
	}
	if second != "dj-2" {
		test.Errorf("expected dj-2, got %s", second)
	}
}

func TestClientSendUserInput(test *testing.T) {
	client := NewClient(clientTestCommand)

	ctx, cancel := context.WithTimeout(context.Background(), clientTestTimeout)
	defer cancel()

	if err := client.Start(ctx); err != nil {
		test.Fatalf(clientTestStartFail, err)
	}
	defer client.Stop()

	messages := make(chan JSONRPCMessage, clientTestChannelSize)
	go client.ReadLoop(func(message JSONRPCMessage) {
		messages <- message
	})

	requestID, err := client.SendUserInput("Hello")
	if err != nil {
		test.Fatalf("SendUserInput failed: %v", err)
	}
	if requestID == "" {
		test.Error("expected non-empty id")
	}

	select {
	case message := <-messages:
		if message.ID != requestID {
			test.Errorf(clientTestExpectedID, requestID, message.ID)
		}
		if message.Method != MethodTurnStart {
			test.Errorf("expected method %s, got %s", MethodTurnStart, message.Method)
		}
	case <-time.After(clientTestEventWait):
		test.Fatal(clientTestTimeoutMsg)
	}
}

func TestReadLoopParsesV2Notification(test *testing.T) {
	clientRead, serverWrite := io.Pipe()

	client := &Client{}
	client.stdout = clientRead
	client.scanner = bufio.NewScanner(clientRead)
	client.scanner.Buffer(make([]byte, scannerBufferSize), scannerBufferSize)
	client.running.Store(true)

	messages := make(chan JSONRPCMessage, clientTestChannelSize)
	go client.ReadLoop(func(message JSONRPCMessage) {
		messages <- message
	})

	line := `{"jsonrpc":"2.0","method":"thread/started","params":{"thread":{"id":"t-1"}}}` + clientTestNewline
	serverWrite.Write([]byte(line))

	select {
	case message := <-messages:
		if message.Method != MethodThreadStarted {
			test.Errorf(clientTestExpectedValue, MethodThreadStarted, message.Method)
		}
	case <-time.After(clientTestEventWait):
		test.Fatal(clientTestTimeoutMsg)
	}

	serverWrite.Close()
}

func TestReadLoopParsesV2Request(test *testing.T) {
	clientRead, serverWrite := io.Pipe()

	client := &Client{}
	client.stdout = clientRead
	client.scanner = bufio.NewScanner(clientRead)
	client.scanner.Buffer(make([]byte, scannerBufferSize), scannerBufferSize)
	client.running.Store(true)

	messages := make(chan JSONRPCMessage, clientTestChannelSize)
	go client.ReadLoop(func(message JSONRPCMessage) {
		messages <- message
	})

	line := `{"jsonrpc":"2.0","id":"req-1","method":"item/commandExecution/requestApproval","params":{"command":"ls"}}` + clientTestNewline
	serverWrite.Write([]byte(line))

	select {
	case message := <-messages:
		if message.ID != clientTestRequestID {
			test.Errorf(clientTestExpectedValue, clientTestRequestID, message.ID)
		}
		if !message.IsRequest() {
			test.Error("should be a request")
		}
	case <-time.After(clientTestEventWait):
		test.Fatal(clientTestTimeoutMsg)
	}

	serverWrite.Close()
}

func TestClientInitialize(test *testing.T) {
	client := NewClient(clientTestCommand)
	ctx, cancel := context.WithTimeout(context.Background(), clientTestTimeout)
	defer cancel()

	if err := client.Start(ctx); err != nil {
		test.Fatalf(clientTestStartFail, err)
	}
	defer client.Stop()

	messages := make(chan JSONRPCMessage, clientTestChannelSize)
	go client.ReadLoop(func(message JSONRPCMessage) {
		messages <- message
	})

	if err := client.Initialize(); err != nil {
		test.Fatalf("Initialize failed: %v", err)
	}

	select {
	case message := <-messages:
		if message.Method != MethodInitialize {
			test.Errorf(clientTestExpectedValue, MethodInitialize, message.Method)
		}
	case <-time.After(clientTestEventWait):
		test.Fatal(clientTestTimeoutMsg)
	}
}

func TestClientSendApproval(test *testing.T) {
	client := NewClient(clientTestCommand)

	ctx, cancel := context.WithTimeout(context.Background(), clientTestTimeout)
	defer cancel()

	if err := client.Start(ctx); err != nil {
		test.Fatalf(clientTestStartFail, err)
	}
	defer client.Stop()

	messages := make(chan JSONRPCMessage, clientTestChannelSize)
	go client.ReadLoop(func(message JSONRPCMessage) {
		messages <- message
	})

	err := client.SendApproval(clientTestRequestID, true)
	if err != nil {
		test.Fatalf("SendApproval failed: %v", err)
	}

	select {
	case message := <-messages:
		if message.ID != clientTestRequestID {
			test.Errorf(clientTestExpectedID, clientTestRequestID, message.ID)
		}
	case <-time.After(clientTestEventWait):
		test.Fatal(clientTestTimeoutMsg)
	}
}
