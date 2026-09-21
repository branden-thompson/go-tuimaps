package archive

import (
	"bytes"
	"context"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/fetch"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// TestRangeReaderInsistsOn206 is plan task 04.6: over a network the archive
// is read by range requests through the fetcher, and a source that answers a
// range with the whole file is refused, not read.
func TestRangeReaderInsistsOn206(t *testing.T) {
	file := buildArchive([]testTile{{z: 0, x: 0, y: 0, data: []byte("tile zero")}, {z: 1, x: 1, y: 0, data: []byte("tile east")}}, true, nil)
	var honest atomic.Bool
	honest.Store(true)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !honest.Load() {
			w.Write(file) // a 200 with the whole file, whatever range was asked for
			return
		}
		http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(file))
	}))
	t.Cleanup(srv.Close)
	pool := x509.NewCertPool()
	pool.AddCert(srv.Certificate())
	f, err := fetch.ForSource(srv.URL, fetch.Options{RootCAs: pool, Token: "archive-test"})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	info, err := f.Describe(ctx, srv.URL+"/planet", HeaderLen)
	if err != nil || info.Length != int64(len(file)) {
		t.Fatalf("describe: %+v, %v", info, err)
	}
	read, err := OverNetwork(f.Fetch, srv.URL+"/planet")
	if err != nil {
		t.Fatal(err)
	}
	a, err := Open(ctx, read, info.Length)
	if err != nil {
		t.Fatal(err)
	}
	got, ok, err := a.Tile(ctx, scene.TileID{Z: 1, X: 1, Y: 0})
	if err != nil || !ok || string(got) != "tile east" {
		t.Errorf("a tile by ranges: %q, %v, %v", got, ok, err)
	}

	honest.Store(false)
	if _, err := Open(ctx, read, info.Length); !isKind(err, fault.FetchFailed) {
		t.Errorf("a 200 to a range: %v; want the fetch-failed kind", err)
	}
	if _, _, err := a.Tile(ctx, scene.TileID{Z: 0}); !isKind(err, fault.FetchFailed) {
		t.Errorf("a 200 to a tile's range: %v; want the fetch-failed kind", err)
	}
	if _, err := OverNetwork(nil, srv.URL); err == nil {
		t.Error("no fetcher must be an error")
	}
	if _, err := OverNetwork(f.Fetch, ""); err == nil {
		t.Error("no address must be an error")
	}
}
