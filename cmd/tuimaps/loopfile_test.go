package main

// loopfile_test.go — v0.2.0 M1 (L10.9) and the BUILD-exit red team's
// accessibility findings: a recorded loop plays, the non-visual path says
// its motion with no place named, and Esc closes a panel before it quits.

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
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
	if !strings.Contains(said, "heavier rain near the view's centre is moving ") || !strings.Contains(said, "kilometres an hour") || !strings.Contains(said, "the nearest is ") {
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

// TestAnOfflineLoopIsDrawnThroughThePump (M1 sitting, 2026-10-04): with
// --offline the tiles the view wants deeper than the built-in ones fail,
// and a failed tile is one job's failure, not the pump's. The pump keeps
// working, so the loop queued behind those tiles is prepared and drawn
// with no Settle first.
func TestAnOfflineLoopIsDrawnThroughThePump(t *testing.T) {
	m, err := build(settings{offline: true, noColour: true, cols: 150, rows: 40, loopDir: outbreak, language: "en"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, quit := context.WithCancel(context.Background())
	events := make(chan event, eventRoom)
	working, err := pump(ctx, m, events)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { quit(); working.Wait(); m.Close() })
	shaded := func() int {
		frame, err := m.Render(tuimaps.Size{Cols: 150, Rows: 38}, time.Now())
		if err != nil {
			t.Fatal(err)
		}
		n := 0
		for _, line := range frame.Lines {
			n += strings.Count(line, "░") + strings.Count(line, "▒") + strings.Count(line, "▓")
		}
		return n
	}
	shaded() // the first render asks for the work
	deadline := time.Now().Add(10 * time.Second)
	for shaded() == 0 {
		if time.Now().After(deadline) {
			t.Fatalf("no radar drawn after 10 s offline; %d jobs still pending", m.Pending())
		}
		select {
		case <-events:
		case <-time.After(50 * time.Millisecond):
		}
	}
}

// TestTheMotionWordsSayTheHeadingFirst (D-122, D-111): the words give the
// way the rain moves and how fast, then the nearest of it; rain over the
// place is said to be over it, not 0 km away; rain that barely moved has no
// heading; a loop with no motion says why.
func TestTheMotionWordsSayTheHeadingFirst(t *testing.T) {
	at := time.Date(2026, 5, 1, 20, 20, 0, 0, time.UTC)
	moving := tuimaps.MotionReport{Overlay: "radar", Place: "Home", Trend: tuimaps.Closer, Span: 55 * time.Minute,
		Moving: true, HeadingCompass: "north-east", SpeedKmh: 41,
		To: tuimaps.Sighting{Valid: at, Distance: 12, Unit: "kilometres", Compass: "south-west"}}
	for _, c := range []struct {
		mo   tuimaps.MotionReport
		want string
	}{
		{moving, "radar: heavier rain near Home is moving north-east at about 40 kilometres an hour, over the 55 minutes to 20:20 UTC; the nearest is 12 kilometres south-west, coming closer"},
		{func() tuimaps.MotionReport { m := moving; m.To.Distance = 0.3; return m }(), "radar: heavier rain near Home is moving north-east at about 40 kilometres an hour, over the 55 minutes to 20:20 UTC; it is over Home now"},
		{func() tuimaps.MotionReport {
			m := moving
			m.Moving, m.HeadingCompass, m.Trend = false, "", tuimaps.Held
			return m
		}(), "radar: heavier rain near Home has barely moved over the 55 minutes to 20:20 UTC; the nearest is 12 kilometres south-west, holding its distance"},
		{tuimaps.MotionReport{Overlay: "radar", Missing: tuimaps.MotionNoHeavierRain}, "radar: no motion to tell near the view's centre: no heavier rain"},
		{func() tuimaps.MotionReport { m := moving; m.To.Distance = 1.2; return m }(), "radar: heavier rain near Home is moving north-east at about 40 kilometres an hour, over the 55 minutes to 20:20 UTC; the nearest is 1 kilometre south-west, coming closer"},
	} {
		if got := moved(c.mo); got != c.want {
			t.Errorf("said %q\n want %q", got, c.want)
		}
	}
}

// TestTheLoopIsSaidPlayingOnlyWhileItMoves (D-109): the status row says
// "playing" only while the frames advance; play pressed over a still
// picture is "held"; playback off says why; a forecast or gap frame says so,
// and the time names its zone.
func TestTheLoopIsSaidPlayingOnlyWhileItMoves(t *testing.T) {
	at := time.Date(2026, 5, 1, 20, 15, 0, 0, time.UTC)
	for _, c := range []struct {
		st   tuimaps.LoopState
		want string
	}{
		{tuimaps.LoopState{At: at, Index: 2, Count: 12, Playing: true, Advancing: true}, "20:15 UTC frame 3 of 12 playing"},
		{tuimaps.LoopState{At: at, Index: 2, Count: 12, Playing: true}, "20:15 UTC frame 3 of 12 held"},
		{tuimaps.LoopState{At: at, Index: 2, Count: 12}, "20:15 UTC frame 3 of 12 stopped"},
		{tuimaps.LoopState{At: at, Index: 11, Count: 12, Off: tuimaps.OffReduceMotion}, "20:15 UTC frame 12 of 12 off: reduced motion"},
		{tuimaps.LoopState{At: at, Index: 11, Count: 12, Forecast: true, Playing: true, Advancing: true}, "forecast 20:15 UTC frame 12 of 12 playing"},
		{tuimaps.LoopState{At: at, Index: 4, Count: 12, Gap: true}, "gap 20:15 UTC frame 5 of 12 stopped"},
	} {
		if got := loopWords(c.st); got != c.want {
			t.Errorf("said %q; want %q", got, c.want)
		}
	}
}
