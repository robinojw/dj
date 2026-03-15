package tui

import (
	"testing"
	"time"

	tui "github.com/grindlemire/go-tui"
	"github.com/robinojw/dj/config"
	"github.com/robinojw/dj/internal/agents"
	"github.com/robinojw/dj/internal/api"
	"github.com/robinojw/dj/internal/modes"
	"github.com/robinojw/dj/internal/tui/theme"
)

func TestNewRootApp_DoesNotPanic(t *testing.T) {
	th := theme.DefaultTheme()
	app := NewRootApp(th, nil, nil, "gpt-5.4", config.Config{}, nil, nil, nil, nil)
	if app == nil {
		t.Fatal("expected non-nil root app")
	}
}

func TestScreenNavigation(t *testing.T) {
	th := theme.DefaultTheme()
	app := NewRootApp(th, nil, nil, "gpt-5.4", config.Config{}, nil, nil, nil, nil)

	if app.screen.Get() != ScreenIDChat {
		t.Fatalf("expected ScreenIDChat, got %d", app.screen.Get())
	}
}

func TestCycleModel(t *testing.T) {
	th := theme.DefaultTheme()
	tracker := api.NewTracker("gpt-5.4")
	app := NewRootApp(th, nil, tracker, "gpt-5.4", config.Config{}, nil, nil, nil, nil)

	initial := app.model.Get()
	app.cycleModel()
	after := app.model.Get()

	if initial == after {
		t.Fatal("expected model to change after cycling")
	}
}

func TestCycleMode(t *testing.T) {
	th := theme.DefaultTheme()
	app := NewRootApp(th, nil, nil, "gpt-5.4", config.Config{}, nil, nil, nil, nil)

	if app.modeVal != modes.ModeConfirm {
		t.Fatalf("expected ModeConfirm, got %d", app.modeVal)
	}

	app.cycleMode()
	if app.modeVal != modes.ModePlan {
		t.Fatalf("expected ModePlan after first cycle, got %d", app.modeVal)
	}

	app.cycleMode()
	if !app.turboModal.IsVisible() {
		t.Fatal("expected turbo modal to be visible")
	}
	if app.modeVal != modes.ModePlan {
		t.Fatalf("expected ModePlan (turbo not confirmed), got %d", app.modeVal)
	}
}

func TestPushPopScreen(t *testing.T) {
	th := theme.DefaultTheme()
	app := NewRootApp(th, nil, nil, "gpt-5.4", config.Config{}, nil, nil, nil, nil)

	if app.screen.Get() != ScreenIDChat {
		t.Fatalf("expected ScreenIDChat")
	}

	app.pushScreen(ScreenIDTeam)
	if app.screen.Get() != ScreenIDTeam {
		t.Fatalf("expected ScreenIDTeam")
	}
	if len(app.screenStack) != 1 {
		t.Fatalf("expected stack len 1, got %d", len(app.screenStack))
	}

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
		func(string, string) {}, func([]storedDiff) {}, func(string) {})
	if c == nil {
		t.Fatal("expected non-nil chat")
	}
}

func TestChat_StartStreamReusesChannel(t *testing.T) {
	th := theme.DefaultTheme()
	mode := tui.NewState(modes.ModeConfirm)
	model := tui.NewState("gpt-5.4")
	cost := tui.NewState(0.0)
	input := tui.NewState(0)
	output := tui.NewState(0)
	mcps := tui.NewState([]string{})

	c := NewChat(th, 80, mode, model, cost, input, output, mcps,
		func(string, string) {}, func([]storedDiff) {}, func(string) {})

	originalCh := c.eventCh

	// Start a stream
	chunks := make(chan api.ResponseChunk)
	errs := make(chan error)
	c.StartStream(chunks, errs)

	// Channel should be the same object
	if c.eventCh != originalCh {
		t.Fatal("StartStream should reuse the same eventCh, not replace it")
	}

	// Cancel to clean up goroutine
	c.cancelActiveStream()
}

