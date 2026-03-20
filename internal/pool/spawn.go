package pool

import (
	"context"
	"fmt"

	"github.com/robinojw/dj/internal/appserver"
)

func isApprovalRequest(message appserver.JSONRPCMessage) bool {
	isRequest := message.IsRequest()
	isExecApproval := message.Method == appserver.MethodExecApproval
	isFileApproval := message.Method == appserver.MethodFileApproval
	return isRequest && (isExecApproval || isFileApproval)
}

func startAgentProcess(
	ctx context.Context,
	agent *AgentProcess,
	command string,
	args []string,
	events chan<- PoolEvent,
	prompt string,
) error {
	client := appserver.NewClient(command, args...)
	if err := client.Start(ctx); err != nil {
		return fmt.Errorf("start agent %s: %w", agent.ID, err)
	}

	if err := client.Initialize(); err != nil {
		client.Stop()
		return fmt.Errorf("initialize agent %s: %w", agent.ID, err)
	}

	agent.Client = client
	agent.Status = AgentStatusActive

	go runAgentReadLoop(agent, events)

	hasPrompt := prompt != ""
	if !hasPrompt {
		return nil
	}

	if _, err := client.SendUserInput(prompt); err != nil {
		return fmt.Errorf("send prompt to %s: %w", agent.ID, err)
	}

	return nil
}

func runAgentReadLoop(agent *AgentProcess, events chan<- PoolEvent) {
	agent.Client.ReadLoop(func(message appserver.JSONRPCMessage) {
		if isApprovalRequest(message) {
			agent.Client.SendApproval(message.ID, true)
		}
		events <- PoolEvent{AgentID: agent.ID, Message: message}
	})
}
