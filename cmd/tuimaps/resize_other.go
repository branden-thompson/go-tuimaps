//go:build !darwin && !linux

package main

import "context"

// watchingSize has no signal to watch on this system, so a map redrawn
// after a resize waits for the next key. It is said here rather than left
// to be discovered.
func watchingSize(_ context.Context, _ chan<- event) func() { return func() {} }
