package tui

import (
	"strings"
	"testing"

	poolpkg "github.com/robinojw/dj/internal/pool"
	"github.com/robinojw/dj/internal/state"
)

const testSwarmMaxAgents = 10

func TestAppModelSwarmFieldsDefault(testing *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)
	if app.menuIntent != MenuIntentThread {
		testing.Error("expected default menu intent to be thread")
	}
	if app.inputBarVisible {
		testing.Error("expected input bar hidden by default")
	}
	if app.swarmFilter {
		testing.Error("expected swarm filter off by default")
	}
}

func TestNewAppModelPoolSetsSwarmActive(testing *testing.T) {
	store := state.NewThreadStore()
	agentPool := poolpkg.NewAgentPool("echo", []string{}, nil, testSwarmMaxAgents)
	app := NewAppModel(store, WithPool(agentPool))
	view := app.header.View()
	if !strings.Contains(view, "p: persona") {
		testing.Error("expected swarm hints in header when pool is set")
	}
}
