package appserver

import (
	"encoding/json"
	"sync/atomic"
	"testing"
)

func TestDispatchRoutesNotificationToRouter(t *testing.T) {
	client := &Client{}
	router := NewNotificationRouter()

	var called atomic.Bool
	router.OnThreadStatusChanged(func(params ThreadStatusChanged) {
		called.Store(true)
		if params.ThreadID != "t-1" {
			t.Errorf("expected t-1, got %s", params.ThreadID)
		}
	})

	client.Router = router

	msg := Message{
		JSONRPC: "2.0",
		Method:  NotifyThreadStatusChanged,
		Params:  json.RawMessage(`{"threadId":"t-1","status":"active","title":"Test"}`),
	}
	client.Dispatch(msg)

	if !called.Load() {
		t.Error("router handler was not called")
	}
}

func TestDispatchFallsBackToOnNotification(t *testing.T) {
	client := &Client{}

	var called atomic.Bool
	client.OnNotification = func(method string, params json.RawMessage) {
		called.Store(true)
	}

	msg := Message{
		JSONRPC: "2.0",
		Method:  "custom/notification",
		Params:  json.RawMessage(`{}`),
	}
	client.Dispatch(msg)

	if !called.Load() {
		t.Error("OnNotification was not called")
	}
}
