package tuimaps_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// TestHostIndependence is plan task 14.12, and metric M5: **the library is
// no one's widget.** Its module graph names no terminal library, no
// user-interface framework and no way of drawing; a host brings its own,
// and the library hands it cells and facts (D-73, NFR-9).
func TestHostIndependence(t *testing.T) {
	out, err := exec.Command("go", "list", "-m", "all").Output()
	if err != nil {
		t.Fatal(err)
	}
	// The names of what a terminal application usually reaches for. A
	// library that reached for any of them would be making the host's
	// decisions for it.
	for _, framework := range []string{
		"bubbletea", "tcell", "termbox", "tview", "termui", "gocui", "pterm",
		"lipgloss", "bubbles", "readline", "golang.org/x/term", "curses", "notcurses",
	} {
		if strings.Contains(string(out), framework) {
			t.Errorf("the library's module graph names %s; the library draws no terminal of its own (M5)", framework)
		}
	}
	// And the two it does name are the two that were ruled (D-75, D-81).
	for _, m := range disallowedModules(string(out)) {
		t.Errorf("module %s is in the library's graph", m)
	}
	// The library's own source imports nothing that owns a terminal, which
	// is the same rule read from the code rather than from the graph.
	for _, dir := range []string{".", "assets"} {
		for _, file := range goFilesIn(t, dir) {
			body, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			for _, owned := range []string{`"os/exec"`, `"os/signal"`, `"syscall"`} {
				if strings.Contains(string(body), owned) && !strings.HasSuffix(file, "_test.go") {
					t.Errorf("%s imports %s; the terminal, its signals and its processes are the host's (D-73)", file, owned)
				}
			}
		}
	}
}

// goFilesIn is the Go files of one directory, not of the tree below it.
func goFilesIn(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var found []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			found = append(found, filepath.Join(dir, entry.Name()))
		}
	}
	return found
}
