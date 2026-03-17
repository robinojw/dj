package appserver

import "encoding/json"

type NotificationRouter struct {
	handlers map[string]func(json.RawMessage)
}

func NewNotificationRouter() *NotificationRouter {
	return &NotificationRouter{
		handlers: make(map[string]func(json.RawMessage)),
	}
}

func (router *NotificationRouter) Handle(method string, params json.RawMessage) {
	handler, exists := router.handlers[method]
	if !exists {
		return
	}
	handler(params)
}

func (router *NotificationRouter) OnThreadStatusChanged(fn func(ThreadStatusChanged)) {
	router.handlers[NotifyThreadStatusChanged] = func(raw json.RawMessage) {
		var params ThreadStatusChanged
		if err := json.Unmarshal(raw, &params); err != nil {
			return
		}
		fn(params)
	}
}

func (router *NotificationRouter) OnThreadMessageCreated(fn func(ThreadMessageCreated)) {
	router.handlers[NotifyThreadMessageCreated] = func(raw json.RawMessage) {
		var params ThreadMessageCreated
		if err := json.Unmarshal(raw, &params); err != nil {
			return
		}
		fn(params)
	}
}

func (router *NotificationRouter) OnThreadMessageDelta(fn func(ThreadMessageDelta)) {
	router.handlers[NotifyThreadMessageDelta] = func(raw json.RawMessage) {
		var params ThreadMessageDelta
		if err := json.Unmarshal(raw, &params); err != nil {
			return
		}
		fn(params)
	}
}

func (router *NotificationRouter) OnCommandOutput(fn func(CommandOutput)) {
	router.handlers[NotifyCommandOutput] = func(raw json.RawMessage) {
		var params CommandOutput
		if err := json.Unmarshal(raw, &params); err != nil {
			return
		}
		fn(params)
	}
}

func (router *NotificationRouter) OnCommandFinished(fn func(CommandFinished)) {
	router.handlers[NotifyCommandFinished] = func(raw json.RawMessage) {
		var params CommandFinished
		if err := json.Unmarshal(raw, &params); err != nil {
			return
		}
		fn(params)
	}
}
