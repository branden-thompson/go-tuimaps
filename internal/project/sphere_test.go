package project

import (
	"math"
	"testing"
)

// TestGreatCircle is plan task 01.5: three published great-circle distances,
// each within half a percent.
func TestGreatCircle(t *testing.T) {
	cases := []struct {
		name string
		a, b LonLat
		km   float64
	}{
		{"New York to London", LonLat{-74.0060, 40.7128}, LonLat{-0.1278, 51.5074}, 5570},
		{"Tokyo to Sydney", LonLat{139.6503, 35.6762}, LonLat{151.2093, -33.8688}, 7823},
		{"Paris to Cairo", LonLat{2.3522, 48.8566}, LonLat{31.2357, 30.0444}, 3210},
	}
	for _, c := range cases {
		got, err := GreatCircleKm(c.a, c.b)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(got-c.km)/c.km > 0.005 {
			t.Errorf("%s: %.1f km, published %.0f km", c.name, got, c.km)
		}
		back, _ := GreatCircleKm(c.b, c.a)
		if math.Abs(back-got) > 1e-9 {
			t.Errorf("%s: %.6f one way and %.6f the other", c.name, got, back)
		}
	}
	if d, _ := GreatCircleKm(LonLat{-85.14, 41.08}, LonLat{-85.14, 41.08}); d != 0 {
		t.Errorf("a place is %v km from itself", d)
	}
}

// TestBearingAndCompass is plan task 01.6.
func TestBearingAndCompass(t *testing.T) {
	from := LonLat{-86.34, 37.47}
	step := 0.2
	cases := []struct {
		to   LonLat
		word string
	}{
		{LonLat{from.Lon, from.Lat + step}, "north"},
		{LonLat{from.Lon + step, from.Lat + step*0.79}, "north-east"},
		{LonLat{from.Lon + step, from.Lat}, "east"},
		{LonLat{from.Lon + step, from.Lat - step*0.79}, "south-east"},
		{LonLat{from.Lon, from.Lat - step}, "south"},
		{LonLat{from.Lon - step, from.Lat - step*0.79}, "south-west"},
		{LonLat{from.Lon - step, from.Lat}, "west"},
		{LonLat{from.Lon - step, from.Lat + step*0.79}, "north-west"},
	}
	for _, c := range cases {
		deg, err := Bearing(from, c.to)
		if err != nil {
			t.Fatal(err)
		}
		word, err := Compass(deg)
		if err != nil || word.String() != c.word {
			t.Errorf("towards %v: bearing %.1f is %q, %v; want %q", c.to, deg, word, err, c.word)
		}
	}
	// Scenario 2's answer key: 342 degrees is "north".
	for deg, want := range map[float64]string{342: "north", 0: "north", 359.9: "north", 22.4: "north", 22.5: "north-east", 337.4: "north-west", 337.5: "north", 180: "south", 360: "north", -18: "north"} {
		if word, err := Compass(deg); err != nil || word.String() != want {
			t.Errorf("Compass(%v) = %q, %v; want %q", deg, word, err, want)
		}
	}
	if _, err := Compass(math.NaN()); err == nil {
		t.Error("Compass(NaN) must be an error")
	}
}

// TestCellSpan is plan task 01.7, against specimen 17's stated scale.
func TestCellSpan(t *testing.T) {
	cases := []struct {
		v        View
		col, row float64
	}{
		{View{Centre: LonLat{-86.34, 37.47}, Zoom: 8.3, Cols: 149, Rows: 38}, 0.8, 1.6},
		{View{Centre: LonLat{-86.34, 37.47}, Zoom: 7.2, Cols: 69, Rows: 12}, 1.7, 3.4},
	}
	for _, c := range cases {
		col, row, err := c.v.CellSpanKm()
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(col-c.col) > 0.05 || math.Abs(row-c.row) > 0.05 {
			t.Errorf("zoom %v: a column is %.2f km and a row %.2f km; the specimen says %.1f and %.1f", c.v.Zoom, col, row, c.col, c.row)
		}
		if ratio := row / col; math.Abs(ratio-2) > 0.01 {
			t.Errorf("a row covers %.3f times a column's ground, want 2", ratio)
		}
	}
}

// TestLongitudeIsCircular is plan task 01.8: the seam at 180 degrees is not
// an edge.
func TestLongitudeIsCircular(t *testing.T) {
	a, b := LonLat{179.5, 0}, LonLat{-179.5, 0}
	d, err := GreatCircleKm(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(d-111.2) > 0.5 {
		t.Errorf("one degree across the seam is %.1f km; the short way is about 111", d)
	}
	if deg, _ := Bearing(a, b); math.Abs(deg-90) > 0.01 {
		t.Errorf("the bearing across the seam is %.2f, want 90: east, the short way", deg)
	}
	if deg, _ := Bearing(b, a); math.Abs(deg-270) > 0.01 {
		t.Errorf("the bearing back across the seam is %.2f, want 270", deg)
	}
}

func TestSphereRefusesNonFinite(t *testing.T) {
	bad := LonLat{math.NaN(), 0}
	ok := LonLat{0, 0}
	if _, err := GreatCircleKm(bad, ok); !isInvalidCoordinates(err) {
		t.Errorf("GreatCircleKm: %v", err)
	}
	if _, err := GreatCircleKm(ok, LonLat{0, 95}); !isInvalidCoordinates(err) {
		t.Errorf("GreatCircleKm with a latitude beyond the pole: %v", err)
	}
	if _, err := Bearing(ok, bad); !isInvalidCoordinates(err) {
		t.Errorf("Bearing: %v", err)
	}
}
