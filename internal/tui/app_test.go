package tui

import (
	"testing"

	tui "github.com/grindlemire/go-tui"
	"github.com/robinojw/dj/config"
	"github.com/robinojw/dj/internal/api"
	"github.com/robinojw/dj/internal/modes"
	"github.com/robinojw/dj/internal/tui/theme"
)

func TestNewRootApp_DoesNotPanic(t *testing.T) {
	th := theme.DefaultTheme()
	app := NewRootApp(th, nil, nil, "gpt-5.4", config.Config{}, nil, nil)
	if app == nil {
		t.Fatal("expected non-nil root app")
	}
}

func TestScreenNavigation(t *testing.T) {
	th := theme.DefaultTheme()
	app := NewRootApp(th, nil, nil, "gpt-5.4", config.Config{}, nil, nil)

	// Should start on chat
	if app.screen.Get() != ScreenIDChat {
		t.Fatalf("expected ScreenIDChat, got %d", app.screen.Get())
	}
}

func TestCycleModel(t *testing.T) {
	th := theme.DefaultTheme()
	tracker := api.NewTracker("gpt-5.4")
	app := NewRootApp(th, nil, tracker, "gpt-5.4", config.Config{}, nil, nil)

	initial := app.model.Get()
	app.cycleModel()
	after := app.model.Get()

	if initial == after {
		t.Fatal("expected model to change after cycling")
	}
}

func TestCycleMode(t *testing.T) {
	th := theme.DefaultTheme()
	app := NewRootApp(th, nil, nil, "gpt-5.4", config.Config{}, nil, nil)

	if app.modeVal != modes.ModeConfirm {
		t.Fatalf("expected ModeConfirm, got %d", app.modeVal)
	}

	// Cycling from Confirm should go to Plan
	app.cycleMode()
	if app.modeVal != modes.ModePlan {
		t.Fatalf("expected ModePlan after first cycle, got %d", app.modeVal)
	}

	// Cycling from Plan should trigger turbo modal (not confirmed yet)
	app.cycleMode()
	if !app.turboModal.IsVisible() {
		t.Fatal("expected turbo modal to be visible")
	}
	// Mode should still be Plan since turbo not confirmed
	if app.modeVal != modes.ModePlan {
		t.Fatalf("expected ModePlan (turbo not confirmed), got %d", app.modeVal)
	}
}

func TestPushPopScreen(t *testing.T) {
	th := theme.DefaultTheme()
	app := NewRootApp(th, nil, nil, "gpt-5.4", config.Config{}, nil, nil)

	// Start at chat
	if app.screen.Get() != ScreenIDChat {
		t.Fatalf("expected ScreenIDChat")
	}

	// Push team screen
	app.pushScreen(ScreenIDTeam)
	if app.screen.Get() != ScreenIDTeam {
		t.Fatalf("expected ScreenIDTeam")
	}
	if len(app.screenStack) != 1 {
		t.Fatalf("expected stack len 1, got %d", len(app.screenStack))
	}

	// Pop back to chat
	app.popScreenFn()
	if app.screen.Get() != ScreenIDChat {
		t.Fatalf("expected ScreenIDChat after pop")
	}
	if len(app.screenStack) != 0 {
		t.Fatalf("expected empty stack after pop")
	}
}

func TestNewChat_DoesNotPanic(t *testing.T) {
	th := theme.DefaultTheme()
	mode := tui.NewState(modes.ModeConfirm)
	model := tui.NewState("gpt-5.4")
	cost := tui.NewState(0.0)
	input := tui.NewState(0)
	output := tui.NewState(0)
	mcps := tui.NewState([]string{})

	c := NewChat(th, 80, mode, model, cost, input, output, mcps,
		func(string, string) {}, func([]storedDiff) {})
	if c == nil {
		t.Fatal("expected non-nil chat")
	}
}

func TestNewDiffPager_DoesNotPanic(t *testing.T) {
	th := theme.DefaultTheme()
	diffs := []storedDiff{
		{FilePath: "test.go", DiffLines: []string{"+added"}},
	}
	dp := NewDiffPager(th, diffs, func() {})
	if dp == nil {
		t.Fatal("expected non-nil diff pager")
	}
}

func TestNewPermissionModal_DoesNotPanic(t *testing.T) {
	th := theme.DefaultTheme()
	pm := NewPermissionModal(th)
	if pm == nil {
		t.Fatal("expected non-nil permission modal")
	}
	if pm.Visible() {
		t.Fatal("expected modal not visible initially")
	}
}

func TestDebugOverlay_AddAndToggle(t *testing.T) {
	th := theme.DefaultTheme()
	d := NewDebugOverlay(th)

	if d.IsVisible() {
		t.Fatal("expected not visible initially")
	}

	d.Toggle()
	if !d.IsVisible() {
		t.Fatal("expected visible after toggle")
	}

	d.AddError("test error")
	d.AddInfo("test info")
	d.AddWarn("test warn")

	entries := d.entries.Get()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].Level != "ERROR" {
		t.Fatalf("expected ERROR, got %s", entries[0].Level)
	}
}
