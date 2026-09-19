package mvt

import (
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

// drawn is the set of layers the map draws in the schema the fixture's
// tiles use.
var drawn = Want{Layers: []string{"water", "waterway", "landcover", "park", "boundary", "transportation", "aeroway", "place", "water_name", "aerodrome_label"}, Language: "en"}

// realTiles loads every tile of the pinned fixture.
func realTiles(t testing.TB) map[string][]byte {
	t.Helper()
	root, err := testkit.FixtureRoot()
	if err != nil {
		t.Fatal(err)
	}
	tiles := map[string][]byte{}
	for _, rel := range []string{
		"tiles-gulf-z6/6-16-26.pbf", "tiles-gulf-z6/6-16-27.pbf", "tiles-gulf-z6/6-17-26.pbf", "tiles-gulf-z6/6-17-27.pbf",
		"tiles-midwest-z5/5-7-11.pbf", "tiles-midwest-z5/5-7-12.pbf", "tiles-midwest-z5/5-8-11.pbf", "tiles-midwest-z5/5-8-12.pbf",
		"tiles-urban-z14/14-4824-6157.pbf", "tiles-urban-z14/14-14552-6451.pbf", "tiles-urban-z14/14-8186-5448.pbf", "tiles-urban-z14/14-8299-5636.pbf",
	} {
		data, err := testkit.LoadFixture(root, rel)
		if err != nil {
			t.Fatal(err)
		}
		tiles[rel] = data
	}
	return tiles
}

// TestCompactNeverLargerThanSource is plan task 03.12: a one-sided bound
// from NFR-10, for every tile of the fixture and the urban set.
func TestCompactNeverLargerThanSource(t *testing.T) {
	for rel, data := range realTiles(t) {
		tile, err := Decode(data, drawn, DefaultLimits())
		if err != nil {
			t.Errorf("%s: %v", rel, err)
			continue
		}
		kept, features, coords := tile.Bytes(), 0, 0
		for _, l := range tile.Layers {
			features += len(l.Features)
			coords += len(l.Coords)
			if len(l.Coords) != cap(l.Coords) || len(l.Parts) != cap(l.Parts) {
				t.Errorf("%s, layer %s: %d of %d coordinates and %d of %d parts used; each slab is allocated once, at its exact size", rel, l.Name, len(l.Coords), cap(l.Coords), len(l.Parts), cap(l.Parts))
			}
		}
		if kept >= len(data) || kept > DefaultLimits().RetainedBytes {
			t.Errorf("%s: %d bytes of tile became %d bytes kept", rel, len(data), kept)
		}
		if len(tile.Layers) == 0 || features == 0 || coords == 0 {
			t.Errorf("%s: nothing was kept: %d layers, %d features", rel, len(tile.Layers), features)
		}
		t.Logf("%-36s %8d bytes -> %7d kept, %2d layers, %5d features, %6d coordinates", rel, len(data), kept, len(tile.Layers), features, coords/2)
	}
}

// TestRealTilesKeepOneLanguage: in a real tile a place has names in dozens
// of languages; the kept form has one.
func TestRealTilesKeepOneLanguage(t *testing.T) {
	data := realTiles(t)["tiles-midwest-z5/5-8-11.pbf"]
	for _, lang := range []string{"en", "ja"} {
		w := drawn
		w.Language = lang
		tile, err := Decode(data, w, DefaultLimits())
		if err != nil {
			t.Fatal(err)
		}
		names := 0
		for _, l := range tile.Layers {
			if l.Name != "place" {
				continue
			}
			for _, f := range l.Features {
				if f.Name != "" {
					names++
				}
				if lang == "en" && strings.ContainsAny(f.Name, "\u30B7\u30AB\u30B4") {
					t.Errorf("an English decode kept the Japanese name %q", f.Name)
				}
			}
		}
		if names == 0 {
			t.Errorf("language %s: no place in the tile has a name", lang)
		}
	}
}

// TestNeverPanics is plan task 03.13: real tiles cut short. Every byte of
// the first 2 KiB, where the headers are, and every 509th after.
func TestNeverPanics(t *testing.T) {
	for rel, data := range realTiles(t) {
		for cut := 0; cut < len(data); cut++ {
			if cut > 2048 && cut%509 != 0 {
				continue
			}
			if _, err := Decode(data[:cut], drawn, DefaultLimits()); err == nil && cut < len(data)/2 {
				// A tile cut at a field boundary is a shorter, valid tile; that
				// is fine. What must never happen is a panic, which fails the test.
				continue
			}
		}
		_ = rel
	}
}

func TestNeverPanicsOnDamage(t *testing.T) {
	data := append([]byte(nil), realTiles(t)["tiles-gulf-z6/6-16-27.pbf"]...)
	for i := 0; i < len(data); i += 251 {
		old := data[i]
		for _, b := range []byte{0x00, 0x7f, 0x80, 0xff} {
			data[i] = b
			_, _ = Decode(data, drawn, DefaultLimits())
		}
		data[i] = old
	}
}

// FuzzDecode is plan task 03.14, seeded with real tiles.
func FuzzDecode(f *testing.F) {
	for _, data := range realTiles(f) {
		if len(data) < 200<<10 {
			f.Add(data)
		}
	}
	f.Add(testTile(testLayer("water", 4096, []string{"class"}, [][]byte{stringValue("lake")}, testFeature(3, []uint32{0, 0}, []uint32{moveTo(1), zz(0), zz(0), lineTo(2), zz(5), zz(0), zz(0), zz(5), closePath()}))))
	lim := DefaultLimits()
	f.Fuzz(func(t *testing.T, data []byte) {
		tile, err := Decode(data, drawn, lim)
		if err != nil {
			return
		}
		if tile.Bytes() > lim.RetainedBytes {
			t.Fatalf("an accepted tile keeps %d bytes", tile.Bytes())
		}
		for _, l := range tile.Layers {
			end := uint32(0)
			for _, p := range l.Parts {
				if p < end || p%2 != 0 || int(p) > len(l.Coords) {
					t.Fatalf("part end %d after %d with %d coordinates", p, end, len(l.Coords))
				}
				end = p
			}
			if l.Extent == 0 || int(l.Extent) > lim.MaxExtent {
				t.Fatalf("extent %d", l.Extent)
			}
			for _, ft := range l.Features {
				if ft.FirstPart >= ft.EndPart || int(ft.EndPart) > len(l.Parts) {
					t.Fatalf("feature parts %d..%d of %d", ft.FirstPart, ft.EndPart, len(l.Parts))
				}
			}
		}
	})
}

// BenchmarkDecode is plan task 03.16: the baseline for NFR-4.
func BenchmarkDecode(b *testing.B) {
	data := realTiles(b)["tiles-midwest-z5/5-8-12.pbf"]
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	for b.Loop() {
		if _, err := Decode(data, drawn, DefaultLimits()); err != nil {
			b.Fatal(err)
		}
	}
}
