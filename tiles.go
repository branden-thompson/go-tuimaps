package tuimaps

import (
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/fetch"
	"github.com/branden-thompson/go-tuimaps/internal/mvt"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
	"github.com/branden-thompson/go-tuimaps/internal/tiles"
)

// FetchOptions is how a host shapes the library's own fetching (L-7.1, D-55):
// its Transport, its UserAgent, its Timeout, and the hosts it lets be
// fetched over plain http (AllowHTTP). The library keeps its client: it
// builds each request, confines it and every redirect to the source's own
// scheme, host and port, sets the headers, and reads the body through its
// limit, throwing away an answer that comes back after its request's end. A
// host Transport dials as it likes, and keeps the refusal of private
// addresses only if it dials through CheckedDialer; its time is bounded only
// while it honours its request's context (L-7.3).
type FetchOptions = fetch.HostOptions

// Dialer is how a transport connects: the shape of an http.Transport's
// DialContext.
type Dialer = fetch.Dialer

// CheckedDialer is the library's own dialer, which refuses a private or
// reserved address at the moment of connection. A host transport that dials
// through it keeps that refusal (L-10.3); one that dials some other way does
// not, and nothing the library does can check where a proxy goes beyond it.
func CheckedDialer() Dialer {
	return fetch.CheckedDialer()
}

// CacheUse is what one cache holds and is allowed to hold (D-90).
type CacheUse struct {
	Need  int64 // the bytes the views on hand need
	Held  int64 // the bytes held now
	Limit int64 // the cap
}

// Caches is what every cache of one map holds: the tiles in memory, the
// prepared shapes, and the tiles on disk.
type Caches struct {
	Tiles  CacheUse
	Shapes CacheUse
	Disk   CacheUse
}

// Source names where tiles come from, and is the only call that makes the
// library reach anything at all: until it is called nothing is fetched and
// no connection is made (D-65). The address is an http or https one, either
// a TileJSON document or a directory ending in a slash.
func (m *Map) Source(address string) (err error) {
	defer guard("Source", &err)
	m.plant("Source")

	if m == nil {
		return closed()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return closed()
	}
	if address == "" {
		m.remote, m.address = nil, ""
		return m.pipe.SetNetwork(nil)
	}
	return m.useSourceLocked(address)
}

// useSourceLocked names a source with the fetch options in effect.
func (m *Map) useSourceLocked(address string) error {
	own, err := fetch.ForSource(address, fetch.Options{Transport: m.fetchOpts.Transport, Token: m.fetchOpts.UserAgent,
		Timeout: m.fetchOpts.Timeout, AllowHTTP: m.fetchOpts.AllowHTTP})
	if err != nil {
		return err
	}
	remote, err := tiles.NewRemote(address, own.Fetch)
	if err != nil {
		return err
	}
	m.remote, m.address = remote, address
	m.changed++
	return m.pipe.SetNetwork(remote.Network())
}

// SetFetchOptions sets how the library fetches, and takes effect at once
// (L-7.4): with a source named, the next tile is fetched the new way. The
// zero FetchOptions is the library's own.
func (m *Map) SetFetchOptions(o FetchOptions) (err error) {
	defer guard("SetFetchOptions", &err)
	m.plant("SetFetchOptions")

	if m == nil {
		return closed()
	}
	if err := fetch.CheckToken(o.UserAgent); err != nil {
		return err
	}
	if o.Timeout < 0 {
		return fault.Make(fault.FetchRefused, textsafe.Const("the fetch options were refused"),
			textsafe.Const("the timeout is negative"), textsafe.Const("give a timeout of zero or more; zero is the library's own"))
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return closed()
	}
	was := m.fetchOpts
	m.fetchOpts = o
	if m.remote == nil {
		return nil
	}
	if err := m.useSourceLocked(m.address); err != nil {
		m.fetchOpts = was
		return err
	}
	return nil
}

// CacheRoot names a directory to keep tiles in between runs, or takes the
// disk cache away again with an empty path (FR-21b). The cap is in bytes; a
// cap of zero is the default. The root it replaces, or turns off, is let go:
// nothing is written to it after, a fetch already in flight included, and
// Purge does not reach it (L-9.3). The maximum age is kept at once, by the
// clock of the last frame (L-9.4).
//
// Over its cap the cache drops the tiles fetched longest ago, never one the
// view needs. A file's time is when its tile was fetched; reading it changes
// nothing on disk.
func (m *Map) CacheRoot(dir string, capBytes int64) (err error) {
	defer guard("CacheRoot", &err)
	m.plant("CacheRoot")

	if m == nil {
		return closed()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return closed()
	}
	if dir == "" {
		m.disk.Release()
		m.disk = nil
		return m.pipe.SetDisk(nil)
	}
	if capBytes <= 0 {
		capBytes = tiles.DefaultDiskBytes // a cap of zero is the library's own
	}
	disk, err := tiles.OpenDisk(dir, capBytes)
	if err != nil {
		return err
	}
	m.disk.Release()
	m.disk = disk
	disk.SetMaxAge(m.cacheMaxAge)
	disk.Expire(m.wallClock)
	return m.pipe.SetDisk(disk)
}

