package overlay

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/project"
)

// borrowCase names the program the sub-process runs.
const borrowCase = "TUIMAPS_BORROW_CASE"

// reuse runs one of plan task 10.5's two programs. The library reads the
// host's geometry where it lies, so the host may write over it only once the
// release has been reported. "early" writes without waiting for that word;
// "patient" waits for it. The write is the same in both.
func reuse(wait bool) {
	s, err := NewStore(Caps{})
	if err != nil {
		panic(err)
	}
	held := ringOverlay("warnings", 200_000)
	ring := held.Features[0].Rings[0]
	if _, err = s.HandIn(held); err != nil {
		panic(err)
	}
	reader, ok := s.Read("warnings")
	if !ok {
		panic("nothing to read")
	}
	done, reading, writing := make(chan struct{}), make(chan struct{}), make(chan struct{})
	go func() { // what a Work call does: it reads the host's geometry where it lies
		defer close(done)
		close(reading)
		if !wait {
			<-writing // early: the read begins once the host has begun to write
		}
		Prepare(reader.Overlay(), 8)
		reader.Done()
	}()
	<-reading
	if _, err = s.HandIn(ringOverlay("warnings", 8)); err != nil {
		panic(err)
	}
	if wait {
		<-done // the call that was reading the old geometry has returned
		scribble(ring)
		return
	}
	// EARLY, THE WRITES OVERLAP THE READ BY CONSTRUCTION, not by timing: the
	// reader starts after the first write, and every write after that one is
	// unordered with its reads until done is received - so the detector sees
	// the race on every run, however fast the reader.
	scribble(ring)
	close(writing)
	for {
		scribble(ring)
		select {
		case <-done:
			return
		default:
		}
	}
}

// scribble is the host writing over the memory it handed in.
func scribble(ring []project.LonLat) {
	for i := range ring {
		ring[i].Lon = -ring[i].Lon
	}
}

// TestBorrowReuseCase is not a test of its own: it is the body of the
// sub-process that TestBorrowReuseUnderRaceDetector runs.
func TestBorrowReuseCase(t *testing.T) {
	switch os.Getenv(borrowCase) {
	case "early":
		reuse(false)
	case "patient":
		reuse(true)
	default:
		t.Skip("run as a sub-process by TestBorrowReuseUnderRaceDetector")
	}
}

// TestBorrowReuseUnderRaceDetector is the rest of plan task 10.5 (FR-11,
// D-86): borrowed geometry is the host's to write over only after the release
// is reported. The rule is shown, not asserted about, by running both
// programs under the race detector - a detector finding inside this process
// would fail the run itself, so each is a sub-process.
func TestBorrowReuseUnderRaceDetector(t *testing.T) {
	if !raceDetector {
		t.Skip("the race detector is not on in this run; the gate's first leg runs it")
	}
	if testing.Short() {
		t.Skip("each sub-process prepares 200,000 vertices")
	}
	cases := []struct {
		name string
		race bool
	}{
		{"early", true},
		{"patient", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// THE DETECTOR KEEPS A BOUNDED HISTORY of each word's accesses, so
			// a race that happened is reported in nearly every run, not in
			// every one (about one run in thirty missed it under load). The
			// early program is run up to earlyTries times and passes on the
			// first run that reports the race; it fails only when none does.
			// The patient program runs once and must be clean: a race reported
			// there is a defect, never the detector's luck.
			tries := 1
			if c.race {
				tries = earlyTries
			}
			var err error
			var found bool
			var out []byte
			for range tries {
				run := exec.Command(os.Args[0], "-test.run=^TestBorrowReuseCase$", "-test.count=1")
				run.Env = append(os.Environ(), borrowCase+"="+c.name)
				out, err = run.CombinedOutput()
				found = strings.Contains(string(out), "DATA RACE")
				if c.race && err != nil && found {
					return
				}
			}
			if c.race {
				t.Errorf("writing over borrowed geometry before the release was reported: in %d runs no data race was reported (last: exit %v); want a failed run with a data race", tries, err)
			}
			if !c.race && (err != nil || found) {
				t.Errorf("writing over it after the release was reported: exit %v, race reported %v; want a clean run\n%s", err, found, out)
			}
		})
	}
}

// earlyTries is how many runs of the early program may pass before one
// reports its race.
const earlyTries = 5
