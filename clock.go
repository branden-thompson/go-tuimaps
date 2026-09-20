package tuimaps

import (
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/overlay"
)

// Animate takes the animation clock over: from then on markers move on the
// time given here, and not on the wall clock given to Render (D-114). It is
// for a host that drives animation itself, or freezes it. A frozen animation
// clock cannot hide data that is out of date: staleness is the wall clock's.
func (m *Map) Animate(at time.Time) {
	defer m.guardQuiet("Animate")
	m.plant("Animate")

	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return
	}
	m.driven, m.animation = true, at
	m.changed++
}

// FollowClock gives the animation clock back to Render's own: markers move
// again on the wall clock the host passes it.
func (m *Map) FollowClock() {
	defer m.guardQuiet("FollowClock")
	m.plant("FollowClock")

	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut || !m.driven {
		return
	}
	m.driven = false
	m.changed++
}

// animationAt is the moment the markers are drawn at: the host's own if it
// has taken the clock over, and otherwise the wall clock of this frame.
func (m *Map) animationAt(wall time.Time) time.Time {
	if m.driven {
		return m.animation
	}
	return wall
}

// stale reports whether any overlay on the frame is out of date on the wall
// clock. Data whose valid time is far ahead of the clock counts, so that a
// bad timestamp cannot keep old data looking fresh (FR-32).
func (m *Map) stale(wall time.Time) bool {
	if wall.IsZero() {
		return false // a host that gives no time is told nothing about time
	}
	for _, id := range m.store.IDs() {
		reader, ok := m.store.Read(id)
		if !ok {
			continue
		}
		o := reader.Overlay()
		reader.Done()
		if overlay.FreshnessAt(o.Valid, o.Keeps, wall).DrawnStale() {
			return true
		}
	}
	return false
}

// NextCall is the wall-clock moment by which the map wants another Render:
// the earliest of the next marker phase, a failed tile's retry time and an
// overlay going stale (FR-25). It is false when nothing at all is due, and
// then a host may sleep until something happens to it instead.
func (m *Map) NextCall(wall time.Time) (time.Time, bool) {
	defer m.guardQuiet("NextCall")
	m.plant("NextCall")

	if m == nil {
		return time.Time{}, false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return time.Time{}, false
	}
	var due time.Time
	if !m.driven && m.blinking() {
		due = soonest(due, m.motion.Next(wall))
	}
	if !wall.IsZero() {
		for _, id := range m.store.IDs() {
			reader, ok := m.store.Read(id)
			if !ok {
				continue
			}
			o := reader.Overlay()
			reader.Done()
			if at, ok := overlay.NextChange(o.Valid, o.Keeps, wall); ok {
				due = soonest(due, at)
			}
		}
	}
	if at, ok := m.member.DueAt(wall); ok {
		due = soonest(due, at)
	}
	return due, !due.IsZero()
}

// blinking reports whether anything on the map is animated at all: with no
// blinking place there is no next phase to wake a host for.
func (m *Map) blinking() bool {
	for _, p := range m.places {
		if p.Blink {
			return true
		}
	}
	return false
}

// soonest is the earlier of two moments, a zero one meaning nothing is due.
func soonest(a, b time.Time) time.Time {
	if b.IsZero() {
		return a
	}
	if a.IsZero() || b.Before(a) {
		return b
	}
	return a
}
