package tui

const (
	defaultSplitRatio     = 0.5
	dividerHeight         = 1
	sessionPaneBorderSize = 2
)

type SessionPanelModel struct {
	pinnedSessions []string
	activePaneIdx  int
	splitRatio     float64
	zoomed         bool
}

func NewSessionPanelModel() SessionPanelModel {
	return SessionPanelModel{
		splitRatio: defaultSplitRatio,
	}
}

func (panel *SessionPanelModel) Pin(threadID string) {
	if panel.IsPinned(threadID) {
		return
	}
	panel.pinnedSessions = append(panel.pinnedSessions, threadID)
}

func (panel *SessionPanelModel) Unpin(threadID string) {
	filtered := make([]string, 0, len(panel.pinnedSessions))
	for _, pinned := range panel.pinnedSessions {
		if pinned != threadID {
			filtered = append(filtered, pinned)
		}
	}
	panel.pinnedSessions = filtered
	panel.clampActivePaneIdx()
}

func (panel *SessionPanelModel) IsPinned(threadID string) bool {
	for _, pinned := range panel.pinnedSessions {
		if pinned == threadID {
			return true
		}
	}
	return false
}

func (panel *SessionPanelModel) PinnedSessions() []string {
	return panel.pinnedSessions
}

func (panel *SessionPanelModel) ActivePaneIdx() int {
	return panel.activePaneIdx
}

func (panel *SessionPanelModel) SetActivePaneIdx(index int) {
	panel.activePaneIdx = index
	panel.clampActivePaneIdx()
}

func (panel *SessionPanelModel) ActiveThreadID() string {
	if len(panel.pinnedSessions) == 0 {
		return ""
	}
	return panel.pinnedSessions[panel.activePaneIdx]
}

func (panel *SessionPanelModel) CycleRight() {
	maxIdx := len(panel.pinnedSessions) - 1
	if panel.activePaneIdx < maxIdx {
		panel.activePaneIdx++
	}
}

func (panel *SessionPanelModel) CycleLeft() {
	if panel.activePaneIdx > 0 {
		panel.activePaneIdx--
	}
}

func (panel SessionPanelModel) SplitRatio() float64 {
	return panel.splitRatio
}

func (panel SessionPanelModel) SessionDimensions(panelWidth int, panelHeight int) (int, int) {
	count := len(panel.pinnedSessions)
	if count == 0 {
		return 0, 0
	}
	sessionWidth := panelWidth / count
	sessionHeight := panelHeight - dividerHeight
	return sessionWidth, sessionHeight
}

func (panel *SessionPanelModel) Zoomed() bool {
	return panel.zoomed
}

func (panel *SessionPanelModel) ToggleZoom() {
	panel.zoomed = !panel.zoomed
}

func (panel *SessionPanelModel) clampActivePaneIdx() {
	maxIdx := len(panel.pinnedSessions) - 1
	if maxIdx < 0 {
		panel.activePaneIdx = 0
		return
	}
	if panel.activePaneIdx > maxIdx {
		panel.activePaneIdx = maxIdx
	}
}
