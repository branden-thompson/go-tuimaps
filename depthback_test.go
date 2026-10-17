package tuimaps_test

// depthback_test.go — the BUILD exit's API review (watchpost D-207, D-216):
// a host offering "follow the terminal" beside the depths can return to it.

import (
	"strings"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// TestADepthHintIsTakenBack is the API review's TerminalDepth: on a terminal
// that says 256 colours and no COLORTERM, a map hinted truecolor draws
// truecolor, and hinted TerminalDepth draws as the terminal says again; a
// depth that is none of the five is no hint.
func TestADepthHintIsTakenBack(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("COLORTERM", "")
	t.Setenv("TERM", "xterm-256color")
	m := world(t, 40, 12)
	draw := func() string {
		t.Helper()
		f, err := m.Render(tuimaps.Size{Cols: 40, Rows: 12}, time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatal(err)
		}
		return strings.Join(f.Lines, "\n")
	}
	truecolor, palette := "38;2;", "38;5;"
	if f := draw(); strings.Contains(f, truecolor) || !strings.Contains(f, palette) {
		t.Fatal("control: with no hint the terminal's 256 colours are not the depth")
	}
	m.ColourDepth(tuimaps.Truecolor)
	if f := draw(); !strings.Contains(f, truecolor) {
		t.Fatal("hinted truecolor, the map does not draw truecolor")
	}
	m.ColourDepth(tuimaps.TerminalDepth + 1)
	if f := draw(); !strings.Contains(f, truecolor) {
		t.Error("a depth that is none of the five took the hint back")
	}
	m.ColourDepth(tuimaps.TerminalDepth)
	if f := draw(); strings.Contains(f, truecolor) || !strings.Contains(f, palette) {
		t.Error("TerminalDepth did not take the hint back: the map still draws truecolor")
	}
	if tuimaps.TerminalDepth.String() != "the terminal's depth" {
		t.Errorf("TerminalDepth is named %q", tuimaps.TerminalDepth.String())
	}
}
