package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const defaultBaseURL = "https://api.openai.com"

// ResponsesClient streams responses from the OpenAI Responses API.
type ResponsesClient struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

func NewResponsesClient(apiKey string) *ResponsesClient {
	return &ResponsesClient{
		httpClient: &http.Client{},
		apiKey:     apiKey,
		baseURL:    defaultBaseURL,
	}
}

func (c *ResponsesClient) WithBaseURL(url string) *ResponsesClient {
	c.baseURL = strings.TrimRight(url, "/")
	return c
}

// Stream sends a request to POST /v1/responses with stream:true and returns
// channels for response chunks and errors.
func (c *ResponsesClient) Stream(
	ctx context.Context,
	req CreateResponseRequest,
) (<-chan ResponseChunk, <-chan error) {
	ch := make(chan ResponseChunk, 64)
	errs := make(chan error, 1)

	go func() {
		defer close(ch)
		defer close(errs)

		req.Stream = true
		body, err := json.Marshal(req)
		if err != nil {
			errs <- fmt.Errorf("marshal request: %w", err)
			return
		}

		httpReq, err := http.NewRequestWithContext(
			ctx, http.MethodPost,
			c.baseURL+"/v1/responses",
			bytes.NewReader(body),
		)
		if err != nil {
			errs <- fmt.Errorf("create request: %w", err)
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			errs <- fmt.Errorf("do request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			errs <- fmt.Errorf("API error %d: %s", resp.StatusCode, string(b))
			return
		}

		c.parseSSE(ctx, resp.Body, ch, errs)
	}()

	return ch, errs
}

// Send sends a non-streaming request and returns the complete response.
func (c *ResponsesClient) Send(
	ctx context.Context,
	req CreateResponseRequest,
) (*ResponseObject, error) {
	req.Stream = false
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx, http.MethodPost,
		c.baseURL+"/v1/responses",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(b))
	}

	var result ResponseObject
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &result, nil
}

func (c *ResponsesClient) parseSSE(
	ctx context.Context,
	body io.Reader,
	ch chan<- ResponseChunk,
	errs chan<- error,
) {
	scanner := bufio.NewScanner(body)
	// Increase buffer for large SSE events
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := scanner.Text()

		// SSE format: "data: {...}" or "data: [DONE]"
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			return
		}

		var event sseEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue // skip malformed events
		}

		chunk := c.eventToChunk(event)
		if chunk != nil {
			ch <- *chunk
		}
	}

	if err := scanner.Err(); err != nil {
		errs <- fmt.Errorf("SSE scan: %w", err)
	}
}

type sseEvent struct {
	Type     string          `json:"type"`
	Delta    string          `json:"delta,omitempty"`
	Item     json.RawMessage `json:"item,omitempty"`
	Response json.RawMessage `json:"response,omitempty"`
}

func (c *ResponsesClient) eventToChunk(event sseEvent) *ResponseChunk {
	switch event.Type {
	case "response.output_text.delta":
		return &ResponseChunk{
			Type:  event.Type,
			Delta: event.Delta,
		}

	case "response.output_item.added":
		var item OutputItem
		if err := json.Unmarshal(event.Item, &item); err == nil {
			return &ResponseChunk{
				Type: event.Type,
				Item: &item,
			}
		}

	case "response.function_call_arguments.delta":
		return &ResponseChunk{
			Type:  event.Type,
			Delta: event.Delta,
		}

	case "response.completed":
		var resp ResponseObject
		if err := json.Unmarshal(event.Response, &resp); err == nil {
			return &ResponseChunk{
				Type:     event.Type,
				Response: &resp,
			}
		}
	}

	return nil
}
