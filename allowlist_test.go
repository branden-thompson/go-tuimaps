package tuimaps_test

import (
	"os"
	"os/exec"
	"testing"
)

// TestDisallowedModulesFindsStrangers checks the checker: only the library
// itself and the two text modules of rulings D-75 and D-81 are allowed.
func TestDisallowedModulesFindsStrangers(t *testing.T) {
	cases := []struct {
		name string
		list string
		want []string
	}{
		{"the library alone", "github.com/branden-thompson/go-tuimaps\n", nil},
		{"both allowed modules", "github.com/branden-thompson/go-tuimaps\ngithub.com/clipperhouse/uax29/v2 v2.2.0\ngithub.com/mattn/go-runewidth v0.0.24\n", nil},
		{"a stranger", "github.com/branden-thompson/go-tuimaps\ngithub.com/paulmach/orb v0.11.1\n", []string{"github.com/paulmach/orb"}},
		{"an older major version of an allowed module", "github.com/clipperhouse/uax29 v1.14.0\n", []string{"github.com/clipperhouse/uax29"}},
		{"a module that only starts like an allowed one", "github.com/mattn/go-runewidth-extra v1.0.0\n", []string{"github.com/mattn/go-runewidth-extra"}},
		{"blank lines and a replacement arrow", "\ngithub.com/x/y v1.0.0 => ../y\n", []string{"github.com/x/y"}},
	}
	for _, c := range cases {
		got := disallowedModules(c.list)
		if len(got) != len(c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%s: got %v, want %v", c.name, got, c.want)
			}
		}
	}
}

// TestAllowList is plan task 00.3: with the workspace file switched off, the
// library's module graph names nothing beyond the two allowed modules.
func TestAllowList(t *testing.T) {
	cmd := exec.Command("go", "list", "-m", "all")
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=-mod=readonly")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list -m all: %v", err)
	}
	for _, m := range disallowedModules(string(out)) {
		t.Errorf("module %s is in the library's graph; only go-runewidth and uax29/v2 are allowed (D-75, D-81)", m)
	}
}
