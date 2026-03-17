package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/state"
)

const canvasColumns = 3

type CanvasModel struct {
	store    *state.ThreadStore
	selected int
}

func NewCanvasModel(store *state.ThreadStore) CanvasModel {
	return CanvasModel{store: store}
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

func (canvas *CanvasModel) View() string {
	threads := canvas.store.All()
	if len(threads) == 0 {
		return "No active threads. Press 'n' to create one."
	}

	var rows []string
	for rowStart := 0; rowStart < len(threads); rowStart += canvasColumns {
		rowEnd := rowStart + canvasColumns
		if rowEnd > len(threads) {
			rowEnd = len(threads)
		}

		var cards []string
		for index := rowStart; index < rowEnd; index++ {
			isSelected := index == canvas.selected
			card := NewCardModel(threads[index], isSelected)
			cards = append(cards, card.View())
		}

		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cards...))
	}

	return strings.Join(rows, "\n")
}
