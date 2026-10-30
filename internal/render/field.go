package render

import (
	"math"
	"slices"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

const (
	// shadeLight, shadeMedium and shadeHeavy are an image's classes with no
	// colour: block shades, heavier for heavier classes (FR-18).
	shadeLight  = "\u2591"
	shadeMedium = "\u2592"
	shadeHeavy  = "\u2593"
	// labelEvery is how many dot rows apart a contour's value is offered as a label.
	labelEvery = 24
)

// axes are the longitude of every dot column's centre and the latitude of
// every dot row's. The projection keeps the two apart, so a frame needs one
// of each a dot, not one a sample.
func (r *Renderer) axes(v project.View) bool {
	w, h := r.painter.lines.Dots()
	r.lons, r.lats = r.lons[:0], r.lats[:0]
	for x := range w {
		p, err := v.FromDot(float64(x)+0.5, 0.5)
		if err != nil {
			return false
		}
		r.lons = append(r.lons, p.Lon)
	}
	for y := range h {
		p, err := v.FromDot(0.5, float64(y)+0.5)
		if err != nil {
			return false
		}
		r.lats = append(r.lats, p.Lat)
	}
	return true
}

// eastOf is how far east b lies of a, the short way round: negative to the
// west and never more than half a turn, so two dots either side of the
// antimeridian are a dot apart, not the world's width.
func eastOf(a, b float64) float64 {
	return math.Mod(b-a+540, 360) - 180
}

// onWorld is a longitude put back in [-180, 180).
func onWorld(lon float64) float64 {
	return math.Mod(math.Mod(lon+180, 360)+360, 360) - 180
}

// fieldClass is a field's class at a position: the cell of the host's grid it
// falls in, rows from the north. Outside the grid there is no data.
func fieldClass(f *scene.Field, lon, lat float64) int8 {
	if f == nil || f.Cols <= 0 || f.Rows <= 0 || len(f.Classes) != f.Cols*f.Rows {
		return -1
	}
	if !(lon >= f.West && lon < f.East && lat > f.South && lat <= f.North) {
		return -1
	}
	col := int((lon - f.West) / (f.East - f.West) * float64(f.Cols))
	row := int((f.North - lat) / (f.North - f.South) * float64(f.Rows))
	return f.Classes[min(row, f.Rows-1)*f.Cols+min(col, f.Cols-1)]
}

// mercator is a latitude's height in the web map projection.
func mercator(lat float64) float64 {
	return math.Atanh(math.Sin(lat * math.Pi / 180))
}

// presetOf is the preset an image's or field's class ink belongs to.
func presetOf(ink uint8) colour.Preset {
	switch t := colour.Token(ink); {
	case t >= colour.Radar1 && t <= colour.Radar6:
		return colour.Radar
	case t >= colour.Temperature1 && t <= colour.Temperature17:
		return colour.Temperature
	case t >= colour.Wind1 && t <= colour.Wind6:
		return colour.Wind
	case t >= colour.Wave1 && t <= colour.Wave6:
		return colour.Waves
	case t >= colour.UV1 && t <= colour.UV5:
		return colour.UV
	case t >= colour.AQI1 && t <= colour.AQI6:
		return colour.AirQuality
	case t >= colour.QPF1 && t <= colour.QPF7:
		return colour.QPF
	case t >= colour.MUF1 && t <= colour.MUF10:
		return colour.MUF
	case t >= colour.FoF21 && t <= colour.FoF26:
		return colour.FoF2
	}
	return 0
}

// classInk is the ink of a preset's class. Radar's class 0 is below its first
// floor: no rain, and nothing is drawn; so is a total's, a trace (L-26), and
// a foF2 that reaches no band (D-149).
func classInk(preset uint8, class int8) uint8 {
	if class < 0 || (class == 0 && colour.FloorsFirst(colour.Preset(preset))) {
		return 0
	}
	switch colour.Preset(preset) {
	case colour.Temperature:
		return uint8(colour.Temperature1) + uint8(min(int(class), 16))
	case colour.Radar:
		return uint8(colour.Radar1) + uint8(min(int(class), 6)) - 1
	case colour.QPF:
		return uint8(colour.QPF1) + uint8(min(int(class), 7)) - 1
	case colour.MUF:
		return uint8(colour.MUF1) + uint8(min(int(class), 9))
	case colour.FoF2:
		return uint8(colour.FoF21) + uint8(min(int(class), 6)) - 1
	case colour.Wind:
		return uint8(colour.Wind1) + uint8(min(int(class), 5))
	case colour.Waves:
		return uint8(colour.Wave1) + uint8(min(int(class), 5))
	case colour.UV:
		return uint8(colour.UV1) + uint8(min(int(class), 4))
	case colour.AirQuality:
		return uint8(colour.AQI1) + uint8(min(int(class), 5))
	}
	return 0
}

// ShadeFor is the block shade an image's or rain's class is drawn in where
// the depth draws no ramp (FR-18): heavier for heavier classes; empty where
// the depth draws colour or the class draws nothing. The legend keys it
// (L-7.1), so the key and the picture are one rule.
func ShadeFor(preset uint8, class int8, classCount int, depth Depth) string {
	drawn := rampless(depth) && classInk(preset, class) != 0
	if !drawn {
		return ""
	}
	heavy, medium := 3*int(class) > 2*(classCount-1), 3*int(class) > classCount-1
	switch {
	case heavy:
		return shadeHeavy
	case medium:
		return shadeMedium
	}
	return shadeLight
}

// rampless reports whether ramps cannot be drawn at a depth, so that what
// needs one is drawn in its no-colour form (D-59).
func rampless(d Depth) bool {
	return d == colour.NoColour || d == colour.Colours16
}

// underlays sets the background of the cells that fields and images colour
// (L2 Render, step 3), or with no ramp to draw them in, their no-colour forms.
func (r *Renderer) underlays(in Input) {
	if len(in.Fields) == 0 && len(in.Rasters) == 0 {
		return
	}
	if !r.axes(in.View) {
		return
	}
	g := r.grid
	r.water = slices.Grow(r.water[:0], g.cols*g.rows)[:g.cols*g.rows]
	clear(r.water) // the areas are whole: what water owns is asked again this frame
	// WHERE AN IMAGE SHARES THE MAP, A FIELD IS ITS LINES (L-15.3, D-88):
	// both ramps run blue to red, so a band under an echo would read as
	// rain. The field is its labelled contours, over its bands drawn faintly
	// where the image has no echo; the image keeps its own colours.
	shared := in.ImageHeld || len(in.Rasters) > 0 || slices.ContainsFunc(in.Fields, func(f scene.Field) bool { return isRain(&f) })
	for i := range in.Fields {
		if f := &in.Fields[i]; f.Marks != nil {
			r.gridMarks(in, f) // before any value the fields write: a day's total takes the place of a speed or an isotherm's value beside it (L-17.2)
		}
	}
	filled := false // a field before this one filled its bands
	for i := range in.Fields {
		f := &in.Fields[i]
		if f.From != nil {
			r.arrows(in, f) // a vector field is its arrows, never bands (L-16)
			continue
		}
		if isRain(f) {
			continue // drawn as an image is, after every field (L-17.1)
		}
		lined := shared || f.Lines // L-15.4: the host may ask for the look alone
		if f.Edge {
			r.edgeMark(in, f) // L-2.6: its edge broken at every depth, never a contour
			if rampless(in.Depth) || filled {
				continue // over a filled field, its edge alone (A-10)
			}
		} else if rampless(in.Depth) || lined {
			r.contours(in, f)
		}
		if rampless(in.Depth) {
			continue
		}
		r.fill(in, f, lined)
		filled = true
	}
	for i := range in.Fields {
		if f := &in.Fields[i]; isRain(f) {
			r.rain(in, f.Preset, f.ClassCount, false, func(col, row int) int8 { return r.heaviestOf(f, col, row) }) // rain falls on the sea (D-87)
		}
	}
	for i := range in.Rasters {
		r.image(in, &in.Rasters[i])
	}
}

// fill colours the cells under a scalar field's bands, faintly when its
// lines are drawn over them.
func (r *Renderer) fill(in Input, f *scene.Field, lined bool) {
	g, inks := r.grid, r.fieldInks(in, f)
	matters := groundMatters(f)
	for row := range g.rows {
		for col := range g.cols {
			if matters && offItsGround(f, r.waterAt(col, row)) {
				continue
			}
			west := r.lons[2*col] // a cell on the antimeridian is sampled there, not halfway round the world
			lon, lat := onWorld(west+eastOf(west, r.lons[2*col+1])/2), (r.lats[4*row+1]+r.lats[4*row+2])/2
			if class := fieldClass(f, lon, lat); class >= 0 && int(class) < len(inks) && inks[class] != 0 {
				g.cells[row*g.cols+col].under, g.cells[row*g.cols+col].faint = inks[class], lined
			}
		}
	}
}

// fieldInks are a scalar field's ink a class: a preset's tokens, or a host
// type's colours as literal inks (L-2.2); 0 for a class not drawn.
func (r *Renderer) fieldInks(in Input, f *scene.Field) []uint8 {
	inks := make([]uint8, f.ClassCount)
	if f.Preset != 0 {
		for c := range inks {
			inks[c] = classInk(f.Preset, int8(c))
		}
		return inks
	}
	own := make([]colour.RGB, len(f.Colours))
	for i, c := range f.Colours {
		own[i] = colour.RGB{R: c[0], G: c[1], B: c[2]}
	}
	for c, rgb := range colour.HostRamp(own, f.ClassCount, in.Palette, in.Ground.Kind(in.Palette), in.Depth) {
		inks[c] = r.painter.literalInk(rgb)
	}
	return inks
}

// stipple marks a host's zone over a field or an image its tint cannot blend
// with (L-8.2, watchpost D-196): every fourth cell, rows staggered, carries
// one dot in the zone's ink, where the cell has no line work of its own, so
// the field's colours stay exact and its legend holds inside the zone.
func (r *Renderer) stipple(in Input) {
	if rampless(in.Depth) {
		return // the outline alone at sixteen colours and none (L-8.1)
	}
	g := r.grid
	for row := range g.rows {
		for col := range g.cols {
			c := g.cells[row*g.cols+col]
			if c.under == 0 || c.faint || row%2 != 0 || (col+row/2)%2 != 0 || !r.painter.InZone(col, row) {
				continue
			}
			if _, blends := in.Blends.ZoneStrength(presetOf(c.under)); blends {
				continue // the tint is blended into the field there instead
			}
			if glyph, _ := r.painter.lines.Cell(col, row); glyph != blank {
				continue // the cell's own line work keeps it
			}
			r.painter.lines.Set(2*col, 4*row+1, uint8(colour.Zone))
		}
	}
}

// edgeMark marks the edge between a two-class host field's classes with a
// broken line that does not depend on colour (L-2.6): about every other dot
// of the edge, picked by a hash of its place so the line is broken whichever
// way it runs, in the contours' ink and with no label - never a contour.
//
// Each dot is classified once: two rows of classes are held, this one and
// the next, and each dot's east and south neighbours read from them
// (watchpost D-209).
func (r *Renderer) edgeMark(in Input, f *scene.Field) {
	w, h := r.painter.lines.Dots()
	if w < 2 || h < 2 {
		return
	}
	ink := uint8(colour.LabelRegion)
	matters := groundMatters(f)
	row, next := make([]int8, w), make([]int8, w)
	for x := range w {
		row[x] = fieldClass(f, r.lons[x], r.lats[0])
	}
	for y := range h - 1 {
		for x := range w {
			next[x] = fieldClass(f, r.lons[x], r.lats[y+1])
		}
		for x := range w - 1 {
			here := row[x]
			if here < 0 || (matters && offItsGround(f, r.waterAt(x/2, y/4))) {
				continue
			}
			east, south := row[x+1], next[x]
			edge := (east >= 0 && east != here) || (south >= 0 && south != here)
			if edge && brokenAt(x, y) {
				r.painter.lines.Set(x, y, ink)
			}
		}
		row, next = next, row
	}
}

// brokenAt reports whether a dot of a broken edge is lit: about half of
// them, by a hash of the dot's place, the same in every frame (NFR-6).
func brokenAt(x, y int) bool { return (uint32(x)*73856093^uint32(y)*19349663)>>7&1 == 0 }

// offItsGround reports whether a field whose ground matters (groundMatters)
// is not drawn over a cell: waves are the sea's alone (L-20.1), whatever the
// host says; any other such field stops at the shore (D-32).
func offItsGround(f *scene.Field, water bool) bool {
	if colour.Preset(f.Preset) == colour.Waves {
		return !water
	}
	return water
}

// waterAt reports whether water owns a cell: the painter's answer, asked
// once a frame for a cell, since it counts the cell's dots and fields and
// images ask it of every dot. The areas are whole before any field is drawn.
func (r *Renderer) waterAt(col, row int) bool {
	i := row*r.grid.cols + col
	if known := r.water[i]; known != waterUnknown {
		return known == waterOwns
	}
	owns := r.painter.Water(col, row)
	r.water[i] = waterNot
	if owns {
		r.water[i] = waterOwns
	}
	return owns
}

// A cell's water, as waterAt holds it: not asked yet this frame, or the
// answer.
const (
	waterUnknown uint8 = iota
	waterNot
	waterOwns
)

// groundMatters reports whether water decides any of a field's cells: for
// waves, and for a field its host has not continued over water (L-1.1); a
// field continued over it never asks (the test counts a cell's dots).
func groundMatters(f *scene.Field) bool {
	return colour.Preset(f.Preset) == colour.Waves || !f.OverWater
}

// isRain reports whether a field is in radar's scale: A GRID OF RAIN IS
// DRAWN AS RAIN (L-17.1, watchpost D-115, D-116) - in its own colours at full
// strength, over the sea, over any field, never lined - and a field beside it
// takes its lines, as beside an image.
func isRain(f *scene.Field) bool {
	return f.From == nil && colour.Preset(f.Preset) == colour.Radar
}

// heaviestOf is the heaviest class of a field among a cell's eight dots, as
// an image's is (D-78).
func (r *Renderer) heaviestOf(f *scene.Field, col, row int) int8 {
	best := int8(-1)
	for dy := range 4 {
		for dx := range 2 {
			best = max(best, fieldClass(f, r.lons[2*col+dx], r.lats[4*row+dy]))
		}
	}
	return best
}

// maxSpan bounds how many pixels across and down one dot's footprint is
// searched; an image finer than that is strode over, evenly.
const maxSpan = 4

// under is the heaviest class among the pixels one dot covers. A dot is a
// sample of D-78's eight, but where an image is finer than the dots a sample
// at the dot's centre alone can fall either side of a storm's core one pixel
// wide - and D-78's own words are "better to overstate than under".
func (r *Renderer) under(ra *scene.Raster, x, y int) int8 {
	if ra == nil || ra.Width <= 0 || ra.Height <= 0 || len(ra.Classes) != ra.Width*ra.Height {
		return -1
	}
	w, h := len(r.lons), len(r.lats)
	halfLon := eastOf(r.lons[max(x-1, 0)], r.lons[min(x+1, w-1)]) / 4
	halfLat := (r.lats[max(y-1, 0)] - r.lats[min(y+1, h-1)]) / 4
	// Each of the footprint's rows and columns is placed in the image once,
	// and a web map image's projected bounds once a dot: the same arithmetic
	// as sampling each of the sixteen points, done four times, not sixteen.
	top, span := ra.North, ra.North-ra.South
	if ra.Projection == 2 {
		top = mercator(ra.North)
		span = top - mercator(ra.South)
	}
	var rows, cols [maxSpan]int
	for i := range maxSpan {
		lat := r.lats[y] + halfLat - 2*halfLat*(float64(i)+0.5)/maxSpan
		rows[i] = -1
		if lat > ra.South && lat <= ra.North {
			down := (ra.North - lat) / span
			if ra.Projection == 2 {
				down = (top - mercator(lat)) / span
			}
			rows[i] = min(max(int(down*float64(ra.Height)), 0), ra.Height-1)
		}
		lon := onWorld(r.lons[x] - halfLon + 2*halfLon*(float64(i)+0.5)/maxSpan)
		cols[i] = -1
		if lon >= ra.West && lon < ra.East {
			cols[i] = min(int((lon-ra.West)/(ra.East-ra.West)*float64(ra.Width)), ra.Width-1)
		}
	}
	best := int8(-1)
	for _, row := range rows {
		for _, col := range cols {
			if row >= 0 && col >= 0 {
				best = max(best, ra.Classes[row*ra.Width+col])
			}
		}
	}
	return best
}

// heaviest is the heaviest class among a cell's eight dots (D-78): a storm's
// core is not averaged away.
func (r *Renderer) heaviest(ra *scene.Raster, col, row int) int8 {
	best := int8(-1)
	for dy := range 4 {
		for dx := range 2 {
			best = max(best, r.under(ra, 2*col+dx, 4*row+dy))
		}
	}
	return best
}

// image colours the cells an image covers.
func (r *Renderer) image(in Input, ra *scene.Raster) {
	r.rain(in, ra.Preset, ra.ClassCount, ra.MaskedByWater, func(col, row int) int8 { return r.heaviest(ra, col, row) })
}

// rain colours the cells an image or a grid of rain covers - over water as
// over land, because rain falls on the sea (D-87), unless masked keeps it off
// the sea (L-1.2) - or with no ramp shades them; class is a cell's heaviest.
func (r *Renderer) rain(in Input, preset uint8, classCount int, masked bool, class func(col, row int) int8) {
	g := r.grid
	for row := range g.rows {
		for col := range g.cols {
			if masked && r.waterAt(col, row) {
				continue
			}
			k := class(col, row)
			ink := classInk(preset, k)
			if ink == 0 {
				continue
			}
			c := &g.cells[row*g.cols+col]
			if !rampless(in.Depth) {
				c.under, c.faint = ink, false // the echo in its own colour, over any field (L-15.3)
				continue
			}
			c.shade = max(c.shade, ShadeFor(preset, k, classCount, in.Depth)) // laid last, by shades: it claims nothing now (L-8.3)
		}
	}
}

// contours draws a field with no colour (D-35): a line wherever the class of
// one dot differs from the next, which bends smoothly at a terminal's
// resolution and needs no special cases (S25-1); and each line's value, as a
// label, every so often. A field all of one class has no line (D-68).
func (r *Renderer) contours(in Input, f *scene.Field) {
	w, h := r.painter.lines.Dots()
	ink := uint8(colour.LabelRegion)
	matters := groundMatters(f)
	for y := range h - 1 {
		for x := range w - 1 {
			here, east, south, ok := r.edgeAt(f, matters, x, y)
			if !ok {
				continue
			}
			if (east >= 0 && east != here) || (south >= 0 && south != here) {
				r.painter.lines.Set(x, y, ink)
			}
			if y%labelEvery == labelEvery/2 && east >= 0 && east != here {
				r.painter.valueLabel(x, y, f.Labels, max(here, east), ink)
			}
			if x%labelEvery == labelEvery/2 && south >= 0 && south != here { // a line running east-west is labelled too (L-3.2)
				r.painter.valueLabel(x, y, f.Labels, max(here, south), ink)
			}
		}
	}
}

// edgeAt is a field's class at a dot and its neighbours' to the east and
// the south, -1 where they have none; false for a dot with no class, or off
// the field's ground.
func (r *Renderer) edgeAt(f *scene.Field, matters bool, x, y int) (here, east, south int8, ok bool) {
	here = fieldClass(f, r.lons[x], r.lats[y])
	if here < 0 || (matters && offItsGround(f, r.waterAt(x/2, y/4))) {
		return here, -1, -1, false
	}
	return here, fieldClass(f, r.lons[x+1], r.lats[y]), fieldClass(f, r.lons[x], r.lats[y+1]), true
}

// valueLabel offers a contour's value as a label at a dot.
func (p *Painter) valueLabel(x, y int, labels []string, class int8, ink uint8) {
	if class < 0 || int(class) >= len(labels) || labels[class] == "" || len(p.bandLabels) >= maxLabels {
		return
	}
	p.bandLabels = append(p.bandLabels, Label{X: x, Y: y, Name: textsafe.Clean(labels[class]), Ink: ink})
}
