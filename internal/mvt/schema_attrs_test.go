package mvt

import (
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

// TestKeepsWhatTheRolesNeed: the attributes OpenMapTiles really carries. A
// boundary has no class: it has an administrative level, and a flag for a
// line drawn across the sea. A place's importance is its rank. Without them
// a country border cannot be told from a region's, a maritime line cannot
// be left out, and the largest places cannot be labelled first.
func TestKeepsWhatTheRolesNeed(t *testing.T) {
	line := []uint32{moveTo(1), zz(10), zz(10), lineTo(1), zz(20), zz(0)}
	point := []uint32{moveTo(1), zz(5), zz(5)}
	boundary := testLayer("boundary", 4096,
		[]string{"admin_level", "maritime", "disputed"},
		[][]byte{intValue(2), intValue(4), intValue(0), intValue(1)},
		testFeature(2, []uint32{0, 0, 1, 2, 2, 2}, line), // level 2, on land
		testFeature(2, []uint32{0, 1, 1, 3}, line),       // level 4, maritime
		testFeature(2, []uint32{1, 2}, line))             // no level at all
	place := testLayer("place", 4096,
		[]string{"class", "name", "rank"},
		[][]byte{stringValue("city"), stringValue("Chicago"), intValue(3)},
		testFeature(1, []uint32{0, 0, 1, 1, 2, 2}, point))
	tile, err := Decode(testTile(boundary, place), Want{Layers: []string{"boundary", "place"}, Language: "en"}, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	b := tile.Layers[0].Features
	if len(b) != 3 {
		t.Fatalf("%d boundary features", len(b))
	}
	if b[0].AdminLevel != 2 || b[0].Maritime {
		t.Errorf("a country border on land: %+v", b[0])
	}
	if b[1].AdminLevel != 4 || !b[1].Maritime {
		t.Errorf("a region border at sea: %+v", b[1])
	}
	if b[2].AdminLevel != 0 || b[2].Maritime {
		t.Errorf("a boundary that says nothing: %+v", b[2])
	}
	if p := tile.Layers[1].Features[0]; p.Rank != 3 || p.Class != "city" || p.Name != "Chicago" {
		t.Errorf("a place: %+v; its rank is its importance", p)
	}

	// Upstream's keys still come first where a tile has them (P-35).
	both := testLayer("place", 4096, []string{"rank", "localrank"}, [][]byte{intValue(3), intValue(7)},
		testFeature(1, []uint32{0, 0, 1, 1}, point))
	tile, err = Decode(testTile(both), Want{Layers: []string{"place"}, Language: "en"}, DefaultLimits())
	if err != nil || tile.Layers[0].Features[0].Rank != 7 {
		t.Errorf("%v, %v; the local rank is read before the rank", tile, err)
	}
	// A level that is no level is kept as none.
	silly := testLayer("boundary", 4096, []string{"admin_level"}, [][]byte{intValue(4000), stringValue("two")},
		testFeature(2, []uint32{0, 0}, line), testFeature(2, []uint32{0, 1}, line))
	tile, err = Decode(testTile(silly), Want{Layers: []string{"boundary"}, Language: "en"}, DefaultLimits())
	if err != nil || tile.Layers[0].Features[0].AdminLevel != 0 || tile.Layers[0].Features[1].AdminLevel != 0 {
		t.Errorf("%+v, %v", tile, err)
	}
}

// TestRealBoundariesAndPlaces holds the same against a real tile.
func TestRealBoundariesAndPlaces(t *testing.T) {
	root, err := testkit.FixtureRoot()
	if err != nil {
		t.Fatal(err)
	}
	body, err := testkit.LoadFixture(root, "tiles-midwest-z5/5-8-12.pbf")
	if err != nil {
		t.Fatal(err)
	}
	tile, err := Decode(body, Want{Layers: []string{"boundary", "place"}, Language: "en"}, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	levels, maritime, ranked := map[uint8]int{}, 0, 0
	for _, l := range tile.Layers {
		for _, f := range l.Features {
			if l.Name == "boundary" {
				levels[f.AdminLevel]++
				if f.Maritime {
					maritime++
				}
			}
			if l.Name == "place" && f.Kind == scene.GeomPoint && f.Rank > 0 {
				ranked++
			}
		}
	}
	if levels[2] == 0 || levels[4] == 0 || levels[0] != 0 || maritime == 0 {
		t.Errorf("boundary levels %v, %d maritime; the tile has country and region borders, some at sea", levels, maritime)
	}
	if ranked < 190 {
		t.Errorf("%d places have a rank; all 196 do in the tile", ranked)
	}
}
