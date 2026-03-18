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
	events           chan appserver.ProtoEvent
	ptySessions      map[string]*PTYSession
	ptyEvents        chan PTYOutputMsg
	interactiveCmd   string
	interactiveArgs  []string
	header           HeaderBar
	sessionPanel     SessionPanelModel
}

func NewAppModel(store *state.ThreadStore, opts ...AppOption) AppModel {
	app := AppModel{
		store:        store,
		statusBar:    NewStatusBar(),
		canvas:       NewCanvasModel(store),
		tree:         NewTreeModel(store),
		prefix:       NewPrefixHandler(),
		help:         NewHelpModel(),
		events:       make(chan appserver.ProtoEvent, eventChannelSize),
		ptySessions:  make(map[string]*PTYSession),
		ptyEvents:    make(chan PTYOutputMsg, eventChannelSize),
		header:       NewHeaderBar(0),
		sessionPanel: NewSessionPanelModel(),
	}
	for _, opt := range opts {
		opt(&app)
	}
	return app
}

type AppOption func(*AppModel)

func WithClient(client *appserver.Client) AppOption {
	return func(a *AppModel) {
		a.client = client
	}
}

func WithInteractiveCommand(command string, args ...string) AppOption {
	return func(a *AppModel) {
		a.interactiveCmd = command
		a.interactiveArgs = args
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
	case tea.WindowSizeMsg:
		app.width = msg.Width
		app.height = msg.Height
		app.header.SetWidth(msg.Width)
		app.statusBar.SetWidth(msg.Width)
		return app, app.rebalancePTYSizes()
	case protoEventMsg:
		return app.handleProtoEvent(msg.Event)
	case PTYOutputMsg:
		return app.handlePTYOutput(msg)
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
	case AppServerErrorMsg:
		app.statusBar.SetError(msg.Error())
		return app, nil
	case ThreadCreatedMsg:
		app.store.Add(msg.ThreadID, msg.Title)
		app.statusBar.SetThreadCount(len(app.store.All()))
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

	if app.focusPane == FocusPaneSession {
		return app.handleSessionKey(msg)
	}

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
	case " ":
		return app.togglePin()
	}
	return app, nil
}

func (app AppModel) createThread() tea.Cmd {
	if app.client == nil {
		return func() tea.Msg {
			return ThreadCreatedMsg{
				ThreadID: "local",
				Title:    "New Thread",
			}
		}
	}
	return nil
}

func (app AppModel) handleHelpKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	isToggle := msg.Type == tea.KeyRunes && msg.String() == "?"
	isEsc := msg.Type == tea.KeyEsc
	if isToggle || isEsc {
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
