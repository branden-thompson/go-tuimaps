package tiles

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/assets"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/mvt"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

const planet = "https://tiles.example/planet/"

func openDisk(t *testing.T, root string, capBytes int64) *Disk {
	t.Helper()
	d, err := OpenDisk(root, capBytes)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(d.Release)
	return d
}

func tileBytes(t *testing.T) []byte {
	t.Helper()
	body, ok := assets.Tile(2, 1, 1)
	if !ok {
		t.Fatal("no embedded tile")
	}
	return body
}

// files lists every regular file under root, relative to it.
func files(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.Mode().IsRegular() {
			rel, _ := filepath.Rel(root, path)
			out = append(out, filepath.ToSlash(rel))
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// TestDiskCacheOffByDefault is plan task 06.8 (FR-21b): with no root set,
// nothing is written anywhere.
func TestDiskCacheOffByDefault(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(dir, "cache"))
	net := &network{}
	p := pipeline(t, Options{Network: net.source()})
	runAll(t, p.Plan(t0, []scene.TileID{id(0, 0, 0)}))
	if got := files(t, dir); len(got) != 0 {
		t.Errorf("files written with no cache root set: %v", got)
	}
	if _, err := OpenDisk("", DefaultDiskBytes); !isKind(err, fault.CacheRefused) {
		t.Errorf("an empty root: %v", err)
	}
	if DefaultDiskBytes != 256<<20 {
		t.Errorf("the default cap is %d, want 256 MB", DefaultDiskBytes)
	}
}

// TestCachePathConfined is plan task 06.9, and TestParityP48_DiskCache the
// layout: a hash of the source identity and integers, nothing else.
func TestCachePathConfined(t *testing.T) {
	root := t.TempDir()
	d := openDisk(t, root, DefaultDiskBytes)
	hostile := []string{"../../../etc/passwd", "/etc/passwd", `C:\Windows\system32`, "a/../../b", strings.Repeat("x", 5000), "https://tiles.example/?key=SECRET", "nul\x00byte"}
	for _, identity := range hostile {
		if err := d.Store(identity, id(3, 1, 2), tileBytes(t), t0); err != nil {
			t.Errorf("%q: %v", identity, err)
		}
	}
	got := files(t, root)
	if len(got) != len(hostile) {
		t.Fatalf("%d files for %d sources: %v", len(got), len(hostile), got)
	}
	for _, rel := range got {
		parts := strings.Split(rel, "/")
		if len(parts) != 4 || parts[0] != "v1" || len(parts[1]) != 32 || strings.Trim(parts[1], "0123456789abcdef") != "" || parts[2] != "3" || parts[3] != "1-2.pbf" {
			t.Errorf("path %q is not v1/<hash>/<z>/<x>-<y>.pbf", rel)
		}
		if strings.Contains(rel, "SECRET") {
			t.Errorf("the path %q repeats the source", rel)
		}
	}
	if parent := files(t, filepath.Dir(root)); len(parent) != len(got) {
		t.Errorf("files outside the root: %v", parent)
	}
}

func TestParityP48_DiskCache(t *testing.T) {
	root := t.TempDir()
	d := openDisk(t, root, DefaultDiskBytes)
	body := tileBytes(t)
	if err := d.Store(planet, id(6, 16, 26), body, t0); err != nil {
		t.Fatal(err)
	}
	// Raw bytes, no expiry: a file a year old is served.
	got := files(t, root)
	old := t0.Add(-365 * 24 * time.Hour)
	if err := os.Chtimes(filepath.Join(root, got[0]), old, old); err != nil {
		t.Fatal(err)
	}
	read, ok := d.Load(planet, id(6, 16, 26), t0)
	if !ok || !bytes.Equal(read, body) {
		t.Error("a year-old file was not served as it was written")
	}
	if _, ok := d.Load("https://other.example/", id(6, 16, 26), t0); ok {
		t.Error("one source's tile was served for another")
	}
	if info, err := os.Stat(filepath.Join(root, got[0])); err != nil || (runtime.GOOS != "windows" && info.Mode().Perm() != 0o600) {
		t.Errorf("file mode %v, want private to the user", info.Mode())
	}
	if info, _ := os.Stat(filepath.Join(root, "v1")); runtime.GOOS != "windows" && info.Mode().Perm() != 0o700 {
		t.Errorf("directory mode %v, want private to the user", info.Mode())
	}
	for _, rel := range files(t, root) {
		if strings.Contains(rel, ".tmp") {
			t.Errorf("a temporary file was left: %s", rel)
		}
	}
}

// TestWriteOnlyAfterFullDecode and TestBadCachedFileDeleted are plan task
// 06.10 (FR-22a).
func TestWriteOnlyAfterFullDecode(t *testing.T) {
	root := t.TempDir()
	d := openDisk(t, root, DefaultDiskBytes)
	bad := pipeline(t, Options{Disk: d, Network: &Network{Identity: planet, MaxZoom: 14, Get: func(context.Context, scene.TileID) ([]byte, error) {
		return []byte("<html>busy</html>"), nil
	}}})
	runAll(t, bad.Plan(t0, []scene.TileID{id(0, 0, 0)}))
	if got := files(t, root); len(got) != 0 {
		t.Errorf("bytes that did not decode were cached: %v", got)
	}
	net := &network{}
	good := pipeline(t, Options{Disk: d, Network: &Network{Identity: planet, MaxZoom: 14, Get: net.get}})
	runAll(t, good.Plan(t0, []scene.TileID{id(0, 0, 0)}))
	if got := files(t, root); len(got) != 1 {
		t.Errorf("after a complete decode: %v", got)
	}

	// The disk is asked before the network (06.2): a second map, same source.
	again := pipeline(t, Options{Disk: d, Network: &Network{Identity: planet, MaxZoom: 14, Get: net.get}})
	runAll(t, again.Plan(t0, []scene.TileID{id(0, 0, 0)}))
	if len(net.asked) != 1 {
		t.Errorf("the network was asked %d times; the second map's tile was on disk", len(net.asked))
	}
	if again.State(id(0, 0, 0)) != OnHand {
		t.Error("the tile from disk is not on hand")
	}
}

func TestBadCachedFileDeleted(t *testing.T) {
	root := t.TempDir()
	d := openDisk(t, root, DefaultDiskBytes)
	if err := d.Store(planet, id(0, 0, 0), []byte("rotted on disk"), t0); err != nil {
		t.Fatal(err)
	}
	net := &network{}
	p := pipeline(t, Options{Disk: d, Network: &Network{Identity: planet, MaxZoom: 14, Get: net.get}})
	if failed := runAll(t, p.Plan(t0, []scene.TileID{id(0, 0, 0)})); failed != 0 {
		t.Error("a bad cached file failed the tile; it is deleted and fetched again")
	}
	if len(net.asked) != 1 {
		t.Errorf("the network was asked %d times", len(net.asked))
	}
	body, ok := d.Load(planet, id(0, 0, 0), t0)
	if !ok || string(body) == "rotted on disk" {
		t.Error("the bad file was not replaced by the fetched one")
	}
}

// TestRootWritableByOthersRefused is plan task 06.11.
func TestRootWritableByOthersRefused(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("NOT RUN: Windows has no permission bits to check; the root lies under the user's profile and the check is documented as not performed (FR-21b)")
	}
	root := t.TempDir()
	if err := os.Chmod(root, 0o777); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenDisk(root, DefaultDiskBytes); !isKind(err, fault.CacheRefused) {
		t.Errorf("a root anyone can write to: %v", err)
	}
	if err := os.Chmod(root, 0o750); err != nil {
		t.Fatal(err)
	}
	openDisk(t, root, DefaultDiskBytes)
	fresh := filepath.Join(t.TempDir(), "made", "by", "the", "library")
	openDisk(t, fresh, DefaultDiskBytes)
	if info, err := os.Stat(fresh); err != nil || info.Mode().Perm() != 0o700 {
		t.Errorf("a root the library made: %v, %v; want private to the user", info, err)
	}
	if _, err := OpenDisk(root, 0); !isKind(err, fault.CacheRefused) {
		t.Errorf("a cap of zero: %v", err)
	}
}

