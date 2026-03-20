package appserver

import (
	"encoding/json"
	"fmt"
)

const (
	initClientName    = "dj"
	initClientVersion = "0.1.0"
)

func (client *Client) Initialize() error {
	requestID := client.NextID()
	request := &JSONRPCRequest{
		jsonRPCOutgoing: jsonRPCOutgoing{JSONRPC: jsonRPCVersion, ID: requestID},
		Method:          MethodInitialize,
		Params: map[string]interface{}{
			"clientInfo": map[string]string{
				"name":    initClientName,
				"version": initClientVersion,
			},
		},
	}
	return client.Send(request)
}

// Send writes a JSON-RPC request to the child's stdin as a JSONL line.
func (client *Client) Send(request *JSONRPCRequest) error {
	return client.writeJSON(request)
}

// SendResponse writes a JSON-RPC response to the child's stdin as a JSONL line.
func (client *Client) SendResponse(response *JSONRPCResponse) error {
	return client.writeJSON(response)
}

func (client *Client) writeJSON(payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	client.mu.Lock()
	defer client.mu.Unlock()

	data = append(data, '\n')
	_, err = client.stdin.Write(data)
	return err
}

// SendUserInput sends a text message via turn/start.
func (client *Client) SendUserInput(text string) (string, error) {
	requestID := client.NextID()
	request := &JSONRPCRequest{
		jsonRPCOutgoing: jsonRPCOutgoing{JSONRPC: jsonRPCVersion, ID: requestID},
		Method:          MethodTurnStart,
		Params:          map[string]string{"message": text},
	}
	return requestID, client.Send(request)
}

// SendInterrupt sends a turn/interrupt request.
func (client *Client) SendInterrupt() error {
	requestID := client.NextID()
	request := &JSONRPCRequest{
		jsonRPCOutgoing: jsonRPCOutgoing{JSONRPC: jsonRPCVersion, ID: requestID},
		Method:          MethodTurnInterrupt,
	}
	return client.Send(request)
}

// SendApproval responds to a server approval request.
func (client *Client) SendApproval(requestID string, approved bool) error {
	response := &JSONRPCResponse{
		jsonRPCOutgoing: jsonRPCOutgoing{JSONRPC: jsonRPCVersion, ID: requestID},
		Result:          map[string]bool{"approved": approved},
	}
	return client.SendResponse(response)
}
