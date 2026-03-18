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
	sessionID        string
	currentMessageID string
	events           chan appserver.JsonRpcMessage
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
		events:         make(chan appserver.JsonRpcMessage, eventChannelSize),
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
	case jsonRpcEventMsg:
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
	result, cmd, handled := app.handleV2Msg(msg)
	if handled {
		return result, cmd
	}
	return app.handleLegacyMsg(msg)
}

func (app AppModel) handleV2Msg(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case ThreadStartedMsg:
		model, cmd := app.handleThreadStarted(msg)
		return model, cmd, true
	case TurnStartedMsg:
		model, cmd := app.handleTurnStarted(msg)
		return model, cmd, true
	case TurnCompletedMsg:
		model, cmd := app.handleTurnCompleted(msg)
		return model, cmd, true
	case V2AgentDeltaMsg:
		model, cmd := app.handleV2AgentDelta(msg)
		return model, cmd, true
	case CollabSpawnMsg:
		model, cmd := app.handleCollabSpawn(msg)
		return model, cmd, true
	case CollabCloseMsg:
		model, cmd := app.handleCollabClose(msg)
		return model, cmd, true
	case ThreadStatusChangedMsg:
		model, cmd := app.handleThreadStatusChanged(msg)
		return model, cmd, true
	case V2ExecApprovalMsg:
		model, cmd := app.handleV2ExecApproval(msg)
		return model, cmd, true
	case V2FileApprovalMsg:
		model, cmd := app.handleV2FileApproval(msg)
		return model, cmd, true
	}
	return app, nil, false
}

func (app AppModel) handleLegacyMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	case AgentReasoningDeltaMsg:
		return app.handleReasoningDelta()
	}
	return app, nil
}
