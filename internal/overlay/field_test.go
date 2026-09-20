package overlay

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

func temperatures(cols, rows int, at func(col, row int) float64) Overlay {
	values := make([]float64, 0, cols*rows)
	for row := range rows {
		for col := range cols {
			values = append(values, at(col, row))
		}
	}
	return Overlay{ID: "temperature", Valid: noon, Keeps: time.Hour, Credit: "Open-Meteo",
		Grid: &Grid{West: -110, South: 25, East: -80, North: 50, Cols: cols, Rows: rows, Values: values, Type: Type{Preset: "temperature", Unit: "C"}}}
}

// TestGridHandIn: a scalar grid is a regular longitude and latitude grid of
// values with a unit and a type (FR-7); its likely mistakes each get a kind.
func TestGridHandIn(t *testing.T) {
	s := store(t)
	good := temperatures(30, 25, func(col, row int) float64 { return float64(col) - 10 })
	if res, err := s.Set(good); err != nil || !res.Created {
		t.Fatalf("%+v, %v", res, err)
	}
	cases := []struct {
		name  string
		spoil func(o *Overlay)
		kind  fault.Kind
		says  string
	}{
		{"fewer values than cells", func(o *Overlay) { o.Grid.Values = o.Grid.Values[:100] }, fault.SizeMismatch, "750"},
		{"no columns", func(o *Overlay) { o.Grid.Cols = 0 }, fault.SizeMismatch, "column"},
		{"west of east the wrong way round", func(o *Overlay) { o.Grid.West, o.Grid.East = o.Grid.East, o.Grid.West }, fault.InvalidCoordinates, "west"},
		{"a bound that is no number", func(o *Overlay) { o.Grid.North = math.NaN() }, fault.InvalidCoordinates, "number"},
		{"off the globe", func(o *Overlay) { o.Grid.North = 95 }, fault.InvalidCoordinates, "globe"},
		{"an unknown preset", func(o *Overlay) { o.Grid.Type.Preset = "rainbow" }, fault.UnknownPreset, "preset"},
		{"its own type and no breaks", func(o *Overlay) { o.Grid.Type = Type{Unit: "mm"} }, fault.UnsortedBreaks, "breaks"},
		{"a grid and features together", func(o *Overlay) { o.Features = alert("x", square(-95, 38, 1)).Features }, fault.SizeMismatch, "one"},
		{"more cells than a grid may have", func(o *Overlay) { o.Grid.Cols, o.Grid.Rows, o.Grid.Values = 2000, 2000, make([]float64, 4_000_000) }, fault.OverImageCap, "1,048,576"},
	}
	for _, c := range cases {
		o := temperatures(30, 25, func(col, row int) float64 { return 12 })
		c.spoil(&o)
		_, err := store(t).Set(o)
		if !isKind(err, c.kind) || !strings.Contains(err.Error(), c.says) {
			t.Errorf("%s: %v; want the %v kind, saying %q", c.name, err, c.kind, c.says)
		}
	}
}

// TestImplausibleUnitWarned is part of 10.25: values that make no sense in
// the unit declared are accepted, with a warning.
func TestImplausibleUnitWarned(t *testing.T) {
	s := store(t)
	fahrenheitAsCelsius := temperatures(10, 10, func(col, row int) float64 { return 68 + float64(col) })
	if _, err := s.Set(fahrenheitAsCelsius); err != nil {
		t.Fatalf("implausible values are accepted: %v", err)
	}
	w := s.TakeWarnings()
	if len(w) != 1 || w[0].Kind != fault.ImplausibleUnit {
		t.Errorf("temperatures all above 60 declared as C: warnings %+v", w)
	}
	s.Set(temperatures(10, 10, func(col, row int) float64 { return 12 + float64(col) }))
	if w := s.TakeWarnings(); len(w) != 0 {
		t.Errorf("plausible temperatures: %+v", w)
	}
	withGaps := temperatures(10, 10, func(col, row int) float64 {
		if col == 0 {
			return math.NaN()
		}
		return 15
	})
	s.Set(withGaps)
	if w := s.TakeWarnings(); len(w) != 0 {
		t.Errorf("values that are no number are no data, not a wrong unit: %+v", w)
	}
}

// TestGridPrepared: preparing a grid classifies each value once, on the
// host's own grid. That is where preparing ends: sampling it into cells is
// the renderer's, at draw time (L2 Overlays).
func TestGridPrepared(t *testing.T) {
	s := store(t)
	s.Set(temperatures(4, 2, func(col, row int) float64 {
		if col == 3 {
			return math.Inf(1)
		}
		return []float64{-35, -0.5, 0, 46}[col] + float64(row)*0.1
	}))
	if err := s.PrepareJob("temperature", 5).Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	field, ok := s.Field("temperature")
	if !ok || field.Cols != 4 || field.Rows != 2 || len(field.Classes) != 8 {
		t.Fatalf("%+v, %v", field, ok)
	}
	want := []int8{0, 6, 7, -1, 0, 6, 7, -1}
	for i, c := range field.Classes {
		if c != want[i] {
			t.Errorf("cell %d is class %d, want %d", i, c, want[i])
		}
	}
	if field.Preset != uint8(colour.Temperature) || field.ClassCount != 17 || field.West != -110 || field.North != 50 {
		t.Errorf("%+v", field)
	}
	if shapes, _, path := s.Drawn("temperature", 5); path != NotReady || shapes != nil {
		t.Errorf("a grid has no shapes: %v", path)
	}
	s.Remove("temperature")
	if _, ok := s.Field("temperature"); ok {
		t.Error("the prepared field outlived its overlay")
	}
}
