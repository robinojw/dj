package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinojw/dj/config"
	"github.com/robinojw/dj/internal/agents"
	"github.com/robinojw/dj/internal/api"
	"github.com/robinojw/dj/internal/checkpoint"
	"github.com/robinojw/dj/internal/hooks"
	"github.com/robinojw/dj/internal/modes"
	"github.com/robinojw/dj/internal/tui/components"
	"github.com/robinojw/dj/internal/tui/screens"
	"github.com/robinojw/dj/internal/tui/theme"
)

// Screen identifies the active TUI screen.
type Screen int

const (
	ScreenChat Screen = iota
	ScreenTeam
	ScreenEnhance
	ScreenMCP
	ScreenSkills
)

// App is the root bubbletea model.
type App struct {
	screen          Screen
	screenStack     []Screen
	chat            screens.ChatModel
	team            screens.TeamModel
	enhance         screens.EnhanceModel
	mcpManager      screens.MCPManagerModel
	skillBrowser    screens.SkillBrowserModel
	theme           *theme.Theme
	tracker         *api.Tracker
	client          *api.ResponsesClient
	model           string
	mode            modes.ExecutionMode
	gate            *modes.Gate
	permissionModal components.PermissionModal
	turboModal      components.TurboModal
	turboConfirmed  bool
	permRequestCh   chan modes.PermissionRequest
	checkpoints     *checkpoint.Manager
	hooks           *hooks.Runner
	debugOverlay    components.DebugOverlay
	debugMode       bool
	width           int
	height          int

	// Active stream state
	streamChunks <-chan api.ResponseChunk
	streamErrs   <-chan error
}

// NewApp creates the root application model.
func NewApp(
	t *theme.Theme,
	client *api.ResponsesClient,
	tracker *api.Tracker,
	model string,
	cfg config.Config,
) App {
	gate := modes.NewGate(
		modes.ModeConfirm,
		cfg.Execution.Allow.Tools,
		cfg.Execution.Deny.Tools,
	)

	app := App{
		screen:          ScreenChat,
		chat:            screens.NewChatModel(t),
		team:            screens.NewTeamModel(t),
		enhance:         screens.NewEnhanceModel(t),
		mcpManager:      screens.NewMCPManagerModel(t),
		skillBrowser:    screens.NewSkillBrowserModel(t),
		theme:           t,
		tracker:         tracker,
		client:          client,
		model:           model,
		mode:            modes.ModeConfirm,
		gate:            gate,
		permissionModal: components.NewPermissionModal(t),
		turboModal:      components.NewTurboModal(t),
		permRequestCh:   make(chan modes.PermissionRequest, 10),
		checkpoints:     checkpoint.NewManager(20),
		debugOverlay:    components.NewDebugOverlay(t),
	}
	app.chat.SetModel(model)
	return app
}

