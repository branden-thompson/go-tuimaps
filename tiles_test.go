package tuimaps_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// answering is a host's own transport that answers each tile address from a
// function: what a host that serves tiles from a store of its own hands the
// library (L-7.1). Nothing is dialled.
type answering func(address string) ([]byte, error)

func (a answering) RoundTrip(r *http.Request) (*http.Response, error) {
	body, err := a(r.URL.String())
	if err != nil {
		return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader("")), Header: http.Header{}, Request: r}, nil
	}
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(body)), ContentLength: int64(len(body)), Header: http.Header{}, Request: r}, nil
}

// served is a transport that answers with the embedded tiles, so that a
// test of the source path reaches nothing at all.
func served(t *testing.T, asked *int) http.RoundTripper {
	t.Helper()
	return answering(func(address string) ([]byte, error) {
		*asked++
		z, x, y, ok := tileOf(address)
		if !ok {
			return nil, errors.New("not a tile address")
		}
		body, held := assets.Tile(z, x, y)
		if !held {
			return nil, errors.New("no such tile")
		}
		return body, nil
	})
}

// useTransport hands the map a transport of the host's own.
func useTransport(t *testing.T, m *tuimaps.Map, tr http.RoundTripper) {
	t.Helper()
	if err := m.SetFetchOptions(tuimaps.FetchOptions{Transport: tr}); err != nil {
		t.Fatal(err)
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
// the host's own transport, and nothing else.
func TestSourceReachesNothingUntilNamed(t *testing.T) {
	m, err := tuimaps.New(tuimaps.WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	asked := 0
	useTransport(t, m, served(t, &asked))
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
	useTransport(t, m, served(t, &asked))
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

// recording is a transport that answers with the embedded tiles and keeps
// every user-agent it was sent.
type recording struct {
	agents []string
	tiles  http.RoundTripper
}

func (r *recording) RoundTrip(req *http.Request) (*http.Response, error) {
	r.agents = append(r.agents, req.Header.Get("User-Agent"))
	return r.tiles.RoundTrip(req)
}

// TestFetchOptionsTakeEffectAtOnce is L8.1 (L-7.1, L-7.2, L-7.4): a
// transport set after the source is the one the next fetch goes through, and
// the host's name reaches the user-agent beside the library's.
func TestFetchOptionsTakeEffectAtOnce(t *testing.T) {
	m, err := tuimaps.New(tuimaps.WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	asked := 0
	first, second := &recording{tiles: served(t, &asked)}, &recording{tiles: served(t, &asked)}
	must(t, m.SetFetchOptions(tuimaps.FetchOptions{Transport: first}))
	must(t, m.Source("https://tiles.example.test/"))
	must(t, m.SetFetchOptions(tuimaps.FetchOptions{Transport: second, UserAgent: "watchpost"}))
	if _, err := m.Render(tuimaps.Size{Cols: 80, Rows: 24}, noon); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(first.agents) != 0 || len(second.agents) == 0 {
		t.Fatalf("the first transport was asked %d times and the second %d; want the second only", len(first.agents), len(second.agents))
	}
	for _, a := range second.agents {
		if !strings.HasPrefix(a, "go-tuimaps/") || !strings.HasSuffix(a, "(watchpost)") {
			t.Errorf("user-agent %q; want the library's name and the host's", a)
		}
	}
	if use := m.CacheUse(); use.Tiles.Held == 0 {
		t.Error("the second transport's tiles did not arrive")
	}
	fresh, err := tuimaps.New()
	if err != nil {
		t.Fatal(err)
	}
	defer fresh.Close()
	if err := fresh.SetFetchOptions(tuimaps.FetchOptions{UserAgent: "two words"}); err == nil {
		t.Error("a user-agent that would break the header was accepted, with no source named yet")
	}
}
