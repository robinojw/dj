package tui

import (
	"context"
	"fmt"

	tui "github.com/grindlemire/go-tui"
	"github.com/robinojw/dj/config"
	"github.com/robinojw/dj/internal/agents"
	"github.com/robinojw/dj/internal/api"
	"github.com/robinojw/dj/internal/checkpoint"
	"github.com/robinojw/dj/internal/hooks"
	"github.com/robinojw/dj/internal/memory"
	"github.com/robinojw/dj/internal/modes"
	"github.com/robinojw/dj/internal/skills"
	"github.com/robinojw/dj/internal/tools"
	"github.com/robinojw/dj/internal/tui/theme"
)

type ScreenID int

const (
	ScreenIDChat ScreenID = iota
	ScreenIDTeam
	ScreenIDEnhance
	ScreenIDMCP
	ScreenIDSkills
	ScreenIDCheatSheet
	ScreenIDDiffPager
)

type rootApp struct {
	app            *tui.App
	screen         *tui.State[ScreenID]
	screenStack    []ScreenID
	chatView       *chat
	teamView       *teamScreen
	enhanceView    *enhanceScreen
	mcpView        *mcpManager
	skillsView     *skillBrowser
	cheatSheetView *cheatSheet
	diffPagerView  *diffPager
	topBarView     *topBar
	permModal      *permissionModal
	turboModal     *turboModal
	debugOverlay   *debugOverlayComponent
	debugMode      *tui.State[bool]

	// Shared state
	mode         *tui.State[modes.ExecutionMode]
	model        *tui.State[string]
	cost         *tui.State[float64]
	inputTokens  *tui.State[int]
	outputTokens *tui.State[int]
	activeMCPs   *tui.State[[]string]

	// Multi-agent orchestrator — set when a multi-agent task is dispatched
	orchestrator  *agents.Orchestrator
	activeAgentID *tui.State[string]

	// Permission request channel — workers send requests here
	permRequestCh chan modes.PermissionRequest

	// Non-TUI dependencies
	t              *theme.Theme
	client         api.Client
	tracker        *api.Tracker
	modeVal        modes.ExecutionMode
	gate           *modes.Gate
	turboConfirmed bool
	checkpoints    *checkpoint.Manager
	toolRegistry   *tools.ToolRegistry
	skillsRegistry *skills.Registry
	memoryMgr      *memory.Manager
	hooks          *hooks.Runner
	width          int
}

func NewRootApp(
	t *theme.Theme,
	client api.Client,
	tracker *api.Tracker,
	modelName string,
	cfg config.Config,
	toolRegistry *tools.ToolRegistry,
	skillsRegistry *skills.Registry,
	memoryMgr *memory.Manager,
	hookRunner *hooks.Runner,
) *rootApp {
	gate := modes.NewGateWithRegistry(
		modes.ModeConfirm,
		cfg.Execution.Allow.Tools,
		cfg.Execution.Deny.Tools,
		toolRegistry,
	)

	modeState := tui.NewState(modes.ModeConfirm)
	modelState := tui.NewState(modelName)
	costState := tui.NewState(0.0)
	inputTokensState := tui.NewState(0)
	outputTokensState := tui.NewState(0)
	activeMCPsState := tui.NewState([]string{})

	a := &rootApp{
		screen:        tui.NewState(ScreenIDChat),
		activeAgentID: tui.NewState(""),
		mode:          modeState,
		model:         modelState,
		cost:          costState,
		inputTokens:   inputTokensState,
		outputTokens:  outputTokensState,
		activeMCPs:    activeMCPsState,
		debugMode:     tui.NewState(false),
		permRequestCh: make(chan modes.PermissionRequest, 10),
		t:             t,
		client:        client,
		tracker:       tracker,
		modeVal:       modes.ModeConfirm,
		gate:          gate,
		checkpoints:    checkpoint.NewManager(20),
		toolRegistry:   toolRegistry,
		skillsRegistry: skillsRegistry,
		memoryMgr:      memoryMgr,
		hooks:          hookRunner,
		width:          80,
	}

	branchState := tui.NewState(resolveGitBranch())
	cwdState := tui.NewState(resolveWorkingDir())
	titleState := tui.NewState("New Session")
	a.topBarView = NewTopBar(t, branchState, cwdState, titleState)

	// Create child components with callbacks
	a.chatView = NewChat(t, 80, modeState, modelState, costState, inputTokensState, outputTokensState, activeMCPsState,
		a.handleSubmit, a.openDiffPager)
	a.cheatSheetView = NewCheatSheet(t, a.popScreenFn)
	a.teamView = NewTeamScreen(t, a.popScreenFn)
	a.enhanceView = NewEnhanceScreen(t, a.popScreenFn, nil)
	a.mcpView = NewMCPManager(t, a.popScreenFn, nil)
	a.skillsView = NewSkillBrowser(t, a.popScreenFn)
	a.diffPagerView = NewDiffPager(t, nil, a.popScreenFn)
	a.permModal = NewPermissionModal(t)
	a.turboModal = NewTurboModal(t, a.handleTurboResult)
	a.debugOverlay = NewDebugOverlay(t)

	return a
}

