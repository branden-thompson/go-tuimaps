// Package scene holds the types every internal package shares: a tile's
// identity, the kinds of slow job, and the job itself. The packages that
// prepare data and the packages that draw it meet only here, so scene
// imports nothing but the standard library; that is what keeps the import
// graph free of cycles.
//
// The decoded tile and the prepared overlay join this package with the work
// packages that first produce them, their fields driven by those tests.
package scene

import (
	"context"
	"errors"
	"fmt"
)

// MaxTileZoom is the deepest tile zoom the library will address. Sources in
// the schema the library reads stop at 14 and the view stops at 18; 22 is
// the deepest a web-map tile source goes, and it keeps a column or row
// number inside 32 bits with room to spare.
const MaxTileZoom = 22

// ErrNoParent is returned by TileID.Parent for the world tile.
var ErrNoParent = errors.New("scene: the world tile has no parent")

// TileID names one tile of the web-mercator pyramid: zoom, column, row.
type TileID struct {
	Z uint8
	X uint32
	Y uint32
}

// Validate reports whether the tile exists: its zoom is one the library
// addresses, and its column and row lie inside that zoom's grid.
func (t TileID) Validate() error {
	if t.Z > MaxTileZoom {
		return fmt.Errorf("scene: tile zoom %d is deeper than %d", t.Z, MaxTileZoom)
	}
	side := uint32(1) << t.Z
	if t.X >= side {
		return fmt.Errorf("scene: tile column %d is outside zoom %d's %d columns", t.X, t.Z, side)
	}
	if t.Y >= side {
		return fmt.Errorf("scene: tile row %d is outside zoom %d's %d rows", t.Y, t.Z, side)
	}
	return nil
}

// Parent returns the tile one zoom out that contains this one: the tile
// that stands in for it while it is missing.
func (t TileID) Parent() (TileID, error) {
	if err := t.Validate(); err != nil {
		return TileID{}, err
	}
	if t.Z == 0 {
		return TileID{}, ErrNoParent
	}
	return TileID{Z: t.Z - 1, X: t.X / 2, Y: t.Y / 2}, nil
}

// JobKind says what a slow job does. The queue orders and counts by kind;
// it knows nothing else about the packages that supply jobs.
type JobKind uint8

// The kinds of slow job in this release. Zero is not a kind, so a job whose
// kind was never set is refused.
const (
	KindTile JobKind = iota + 1
	KindOverlayPrepare
	KindDescribe
)

// Validate reports whether k is one of the kinds above.
func (k JobKind) Validate() error {
	if k == 0 {
		return errors.New("scene: the job's kind was never set")
	}
	if k > KindDescribe {
		return fmt.Errorf("scene: %d is not a job kind", uint8(k))
	}
	return nil
}

// Job is one piece of slow work: fetching and decoding a tile, preparing an
// overlay, computing a description. The library never runs one by itself; a
// job runs inside a host's Work call, on the host's goroutine.
type Job interface {
	// Kind says what the job does.
	Kind() JobKind
	// Key identifies the work, so the same work is never queued twice.
	Key() string
	// Run does the slow part and returns when it is done or ctx has ended.
	// It holds no lock of the map's while it fetches, decodes or computes.
	Run(ctx context.Context) error
}
