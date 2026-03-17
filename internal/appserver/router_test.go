package appserver

import (
	"encoding/json"
	"sync/atomic"
	"testing"
)

func TestRouterDispatchesNotification(t *testing.T) {
	router := NewNotificationRouter()

	var called atomic.Bool
	router.OnThreadStatusChanged(func(params ThreadStatusChanged) {
		called.Store(true)
		if params.ThreadID != "t-1" {
			t.Errorf("expected t-1, got %s", params.ThreadID)
		}
	})

	raw := json.RawMessage(`{"threadId":"t-1","status":"active","title":"Test"}`)
	router.Handle(NotifyThreadStatusChanged, raw)

	if !called.Load() {
		t.Error("handler was not called")
	}
}

func TestRouterIgnoresUnregisteredMethod(t *testing.T) {
	router := NewNotificationRouter()
	router.Handle("unknown/method", json.RawMessage(`{}`))
}

func TestRouterDispatchesMessageDelta(t *testing.T) {
	router := NewNotificationRouter()

	var receivedDelta string
	router.OnThreadMessageDelta(func(params ThreadMessageDelta) {
		receivedDelta = params.Delta
	})

	raw := json.RawMessage(`{"threadId":"t-1","messageId":"m-1","delta":"hello"}`)
	router.Handle(NotifyThreadMessageDelta, raw)

	if receivedDelta != "hello" {
		t.Errorf("expected hello, got %s", receivedDelta)
	}
}

func TestRouterDispatchesCommandOutput(t *testing.T) {
	router := NewNotificationRouter()

	var receivedData string
	router.OnCommandOutput(func(params CommandOutput) {
		receivedData = params.Data
	})

	raw := json.RawMessage(`{"threadId":"t-1","execId":"e-1","data":"output line\n"}`)
	router.Handle(NotifyCommandOutput, raw)

	if receivedData != "output line\n" {
		t.Errorf("expected output, got %s", receivedData)
	}
}
