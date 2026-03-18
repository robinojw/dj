package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/appserver"
	"github.com/robinojw/dj/internal/state"
)

const (
	CanvasModeGrid = iota
	CanvasModeTree
)

const eventChannelSize = 64

type AppModel struct {
	store            *state.ThreadStore
	client           *appserver.Client
	statusBar        *StatusBar
	canvas           CanvasModel
	tree             TreeModel
	prefix           *PrefixHandler
	menu             MenuModel
	help             HelpModel
	menuVisible      bool
	helpVisible      bool
	focusPane        FocusPane
	canvasMode       int
	width            int
	height           int
	sessionID        string
	currentMessageID string
	events           chan appserver.ProtoEvent
	ptySessions      map[string]*PTYSession
	ptyEvents        chan PTYOutputMsg
	sessionCounter   *int
	interactiveCmd   string
	interactiveArgs  []string
	header           HeaderBar
	sessionPanel     SessionPanelModel
}

func NewAppModel(store *state.ThreadStore, opts ...AppOption) AppModel {
	app := AppModel{
		store:          store,
		statusBar:      NewStatusBar(),
		canvas:         NewCanvasModel(store),
		tree:           NewTreeModel(store),
		prefix:         NewPrefixHandler(),
		help:           NewHelpModel(),
		events:         make(chan appserver.ProtoEvent, eventChannelSize),
		ptySessions:    make(map[string]*PTYSession),
		ptyEvents:      make(chan PTYOutputMsg, eventChannelSize),
		sessionCounter: new(int),
		header:         NewHeaderBar(0),
		sessionPanel:   NewSessionPanelModel(),
	}
	for _, opt := range opts {
		opt(&app)
	}
	return app
}

type AppOption func(*AppModel)

func WithClient(client *appserver.Client) AppOption {
	return func(app *AppModel) {
		app.client = client
	}
}

func WithInteractiveCommand(command string, args ...string) AppOption {
	return func(app *AppModel) {
		app.interactiveCmd = command
		app.interactiveArgs = args
	}
}

func (app AppModel) FocusPane() FocusPane {
	return app.focusPane
}

func (app AppModel) CanvasMode() int {
	return app.canvasMode
}

func (app AppModel) HelpVisible() bool {
	return app.helpVisible
}

func (app AppModel) Init() tea.Cmd {
	if app.client == nil {
		return nil
	}
	return tea.Batch(
		app.connectClient(),
		app.listenForEvents(),
		app.listenForPTYEvents(),
	)
}

func (app AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return app.handleKey(msg)
	case tea.MouseMsg:
		return app.handleMouse(msg)
	case tea.WindowSizeMsg:
		return app.handleWindowSize(msg)
	case protoEventMsg:
		return app.handleProtoEvent(msg.Event)
	case PTYOutputMsg:
		return app.handlePTYOutput(msg)
	case AppServerErrorMsg:
		app.statusBar.SetError(msg.Error())
		return app, nil
	case ThreadCreatedMsg:
		return app.handleThreadCreated(msg)
	default:
		return app.handleAgentMsg(msg)
	}
}

func (app AppModel) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	app.width = msg.Width
	app.height = msg.Height
	app.header.SetWidth(msg.Width)
	app.statusBar.SetWidth(msg.Width)
	return app, app.rebalancePTYSizes()
}

func (app AppModel) handleThreadCreated(msg ThreadCreatedMsg) (tea.Model, tea.Cmd) {
	app.store.Add(msg.ThreadID, msg.Title)
	app.statusBar.SetThreadCount(len(app.store.All()))
	app.canvas.SetSelected(len(app.store.All()) - 1)
	return app.openSession()
}

func (app AppModel) handleAgentMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SessionConfiguredMsg:
		return app.handleSessionConfigured(msg)
	case TaskStartedMsg:
		return app.handleTaskStarted()
	case AgentDeltaMsg:
		return app.handleAgentDelta(msg)
	case AgentMessageCompletedMsg:
		return app.handleAgentMessageCompleted()
	case TaskCompleteMsg:
		return app.handleTaskComplete()
	case ExecApprovalRequestMsg:
		return app.handleExecApproval(msg)
	case PatchApprovalRequestMsg:
		return app.handlePatchApproval(msg)
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

	if result, model, cmd := app.handlePrefix(msg); result {
		return model, cmd
	}

	if app.focusPane == FocusPaneSession {
		return app.handleSessionKey(msg)
	}

	return app.handleCanvasKey(msg)
}

func (app AppModel) handlePrefix(msg tea.KeyMsg) (bool, tea.Model, tea.Cmd) {
	prefixResult := app.prefix.HandleKey(msg)
	switch prefixResult {
	case PrefixWaiting:
		return true, app, nil
	case PrefixComplete:
		model, cmd := app.handlePrefixAction()
		return true, model, cmd
	case PrefixCancelled:
		return true, app, nil
	}
	return false, app, nil
}

func (app AppModel) handleCanvasKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		return app, tea.Quit
	case tea.KeyEnter:
		return app.openSession()
	case tea.KeyTab:
		return app.switchToSessionPanel()
	case tea.KeyRunes:
		return app.handleRune(msg)
	default:
		return app.handleArrow(msg)
	}
}

func (app AppModel) handleRune(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "t":
		app.toggleCanvasMode()
	case "n":
		return app, app.createThread()
	case "?":
		app.helpVisible = !app.helpVisible
	case " ", "s":
		return app.togglePin()
	}
	return app, nil
}

func (app AppModel) createThread() tea.Cmd {
	*app.sessionCounter++
	counter := *app.sessionCounter
	return func() tea.Msg {
		return ThreadCreatedMsg{
			ThreadID: fmt.Sprintf("session-%d", counter),
			Title:    fmt.Sprintf("Session %d", counter),
		}
	}
}

func (app AppModel) handleHelpKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	isToggle := msg.Type == tea.KeyRunes && msg.String() == "?"
	isEsc := msg.Type == tea.KeyEsc
	shouldDismissHelp := isToggle || isEsc
	if shouldDismissHelp {
		app.helpVisible = false
	}
	return app, nil
}

func (app *AppModel) toggleCanvasMode() {
	if app.canvasMode == CanvasModeGrid {
		app.canvasMode = CanvasModeTree
		return
	}
	app.canvasMode = CanvasModeGrid
}

func (app AppModel) switchToSessionPanel() (tea.Model, tea.Cmd) {
	hasPinned := len(app.sessionPanel.PinnedSessions()) > 0
	if !hasPinned {
		return app, nil
	}
	app.focusPane = FocusPaneSession
	return app, nil
}

func (app AppModel) handleArrow(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if app.canvasMode == CanvasModeTree {
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
	case tea.KeyRight:
		app.canvas.MoveRight()
	case tea.KeyLeft, tea.KeyShiftTab:
		app.canvas.MoveLeft()
	case tea.KeyDown:
		app.canvas.MoveDown()
	case tea.KeyUp:
		app.canvas.MoveUp()
	}
}
