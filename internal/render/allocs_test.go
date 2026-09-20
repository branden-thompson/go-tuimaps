//go:build !race

package render

import "testing"

// TestUnchangedFrameZeroAllocs is plan task 09.19 (NFR-4). It is built only
// without the race detector, which allocates on its own account.
func TestUnchangedFrameZeroAllocs(t *testing.T) {
	v := fitted(149, 38)
	in := input(t, v)
	r, err := NewRenderer(v.Cols, v.Rows)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Render(in); err != nil {
		t.Fatal(err)
	}
	allocs := testing.AllocsPerRun(100, func() {
		if _, err := r.Render(in); err != nil {
			t.Fatal(err)
		}
	})
	if allocs != 0 {
		t.Errorf("an unchanged frame cost %v allocations, want none", allocs)
	}
}
