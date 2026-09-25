package tuimaps_test

// motion_test.go — v0.2.0 L5.5 (L-1.12, D-42, D-72): where the heavier rain
// was at the oldest usable frame and where it is at the newest, relative to a
// named place or, with none, to the view's centre - observation, never
// forecast.

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

const (
	motionWest, motionSouth, motionEast, motionNorth = -92.0, 34.0, -90.0, 36.0
	motionSize                                       = 100 // pixels a side: 0.02 degrees a pixel
)

// blobs is a picture of heavy rain in squares of six pixels, one centred on
// each longitude and latitude given, light rain around it, clear elsewhere.
func blobs(t *testing.T, centres ...tuimaps.LonLat) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, motionSize, motionSize))
	for _, c := range centres {
		cx := int((c.Lon - motionWest) / (motionEast - motionWest) * motionSize)
		cy := int((motionNorth - c.Lat) / (motionNorth - motionSouth) * motionSize)
		for dy := -6; dy < 6; dy++ {
			for dx := -6; dx < 6; dx++ {
				ink := color.NRGBA{G: 200, A: 255} // light
				if dx >= -3 && dx < 3 && dy >= -3 && dy < 3 {
					ink = color.NRGBA{R: 200, A: 255} // heavy
				}
				img.Set(cx+dx, cy+dy, ink)
			}
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// motionLoop is a loop of the given frames, oldest first.
func motionLoop(frames ...tuimaps.LoopFrame) tuimaps.Overlay {
	return tuimaps.RadarImage("radar", tuimaps.Image{Frames: frames, West: motionWest, South: motionSouth, East: motionEast, North: motionNorth,
		Projection: tuimaps.PlateCarree, Exact: true, Table: []tuimaps.TableEntry{{Colour: tuimaps.RGB{G: 200}, Value: 20}, {Colour: tuimaps.RGB{R: 200}, Value: 55}}}, noon)
}

func motionMap(t *testing.T, o tuimaps.Overlay, places ...tuimaps.Place) tuimaps.Report {
	t.Helper()
	m := world(t, 80, 24)
	must(t, m.Recentre(tuimaps.LonLat{Lon: -91, Lat: 35}))
	must(t, m.Zoom(6))
	mustSet(t, m, o)
	settle(t, m)
	r, err := m.Report(places)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// kmEast is a longitude a distance east of another at 35 north.
func kmEast(lon, km float64) float64 { return lon + km/(111.32*math.Cos(35*math.Pi/180)) }

// TestMotionTowardAPlace is L5.5: a cell 12 km further east after ten
// minutes, toward a place east of it: from about 32 km to about 20 km, came
// closer, over ten minutes, at the heavy class.
func TestMotionTowardAPlace(t *testing.T) {
	start := tuimaps.LonLat{Lon: -91.5, Lat: 35}
	then := tuimaps.LonLat{Lon: kmEast(start.Lon, 12), Lat: 35}
	home := tuimaps.Place{ID: "home", Name: "Home", At: tuimaps.LonLat{Lon: kmEast(then.Lon, 20), Lat: 35}}
	r := motionMap(t, motionLoop(
		tuimaps.LoopFrame{Valid: noon.Add(-10 * time.Minute), PNG: blobs(t, start)},
		tuimaps.LoopFrame{Valid: noon, PNG: blobs(t, then)},
	), home)
	if len(r.Motion) != 1 {
		t.Fatalf("motion: %+v; want one report, for Home", r.Motion)
	}
	mo := r.Motion[0]
	if mo.Place != "Home" || mo.Overlay != "radar" || mo.Trend != tuimaps.Closer || mo.Span != 10*time.Minute {
		t.Errorf("motion: %+v; want Home, radar, closer, ten minutes", mo)
	}
	if math.Abs(mo.From.Distance-32) > 2.5 || math.Abs(mo.To.Distance-20) > 2.5 || mo.To.Unit != "kilometres" || mo.To.Compass != "west" {
		t.Errorf("from %.1f to %.1f %s, %s; want about 32 to about 20 kilometres, to the west", mo.From.Distance, mo.To.Distance, mo.To.Unit, mo.To.Compass)
	}
	if !mo.From.Valid.Equal(noon.Add(-10*time.Minute)) || !mo.To.Valid.Equal(noon) || mo.Threshold < 1 {
		t.Errorf("sightings at %v and %v, threshold class %d", mo.From.Valid, mo.To.Valid, mo.Threshold)
	}
}

// TestEachPlaceGetsItsNearestCell: two cells, one near each place; each
// place's report follows its own.
func TestEachPlaceGetsItsNearestCell(t *testing.T) {
	west, east := tuimaps.LonLat{Lon: -91.7, Lat: 35.4}, tuimaps.LonLat{Lon: -90.4, Lat: 34.6}
	a := tuimaps.Place{ID: "a", Name: "A", At: tuimaps.LonLat{Lon: -91.6, Lat: 35.4}}
	b := tuimaps.Place{ID: "b", Name: "B", At: tuimaps.LonLat{Lon: -90.3, Lat: 34.6}}
	r := motionMap(t, motionLoop(
		tuimaps.LoopFrame{Valid: noon.Add(-10 * time.Minute), PNG: blobs(t, west, east)},
		tuimaps.LoopFrame{Valid: noon, PNG: blobs(t, west, east)},
	), a, b)
	if len(r.Motion) != 2 {
		t.Fatalf("motion: %+v; want one for each place", r.Motion)
	}
	for _, mo := range r.Motion {
		if mo.To.Distance > 15 || mo.Trend != tuimaps.Held {
			t.Errorf("%s: %+v; want its own cell, within 15 km, held", mo.Place, mo)
		}
	}
}

// TestAGapAtTheOldestFrame: the oldest usable frame is the oldest that is
// not a gap.
func TestAGapAtTheOldestFrame(t *testing.T) {
	c := tuimaps.LonLat{Lon: -91, Lat: 35}
	home := tuimaps.Place{ID: "home", Name: "Home", At: tuimaps.LonLat{Lon: -90.5, Lat: 35}}
	r := motionMap(t, motionLoop(
		tuimaps.LoopFrame{Valid: noon.Add(-20 * time.Minute), Gap: true},
		tuimaps.LoopFrame{Valid: noon.Add(-10 * time.Minute), PNG: blobs(t, c)},
		tuimaps.LoopFrame{Valid: noon, PNG: blobs(t, c)},
	), home)
	if len(r.Motion) != 1 || !r.Motion[0].From.Valid.Equal(noon.Add(-10*time.Minute)) || r.Motion[0].Span != 10*time.Minute {
		t.Errorf("a gap at the oldest frame: %+v; want the sighting from -10 minutes", r.Motion)
	}
}

// TestAForecastIsNeverASighting (D-42): motion is observation; a forecast
// frame after the newest observed one is not where the rain is.
func TestAForecastIsNeverASighting(t *testing.T) {
	c := tuimaps.LonLat{Lon: -91, Lat: 35}
	home := tuimaps.Place{ID: "home", Name: "Home", At: tuimaps.LonLat{Lon: -90.5, Lat: 35}}
	r := motionMap(t, motionLoop(
		tuimaps.LoopFrame{Valid: noon.Add(-10 * time.Minute), PNG: blobs(t, c)},
		tuimaps.LoopFrame{Valid: noon, PNG: blobs(t, c)},
		tuimaps.LoopFrame{Valid: noon.Add(10 * time.Minute), PNG: blobs(t, tuimaps.LonLat{Lon: -90.6, Lat: 35}), Forecast: true},
	), home)
	if len(r.Motion) != 1 || !r.Motion[0].To.Valid.Equal(noon) || r.Motion[0].Trend != tuimaps.Held {
		t.Errorf("with a forecast frame after the newest observed: %+v; want the newest sighting at noon, held", r.Motion)
	}
}

// TestMotionWithNoPlaceIsRelativeToTheView (D-42): with no place, the report
// is relative to the view's centre, and its Place is empty.
func TestMotionWithNoPlaceIsRelativeToTheView(t *testing.T) {
	start := tuimaps.LonLat{Lon: -91.5, Lat: 35}
	r := motionMap(t, motionLoop(
		tuimaps.LoopFrame{Valid: noon.Add(-10 * time.Minute), PNG: blobs(t, start)},
		tuimaps.LoopFrame{Valid: noon, PNG: blobs(t, tuimaps.LonLat{Lon: kmEast(start.Lon, 12), Lat: 35})},
	))
	if len(r.Motion) != 1 || r.Motion[0].Place != "" || r.Motion[0].Trend != tuimaps.Closer {
		t.Fatalf("motion with no place: %+v; want one, relative to the view's centre, closer", r.Motion)
	}
	if d := r.Motion[0].To.Distance; math.Abs(d-kmFromCentre(kmEast(start.Lon, 12))) > 2.5 {
		t.Errorf("to %.1f km from the centre", d)
	}
}

// kmFromCentre is how far a longitude at 35 north is from the view's
// centre, at -91.
func kmFromCentre(lon float64) float64 { return math.Abs(lon+91) * 111.32 * math.Cos(35*math.Pi/180) }

// TestAHostTypesHeavierRainIsItsTopThird: for a type of the host's own, the
// heavier rain is the top third of its classes - here six classes, so the
// fifth and sixth (indexes 4 and 5) - and a cell of them is followed.
func TestAHostTypesHeavierRainIsItsTopThird(t *testing.T) {
	c := tuimaps.LonLat{Lon: -91, Lat: 35}
	o := motionLoop(
		tuimaps.LoopFrame{Valid: noon.Add(-10 * time.Minute), PNG: blobs(t, c)},
		tuimaps.LoopFrame{Valid: noon, PNG: blobs(t, c)},
	)
	o.Image.Type = tuimaps.Type{Unit: "mm/h", Breaks: []float64{1, 5, 10, 20, 50}}
	r := motionMap(t, o, tuimaps.Place{ID: "home", Name: "Home", At: tuimaps.LonLat{Lon: -90.5, Lat: 35}})
	if len(r.Motion) != 1 || r.Motion[0].Threshold != 4 {
		t.Errorf("a host type of six classes: %+v; want the heavier rain from class 4", r.Motion)
	}
}
