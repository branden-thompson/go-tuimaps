package overlay

import (
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/project"
)

// BenchmarkSetIndexAtVertexCap is plan task 10.26's measurement: the one
// pass Set makes for a shape at the vertex cap, 2,000,000 vertices, with the
// borrow check off and on. PLAN estimated "a few milliseconds" by arithmetic;
// what it really costs is recorded in the build log, not gated.
func BenchmarkSetIndexAtVertexCap(b *testing.B) {
	ring := circle(project.LonLat{Lon: -95, Lat: 38}, 20, 1_999_999)
	for _, check := range []bool{false, true} {
		name := "borrow check off"
		if check {
			name = "borrow check on"
		}
		b.Run(name, func(b *testing.B) {
			s, err := NewStore(Caps{BorrowCheck: check})
			if err != nil {
				b.Fatal(err)
			}
			o := alert("huge", ring)
			b.ResetTimer()
			for range b.N {
				if _, err := s.Set(o); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
