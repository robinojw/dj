package tui

import (
	"testing"

	poolpkg "github.com/robinojw/dj/internal/pool"
	"github.com/robinojw/dj/internal/state"
)

const testPoolMaxAgents = 10

func TestListenForPoolEvents(testing *testing.T) {
	store := state.NewThreadStore()
	agentPool := poolpkg.NewAgentPool("echo", []string{}, nil, testPoolMaxAgents)
	app := NewAppModel(store, WithPool(agentPool))

	cmd := app.listenForPoolEvents()
	if cmd == nil {
		testing.Error("expected non-nil command")
	}
}

func TestListenForPoolEventsNilPool(testing *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)

	cmd := app.listenForPoolEvents()
	if cmd != nil {
		testing.Error("expected nil command when pool is nil")
	}
}
