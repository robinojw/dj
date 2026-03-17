package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/appserver"
	"github.com/robinojw/dj/internal/state"
)

const (
	FocusCanvas = iota
	FocusTree
	FocusSession
)

var titleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("39")).
	MarginBottom(1)

type AppModel struct {
	store     *state.ThreadStore
	client    *appserver.Client
	canvas    CanvasModel
	tree      TreeModel
	session   *SessionModel
	prefix    *PrefixHandler
	menu      MenuModel
	help      HelpModel
	statusBar *StatusBar

	connected   bool
	menuVisible bool
	helpVisible bool
	focus       int
	width       int
	height      int
}

func NewAppModel(store *state.ThreadStore, client *appserver.Client) AppModel {
	bar := NewStatusBar()
	if client != nil {
		bar.SetConnecting()
	}

	return AppModel{
		store:     store,
		client:    client,
		canvas:    NewCanvasModel(store),
		tree:      NewTreeModel(store),
		prefix:    NewPrefixHandler(),
		help:      NewHelpModel(),
		statusBar: bar,
	}
}

func (app AppModel) Focus() int {
	return app.focus
}

func (app AppModel) HelpVisible() bool {
	return app.helpVisible
}

func (app AppModel) Init() tea.Cmd {
	if app.client == nil {
		return nil
	}

	return func() tea.Msg {
		if err := app.client.Start(context.Background()); err != nil {
			return AppServerErrorMsg{Err: err}
		}

		go app.client.ReadLoop()

		return nil
	}
}

func (app AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return app.handleKey(msg)
	case tea.WindowSizeMsg:
		app.width = msg.Width
		app.height = msg.Height
		app.statusBar.SetWidth(msg.Width)
		if app.session != nil {
			app.session.SetSize(msg.Width, msg.Height)
		}
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
	case ThreadCreatedMsg:
		app.store.Add(msg.ThreadID, msg.Title)
		app.statusBar.SetThreadCount(len(app.store.All()))
		return app, nil
	case AppServerConnectedMsg:
		app.connected = true
		app.statusBar.SetConnected(true)
		app.store.Add(msg.SessionID, msg.Model)
		app.statusBar.SetThreadCount(len(app.store.All()))
		app.statusBar.SetSelectedThread(msg.Model)
		return app, nil
	case AppServerErrorMsg:
		app.statusBar.SetError(msg.Error())
		return app, nil
	}
	return app, nil
}

func (app AppModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if app.helpVisible {
		return app.handleHelpKey(msg)
	}

	if app.menuVisible {
		return app.handleMenuKey(msg)
	}

	prefixResult := app.prefix.HandleKey(msg)
	switch prefixResult {
	case PrefixWaiting:
		return app, nil
	case PrefixComplete:
		return app.handlePrefixAction()
	case PrefixCancelled:
		return app, nil
	}

	if app.focus == FocusSession {
		return app.handleSessionKey(msg)
	}

	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		return app, tea.Quit
	case tea.KeyEnter:
		return app.openSession()
	case tea.KeyRunes:
		return app.handleRune(msg)
	default:
		return app.handleArrow(msg)
	}
}

func (app AppModel) handleRune(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "t":
		app.toggleFocus()
	case "?":
		app.helpVisible = !app.helpVisible
	case "n":
		if !app.connected {
			app.statusBar.SetError("waiting for codex — is codex CLI installed?")
			return app, nil
		}
		app.statusBar.SetError("single session mode — session auto-created on connect")
		return app, nil
	}
	return app, nil
}


func (app AppModel) handleHelpKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	isToggle := msg.Type == tea.KeyRunes && msg.String() == "?"
	isEsc := msg.Type == tea.KeyEsc
	if isToggle || isEsc {
		app.helpVisible = false
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

func (app AppModel) canvasView() string {
	threads := app.store.All()
	if len(threads) > 0 {
		return app.canvas.View()
	}
	if !app.connected {
		return "Waiting for app-server connection..."
	}
	return "No active threads. Press 'n' to create one."
}

func (app AppModel) View() string {
	title := titleStyle.Render("DJ — Codex TUI Visualizer")
	status := app.statusBar.View()

	if app.helpVisible {
		return title + "\n" + app.help.View() + "\n" + status
	}

	if app.menuVisible {
		return title + "\n" + app.menu.View() + "\n" + status
	}

	if app.focus == FocusSession && app.session != nil {
		return title + "\n" + app.session.View() + "\n" + status
	}

	canvas := app.canvasView()

	if app.focus == FocusTree {
		treeView := app.tree.View()
		body := lipgloss.JoinHorizontal(lipgloss.Top, treeView+"  ", canvas)
		return title + "\n" + body + "\n" + status
	}

	return title + "\n" + canvas + "\n" + status
}
