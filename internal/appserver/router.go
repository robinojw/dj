package appserver

import "encoding/json"

// EventRouter dispatches incoming events by their type field.
type EventRouter struct {
	handlers map[string]func(string, json.RawMessage)
}

// NewEventRouter creates a new event router.
func NewEventRouter() *EventRouter {
	return &EventRouter{
		handlers: make(map[string]func(string, json.RawMessage)),
	}
}

// HandleEvent extracts the event type and dispatches to the registered handler.
func (router *EventRouter) HandleEvent(event Event) {
	var header EventHeader
	if err := json.Unmarshal(event.Msg, &header); err != nil {
		return
	}

	handler, exists := router.handlers[header.Type]
	if !exists {
		return
	}
	handler(event.ID, event.Msg)
}

func (router *EventRouter) OnSessionConfigured(fn func(SessionConfigured)) {
	router.handlers[EventSessionConfigured] = func(_ string, raw json.RawMessage) {
		var event SessionConfigured
		if err := json.Unmarshal(raw, &event); err != nil {
			return
		}
		fn(event)
	}
}

func (router *EventRouter) OnTaskStarted(fn func(TaskStarted)) {
	router.handlers[EventTaskStarted] = func(_ string, raw json.RawMessage) {
		var event TaskStarted
		if err := json.Unmarshal(raw, &event); err != nil {
			return
		}
		fn(event)
	}
}

func (router *EventRouter) OnTaskComplete(fn func(TaskComplete)) {
	router.handlers[EventTaskComplete] = func(_ string, raw json.RawMessage) {
		var event TaskComplete
		if err := json.Unmarshal(raw, &event); err != nil {
			return
		}
		fn(event)
	}
}

func (router *EventRouter) OnAgentMessage(fn func(AgentMessage)) {
	router.handlers[EventAgentMessage] = func(_ string, raw json.RawMessage) {
		var event AgentMessage
		if err := json.Unmarshal(raw, &event); err != nil {
			return
		}
		fn(event)
	}
}

func (router *EventRouter) OnAgentMessageDelta(fn func(AgentMessageDelta)) {
	router.handlers[EventAgentMessageDelta] = func(_ string, raw json.RawMessage) {
		var event AgentMessageDelta
		if err := json.Unmarshal(raw, &event); err != nil {
			return
		}
		fn(event)
	}
}

func (router *EventRouter) OnExecCommandBegin(fn func(ExecCommandBegin)) {
	router.handlers[EventExecCommandBegin] = func(_ string, raw json.RawMessage) {
		var event ExecCommandBegin
		if err := json.Unmarshal(raw, &event); err != nil {
			return
		}
		fn(event)
	}
}

func (router *EventRouter) OnExecCommandOutputDelta(fn func(ExecCommandOutputDelta)) {
	router.handlers[EventExecCommandOutputDelta] = func(_ string, raw json.RawMessage) {
		var event ExecCommandOutputDelta
		if err := json.Unmarshal(raw, &event); err != nil {
			return
		}
		fn(event)
	}
}

func (router *EventRouter) OnExecCommandEnd(fn func(ExecCommandEnd)) {
	router.handlers[EventExecCommandEnd] = func(_ string, raw json.RawMessage) {
		var event ExecCommandEnd
		if err := json.Unmarshal(raw, &event); err != nil {
			return
		}
		fn(event)
	}
}

func (router *EventRouter) OnExecApprovalRequest(fn func(ExecApprovalRequest)) {
	router.handlers[EventExecApprovalRequest] = func(_ string, raw json.RawMessage) {
		var event ExecApprovalRequest
		if err := json.Unmarshal(raw, &event); err != nil {
			return
		}
		fn(event)
	}
}

func (router *EventRouter) OnError(fn func(ServerError)) {
	router.handlers[EventError] = func(_ string, raw json.RawMessage) {
		var event ServerError
		if err := json.Unmarshal(raw, &event); err != nil {
			return
		}
		fn(event)
	}
}
