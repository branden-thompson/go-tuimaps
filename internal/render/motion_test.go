package render

import (
	"testing"
	"time"
)

var start = time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)

// TestBlinkPhaseFromTime is plan task 09.13 (FR-25, P-59a): the phase is
// worked out from the time the host supplies and never from a count of
// calls, and a host sampling at 100, 300 or 1,000 ms sees both phases.
func TestBlinkPhaseFromTime(t *testing.T) {
	for _, step := range []time.Duration{100 * time.Millisecond, 300 * time.Millisecond, time.Second} {
		var m Motion
		on, off := 0, 0
		for i := range 40 {
			if m.Phase(start.Add(time.Duration(i) * step)) {
				on++
			} else {
				off++
			}
		}
		if on == 0 || off == 0 {
			t.Errorf("sampling every %v over %v: %d frames on and %d off; a host must see a marker blink", step, 40*step, on, off)
		}
	}
	// However many extra calls are made at instants already seen, the phase
	// at each instant that does move the clock on is the same.
	var once, again Motion
	for i := range 24 {
		at := start.Add(time.Duration(i) * 150 * time.Millisecond)
		want := once.Phase(at)
		got := again.Phase(at)
		again.Phase(at) // called again at the same instant: the animation does not move
		again.Phase(at)
		if got != want {
			t.Fatalf("at step %d the phase was %v with one call an instant and %v with three; the phase is the time's, not the call's", i, want, got)
		}
	}
}

// TestBlinkKeepsUpstreamsRate is parity row P-59a: upstream blinks on eight
// ticks of its 50 ms loop and off for eight - an 800 ms period - and starts
// on. Nothing in the library flashes faster than 2.5 times a second (D-56).
func TestParityP59a_MarkerAnimation(t *testing.T) {
	if BlinkPeriod != 800*time.Millisecond {
		t.Errorf("the blink period is %v, want upstream's 800 ms", BlinkPeriod)
	}
	var m Motion
	for _, c := range []struct {
		at time.Duration
		on bool
	}{{0, true}, {399 * time.Millisecond, true}, {400 * time.Millisecond, false},
		{799 * time.Millisecond, false}, {800 * time.Millisecond, true}, {1200 * time.Millisecond, false}} {
		if got := m.Phase(start.Add(c.at)); got != c.on {
			t.Errorf("%v after it started, the marker is drawn %v, want %v", c.at, got, c.on)
		}
	}
	if 2*float64(time.Second)/float64(BlinkPeriod) > 2.5 {
		t.Error("the blink shows more than 2.5 changes a second")
	}
	var next Motion
	next.Phase(start)
	if got, want := next.Next(start), start.Add(400*time.Millisecond); !got.Equal(want) {
		t.Errorf("the next change is %v, want %v: half a period from the start", got, want)
	}
}

// TestFrozenClockDrawsOn is plan task 09.14 (NFR-21): a clock that does not
// advance never leaves a marker hidden - whether it never moved at all, or
// stopped in the half of the blink where the marker is not drawn.
func TestFrozenClockDrawsOn(t *testing.T) {
	var never Motion
	for range 10 {
		if !never.Phase(start) {
			t.Fatal("a clock that never advanced hid the marker")
		}
	}
	var stopped Motion
	stopped.Phase(start)
	if stopped.Phase(start.Add(500 * time.Millisecond)) {
		t.Fatal("500 ms in, the marker should be in the half of the blink it is not drawn")
	}
	for range 10 {
		if !stopped.Phase(start.Add(500 * time.Millisecond)) {
			t.Fatal("a clock stopped in the unlit half of a blink left the marker hidden")
		}
	}
	// A clock that goes backwards is treated the same way.
	if !stopped.Phase(start.Add(100 * time.Millisecond)) {
		t.Error("a clock that went backwards hid the marker")
	}
}

// TestReduceMotionSteady is the other half of 09.14 (NFR-21): with
// reduce-motion on, a marker is drawn steadily and no change is ever due.
func TestReduceMotionSteady(t *testing.T) {
	var m Motion
	m.Reduce(true)
	for i := range 20 {
		if !m.Phase(start.Add(time.Duration(i) * 137 * time.Millisecond)) {
			t.Fatalf("step %d: a marker was hidden with reduce-motion on", i)
		}
	}
	if got := m.Next(start.Add(time.Second)); !got.IsZero() {
		t.Errorf("the next change is %v; with reduce-motion on nothing is ever due", got)
	}
	// Switched off again, it blinks as before, from that moment.
	m.Reduce(false)
	m.Phase(start.Add(2 * time.Second))
	if m.Phase(start.Add(2500 * time.Millisecond)) {
		t.Error("with reduce-motion switched off the marker does not blink again")
	}
}
