package tuimaps

import (
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/overlay"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
	"github.com/branden-thompson/go-tuimaps/internal/tiles"
)

// Shared is a set of caches more than one map may draw from (FR-27). Two
// maps of the same data - a small one beside a large one, say - then hold
// one copy of each tile between them rather than one each.
//
// A shared set is made once and passed to each map as an option. It is safe
// to use from the maps that share it; it is not a call a host makes
// anything else of.
type Shared struct {
	tiles *tiles.Cache
	// pictures are the images the maps of this set have already read: a
	// picture is reduced to one class a pixel once between them, not once
	// each (D-116).
	pictures *overlay.Classified
}

// NewShared makes a set of caches for maps to share. The cap is the bytes
// of tiles held in memory for all of them together; zero is the library's
// own default.
func NewShared(tileBytes int64) (*Shared, error) {
	if tileBytes <= 0 {
		tileBytes = tiles.DefaultCacheBytes
	}
	cache, err := tiles.NewCache(tileBytes)
	if err != nil {
		return nil, err
	}
	return &Shared{tiles: cache, pictures: overlay.NewClassified(0)}, nil
}

// Use is what the shared caches need, hold and may hold. Need over cap is
// worth watching: it means the views together want more than they can keep,
// so tiles are fetched again as the views take turns (D-90).
func (s *Shared) Use() CacheUse {
	if s == nil || s.tiles == nil {
		return CacheUse{}
	}
	held := s.tiles.Bytes()
	return CacheUse{Need: held.Need, Held: held.Held, Limit: held.Cap}
}

// SharedCaches makes a map draw from a shared set rather than caches of its
// own (FR-27). Every map that shares a set must be created with it.
func SharedCaches(set *Shared) Option {
	return func(c *config) error {
		if c == nil || set == nil || set.tiles == nil {
			return badShared()
		}
		c.shared = set
		return nil
	}
}

func badShared() error {
	return fault.Make(fault.Internal, textsafe.Const("the shared caches were refused"),
		textsafe.Const("no set of caches was passed, or the set was never made"),
		textsafe.Const("make one with NewShared and pass it to every map that shares it"))
}
