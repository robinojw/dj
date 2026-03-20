package tui

import (
	"strings"
	"testing"

	"github.com/robinojw/dj/internal/state"
)

const (
	appViewTestWidth  = 80
	appViewTestHeight = 24
	appViewPrompt     = "Task: "
)

func TestAppViewShowsInputBar(testing *testing.T) {
	store := state.NewThreadStore()
	app := NewAppModel(store)
	app.inputBarVisible = true
	app.inputBar = NewInputBarModel(appViewPrompt)
	app.width = appViewTestWidth
	app.height = appViewTestHeight

	view := app.View()
	if !strings.Contains(view, appViewPrompt) {
		testing.Error("expected input bar prompt in view")
	}
}
