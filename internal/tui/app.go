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
	store       *state.ThreadStore
	client      *appserver.Client
	program     *tea.Program
	statusBar   *StatusBar
	canvas      CanvasModel
	tree        TreeModel
	session     *SessionModel
	prefix      *PrefixHandler
	menu        MenuModel
	help        HelpModel
	menuVisible bool
	helpVisible bool
	focus       int
	width       int
	height      int
}

func NewAppModel(store *state.ThreadStore, opts ...AppOption) AppModel {
	app := AppModel{
		store:     store,
		statusBar: NewStatusBar(),
		canvas:    NewCanvasModel(store),
		tree:      NewTreeModel(store),
		prefix:    NewPrefixHandler(),
		help:      NewHelpModel(),
	}
	for _, opt := range opts {
		opt(&app)
	}
	return app
}

// AppOption configures optional AppModel fields.
type AppOption func(*AppModel)

// WithClient sets the app-server client.
func WithClient(client *appserver.Client) AppOption {
	return func(a *AppModel) {
		a.client = client
	}
}

// SetProgram stores the tea.Program for sending async messages.
func (app *AppModel) SetProgram(p *tea.Program) {
	app.program = p
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
		ctx := context.Background()
		if err := app.client.Start(ctx); err != nil {
			return AppServerErrorMsg{Err: err}
		}

		router := appserver.NewNotificationRouter()
		app.client.Router = router
		if app.program != nil {
			WireEventBridge(router, app.program)
		}

		go app.client.ReadLoop(app.client.Dispatch)

		caps, err := app.client.Initialize(ctx)
		if err != nil {
			return AppServerErrorMsg{Err: err}
		}

		return AppServerConnectedMsg{
			ServerName:    caps.ServerInfo.Name,
			ServerVersion: caps.ServerInfo.Version,
		}
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
	case AppServerConnectedMsg:
		app.statusBar.SetConnected(true)
		return app, nil
	case AppServerErrorMsg:
		app.statusBar.SetError(msg.Error())
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

func (app AppModel) handleSessionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return app, tea.Quit
	case tea.KeyEsc:
		app.closeSession()
		return app, nil
	case tea.KeyUp, tea.KeyDown, tea.KeyPgUp, tea.KeyPgDown:
		app.scrollSession(msg)
		return app, nil
	}
	return app, nil
}

func (app AppModel) openSession() (tea.Model, tea.Cmd) {
	threadID := app.canvas.SelectedThreadID()
	if threadID == "" {
		return app, nil
	}

	thread, exists := app.store.Get(threadID)
	if !exists {
		return app, nil
	}

	session := NewSessionModel(thread)
	session.SetSize(app.width, app.height)
	app.session = &session
	app.focus = FocusSession
	return app, nil
}

func (app *AppModel) closeSession() {
	app.session = nil
	app.focus = FocusCanvas
}

func (app *AppModel) scrollSession(msg tea.KeyMsg) {
	if app.session == nil {
		return
	}

	switch msg.Type {
	case tea.KeyUp:
		app.session.viewport.ScrollUp(1)
	case tea.KeyDown:
		app.session.viewport.ScrollDown(1)
	case tea.KeyPgUp:
		app.session.viewport.HalfPageUp()
	case tea.KeyPgDown:
		app.session.viewport.HalfPageDown()
	}
}

func (app *AppModel) refreshSession() {
	if app.session == nil {
		return
	}
	app.session.Refresh()
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
	app.refreshSession()
	return app, nil
}

func (app AppModel) handleThreadDelta(msg ThreadDeltaMsg) (tea.Model, tea.Cmd) {
	thread, exists := app.store.Get(msg.ThreadID)
	if !exists {
		return app, nil
	}
	thread.AppendDelta(msg.MessageID, msg.Delta)
	app.refreshSession()
	return app, nil
}

func (app AppModel) handleCommandOutput(msg CommandOutputMsg) (tea.Model, tea.Cmd) {
	thread, exists := app.store.Get(msg.ThreadID)
	if !exists {
		return app, nil
	}
	thread.AppendOutput(msg.ExecID, msg.Data)
	app.refreshSession()
	return app, nil
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

	canvas := app.canvas.View()

	if app.focus == FocusTree {
		treeView := app.tree.View()
		body := lipgloss.JoinHorizontal(lipgloss.Top, treeView+"  ", canvas)
		return title + "\n" + body + "\n" + status
	}

	return title + "\n" + canvas + "\n" + status
}
