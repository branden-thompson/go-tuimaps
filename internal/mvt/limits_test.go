package mvt

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// want is what the tests ask the decoder to keep.
func want(layers ...string) Want {
	return Want{Layers: layers, Language: "en"}
}

func gz(t *testing.T, body []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestDefaultLimitsAreTheRequirementsNumbers(t *testing.T) {
	d := DefaultLimits()
	if d.BodyBytes != 2<<20 || d.DecompressedBytes != 8<<20 || d.Layers != 64 || d.Features != 100000 ||
		d.GeometryIntegers != 2000000 || d.RetainedBytes != 4<<20 || d.MaxExtent != 8192 {
		t.Errorf("%+v does not match NFR-10 and the constants file", d)
	}
}

// TestBodyLimit is plan task 03.2.
func TestBodyLimit(t *testing.T) {
	lim := DefaultLimits()
	lim.BodyBytes = 64
	big := testTile(testLayer("water", 4096, nil, nil, testFeature(1, nil, make([]uint32, 80))))
	if _, err := Decode(big, want("water"), lim); !isKind(err, fault.OverLimit) {
		t.Errorf("a body over the limit: %v; want an over-limit error", err)
	}
	if _, err := Decode(nil, want("water"), DefaultLimits()); !isKind(err, fault.UnsupportedTile) {
		t.Errorf("an empty body: %v", err)
	}
}

// TestGzipLimit is plan task 03.3: a small body that decompresses without
// end is stopped at the limit.
func TestGzipLimit(t *testing.T) {
	lim := DefaultLimits()
	bomb := gz(t, make([]byte, lim.DecompressedBytes+1))
	if len(bomb) > lim.BodyBytes {
		t.Fatalf("the bomb is %d bytes; it must pass the body limit to test the other one", len(bomb))
	}
	if _, err := Decode(bomb, want("water"), lim); !isKind(err, fault.OverLimit) {
		t.Errorf("a gzip bomb: %v; want an over-limit error", err)
	}
}

// TestParityP37_GzipSniff: a gzipped tile is known by its first two bytes,
// and decodes to the same tile as the plain one.
func TestParityP37_GzipSniff(t *testing.T) {
	plain := testTile(testLayer("water", 4096, nil, nil), testLayer("place", 4096, nil, nil))
	a, err := Decode(plain, want("water", "place"), DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	b, err := Decode(gz(t, plain), want("water", "place"), DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Layers) != 2 || len(b.Layers) != 2 || a.Layers[1].Name != "place" || b.Layers[1].Name != "place" {
		t.Errorf("plain gave %+v, gzipped gave %+v", a.Layers, b.Layers)
	}
	if _, err := Decode([]byte{0x1f, 0x8b, 0x08, 0xff, 0xff}, want("water"), DefaultLimits()); !isKind(err, fault.UnsupportedTile) {
		t.Errorf("damaged gzip: %v", err)
	}
}

// TestLayerLimit is plan task 03.4.
func TestLayerLimit(t *testing.T) {
	var layers [][]byte
	for i := range 65 {
		layers = append(layers, testLayer(fmt.Sprintf("l%d", i), 4096, nil, nil))
	}
	if _, err := Decode(testTile(layers...), want("l1"), DefaultLimits()); !isKind(err, fault.OverLimit) {
		t.Errorf("65 layers: %v; want an over-limit error", err)
	}
	if _, err := Decode(testTile(layers[:64]...), want("l1"), DefaultLimits()); err != nil {
		t.Errorf("64 layers were refused: %v", err)
	}
}

// TestUnsupported is plan task 03.11.
func TestUnsupported(t *testing.T) {
	cases := map[string][]byte{
		"a PNG image":        {0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 0, 0},
		"a JPEG image":       {0xff, 0xd8, 0xff, 0xe0, 0, 0},
		"a WebP image":       []byte("RIFF\x00\x00\x00\x00WEBPVP8 "),
		"zstd compression":   {0x28, 0xb5, 0x2f, 0xfd, 0, 0},
		"an HTML error page": []byte("<!doctype html><html>not found</html>"),
		"a JSON error":       []byte(`{"error":"no such tile"}`),
	}
	for name, body := range cases {
		_, err := Decode(body, want("water"), DefaultLimits())
		if !isKind(err, fault.UnsupportedTile) {
			t.Errorf("%s: %v; want an unsupported-tile error", name, err)
		}
	}
}

func TestOnlyWantedLayersAreKept(t *testing.T) {
	tile := testTile(testLayer("water", 4096, nil, nil), testLayer("housenumber", 4096, nil, nil), testLayer("place", 0, nil, nil))
	got, err := Decode(tile, want("place", "water"), DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Layers) != 2 || got.Layers[0].Name != "water" || got.Layers[1].Name != "place" {
		t.Errorf("kept %+v; want water and place, in the tile's own order", got.Layers)
	}
	if got.Layers[1].Extent != 4096 {
		t.Errorf("extent %d; a layer that leaves the extent out has the default, 4096", got.Layers[1].Extent)
	}
}

// TestLimitsValidate: a host may lower a limit; it may not remove one, and
// it may not raise the extent past what a 16-bit coordinate can hold.
func TestLimitsValidate(t *testing.T) {
	if err := DefaultLimits().Validate(); err != nil {
		t.Fatalf("the defaults were refused: %v", err)
	}
	lower := DefaultLimits()
	lower.BodyBytes, lower.Features, lower.MaxExtent = 1<<16, 100, 4096
	if err := lower.Validate(); err != nil {
		t.Errorf("lowered limits were refused: %v", err)
	}
	for name, change := range map[string]func(*Limits){
		"no body limit":           func(l *Limits) { l.BodyBytes = 0 },
		"a negative size":         func(l *Limits) { l.DecompressedBytes = -1 },
		"no layer limit":          func(l *Limits) { l.Layers = 0 },
		"no feature limit":        func(l *Limits) { l.Features = 0 },
		"no geometry limit":       func(l *Limits) { l.GeometryIntegers = 0 },
		"no key limit":            func(l *Limits) { l.Keys = 0 },
		"no value limit":          func(l *Limits) { l.Values = 0 },
		"no retained limit":       func(l *Limits) { l.RetainedBytes = 0 },
		"an extent of zero":       func(l *Limits) { l.MaxExtent = 0 },
		"an extent past 16 bits":  func(l *Limits) { l.MaxExtent = 8193 },
		"a body limit raised":     func(l *Limits) { l.BodyBytes = 3 << 20 },
		"a retained limit raised": func(l *Limits) { l.RetainedBytes = 5 << 20 },
		"a geometry limit raised": func(l *Limits) { l.GeometryIntegers = 2000001 },
	} {
		lim := DefaultLimits()
		change(&lim)
		if err := lim.Validate(); !isKind(err, fault.OverLimit) {
			t.Errorf("%s: Validate gave %v", name, err)
		}
		if _, err := Decode(testTile(testLayer("l", 4096, nil, nil)), want("l"), lim); !isKind(err, fault.OverLimit) {
			t.Errorf("%s: Decode gave %v", name, err)
		}
	}
}

func TestNothingWantedDecodesNothing(t *testing.T) {
	tile, err := Decode(testTile(testLayer("l", 4096, nil, nil)), Want{}, DefaultLimits())
	if err != nil || len(tile.Layers) != 0 {
		t.Errorf("%+v, %v", tile, err)
	}
}
