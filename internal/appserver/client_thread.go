package appserver

import (
	"encoding/json"
	"fmt"
)

// SendUserTurn sends a user_turn submission with the given text content.
func (c *Client) SendUserTurn(text string, cwd string, model string) error {
	op := UserTurnOp{
		Type:           OpUserTurn,
		Items:          []UserInput{NewTextInput(text)},
		Cwd:            cwd,
		ApprovalPolicy: "on-request",
		SandboxPolicy:  SandboxPolicyReadOnly(),
		Model:          model,
	}

	opData, err := json.Marshal(op)
	if err != nil {
		return fmt.Errorf("marshal user_turn op: %w", err)
	}

	sub := &Submission{
		ID: c.NextID(),
		Op: opData,
	}
	return c.Send(sub)
}

// SendInterrupt sends an interrupt submission to stop the current turn.
func (c *Client) SendInterrupt() error {
	op := InterruptOp{Type: OpInterrupt}
	opData, _ := json.Marshal(op)

	sub := &Submission{
		ID: c.NextID(),
		Op: opData,
	}
	return c.Send(sub)
}

// SendShutdown sends a shutdown submission.
func (c *Client) SendShutdown() error {
	op := ShutdownOp{Type: OpShutdown}
	opData, _ := json.Marshal(op)

	sub := &Submission{
		ID: c.NextID(),
		Op: opData,
	}
	return c.Send(sub)
}
