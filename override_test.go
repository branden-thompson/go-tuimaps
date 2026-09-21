package tuimaps_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// hostMain is the whole of the throw-away host: it draws a map, which is
// enough to make the toolchain build the library and everything it needs.
const hostMain = `package main

import (
	"context"
	"fmt"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
)

func main() {
	m, err := tuimaps.New(tuimaps.WithSize(40, 12), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		panic(err)
	}
	defer m.Close()
	if _, err := m.Settle(context.Background()); err != nil {
		panic(err)
	}
	frame, err := m.Render(tuimaps.Size{Cols: 40, Rows: 12}, time.Unix(0, 0))
	if err != nil {
		panic(err)
	}
	fmt.Println(len(frame.Lines), frame.Status)
}
`

// TestLocalOverrideRecipe is plan task 12.24 (D-19, CD-4): no remote exists
// yet, so the first host builds against this tree by the recipe the
// documents give. The test follows it exactly, with the module proxy off,
// so the recipe cannot rot unnoticed and the test never reaches a network.
func TestLocalOverrideRecipe(t *testing.T) {
	if testing.Short() {
		t.Skip("it builds a module of its own")
	}
	tree, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	host := t.TempDir()
	// Step 1 of the recipe: require the library at the placeholder version,
	// and replace that version with the tree on disk. The host's own floor
	// is the library's.
	modFile := "module example.test/host\n\ngo 1.25.0\n\nrequire github.com/branden-thompson/go-tuimaps v0.0.0\n\n" +
		"replace github.com/branden-thompson/go-tuimaps v0.0.0 => " + tree + "\n"
	write(t, filepath.Join(host, "go.mod"), modFile)
	write(t, filepath.Join(host, "main.go"), hostMain)
	// The library's own requirements and sums come with it: a host that
	// already carries both text modules gains nothing else (D-81).
	for _, name := range []string{"go.sum"} {
		body, err := os.ReadFile(filepath.Join(tree, name))
		if err != nil {
			t.Fatal(err)
		}
		write(t, filepath.Join(host, name), string(body))
	}
	appendTo(t, filepath.Join(host, "go.mod"), "\nrequire (\n\tgithub.com/clipperhouse/uax29/v2 v2.7.0 // indirect\n\tgithub.com/mattn/go-runewidth v0.0.24 // indirect\n)\n")

	// Step 3: the host builds, with nothing fetched.
	out := filepath.Join(host, "host")
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	build := exec.Command("go", "build", "-o", out, ".")
	build.Dir = host
	build.Env = append(os.Environ(), "GOPROXY=off", "GOFLAGS=-mod=mod", "GOWORK=off")
	if report, err := build.CombinedOutput(); err != nil {
		t.Fatalf("the recipe no longer builds a host:\n%s\n%v", report, err)
	}
	// And the host it built draws a map.
	run := exec.Command(out)
	run.Dir = host
	report, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("the host built by the recipe does not run:\n%s\n%v", report, err)
	}
	if got := strings.TrimSpace(string(report)); got != "12 complete" {
		t.Errorf("the host printed %q, want a complete frame of twelve rows", got)
	}
	// The module graph gained the library and the two text modules, and
	// nothing else (D-81).
	list := exec.Command("go", "list", "-m", "all")
	list.Dir = host
	list.Env = build.Env
	graph, err := list.CombinedOutput()
	if err != nil {
		t.Fatalf("listing the host's modules:\n%s\n%v", graph, err)
	}
	lines := strings.Split(strings.TrimSpace(string(graph)), "\n")
	if len(lines) != 4 {
		t.Errorf("the host's module graph is\n%s\nwant itself, the library and the two text modules", graph)
	}
}

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func appendTo(t *testing.T, path, body string) {
	t.Helper()
	was, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	write(t, path, string(was)+body)
}
