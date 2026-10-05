package tuimaps_test

// contract_claims_test.go — v0.2.0 L1.3 (L-4.2): the contract's behavioural
// claims that were false are corrected, and each corrected claim is held here
// by a test of the behaviour it now describes, so a sentence cannot be right
// while the code does something else.

import (
	"context"
	"image/color"
	"io/fs"
	"path/filepath"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// TestAPanickingRenderReturnsAnEmptyFrameAndAnInternalError holds contract
// section 6, rule 4: a panic inside Render gives an error of the internal kind
// and an empty frame — there is no "failed" status and no last good rows.
func TestAPanickingRenderReturnsAnEmptyFrameAndAnInternalError(t *testing.T) {
	m := world(t, 40, 12)
	tuimaps.PlantPanic(m, "Render")
	f, err := m.Render(tuimaps.Size{Cols: 40, Rows: 12}, noon)
	tuimaps.ClearPanic(m)
	if !isKind(err, fault.Internal) {
		t.Fatalf("a panic inside Render gave %v; want an error of the internal kind", err)
	}
	if len(f.Lines) != 0 || f.Status != 0 {
		t.Errorf("the frame after a panic has %d rows and status %v; want an empty frame", len(f.Lines), f.Status)
	}
}

// TestSetAndRemoveReportTheRelease holds contract section 4: with one
// goroutine making every call, Set and Remove report the old geometry released
// at once; Work and Settle carry no released ids.
func TestSetAndRemoveReportTheRelease(t *testing.T) {
	m := world(t, 40, 12)
	if _, err := m.Set(warning("alerts")); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Render(tuimaps.Size{Cols: 40, Rows: 12}, noon); err != nil {
		t.Fatal(err)
	}
	replaced, err := m.Set(warning("alerts"))
	if err != nil {
		t.Fatal(err)
	}
	if replaced.Created || !replaced.Released {
		t.Errorf("replacing on one goroutine: %+v; want replaced and released", replaced)
	}
	removed, err := m.Remove("alerts")
	if err != nil {
		t.Fatal(err)
	}
	if !removed.Found || !removed.Released {
		t.Errorf("removing on one goroutine: %+v; want found and released", removed)
	}
}

// TestChangedCountsInputsAndWhatWorkLands holds contract section 1 as
// v0.2.0 has it (D-66, L4.7): the counter moves on every host call that
// changes an input and when Work lands something, with no Render; Render
// never moves it, whether or not the frame it draws differs.
func TestChangedCountsInputsAndWhatWorkLands(t *testing.T) {
	m := world(t, 40, 12)
	size := tuimaps.Size{Cols: 40, Rows: 12}
	before := m.Changed()
	if _, err := m.Set(warning("alerts")); err != nil {
		t.Fatal(err)
	}
	if m.Changed() == before {
		t.Error("Set changed an input and did not move the counter")
	}
	if err := m.Zoom(3); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Render(size, noon); err != nil { // notes the tiles zoom 3 wants
		t.Fatal(err)
	}
	quiet := m.Changed()
	if _, err := m.Render(size, noon); err != nil {
		t.Fatal(err)
	}
	if m.Changed() != quiet {
		t.Errorf("a Render of unchanged inputs moved the counter from %d to %d", quiet, m.Changed())
	}
	did, err := m.Work(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !did {
		t.Fatal("zoom 3 wanted no tile, so this proves nothing: pick a zoom that needs one")
	}
	landed := m.Changed()
	if landed == quiet {
		t.Error("Work landed a tile and the counter did not move: a host would not know to render")
	}
	if _, err := m.Render(size, noon); err != nil {
		t.Fatal(err)
	}
	if m.Changed() != landed {
		t.Errorf("the Render that drew the landed tile moved the counter from %d to %d; Render never does", landed, m.Changed())
	}
	// A picture the work prepares counts as a tile does. The tiles are all
	// landed first, so the picture is the only work left.
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	loop := tuimaps.RadarImage("radar", tuimaps.Image{PNG: solidPNG(t, 8, 6, color.NRGBA{R: 200, A: 255}), West: -90, South: 30, East: -80, North: 40,
		Projection: tuimaps.PlateCarree, Table: []tuimaps.TableEntry{{Colour: tuimaps.RGB{R: 200}, Value: 25}}, Exact: true}, noon)
	if _, err := m.Set(loop); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Render(size, noon); err != nil { // notes the picture as wanted
		t.Fatal(err)
	}
	set := m.Changed()
	for {
		did, err := m.Work(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if !did {
			break
		}
	}
	if m.Changed() == set {
		t.Error("Work prepared a picture and the counter did not move")
	}
}

// TestPurgeDoesNotReachAReleasedRoot holds contract section 11, L-9.3: a root
// that CacheRoot replaced is released, and Purge empties only the root the map
// holds now. L9.4 widens what Purge empties; this stays true after it.
func TestPurgeDoesNotReachAReleasedRoot(t *testing.T) {
	dir := t.TempDir()
	old, now := filepath.Join(dir, "old"), filepath.Join(dir, "now")
	srv := serveTiles(t)
	m, err := tuimaps.New(tuimaps.WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if err := m.Source(srv.url + "/tiles/"); err != nil {
		t.Fatal(err)
	}
	if err := m.CacheRoot(old, 0); err != nil {
		t.Fatal(err)
	}
	drawAndSettle(t, m)
	held := filesUnder(t, old)
	if held == 0 {
		t.Fatal("nothing reached the first root, so this proves nothing")
	}
	if err := m.CacheRoot(now, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Purge(); err != nil {
		t.Fatal(err)
	}
	if after := filesUnder(t, old); after != held {
		t.Errorf("Purge reached the released root: %d files before, %d after", held, after)
	}
}

// filesUnder counts the regular files below dir.
func filesUnder(t *testing.T, dir string) int {
	t.Helper()
	n := 0
	err := filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
		if err == nil && d.Type().IsRegular() {
			n++
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return n
}
