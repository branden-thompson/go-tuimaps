package main

import (
	"context"
	"io"
	"math"
	"strings"
	"sync"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
)

// upFor is an app over a map of the world drawn from the tiles built in,
// with the places a test names. It reaches nothing.
func upFor(t *testing.T, places ...tuimaps.Place) (*app, *shown) {
	t.Helper()
	m, err := tuimaps.New(tuimaps.WithSize(80, 22), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { m.Close() })
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(places) > 0 {
		if _, err := m.SetPlaces(places); err != nil {
			t.Fatal(err)
		}
	}
	out := &shown{}
	a := newApp(m, settings{places: places}, out, 80, 24)
	return a, out
}

// TestParityP68a_Keys is plan task 13.1: upstream's own keys, each doing
// what upstream's does - pan, zoom, the layer toggles, fit world, the
// markers, and the two ways out.
func TestParityP68a_Keys(t *testing.T) {
	a, _ := upFor(t, tuimaps.Place{Name: "Home", At: tuimaps.LonLat{Lon: -84.5, Lat: 33.8}})
	if err := a.m.Zoom(4); err != nil {
		t.Fatal(err)
	}
	if err := a.m.Recentre(tuimaps.LonLat{Lon: 0, Lat: 0}); err != nil {
		t.Fatal(err)
	}

	// The two ways out, and nothing else ends the app.
	for _, key := range []string{keyQuit} {
		if _, done := a.act(key); !done {
			t.Errorf("%s did not end the app", key)
		}
	}
	for _, key := range []string{keyZoomIn, keyZoomOut, keyLeft, keyRight, keyUp, keyDown,
		keyNames, keyWater, keyWorld, keyMarkers, keyFocus, keySafeRamps, keyReduce, keyColour} {
		if _, done := a.act(key); done {
			t.Errorf("%s ended the app", key)
		}
	}

	// Zoom, with nothing focused: one level a press, as upstream.
	_, was := a.m.Centre()
	a.act(keyZoomIn)
	if _, now := a.m.Centre(); math.Abs(now-was-1) > 0.001 {
		t.Errorf("zoom in went from %v to %v", was, now)
	}
	a.act(keyZoomOut)
	if _, now := a.m.Centre(); math.Abs(now-was) > 0.001 {
		t.Errorf("zoom out did not undo zoom in: %v, want %v", now, was)
	}

	// The layer switches turn something off and on again.
	for _, key := range []string{keyNames, keyWater, keyMarkers} {
		before := a.frameText(t)
		a.act(key)
		off := a.frameText(t)
		if off == before {
			t.Errorf("%s changed nothing on the map", key)
		}
		a.act(key)
		if back := a.frameText(t); back != before {
			t.Errorf("%s did not put back what it took away", key)
		}
	}

	// Fit world puts the whole world on the map, wherever it was.
	if err := a.m.Zoom(8); err != nil {
		t.Fatal(err)
	}
	a.act(keyWorld)
	if _, zoom := a.m.Centre(); zoom > 2 {
		t.Errorf("fit world left the map at zoom %v", zoom)
	}
}

// frameText is the map as it would be drawn, without its colours, so that
// two frames can be compared as pictures.
func (a *app) frameText(t *testing.T) string {
	t.Helper()
	frame, err := a.m.Render(tuimaps.Size{Cols: a.cols, Rows: mapRows(a.rows)}, noon)
	if err != nil {
		t.Fatal(err)
	}
	return colourless(strings.Join(frame.Lines, "\n"))
}

// TestParityP69_PanStep is plan task 13.16's first row: upstream's own pan
// step, eight degrees of longitude and six of latitude at zoom 0, halving
// with every level.
func TestParityP69_PanStep(t *testing.T) {
	a, _ := upFor(t)
	for _, zoom := range []float64{0, 1, 4, 7.5} {
		if err := a.m.Zoom(zoom); err != nil {
			t.Fatal(err)
		}
		if err := a.m.Recentre(tuimaps.LonLat{Lon: 10, Lat: 20}); err != nil {
			t.Fatal(err)
		}
		wantEast, wantNorth := 8/math.Pow(2, zoom), 6/math.Pow(2, zoom)
		a.act(keyRight)
		at, _ := a.m.Centre()
		if math.Abs(at.Lon-(10+wantEast)) > 1e-9 {
			t.Errorf("at zoom %v, east went to %v, want %v", zoom, at.Lon, 10+wantEast)
		}
		a.act(keyLeft)
		a.act(keyUp)
		at, _ = a.m.Centre()
		if math.Abs(at.Lat-(20+wantNorth)) > 1e-9 {
			t.Errorf("at zoom %v, north went to %v, want %v", zoom, at.Lat, 20+wantNorth)
		}
		a.act(keyDown)
		at, _ = a.m.Centre()
		if math.Abs(at.Lon-10) > 1e-9 || math.Abs(at.Lat-20) > 1e-9 {
			t.Errorf("at zoom %v, four presses did not come back: %+v", zoom, at)
		}
	}
	// A step over the seam comes out the other side, not off the world.
	if err := a.m.Zoom(0); err != nil {
		t.Fatal(err)
	}
	if err := a.m.Recentre(tuimaps.LonLat{Lon: 177, Lat: 0}); err != nil {
		t.Fatal(err)
	}
	a.act(keyRight)
	if at, _ := a.m.Centre(); at.Lon > 0 {
		t.Errorf("a step east from 177 went to %v; longitude is circular", at.Lon)
	}
}

