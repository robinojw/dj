package appserver

import (
	"encoding/json"
	"sync/atomic"
	"testing"
)

func TestEventRouterDispatchesSessionConfigured(t *testing.T) {
	router := NewEventRouter()

	var called atomic.Bool
	router.OnSessionConfigured(func(event SessionConfigured) {
		called.Store(true)
		if event.SessionID != "sess-1" {
			t.Errorf("expected sess-1, got %s", event.SessionID)
		}
	})

	event := Event{
		ID:  "",
		Msg: json.RawMessage(`{"type":"session_configured","session_id":"sess-1","model":"gpt-4o"}`),
	}
	router.HandleEvent(event)

	if !called.Load() {
		t.Error("handler was not called")
	}
}

func TestEventRouterIgnoresUnregisteredType(t *testing.T) {
	router := NewEventRouter()
	event := Event{
		ID:  "",
		Msg: json.RawMessage(`{"type":"unknown_event"}`),
	}
	router.HandleEvent(event)
}

func TestEventRouterDispatchesError(t *testing.T) {
	router := NewEventRouter()

	var receivedMsg string
	router.OnError(func(event ServerError) {
		receivedMsg = event.Message
	})

	event := Event{
		ID:  "",
		Msg: json.RawMessage(`{"type":"error","message":"test error"}`),
	}
	router.HandleEvent(event)

	if receivedMsg != "test error" {
		t.Errorf("expected 'test error', got %s", receivedMsg)
	}
}
