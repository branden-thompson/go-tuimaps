package colour

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

var (
	darkGround  = RGB{16, 22, 28}
	lightGround = RGB{245, 245, 240}
)

// specimen reads a ramp from one of PLAN's data files, which were written
// first and which HUM LEAD's looks were taken from.
func specimen(t *testing.T, file string, path ...string) []RGB {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("..", "..", "06_docs", "02_features", "go-tuimaps", "02-analysis", "specimens", file))
	if err != nil {
		t.Fatal(err)
	}
	var node any
	if err := json.Unmarshal(body, &node); err != nil {
		t.Fatal(err)
	}
	for _, key := range path {
		node = node.(map[string]any)[key]
	}
	var out []RGB
	for _, c := range node.([]any) {
		v := c.([]any)
		out = append(out, RGB{uint8(v[0].(float64)), uint8(v[1].(float64)), uint8(v[2].(float64))})
	}
	return out
}

// rainbow is a many-hued scale in the style of broadcast maps (S21-2).
func rainbow() []RGB {
	return []RGB{{145, 0, 160}, {90, 0, 200}, {0, 0, 255}, {0, 110, 255}, {0, 190, 255}, {0, 230, 200}, {0, 200, 90}, {60, 220, 0},
		{170, 240, 0}, {255, 255, 0}, {255, 200, 0}, {255, 140, 0}, {255, 70, 0}, {230, 0, 0}, {255, 0, 180}}
}

func count(findings []Finding, rule Rule) int {
	n := 0
	for _, f := range findings {
		if f.Rule == rule {
			n++
		}
	}
	return n
}

// Plan task 08.7 (D-53): each rule with a ramp that passes and one that fails.
func TestCheckerOrdered(t *testing.T) {
	temperature := specimen(t, "21-temperature-scale-candidate.json", "truecolor")
	// The break at 0 C lies between classes 6 and 7, both pale; class 6, just
	// below freezing, is the lightest, and luminance falls from it.
	if f := CheckRamp(temperature, RampCheck{Ground: darkGround, Midpoint: 6}); count(f, Ordered) != 0 {
		t.Errorf("the temperature scale, lightest at freezing: %+v", f)
	}
	light := specimen(t, "21-temperature-scale-candidate.json", "on_a_light_ground", "truecolor")
	// On a light ground three bands differ (D-91), and the lightest is class 7,
	// just above freezing: each variant carries its own midpoint.
	if f := CheckRamp(light, RampCheck{Ground: lightGround, Midpoint: 7}); count(f, Ordered) != 0 {
		t.Errorf("the light-ground temperature scale: %+v", f)
	}
	if f := CheckRamp(temperature, RampCheck{Ground: darkGround, Midpoint: -1}); count(f, Ordered) == 0 {
		t.Error("the temperature scale read as one-way passed; it rises and then falls")
	}
	radar := specimen(t, "22-radar-ramp-candidate.json", "truecolor")
	if f := CheckRamp(radar, RampCheck{Ground: darkGround, Midpoint: -1}); count(f, Ordered) != 0 {
		t.Errorf("the radar ramp, lighter at every step: %+v", f)
	}
	first := specimen(t, "22-radar-ramp-candidate.json", "on_a_light_ground", "first_version", "truecolor")
	if f := CheckRamp(first, RampCheck{Ground: lightGround, Midpoint: -1}); count(f, Ordered) != 1 {
		t.Errorf("the light-ground radar ramp as first filed got lighter from class 2 to 3: %+v", f)
	}
	if f := CheckRamp(rainbow(), RampCheck{Ground: darkGround, Midpoint: -1}); count(f, Ordered) == 0 {
		t.Error("the broadcast-style scale passed as ordered")
	}
}

func TestCheckerVisionSafe(t *testing.T) {
	for _, c := range []struct {
		name   string
		ramp   []RGB
		ground RGB
		depth  Depth
	}{
		{"temperature, dark", specimen(t, "21-temperature-scale-candidate.json", "truecolor"), darkGround, Truecolor},
		{"temperature, light", specimen(t, "21-temperature-scale-candidate.json", "on_a_light_ground", "truecolor"), lightGround, Truecolor},
		{"radar, dark", specimen(t, "22-radar-ramp-candidate.json", "truecolor"), darkGround, Truecolor},
		{"radar, light", specimen(t, "22-radar-ramp-candidate.json", "on_a_light_ground", "truecolor"), lightGround, Truecolor},
	} {
		if f := CheckRamp(c.ramp, RampCheck{Ground: c.ground, Midpoint: -1, Depth: c.depth}); count(f, VisionSafe) != 0 {
			t.Errorf("%s: %+v", c.name, f)
		}
	}
	// The broadcast-style scale fails three ways (S21-2): not ordered, a pair
	// of neighbours too close, and a pair that are not neighbours too close.
	findings := CheckRamp(rainbow(), RampCheck{Ground: darkGround, Midpoint: -1})
	neighbours, others := 0, 0
	for _, f := range findings {
		if f.Rule != VisionSafe || f.B < 0 {
			continue
		}
		if f.B-f.A == 1 {
			neighbours++
		} else {
			others++
		}
		if f.Value >= 10 {
			t.Errorf("%+v is reported as too close", f)
		}
	}
	if count(findings, Ordered) == 0 || neighbours == 0 || others == 0 {
		t.Errorf("ordered %d, neighbours %d, others %d; the broadcast-style scale fails all three ways", count(findings, Ordered), neighbours, others)
	}
}

