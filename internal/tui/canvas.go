package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/state"
)

const (
	canvasColumns = 3
	rowGap        = 1
	columnGap     = 2
)

type CanvasModel struct {
	store     *state.ThreadStore
	selected  int
	pinnedIDs map[string]bool
	width     int
	height    int
}

func NewCanvasModel(store *state.ThreadStore) CanvasModel {
	return CanvasModel{store: store}
}

func (canvas *CanvasModel) SetPinnedIDs(pinned []string) {
	canvas.pinnedIDs = make(map[string]bool, len(pinned))
	for _, id := range pinned {
		canvas.pinnedIDs[id] = true
	}
}

func (canvas *CanvasModel) SetDimensions(width int, height int) {
	canvas.width = width
	canvas.height = height
}

func (canvas *CanvasModel) SelectedIndex() int {
	return canvas.selected
}

func (canvas *CanvasModel) SelectedThreadID() string {
	threads := canvas.store.All()
	if len(threads) == 0 {
		return ""
	}
	return threads[canvas.selected].ID
}

func (canvas *CanvasModel) SetSelected(index int) {
	threads := canvas.store.All()
	isValidIndex := index >= 0 && index < len(threads)
	if isValidIndex {
		canvas.selected = index
	}
}

func (canvas *CanvasModel) MoveRight() {
	threads := canvas.store.All()
	if canvas.selected < len(threads)-1 {
		canvas.selected++
	}
}

func (canvas *CanvasModel) MoveLeft() {
	if canvas.selected > 0 {
		canvas.selected--
	}
}

func (canvas *CanvasModel) MoveDown() {
	threads := canvas.store.All()
	next := canvas.selected + canvasColumns
	if next < len(threads) {
		canvas.selected = next
	}
}

func (canvas *CanvasModel) MoveUp() {
	next := canvas.selected - canvasColumns
	if next >= 0 {
		canvas.selected = next
	}
}

func (canvas *CanvasModel) hasDimensions() bool {
	return canvas.width > 0 && canvas.height > 0
}

func (canvas *CanvasModel) centerContent(content string) string {
	return lipgloss.Place(canvas.width, canvas.height,
		lipgloss.Center, lipgloss.Center, content)
}

func (canvas *CanvasModel) View() string {
	threads := canvas.store.All()
	if len(threads) == 0 {
		return canvas.renderEmpty()
	}

	grid := canvas.renderGrid(threads)
	if canvas.hasDimensions() {
		return canvas.centerContent(grid)
	}
	return grid
}

func (canvas *CanvasModel) renderEmpty() string {
	emptyMessage := "No active threads. Press 'n' to create one."
	if canvas.hasDimensions() {
		return canvas.centerContent(emptyMessage)
	}
	return emptyMessage
}

func (canvas *CanvasModel) renderGrid(threads []*state.ThreadState) string {
	numRows := (len(threads) + canvasColumns - 1) / canvasColumns
	cardWidth, cardHeight := canvas.cardDimensions(numRows)

	gapStyle := lipgloss.NewStyle().Width(columnGap)

	var rows []string
	for rowStart := 0; rowStart < len(threads); rowStart += canvasColumns {
		rowEnd := rowStart + canvasColumns
		if rowEnd > len(threads) {
			rowEnd = len(threads)
		}

		var parts []string
		for index := rowStart; index < rowEnd; index++ {
			isNotFirstInRow := index > rowStart
			if isNotFirstInRow {
				parts = append(parts, gapStyle.Render(""))
			}
			isSelected := index == canvas.selected
			isPinned := canvas.pinnedIDs[threads[index].ID]
			card := NewCardModel(threads[index], isSelected, isPinned)
			card.SetSize(cardWidth, cardHeight)
			parts = append(parts, card.View())
		}

		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, parts...))
	}

	return strings.Join(rows, "\n")
}

func (canvas CanvasModel) cardDimensions(numRows int) (int, int) {
	missingDimensions := canvas.width == 0 || canvas.height == 0
	if missingDimensions {
		return minCardWidth, minCardHeight
	}

	totalColumnGaps := columnGap * (canvasColumns - 1)
	cardWidth := (canvas.width - totalColumnGaps) / canvasColumns
	if cardWidth < minCardWidth {
		cardWidth = minCardWidth
	}
	if cardWidth > maxCardWidth {
		cardWidth = maxCardWidth
	}

	totalRowGaps := rowGap * (numRows - 1)
	cardHeight := (canvas.height - totalRowGaps) / numRows
	if cardHeight < minCardHeight {
		cardHeight = minCardHeight
	}
	if cardHeight > maxCardHeight {
		cardHeight = maxCardHeight
	}

	return cardWidth, cardHeight
}
