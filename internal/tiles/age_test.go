package tiles

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// fileTime is the modification time of the one cached file for a tile.
func fileTime(t *testing.T, root, identity string, tile scene.TileID) time.Time {
	t.Helper()
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(tilePath(identity, tile))))
	if err != nil {
		t.Fatal(err)
	}
	return info.ModTime()
}

// TestReadsNeverTouchTheFileTime is v0.2.0 L9.1 (L-9.4, D-56): a tile's file
// time is when it was fetched. Read from disk every hour for a day, or served
// from memory, it never changes: the disk keeps no record of when a place was
// looked at.
func TestReadsNeverTouchTheFileTime(t *testing.T) {
	root := t.TempDir()
	d := openDisk(t, root, DefaultDiskBytes)
	if err := d.Store(planet, id(1, 0, 0), tileBytes(t), t0, d.Generation()); err != nil {
		t.Fatal(err)
	}
	for hour := 1; hour <= 24; hour++ {
		if _, ok := d.Load(planet, id(1, 0, 0), t0.Add(time.Duration(hour)*time.Hour)); !ok {
			t.Fatalf("hour %d: miss", hour)
		}
		if got := fileTime(t, root, planet, id(1, 0, 0)); !got.Equal(t0) {
			t.Fatalf("hour %d: a read moved the file time to %v", hour, got)
		}
	}

	net := &network{}
	p := pipeline(t, Options{Disk: d, Network: &Network{Identity: planet, MaxZoom: 14, Get: net.get}})
	runAll(t, p.Plan(t0, []scene.TileID{id(0, 0, 0)}))
	later := t0.Add(5 * time.Hour)
	runAll(t, p.Plan(later, []scene.TileID{id(0, 0, 0), id(1, 1, 0)})) // the first from memory, the second fetched
	if got := fileTime(t, root, planet, id(0, 0, 0)); !got.Equal(t0) {
		t.Errorf("a tile served from memory had its file time moved to %v", got)
	}
	if got := fileTime(t, root, planet, id(1, 1, 0)); !got.Equal(later) {
		t.Errorf("a tile fetched at %v was dated %v", later, got)
	}
}

// TestATilePastItsAgeIsFetchedAgain is v0.2.0 L9.2 (L-9.4): a tile at its
// maximum age is not served, and its file is gone; a tile dated after now
// counts as expired.
func TestATilePastItsAgeIsFetchedAgain(t *testing.T) {
	root := t.TempDir()
	d := openDisk(t, root, DefaultDiskBytes)
	d.SetMaxAge(time.Hour)
	for _, c := range []struct {
		tile    scene.TileID
		fetched time.Time
	}{{id(3, 0, 0), t0}, {id(3, 1, 0), t0.Add(2 * time.Hour)}} {
		if err := d.Store(planet, c.tile, tileBytes(t), c.fetched, d.Generation()); err != nil {
			t.Fatal(err)
		}
	}
	if _, ok := d.Load(planet, id(3, 0, 0), t0.Add(59*time.Minute)); !ok {
		t.Error("a tile inside its age was not served")
	}
	if _, ok := d.Load(planet, id(3, 0, 0), t0.Add(time.Hour)); ok {
		t.Error("a tile at its maximum age was served")
	}
	if _, ok := d.Load(planet, id(3, 1, 0), t0.Add(time.Hour)); ok {
		t.Error("a tile dated after now was served")
	}
	if got := files(t, root); len(got) != 0 {
		t.Errorf("expired files were kept: %v", got)
	}

	// Through the pipeline: the aged tile is fetched again and dated anew.
	if err := d.Store(planet, id(0, 0, 0), tileBytes(t), t0, d.Generation()); err != nil {
		t.Fatal(err)
	}
	net := &network{}
	p := pipeline(t, Options{Disk: d, Network: &Network{Identity: planet, MaxZoom: 14, Get: net.get}})
	later := t0.Add(3 * time.Hour)
	runAll(t, p.Plan(later, []scene.TileID{id(0, 0, 0)}))
	if len(net.asked) != 1 {
		t.Errorf("%d fetches; the aged tile was served from disk", len(net.asked))
	}
	if got := fileTime(t, root, planet, id(0, 0, 0)); !got.Equal(later) {
		t.Errorf("the tile fetched again is dated %v, want %v", got, later)
	}
}

