package tui

import (
	"strings"

	"github.com/robinojw/dj/internal/tui/theme"
)

templ RenderChatMsg(t *theme.Theme, msg chatMsg) {
	if msg.Kind == chatMsgUser {
		<div class="flex-col my-1 px-1">
			<span class="text-cyan font-bold">{"You"}</span>
			<span>{msg.Content}</span>
		</div>
	} else if msg.Kind == chatMsgAgent {
		<div class="flex-col my-1 px-1">
			<span class="text-magenta font-bold">{"DJ"}</span>
			<span>{msg.Content}</span>
		</div>
	} else if msg.Kind == chatMsgToolCall {
		<div class="border-rounded border-yellow px-1 my-1 flex-col">
			<span class="text-yellow font-bold">{"Tool: " + msg.ToolName}</span>
			<span class="text-dim">{msg.Content}</span>
		</div>
	} else if msg.Kind == chatMsgToolResult {
		<div class="border-rounded px-1 my-1 flex-col">
			<span class="text-dim font-bold">{"result"}</span>
			<span class="text-dim">{truncateMsg(msg.Content, 300)}</span>
		</div>
	} else if msg.Kind == chatMsgDiff {
		<div class="border-rounded px-1 my-1 flex-col">
			<span class="text-yellow font-bold">{"  " + msg.FilePath}</span>
			for _, line := range msg.DiffLines {
				if isDiffAdd(line) {
					<span class="text-green">{line}</span>
				} else if isDiffRemove(line) {
					<span class="text-red">{line}</span>
				} else {
					<span class="text-dim">{line}</span>
				}
			}
		</div>
	} else if msg.Kind == chatMsgError {
		<span class="text-red">{"Error: " + msg.Content}</span>
	}
}

func isDiffAdd(line string) bool {
	return strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++")
}

func isDiffRemove(line string) bool {
	return strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---")
}
