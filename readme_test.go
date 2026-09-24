package tuimaps_test

import (
	"context"
	"go/parser"
	"go/token"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
	"github.com/branden-thompson/go-tuimaps/internal/fetch"
)

// TestReadmeQuickStartBuilds is plan task 12.15: the quick start in the
// README is taken out of the file and parsed, so it cannot rot unnoticed.
// Its calls are the ones the example test runs, which is what proves it
// works; this holds it to being real Go that names the real calls.
func TestReadmeQuickStartBuilds(t *testing.T) {
	body, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	blocks := goBlocks(string(body))
	if len(blocks) == 0 {
		t.Fatal("the README has no Go in it at all")
	}
	for i, block := range blocks {
		if _, err := parser.ParseFile(token.NewFileSet(), "readme.go", block, parser.AllErrors); err != nil {
			t.Errorf("block %d of the README is not Go: %v", i, err)
		}
	}
	quick := blocks[0]
	for _, call := range []string{"tuimaps.New", "tuimaps.WithSize", "tuimaps.Embed", "assets.Tile", "m.Settle", "m.Render", "m.Close"} {
		if !strings.Contains(quick, call) {
			t.Errorf("the quick start does not use %s", call)
		}
	}
	// The three calls are three: create, settle, render.
	if n := strings.Count(quick, "m.Settle") + strings.Count(quick, "m.Render") + strings.Count(quick, "tuimaps.New"); n != 3 {
		t.Errorf("the quick start makes %d of the three calls", n)
	}
}

// goBlocks are the fenced Go blocks of a markdown file.
func goBlocks(body string) []string {
	var out []string
	parts := strings.Split(body, "```")
	for i := 1; i < len(parts); i += 2 {
		block := parts[i]
		if head, rest, ok := strings.Cut(block, "\n"); ok && strings.TrimSpace(head) == "go" {
			out = append(out, rest)
		}
	}
	return out
}

// TestReadmeSaysWhatIsOwed: the README tells a host the things it cannot
// find out for itself - the braille font, what is sent and stored, the
// credit the data asks for, and what is not built yet.
func TestReadmeSaysWhatIsOwed(t *testing.T) {
	body, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, owed := range []string{"braille", "NO_COLOR", "OpenStreetMap", "About", "CacheRoot", "Not built yet", "MIT"} {
		if !strings.Contains(text, owed) {
			t.Errorf("the README does not mention %q", owed)
		}
	}
}

// tileServer serves the embedded tiles over plain http on loopback under
// /<dir>/z/x/y.pbf, which the library's own fetcher reaches with no
// replacement (plain http is allowed to a literal loopback address). It
// records every request, so a test sees what the library really sent.
type tileServer struct {
	mu     sync.Mutex
	agents []string
	paths  []string
	url    string
}

func serveTiles(t *testing.T) *tileServer {
	t.Helper()
	s := &tileServer{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.agents = append(s.agents, r.Header.Get("User-Agent"))
		s.paths = append(s.paths, r.URL.Path)
		s.mu.Unlock()
		z, x, y, ok := tileOf(r.URL.Path)
		body, held := assets.Tile(z, x, y)
		if !ok || !held {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	s.url = srv.URL
	return s
}

func (s *tileServer) requests() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.paths)
}

// drawAndSettle asks for one view's tiles and waits for them.
func drawAndSettle(t *testing.T, m *tuimaps.Map) {
	t.Helper()
	if _, err := m.Render(tuimaps.Size{Cols: 80, Rows: 24}, noon); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
}

