package tui

import (
	"context"
	"strings"
	"time"

	tui "github.com/grindlemire/go-tui"
	"github.com/robinojw/dj/internal/agents"
	"github.com/robinojw/dj/internal/api"
	"github.com/robinojw/dj/internal/mentions"
	"github.com/robinojw/dj/internal/modes"
	"github.com/robinojw/dj/internal/tui/theme"
)

type chat struct {
	app          *tui.App
	textareaRef  *tui.Ref
	scrollRef    *tui.Ref
	scrollY      *tui.State[int]
	sticky       *tui.State[bool]
	streaming    *tui.State[bool]
	chatMessages *tui.State[[]chatMsg]
	eventCh      chan streamEvent
	cancelStream context.CancelFunc
	apiMessages  []chatMessage
	diffs        []storedDiff
	onSubmit      func(text string, mentionCtx string)
	onOpenDiffs   func(diffs []storedDiff)
	onTitleChange func(title string)
	mode         *tui.State[modes.ExecutionMode]
	model        *tui.State[string]
	cost         *tui.State[float64]
	inputTokens  *tui.State[int]
	outputTokens *tui.State[int]
	activeMCPs   *tui.State[[]string]
	width        int
	t            *theme.Theme
}

type chatMessage struct {
	Role    string
	Content string
}

func NewChat(
	t *theme.Theme,
	width int,
	mode *tui.State[modes.ExecutionMode],
	model *tui.State[string],
	cost *tui.State[float64],
	inputTokens *tui.State[int],
	outputTokens *tui.State[int],
	activeMCPs *tui.State[[]string],
	onSubmit func(string, string),
	onOpenDiffs func([]storedDiff),
	onTitleChange func(string),
) *chat {
	return &chat{
		textareaRef:   tui.NewRef(),
		scrollRef:     tui.NewRef(),
		scrollY:       tui.NewState(0),
		sticky:        tui.NewState(true),
		streaming:     tui.NewState(false),
		chatMessages:  tui.NewState([]chatMsg{}),
		eventCh:       make(chan streamEvent, 100),
		apiMessages:   make([]chatMessage, 0),
		diffs:         make([]storedDiff, 0),
		onSubmit:      onSubmit,
		onOpenDiffs:   onOpenDiffs,
		onTitleChange: onTitleChange,
		mode:         mode,
		model:        model,
		cost:         cost,
		inputTokens:  inputTokens,
		outputTokens: outputTokens,
		activeMCPs:   activeMCPs,
		width:        width,
		t:            t,
	}
}

func (c *chat) Init() func() {
	return func() {
		c.cancelActiveStream()
	}
}

func (c *chat) Watchers() []tui.Watcher {
	return []tui.Watcher{
		tui.NewChannelWatcher(c.eventCh, c.onStreamEvent),
	}
}

func (c *chat) scrollToBottom() {
	if el := c.scrollRef.El(); el != nil {
		_, maxY := el.MaxScroll()
		c.scrollY.Set(maxY + 1)
	}
}

func (c *chat) scrollBy(delta int) {
	el := c.scrollRef.El()
	if el == nil {
		return
	}
	_, maxY := el.MaxScroll()
	newY := c.scrollY.Get() + delta
	if newY < 0 {
		newY = 0
	}
	if newY > maxY {
		newY = maxY
	}
	c.scrollY.Set(newY)
	c.sticky.Set(newY >= maxY)
}

func (c *chat) autoScroll() {
	if c.sticky.Get() {
		c.scrollToBottom()
	}
}

func (c *chat) onStreamEvent(ev streamEvent) {
	if !c.streaming.Get() && ev.Type == eventText {
		return
	}

	switch ev.Type {
	case eventText:
		c.chatMessages.Update(func(msgs []chatMsg) []chatMsg {
			lastIsAgent := len(msgs) > 0 && msgs[len(msgs)-1].Kind == chatMsgAgent
			if lastIsAgent {
				msgs[len(msgs)-1].Content += ev.Delta
				return msgs
			}
			return append(msgs, chatMsg{Kind: chatMsgAgent, Content: ev.Delta, Timestamp: time.Now()})
		})
		c.autoScroll()

	case eventDone:
		c.inputTokens.Update(func(v int) int { return v + ev.Usage.InputTokens })
		c.outputTokens.Update(func(v int) int { return v + ev.Usage.OutputTokens })
		c.streaming.Set(false)

	case eventError:
		c.chatMessages.Update(func(msgs []chatMsg) []chatMsg {
			return append(msgs, chatMsg{Kind: chatMsgError, Content: ev.Err.Error(), Timestamp: time.Now()})
		})
		c.apiMessages = append(c.apiMessages, chatMessage{Role: "assistant", Content: "Error: " + ev.Err.Error()})
		c.streaming.Set(false)
		c.autoScroll()

	case eventDiff:
		c.chatMessages.Update(func(msgs []chatMsg) []chatMsg {
			return append(msgs, chatMsg{
				Kind:      chatMsgDiff,
				FilePath:  ev.FilePath,
				DiffLines: strings.Split(ev.DiffText, "\n"),
				Timestamp: ev.Timestamp,
			})
		})
		c.diffs = append(c.diffs, storedDiff{
			FilePath:  ev.FilePath,
			DiffLines: strings.Split(ev.DiffText, "\n"),
			Timestamp: ev.Timestamp,
		})
		c.autoScroll()
	}
}

