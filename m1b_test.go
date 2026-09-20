package tuimaps_test

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// scenario is the shape of a scenario file, as the answer-key tool reads
// it. It is written out here rather than imported, because the tool shares
// no code with the library: that is what makes its key independent (D-43).
type scenario struct {
	Name   string `json:"name"`
	Places []struct {
		Name string  `json:"name"`
		Lon  float64 `json:"lon"`
		Lat  float64 `json:"lat"`
	} `json:"places"`
	Shapes []struct {
		ID    string        `json:"id"`
		Kind  string        `json:"kind"`
		Rings [][][]float64 `json:"rings"`
	} `json:"shapes"`
	Fields []struct {
		ID     string    `json:"id"`
		West   float64   `json:"west"`
		South  float64   `json:"south"`
		East   float64   `json:"east"`
		North  float64   `json:"north"`
		Cols   int       `json:"cols"`
		Rows   int       `json:"rows"`
		Values []float64 `json:"values"`
	} `json:"fields"`
	Images []struct {
		ID string `json:"id"`
	} `json:"images"`
}

// wholeKey is the key as the tool writes it: the shapes, the fields and
// the images.
type wholeKey struct {
	Shapes []keyed `json:"shapes"`
	Fields []struct {
		Place string  `json:"place"`
		Field string  `json:"field"`
		Value float64 `json:"value"`
		Band  int     `json:"band"`
		Rises string  `json:"rises"`
		Flat  bool    `json:"flat"`
	} `json:"fields"`
}

// keyed is one answer of the independent key.
type keyed struct {
	Place   string  `json:"place"`
	Shape   string  `json:"shape"`
	Inside  *bool   `json:"inside,omitempty"`
	Km      float64 `json:"km"`
	Bearing float64 `json:"bearing_deg"`
	Compass string  `json:"compass"`
}

// TestM1bAgreesWithTheIndependentKey is plan task 11.13, and the metric M1b
// itself (D-67): the library's own description must equal a key worked out
// by a separate program, over the same data, sharing no code with it.
func TestM1bAgreesWithTheIndependentKey(t *testing.T) {
	for _, name := range []string{"hand-made", "scenario-1", "scenario-3", "scenario-4"} {
		t.Run(name, func(t *testing.T) { againstKey(t, name) })
	}
}