// TestEveryJobDropsWhatHasAged is v0.2.0 L9.2 (L-9.4): the maximum age is
// kept on disk, not only on reading. A job for one tile removes every other
// tile past its age, and a tile dated after now.
func TestEveryJobDropsWhatHasAged(t *testing.T) {
	root := t.TempDir()
	d := openDisk(t, root, DefaultDiskBytes)
	d.SetMaxAge(time.Hour)
	for x := uint32(0); x < 3; x++ {
		if err := d.Store(planet, id(5, x, 0), tileBytes(t), t0.Add(time.Duration(x)*20*time.Minute), d.Generation()); err != nil {
			t.Fatal(err)
		}
	}
	if err := d.Store(planet, id(5, 9, 0), tileBytes(t), t0.Add(24*time.Hour), d.Generation()); err != nil {
		t.Fatal(err)
	}
	net := &network{}
	p := pipeline(t, Options{Disk: d, Network: &Network{Identity: planet, MaxZoom: 14, Get: net.get}})
	runAll(t, p.Plan(t0.Add(70*time.Minute), []scene.TileID{id(0, 0, 0)}))
	want := map[string]bool{tilePath(planet, id(5, 1, 0)): true, tilePath(planet, id(5, 2, 0)): true, tilePath(planet, id(0, 0, 0)): true}
	got := files(t, root)
	for _, f := range got {
		if !want[f] {
			t.Errorf("kept %s", f)
		}
	}
	if len(got) != len(want) {
		t.Errorf("kept %v, want the three inside their age", got)
	}
	if held, _ := d.Held(); held != 3*int64(len(tileBytes(t))) {
		t.Errorf("held %d after the job", held)
	}
}

// TestEvictOldestFetchedFirst is v0.2.0 L9.2 (L-9.4, D-56): over its cap the
// cache drops the tiles fetched longest ago, however recently they were read,
// and never a tile the view needs.
func TestEvictOldestFetchedFirst(t *testing.T) {
	root := t.TempDir()
	body := tileBytes(t)
	size := int64(len(body))
	d := openDisk(t, root, 10*size)
	for x := uint32(0); x < 10; x++ {
		if err := d.Store(planet, id(8, x, 0), body, t0.Add(time.Duration(x)*time.Minute), d.Generation()); err != nil {
			t.Fatal(err)
		}
	}
	later := t0.Add(2 * time.Hour)
	if _, ok := d.Load(planet, id(8, 1, 0), later); !ok { // read recently: it makes no difference
		t.Fatal("miss")
	}
	d.InView(planet, []scene.TileID{id(8, 0, 0)})
	if err := d.Store(planet, id(8, 10, 0), body, later, d.Generation()); err != nil {
		t.Fatal(err)
	}
	held, limit := d.Held()
	if held > limit/10*9 {
		t.Errorf("held %d after pruning, want at most nine tenths of %d", held, limit)
	}
	for x, kept := range map[uint32]bool{0: true, 1: false, 2: false, 3: true, 10: true} {
		if _, ok := d.Load(planet, id(8, x, 0), later); ok != kept {
			t.Errorf("tile %d: kept %v, want %v", x, ok, kept)
		}
	}
	d.Release()
	reopened := openDisk(t, root, 10*size)
	if again, _ := reopened.Held(); again != held {
		t.Errorf("reopened: held %d, want %d", again, held)
	}
}

// TestAStoreFromAnEarlierGenerationWritesNothing is v0.2.0 L9.3 (L-9.3) at
// the disk: a job that began before a purge or a release lands nothing after
// it.
func TestAStoreFromAnEarlierGenerationWritesNothing(t *testing.T) {
	root := t.TempDir()
	d := openDisk(t, root, DefaultDiskBytes)
	before := d.Generation()
	if _, err := d.Empty(); err != nil {
		t.Fatal(err)
	}
	if err := d.Store(planet, id(0, 0, 0), tileBytes(t), t0, before); err == nil {
		t.Error("a store begun before the purge was accepted")
	}
	now := d.Generation()
	d.Release()
	if err := d.Store(planet, id(0, 0, 0), tileBytes(t), t0, now); err == nil {
		t.Error("a store after the release was accepted")
	}
	if got := files(t, root); len(got) != 0 {
		t.Errorf("written: %v", got)
	}
}

