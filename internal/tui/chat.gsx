package tui

import (
	"context"
	"strings"

	tui "github.com/grindlemire/go-tui"
	"github.com/robinojw/dj/internal/api"
	"github.com/robinojw/dj/internal/mentions"
	"github.com/robinojw/dj/internal/modes"
	"github.com/robinojw/dj/internal/tui/theme"
)

type chat struct {
	app          *tui.App
	textareaRef  *tui.Ref
	streaming    *tui.State[bool]
	streamWriter *tui.StreamWriter
	eventCh      chan streamEvent
	cancelStream context.CancelFunc // cancels the bridge goroutine
	messages     []chatMessage      // kept for API context
	diffs        []storedDiff       // stored for diff pager
	onSubmit     func(text string, mentionCtx string) // callback to root
	onOpenDiffs  func(diffs []storedDiff)              // callback to open diff pager
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
) *chat {
	return &chat{
		textareaRef:  tui.NewRef(),
		streaming:    tui.NewState(false),
		eventCh:      make(chan streamEvent, 100),
		messages:     make([]chatMessage, 0),
		diffs:        make([]storedDiff, 0),
		onSubmit:     onSubmit,
		onOpenDiffs:  onOpenDiffs,
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

// Watchers returns the channel watcher for stream events.
// The eventCh is stable for the lifetime of the component — never replaced.
func (c *chat) Watchers() []tui.Watcher {
	return []tui.Watcher{
		tui.NewChannelWatcher(c.eventCh, c.onStreamEvent),
	}
}

func (c *chat) onStreamEvent(ev streamEvent) {
	// Ignore events if streaming was cancelled (e.g. user pressed Escape)
	if !c.streaming.Get() && ev.Type == eventText {
		return
	}

	switch ev.Type {
	case eventText:
		if c.app == nil {
			return
		}
		if c.streamWriter == nil {
			c.streamWriter = c.app.StreamAbove()
			c.streamWriter.WriteStyled("DJ: ", c.t.TuiPrimaryStyle())
		}
		c.streamWriter.Write([]byte(ev.Delta))

	case eventDone:
		if c.streamWriter != nil {
			c.streamWriter.Close()
			c.streamWriter = nil
		}
		c.inputTokens.Update(func(v int) int { return v + ev.Usage.InputTokens })
		c.outputTokens.Update(func(v int) int { return v + ev.Usage.OutputTokens })
		c.streaming.Set(false)

	case eventError:
		if c.streamWriter != nil {
			c.streamWriter.Close()
			c.streamWriter = nil
		}
		if c.app != nil {
			c.app.PrintAboveln("DJ: Error: %s", ev.Err.Error())
		}
		c.messages = append(c.messages, chatMessage{Role: "assistant", Content: "Error: " + ev.Err.Error()})
		c.streaming.Set(false)

	case eventDiff:
		summary := formatDiffSummary(ev.FilePath, ev.DiffText)
		if c.app != nil {
			c.app.PrintAboveln("%s", summary)
		}
		c.diffs = append(c.diffs, storedDiff{
			FilePath:  ev.FilePath,
			DiffLines: strings.Split(ev.DiffText, "\n"),
			Timestamp: ev.Timestamp,
		})
	}
}

func (c *chat) submit(text string) {
	text = strings.TrimSpace(text)
	if text == "" || c.streaming.Get() {
		return
	}

	if c.app != nil {
		c.app.PrintAboveln("You: %s", text)
	}
	c.messages = append(c.messages, chatMessage{Role: "user", Content: text})
	c.streaming.Set(true)

	// Parse and resolve @mentions
	parsed := mentions.Parse(text)
	var mentionCtx string
	if len(parsed) > 0 {
		resolved := mentions.Resolve(context.Background(), parsed)
		mentionCtx = mentions.FormatResolved(resolved)
		text = mentions.StripMentions(text)
	}

	c.onSubmit(text, mentionCtx)
}

// StartStream is called by the root app after api.Stream() returns channels.
// Uses the stable eventCh — the watcher stays attached.
func (c *chat) StartStream(chunks <-chan api.ResponseChunk, errs <-chan error) {
	// Cancel any previous stream goroutine
	c.cancelActiveStream()

	ctx, cancel := context.WithCancel(context.Background())
	c.cancelStream = cancel
	go bridgeStreamToChannel(ctx, chunks, errs, c.eventCh)
}

// cancelActiveStream cancels the current bridge goroutine and cleans up the stream writer.
func (c *chat) cancelActiveStream() {
	if c.cancelStream != nil {
		c.cancelStream()
		c.cancelStream = nil
	}
	if c.streamWriter != nil {
		c.streamWriter.Close()
		c.streamWriter = nil
	}
}

// Messages returns the message history for API context.
func (c *chat) Messages() []chatMessage {
	return c.messages
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
	}
}

templ (c *chat) Render() {
	<div class="flex-col">
		if c.streaming.Get() {
			<span class="text-dim">{"Streaming... (Esc to cancel)"}</span>
		}
		<textarea ref={c.textareaRef} autoFocus={true}
			placeholder="Send a message... (/skills name)"
			width={c.width - 2}
			border={tui.BorderRounded}
			onSubmit={c.submit} />
		@StatusBar(c.t, c.model.Get(), c.mode.Get(), c.inputTokens.Get(), c.outputTokens.Get(), c.cost.Get(), c.activeMCPs.Get())
	</div>
}
