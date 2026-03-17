package appserver

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) CreateThread(ctx context.Context, instructions string) (*ThreadCreateResult, error) {
	params, _ := json.Marshal(ThreadCreateParams{
		Instructions: instructions,
	})

	resp, err := c.Call(ctx, MethodThreadCreate, params)
	if err != nil {
		return nil, fmt.Errorf("thread/create: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("thread/create: %w", resp.Error)
	}

	var result ThreadCreateResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("unmarshal thread/create result: %w", err)
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

func (c *Client) DeleteThread(ctx context.Context, threadID string) error {
	params, _ := json.Marshal(ThreadDeleteParams{
		ThreadID: threadID,
	})

	resp, err := c.Call(ctx, MethodThreadDelete, params)
	if err != nil {
		return fmt.Errorf("thread/delete: %w", err)
	}
	if resp.Error != nil {
		return fmt.Errorf("thread/delete: %w", resp.Error)
	}
	return nil
}
