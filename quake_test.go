package tuimaps_test

// quake_test.go — v0.2.0 L-19 (watchpost D-122, D-123): a circle sized on
// the screen - a quake's ring, fixed by its magnitude at every zoom, as
// USGS's map draws it - and the quakes' colours by age.

import (
	"strings"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// quakeRing is a ring of a radius in dots at Memphis, labelled.
func quakeRing(dots int, km float64) tuimaps.Overlay {
	return tuimaps.Overlay{ID: "quake", Valid: noon, Keeps: 24 * time.Hour, Features: []tuimaps.Feature{{Kind: tuimaps.Circle,
		Centre: tuimaps.LonLat{Lon: -90, Lat: 35.1}, RadiusDots: dots, RadiusKm: km, Role: tuimaps.QuakeDay, Label: "M4.1 2:14 PM"}}}
}

// TestAScreenRingKeepsItsSizeAtEveryZoom is L-19.1: a circle given a radius
// in dots is a braille ring that size on the screen, at county zoom as at
// national, its label beside it - where a circle in km grows as the map zooms.
func TestAScreenRingKeepsItsSizeAtEveryZoom(t *testing.T) {
	m := reportMap(t)
	m.ColourDepth(tuimaps.Truecolor)
	if unknown, err := m.SetPalette(map[string]tuimaps.RGB{"quake.day": {R: 251, G: 171, B: 41}}); err != nil || len(unknown) != 0 {
		t.Fatalf("quake.day is not a palette's token: %v %v", unknown, err)
	}
	const fg = "38;2;251;171;41"
	mustSet(t, m, quakeRing(8, 0))
	counts := []int{}
	for _, z := range []float64{4, 7} {
		must(t, m.Recentre(tuimaps.LonLat{Lon: -90, Lat: 35.1}))
		must(t, m.Zoom(z))
		settle(t, m)
		frame := momentFrame(t, m, noon)
		found := false
		for _, row := range strings.Split(colours.ReplaceAllString(frame, ""), "\n") {
			if at := strings.Index(row, "M4.1 2:14 PM"); at >= 0 {
				found = len([]rune(row[:at])) > 40 // the frame is 80 wide, centred on the quake: its words begin right of it, never over it
			}
		}
		if !found {
			t.Errorf("zoom %v: the ring's label is not beside it:\n%s", z, colours.ReplaceAllString(frame, ""))
		}
		counts = append(counts, strings.Count(frame, fg))
	}
	if counts[0] < 4 || counts[1] > counts[0]*3/2+2 || counts[1] < counts[0]*2/3-2 {
		t.Errorf("the ring's cells are %v at zoom 4 and 7; want the same size on the screen", counts)
	}
}

// TestAScreenRingIsChecked is L-19.2: a radius in dots or in km, never both;
// in dots, from 1 to 48.
func TestAScreenRingIsChecked(t *testing.T) {
	m := reportMap(t)
	for _, c := range []struct {
		dots int
		km   float64
	}{{8, 10}, {49, 0}, {-1, 0}} {
		if _, err := m.Set(quakeRing(c.dots, c.km)); !isKind(err, fault.InvalidCoordinates) {
			t.Errorf("a ring of %d dots and %v km: %v; want refused", c.dots, c.km, err)
		}
	}
	if _, err := m.Set(quakeRing(48, 0)); err != nil {
		t.Errorf("a ring of 48 dots was refused: %v", err)
	}
}

// TestTheQuakesColoursAreTokens is L-19.3: the past hour's, the past day's
// and older, named and settable.
func TestTheQuakesColoursAreTokens(t *testing.T) {
	names := " " + strings.Join(tuimaps.TokenNames(), " ") + " "
	for _, n := range []string{"quake.hour", "quake.day", "quake.older"} {
		if !strings.Contains(names, " "+n+" ") {
			t.Errorf("the tokens lack %s", n)
		}
	}
	_ = []tuimaps.Token{tuimaps.QuakeHour, tuimaps.QuakeDay, tuimaps.QuakeOlder}
}
