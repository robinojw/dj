package screens

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/robinojw/dj/internal/agents"
	"github.com/robinojw/dj/internal/api"
	"github.com/robinojw/dj/internal/mentions"
	"github.com/robinojw/dj/internal/tui/components"
	"github.com/robinojw/dj/internal/tui/theme"
)

// SubmitMsg is sent when the user presses Enter to submit a message.
type SubmitMsg struct {
	Text           string
	MentionContext string // resolved @mention content appended to prompt
}

// StreamDeltaMsg carries a text delta from the SSE stream.
type StreamDeltaMsg struct {
	Delta string
}

// StreamDoneMsg signals the stream has completed.
type StreamDoneMsg struct {
	Usage api.Usage
}

// StreamErrorMsg signals a streaming error.
type StreamErrorMsg struct {
	Err error
}

// StreamDiffMsg carries a git diff result.
type StreamDiffMsg struct {
	FilePath  string
	DiffText  string
	Timestamp time.Time
}

// ChatModel is the single-agent chat screen.
type ChatModel struct {
	viewport         viewport.Model
	input            components.ChatInput
	statusBar        components.StatusBar
	messages         []chatMessage
	diffs            []CollapsibleDiff
	focusedDiffIndex int // -1 when no diff focused
	viewportMode     string
	streaming        bool
	buffer           strings.Builder // accumulates current assistant response
	Mode             agents.AgentMode
	width            int
	height           int
	theme            *theme.Theme
}

type chatMessage struct {
	Role    string // "user" or "assistant"
	Content string
}

// CollapsibleDiff represents a git diff that can be expanded/collapsed.
type CollapsibleDiff struct {
	ID        string
	FilePath  string
	DiffLines []string
	Expanded  bool
	Timestamp time.Time
}

func NewChatModel(t *theme.Theme) ChatModel {
	vp := viewport.New(80, 20)
	vp.SetContent("")

	return ChatModel{
		viewport:         vp,
		input:            components.NewChatInput(t),
		statusBar:        components.NewStatusBar(t),
		theme:            t,
		diffs:            make([]CollapsibleDiff, 0),
		focusedDiffIndex: -1,
		viewportMode:     "chat",
	}
}

func (m ChatModel) Init() tea.Cmd {
	return m.input.Focus()
}

func (m ChatModel) Update(msg tea.Msg) (ChatModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 4 // room for input + status
		m.input.SetWidth(msg.Width)
		m.statusBar.Width = msg.Width
		m.updateViewport()

	case tea.KeyMsg:
		// Diff navigation mode
		if m.viewportMode == "diff_nav" && len(m.diffs) > 0 {
			switch msg.String() {
			case "tab":
				m.focusedDiffIndex = (m.focusedDiffIndex + 1) % len(m.diffs)
				m.updateViewport()
				return m, nil

			case "shift+tab":
				m.focusedDiffIndex--
				if m.focusedDiffIndex < 0 {
					m.focusedDiffIndex = len(m.diffs) - 1
				}
				m.updateViewport()
				return m, nil

			case "enter", " ":
				if m.focusedDiffIndex >= 0 && m.focusedDiffIndex < len(m.diffs) {
					m.diffs[m.focusedDiffIndex].Expanded = !m.diffs[m.focusedDiffIndex].Expanded
					m.updateViewport()
				}
				return m, nil

			case "esc":
				m.viewportMode = "chat"
				m.focusedDiffIndex = -1
				m.updateViewport()
				return m, nil
			}
		}

		// Existing message submission logic
		if msg.String() == "enter" && !m.streaming {
			text := strings.TrimSpace(m.input.Value())
			if text != "" {
				m.messages = append(m.messages, chatMessage{Role: "user", Content: text})
				m.input.Reset()
				m.streaming = true
				m.buffer.Reset()
				m.updateViewport()

				// Parse and resolve @mentions
				parsed := mentions.Parse(text)
				var mentionCtx string
				if len(parsed) > 0 {
					resolved := mentions.Resolve(context.Background(), parsed)
					mentionCtx = mentions.FormatResolved(resolved)
					text = mentions.StripMentions(text)
				}

				return m, func() tea.Msg {
					return SubmitMsg{Text: text, MentionContext: mentionCtx}
				}
			}
		}

	case StreamDeltaMsg:
		m.buffer.WriteString(msg.Delta)
		m.updateViewport()

	case StreamDoneMsg:
		m.messages = append(m.messages, chatMessage{
			Role:    "assistant",
			Content: m.buffer.String(),
		})
		m.buffer.Reset()
		m.streaming = false
		m.statusBar.InputTokens += msg.Usage.InputTokens
		m.statusBar.OutputTokens += msg.Usage.OutputTokens
		m.updateViewport()

	case StreamErrorMsg:
		m.messages = append(m.messages, chatMessage{
			Role:    "assistant",
			Content: "Error: " + msg.Err.Error(),
		})
		m.buffer.Reset()
		m.streaming = false
		m.updateViewport()

	case StreamDiffMsg:
		diff := CollapsibleDiff{
			ID:        fmt.Sprintf("diff-%d", time.Now().UnixNano()),
			FilePath:  msg.FilePath,
			DiffLines: parseDiffLines(msg.DiffText),
			Expanded:  false, // collapsed by default
			Timestamp: msg.Timestamp,
		}
		m.diffs = append(m.diffs, diff)
		m.focusedDiffIndex = len(m.diffs) - 1 // auto-focus latest
		m.viewportMode = "diff_nav"
		m.updateViewport()
	}

	// Update sub-components
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *ChatModel) updateViewport() {
	var lines []string
	for _, msg := range m.messages {
		switch msg.Role {
		case "user":
			label := m.theme.AccentStyle().Render("You: ")
			lines = append(lines, label+msg.Content)
		case "assistant":
			label := m.theme.PrimaryStyle().Render("DJ: ")
			lines = append(lines, label+msg.Content)
		}
		lines = append(lines, "")
	}

	// Show streaming buffer
	if m.streaming && m.buffer.Len() > 0 {
		label := m.theme.PrimaryStyle().Render("DJ: ")
		lines = append(lines, label+m.buffer.String())
		lines = append(lines, "")
	}

	if m.streaming && m.buffer.Len() == 0 {
		lines = append(lines, m.theme.MutedStyle().Render("Thinking..."))
	}

	m.viewport.SetContent(strings.Join(lines, "\n"))
	m.viewport.GotoBottom()
}

func (m ChatModel) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		m.viewport.View(),
		m.input.View(),
		m.statusBar.View(),
	)
}

func (m *ChatModel) SetCost(cost float64) {
	m.statusBar.CumulativeCost = cost
}

func (m *ChatModel) SetActiveMCPs(names []string) {
	m.statusBar.ActiveMCPs = names
}

func (m *ChatModel) SetMode(mode agents.AgentMode) {
	m.Mode = mode
	m.statusBar.Mode = mode
}

func (m *ChatModel) SetModel(model string) {
	m.statusBar.Model = model
}

// parseDiffLines splits git diff output into individual lines.
func parseDiffLines(diffText string) []string {
	if diffText == "" {
		return []string{}
	}
	return strings.Split(diffText, "\n")
}
