package overlay

import (
	"context"
	"encoding/binary"
	"errors"
	"image/color"
	"math"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

func ownKind(t *testing.T, err error) {
	t.Helper()
	var own *fault.Error
	if !errors.As(err, &own) || own.Kind().String() == "unknown" {
		t.Fatalf("an error that is not one of the library's kinds: %v", err)
	}
}

// FuzzImage is part of plan task 10.22 (PL-IS-1): whatever bytes are handed
// in as a PNG, nothing panics, a refusal is one of the library's kinds, and
// what is kept is one byte a pixel, within the cap, of classes that exist.
func FuzzImage(f *testing.F) {
	table := []TableEntry{{Colour: colour.RGB{G: 200}, Value: 15}, {Colour: colour.RGB{R: 250}, Value: 55}, {Colour: colour.RGB{}, Missing: true}}
	f.Add(picturePNG(f, 8, 8, func(x, y int) color.Color { return color.NRGBA{G: 200, A: uint8(32 * x)} }))
	f.Add(picturePNG(f, 3, 5, func(x, y int) color.Color { return color.Gray{Y: uint8(40 * y)} }))
	f.Add(headerOnly(20, 20, 8, 6))
	f.Add(headerOnly(1, 1, 8, 3))
	f.Add([]byte("\x89PNG\r\n\x1a\n and then nothing a decoder wants"))
	f.Fuzz(func(t *testing.T, data []byte) {
		s, err := NewStore(Caps{ImageBytes: 4096})
		if err != nil {
			t.Fatal(err)
		}
		o := Overlay{ID: "radar", Valid: noon, Keeps: time.Hour, Image: &Image{PNG: data, West: -90, South: 30, East: -80, North: 40,
			Projection: PlateCarree, Table: table, Type: Type{Preset: "radar", Unit: "dBZ"}}}
		if _, err := s.HandIn(o); err != nil {
			ownKind(t, err)
			return
		}
		if err := s.PrepareJob("radar", 6).Run(context.Background()); err != nil {
			ownKind(t, err)
			if _, _, kept := s.Raster("radar"); kept {
				t.Fatal("a raster was kept from an image that failed")
			}
			return
		}
		raster, report, ok := s.Raster("radar")
		if !ok || len(raster.Classes) != raster.Width*raster.Height || len(raster.Classes) > 4096 {
			t.Fatalf("%dx%d with %d class bytes, kept=%v", raster.Width, raster.Height, len(raster.Classes), ok)
		}
		for _, c := range raster.Classes {
			if c < NoData || int(c) >= raster.ClassCount {
				t.Fatalf("class %d of %d", c, raster.ClassCount)
			}
		}
		if report.Unmatched < 0 || report.Unmatched > len(raster.Classes) || len(report.Samples) > 16 {
			t.Fatalf("%+v", report)
		}
	})
}

// FuzzTable is the other part: any table at all is refused with one of the
// library's kinds or is usable, and a usable one gives every colour a class
// that exists.
func FuzzTable(f *testing.F) {
	f.Add([]byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0x24, 0x40, 0, 9, 9, 9, 0, 0, 0, 0, 0, 0, 0x34, 0x40, 1})
	f.Add([]byte("a table of nothing in particular, long enough for three entries or so"))
	f.Fuzz(func(t *testing.T, data []byte) {
		var table []TableEntry
		for len(data) >= 12 && len(table) < 300 {
			table = append(table, TableEntry{Colour: colour.RGB{R: data[0], G: data[1], B: data[2]},
				Value: math.Float64frombits(binary.LittleEndian.Uint64(data[3:11])), Missing: data[11]&1 == 1})
			data = data[12:]
		}
		if err := CheckTable(table); err != nil {
			ownKind(t, err)
			return
		}
		m := &matcher{table: table, breaks: []float64{10, 20, 30, 40, 50, 60}, tolerance: 10, known: map[colour.RGB]int8{}, sampled: map[colour.RGB]bool{}}
		for _, c := range []color.NRGBA{{A: 255}, {R: 255, G: 255, B: 255, A: 255}, {R: 1, G: 2, B: 3, A: 255}, {R: 9, G: 9, B: 9, A: 0}, {R: table[0].Colour.R, G: table[0].Colour.G, B: table[0].Colour.B, A: 255}} {
			if got := m.pixel(c); got < NoData || got > 6 {
				t.Fatalf("colour %v is class %d", c, got)
			}
		}
	})
}
