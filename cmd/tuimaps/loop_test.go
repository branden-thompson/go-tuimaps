package main

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
)

// TestParityP72a_Chrome is plan task 13.2: the app's own chrome, as
// upstream draws it - a row of keys at rows-2 and a status row at rows-1,
// the status row showing the focused marker's label after two arrows.
func TestParityP72a_Chrome(t *testing.T) {
	home := tuimaps.Place{Name: "Home", At: tuimaps.LonLat{Lon: -84.5, Lat: 33.8}}
	a, out := upFor(t, home)
	if err := a.drawn(noon); err != nil {
		t.Fatal(err)
	}
	rows := strings.Split(out.String(), "\r\n")
	if len(rows) != a.rows {
		t.Fatalf("%d rows for a terminal of %d", len(rows), a.rows)
	}
	keys := colourless(rows[a.rows-2])
	if !strings.Contains(keys, "quit") || !strings.Contains(keys, "zoom") || !strings.Contains(keys, "pan") {
		t.Errorf("the row at rows-2 is not the key row: %q", keys)
	}
	// With nothing focused the status row says where the map is.
	status := colourless(rows[a.rows-1])
	if strings.TrimSpace(status) == "" {
		t.Error("the status row is empty")
	}
	// Focused, it is upstream's own: two arrows and the label.
	a.act(keyFocus)
	out.Reset()
	if err := a.drawn(noon); err != nil {
		t.Fatal(err)
	}
	rows = strings.Split(out.String(), "\r\n")
	if got := colourless(rows[a.rows-1]); !strings.Contains(got, ">> Home") {
		t.Errorf("the status row with a place focused is %q, want it to carry >> Home", got)
	}
	// The map keeps the rest, and no more: the library is given the exact
	// rectangle that is left (L-17 g).
	if mapRows(a.rows) != a.rows-2 {
		t.Errorf("the map was given %d rows of %d", mapRows(a.rows), a.rows)
	}
	// Every row is cut to the terminal's width by the app's own rule.
	for i, row := range rows {
		if len([]rune(colourless(row))) > a.cols {
			t.Errorf("row %d is wider than the terminal", i)
		}
	}
}

// TestPanelsReplaceTheMapAndGoAway: the two panels a key opens - the
// description and the keys in full - are drawn where the map is and leave
// the chrome alone, so a person never loses the way back.
func TestPanelsReplaceTheMapAndGoAway(t *testing.T) {
	home := tuimaps.Place{Name: "Home", At: tuimaps.LonLat{Lon: -84.5, Lat: 33.8}}
	a, out := upFor(t, home)
	for _, c := range []struct{ key, says string }{
		{keyHelp, "braille"},
		{keyDescribe, "Home"},
	} {
		a.act(c.key)
		out.Reset()
		if err := a.drawn(noon); err != nil {
			t.Fatal(err)
		}
		shown := colourless(out.String())
		if !strings.Contains(shown, c.says) {
			t.Errorf("the panel %s does not say %q", c.key, c.says)
		}
		if !strings.Contains(shown, "quit") {
			t.Errorf("the panel %s took the key row with it", c.key)
		}
		a.act(c.key)
		out.Reset()
		if err := a.drawn(noon); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(colourless(out.String()), c.says) {
			t.Errorf("the panel %s would not go away", c.key)
		}
	}
}

