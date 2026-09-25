package tuimaps_test

import (
	"context"
	"strings"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// TestRenderTakesTheWallClock is ruling D-114: the time given to Render is
// the wall clock. Data that is out of date is marked from it, and the
// marker phase follows it unless the host drives animation itself.
func TestRenderTakesTheWallClock(t *testing.T) {
	const cols, rows = 149, 38
	m := gulfMap(t, cols, rows)
	if _, err := m.Set(warning("alerts")); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	// Current data: no mark.
	fresh := frameAtTime(t, m, cols, rows, noon.Add(30*time.Minute))
	if strings.Contains(fresh, "stale") {
		t.Errorf("data still current is marked stale:\n%s", fresh)
	}
	// The same data, hours later on the wall clock: marked.
	old := frameAtTime(t, m, cols, rows, noon.Add(6*time.Hour))
	if !strings.Contains(old, "stale") {
		t.Errorf("data six hours past its currency is not marked:\n%s", old)
	}
	// And the mark goes when fresher data arrives.
	newer := warning("alerts")
	newer.Valid = noon.Add(6 * time.Hour)
	if _, err := m.Set(newer); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	if again := frameAtTime(t, m, cols, rows, noon.Add(6*time.Hour)); strings.Contains(again, "stale") {
		t.Error("fresher data did not take the mark away")
	}
}

// TestAnimateDrivesTheMarkerPhase is the other half of D-114: a host that
// drives animation itself has a clock of its own, and a frozen animation
// clock cannot hide data that is out of date.
func TestAnimateDrivesTheMarkerPhase(t *testing.T) {
	const cols, rows = 149, 38
	m := gulfMap(t, cols, rows)
	if _, err := m.SetPlaces([]tuimaps.Place{{Name: "Beacon", At: tuimaps.LonLat{Lon: -84, Lat: 26}, Blink: true}}); err != nil {
		t.Fatal(err)
	}
	// With animation following the wall clock, the marker blinks.
	on := frameAtTime(t, m, cols, rows, noon)
	off := frameAtTime(t, m, cols, rows, noon.Add(500*time.Millisecond))
	if on == off {
		t.Error("a blinking place does not follow the wall clock when nothing else drives it")
	}
	// Driven, the animation clock is the host's and the wall clock is not.
	m.Animate(noon)
	held := frameAtTime(t, m, cols, rows, noon.Add(500*time.Millisecond))
	again := frameAtTime(t, m, cols, rows, noon.Add(900*time.Millisecond))
	if held != again {
		t.Error("the marker moved on a clock the host had taken over")
	}
	// The frozen animation clock does not stop the data going out of date.
	if _, err := m.Set(warning("alerts")); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	if marked := frameAtTime(t, m, cols, rows, noon.Add(9*time.Hour)); !strings.Contains(marked, "stale") {
		t.Error("a frozen animation clock hid data that is out of date")
	}
}

// TestNextCall is plan task 12.20 (FR-25): the deadline a host may sleep
// until - the earliest of the next marker phase, a failed tile's retry and
// an overlay going stale - and nothing at all when nothing is due.
func TestNextCall(t *testing.T) {
	const cols, rows = 80, 24
	m := gulfMap(t, cols, rows)
	frameAtTime(t, m, cols, rows, noon)
	if due, ok := m.NextCall(noon); ok {
		t.Errorf("a map with nothing animated and nothing to go stale is due at %v", due)
	}
	// An overlay is due when it goes stale.
	if _, err := m.Set(warning("alerts")); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	frameAtTime(t, m, cols, rows, noon)
	due, ok := m.NextCall(noon)
	if !ok {
		t.Fatal("an overlay that will go stale is not due at all")
	}
	// The first instant at which it is stale, not the last at which it is
	// current: at exactly an hour the data still counts as current.
	if want := noon.Add(time.Hour + time.Nanosecond); !due.Equal(want) {
		t.Errorf("due at %v, want %v: the first instant its hour of currency has run out", due, want)
	}
	// A blinking place is due sooner: at the next half of its blink.
	if _, err := m.SetPlaces([]tuimaps.Place{{Name: "Beacon", At: tuimaps.LonLat{Lon: -84, Lat: 26}, Blink: true}}); err != nil {
		t.Fatal(err)
	}
	frameAtTime(t, m, cols, rows, noon)
	due, ok = m.NextCall(noon)
	if !ok || due.After(noon.Add(time.Second)) {
		t.Errorf("with a blinking place the map is due at %v; the next phase is within a second", due)
	}
	// With reduce-motion on, the blink is not what is due any more, but the
	// overlay's stale moment still is: reduce-motion stops motion, not the
	// clock's other work (v0.2.0 L10.2, L-13.4, ReduceMotion's comment).
	m.ReduceMotion(true)
	frameAtTime(t, m, cols, rows, noon)
	due, ok = m.NextCall(noon)
	if want := noon.Add(time.Hour + time.Nanosecond); !ok || !due.Equal(want) {
		t.Errorf("with reduce-motion on the map is due at %v (%v); nothing animates, so the stale moment %v is the deadline", due, ok, want)
	}
}

// frameAtTime is a frame drawn at a moment of the wall clock, as its own
// string with the colour sequences taken out.
func frameAtTime(t *testing.T, m *tuimaps.Map, cols, rows int, at time.Time) string {
	t.Helper()
	f, err := m.Render(tuimaps.Size{Cols: cols, Rows: rows}, at)
	if err != nil {
		t.Fatal(err)
	}
	return colours.ReplaceAllString(strings.Join(f.Lines, "\n"), "")
}
