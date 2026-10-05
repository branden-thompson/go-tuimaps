package describe

// flow_test.go — v0.2.0 D-122: the heavier rain's motion is measured over
// every pair of frames, block by block, placed between blocks by the
// parabola, from the rain near the place alone.

import (
	"math"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/project"
)

// flowSize is the side of a test image: 0.01 degrees a pixel at the
// equator, about 1.1 km, so a block is two pixels.
const flowSize = 400

// squares is a frame with a square of heavier rain, side pixels a side, its
// top left corner at each pixel given.
func squares(at time.Time, side int, corners ...[2]int) Frame {
	cls := make([]int8, flowSize*flowSize)
	for _, c := range corners {
		for y := c[1]; y < c[1]+side; y++ {
			for x := c[0]; x < c[0]+side; x++ {
				if x >= 0 && y >= 0 && x < flowSize && y < flowSize {
					cls[y*flowSize+x] = 5
				}
			}
		}
	}
	return Frame{Valid: at, Image: Image{West: 0, South: -2, East: 4, North: 2, Width: flowSize, Height: flowSize, Classes: cls}}
}

// TestAStepBetweenBlocksIsPlacedByTheParabola: a square three pixels
// further east every five minutes - a block and a half - moves at about
// 40 km/h, not the 27 or 53 of one block or two.
func TestAStepBetweenBlocksIsPlacedByTheParabola(t *testing.T) {
	t0 := time.Unix(0, 0)
	var frames []Frame
	for k := range 6 {
		frames = append(frames, squares(t0.Add(time.Duration(k)*5*time.Minute), 20, [2]int{100 + 3*k, 190}))
	}
	mo, missing := Track(frames, 3, project.LonLat{Lon: 2, Lat: 0})
	want := 3 * 0.01 * kmPerDegree * 12 // three pixels each five minutes, an hour of them
	if missing != 0 || !mo.Moving || math.Abs(mo.SpeedKmh-want) > 5 || math.Abs(mo.Heading-90) > 5 {
		t.Errorf("motion %+v, missing %v; want east at about %.0f km/h", mo, missing, want)
	}
}

// TestTooLittleRainIsNoMotion: fewer heavier pixels than leastHot near the
// place say nothing about where rain moves.
func TestTooLittleRainIsNoMotion(t *testing.T) {
	t0 := time.Unix(0, 0)
	frames := []Frame{squares(t0, 2, [2]int{200, 200}), squares(t0.Add(5*time.Minute), 2, [2]int{202, 200})}
	if _, missing := Track(frames, 3, project.LonLat{Lon: 2, Lat: 0}); missing != TooFewFrames {
		t.Errorf("four heavier pixels: missing %v; want too few frames", missing)
	}
}

// TestRainFarAwayDoesNotSteerTheMotion: a small storm near the place moves
// east; a large one nearly 300 km away moves north. The motion near the place
// is the near storm's.
func TestRainFarAwayDoesNotSteerTheMotion(t *testing.T) {
	t0 := time.Unix(0, 0)
	var frames []Frame
	for k := range 4 {
		frames = append(frames, squares(t0.Add(time.Duration(k)*5*time.Minute), 12, [2]int{40 + 4*k, 190}, [2]int{300, 200 - 4*k}))
		// the far storm, 60 pixels a side, is drawn over the second corner
		for y := 200 - 4*k; y < 260-4*k; y++ {
			for x := 300; x < 360; x++ {
				frames[k].Image.Classes[y*flowSize+x] = 5
			}
		}
	}
	mo, missing := Track(frames, 3, project.LonLat{Lon: 0.4, Lat: 0})
	if missing != 0 || math.Abs(mo.Heading-90) > 10 {
		t.Errorf("near the small storm: heading %.0f, missing %v; want east, its own", mo.Heading, missing)
	}
}

// TestAPairWithNothingToMeetIsNotMeasured: where the next frame has no
// heavier rain within what rain can travel, every shift fits nothing, and
// the pair says nothing about motion - it is not rain that held still.
func TestAPairWithNothingToMeetIsNotMeasured(t *testing.T) {
	t0 := time.Unix(0, 0)
	frames := []Frame{squares(t0, 20, [2]int{100, 190}), squares(t0.Add(5*time.Minute), 20, [2]int{160, 190})} // 66 km in five minutes
	if mo, missing := Track(frames, 3, project.LonLat{Lon: 1.7, Lat: 0}); missing != TooFewFrames {
		t.Errorf("a pair with nothing to meet: %+v, missing %v; want too few frames", mo, missing)
	}
}

// TestTheSpanIsTheTimeMeasured: a frame with no heavier rain leaves the
// pairs either side of it unmeasured; the span is the time the measured
// pairs cover, and From is where the rain was that long before.
func TestTheSpanIsTheTimeMeasured(t *testing.T) {
	t0 := time.Unix(0, 0)
	frames := []Frame{
		squares(t0, 20, [2]int{100, 190}),
		squares(t0.Add(5*time.Minute), 20),
		squares(t0.Add(10*time.Minute), 20, [2]int{106, 190}),
		squares(t0.Add(15*time.Minute), 20, [2]int{109, 190}),
	}
	mo, missing := Track(frames, 3, project.LonLat{Lon: 1.2, Lat: 0})
	if missing != 0 || !mo.From.Valid.Equal(t0.Add(10*time.Minute)) || !mo.To.Valid.Equal(t0.Add(15*time.Minute)) {
		t.Errorf("motion %+v, missing %v; want it measured over the last five minutes alone", mo, missing)
	}
}

// TestAWideImageFindsTheNearestRain: an image wider than half the world
// places a point east of its middle on its east side, not off its west edge.
func TestAWideImageFindsTheNearestRain(t *testing.T) {
	cls := make([]int8, 360*10)
	cls[5*360+340] = 5 // a heavier pixel at about 170 east
	cls[5*360+2] = 5   // and one at about 178 west
	img := Image{West: -180, South: -5, East: 180, North: 5, Width: 360, Height: 10, Classes: cls}
	at, ok := img.nearestHot(3, project.LonLat{Lon: 160, Lat: 0})
	if !ok || math.Abs(at.Lon-160.5) > 1 {
		t.Errorf("nearest to 160 east: %v, %v; want the pixel near 160 east", at, ok)
	}
}

// TestHourlyFramesAreMeasuredWithinBounds (REVIEW, perf F9): frames an hour
// apart give a reach of dozens of blocks; the search stays bounded - every
// second shift, then each around the best - and still finds the motion. The
// bound is held as a count of shifts tried, never as a time.
func TestHourlyFramesAreMeasuredWithinBounds(t *testing.T) {
	t0 := time.Unix(0, 0)
	var frames []Frame
	for k := range 3 {
		frames = append(frames, squares(t0.Add(time.Duration(k)*time.Hour), 30, [2]int{120 + 20*k, 185}))
	}
	mo, missing := Track(frames, 3, project.LonLat{Lon: 1.7, Lat: 0})
	want := 20 * 0.01 * kmPerDegree // twenty pixels an hour
	if missing != 0 || math.Abs(mo.Heading-90) > 5 || math.Abs(mo.SpeedKmh-want) > 5 {
		t.Errorf("hourly frames: %+v, missing %v; want east at about %.0f km/h", mo, missing, want)
	}
	for _, f := range frames[1:] {
		for _, moved := range f.steps {
			if moved.tries > 70*70 {
				t.Errorf("the search tried %d shifts for one pair; want the coarse grid and its refinement, not every shift", moved.tries)
			}
		}
	}
}
