package tui

import (
	"strings"
	"testing"
)

const (
	edgeTestCardWidth    = 20
	edgeTestGap          = 2
	edgeTestSecondColumn = 2
)

func TestRenderConnectorSimple(test *testing.T) {
	parentCol := 0
	childCols := []int{0}
	connector := renderConnectorRow(parentCol, childCols, edgeTestCardWidth, edgeTestGap)
	if !strings.Contains(connector, edgeVertical) {
		test.Error("expected vertical connector")
	}
}

func TestRenderConnectorBranching(test *testing.T) {
	parentCol := 0
	childCols := []int{0, edgeTestSecondColumn}
	connector := renderConnectorRow(parentCol, childCols, edgeTestCardWidth, edgeTestGap)
	hasBranch := strings.Contains(connector, edgeTeeDown) || strings.Contains(connector, edgeHorizontal)
	if !hasBranch {
		test.Error("expected branching connector with horizontal lines")
	}
}

func TestRenderConnectorNoChildren(test *testing.T) {
	parentCol := 0
	childCols := []int{}
	connector := renderConnectorRow(parentCol, childCols, edgeTestCardWidth, edgeTestGap)
	if connector != "" {
		test.Error("expected empty string for no children")
	}
}
