//go:build !race

package mvt

import "testing"

// TestWireAllocatesNothing is the allocation half of plan task 03.1. It is
// built only without the race detector, under which counts are not stable.
func TestWireAllocatesNothing(t *testing.T) {
	msg := []byte{0x1a, 0x03, 'a', 'b', 'c', 0x78, 0xac, 0x02, 0x15, 1, 2, 3, 4}
	allocs := testing.AllocsPerRun(100, func() {
		rest := msg
		for len(rest) > 0 {
			var err error
			_, rest, err = nextField(rest)
			if err != nil {
				t.Fatal(err)
			}
		}
	})
	if allocs != 0 {
		t.Errorf("scanning fields allocated %v times a run; want none", allocs)
	}
}
