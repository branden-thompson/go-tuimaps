package render

import (
	"math"
	"strconv"

	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// A vector field is drawn as arrows (L-16, FR-8; watchpost D-109): braille
// strokes on an even spacing, each pointing where its wind blows TO, its
// length and its colour by speed, every other one labelled with its speed.
// There is no fill, so it lies over radar and over a field's faint bands.
const (
	arrowStepX   = 16 // dots between arrows across: eight cells
	arrowStepY   = 16 // and down: four rows - square on the screen
	arrowMin     = 4  // the calmest arrow's length in dots
	arrowGrow    = 2  // and what each stronger class adds
	arrowHead    = 3  // each stroke of the head, in dots
	arrowHeadDeg = 35 // the head's strokes, either side of the shaft
)

// arrows draws a vector field.
func (r *Renderer) arrows(in Input, f *scene.Field) {
	if !r.axes(in.View) {
		return
	}
	w, h := r.painter.lines.Dots()
	row := 0
	for y := arrowStepY / 2; y < h; y += arrowStepY {
		shift := (row % 2) * arrowStepX / 2 // every other row staggered, so the arrows do not stand in columns
		col := 0
		for x := arrowStepX/2 + shift; x < w; x += arrowStepX {
			if i, ok := fieldIndex(f, r.lons[x], r.lats[y]); ok {
				r.arrow(f, i, x, y, (row+col)%2 == 0)
			}
			col++
		}
		row++
	}
}

// arrow draws one arrow centred on a dot, and its speed when labelled.
func (r *Renderer) arrow(f *scene.Field, i, x, y int, labelled bool) {
	speed, from, class := f.Speeds[i], f.From[i], f.Classes[i]
	if math.IsNaN(speed) || math.IsNaN(from) || class < 0 {
		return // no wind known here: nothing is drawn, never a calm arrow
	}
	ink := classInk(f.Preset, class)
	length := float64(arrowMin + arrowGrow*int(class))
	to := (from + 180) * math.Pi / 180    // where it blows to, clockwise from north
	dx, dy := math.Sin(to), -math.Cos(to) // north is up the screen
	tipX, tipY := float64(x)+dx*length/2, float64(y)+dy*length/2
	r.stroke(float64(x)-dx*length/2, float64(y)-dy*length/2, tipX, tipY, ink)
	for _, side := range []float64{-1, 1} {
		back := to + math.Pi + side*arrowHeadDeg*math.Pi/180
		r.stroke(tipX, tipY, tipX+math.Sin(back)*arrowHead, tipY-math.Cos(back)*arrowHead, ink)
	}
	if labelled && len(r.painter.bandLabels) < maxLabels {
		r.painter.bandLabels = append(r.painter.bandLabels, Label{X: x + 3, Y: y + 2,
			Name: textsafe.Clean(strconv.Itoa(int(math.Round(speed)))), Ink: ink})
	}
}

// stroke sets the dots along a line, half a dot at a time.
func (r *Renderer) stroke(x0, y0, x1, y1 float64, ink uint8) {
	n := int(math.Ceil(math.Max(math.Abs(x1-x0), math.Abs(y1-y0))*2)) + 1
	for k := 0; k <= n; k++ {
		t := float64(k) / float64(n)
		r.painter.lines.Set(int(math.Round(x0+(x1-x0)*t)), int(math.Round(y0+(y1-y0)*t)), ink)
	}
}

// fieldIndex is the cell of a field's grid a position falls in, rows from
// the north, and false outside it.
func fieldIndex(f *scene.Field, lon, lat float64) (int, bool) {
	if f == nil || f.Cols <= 0 || f.Rows <= 0 || !(lon >= f.West && lon < f.East && lat > f.South && lat <= f.North) {
		return 0, false
	}
	col := int((lon - f.West) / (f.East - f.West) * float64(f.Cols))
	row := int((f.North - lat) / (f.North - f.South) * float64(f.Rows))
	i := min(row, f.Rows-1)*f.Cols + min(col, f.Cols-1)
	return i, i < len(f.Speeds) && i < len(f.From)
}
