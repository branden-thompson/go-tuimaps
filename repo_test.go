package tuimaps_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// moduleFiles lists every module file in the repository, found as the gate
// finds them: every go.mod outside testdata, hidden directories, `_`-prefixed
// directories and dist. The separate modules resolve the library through an
// untracked workspace file, so none of them may carry a replace directive
// (plan task 00.2).
func moduleFiles(t *testing.T) []string {
	t.Helper()
	var found []string
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && path != "." {
			name := d.Name()
			if name == "testdata" || name == "dist" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
				return filepath.SkipDir
			}
		}
		if !d.IsDir() && d.Name() == "go.mod" {
			found = append(found, filepath.ToSlash(path))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(found)
	return found
}

// TestModuleFilesFindsEveryModule: the walk finds the modules the repository
// is known to hold, so a walk that finds nothing cannot pass the checks that
// range over it.
func TestModuleFilesFindsEveryModule(t *testing.T) {
	found := moduleFiles(t)
	got := map[string]bool{}
	for _, f := range found {
		got[f] = true
	}
	for _, want := range []string{"go.mod", "cmd/tuimaps/go.mod", "examples/go.mod", "tools/answer-key/go.mod", "tools/atlas/go.mod", "tools/gen-assets/go.mod", "tools/oracle/go.mod"} {
		if !got[want] {
			t.Errorf("the module walk did not find %s; found %v", want, found)
		}
	}
}

func TestReplaceLinesFindsEveryForm(t *testing.T) {
	cases := []struct {
		name string
		mod  string
		want int
	}{
		{"none", "module m\n\ngo 1.25.13\n", 0},
		{"single line", "module m\nreplace a => ../a\n", 1},
		{"indented", "module m\n\t replace a => ../a\n", 1},
		{"block", "module m\nreplace (\n\ta => ../a\n\tb => ../b\n)\n", 1},
		{"in a comment", "module m\n// replace a => ../a\n", 0},
		{"a module named replacer", "module m\nrequire replacer.example/x v1.0.0\n", 0},
	}
	for _, c := range cases {
		if got := len(replaceLines([]byte(c.mod))); got != c.want {
			t.Errorf("%s: found %d replace directives, want %d", c.name, got, c.want)
		}
	}
}

func TestModuleHasNoReplace(t *testing.T) {
	for _, f := range moduleFiles(t) {
		data, err := os.ReadFile(filepath.FromSlash(f))
		if err != nil {
			t.Errorf("%s: %v", f, err)
			continue
		}
		for _, l := range replaceLines(data) {
			t.Errorf("%s: replace directive %q; tracked module files carry none", f, l)
		}
	}
}

// moduleFloor is the go directive every module carries (D-28, after v0.2.0
// D-133): the first release whose standard library has none of the
// vulnerabilities the code reaches, GO-2026-6617 the latest. The gate reads
// its floor toolchain from the root module's directive.
const moduleFloor = "1.26.9"

// TestEveryModuleIsAtTheFloor (D-133): every module file's go directive is
// the floor, so no module builds against an older standard library.
func TestEveryModuleIsAtTheFloor(t *testing.T) {
	directive := regexp.MustCompile(`(?m)^go[ \t]+(\S+)[ \t]*$`)
	for _, f := range moduleFiles(t) {
		data, err := os.ReadFile(filepath.FromSlash(f))
		if err != nil {
			t.Errorf("%s: %v", f, err)
			continue
		}
		m := directive.FindAllSubmatch(data, -1)
		if len(m) != 1 {
			t.Errorf("%s: %d go directives; want one", f, len(m))
			continue
		}
		if got := string(m[0][1]); got != moduleFloor {
			t.Errorf("%s: go %s; every module is at go %s (D-133)", f, got, moduleFloor)
		}
	}
}

func TestWorkspaceFileIsIgnored(t *testing.T) {
	data, err := os.ReadFile(".gitignore")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"go.work", "go.work.sum"} {
		if !hasLine(data, want) {
			t.Errorf(".gitignore does not list %q; the workspace file must never be tracked", want)
		}
	}
}
