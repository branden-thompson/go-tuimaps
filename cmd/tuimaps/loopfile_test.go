package main

// loopfile_test.go — v0.2.0 M1 (L10.9) and the BUILD-exit red team's
// accessibility findings: a recorded loop plays, the non-visual path says
// its motion with no place named, and Esc closes a panel before it quits.

import (
	"bytes"
	"strings"
	"testing"
)

const outbreak = "../../testdata/loops/outbreak-alabama-2011-04-27"

// TestARecordedLoopLoads: the loop file's frames come in oldest first, as
// a radar overlay read with IEM's table.
func TestARecordedLoopLoads(t *testing.T) {
	o, err := loadLoop(outbreak)
	if err != nil {
		t.Fatal(err)
	}
	if o.ID != loopID || o.Image == nil || len(o.Image.Frames) != 12 {
		t.Fatalf("the loop loaded as %q with %d frames; want %q with 12", o.ID, len(o.Image.Frames), loopID)
	}
	if !o.Image.Frames[0].Valid.Before(o.Image.Frames[11].Valid) {
		t.Error("the frames are not oldest first")
	}
	if _, err := loadLoop(t.TempDir()); err == nil {
		t.Error("a directory with no loop file loaded")
	}
}

// TestTheMotionIsSaidWithNoPlace (A11y 3.2): --describe over a recorded
// outbreak says where the heavier rain moved, measured from the view, with
// no place named - the release's answer for a listener who cannot watch.
func TestTheMotionIsSaidWithNoPlace(t *testing.T) {
	var out, errs bytes.Buffer
	if code := run([]string{"--describe", "--offline", "--size", "149x38", "--loop", outbreak}, &out, &errs); code != exitFine {
		t.Fatalf("exit %d: %s", code, errs.String())
	}
	said := out.String()
	if !strings.Contains(said, "heavier rain from the view's centre") || !strings.Contains(said, "minutes") {
		t.Errorf("the description says no motion:\n%s", said)
	}
}

// TestEscapeClosesAPanelBeforeItQuits (A11y 2.2): with a panel open, Esc
// closes it; with none, it quits. q always quits.
func TestEscapeClosesAPanelBeforeItQuits(t *testing.T) {
	a := &app{describeUp: true}
	if redraw, done := a.act(keyEscape); done || !redraw || a.describeUp {
		t.Fatalf("Esc with the describe panel open: redraw %v, done %v, panel %v; want it closed and the app running", redraw, done, a.describeUp)
	}
	if _, done := a.act(keyEscape); !done {
		t.Error("Esc with no panel open did not quit")
	}
}

// TestALongPanelSaysWhatItCannotShow (A11y 2.1): a panel longer than the
// map's rows ends with how many lines are not shown, and no line is wider
// than the terminal.
func TestALongPanelSaysWhatItCannotShow(t *testing.T) {
	var lines []string
	for range 30 {
		lines = append(lines, strings.Repeat("words ", 30))
	}
	got := fitted(lines, 10, 69)
	if len(got) != 10 || !strings.Contains(got[9], "21 more lines") {
		t.Fatalf("30 lines on 10 rows: %d rows, the last %q; want 10, the last naming 21 more", len(got), got[len(got)-1])
	}
	for _, l := range got {
		if len([]rune(l)) > 69 {
			t.Errorf("a line is %d wide on a 69-column terminal", len([]rune(l)))
		}
	}
}
