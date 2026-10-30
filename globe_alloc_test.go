//go:build !race

package tuimaps_test

// globe_alloc_test.go — plan task P7.1 (M5, watchpost D-209, D-350): the work
// of M5's frame, gated on every machine by its allocations. Its time is held
// on the reference machine by M5's record (m5_record_test.go). Built without
// the race detector, which allocates on its own account.

import (
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// globePanAllocs is the budget a pan frame of M5's at 400 by 110 may
// allocate: set from the first measurement, and only ever lowered by the
// work that follows.
const globePanAllocs = 4600 // measured 2026-10-30: 4,200 on arm64 and on amd64, both choices over water

// TestAGlobeFieldPanAllocates holds M5's frame at 400 by 110, over the sea and
// stopping at the shore, to globePanAllocs: a frame whose allocations grew
// with its 44,000 cells - the one an allocation a cell D-209 measured - fails
// it on any machine.
func TestAGlobeFieldPanAllocates(t *testing.T) {
	size := tuimaps.Size{Cols: 400, Rows: 110}
	for _, overWater := range []bool{true, false} {
		step := globePan(t, size, overWater)
		step(0)
		step(1)
		i := 0
		allocs := testing.AllocsPerRun(20, func() {
			step(i)
			i++
		})
		t.Logf("a pan frame at 400x110 %s: %.0f allocations", waterName[overWater], allocs)
		if allocs > globePanAllocs {
			t.Errorf("a pan frame at 400x110 %s allocates %.0f times; its budget is %d", waterName[overWater], allocs, globePanAllocs)
		}
	}
}
