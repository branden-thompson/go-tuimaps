package overlay

import (
	"context"
	"encoding/binary"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/project"
)

// fuzzOverlay makes an overlay out of arbitrary bytes: any id, any times,
// every kind of feature and some that are none, coordinates that are any
// 64-bit pattern at all - numbers, infinities, not-a-numbers.
func fuzzOverlay(data []byte) Overlay {
	take := func(n int) []byte {
		if len(data) < n {
			out := make([]byte, n)
			copy(out, data)
			data = nil
			return out
		}
		out := data[:n]
		data = data[n:]
		return out
	}
	number := func() float64 {
		b := take(8)
		if b[7]&3 == 0 { // often a plausible coordinate, so that some overlays are accepted
			return float64(int16(binary.LittleEndian.Uint16(b))) / 180
		}
		return math.Float64frombits(binary.LittleEndian.Uint64(b))
	}
	head := take(4)
	o := Overlay{ID: string(take(int(head[0]) % 40)), Keeps: time.Duration(int64(binary.LittleEndian.Uint16(take(2)))-100) * time.Minute, Credit: string(take(int(head[1]) % 20))}
	if head[2]&7 != 0 {
		o.Valid = noon.Add(time.Duration(int8(head[3])) * time.Hour)
	}
	for range int(head[2]>>3) % 5 {
		f := Feature{Kind: FeatureKind(take(1)[0] % 6), Role: colour.Token(take(1)[0] % 70), Label: string(take(int(take(1)[0]) % 12)), RadiusKm: number(),
			Centre: project.LonLat{Lon: number(), Lat: number()}}
		for range int(take(1)[0]) % 4 {
			var ring []project.LonLat
			for range int(take(1)[0]) % 9 {
				ring = append(ring, project.LonLat{Lon: number(), Lat: number()})
			}
			f.Rings = append(f.Rings, ring)
		}
		o.Features = append(o.Features, f)
	}
	return o
}

// FuzzHandIn is plan task 10.24 (PL-IS-1): whatever struct a host hands in,
// nothing panics, a refusal is one of the library's own kinds, a refused Set
// leaves its warning and keeps nothing, and an accepted one can be read,
// projected, simplified, split and removed.
func FuzzHandIn(f *testing.F) {
	f.Add([]byte("warnings and not much else"))
	f.Add([]byte{8, 3, 15, 0, 'w', 'a', 'r', 'n', 'i', 'n', 'g', 's', 160, 0, 'N', 'W', 'S', 3, 23, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 5, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		s, err := NewStore(Caps{OverlayVertices: 64, StoreVertices: 100})
		if err != nil {
			t.Fatal(err)
		}
		o := fuzzOverlay(data)
		res, err := s.Set(o)
		if err != nil {
			var own *fault.Error
			if !errors.As(err, &own) || own.Kind().String() == "unknown" {
				t.Fatalf("a refusal that is not one of the library's kinds: %v", err)
			}
			if w := s.TakeWarnings(); len(w) != 1 || w[0].Kind != fault.SetRefused {
				t.Fatalf("a refused Set left %+v", w)
			}
			if len(s.IDs()) != 0 {
				t.Fatal("a refused overlay was kept")
			}
			return
		}
		if !res.Created || !res.Released {
			t.Fatalf("a first Set: %+v", res)
		}
		reader, ok := s.Read(o.ID)
		if !ok {
			t.Fatal("an accepted overlay cannot be read")
		}
		for _, feature := range reader.Overlay().Features {
			for _, ring := range feature.Rings {
				parts, err := Split(ring)
				if err != nil {
					t.Fatalf("an accepted ring could not be split: %v", err)
				}
				for _, part := range parts {
					vertices, err := Project(part)
					if err != nil {
						t.Fatalf("an accepted ring could not be projected: %v", err)
					}
					Keep(Simplify(vertices, Tolerance(6)))
				}
			}
		}
		view := s.Register()
		view.Publish(6)
		if err := s.PrepareJob(o.ID, 6).Run(context.Background()); err != nil {
			t.Fatalf("an accepted overlay could not be prepared: %v", err)
		}
		if _, _, path := s.Drawn(o.ID, 6); path == NotReady {
			t.Fatal("a prepared overlay is not ready to draw")
		}
		if use := s.ShapeUse(); use.Held < 0 || use.Need < 0 {
			t.Fatalf("%+v", use)
		}
		view.Withdraw()
		if again, _ := s.Set(o); again.Created || again.Released {
			t.Fatalf("a replace under a reader: %+v", again)
		}
		if released := reader.Done(); len(released) != 1 {
			t.Fatalf("released %v", released)
		}
		if gone, err := s.Remove(o.ID); err != nil || !gone.Found || !gone.Released {
			t.Fatalf("%+v, %v", gone, err)
		}
	})
}
