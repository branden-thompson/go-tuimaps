package mvt

import (
	"reflect"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// points returns a feature's parts as slices of x, y pairs.
func parts(t *scene.Tile, f scene.Feature) [][]int16 {
	var out [][]int16
	for p := f.FirstPart; p < f.EndPart; p++ {
		part, err := t.Layers[0].Part(int(p))
		if err != nil {
			panic(err)
		}
		out = append(out, part)
	}
	return out
}

func decodeOne(t *testing.T, layer []byte) *scene.Tile {
	t.Helper()
	tile, err := Decode(testTile(layer), want("l"), DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	return tile
}

// TestGeometryCommands is plan task 03.8, and TestParityP38_MVTDecode with it.
func TestGeometryCommands(t *testing.T) {
	line := testFeature(2, nil, []uint32{moveTo(1), zz(2), zz(2), lineTo(2), zz(0), zz(8), zz(8), zz(0)})
	twoLines := testFeature(2, nil, []uint32{moveTo(1), zz(2), zz(2), lineTo(1), zz(5), zz(5), moveTo(1), zz(-6), zz(-6), lineTo(1), zz(1), zz(0)})
	somePoints := testFeature(1, nil, []uint32{moveTo(2), zz(5), zz(7), zz(-2), zz(-3)})
	tile := decodeOne(t, testLayer("l", 4096, nil, nil, line, twoLines, somePoints))
	fs := tile.Layers[0].Features
	if len(fs) != 3 {
		t.Fatalf("%d features, want 3", len(fs))
	}
	if got := parts(tile, fs[0]); fs[0].Kind != scene.GeomLine || !reflect.DeepEqual(got, [][]int16{{2, 2, 2, 10, 10, 10}}) {
		t.Errorf("a line: %v %v", fs[0].Kind, got)
	}
	if got := parts(tile, fs[1]); !reflect.DeepEqual(got, [][]int16{{2, 2, 7, 7}, {1, 1, 2, 1}}) {
		t.Errorf("two lines in one feature; the cursor carries over: %v", got)
	}
	if got := parts(tile, fs[2]); fs[2].Kind != scene.GeomPoint || !reflect.DeepEqual(got, [][]int16{{5, 7}, {3, 4}}) {
		t.Errorf("two points: %v %v", fs[2].Kind, got)
	}
}

func TestParityP38_MVTDecode(t *testing.T) {
	// ClosePath re-pushes the ring's first point, as upstream does.
	square := testFeature(3, nil, []uint32{moveTo(1), zz(0), zz(0), lineTo(3), zz(10), zz(0), zz(0), zz(10), zz(-10), zz(0), closePath()})
	tile := decodeOne(t, testLayer("l", 0, nil, nil, square))
	got := parts(tile, tile.Layers[0].Features[0])
	if !reflect.DeepEqual(got, [][]int16{{0, 0, 10, 0, 10, 10, 0, 10, 0, 0}}) {
		t.Errorf("a closed square: %v", got)
	}
	if tile.Layers[0].Extent != 4096 {
		t.Errorf("default extent %d", tile.Layers[0].Extent)
	}
}

// TestParityP39_RingGrouping: a ring of area >= 0 starts a new polygon, a
// negative one is a hole in the polygon before it, and each polygon becomes
// a feature of its own.
func TestParityP39_RingGrouping(t *testing.T) {
	outer := []uint32{moveTo(1), zz(0), zz(0), lineTo(3), zz(20), zz(0), zz(0), zz(20), zz(-20), zz(0), closePath()}
	hole := []uint32{moveTo(1), zz(5), zz(5), lineTo(3), zz(0), zz(5), zz(5), zz(0), zz(0), zz(-5), closePath()}
	second := []uint32{moveTo(1), zz(25), zz(-5), lineTo(3), zz(5), zz(0), zz(0), zz(5), zz(-5), zz(0), closePath()}
	geometry := append(append(append([]uint32{}, outer...), hole...), second...)
	keys, values := []string{"class"}, [][]byte{stringValue("lake")}
	tile := decodeOne(t, testLayer("l", 4096, keys, values, testFeature(3, []uint32{0, 0}, geometry)))
	fs := tile.Layers[0].Features
	if len(fs) != 2 {
		t.Fatalf("%d features; a polygon with a hole and a second polygon are two", len(fs))
	}
	if n := fs[0].EndPart - fs[0].FirstPart; n != 2 {
		t.Errorf("the first polygon has %d rings, want its outline and its hole", n)
	}
	if n := fs[1].EndPart - fs[1].FirstPart; n != 1 {
		t.Errorf("the second polygon has %d rings, want 1", n)
	}
	if fs[0].Class != "lake" || fs[1].Class != "lake" {
		t.Errorf("each polygon carries the feature's attributes: %q %q", fs[0].Class, fs[1].Class)
	}
	// An orphan hole, with no outline before it, is its own polygon, as upstream.
	orphan := decodeOne(t, testLayer("l", 4096, nil, nil, testFeature(3, nil, hole)))
	if len(orphan.Layers[0].Features) != 1 {
		t.Errorf("an orphan hole gave %d features", len(orphan.Layers[0].Features))
	}
}

// TestKeysAndValues is plan task 03.6, with TestParityP36_LabelLanguage and
// TestParityP26_SortKey.
func TestKeysAndValues(t *testing.T) {
	keys := []string{"class", "name", "name:fr", "name:en", "name_de", "house_num", "localrank", "scalerank", "wikidata", "name:ja"}
	values := [][]byte{stringValue("city"), stringValue("Wien"), stringValue("Vienne"), stringValue("Vienna"), stringValue("Wien (de)"), stringValue("12"), intValue(7), intValue(3), stringValue("Q1741"), stringValue("\xe3\x82\xa6")}
	all := []uint32{0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9}
	pt := []uint32{moveTo(1), zz(1), zz(1)}
	layer := testLayer("l", 4096, keys, values,
		testFeature(1, all, pt),
		testFeature(1, []uint32{1, 1, 7, 7}, pt), // no English name: the local name; scalerank alone
		testFeature(1, []uint32{5, 5}, pt),       // a house number and nothing else
		testFeature(1, nil, pt))
	cases := []struct {
		lang  string
		names []string
	}{
		{"en", []string{"Vienna", "Wien", "12", ""}},
		{"fr", []string{"Vienne", "Wien", "12", ""}},
		{"de", []string{"Wien (de)", "Wien", "12", ""}}, // the underscore form is found too
		{"it", []string{"Wien", "Wien", "12", ""}},      // no Italian name: the local name, not English (D-82)
	}
	for _, c := range cases {
		tile, err := Decode(testTile(layer), Want{Layers: []string{"l"}, Language: c.lang}, DefaultLimits())
		if err != nil {
			t.Fatal(err)
		}
		for i, f := range tile.Layers[0].Features {
			if f.Name != c.names[i] {
				t.Errorf("language %s, feature %d: name %q, want %q", c.lang, i, f.Name, c.names[i])
			}
		}
		fs := tile.Layers[0].Features
		if fs[0].Class != "city" || fs[0].Rank != 7 || fs[1].Rank != 3 || fs[2].Rank != 0 {
			t.Errorf("class %q, ranks %d %d %d; want city, 7 (localrank first), 3 (scalerank), 0", fs[0].Class, fs[0].Rank, fs[1].Rank, fs[2].Rank)
		}
	}
}

func TestParityP36_LabelLanguage(t *testing.T) { TestKeysAndValues(t) }

func TestParityP26_SortKey(t *testing.T) {
	keys := []string{"localrank", "scalerank"}
	values := [][]byte{stringValue("9"), intValue(4)}
	pt := []uint32{moveTo(1), zz(1), zz(1)}
	tile := decodeOne(t, testLayer("l", 4096, keys, values, testFeature(1, []uint32{0, 0, 1, 1}, pt), testFeature(1, []uint32{0, 0}, pt)))
	fs := tile.Layers[0].Features
	// Integers only: a rank written as text does not count, and - as
	// upstream - it is localrank that was found, so scalerank is not read.
	if fs[0].Rank != 0 || fs[1].Rank != 0 {
		t.Errorf("ranks %d %d; a localrank that is not an integer gives 0, as upstream", fs[0].Rank, fs[1].Rank)
	}
}

// TestFeatureLimit and TestGeometryIntegerLimit are plan task 03.7.
func TestFeatureLimit(t *testing.T) {
	pt := testFeature(1, nil, []uint32{moveTo(1), zz(1), zz(1)})
	lim := DefaultLimits()
	lim.Features = 3
	four := testLayer("l", 4096, nil, nil, pt, pt, pt, pt)
	if _, err := Decode(testTile(four), want("l"), lim); !isKind(err, fault.OverLimit) {
		t.Errorf("four features against a limit of three: %v", err)
	}
	// Features of a layer the map does not draw are not counted.
	if _, err := Decode(testTile(testLayer("other", 4096, nil, nil, pt, pt, pt, pt), testLayer("l", 4096, nil, nil, pt)), want("l"), lim); err != nil {
		t.Errorf("features of a dropped layer were counted: %v", err)
	}
}

func TestGeometryIntegerLimit(t *testing.T) {
	lim := DefaultLimits()
	lim.GeometryIntegers = 8
	long := testFeature(2, nil, []uint32{moveTo(1), zz(0), zz(0), lineTo(3), zz(1), zz(1), zz(1), zz(1), zz(1), zz(1)})
	if _, err := Decode(testTile(testLayer("l", 4096, nil, nil, long)), want("l"), lim); !isKind(err, fault.OverLimit) {
		t.Errorf("ten geometry integers against a limit of eight: %v", err)
	}
}

// TestExtentRange and TestCursorLeavesInt16IsError are plan task 03.9.
func TestExtentRange(t *testing.T) {
	for _, extent := range []uint32{8193, 65536, 1 << 31} {
		if _, err := Decode(testTile(testLayer("l", extent, nil, nil)), want("l"), DefaultLimits()); !isKind(err, fault.OverLimit) {
			t.Errorf("extent %d: %v", extent, err)
		}
	}
	zero := putUint(testLayer("l", 0, nil, nil), 5, 0)
	if _, err := Decode(testTile(zero), want("l"), DefaultLimits()); !isKind(err, fault.OverLimit) {
		t.Errorf("extent 0: %v", err)
	}
	if _, err := Decode(testTile(testLayer("l", 8192, nil, nil)), want("l"), DefaultLimits()); err != nil {
		t.Errorf("extent 8192 was refused: %v", err)
	}
}

func TestCursorLeavesInt16IsError(t *testing.T) {
	walk := testFeature(2, nil, []uint32{moveTo(1), zz(30000), zz(0), lineTo(1), zz(3000), zz(0)})
	if _, err := Decode(testTile(testLayer("l", 4096, nil, nil, walk)), want("l"), DefaultLimits()); !isKind(err, fault.UnsupportedTile) {
		t.Errorf("a cursor that walks past 32767: %v; it must be an error, never a wrap", err)
	}
	jump := testFeature(1, nil, []uint32{moveTo(1), zz(-2147483648), zz(0)})
	if _, err := Decode(testTile(testLayer("l", 4096, nil, nil, jump)), want("l"), DefaultLimits()); !isKind(err, fault.UnsupportedTile) {
		t.Errorf("a delta of the most negative 32-bit value: %v", err)
	}
	edge := testFeature(1, nil, []uint32{moveTo(1), zz(32767), zz(-32768)})
	if _, err := Decode(testTile(testLayer("l", 4096, nil, nil, edge)), want("l"), DefaultLimits()); err != nil {
		t.Errorf("the corners of the 16-bit range were refused: %v", err)
	}
}

// TestRetainedLimit is plan task 03.10.
func TestRetainedLimit(t *testing.T) {
	pt := testFeature(1, nil, []uint32{moveTo(1), zz(1), zz(1)})
	var fs [][]byte
	for range 200 {
		fs = append(fs, pt)
	}
	lim := DefaultLimits()
	lim.RetainedBytes = 1024
	if _, err := Decode(testTile(testLayer("l", 4096, nil, nil, fs...)), want("l"), lim); !isKind(err, fault.OverLimit) {
		t.Errorf("200 features against a kept form of 1 KiB: %v", err)
	}
	tile := decodeOne(t, testLayer("l", 4096, nil, nil, fs...))
	if got := tile.Bytes(); got <= 0 || got > 64<<10 {
		t.Errorf("200 point features are accounted as %d bytes", got)
	}
}

// TestProtobufPitfalls is the decoder's share of plan task 03.17.
func TestProtobufPitfalls(t *testing.T) {
	pt := []uint32{moveTo(1), zz(1), zz(1)}
	unpackedGeometry := putUint(putUint(putUint(nil, 3, 1), 4, uint64(moveTo(1))), 4, 2)
	cases := map[string][]byte{
		"an odd number of tag integers":            testLayer("l", 4096, []string{"class"}, [][]byte{stringValue("a")}, testFeature(1, []uint32{0, 0, 0}, pt)),
		"a key index out of range":                 testLayer("l", 4096, []string{"class"}, [][]byte{stringValue("a")}, testFeature(1, []uint32{4, 0}, pt)),
		"a value index out of range":               testLayer("l", 4096, []string{"class"}, [][]byte{stringValue("a")}, testFeature(1, []uint32{0, 9}, pt)),
		"a command count beyond the integers left": testLayer("l", 4096, nil, nil, testFeature(2, nil, []uint32{moveTo(1), zz(0), zz(0), lineTo(5), zz(1), zz(1)})),
		"LineTo before MoveTo":                     testLayer("l", 4096, nil, nil, testFeature(2, nil, []uint32{lineTo(1), zz(1), zz(1)})),
		"a ClosePath with a count of two":          testLayer("l", 4096, nil, nil, testFeature(3, nil, []uint32{moveTo(1), zz(0), zz(0), lineTo(2), zz(1), zz(0), zz(0), zz(1), 7 | 2<<3})),
		"an unknown command":                       testLayer("l", 4096, nil, nil, testFeature(2, nil, []uint32{3 | 1<<3, zz(0), zz(0)})),
		"a MoveTo with a count of zero":            testLayer("l", 4096, nil, nil, testFeature(1, nil, []uint32{moveTo(0)})),
		"an unknown geometry type":                 testLayer("l", 4096, nil, nil, testFeature(9, nil, pt)),
		"geometry that is not packed":              testLayer("l", 4096, nil, nil, unpackedGeometry),
		"a geometry varint cut short":              testLayer("l", 4096, nil, nil, putBytes(putUint(nil, 3, 1), 4, []byte{0x09, 0x80})),
		"a version the library does not read":      putUint(testLayer("l", 4096, nil, nil)[2:], 15, 3),
	}
	for name, layer := range cases {
		if _, err := Decode(testTile(layer), want("l"), DefaultLimits()); !isKind(err, fault.UnsupportedTile) {
			t.Errorf("%s: %v; want a clear unsupported-tile error", name, err)
		}
	}
}
