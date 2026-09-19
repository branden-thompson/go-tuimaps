package archive

import (
	"context"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// FuzzHeader and FuzzDirectory are plan task 04.7.
func FuzzHeader(f *testing.F) {
	f.Add(buildArchive(sampleTiles, false, nil)[:HeaderLen], int64(1<<20))
	f.Add(buildArchive(sampleTiles, true, []byte("{}"))[:HeaderLen], int64(HeaderLen))
	f.Fuzz(func(t *testing.T, b []byte, size int64) {
		h, err := ParseHeader(b, size)
		if err != nil {
			return
		}
		for _, span := range [][2]uint64{{h.RootOffset, h.RootLength}, {h.MetadataOffset, h.MetadataLength}, {h.LeafOffset, h.LeafLength}, {h.TileDataOffset, h.TileDataLength}} {
			if span[0]+span[1] < span[0] || span[0]+span[1] > uint64(size) {
				t.Fatalf("an accepted header points outside the file: %+v against %d bytes", h, size)
			}
		}
	})
}

func FuzzDirectory(f *testing.F) {
	f.Add(gzipBytes(encodeDir([]dirEntry{{id: 0, run: 1, length: 5, offset: 0}, {id: 1, run: 4, length: 7, offset: 5}})), uint64(100), uint64(0))
	f.Add(gzipBytes(encodeDir([]dirEntry{{id: 3, run: 0, length: 9, offset: 0}})), uint64(0), uint64(9))
	f.Add([]byte{0x1f, 0x8b}, uint64(1), uint64(1))
	f.Fuzz(func(t *testing.T, raw []byte, tileData, leaves uint64) {
		entries, err := decodeDirectory(raw, tileData, leaves)
		if err != nil {
			return
		}
		if len(entries) > MaxEntries {
			t.Fatalf("%d entries accepted", len(entries))
		}
		for i, e := range entries {
			space := tileData
			if e.run == 0 {
				space = leaves
			}
			if e.length == 0 || e.offset+e.length < e.offset || e.offset+e.length > space {
				t.Fatalf("entry %d reaches outside its space: %+v against %d", i, e, space)
			}
			if i > 0 && e.id <= entries[i-1].id {
				t.Fatalf("entries out of order at %d", i)
			}
		}
	})
}

func FuzzOpenAndFind(f *testing.F) {
	f.Add(buildArchive(sampleTiles, false, nil), uint8(1), uint32(1), uint32(0))
	f.Add(buildArchive(sampleTiles, true, []byte("{}")), uint8(3), uint32(2), uint32(3))
	f.Fuzz(func(t *testing.T, data []byte, z uint8, x, y uint32) {
		r := &reader{data: data}
		a, err := Open(context.Background(), r.read, int64(len(data)))
		if err != nil {
			return
		}
		_, _, _ = a.Tile(context.Background(), scene.TileID{Z: z, X: x, Y: y})
		if len(r.asked) > MaxLeafDepth+4 {
			t.Fatalf("%d reads for one tile", len(r.asked))
		}
	})
}
