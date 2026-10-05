package tuimaps_test

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// held is a transport that keeps its first request waiting until the test
// lets it go, then answers it with a real tile; every later request is
// answered "not found", so the one held is the only tile that could land.
type held struct {
	once    sync.Once
	started chan struct{}
	release chan struct{}
}

func holding() *held {
	return &held{started: make(chan struct{}), release: make(chan struct{})}
}

func (h *held) RoundTrip(r *http.Request) (*http.Response, error) {
	first := false
	h.once.Do(func() { first = true })
	if !first {
		return answering(func(string) ([]byte, error) { return nil, errors.New("not this one") }).RoundTrip(r)
	}
	close(h.started)
	<-h.release
	return answering(func(string) ([]byte, error) {
		body, _ := assets.Tile(0, 0, 0)
		return body, nil
	}).RoundTrip(r)
}

// cachedFiles counts the regular files under a cache root.
func cachedFiles(t *testing.T, root string) int {
	t.Helper()
	n := 0
	_ = filepath.Walk(root, func(_ string, info os.FileInfo, err error) error {
		if err == nil && info.Mode().IsRegular() {
			n++
		}
		return nil
	})
	return n
}

// TestNoWriteLandsAfterPurgeRootChangeOrClose is v0.2.0 L9.3 (L-9.3): a fetch
// in flight across a Purge, a CacheRoot change, turning the cache off, or
// Close writes nothing to the disk.
func TestNoWriteLandsAfterPurgeRootChangeOrClose(t *testing.T) {
	for _, c := range []struct {
		name string
		do   func(m *tuimaps.Map, other string) error
	}{
		{"Purge", func(m *tuimaps.Map, _ string) error { _, err := m.Purge(); return err }},
		{"CacheRoot to another root", func(m *tuimaps.Map, other string) error { return m.CacheRoot(other, 0) }},
		{"CacheRoot off", func(m *tuimaps.Map, _ string) error { return m.CacheRoot("", 0) }},
		{"Close", func(m *tuimaps.Map, _ string) error { m.Close(); return nil }},
	} {
		t.Run(c.name, func(t *testing.T) {
			root, other := filepath.Join(t.TempDir(), "cache"), filepath.Join(t.TempDir(), "other")
			m, err := tuimaps.New(tuimaps.WithSize(40, 12))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			h := holding()
			must(t, m.CacheRoot(root, 0))
			useTransport(t, m, h)
			must(t, m.Source("https://tiles.example.test/"))
			if _, err := m.Render(tuimaps.Size{Cols: 40, Rows: 12}, noon); err != nil {
				t.Fatal(err)
			}
			done := make(chan struct{})
			go func() {
				defer close(done)
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_, _ = m.Settle(ctx)
			}()
			select {
			case <-h.started:
			case <-time.After(5 * time.Second):
				t.Fatal("no fetch began")
			}
			must(t, c.do(m, other))
			close(h.release)
			<-done
			if n := cachedFiles(t, root) + cachedFiles(t, other); n != 0 {
				t.Errorf("%d files written by a fetch that began before it", n)
			}
		})
	}
}

// TestPurgeEmptiesEverything is v0.2.0 L9.4 (L-9.3, L-9.5): every source's
// tiles on disk and the tiles in memory go, and the report counts them.
func TestPurgeEmptiesEverything(t *testing.T) {
	root := filepath.Join(t.TempDir(), "cache")
	m, err := tuimaps.New(tuimaps.WithSize(40, 12))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	must(t, m.CacheRoot(root, 0))
	asked := 0
	useTransport(t, m, served(t, &asked))
	for _, source := range []string{"https://one.example.test/", "https://two.example.test/"} {
		must(t, m.Source(source))
		drawAndSettle(t, m)
	}
	before := cachedFiles(t, root)
	if before == 0 || m.CacheUse().Tiles.Held == 0 {
		t.Fatal("nothing was cached, so this proves nothing")
	}
	report, err := m.Purge()
	if err != nil {
		t.Fatal(err)
	}
	if report.Removed != before || report.Failed != 0 {
		t.Errorf("report %+v; want %d removed", report, before)
	}
	if n := cachedFiles(t, root); n != 0 {
		t.Errorf("%d files left", n)
	}
	if held := m.CacheUse().Tiles.Held; held != 0 {
		t.Errorf("%d bytes of tiles left in memory", held)
	}
	if use := m.CacheUse().Disk; use.Held != 0 {
		t.Errorf("the disk reports %d bytes held", use.Held)
	}
}

