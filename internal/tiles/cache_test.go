package tiles

import (
	"os"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

func TestMain(m *testing.M) { os.Exit(testkit.Main(m)) }

// sized returns a tile whose kept form costs about n bytes.
func sized(n int) *scene.Tile {
	return &scene.Tile{Layers: []scene.Layer{{Name: "water", Extent: 4096, Coords: make([]int16, n/2)}}}
}

func key(z uint8, x, y uint32) Key {
	return Key{Source: "https://tiles.example", Language: "en", Tile: scene.TileID{Z: z, X: x, Y: y}}
}

func mustCache(t *testing.T, capBytes int64) *Cache {
	t.Helper()
	c, err := NewCache(capBytes)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// TestMemoryCacheByteCap is plan task 06.3: with no view registered, the
// cache holds what fits under its cap and drops what was read longest ago.
func TestMemoryCacheByteCap(t *testing.T) {
	if DefaultCacheBytes != 500_000 {
		t.Errorf("the default cap is %d, want 500,000 (D-85)", DefaultCacheBytes)
	}
	c := mustCache(t, DefaultCacheBytes)
	for x := uint32(0); x < 8; x++ {
		c.Put(key(6, x, 0), sized(100_000))
	}
	use := c.Bytes()
	if use.Held > use.Cap || use.Cap != DefaultCacheBytes || use.Need != 0 {
		t.Errorf("%+v: over the cap with nothing needed", use)
	}
	if _, ok := c.Get(key(6, 0, 0)); ok {
		t.Error("the tile put first is still held; the cap drops the least recently read")
	}
	if _, ok := c.Get(key(6, 7, 0)); !ok {
		t.Error("the tile put last was dropped")
	}

	// Reading a tile makes it the most recent.
	c = mustCache(t, 250_000)
	c.Put(key(6, 0, 0), sized(100_000))
	c.Put(key(6, 1, 0), sized(100_000))
	c.Get(key(6, 0, 0))
	c.Put(key(6, 2, 0), sized(100_000))
	if _, ok := c.Get(key(6, 0, 0)); !ok {
		t.Error("a tile just read was dropped before one read longer ago")
	}
	if _, ok := c.Get(key(6, 1, 0)); ok {
		t.Error("the least recently read tile was kept")
	}

	for _, bad := range []int64{0, -1} {
		if _, err := NewCache(bad); !isKind(err, fault.Internal) {
			t.Errorf("a cap of %d: %v", bad, err)
		}
	}
}

// TestPinnedNeverEvicted is D-90: what a live view draws is never evicted,
// even when two views together need more than the cap.
func TestPinnedNeverEvicted(t *testing.T) {
	c := mustCache(t, 300_000)
	east, west := c.Register(), c.Register()
	var eastKeys, westKeys []Key
	for x := uint32(0); x < 3; x++ {
		eastKeys = append(eastKeys, key(8, x, 0))
		westKeys = append(westKeys, key(8, x+100, 0))
	}
	east.Publish(lineages(eastKeys...))
	west.Publish(lineages(westKeys...))
	for i := range eastKeys {
		c.Put(eastKeys[i], sized(100_000))
		c.Put(westKeys[i], sized(100_000))
	}
	for _, k := range append(eastKeys, westKeys...) {
		if _, ok := c.Get(k); !ok {
			t.Errorf("%v is needed by a live view and was evicted; it would be fetched again without end", k.Tile)
		}
	}
	use := c.Bytes()
	if use.Need <= use.Cap || use.Held != use.Need {
		t.Errorf("%+v: over its cap the cache holds exactly the need", use)
	}

	// A spare has no room while need alone is over the cap.
	c.Put(key(8, 50, 0), sized(100_000))
	if _, ok := c.Get(key(8, 50, 0)); ok {
		t.Error("a spare was kept with need already over the cap")
	}

	// A withdrawn pin set frees its tiles.
	west.Withdraw()
	c.Put(key(8, 51, 0), sized(1_000))
	for _, k := range westKeys {
		if _, ok := c.Get(k); ok {
			t.Errorf("%v outlived the view that needed it, over the cap", k.Tile)
		}
	}
	for _, k := range eastKeys {
		if _, ok := c.Get(k); !ok {
			t.Errorf("%v was evicted while its view is live", k.Tile)
		}
	}
}

// TestStandInIsPinnedWhileItStandsIn: the ancestor drawn in place of a
// missing tile is need; once the tile arrives the ancestor is a spare.
func TestStandInIsPinnedWhileItStandsIn(t *testing.T) {
	c := mustCache(t, 150_000)
	view := c.Register()
	c.Put(key(3, 1, 1), sized(100_000)) // the ancestor of 6/8..15/8..15
	view.Publish(lineages(key(6, 9, 9)))
	c.Put(key(6, 40, 40), sized(100_000)) // a spare, pushing the cache over its cap
	if _, ok := c.Get(key(3, 1, 1)); !ok {
		t.Fatal("the stand-in for a missing tile was evicted")
	}
	c.Put(key(6, 9, 9), sized(100_000)) // the tile arrives; the ancestor is now spare and the oldest
	if _, ok := c.Get(key(6, 9, 9)); !ok {
		t.Error("the tile that arrived is not on hand")
	}
	if _, ok := c.Get(key(3, 1, 1)); ok {
		t.Error("the ancestor is still held over the cap, though nothing needs it now")
	}
}

// TestCacheUnderNeedWarnedOnce and TestCacheUseReportsNeedHeldCap are D-90's
// reporting half.
func TestCacheUnderNeedWarnedOnce(t *testing.T) {
	c := mustCache(t, 150_000)
	view := c.Register()
	view.Publish(lineages(key(8, 0, 0), key(8, 1, 0)))
	if _, ok := c.TakeWarning(); ok {
		t.Error("warned before need was over the cap")
	}
	c.Put(key(8, 0, 0), sized(100_000))
	c.Put(key(8, 1, 0), sized(100_000))
	w, ok := c.TakeWarning()
	if !ok || w.Kind != fault.CacheUnderNeed {
		t.Fatalf("%+v, %v; want one cache-under-need warning", w, ok)
	}
	c.Put(key(8, 2, 0), sized(100_000))
	view.Publish(lineages(key(8, 0, 0), key(8, 1, 0)))
	if _, ok := c.TakeWarning(); ok {
		t.Error("warned twice for one stretch over the cap")
	}
	// Back under, then over again: that is a new stretch.
	view.Publish(nil)
	c.Put(key(8, 3, 0), sized(1_000))
	view.Publish(lineages(key(9, 0, 0), key(9, 1, 0)))
	c.Put(key(9, 0, 0), sized(100_000))
	c.Put(key(9, 1, 0), sized(100_000))
	if _, ok := c.TakeWarning(); !ok {
		t.Error("a second stretch over the cap was not warned of")
	}
}

func TestCacheUseReportsNeedHeldCap(t *testing.T) {
	c := mustCache(t, 400_000)
	view := c.Register()
	a, b := sized(100_000), sized(50_000)
	c.Put(key(8, 0, 0), a)
	c.Put(key(8, 1, 0), b)
	view.Publish(lineages(key(8, 0, 0), key(8, 9, 9))) // one on hand, one not: it counts for nothing until it arrives
	use := c.Bytes()
	if use.Need != int64(a.Bytes()) || use.Held != int64(a.Bytes()+b.Bytes()) || use.Cap != 400_000 {
		t.Errorf("%+v; want need %d, held %d", use, a.Bytes(), a.Bytes()+b.Bytes())
	}
}

// TestCacheKeyIncludesLanguageNotStyle is plan task 06.4 (FR-31, D-82).
func TestCacheKeyIncludesLanguageNotStyle(t *testing.T) {
	c := mustCache(t, DefaultCacheBytes)
	english := key(5, 1, 1)
	c.Put(english, sized(1_000))
	german := english
	german.Language = "de"
	if _, ok := c.Get(german); ok {
		t.Error("a tile decoded for English was served for German")
	}
	other := english
	other.Source = "https://other.example"
	if _, ok := c.Get(other); ok {
		t.Error("a tile from one source was served for another")
	}
	if _, ok := c.Get(english); !ok {
		t.Error("the same source and language missed")
	}
}

// TestAncestorStandIn is plan task 06.5 (D-30): with only a zoom-3 tile on
// hand, a zoom-6 tile is drawn from it, and from a nearer ancestor once one
// is on hand.
func TestAncestorStandIn(t *testing.T) {
	c := mustCache(t, DefaultCacheBytes)
	world := sized(1_000)
	c.Put(key(3, 1, 1), world)
	got, at, ok := c.First(Lineage(key(6, 9, 9)))
	if !ok || got != world || at.Tile != (scene.TileID{Z: 3, X: 1, Y: 1}) {
		t.Errorf("%v, %v; want the zoom-3 ancestor", at.Tile, ok)
	}
	nearer := sized(1_000)
	c.Put(key(5, 4, 4), nearer)
	if got, at, _ := c.First(Lineage(key(6, 9, 9))); got != nearer || at.Tile.Z != 5 {
		t.Errorf("stand-in at zoom %d; the nearest ancestor on hand is drawn", at.Tile.Z)
	}
	if _, _, ok := c.First(Lineage(key(6, 40, 40))); ok {
		t.Error("a tile with no ancestor on hand got a stand-in")
	}
	if n := len(Lineage(key(6, 9, 9))); n != 7 {
		t.Errorf("a zoom-6 lineage has %d keys, want 7", n)
	}
	if Lineage(key(200, 0, 0)) != nil {
		t.Error("a tile deeper than any zoom has a lineage")
	}

	// A chain may cross sources: the embedded tile stands in for the network's.
	embedded := Key{Source: "embedded", Language: "en", Tile: scene.TileID{Z: 2, X: 1, Y: 1}}
	network := key(2, 1, 1)
	c = mustCache(t, DefaultCacheBytes)
	c.Put(embedded, world)
	if _, at, _ := c.First([]Key{network, embedded}); at != embedded {
		t.Errorf("%v; the embedded tile is drawn until the network's arrives", at)
	}
	c.Put(network, nearer)
	if _, at, _ := c.First([]Key{network, embedded}); at != network {
		t.Errorf("%v; the network's tile wins once it arrives (L-13)", at)
	}
}

func lineages(keys ...Key) [][]Key {
	var out [][]Key
	for _, k := range keys {
		out = append(out, Lineage(k))
	}
	return out
}

// TestEmptyFetchedKeepsTheEmbeddedTiles is v0.2.0 L9.4 (L-9.3): a purge drops
// every fetched tile, needed or not, and keeps the embedded ones, which
// record nothing of where anyone looked.
func TestEmptyFetchedKeepsTheEmbeddedTiles(t *testing.T) {
	c := mustCache(t, DefaultCacheBytes)
	embedded := Key{Source: embeddedIdentity, Language: "en", Tile: scene.TileID{Z: 1}}
	c.Put(embedded, sized(10_000))
	c.Put(key(6, 0, 0), sized(10_000))
	pins := c.Register()
	pins.Publish([][]Key{{key(6, 0, 0)}})
	c.EmptyFetched()
	if c.Has(key(6, 0, 0)) {
		t.Error("a fetched tile the view needs survived the purge")
	}
	if !c.Has(embedded) {
		t.Error("an embedded tile was purged")
	}
	if use := c.Bytes(); use.Held != int64(sized(10_000).Bytes()) {
		t.Errorf("held %d after the purge, want the embedded tile's %d", use.Held, sized(10_000).Bytes())
	}
}