// PermRequestCh returns the channel for sending permission requests to the TUI.
func (a *rootApp) PermRequestCh() chan<- modes.PermissionRequest {
	return a.permRequestCh
}

// Watchers registers channel watchers for the root app.
func (a *rootApp) Watchers() []tui.Watcher {
	watchers := []tui.Watcher{
		tui.NewChannelWatcher(a.permRequestCh, a.onPermissionRequest),
	}
	if a.orchestrator != nil {
		watchers = append(watchers,
			tui.NewChannelWatcher(a.orchestrator.UpdatesCh, a.onWorkerUpdate),
		)
	}
	return watchers
}

func (a *rootApp) onPermissionRequest(req modes.PermissionRequest) {
	a.permModal.Show(&req)
}

// onWorkerUpdate processes a worker update from the orchestrator's UpdatesCh.
func (a *rootApp) onWorkerUpdate(update agents.WorkerUpdate) {
	switch update.Type {
	case agents.UpdateDelta:
		a.teamView.AppendAgentDelta(update.WorkerID, update.Content)

	case agents.UpdateToolCall:
		a.teamView.AppendToolCall(update.WorkerID, update.Content)
		if a.activeAgentID.Get() == update.WorkerID {
			a.chatView.AppendToolCallBlock(update.WorkerID, update.Content)
		}

	case agents.UpdateToolResult:
		a.teamView.AppendToolResult(update.WorkerID, update.Content)
		if a.activeAgentID.Get() == update.WorkerID {
			a.chatView.AppendToolResultBlock(update.WorkerID, update.Content)
		}

	case agents.UpdateDiffResult:
		if update.DiffInfo != nil {
			a.chatView.AppendDiffBlock(update.DiffInfo)
		}

	case agents.UpdateCompleted:
		a.teamView.SetAgentStatus(update.WorkerID, "completed")
		a.inputTokens.Update(func(v int) int { return v + update.Usage.InputTokens })
		a.outputTokens.Update(func(v int) int { return v + update.Usage.OutputTokens })

	case agents.UpdateError:
		a.teamView.SetAgentStatus(update.WorkerID, "error")

	case agents.UpdateSkipped:
		a.teamView.SetAgentStatus(update.WorkerID, "skipped")

	case agents.UpdateHookResult:
		if update.HookResult != nil && a.debugOverlay.IsVisible() {
			a.debugOverlay.AddInfo(fmt.Sprintf("Hook %s: exit=%d stdout=%q",
				update.HookResult.Event, update.HookResult.ExitCode, update.HookResult.Stdout))
		}
	}
}

func (a *rootApp) handleSubmit(text string, mentionCtx string) {
	fullText := text
	if mentionCtx != "" {
		fullText = fullText + "\n\n" + mentionCtx
	}

	router := agents.NewTaskRouter(a.client, a.model.Get(), 0)
	analysis, err := router.Analyze(context.Background(), fullText)

	isMultiAgent := err == nil && router.ShouldSpawnTeam(analysis)

	if isMultiAgent {
		a.orchestrator = agents.NewOrchestrator(
			a.client, a.skillsRegistry, a.tracker, a.model.Get(),
		)
		a.orchestrator.Mode = a.modeVal
		a.orchestrator.Gate = a.gate
		a.orchestrator.Registry = a.toolRegistry
		a.orchestrator.PermReqCh = a.permRequestCh
		a.orchestrator.Hooks = a.hooks
		a.orchestrator.Memory = a.memoryMgr

		statuses := make([]AgentStatus, 0, len(analysis.Subtasks))
		for _, task := range analysis.Subtasks {
			parentID := ""
			if len(task.DependsOn) > 0 {
				parentID = task.DependsOn[0]
			}
			statuses = append(statuses, AgentStatus{
				ID:       task.ID,
				Name:     task.Description,
				Status:   "pending",
				ParentID: parentID,
			})
		}
		a.teamView.SetAgents(statuses)
		a.pushScreen(ScreenIDTeam)

		go a.orchestrator.Dispatch(context.Background(), analysis.Subtasks)
	} else {
		modeCfg := modes.Modes[a.modeVal]

		var toolDefs []api.Tool
		if a.toolRegistry != nil {
			toolDefs = a.toolRegistry.ToolDefinitions(modeCfg.AllowedTools)
		}

		instructions := modeCfg.SystemPrompt
		if mentionCtx != "" {
			instructions = instructions + "\n\n" + mentionCtx
		}

		req := api.CreateResponseRequest{
			Model:        a.model.Get(),
			Input:        api.MakeStringInput(text),
			Instructions: instructions,
			Tools:        toolDefs,
			Reasoning:    &api.Reasoning{Effort: modeCfg.ReasoningEffort},
			Stream:       true,
		}

		if a.debugOverlay.IsVisible() {
			a.debugOverlay.AddInfo(fmt.Sprintf("Starting stream with model: %s", a.model.Get()))
		}

		ctx := context.Background()
		chunks, errs := a.client.Stream(ctx, req)
		a.chatView.StartStream(chunks, errs)
	}
}

