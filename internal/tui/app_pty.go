package tui

import tea "github.com/charmbracelet/bubbletea"

func (app AppModel) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	isScrollWheel := msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown
	if !isScrollWheel {
		return app, nil
	}

	if app.focusPane != FocusPaneSession {
		return app, nil
	}

	activeID := app.sessionPanel.ActiveThreadID()
	if activeID == "" {
		return app, nil
	}

	ptySession, exists := app.ptySessions[activeID]
	if !exists {
		return app, nil
	}

	if msg.Button == tea.MouseButtonWheelUp {
		ptySession.ScrollUp(scrollStep)
	} else {
		ptySession.ScrollDown(scrollStep)
	}

	return app, nil
}

func (app AppModel) handleSessionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return app, tea.Quit
	case tea.KeyEsc:
		app.closeSession()
		return app, nil
	default:
		return app.forwardKeyToPTY(msg)
	}
}

func (app AppModel) forwardKeyToPTY(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	activeID := app.sessionPanel.ActiveThreadID()
	if activeID == "" {
		return app, nil
	}

	ptySession, exists := app.ptySessions[activeID]
	if !exists {
		return app, nil
	}

	data := KeyMsgToBytes(msg)
	if data == nil {
		return app, nil
	}

	ptySession.WriteBytes(data)
	return app, nil
}

func (app AppModel) openSession() (tea.Model, tea.Cmd) {
	threadID := app.canvas.SelectedThreadID()
	if threadID == "" {
		return app, nil
	}

	if !app.sessionPanel.IsPinned(threadID) {
		pinned, _ := app.pinSession(threadID)
		app = pinned.(AppModel)
	}

	app.focusPane = FocusPaneSession
	app.sessionPanel.SetActivePaneIdx(app.pinnedIndex(threadID))
	return app, app.rebalancePTYSizes()
}

func (app AppModel) selectedThreadID() string {
	if app.canvasMode == CanvasModeTree {
		return app.tree.SelectedID()
	}
	return app.canvas.SelectedThreadID()
}

func (app AppModel) killSession() (tea.Model, tea.Cmd) {
	threadID := app.selectedThreadID()
	if threadID == "" {
		return app, nil
	}

	app.stopAndRemovePTY(threadID)
	app.sessionPanel.Unpin(threadID)
	app.store.Delete(threadID)
	app.canvas.ClampSelected()
	app.tree.Refresh()
	app.statusBar.SetThreadCount(len(app.store.All()))

	hasPinned := len(app.sessionPanel.PinnedSessions()) > 0
	if !hasPinned {
		app.focusPane = FocusPaneCanvas
	}

	return app, app.rebalancePTYSizes()
}

func (app *AppModel) stopAndRemovePTY(threadID string) {
	ptySession, exists := app.ptySessions[threadID]
	if !exists {
		return
	}
	ptySession.Stop()
	delete(app.ptySessions, threadID)
	hasNoSessions := len(app.ptySessions) == 0
	if hasNoSessions {
		app.statusBar.SetConnected(false)
	}
}

func (app AppModel) togglePin() (tea.Model, tea.Cmd) {
	threadID := app.canvas.SelectedThreadID()
	if threadID == "" {
		return app, nil
	}

	if app.sessionPanel.IsPinned(threadID) {
		app.sessionPanel.Unpin(threadID)
		return app, app.rebalancePTYSizes()
	}

	return app.pinSession(threadID)
}

func (app AppModel) pinSession(threadID string) (tea.Model, tea.Cmd) {
	_, exists := app.store.Get(threadID)
	if !exists {
		return app, nil
	}

	app.ensurePTYSession(threadID)

	app.sessionPanel.Pin(threadID)
	return app, app.rebalancePTYSizes()
}

func (app *AppModel) ensurePTYSession(threadID string) {
	_, hasPTY := app.ptySessions[threadID]
	if hasPTY {
		return
	}

	ptySession := NewPTYSession(PTYSessionConfig{
		ThreadID: threadID,
		Command:  app.resolveInteractiveCmd(),
		Args:     app.interactiveArgs,
		SendMsg:  app.ptyEventCallback(),
	})
	if err := ptySession.Start(); err != nil {
		app.statusBar.SetError(err.Error())
		return
	}
	app.ptySessions[threadID] = ptySession
	app.statusBar.SetConnected(true)
}

func (app AppModel) pinnedIndex(threadID string) int {
	for index, pinned := range app.sessionPanel.PinnedSessions() {
		if pinned == threadID {
			return index
		}
	}
	return 0
}

func (app AppModel) resolveInteractiveCmd() string {
	if app.interactiveCmd != "" {
		return app.interactiveCmd
	}
	return "codex"
}

func (app AppModel) ptyEventCallback() func(PTYOutputMsg) {
	events := app.ptyEvents
	return func(msg PTYOutputMsg) {
		events <- msg
	}
}

func (app *AppModel) closeSession() {
	app.focusPane = FocusPaneCanvas
}

func (app AppModel) handlePTYOutput(msg PTYOutputMsg) (tea.Model, tea.Cmd) {
	return app, app.listenForPTYEvents()
}

func (app AppModel) listenForPTYEvents() tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-app.ptyEvents
		if !ok {
			return nil
		}
		return msg
	}
}

func (app AppModel) rebalancePTYSizes() tea.Cmd {
	pinned := app.sessionPanel.PinnedSessions()
	if len(pinned) == 0 {
		return nil
	}

	hasDimensions := app.width > 0 && app.height > 0
	if !hasDimensions {
		return nil
	}

	contentWidth, contentHeight := app.panelContentDimensions()

	if app.sessionPanel.Zoomed() {
		app.resizeZoomedSession(contentWidth, contentHeight)
		return nil
	}

	app.resizeSplitSessions(pinned, contentWidth, contentHeight)
	return nil
}

func (app AppModel) panelContentDimensions() (int, int) {
	canvasHeight := int(float64(app.height) * app.sessionPanel.SplitRatio())
	panelHeight := app.height - canvasHeight - dividerHeight
	if panelHeight < 1 {
		panelHeight = 1
	}

	contentWidth := app.width - sessionPaneBorderSize
	contentHeight := panelHeight - sessionPaneBorderSize
	if contentWidth < 1 {
		contentWidth = 1
	}
	if contentHeight < 1 {
		contentHeight = 1
	}
	return contentWidth, contentHeight
}

func (app AppModel) resizeZoomedSession(contentWidth int, contentHeight int) {
	activeID := app.sessionPanel.ActiveThreadID()
	ptySession, exists := app.ptySessions[activeID]
	if !exists {
		return
	}
	ptySession.Resize(contentWidth, contentHeight)
}

func (app AppModel) resizeSplitSessions(pinned []string, contentWidth int, contentHeight int) {
	count := len(pinned)
	paneWidth := app.width/count - sessionPaneBorderSize
	if paneWidth < 1 {
		paneWidth = 1
	}
	for _, threadID := range pinned {
		ptySession, exists := app.ptySessions[threadID]
		if !exists {
			continue
		}
		ptySession.Resize(paneWidth, contentHeight)
	}
}

func (app *AppModel) StopAllPTYSessions() {
	for _, ptySession := range app.ptySessions {
		ptySession.Stop()
	}
}
