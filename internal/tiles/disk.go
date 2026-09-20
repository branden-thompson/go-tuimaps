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
	// recencyStep is how stale a file's time may be before a read writes it again.
	recencyStep = time.Hour
	// maxNoted bounds the reads from memory waiting to be written down.
	maxNoted = 1024
)

// Disk is the disk cache: the raw bytes of tiles that decoded, kept without
// expiry under a byte cap (D-18, FR-21a). It is off unless a host sets a
// root, because its file names are a record of where the user has looked
// (FR-21b).
//
// The layout is stable and documented: <root>/v1/<hash>/<z>/<x>-<y>.pbf,
// where <hash> is the first 32 hex digits of the SHA-256 of the source's
// identity. Nothing of the source appears in a path, and the only other
// parts are integers the library formats itself. The root is opened once,
// as a handle nothing can be reached outside of, and every access goes
// through it.
type Disk struct {
	mu     sync.Mutex
	root   *os.Root
	limit  int64
	held   int64
	serial uint64
	noted  map[string]bool // tiles served from memory since the last job; their recency is owed
}

func cacheRefused(why textsafe.Text) error {
	return fault.Make(fault.CacheRefused, textsafe.Const("the disk cache was refused"), why,
		textsafe.Const("name a directory that only this user can write to, and a cap above zero; or set no root, and nothing is kept on disk"))
}

// OpenDisk opens, and if need be makes, the cache under dir. A root that
// others can write to is refused where the platform can tell; on Windows
// the check is not performed.
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
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o022 != 0 {
		_ = root.Close() // nothing was written; the refusal is the error
		return nil, cacheRefused(textsafe.Const("its root can be written to by other users"))
	}
	d := &Disk{root: root, limit: capBytes, noted: map[string]bool{}}
	d.held = d.walk(nil)
	return d, nil
}

// Release lets go of the root. The cache is not used after it.
func (d *Disk) Release() {
	if d == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	_ = d.root.Close() // nothing is buffered; a close error changes nothing
}

// Use reports the bytes held and the cap.
func (d *Disk) Held() (held, limit int64) {
	if d == nil {
		return 0, 0
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.held, d.limit
}

type cached struct {
	rel  string
	size int64
	read time.Time
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
			visit(cached{rel: rel, size: info.Size(), read: info.ModTime()})
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

// plainLocked looks at a cached file without following anything: each
// directory on the way must be a directory, and the file a regular file. The
// handle stops a link from leaving the root, but it follows one that stays
// inside, so each entry is checked before it is opened.
func (d *Disk) plainLocked(rel string) (fs.FileInfo, bool) {
	if rel == "" {
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
// and takes a regular file's size off the total.
func (d *Disk) removeLocked(rel string) {
	if rel == "" {
		return
	}
	info, err := d.root.Lstat(rel)
	if err != nil {
		return
	}
	if d.root.Remove(rel) == nil && info.Mode().IsRegular() {
		d.held -= info.Size()
	}
}

// Load returns a cached tile's bytes. now is the host's wall clock: a file
// whose time is more than an hour behind it has the time written again,
// which is what "least recently read" is judged by.
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
	if info.Size() == 0 || info.Size() > maxTileBytes {
		d.removeLocked(rel) // no tile is this size: it is not one of ours, or it rotted
		return nil, false
	}
	f, err := d.root.Open(rel)
	if err != nil {
		return nil, false
	}
	opened, err := f.Stat()
	same := err == nil && os.SameFile(info, opened) // what was opened must be what was checked
	body, err := io.ReadAll(io.LimitReader(f, maxTileBytes+1))
	_ = f.Close() // it was only read; a close error changes nothing
	if !same || err != nil || len(body) > maxTileBytes {
		return nil, false
	}
	if now.Sub(info.ModTime()) >= recencyStep {
		_ = d.root.Chtimes(rel, now, now) // a cache that cannot be written still serves
	}
	return body, true
}

// Store keeps a tile's raw bytes: a temporary file, then a rename. The
// pipeline calls it only after the bytes decoded completely (FR-22a). If the
// cache is then over its cap it is pruned to nine tenths of it, the least
// recently read first; this runs inside a job, never while drawing.
func (d *Disk) Store(identity string, tile scene.TileID, body []byte, now time.Time) error {
	if d == nil || tile.Validate() != nil {
		return cacheRefused(textsafe.Const("there is no cache, or the tile is not a tile"))
	}
	if len(body) == 0 || len(body) > maxTileBytes {
		return cacheRefused(textsafe.Const("the bytes to keep are empty or larger than any tile"))
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	rel := tilePath(identity, tile)
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
	if err != nil || closeErr != nil {
		_ = d.root.Remove(tmp) // best effort: a stray temporary file is pruned like any other
		return cacheRefused(textsafe.Const("a file could not be written in it"))
	}
	_ = d.root.Chtimes(tmp, now, now) // recency starts at the host's clock, not the machine's
	d.removeLocked(rel)               // whatever was there, a link included, is replaced and never written through
	err = d.root.Rename(tmp, rel)
	if err != nil {
		_ = d.root.Remove(tmp) // best effort, as above
		return cacheRefused(textsafe.Const("a file could not be put in place in it"))
	}
	d.held += int64(len(body))
	if d.held > d.limit {
		d.pruneLocked()
	}
	return nil
}

// pruneLocked deletes the least recently read files until the cache is at
// nine tenths of its cap.
func (d *Disk) pruneLocked() {
	var all []cached
	d.held = d.walk(func(c cached) { all = append(all, c) })
	sort.Slice(all, func(i, j int) bool {
		if !all[i].read.Equal(all[j].read) {
			return all[i].read.Before(all[j].read)
		}
		return all[i].rel < all[j].rel
	})
	for _, c := range all {
		if d.held <= d.limit/10*9 {
			return
		}
		d.removeLocked(c.rel)
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

// NoteRead records, in memory only, that a tile was served from the memory
// cache: it is still being read, and the next job writes that down. A plan
// does no input or output, so it cannot.
func (d *Disk) NoteRead(identity string, tile scene.TileID) {
	if d == nil || tile.Validate() != nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.noted) < maxNoted {
		d.noted[tilePath(identity, tile)] = true
	}
}

// Flush writes down the recency of every noted tile whose time is more than
// an hour behind now. It runs inside a job.
func (d *Disk) Flush(now time.Time) {
	if d == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	for rel := range d.noted {
		info, ok := d.plainLocked(rel)
		if ok && now.Sub(info.ModTime()) >= recencyStep {
			_ = d.root.Chtimes(rel, now, now) // a cache that cannot be written still serves
		}
	}
	d.noted = map[string]bool{}
}

// Purge deletes one source's tiles, or with an empty identity every tile.
func (d *Disk) Empty(identity string) error {
	if d == nil {
		return cacheRefused(textsafe.Const("there is no cache"))
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	target := layout
	if identity != "" {
		target = sourceDir(identity)
	}
	err := d.root.RemoveAll(target)
	d.held = d.walk(nil)
	if err != nil {
		return cacheRefused(textsafe.Const("some of its files could not be deleted"))
	}
	return nil
}

// Verify decodes every cached tile through the gate and deletes those that
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
		body, readErr := d.root.ReadFile(c.rel)
		if readErr == nil && c.size <= maxTileBytes {
			if _, decodeErr := mvt.Decode(body, want, lim); decodeErr == nil {
				continue
			}
		}
		d.removeLocked(c.rel)
		removed++
	}
	return checked, removed, nil
}