// againstKey runs one scenario against its own key.
func againstKey(t *testing.T, name string) {
	dir := filepath.Join("06_docs", "02_features", "go-tuimaps", "02-analysis", "scenarios")
	var s scenario
	read(t, filepath.Join(dir, name+".json"), &s)
	var whole wholeKey
	read(t, filepath.Join(dir, name+"-key.json"), &whole)
	key := whole.Shapes
	if len(s.Images) > 0 {
		// The library takes an image as a picture and a table and classifies
		// it; this key takes the classes themselves. Comparing them would
		// compare two different inputs, so an image scenario is not compared
		// here, and says so rather than passing quietly.
		t.Skip("an image scenario: the library classifies a picture, the key is given classes")
	}
	if len(key)+len(whole.Fields) == 0 {
		t.Fatal("the key is empty")
	}

	m := world(t, 149, 38)
	places := make([]tuimaps.Place, 0, len(s.Places))
	for _, p := range s.Places {
		places = append(places, tuimaps.Place{Name: p.Name, At: tuimaps.LonLat{Lon: p.Lon, Lat: p.Lat}})
	}
	if _, err := m.SetPlaces(places); err != nil {
		t.Fatal(err)
	}
	for _, shape := range s.Shapes {
		if _, err := m.Set(overlayOf(t, shape.ID, shape.Kind, shape.Rings)); err != nil {
			t.Fatalf("%s: %v", shape.ID, err)
		}
	}
	for _, f := range s.Fields {
		grid := tuimaps.Grid{West: f.West, South: f.South, East: f.East, North: f.North,
			Cols: f.Cols, Rows: f.Rows, Values: f.Values}
		if _, err := m.Set(tuimaps.TemperatureGrid(f.ID, grid, tuimaps.Celsius, noon)); err != nil {
			t.Fatalf("%s: %v", f.ID, err)
		}
	}
	got, err := m.Describe(nil)
	if err != nil {
		t.Fatal(err)
	}
	// The fields: the value, the band and which way it rises.
	byPlace := map[string][]tuimaps.Answer{}
	for _, one := range got {
		byPlace[one.Place] = one.Answers
	}
	for _, want := range whole.Fields {
		var mine tuimaps.Answer
		for _, a := range byPlace[want.Place] {
			if a.Overlay == want.Field {
				mine = a
			}
		}
		if mine.Overlay == "" {
			t.Errorf("the library says nothing about %s and %s", want.Place, want.Field)
			continue
		}
		if math.Abs(mine.Value-want.Value) > 0.001 {
			t.Errorf("%s in %s: the library reads %v, the key reads %v", want.Place, want.Field, mine.Value, want.Value)
		}
		if mine.Band != want.Band {
			t.Errorf("%s in %s: the library says band %d, the key says %d", want.Place, want.Field, mine.Band, want.Band)
		}
		if mine.Rises != want.Rises {
			t.Errorf("%s in %s: the library says it rises %q, the key says %q", want.Place, want.Field, mine.Rises, want.Rises)
		}
	}

	// Every answer of the key has one of the library's, and they agree.
	answers := map[[2]string]tuimaps.Answer{}
	for _, one := range got {
		for _, a := range one.Answers {
			answers[[2]string{one.Place, a.Overlay}] = a
		}
	}
	for _, want := range key {
		mine, ok := answers[[2]string{want.Place, want.Shape}]
		if !ok {
			t.Errorf("the library says nothing about %s and %s", want.Place, want.Shape)
			continue
		}
		if want.Inside != nil {
			inside := mine.Relation.String() == "inside"
			if inside != *want.Inside {
				t.Errorf("%s in %s: the library says %v, the key says %v", want.Place, want.Shape, inside, *want.Inside)
			}
		}
		// The key steps along each edge a hundred times; the library solves
		// for the nearest point. They agree to within that step.
		if diff := math.Abs(mine.Distance - want.Km); diff > math.Max(1, want.Km*0.01) {
			t.Errorf("%s to %s: the library says %.1f km, the key says %.1f km", want.Place, want.Shape, mine.Distance, want.Km)
		}
		if mine.Compass != want.Compass {
			t.Errorf("%s to %s: the library says %q, the key says %q (%.0f vs %.0f degrees)",
				want.Place, want.Shape, mine.Compass, want.Compass, mine.Bearing, want.Bearing)
		}
	}
}

// overlayOf turns a scenario's shape into an overlay the library takes.
func overlayOf(t *testing.T, id, kind string, rings [][][]float64) tuimaps.Overlay {
	t.Helper()
	feature := tuimaps.Feature{Label: id, Role: tuimaps.AlertSevere}
	switch kind {
	case "area":
		feature.Kind = tuimaps.Polygon
	case "point":
		feature.Kind, feature.Role = tuimaps.Point, tuimaps.Track
	default:
		feature.Kind, feature.Role = tuimaps.Line, tuimaps.Track
	}
	for _, ring := range rings {
		run := make([]tuimaps.LonLat, 0, len(ring)+1)
		for _, at := range ring {
			run = append(run, tuimaps.LonLat{Lon: at[0], Lat: at[1]})
		}
		if feature.Kind == tuimaps.Polygon && len(run) > 0 {
			run = append(run, run[0]) // the library takes a closed ring
		}
		feature.Rings = append(feature.Rings, run)
	}
	return tuimaps.Overlay{ID: id, Valid: noon, Keeps: time.Hour, Credit: "hand-made", Features: []tuimaps.Feature{feature}}
}

func read(t *testing.T, path string, into any) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, into); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
}
