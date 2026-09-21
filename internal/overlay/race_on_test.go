//go:build race

package overlay

// raceDetector says the test binary was built with the race detector, which
// is what plan task 10.5's sub-process run needs.
const raceDetector = true
