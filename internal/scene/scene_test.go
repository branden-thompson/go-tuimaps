package scene

import (
	"context"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestSceneImportsNothingHere is plan task 00.10. Every internal package
// meets the others through scene, so scene itself may import only the
// standard library; that is what keeps the import graph free of cycles.
func TestSceneImportsNothingHere(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), f, src, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		checked++
		for _, imp := range parsed.Imports {
			path, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			first, _, _ := strings.Cut(path, "/")
			if strings.Contains(first, ".") {
				t.Errorf("%s imports %s; scene may import the standard library only", f, path)
			}
		}
	}
	if checked == 0 {
		t.Fatal("no non-test file was checked")
	}
}

func TestTileIDValidate(t *testing.T) {
	cases := []struct {
		name string
		id   TileID
		ok   bool
	}{
		{"the world tile", TileID{0, 0, 0}, true},
		{"the last tile of zoom 3", TileID{3, 7, 7}, true},
		{"the deepest zoom", TileID{MaxTileZoom, 1<<MaxTileZoom - 1, 0}, true},
		{"x one past the edge", TileID{3, 8, 0}, false},
		{"y one past the edge", TileID{3, 0, 8}, false},
		{"zoom 0 with a column", TileID{0, 1, 0}, false},
		{"deeper than any source", TileID{MaxTileZoom + 1, 0, 0}, false},
	}
	for _, c := range cases {
		if err := c.id.Validate(); (err == nil) != c.ok {
			t.Errorf("%s: Validate() = %v, want ok=%v", c.name, err, c.ok)
		}
	}
}

func TestTileIDParent(t *testing.T) {
	got, err := TileID{3, 5, 6}.Parent()
	if err != nil || got != (TileID{2, 2, 3}) {
		t.Errorf("parent of 3/5/6 = %v, %v; want 2/2/3", got, err)
	}
	if _, err := (TileID{0, 0, 0}).Parent(); err == nil {
		t.Error("the world tile has no parent, and Parent must say so")
	}
	if _, err := (TileID{3, 9, 0}).Parent(); err == nil {
		t.Error("an invalid tile has no parent")
	}
}

func TestJobKindValidate(t *testing.T) {
	for _, k := range []JobKind{KindTile, KindOverlayPrepare, KindDescribe} {
		if err := k.Validate(); err != nil {
			t.Errorf("kind %d: %v", k, err)
		}
	}
	for _, k := range []JobKind{0, KindDescribe + 1, 255} {
		if err := k.Validate(); err == nil {
			t.Errorf("kind %d is not a kind, and Validate must say so", k)
		}
	}
}

// oneJob shows the interface can be met by a plain struct in another
// package's shape: a kind, a key for de-duplication, and the slow part.
type oneJob struct{ ran bool }

func (j *oneJob) Kind() JobKind { return KindTile }
func (j *oneJob) Key() string   { return "tile/0/0/0" }
func (j *oneJob) Run(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	j.ran = true
	return nil
}

func TestJobIsSatisfiedByAPlainStruct(t *testing.T) {
	var j Job = &oneJob{}
	if err := j.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := j.Run(ctx); err == nil {
		t.Error("a job must give up when its context has ended")
	}
}
