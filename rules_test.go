package tuimaps_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/rules"
	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

func TestMain(m *testing.M) { os.Exit(testkit.Main(m)) }

// TestStaticRules is plan task 00.11 applied to the library itself: no
// goroutine, no clock, no output, no forbidden import, and every test
// package held to loopback.
func TestStaticRules(t *testing.T) {
	found, err := rules.Check(".", testkit.ModulePath)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range found {
		t.Error(f)
	}
}

// TestTheImagePathNeverFetches is v0.2.0 L2.10 (L-1.4): the library never
// fetches radar; the host hands in the data. No package on the image path -
// the overlay store, the renderer, the scene they share, the colours - reaches
// the fetcher or the network, directly or through anything it imports.
func TestTheImagePathNeverFetches(t *testing.T) {
	forbidden := map[string]bool{testkit.ModulePath + "/internal/fetch": true, "net/http": true, "net": true}
	seen := map[string]bool{}
	var walk func(pkg string, path []string)
	walk = func(pkg string, path []string) {
		if seen[pkg] {
			return
		}
		seen[pkg] = true
		dir := strings.TrimPrefix(strings.TrimPrefix(pkg, testkit.ModulePath), "/")
		names, err := filepath.Glob(filepath.Join(dir, "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range names {
			if strings.HasSuffix(name, "_test.go") {
				continue
			}
			file, err := parser.ParseFile(token.NewFileSet(), name, nil, parser.ImportsOnly)
			if err != nil {
				t.Fatal(err)
			}
			for _, spec := range file.Imports {
				imported := strings.Trim(spec.Path.Value, `"`)
				if forbidden[imported] {
					t.Errorf("%s imports %s", strings.Join(append(path, pkg), " → "), imported)
				}
				if strings.HasPrefix(imported, testkit.ModulePath+"/") {
					walk(imported, append(path, pkg))
				}
			}
		}
	}
	for _, pkg := range []string{"overlay", "render", "scene", "colour"} {
		walk(testkit.ModulePath+"/internal/"+pkg, nil)
	}
	if len(seen) < 4 {
		t.Fatalf("walked %d packages; the image path has at least four", len(seen))
	}
}
