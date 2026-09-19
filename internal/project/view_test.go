package project

import (
	"math"
	"reflect"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// fixtureView is the pinned fixture's view: 149 by 38 cells at zoom 6.4 over
// Tampa Bay (memory-measurement.md).
var fixtureView = View{Centre: LonLat{Lon: -82.4, Lat: 27.6}, Zoom: 6.4, Cols: 149, Rows: 38}

// TestViewToDot is plan task 01.2.
func TestViewToDot(t *testing.T) {
	v := fixtureView
	x, y, err := v.ToDot(v.Centre)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(x-149) > 1e-6 || math.Abs(y-76) > 1e-6 {
		t.Errorf("the centre is at dot (%v, %v); a 149 by 38 view is 298 by 152 dots, so its centre is (149, 76)", x, y)
	}
	east := LonLat{Lon: v.Centre.Lon + 1, Lat: v.Centre.Lat}
	x6, _, _ := v.ToDot(east)
	closer := v
	closer.Zoom = v.Zoom + 1
	x7, _, _ := closer.ToDot(east)
	if got := (x7 - 149) / (x6 - 149); math.Abs(got-2) > 1e-9 {
		t.Errorf("one zoom level scales by %v, want 2", got)
	}
	north := LonLat{Lon: v.Centre.Lon, Lat: v.Centre.Lat + 1}
	if _, yn, _ := v.ToDot(north); yn >= 76 {
		t.Errorf("a point to the north is at row %v; north is up, so it must be above the centre", yn)
	}
}

// TestFractionalZoom is plan task 01.3.
func TestFractionalZoom(t *testing.T) {
	east := LonLat{Lon: -81.4, Lat: 27.6}
	at := func(zoom float64) float64 {
		v := fixtureView
		v.Zoom = zoom
		x, _, err := v.ToDot(east)
		if err != nil {
			t.Fatal(err)
		}
		return x - 149
	}
	if got := at(6.5) / at(6); math.Abs(got-math.Sqrt2) > 1e-9 {
		t.Errorf("half a zoom level scales by %v, want the square root of 2", got)
	}
	if got := at(6.999) / at(7); got >= 1 || got < 0.999 {
		t.Errorf("the scale is not continuous across a whole zoom: ratio %v", got)
	}
}

func TestDotRoundTrip(t *testing.T) {
	v := fixtureView
	for _, d := range [][2]float64{{0, 0}, {149, 76}, {298, 152}, {12.25, 140.5}} {
		p, err := v.FromDot(d[0], d[1])
		if err != nil {
			t.Fatal(err)
		}
		x, y, err := v.ToDot(p)
		if err != nil || math.Abs(x-d[0]) > 1e-6 || math.Abs(y-d[1]) > 1e-6 {
			t.Errorf("dot %v went to %v and came back as (%v, %v), %v", d, p, x, y, err)
		}
	}
}

// TestTilesForView is plan task 01.4: the fixture's view needs exactly the
// four zoom-6 tiles committed with the fixture, in a fixed order (NFR-6).
func TestTilesForView(t *testing.T) {
	got, err := fixtureView.Tiles()
	if err != nil {
		t.Fatal(err)
	}
	want := []scene.TileID{{Z: 6, X: 16, Y: 26}, {Z: 6, X: 16, Y: 27}, {Z: 6, X: 17, Y: 26}, {Z: 6, X: 17, Y: 27}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	again, _ := fixtureView.Tiles()
	if !reflect.DeepEqual(got, again) {
		t.Error("the order changed between two calls")
	}
}

func TestTilesStayInsideTheWorld(t *testing.T) {
	world := View{Centre: LonLat{}, Zoom: 0.2, Cols: 149, Rows: 38}
	got, err := world.Tiles()
	if err != nil || !reflect.DeepEqual(got, []scene.TileID{{Z: 0, X: 0, Y: 0}}) {
		t.Errorf("the world view needs %v, %v; want the world tile alone", got, err)
	}
	edge := View{Centre: LonLat{Lon: 179.9, Lat: 84}, Zoom: 3, Cols: 149, Rows: 38}
	tiles, err := edge.Tiles()
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range tiles {
		if err := id.Validate(); err != nil {
			t.Errorf("a view at the edge of the world asked for a tile that does not exist: %v", err)
		}
	}
	deep := View{Centre: LonLat{Lon: 2.35, Lat: 48.86}, Zoom: 17.5, Cols: 69, Rows: 12}
	tiles, err = deep.Tiles()
	if err != nil || len(tiles) == 0 {
		t.Fatalf("%v, %v", tiles, err)
	}
	for _, id := range tiles {
		if id.Z != MaxSourceZoom {
			t.Errorf("a view closer than the deepest source zoom asked for zoom %d; it is drawn from zoom-14 tiles, stretched", id.Z)
		}
	}
}

func TestViewValidate(t *testing.T) {
	bad := []View{
		{Centre: LonLat{}, Zoom: 6, Cols: 0, Rows: 38},
		{Centre: LonLat{}, Zoom: 6, Cols: 149, Rows: -1},
		{Centre: LonLat{}, Zoom: 6, Cols: MaxCells + 1, Rows: 38},
		{Centre: LonLat{}, Zoom: math.NaN(), Cols: 149, Rows: 38},
		{Centre: LonLat{}, Zoom: 40, Cols: 149, Rows: 38},
		{Centre: LonLat{Lon: math.Inf(1)}, Zoom: 6, Cols: 149, Rows: 38},
		{Centre: LonLat{Lat: 91}, Zoom: 6, Cols: 149, Rows: 38},
	}
	for _, v := range bad {
		if err := v.Validate(); err == nil {
			t.Errorf("%+v passed validation", v)
		}
		if _, err := v.Tiles(); err == nil {
			t.Errorf("%+v: Tiles did not refuse it", v)
		}
	}
	if err := fixtureView.Validate(); err != nil {
		t.Errorf("the fixture's view was refused: %v", err)
	}
}
