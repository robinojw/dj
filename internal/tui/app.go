package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinojw/dj/internal/agents"
	"github.com/robinojw/dj/internal/api"
	"github.com/robinojw/dj/internal/checkpoint"
	"github.com/robinojw/dj/internal/hooks"
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
	screen       Screen
	screenStack  []Screen
	chat         screens.ChatModel
	team         screens.TeamModel
	enhance      screens.EnhanceModel
	mcpManager   screens.MCPManagerModel
	skillBrowser screens.SkillBrowserModel
	theme        *theme.Theme
	tracker      *api.Tracker
	client       *api.ResponsesClient
	model        string
	mode         agents.AgentMode
	checkpoints  *checkpoint.Manager
	hooks        *hooks.Runner
	width        int
	height       int
}

// NewApp creates the root application model.
func NewApp(
	t *theme.Theme,
	client *api.ResponsesClient,
	tracker *api.Tracker,
	model string,
) App {
	return App{
		screen:       ScreenChat,
		chat:         screens.NewChatModel(t),
		team:         screens.NewTeamModel(t),
		enhance:      screens.NewEnhanceModel(t),
		mcpManager:   screens.NewMCPManagerModel(t),
		skillBrowser: screens.NewSkillBrowserModel(t),
		theme:        t,
		tracker:      tracker,
		client:       client,
		model:        model,
		checkpoints:  checkpoint.NewManager(20),
	}
}

func (a App) Init() tea.Cmd {
	return a.chat.Init()
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+q":
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
		case "tab":
			if a.mode == agents.ModeBuild {
				a.mode = agents.ModePlan
			} else {
				a.mode = agents.ModeBuild
			}
			a.chat.SetMode(a.mode)
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
		return a, a.handleSubmit(msg.Text)

	case screens.TeamSpawnedMsg:
		return a, a.pushScreen(ScreenTeam)
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
	switch a.screen {
	case ScreenTeam:
		return a.team.View()
	case ScreenEnhance:
		return a.enhance.View()
	case ScreenMCP:
		return a.mcpManager.View()
	case ScreenSkills:
		return a.skillBrowser.View()
	default:
		return a.chat.View()
	}
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

func (a *App) handleSubmit(text string) tea.Cmd {
	return func() tea.Msg {
		req := api.CreateResponseRequest{
			Model: a.model,
			Input: api.MakeStringInput(text),
			Reasoning: &api.Reasoning{
				Effort: "medium",
			},
			Stream: true,
		}

		ctx := context.Background()
		chunks, errs := a.client.Stream(ctx, req)

		// Process stream in a goroutine, emitting tea messages
		go func() {
			for chunk := range chunks {
				switch chunk.Type {
				case "response.output_text.delta":
					// We can't directly send tea messages from here;
					// the real integration uses tea.Program.Send()
					_ = chunk.Delta
				}
			}
			// Check for errors
			for err := range errs {
				_ = err
			}
		}()

		return nil
	}
}
