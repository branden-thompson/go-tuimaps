package tuimaps

import (
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/overlay"
	"github.com/branden-thompson/go-tuimaps/internal/render"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// Class is one step of a legend: the values it covers, the words for them,
// and the colour it is drawn in where there is one.
type Class struct {
	Label  string // the range in words, in the overlay's own unit
	Colour RGB
	Drawn  bool // false where the depth in use draws no colour at all
}

// LegendEntry is what one overlay puts on the map, as data (FR-18).
type LegendEntry struct {
	ID      string
	Unit    string
	Preset  string // "temperature", "radar", or empty for a host's own type
	Classes []Class
}

// Scaled is the scale of the map as data, in the terms the mark on the frame
// draws it (FR-33).
type Scaled struct {
	KmPerCell float64 // how far one cell spans, across the middle of the map
	Km        float64 // the round distance the bar stands for
	Cells     int     // how many cells long the bar is
}

// Legend is what is drawn on the map now, overlay by overlay and class by
// class, in the colours actually in use at the depth in effect (FR-18). It
// is the same facts the picture carries, for a host that cannot show them.
func (m *Map) Legend() []LegendEntry {
	defer m.guardQuiet("Legend")
	m.plant("Legend")

	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return nil
	}
	var out []LegendEntry
	for _, id := range m.store.IDs() {
		entry, ok := m.legendOf(id)
		if ok {
			out = append(out, entry)
		}
	}
	return out
}

// legendOf is one overlay's legend, or false for one that has no classes:
// features carry their own labels and need none.
func (m *Map) legendOf(id string) (LegendEntry, bool) {
	reader, ok := m.store.Read(id)
	if !ok {
		return LegendEntry{}, false
	}
	o := reader.Overlay()
	reader.Done()
	var kind overlay.Type
	switch {
	case o.Grid != nil:
		kind = o.Grid.Type
	case o.Image != nil:
		kind = o.Image.Type
	default:
		return LegendEntry{}, false
	}
	resolved, err := overlay.ResolveType(kind)
	if err != nil {
		return LegendEntry{}, false
	}
	return LegendEntry{ID: id, Unit: kind.Unit, Preset: kind.Preset,
		Classes: m.classesOf(resolved)}, true
}

// classesOf is one type's classes: one more than its breaks, each labelled
// in the overlay's own unit and coloured as the frame draws it. The colour
// is asked for the way the renderer asks for it, so the legend and the
// picture can never disagree - at sixteen colours it is the sixteen-colour
// approximation, and with no colour there is none.
func (m *Map) classesOf(kind overlay.Kind) []Class {
	depth := m.depthInEffect()
	ground := m.look.ground.Kind(m.look.palette)
	first, ramped := rampToken(kind.Preset)
	out := make([]Class, 0, len(kind.Breaks)+1)
	for i := range len(kind.Breaks) + 1 {
		one := Class{Label: classLabel(kind.Breaks, i)}
		if ramped && depth != NoColour {
			one.Colour, one.Drawn = m.look.palette.ResolveAt(first+Token(i), ground, depth)
		}
		out = append(out, one)
	}
	return out
}

// rampToken is the token of a preset's first class; the classes after it
// follow in order. A host's own type has no ramp of its own yet.
func rampToken(preset colour.Preset) (Token, bool) {
	switch preset {
	case colour.Temperature:
		return colour.Temperature1, true
	case colour.Radar:
		return colour.Radar1, true
	}
	return 0, false
}

// classLabel is the range a class covers, in words: below the first break,
// between two, or above the last.
func classLabel(breaks []float64, i int) string {
	switch {
	case len(breaks) == 0:
		return "all values"
	case i == 0:
		return "under " + number(breaks[0])
	case i >= len(breaks):
		return number(breaks[len(breaks)-1]) + " and above"
	}
	return number(breaks[i-1]) + " to " + number(breaks[i])
}

// number is a break written as shortly as it can be read.
func number(v float64) string {
	if v == math.Trunc(v) {
		return strconv.FormatFloat(v, 'f', 0, 64)
	}
	return strconv.FormatFloat(v, 'f', 1, 64)
}

// Credits are every credit the map owes: the basemap's, then each overlay's
// in the order they were set, once each (FR-14).
func (m *Map) Credits() []string {
	defer m.guardQuiet("Credits")
	m.plant("Credits")

	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return nil
	}
	out := []string{basemapCredit}
	for _, id := range m.store.IDs() {
		reader, ok := m.store.Read(id)
		if !ok {
			continue
		}
		credit := textsafe.Clean(reader.Overlay().Credit).String()
		reader.Done()
		if credit != "" && !slices.Contains(out, credit) {
			out = append(out, credit)
		}
	}
	return out
}

// basemapCredit is what the tiles are owed, drawn on every frame (FR-14).
const basemapCredit = "OpenFreeMap (c) OpenMapTiles Data from OpenStreetMap"

// Scale is how far the map spans, as data (FR-33): how far one cell is, and
// the round distance the bar on the frame stands for.
func (m *Map) Scale() (Scaled, bool) {
	defer m.guardQuiet("Scale")
	m.plant("Scale")

	if m == nil {
		return Scaled{}, false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut || !m.sized {
		return Scaled{}, false
	}
	perCol, _, err := m.view.CellSpanKm()
	if err != nil || !(perCol > 0) {
		return Scaled{}, false
	}
	// The same rule the mark on the frame is drawn by, so the two never
	// disagree about how far the map spans.
	km, cells := render.RoundBar(perCol, m.view.Cols/3)
	return Scaled{KmPerCell: perCol, Km: km, Cells: cells}, true
}

// Footer is the centre and the zoom in upstream's own wording, cut with
// floor as upstream cuts them (P-57). It is drawn inside the map only when
// the host turns it on, and it is off by default.
func (m *Map) Footer() string {
	defer m.guardQuiet("Footer")
	m.plant("Footer")

	if m == nil {
		return ""
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut || !m.sized {
		return ""
	}
	return footerOf(m.view.Centre.Lat, m.view.Centre.Lon, m.view.Zoom)
}

// footerOf writes one footer line. Upstream cuts rather than rounds.
func footerOf(lat, lon, zoom float64) string {
	var b strings.Builder
	b.WriteString("center: ")
	b.WriteString(cut(lat, 3))
	b.WriteString(", ")
	b.WriteString(cut(lon, 3))
	b.WriteString("   zoom: ")
	b.WriteString(cut(zoom, 0))
	return b.String()
}

// cut writes a number to a number of decimal places, cutting what is left
// rather than rounding it, as upstream does (P-57).
func cut(v float64, places int) string {
	scale := math.Pow(10, float64(places))
	whole := math.Trunc(v*scale) / scale
	return strconv.FormatFloat(whole, 'f', places, 64)
}

// ShowFooter draws the footer inside the map, or takes it off again. It is
// off by default (P-57).
func (m *Map) ShowFooter(on bool) {
	defer m.guardQuiet("ShowFooter")
	m.plant("ShowFooter")

	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut || m.footer == on {
		return
	}
	m.footer = on
	m.changed++
}