// TestPurgeCountsWhatItRemoved is v0.2.0 L9.4 and L9.5 (L-9.3, L-9.5): every
// source's tiles go, and the report counts only the removals that succeeded.
func TestPurgeCountsWhatItRemoved(t *testing.T) {
	root := t.TempDir()
	d := openDisk(t, root, DefaultDiskBytes)
	for _, source := range []string{planet, "https://other.example/"} {
		for x := uint32(0); x < 3; x++ {
			if err := d.Store(source, id(4, x, 0), tileBytes(t), t0, d.Generation()); err != nil {
				t.Fatal(err)
			}
		}
	}
	report, err := d.Empty()
	if err != nil || report.Removed != 6 || report.Failed != 0 {
		t.Errorf("purge: %+v, %v; want 6 removed", report, err)
	}
	if got := files(t, root); len(got) != 0 {
		t.Errorf("after the purge: %v", got)
	}
	if held, _ := d.Held(); held != 0 {
		t.Errorf("held %d after a purge", held)
	}

	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("NOT RUN: a directory this user cannot remove from cannot be made here")
	}
	if err := d.Store(planet, id(4, 0, 0), tileBytes(t), t0, d.Generation()); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, filepath.FromSlash(sourceDir(planet)), "4")
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o700) })
	report, err = d.Empty()
	if err == nil || report.Removed != 0 || report.Failed != 1 {
		t.Errorf("purge of a file that cannot be removed: %+v, %v; want 1 failed, and an error", report, err)
	}
}

// TestAFailedWriteIsWarnedOf is v0.2.0 L9.5 (L-9.5): a cache that cannot be
// written still serves, and says so, counting each failure.
func TestAFailedWriteIsWarnedOf(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("NOT RUN: a read-only directory cannot be made for this user here")
	}
	root := t.TempDir()
	d := openDisk(t, root, DefaultDiskBytes)
	if err := os.Chmod(root, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(root, 0o700) })
	net := &network{}
	p := pipeline(t, Options{Disk: d, Network: &Network{Identity: planet, MaxZoom: 14, Get: net.get}})
	if failed := runAll(t, p.Plan(t0, []scene.TileID{id(1, 0, 0), id(1, 1, 0)})); failed != 0 {
		t.Error("a cache that cannot be written failed the tiles")
	}
	var got []fault.Warning
	for _, w := range p.TakeWarnings() {
		if w.Kind == fault.CacheWriteFailed {
			got = append(got, w)
		}
	}
	if len(got) != 1 || got[0].Count != len(net.asked) || len(net.asked) < 2 {
		t.Errorf("warnings %+v; want one cache-write-failed counting the %d tiles fetched", got, len(net.asked))
	}
	if again := p.TakeWarnings(); len(again) != 0 {
		t.Errorf("warned again: %+v", again)
	}
}

// TestARootOthersCanReadIsWarnedOf is v0.2.0 L9.6 (L-9.6): the file names
// are a record of where the user looked, so a root others can read is
// warned of, once.
func TestARootOthersCanReadIsWarnedOf(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("NOT RUN: the check is not performed on Windows")
	}
	for _, c := range []struct {
		mode os.FileMode
		warn bool
	}{{0o700, false}, {0o750, true}, {0o705, true}} {
		root := filepath.Join(t.TempDir(), "cache")
		if err := os.Mkdir(root, c.mode); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(root, c.mode); err != nil { // the umask may have taken bits off
			t.Fatal(err)
		}
		d := openDisk(t, root, DefaultDiskBytes)
		got := d.TakeWarnings()
		if warned := len(got) == 1 && got[0].Kind == fault.CacheRootReadable; warned != c.warn {
			t.Errorf("%v: warnings %+v, want a warning %v", c.mode, got, c.warn)
		}
		if again := d.TakeWarnings(); len(again) != 0 {
			t.Errorf("%v: warned twice", c.mode)
		}
	}
}

// endless is a file that never ends, counting what was read of it.
type endless struct{ read int }

func (e *endless) Read(p []byte) (int, error) {
	e.read += len(p)
	return len(p), nil
}

// TestACachedFileIsReadThroughItsLimit is v0.2.0 L9.6 (L-12.6): every read of
// a cached file, Verify's among them, stops past the largest tile.
func TestACachedFileIsReadThroughItsLimit(t *testing.T) {
	e := &endless{}
	if _, ok := readTile(e); ok {
		t.Error("a file larger than any tile was read as one")
	}
	if e.read > maxTileBytes+64<<10 {
		t.Errorf("read %d bytes of it; the limit is %d", e.read, maxTileBytes)
	}
}

// TestNoClockJudgesNothing is v0.2.0 L9.2 (L-9.4): a job with no host clock
// yet, as a Settle before any frame, neither expires a tile nor dates one to
// the year 1; the system dates the file, and age is judged when a clock comes.
func TestNoClockJudgesNothing(t *testing.T) {
	root := t.TempDir()
	d := openDisk(t, root, DefaultDiskBytes)
	d.SetMaxAge(time.Hour)
	if err := d.Store(planet, id(2, 0, 0), tileBytes(t), t0, d.Generation()); err != nil {
		t.Fatal(err)
	}
	if err := d.Store(planet, id(2, 1, 0), tileBytes(t), time.Time{}, d.Generation()); err != nil {
		t.Fatal(err)
	}
	if got := fileTime(t, root, planet, id(2, 1, 0)); got.Year() < 2000 {
		t.Errorf("a tile stored with no clock was dated %v", got)
	}
	d.Expire(time.Time{})
	if _, ok := d.Load(planet, id(2, 0, 0), time.Time{}); !ok {
		t.Error("with no clock, a tile was judged aged")
	}
}

