package render

import (
	"math"

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

// rasterClass is an image's class at a position, in the projection the host
// stated: rows in equal steps of latitude, or in equal steps of the web map's
// height.
func rasterClass(ra *scene.Raster, lon, lat float64) int8 {
	if ra == nil || ra.Width <= 0 || ra.Height <= 0 || len(ra.Classes) != ra.Width*ra.Height {
		return -1
	}
	if !(lon >= ra.West && lon < ra.East && lat > ra.South && lat <= ra.North) {
		return -1
	}
	col := int((lon - ra.West) / (ra.East - ra.West) * float64(ra.Width))
	down := (ra.North - lat) / (ra.North - ra.South)
	if ra.Projection == 2 {
		down = (mercator(ra.North) - mercator(lat)) / (mercator(ra.North) - mercator(ra.South))
	}
	row := int(down * float64(ra.Height))
	return ra.Classes[min(max(row, 0), ra.Height-1)*ra.Width+min(col, ra.Width-1)]
}

// classInk is the ink of a preset's class. Radar's class 0 is below its first
// floor: no rain, and nothing is drawn.
// presetOf is the preset an image's or field's class ink belongs to.
func presetOf(ink uint8) colour.Preset {
	switch t := colour.Token(ink); {
	case t >= colour.Radar1 && t <= colour.Radar6:
		return colour.Radar
	case t >= colour.Temperature1 && t <= colour.Temperature17:
		return colour.Temperature
	}
	return 0
}

func classInk(preset uint8, class int8) uint8 {
	if class < 0 {
		return 0
	}
	switch colour.Preset(preset) {
	case colour.Temperature:
		return uint8(colour.Temperature1) + uint8(min(int(class), 16))
	case colour.Radar:
		if class == 0 {
			return 0
		}
		return uint8(colour.Radar1) + uint8(min(int(class), 6)) - 1
	}
	return 0
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
	for i := range in.Fields {
		f := &in.Fields[i]
		if rampless(in.Depth) {
			r.contours(in, f)
			continue
		}
		for row := range g.rows {
			for col := range g.cols {
				if !in.FieldsOverWater && r.painter.Water(col, row) {
					continue // temperature stops at the shore (D-32)
				}
				lon, lat := (r.lons[2*col]+r.lons[2*col+1])/2, (r.lats[4*row+1]+r.lats[4*row+2])/2
				if ink := classInk(f.Preset, fieldClass(f, lon, lat)); ink != 0 {
					g.cells[row*g.cols+col].under = ink
				}
			}
		}
	}
	for i := range in.Rasters {
		r.image(in, &in.Rasters[i])
	}
}

// maxSpan bounds how many pixels across and down one dot's footprint is
// searched; an image finer than that is strode over, evenly.
const maxSpan = 4

// under is the heaviest class among the pixels one dot covers. A dot is a
// sample of D-78's eight, but where an image is finer than the dots a sample
// at the dot's centre alone can fall either side of a storm's core one pixel
// wide - and D-78's own words are "better to overstate than under".
func (r *Renderer) under(ra *scene.Raster, x, y int) int8 {
	w, h := len(r.lons), len(r.lats)
	halfLon := (r.lons[min(x+1, w-1)] - r.lons[max(x-1, 0)]) / 4
	halfLat := (r.lats[max(y-1, 0)] - r.lats[min(y+1, h-1)]) / 4
	best := int8(-1)
	for i := range maxSpan {
		lat := r.lats[y] + halfLat - 2*halfLat*(float64(i)+0.5)/maxSpan
		for j := range maxSpan {
			lon := r.lons[x] - halfLon + 2*halfLon*(float64(j)+0.5)/maxSpan
			best = max(best, rasterClass(ra, lon, lat))
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

// image colours the cells an image covers - over water as over land, because
// rain falls on the sea (D-87) - or with no ramp shades them.
func (r *Renderer) image(in Input, ra *scene.Raster) {
	g := r.grid
	for row := range g.rows {
		for col := range g.cols {
			if in.ImagesMaskedByWater && r.painter.Water(col, row) {
				continue
			}
			class := r.heaviest(ra, col, row)
			ink := classInk(ra.Preset, class)
			if ink == 0 {
				continue
			}
			c := &g.cells[row*g.cols+col]
			if !rampless(in.Depth) {
				c.under = ink
				continue
			}
			shade := shadeLight
			if 3*int(class) > 2*(ra.ClassCount-1) {
				shade = shadeHeavy
			} else if 3*int(class) > ra.ClassCount-1 {
				shade = shadeMedium
			}
			c.shade = max(c.shade, shade) // laid last, by shades: it claims nothing now (L-8.3)
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
	for y := range h - 1 {
		for x := range w - 1 {
			here := fieldClass(f, r.lons[x], r.lats[y])
			if here < 0 || (!in.FieldsOverWater && r.painter.Water(x/2, y/4)) {
				continue
			}
			east, south := fieldClass(f, r.lons[x+1], r.lats[y]), fieldClass(f, r.lons[x], r.lats[y+1])
			if (east >= 0 && east != here) || (south >= 0 && south != here) {
				r.painter.lines.Set(x, y, ink)
			}
			if y%labelEvery == labelEvery/2 && east >= 0 && east != here {
				r.painter.valueLabel(x, y, f.Labels, max(here, east), ink)
			}
		}
	}
}

// valueLabel offers a contour's value as a label at a dot.
func (p *Painter) valueLabel(x, y int, labels []string, class int8, ink uint8) {
	if class < 0 || int(class) >= len(labels) || labels[class] == "" || len(p.bandLabels) >= maxLabels {
		return
	}
	p.bandLabels = append(p.bandLabels, Label{X: x, Y: y, Name: textsafe.Clean(labels[class]), Ink: ink})
}
