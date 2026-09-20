// Package tiles is the tile pipeline: where a tile can come from, what is
// kept of it and for how long, what is drawn while it is missing, and when
// a tile that failed may be tried again. It starts nothing and owns no
// clock: every slow step is a job a host's Work call runs, and every time is
// one the host passes in.
package tiles

import (
	"sync"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// DefaultCacheBytes is the memory cache's default cap (D-85).
const DefaultCacheBytes = 500_000

// Key names a decoded tile: where it came from, the one label language it
// was decoded for, and which tile it is. Style is no part of it, because a
// tile is decoded without reference to style (FR-31, D-82).
type Key struct {
	Source   string
	Language string
	Tile     scene.TileID
}

// Use is what a cache reports of itself (D-90): the bytes live views need,
// the bytes held, and the cap.
type Use struct {
	Need, Held, Cap int64
}

type item struct {
	tile  *scene.Tile
	bytes int64
	read  uint64 // the cache's count when the tile was last put or read
}

// Cache is the memory cache of decoded tiles. Need comes first and the cap
// second (D-90): a tile a live view draws is never evicted, spares are kept
// only in the room left under the cap, and need alone over the cap is
// reported, not enforced. It has a lock of its own and is safe beside
// everything.
type Cache struct {
	mu      sync.Mutex
	limit   int64
	held    int64
	reads   uint64
	items   map[Key]*item
	views   map[*PinSet][][]Key
	over    bool // need alone is over the cap
	warning bool // and that has not been told yet
}

// PinSet is one live view's say in what the cache must keep. For each tile
// the view wants it publishes a chain - the tile, then what may stand in for
// it, nearest first - and the first of the chain on hand is need: it is what
// would be drawn. A map registers one, publishes at every owner call that
// changes its view or size, and withdraws it at Close.
type PinSet struct {
	cache *Cache
}

// NewCache makes a cache with a cap in bytes, fixed for its life.
func NewCache(capBytes int64) (*Cache, error) {
	if capBytes <= 0 {
		return nil, fault.Make(fault.Internal,
			textsafe.Const("a tile cache was given no room"),
			textsafe.Const("its cap must be more than zero bytes"),
			textsafe.Const("give a cap in bytes, or none for the default"))
	}
	return &Cache{limit: capBytes, items: map[Key]*item{}, views: map[*PinSet][][]Key{}}, nil
}

// Put stores a decoded tile, then drops spares until the cache is under its
// cap or only need is left.
func (c *Cache) Put(k Key, tile *scene.Tile) {
	if c == nil || tile == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if old, ok := c.items[k]; ok {
		c.held -= old.bytes
	}
	c.reads++
	it := &item{tile: tile, bytes: int64(tile.Bytes()), read: c.reads}
	c.items[k] = it
	c.held += it.bytes
	c.settleLocked()
}

// Get returns a tile on hand and notes that it was read.
func (c *Cache) Get(k Key) (*scene.Tile, bool) {
	if c == nil {
		return nil, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	it, ok := c.items[k]
	if !ok {
		return nil, false
	}
	c.reads++
	it.read = c.reads
	return it.tile, true
}

// Has reports whether a tile is on hand, without counting as a read.
func (c *Cache) Has(k Key) bool {
	if c == nil {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.items[k]
	return ok
}

// Lineage is a tile's key and the keys of its ancestors, nearest first: the
// order in which they are looked for when the tile is to be drawn.
func Lineage(k Key) []Key {
	if k.Tile.Z > scene.MaxTileZoom {
		return nil
	}
	chain := make([]Key, 0, int(k.Tile.Z)+1)
	chain = append(chain, k)
	for range k.Tile.Z {
		parent, err := k.Tile.Parent()
		if err != nil {
			return chain
		}
		k.Tile = parent
		chain = append(chain, k)
	}
	return chain
}

// First returns the first tile of a chain that is on hand, and notes that it
// was read. With a chain that starts at the wanted tile and runs through
// its ancestors, anything but the first key is a stand-in (D-30).
func (c *Cache) First(chain []Key) (*scene.Tile, Key, bool) {
	if c == nil || len(chain) == 0 {
		return nil, Key{}, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	k, ok := c.firstLocked(chain)
	if !ok {
		return nil, Key{}, false
	}
	c.reads++
	c.items[k].read = c.reads
	return c.items[k].tile, k, true
}

func (c *Cache) firstLocked(chain []Key) (Key, bool) {
	if len(chain) == 0 {
		return Key{}, false
	}
	for _, k := range chain {
		if _, ok := c.items[k]; ok {
			return k, true
		}
	}
	return Key{}, false
}

// Register adds a live view. It needs nothing until it publishes.
func (c *Cache) Register() *PinSet {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	p := &PinSet{cache: c}
	c.views[p] = nil
	return p
}

// Publish replaces what the view wants: one chain for each tile. A tile that
// has not arrived, and has nothing on hand to stand in for it, counts for
// nothing until it arrives.
func (p *PinSet) Publish(wanted [][]Key) {
	if p == nil || p.cache == nil {
		return
	}
	c := p.cache
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, live := c.views[p]; !live {
		return
	}
	c.views[p] = append([][]Key(nil), wanted...)
	c.settleLocked()
}

// Withdraw ends the view: what only it needed becomes spare.
func (p *PinSet) Withdraw() {
	if p == nil || p.cache == nil {
		return
	}
	c := p.cache
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.views, p)
	c.settleLocked()
}

// neededLocked is the need: the first tile on hand of every chain.
func (c *Cache) neededLocked() (map[Key]bool, int64) {
	needed := map[Key]bool{}
	total := int64(0)
	for _, wanted := range c.views {
		for _, chain := range wanted {
			k, found := c.firstLocked(chain)
			if !found {
				continue
			}
			if !needed[k] {
				needed[k] = true
				total += c.items[k].bytes
			}
		}
	}
	return needed, total
}

// settleLocked drops the spare read longest ago until the cache is under its
// cap or holds need alone, and notes when need alone went over the cap.
func (c *Cache) settleLocked() {
	needed, need := c.neededLocked()
	for range len(c.items) {
		if c.held <= c.limit {
			break
		}
		var oldest Key
		found := false
		for k, it := range c.items {
			if !needed[k] && (!found || it.read < c.items[oldest].read) {
				oldest, found = k, true
			}
		}
		if !found {
			break
		}
		c.held -= c.items[oldest].bytes
		delete(c.items, oldest)
	}
	if need > c.limit && !c.over {
		c.warning = true
	}
	c.over = need > c.limit
}

// Use reports the bytes live views need, the bytes held, and the cap.
func (c *Cache) Use() Use {
	if c == nil {
		return Use{}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_, need := c.neededLocked()
	return Use{Need: need, Held: c.held, Cap: c.limit}
}

// TakeWarning returns the cache-under-need warning once for each stretch
// that need alone spends over the cap.
func (c *Cache) TakeWarning() (fault.Warning, bool) {
	if c == nil {
		return fault.Warning{}, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.warning {
		return fault.Warning{}, false
	}
	c.warning = false
	return fault.Warning{Kind: fault.CacheUnderNeed, Subject: textsafe.Const("the tile cache"), Count: 1}, true
}
