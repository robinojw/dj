package tools

import (
	"context"
	"sync"
)

// ToolHandler is the function signature for tool implementations.
type ToolHandler func(ctx context.Context, args map[string]any) (string, error)

// ToolAnnotations provides metadata about a tool's behavior.
type ToolAnnotations struct {
	ReadOnly      bool   // tool only reads state, never mutates
	Destructive   bool   // tool may delete or overwrite data
	Idempotent    bool   // repeated calls with same args produce same result
	MutatesFiles  bool   // tool writes to filesystem (triggers diff generation)
	FilePathParam string // arg key holding the target file path (e.g. "file_path")
}

// ToolRegistry maps tool names to handlers and annotations.
type ToolRegistry struct {
	mu          sync.RWMutex
	handlers    map[string]ToolHandler
	annotations map[string]ToolAnnotations
}

// NewRegistry creates an empty ToolRegistry.
func NewRegistry() *ToolRegistry {
	return &ToolRegistry{
		handlers:    make(map[string]ToolHandler),
		annotations: make(map[string]ToolAnnotations),
	}
}

// Register adds a tool with its handler and annotations.
func (r *ToolRegistry) Register(name string, handler ToolHandler, ann ToolAnnotations) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[name] = handler
	r.annotations[name] = ann
}

// RegisterAnnotationsOnly stores annotations for a tool without a native handler.
// Use this for MCP tools that should participate in diff generation.
func (r *ToolRegistry) RegisterAnnotationsOnly(name string, ann ToolAnnotations) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.annotations[name] = ann
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

// Annotations returns the annotations for the given tool name.
// Returns a zero-value ToolAnnotations if not found.
func (r *ToolRegistry) Annotations(name string) ToolAnnotations {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.annotations[name]
}

// ToolAnnotationsForClassifier returns annotation flags for use by the permission gate.
// Returns ok=false if the tool has no annotations.
func (r *ToolRegistry) ToolAnnotations(name string) (readOnly, destructive, mutatesFiles bool, ok bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ann, exists := r.annotations[name]
	if !exists {
		return false, false, false, false
	}
	return ann.ReadOnly, ann.Destructive, ann.MutatesFiles, true
}