// TestCacheMaxAge is v0.2.0 L9.2 (L-9.1, L-9.4, D-70) through the Map: a
// maximum age set before or after the root is kept, at CacheRoot and in
// every job, and a negative one is refused.
func TestCacheMaxAge(t *testing.T) {
	root := filepath.Join(t.TempDir(), "cache")
	m, err := tuimaps.New(tuimaps.WithSize(40, 12))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if err := m.SetCacheMaxAge(-time.Second); !isKind(err, fault.CacheRefused) {
		t.Errorf("a negative age: %v, want cache-refused", err)
	}
	must(t, m.CacheRoot(root, 0))
	asked := 0
	useTransport(t, m, served(t, &asked))
	must(t, m.Source("https://tiles.example.test/"))
	drawAndSettle(t, m)
	if cachedFiles(t, root) == 0 {
		t.Fatal("nothing was cached")
	}
	// The files were dated noon, the host's clock. An hour's age, set once
	// that clock has passed it, empties the root at once.
	if _, err := m.Render(tuimaps.Size{Cols: 40, Rows: 12}, noon.Add(2*time.Hour)); err != nil {
		t.Fatal(err)
	}
	must(t, m.SetCacheMaxAge(time.Hour))
	if n := cachedFiles(t, root); n != 0 {
		t.Errorf("%d files past their age kept after SetCacheMaxAge", n)
	}

	// An age set before the root is kept when the root is named: a root
	// holding tiles fetched at noon is emptied by a map whose clock is later.
	again := filepath.Join(t.TempDir(), "again")
	filler, err := tuimaps.New(tuimaps.WithSize(40, 12))
	if err != nil {
		t.Fatal(err)
	}
	must(t, filler.CacheRoot(again, 0))
	useTransport(t, filler, served(t, &asked))
	must(t, filler.Source("https://tiles.example.test/"))
	drawAndSettle(t, filler)
	filler.Close()
	if cachedFiles(t, again) == 0 {
		t.Fatal("nothing was cached")
	}
	later, err := tuimaps.New(tuimaps.WithSize(40, 12))
	if err != nil {
		t.Fatal(err)
	}
	defer later.Close()
	if _, err := later.Render(tuimaps.Size{Cols: 40, Rows: 12}, noon.Add(2*time.Hour)); err != nil {
		t.Fatal(err)
	}
	must(t, later.SetCacheMaxAge(time.Hour))
	must(t, later.CacheRoot(again, 0))
	if n := cachedFiles(t, again); n != 0 {
		t.Errorf("%d files past their age kept after CacheRoot", n)
	}
}

// TestPurgeCountsWhatItCouldNotRemove is L9.5 through the Map (L-9.5).
func TestPurgeCountsWhatItCouldNotRemove(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("NOT RUN: a directory this user cannot remove from cannot be made here")
	}
	root := filepath.Join(t.TempDir(), "cache")
	m, err := tuimaps.New(tuimaps.WithSize(40, 12))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	must(t, m.CacheRoot(root, 0))
	asked := 0
	useTransport(t, m, served(t, &asked))
	must(t, m.Source("https://tiles.example.test/"))
	drawAndSettle(t, m)
	total := cachedFiles(t, root)
	var locked string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.Mode().IsRegular() && locked == "" {
			locked = filepath.Dir(path)
		}
		return nil
	})
	inLocked := cachedFiles(t, locked)
	must(t, os.Chmod(locked, 0o500))
	t.Cleanup(func() { os.Chmod(locked, 0o700) })
	report, err := m.Purge()
	if err == nil || report.Failed != inLocked || report.Removed != total-inLocked {
		t.Errorf("report %+v, %v; want %d removed and %d failed, and an error", report, err, total-inLocked, inLocked)
	}
}