func (c *chat) submit(text string) {
	text = strings.TrimSpace(text)
	if text == "" || c.streaming.Get() {
		return
	}

	isFirstMessage := len(c.chatMessages.Get()) == 0
	c.chatMessages.Update(func(msgs []chatMsg) []chatMsg {
		return append(msgs, chatMsg{Kind: chatMsgUser, Content: text, Timestamp: time.Now()})
	})
	c.autoScroll()

	if isFirstMessage && c.onTitleChange != nil {
		title := text
		maxTitleLen := 40
		if len(title) > maxTitleLen {
			title = title[:maxTitleLen] + "..."
		}
		c.onTitleChange(title)
	}

	c.apiMessages = append(c.apiMessages, chatMessage{Role: "user", Content: text})
	c.streaming.Set(true)

	parsed := mentions.Parse(text)
	var mentionCtx string
	if len(parsed) > 0 {
		resolved := mentions.Resolve(context.Background(), parsed)
		mentionCtx = mentions.FormatResolved(resolved)
		text = mentions.StripMentions(text)
	}

	c.onSubmit(text, mentionCtx)
}

func (c *chat) StartStream(chunks <-chan api.ResponseChunk, errs <-chan error) {
	c.cancelActiveStream()

	ctx, cancel := context.WithCancel(context.Background())
	c.cancelStream = cancel
	go bridgeStreamToChannel(ctx, chunks, errs, c.eventCh)
}

func (c *chat) cancelActiveStream() {
	if c.cancelStream != nil {
		c.cancelStream()
		c.cancelStream = nil
	}
}

// AppendToolCallBlock adds a tool call message from a worker agent.
func (c *chat) AppendToolCallBlock(workerID, content string) {
	c.chatMessages.Update(func(msgs []chatMsg) []chatMsg {
		return append(msgs, chatMsg{Kind: chatMsgToolCall, WorkerID: workerID, Content: content, Timestamp: time.Now()})
	})
	c.autoScroll()
}

// AppendToolResultBlock adds a tool result message from a worker agent.
func (c *chat) AppendToolResultBlock(workerID, content string) {
	c.chatMessages.Update(func(msgs []chatMsg) []chatMsg {
		return append(msgs, chatMsg{Kind: chatMsgToolResult, WorkerID: workerID, Content: content, Timestamp: time.Now()})
	})
}

// AppendDiffBlock adds a diff block from a worker agent.
func (c *chat) AppendDiffBlock(diff *agents.DiffInfo) {
	c.chatMessages.Update(func(msgs []chatMsg) []chatMsg {
		return append(msgs, chatMsg{
			Kind:      chatMsgDiff,
			FilePath:  diff.FilePath,
			DiffLines: strings.Split(diff.DiffText, "\n"),
			Timestamp: diff.Timestamp,
		})
	})
	c.diffs = append(c.diffs, storedDiff{
		FilePath:  diff.FilePath,
		DiffLines: strings.Split(diff.DiffText, "\n"),
		Timestamp: diff.Timestamp,
	})
	c.autoScroll()
}

// LoadSession replaces the message list with a worker's stored session.
func (c *chat) LoadSession(session *agents.WorkerSession) {
	msgs := make([]chatMsg, 0, len(session.Turns))
	for _, turn := range session.Turns {
		msgs = append(msgs, sessionTurnToChatMsg(turn))
	}
	c.chatMessages.Set(msgs)
	c.scrollToBottom()
}

func sessionTurnToChatMsg(turn agents.SessionTurn) chatMsg {
	switch turn.Kind {
	case agents.TurnText:
		return chatMsg{Kind: chatMsgAgent, Content: turn.Content, Timestamp: turn.Timestamp}
	case agents.TurnToolCall:
		return chatMsg{Kind: chatMsgToolCall, ToolName: turn.ToolName, Content: turn.Content, Timestamp: turn.Timestamp}
	case agents.TurnToolResult:
		return chatMsg{Kind: chatMsgToolResult, Content: turn.Content, Timestamp: turn.Timestamp}
	case agents.TurnDiff:
		return chatMsg{Kind: chatMsgDiff, Content: turn.Content, Timestamp: turn.Timestamp}
	case agents.TurnError:
		return chatMsg{Kind: chatMsgError, Content: turn.Content, Timestamp: turn.Timestamp}
	default:
		return chatMsg{Kind: chatMsgSystem, Content: turn.Content, Timestamp: turn.Timestamp}
	}
}

// Messages returns the message history for API context.
func (c *chat) Messages() []chatMessage {
	return c.apiMessages
}

func (c *chat) KeyMap() tui.KeyMap {
	if c.streaming.Get() {
		return tui.KeyMap{
			tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) {
				c.cancelActiveStream()
				c.streaming.Set(false)
			}),
		}
	}

	return tui.KeyMap{
		tui.OnKey(tui.KeyCtrlF, func(ke tui.KeyEvent) {
			if len(c.diffs) > 0 {
				c.onOpenDiffs(c.diffs)
			}
		}),
		tui.OnKey(tui.KeyPageUp, func(ke tui.KeyEvent) {
			c.scrollBy(-10)
		}),
		tui.OnKey(tui.KeyPageDown, func(ke tui.KeyEvent) {
			c.scrollBy(10)
		}),
	}
}

templ (c *chat) Render() {
	<div class="flex-col h-full">
		<div class="overflow-y-scroll flex-1" ref={c.scrollRef} scrollOffset={0, c.scrollY.Get()}>
			for _, msg := range c.chatMessages.Get() {
				@RenderChatMsg(c.t, msg)
			}
			if c.streaming.Get() {
				<span class="text-dim">{"Streaming..."}</span>
			}
		</div>
		<textarea ref={c.textareaRef} autoFocus={true}
			placeholder="Send a message... (/skills name)"
			width={c.width - 2}
			border={tui.BorderRounded}
			onSubmit={c.submit} />
		@StatusBar(c.t, c.model.Get(), c.mode.Get(), c.inputTokens.Get(), c.outputTokens.Get(), c.cost.Get(), c.activeMCPs.Get())
	</div>
}
