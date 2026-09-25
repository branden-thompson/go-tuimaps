package tiles

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/mvt"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

const (
	// DefaultDiskBytes is the disk cache's default cap (FR-21a).
	DefaultDiskBytes = 256 << 20
	// layout names the directory layout, so that a later one can sit beside it.
	layout = "v1"
)

// Disk is the disk cache: the raw bytes of tiles that decoded, kept under a
// byte cap and, if the host sets one, a maximum age (D-18, FR-21a, L-9.1).
// It is off unless a host sets a root, because its file names are a record
// of where the user has looked (FR-21b).
//
// The layout is stable and documented: <root>/v1/<hash>/<z>/<x>-<y>.pbf,
// where <hash> is the first 32 hex digits of the SHA-256 of the source's
// identity. Nothing of the source appears in a path, and the only other
// parts are integers the library formats itself. The root is opened once,
// as a handle nothing can be reached outside of, and every access goes
// through it.
//
// A file's time is when its tile was fetched, set once when it is written
// and never touched by a read (L-9.4, D-56): the disk records which places
// were fetched, and when, but not when they were looked at.
type Disk struct {
	mu       sync.Mutex
	root     *os.Root
	released bool
	limit    int64
	held     int64
	serial   uint64
	maxAge   time.Duration   // zero: tiles do not age
	due      time.Time       // when the oldest file held reaches its age; zero with dueKnown: nothing is held
	dueKnown bool            // due was worked out by a walk, and kept up by each store since
	inView   map[string]bool // the files the view last planned for, which the cap never evicts
	gen      uint64          // raised by a purge and a release: a store begun before either writes nothing (L-9.3)
	readable bool            // others can read the root, and the host has not yet been told (L-9.6)
	failed   int             // writes that failed since the host was last told (L-9.5)
}

func cacheRefused(why textsafe.Text) error {
	return fault.Make(fault.CacheRefused, textsafe.Const("the disk cache was refused"), why,
		textsafe.Const("name a directory that only this user can write to, and a cap above zero; or set no root, and nothing is kept on disk"))
}

// OpenDisk opens, and if need be makes, the cache under dir. A root that
// others can write to is refused, and one they can read is warned of, where
// the platform can tell; on Windows neither check is performed.
func OpenDisk(dir string, capBytes int64) (*Disk, error) {
	if dir == "" {
		return nil, cacheRefused(textsafe.Const("no root was named"))
	}
	if capBytes <= 0 {
		return nil, cacheRefused(textsafe.Const("its cap is not above zero"))
	}
	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		return nil, cacheRefused(textsafe.Const("its root could not be made"))
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, cacheRefused(textsafe.Const("its root could not be opened"))
	}
	info, err := root.Stat(".") // the handle, not the path: what was checked is what is used
	if err != nil || !info.IsDir() {
		_ = root.Close() // nothing was written; the refusal is the error
		return nil, cacheRefused(textsafe.Const("its root is not a directory"))
	}
	checked := runtime.GOOS != "windows"
	if checked && info.Mode().Perm()&0o022 != 0 {
		_ = root.Close() // nothing was written; the refusal is the error
		return nil, cacheRefused(textsafe.Const("its root can be written to by other users"))
	}
	d := &Disk{root: root, limit: capBytes, readable: checked && info.Mode().Perm()&0o044 != 0}
	d.held = d.walk(nil)
	return d, nil
}

// Release lets go of the root. Nothing is written through the cache after
// it, a store begun before it included.
func (d *Disk) Release() {
	if d == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.released {
		return
	}
	d.released = true
	d.gen++
	_ = d.root.Close() // nothing is buffered; a close error changes nothing
}

