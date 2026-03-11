package tools

import (
	"context"
	"fmt"
	"sync"
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

// ToolRegistry maps tool names to native Go handlers and annotations.
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
