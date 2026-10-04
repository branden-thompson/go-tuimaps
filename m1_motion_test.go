package tuimaps_test

// m1_motion_test.go — v0.2.0 D-122, D-123, D-127: a regression pin, not
// evidence for M1. The motion the library measures over the five loops of
// M1's first sitting - the loops the method was built on, with the HUM
// LEAD's answers in view - stays within one compass point of the direction
// the HUM LEAD saw the heavier rain move in each (07-readiness/m1-sitting.md,
// sitting 1's ground truth), measured over every pair of frames.

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// sittingOne is each loop of M1's first sitting, with the HUM LEAD's ground
// truth as a bearing on the 8-point compass.
var sittingOne = []struct {
	dir   string
	truth float64
}{
	{"outbreak-ohio-2019-05-28", 90},     // E
	{"moore-oklahoma-2013-05-20", 45},    // NE
	{"derecho-indiana-2012-06-29", 135},  // SE
	{"mayfield-kentucky-2021-12-11", 45}, // NE
	{"harvey-houston-2017-08-26", 0},     // N
}

// recordedLoop is a loop as testdata/loops keeps one, as a radar overlay
// read with IEM's table - as the demo app reads it.
func recordedLoop(t *testing.T, dir string) tuimaps.Overlay {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(dir, "loop.json"))
	if err != nil {
		t.Fatal(err)
	}
	var lf struct {
		West, South, East, North float64
		Frames                   []struct{ Valid, File string }
	}
	if err := json.Unmarshal(body, &lf); err != nil {
		t.Fatal(err)
	}
	frames := make([]tuimaps.LoopFrame, 0, len(lf.Frames))
	for _, f := range lf.Frames {
		valid, err := time.Parse(time.RFC3339, f.Valid)
		if err != nil {
			t.Fatal(err)
		}
		png, err := os.ReadFile(filepath.Join(dir, filepath.Base(f.File)))
		if err != nil {
			t.Fatal(err)
		}
		frames = append(frames, tuimaps.LoopFrame{Valid: valid, PNG: png})
	}
	return tuimaps.RadarImage("loop", tuimaps.Image{Frames: frames, West: lf.West, South: lf.South, East: lf.East, North: lf.North,
		Projection: tuimaps.PlateCarree, Provider: tuimaps.ProviderIEM}, frames[len(frames)-1].Valid)
}

// TestTheMotionOfSittingOneIsWithinAPoint (D-123): over each loop, framed
// as the demo app frames it and with no place named, the heading is within
// one point of the HUM LEAD's ground truth (45 degrees either way of it,
// each heading taken to its nearest point), over the whole loop.
func TestTheMotionOfSittingOneIsWithinAPoint(t *testing.T) {
	for _, loop := range sittingOne {
		t.Run(loop.dir, func(t *testing.T) {
			m := world(t, 149, 38)
			mustSet(t, m, recordedLoop(t, filepath.Join("testdata", "loops", loop.dir)))
			must(t, m.FitTo(nil, []string{"loop"}, 2))
			settle(t, m)
			r, err := m.Report(nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(r.Motion) != 1 || r.Motion[0].Missing != 0 || !r.Motion[0].Moving {
				t.Fatalf("motion: %+v; want one measured, moving", r.Motion)
			}
			mo := r.Motion[0]
			point := math.Mod(math.Round(mo.Heading/45)*45, 360)
			off := math.Abs(math.Mod(point-loop.truth+540, 360) - 180)
			if off > 45 {
				t.Errorf("heading %.0f (%s), the point %.0f, is %.0f degrees from the ground truth %.0f; want within one point", mo.Heading, mo.HeadingCompass, point, off, loop.truth)
			}
			if mo.Span != 55*time.Minute {
				t.Errorf("measured over %v; want every pair of frames, 55 minutes", mo.Span)
			}
			raw := math.Abs(math.Mod(mo.Heading-loop.truth+540, 360) - 180)
			t.Logf("heading %.0f (%s), %.0f degrees from the ground truth, at %.0f km/h over %v", mo.Heading, mo.HeadingCompass, raw, mo.SpeedKmh, mo.Span)

			// A named place, as a host names one: the view's centre by name
			// gives the same motion.
			centre, _ := m.Centre()
			named, err := m.Report([]tuimaps.Place{{ID: "here", Name: "Here", At: centre}})
			if err != nil {
				t.Fatal(err)
			}
			if len(named.Motion) != 1 || named.Motion[0].Place != "Here" || named.Motion[0].Heading != mo.Heading {
				t.Errorf("named at the centre: %+v; want Here, heading %.0f", named.Motion, mo.Heading)
			}
		})
	}
}
