package mcp

import (
	"context"
	"fmt"
	"sync"
)

// Registry manages MCP server lifecycle and tool caching.
type Registry struct {
	configs    []MCPServerConfig
	clients    map[string]*Client
	toolsCache map[string][]MCPTool
	mu         sync.RWMutex
}

func NewRegistry(configs []MCPServerConfig) *Registry {
	return &Registry{
		configs:    configs,
		clients:    make(map[string]*Client),
		toolsCache: make(map[string][]MCPTool),
	}
}

// Start connects to all auto-start MCP servers.
func (r *Registry) Start(ctx context.Context) error {
	for _, cfg := range r.configs {
		if !cfg.AutoStart {
			continue
		}
		if err := r.Connect(ctx, cfg.Name); err != nil {
			// Log but don't fail — partial MCP is fine
			fmt.Printf("Warning: MCP server %s failed to start: %v\n", cfg.Name, err)
		}
	}
	return nil
}

// Connect starts and initializes a specific MCP server.
func (r *Registry) Connect(ctx context.Context, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cfg := r.findConfig(name)
	if cfg == nil {
		return fmt.Errorf("MCP server %q not configured", name)
	}

	client := NewClient(*cfg)
	if err := client.Connect(ctx); err != nil {
		return err
	}

	tools, err := client.ListTools(ctx)
	if err != nil {
		client.Close()
		return fmt.Errorf("list tools for %s: %w", name, err)
	}

	r.clients[name] = client
	r.toolsCache[name] = tools
	return nil
}

// Disconnect stops a specific MCP server.
func (r *Registry) Disconnect(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	client, ok := r.clients[name]
	if !ok {
		return nil
	}

	err := client.Close()
	delete(r.clients, name)
	delete(r.toolsCache, name)
	return err
}

// StopAll shuts down all connected MCP servers.
func (r *Registry) StopAll() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for name, client := range r.clients {
		client.Close()
		delete(r.clients, name)
		delete(r.toolsCache, name)
	}
}

// ActiveClients returns all currently connected clients.
func (r *Registry) ActiveClients() []*Client {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var clients []*Client
	for _, c := range r.clients {
		clients = append(clients, c)
	}
	return clients
}

// ActiveNames returns names of all active MCP servers.
func (r *Registry) ActiveNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var names []string
	for name := range r.clients {
		names = append(names, name)
	}
	return names
}

// AllConfigs returns all configured servers with their connection status.
func (r *Registry) AllConfigs() []ServerStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var statuses []ServerStatus
	for _, cfg := range r.configs {
		status := ServerStatus{
			Config: cfg,
			Active: r.clients[cfg.Name] != nil,
		}
		if tools, ok := r.toolsCache[cfg.Name]; ok {
			status.ToolCount = len(tools)
		}
		statuses = append(statuses, status)
	}
	return statuses
}

// ToolsForServer returns cached tools for a specific server.
func (r *Registry) ToolsForServer(name string) []MCPTool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.toolsCache[name]
}

// ServerStatus combines config with runtime state.
type ServerStatus struct {
	Config    MCPServerConfig
	Active    bool
	ToolCount int
}

func (r *Registry) findConfig(name string) *MCPServerConfig {
	for i, cfg := range r.configs {
		if cfg.Name == name {
			return &r.configs[i]
		}
	}
	return nil
}
