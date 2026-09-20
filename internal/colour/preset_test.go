package colour

import (
	"math"
	"testing"
)

func sameRamp(a, b []RGB) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// lightnessSteps lists the steps along a ramp, between from and to, where
// lightness under a kind of vision does not move the way it should.
func lightnessSteps(ramp []RGB, v Vision, from, to int, rising bool) [][2]int {
	var wrong [][2]int
	for i := from; i < to; i++ {
		step := InLab(ramp[i+1], v).L - InLab(ramp[i], v).L
		if (rising && step <= 0) || (!rising && step >= 0) {
			wrong = append(wrong, [2]int{i, i + 1})
		}
	}
	return wrong
}

func groundOf(kind GroundKind, depth Depth) RGB {
	var none Palette
	c, _ := none.Resolve(Ground, kind)
	if depth == Colours256 {
		_, c = To256(c)
	}
	return c
}

// TestTemperaturePresetPasses is plan task 08.8 (D-62, D-88, D-91).
func TestTemperaturePresetPasses(t *testing.T) {
	file := "21-temperature-scale-candidate.json"
	filed := map[GroundKind]map[Depth][]RGB{
		Dark:  {Truecolor: specimen(t, file, "truecolor"), Colours256: specimen(t, file, "colours_256_palette")},
		Light: {Truecolor: specimen(t, file, "on_a_light_ground", "truecolor"), Colours256: specimen(t, file, "on_a_light_ground", "colours_256_palette")},
	}
	for kind, byDepth := range filed {
		for depth, want := range byDepth {
			ramp, ok := Ramp(Temperature, kind, depth)
			if !ok || len(ramp) != 17 || !sameRamp(ramp, want) {
				t.Fatalf("%v, %v: the preset is not the ramp HUM LEAD was shown (specimen 21)", kind, depth)
			}
			mid := Midpoint(Temperature, kind, depth)
			if f := Check(ramp, RampCheck{Ground: groundOf(kind, depth), Depth: depth, Midpoint: mid}); len(f) != 0 {
				t.Errorf("%v, %v: %+v", kind, depth, f)
			}
			// Lightness, not only luminance, rises to the pale class and falls
			// from it under every kind of vision - in truecolor. At 256 colours
			// the palette is too coarse: the known inversions are recorded.
			inversions := 0
			for v := Normal; v <= Tritanopia; v++ {
				inversions += len(lightnessSteps(ramp, v, 0, mid, true)) + len(lightnessSteps(ramp, v, mid, 16, false))
			}
			if depth == Truecolor && inversions != 0 {
				t.Errorf("%v truecolor: %d steps where lightness goes the wrong way", kind, inversions)
			}
			// Recorded, not gated (constants, section 4): classes 1 to 2 under
			// protanopia and tritanopia, 3 to 4 under tritanopia; and on the dark
			// ground the two pale classes, 6 to 7, under protanopia, by 0.4.
			// Held to the exact count so that a change to a ramp is noticed.
			if known := map[GroundKind]int{Dark: 4, Light: 3}[kind]; depth == Colours256 && inversions != known {
				t.Errorf("%v at 256 colours: %d lightness inversions, where %d are known and recorded", kind, inversions, known)
			}
		}
	}
	dark, _ := Ramp(Temperature, Dark, Truecolor)
	light, _ := Ramp(Temperature, Light, Truecolor)
	var differ []int
	for i := range dark {
		if dark[i] != light[i] {
			differ = append(differ, i)
		}
	}
	if len(differ) != 3 || differ[0] != 5 || differ[2] != 7 {
		t.Errorf("the two grounds differ in classes %v; D-91 changed three bands, 5 to 7", differ)
	}
	if _, ok := Ramp(Temperature, Dark, NoColour); ok {
		t.Error("with no colour there is no ramp: the no-colour form is drawn")
	}
}

