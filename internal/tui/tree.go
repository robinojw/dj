package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/state"
)

var (
	treeSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")).
				Bold(true)
	treeNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))
	treeDimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
)

type TreeModel struct {
	store    *state.ThreadStore
	selected int
	flatList []string
}

func NewTreeModel(store *state.ThreadStore) TreeModel {
	tree := TreeModel{store: store}
	tree.rebuild()
	return tree
}

func (tree *TreeModel) SelectedID() string {
	if len(tree.flatList) == 0 {
		return ""
	}
	return tree.flatList[tree.selected]
}

func (tree *TreeModel) MoveDown() {
	if tree.selected < len(tree.flatList)-1 {
		tree.selected++
	}
}

func (tree *TreeModel) MoveUp() {
	if tree.selected > 0 {
		tree.selected--
	}
}

func (tree *TreeModel) Refresh() {
	tree.rebuild()
}

func (tree *TreeModel) rebuild() {
	tree.flatList = nil
	roots := tree.store.Roots()
	for _, root := range roots {
		tree.flatList = append(tree.flatList, root.ID)
		tree.addChildren(root.ID)
	}
}

func (tree *TreeModel) addChildren(parentID string) {
	children := tree.store.Children(parentID)
	for _, child := range children {
		tree.flatList = append(tree.flatList, child.ID)
		tree.addChildren(child.ID)
	}
}

func (tree *TreeModel) depthOf(id string) int {
	thread, exists := tree.store.Get(id)
	if !exists || thread.ParentID == "" {
		return 0
	}
	return 1 + tree.depthOf(thread.ParentID)
}

func (tree *TreeModel) View() string {
	if len(tree.flatList) == 0 {
		return treeDimStyle.Render("No threads")
	}

	var lines []string
	for index, id := range tree.flatList {
		thread, exists := tree.store.Get(id)
		if !exists {
			continue
		}

		depth := tree.depthOf(id)
		indent := strings.Repeat("  ", depth)
		prefix := "├─"
		if depth == 0 {
			prefix = "●"
		}

		label := fmt.Sprintf("%s%s %s", indent, prefix, thread.Title)

		style := treeNormalStyle
		if index == tree.selected {
			style = treeSelectedStyle
		}
		lines = append(lines, style.Render(label))
	}

	return strings.Join(lines, "\n")
}
