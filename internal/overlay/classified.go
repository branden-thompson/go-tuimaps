package overlay

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"
	"math"
	"sync"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// DefaultClassifiedBytes is how much of a shared set is kept for classified
// images: one byte a pixel, so this is about a megapixel between every map
// that shares it.
const DefaultClassifiedBytes = 1 << 20

// Classified is the pictures more than one map has already read. Reducing a
// picture to one class a pixel means decoding it into full colour first -
// about four bytes a pixel, held for as long as the reading takes - and
// three maps showing the same radar frame were each paying it (D-116).
//
// It is keyed by what the reading depends on and nothing else: the picture
// itself, the table it is matched against, and how closely. Two maps that
// hand in the same picture with the same table get the same answer, which
// is the only reason it is safe to share.
type Classified struct {
	cap int64

	mu    sync.Mutex
	held  int64
	seen  map[[32]byte]classifiedEntry
	order [][32]byte // oldest first, for making room
}

type classifiedEntry struct {
	raster scene.Raster
	report Report
}

// NewClassified makes a shared set of read pictures. A cap of zero or less
// is the library's own.
func NewClassified(bytes int64) *Classified {
	if bytes <= 0 {
		bytes = DefaultClassifiedBytes
	}
	return &Classified{cap: bytes, seen: map[[32]byte]classifiedEntry{}}
}

// ImageKey is what a reading depends on: the picture's bytes, the table it
// is matched against, how closely, and how it is laid on the world. Every
// number goes in as its exact bits, so two tables never share a key (L-12.6).
func ImageKey(img *Image, kind Kind) ([32]byte, bool) {
	if img == nil || len(img.PNG) == 0 {
		return [32]byte{}, false
	}
	return keyOf(img, img.PNG, func(sum hash.Hash) { writeKind(sum, kind) }), true
}

// frameKey is one frame's key: what its reading depends on, and its valid
// time, so that a refresh keeps the reading of every frame it keeps (L-1.7).
func frameKey(img *Image, file []byte, valid time.Time, kind Kind) [32]byte {
	return keyOf(img, file, func(sum hash.Hash) {
		var number [8]byte
		binary.LittleEndian.PutUint64(number[:], uint64(valid.UnixNano()))
		sum.Write(number[:])
		writeKind(sum, kind)
	})
}

// keyOf hashes a picture's bytes, its bounds, its tolerance and its table,
// and then whatever more the caller adds: the type, and a frame's time.
func keyOf(img *Image, file []byte, more func(hash.Hash)) [32]byte {
	sum := sha256.New()
	sum.Write(file)
	var number [8]byte
	bits := func(v float64) {
		binary.LittleEndian.PutUint64(number[:], math.Float64bits(v))
		sum.Write(number[:])
	}
	for _, v := range []float64{img.West, img.South, img.East, img.North, img.Tolerance} {
		bits(v)
	}
	sum.Write([]byte{uint8(img.Projection), boolByte(img.Exact)})
	for _, e := range img.Table {
		bits(e.Value)
		sum.Write([]byte{e.Colour.R, e.Colour.G, e.Colour.B, boolByte(e.Missing)})
	}
	if more != nil {
		more(sum)
	}
	var key [32]byte
	copy(key[:], sum.Sum(nil))
	return key
}

// writeKind hashes a type: its preset and its breaks, each as exact bits.
func writeKind(sum hash.Hash, kind Kind) {
	var number [8]byte
	sum.Write([]byte{uint8(kind.Preset)})
	for _, b := range kind.Breaks {
		binary.LittleEndian.PutUint64(number[:], math.Float64bits(b))
		sum.Write(number[:])
	}
}

func boolByte(on bool) byte {
	if on {
		return 1
	}
	return 0
}

// Read is a picture already read, if this set has it.
func (c *Classified) Read(key [32]byte) (scene.Raster, Report, bool) {
	if c == nil {
		return scene.Raster{}, Report{}, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.seen[key]
	return entry.raster, entry.report, ok
}

// Keep puts a reading in the set, making room by dropping the oldest. A
// reading larger than the whole cap is not kept: the map that made it still
// has it, and nothing is evicted to make room for something that cannot fit.
func (c *Classified) Keep(key [32]byte, raster scene.Raster, report Report) {
	if c == nil {
		return
	}
	size := int64(len(raster.Classes))
	if size <= 0 || size > c.cap {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, already := c.seen[key]; already {
		return
	}
	// Dropping the oldest until there is room: bounded by what is held,
	// because each turn drops one and nothing is added while it runs.
	for range len(c.order) {
		if c.held+size <= c.cap {
			break
		}
		oldest := c.order[0]
		c.order = c.order[1:]
		c.held -= int64(len(c.seen[oldest].raster.Classes))
		delete(c.seen, oldest)
	}
	c.seen[key], c.held = classifiedEntry{raster: raster, report: report}, c.held+size
	c.order = append(c.order, key)
}

// Bytes is what the set holds and may hold.
func (c *Classified) Bytes() (held, limit int64) {
	if c == nil {
		return 0, 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.held, c.cap
}