// TestBreaksExactInFahrenheit: the breaks are defined once, in Celsius, and
// converted exactly.
func TestBreaksExactInFahrenheit(t *testing.T) {
	c, f := TemperatureBreaks(Celsius), TemperatureBreaks(Fahrenheit)
	if len(c) != 16 || c[0] != -30 || c[15] != 45 || c[6] != 0 {
		t.Fatalf("Celsius breaks %v", c)
	}
	for i := range c {
		if f[i] != c[i]*9/5+32 || f[i] != math.Trunc(f[i]) {
			t.Errorf("break %d: %v C is %v F", i, c[i], f[i])
		}
	}
	if f[6] != 32 || f[0] != -22 || f[15] != 113 {
		t.Errorf("Fahrenheit breaks %v", f)
	}
	c[0] = 999
	if TemperatureBreaks(Celsius)[0] != -30 {
		t.Error("a caller changed the library's breaks")
	}
}

// TestRadarPresetPasses is plan task 08.9 (D-69, D-88, PL-AX-1).
func TestRadarPresetPasses(t *testing.T) {
	file := "22-radar-ramp-candidate.json"
	filed := map[GroundKind]map[Depth][]RGB{
		Dark:  {Truecolor: specimen(t, file, "truecolor"), Colours256: specimen(t, file, "colours_256_palette")},
		Light: {Truecolor: specimen(t, file, "on_a_light_ground", "truecolor"), Colours256: specimen(t, file, "on_a_light_ground", "colours_256_palette")},
	}
	for kind, byDepth := range filed {
		for depth, want := range byDepth {
			ramp, ok := Ramp(Radar, kind, depth)
			if !ok || len(ramp) != 6 || !sameRamp(ramp, want) {
				t.Fatalf("%v, %v: the preset is not the ramp HUM LEAD was shown (specimen 22)", kind, depth)
			}
			if f := Check(ramp, RampCheck{Ground: groundOf(kind, depth), Depth: depth, Midpoint: -1}); len(f) != 0 {
				t.Errorf("%v, %v: %+v", kind, depth, f)
			}
			// Lighter means heavier on a dark ground, darker on a light one -
			// at every step, under every kind of vision, at both depths.
			for v := Normal; v <= Tritanopia; v++ {
				if wrong := lightnessSteps(ramp, v, 0, 5, kind == Dark); len(wrong) != 0 {
					t.Errorf("%v, %v, %v: lightness goes the wrong way at %v", kind, depth, v, wrong)
				}
			}
		}
	}
	floors := RadarFloors()
	if len(floors) != 6 || floors[0] != 10 || floors[5] != 60 {
		t.Errorf("floors %v", floors)
	}
	if Midpoint(Radar, Dark, Truecolor) != -1 {
		t.Error("the radar ramp runs one way")
	}
}

// TestRadarRampFollowsDeclaredGround: the ramp is chosen by the ground in
// effect - painted, or the colour the host declared.
func TestRadarRampFollowsDeclaredGround(t *testing.T) {
	var none Palette
	white, _ := Declared(RGB{255, 255, 255})
	dark, _ := Ramp(Radar, Dark, Truecolor)
	light, _ := Ramp(Radar, Light, Truecolor)
	got, _ := none.ResolveAt(Radar1+5, white.Kind(none), Truecolor)
	if got != light[5] {
		t.Errorf("heaviest rain on a declared white ground: %v, want the light ramp's %v", got, light[5])
	}
	var painted GroundChoice
	if got, _ := none.ResolveAt(Radar1, painted.Kind(none), Truecolor); got != dark[0] {
		t.Errorf("lightest rain on the painted dark ground: %v", got)
	}
	if got, _ := none.ResolveAt(Temperature1+8, Dark, Colours256); got != (RGB{215, 215, 135}) {
		t.Errorf("temperature.9 at 256 colours: %v; the library's own ramp for the depth, never a downgrade", got)
	}
	host, _ := NewPalette(map[string]RGB{"radar.6": {1, 2, 3}})
	if got, _ := host.ResolveAt(Radar6, Dark, Truecolor); got != (RGB{1, 2, 3}) {
		t.Errorf("a host's colour for a ramp token: %v", got)
	}
	if got, ok := none.Resolve(Radar6, Dark); !ok || got != dark[5] {
		t.Errorf("Resolve gives the truecolor default: %v, %v", got, ok)
	}
}
