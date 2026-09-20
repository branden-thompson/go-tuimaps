package project

import (
	"math"
	"testing"
)

// destination is the point km away from p on the given bearing: a test's own
// way to place scenario 6's hazards from its answer key.
func destination(p LonLat, km, bearing float64) LonLat {
	const rad = math.Pi / 180
	d, b, lat1 := km/EarthRadiusKm, bearing*rad, p.Lat*rad
	lat2 := math.Asin(math.Sin(lat1)*math.Cos(d) + math.Cos(lat1)*math.Sin(d)*math.Cos(b))
	lon2 := p.Lon*rad + math.Atan2(math.Sin(b)*math.Sin(d)*math.Cos(lat1), math.Cos(d)-math.Sin(lat1)*math.Sin(lat2))
	return LonLat{Lon: lon2 / rad, Lat: lat2 / rad}
}

// scenario6 is M1 scenario 6 by its answer key: the place, an earthquake 6.0
// km away at 343 degrees, a fire 43.3 km away at 115 degrees.
func scenario6() []LonLat {
	place := LonLat{Lon: -85.14, Lat: 41.08}
	return []LonLat{place, destination(place, 6.0, 343), destination(place, 43.3, 115)}
}

// inside reports whether every point lies within the view's rectangle less
// the margin, and returns the first that does not.
func inside(t *testing.T, v View, points []LonLat, margin int) (bool, LonLat) {
	t.Helper()
	const slack = 1e-6
	for _, p := range points {
		x, y, err := v.ToDot(p)
		if err != nil {
			t.Fatal(err)
		}
		if x < float64(margin*DotsPerCol)-slack || x > float64((v.Cols-margin)*DotsPerCol)+slack ||
			y < float64(margin*DotsPerRow)-slack || y > float64((v.Rows-margin)*DotsPerRow)+slack {
			return false, p
		}
	}
	return true, LonLat{}
}

var small = View{Centre: LonLat{Lon: 10, Lat: 10}, Zoom: 3, Cols: 69, Rows: 12}

// TestFitToContainsAll is plan task 01.9.
func TestFitToContainsAll(t *testing.T) {
	for _, margin := range []int{0, 1, 2} {
		v, err := Frame(small, scenario6(), nil, margin)
		if err != nil {
			t.Fatal(err)
		}
		if ok, p := inside(t, v, scenario6(), margin); !ok {
			t.Errorf("margin %d: %v is outside the fitted view %+v", margin, p, v)
		}
		if v.Cols != small.Cols || v.Rows != small.Rows {
			t.Errorf("fit-to changed the view's size: %+v", v)
		}
	}
}

// TestFitToLargestZoom is plan task 01.10: the zoom is the largest that
// fits, solved directly; 1/256 of a level closer and a named point is out.
func TestFitToLargestZoom(t *testing.T) {
	for _, base := range []View{small, {Centre: LonLat{}, Zoom: 5, Cols: 149, Rows: 38}} {
		v, err := Frame(base, scenario6(), nil, 1)
		if err != nil {
			t.Fatal(err)
		}
		closer := v
		closer.Zoom += 1.0 / 256
		if ok, _ := inside(t, closer, scenario6(), 1); ok {
			t.Errorf("%dx%d: zoom %v fits, and so does 1/256 of a level closer; the zoom is not the largest", base.Cols, base.Rows, v.Zoom)
		}
	}
}

// TestFitToEdgeCases is plan task 01.12.
func TestFitToEdgeCases(t *testing.T) {
	one := []LonLat{{Lon: -85.14, Lat: 41.08}}
	v, err := Frame(small, one, nil, 1)
	if err != nil || v.Zoom != small.Zoom || math.Abs(v.Centre.Lon+85.14) > 1e-9 || math.Abs(v.Centre.Lat-41.08) > 1e-9 {
		t.Errorf("a single point must be centred at the zoom the view already has: %+v, %v", v, err)
	}

	same, err := Frame(small, nil, nil, 1)
	if err != nil || same != small {
		t.Errorf("with nothing named the view must not change: %+v, %v", same, err)
	}

	// Across the seam: the short way round is two degrees wide, not 358.
	seam := []LonLat{{Lon: 179, Lat: 50}, {Lon: -179, Lat: 52}}
	v, err = Frame(small, seam, nil, 1)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(math.Abs(v.Centre.Lon)-180) > 1e-6 {
		t.Errorf("two points either side of the seam are centred at longitude %v; the short way round centres them at 180", v.Centre.Lon)
	}
	if v.Zoom < 4 {
		t.Errorf("zoom %v: the fit took the long way round the world", v.Zoom)
	}

	// A box recorded at hand-in is enough: no shape, no Work.
	box := []Box{{West: -87, South: 40, East: -84, North: 42}}
	v, err = Frame(small, nil, box, 1)
	if err != nil {
		t.Fatal(err)
	}
	corners := []LonLat{{Lon: -87, Lat: 40}, {Lon: -84, Lat: 42}, {Lon: -87, Lat: 42}, {Lon: -84, Lat: 40}}
	if ok, p := inside(t, v, corners, 1); !ok {
		t.Errorf("corner %v of the box is outside the fitted view", p)
	}
	// A box wider than half the world is not mistaken for its complement.
	wide := []Box{{West: -100, South: -10, East: 100, North: 10}}
	v, err = Frame(View{Centre: LonLat{}, Zoom: 3, Cols: 149, Rows: 38}, nil, wide, 0)
	if err != nil || math.Abs(v.Centre.Lon) > 1e-6 {
		t.Errorf("a box from 100 west to 100 east is centred at %v, %v; want longitude 0", v.Centre.Lon, err)
	}
	// A box that itself crosses the seam.
	crossing := []Box{{West: 170, South: 50, East: -170, North: 55}}
	v, err = Frame(small, nil, crossing, 0)
	if err != nil || math.Abs(math.Abs(v.Centre.Lon)-180) > 1e-6 {
		t.Errorf("a box from 170 east to 170 west is centred at %v, %v; want 180", v.Centre.Lon, err)
	}
}

func TestFitToClampsAndRefuses(t *testing.T) {
	near := []LonLat{{Lon: 2.3500, Lat: 48.8600}, {Lon: 2.3501, Lat: 48.8601}}
	v, err := Frame(small, near, nil, 0)
	if err != nil || v.Zoom != MaxViewZoom {
		t.Errorf("two points a few metres apart: zoom %v, %v; want the closest zoom, %d", v.Zoom, err, MaxViewZoom)
	}
	if _, err := Frame(small, scenario6(), nil, 6); !isInvalidCoordinates(err) {
		t.Errorf("a margin that leaves no room in a 12-row view: %v", err)
	}
	if _, err := Frame(small, scenario6(), nil, -1); !isInvalidCoordinates(err) {
		t.Errorf("a negative margin: %v", err)
	}
	if _, err := Frame(small, []LonLat{{Lon: math.NaN(), Lat: 0}}, nil, 0); !isInvalidCoordinates(err) {
		t.Errorf("a point that is not a number: %v", err)
	}
	if _, err := Frame(small, nil, []Box{{West: 0, South: 10, East: 5, North: 5}}, 0); !isInvalidCoordinates(err) {
		t.Errorf("a box whose south is north of its north: %v", err)
	}
	if _, err := Frame(View{}, scenario6(), nil, 0); !isInvalidCoordinates(err) {
		t.Errorf("a view with no size: %v", err)
	}
}
