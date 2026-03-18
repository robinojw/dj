package tui

import tea "github.com/charmbracelet/bubbletea"

const (
	PrefixNone      = iota
	PrefixWaiting
	PrefixComplete
	PrefixCancelled
)

type PrefixHandler struct {
	active  bool
	action  rune
	keyType tea.KeyType
}

func NewPrefixHandler() *PrefixHandler {
	return &PrefixHandler{}
}

func (handler *PrefixHandler) Active() bool {
	return handler.active
}

func (handler *PrefixHandler) Action() rune {
	return handler.action
}

func (handler *PrefixHandler) KeyType() tea.KeyType {
	return handler.keyType
}

func (handler *PrefixHandler) HandleKey(msg tea.KeyMsg) int {
	if !handler.active {
		if msg.Type == tea.KeyCtrlB {
			handler.active = true
			return PrefixWaiting
		}
		return PrefixNone
	}

	handler.active = false

	if msg.Type == tea.KeyEsc {
		return PrefixCancelled
	}

	hasRunes := msg.Type == tea.KeyRunes && len(msg.Runes) > 0
	if hasRunes {
		handler.action = msg.Runes[0]
		handler.keyType = msg.Type
		return PrefixComplete
	}

	isArrow := msg.Type == tea.KeyLeft || msg.Type == tea.KeyRight ||
		msg.Type == tea.KeyUp || msg.Type == tea.KeyDown
	if isArrow {
		handler.action = 0
		handler.keyType = msg.Type
		return PrefixComplete
	}

	return PrefixCancelled
}
