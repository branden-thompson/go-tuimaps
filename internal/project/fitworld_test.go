package project

import (
	"math"
	"testing"
)

// TestFitWorld is upstream's fit_world (P-56): latitude 84 to -56, longitude
// 0, the latitude at the Mercator midpoint, and the zoom the smaller of the
// one that fits that span's height and the one that fits the world's width.
func TestFitWorld(t *testing.T) {
	_, top, _ := ToTile(LonLat{Lat: 84}, 0)
	_, bottom, _ := ToTile(LonLat{Lat: -56}, 0)
	span := (bottom - top) * TileSize
	for _, size := range [][2]int{{149, 38}, {69, 12}, {40, 40}, {300, 20}, {2, 1}} {
		v, err := WholeWorld(size[0], size[1])
		if err != nil {
			t.Fatalf("%v: %v", size, err)
		}
		w, h := float64(size[0]*DotsPerCol), float64(size[1]*DotsPerRow)
		want := math.Min(math.Log2(h/span), math.Log2(w/TileSize))
		want = math.Max(MinViewZoom, math.Min(MaxViewZoom, want))
		if math.Abs(v.Zoom-want) > 1e-12 || v.Centre.Lon != 0 || v.Cols != size[0] || v.Rows != size[1] {
			t.Errorf("%v: %+v, want zoom %v at longitude 0", size, v, want)
		}
		_, mid, _ := ToTile(v.Centre, 0)
		if math.Abs(mid-(top+bottom)/2) > 1e-9 {
			t.Errorf("%v: the centre's latitude %v is not the Mercator midpoint of 84 and -56", size, v.Centre.Lat)
		}
		if v.Validate() != nil {
			t.Errorf("%v: the view is not valid", size)
		}
	}
	// The whole span is in view: its top and bottom are inside the rectangle.
	v, _ := WholeWorld(149, 38)
	for _, lat := range []float64{83.9, -55.9} {
		_, y, err := v.ToDot(LonLat{Lat: lat})
		if err != nil || y < 0 || y > float64(38*DotsPerRow) {
			t.Errorf("latitude %v is at dot row %v of %d", lat, y, 38*DotsPerRow)
		}
	}
	for _, bad := range [][2]int{{0, 10}, {10, 0}, {-1, -1}} {
		if _, err := WholeWorld(bad[0], bad[1]); err == nil {
			t.Errorf("%v: no error", bad)
		}
	}
}
