package project

import (
	"errors"
	"math"
	"os"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

func TestMain(m *testing.M) { os.Exit(testkit.Main(m)) }

var points = []LonLat{
	{0, 0}, {-85.14, 41.08}, {-82.4, 27.6}, {139.69, 35.68}, {2.35, 48.86}, {-0.13, 51.51},
	{179.999999, 0}, {-180, 0}, {180, 10}, {0, 85}, {0, -85}, {-122.4, 85.0511}, {151.2, -85.0511}, {12.5, 0.000001},
}

// TestMercatorRoundTrip is plan task 01.1.
func TestMercatorRoundTrip(t *testing.T) {
	for _, zoom := range []uint8{0, 6, 14, 22} {
		for _, p := range points {
			x, y, err := ToTile(p, zoom)
			if err != nil {
				t.Fatalf("ToTile(%v, %d): %v", p, zoom, err)
			}
			back, err := FromTile(x, y, zoom)
			if err != nil {
				t.Fatalf("FromTile(%v, %v, %d): %v", x, y, zoom, err)
			}
			if math.Abs(back.Lon-p.Lon) > 1e-9 || math.Abs(back.Lat-p.Lat) > 1e-9 {
				t.Errorf("zoom %d: %v went to (%v, %v) and came back as %v", zoom, p, x, y, back)
			}
		}
	}
}

// upstream is TerminalMap's own forward projection, utils.rs ll2tile, written
// with the logarithm and tangent it uses.
func upstream(lon, lat float64, zoom uint8) (float64, float64) {
	n := math.Exp2(float64(zoom))
	rad := lat * math.Pi / 180
	return (lon + 180) / 360 * n, (1 - math.Log(math.Tan(rad)+1/math.Cos(rad))/math.Pi) / 2 * n
}

// TestParityP20_Projection: standard web mercator, equal to upstream's
// formula although written with different functions (constants, section 6).
func TestParityP20_Projection(t *testing.T) {
	for _, zoom := range []uint8{0, 3, 6, 14} {
		for _, p := range points {
			x, y, err := ToTile(p, zoom)
			if err != nil {
				t.Fatal(err)
			}
			wantX, wantY := upstream(p.Lon, p.Lat, zoom)
			scale := math.Exp2(float64(zoom))
			if math.Abs(x-wantX) > 1e-9*scale || math.Abs(y-wantY) > 1e-9*scale {
				t.Errorf("zoom %d %v: got (%v, %v), upstream gives (%v, %v)", zoom, p, x, y, wantX, wantY)
			}
		}
	}
	if x, y, _ := ToTile(LonLat{0, 0}, 0); x != 0.5 || math.Abs(y-0.5) > 1e-15 {
		t.Errorf("the origin is at (%v, %v) of the world tile, want its centre", x, y)
	}
	if _, y, _ := ToTile(LonLat{0, MaxLatitude}, 0); math.Abs(y) > 1e-6 {
		t.Errorf("the clamp latitude is at row %v of the world tile, want its top edge", y)
	}
}

// TestParityP21_Normalize: longitude wraps by a single turn and latitude is
// clamped, as upstream (defect ledger L-17 c and e, replicated).
func TestParityP21_Normalize(t *testing.T) {
	cases := []struct{ in, want LonLat }{
		{LonLat{-85, 41}, LonLat{-85, 41}},
		{LonLat{190, 0}, LonLat{-170, 0}},
		{LonLat{-190, 0}, LonLat{170, 0}},
		{LonLat{180, 0}, LonLat{180, 0}},
		{LonLat{-180, 0}, LonLat{-180, 0}},
		{LonLat{550, 0}, LonLat{190, 0}}, // one turn only, as upstream
		{LonLat{0, 89}, LonLat{0, MaxLatitude}},
		{LonLat{0, -90}, LonLat{0, -MaxLatitude}},
	}
	for _, c := range cases {
		got, err := Normalize(c.in)
		if err != nil || got != c.want {
			t.Errorf("Normalize(%v) = %v, %v; want %v", c.in, got, err, c.want)
		}
	}
}

// TestParityP17_TileZoom: the tile zoom is the whole part of the view's
// zoom, held between 0 and the deepest zoom a source has.
func TestParityP17_TileZoom(t *testing.T) {
	cases := map[float64]uint8{-2: 0, 0: 0, 0.99: 0, 1: 1, 6.4: 6, 13.999: 13, 14: 14, 14.5: 14, 18: 14, 99: 14}
	for zoom, want := range cases {
		got, err := TileZoom(zoom)
		if err != nil || got != want {
			t.Errorf("TileZoom(%v) = %d, %v; want %d", zoom, got, err, want)
		}
	}
}

// TestParityP18_TilePixelSize: a tile is 256 dots at its own zoom and grows
// with the fraction; above the deepest source zoom it is stretched, to 4096
// at zoom 18.
func TestParityP18_TilePixelSize(t *testing.T) {
	cases := map[float64]float64{0: 256, 6: 256, 6.5: 256 * math.Sqrt2, 14: 256, 15: 512, 18: 4096, -1: 128}
	for zoom, want := range cases {
		got, err := TileSizeAt(zoom)
		if err != nil || math.Abs(got-want) > 1e-9 {
			t.Errorf("TileSizeAt(%v) = %v, %v; want %v", zoom, got, err, want)
		}
	}
}

// TestNoNonFiniteReachesInt is plan task 01.11.
func TestNoNonFiniteReachesInt(t *testing.T) {
	nan, inf := math.NaN(), math.Inf(1)
	bad := []LonLat{{nan, 0}, {0, nan}, {inf, 0}, {0, -inf}}
	if _, _, err := ToTile(LonLat{0, 90.5}, 6); !isInvalidCoordinates(err) {
		t.Errorf("a latitude beyond the pole: %v", err)
	}
	for _, p := range bad {
		if _, _, err := ToTile(p, 6); !isInvalidCoordinates(err) {
			t.Errorf("ToTile(%v): %v; want an invalid-coordinates error", p, err)
		}
		if _, err := Normalize(p); !isInvalidCoordinates(err) {
			t.Errorf("Normalize(%v): %v; want an invalid-coordinates error", p, err)
		}
	}
	for _, v := range []float64{nan, inf, -inf} {
		if _, err := FromTile(v, 0, 6); !isInvalidCoordinates(err) {
			t.Errorf("FromTile(%v, 0): %v", v, err)
		}
		if _, err := TileZoom(v); !isInvalidCoordinates(err) {
			t.Errorf("TileZoom(%v): %v", v, err)
		}
		if _, err := TileSizeAt(v); !isInvalidCoordinates(err) {
			t.Errorf("TileSizeAt(%v): %v", v, err)
		}
	}
	if _, _, err := ToTile(LonLat{0, 0}, 23); !isInvalidCoordinates(err) {
		t.Errorf("a zoom deeper than the library addresses: %v", err)
	}
}

func isInvalidCoordinates(err error) bool {
	var f *fault.Error
	return errors.As(err, &f) && f.Kind() == fault.InvalidCoordinates
}