func (a App) Init() tea.Cmd {
	return a.chat.Init()
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.debugOverlay.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			return a, tea.Quit
		case "ctrl+e":
			if a.screen != ScreenEnhance {
				return a, a.pushScreen(ScreenEnhance)
			}
		case "ctrl+m":
			if a.screen != ScreenMCP {
				return a, a.pushScreen(ScreenMCP)
			}
		case "ctrl+k":
			if a.screen != ScreenSkills {
				return a, a.pushScreen(ScreenSkills)
			}
		case "ctrl+t":
			if a.screen != ScreenTeam {
				return a, a.pushScreen(ScreenTeam)
			}
		case "ctrl+/":
			a.cycleModel()
			return a, nil
		case "ctrl+d":
			a.debugMode = !a.debugMode
			a.debugOverlay.Toggle()
			return a, nil
		case "tab":
			// Cycle: Confirm → Plan → Turbo → Confirm
			newMode := (a.mode + 1) % 3

			// Check if switching to Turbo
			if newMode == modes.ModeTurbo && !a.turboConfirmed {
				a.turboModal.Show()
				respCh := make(chan bool, 1)
				a.turboModal.SetResponseChannel(respCh)

				go func() {
					confirmed := <-respCh
					if confirmed {
						a.turboConfirmed = true
						a.mode = modes.ModeTurbo
						a.gate.SetMode(modes.ModeTurbo)
						a.chat.SetMode(modes.ModeTurbo)
					}
				}()
				return a, nil
			}

			a.mode = newMode
			a.gate.SetMode(newMode)
			a.chat.SetMode(newMode)
			return a, nil
		case "ctrl+z":
			cp := a.checkpoints.Pop()
			if cp != nil {
				if err := a.checkpoints.Restore(*cp); err == nil {
					return a, func() tea.Msg {
						return screens.StreamDeltaMsg{Delta: fmt.Sprintf("\n[Restored: %s]\n", cp.Description)}
					}
				}
			}
			return a, nil
		case "esc":
			if a.screen != ScreenChat {
				return a, a.popScreen()
			}
		}

	case screens.SubmitMsg:
		cmd := a.handleSubmit(msg.Text)
		return a, cmd

	case screens.StreamDeltaMsg:
		// Forward to chat screen AND continue reading stream
		var chatCmd tea.Cmd
		a.chat, chatCmd = a.chat.Update(msg)

		// If stream is still active, read next chunk
		if a.streamChunks != nil {
			return a, tea.Batch(chatCmd, waitForStreamMessage(a.streamChunks, a.streamErrs))
		}
		return a, chatCmd

	case screens.StreamDoneMsg:
		// Stream completed successfully, clear state
		a.streamChunks = nil
		a.streamErrs = nil
		var chatCmd tea.Cmd
		a.chat, chatCmd = a.chat.Update(msg)
		return a, chatCmd

	case screens.StreamErrorMsg:
		// Stream errored, clear state
		a.streamChunks = nil
		a.streamErrs = nil
		if a.debugMode {
			a.debugOverlay.AddError(msg.Err.Error())
		}
		var chatCmd tea.Cmd
		a.chat, chatCmd = a.chat.Update(msg)
		return a, chatCmd

	case screens.TeamSpawnedMsg:
		return a, a.pushScreen(ScreenTeam)

	case agents.WorkerUpdate:
		// Convert UpdateDiffResult to StreamDiffMsg for the UI
		if msg.Type == agents.UpdateDiffResult && msg.DiffInfo != nil {
			return a, func() tea.Msg {
				return screens.StreamDiffMsg{
					FilePath:  msg.DiffInfo.FilePath,
					DiffText:  msg.DiffInfo.DiffText,
					Timestamp: msg.DiffInfo.Timestamp,
				}
			}
		}

	case modes.PermissionRequest:
		var cmd tea.Cmd
		a.permissionModal, cmd = a.permissionModal.Update(msg)
		return a, cmd
	}

	// Delegate to the active screen
	var cmd tea.Cmd
	switch a.screen {
	case ScreenChat:
		a.chat, cmd = a.chat.Update(msg)
	case ScreenTeam:
		a.team, cmd = a.team.Update(msg)
	case ScreenEnhance:
		a.enhance, cmd = a.enhance.Update(msg)
	case ScreenMCP:
		a.mcpManager, cmd = a.mcpManager.Update(msg)
	case ScreenSkills:
		a.skillBrowser, cmd = a.skillBrowser.Update(msg)
	}

	return a, cmd
}

func (a App) View() string {
	var base string
	switch a.screen {
	case ScreenTeam:
		base = a.team.View()
	case ScreenEnhance:
		base = a.enhance.View()
	case ScreenMCP:
		base = a.mcpManager.View()
	case ScreenSkills:
		base = a.skillBrowser.View()
	default:
		base = a.chat.View()
	}

	if a.debugMode {
		overlay := a.debugOverlay.View()
		if overlay != "" {
			// Place the debug panel in the top-right, overlaid on the base
			return lipgloss.Place(a.width, a.height,
				lipgloss.Right, lipgloss.Top,
				overlay,
				lipgloss.WithWhitespaceChars(" "),
				lipgloss.WithWhitespaceForeground(lipgloss.NoColor{}),
			) + "\n" + base
		}
	}

	return base
}