// TestPump is plan task 13.3: the app's own pump sharpens the map, tells
// the loop each time it has done something, and stops when the app does.
func TestPump(t *testing.T) {
	m, err := tuimaps.New(tuimaps.WithSize(80, 22), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	ctx, quit := context.WithCancel(context.Background())
	defer quit()
	events := make(chan event, eventRoom)
	working, err := pump(ctx, m, events)
	if err != nil {
		t.Fatal(err)
	}
	// The map starts with nothing drawn and sharpens as the pump works.
	frame, err := m.Render(tuimaps.Size{Cols: 80, Rows: 22}, noon)
	if err != nil {
		t.Fatal(err)
	}
	for frame.Status != tuimaps.Complete {
		select {
		case <-events:
		case <-time.After(10 * time.Second):
			t.Fatal("the pump said nothing for ten seconds")
		}
		if frame, err = m.Render(tuimaps.Size{Cols: 80, Rows: 22}, noon); err != nil {
			t.Fatal(err)
		}
	}
	quit()
	done := make(chan struct{})
	go func() { working.Wait(); close(done) }()
	// Whatever the pump still has to say, it stops saying it.
	for {
		select {
		case <-events:
		case <-done:
			return
		case <-time.After(10 * time.Second):
			t.Fatal("the pump did not stop")
		}
	}
}

// TestParityP71_Loop is plan task 13.16's other row (P-71, a fix): upstream
// polls every fifty milliseconds; this loop draws when something would make
// the frame differ - a key, a resize, a unit of work - and otherwise waits
// for the moment the map itself said to call back by. It never polls.
func TestParityP71_Loop(t *testing.T) {
	a, out := upFor(t)
	events := make(chan event, 8)
	waits := make(chan time.Duration, 8)
	release := make(chan time.Time, 8)
	tick := clock{
		now: func() time.Time { return noon },
		after: func(d time.Duration) <-chan time.Time {
			waits <- d
			return release
		},
	}
	done := make(chan error, 1)
	go func() { done <- a.loop(events, tick) }()

	// It drew once at the start, without being asked to.
	waitFor(t, func() bool { return strings.Contains(out.String(), "\x1b[H") })

	// A unit of work done makes it draw again.
	before := len(out.String())
	events <- event{worked: true}
	waitFor(t, func() bool { return len(out.String()) > before })

	// A resize makes it draw at the new size.
	out.Reset()
	events <- event{resize: true, cols: 60, rows: 14}
	waitFor(t, func() bool { return out.rows() == 14 })

	// A key it has no use for makes it draw nothing at all.
	quiet := len(out.String())
	events <- event{key: "nonsense"}
	events <- event{key: keyNames} // and one it does, to know the first was read
	waitFor(t, func() bool { return len(out.String()) > quiet })

	// And quit ends it.
	events <- event{key: keyQuit}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the loop did not stop when it was told to quit")
	}
	// Nothing it waited for was a fifty-millisecond poll: every wait it
	// asked for was the map's own next call.
	close(waits)
	for wait := range waits {
		if wait > 0 && wait < time.Second {
			t.Errorf("the loop waited %v, which is a poll and not a deadline", wait)
		}
	}
}

// waitFor waits for something to become true, or fails the test.
func waitFor(t *testing.T, done func() bool) {
	t.Helper()
	for range 1000 {
		if done() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("it did not happen")
}

// TestTerminalRestoredOnPanic is plan task 13.9 (L-16): the terminal is put
// back on every way out - a clean return, an error, and a panic, which is
// re-raised only once the screen is usable again.
func TestTerminalRestoredOnPanic(t *testing.T) {
	var put putBack
	if err := kept(&put, func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	if put.times != 1 {
		t.Errorf("a clean return put the terminal back %d times", put.times)
	}

	put = putBack{}
	if err := kept(&put, func() error { return errNothingToDescribe() }); err == nil {
		t.Error("the error was swallowed")
	}
	if put.times != 1 {
		t.Errorf("an error put the terminal back %d times", put.times)
	}

	put = putBack{}
	caught := func() (caught any) {
		defer func() { caught = recover() }()
		kept(&put, func() error { panic("the map fell over") })
		return nil
	}()
	if caught == nil {
		t.Error("the panic was swallowed; a person would never know what happened")
	}
	if put.times != 1 {
		t.Errorf("a panic put the terminal back %d times, want once", put.times)
	}
}

// putBack is a terminal that only counts having been put back.
type putBack struct {
	mu    sync.Mutex
	times int
}

func (p *putBack) restore() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.times++
}

// TestInteractiveSaysSoWithNoTerminal: run with its output going to a pipe
// rather than a terminal, the app says plainly that this is not a terminal
// it can draw on, and leaves it alone.
func TestInteractiveSaysSoWithNoTerminal(t *testing.T) {
	out, errs, code := ran(t, "--offline", "--size", "69x12")
	if code == 0 {
		t.Error("the app drew an interactive map on something that is not a terminal")
	}
	if !strings.Contains(errs, "terminal") {
		t.Errorf("the complaint was %q", errs)
	}
	if strings.Contains(out, "\x1b[?1049h") {
		t.Error("the app took a screen of its own on something that is not a terminal")
	}
	if !clean(errs) {
		t.Errorf("the complaint is not clean text: %q", errs)
	}
}
