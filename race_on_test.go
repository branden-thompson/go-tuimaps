//go:build race

package tuimaps_test

// raceDetector says the test binary was built with the race detector, whose
// own allocations and collections move a heap's peak: NFR-3's peak is judged
// with it off.
const raceDetector = true