// TestZoomAroundFocusKey is the second half of plan task 13.14: with a
// place focused, zooming zooms about that place - the keyboard's equivalent
// of zooming towards what a pointer points at (FR-5, D-17), so that nothing
// on the map needs a pointer to reach (NFR-15).
func TestZoomAroundFocusKey(t *testing.T) {
	home := tuimaps.Place{Name: "Home", At: tuimaps.LonLat{Lon: 20, Lat: 10}}
	a, _ := upFor(t, home)
	if err := a.m.Zoom(4); err != nil {
		t.Fatal(err)
	}
	if err := a.m.Recentre(tuimaps.LonLat{Lon: 0, Lat: 0}); err != nil {
		t.Fatal(err)
	}

	// With nothing focused, the middle stays where it is.
	a.act(keyZoomIn)
	if at, _ := a.m.Centre(); math.Abs(at.Lon) > 1e-9 || math.Abs(at.Lat) > 1e-9 {
		t.Errorf("with nothing focused the middle moved to %+v", at)
	}

	// Focused, and zooming in: the middle comes half way to the place, so
	// the place stays where it is on the screen.
	if err := a.m.Zoom(4); err != nil {
		t.Fatal(err)
	}
	if err := a.m.Recentre(tuimaps.LonLat{Lon: 0, Lat: 0}); err != nil {
		t.Fatal(err)
	}
	a.act(keyFocus)
	if place, ok := a.focusedPlace(); !ok || place.Name != "Home" {
		t.Fatalf("the focus is %+v", place)
	}
	a.act(keyZoomIn)
	at, zoom := a.m.Centre()
	if math.Abs(zoom-5) > 1e-9 {
		t.Errorf("the zoom went to %v", zoom)
	}
	if math.Abs(at.Lon-10) > 1e-6 || math.Abs(at.Lat-5) > 1e-6 {
		t.Errorf("the middle went to %+v, want half way to the place at 20,10", at)
	}
	// And out again puts it back.
	a.act(keyZoomOut)
	if at, _ := a.m.Centre(); math.Abs(at.Lon) > 1e-6 || math.Abs(at.Lat) > 1e-6 {
		t.Errorf("zooming out about the place left the middle at %+v", at)
	}
	// The focus goes round the places and back to none.
	a.act(keyFocus)
	if _, ok := a.focusedPlace(); ok {
		t.Error("the focus cannot be given up")
	}
}

// TestKeysAreDecodedFromWhatATerminalSends: arrow keys arrive as escape
// sequences, several keys may arrive in one read, and what the app has no
// use for is passed over rather than acted on.
func TestKeysAreDecodedFromWhatATerminalSends(t *testing.T) {
	for _, c := range []struct {
		name string
		sent string
		want []string
	}{
		{"one letter", "q", []string{keyQuit}},
		{"several at once", "anw", []string{keyZoomIn, keyNames, keyWorld}},
		{"an arrow", "\x1b[A", []string{keyUp}},
		{"every arrow", "\x1b[A\x1b[B\x1b[C\x1b[D", []string{keyUp, keyDown, keyRight, keyLeft}},
		{"an arrow among letters", "n\x1b[Cw", []string{keyNames, keyRight, keyWorld}},
		{"escape on its own", "\x1b", []string{keyQuit}},
		{"a sequence the app has no use for", "\x1b[5~", nil},
		{"letters it has no use for", "xvp", nil},
		{"the interrupt character", "\x03", []string{keyQuit}},
		{"tab", "\t", []string{keyFocus}},
	} {
		got := decode([]byte(c.sent))
		if len(got) != len(c.want) {
			t.Errorf("%s: %v, want %v", c.name, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%s: %v, want %v", c.name, got, c.want)
				break
			}
		}
	}
}

