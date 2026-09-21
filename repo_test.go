package tuimaps_test

import (
	"os"
	"path/filepath"
	"testing"
)

// moduleFiles lists every module file in the repository. The separate
// modules resolve the library through an untracked workspace file, so none
// of them may carry a replace directive (plan task 00.2).
var moduleFiles = []string{
	"go.mod",
	"cmd/tuimaps/go.mod",
	"examples/go.mod",
	"tools/gen-assets/go.mod",
	"tools/oracle/go.mod",
	"tools/answer-key/go.mod",
}

func TestReplaceLinesFindsEveryForm(t *testing.T) {
	cases := []struct {
		name string
		mod  string
		want int
	}{
		{"none", "module m\n\ngo 1.25.0\n", 0},
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
	for _, f := range moduleFiles {
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
