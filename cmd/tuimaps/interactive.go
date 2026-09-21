package main

import (
	"context"
	"io"
	"os"
)

// eventRoom is how many events may wait for the loop. The pump says
// something every time it finishes a unit and must never be held up by a
// loop that is drawing.
const eventRoom = 64

// interactive is the map a person moves about with the keys: the terminal
// taken over and always given back (L-16), a pump of the app's own doing
// the library's work (D-73), and a loop that draws when something would
// make the frame differ and sleeps when nothing would (P-71).
func interactive(s settings, out, errs io.Writer) int {
	m, err := build(s)
	if err != nil {
		return complain(errs, err, exitMistake)
	}
	defer m.Close()

	term, err := open(out)
	if err != nil {
		return complain(errs, err, exitFailed)
	}
	cols, rows := sizeOf(s)
	a := newApp(m, s, out, cols, rows)
	a.places = m.Places() // what the scenario loaded, and what was named

	ctx, quit := context.WithCancel(context.Background())
	defer quit()
	events := make(chan event, eventRoom)
	working, err := pump(ctx, m, events)
	if err != nil {
		term.restore()
		return complain(errs, err, exitFailed)
	}
	stopWatching := watchingSize(ctx, events)
	go readingKeys(ctx, os.Stdin, events)

	err = kept(term, func() error { return a.loop(events, realClock()) })
	quit()
	stopWatching()
	working.Wait()
	if err != nil {
		return complain(errs, err, exitFailed)
	}
	return exitFine
}
