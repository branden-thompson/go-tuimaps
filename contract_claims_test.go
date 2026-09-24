package tuimaps_test

// contract_claims_test.go — v0.2.0 L1.3 (L-4.2): the contract's behavioural
// claims that were false are corrected, and each corrected claim is held here
// by a test of the behaviour it now describes, so a sentence cannot be right
// while the code does something else.

import (
	"context"
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

// TestChangedMissesWhatWorkLands holds contract section 1 as v0.1.0 has it:
// the counter moves on every host call that changes an input (Set, the look,
// places, the view) and inside Render when the frame differs — but NOT when
// Work lands a tile, which reaches the counter only at the next Render. That
// last is the gap v0.2.0 closes (D-66, L4.7), and this test changes with it.
func TestChangedMissesWhatWorkLands(t *testing.T) {
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
	if _, err := m.Render(size, noon); err != nil {
		t.Fatal(err)
	}
	settled := m.Changed()
	did, err := m.Work(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !did {
		t.Fatal("zoom 3 wanted no tile, so this proves nothing: pick a zoom that needs one")
	}
	if m.Changed() != settled {
		t.Errorf("Work landed a tile and moved the counter from %d to %d; in v0.1.0 it does not", settled, m.Changed())
	}
	if _, err := m.Render(size, noon); err != nil {
		t.Fatal(err)
	}
	if m.Changed() == settled {
		t.Error("the render that drew the landed tile did not move the counter")
	}
}
