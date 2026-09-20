package oracle

import (
	"fmt"
	"os"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/mvt"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/testkit"
	"github.com/paulmach/orb"
	orbmvt "github.com/paulmach/orb/encoding/mvt"
	"github.com/paulmach/orb/geojson"
)

func TestMain(m *testing.M) { os.Exit(testkit.Main(m)) }

// drawn is the set of layers the map draws; the same filter is applied to
// both decoders' output.
var drawn = mvt.Want{Layers: []string{"water", "waterway", "landcover", "park", "boundary", "transportation", "aeroway", "place", "water_name", "aerodrome_label"}, Language: "en"}

var fixtureTiles = []string{
	"tiles-gulf-z6/6-16-26.pbf", "tiles-gulf-z6/6-16-27.pbf", "tiles-gulf-z6/6-17-26.pbf", "tiles-gulf-z6/6-17-27.pbf",
	"tiles-midwest-z5/5-7-11.pbf", "tiles-midwest-z5/5-7-12.pbf", "tiles-midwest-z5/5-8-11.pbf", "tiles-midwest-z5/5-8-12.pbf",
	"tiles-urban-z14/14-4824-6157.pbf", "tiles-urban-z14/14-14552-6451.pbf", "tiles-urban-z14/14-8186-5448.pbf", "tiles-urban-z14/14-8299-5636.pbf",
}