// TestEvictLeastRecentlyRead is plan task 06.12 (FR-21a).
func TestEvictLeastRecentlyRead(t *testing.T) {
	root := t.TempDir()
	body := tileBytes(t)
	size := int64(len(body))
	d := openDisk(t, root, 10*size)
	for x := uint32(0); x < 10; x++ {
		if err := d.Store(planet, id(8, x, 0), body, t0.Add(time.Duration(x)*time.Minute)); err != nil {
			t.Fatal(err)
		}
	}
	if held, _ := d.Held(); held != 10*size {
		t.Fatalf("held %d, want %d", held, 10*size)
	}
	// Reading the oldest, more than an hour on, makes it the most recent.
	later := t0.Add(2 * time.Hour)
	if _, ok := d.Load(planet, id(8, 0, 0), later); !ok {
		t.Fatal("miss")
	}
	// Within the hour the time is not written again.
	path := filepath.Join(root, files(t, root)[0])
	_ = path
	if err := d.Store(planet, id(8, 10, 0), body, later); err != nil {
		t.Fatal(err)
	}
	held, limit := d.Held()
	if held > limit*9/10 {
		t.Errorf("held %d after pruning, want at most 90%% of %d", held, limit)
	}
	if _, ok := d.Load(planet, id(8, 0, 0), later); !ok {
		t.Error("the tile just read was evicted")
	}
	if _, ok := d.Load(planet, id(8, 1, 0), later); ok {
		t.Error("the least recently read tile was kept")
	}
	if _, ok := d.Load(planet, id(8, 10, 0), later); !ok {
		t.Error("the tile just written was evicted")
	}

	// The total is seeded by one walk when the cache is opened.
	d.Release()
	reopened := openDisk(t, root, 10*size)
	if again, _ := reopened.Held(); again != held {
		t.Errorf("reopened: held %d, want %d", again, held)
	}
}

