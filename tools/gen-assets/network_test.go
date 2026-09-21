package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/x509"
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/archive"
	"github.com/branden-thompson/go-tuimaps/internal/fetch"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// planetFile lays the source's tiles out as a small archive: a header, one
// flat root directory, and the tile data. skip leaves one tile out.
func planetFile(t *testing.T, source Source, skip scene.TileID) []byte {
	t.Helper()
	type entry struct {
		id   uint64
		data []byte
	}
	var entries []entry
	for z := uint8(0); z <= maxZoom; z++ {
		for x := uint32(0); x < 1<<z; x++ {
			for y := uint32(0); y < 1<<z; y++ {
				tile := scene.TileID{Z: z, X: x, Y: y}
				if tile == skip {
					continue
				}
				id, err := archive.TileID(z, x, y)
				if err != nil {
					t.Fatal(err)
				}
				data, _, _ := source(context.Background(), tile)
				entries = append(entries, entry{id, data})
			}
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].id < entries[j].id })
	var tiles []byte
	dir := binary.AppendUvarint(nil, uint64(len(entries)))
	last := uint64(0)
	for _, e := range entries {
		dir = binary.AppendUvarint(dir, e.id-last)
		last = e.id
	}
	for range entries {
		dir = binary.AppendUvarint(dir, 1)
	}
	for _, e := range entries {
		dir = binary.AppendUvarint(dir, uint64(len(e.data)))
	}
	for _, e := range entries {
		dir = binary.AppendUvarint(dir, uint64(len(tiles))+1)
		tiles = append(tiles, e.data...)
	}
	var root bytes.Buffer
	zw := gzip.NewWriter(&root)
	zw.Write(dir)
	zw.Close()

	head := make([]byte, archive.HeaderLen)
	copy(head, "PMTiles")
	head[7] = 3
	put := func(at int, v int) { binary.LittleEndian.PutUint64(head[at:], uint64(v)) }
	put(8, archive.HeaderLen)
	put(16, root.Len())
	put(24, archive.HeaderLen+root.Len()) // no metadata
	put(40, archive.HeaderLen+root.Len()) // no leaves
	put(56, archive.HeaderLen+root.Len())
	put(64, len(tiles))
	head[97], head[98], head[99], head[101] = 2, 2, 1, maxZoom
	return append(append(head, root.Bytes()...), tiles...)
}

// TestGeneratesFromThePinnedArchive: the generator reads the planet file by
// range requests, and pins it by the name it is given and the length and
// entity tag the source reports (FR-28a).
func TestGeneratesFromThePinnedArchive(t *testing.T) {
	whole := planetFile(t, fixtureSource(t), scene.TileID{Z: 99})
	holed := planetFile(t, fixtureSource(t), scene.TileID{Z: 3, X: 5, Y: 2})
	var mode atomic.Value
	mode.Store("honest")
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", `"76dd-9168"`)
		switch mode.Load() {
		case "whole file":
			w.Write(whole)
		case "a tile missing":
			http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(holed))
		default:
			http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(whole))
		}
	}))
	t.Cleanup(srv.Close)
	pool := x509.NewCertPool()
	pool.AddCert(srv.Certificate())
	opts := fetch.Options{RootCAs: pool}
	address := srv.URL + "/areas/planet/test/tiles.pmtiles"

	out := t.TempDir()
	if err := fromNetwork(context.Background(), address, "planet/test", out, opts); err != nil {
		t.Fatal(err)
	}
	pin, err := os.ReadFile(filepath.Join(out, "PIN"))
	if err != nil {
		t.Fatal(err)
	}
	want := "archive planet/test\nlength " + strconv.Itoa(len(whole)) + "\nentity-tag \"76dd-9168\"\n"
	if string(pin) != want {
		t.Errorf("PIN is\n%s\nwant\n%s", pin, want)
	}
	if strings.Contains(string(pin), "127.0.0.1") {
		t.Error("the pin holds the address; it holds the archive's name only")
	}
	if err := Verify(out); err != nil {
		t.Error(err)
	}

	for _, m := range []string{"whole file", "a tile missing"} {
		mode.Store(m)
		if err := fromNetwork(context.Background(), address, "planet/test", t.TempDir(), opts); err == nil {
			t.Errorf("%s: no error", m)
		}
	}
}