// Generation is the cache's count of purges and releases. A job reads it
// when it starts and hands it to Store, which writes nothing if it moved.
func (d *Disk) Generation() uint64 {
	if d == nil {
		return 0
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.gen
}

// SetMaxAge sets how long a tile is kept from when it was fetched; zero
// keeps tiles until the cap needs their room.
func (d *Disk) SetMaxAge(age time.Duration) {
	if d == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.maxAge = max(age, 0)
	d.dueKnown = false // the next expiry walks
}

// InView names the tiles the view needs, which the cap never evicts. It
// does no input or output, so a plan may call it.
func (d *Disk) InView(identity string, tiles []scene.TileID) {
	if d == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.inView = make(map[string]bool, len(tiles))
	for _, tile := range tiles {
		if tile.Validate() == nil {
			d.inView[tilePath(identity, tile)] = true
		}
	}
}

// Held reports the bytes held and the cap.
func (d *Disk) Held() (held, limit int64) {
	if d == nil {
		return 0, 0
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.held, d.limit
}

// TakeWarnings returns, once, what the cache has to tell the host: that
// others can read its root, and how many writes failed since it last said.
func (d *Disk) TakeWarnings() []fault.Warning {
	if d == nil {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	var out []fault.Warning
	if d.readable {
		out = append(out, fault.Warning{Kind: fault.CacheRootReadable, Count: 1})
		d.readable = false
	}
	if d.failed > 0 {
		out = append(out, fault.Warning{Kind: fault.CacheWriteFailed, Count: d.failed})
		d.failed = 0
	}
	return out
}

type cached struct {
	rel     string
	size    int64
	fetched time.Time
}

// walk visits every regular file of the cache, and returns their total size.
// Links are not followed: the walk sees them as links and passes them by.
func (d *Disk) walk(visit func(cached)) int64 {
	total := int64(0)
	_ = fs.WalkDir(d.root.FS(), layout, func(rel string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil // a directory that cannot be read holds nothing the cache can use
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() {
			return nil
		}
		total += info.Size()
		if visit != nil {
			visit(cached{rel: rel, size: info.Size(), fetched: info.ModTime()})
		}
		return nil
	})
	return total
}

func sourceDir(identity string) string {
	sum := sha256.Sum256([]byte(identity))
	return layout + "/" + hex.EncodeToString(sum[:16])
}

func tilePath(identity string, tile scene.TileID) string {
	return sourceDir(identity) + "/" + strconv.Itoa(int(tile.Z)) + "/" +
		strconv.FormatUint(uint64(tile.X), 10) + "-" + strconv.FormatUint(uint64(tile.Y), 10) + ".pbf"
}

// expiredLocked says whether a tile fetched at fetched is past its age at
// now. A time after now counts as expired: a clock that moved back, or a
// file the library did not date, is not trusted to be fresh (L-9.4). A zero
// now is a clock not yet known, which judges nothing.
func (d *Disk) expiredLocked(fetched, now time.Time) bool {
	if now.IsZero() {
		return false
	}
	if fetched.After(now) {
		return true
	}
	return d.maxAge > 0 && now.Sub(fetched) >= d.maxAge
}

// plainLocked looks at a cached file without following anything: each
// directory on the way must be a directory, and the file a regular file. The
// handle stops a link from leaving the root, but it follows one that stays
// inside, so each entry is checked before it is opened.
func (d *Disk) plainLocked(rel string) (fs.FileInfo, bool) {
	if rel == "" || d.released {
		return nil, false
	}
	dir := path.Dir(rel)
	for range 4 {
		if dir == "." {
			break
		}
		info, err := d.root.Lstat(dir)
		if err != nil || !info.IsDir() {
			return nil, false
		}
		dir = path.Dir(dir)
	}
	info, err := d.root.Lstat(rel)
	if err != nil || !info.Mode().IsRegular() {
		return nil, false
	}
	return info, true
}

// removeLocked deletes whatever is at rel - a link is deleted, not followed -
// takes a regular file's size off the total, and says whether it went.
func (d *Disk) removeLocked(rel string) bool {
	if rel == "" || d.released {
		return false
	}
	info, err := d.root.Lstat(rel)
	if err != nil {
		return false
	}
	if d.root.Remove(rel) != nil {
		return false
	}
	if info.Mode().IsRegular() {
		d.held -= info.Size()
	}
	return true
}

// readTile reads one cached file through the largest tile's size: a file
// that grew after it was looked at is not read whole (L-12.6).
func readTile(r io.Reader) ([]byte, bool) {
	body, err := io.ReadAll(io.LimitReader(r, maxTileBytes+1))
	if err != nil || len(body) == 0 || len(body) > maxTileBytes {
		return nil, false
	}
	return body, true
}

// readLocked opens a file the cache checked, and reads it if what was opened
// is what was checked.
func (d *Disk) readLocked(rel string, info fs.FileInfo) ([]byte, bool) {
	f, err := d.root.Open(rel)
	if err != nil {
		return nil, false
	}
	opened, err := f.Stat()
	same := err == nil && os.SameFile(info, opened) // what was opened must be what was checked
	body, ok := readTile(f)
	_ = f.Close() // it was only read; a close error changes nothing
	return body, same && ok
}

// Load returns a cached tile's bytes, unless it is past its age at now, the
// host's wall clock; then its file is removed and it is fetched again. A
// read writes nothing.
func (d *Disk) Load(identity string, tile scene.TileID, now time.Time) ([]byte, bool) {
	if d == nil || tile.Validate() != nil {
		return nil, false
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	rel := tilePath(identity, tile)
	info, ok := d.plainLocked(rel)
	if !ok {
		return nil, false
	}
	if info.Size() == 0 || info.Size() > maxTileBytes || d.expiredLocked(info.ModTime(), now) {
		d.removeLocked(rel) // no tile is this size, or it has aged: it is not served
		return nil, false
	}
	return d.readLocked(rel, info)
}

// Store keeps a tile's raw bytes, dated now, its fetch time: a temporary
// file, then a rename. The pipeline calls it only after the bytes decoded
// completely (FR-22a), with the generation its job began in; if the cache
// was purged or released since, nothing is written. If the cache is then
// over its cap it is pruned to nine tenths of it, the oldest fetched first;
// this runs inside a job, never while drawing.
func (d *Disk) Store(identity string, tile scene.TileID, body []byte, now time.Time, gen uint64) error {
	if d == nil || tile.Validate() != nil {
		return cacheRefused(textsafe.Const("there is no cache, or the tile is not a tile"))
	}
	if len(body) == 0 || len(body) > maxTileBytes {
		return cacheRefused(textsafe.Const("the bytes to keep are empty or larger than any tile"))
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.released || gen != d.gen {
		return cacheRefused(textsafe.Const("the cache was purged or let go while the tile was fetched, so it is not kept"))
	}
	err := d.writeLocked(tilePath(identity, tile), body, now)
	if err != nil {
		d.failed++
		return err
	}
	if at := now.Add(d.maxAge); d.dueKnown && (now.IsZero() || d.due.IsZero() || at.Before(d.due)) {
		d.due = at
		d.dueKnown = !now.IsZero() // a file the system dated is found by the next walk
	}
	if d.held > d.limit {
		d.pruneLocked()
	}
	return nil
}

func (d *Disk) writeLocked(rel string, body []byte, now time.Time) error {
	err := d.root.MkdirAll(path.Dir(rel), 0o700)
	if err != nil {
		return cacheRefused(textsafe.Const("its directories could not be made; a cache that cannot be written still serves what it holds"))
	}
	d.serial++
	tmp := rel + ".tmp-" + strconv.FormatUint(d.serial, 10)
	f, err := d.root.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return cacheRefused(textsafe.Const("a file could not be made in it"))
	}
	_, err = f.Write(body)
	closeErr := f.Close()
	if err == nil && closeErr == nil {
		err = d.root.Chtimes(tmp, now, now) // the fetch time, by the host's clock; a zero time is left as the system dated it
	}
	if err != nil || closeErr != nil {
		_ = d.root.Remove(tmp) // best effort: a stray temporary file is pruned like any other
		return cacheRefused(textsafe.Const("a file could not be written in it"))
	}
	d.removeLocked(rel) // whatever was there, a link included, is replaced and never written through
	err = d.root.Rename(tmp, rel)
	if err != nil {
		_ = d.root.Remove(tmp) // best effort, as above
		return cacheRefused(textsafe.Const("a file could not be put in place in it"))
	}
	d.held += int64(len(body))
	return nil
}

// pruneLocked deletes the oldest fetched files until the cache is at nine
// tenths of its cap, passing by the files the view needs.
func (d *Disk) pruneLocked() {
	var all []cached
	d.held = d.walk(func(c cached) { all = append(all, c) })
	sort.Slice(all, func(i, j int) bool {
		if !all[i].fetched.Equal(all[j].fetched) {
			return all[i].fetched.Before(all[j].fetched)
		}
		return all[i].rel < all[j].rel
	})
	for _, c := range all {
		if d.held <= d.limit/10*9 {
			return
		}
		if !d.inView[c.rel] {
			d.removeLocked(c.rel)
		}
	}
}

// Expire removes every file past its age at now, and every file dated after
// it. It runs in every job, and walks the cache only once the oldest file
// held is due: a job otherwise pays one comparison. A clock set back after a
// walk is not looked for here; Load refuses what it dates after now.
func (d *Disk) Expire(now time.Time) {
	if d == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.released || d.maxAge <= 0 || now.IsZero() {
		return // with no age, Load still refuses a tile dated after now; with no clock, nothing is judged
	}
	if d.dueKnown && (d.due.IsZero() || now.Before(d.due)) {
		return
	}
	var all []cached
	d.walk(func(c cached) { all = append(all, c) })
	d.due, d.dueKnown = time.Time{}, true
	for _, c := range all {
		if d.expiredLocked(c.fetched, now) {
			d.removeLocked(c.rel)
			continue
		}
		if at := c.fetched.Add(d.maxAge); d.due.IsZero() || at.Before(d.due) {
			d.due = at
		}
	}
}

// Delete removes one cached tile: the pipeline calls it for a file that did
// not decode, which is then fetched again (FR-22a).
func (d *Disk) Delete(identity string, tile scene.TileID) {
	if d == nil || tile.Validate() != nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.removeLocked(tilePath(identity, tile))
}

// Emptied is what a purge removed, and what it could not.
type Emptied struct{ Removed, Failed int }

// Empty deletes every tile of every source, counting each removal that
// succeeded and each that failed, and raises the generation: a store begun
// before it writes nothing (L-9.3, L-9.5).
func (d *Disk) Empty() (Emptied, error) {
	if d == nil {
		return Emptied{}, cacheRefused(textsafe.Const("there is no cache"))
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.gen++
	if d.released {
		return Emptied{}, cacheRefused(textsafe.Const("the cache was let go"))
	}
	var all []cached
	d.walk(func(c cached) { all = append(all, c) })
	out := Emptied{}
	for _, c := range all {
		if d.removeLocked(c.rel) {
			out.Removed++
		} else {
			out.Failed++
		}
	}
	left := d.root.RemoveAll(layout) // the directories, and anything the walk could not see
	d.held = d.walk(nil)
	d.dueKnown = false
	if out.Failed > 0 || left != nil {
		return out, cacheRefused(textsafe.Const("some of its files could not be deleted"))
	}
	return out, nil
}

// ReadBack decodes every cached tile through the gate and deletes those that
// fail. It says how many it checked and how many it removed.
func (d *Disk) ReadBack(lim mvt.Limits) (checked, removed int, err error) {
	if d == nil {
		return 0, 0, cacheRefused(textsafe.Const("there is no cache"))
	}
	err = lim.Validate()
	if err != nil {
		return 0, 0, err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	var all []cached
	d.walk(func(c cached) { all = append(all, c) })
	want := mvt.Want{Layers: openMapTiles().Layers, Language: "en"}
	for _, c := range all {
		checked++
		if info, ok := d.plainLocked(c.rel); ok {
			if body, ok := d.readLocked(c.rel, info); ok {
				if _, decodeErr := mvt.Decode(body, want, lim); decodeErr == nil {
					continue
				}
			}
		}
		if d.removeLocked(c.rel) {
			removed++
		}
	}
	return checked, removed, nil
}
