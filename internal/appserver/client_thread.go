package appserver

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) StartThread(ctx context.Context, model string) (*ThreadStartResult, error) {
	params, _ := json.Marshal(ThreadStartParams{
		Model: model,
	})

	resp, err := c.Call(ctx, MethodThreadStart, params)
	if err != nil {
		return nil, fmt.Errorf("thread/start: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("thread/start: %w", resp.Error)
	}

	var result ThreadStartResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("unmarshal thread/start result: %w", err)
	}
	return &result, nil
}

func (c *Client) ListThreads(ctx context.Context) (*ThreadListResult, error) {
	resp, err := c.Call(ctx, MethodThreadList, json.RawMessage(`{}`))
	if err != nil {
		return nil, fmt.Errorf("thread/list: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("thread/list: %w", resp.Error)
	}

	var result ThreadListResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("unmarshal thread/list result: %w", err)
	}
	return &result, nil
}

func (c *Client) ArchiveThread(ctx context.Context, threadID string) error {
	params, _ := json.Marshal(ThreadArchiveParams{
		ThreadID: threadID,
	})

	resp, err := c.Call(ctx, MethodThreadArchive, params)
	if err != nil {
		return fmt.Errorf("thread/archive: %w", err)
	}
	if resp.Error != nil {
		return fmt.Errorf("thread/archive: %w", resp.Error)
	}
	return nil
}
