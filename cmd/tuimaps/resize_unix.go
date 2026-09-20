//go:build darwin || linux

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// watchingSize tells the loop when the terminal changes size. A terminal
// says so with a signal, so nothing here polls (P-71). The function it
// gives back stops the watch.
func watchingSize(ctx context.Context, events chan<- event) func() {
	changed := make(chan os.Signal, 1)
	signal.Notify(changed, syscall.SIGWINCH)
	go func() {
		defer signal.Stop(changed)
		for {
			select {
			case <-ctx.Done():
				return
			case <-changed:
				cols, rows := terminalSize()
				select {
				case events <- event{resize: true, cols: cols, rows: rows}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return func() { signal.Stop(changed) }
}
