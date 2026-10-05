package overlay

// fallback_test.go — v0.2.0 L6.4-L6.6 and L6.9 (L-2.3): MRMS's heavy end,
// valued along its legend's gradient where the observed palette has no
// colour, counted, warned of, and held to the fallback-share threshold.

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// mrmsImage is an image read with MRMS's table, of one colour all over but
// for a square of another in its corner.
func mrmsImage(t testing.TB, ground, corner color.NRGBA, side int) *Image {
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
	return &Image{PNG: buf.Bytes(), West: -95, South: 30, East: -85, North: 40, Projection: PlateCarree, Provider: ProviderMRMS,
		Type: Type{Preset: "radar", Unit: "dBZ"}}
}

// TestTheHeavyEndIsValuedAlongTheGradient is L6.5: wave 2's colours 3 to 30
// from any single legend colour, taken from yellow to red, are valued in
// order along the legend's heavy-end gradient, each close to it.
func TestTheHeavyEndIsValuedAlongTheGradient(t *testing.T) {
	table, ok := TableOf(ProviderMRMS)
	if !ok || len(table.Gradient) == 0 {
		t.Fatal("MRMS carries no gradient")
	}
	g := gradientOf(table.Gradient)
	last := -1.0
	for _, green := range []uint8{0xd3, 0xce, 0xca, 0xc5, 0xc0, 0xbb, 0x9f, 0x8e, 0x7c, 0x6a, 0x59, 0x47, 0x35, 0x23} {
		v, off := g.value(colour.RGB{R: 0xff, G: green})
		if v <= last {
			t.Errorf("ff%02x00 valued %.2f dBZ, not above the colour before it (%.2f)", green, v, last)
		}
		if off > fallbackReach {
			t.Errorf("ff%02x00 is %.1f from the gradient, beyond the reach of %v", green, off, fallbackReach)
		}
		last = v
	}
}

// TestAFallbackIsCountedAndWarned is L6.4 and L6.6: a colour MRMS was never
// seen to draw but its legend holds - 66 dBZ, a purple - is valued along the
// gradient, counted apart from colours that matched nothing, and warned of
// with the count; a colour far from any rain matches nothing, as before.
func TestAFallbackIsCountedAndWarned(t *testing.T) {
	light := color.NRGBA{R: 183, G: 188, B: 180, A: 255} // observed, -2.99 dBZ
	img := mrmsImage(t, light, color.NRGBA{R: 0xa0, B: 0xf6, A: 255}, 2)
	s := store(t)
	if _, err := s.HandIn(Overlay{ID: "radar", Valid: noon, Keeps: time.Hour, Image: img}); err != nil {
		t.Fatal(err)
	}
	if err := s.PrepareJob("radar", 0).Run(t.Context()); err != nil {
		t.Fatal(err)
	}
	raster, report, ok := s.Raster("radar")
	if !ok {
		t.Fatal("nothing read")
	}
	if report.Fallback != 4 || report.Unmatched != 0 {
		t.Errorf("four pixels of 66 dBZ: %d by the fallback, %d unmatched; want 4 and 0", report.Fallback, report.Unmatched)
	}
	if top := int8(raster.ClassCount - 1); raster.Classes[0] != top {
		t.Errorf("66 dBZ read as class %d; want the heaviest, %d", raster.Classes[0], top)
	}
	found := false
	for _, w := range s.TakeWarnings() {
		if w.Kind == fault.TableFallback && w.Count == 4 {
			found = true
		}
	}
	if !found {
		t.Error("no table-fallback warning with the count")
	}
	cyan, err := copied(mrmsImage(t, light, color.NRGBA{G: 255, B: 255, A: 255}, 2))
	if err != nil {
		t.Fatal(err)
	}
	if _, report, err = rasterise(cyan, cyan.PNG, mustKind(t), defaultImageBytes); err != nil {
		t.Fatal(err)
	}
	if report.Fallback != 0 || report.Unmatched != 4 {
		t.Errorf("cyan, far from any rain: %d by the fallback, %d unmatched; want 0 and 4", report.Fallback, report.Unmatched)
	}
}

func mustKind(t testing.TB) Kind {
	t.Helper()
	k, err := ResolveType(Type{Preset: "radar", Unit: "dBZ"})
	if err != nil {
		t.Fatal(err)
	}
	return k
}

// fallbackShare is how much of a frame's rain was read by the fallback.
func fallbackShare(t testing.TB, img *Image) float64 {
	t.Helper()
	_, report, err := rasterise(img, img.PNG, mustKind(t), 1_048_576)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := png.Decode(bytes.NewReader(img.PNG))
	if err != nil {
		t.Fatal(err)
	}
	rain := 0
	b := decoded.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if color.NRGBAModel.Convert(decoded.At(x, y)).(color.NRGBA).A != 0 {
				rain++
			}
		}
	}
	if rain == 0 {
		return 0
	}
	return float64(report.Fallback) / float64(rain)
}

// fallbackShareLimit is the PLAN report's threshold (D-73): more than this
// share of a frame's rain read by the fallback, and the table is failing it.
const fallbackShareLimit = 0.01

// TestTheFallbackShare is L6.9 (L-2.3): every recorded MRMS frame keeps its
// fallback share within the threshold; IEM's archive frames (OW-11), read
// with IEM's published table, are the control; and a frame whose heavy end
// is mostly off the observed palette is over it. **MRMS's heavy end has no
// oracle before OW-12's capture** (D-72): until then, this test holds the
// frames we have, and cannot say the heavy end is right.
func TestTheFallbackShare(t *testing.T) {
	root := "../../06_docs/02_features/radar-loops/02-analysis/programs/inputs/"
	check := func(pattern string, p Provider) int {
		files, err := filepath.Glob(root + pattern)
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range files {
			body, err := os.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}
			img := &Image{PNG: body, West: -125, South: 24, East: -66, North: 50, Projection: PlateCarree, Provider: p, Type: Type{Preset: "radar", Unit: "dBZ"}}
			own, err := copied(img)
			if err != nil {
				t.Fatal(err)
			}
			if share := fallbackShare(t, own); share > fallbackShareLimit {
				t.Errorf("%s: %.2f %% of its rain by the fallback, over %.0f %%", name, 100*share, 100*fallbackShareLimit)
			}
		}
		return len(files)
	}
	if n := check("mrms/nat/*.png", ProviderMRMS) + check("mrms/national.png", ProviderMRMS) + check("mrms/region.png", ProviderMRMS) + check("mrms/state.png", ProviderMRMS); n != 18 {
		t.Errorf("%d MRMS frames; wave 2 recorded 18", n)
	}
	if n := check("ow11/*.png", ProviderIEM); n == 0 {
		t.Error("no IEM archive frames to hold as the control")
	}
	heavy, _ := copied(mrmsImage(t, color.NRGBA{R: 183, G: 188, B: 180, A: 255}, color.NRGBA{R: 0xa0, B: 0xf6, A: 255}, 6))
	if share := fallbackShare(t, heavy); share <= fallbackShareLimit {
		t.Errorf("a frame with 9 %% of its rain off the palette: %.2f %% by the fallback; the threshold should catch it", 100*share)
	}
}