// TestRecencyWrittenAtMostHourly: the modification time is the recency, and
// a read within the hour does not write it again.
func TestRecencyWrittenAtMostHourly(t *testing.T) {
	root := t.TempDir()
	d := openDisk(t, root, DefaultDiskBytes)
	if err := d.Store(planet, id(1, 0, 0), tileBytes(t), t0); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, files(t, root)[0])
	mtime := func() time.Time {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		return info.ModTime()
	}
	d.Load(planet, id(1, 0, 0), t0.Add(59*time.Minute))
	if !mtime().Equal(t0) {
		t.Errorf("read within the hour moved the time to %v", mtime())
	}
	d.Load(planet, id(1, 0, 0), t0.Add(61*time.Minute))
	if !mtime().Equal(t0.Add(61 * time.Minute)) {
		t.Errorf("read after the hour left the time at %v", mtime())
	}
}

// TestPurgeAndVerify is plan task 06.13 (FR-22a).
func TestPurgeAndVerify(t *testing.T) {
	root := t.TempDir()
	d := openDisk(t, root, DefaultDiskBytes)
	other := "https://other.example/"
	for _, source := range []string{planet, other} {
		for x := uint32(0); x < 3; x++ {
			if err := d.Store(source, id(4, x, 0), tileBytes(t), t0); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := d.Store(planet, id(4, 9, 9), []byte("rotted"), t0); err != nil {
		t.Fatal(err)
	}
	checked, removed, err := d.ReadBack(mvt.DefaultLimits())
	if err != nil || checked != 7 || removed != 1 {
		t.Errorf("verify: %d checked, %d removed, %v; want 7 and 1", checked, removed, err)
	}
	if err := d.Empty(planet); err != nil {
		t.Fatal(err)
	}
	if got := files(t, root); len(got) != 3 {
		t.Errorf("after purging one source: %v", got)
	}
	if _, ok := d.Load(other, id(4, 0, 0), t0); !ok {
		t.Error("purging one source removed another's tiles")
	}
	if err := d.Empty(""); err != nil {
		t.Fatal(err)
	}
	if got := files(t, root); len(got) != 0 {
		t.Errorf("after purging everything: %v", got)
	}
	if held, _ := d.Held(); held != 0 {
		t.Errorf("held %d after a purge", held)
	}
}

// The hardening of plan task 06.17 (PL-IS-4).
func TestSymlinkInsideRootNotFollowed(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("NOT RUN: making a symbolic link needs a privilege on Windows")
	}
	root := t.TempDir()
	d := openDisk(t, root, DefaultDiskBytes)
	if err := d.Store(planet, id(2, 1, 1), tileBytes(t), t0); err != nil {
		t.Fatal(err)
	}
	real := filepath.Join(root, files(t, root)[0])
	secret := filepath.Join(root, "secret")
	if err := os.WriteFile(secret, []byte("not a tile"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(real); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(secret, real); err != nil {
		t.Fatal(err)
	}
	if body, ok := d.Load(planet, id(2, 1, 1), t0); ok {
		t.Errorf("a link inside the root was followed: %q", body)
	}
	if err := d.Store(planet, id(2, 1, 1), tileBytes(t), t0); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(secret); string(got) != "not a tile" {
		t.Error("a write went through the link")
	}
}

func TestCachedFileSizeCheckedBeforeRead(t *testing.T) {
	root := t.TempDir()
	d := openDisk(t, root, DefaultDiskBytes)
	if err := d.Store(planet, id(2, 1, 1), tileBytes(t), t0); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, files(t, root)[0])
	if err := os.Truncate(path, 3<<20); err != nil {
		t.Fatal(err)
	}
	if _, ok := d.Load(planet, id(2, 1, 1), t0); ok {
		t.Error("a file larger than any tile was read")
	}
	if got := files(t, root); len(got) != 0 {
		t.Errorf("the oversized file was kept: %v", got)
	}
	if err := d.Store(planet, id(2, 1, 1), make([]byte, 3<<20), t0); err == nil {
		t.Error("a body larger than any tile was written")
	}
}

func TestReadOnlyRootTolerated(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("NOT RUN: a read-only directory cannot be made for this user here")
	}
	root := t.TempDir()
	d := openDisk(t, root, DefaultDiskBytes)
	if err := d.Store(planet, id(0, 0, 0), tileBytes(t), t0); err != nil {
		t.Fatal(err)
	}
	d.Release()
	var dirs []string
	filepath.Walk(root, func(path string, info os.FileInfo, _ error) error {
		if info != nil && info.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	})
	for _, dir := range dirs {
		os.Chmod(dir, 0o500)
	}
	t.Cleanup(func() {
		for _, dir := range dirs {
			os.Chmod(dir, 0o700)
		}
	})
	ro := openDisk(t, root, DefaultDiskBytes)
	if _, ok := ro.Load(planet, id(0, 0, 0), t0.Add(3*time.Hour)); !ok {
		t.Error("a read-only cache did not serve what it holds")
	}
	net := &network{}
	p := pipeline(t, Options{Disk: ro, Network: &Network{Identity: planet, MaxZoom: 14, Get: net.get}})
	if failed := runAll(t, p.Plan(t0, []scene.TileID{id(1, 1, 1)})); failed != 0 {
		t.Error("a cache that cannot be written failed the tile")
	}
}

func TestMemoryHitRefreshesDiskRecency(t *testing.T) {
	root := t.TempDir()
	d := openDisk(t, root, DefaultDiskBytes)
	net := &network{}
	p := pipeline(t, Options{Disk: d, Network: &Network{Identity: planet, MaxZoom: 14, Get: net.get}})
	runAll(t, p.Plan(t0, []scene.TileID{id(0, 0, 0)}))
	path := filepath.Join(root, files(t, root)[0])
	// Hours later the tile is still drawn from memory. The plan only notes
	// it - a plan does no input or output - and the next job writes it down.
	later := t0.Add(5 * time.Hour)
	plan := p.Plan(later, []scene.TileID{id(0, 0, 0), id(1, 0, 0)})
	if info, _ := os.Stat(path); !info.ModTime().Equal(t0) {
		t.Error("the plan itself touched the disk")
	}
	runAll(t, plan)
	if info, _ := os.Stat(path); !info.ModTime().Equal(later) {
		t.Errorf("recency %v; a tile served from memory is still being read", info.ModTime())
	}
}

func TestCacheRootOpenedOnce(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("NOT RUN: a directory in use cannot be renamed on Windows")
	}
	base := t.TempDir()
	root := filepath.Join(base, "cache")
	d := openDisk(t, root, DefaultDiskBytes)
	// The root is moved aside and a stranger's directory put in its place:
	// the cache keeps to the handle it opened and checked.
	if err := os.Rename(root, filepath.Join(base, "moved")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(root, 0o777); err != nil {
		t.Fatal(err)
	}
	if err := d.Store(planet, id(0, 0, 0), tileBytes(t), t0); err != nil {
		t.Fatal(err)
	}
	if got := files(t, root); len(got) != 0 {
		t.Errorf("written under the path, not the handle: %v", got)
	}
	if got := files(t, filepath.Join(base, "moved")); len(got) != 1 {
		t.Errorf("the handle's directory holds %v", got)
	}
}