// TestCheckerAllPairs and TestCheckerGroundRule are plan task 08.18 (D-88,
// PL-AX-1, PL-AX-2). The temperature scale as first filed is not kept; this
// ramp has its flaw: neighbours well apart, and two classes that are not
// neighbours nearly the same.
func TestCheckerAllPairs(t *testing.T) {
	ramp := []RGB{{240, 232, 144}, {255, 136, 68}, {244, 230, 150}, {85, 0, 17}}
	findings := CheckRamp(ramp, RampCheck{Ground: darkGround, Midpoint: -1})
	found := false
	for _, f := range findings {
		if f.Rule == VisionSafe && f.A == 0 && f.B == 2 {
			found = true
		}
		if f.Rule == VisionSafe && f.B-f.A == 1 {
			t.Errorf("neighbours reported: %+v; the case is about classes that are not", f)
		}
	}
	if !found {
		t.Errorf("classes 0 and 2 are nearly one colour and passed: %+v", findings)
	}
}

func TestCheckerGroundRule(t *testing.T) {
	dark := specimen(t, "22-radar-ramp-candidate.json", "truecolor")
	findings := CheckRamp(dark, RampCheck{Ground: lightGround, Midpoint: -1})
	worst := Finding{Value: 1e9}
	for _, f := range findings {
		if f.Rule == VisionSafe && f.B == -1 && f.Value < worst.Value {
			worst = f
		}
	}
	if worst.A != 5 || worst.Value < 3.0 || worst.Value > 3.2 {
		t.Errorf("%+v; on a light ground the dark-ground ramp's heaviest class is 3.1 from the ground", worst)
	}
}

func TestCheckerReadable(t *testing.T) {
	outlines := []RGB{{255, 128, 80}, {255, 214, 90}}
	if f := CheckRamp(outlines, RampCheck{Ground: darkGround, Midpoint: -1, Lines: true}); count(f, Readable) != 0 {
		t.Errorf("two bright outlines on the dark ground: %+v", f)
	}
	if f := CheckRamp(outlines, RampCheck{Ground: lightGround, Midpoint: -1, Lines: true}); count(f, Readable) != 2 {
		t.Errorf("the same outlines on a light ground both fail 3:1 (1.3 and 2.3, constants section 4): %+v", f)
	}
	// An area ramp is read through the text drawn on it, which the foreground
	// rule picks: every colour takes black or white at 4.5:1 or better.
	if f := CheckRamp(rainbow(), RampCheck{Ground: darkGround, Midpoint: -1}); count(f, Readable) != 0 {
		t.Errorf("%+v", f)
	}
}

func TestCheckerDistinct(t *testing.T) {
	close := []RGB{{10, 60, 120}, {14, 64, 124}, {200, 200, 60}}
	if f := CheckRamp(close, RampCheck{Ground: darkGround, Midpoint: -1, Depth: Truecolor}); count(f, Distinct) != 0 {
		t.Errorf("three different colours at truecolor: %+v", f)
	}
	if f := CheckRamp(close, RampCheck{Ground: darkGround, Midpoint: -1, Depth: Colours256}); count(f, Distinct) != 1 {
		t.Errorf("two colours that fall on one entry of the 256 palette: %+v", f)
	}
	same := []RGB{{1, 2, 3}, {1, 2, 3}}
	if f := CheckRamp(same, RampCheck{Ground: darkGround, Midpoint: -1}); count(f, Distinct) != 1 {
		t.Errorf("one colour twice: %+v", f)
	}
	if f := CheckRamp(nil, RampCheck{}); len(f) != 0 {
		t.Errorf("no ramp: %+v", f)
	}
	for r := Ordered; r <= VisionSafe; r++ {
		if r.String() == "" || r.String() == "unknown" {
			t.Errorf("rule %d has no name", r)
		}
	}
}

// Test256UsesOnlyFixedIndices is plan task 08.13: indices 16 to 255, whose
// colours every terminal agrees on; never 0 to 15, which a theme repaints.
func Test256UsesOnlyFixedIndices(t *testing.T) {
	for _, c := range []RGB{{0, 0, 0}, {255, 255, 255}, {255, 0, 0}, {128, 128, 128}, {12, 34, 56}, {250, 128, 114}, {8, 8, 8}, {238, 238, 238}} {
		index, shown := To256(c)
		if index < 16 {
			t.Errorf("%v maps to index %d; only 16 to 255 are fixed", c, index)
		}
		if back, _ := To256(shown); back != index {
			t.Errorf("%v: index %d shows as %v, which maps to %d", c, index, shown, back)
		}
	}
	if i, shown := To256(RGB{0, 0, 95}); i != 17 || shown != (RGB{0, 0, 95}) {
		t.Errorf("a colour of the cube maps to itself: %d, %v", i, shown)
	}
	if i, shown := To256(RGB{128, 128, 128}); i != 244 || shown != (RGB{128, 128, 128}) {
		t.Errorf("a grey of the ramp maps to itself: %d, %v", i, shown)
	}
	for _, ramp := range [][]RGB{specimen(t, "21-temperature-scale-candidate.json", "colours_256_palette"), specimen(t, "22-radar-ramp-candidate.json", "colours_256_palette")} {
		for _, c := range ramp {
			if _, shown := To256(c); shown != c {
				t.Errorf("%v is filed as a 256-palette colour and is not one: nearest is %v", c, shown)
			}
		}
	}
}