// openFiles counts this process's open file descriptors.
func openFiles(t *testing.T) int {
	t.Helper()
	entries, err := os.ReadDir("/dev/fd")
	if err != nil {
		t.Skipf("NOT RUN: open descriptors cannot be counted here: %v", err)
	}
	return len(entries)
}

// TestAReplacedRootIsLetGo is v0.2.0 L9.4 (L-9.3): a root that CacheRoot
// replaces or turns off, and the root a closed map held, has its descriptor
// closed.
func TestAReplacedRootIsLetGo(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("NOT RUN: open descriptors are not counted on Windows")
	}
	m, err := tuimaps.New(tuimaps.WithSize(40, 12))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	before := openFiles(t)
	for i := range 5 {
		must(t, m.CacheRoot(filepath.Join(t.TempDir(), "cache"+string(rune('a'+i))), 0))
	}
	must(t, m.CacheRoot("", 0))
	if after := openFiles(t); after > before {
		t.Errorf("%d descriptors open after replacing roots and turning the cache off; %d before", after, before)
	}
	must(t, m.CacheRoot(filepath.Join(t.TempDir(), "last"), 0))
	m.Close()
	if after := openFiles(t); after > before {
		t.Errorf("%d descriptors open after Close; %d before", after, before)
	}
}

// TestPurgeEmptiesTheSharedReadings is L9.4 through the Map (L-9.3): the
// decoded pictures kept for reuse go with a purge, even with no disk cache.
func TestPurgeEmptiesTheSharedReadings(t *testing.T) {
	shared, err := tuimaps.NewShared(0)
	if err != nil {
		t.Fatal(err)
	}
	m, err := tuimaps.New(tuimaps.WithSize(40, 12), tuimaps.SharedCaches(shared))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if _, err := m.Set(radarLoop(t, "radar", 3, 8, 6)); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	before := tuimaps.ImageUse(m)
	if _, err := m.Purge(); !isKind(err, fault.CacheRefused) {
		t.Errorf("Purge with no disk cache: %v, want cache-refused", err)
	}
	if after := tuimaps.ImageUse(m); after != before-8*6 {
		t.Errorf("image use %d after a purge, %d before; want the shared reading's %d bytes gone", after, before, 8*6)
	}
}

// TestTheDefaultTileCacheAgainstALargeView is v0.2.0 L10.6 (L-6.3): one
// large view needs more than the default memory cache. The cap is a target,
// not a limit: the view keeps every tile it draws, no spares, and the host is
// told once, with the figures in CacheUse.
func TestTheDefaultTileCacheAgainstALargeView(t *testing.T) {
	size := tuimaps.Size{Cols: 200, Rows: 60}
	m, err := tuimaps.New(tuimaps.WithSize(size.Cols, size.Rows))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	useTransport(t, m, answering(func(string) ([]byte, error) { // every tile real bytes, so the view is whole
		body, _ := assets.Tile(2, 1, 1)
		return body, nil
	}))
	must(t, m.Source("https://tiles.example.test/"))
	must(t, m.Recentre(tuimaps.LonLat{Lon: -97, Lat: 38}))
	must(t, m.Zoom(4))
	if _, err := m.Render(size, noon); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Render(size, noon); err != nil {
		t.Fatal(err)
	}
	use := m.CacheUse().Tiles
	if use.Limit != 500_000 || use.Need <= use.Limit || use.Held != use.Need {
		t.Fatalf("tile cache %+v; want the default 500000, a need over it, and exactly the need held", use)
	}
	told := 0
	for range 2 {
		for _, w := range m.Warnings() {
			if w.Kind == tuimaps.CacheUnderNeed {
				told++
			}
		}
	}
	if told != 1 {
		t.Errorf("cache-under-need told %d times; want once", told)
	}
}