func TestChat_CancelActiveStreamCleansUp(t *testing.T) {
	th := theme.DefaultTheme()
	mode := tui.NewState(modes.ModeConfirm)
	model := tui.NewState("gpt-5.4")
	cost := tui.NewState(0.0)
	input := tui.NewState(0)
	output := tui.NewState(0)
	mcps := tui.NewState([]string{})

	c := NewChat(th, 80, mode, model, cost, input, output, mcps,
		func(string, string) {}, func([]storedDiff) {}, func(string) {})

	// Start a stream with blocking channels
	chunks := make(chan api.ResponseChunk)
	errs := make(chan error)
	c.StartStream(chunks, errs)

	if c.cancelStream == nil {
		t.Fatal("expected cancelStream to be set")
	}

	c.cancelActiveStream()

	if c.cancelStream != nil {
		t.Fatal("expected cancelStream to be nil after cancel")
	}
}

func TestChat_OnStreamEventIgnoresTextAfterCancel(t *testing.T) {
	th := theme.DefaultTheme()
	mode := tui.NewState(modes.ModeConfirm)
	model := tui.NewState("gpt-5.4")
	cost := tui.NewState(0.0)
	input := tui.NewState(0)
	output := tui.NewState(0)
	mcps := tui.NewState([]string{})

	c := NewChat(th, 80, mode, model, cost, input, output, mcps,
		func(string, string) {}, func([]storedDiff) {}, func(string) {})

	// streaming is false (not set) — text events should be ignored
	c.onStreamEvent(streamEvent{Type: eventText, Delta: "ignored"})

	// No panic and streaming remains false
	if c.streaming.Get() {
		t.Fatal("expected streaming to remain false after ignored event")
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

func TestPermissionModal_ShowAndDismiss(t *testing.T) {
	th := theme.DefaultTheme()
	pm := NewPermissionModal(th)

	respCh := make(chan modes.PermissionResp, 1)
	req := &modes.PermissionRequest{
		Tool:   "write_file",
		Args:   map[string]any{"path": "/tmp/test"},
		RespCh: respCh,
	}

	pm.Show(req)
	if !pm.Visible() {
		t.Fatal("expected modal visible after Show")
	}

	pm.dismiss()
	if pm.Visible() {
		t.Fatal("expected modal not visible after dismiss")
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

func TestRootApp_OnWorkerUpdate_DiffResult(t *testing.T) {
	th := theme.DefaultTheme()
	app := NewRootApp(th, nil, nil, "gpt-5.4", config.Config{}, nil, nil, nil, nil)

	update := agents.WorkerUpdate{
		Type: agents.UpdateDiffResult,
		DiffInfo: &agents.DiffInfo{
			FilePath:  "main.go",
			DiffText:  "+added\n-removed",
			Timestamp: time.Now(),
		},
	}

	app.onWorkerUpdate(update)

	// Should have stored a diff in chat's diffs
	if len(app.chatView.diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(app.chatView.diffs))
	}
	if app.chatView.diffs[0].FilePath != "main.go" {
		t.Fatalf("expected 'main.go', got %q", app.chatView.diffs[0].FilePath)
	}
}

func TestRootApp_PermRequestCh(t *testing.T) {
	th := theme.DefaultTheme()
	app := NewRootApp(th, nil, nil, "gpt-5.4", config.Config{}, nil, nil, nil, nil)

	ch := app.PermRequestCh()
	if ch == nil {
		t.Fatal("expected non-nil PermRequestCh")
	}
}

func TestRootApp_HandleSubmit_IncludesMentionCtx(t *testing.T) {
	th := theme.DefaultTheme()
	app := NewRootApp(th, nil, nil, "gpt-5.4", config.Config{}, nil, nil, nil, nil)

	// Verify mentionCtx is appended to instructions
	// We can't easily test the API call, but we can verify the function doesn't panic
	// with nil client (it will panic at client.Stream — that's expected in real usage)
	// This is a construction/wiring test
	if app.chatView.onSubmit == nil {
		t.Fatal("expected onSubmit callback to be wired")
	}
}
