// Package archive is a minimal reader for the single-file tile archive the
// embedded tiles are cut from (D-58): a header, a root directory, optional
// leaf directories, and tile data, read by ranges. It reads exactly what it
// needs to find a tile. The archive's metadata block is never read.
//
// Every byte is untrusted (NFR-10): offsets and lengths are checked against
// the file before anything is read, a directory is held to a size and an
// entry count before anything is allocated for it, and the walk through
// leaf directories is bounded, so a directory that points at itself ends in
// an error and not in a loop.
package archive

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"io"
	"sort"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

const (
	// HeaderLen is the length of the archive's header.
	HeaderLen = 127
	// MaxDirectoryBytes bounds one directory, compressed and decompressed.
	MaxDirectoryBytes = 1 << 20
	// MaxEntries bounds the entries of one directory.
	MaxEntries = 65536
	// MaxLeafDepth bounds the walk from the root through leaf directories.
	MaxLeafDepth = 3
	// maxTileBytes bounds one tile's stored bytes, as the decoder's body
	// limit does.
	maxTileBytes = 2 << 20

	compressionGzip = 2
	tileTypeVector  = 1
)

// RangeReader returns exactly length bytes of the archive starting at
// offset. Over a network it is a range request, which the fetcher holds to
// a 206 reply with exactly that range.
type RangeReader func(ctx context.Context, offset, length int64) ([]byte, error)

// Header is what the archive says about itself, as far as the reader needs.
type Header struct {
	RootOffset, RootLength         uint64
	MetadataOffset, MetadataLength uint64
	LeafOffset, LeafLength         uint64
	TileDataOffset, TileDataLength uint64
	MinZoom, MaxZoom               uint8
}

// entry is one directory entry. A run of zero marks a leaf directory; a run
// of n serves n consecutive tile ids from the same bytes.
type entry struct {
	id, run, length, offset uint64
}

// Archive is an open archive: its header and its root directory.
type Archive struct {
	header Header
	read   RangeReader
	root   []entry
}

// bad is the error for bytes that are not a well-formed archive of vector
// tiles.
func bad() error {
	return fault.Make(fault.UnsupportedTile,
		textsafe.Const("the tile archive could not be read"),
		textsafe.Const("it is not a well-formed version 3 archive of vector tiles with gzip directories"),
		textsafe.Const("check the archive; if it is another format or version, it is not supported"))
}

// tooLarge is the error for a directory or a tile over its limit.
func tooLarge() error {
	return fault.Make(fault.OverLimit,
		textsafe.Const("the tile archive was refused"),
		textsafe.Const("a directory or a tile in it is larger than a limit allows"),
		textsafe.Const("check the archive; the limits protect the host's memory"))
}

// unread is the error for a range that could not be read. The reader's own
// error is not passed on: over a network it may hold the archive's address.
func unread() error {
	return fault.Make(fault.FetchFailed,
		textsafe.Const("part of the tile archive could not be read"),
		textsafe.Const("the read failed, or returned a different number of bytes than was asked for"),
		textsafe.Const("check the archive and the network; the read is tried again later"))
}

// ParseHeader reads the header and checks every offset and length in it
// against the size of the file.
func ParseHeader(b []byte, fileSize int64) (Header, error) {
	if len(b) < HeaderLen || fileSize < HeaderLen {
		return Header{}, bad()
	}
	if string(b[:7]) != "PMTiles" || b[7] != 3 {
		return Header{}, bad()
	}
	if b[97] != compressionGzip || b[99] != tileTypeVector {
		return Header{}, bad()
	}
	u := func(at int) uint64 { return binary.LittleEndian.Uint64(b[at:]) }
	h := Header{
		RootOffset: u(8), RootLength: u(16),
		MetadataOffset: u(24), MetadataLength: u(32),
		LeafOffset: u(40), LeafLength: u(48),
		TileDataOffset: u(56), TileDataLength: u(64),
		MinZoom: b[100], MaxZoom: b[101],
	}
	if h.RootLength == 0 || h.RootLength > MaxDirectoryBytes {
		return Header{}, bad()
	}
	size := uint64(fileSize)
	for _, span := range [][2]uint64{{h.RootOffset, h.RootLength}, {h.MetadataOffset, h.MetadataLength}, {h.LeafOffset, h.LeafLength}, {h.TileDataOffset, h.TileDataLength}} {
		if span[0] > size || span[1] > size-span[0] { // written so that it cannot overflow
			return Header{}, bad()
		}
	}
	return h, nil
}

// Open reads the header and the root directory. fileSize is the archive's
// length, against which every offset is checked.
func Open(ctx context.Context, read RangeReader, fileSize int64) (*Archive, error) {
	if read == nil || ctx == nil {
		return nil, bad()
	}
	if fileSize < HeaderLen {
		return nil, bad()
	}
	a := &Archive{read: read}
	head, err := a.exactly(ctx, 0, HeaderLen)
	if err != nil {
		return nil, err
	}
	a.header, err = ParseHeader(head, fileSize)
	if err != nil {
		return nil, err
	}
	raw, err := a.exactly(ctx, a.header.RootOffset, a.header.RootLength)
	if err != nil {
		return nil, err
	}
	a.root, err = decodeDirectory(raw, a.header.TileDataLength, a.header.LeafLength)
	if err != nil {
		return nil, err
	}
	return a, nil
}

// Header returns the archive's header.
func (a *Archive) Header() Header {
	if a == nil {
		return Header{}
	}
	return a.header
}

