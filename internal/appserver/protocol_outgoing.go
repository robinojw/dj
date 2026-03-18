package appserver

type jsonRPCOutgoing struct {
	JSONRPC string `json:"jsonrpc"`
	ID      string `json:"id"`
}

// JSONRPCRequest is an outgoing client-to-server request.
type JSONRPCRequest struct {
	jsonRPCOutgoing
	Method string      `json:"method"`
	Params interface{} `json:"params,omitempty"`
}

// JSONRPCResponse is an outgoing client response to a server request.
type JSONRPCResponse struct {
	jsonRPCOutgoing
	Result interface{} `json:"result"`
}
