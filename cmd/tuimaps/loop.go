package main

import (
	"context"
	"io"
	"sync"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// app is the interactive map: the library's map, the terminal it is drawn
// on, and everything the keys have turned on and off.
type app struct {
	m    *tuimaps.Map
	out  io.Writer
	cols int
	rows int

	places []tuimaps.Place
	focus  int // which place is focused; none is -1

	labels  bool
	water   bool
	markers bool

	safeRamps    bool
	reduceMotion bool
	noColour     bool

	offline    bool
	helpUp     bool
	describeUp bool
	said       string // what went wrong, for the status row
}

// event is something the app has to answer: a key, a change of size, or the
// pump saying it has done some work.
type event struct {
	key        string
	resize     bool
	cols, rows int // the terminal's new size, where it changed
	worked     bool
}

// newApp is the app with the settings it was started with already in it, so
// that what a flag turned on and what a key turns on are one state.
func newApp(m *tuimaps.Map, s settings, out io.Writer, cols, rows int) *app {
	return &app{
		m: m, out: out, cols: cols, rows: rows,
		places: s.places, focus: -1, offline: s.offline,
		labels: true, water: true, markers: true,
		safeRamps: s.safeRamps, reduceMotion: s.reduceMotion, noColour: s.noColour,
	}
}

// note keeps what went wrong for the status row and says whether the call
// worked. Nothing a key does is worth ending the app for: a refused pan
// leaves the map where it was and says so.
func (a *app) note(err error) bool {
	if err == nil {
		a.said = ""
		return true
	}
	a.said = printable(err.Error(), wholeComplaint)
	return true // the status row changed, so the frame differs
}

// focusedPlace is the place the keys are working about, if any.
func (a *app) focusedPlace() (tuimaps.Place, bool) {
	if a.focus < 0 || a.focus >= len(a.places) {
		return tuimaps.Place{}, false
	}
	return a.places[a.focus], true
}

// focused is where the focused place is.
func (a *app) focused() (tuimaps.LonLat, bool) {
	place, ok := a.focusedPlace()
	return place.At, ok
}

// focusNext moves the focus to the next place, and past the last one back
// to none - so that a person can always get back to no focus at all.
func (a *app) focusNext() bool {
	if len(a.places) == 0 {
		a.said = "no place is named; --place names one, and --scenario loads some"
		return true
	}
	a.focus++
	if a.focus >= len(a.places) {
		a.focus = -1
	}
	a.said = ""
	return true
}

// showPlaces puts the places on the map or takes them off, which is what
// upstream's marker key does (P-68a).
func (a *app) showPlaces() error {
	if !a.markers {
		_, err := a.m.SetPlaces(nil)
		return err
	}
	_, err := a.m.SetPlaces(a.places)
	return err
}

// drawn writes one frame and the chrome around it. It is the only place the
// app writes to the terminal.
func (a *app) drawn(now time.Time) error {
	frame, err := a.m.Render(tuimaps.Size{Cols: a.cols, Rows: mapRows(a.rows)}, now)
	if err != nil {
		return err
	}
	return a.put(a.screenful(frame))
}

// put writes a screenful: home the cursor, then every row, each row ending
// where the terminal would end it anyway.
func (a *app) put(rows []string) error {
	var b []byte
	b = append(b, "\x1b[H"...) // the top left corner
	for i, row := range rows {
		b = append(b, "\x1b[2K"...) // this row, cleared, so nothing is left behind
		b = append(b, row...)
		if i+1 < len(rows) {
			b = append(b, '\r', '\n')
		}
	}
	_, err := a.out.Write(b)
	return err
}

// resized takes the terminal's new size. The map is asked for the new
// rectangle at the next frame; nothing is re-fetched for it.
func (a *app) resized(cols, rows int) bool {
	if cols < 1 || rows < 1 || (cols == a.cols && rows == a.rows) {
		return false
	}
	a.cols, a.rows = cols, rows
	return true
}

// pump is the work the library never does for itself (D-73): two goroutines
// calling Work, woken by the library's own hook, each unit done telling the
// loop that there is more of the map to see.
func pump(ctx context.Context, m *tuimaps.Map, events chan<- event) (*sync.WaitGroup, error) {
	wake := make(chan struct{}, 1)
	nudge := func() {
		select {
		case wake <- struct{}{}:
		default: // already awake
		}
	}
	if err := m.OnPending(nudge); err != nil {
		return nil, err
	}
	var running sync.WaitGroup
	for range 2 {
		running.Add(1)
		go func() {
			defer running.Done()
			for {
				did, err := m.Work(ctx)
				if err != nil {
					return
				}
				if did {
					select {
					case events <- event{worked: true}:
					case <-ctx.Done():
						return
					}
					continue // there may be more
				}
				select {
				case <-ctx.Done():
					return
				case <-wake:
				}
			}
		}()
	}
	nudge() // whatever is already waiting
	return &running, nil
}

// clock is how the loop tells the time and how it waits. A test hands in
// one of its own, so that the loop can be shown to sleep until the map says
// to call again rather than polling (P-71, L-15).
type clock struct {
	now   func() time.Time
	after func(time.Duration) <-chan time.Time
}

// realClock is the wall clock, which is what an interactive host runs on
// (D-114).
func realClock() clock {
	return clock{now: time.Now, after: time.After}
}

// loop draws, and draws again whenever something would make the frame
// differ: a key, a change of size, a unit of work done, or the moment the
// map itself said to call back by. **It never polls** - with nothing
// animated and nothing going stale it waits on the events alone, which is
// this release's fix to upstream's fifty-millisecond poll (P-71, L-15).
func (a *app) loop(events <-chan event, tick clock) error {
	if err := a.drawn(tick.now()); err != nil {
		return err
	}
	for {
		var due <-chan time.Time
		if at, ok := a.m.NextCall(tick.now()); ok {
			due = tick.after(max(0, at.Sub(tick.now())))
		}
		select {
		case e, open := <-events:
			if !open {
				return nil
			}
			redraw, done := a.answer(e)
			if done {
				return nil
			}
			if !redraw {
				continue
			}
		case <-due:
		}
		if err := a.drawn(tick.now()); err != nil {
			return err
		}
	}
}

// answer is what one event does to the map.
func (a *app) answer(e event) (redraw, done bool) {
	switch {
	case e.resize:
		return a.resized(e.cols, e.rows), false
	case e.worked:
		return true, false
	case e.key != "":
		return a.act(e.key)
	}
	return false, false
}
