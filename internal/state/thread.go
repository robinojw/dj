package state

const (
	StatusIdle      = "idle"
	StatusActive    = "active"
	StatusCompleted = "completed"
	StatusError     = "error"
)

type ChatMessage struct {
	ID      string
	Role    string
	Content string
}

type ThreadState struct {
	ID             string
	Title          string
	Status         string
	Activity       string
	ParentID       string
	AgentNickname  string
	AgentRole      string
	AgentProcessID string
	Depth          int
	Messages       []ChatMessage
	CommandOutput  map[string]string
}

func NewThreadState(id string, title string) *ThreadState {
	return &ThreadState{
		ID:            id,
		Title:         title,
		Status:        StatusIdle,
		Messages:      make([]ChatMessage, 0),
		CommandOutput: make(map[string]string),
	}
}

func (threadState *ThreadState) AppendMessage(msg ChatMessage) {
	threadState.Messages = append(threadState.Messages, msg)
}

func (threadState *ThreadState) AppendDelta(messageID string, delta string) {
	for index := range threadState.Messages {
		if threadState.Messages[index].ID == messageID {
			threadState.Messages[index].Content += delta
			return
		}
	}
}

func (threadState *ThreadState) AppendOutput(execID string, data string) {
	threadState.CommandOutput[execID] += data
}

func (threadState *ThreadState) SetActivity(activity string) {
	threadState.Activity = activity
}

func (threadState *ThreadState) ClearActivity() {
	threadState.Activity = ""
}
