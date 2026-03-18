package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/appserver"
)

// V2MessageToMsg routes a JSON-RPC message by method to the appropriate Bubble Tea message.
func V2MessageToMsg(message appserver.JsonRpcMessage) tea.Msg {
	switch message.Method {
	case appserver.MethodThreadStarted:
		return decodeThreadStarted(message.Params)
	case appserver.MethodTurnStarted:
		return decodeTurnStarted(message.Params)
	case appserver.MethodTurnCompleted:
		return decodeTurnCompleted(message.Params)
	case appserver.MethodAgentMessageDelta:
		return decodeV2AgentDelta(message.Params)
	case appserver.MethodThreadStatusChanged:
		return decodeThreadStatusChanged(message.Params)
	case appserver.MethodExecApproval:
		return decodeV2ExecApproval(message)
	case appserver.MethodFileApproval:
		return decodeV2FileApproval(message)
	case appserver.MethodCollabSpawnEnd:
		return decodeCollabSpawnEnd(message.Params)
	case appserver.MethodCollabCloseEnd:
		return decodeCollabCloseEnd(message.Params)
	}
	return nil
}
