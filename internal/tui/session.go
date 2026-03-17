package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/internal/state"
)

const sessionHeaderHeight = 3

var sessionHeaderStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("39")).
	BorderStyle(lipgloss.NormalBorder()).
	BorderBottom(true).
	BorderForeground(lipgloss.Color("240"))

type SessionModel struct {
	thread   *state.ThreadState
	viewport viewport.Model
	ready    bool
}

func NewSessionModel(thread *state.ThreadState) SessionModel {
	return SessionModel{thread: thread}
}

func (session *SessionModel) SetSize(width int, height int) {
	viewHeight := height - sessionHeaderHeight
	if viewHeight < 1 {
		viewHeight = 1
	}
	session.viewport = viewport.New(width, viewHeight)
	session.viewport.SetContent(RenderMessages(session.thread))
	session.ready = true
}

func (session *SessionModel) Refresh() {
	if !session.ready {
		return
	}
	session.viewport.SetContent(RenderMessages(session.thread))
	session.viewport.GotoBottom()
}

func (session SessionModel) View() string {
	header := sessionHeaderStyle.Render(
		fmt.Sprintf("%s [%s]", session.thread.Title, session.thread.Status),
	)

	if !session.ready {
		return header + "\n" + RenderMessages(session.thread)
	}

	return header + "\n" + session.viewport.View()
}
