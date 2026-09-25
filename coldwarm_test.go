package tuimaps_test

import (
	"context"
	"crypto/tls"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

// The link NFR-5 is stated against: 150 ms of latency, paid on the
// connection as well as on each round of requests, and 8 Mbit/s in
// aggregate. The pump is two Work calls wide (D-84), so two requests
// travel together.
func statedLink() testkit.Link {
	return testkit.Link{
		Setup:         300 * time.Millisecond, // the connection: two 150 ms trips
		RoundTrip:     300 * time.Millisecond, // a round of requests: out and back
		BitsPerSecond: 8_000_000,
		Width:         2,
	}
}

// hostTick is the simulated interval at which a host draws again, which is
// what the clock is stopped on (D-30).
const hostTick = 300 * time.Millisecond

// The targets, from NFR-5.
const (
	coldTarget = 3 * time.Second
	warmTarget = 1 * time.Second
)

// fixtureTiles serves the pinned fixture's own zoom-5 and zoom-6 tiles, and
// a TileJSON that points at them: the 1.2 MB NFR-5 names.
func fixtureTiles(t *testing.T, served *int) http.Handler {
	t.Helper()
	root, err := testkit.FixtureRoot()
	if err != nil {
		t.Fatal(err)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		z, x, y, ok := tileOf(r.URL.Path)
		if !ok {
			http.Error(w, "not a tile address: "+r.URL.Path, http.StatusNotFound)
			return
		}
		last := ""
		for _, dir := range []string{"tiles-midwest-z5", "tiles-gulf-z6"} {
			name := filepath.ToSlash(filepath.Join(dir,
				itoa(int(z))+"-"+itoa(int(x))+"-"+itoa(int(y))+".pbf"))
			body, err := testkit.LoadFixture(root, name)
			if err != nil {
				last = name
				continue
			}
			*served++
			w.Header().Set("Content-Type", "application/x-protobuf")
			w.Write(body)
			return
		}
		http.Error(w, "no such tile: "+last, http.StatusNotFound)
	})
}

// TestColdAndWarm is plan task 14.9 and metric M2 (NFR-5): **how long a
// host waits for its map.** It is measured on a virtual clock - the link
// charges for what it serves and nothing sleeps - so the figure is the same
// on every machine and in every run. A real-time run is recorded by the
// benchmark and gates nothing.
func TestColdAndWarm(t *testing.T) {
	served := 0
	link := statedLink()
	server, err := testkit.NewShapedServer(fixtureTiles(t, &served), link)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Shutdown()

	dir := t.TempDir()
	cold := timed(t, server, dir, link)
	if served == 0 {
		t.Fatal("nothing was fetched; the map drew from somewhere else")
	}
	if cold.elapsed > coldTarget {
		t.Errorf("cold: the map took %v to full detail; NFR-5 says 3 s", cold.elapsed)
	}
	if !cold.firstFrameDrawn {
		t.Error("cold: the first render was blank; FR-23's floor is a frame that is never blank")
	}
	t.Logf("cold: %v to full detail over the stated link, %d tiles, %d requests, %d bytes",
		cold.elapsed.Round(time.Millisecond), served, cold.requests, cold.bytes)

	// Warm: the same source, the same cache on disk, nothing to fetch.
	before := served
	warm := timed(t, server, dir, link)
	if warm.elapsed > warmTarget {
		t.Errorf("warm: the map took %v to full detail; NFR-5 says 1 s", warm.elapsed)
	}
	if served != before {
		t.Errorf("warm: %d tiles were fetched again; the cache on disk held them", served-before)
	}
	t.Logf("warm: %v to full detail, nothing fetched", warm.elapsed.Round(time.Millisecond))
}

// run is what one start cost.
type run struct {
	elapsed         time.Duration
	requests        int
	bytes           int64
	firstFrameDrawn bool
}

// timed starts a map against the shaped server and stops the clock at the
// first frame drawn to full detail, charging a host tick for each round of
// drawing as D-30 says.
func timed(t *testing.T, server *testkit.ShapedServer, cacheDir string, link testkit.Link) run {
	t.Helper()
	before := server.Stats()
	m, err := tuimaps.New(tuimaps.WithSize(149, 38))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if err := m.CacheRoot(cacheDir, 0); err != nil {
		t.Fatal(err)
	}
	if err := m.SetFetchOptions(tuimaps.FetchOptions{Transport: trusting(t, server), UserAgent: "cold-and-warm"}); err != nil {
		t.Fatal(err)
	}
	if err := m.Source(server.URL + "/"); err != nil {
		t.Fatal(err)
	}
	// The fixture's own view: 149 by 38 cells at zoom 6.4 over Tampa Bay,
	// which is the view every memory and timing figure is stated against
	// (memory-measurement.md).
	if err := m.Recentre(tuimaps.LonLat{Lon: -82.4, Lat: 27.6}); err != nil {
		t.Fatal(err)
	}
	if err := m.Zoom(6.4); err != nil {
		t.Fatal(err)
	}

	// The host's first render, before any work: FR-23's floor says it is
	// never blank, and it is what a person sees while the rest arrives.
	size := tuimaps.Size{Cols: 149, Rows: 38}
	first, err := m.Render(size, noon)
	if err != nil {
		t.Fatal(err)
	}
	out := run{firstFrameDrawn: strings.TrimSpace(plainText(strings.Join(first.Lines, "\n"))) != ""}

	// The host's pump, two Work calls wide, redrawing on its own tick until
	// the frame is as good as it gets.
	// The clock starts when the host hands in its data, and the first
	// render is that moment: a tick is what each *further* round of drawing
	// costs (D-30).
	ticks := 0
	for range 64 {
		did := false
		for range link.Width {
			// A tile the fixture does not carry answers 404, which the
			// library reports and carries on from: what is being measured
			// is how long the tiles that *are* there take to arrive.
			ran, err := m.Work(context.Background())
			if err != nil && !isKind(err, fault.FetchFailed) {
				t.Fatal(err)
			}
			did = did || ran
		}
		frame, err := m.Render(size, noon)
		if err != nil {
			t.Fatal(err)
		}
		ticks++
		if frame.Status == tuimaps.Complete {
			break
		}
		if !did && m.Pending() == 0 {
			t.Fatalf("the map stopped at %v with nothing left to do", frame.Status)
		}
	}
	after := server.Stats()
	out.requests = after.Requests - before.Requests
	out.bytes = after.Bytes - before.Bytes
	// What the link charged for what was fetched, plus a host tick for each
	// round of drawing after the first (D-30).
	out.elapsed = after.VirtualElapsed - before.VirtualElapsed + time.Duration(ticks)*hostTick
	return out
}

// trusting is a host transport that trusts the shaped server's own anchor:
// the server signs for itself, and the library's own transport trusts the
// system's roots and nothing else (NFR-11). It dials the loopback server,
// which the library's own transport would refuse as a private address.
func trusting(t *testing.T, server *testkit.ShapedServer) http.RoundTripper {
	t.Helper()
	return &http.Transport{TLSClientConfig: &tls.Config{RootCAs: server.RootCAs, MinVersion: tls.VersionTLS12}}
}
