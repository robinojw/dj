package tui

import (
	"strings"
	"testing"
)

func TestMenuRender(t *testing.T) {
	items := []MenuItem{
		{Label: "Fork Thread", Key: 'f'},
		{Label: "Delete Thread", Key: 'd'},
		{Label: "Rename Thread", Key: 'r'},
	}
	menu := NewMenuModel("Thread Actions", items)

	output := menu.View()
	if !strings.Contains(output, "Fork Thread") {
		t.Errorf("expected Fork Thread in output:\n%s", output)
	}
	if !strings.Contains(output, "Delete Thread") {
		t.Errorf("expected Delete Thread in output:\n%s", output)
	}
}

func TestMenuNavigation(t *testing.T) {
	items := []MenuItem{
		{Label: "First", Key: 'a'},
		{Label: "Second", Key: 'b'},
	}
	menu := NewMenuModel("Test", items)

	if menu.SelectedIndex() != 0 {
		t.Errorf("expected 0, got %d", menu.SelectedIndex())
	}

	menu.MoveDown()
	if menu.SelectedIndex() != 1 {
		t.Errorf("expected 1, got %d", menu.SelectedIndex())
	}

	menu.MoveDown()
	if menu.SelectedIndex() != 1 {
		t.Errorf("expected clamped at 1, got %d", menu.SelectedIndex())
	}
}

func TestMenuSelect(t *testing.T) {
	items := []MenuItem{
		{Label: "Fork", Key: 'f'},
		{Label: "Delete", Key: 'd'},
	}
	menu := NewMenuModel("Test", items)

	selected := menu.Selected()
	if selected.Key != 'f' {
		t.Errorf("expected f, got %c", selected.Key)
	}
}

func TestMenuMoveUp(t *testing.T) {
	items := []MenuItem{
		{Label: "First", Key: 'a'},
		{Label: "Second", Key: 'b'},
	}
	menu := NewMenuModel("Test", items)

	menu.MoveUp()
	if menu.SelectedIndex() != 0 {
		t.Errorf("expected clamped at 0, got %d", menu.SelectedIndex())
	}

	menu.MoveDown()
	menu.MoveUp()
	if menu.SelectedIndex() != 0 {
		t.Errorf("expected 0 after down+up, got %d", menu.SelectedIndex())
	}
}
