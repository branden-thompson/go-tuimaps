package main

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// scenarioFiles are the M1 scenarios, carried inside the app so that a
// person judging M1a needs nothing but the program. They are copies of the
// files the independent answer key was computed against, and a test holds
// them to being byte for byte the same.
//
//go:embed scenarios/*.json
var scenarioFiles embed.FS

// scenarioValid is the instant the scenarios' data is valid at. They carry
// hand-made data with no time of its own, so the app says "now": a frame
// that marked every scenario stale would teach the wrong thing about the
// staleness mark.
func scenarioValid() time.Time { return time.Now() }

// scenario is one of the M1 scenarios as its file is written.
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
	Pictures []struct {
		ID        string  `json:"id"`
		West      float64 `json:"west"`
		South     float64 `json:"south"`
		East      float64 `json:"east"`
		North     float64 `json:"north"`
		PNG       string  `json:"png_base64"`
		Tolerance float64 `json:"tolerance"`
		Table     []struct {
			R       uint8   `json:"r"`
			G       uint8   `json:"g"`
			B       uint8   `json:"b"`
			Value   float64 `json:"value"`
			Missing bool    `json:"missing"`
		} `json:"table"`
	} `json:"pictures"`

	places []tuimaps.Place
}

// loadScenario reads one scenario out of the app.
func loadScenario(n int) (*scenario, error) {
	if n < 1 || n > 7 {
		return nil, fmt.Errorf("there is no M1 scenario %d", n)
	}
	body, err := scenarioFiles.ReadFile("scenarios/scenario-" + strconv.Itoa(n) + ".json")
	if err != nil {
		return nil, fmt.Errorf("M1 scenario %d is not carried by this app", n)
	}
	var s scenario
	if err := json.Unmarshal(body, &s); err != nil {
		return nil, fmt.Errorf("M1 scenario %d cannot be read: %w", n, err)
	}
	for _, p := range s.Places {
		s.places = append(s.places, tuimaps.Place{Name: p.Name, At: tuimaps.LonLat{Lon: p.Lon, Lat: p.Lat}})
	}
	if len(s.places) == 0 {
		return nil, fmt.Errorf("M1 scenario %d names no place", n)
	}
	return &s, nil
}

// onto hands the scenario's overlays to a map, each in the shape the
// library takes it in.
func (s *scenario) onto(m *tuimaps.Map) error {
	if s == nil || m == nil {
		return errNothingToDescribe()
	}
	valid := scenarioValid()
	for _, shape := range s.Shapes {
		if _, err := m.Set(featureOverlay(shape.ID, shape.Kind, shape.Rings, valid)); err != nil {
			return fmt.Errorf("%s: %w", shape.ID, err)
		}
	}
	for _, f := range s.Fields {
		grid := tuimaps.Grid{West: f.West, South: f.South, East: f.East, North: f.North,
			Cols: f.Cols, Rows: f.Rows, Values: f.Values}
		if _, err := m.Set(tuimaps.TemperatureGrid(f.ID, grid, tuimaps.Celsius, valid)); err != nil {
			return fmt.Errorf("%s: %w", f.ID, err)
		}
	}
	for _, picture := range s.Pictures {
		png, err := base64.StdEncoding.DecodeString(picture.PNG)
		if err != nil {
			return fmt.Errorf("%s: its picture is not written as bytes", picture.ID)
		}
		image := tuimaps.Image{West: picture.West, South: picture.South, East: picture.East, North: picture.North,
			Projection: tuimaps.PlateCarree, PNG: png, Tolerance: picture.Tolerance}
		for _, e := range picture.Table {
			image.Table = append(image.Table, tuimaps.TableEntry{
				Colour: tuimaps.RGB{R: e.R, G: e.G, B: e.B}, Value: e.Value, Missing: e.Missing})
		}
		if _, err := m.Set(tuimaps.RadarImage(picture.ID, image, valid)); err != nil {
			return fmt.Errorf("%s: %w", picture.ID, err)
		}
	}
	return nil
}

// featureOverlay turns one of a scenario's shapes into an overlay: an area
// is an alert, a point or a line is a track, which is how the scenarios
// describe them.
func featureOverlay(id, kind string, rings [][][]float64, valid time.Time) tuimaps.Overlay {
	feature := tuimaps.Feature{Label: id, Role: tuimaps.AlertSevere, Kind: tuimaps.Polygon}
	switch kind {
	case "area":
	case "point":
		feature.Kind, feature.Role = tuimaps.Point, tuimaps.Track
	default:
		feature.Kind, feature.Role = tuimaps.Line, tuimaps.Track
	}
	for _, ring := range rings {
		run := make([]tuimaps.LonLat, 0, len(ring)+1)
		for _, at := range ring {
			if len(at) < 2 {
				continue
			}
			run = append(run, tuimaps.LonLat{Lon: at[0], Lat: at[1]})
		}
		if feature.Kind == tuimaps.Polygon && len(run) > 0 {
			run = append(run, run[0]) // the library takes a closed ring
		}
		feature.Rings = append(feature.Rings, run)
	}
	return tuimaps.Overlay{ID: id, Valid: valid, Keeps: time.Hour, Credit: "M1 scenario data", Features: []tuimaps.Feature{feature}}
}
