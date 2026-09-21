package archive

import (
	"context"
	"encoding/binary"
	"errors"
	"os"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

func TestMain(m *testing.M) { os.Exit(testkit.Main(m)) }

var errShort = errors.New("test reader: range outside the file")

func isKind(err error, k fault.Kind) bool {
	var f *fault.Error
	return errors.As(err, &f) && f.Kind() == k
}

var sampleTiles = []testTile{
	{0, 0, 0, []byte("world")}, {1, 0, 0, []byte("nw")}, {1, 1, 0, []byte("ne")}, {1, 0, 1, []byte("sw")}, {1, 1, 1, []byte("se")},
	{3, 2, 3, []byte("three-two-three")},
}

// TestHeader is plan task 04.1.
func TestHeader(t *testing.T) {
	good := buildArchive(sampleTiles, false, []byte(`{"name":"x"}`))
	h, err := ParseHeader(good[:HeaderLen], int64(len(good)))
	if err != nil {
		t.Fatal(err)
	}
	if h.RootOffset != HeaderLen || h.RootLength == 0 || h.TileDataLength == 0 || h.MaxZoom != 3 {
		t.Errorf("%+v", h)
	}
	damage := map[string]func([]byte) []byte{
		"bad magic":                      func(b []byte) []byte { b[0] = 'X'; return b },
		"version 2":                      func(b []byte) []byte { b[7] = 2; return b },
		"cut short":                      func(b []byte) []byte { return b[:100] },
		"root directory beyond the file": func(b []byte) []byte { binary.LittleEndian.PutUint64(b[8:], 1<<40); return b },
		"tile data beyond the file":      func(b []byte) []byte { binary.LittleEndian.PutUint64(b[64:], 1<<40); return b },
		"an offset that overflows": func(b []byte) []byte {
			binary.LittleEndian.PutUint64(b[56:], 1<<63)
			binary.LittleEndian.PutUint64(b[64:], 1<<63)
			return b
		},
		"a root directory over the limit":          func(b []byte) []byte { binary.LittleEndian.PutUint64(b[16:], MaxDirectoryBytes+1); return b },
		"an internal compression it does not read": func(b []byte) []byte { b[97] = 3; return b },
		"tiles that are not vector tiles":          func(b []byte) []byte { b[99] = 2; return b },
	}
	for name, change := range damage {
		b := change(append([]byte(nil), good[:HeaderLen]...))
		if _, err := ParseHeader(b, 1<<30); !isKind(err, fault.UnsupportedTile) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// TestTileIDHilbert is plan task 04.4: known values from the format's own
// description, and the round trip.
func TestTileIDHilbert(t *testing.T) {
	known := []struct {
		z    uint8
		x, y uint32
		id   uint64
	}{
		{0, 0, 0, 0}, {1, 0, 0, 1}, {1, 0, 1, 2}, {1, 1, 1, 3}, {1, 1, 0, 4}, {2, 0, 0, 5}, {12, 3423, 1763, 19078479},
	}
	for _, k := range known {
		id, err := TileID(k.z, k.x, k.y)
		if err != nil || id != k.id {
			t.Errorf("TileID(%d,%d,%d) = %d, %v; want %d", k.z, k.x, k.y, id, err, k.id)
		}
	}
	seen := map[uint64]bool{}
	for z := uint8(0); z <= 4; z++ {
		for x := uint32(0); x < 1<<z; x++ {
			for y := uint32(0); y < 1<<z; y++ {
				id, err := TileID(z, x, y)
				if err != nil || seen[id] {
					t.Fatalf("TileID(%d,%d,%d) = %d, %v; ids must be distinct", z, x, y, id, err)
				}
				seen[id] = true
			}
		}
	}
	if len(seen) != 1+4+16+64+256 {
		t.Errorf("%d ids for zooms 0 to 4", len(seen))
	}
	if _, err := TileID(3, 8, 0); err == nil {
		t.Error("a column outside the zoom must be an error")
	}
	if _, err := TileID(scene.MaxTileZoom+1, 0, 0); err == nil {
		t.Error("a zoom deeper than the library addresses must be an error")
	}
}

// TestFindTile is plan task 04.5, flat and through a leaf directory.
func TestFindTile(t *testing.T) {
	for _, leaf := range []bool{false, true} {
		r := &reader{data: buildArchive(sampleTiles, leaf, []byte(`{"vector_layers":[]}`))}
		a, err := Open(context.Background(), r.read, int64(len(r.data)))
		if err != nil {
			t.Fatalf("leaf=%v: %v", leaf, err)
		}
		for _, want := range sampleTiles {
			got, ok, err := a.Tile(context.Background(), scene.TileID{Z: want.z, X: want.x, Y: want.y})
			if err != nil || !ok || string(got) != string(want.data) {
				t.Errorf("leaf=%v %d/%d/%d: %q, %v, %v", leaf, want.z, want.x, want.y, got, ok, err)
			}
		}
		for _, missing := range []scene.TileID{{Z: 2, X: 1, Y: 1}, {Z: 3, X: 7, Y: 7}, {Z: 9, X: 1, Y: 1}} {
			if got, ok, err := a.Tile(context.Background(), missing); ok || err != nil || got != nil {
				t.Errorf("leaf=%v %v: %q, %v, %v; a tile the archive does not hold is simply absent", leaf, missing, got, ok, err)
			}
		}
	}
}

func TestRunLengthsServeManyTilesFromOneEntry(t *testing.T) {
	// One entry with a run of 4 covers the four tiles of zoom 1: the way an
	// archive stores a stretch of identical ocean.
	root := gzipBytes(encodeDir([]dirEntry{{id: 1, run: 4, length: 5, offset: 0}}))
	h := buildArchive(nil, false, nil)[:HeaderLen]
	binary.LittleEndian.PutUint64(h[16:], uint64(len(root)))
	binary.LittleEndian.PutUint64(h[24:], uint64(HeaderLen+len(root)))
	binary.LittleEndian.PutUint64(h[40:], uint64(HeaderLen+len(root)))
	binary.LittleEndian.PutUint64(h[56:], uint64(HeaderLen+len(root)))
	binary.LittleEndian.PutUint64(h[64:], 5)
	data := append(append(h, root...), []byte("ocean")...)
	r := &reader{data: data}
	a, err := Open(context.Background(), r.read, int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []scene.TileID{{Z: 1, X: 0, Y: 0}, {Z: 1, X: 1, Y: 1}, {Z: 1, X: 1, Y: 0}} {
		if got, ok, _ := a.Tile(context.Background(), id); !ok || string(got) != "ocean" {
			t.Errorf("%v: %q, %v", id, got, ok)
		}
	}
	if _, ok, _ := a.Tile(context.Background(), scene.TileID{Z: 2, X: 0, Y: 0}); ok {
		t.Error("the tile just past the run was served")
	}
}

// TestDirectoryLimits is plan task 04.2.
func TestDirectoryLimits(t *testing.T) {
	huge := make([]byte, MaxDirectoryBytes+1)
	if _, err := decodeDirectory(gzipBytes(huge), 1<<30, 1<<30); !isKind(err, fault.OverLimit) {
		t.Errorf("a directory that decompresses past the limit: %v", err)
	}
	tooMany := putUvarint(nil, MaxEntries+1)
	if _, err := decodeDirectory(gzipBytes(append(tooMany, make([]byte, 4*(MaxEntries+1))...)), 1<<30, 1<<30); !isKind(err, fault.OverLimit) {
		t.Errorf("more entries than the limit: %v", err)
	}
	lying := putUvarint(nil, 5000) // claims 5000 entries in a handful of bytes
	if _, err := decodeDirectory(gzipBytes(append(lying, 1, 2, 3)), 1<<30, 1<<30); !isKind(err, fault.UnsupportedTile) {
		t.Errorf("an entry count larger than a quarter of the bytes that follow: %v; it must be refused before anything is allocated", err)
	}
	outside := encodeDir([]dirEntry{{id: 0, run: 1, length: 100, offset: 50}})
	if _, err := decodeDirectory(gzipBytes(outside), 120, 1<<30); !isKind(err, fault.UnsupportedTile) {
		t.Errorf("an entry that reaches past the tile data: %v", err)
	}
	unordered := encodeDir([]dirEntry{{id: 5, run: 1, length: 1, offset: 0}, {id: 5, run: 1, length: 1, offset: 1}})
	if _, err := decodeDirectory(gzipBytes(unordered), 1<<30, 1<<30); !isKind(err, fault.UnsupportedTile) {
		t.Errorf("entries that are not in rising order: %v", err)
	}
	if _, err := decodeDirectory([]byte("not gzip at all"), 1<<30, 1<<30); !isKind(err, fault.UnsupportedTile) {
		t.Errorf("a directory that is not gzip: %v", err)
	}
}

// TestLeafDepthAndCycle is plan task 04.3.
func TestLeafDepthAndCycle(t *testing.T) {
	// A leaf directory whose only entry points back at itself.
	self := gzipBytes(encodeDir([]dirEntry{{id: 0, run: 0, length: 0, offset: 0}}))
	loop := encodeDir([]dirEntry{{id: 0, run: 0, length: uint64(len(self)), offset: 0}})
	loopGz := gzipBytes(loop)
	// Make the leaf point at itself with its true length.
	leaf := gzipBytes(encodeDir([]dirEntry{{id: 0, run: 0, length: uint64(len(loopGz)), offset: 0}}))
	for range 4 {
		leaf = gzipBytes(encodeDir([]dirEntry{{id: 0, run: 0, length: uint64(len(leaf)), offset: 0}}))
	}
	h := buildArchive(nil, false, nil)[:HeaderLen]
	root := gzipBytes(encodeDir([]dirEntry{{id: 0, run: 0, length: uint64(len(leaf)), offset: 0}}))
	binary.LittleEndian.PutUint64(h[16:], uint64(len(root)))
	binary.LittleEndian.PutUint64(h[24:], uint64(HeaderLen+len(root)))
	binary.LittleEndian.PutUint64(h[40:], uint64(HeaderLen+len(root)))
	binary.LittleEndian.PutUint64(h[48:], uint64(len(leaf)))
	binary.LittleEndian.PutUint64(h[56:], uint64(HeaderLen+len(root)+len(leaf)))
	data := append(append(h, root...), leaf...)
	r := &reader{data: data}
	a, err := Open(context.Background(), r.read, int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	_, ok, err := a.Tile(context.Background(), scene.TileID{})
	if ok || !isKind(err, fault.UnsupportedTile) {
		t.Errorf("leaves that lead only to more leaves: ok=%v, %v; the walk stops at depth %d", ok, err, MaxLeafDepth)
	}
	if len(r.asked) > MaxLeafDepth+3 {
		t.Errorf("%d reads; the walk did not stop", len(r.asked))
	}
}

// TestMetadataNeverParsed is plan task 04.15.
func TestMetadataNeverParsed(t *testing.T) {
	meta := []byte(`{"this":"is never read","bomb":"` + string(make([]byte, 1<<16)) + `"}`)
	r := &reader{data: buildArchive(sampleTiles, true, meta)}
	a, err := Open(context.Background(), r.read, int64(len(r.data)))
	if err != nil {
		t.Fatal(err)
	}
	for _, tile := range sampleTiles {
		a.Tile(context.Background(), scene.TileID{Z: tile.z, X: tile.x, Y: tile.y})
	}
	metaStart, metaEnd := int64(a.Header().MetadataOffset), int64(a.Header().MetadataOffset+a.Header().MetadataLength)
	for _, asked := range r.asked {
		if asked[0] < metaEnd && asked[0]+asked[1] > metaStart {
			t.Errorf("bytes %d to %d were read; the metadata block lies at %d to %d and is never read", asked[0], asked[0]+asked[1], metaStart, metaEnd)
		}
	}
}

func TestReaderFailuresAreTheLibrarysOwnErrors(t *testing.T) {
	r := &reader{data: buildArchive(sampleTiles, false, nil)}
	short := func(ctx context.Context, offset, length int64) ([]byte, error) {
		b, err := r.read(ctx, offset, length)
		if err != nil || len(b) < 2 {
			return b, err
		}
		return b[:len(b)-1], nil // one byte short of what was asked
	}
	if _, err := Open(context.Background(), short, int64(len(r.data))); !isKind(err, fault.FetchFailed) {
		t.Errorf("a reader that returns fewer bytes than asked: %v", err)
	}
	failing := func(context.Context, int64, int64) ([]byte, error) { return nil, errShort }
	_, err := Open(context.Background(), failing, 1<<20)
	if !isKind(err, fault.FetchFailed) || errors.Is(err, errShort) {
		t.Errorf("%v; the reader's own error is not passed on", err)
	}
	if _, err := Open(context.Background(), nil, 10); err == nil {
		t.Error("no reader must be an error")
	}
}
