package describe

import (
	"math"

	"github.com/branden-thompson/go-tuimaps/internal/project"
)

// Grid is a scalar field as the host handed it in: values on a regular grid
// of longitude and latitude, rows from the north, each row west to east.
type Grid struct {
	West, South, East, North float64
	Cols, Rows               int
	Values                   []float64
}

// Reading is what a field says at a place.
type Reading struct {
	Value  float64
	Band   int     // which class it falls in, by the breaks the overlay carries
	Rises  float64 // the bearing the values rise fastest along
	Rising bool    // false on a flat day, and then Rises means nothing (D-68)
	NoData bool    // the place is outside the grid, or the value is not a number
}

// At is the field's reading at a place: the value of the sample the place
// falls in, which class it is in, and which way the field rises.
//
// **The value is the data's, not the picture's.** The renderer samples a
// grid into cells and a cell may be kilometres across; this answers from
// the sample itself, which is the whole reason the description exists.
func (g Grid) At(at project.LonLat, breaks []float64) Reading {
	col, row, ok := g.cell(at)
	if !ok {
		return Reading{NoData: true}
	}
	value, ok := g.value(col, row)
	if !ok {
		return Reading{NoData: true}
	}
	out := Reading{Value: value, Band: band(value, breaks)}
	out.Rises, out.Rising = g.rise(col, row, value)
	return out
}

// cell is the sample a place falls in, or false for a place outside the
// grid. Longitude is circular, so a grid that crosses the seam is asked
// about in its own span rather than in the world's numbering.
func (g Grid) cell(at project.LonLat) (int, int, bool) {
	if g.Cols <= 0 || g.Rows <= 0 || !onGlobe(at) {
		return 0, 0, false
	}
	span := eastward(g.West, g.East)
	if span <= 0 {
		span += 360 // a grid that crosses the seam spans the long way round
	}
	east := eastward(g.West, at.Lon)
	if east < 0 {
		east += 360
	}
	if east > span || at.Lat < g.South || at.Lat > g.North {
		return 0, 0, false
	}
	col := int(east / span * float64(g.Cols))
	row := int((g.North - at.Lat) / (g.North - g.South) * float64(g.Rows))
	return min(col, g.Cols-1), min(row, g.Rows-1), true
}

// value is one sample, or false for one that is not a number.
func (g Grid) value(col, row int) (float64, bool) {
	at := row*g.Cols + col
	if at < 0 || at >= len(g.Values) {
		return 0, false
	}
	v := g.Values[at]
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, false
	}
	return v, true
}

// rise is the bearing along which the values rise fastest from one sample,
// and whether they rise at all. On a flat day nothing rises, and saying so
// is the answer (D-68).
func (g Grid) rise(col, row int, here float64) (float64, bool) {
	east, eastOK := g.step(col+1, row, here)
	west, westOK := g.step(col-1, row, here)
	north, northOK := g.step(col, row-1, here)
	south, southOK := g.step(col, row+1, here)
	x, y := 0.0, 0.0
	if eastOK {
		x += east
	}
	if westOK {
		x -= west
	}
	if northOK {
		y += north
	}
	if southOK {
		y -= south
	}
	if x == 0 && y == 0 {
		return 0, false
	}
	bearing := math.Atan2(x, y) * 180 / math.Pi
	return math.Mod(bearing+360, 360), true
}

// step is how much higher a neighbouring sample is than this one, or false
// where there is no neighbour.
func (g Grid) step(col, row int, here float64) (float64, bool) {
	if col < 0 || row < 0 || col >= g.Cols || row >= g.Rows {
		return 0, false
	}
	there, ok := g.value(col, row)
	if !ok {
		return 0, false
	}
	return there - here, true
}

// band is the class a value falls in: the number of breaks at or below it,
// which is the same rule the renderer colours by, so the words and the
// picture can never disagree.
func band(value float64, breaks []float64) int {
	n := 0
	for _, b := range breaks {
		if value >= b {
			n++
		}
	}
	return n
}

// Image is a classified image as the renderer draws it: one class a pixel,
// rows from the north, with the box it covers.
type Image struct {
	West, South, East, North float64
	Width, Height            int
	Classes                  []int8 // -1 is no data
}

// Heavier is the nearest pixel of a heavier class than the one at a place:
// where the rain gets worse, and how far away that is.
type Heavier struct {
	Class   int
	Km      float64
	Bearing float64
	At      project.LonLat
	Found   bool
}

// ClassAt is the class of the pixel a place falls in, and whether there is
// data there at all.
func (i Image) ClassAt(at project.LonLat) (int, bool) {
	x, y, ok := i.pixel(at)
	if !ok {
		return 0, false
	}
	class := i.Classes[y*i.Width+x]
	if class < 0 {
		return 0, false
	}
	return int(class), true
}

// pixel is the pixel a place falls in.
func (i Image) pixel(at project.LonLat) (int, int, bool) {
	grid := Grid{West: i.West, South: i.South, East: i.East, North: i.North, Cols: i.Width, Rows: i.Height}
	x, y, ok := grid.cell(at)
	if !ok || y*i.Width+x >= len(i.Classes) {
		return 0, 0, false
	}
	return x, y, true
}

// NearestHeavier is the nearest pixel of a heavier class than the one here,
// with how far and which way it is. It is what a host says when the answer
// "light rain" is true and not the whole truth.
func (i Image) NearestHeavier(at project.LonLat) Heavier {
	here, ok := i.ClassAt(at)
	if !ok {
		here = -1 // no data here: anything at all is heavier
	}
	best := Heavier{Km: math.Inf(1)}
	for y := range i.Height {
		for x := range i.Width {
			class := int(i.Classes[y*i.Width+x])
			if class <= here {
				continue
			}
			centre := i.centre(x, y)
			edge, ok := measured(onWorld(at), centre)
			if !ok || edge.Km >= best.Km {
				continue
			}
			best = Heavier{Class: class, Km: edge.Km, Bearing: edge.Bearing, At: centre, Found: true}
		}
	}
	if !best.Found {
		return Heavier{}
	}
	return best
}

// centre is the middle of one pixel, in degrees.
func (i Image) centre(x, y int) project.LonLat {
	span := eastward(i.West, i.East)
	if span <= 0 {
		span += 360
	}
	lon := i.West + span*(float64(x)+0.5)/float64(i.Width)
	lat := i.North - (i.North-i.South)*(float64(y)+0.5)/float64(i.Height)
	return project.LonLat{Lon: wrapped(lon), Lat: lat}
}
