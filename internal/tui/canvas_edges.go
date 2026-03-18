package tui

import "strings"

const (
	edgeVertical    = "│"
	edgeHorizontal  = "─"
	edgeTeeRight    = "├"
	edgeElbow       = "└"
	edgeTeeDown     = "┬"
	edgeCornerRight = "┐"
	cardCenterDiv   = 2
)

func renderConnectorRow(parentCol int, childCols []int, cardWidth int, gap int) string {
	if len(childCols) == 0 {
		return ""
	}

	cellWidth := cardWidth + gap
	parentCenter := parentCol*cellWidth + cardWidth/cardCenterDiv
	totalWidth := computeConnectorWidth(childCols, cellWidth, cardWidth)

	return buildConnectorLine(parentCenter, childCols, cellWidth, cardWidth, totalWidth)
}

func computeConnectorWidth(childCols []int, cellWidth int, cardWidth int) int {
	maxCol := 0
	for _, col := range childCols {
		if col > maxCol {
			maxCol = col
		}
	}
	return maxCol*cellWidth + cardWidth
}

func buildConnectorLine(parentCenter int, childCols []int, cellWidth int, cardWidth int, totalWidth int) string {
	childCenters := buildChildCenterSet(childCols, cellWidth, cardWidth)
	spanStart, spanEnd := computeSpan(parentCenter, childCenters, totalWidth)

	topLine := strings.Repeat(" ", parentCenter) + edgeVertical
	bottomLine := renderBottomLine(parentCenter, childCenters, spanStart, spanEnd)

	return topLine + "\n" + bottomLine
}

func buildChildCenterSet(childCols []int, cellWidth int, cardWidth int) map[int]bool {
	childCenters := make(map[int]bool, len(childCols))
	for _, col := range childCols {
		center := col*cellWidth + cardWidth/cardCenterDiv
		childCenters[center] = true
	}
	return childCenters
}

func computeSpan(parentCenter int, childCenters map[int]bool, totalWidth int) (int, int) {
	spanStart := totalWidth
	spanEnd := 0
	for center := range childCenters {
		if center < spanStart {
			spanStart = center
		}
		if center > spanEnd {
			spanEnd = center
		}
	}
	if parentCenter < spanStart {
		spanStart = parentCenter
	}
	if parentCenter > spanEnd {
		spanEnd = parentCenter
	}
	return spanStart, spanEnd
}

func renderBottomLine(parentCenter int, childCenters map[int]bool, spanStart int, spanEnd int) string {
	var builder strings.Builder
	for position := 0; position <= spanEnd; position++ {
		isChild := childCenters[position]
		isParent := position == parentCenter
		isInSpan := position >= spanStart && position <= spanEnd

		builder.WriteString(resolveConnectorChar(isChild, isParent, isInSpan))
	}
	return builder.String()
}

func resolveConnectorChar(isChild bool, isParent bool, inSpan bool) string {
	isParentAndChild := isParent && isChild
	if isParentAndChild {
		return edgeTeeDown
	}
	if isParent {
		return edgeTeeDown
	}
	isChildInSpan := isChild && inSpan
	if isChildInSpan {
		return edgeElbow
	}
	if isChild {
		return edgeVertical
	}
	if inSpan {
		return edgeHorizontal
	}
	return " "
}
