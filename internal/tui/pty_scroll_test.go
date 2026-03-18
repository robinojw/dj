package tui

import (
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

const (
	scrollTestLineWidth     = 10
	scrollTestViewportThree = 3
	scrollTestViewportTwo   = 2
	scrollTestOffsetOne     = 1
	scrollTestOffsetFour    = 4

	scrollTestVisible2  = "visible-2"
	scrollTestVis1      = "vis-1"
	scrollTestVis2      = "vis-2"
	scrollTestExpectFmt    = "expected %d lines, got %d"
	scrollTestExpectStrFmt = "expected %s, got %q"
)

func TestRenderScrolledViewport(testing *testing.T) {
	scrollbackLines := []uv.Line{
		uv.NewLine(scrollTestLineWidth),
		uv.NewLine(scrollTestLineWidth),
	}
	screenLines := []string{"visible-1", scrollTestVisible2, "visible-3"}

	result := renderScrolledViewport(scrollbackLines, screenLines, scrollTestViewportThree, scrollTestOffsetOne)

	if len(result) != scrollTestViewportThree {
		testing.Fatalf(scrollTestExpectFmt, scrollTestViewportThree, len(result))
	}

	if !strings.Contains(result[scrollTestViewportThree-1], scrollTestVisible2) {
		testing.Errorf("expected %s at bottom, got %q", scrollTestVisible2, result[scrollTestViewportThree-1])
	}
}

func TestRenderScrolledViewportAtMaxOffset(testing *testing.T) {
	scrollbackLines := []uv.Line{
		uv.NewLine(scrollTestLineWidth),
		uv.NewLine(scrollTestLineWidth),
	}
	screenLines := []string{scrollTestVis1, scrollTestVis2}

	result := renderScrolledViewport(scrollbackLines, screenLines, scrollTestViewportTwo, scrollTestOffsetFour)

	if len(result) != scrollTestViewportTwo {
		testing.Fatalf(scrollTestExpectFmt, scrollTestViewportTwo, len(result))
	}
}

func TestRenderScrolledViewportZeroOffset(testing *testing.T) {
	scrollbackLines := []uv.Line{
		uv.NewLine(scrollTestLineWidth),
	}
	screenLines := []string{scrollTestVis1, scrollTestVis2}

	result := renderScrolledViewport(scrollbackLines, screenLines, scrollTestViewportTwo, 0)

	if len(result) != scrollTestViewportTwo {
		testing.Fatalf(scrollTestExpectFmt, scrollTestViewportTwo, len(result))
	}
	if result[0] != scrollTestVis1 {
		testing.Errorf(scrollTestExpectStrFmt, scrollTestVis1, result[0])
	}
	if result[1] != scrollTestVis2 {
		testing.Errorf(scrollTestExpectStrFmt, scrollTestVis2, result[1])
	}
}
