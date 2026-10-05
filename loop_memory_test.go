package tuimaps_test

// loop_memory_test.go — v0.2.0 M4 (D-61, D-68; plan task L10.8): a 12-frame
// loop at dot resolution adds at most 1 MiB of heap over an empty map, and the
// heap holds within ±5 % while the loop plays. The play runs for
// TUIMAPS_M4_SOAK (a Go duration): the full gate runs five minutes, the soak
// an hour; an ordinary run measures the cost and plays one round.

import (
	"image/color"
	"os"
	"runtime"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// m4Soak names how long the loop plays while the heap is watched.
const m4Soak = "TUIMAPS_M4_SOAK"

// m4Line and m4Hold are M4's target: the loop's cost, and how far the heap
// may move while it plays.
const (
	m4Line = 1 << 20
	m4Hold = 0.05
)

// settledHeap is the live heap after collection: the median of three, so one
// collection's timing does not decide it.
func settledHeap() uint64 {
	var got []uint64
	for range 3 {
		runtime.GC()
		var s runtime.MemStats
		runtime.ReadMemStats(&s)
		got = append(got, s.HeapAlloc)
	}
	return median(got)
}

// playRound steps a loop once through its frames, drawing each at a wall
// clock.
func playRound(t *testing.T, m *tuimaps.Map, frames int, size tuimaps.Size, clock time.Time) {
	t.Helper()
	for range frames {
		if err := m.Step(1); err != nil {
			t.Fatal(err)
		}
		if _, err := m.Render(size, clock); err != nil {
			t.Fatal(err)
		}
	}
}

// TestALoopsHeapCostAndHold is M4: the loop's cost over an empty map, then
// the heap held while it plays.
func TestALoopsHeapCostAndHold(t *testing.T) {
	soak := time.Duration(0)
	if v := os.Getenv(m4Soak); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d < 0 {
			t.Fatalf("%s=%q is not a duration", m4Soak, v)
		}
		soak = d
	}
	size := tuimaps.Size{Cols: 149, Rows: 38}
	m := world(t, size.Cols, size.Rows)
	settle(t, m)
	if _, err := m.Render(size, noon); err != nil {
		t.Fatal(err)
	}
	empty := settledHeap()
	if _, err := m.Set(radarLoop(t, "radar", 12, 2*size.Cols, 4*size.Rows)); err != nil {
		t.Fatal(err)
	}
	settle(t, m)
	playRound(t, m, 12, size, noon)
	held := settledHeap()
	cost := int64(held) - int64(empty)
	t.Logf("M4: a 12-frame loop at dot resolution (%dx%d) adds %s over an empty map; the line is %s", 2*size.Cols, 4*size.Rows, asBytes(cost), asBytes(m4Line))
	if cost > m4Line {
		t.Errorf("the loop adds %s; M4's line is %s", asBytes(cost), asBytes(m4Line))
	}
	if soak == 0 {
		return
	}
	lo, hi := float64(held)*(1-m4Hold), float64(held)*(1+m4Hold)
	end := time.Now().Add(soak)
	next := time.Now().Add(time.Minute)
	worst, samples, refreshes := held, 0, 0
	goroutines := runtime.NumGoroutine()
	for time.Now().Before(end) {
		playRound(t, m, 12, size, noon.Add(time.Duration(refreshes)*5*time.Minute)) // the clock moves on with the data
		if time.Now().Before(next) {
			continue
		}
		next = next.Add(time.Minute)
		// A host refreshes its loop every few minutes: a new frame on, the
		// oldest off, eleven frames shared. The soak refreshes once a
		// minute, so what a refresh keeps and lets go is soaked too.
		refreshes++
		if _, err := m.Set(shiftedLoop(t, refreshes, 2*size.Cols, 4*size.Rows)); err != nil {
			t.Fatal(err)
		}
		settle(t, m)
		samples++
		if n := runtime.NumGoroutine(); n > goroutines {
			t.Fatalf("after %d refreshes %d goroutines run; %d did at the start", refreshes, n, goroutines)
		}
		now := settledHeap()
		if float64(now) < lo || float64(now) > hi {
			t.Fatalf("after %s of play the heap is %s; it held %s at the start, and M4 holds it within ±5 %%", soak-time.Until(end).Round(time.Second), asBytes(int64(now)), asBytes(int64(held)))
		}
		if now > worst {
			worst = now
		}
	}
	if samples == 0 {
		t.Fatalf("a soak of %s took no sample; a soak is at least a minute", soak)
	}
	t.Logf("M4: over %s of play, refreshed %d times, the heap held within ±5 %% (start %s, highest %s)", soak, refreshes, asBytes(int64(held)), asBytes(int64(worst)))
}

// shiftedLoop is radarLoop's twelve frames moved on k frames: what a host
// hands in when it refreshes, sharing all but k frames with the last.
func shiftedLoop(t testing.TB, k, w, h int) tuimaps.Overlay {
	t.Helper()
	newest := noon.Add(time.Duration(k) * 5 * time.Minute)
	var frames []tuimaps.LoopFrame
	for i := 11; i >= 0; i-- {
		frames = append(frames, tuimaps.LoopFrame{Valid: newest.Add(-time.Duration(i) * 5 * time.Minute), PNG: solidPNG(t, w, h, color.NRGBA{R: 200, A: 255})})
	}
	return tuimaps.RadarImage("radar", tuimaps.Image{Frames: frames, West: -90, South: 30, East: -80, North: 40, Projection: tuimaps.PlateCarree,
		Table: []tuimaps.TableEntry{{Colour: tuimaps.RGB{R: 200}, Value: 25}}, Exact: true}, newest)
}

// TestTwentyFourRegionFramesCost is D-68's re-measure, brought back as a
// number: a two-hour loop of region-sized frames, which D-68 let past D-61's
// budget, costs what this logs.
func TestTwentyFourRegionFramesCost(t *testing.T) {
	size := tuimaps.Size{Cols: 149, Rows: 38}
	m := world(t, size.Cols, size.Rows)
	if err := m.SetImageBudget(16 << 20); err != nil {
		t.Fatal(err)
	}
	settle(t, m)
	if _, err := m.Render(size, noon); err != nil {
		t.Fatal(err)
	}
	empty := settledHeap()
	if _, err := m.Set(radarLoop(t, "radar", 24, 600, 400)); err != nil {
		t.Fatal(err)
	}
	settle(t, m)
	playRound(t, m, 24, size, noon)
	cost := int64(settledHeap()) - int64(empty)
	if cost <= 0 {
		t.Fatalf("24 region frames measured %s over an empty map: the measure is not reading the loop", asBytes(cost))
	}
	t.Logf("M4 (D-68): 24 region frames (600x400) add %s over an empty map", asBytes(cost))
}
