package tui

import tea "github.com/charmbracelet/bubbletea"

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

	_, hasPTY := app.ptySessions[threadID]
	if !hasPTY {
		ptySession := NewPTYSession(PTYSessionConfig{
			ThreadID: threadID,
			Command:  app.resolveInteractiveCmd(),
			Args:     app.interactiveArgs,
			SendMsg:  app.ptyEventCallback(),
		})
		if err := ptySession.Start(); err != nil {
			app.statusBar.SetError(err.Error())
			return app, nil
		}
		app.ptySessions[threadID] = ptySession
	}

	app.sessionPanel.Pin(threadID)
	return app, app.rebalancePTYSizes()
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

	if app.sessionPanel.Zoomed() {
		activeID := app.sessionPanel.ActiveThreadID()
		ptySession, exists := app.ptySessions[activeID]
		if exists {
			ptySession.Resize(contentWidth, contentHeight)
		}
		return nil
	}

	count := len(pinned)
	paneWidth := app.width/count - sessionPaneBorderSize
	if paneWidth < 1 {
		paneWidth = 1
	}
	for _, threadID := range pinned {
		ptySession, exists := app.ptySessions[threadID]
		if exists {
			ptySession.Resize(paneWidth, contentHeight)
		}
	}
	return nil
}

func (app *AppModel) StopAllPTYSessions() {
	for _, ptySession := range app.ptySessions {
		ptySession.Stop()
	}
}