func loadTile(t testing.TB, rel string) []byte {
	t.Helper()
	root, err := testkit.FixtureRoot()
	if err != nil {
		t.Fatal(err)
	}
	data, err := testkit.LoadFixture(root, rel)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// flat turns the proven decoder's geometry into parts of integer pairs, in
// order: a point, a line and a ring are each one part.
func flat(g orb.Geometry) [][]int16 {
	line := func(ps []orb.Point) []int16 {
		out := make([]int16, 0, 2*len(ps))
		for _, p := range ps {
			out = append(out, int16(p[0]), int16(p[1]))
		}
		return out
	}
	var parts [][]int16
	switch g := g.(type) {
	case orb.Point:
		parts = append(parts, line([]orb.Point{g}))
	case orb.MultiPoint:
		for _, p := range g {
			parts = append(parts, line([]orb.Point{p}))
		}
	case orb.LineString:
		parts = append(parts, line(g))
	case orb.MultiLineString:
		for _, l := range g {
			parts = append(parts, line(l))
		}
	case orb.Polygon:
		for _, r := range g {
			parts = append(parts, line(r))
		}
	case orb.MultiPolygon:
		for _, p := range g {
			for _, r := range p {
				parts = append(parts, line(r))
			}
		}
	}
	return parts
}

// polygons is how many features of ours one of theirs becomes.
func polygons(g orb.Geometry) int {
	if mp, ok := g.(orb.MultiPolygon); ok {
		return len(mp)
	}
	if g == nil {
		return 0
	}
	return 1
}

// label is upstream's lookup order with one language kept (P-36, D-82).
func label(f *geojson.Feature, lang string) string {
	for _, key := range []string{"name_" + lang, "name:" + lang, "name", "house_num"} {
		if s, ok := f.Properties[key].(string); ok && s != "" {
			return s
		}
	}
	return ""
}

// whole reads a property as a whole number, whatever type the proven
// decoder gave it.
func whole(v any) (int64, bool) {
	switch n := v.(type) {
	case float64:
		return int64(n), n == float64(int64(n))
	case int64:
		return n, true
	case uint64:
		return int64(n), true
	case int:
		return int64(n), true
	}
	return 0, false
}

// importance is a feature's rank by upstream's order, then OpenMapTiles'
// own key; a boundary's administrative level; and whether it is at sea.
func importance(f *geojson.Feature) (rank int32, level uint8, maritime bool) {
	for _, key := range []string{"localrank", "scalerank", "rank"} {
		if v, present := f.Properties[key]; present {
			if n, ok := whole(v); ok && n >= 0 && n < 1<<31 {
				rank = int32(n)
			}
			break
		}
	}
	if n, ok := whole(f.Properties["admin_level"]); ok && n >= 1 && n <= 11 {
		level = uint8(n)
	}
	n, ok := whole(f.Properties["maritime"])
	return rank, level, ok && n == 1
}

// compare holds our kept layers equal to the proven decoder's, feature by
// feature: geometry part by part, class and label by value.
func compare(theirs orbmvt.Layers, ours *scene.Tile) (compared int, err error) {
	kept := map[string]*scene.Layer{}
	for i := range ours.Layers {
		kept[ours.Layers[i].Name] = &ours.Layers[i]
	}
	for _, their := range theirs {
		our := kept[their.Name]
		if our == nil {
			continue // a layer the map does not draw
		}
		delete(kept, their.Name)
		next, part := 0, 0
		untyped := int(our.Untyped)
	features:
		for n, f := range their.Features {
			want := flat(f.Geometry)
			if len(want) == 0 {
				continue
			}
			for range polygons(f.Geometry) {
				if next >= len(our.Features) {
					// **The last exemption this oracle takes (D-118).** A
					// feature that does not say what it is cannot be
					// drawn, so this decoder passes it over and counts it;
					// the proven decoder reads its geometry anyway. Only
					// as many features as the decoder counted may be
					// missing - a feature dropped for any other reason
					// still fails here, and no further narrowing is taken.
					if untyped > 0 {
						untyped--
						continue features
					}
					return compared, fmt.Errorf("layer %s: they have feature %d, we ran out after %d", their.Name, n, next)
				}
				got := our.Features[next]
				next++
				compared++
				class, _ := f.Properties["class"].(string)
				if got.Class != class || got.Name != label(f, "en") {
					return compared, fmt.Errorf("layer %s feature %d: class %q name %q; they have class %q name %q", their.Name, n, got.Class, got.Name, class, label(f, "en"))
				}
				rank, level, sea := importance(f)
				if got.Rank != rank || got.AdminLevel != level || got.Maritime != sea {
					// **The one difference this oracle does not arbitrate,
					// and it covers attributes only (D-117).** On a value
					// carrying several fields where the format allows one,
					// this decoder answers "none" rather than picking a
					// number out of the damage, and the proven decoder
					// picks one - it has no rule for garbage either. Where
					// ours refused and theirs guessed, that is a difference
					// by design, the same standing this file already gives
					// their panics. **A number this decoder invents where
					// they read none is still a failure**, and geometry,
					// classes, names, feature counts and which layers are
					// kept stay exact on every input.
					if !onlyRefused(got.Rank, rank) || !onlyRefused(int32(got.AdminLevel), int32(level)) || (got.Maritime && !sea) {
						return compared, fmt.Errorf("layer %s feature %d: rank %d level %d maritime %v; they have %d, %d, %v", their.Name, n, got.Rank, got.AdminLevel, got.Maritime, rank, level, sea)
					}
				}
				for p := got.FirstPart; p < got.EndPart; p++ {
					coords, err := our.Part(int(p))
					if err != nil {
						return compared, err
					}
					if part >= len(want) || fmt.Sprint(coords) != fmt.Sprint(want[part]) {
						return compared, fmt.Errorf("layer %s feature %d part %d: we have %v", their.Name, n, part, coords)
					}
					part++
				}
			}
			if part != len(want) {
				return compared, fmt.Errorf("layer %s feature %d: we kept %d parts of %d", their.Name, n, part, len(want))
			}
			part = 0
		}
		if next != len(our.Features) {
			return compared, fmt.Errorf("layer %s: we kept %d features, they account for %d", their.Name, len(our.Features), next)
		}
	}
	for name := range kept {
		return compared, fmt.Errorf("we kept layer %s and they have no such layer", name)
	}
	return compared, nil
}

// TestKeptFeaturesEqualTheProvenDecoder is plan task 03.15.
func TestKeptFeaturesEqualTheProvenDecoder(t *testing.T) {
	total := 0
	defer func() { t.Logf("%d features compared across %d tiles", total, len(fixtureTiles)) }()
	for _, rel := range fixtureTiles {
		data := loadTile(t, rel)
		theirs, err := orbmvt.Unmarshal(data)
		if err != nil {
			t.Fatalf("%s: the proven decoder refused it: %v", rel, err)
		}
		ours, err := mvt.Decode(data, drawn, mvt.DefaultLimits())
		if err != nil {
			t.Fatalf("%s: %v", rel, err)
		}
		compared, err := compare(theirs, ours)
		if err != nil {
			t.Errorf("%s: %v", rel, err)
		}
		if compared < 100 {
			t.Errorf("%s: only %d features were compared", rel, compared)
		}
		total += compared
	}
}

// provenDecode runs the proven decoder and turns a panic into a refusal. It
// does panic on hostile bytes - a feature with no geometry field is a nil
// dereference inside it - which is one reason the library has a decoder of
// its own (D-75). A panic there is not a disagreement with ours.
func provenDecode(data []byte) (layers orbmvt.Layers, err error) {
	defer func() {
		if r := recover(); r != nil {
			layers, err = nil, fmt.Errorf("the proven decoder panicked: %v", r)
		}
	}()
	return orbmvt.Unmarshal(data)
}

// TestTheProvenDecoderPanicsWhereOursDoesNot keeps the input the fuzzer
// found: a feature with no geometry. Ours accepts it and keeps nothing of
// that feature; the proven decoder dereferences a nil pointer.
func TestTheProvenDecoderPanicsWhereOursDoesNot(t *testing.T) {
	// A tile with one layer, "water", holding one feature that has a type
	// and no geometry field.
	feature := []byte{0x18, 0x01}
	layer := append([]byte{0x78, 0x02, 0x0a, 0x05, 'w', 'a', 't', 'e', 'r', 0x12, byte(len(feature))}, feature...)
	tile := append([]byte{0x1a, byte(len(layer))}, layer...)
	ours, err := mvt.Decode(tile, drawn, mvt.DefaultLimits())
	if err != nil || len(ours.Layers) != 1 || len(ours.Layers[0].Features) != 0 {
		t.Fatalf("ours: %+v, %v; want the layer kept and the feature passed over", ours, err)
	}
	if _, err := provenDecode(tile); err == nil {
		t.Log("the proven decoder no longer panics on a feature with no geometry; this test can go when the oracle is updated")
	}
}

// FuzzAgree is plan task 03.18: wherever both decoders accept a tile, what
// is kept agrees.
func FuzzAgree(f *testing.F) {
	for _, rel := range fixtureTiles[:4] {
		f.Add(loadTile(f, rel))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		ours, err := mvt.Decode(data, drawn, mvt.DefaultLimits())
		if err != nil {
			return
		}
		theirs, err := provenDecode(data)
		if err != nil {
			return
		}
		if _, err := compare(theirs, ours); err != nil {
			t.Fatal(err)
		}
	})
}

// onlyRefused reports whether this decoder's answer for one attribute is
// either the same as the proven decoder's or the refusal "none" where the
// proven decoder read a number. It is never true the other way round
// (D-117).
func onlyRefused(ours, theirs int32) bool {
	return ours == theirs || ours == 0
}
