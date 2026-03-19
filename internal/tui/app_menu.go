package tui

import tea "github.com/charmbracelet/bubbletea"

var threadMenuItems = []MenuItem{
	{Label: "Fork Thread", Key: 'f'},
	{Label: "Delete Thread", Key: 'd'},
	{Label: "Rename Thread", Key: 'r'},
}

func (app AppModel) MenuVisible() bool {
	return app.menuVisible
}

func (app AppModel) handleMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		app.closeMenu()
		return app, nil
	case tea.KeyUp:
		app.menu.MoveUp()
		return app, nil
	case tea.KeyDown:
		app.menu.MoveDown()
		return app, nil
	case tea.KeyEnter:
		selected := app.menu.Selected()
		intent := app.menuIntent
		app.closeMenu()
		return app.dispatchMenuByIntent(intent, selected)
	}
	return app, nil
}

func (app AppModel) handlePrefixAction() (tea.Model, tea.Cmd) {
	action := app.prefix.Action()
	keyType := app.prefix.KeyType()

	switch {
	case action == 'm':
		app.showMenu()
	case action == 'x':
		return app.unpinActiveSession()
	case action == 'z':
		return app.toggleZoom()
	case keyType == tea.KeyRight:
		app.sessionPanel.CycleRight()
	case keyType == tea.KeyLeft:
		app.sessionPanel.CycleLeft()
	case action >= '1' && action <= '9':
		return app.jumpToPane(action)
	}
	return app, nil
}

func (app AppModel) unpinActiveSession() (tea.Model, tea.Cmd) {
	activeID := app.sessionPanel.ActiveThreadID()
	if activeID == "" {
		return app, nil
	}
	app.sessionPanel.Unpin(activeID)
	hasPinned := len(app.sessionPanel.PinnedSessions()) > 0
	if !hasPinned {
		app.focusPane = FocusPaneCanvas
	}
	return app, app.rebalancePTYSizes()
}

func (app AppModel) toggleZoom() (tea.Model, tea.Cmd) {
	app.sessionPanel.ToggleZoom()
	return app, app.rebalancePTYSizes()
}

func (app AppModel) jumpToPane(digit rune) (tea.Model, tea.Cmd) {
	index := int(digit - '1')
	pinned := app.sessionPanel.PinnedSessions()
	if index >= len(pinned) {
		return app, nil
	}
	app.sessionPanel.SetActivePaneIdx(index)
	app.focusPane = FocusPaneSession
	return app, nil
}

func (app AppModel) dispatchMenuByIntent(intent MenuIntent, item MenuItem) (tea.Model, tea.Cmd) {
	switch intent {
	case MenuIntentPersonaPicker:
		return app.dispatchPersonaPick(item)
	case MenuIntentAgentPicker:
		return app.dispatchAgentPick(item)
	default:
		return app.dispatchMenuAction(item)
	}
}

func (app AppModel) dispatchMenuAction(item MenuItem) (tea.Model, tea.Cmd) {
	threadID := app.canvas.SelectedThreadID()
	if threadID == "" {
		return app, nil
	}

	switch item.Key {
	case 'f':
		return app, func() tea.Msg {
			return ForkThreadMsg{ParentID: threadID}
		}
	case 'd':
		return app, func() tea.Msg {
			return DeleteThreadMsg{ThreadID: threadID}
		}
	case 'r':
		return app, func() tea.Msg {
			return RenameThreadMsg{ThreadID: threadID}
		}
	}
	return app, nil
}

func (app *AppModel) showMenu() {
	app.menu = NewMenuModel("Thread Actions", threadMenuItems)
	app.menuVisible = true
}

func (app *AppModel) closeMenu() {
	app.menuVisible = false
	app.menuIntent = MenuIntentThread
}