func (a *App) pushScreen(s Screen) tea.Cmd {
	a.screenStack = append(a.screenStack, a.screen)
	a.screen = s
	return nil
}

func (a *App) popScreen() tea.Cmd {
	if len(a.screenStack) == 0 {
		a.screen = ScreenChat
		return nil
	}
	a.screen = a.screenStack[len(a.screenStack)-1]
	a.screenStack = a.screenStack[:len(a.screenStack)-1]
	return nil
}

func (a *App) cycleModel() {
	models := api.CycleModels
	current := -1
	for i, m := range models {
		if m == a.model {
			current = i
			break
		}
	}
	next := models[(current+1)%len(models)]
	a.model = next
	a.tracker.SetModel(next)
	a.chat.SetModel(next)

	if a.debugMode {
		a.debugOverlay.AddInfo("Model switched to " + next)
	}
}

func (a *App) handleSubmit(text string) tea.Cmd {
	req := api.CreateResponseRequest{
		Model: a.model,
		Input: api.MakeStringInput(text),
		Reasoning: &api.Reasoning{
			Effort: "medium",
		},
		Stream: true,
	}

	if a.debugMode {
		a.debugOverlay.AddInfo(fmt.Sprintf("Starting stream with model: %s", a.model))
	}

	ctx := context.Background()
	chunks, errs := a.client.Stream(ctx, req)

	// Store channels for continued reading
	a.streamChunks = chunks
	a.streamErrs = errs

	// Return command to wait for first chunk/error
	return waitForStreamMessage(chunks, errs)
}

// waitForStreamMessage returns a command that waits for the next chunk or error
func waitForStreamMessage(chunks <-chan api.ResponseChunk, errs <-chan error) tea.Cmd {
	return func() tea.Msg {
		select {
		case chunk, ok := <-chunks:
			if !ok {
				// Chunks channel closed, check for final errors
				select {
				case err := <-errs:
					// Provide more helpful error messages
					errMsg := err.Error()
					if strings.Contains(errMsg, "401") || strings.Contains(errMsg, "authentication") {
						return screens.StreamErrorMsg{
							Err: fmt.Errorf("authentication failed: check OPENAI_API_KEY environment variable"),
						}
					} else if strings.Contains(errMsg, "404") {
						return screens.StreamErrorMsg{
							Err: fmt.Errorf("model not found: check model name in config (current: see debug overlay)"),
						}
					} else if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
						return screens.StreamErrorMsg{
							Err: fmt.Errorf("request timeout: API server took too long to respond"),
						}
					}
					return screens.StreamErrorMsg{Err: fmt.Errorf("API error: %w", err)}
				default:
					// Stream completed successfully
					return screens.StreamDoneMsg{Usage: api.Usage{}}
				}
			}

			// Process chunk based on type
			switch chunk.Type {
			case "response.output_text.delta":
				if chunk.Delta != "" {
					return screens.StreamDeltaMsg{Delta: chunk.Delta}
				}
			case "response.completed":
				if chunk.Response != nil {
					return screens.StreamDoneMsg{Usage: chunk.Response.Usage}
				}
			}

			// Got a chunk but didn't return a message (unknown type), try again
			return waitForStreamMessage(chunks, errs)()

		case err, ok := <-errs:
			if ok {
				// Provide helpful error context
				errMsg := err.Error()
				if strings.Contains(errMsg, "401") || strings.Contains(errMsg, "authentication") {
					return screens.StreamErrorMsg{
						Err: fmt.Errorf("authentication failed: check OPENAI_API_KEY environment variable"),
					}
				} else if strings.Contains(errMsg, "connection refused") {
					return screens.StreamErrorMsg{
						Err: fmt.Errorf("cannot connect to API: check network and base URL"),
					}
				}
				return screens.StreamErrorMsg{Err: fmt.Errorf("stream error: %w", err)}
			}
			// Error channel closed with no error
			return screens.StreamDoneMsg{Usage: api.Usage{}}
		}
	}
}
