package tuimaps_test

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/fetch"
)

// served is a fetcher that answers with the embedded tiles, so that a test
// of the source path reaches nothing at all.
func served(t *testing.T, asked *int) tuimaps.Fetcher {
	t.Helper()
	return func(ctx context.Context, r fetch.Request) ([]byte, error) {
		*asked++
		z, x, y, ok := tileOf(r.URL)
		if !ok {
			return nil, errors.New("not a tile address")
		}
		body, held := assets.Tile(z, x, y)
		if !held {
			return nil, errors.New("no such tile")
		}
		return body, nil
	}
}

// tileOf reads z, x and y out of an address ending in z/x/y.pbf.
func tileOf(address string) (uint8, uint32, uint32, bool) {
	address = strings.TrimSuffix(address, ".pbf")
	parts := strings.Split(address, "/")
	if len(parts) < 3 {
		return 0, 0, 0, false
	}
	var z, x, y int
	for i, at := range []*int{&z, &x, &y} {
		n := 0
		text := parts[len(parts)-3+i]
		if text == "" {
			return 0, 0, 0, false
		}
		for _, r := range text {
			if r < '0' || r > '9' {
				return 0, 0, 0, false
			}
			n = n*10 + int(r-'0')
		}
		*at = n
	}
	return uint8(z), uint32(x), uint32(y), true
}

// TestSourceReachesNothingUntilNamed is plan task 12.18 (D-65): a map with
// no source named reaches nothing; one with a source named fetches through
// the host's own fetcher, and nothing else.
func TestSourceReachesNothingUntilNamed(t *testing.T) {
	m, err := tuimaps.New(tuimaps.WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	asked := 0
	m.Fetcher(served(t, &asked))
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	if asked != 0 {
		t.Errorf("a map with no source named reached out %d times", asked)
	}
	if err := m.Source("https://tiles.example.test/"); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Render(tuimaps.Size{Cols: 80, Rows: 24}, noon); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	if asked == 0 {
		t.Error("a map with a source named fetched nothing")
	}
	// A source that is not an address is refused, and the map keeps the one
	// it had.
	for _, bad := range []string{"tiles.example.test", "ftp://tiles.example.test/", "https://user:pw@tiles.example.test/"} {
		if err := m.Source(bad); err == nil {
			t.Errorf("Source(%q) was accepted", bad)
		}
	}
}

// TestCacheRootAndMaintenance is the rest of 12.18 (FR-21b, FR-22a): tiles
// kept on disk between runs, and the two maintenance calls that are the
// host's to make.
func TestCacheRootAndMaintenance(t *testing.T) {
	dir := t.TempDir()
	m, err := tuimaps.New(tuimaps.WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	// Before a cache is named, the maintenance calls say so rather than
	// pretending to work.
	if err := m.Purge(); !isKind(err, fault.CacheRefused) {
		t.Errorf("Purge with no cache: %v", err)
	}
	if _, _, err := m.Verify(); !isKind(err, fault.CacheRefused) {
		t.Errorf("Verify with no cache: %v", err)
	}
	asked := 0
	m.Fetcher(served(t, &asked))
	if err := m.CacheRoot(filepath.Join(dir, "tiles"), 0); err != nil {
		t.Fatal(err)
	}
	if err := m.Source("https://tiles.example.test/"); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Render(tuimaps.Size{Cols: 80, Rows: 24}, noon); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	use := m.CacheUse()
	if use.Tiles.Held <= 0 || use.Tiles.Limit <= 0 {
		t.Errorf("the tile cache holds %+v", use.Tiles)
	}
	if use.Disk.Held <= 0 {
		t.Errorf("nothing reached the disk cache: %+v", use.Disk)
	}
	// Verify reads what is there and removes nothing that still decodes.
	checked, removed, err := m.Verify()
	if err != nil || checked == 0 || removed != 0 {
		t.Errorf("Verify: %d checked, %d removed, %v", checked, removed, err)
	}
	// Purge empties it.
	if err := m.Purge(); err != nil {
		t.Fatal(err)
	}
	if after := m.CacheUse(); after.Disk.Held != 0 {
		t.Errorf("the disk cache holds %d bytes after a purge", after.Disk.Held)
	}
}

// TestCacheUseWithNoCaches: a map with nothing named still answers, with the
// caps it has and nothing held.
func TestCacheUseWithNoCaches(t *testing.T) {
	m := world(t, 80, 24)
	use := m.CacheUse()
	if use.Tiles.Limit <= 0 || use.Shapes.Limit <= 0 {
		t.Errorf("the caps are %+v", use)
	}
	if use.Disk.Limit != 0 || use.Disk.Held != 0 {
		t.Errorf("a map with no disk cache reports %+v", use.Disk)
	}
	if got := m.SourceCredit(); got != "" {
		t.Errorf("a map drawing the embedded tiles credits a source: %q", got)
	}
}
