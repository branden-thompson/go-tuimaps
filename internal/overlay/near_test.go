package overlay

// near_test.go — L-2.2, OW-10 (D-104): a pixel matched by the tolerance, not
// exactly, is counted and warned of, where the table is one the provider
// publishes or the host gives; a table read off a legend expects them.

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// cornered is a 20-pixel image of one colour but for a square of another in
// its corner, read with the provider or table given.
func cornered(t testing.TB, ground, corner color.NRGBA, side int, provider Provider, table []TableEntry) *Image {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 20, 20))
	for y := range 20 {
		for x := range 20 {
			c := ground
			if x < side && y < side {
				c = corner
			}
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return &Image{PNG: buf.Bytes(), West: -95, South: 30, East: -85, North: 40, Projection: PlateCarree, Provider: provider,
		Table: table, Type: Type{Preset: "radar", Unit: "dBZ"}}
}

// nudged is a colour moved a few steps, inside the tolerance of 10.
func nudged(c colour.RGB) color.NRGBA {
	step := func(v uint8) uint8 {
		if v > 250 {
			return v - 3
		}
		return v + 3
	}
	return color.NRGBA{R: step(c.R), G: c.G, B: c.B, A: 255}
}

// nearAfter hands in an image, reads it, and returns its report and the
// near-image-colours warning's count (0 for none).
func nearAfter(t *testing.T, img *Image) (Report, int) {
	t.Helper()
	s := store(t)
	if _, err := s.HandIn(Overlay{ID: "radar", Valid: noon, Keeps: time.Hour, Image: img}); err != nil {
		t.Fatal(err)
	}
	if err := s.PrepareJob("radar", 0).Run(t.Context()); err != nil {
		t.Fatal(err)
	}
	_, report, ok := s.Raster("radar")
	if !ok {
		t.Fatal("nothing read")
	}
	count := 0
	for _, w := range s.TakeWarnings() {
		if w.Kind == fault.NearImageColours {
			count = w.Count
		}
	}
	return report, count
}

// TestANearColourOnAPublishedTableIsCountedAndWarned is OW-10: IEM's table
// is published and its pictures match it exactly, so four pixels a few steps
// off one of its colours are a palette that moved - counted, warned of with
// the count, and still read as that colour's class.
func TestANearColourOnAPublishedTableIsCountedAndWarned(t *testing.T) {
	iem, ok := TableOf(ProviderIEM)
	if !ok || len(iem.Entries) < 4 {
		t.Fatal("no IEM table")
	}
	c := iem.Entries[3].Colour
	exact := color.NRGBA{R: c.R, G: c.G, B: c.B, A: 255}
	report, warned := nearAfter(t, cornered(t, exact, nudged(c), 2, ProviderIEM, nil))
	if report.Near != 4 || report.Unmatched != 0 || warned != 4 {
		t.Errorf("four nudged pixels on IEM: %d near, %d unmatched, warned %d; want 4, 0, 4", report.Near, report.Unmatched, warned)
	}
	report, warned = nearAfter(t, cornered(t, exact, exact, 2, ProviderIEM, nil))
	if report.Near != 0 || warned != 0 {
		t.Errorf("an exact picture: %d near, warned %d; want none", report.Near, warned)
	}
}

// TestANearColourOnAHostsTableIsCounted: a host's own table is the host's
// word for its colours, and a near match on it is counted as on IEM's.
func TestANearColourOnAHostsTableIsCounted(t *testing.T) {
	c := colour.RGB{R: 40, G: 160, B: 40}
	table := []TableEntry{{Colour: c, Value: 20}, {Colour: colour.RGB{R: 200, G: 30, B: 30}, Value: 50}}
	report, warned := nearAfter(t, cornered(t, color.NRGBA{R: c.R, G: c.G, B: c.B, A: 255}, nudged(c), 3, 0, table))
	if report.Near != 9 || warned != 9 {
		t.Errorf("nine nudged pixels on a host's table: %d near, warned %d; want 9 and 9", report.Near, warned)
	}
}

// TestALegendTableExpectsNearColours: MRMS's table was read off its legend
// (Approximate), so a colour near one of its entries is what it expects, and
// is not counted.
func TestALegendTableExpectsNearColours(t *testing.T) {
	mrms, ok := TableOf(ProviderMRMS)
	if !ok || !mrms.Approximate || len(mrms.Entries) < 4 {
		t.Fatal("MRMS's table is not the approximate one")
	}
	c := mrms.Entries[3].Colour
	report, warned := nearAfter(t, cornered(t, color.NRGBA{R: c.R, G: c.G, B: c.B, A: 255}, nudged(c), 2, ProviderMRMS, nil))
	if report.Near != 0 || warned != 0 {
		t.Errorf("nudged pixels on MRMS's legend table: %d near, warned %d; want none", report.Near, warned)
	}
}
