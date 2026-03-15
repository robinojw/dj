package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

const wsReadBufferSize = 4096
const wsWriteBufferSize = 4096
const wsMaxMessageSize = 10 * 1024 * 1024

// WebSocketClient connects to the OpenAI Responses API over WebSocket.
// It maintains a map of response-ID-to-connection so that multi-turn
// tool-call loops reuse the same connection and benefit from the
// server's in-memory previous-response cache.
type WebSocketClient struct {
	apiKey  string
	baseURL string
	mu      sync.Mutex
	conns   map[string]*websocket.Conn
}

func NewWebSocketClient(apiKey string) *WebSocketClient {
	return &WebSocketClient{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		conns:   make(map[string]*websocket.Conn),
	}
}

func (c *WebSocketClient) WithBaseURL(url string) *WebSocketClient {
	c.baseURL = strings.TrimRight(url, "/")
	return c
}

// Stream sends a response.create event over WebSocket and returns channels
// for streaming response chunks. The connection is reused when the request
// includes a PreviousResponseID from a prior response on the same connection.
func (c *WebSocketClient) Stream(
	ctx context.Context,
	req CreateResponseRequest,
) (<-chan ResponseChunk, <-chan error) {
	chunks := make(chan ResponseChunk, 64)
	errs := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errs)

		conn, dialErr := c.getOrDial(ctx, req.PreviousResponseID)
		if dialErr != nil {
			errs <- dialErr
			return
		}

		event := newWSCreateEvent(req)
		writeErr := conn.WriteJSON(event)
		if writeErr != nil {
			conn.Close()
			errs <- fmt.Errorf("websocket send: %w", writeErr)
			return
		}

		c.readEvents(ctx, conn, chunks, errs)
	}()

	return chunks, errs
}

// Send sends a request over WebSocket and collects the full response.
func (c *WebSocketClient) Send(
	ctx context.Context,
	req CreateResponseRequest,
) (*ResponseObject, error) {
	chunks, errs := c.Stream(ctx, req)

	var response *ResponseObject
	for chunk := range chunks {
		if chunk.Type == "response.completed" && chunk.Response != nil {
			response = chunk.Response
		}
	}

	for err := range errs {
		return nil, err
	}

	if response == nil {
		return nil, fmt.Errorf("websocket: no response received")
	}
	return response, nil
}

// Close tears down all pooled WebSocket connections.
func (c *WebSocketClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for responseID, conn := range c.conns {
		conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		)
		conn.Close()
		delete(c.conns, responseID)
	}
	return nil
}

var _ Client = (*WebSocketClient)(nil)

func (c *WebSocketClient) wsURL() string {
	wsBase := strings.Replace(c.baseURL, "https://", "wss://", 1)
	wsBase = strings.Replace(wsBase, "http://", "ws://", 1)
	return wsBase + responsesPath
}

func (c *WebSocketClient) dial(ctx context.Context) (*websocket.Conn, error) {
	dialer := websocket.Dialer{
		ReadBufferSize:  wsReadBufferSize,
		WriteBufferSize: wsWriteBufferSize,
	}

	header := http.Header{}
	header.Set("Authorization", "Bearer "+c.apiKey)

	conn, _, dialErr := dialer.DialContext(ctx, c.wsURL(), header)
	if dialErr != nil {
		return nil, fmt.Errorf("websocket dial: %w", dialErr)
	}

	conn.SetReadLimit(wsMaxMessageSize)
	return conn, nil
}

// getOrDial returns a pooled connection for the given previousResponseID,
// or dials a new one if none is available.
func (c *WebSocketClient) getOrDial(ctx context.Context, previousResponseID string) (*websocket.Conn, error) {
	if previousResponseID != "" {
		c.mu.Lock()
		conn, found := c.conns[previousResponseID]
		if found {
			delete(c.conns, previousResponseID)
		}
		c.mu.Unlock()
		if found {
			return conn, nil
		}
	}
	return c.dial(ctx)
}

// storeConn saves a connection keyed by its response ID for future reuse.
func (c *WebSocketClient) storeConn(responseID string, conn *websocket.Conn) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conns[responseID] = conn
}

// readEvents reads WebSocket messages and dispatches them as ResponseChunks.
// On successful completion (response.completed), the connection is stored
// for reuse. On error or context cancellation, the connection is closed.
func (c *WebSocketClient) readEvents(
	ctx context.Context,
	conn *websocket.Conn,
	chunks chan<- ResponseChunk,
	errs chan<- error,
) {
	readDone := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			conn.Close()
		case <-readDone:
		}
	}()

	for {
		_, message, readErr := conn.ReadMessage()
		if readErr != nil {
			close(readDone)
			if ctx.Err() == nil {
				errs <- fmt.Errorf("websocket read: %w", readErr)
			}
			return
		}

		var raw sseEvent
		if err := json.Unmarshal(message, &raw); err != nil {
			continue
		}

		if raw.Type == "error" {
			close(readDone)
			c.handleWSError(message, conn, errs)
			return
		}

		chunk := eventToChunk(raw)
		if chunk == nil {
			continue
		}

		chunks <- *chunk

		if chunk.Type == "response.completed" && chunk.Response != nil {
			close(readDone)
			c.storeConn(chunk.Response.ID, conn)
			return
		}
	}
}

func (c *WebSocketClient) handleWSError(
	message []byte,
	conn *websocket.Conn,
	errs chan<- error,
) {
	var wsErr wsErrorEvent
	if err := json.Unmarshal(message, &wsErr); err == nil {
		errs <- fmt.Errorf("API error %d (%s): %s", wsErr.Status, wsErr.Error.Code, wsErr.Error.Message)
	} else {
		errs <- fmt.Errorf("API error: %s", string(message))
	}
	conn.Close()
}
