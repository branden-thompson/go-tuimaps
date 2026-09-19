//go:build !race

package mvt

import "testing"

// TestDropUnusedLayer is plan task 03.5: a layer the map does not draw costs
// no allocation at all, however large it is.
func TestDropUnusedLayer(t *testing.T) {
	var features [][]byte
	for range 500 {
		features = append(features, testFeature(2, []uint32{0, 0}, []uint32{moveTo(1), zz(1), zz(1), lineTo(2), zz(3), zz(3), zz(4), zz(4)}))
	}
	heavy := testTile(testLayer("building", 4096, []string{"class"}, [][]byte{stringValue("house")}, features...))
	lim := DefaultLimits()
	water := Want{Layers: []string{"water"}} // made once: the test must not count its own allocation
	allocs := testing.AllocsPerRun(50, func() {
		tile, err := Decode(heavy, water, lim)
		if err != nil || len(tile.Layers) != 0 {
			t.Fatalf("%v, %v", tile, err)
		}
	})
	if allocs > 1 { // the tile value itself
		t.Errorf("dropping a 500-feature layer allocated %v times; only the empty tile may be allocated", allocs)
	}
}
