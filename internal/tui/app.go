package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/state"
)

var titleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("39")).
	MarginBottom(1)

type AppModel struct {
	store  *state.ThreadStore
	canvas CanvasModel
	width  int
	height int
}

func NewAppModel(store *state.ThreadStore) AppModel {
	return AppModel{
		store:  store,
		canvas: NewCanvasModel(store),
	}
}

func (app AppModel) Init() tea.Cmd {
	return nil
}

func (app AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return app.handleKey(msg)
	case tea.WindowSizeMsg:
		app.width = msg.Width
		app.height = msg.Height
		return app, nil
	case ThreadStatusMsg:
		app.store.UpdateStatus(msg.ThreadID, msg.Status, msg.Title)
		return app, nil
	case ThreadMessageMsg:
		return app.handleThreadMessage(msg)
	case ThreadDeltaMsg:
		return app.handleThreadDelta(msg)
	case CommandOutputMsg:
		return app.handleCommandOutput(msg)
	}
	return app, nil
}

func (app AppModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		return app, tea.Quit
	case tea.KeyRight, tea.KeyTab:
		app.canvas.MoveRight()
	case tea.KeyLeft, tea.KeyShiftTab:
		app.canvas.MoveLeft()
	case tea.KeyDown:
		app.canvas.MoveDown()
	case tea.KeyUp:
		app.canvas.MoveUp()
	}
	return app, nil
}

func (app AppModel) handleThreadMessage(msg ThreadMessageMsg) (tea.Model, tea.Cmd) {
	thread, exists := app.store.Get(msg.ThreadID)
	if !exists {
		return app, nil
	}
	thread.AppendMessage(state.ChatMessage{
		ID:      msg.MessageID,
		Role:    msg.Role,
		Content: msg.Content,
	})
	return app, nil
}

func (app AppModel) handleThreadDelta(msg ThreadDeltaMsg) (tea.Model, tea.Cmd) {
	thread, exists := app.store.Get(msg.ThreadID)
	if !exists {
		return app, nil
	}
	thread.AppendDelta(msg.MessageID, msg.Delta)
	return app, nil
}

func (app AppModel) handleCommandOutput(msg CommandOutputMsg) (tea.Model, tea.Cmd) {
	thread, exists := app.store.Get(msg.ThreadID)
	if !exists {
		return app, nil
	}
	thread.AppendOutput(msg.ExecID, msg.Data)
	return app, nil
}

func (app AppModel) View() string {
	title := titleStyle.Render("DJ — Codex TUI Visualizer")
	canvas := app.canvas.View()
	return title + "\n" + canvas + "\n"
}
