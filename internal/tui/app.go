package tui

import (
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
	events           chan appserver.JSONRPCMessage
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
		events:         make(chan appserver.JSONRPCMessage, eventChannelSize),
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
	case jsonRPCEventMsg:
		return app.handleProtoEvent(msg.Message)
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
	case ThreadStartedMsg:
		return app.handleThreadStarted(msg)
	case TurnStartedMsg:
		return app.handleTurnStarted(msg)
	case TurnCompletedMsg:
		return app.handleTurnCompleted(msg)
	case V2AgentDeltaMsg:
		return app.handleV2AgentDelta(msg)
	case CollabSpawnMsg:
		return app.handleCollabSpawn(msg)
	case CollabCloseMsg:
		return app.handleCollabClose(msg)
	case ThreadStatusChangedMsg:
		return app.handleThreadStatusChanged(msg)
	case V2ExecApprovalMsg:
		return app.handleV2ExecApproval(msg)
	case V2FileApprovalMsg:
		return app.handleV2FileApproval(msg)
	}
	return app, nil
}
