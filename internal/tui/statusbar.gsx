package tui

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/robinojw/dj/internal/modes"
	"github.com/robinojw/dj/internal/tui/theme"
)

templ StatusBar(t *theme.Theme, model string, mode modes.ExecutionMode, inputTokens int, outputTokens int, cost float64, activeMCPs []string) {
	<div class="flex-row">
		<span class="text-cyan font-bold">{"● " + model}</span>
		<span class="text-dim">{" │ "}</span>
		<span class={modeClass(mode)}>{modeLabel(mode)}</span>
		<span class="text-dim">{" │ "}</span>
		<span class="text-dim">{formatTokens(inputTokens, outputTokens)}</span>
		<span class="text-dim">{" │ $" + fmt.Sprintf("%.4f", cost)}</span>
		if len(activeMCPs) > 0 {
			<span class="text-dim">{" │ MCP: " + strings.Join(activeMCPs, ", ")}</span>
		}
	</div>
}

func modeLabel(mode modes.ExecutionMode) string {
	switch mode {
	case modes.ModeTurbo:
		return "Turbo"
	case modes.ModePlan:
		return "Plan"
	default:
		return "Confirm"
	}
}

func modeClass(mode modes.ExecutionMode) string {
	switch mode {
	case modes.ModeTurbo:
		return "text-red font-bold"
	case modes.ModePlan:
		return "text-blue font-bold"
	default:
		return "text-yellow font-bold"
	}
}

func formatTokens(input, output int) string {
	return humanize.Comma(int64(input)) + "/" + humanize.Comma(int64(output))
}
