package archive

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"sort"
)

// A small archive writer for tests. The library never writes archives.

type testTile struct {
	z    uint8
	x, y uint32
	data []byte
}

func putUvarint(b []byte, v uint64) []byte {
	for v >= 0x80 {
		b = append(b, byte(v)|0x80)
		v >>= 7
	}
	return append(b, byte(v))
}

func gzipBytes(b []byte) []byte {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	w.Write(b)
	w.Close()
	return buf.Bytes()
}

type dirEntry struct {
	id, run, length, offset uint64
}

// encodeDir writes a directory in the archive's column layout.
func encodeDir(entries []dirEntry) []byte {
	b := putUvarint(nil, uint64(len(entries)))
	last := uint64(0)
	for _, e := range entries {
		b = putUvarint(b, e.id-last)
		last = e.id
	}
	for _, e := range entries {
		b = putUvarint(b, e.run)
	}
	for _, e := range entries {
		b = putUvarint(b, e.length)
	}
	for i, e := range entries {
		if i > 0 && e.offset == entries[i-1].offset+entries[i-1].length {
			b = putUvarint(b, 0)
		} else {
			b = putUvarint(b, e.offset+1)
		}
	}
	return b
}

// buildArchive lays out header, root directory, metadata, optional leaf
// directory and tile data. With leaf true the root points at one leaf that
// holds every tile.
func buildArchive(tiles []testTile, leaf bool, metadata []byte) []byte {
	sort.Slice(tiles, func(i, j int) bool {
		a, _ := TileID(tiles[i].z, tiles[i].x, tiles[i].y)
		b, _ := TileID(tiles[j].z, tiles[j].x, tiles[j].y)
		return a < b
	})
	var data []byte
	var entries []dirEntry
	for _, t := range tiles {
		id, _ := TileID(t.z, t.x, t.y)
		entries = append(entries, dirEntry{id: id, run: 1, length: uint64(len(t.data)), offset: uint64(len(data))})
		data = append(data, t.data...)
	}
	var root, leaves []byte
	if leaf {
		leaves = gzipBytes(encodeDir(entries))
		first := uint64(0)
		if len(entries) > 0 {
			first = entries[0].id
		}
		root = gzipBytes(encodeDir([]dirEntry{{id: first, run: 0, length: uint64(len(leaves)), offset: 0}}))
	} else {
		root = gzipBytes(encodeDir(entries))
	}
	h := make([]byte, HeaderLen)
	copy(h, "PMTiles")
	h[7] = 3
	pos := uint64(HeaderLen)
	put := func(at int, length int) {
		binary.LittleEndian.PutUint64(h[at:], pos)
		binary.LittleEndian.PutUint64(h[at+8:], uint64(length))
		pos += uint64(length)
	}
	put(8, len(root))
	put(24, len(metadata))
	put(40, len(leaves))
	put(56, len(data))
	h[97], h[98], h[99] = 2, 2, 1 // gzip inside, gzip tiles, vector tiles
	h[101] = 3
	out := append(h, root...)
	out = append(out, metadata...)
	out = append(out, leaves...)
	return append(out, data...)
}

// reader serves ranges of a byte slice, and counts what was asked for.
type reader struct {
	data  []byte
	asked [][2]int64
}

func (r *reader) read(_ context.Context, offset, length int64) ([]byte, error) {
	r.asked = append(r.asked, [2]int64{offset, length})
	if offset < 0 || length < 0 || offset+length > int64(len(r.data)) {
		return nil, errShort
	}
	return r.data[offset : offset+length], nil
}
