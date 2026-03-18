package tui

import (
	"encoding/json"
	"testing"

	"github.com/robinojw/dj/internal/appserver"
)

const (
	bridgeV2ExpectedTypeFormat = "expected %T, got %T"
	bridgeV2ThreadID           = "t-1"
	bridgeV2ExpectedThreadFmt  = "expected t-1, got %s"
)

func TestBridgeV2ThreadStarted(test *testing.T) {
	message := appserver.JSONRPCMessage{
		Method: appserver.MethodThreadStarted,
		Params: json.RawMessage(`{"thread":{"id":"t-1","status":"idle","source":{"type":"cli"}}}`),
	}
	msg := V2MessageToMsg(message)
	started, ok := msg.(ThreadStartedMsg)
	if !ok {
		test.Fatalf(bridgeV2ExpectedTypeFormat, ThreadStartedMsg{}, msg)
	}
	if started.ThreadID != bridgeV2ThreadID {
		test.Errorf(bridgeV2ExpectedThreadFmt, started.ThreadID)
	}
}

func TestBridgeV2SubAgentThread(test *testing.T) {
	message := appserver.JSONRPCMessage{
		Method: appserver.MethodThreadStarted,
		Params: json.RawMessage(`{"thread":{"id":"t-2","status":"idle","source":{"type":"sub_agent","parent_thread_id":"t-1","depth":1,"agent_nickname":"scout","agent_role":"researcher"}}}`),
	}
	msg := V2MessageToMsg(message)
	started, ok := msg.(ThreadStartedMsg)
	if !ok {
		test.Fatalf(bridgeV2ExpectedTypeFormat, ThreadStartedMsg{}, msg)
	}
	if started.ParentID != bridgeV2ThreadID {
		test.Errorf(bridgeV2ExpectedThreadFmt, started.ParentID)
	}
	if started.AgentRole != "researcher" {
		test.Errorf("expected researcher, got %s", started.AgentRole)
	}
}

func TestBridgeV2TurnStarted(test *testing.T) {
	message := appserver.JSONRPCMessage{
		Method: appserver.MethodTurnStarted,
		Params: json.RawMessage(`{"thread_id":"t-1","turn":{"id":"turn-1","status":"in_progress"}}`),
	}
	msg := V2MessageToMsg(message)
	turn, ok := msg.(TurnStartedMsg)
	if !ok {
		test.Fatalf(bridgeV2ExpectedTypeFormat, TurnStartedMsg{}, msg)
	}
	if turn.ThreadID != bridgeV2ThreadID {
		test.Errorf(bridgeV2ExpectedThreadFmt, turn.ThreadID)
	}
}

func TestBridgeV2AgentDelta(test *testing.T) {
	message := appserver.JSONRPCMessage{
		Method: appserver.MethodAgentMessageDelta,
		Params: json.RawMessage(`{"thread_id":"t-1","delta":"hello"}`),
	}
	msg := V2MessageToMsg(message)
	delta, ok := msg.(V2AgentDeltaMsg)
	if !ok {
		test.Fatalf(bridgeV2ExpectedTypeFormat, V2AgentDeltaMsg{}, msg)
	}
	if delta.Delta != "hello" {
		test.Errorf("expected hello, got %s", delta.Delta)
	}
}

func TestBridgeV2UnknownMethodReturnsNil(test *testing.T) {
	message := appserver.JSONRPCMessage{
		Method: "some/unknown/method",
	}
	msg := V2MessageToMsg(message)
	if msg != nil {
		test.Errorf("expected nil for unknown method, got %T", msg)
	}
}
