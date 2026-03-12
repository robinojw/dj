package components

import (
	"testing"

	"github.com/robinojw/dj/internal/tui/theme"
)

func TestDebugOverlay_InitialState(t *testing.T) {
	d := NewDebugOverlay(theme.DefaultTheme())

	if d.IsVisible() {
		t.Error("should be hidden initially")
	}
	if len(d.entries) != 0 {
		t.Errorf("expected no entries, got %d", len(d.entries))
	}
}

func TestDebugOverlay_Toggle(t *testing.T) {
	d := NewDebugOverlay(theme.DefaultTheme())

	d.Toggle()
	if !d.IsVisible() {
		t.Error("should be visible after first toggle")
	}

	d.Toggle()
	if d.IsVisible() {
		t.Error("should be hidden after second toggle")
	}
}

func TestDebugOverlay_AddEntries(t *testing.T) {
	d := NewDebugOverlay(theme.DefaultTheme())

	d.AddError("something broke")
	d.AddWarn("something concerning")
	d.AddInfo("something normal")

	if len(d.entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(d.entries))
	}

	if d.entries[0].Level != "ERROR" {
		t.Errorf("entry[0].Level = %q, want ERROR", d.entries[0].Level)
	}
	if d.entries[0].Message != "something broke" {
		t.Errorf("entry[0].Message = %q, want %q", d.entries[0].Message, "something broke")
	}

	if d.entries[1].Level != "WARN" {
		t.Errorf("entry[1].Level = %q, want WARN", d.entries[1].Level)
	}

	if d.entries[2].Level != "INFO" {
		t.Errorf("entry[2].Level = %q, want INFO", d.entries[2].Level)
	}
}

func TestDebugOverlay_MaxEntries(t *testing.T) {
	d := NewDebugOverlay(theme.DefaultTheme())

	// Add more than maxDebugEntries (50).
	for i := 0; i < 60; i++ {
		d.AddInfo("entry")
	}

	if len(d.entries) != maxDebugEntries {
		t.Errorf("entries = %d, want %d (max)", len(d.entries), maxDebugEntries)
	}
}

func TestDebugOverlay_Clear(t *testing.T) {
	d := NewDebugOverlay(theme.DefaultTheme())
	d.AddError("error 1")
	d.AddInfo("info 1")

	d.Clear()

	if len(d.entries) != 0 {
		t.Errorf("expected 0 entries after clear, got %d", len(d.entries))
	}
}

func TestDebugOverlay_SetSize(t *testing.T) {
	d := NewDebugOverlay(theme.DefaultTheme())
	d.SetSize(120, 40)

	if d.width != 120 || d.height != 40 {
		t.Errorf("size = %dx%d, want 120x40", d.width, d.height)
	}
}

func TestDebugOverlay_View_WhenHidden(t *testing.T) {
	d := NewDebugOverlay(theme.DefaultTheme())
	d.SetSize(80, 24)
	d.AddError("should not render")

	v := d.View()
	if v != "" {
		t.Errorf("View() should return empty when hidden, got %q", v)
	}
}

func TestDebugOverlay_View_WhenVisibleNoEntries(t *testing.T) {
	d := NewDebugOverlay(theme.DefaultTheme())
	d.SetSize(80, 24)
	d.Toggle()

	v := d.View()
	if v == "" {
		t.Error("View() returned empty string when visible")
	}
}

func TestDebugOverlay_View_WhenVisibleWithEntries(t *testing.T) {
	d := NewDebugOverlay(theme.DefaultTheme())
	d.SetSize(80, 24)
	d.Toggle()

	d.AddError("test error")
	d.AddWarn("test warning")
	d.AddInfo("test info")

	v := d.View()
	if v == "" {
		t.Error("View() returned empty string with entries")
	}
}

func TestDebugOverlay_View_ZeroSize(t *testing.T) {
	d := NewDebugOverlay(theme.DefaultTheme())
	d.Toggle()
	// Don't set size — width and height are 0.

	v := d.View()
	if v != "" {
		t.Error("View() should return empty with zero dimensions")
	}
}

func TestDebugOverlay_EntryTimestamps(t *testing.T) {
	d := NewDebugOverlay(theme.DefaultTheme())
	d.AddInfo("first")
	d.AddInfo("second")

	if d.entries[0].Time.After(d.entries[1].Time) {
		t.Error("first entry should have earlier or equal timestamp")
	}
}
