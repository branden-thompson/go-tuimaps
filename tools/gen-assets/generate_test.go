package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/mvt"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// fixtureSource stands for the planet archive: it answers every tile of zoom
// 0 to 3 with one of the fixture's real tiles, gzipped as an archive stores
// them.
func fixtureSource(t *testing.T) Source {
	t.Helper()
	var stored [][]byte
	for _, rel := range realTiles {
		var buf bytes.Buffer
		w := gzip.NewWriter(&buf)
		w.Write(fixtureTile(t, rel))
		w.Close()
		stored = append(stored, buf.Bytes())
	}
	return func(_ context.Context, id scene.TileID) ([]byte, bool, error) {
		return stored[(int(id.Z)+int(id.X)+int(id.Y))%len(stored)], true, nil
	}
}

var testPin = Pin{Archive: "planet/test", Length: 12345, EntityTag: `"ab-1"`}

// generated is one run of the generator, shared by the tests that only read
// it; a test that damages the output works on a copy.
var generated struct {
	once sync.Once
	dir  string
	err  error
}

func generate(t *testing.T) string {
	t.Helper()
	generated.once.Do(func() {
		generated.dir, generated.err = os.MkdirTemp("", "gen-assets-test")
		if generated.err == nil {
			generated.err = Generate(context.Background(), fixtureSource(t), testPin, generated.dir)
		}
	})
	if generated.err != nil {
		t.Fatal(generated.err)
	}
	copied := t.TempDir()
	if err := os.CopyFS(copied, os.DirFS(generated.dir)); err != nil {
		t.Fatal(err)
	}
	return copied
}

func gunzip(t *testing.T, path string) []byte {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(zr)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// TestGeneratorStripsTranslations is plan task 04.13.
func TestGeneratorStripsTranslations(t *testing.T) {
	dir := generate(t)
	files, _ := filepath.Glob(filepath.Join(dir, "tiles", "*.pbf.gz"))
	if len(files) != 85 {
		t.Fatalf("%d tiles written, want the 85 of zoom 0 to 3", len(files))
	}
	named := 0
	for _, f := range files {
		data := gunzip(t, f)
		for _, other := range []string{"name:", "name_", "wikidata", "iso_a2"} {
			if bytes.Contains(data, []byte(other)) {
				t.Errorf("%s still carries %q", filepath.Base(f), other)
			}
		}
		tile, err := mvt.Decode(data, drawn(), mvt.DefaultLimits())
		if err != nil {
			t.Fatalf("%s: the library's decoder refused it: %v", filepath.Base(f), err)
		}
		for _, l := range tile.Layers {
			for _, ft := range l.Features {
				if ft.Name.String() != "" {
					named++
				}
			}
		}
	}
	if named == 0 {
		t.Error("no feature in any output tile has a name: English was stripped too")
	}
}

// TestGeneratorWritesHashList is plan task 04.9.
func TestGeneratorWritesHashList(t *testing.T) {
	dir := generate(t)
	list, err := os.ReadFile(filepath.Join(dir, "HASHES"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(list)), "\n")
	sources, outputs := 0, 0
	for _, l := range lines {
		switch {
		case strings.Contains(l, "  source/"):
			sources++
		case strings.Contains(l, "  tiles/"):
			outputs++
		}
	}
	if sources != 85 || outputs != 85 {
		t.Errorf("%d source tiles and %d output tiles listed; want 85 of each", sources, outputs)
	}
	if err := Verify(dir); err != nil {
		t.Errorf("a fresh output does not verify: %v", err)
	}
}

// TestPinChangesTogether is plan task 04.10: the pin, the hash list and the
// tiles change together or not at all.
func TestPinChangesTogether(t *testing.T) {
	damage := map[string]func(dir string) error{
		"the pin edited alone": func(dir string) error {
			return os.WriteFile(filepath.Join(dir, "PIN"), []byte("archive planet/other\nlength 1\nsha256 "+strings.Repeat("cd", 32)+"\n"), 0o644)
		},
		"a tile edited alone": func(dir string) error {
			return os.WriteFile(filepath.Join(dir, "tiles", "0-0-0.pbf.gz"), []byte("changed"), 0o644)
		},
		"a tile removed": func(dir string) error { return os.Remove(filepath.Join(dir, "tiles", "3-7-7.pbf.gz")) },
		"a tile added": func(dir string) error {
			return os.WriteFile(filepath.Join(dir, "tiles", "4-0-0.pbf.gz"), []byte("extra"), 0o644)
		},
		"the hash list removed": func(dir string) error { return os.Remove(filepath.Join(dir, "HASHES")) },
	}
	for name, change := range damage {
		dir := generate(t)
		if err := change(dir); err != nil {
			t.Fatal(err)
		}
		if err := Verify(dir); err == nil {
			t.Errorf("%s: Verify passed", name)
		}
	}
}

// TestGeneratorDeterministic is plan task 04.14.
func TestGeneratorDeterministic(t *testing.T) {
	a, b := generate(t), t.TempDir()
	if err := Generate(context.Background(), fixtureSource(t), testPin, b); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"HASHES", "PIN", "tiles/0-0-0.pbf.gz", "tiles/2-1-3.pbf.gz", "tiles/3-7-7.pbf.gz"} {
		x, err := os.ReadFile(filepath.Join(a, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatal(err)
		}
		y, _ := os.ReadFile(filepath.Join(b, filepath.FromSlash(rel)))
		if !bytes.Equal(x, y) {
			t.Errorf("%s differs between two runs", rel)
		}
	}
}

func TestGenerateRefusesAnIncompleteSource(t *testing.T) {
	missing := func(_ context.Context, id scene.TileID) ([]byte, bool, error) { return nil, false, nil }
	if err := Generate(context.Background(), missing, testPin, t.TempDir()); err == nil {
		t.Error("a source that lacks a tile of zoom 0 to 3 must fail the run, not leave a hole")
	}
	if err := Generate(context.Background(), nil, testPin, t.TempDir()); err == nil {
		t.Error("no source must be an error")
	}
	if err := Generate(context.Background(), fixtureSource(t), Pin{}, t.TempDir()); err == nil {
		t.Error("an empty pin must be an error")
	}
}