// TestSafeRampsKeyAndFlag is plan task 13.8: each accessibility switch has
// a key as well as a flag, and the key and the flag reach the same setting
// (NFR-15). A person who cannot use one must be able to use the other.
func TestSafeRampsKeyAndFlag(t *testing.T) {
	byFlag, _ := upFor(t)
	byFlag.safeRamps, byFlag.reduceMotion, byFlag.noColour = true, true, true
	byKey, _ := upFor(t)
	byKey.act(keySafeRamps)
	byKey.act(keyReduce)
	byKey.act(keyColour)
	if byKey.safeRamps != byFlag.safeRamps || byKey.reduceMotion != byFlag.reduceMotion || byKey.noColour != byFlag.noColour {
		t.Errorf("the keys gave %+v, the flags %+v", byKey, byFlag)
	}
	// And the colour key really takes the colour off the map.
	withColour := byKey.frameText(t)
	raw, err := byKey.m.Render(tuimaps.Size{Cols: byKey.cols, Rows: mapRows(byKey.rows)}, noon)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.Join(raw.Lines, ""), "\x1b[") {
		t.Error("the colour key was pressed and the map still has colour in it")
	}
	byKey.act(keyColour)
	back, err := byKey.m.Render(tuimaps.Size{Cols: byKey.cols, Rows: mapRows(byKey.rows)}, noon)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(back.Lines, ""), "\x1b[") {
		t.Error("the colour key did not put the colour back")
	}
	if withColour == "" {
		t.Error("the map drew nothing at all")
	}
}

// TestEveryStateIsReachableByKeysAlone is NFR-15 stated as a test: every
// switch the help lists has a key, and pressing that key changes the app's
// own state. A state reachable only by a flag is a state a person cannot
// reach while the map is up.
func TestEveryStateIsReachableByKeysAlone(t *testing.T) {
	for _, switched := range accessibility() {
		if switched.key == "" {
			continue // the environment's, which has no key by its nature
		}
		a, _ := upFor(t)
		before := a.state()
		keys := decode([]byte(switched.key))
		if len(keys) != 1 {
			t.Errorf("%s: the key %q is not one the app reads", switched.what, switched.key)
			continue
		}
		a.act(keys[0])
		if a.state() == before {
			t.Errorf("%s: pressing %q changed nothing", switched.what, switched.key)
		}
	}
}

// state is what the switches are set to, as one string, so that a test can
// say whether a key changed anything at all.
func (a *app) state() string {
	return strings.Join([]string{
		yesNo(a.safeRamps), yesNo(a.reduceMotion), yesNo(a.noColour),
		yesNo(a.labels), yesNo(a.water), yesNo(a.markers),
		yesNo(a.helpUp), yesNo(a.describeUp),
	}, "")
}

func yesNo(on bool) string {
	if on {
		return "y"
	}
	return "n"
}

// TestNothingIsWrittenOutsideTheTerminalsRows: the app writes whole
// screenfuls and never more rows than the terminal has.
func TestNothingIsWrittenOutsideTheTerminalsRows(t *testing.T) {
	a, out := upFor(t)
	if err := a.drawn(noon); err != nil {
		t.Fatal(err)
	}
	written := out.String()
	if rows := strings.Count(written, "\r\n") + 1; rows != a.rows {
		t.Errorf("%d rows were written, and the terminal has %d", rows, a.rows)
	}
	if !strings.HasPrefix(written, "\x1b[H") {
		t.Error("the app drew without putting the cursor back at the top")
	}
}

// shown is what the app has written, readable from another goroutine. The
// app itself is the loop's, and a test that reached into it while the loop
// was running would be reading a map someone else is drawing.
type shown struct {
	mu      sync.Mutex
	written strings.Builder
}

func (s *shown) Write(b []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.written.Write(b)
}

// String is everything written so far.
func (s *shown) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.written.String()
}

// Reset forgets what was written, so that a test can watch for what comes
// next.
func (s *shown) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.written.Reset()
}

// rows is how many rows the last screenful held, which is how a test sees
// the size the loop is drawing at without reading the app itself.
func (s *shown) rows() int {
	whole := s.String()
	at := strings.LastIndex(whole, "\x1b[H")
	if at < 0 {
		return 0
	}
	return strings.Count(whole[at:], "\r\n") + 1
}

var _ io.Writer = (*shown)(nil)
