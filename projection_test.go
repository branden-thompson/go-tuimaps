package tuimaps_test

// projection_test.go — F-3: an image handed in as web Mercator is answered in
// web Mercator, as it is drawn.

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

// A WEB MERCATOR IMAGE'S ANSWERS ARE ITS PICTURE'S (F-3): heavy rain in the
// top fifth of an image from the equator to 70 north reaches, in the web
// map's height, down to about 59 north. At 58 north the picture is light
// rain - the answer a reading in equal steps of latitude would get wrong.
func TestAWebMercatorImagesAnswersAreItsPictures(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 20, 10))
	for y := range 10 {
		ink := color.NRGBA{G: 200, A: 255} // light
		if y < 2 {
			ink = color.NRGBA{R: 200, A: 255} // heavy
		}
		for x := range 20 {
			img.Set(x, y, ink)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	o := tuimaps.RadarImage("radar", tuimaps.Image{PNG: buf.Bytes(), West: -10, South: 0, East: 10, North: 70,
		Projection: tuimaps.WebMercator, Exact: true, Table: []tuimaps.TableEntry{{Colour: tuimaps.RGB{G: 200}, Value: 20}, {Colour: tuimaps.RGB{R: 200}, Value: 55}}}, noon)
	m := world(t, 80, 24)
	must(t, m.Recentre(tuimaps.LonLat{Lon: 0, Lat: 35}))
	must(t, m.Zoom(2))
	mustSet(t, m, o)
	settle(t, m)
	places := []tuimaps.Place{{Name: "Heavy", At: tuimaps.LonLat{Lon: 0, Lat: 69}}, {Name: "Light", At: tuimaps.LonLat{Lon: 0, Lat: 40}}, {Name: "Edge", At: tuimaps.LonLat{Lon: 0, Lat: 58}}}
	r, err := m.Report(places)
	if err != nil {
		t.Fatal(err)
	}
	class := map[string]int{}
	for _, p := range r.Places {
		for _, a := range p.Answers {
			if a.Overlay == "radar" {
				class[a.Place] = a.Class
			}
		}
	}
	if len(class) != 3 || class["Heavy"] == class["Light"] {
		t.Fatalf("the answers do not tell heavy from light: %v", class)
	}
	if class["Edge"] != class["Light"] {
		t.Errorf("at 58 north the picture is light rain; the answer says class %d (heavy is %d, light %d)", class["Edge"], class["Heavy"], class["Light"])
	}
}

// A WEB MERCATOR LOOP'S SIGHTINGS ARE WHERE ITS PICTURE PUTS THEM (F-3): a
// cell drawn at 60 north on a loop from 50 to 70 north, in the web map's
// height, is sighted at 60 north - not a degree and a half south of it, where
// a reading in equal steps of latitude would put it.
func TestAWebMercatorLoopsSightingsAreWhereItsPicturePutsThem(t *testing.T) {
	const west, south, east, north, size = -92.0, 50.0, -90.0, 70.0, 100
	h := func(lat float64) float64 { return math.Atanh(math.Sin(lat * math.Pi / 180)) }
	frame := func(at tuimaps.LonLat) []byte {
		img := image.NewNRGBA(image.Rect(0, 0, size, size))
		cx := int((at.Lon - west) / (east - west) * size)
		cy := int((h(north) - h(at.Lat)) / (h(north) - h(south)) * size)
		for dy := -3; dy < 3; dy++ {
			for dx := -3; dx < 3; dx++ {
				img.Set(cx+dx, cy+dy, color.NRGBA{R: 200, A: 255})
			}
		}
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			t.Fatal(err)
		}
		return buf.Bytes()
	}
	cell := tuimaps.LonLat{Lon: -91, Lat: 60}
	o := tuimaps.RadarImage("radar", tuimaps.Image{Frames: []tuimaps.LoopFrame{
		{Valid: noon.Add(-10 * time.Minute), PNG: frame(cell)}, {Valid: noon, PNG: frame(cell)},
	}, West: west, South: south, East: east, North: north, Projection: tuimaps.WebMercator, Exact: true,
		Table: []tuimaps.TableEntry{{Colour: tuimaps.RGB{G: 200}, Value: 20}, {Colour: tuimaps.RGB{R: 200}, Value: 55}}}, noon)
	m := world(t, 80, 24)
	must(t, m.Recentre(tuimaps.LonLat{Lon: -91, Lat: 60}))
	must(t, m.Zoom(6))
	mustSet(t, m, o)
	settle(t, m)
	r, err := m.Report([]tuimaps.Place{{ID: "camp", Name: "Camp", At: tuimaps.LonLat{Lon: -90.5, Lat: 60}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Motion) != 1 {
		t.Fatalf("motion: %+v; want one report", r.Motion)
	}
	if got := r.Motion[0].To.At.Lat; math.Abs(got-60) > 0.5 {
		t.Errorf("the cell drawn at 60 north is sighted at %.2f north", got)
	}
}
