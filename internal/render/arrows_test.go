package render

import (
	"math"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/style"
)

// windField is a vector field over the Gulf view, every cell one class, one
// speed and one direction the wind blows from.
func windField(class int8, speed, from float64) scene.Field {
	const cols, rows = 20, 15
	f := scene.Field{West: -110, South: 15, East: -70, North: 45, Cols: cols, Rows: rows, Preset: uint8(colour.Wind), ClassCount: 6,
		Classes: make([]int8, cols*rows), From: make([]float64, cols*rows), Speeds: make([]float64, cols*rows)}
	for i := range f.Classes {
		f.Classes[i], f.From[i], f.Speeds[i] = class, from, speed
	}
	return f
}

// windDots are the dots the wind's inks set near the first arrow's centre,
// and their mean offset from it.
func windDots(t *testing.T, f scene.Field) (n int, dx, dy float64, r *Renderer) {
	t.Helper()
	v := gulfView()
	r, _ = drawn(t, Input{View: v, Style: style.BuiltIn(), Fields: []scene.Field{f}, OverlaysVersion: 1})
	c := r.painter.lines
	cx, cy := arrowStepX/2, arrowStepY/2
	for y := max(cy-arrowStepY/2, 0); y < cy+arrowStepY/2; y++ {
		for x := max(cx-arrowStepX/2, 0); x < cx+arrowStepX/2; x++ {
			if ink := colour.Token(c.ink[y*c.w+x]); ink >= colour.Wind1 && ink <= colour.Wind6 {
				n++
				dx += float64(x - cx)
				dy += float64(y - cy)
			}
		}
	}
	if n > 0 {
		dx, dy = dx/float64(n), dy/float64(n)
	}
	return n, dx, dy, r
}

// TestAnArrowPointsWhereTheWindBlows is L-16.1: the head - where the dots
// gather - lies downwind: a west wind points east, a south wind north.
func TestAnArrowPointsWhereTheWindBlows(t *testing.T) {
	for _, c := range []struct {
		name     string
		from     float64
		east, up bool
	}{{"from the west", 270, true, false}, {"from the east", 90, false, false}, {"from the south", 180, false, true}, {"from the north", 0, false, false}} {
		n, dx, dy, _ := windDots(t, windField(3, 15, c.from))
		if n == 0 {
			t.Fatalf("%s: no arrow drawn", c.name)
		}
		switch {
		case c.from == 270 && dx <= 0.3, c.from == 90 && dx >= -0.3:
			t.Errorf("%s: the arrow's dots lean %.2f across; want downwind", c.name, dx)
		case c.from == 180 && dy >= -0.3, c.from == 0 && dy <= 0.3:
			t.Errorf("%s: the arrow's dots lean %.2f down; want downwind", c.name, dy)
		}
	}
}

// TestNoWindIsNoArrow is L-16.1: a cell with no speed or no direction draws
// nothing - never a calm arrow in its place - and there is never a fill.
func TestNoWindIsNoArrow(t *testing.T) {
	for _, f := range []scene.Field{windField(3, math.NaN(), 270), windField(3, 15, math.NaN())} {
		if n, _, _, _ := windDots(t, f); n != 0 {
			t.Errorf("%d dots drawn where the wind is not known", n)
		}
	}
	_, _, _, r := windDots(t, windField(3, 15, 270))
	for i, c := range r.grid.cells {
		if c.under != 0 {
			t.Fatalf("cell %d is filled: a vector field is its arrows alone", i)
		}
	}
}

// TestAStrongerWindIsALongerArrow is L-16.1: length by speed.
func TestAStrongerWindIsALongerArrow(t *testing.T) {
	calm, _, _, _ := windDots(t, windField(0, 3, 270))
	gale, _, _, _ := windDots(t, windField(5, 45, 270))
	if gale <= calm {
		t.Errorf("a gale's arrow has %d dots and a calm one %d; want the gale longer", gale, calm)
	}
}

// TestEveryOtherArrowCarriesItsSpeed is L-16.1: the speed, rounded, beside
// every other arrow.
func TestEveryOtherArrowCarriesItsSpeed(t *testing.T) {
	_, _, _, r := windDots(t, windField(3, 14.6, 270))
	labels := 0
	for _, l := range r.painter.bandLabels {
		if l.Name.String() == "15" {
			labels++
		}
	}
	w, h := r.painter.lines.Dots()
	arrows := (w / arrowStepX) * (h / arrowStepY)
	if labels == 0 || labels > arrows*2/3 {
		t.Errorf("%d labels for about %d arrows; want every other one", labels, arrows)
	}
}
