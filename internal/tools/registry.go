package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/robinojw/dj/internal/api"
)

// ToolHandler is the signature for native tool implementations.
type ToolHandler func(ctx context.Context, args map[string]any) (string, error)

// ToolAnnotations provides metadata about a tool's behavior.
type ToolAnnotations struct {
	ReadOnly      bool   // tool only reads state, never mutates
	Destructive   bool   // tool may delete or overwrite data
	Idempotent    bool   // repeated calls with same args produce same result
	MutatesFiles  bool   // tool writes to filesystem (triggers diff generation)
	FilePathParam string // arg key holding the target file path (e.g. "file_path")
}

// ToolSchema holds the API-facing description and parameter JSON schema for a tool.
type ToolSchema struct {
	Description string
	Parameters  json.RawMessage
}

// ToolRegistry maps tool names to native Go handlers and annotations.
type ToolRegistry struct {
	mu          sync.RWMutex
	handlers    map[string]ToolHandler
	annotations map[string]ToolAnnotations
	schemas     map[string]ToolSchema
}

// NewRegistry creates an empty ToolRegistry.
func NewRegistry() *ToolRegistry {
	return &ToolRegistry{
		handlers:    make(map[string]ToolHandler),
		annotations: make(map[string]ToolAnnotations),
		schemas:     make(map[string]ToolSchema),
	}
}

// Register adds a tool handler with its annotations.
func (r *ToolRegistry) Register(name string, h ToolHandler, ann ToolAnnotations) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[name] = h
	r.annotations[name] = ann
}

// RegisterAnnotationsOnly stores annotations for a tool without a native handler.
// Use this for MCP tools that should participate in diff generation.
func (r *ToolRegistry) RegisterAnnotationsOnly(name string, ann ToolAnnotations) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.annotations[name] = ann
}

// Dispatch calls the named tool's handler. Returns an error if the tool is not registered.
func (r *ToolRegistry) Dispatch(ctx context.Context, name string, args map[string]any) (string, error) {
	r.mu.RLock()
	h, ok := r.handlers[name]
	r.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", name)
	}
	return h(ctx, args)
}

// Has returns true if a handler is registered for the given tool name.
func (r *ToolRegistry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.handlers[name]
	return ok
}

// HasAnnotations returns true if annotations are registered for the given tool name,
// regardless of whether a handler exists.
func (r *ToolRegistry) HasAnnotations(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.annotations[name]
	return ok
}

// Annotations returns the annotations for a tool, or empty annotations if not found.
func (r *ToolRegistry) Annotations(name string) ToolAnnotations {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.annotations[name]
}

// IsDestructive returns true if the named tool is annotated as destructive.
func (r *ToolRegistry) IsDestructive(name string) bool {
	return r.Annotations(name).Destructive
}

// Names returns all registered tool names.
func (r *ToolRegistry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.handlers))
	for name := range r.handlers {
		names = append(names, name)
	}
	return names
}

// ToolAnnotations returns annotation flags for use by the permission gate.
// Returns ok=false if the tool has no annotations.
// This method satisfies the modes.ToolClassifier interface.
func (r *ToolRegistry) ToolAnnotations(name string) (readOnly, destructive, mutatesFiles bool, ok bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ann, exists := r.annotations[name]
	if !exists {
		return false, false, false, false
	}
	return ann.ReadOnly, ann.Destructive, ann.MutatesFiles, true
}

// RegisterSchema stores the API-facing schema for a tool.
func (r *ToolRegistry) RegisterSchema(name string, schema ToolSchema) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.schemas[name] = schema
}

// ToolDefinitions returns API tool definitions, optionally filtered by an allow list.
// If allowedTools is nil, all registered schemas are returned.
func (r *ToolRegistry) ToolDefinitions(allowedTools []string) []api.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	allowed := make(map[string]bool, len(allowedTools))
	for _, name := range allowedTools {
		allowed[name] = true
	}
	filterActive := len(allowedTools) > 0

	tools := make([]api.Tool, 0, len(r.schemas))
	for name, schema := range r.schemas {
		if filterActive && !allowed[name] {
			continue
		}
		tools = append(tools, api.Tool{
			Type:        "function",
			Name:        name,
			Description: schema.Description,
			Parameters:  schema.Parameters,
		})
	}
	return tools
}
