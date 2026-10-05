package render

import "time"

// BlinkPeriod is how long a blinking marker takes to come back to where it
// started: upstream blinks for eight ticks of its 50 ms loop and rests for
// eight, which is 800 ms (P-59a). It is two changes a second, inside the
// two and a half a second that is the most anything here may flash (D-56).
const BlinkPeriod = 800 * time.Millisecond

// Motion turns the host's animation clock into a marker's phase (FR-25,
// NFR-21). It is the one piece of the renderer that remembers anything
// between frames, and it remembers only the clock: the phase is worked out
// from the time supplied and never from a count of calls, so a host that
// draws twice at one instant sees one frame twice.
//
// A clock that does not advance never leaves a marker hidden (NFR-21): an
// instant already seen, or one before it, is always drawn. It is the host's
// clock that has stopped, not the marker that has gone.
type Motion struct {
	start, seen time.Time
	going       bool
	reduce      bool
	half        time.Duration // how long each half of a blink lasts; zero means half of BlinkPeriod
}

// Align puts the blink on a loop's grid while it plays (v0.2.0 D-76): it
// starts again at the loop's start and changes every half, so that every
// change on the map falls on one grid.
func (m *Motion) Align(start time.Time, half time.Duration) {
	if m == nil || m.reduce || half <= 0 {
		return
	}
	m.start, m.seen, m.going, m.half = start, start, true, half
}

// Free gives the blink its own period back, from where it stands.
func (m *Motion) Free() {
	if m == nil {
		return
	}
	m.half = 0
}

func (m *Motion) halfPeriod() time.Duration {
	if m.half > 0 {
		return m.half
	}
	return BlinkPeriod / 2
}

// Reduce sets whether motion is reduced. With it on, markers are steady and
// nothing is ever due (NFR-21).
func (m *Motion) Reduce(on bool) {
	if m == nil {
		return
	}
	m.reduce = on
	if on {
		m.going = false // it starts again from the moment motion is allowed
	}
}

// Phase reports whether a blinking marker is drawn at the host's time.
func (m *Motion) Phase(now time.Time) bool {
	if m == nil || m.reduce {
		return true
	}
	if !m.going {
		m.start, m.seen, m.going = now, now, true
		return true
	}
	if !now.After(m.seen) {
		return true
	}
	m.seen = now
	return litHalf(now.Sub(m.start), m.halfPeriod())
}

// Next is when the phase changes next, or the zero time if it never will.
func (m *Motion) Next(now time.Time) time.Time {
	if m == nil || m.reduce || !m.going {
		return time.Time{}
	}
	half := m.halfPeriod()
	elapsed := now.Sub(m.start)
	if elapsed < 0 {
		return m.start
	}
	return m.start.Add((elapsed/half + 1) * half)
}

// litHalf is which half of a blink an elapsed time falls in.
func litHalf(elapsed, half time.Duration) bool {
	if elapsed < 0 {
		return true
	}
	return (elapsed/half)%2 == 0
}
