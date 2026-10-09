package colour

import (
	"fmt"
	"testing"
)

// TestTheMUFAndFoF2RampsPassCheckOnBothGrounds is v0.3.0 L-2.1 (watchpost
// D-144, D-145, D-149): each ramp has a colour for every class it draws - the
// MUF one more than its breaks, foF2 one a floor - its breaks rise, and it
// passes the checker on each ground at truecolor and at 256 colours,
// colour-vision safe. Each foF2 class is the MUF colour of its frequencies.
func TestTheMUFAndFoF2RampsPassCheckOnBothGrounds(t *testing.T) {
	for _, p := range []struct {
		preset Preset
		breaks []float64
	}{{MUF, MUFBreaks()}, {FoF2, FoF2Breaks()}} {
		for i := 1; i < len(p.breaks); i++ {
			if p.breaks[i] <= p.breaks[i-1] {
				t.Fatalf("%v: breaks %v do not rise", p.preset, p.breaks)
			}
		}
		for _, ground := range []GroundKind{Dark, Light} {
			under, _ := Palette{}.ResolveAt(Ground, ground, Truecolor)
			for _, depth := range []Depth{Truecolor, Colours256} {
				ramp, ok := Ramp(p.preset, ground, depth)
				drawn := len(p.breaks) + 1
				if FloorsFirst(p.preset) {
					drawn = len(p.breaks)
				}
				if !ok || len(ramp) != drawn {
					t.Fatalf("%v %v %v: %d classes for %d breaks", p.preset, ground, depth, len(ramp), len(p.breaks))
				}
				if f := Check(ramp, RampCheck{Ground: under, Depth: depth, Midpoint: -1}); len(f) != 0 {
					t.Errorf("%v %v %v: %+v", p.preset, ground, depth, f)
				}
			}
		}
	}
	muf, fof2 := mufRamp(), fof2Ramp()
	for i, c := range fof2 { // fof2.1 is 1.8-3.5 MHz, MUF's class below 3.5; and on, band for band
		if c != muf[i] {
			t.Errorf("foF2 class %d is %v; D-149 gives it the MUF colour of the same frequencies, %v", i+1, c, muf[i])
		}
	}
	if !FloorsFirst(FoF2) {
		t.Error("foF2 draws below 1.8 MHz; D-149 rules nothing there")
	}
	if b, want := MUFBreaks(), []float64{3.5, 5.3, 7, 10.1, 14, 18.068, 21, 24.89, 28}; fmt.Sprint(b) != fmt.Sprint(want) {
		t.Errorf("the MUF breaks are %v; D-144 rules the amateur band edges %v", b, want)
	}
	if b, want := FoF2Breaks(), []float64{1.8, 3.5, 5.3, 7, 10.1, 14}; fmt.Sprint(b) != fmt.Sprint(want) {
		t.Errorf("the foF2 breaks are %v; D-145 rules the band edges below 15 MHz %v", b, want)
	}
}
