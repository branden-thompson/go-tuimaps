package tuimaps

import (
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/overlay"
	"github.com/branden-thompson/go-tuimaps/internal/render"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// Playback is whether a map's loops may play. It is one setting for the map,
// not one a loop, and a refresh Set leaves it as it was (D-67). It is off
// until the host chooses.
type Playback uint8

// The playback settings.
const (
	PlaybackOff Playback = iota // loops show one frame; Play does nothing
	PlaybackOn                  // Play runs the loops, one frame a step
)

// OffReason is why playback is off. The zero value means it is not off.
type OffReason uint8

// The reasons playback is off, in the order the state read names them when
// there is more than one (L-1.10d).
const (
	OffReduceMotion OffReason = iota + 1 // the host asked for reduced motion
	OffByHost                            // the host set playback off
	OffByDefault                         // the host has not chosen: playback is off until it does
)

const (
	// defaultStep is how long each frame shows while a loop plays (D-76).
	defaultStep = 500 * time.Millisecond
	// fastestStep and slowestStep bound the step a host may set (D-76).
	fastestStep = 200 * time.Millisecond
	slowestStep = time.Second
	// ceilingStep is the fastest step that keeps every change on the map
	// within 2.5 a second (L-1.10g); a faster one lifts the ceiling.
	ceilingStep = 400 * time.Millisecond
	// lastHold is the least a loop holds its last frame before it repeats.
	lastHold = 2 * time.Second
)

// LoopState is what a host reads to show the listener where the loops are
// (L-1.10d). At is the moment shown and Now is "right now", the newest
// observed frame. Index counts from zero along every loop's frame times,
// merged; Oldest and Newest are its span.
type LoopState struct {
	At, Now, Oldest, Newest time.Time
	Index, Count            int
	// Playing is true while the host has pressed play and playback is on;
	// Advancing is true only while the shown frame is actually moving, so it
	// is false while the host holds the animation clock still.
	Playing, Advancing bool
	Gap, Forecast      bool // the moment shown is a gap, or a forecast
	CeilingLifted      bool // the step is faster than the 2.5-a-second ceiling allows (D-76)
	Playback           Playback
	Step               time.Duration
	Off                OffReason
}

// playback is a map's playback: the host's settings and the listener's
// place. The moment shown is a valid time, never a frame number, so a
// refresh keeps it (D-67).
type playback struct {
	setting  Playback
	chosen   bool          // the host has set playback, on or off
	step     time.Duration // zero means the default
	playing  bool
	from     time.Time // the animation time play began; zero until the map is next given one
	at       time.Time // the moment held while stopped; zero means "right now"
	given    time.Time // the animation time the map was last given
	previous time.Time // and the one before it
	aligned  bool      // the blink is on the loop's grid
	drawn    time.Time // the moment the last frame drew
}

// SetPlayback sets whether the map's loops may play. Off stops a loop that is
// playing, where it is.
func (m *Map) SetPlayback(p Playback) (err error) {
	defer guard("SetPlayback", &err)
	m.plant("SetPlayback")

	if m == nil {
		return closed()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return closed()
	}
	if p != PlaybackOff && p != PlaybackOn {
		return fault.Make(fault.OverLimit, textsafe.Const("playback was not set"), textsafe.Const("it is neither on nor off"), textsafe.Const("pass PlaybackOn or PlaybackOff"))
	}
	if p == PlaybackOff {
		m.holdLocked()
	}
	m.play.setting, m.play.chosen = p, true
	m.changed++
	return nil
}

// SetPlaybackStep sets how long each frame shows while a loop plays: from
// 200 ms to 1000 ms, zero meaning the default, 500 ms (D-76). A step below
// 400 ms lifts the ceiling of 2.5 changes a second, which is there for people
// that flashing harms; the state read says so, and a host that offers such a
// step should say so beside it.
func (m *Map) SetPlaybackStep(step time.Duration) (err error) {
	defer guard("SetPlaybackStep", &err)
	m.plant("SetPlaybackStep")

	if m == nil {
		return closed()
	}
	if step != 0 && (step < fastestStep || step > slowestStep) {
		return fault.Make(fault.OverLimit, textsafe.Const("the playback step was not set"),
			textsafe.Const("it is outside 200 ms to 1000 ms a frame"), textsafe.Const("pass a step from 200 ms to 1000 ms, or zero for the default of 500 ms"))
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return closed()
	}
	m.play.step = step
	m.play.aligned = false // the blink takes the new step at the next frame
	m.changed++
	return nil
}

// Play runs the loops from the oldest frame through "right now" and on
// through any forecast frames, holds the last frame and repeats (D-67). It
// does nothing while playback is off.
func (m *Map) Play() (err error) {
	return m.control("Play", func() {
		if !m.playsLocked() {
			return
		}
		m.play.playing, m.play.from, m.play.aligned = true, m.play.given, false
	})
}

// Stop holds the frame shown.
func (m *Map) Stop() (err error) {
	return m.control("Stop", m.holdLocked)
}

// Reset returns to "right now", the newest observed frame, and stops.
func (m *Map) Reset() (err error) {
	return m.control("Reset", func() {
		m.play.playing, m.play.at = false, time.Time{}
	})
}

// Step moves by frames along every loop's frame times, merged, and stops;
// it stops at the oldest and at the newest frame.
func (m *Map) Step(by int) (err error) {
	return m.control("Step", func() {
		timeline := m.store.Timeline()
		if len(timeline) == 0 {
			return
		}
		i := min(max(indexOf(timeline, m.shownLocked())+by, 0), len(timeline)-1)
		m.play.playing, m.play.at = false, timeline[i].Valid
	})
}

// control is the shape every listener control shares.
func (m *Map) control(name string, do func()) (err error) {
	defer guard(name, &err)
	m.plant(name)

	if m == nil {
		return closed()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return closed()
	}
	do()
	m.changed++
	return nil
}

// Loop is where the loops are, and why they are not playing if they are not.
func (m *Map) Loop() LoopState {
	defer m.guardQuiet("Loop")
	m.plant("Loop")

	if m == nil {
		return LoopState{}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return LoopState{}
	}
	st := LoopState{Playback: m.play.setting, Step: m.stepLocked(), Off: m.offLocked(), Now: m.store.RightNow()}
	st.CeilingLifted = st.Step < ceilingStep
	st.Playing = m.play.playing && st.Off == 0
	st.Advancing = st.Playing && m.play.given.After(m.play.previous)
	timeline := m.store.Timeline()
	st.Count = len(timeline)
	if st.Count == 0 {
		return st
	}
	st.Oldest, st.Newest = timeline[0].Valid, timeline[st.Count-1].Valid
	st.At = m.shownLocked()
	st.Index = indexOf(timeline, st.At)
	st.Gap, st.Forecast = timeline[st.Index].Gap, timeline[st.Index].Forecast
	return st
}

// playsLocked reports whether playback is in effect: on, and motion allowed.
func (m *Map) playsLocked() bool {
	return m.play.setting == PlaybackOn && !m.look.reduce
}

// offLocked is the first reason playback is off, or zero.
func (m *Map) offLocked() OffReason {
	switch {
	case m.look.reduce:
		return OffReduceMotion
	case m.play.setting == PlaybackOff && m.play.chosen:
		return OffByHost
	case m.play.setting == PlaybackOff:
		return OffByDefault
	}
	return 0
}

func (m *Map) stepLocked() time.Duration {
	if m.play.step == 0 {
		return defaultStep
	}
	return m.play.step
}

// holdLocked stops a loop that is playing on the frame it shows.
func (m *Map) holdLocked() {
	if m.play.playing {
		m.play.at = m.shownLocked()
	}
	m.play.playing = false
}

// giveLocked notes an animation time the map was given, by Render or
// Animate. A loop told to play before the map had one begins at this one.
func (m *Map) giveLocked(at time.Time) {
	m.play.previous, m.play.given = m.play.given, at
	if m.play.playing && m.play.from.IsZero() {
		m.play.from = at
	}
}

// blinkHalf is the blink's half-period while a loop plays: the step, or the
// smallest multiple of it that is at least half the blink's own period, so
// that every blink change falls on an advance (D-76).
func blinkHalf(step time.Duration) time.Duration {
	half := render.BlinkPeriod / 2
	return step * ((half + step - 1) / step)
}

// alignBlinkLocked puts the blink on the loop's grid while a loop plays, and
// gives it its own period back when none does.
func (m *Map) alignBlinkLocked() {
	if m.play.playing && m.playsLocked() && !m.play.from.IsZero() {
		if !m.play.aligned {
			m.motion.Align(m.play.from, blinkHalf(m.stepLocked()))
			m.play.aligned = true
		}
		return
	}
	if m.play.aligned {
		m.motion.Free()
		m.play.aligned = false
	}
}

// nextAdvanceLocked is when the frame shown changes next while a loop plays:
// the next step, or at the end of the last frame's hold, the start again.
func (m *Map) nextAdvanceLocked(wall time.Time) (time.Time, bool) {
	timeline := m.store.Timeline()
	if !m.play.playing || !m.playsLocked() || m.play.from.IsZero() || len(timeline) < 2 {
		return time.Time{}, false
	}
	if wall.Before(m.play.from) {
		return m.play.from, true
	}
	step := m.stepLocked()
	hold := step * ((lastHold + step - 1) / step)
	run := step * time.Duration(len(timeline)-1)
	elapsed := wall.Sub(m.play.from) % (run + hold)
	base := wall.Add(-elapsed)
	if elapsed < run {
		return base.Add((elapsed/step + 1) * step), true
	}
	return base.Add(run + hold), true
}

// shownLocked is the moment shown: while playing, a function of the
// animation time since play began, one frame a step, the last held at least
// two seconds, on whole steps (D-76); while stopped, the moment held - or,
// if a refresh dropped it, the nearest still held - or "right now".
func (m *Map) shownLocked() time.Time {
	timeline := m.store.Timeline()
	if len(timeline) == 0 {
		return time.Time{}
	}
	if !m.play.playing || !m.playsLocked() {
		if m.play.at.IsZero() {
			return m.store.RightNow()
		}
		return timeline[indexOf(timeline, m.play.at)].Valid // the same moment, or the nearest still held (D-67)
	}
	if m.play.from.IsZero() {
		return timeline[0].Valid
	}
	step := m.stepLocked()
	hold := step * ((lastHold + step - 1) / step)
	run := step * time.Duration(len(timeline)-1)
	elapsed := max(m.play.given.Sub(m.play.from), 0) % (run + hold)
	return timeline[min(int(elapsed/step), len(timeline)-1)].Valid
}

// indexOf is where a moment falls on the timeline: the last valid time at or
// before it, or the first when it is before them all.
func indexOf(timeline []overlay.Moment, at time.Time) int {
	i := 0
	for j, moment := range timeline {
		if !moment.Valid.After(at) {
			i = j
		}
	}
	return i
}
