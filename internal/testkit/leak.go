package testkit

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"
)

// ModulePath is the library's module path: the prefix of every function
// name inside it.
const ModulePath = "github.com/branden-thompson/go-tuimaps"

const (
	// leakAttempts and leakPause give a goroutine that is on its way out
	// the time to finish before it is called a leak.
	leakAttempts = 20
	leakPause    = 10 * time.Millisecond
	// maxDump bounds the goroutine dump.
	maxDump = 64 << 20
)

// LeakCheck reports any goroutine whose stack is inside the module and that
// no test is running. The library starts no goroutine, so after Close there
// must be none. Goroutines of the network stack are not the library's, even
// when a call of the library's started them, and are not reported.
func LeakCheck(module string) error {
	if module == "" {
		return errors.New("testkit: empty module path")
	}
	var leaks []string
	for range leakAttempts {
		dump, err := goroutineDump()
		if err != nil {
			return err
		}
		leaks = leakedIn(dump, module)
		if len(leaks) == 0 {
			return nil
		}
		time.Sleep(leakPause)
	}
	return fmt.Errorf("testkit: %d goroutine(s) still inside %s:\n%s", len(leaks), module, strings.Join(leaks, "\n"))
}

// goroutineDump returns the stacks of every goroutine.
func goroutineDump() (string, error) {
	for size := 1 << 16; size <= maxDump; size *= 2 {
		buf := make([]byte, size)
		if n := runtime.Stack(buf, true); n < size {
			return string(buf[:n]), nil
		}
	}
	return "", fmt.Errorf("testkit: goroutine dump is larger than %d bytes", maxDump)
}

// leakedIn reads a goroutine dump and returns one line for each goroutine
// that has a frame inside module and no frame of the testing package.
func leakedIn(dump, module string) []string {
	var leaks []string
	for block := range strings.SplitSeq(dump, "\n\n") {
		header, frames, _ := strings.Cut(strings.TrimSpace(block), "\n")
		inside, underTest := "", false
		for line := range strings.SplitSeq(frames, "\n") {
			if strings.HasPrefix(line, "created by ") {
				break
			}
			if strings.HasPrefix(line, "\t") {
				continue
			}
			if strings.HasPrefix(line, "testing.") {
				underTest = true
			}
			if inside == "" && (strings.HasPrefix(line, module+"/") || strings.HasPrefix(line, module+".")) {
				inside = line
			}
		}
		if inside != "" && !underTest {
			leaks = append(leaks, header+" "+inside)
		}
	}
	return leaks
}