// TestEachJobKeepsTheAgeAsTimeMoves is v0.2.0 L9.2 (L-9.4): a job that finds
// nothing aged does not stop a later one from dropping what has aged since.
func TestEachJobKeepsTheAgeAsTimeMoves(t *testing.T) {
	root := t.TempDir()
	d := openDisk(t, root, DefaultDiskBytes)
	d.SetMaxAge(time.Hour)
	if err := d.Store(planet, id(5, 0, 0), tileBytes(t), t0, d.Generation()); err != nil {
		t.Fatal(err)
	}
	d.Expire(t0.Add(30 * time.Minute))
	if len(files(t, root)) != 1 {
		t.Fatal("a tile inside its age was dropped")
	}
	if err := d.Store(planet, id(5, 1, 0), tileBytes(t), t0.Add(40*time.Minute), d.Generation()); err != nil {
		t.Fatal(err)
	}
	d.Expire(t0.Add(61 * time.Minute))
	if got := files(t, root); len(got) != 1 || got[0] != tilePath(planet, id(5, 1, 0)) {
		t.Errorf("after the first tile aged: %v", got)
	}
	d.Expire(t0.Add(101 * time.Minute))
	if got := files(t, root); len(got) != 0 {
		t.Errorf("after the second tile aged: %v", got)
	}
	// A walk that found nothing, then a store: the new tile still ages.
	if err := d.Store(planet, id(5, 2, 0), tileBytes(t), t0.Add(110*time.Minute), d.Generation()); err != nil {
		t.Fatal(err)
	}
	d.Expire(t0.Add(171 * time.Minute))
	if got := files(t, root); len(got) != 0 {
		t.Errorf("a tile stored after an empty walk did not age: %v", got)
	}
}

// TestAPlanWithNoClockKeepsTheLastOne is v0.2.0 L9.1 (L-9.4): Settle plans
// with no clock; the tiles its jobs fetch are dated by the last clock a
// frame gave, not left undated.
func TestAPlanWithNoClockKeepsTheLastOne(t *testing.T) {
	root := t.TempDir()
	d := openDisk(t, root, DefaultDiskBytes)
	net := &network{}
	p := pipeline(t, Options{Disk: d, Network: &Network{Identity: planet, MaxZoom: 14, Get: net.get}})
	p.Plan(t0, []scene.TileID{id(0, 0, 0)})
	runAll(t, p.Plan(time.Time{}, []scene.TileID{id(0, 0, 0)}))
	if got := fileTime(t, root, planet, id(0, 0, 0)); !got.Equal(t0) {
		t.Errorf("a tile fetched after a plan with no clock was dated %v, want %v", got, t0)
	}
}

// TestTheViewsTilesSurviveThePipelinesPruning is v0.2.0 L9.2 (L-9.4) through
// the pipeline: the plan names the tiles the view needs to the disk cache,
// and pruning over the cap passes them by, however old.
func TestTheViewsTilesSurviveThePipelinesPruning(t *testing.T) {
	root := t.TempDir()
	body := tileBytes(t)
	d := openDisk(t, root, 4*int64(len(body)))
	for x := uint32(0); x < 4; x++ {
		if err := d.Store(planet, id(8, x, 0), body, t0.Add(time.Duration(x)*time.Minute), d.Generation()); err != nil {
			t.Fatal(err)
		}
	}
	net := &network{}
	p := pipeline(t, Options{Disk: d, Network: &Network{Identity: planet, MaxZoom: 14, Get: net.get}})
	runAll(t, p.Plan(t0.Add(time.Hour), []scene.TileID{id(8, 0, 0), id(8, 9, 0)}))
	// The stand-ins are fetched and stored first, over the cap: the view's
	// oldest tile must still be served from disk, not fetched again.
	for _, asked := range net.asked {
		if asked == id(8, 0, 0) {
			t.Error("the oldest tile, which the view needs, was evicted and fetched again")
		}
	}
	if _, ok := d.Load(planet, id(8, 0, 0), t0.Add(time.Hour)); !ok {
		t.Error("the oldest tile, which the view needs, was evicted")
	}
	if _, ok := d.Load(planet, id(8, 1, 0), t0.Add(time.Hour)); ok {
		t.Error("nothing was evicted, so this proves nothing")
	}
}