// SetCacheMaxAge sets how long a tile is kept on disk from when it was
// fetched (L-9.1, D-70). A tile at its age is not served but fetched again,
// and its file is removed; a file dated after the host's clock counts as
// aged. It is kept at once, by the clock of the last frame, and in every job
// after. Zero, the default, keeps tiles until the cap needs their room; a
// negative age is refused.
func (m *Map) SetCacheMaxAge(age time.Duration) (err error) {
	defer guard("SetCacheMaxAge", &err)
	m.plant("SetCacheMaxAge")

	if m == nil {
		return closed()
	}
	if age < 0 {
		return fault.Make(fault.CacheRefused, textsafe.Const("the cache's maximum age was refused"),
			textsafe.Const("it is negative"), textsafe.Const("give an age above zero, or zero to keep tiles until the cap needs the room"))
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return closed()
	}
	m.cacheMaxAge = age
	m.disk.SetMaxAge(age)
	m.disk.Expire(m.wallClock)
	return nil
}

// CacheUse is what each of the map's caches holds and may hold (D-90). A
// host that watches it knows whether its caps are the right size.
func (m *Map) CacheUse() Caches {
	defer m.guardQuiet("CacheUse")
	m.plant("CacheUse")

	if m == nil {
		return Caches{}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return Caches{}
	}
	tileUse := m.pipe.Holding()
	shapeUse := m.store.ShapeUse()
	out := Caches{
		Tiles:  CacheUse{Need: tileUse.Need, Held: tileUse.Held, Limit: tileUse.Cap},
		Shapes: CacheUse{Need: shapeUse.Need, Held: shapeUse.Held, Limit: shapeUse.Cap},
	}
	if m.disk != nil {
		held, limit := m.disk.Held()
		out.Disk = CacheUse{Held: held, Limit: limit}
	}
	return out
}

// PurgeReport is what Purge removed from disk, and what it could not.
// Removed counts only the files that went.
type PurgeReport struct{ Removed, Failed int }

// Purge empties everything the map holds of tiles: every source's tiles on
// disk, the fetched tiles in memory, and the decoded pictures kept only to be
// used again - a replaced loop's spare frames and a shared set's readings
// (L-9.3, FR-22a). A fetch already in flight writes nothing to disk after
// it. The tiles the view needs are fetched again. It is one of the cache's
// two maintenance calls, and it is the host's to make: the library never
// purges by itself. With no disk cache, memory is emptied and the error
// says there was no disk to purge. Purging deletes files; it is not secure
// erasure (L-9.6).
func (m *Map) Purge() (report PurgeReport, err error) {
	defer guard("Purge", &err)
	m.plant("Purge")

	if m == nil {
		return PurgeReport{}, closed()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return PurgeReport{}, closed()
	}
	m.pipe.Purge()
	m.store.Purge()
	if m.disk == nil {
		return PurgeReport{}, noCache()
	}
	emptied, err := m.disk.Empty()
	return PurgeReport(emptied), err
}

// Verify reads every tile the disk cache holds and removes the ones that no
// longer decode, saying how many it read and how many it removed (FR-22a).
func (m *Map) Verify() (checked, removed int, err error) {
	defer guard("Verify", &err)
	m.plant("Verify")

	if m == nil {
		return 0, 0, closed()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return 0, 0, closed()
	}
	if m.disk == nil {
		return 0, 0, noCache()
	}
	return m.disk.ReadBack(mvt.DefaultLimits())
}

// SourceCredit is what the source in use asks to be credited with, once its
// own document has been read. It is empty until then, and empty for the
// embedded tiles, whose credit is the basemap's own.
func (m *Map) SourceCredit() string {
	defer m.guardQuiet("SourceCredit")
	m.plant("SourceCredit")

	if m == nil {
		return ""
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut || m.remote == nil {
		return ""
	}
	return m.remote.Attribution().String()
}

func noCache() error {
	return fault.Make(fault.CacheRefused, textsafe.Const("the map has no disk cache"),
		textsafe.Const("nothing was kept on disk to work on"),
		textsafe.Const("name a directory with CacheRoot first, or leave the cache off"))
}
