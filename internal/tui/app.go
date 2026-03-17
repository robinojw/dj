package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/state"
)

const (
	FocusCanvas = iota
	FocusTree
)

var titleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("39")).
	MarginBottom(1)

type AppModel struct {
	store  *state.ThreadStore
	canvas CanvasModel
	tree   TreeModel
	focus  int
	width  int
	height int
}

func NewAppModel(store *state.ThreadStore) AppModel {
	return AppModel{
		store:  store,
		canvas: NewCanvasModel(store),
		tree:   NewTreeModel(store),
	}
}

func (app AppModel) Focus() int {
	return app.focus
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
	case tea.KeyRunes:
		return app.handleRune(msg)
	default:
		return app.handleArrow(msg)
	}
}

func (app AppModel) handleRune(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "t" {
		app.toggleFocus()
	}
	return app, nil
}

func (app *AppModel) toggleFocus() {
	if app.focus == FocusCanvas {
		app.focus = FocusTree
		return
	}
	app.focus = FocusCanvas
}

func (app AppModel) handleArrow(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if app.focus == FocusTree {
		app.handleTreeArrow(msg)
		return app, nil
	}
	app.handleCanvasArrow(msg)
	return app, nil
}

func (app *AppModel) handleTreeArrow(msg tea.KeyMsg) {
	switch msg.Type {
	case tea.KeyDown:
		app.tree.MoveDown()
	case tea.KeyUp:
		app.tree.MoveUp()
	}
}

func (app *AppModel) handleCanvasArrow(msg tea.KeyMsg) {
	switch msg.Type {
	case tea.KeyRight, tea.KeyTab:
		app.canvas.MoveRight()
	case tea.KeyLeft, tea.KeyShiftTab:
		app.canvas.MoveLeft()
	case tea.KeyDown:
		app.canvas.MoveDown()
	case tea.KeyUp:
		app.canvas.MoveUp()
	}
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

	if app.focus == FocusTree {
		treeView := app.tree.View()
		body := lipgloss.JoinHorizontal(lipgloss.Top, treeView+"  ", canvas)
		return title + "\n" + body + "\n"
	}

	return title + "\n" + canvas + "\n"
}