func (a *rootApp) openDiffPager(diffs []storedDiff) {
	a.diffPagerView = NewDiffPager(a.t, diffs, a.popScreenFn)
	a.pushScreen(ScreenIDDiffPager)
}

func (a *rootApp) pushScreen(s ScreenID) {
	a.screenStack = append(a.screenStack, a.screen.Get())
	a.screen.Set(s)
	if a.app != nil {
		a.app.EnterAlternateScreen()
	}
}

func (a *rootApp) popScreenFn() {
	if len(a.screenStack) == 0 {
		a.screen.Set(ScreenIDChat)
	} else {
		a.screen.Set(a.screenStack[len(a.screenStack)-1])
		a.screenStack = a.screenStack[:len(a.screenStack)-1]
	}
	if a.app != nil {
		a.app.ExitAlternateScreen()
	}
}

func (a *rootApp) handleTurboResult(confirmed bool) {
	if confirmed {
		a.turboConfirmed = true
		a.modeVal = modes.ModeTurbo
		a.gate.SetMode(modes.ModeTurbo)
		a.mode.Set(modes.ModeTurbo)
	}
}

func (a *rootApp) cycleModel() {
	models := api.CycleModels
	current := -1
	for i, m := range models {
		if m == a.model.Get() {
			current = i
			break
		}
	}
	next := models[(current+1)%len(models)]
	a.model.Set(next)
	if a.tracker != nil {
		a.tracker.SetModel(next)
	}

	if a.debugOverlay.IsVisible() {
		a.debugOverlay.AddInfo("Model switched to " + next)
	}
}

func (a *rootApp) cycleMode() {
	newMode := (a.modeVal + 1) % 3
	if newMode == modes.ModeTurbo && !a.turboConfirmed {
		a.turboModal.Show()
		return
	}
	a.modeVal = newMode
	a.gate.SetMode(newMode)
	a.mode.Set(newMode)
}

func (a *rootApp) KeyMap() tui.KeyMap {
	// Turbo modal intercepts all keys when visible
	if a.turboModal.IsVisible() {
		return a.turboModal.KeyMap()
	}
	// Permission modal intercepts all keys when visible
	if a.permModal.Visible() {
		return a.permModal.KeyMap()
	}

	km := tui.KeyMap{
		tui.OnKey(tui.KeyCtrlC, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnKey(tui.KeyCtrlQ, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnKey(tui.KeyCtrlD, func(ke tui.KeyEvent) {
			a.debugOverlay.Toggle()
			a.debugMode.Update(func(v bool) bool { return !v })
		}),
	}

	if a.screen.Get() == ScreenIDChat {
		km = append(km,
			tui.OnKey(tui.KeyCtrlT, func(ke tui.KeyEvent) { a.pushScreen(ScreenIDTeam) }),
			tui.OnKey(tui.KeyCtrlM, func(ke tui.KeyEvent) { a.pushScreen(ScreenIDMCP) }),
			tui.OnKey(tui.KeyCtrlK, func(ke tui.KeyEvent) { a.pushScreen(ScreenIDSkills) }),
			tui.OnKey(tui.KeyCtrlE, func(ke tui.KeyEvent) { a.pushScreen(ScreenIDEnhance) }),
			tui.OnKey(tui.KeyCtrlH, func(ke tui.KeyEvent) { a.pushScreen(ScreenIDCheatSheet) }),
			tui.OnKeyStop(tui.KeyTab, func(ke tui.KeyEvent) { a.cycleMode() }),
			tui.OnKey(tui.KeyCtrlN, func(ke tui.KeyEvent) { a.cycleModel() }),
			tui.OnKey(tui.KeyCtrlZ, func(ke tui.KeyEvent) {
				cp := a.checkpoints.Pop()
				if cp != nil {
					if err := a.checkpoints.Restore(*cp); err == nil && a.app != nil {
						a.app.PrintAboveln("[Restored: %s]", cp.Description)
					}
				}
			}),
		)
	}

	return km
}

templ (a *rootApp) Render() {
	if a.turboModal.IsVisible() {
		@a.turboModal
	} else if a.permModal.Visible() {
		<div class="flex-col">
			@a.permModal
			@a.chatView
		</div>
	} else if a.screen.Get() == ScreenIDChat {
		<div class="flex-col">
			@a.topBarView
			@a.chatView
			if a.debugOverlay.IsVisible() {
				@a.debugOverlay
			}
		</div>
	} else if a.screen.Get() == ScreenIDTeam {
		@a.teamView
	} else if a.screen.Get() == ScreenIDMCP {
		@a.mcpView
	} else if a.screen.Get() == ScreenIDSkills {
		@a.skillsView
	} else if a.screen.Get() == ScreenIDEnhance {
		@a.enhanceView
	} else if a.screen.Get() == ScreenIDCheatSheet {
		@a.cheatSheetView
	} else if a.screen.Get() == ScreenIDDiffPager {
		@a.diffPagerView
	}
}
