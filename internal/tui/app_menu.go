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
		app.closeMenu()
		return app.dispatchMenuAction(selected)
	}
	return app, nil
}

func (app AppModel) handlePrefixAction() (tea.Model, tea.Cmd) {
	action := app.prefix.Action()
	if action == 'm' {
		app.showMenu()
	}
	return app, nil
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
}