// exactly reads a range and holds the reader to the length asked for.
func (a *Archive) exactly(ctx context.Context, offset, length uint64) ([]byte, error) {
	if length == 0 || length > maxTileBytes {
		return nil, tooLarge()
	}
	b, err := a.read(ctx, int64(offset), int64(length))
	if err != nil {
		if ctx.Err() != nil {
			return nil, fault.Make(fault.Cancelled, textsafe.Const("reading the tile archive was abandoned"), textsafe.Const("the work it was part of was cancelled or ran out of time"), textsafe.Const("nothing; it is asked for again if it is still wanted"))
		}
		return nil, unread()
	}
	if uint64(len(b)) != length {
		return nil, unread()
	}
	return b, nil
}

// decodeDirectory decompresses and decodes one directory. The decompressed
// size and the entry count are checked before anything is allocated for the
// entries; every entry is checked against the space it points into.
func decodeDirectory(raw []byte, tileDataLength, leafLength uint64) ([]entry, error) {
	if len(raw) == 0 || len(raw) > MaxDirectoryBytes {
		return nil, tooLarge()
	}
	zr, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, bad()
	}
	b, err := io.ReadAll(io.LimitReader(zr, MaxDirectoryBytes+1))
	if err != nil {
		return nil, bad()
	}
	if len(b) > MaxDirectoryBytes {
		return nil, tooLarge()
	}
	n, used := binary.Uvarint(b)
	if used <= 0 {
		return nil, bad()
	}
	if n > MaxEntries {
		return nil, tooLarge()
	}
	if n > uint64(len(b)-used)/4 { // four columns of at least a byte an entry
		return nil, bad()
	}
	entries := make([]entry, n)
	rest, err := readColumns(entries, b[used:])
	if err != nil || len(rest) != 0 {
		return nil, bad()
	}
	return entries, checkEntries(entries, tileDataLength, leafLength)
}

// readColumns fills the entries from the directory's four columns: tile ids
// as deltas, run lengths, lengths, and offsets, where zero means "straight
// after the entry before".
func readColumns(entries []entry, b []byte) ([]byte, error) {
	last := uint64(0)
	for col := range 4 {
		for i := range entries {
			v, used := binary.Uvarint(b)
			if used <= 0 {
				return nil, bad()
			}
			b = b[used:]
			switch col {
			case 0:
				last += v
				entries[i].id = last
			case 1:
				entries[i].run = v
			case 2:
				entries[i].length = v
			case 3:
				if v == 0 && i > 0 {
					entries[i].offset = entries[i-1].offset + entries[i-1].length
				} else if v == 0 {
					return nil, bad()
				} else {
					entries[i].offset = v - 1
				}
			}
		}
	}
	return b, nil
}

// checkEntries holds the entries to rising order and to the space each
// points into: tile data, or the leaf directories.
func checkEntries(entries []entry, tileDataLength, leafLength uint64) error {
	for i, e := range entries {
		if i > 0 && e.id <= entries[i-1].id {
			return bad()
		}
		space := tileDataLength
		if e.run == 0 {
			space = leafLength
		}
		if e.length == 0 || e.offset > space || e.length > space-e.offset {
			return bad()
		}
	}
	return nil
}

// find returns the entry that serves id: the last whose id is not greater,
// if id falls inside its run, or if it is a leaf.
func find(entries []entry, id uint64) (entry, bool) {
	lo := sort.Search(len(entries), func(i int) bool { return entries[i].id > id })
	if lo == 0 {
		return entry{}, false
	}
	e := entries[lo-1]
	if e.run == 0 || id-e.id < e.run {
		return e, true
	}
	return entry{}, false
}

// Tile returns a tile's stored bytes - still compressed, as the archive
// holds them - or false if the archive does not hold it.
func (a *Archive) Tile(ctx context.Context, tile scene.TileID) ([]byte, bool, error) {
	if a == nil || ctx == nil {
		return nil, false, bad()
	}
	id, err := TileID(tile.Z, tile.X, tile.Y)
	if err != nil {
		return nil, false, err
	}
	entries := a.root
	for range MaxLeafDepth + 1 {
		e, ok := find(entries, id)
		if !ok {
			return nil, false, nil
		}
		if e.run > 0 {
			data, err := a.exactly(ctx, a.header.TileDataOffset+e.offset, e.length)
			return data, err == nil, err
		}
		raw, err := a.exactly(ctx, a.header.LeafOffset+e.offset, e.length)
		if err != nil {
			return nil, false, err
		}
		entries, err = decodeDirectory(raw, a.header.TileDataLength, a.header.LeafLength)
		if err != nil {
			return nil, false, err
		}
	}
	return nil, false, bad() // leaves that lead only to more leaves
}

// TileID is the archive's number for a tile: the tiles of every shallower
// zoom, then the tile's place along its zoom's Hilbert curve.
func TileID(z uint8, x, y uint32) (uint64, error) {
	err := scene.TileID{Z: z, X: x, Y: y}.Validate()
	if err != nil {
		return 0, err
	}
	base := uint64(0)
	for i := range z {
		base += uint64(1) << (2 * i)
	}
	var d uint64
	tx, ty := uint64(x), uint64(y)
	for s := uint64(1) << z >> 1; s > 0; s >>= 1 {
		rx, ry := uint64(0), uint64(0)
		if tx&s > 0 {
			rx = 1
		}
		if ty&s > 0 {
			ry = 1
		}
		d += s * s * ((3 * rx) ^ ry)
		if ry == 0 {
			if rx == 1 {
				tx, ty = s-1-tx, s-1-ty
			}
			tx, ty = ty, tx
		}
	}
	return base + d, nil
}
