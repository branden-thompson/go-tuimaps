package main

import (
	"context"
	"fmt"
	"io"
)

// The sequences the app sends to the terminal itself. The library emits
// colour and nothing else (FR-34); moving the cursor, hiding it and taking
// a screen of one's own are the app's business.
const (
	takeScreen = "\x1b[?1049h\x1b[?25l\x1b[H\x1b[2J"
	giveBack   = "\x1b[?25h\x1b[?1049l"
)

// restorer is a terminal that has been changed and must be put back. A test
// has one of its own, so that "put back on every path out, a panic
// included" can be shown without a terminal (L-16).
type restorer interface {
	restore()
}

// kept runs the app with the terminal put back whatever happens: a clean
// return, an error, or a panic, which is re-raised once the terminal is
// usable again so that nobody reads a stack trace through a half-changed
// screen (L-16).
func kept(term restorer, draw func() error) (err error) {
	defer func() {
		term.restore()
		if caught := recover(); caught != nil {
			panic(caught)
		}
	}()
	return draw()
}

// terminal is the real one: the app's own screen, raw mode, and the
// settings it found, kept so that they can be put back exactly.
type terminal struct {
	out  io.Writer
	was  *state
	told bool
}

// open takes the terminal over. A terminal it cannot take over - a pipe, a
// file, a run with no terminal at all - is said plainly rather than drawn
// on as though it were one.
func open(out io.Writer) (*terminal, error) {
	was, err := raw()
	if err != nil {
		return nil, fmt.Errorf("this is not a terminal the map can be drawn on: %w", err)
	}
	t := &terminal{out: out, was: was}
	if _, err := io.WriteString(out, takeScreen); err != nil {
		t.restore()
		return nil, err
	}
	t.told = true
	return t, nil
}

// restore puts the terminal back as it was found. It may be called twice -
// once on the way out and once by a deferred call - and the second time
// does nothing.
func (t *terminal) restore() {
	if t == nil {
		return
	}
	if t.told {
		t.told = false
		io.WriteString(t.out, giveBack)
	}
	if t.was != nil {
		unraw(t.was)
		t.was = nil
	}
}

// readingKeys sends what is typed, decoded into key names, until the context is
// done. A terminal in raw mode hands over each key as it is pressed, so the
// loop hears a key without waiting for a line.
func readingKeys(ctx context.Context, in io.Reader, events chan<- event) {
	buffer := make([]byte, 64)
	for {
		read, err := in.Read(buffer)
		if read > 0 {
			for _, key := range decode(buffer[:read]) {
				select {
				case events <- event{key: key}:
				case <-ctx.Done():
					return
				}
			}
		}
		if err != nil {
			return
		}
		if ctx.Err() != nil {
			return
		}
	}
}
