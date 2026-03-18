package appserver

type jsonRpcOutgoing struct {
	JsonRpc string `json:"jsonrpc"`
	ID      string `json:"id"`
}

// JsonRpcRequest is an outgoing client-to-server request.
type JsonRpcRequest struct {
	jsonRpcOutgoing
	Method string      `json:"method"`
	Params interface{} `json:"params,omitempty"`
}

// JsonRpcResponse is an outgoing client response to a server request.
type JsonRpcResponse struct {
	jsonRpcOutgoing
	Result interface{} `json:"result"`
}
