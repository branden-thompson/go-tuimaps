package main

import (
	"context"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/mvt"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

func TestMain(m *testing.M) {
	code := testkit.Main(m)
	if generated.dir != "" {
		os.RemoveAll(generated.dir)
	}
	os.Exit(code)
}

func fixtureTile(t testing.TB, rel string) []byte {
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

var realTiles = []string{
	"tiles-gulf-z6/6-16-26.pbf", "tiles-gulf-z6/6-17-27.pbf", "tiles-midwest-z5/5-8-12.pbf", "tiles-urban-z14/14-8299-5636.pbf",
}

// TestEncodeRoundTrip is plan task 04.8: a tile decoded, stripped and
// re-encoded decodes to the same kept features.
func TestEncodeRoundTrip(t *testing.T) {
	for _, rel := range realTiles {
		source := fixtureTile(t, rel)
		kept, err := mvt.Decode(source, drawn(), mvt.DefaultLimits())
		if err != nil {
			t.Fatalf("%s: %v", rel, err)
		}
		encoded, err := Encode(kept)
		if err != nil {
			t.Fatalf("%s: %v", rel, err)
		}
		again, err := mvt.Decode(encoded, drawn(), mvt.DefaultLimits())
		if err != nil {
			t.Fatalf("%s: the library's decoder refused the generator's own output: %v", rel, err)
		}
		if len(again.Layers) != len(kept.Layers) {
			t.Fatalf("%s: %d layers became %d", rel, len(kept.Layers), len(again.Layers))
		}
		for i := range kept.Layers {
			a, b := kept.Layers[i], again.Layers[i]
			if a.Name != b.Name || a.Extent != b.Extent || !reflect.DeepEqual(a.Features, b.Features) ||
				!reflect.DeepEqual(a.Coords, b.Coords) || !reflect.DeepEqual(a.Parts, b.Parts) {
				t.Errorf("%s, layer %s: what is kept changed across a round trip (%d features, then %d)", rel, a.Name, len(a.Features), len(b.Features))
			}
		}
		if len(encoded) >= len(source) {
			t.Errorf("%s: %d bytes of source became %d bytes stripped", rel, len(source), len(encoded))
		}
		t.Logf("%-36s %8d bytes -> %7d stripped", rel, len(source), len(encoded))
	}
}

func TestEncodeRefusesWhatItCannotWrite(t *testing.T) {
	if _, err := Encode(nil); err == nil {
		t.Error("no tile must be an error")
	}
	point := scene.Layer{Name: "place", Extent: 4096, Coords: []int16{1, 2}, Parts: []uint32{2},
		Features: []scene.Feature{{Kind: scene.GeomPoint, FirstPart: 0, EndPart: 1}}}
	if _, err := Encode(&scene.Tile{Layers: []scene.Layer{point}}); err != nil {
		t.Fatalf("the well-formed layer the cases below damage: %v", err)
	}
	cases := map[string]func(l *scene.Layer){
		"a layer with no name":       func(l *scene.Layer) { l.Name = "" },
		"a layer with no extent":     func(l *scene.Layer) { l.Extent = 0 },
		"a feature of no known kind": func(l *scene.Layer) { l.Features[0].Kind = 0 },
		"a feature with no geometry": func(l *scene.Layer) { l.Features[0].EndPart = 0 },
		"a part the layer lacks":     func(l *scene.Layer) { l.Features[0].EndPart = 9 },
	}
	for name, damage := range cases {
		l := point
		l.Features = append([]scene.Feature(nil), point.Features...)
		damage(&l)
		if _, err := Encode(&scene.Tile{Layers: []scene.Layer{l}}); err == nil {
			t.Errorf("%s must be an error", name)
		}
	}
	if _, err := Encode(&scene.Tile{}); err == nil {
		t.Error("a tile with no layer must be an error")
	}
	if _, err := one(context.Background(), nil, scene.TileID{}, t.TempDir()); err == nil {
		t.Error("no source must be an error")
	}
	if code := run([]string{"-url", "https://example.invalid/x"}, io.Discard); code != 1 {
		t.Errorf("a missing flag: exit %d, want 1", code)
	}
}
