package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	menuBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39")).
			Padding(1, 2)
	menuTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			MarginBottom(1)
	menuItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))
	menuSelectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")).
				Bold(true)
)

type MenuItem struct {
	Label string
	Key   rune
}

type MenuModel struct {
	title    string
	items    []MenuItem
	selected int
}

func NewMenuModel(title string, items []MenuItem) MenuModel {
	return MenuModel{
		title: title,
		items: items,
	}
}

func (menu *MenuModel) SelectedIndex() int {
	return menu.selected
}

func (menu *MenuModel) Selected() MenuItem {
	return menu.items[menu.selected]
}

func (menu *MenuModel) MoveDown() {
	if menu.selected < len(menu.items)-1 {
		menu.selected++
	}
}

func (menu *MenuModel) MoveUp() {
	if menu.selected > 0 {
		menu.selected--
	}
}

func (menu MenuModel) View() string {
	title := menuTitleStyle.Render(menu.title)

	var lines []string
	for index, item := range menu.items {
		style := menuItemStyle
		prefix := "  "
		isSelected := index == menu.selected
		if isSelected {
			style = menuSelectedItemStyle
			prefix = "▸ "
		}
		line := style.Render(fmt.Sprintf("%s[%c] %s", prefix, item.Key, item.Label))
		lines = append(lines, line)
	}

	content := title + "\n" + strings.Join(lines, "\n")
	return menuBorderStyle.Render(content)
}
