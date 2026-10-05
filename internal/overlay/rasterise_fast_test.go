package overlay

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math/rand/v2"
	"testing"
)

// A PICTURE IS READ AS FAST AS ITS LAYOUT ALLOWS, AND READ THE SAME (watchpost
// W14, P-12). Read one pixel at a time through image.Image, a 600x275 radar
// frame cost 2.7 ms and 165,049 allocations - one for every pixel's colour. The
// two layouts a PNG of the radar decodes to, NRGBA and paletted, are read
// straight from their pixels; every other layout keeps the general reading.

// frameOf is a w x h picture in the layout named, filled from seed: clear,
// half-clear and solid pixels of the table's colours and of colours near and
// far from them, so every class, the fallback and the unmatched are met.
func frameOf(t testing.TB, layout string, seed uint64) []byte {
	t.Helper()
	table, _ := TableOf(ProviderMRMS)
	var cols []color.NRGBA
	for _, e := range table.Entries {
		cols = append(cols, color.NRGBA{R: e.Colour.R, G: e.Colour.G, B: e.Colour.B, A: 255})
	}
	r := rand.New(rand.NewPCG(seed, 7))
	pick := func() color.NRGBA {
		switch r.IntN(6) {
		case 0, 1:
			return color.NRGBA{}
		case 2:
			c := cols[r.IntN(len(cols))]
			c.A = uint8(1 + r.IntN(254))
			return c
		case 3:
			return color.NRGBA{R: uint8(r.IntN(256)), G: uint8(r.IntN(256)), B: uint8(r.IntN(256)), A: 255}
		}
		return cols[r.IntN(len(cols))]
	}
	const w, h = 61, 37
	bounds := image.Rect(3, 5, 3+w, 5+h) // an origin away from zero, which a direct read must honour
	var img image.Image
	switch layout {
	case "nrgba":
		n := image.NewNRGBA(bounds)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				n.SetNRGBA(x, y, pick())
			}
		}
		img = n
	case "paletted":
		pal := color.Palette{}
		for range 40 {
			pal = append(pal, pick())
		}
		p := image.NewPaletted(bounds, pal)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				p.SetColorIndex(x, y, uint8(r.IntN(len(pal))))
			}
		}
		img = p
	case "rgba":
		n := image.NewRGBA(bounds)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				n.Set(x, y, pick())
			}
		}
		img = n
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestEveryLayoutIsReadAsTheGeneralReadingReadsIt(t *testing.T) {
	kind := mustKind(t)
	for _, layout := range []string{"nrgba", "paletted", "rgba"} {
		for seed := range uint64(6) {
			file := frameOf(t, layout, seed)
			img := &Image{PNG: file, West: -126, South: 23, East: -65, North: 51, Projection: PlateCarree, Provider: ProviderMRMS,
				Type: Type{Preset: "radar", Unit: "dBZ"}}
			got, gotReport, err := rasterise(img, file, kind, 1<<24)
			if err != nil {
				t.Fatal(err)
			}
			want, wantReport := generalReading(t, img, file, kind)
			if !bytes.Equal(int8Bytes(got.Classes), int8Bytes(want)) {
				t.Errorf("%s, seed %d: the classes differ from the general reading", layout, seed)
			}
			if gotReport.Fallback != wantReport.Fallback || gotReport.Unmatched != wantReport.Unmatched || len(gotReport.Samples) != len(wantReport.Samples) {
				t.Errorf("%s, seed %d: report %+v, the general reading's %+v", layout, seed, gotReport, wantReport)
			}
		}
	}
}

// generalReading is the picture read one pixel at a time through image.Image,
// as rasterise read every layout before.
func generalReading(t *testing.T, img *Image, file []byte, kind Kind) ([]int8, Report) {
	t.Helper()
	decoded, err := png.Decode(bytes.NewReader(file))
	if err != nil {
		t.Fatal(err)
	}
	m := newMatcher(img, kind)
	b := decoded.Bounds()
	out := make([]int8, 0, b.Dx()*b.Dy())
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			out = append(out, m.pixel(color.NRGBAModel.Convert(decoded.At(x, y)).(color.NRGBA)))
		}
	}
	return out, m.report
}

// TestAFrameIsReadWithoutAnAllocationAPixel holds the direct reading to its
// purpose: a radar frame's allocations do not grow with its pixels.
func TestAFrameIsReadWithoutAnAllocationAPixel(t *testing.T) {
	kind := mustKind(t)
	file := frameOf(t, "nrgba", 1)
	img := &Image{PNG: file, West: -126, South: 23, East: -65, North: 51, Projection: PlateCarree, Provider: ProviderMRMS,
		Type: Type{Preset: "radar", Unit: "dBZ"}}
	allocs := testing.AllocsPerRun(5, func() { _, _, _ = rasterise(img, file, kind, 1<<24) })
	if pixels := 61.0 * 37; allocs > pixels/4 {
		t.Errorf("reading a %v-pixel frame allocated %.0f times", pixels, allocs)
	}
}
