package tuimaps_test

import (
	"runtime/debug"
	"testing"
)

// TestRaceDetectorIsToldTruly: raceDetector, which decides whether NFR-3's
// peak is judged, says how this test binary was built - the -race setting
// the build records - so a tag swapped or a constant flipped cannot leave
// the peak judged nowhere (watchpost D-299, R6C-7, R6L-8).
func TestRaceDetectorIsToldTruly(t *testing.T) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		t.Skip("the binary carries no build information to read")
	}
	built := false
	for _, s := range info.Settings {
		built = built || s.Key == "-race" && s.Value == "true"
	}
	if built != raceDetector {
		t.Errorf("raceDetector is %v, and the binary was built with -race %v", raceDetector, built)
	}
}
