package tuimaps_test

// playback_test.go — v0.2.0 WP-L4: one playback per map, driven by the
// listener's controls (D-67), at a step the host sets (D-76).

import (
	"context"
	"fmt"
	"image/color"
	"strings"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// loopAt is a loop whose frames are valid at the given minutes after noon;
// those listed in forecast are forecast frames, and those in gaps are gaps.
func loopAt(t *testing.T, id string, minutes []int, forecast, gaps map[int]bool) tuimaps.Overlay {
	t.Helper()
	var frames []tuimaps.LoopFrame
	for _, min := range minutes {
		f := tuimaps.LoopFrame{Valid: noon.Add(time.Duration(min) * time.Minute), Forecast: forecast[min], Gap: gaps[min]}
		if !f.Gap {
			f.PNG = solidPNG(t, 8, 6, color.NRGBA{R: 200, A: 255})
		}
		frames = append(frames, f)
	}
	return tuimaps.RadarImage(id, tuimaps.Image{Frames: frames, West: -90, South: 30, East: -80, North: 40, Projection: tuimaps.PlateCarree,
		Table: []tuimaps.TableEntry{{Colour: tuimaps.RGB{R: 200}, Value: 25}}, Exact: true}, noon)
}

func at(minutes int) time.Time { return noon.Add(time.Duration(minutes) * time.Minute) }

func mustSet(t *testing.T, m *tuimaps.Map, o tuimaps.Overlay) {
	t.Helper()
	if _, err := m.Set(o); err != nil {
		t.Fatal(err)
	}
}

// settle runs the map's work until nothing is pending, so what was set is
// decoded and drawn.
func settle(t *testing.T, m *tuimaps.Map) {
	t.Helper()
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// TestPlaybackIsTheMapsAndPersists is L4.1 (L-1.5, L-1.11, D-67, D-76): a
// new map is off, by default, at 500 ms; the setting and the step survive a
// refresh; a step outside 200-1000 ms is refused, and zero is the default.
func TestPlaybackIsTheMapsAndPersists(t *testing.T) {
	m := world(t, 80, 24)
	st := m.Loop()
	if st.Playback != tuimaps.PlaybackOff || st.Off != tuimaps.OffByDefault || st.Step != 500*time.Millisecond {
		t.Errorf("a new map: %+v; want off by default at 500 ms", st)
	}
	must(t, m.SetPlayback(tuimaps.PlaybackOn))
	must(t, m.SetPlaybackStep(250*time.Millisecond))
	mustSet(t, m, loopAt(t, "radar", []int{-10, -5, 0}, nil, nil))
	mustSet(t, m, loopAt(t, "radar", []int{-5, 0}, nil, nil))
	if st := m.Loop(); st.Playback != tuimaps.PlaybackOn || st.Off != 0 || st.Step != 250*time.Millisecond {
		t.Errorf("after a refresh: %+v; want on at 250 ms", st)
	}
	for _, bad := range []time.Duration{199 * time.Millisecond, 1001 * time.Millisecond, -time.Second} {
		if err := m.SetPlaybackStep(bad); !isKind(err, fault.OverLimit) {
			t.Errorf("a step of %v: %v; want it refused", bad, err)
		}
	}
	must(t, m.SetPlaybackStep(0))
	if st := m.Loop(); st.Step != 500*time.Millisecond {
		t.Errorf("a step of zero: %v; want the default 500 ms", st.Step)
	}
	must(t, m.SetPlayback(tuimaps.PlaybackOff))
	if st := m.Loop(); st.Off != tuimaps.OffByHost {
		t.Errorf("turned off by the host: %v", st.Off)
	}
}

// TestTheListenersControls is L4.4 (L-1.10b, D-67): the map opens on "right
// now"; Play runs from the oldest frame in time order and on through the
// forecast, holds the last frame and repeats; Stop holds; Step moves by
// frames, stops, and stops at each end; Reset returns to "right now"; and a
// refresh keeps the moment on screen, or the nearest one still held.
func TestTheListenersControls(t *testing.T) {
	m := world(t, 80, 24)
	mustSet(t, m, loopAt(t, "radar", []int{-15, -10, -5, 0, 5, 10}, map[int]bool{5: true, 10: true}, nil))
	t0 := noon
	m.Animate(t0)
	st := m.Loop()
	if !st.At.Equal(at(0)) || !st.Now.Equal(at(0)) || st.Playing || st.Count != 6 || st.Index != 3 {
		t.Fatalf("on opening: %+v; want stopped on right now, frame 4 of 6", st)
	}
	if !st.Oldest.Equal(at(-15)) || !st.Newest.Equal(at(10)) {
		t.Errorf("the span: %v to %v", st.Oldest, st.Newest)
	}
	must(t, m.Play()) // playback is off: nothing moves
	if m.Animate(t0.Add(time.Second)); m.Loop().Playing || !m.Loop().At.Equal(at(0)) {
		t.Errorf("Play with playback off moved the loop: %+v", m.Loop())
	}
	must(t, m.SetPlayback(tuimaps.PlaybackOn))
	m.Animate(t0)
	must(t, m.Play())
	step := 500 * time.Millisecond
	hold := 2 * time.Second
	cycle := 5*step + hold
	for _, c := range []struct {
		after    time.Duration
		minute   int
		forecast bool
	}{
		{0, -15, false}, {step - 1, -15, false}, {step, -10, false}, {3 * step, 0, false},
		{4 * step, 5, true}, {5 * step, 10, true}, {5*step + hold - 1, 10, true}, {cycle, -15, false}, {cycle + step, -10, false},
	} {
		m.Animate(t0.Add(c.after))
		st := m.Loop()
		if !st.At.Equal(at(c.minute)) || st.Forecast != c.forecast || !st.Playing {
			t.Errorf("playing, %v in: %+v; want %d minutes, forecast %v", c.after, st, c.minute, c.forecast)
		}
	}
	m.Animate(t0.Add(step))
	must(t, m.Stop())
	m.Animate(t0.Add(time.Minute))
	if st := m.Loop(); st.Playing || !st.At.Equal(at(-10)) {
		t.Errorf("stopped: %+v; want held on -10", st)
	}
	must(t, m.Step(2))
	if st := m.Loop(); !st.At.Equal(at(0)) {
		t.Errorf("two steps on: %v", st.At)
	}
	must(t, m.Step(10))
	if st := m.Loop(); !st.At.Equal(at(10)) {
		t.Errorf("past the end: %v; want the newest frame", st.At)
	}
	must(t, m.Step(-10))
	if st := m.Loop(); !st.At.Equal(at(-15)) {
		t.Errorf("past the start: %v; want the oldest frame", st.At)
	}
	must(t, m.Play())
	must(t, m.Step(1))
	if st := m.Loop(); st.Playing {
		t.Error("Step did not stop playback")
	}
	must(t, m.Reset())
	if st := m.Loop(); !st.At.Equal(at(0)) || st.Playing {
		t.Errorf("reset: %+v; want stopped on right now", st)
	}
	// A refresh keeps the moment on screen.
	must(t, m.Step(-2))
	mustSet(t, m, loopAt(t, "radar", []int{-10, -5, 0, 5}, nil, nil))
	if st := m.Loop(); !st.At.Equal(at(-10)) {
		t.Errorf("a refresh that dropped the oldest frame moved the loop to %v; want -10 kept", st.At)
	}
	mustSet(t, m, loopAt(t, "radar", []int{-5, 0, 5, 10}, nil, nil))
	if st := m.Loop(); !st.At.Equal(at(-5)) {
		t.Errorf("a refresh that dropped the moment shown: %v; want the nearest still held, -5", st.At)
	}
}

// TestTheStateRead is L4.5 (L-1.10d): Playing only while the frame is
// actually advancing, Advancing false when the host has frozen the clock, and
// the first reason playback is off.
func TestTheStateRead(t *testing.T) {
	m := world(t, 80, 24)
	mustSet(t, m, loopAt(t, "radar", []int{-10, -5, 0}, nil, map[int]bool{-5: true}))
	must(t, m.SetPlayback(tuimaps.PlaybackOn))
	m.Animate(noon)
	must(t, m.Play())
	m.Animate(noon.Add(500 * time.Millisecond))
	st := m.Loop()
	if !st.Playing || !st.Advancing || !st.Gap || !st.At.Equal(at(-5)) {
		t.Errorf("playing onto a gap: %+v; want playing, advancing, at the gap's time", st)
	}
	m.Animate(noon.Add(500 * time.Millisecond)) // the same instant again: frozen
	if st := m.Loop(); !st.Playing || st.Advancing {
		t.Errorf("a frozen clock: %+v; want playing but not advancing", st)
	}
	m.ReduceMotion(true)
	if st := m.Loop(); st.Off != tuimaps.OffReduceMotion || st.Playing {
		t.Errorf("reduce motion over a host's on: %+v; want off because of reduce motion", st)
	}
	must(t, m.SetPlayback(tuimaps.PlaybackOff))
	if st := m.Loop(); st.Off != tuimaps.OffReduceMotion {
		t.Errorf("reduce motion and the host's off: %v; reduce motion is named first", st.Off)
	}
	m.ReduceMotion(false)
	if st := m.Loop(); st.Off != tuimaps.OffByHost {
		t.Errorf("the host's off alone: %v", st.Off)
	}
	must(t, m.SetPlayback(tuimaps.PlaybackOn))
	must(t, m.SetPlaybackStep(200*time.Millisecond))
	if st := m.Loop(); !st.CeilingLifted {
		t.Error("a 200 ms step: the state read does not say the ceiling is lifted")
	}
	must(t, m.SetPlaybackStep(400*time.Millisecond))
	if st := m.Loop(); st.CeilingLifted {
		t.Error("a 400 ms step keeps within the ceiling")
	}
}

// flickerLoop is a loop over a band of latitude that is all rain on odd
// frames and clear on even ones, so that every advance changes what is drawn.
func flickerLoop(t *testing.T, id string, minutes []int, south, north float64) tuimaps.Overlay {
	t.Helper()
	var frames []tuimaps.LoopFrame
	for i, min := range minutes {
		c := color.NRGBA{R: 200, A: 255}
		if i%2 == 0 {
			c = color.NRGBA{A: 0}
		}
		frames = append(frames, tuimaps.LoopFrame{Valid: noon.Add(time.Duration(min) * time.Minute), PNG: solidPNG(t, 8, 6, c)})
	}
	return tuimaps.RadarImage(id, tuimaps.Image{Frames: frames, West: -120, South: south, East: 120, North: north, Projection: tuimaps.PlateCarree,
		Table: []tuimaps.TableEntry{{Colour: tuimaps.RGB{R: 200}, Value: 45}}, Exact: true}, noon)
}

// changesOver renders the map every 10 ms of animation time for a span and
// returns the instants at which anything on it changed.
func changesOver(t *testing.T, m *tuimaps.Map, start time.Time, span time.Duration) []time.Duration {
	t.Helper()
	size := tuimaps.Size{Cols: 60, Rows: 18}
	var last string
	var changed []time.Duration
	for d := time.Duration(0); d <= span; d += 10 * time.Millisecond {
		m.Animate(start.Add(d))
		f, err := m.Render(size, noon)
		if err != nil {
			t.Fatal(err)
		}
		now := ""
		for _, line := range f.Lines {
			now += line + "\n"
		}
		if d > 0 && now != last {
			changed = append(changed, d)
		}
		last = now
	}
	return changed
}

// TestOneChangeCeilingPerMap is L4.2 (L-1.10g, D-76): two loops and a
// blinking place change on one grid. From a 400 ms step up nothing changes
// more than 2.5 times a second; at 200 ms it changes once a step, no more.
func TestOneChangeCeilingPerMap(t *testing.T) {
	for _, step := range []time.Duration{200, 400, 500, 600, 700, 800, 900, 1000} {
		step *= time.Millisecond
		m := world(t, 60, 18)
		mustSet(t, m, flickerLoop(t, "north", []int{-15, -10, -5, 0}, 20, 60))
		mustSet(t, m, flickerLoop(t, "south", []int{-12, -7, -2}, -50, -10))
		settle(t, m)
		if _, err := m.AddPlace(tuimaps.Place{ID: "home", Name: "Home", At: tuimaps.LonLat{Lon: 0, Lat: 10}, Blink: true}); err != nil {
			t.Fatal(err)
		}
		must(t, m.SetPlayback(tuimaps.PlaybackOn))
		must(t, m.SetPlaybackStep(step))
		start := noon
		m.Animate(start)
		must(t, m.Play())
		span := 12 * time.Second
		changed := changesOver(t, m, start, span)
		// The two loops cover separate bands, so every advance changes one of
		// them: six advances and the repeat in every cycle.
		cycle := 6*step + step*((2*time.Second+step-1)/step)
		if least := int(span/cycle) * 7; len(changed) < least {
			t.Fatalf("step %v: %d changes in %v, fewer than the %d advances, so the loops are not being drawn", step, len(changed), span, least)
		}
		for _, d := range changed {
			if d%step != 0 {
				t.Errorf("step %v: a change at %v is off the step's grid", step, d)
				break
			}
		}
		if step < 400*time.Millisecond {
			continue // the ceiling is lifted by the host's choice; on the grid, it is one change a step
		}
		for i := range changed {
			in := 0
			for _, d := range changed[i:] {
				if d-changed[i] < time.Second {
					in++
				}
			}
			if in > 3 {
				t.Errorf("step %v: %d changes in the second from %v", step, in, changed[i])
				break
			}
		}
		if rate := float64(len(changed)) / span.Seconds(); rate > 2.5 {
			t.Errorf("step %v: %.2f changes a second", step, rate)
		}
	}
}

// TestNextCallIncludesTheNextAdvance is L4.3 (L-1.8, D-66): with a loop
// playing, NextCall names the next advance, and after the last frame the end
// of the hold.
func TestNextCallIncludesTheNextAdvance(t *testing.T) {
	m := world(t, 80, 24)
	mustSet(t, m, loopAt(t, "radar", []int{-15, -10, -5, 0}, nil, nil))
	size := tuimaps.Size{Cols: 80, Rows: 24}
	if _, err := m.Render(size, noon); err != nil {
		t.Fatal(err)
	}
	if due, ok := m.NextCall(noon); ok && due.Before(noon.Add(time.Minute)) {
		t.Errorf("a loop that is not playing is due at %v", due)
	}
	must(t, m.SetPlayback(tuimaps.PlaybackOn))
	must(t, m.Play())
	if _, err := m.Render(size, noon); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ wall, due time.Duration }{
		{100 * time.Millisecond, 500 * time.Millisecond},
		{500 * time.Millisecond, time.Second},
		{1600 * time.Millisecond, 3500 * time.Millisecond}, // three steps run, then a two-second hold
	} {
		due, ok := m.NextCall(noon.Add(c.wall))
		if !ok || !due.Equal(noon.Add(c.due)) {
			t.Errorf("at %v: next call %v (%v); want %v", c.wall, due.Sub(noon), ok, c.due)
		}
	}
}

// TestReduceMotionForcesOffAndRestores is L4.6 (L-1.10c): on, then reduce
// motion: off because of it, the loop held where it was; reduce motion off:
// the host's on is back, and the loop stays held until the listener plays.
func TestReduceMotionForcesOffAndRestores(t *testing.T) {
	m := world(t, 80, 24)
	mustSet(t, m, loopAt(t, "radar", []int{-15, -10, -5, 0}, nil, nil))
	must(t, m.SetPlayback(tuimaps.PlaybackOn))
	m.Animate(noon)
	must(t, m.Play())
	m.Animate(noon.Add(500 * time.Millisecond))
	m.ReduceMotion(true)
	st := m.Loop()
	if st.Off != tuimaps.OffReduceMotion || st.Playing || !st.At.Equal(at(-10)) {
		t.Errorf("reduce motion while playing: %+v; want off because of it, held on -10", st)
	}
	must(t, m.Play())
	if m.Animate(noon.Add(3 * time.Second)); m.Loop().Playing {
		t.Error("Play under reduce motion played")
	}
	m.ReduceMotion(false)
	st = m.Loop()
	if st.Off != 0 || st.Playback != tuimaps.PlaybackOn || st.Playing || !st.At.Equal(at(-10)) {
		t.Errorf("reduce motion off: %+v; want the host's on back, still held on -10", st)
	}
	must(t, m.Play())
	m.Animate(noon.Add(4 * time.Second))
	if !m.Loop().Playing {
		t.Error("Play after reduce motion was turned off did not play")
	}
	// The host's off, while playing, holds the frame shown; on again does
	// not resume until the listener plays.
	m.Animate(noon.Add(4*time.Second + 500*time.Millisecond))
	shown := m.Loop().At
	must(t, m.SetPlayback(tuimaps.PlaybackOff))
	must(t, m.SetPlayback(tuimaps.PlaybackOn))
	m.Animate(noon.Add(6 * time.Second))
	if st := m.Loop(); st.Playing || !st.At.Equal(shown) {
		t.Errorf("off and on again while playing: %+v; want held on %v", st, shown.Sub(noon))
	}
}

// TestEveryAdvanceIsDrawn is the rest of L4.2: with no blink, the two loops'
// visible changes fall exactly where the timeline says. Merged, the frames
// are -15, -12, -10, -7, -5, -2 and 0 minutes, one a step, the last held at
// least 2 s on whole steps. -12 is the south loop's first frame, which it
// already showed as its nearest; every other advance flips one band between
// rain and clear. At 200 ms the blink's phase changes only every other step,
// so an advance between two phases is drawn only if the advance itself
// redraws the frame.
func TestEveryAdvanceIsDrawn(t *testing.T) {
	for _, c := range []struct {
		step time.Duration
		want []int // milliseconds
	}{
		{500, []int{1000, 1500, 2000, 2500, 3000, 5000, 6000, 6500, 7000}},
		{200, []int{400, 600, 800, 1000, 1200, 3200, 3600, 3800, 4000, 4200, 4400}},
	} {
		step := c.step * time.Millisecond
		m := world(t, 60, 18)
		mustSet(t, m, flickerLoop(t, "north", []int{-15, -10, -5, 0}, 20, 60))
		mustSet(t, m, flickerLoop(t, "south", []int{-12, -7, -2}, -50, -10))
		settle(t, m)
		must(t, m.SetPlayback(tuimaps.PlaybackOn))
		must(t, m.SetPlaybackStep(step))
		m.Animate(noon)
		must(t, m.Play())
		got := changesOver(t, m, noon, time.Duration(c.want[len(c.want)-1]+100)*time.Millisecond)
		var want []time.Duration
		for _, ms := range c.want {
			want = append(want, time.Duration(ms)*time.Millisecond)
		}
		if fmt.Sprint(got) != fmt.Sprint(want) {
			t.Errorf("step %v: changes at %v; want %v", step, got, want)
		}
	}
}

// TestAFrameAdvanceIsATickNotAChange is L4.7 (L-1.10e, D-66): ten frame
// advances move FrameTicks by ten, leave Changed alone, and leave the
// description where it was, remembered.
func TestAFrameAdvanceIsATickNotAChange(t *testing.T) {
	m := world(t, 60, 18)
	var minutes []int
	for i := -55; i <= 0; i += 5 {
		minutes = append(minutes, i)
	}
	mustSet(t, m, flickerLoop(t, "radar", minutes, -40, 60))
	settle(t, m)
	must(t, m.SetPlayback(tuimaps.PlaybackOn))
	size := tuimaps.Size{Cols: 60, Rows: 18}
	m.Animate(noon)
	if _, err := m.Render(size, noon); err != nil {
		t.Fatal(err)
	}
	before := m.FrameTicks()
	must(t, m.Play()) // the jump to the oldest frame is the listener's input, not an advance
	if _, err := m.Render(size, noon); err != nil {
		t.Fatal(err)
	}
	if m.FrameTicks() != before {
		t.Errorf("Play's jump to the oldest frame moved FrameTicks from %d to %d", before, m.FrameTicks())
	}
	places := []tuimaps.Place{{ID: "home", Name: "Home", At: tuimaps.LonLat{Lon: 0, Lat: 10}}}
	if _, err := m.Describe(places); err != nil {
		t.Fatal(err)
	}
	// The advances come from the animation clock at one wall time: a new
	// wall time is itself a reason to work a description out again (D-114),
	// and this is about the advance alone.
	changed, ticks := m.Changed(), m.FrameTicks()
	for i := 1; i <= 10; i++ {
		m.Animate(noon.Add(time.Duration(i) * 500 * time.Millisecond))
		if _, err := m.Render(size, noon); err != nil {
			t.Fatal(err)
		}
	}
	if got := m.FrameTicks() - ticks; got != 10 {
		t.Errorf("ten advances moved FrameTicks by %d", got)
	}
	if m.Changed() != changed {
		t.Errorf("ten advances moved Changed from %d to %d; an advance is not an input", changed, m.Changed())
	}
	if !tuimaps.DescriptionRemembered(m, places) {
		t.Error("an advance changed the description's key")
	}
	// A refresh while playing is an input too: the moment it moves the loop
	// to is not an advance.
	ticks = m.FrameTicks()
	mustSet(t, m, flickerLoop(t, "radar", minutes[3:], -40, 60))
	if _, err := m.Render(size, noon); err != nil {
		t.Fatal(err)
	}
	if m.FrameTicks() != ticks {
		t.Errorf("a refresh while playing moved FrameTicks from %d to %d", ticks, m.FrameTicks())
	}
	// A control is an input: it moves Changed, not FrameTicks.
	ticks = m.FrameTicks()
	must(t, m.Step(-1))
	m.Animate(noon.Add(6 * time.Second))
	if _, err := m.Render(size, noon); err != nil {
		t.Fatal(err)
	}
	if m.FrameTicks() != ticks {
		t.Errorf("a step moved FrameTicks from %d to %d", ticks, m.FrameTicks())
	}
}

// topRow renders the map and returns its top row as plain text.
func topRow(t *testing.T, m *tuimaps.Map, wall time.Time) string {
	t.Helper()
	f, err := m.Render(tuimaps.Size{Cols: 80, Rows: 24}, wall)
	if err != nil {
		t.Fatal(err)
	}
	return strings.ReplaceAll(plainText(f.Lines[0]), "\u2800", " ") // an empty cell is the braille blank
}

// TestTheFrameTimeIsOnTheMap is L4.8 (L-1.2, L-1.10a, L-1.10f, L-1.10g,
// D-67): the moment shown is text on the map beside the stale word; a gap
// reads "gap" with its own time; a forecast frame reads as one; with two
// loops, the newest loop's time; with no loop, nothing.
func TestTheFrameTimeIsOnTheMap(t *testing.T) {
	m := world(t, 80, 24)
	if row := topRow(t, m, noon); strings.Contains(row, ":") {
		t.Errorf("no loop, and the top row reads %q", row)
	}
	mustSet(t, m, loopAt(t, "radar", []int{-10, -5, 0, 5}, map[int]bool{5: true}, map[int]bool{-5: true}))
	settle(t, m)
	for _, c := range []struct {
		step int
		want string
	}{
		{0, "12:00"}, {-1, "gap 11:55"}, {-1, "11:50"}, {3, "forecast 12:05"},
	} {
		must(t, m.Step(c.step))
		row := topRow(t, m, noon)
		if !strings.HasSuffix(strings.TrimRight(row, " ⠀"), c.want) {
			t.Errorf("after a step of %d the top row ends %q; want %q", c.step, row[max(0, len(row)-30):], c.want)
		}
	}
	// Stale: the time sits to the left of the word.
	row := topRow(t, m, noon.Add(time.Hour))
	if !strings.Contains(row, "forecast 12:05 stale") {
		t.Errorf("stale, the top row reads %q; want the time beside the word", row)
	}
	// Two loops: the newest loop's time.
	two := world(t, 80, 24)
	mustSet(t, two, loopAt(t, "west", []int{-10, 0}, nil, nil))
	mustSet(t, two, loopAt(t, "east", []int{-12, -2}, nil, nil))
	settle(t, two)
	must(t, two.Step(-1)) // the merged timeline is -12, -10, -2, 0: one back from right now is -2
	if row := topRow(t, two, noon); !strings.Contains(row, "11:58") {
		t.Errorf("two loops at -2: the top row reads %q; want the newer loop's 11:58", row)
	}
}

// TestTheNewestObservedFrameDrivesStale is L4.9 (L-1.3, D-39, D-67): a loop
// is stale by its newest observed picture, never by the frame shown and never
// by a forecast.
func TestTheNewestObservedFrameDrivesStale(t *testing.T) {
	fresh := world(t, 80, 24)
	mustSet(t, fresh, loopAt(t, "radar", []int{-30, -20, -10, 0}, nil, nil))
	must(t, fresh.Step(-3))
	if row := topRow(t, fresh, noon.Add(10*time.Minute)); strings.Contains(row, "stale") {
		t.Errorf("stepped back to a frame 40 minutes old, with the newest 10: %q", row)
	}
	old := world(t, 80, 24)
	mustSet(t, old, loopAt(t, "radar", []int{-40, -30, -20}, nil, nil)) // handed in as valid at noon
	if row := topRow(t, old, noon.Add(10*time.Minute)); !strings.Contains(row, "stale") {
		t.Errorf("a loop whose newest picture is 30 minutes old, kept 15: %q; want stale", row)
	}
	forecast := world(t, 80, 24)
	mustSet(t, forecast, loopAt(t, "radar", []int{-40, -20, 0, 5}, map[int]bool{0: true, 5: true}, nil))
	if row := topRow(t, forecast, noon.Add(10*time.Minute)); !strings.Contains(row, "stale") {
		t.Errorf("observed up to -20 with forecast frames after: %q; a forecast never counts as newest", row)
	}
	if due, ok := old.NextCall(noon.Add(-10 * time.Minute)); !ok || !due.Before(noon) {
		t.Errorf("the loop goes stale at 11:35 by its newest picture; NextCall says %v", due)
	}
}

// TestOnlyAForecastMayBeInTheFuture is L4.9a (D-67): once the map has a wall
// clock, an observed frame dated past it by more than the skew allowed is
// refused, naming the frame; a forecast frame so dated is accepted.
func TestOnlyAForecastMayBeInTheFuture(t *testing.T) {
	m := world(t, 80, 24)
	if _, err := m.Render(tuimaps.Size{Cols: 80, Rows: 24}, noon); err != nil {
		t.Fatal(err)
	}
	_, err := m.Set(loopAt(t, "radar", []int{-5, 0, 10}, nil, nil))
	if !isKind(err, fault.ImageRefused) || !strings.Contains(err.Error(), "frame 3") {
		t.Errorf("an observed frame ten minutes ahead of the clock: %v; want it refused, naming frame 3", err)
	}
	if _, err := m.Set(loopAt(t, "radar", []int{-5, 0, 4}, nil, nil)); err != nil {
		t.Errorf("an observed frame four minutes ahead, inside the skew allowed: %v", err)
	}
	if _, err := m.Set(loopAt(t, "radar", []int{-5, 0, 10}, map[int]bool{10: true}, nil)); err != nil {
		t.Errorf("a forecast frame ten minutes ahead: %v", err)
	}
}
