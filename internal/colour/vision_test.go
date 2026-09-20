package colour

import (
	"math"
	"testing"
)

func near(a, b, within float64) bool { return math.Abs(a-b) <= within }

// TestSimulation is plan task 08.6 (D-88): the published matrices for the
// three deficiencies at full severity, the published Lab of a primary, and
// differences held to the figures PLAN's specimens were measured with.
func TestSimulation(t *testing.T) {
	red := RGB{255, 0, 0}
	if lab := InLab(red, Normal); !near(lab.L, 53.2408, 0.001) || !near(lab.A, 80.0925, 0.001) || !near(lab.B, 67.2032, 0.001) {
		t.Errorf("pure red in Lab under D65: %+v", lab)
	}
	if lab := InLab(red, Protanopia); !near(lab.L, 40.2912, 0.001) || !near(lab.A, -3.6738, 0.001) || !near(lab.B, 47.4696, 0.001) {
		t.Errorf("pure red as protanopia sees it: %+v; red loses nearly all of its redness", lab)
	}
	cases := []struct {
		a, b RGB
		want [4]float64 // normal, protanopia, deuteranopia, tritanopia
	}{
		{RGB{255, 0, 0}, RGB{0, 255, 0}, [4]float64{170.5652, 65.8748, 27.7523, 153.2419}},
		{RGB{255, 128, 80}, RGB{112, 36, 28}, [4]float64{49.4116, 45.9366, 49.3935, 40.2847}},
		{RGB{34, 0, 68}, RGB{0, 34, 119}, [4]float64{20.397, 18.513, 21.9234, 22.7406}},
		{RGB{16, 22, 28}, RGB{245, 245, 240}, [4]float64{89.7903, 89.3676, 90.0632, 89.5374}},
		{RGB{240, 232, 144}, RGB{255, 204, 102}, [4]float64{21.9614, 14.027, 14.7018, 16.5747}},
	}
	for _, c := range cases {
		worst := math.Inf(1)
		for i, v := range []Vision{Normal, Protanopia, Deuteranopia, Tritanopia} {
			got := InLab(c.a, v).distance(InLab(c.b, v))
			if !near(got, c.want[i], 0.001) {
				t.Errorf("%v and %v under %v: %.4f, want %.4f", c.a, c.b, v, got, c.want[i])
			}
			worst = math.Min(worst, c.want[i])
		}
		if got := Difference(c.a, c.b); !near(got, worst, 0.001) {
			t.Errorf("%v and %v: difference %.4f, want the worst of the four, %.4f", c.a, c.b, got, worst)
		}
	}
	if Difference(red, red) != 0 {
		t.Error("a colour differs from itself")
	}
	for _, v := range []Vision{Normal, Protanopia, Deuteranopia, Tritanopia, Vision(9)} {
		if v.String() == "" {
			t.Errorf("vision %d has no name", v)
		}
	}
	if InLab(red, Vision(9)) != InLab(red, Normal) {
		t.Error("a kind of vision that is none is read as normal")
	}
}

// TestSimulationIsInLinearLight (D-88): the same two colours score very
// differently depending on where the simulation is applied. The matrices
// were fitted in linear light; applied to gamma-encoded values they say two
// dark reds of the temperature scale are 6.8 apart when they are 27.1.
func TestSimulationIsInLinearLight(t *testing.T) {
	a, b := RGB{170, 0, 17}, RGB{85, 0, 17}
	if got := Difference(a, b); !near(got, 27.0612, 0.001) {
		t.Errorf("in linear light: %.4f, want 27.0612", got)
	}
	gamma := math.Inf(1)
	for _, v := range []Vision{Normal, Protanopia, Deuteranopia, Tritanopia} {
		gamma = math.Min(gamma, inLabGammaSpace(a, v).distance(inLabGammaSpace(b, v)))
	}
	if !near(gamma, 6.8087, 0.001) {
		t.Errorf("the wrong way round gives %.4f here; the test no longer shows the trap", gamma)
	}
}

// inLabGammaSpace is the mistake: the matrix applied to gamma-encoded values.
func inLabGammaSpace(c RGB, v Vision) Lab {
	ch := [3]float64{float64(c.R) / 255, float64(c.G) / 255, float64(c.B) / 255}
	if m, ok := matrix(v); ok {
		ch = apply(m, ch)
	}
	return toLab([3]float64{linearOf(ch[0]), linearOf(ch[1]), linearOf(ch[2])})
}