// TestReadmeWhatItSendsIsHeldByBehaviour is v0.2.0 L1.4 (L-4.4, D-34): the
// README's "What it sends and stores" is held by what the library does, over
// its own fetcher, not by the sentences being present. Nothing is reached
// until a source is named or after it is taken away; with one named, only its
// address is asked for, and the user-agent names the library and its fixed
// version and nothing else - no machine, and no name of the host's own.
func TestReadmeWhatItSendsIsHeldByBehaviour(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", home)
	srv := serveTiles(t)
	m, err := tuimaps.New(tuimaps.WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()

	drawAndSettle(t, m)
	if n := srv.requests(); n != 0 {
		t.Errorf("with no source named the library asked %d times", n)
	}

	if err := m.Source(srv.url + "/tiles/"); err != nil {
		t.Fatal(err)
	}
	drawAndSettle(t, m)
	named := srv.requests()
	if named == 0 {
		t.Fatal("with a source named nothing was asked for, so this proves nothing")
	}
	srv.mu.Lock()
	for i, path := range srv.paths {
		if !strings.HasPrefix(path, "/tiles/") {
			t.Errorf("asked for %q, outside the source named", path)
		}
		if agent := srv.agents[i]; agent != "go-tuimaps/"+fetch.Version {
			t.Errorf("user-agent %q; the README promises the library's name and fixed version only", agent)
		}
	}
	srv.mu.Unlock()

	// The source taken away: a new place asks for nothing more.
	if err := m.Source(""); err != nil {
		t.Fatal(err)
	}
	if err := m.Zoom(4); err != nil {
		t.Fatal(err)
	}
	if err := m.Recentre(tuimaps.LonLat{Lon: 151.2, Lat: -33.9}); err != nil {
		t.Fatal(err)
	}
	drawAndSettle(t, m)
	if n := srv.requests(); n != named {
		t.Errorf("with the source taken away the library asked %d more times", n-named)
	}

	// On disk only if a directory is named: none was, and nothing was written.
	if entries, err := os.ReadDir(home); err != nil || len(entries) != 0 {
		t.Errorf("with no CacheRoot the library wrote %d entries under the home directory (%v)", len(entries), err)
	}
	if use := m.CacheUse(); use.Disk.Held != 0 {
		t.Errorf("with no CacheRoot the disk cache holds %d bytes", use.Disk.Held)
	}
}

// TestReadmePurgeEmptiesTheCurrentSourceOnly is the rest of L1.4: Purge
// empties the current source's tiles from the disk cache, and another
// source's tiles in the same directory stay, so a later map on that source
// asks for nothing. L9.4 makes Purge empty every source (L-9.3, the
// changelog's Purge row); the README and this test change with it.
func TestReadmePurgeEmptiesTheCurrentSourceOnly(t *testing.T) {
	root := filepath.Join(t.TempDir(), "tiles")
	one, two := serveTiles(t), serveTiles(t)
	m, err := tuimaps.New(tuimaps.WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	if err := m.CacheRoot(root, 0); err != nil {
		t.Fatal(err)
	}
	for _, s := range []*tileServer{one, two} {
		if err := m.Source(s.url + "/tiles/"); err != nil {
			t.Fatal(err)
		}
		drawAndSettle(t, m)
	}
	if err := m.Purge(); err != nil { // the current source is the second
		t.Fatal(err)
	}
	if use := m.CacheUse(); use.Disk.Held == 0 {
		t.Error("Purge emptied every source's tiles, not the current one's")
	}
	m.Close()

	// A fresh map on the same directory: the first source is still served
	// from disk; the purged one is asked for again.
	for _, c := range []struct {
		s     *tileServer
		again bool
	}{{one, false}, {two, true}} {
		fresh, err := tuimaps.New(tuimaps.WithSize(80, 24))
		if err != nil {
			t.Fatal(err)
		}
		if err := fresh.CacheRoot(root, 0); err != nil {
			t.Fatal(err)
		}
		if err := fresh.Source(c.s.url + "/tiles/"); err != nil {
			t.Fatal(err)
		}
		before := c.s.requests()
		drawAndSettle(t, fresh)
		fresh.Close()
		if asked := c.s.requests() > before; asked != c.again {
			t.Errorf("source %s after the purge: asked again %v, want %v", c.s.url, asked, c.again)
		}
	}
}
