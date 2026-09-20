package render

import (
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/style"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

const (
	// labelMargin is upstream's collision margin, in cells (P-34).
	labelMargin = 5
	// placeGlyph is what a symbol with no name draws (P-35).
	placeGlyph = "\u25C9"
)

// Depth is the colour depth a frame is emitted at: the colour package's.
type Depth = colour.Depth

// The depths, by their short names here.
const (
	Truecolor = colour.Truecolor
	NoColour  = colour.NoColour
)

// colourless reports whether a frame at this depth carries no colour sequence.
func colourless(d Depth) bool {
	return d == colour.NoColour
}

// Status says how finished a frame is.
type Status uint8

// The statuses of a frame.
const (
	Complete   Status = iota + 1 // every tile the view wants is drawn as itself
	Sharpening                   // something is a stand-in, or is missing; work is pending
	NoTiles                      // nothing is on hand from any source
)

// String names the status.
func (s Status) String() string {
	switch s {
	case Complete:
		return "complete"
	case Sharpening:
		return "still sharpening"
	case NoTiles:
		return "no tiles"
	}
	return "unknown"
}

// Drawn is one tile on hand: the tile, which tile it is, and whether it is
// the tile the view wanted or an ancestor standing in for it.
type Drawn struct {
	Tile  *scene.Tile
	At    scene.TileID
	Exact bool
}

// Input is everything a frame is a function of (NFR-6).
type Input struct {
	View  project.View
	Tiles []Drawn // in any order: the renderer draws them in one
	// Shapes are the overlays' prepared shapes, in the order they are drawn;
	// OverlaysVersion counts their changes, since they cannot be compared.
	Shapes []scene.Shape
	// Borrowed are the overlays drawn straight from the host's memory,
	// which are read through their run index and never copied (D-92).
	Borrowed        []Borrowed
	Fields          []scene.Field  // prepared scalar grids, sampled into cells here, at draw time
	Rasters         []scene.Raster // prepared images, resampled here, at draw time
	OverlaysVersion uint64
	// FieldsOverWater and ImagesMaskedByWater flip the two defaults: a field
	// stops at the shore (D-32), and an image never does (D-87).
	FieldsOverWater     bool
	ImagesMaskedByWater bool
	// FieldLabels are a field's values as text, by class: what its contour
	// lines carry when there is no colour (D-35).
	FieldLabels []string
	// Markers are the host's places, drawn over everything beneath them
	// (FR-26); MarkerPhase is which half of the blink this frame draws, and
	// Motion works it out from the host's clock.
	Markers     []Marker
	MarkerPhase bool
	Missing     int            // tiles the view wants with nothing on hand to draw for them
	Layers      style.Switches // the basemap layers the host has switched off (FR-36)
	Style       *style.Style
	Palette     colour.Palette
	// Look counts changes to the palette, which cannot be compared: whoever
	// changes the palette raises it, and a frame is reused only while it and
	// everything else here stay the same (contract, section 5).
	Look   uint64
	Ground colour.GroundChoice
	Depth  Depth
	Labels bool
	Scale  bool
	// Simplify turns on upstream's line simplification, which upstream and
	// this library both leave off (P-31).
	Simplify bool
	// Stale says that the data of an overlay on the frame is no longer
	// current, which the frame marks in a word (FR-32); Footer is the
	// host's own line of text, drawn only when it is given (P-57).
	Stale  bool
	Footer textsafe.Text
	Credit textsafe.Text
}

// Frame is a drawn map: one string a row, each exactly the view's width in
// cells, holding colour sequences and cleaned text and nothing else.
type Frame struct {
	Lines  []string
	Status Status
}

// cell is one cell of the frame being composed.
type cell struct {
	text   string // a cluster of text, which wins over dots (P-10); empty for a dot cell
	glyph  rune
	ink    uint8
	area   uint8 // the ink of the area that owns the cell's background, or 0
	under  uint8 // the ink of the field or image class that colours the cell, or 0
	taken  bool  // text occupies the cell: a label, or the second half of a wide character
	strict bool  // the cell holds text, held to the text contrast
}

type box struct{ left, right, top, bottom int }

// holds reports whether a point is inside the box, its right and bottom edges
// excluded. The zero box bounds nothing.
func (b box) holds(p Point) bool {
	if b == (box{}) {
		return true
	}
	return p.X >= b.left && p.X < b.right && p.Y >= b.top && p.Y < b.bottom
}

// grid is the frame's cells and the boxes of the labels placed so far.
type grid struct {
	cols, rows int
	cells      []cell
	boxes      []box
	world      box // the world's edges in dots; the zero box means no bound is known
}

func newGrid(cols, rows int) (*grid, error) {
	if cols <= 0 || rows <= 0 || cols > maxCells || rows > maxCells || cols*rows > maxCells {
		return nil, fault.Make(fault.NoSize, textsafe.Const("the map has no size to draw at"),
			textsafe.Const("its width and height in cells must each be at least 1"), textsafe.Const("give the map a size before drawing"))
	}
	return &grid{cols: cols, rows: rows, cells: make([]cell, cols*rows)}, nil
}

func (g *grid) reset() {
	clear(g.cells)
	g.boxes = g.boxes[:0]
}

// write puts text into a row from a column, a cluster a cell, a wide cluster
// taking the cell after it as well (P-11). It writes nothing unless the
// whole text fits and every cell is free.
func (g *grid) write(col, row int, text textsafe.Text, ink uint8) bool {
	width := textsafe.Width(text)
	if row < 0 || row >= g.rows || col < 0 || width == 0 || col+width > g.cols {
		return false
	}
	for i := range width {
		if g.cells[row*g.cols+col+i].taken {
			return false
		}
	}
	at := row*g.cols + col
	textsafe.Each(text, func(cluster string, w int) bool {
		if w == 0 {
			return true
		}
		g.cells[at] = cell{text: cluster, ink: ink, taken: true, strict: true, area: g.cells[at].area}
		if w == 2 {
			g.cells[at+1] = cell{ink: ink, taken: true, strict: true, area: g.cells[at+1].area}
		}
		at += w
		return true
	})
	return true
}

// label places one name: centred on its point by its width in cells (L-7),
// skipped whole if it would leave the rectangle (P-33) or if its box - the
// text grown by upstream's margin - meets one already placed (P-34).
func (g *grid) label(l Label) {
	// Whether it was placed is nobody's business here: a name that does not
	// fit is skipped whole, and the frame is right either way.
	_ = g.labelAt(l, []Point{{X: l.X, Y: l.Y}})
}

// labelAt tries each of a name's vertices in turn until one takes it (P-32).
func (g *grid) labelAt(l Label, anchors []Point) bool {
	if g == nil || len(anchors) == 0 {
		return false
	}
	for _, a := range anchors {
		if g.anchor(l, a) {
			return true
		}
	}
	return false
}

// anchor places a name at one vertex.
func (g *grid) anchor(l Label, at Point) bool {
	text := l.Name
	if text.String() == "" {
		text = textsafe.Const(placeGlyph)
	}
	width := textsafe.Width(text)
	if width == 0 || at.X < 0 || at.Y < 0 {
		return false // above or left of the rectangle: never a negative row (P-33, L-8)
	}
	if !g.world.holds(at) {
		return false // beyond the world's edge: a tile's buffer repeats places a world away (P-33)
	}
	col, row := at.X/2, at.Y/4
	if !l.fromPoint {
		col -= width / 2 // a name is centred on its place; a marker's label begins at its point
	}
	mine := box{left: col - labelMargin, right: col + labelMargin + width, top: row - labelMargin/2, bottom: row + labelMargin/2}
	for _, b := range g.boxes {
		if mine.left <= b.right && b.left <= mine.right && mine.top <= b.bottom && b.top <= mine.bottom {
			return false
		}
	}
	if !g.write(col, row, text, l.Ink) {
		return false
	}
	g.boxes = append(g.boxes, mine)
	return true
}

// Renderer draws frames of one size, keeping its buffers between them.
type Renderer struct {
	painter *Painter
	grid    *grid
	order   []Drawn
	labels  []Label
	lons    []float64 // the longitude of each dot column's centre, for this frame
	lats    []float64 // the latitude of each dot row's
	line    strings.Builder

	drawn     bool  // a frame has been drawn, and last describes it
	last      Input // the input of the frame held; its tiles are lastTiles
	lastTiles []Drawn
	held      Frame    // valid until the next Render (contract, section 5)
	rowSums   []uint64 // what each held row was built from
	redraws   int
	rowsBuilt int
}

// Redraws is how many frames were drawn and not reused.
func (r *Renderer) Redraws() int {
	if r == nil {
		return 0
	}
	return r.redraws
}

// RowsBuilt is how many rows of text were built, over every frame.
func (r *Renderer) RowsBuilt() int {
	if r == nil {
		return 0
	}
	return r.rowsBuilt
}

// sameOverlays reports whether the overlays are those of the frame held. They
// cannot be compared, so whoever changes them raises their version.
func (r *Renderer) sameOverlays(in Input) bool {
	if r == nil {
		return false
	}
	l := r.last
	if in.FieldsOverWater != l.FieldsOverWater || in.ImagesMaskedByWater != l.ImagesMaskedByWater {
		return false
	}
	if in.MarkerPhase != l.MarkerPhase || len(in.Markers) != len(l.Markers) {
		return false
	}
	for i, m := range in.Markers {
		if m != l.Markers[i] {
			return false
		}
	}
	return in.OverlaysVersion == l.OverlaysVersion && len(in.Shapes) == len(l.Shapes) && len(in.Borrowed) == len(l.Borrowed) && len(in.Fields) == len(l.Fields) && len(in.Rasters) == len(l.Rasters)
}

// sameLook reports whether everything but the tiles and the overlays is as it
// was: the view, the colours, what is drawn of the basemap, and the furniture.
func (r *Renderer) sameLook(in Input) bool {
	l := r.last
	if in.View != l.View || in.Look != l.Look || in.Ground != l.Ground || in.Depth != l.Depth || in.Style != l.Style {
		return false
	}
	if in.Stale != l.Stale || in.Footer != l.Footer || in.Simplify != l.Simplify {
		return false
	}
	return in.Labels == l.Labels && in.Scale == l.Scale && in.Credit == l.Credit && in.Missing == l.Missing && in.Layers == l.Layers
}

// unchanged reports whether nothing a frame is a function of has changed
// since the frame held was drawn. It allocates nothing.
func (r *Renderer) unchanged(in Input) bool {
	if !r.drawn || len(in.Tiles) != len(r.lastTiles) {
		return false
	}
	if !r.sameOverlays(in) {
		return false
	}
	if !r.sameLook(in) {
		return false
	}
	for i, d := range in.Tiles {
		if d != r.lastTiles[i] {
			return false
		}
	}
	return true
}

// NewRenderer makes a renderer for a map of cols by rows cells.
func NewRenderer(cols, rows int) (*Renderer, error) {
	p, err := NewPainter(cols, rows)
	if err != nil {
		return nil, err
	}
	g, err := newGrid(cols, rows)
	if err != nil {
		return nil, err
	}
	return &Renderer{painter: p, grid: g}, nil
}

func badInput(why textsafe.Text) error {
	return fault.Make(fault.Internal, textsafe.Const("the frame could not be drawn"), why, textsafe.Const("this is a defect in the library; report it"))
}

// Draw draws one frame. It reads only what it is given and never waits.
func (r *Renderer) Draw(in Input) (Frame, error) {
	if r == nil || in.Style == nil {
		return Frame{}, badInput(textsafe.Const("it was given no renderer or no style"))
	}
	if in.View.Cols != r.grid.cols || in.View.Rows != r.grid.rows {
		return Frame{}, badInput(textsafe.Const("the view is not the size the renderer was made for"))
	}
	if r.unchanged(in) {
		return r.held, nil // an unchanged frame costs nothing (NFR-4)
	}
	r.painter.Reset()
	r.grid.reset()
	status, err := r.paint(in)
	if err != nil {
		r.drawn = false
		return Frame{}, err
	}
	r.compose(in, status)
	r.held = Frame{Lines: r.emit(in), Status: status}
	r.lastTiles = append(r.lastTiles[:0], in.Tiles...)
	r.last, r.drawn = in, true
	r.last.Tiles = nil
	r.last.Shapes = in.Shapes[:len(in.Shapes):len(in.Shapes)]
	r.redraws++
	return r.held, nil
}

// paint draws every tile, in one order whatever order they came in (NFR-6):
// shallower tiles first, so a stand-in lies under what is sharper.
func (r *Renderer) paint(in Input) (Status, error) {
	r.order = append(r.order[:0], in.Tiles...)
	sort.SliceStable(r.order, func(i, j int) bool {
		a, b := r.order[i].At, r.order[j].At
		if a.Z != b.Z {
			return a.Z < b.Z
		}
		if a.X != b.X {
			return a.X < b.X
		}
		return a.Y < b.Y
	})
	r.painter.SetProfile(style.NewProfile(load(in), in.View.Cols, in.View.Rows, in.Layers))
	r.painter.SetDepth(in.Depth)
	r.painter.SetSimplify(in.Simplify)
	status := Complete
	if len(r.order) == 0 {
		status = NoTiles
	}
	if in.Missing > 0 && status == Complete {
		status = Sharpening
	}
	if status != NoTiles {
		err := r.painter.World(in.View)
		if err != nil {
			return status, err
		}
	}
	for _, d := range r.order {
		if d.Tile == nil {
			continue
		}
		if !d.Exact {
			status = Sharpening
		}
		err := r.painter.Tile(in.View, d.Tile, d.At, in.Style)
		if err != nil {
			return status, err
		}
	}
	// Overlays are drawn over the basemap, and with no tile at all they are
	// still drawn: the frame is never an empty rectangle (FR-23). Markers
	// go over everything beneath them, a hatch included (FR-18a).
	err := r.overlays(in)
	if err != nil {
		return status, err
	}
	return status, r.marks(in)
}

// overlays draws the prepared shapes, then the ones read straight from
// the host's memory through their run index (D-92).
func (r *Renderer) overlays(in Input) error {
	for _, s := range in.Shapes {
		err := r.painter.Shape(in.View, s)
		if err != nil {
			return err
		}
	}
	for _, b := range in.Borrowed {
		err := r.painter.Borrow(in.View, b)
		if err != nil {
			return err
		}
	}
	return nil
}

// marks draws the host's places over everything already drawn.
func (r *Renderer) marks(in Input) error {
	for _, m := range in.Markers {
		err := r.painter.Mark(in.View, m, in.MarkerPhase)
		if err != nil {
			return err
		}
	}
	return nil
}

// load is what is drawn over the basemap this frame, which decides how much
// of itself the basemap gives up (FR-19).
func load(in Input) style.Load {
	if len(in.Fields) > 0 || len(in.Rasters) > 0 {
		return style.Covered
	}
	return style.Bare
}

// compose builds the cells: areas, line work, then names, then furniture.
func (r *Renderer) compose(in Input, status Status) {
	g := r.grid
	r.underlays(in) // before the cells are read: a field with no colour is lines, drawn on the canvas
	for row := range g.rows {
		for col := range g.cols {
			c := &g.cells[row*g.cols+col]
			if c.taken {
				continue // an image's shade with no colour: text, which wins over dots
			}
			c.glyph, c.ink = r.painter.Lines().Cell(col, row)
			if ink, owned := r.painter.Area(col, row); owned {
				c.area = ink
			}
		}
	}
	// The text of a cell is a contest, not a painting: the first to claim a
	// cell keeps it, so the order below is the order of precedence. The
	// frame's own furniture is never overdrawn; a marker keeps its cell
	// against any name (FR-18a); a marker's own label is checked last (P-60).
	r.furniture(in, status)
	for _, g := range r.painter.Glyphs() {
		_ = r.grid.anchor(g, Point{X: g.X, Y: g.Y}) // a marker outside the rectangle is simply not drawn
	}
	// An overlay's labels are placed before any name of the basemap's, so that
	// a place name never hides a warning's word; they are not the basemap's
	// labels and do not go when those are turned off.
	g.world = box{}
	for _, l := range r.painter.OverlayLabels() {
		g.label(l)
	}
	if !in.Labels {
		r.markerNames() // the host's own places are not the basemap's names
		if colourless(in.Depth) {
			r.hatch()
		}
		return
	}
	if x, y, side, err := in.View.TilePlace(scene.TileID{}); err == nil {
		g.world = box{left: toDot(x), top: toDot(y), right: toDot(x + side), bottom: toDot(y + side)}
	}
	// Names last, most important first; the order among equals is the order
	// the tiles gave them, and the tiles were sorted (P-25).
	r.labels = append(r.labels[:0], r.painter.Labels()...)
	sort.SliceStable(r.labels, func(i, j int) bool { return r.labels[i].Rank < r.labels[j].Rank })
	budget := style.NewProfile(load(in), g.cols, g.rows, in.Layers).Labels()
	placed := 0
	for _, l := range r.labels {
		if budget > 0 && placed >= budget {
			break // the profile's label budget: fewer names under an overlay, or on a small map
		}
		if g.labelAt(l, r.painter.Points(l)) {
			placed++
		}
	}
	r.markerNames()
	if colourless(in.Depth) {
		r.hatch() // last of all: it fills what nothing else has taken (FR-18a)
	}
}

// markerNames places the markers' labels, which are checked against the
// map's own names and so are placed after them (P-60).
func (r *Renderer) markerNames() {
	r.grid.world = box{}
	for _, l := range r.painter.MarkerLabels() {
		r.grid.label(l)
	}
}

// furniture is what is drawn over the map's edge: the notice when there are
// no tiles (FR-23), the credit line (FR-14), the scale mark (FR-33).
func (r *Renderer) furniture(in Input, status Status) {
	g := r.grid
	if status == NoTiles {
		notice := textsafe.Const("no map tiles: name a source, or pass the assets package's tiles")
		fit := textsafe.Fit(notice, g.cols)
		g.write((g.cols-textsafe.Width(fit))/2, g.rows/2, fit, uint8(colour.Notice))
	}
	if in.Stale {
		mark := textsafe.Fit(textsafe.Const(staleMark), g.cols)
		g.write(g.cols-textsafe.Width(mark), 0, mark, uint8(colour.Stale))
	}
	credit := textsafe.Fit(in.Credit, g.cols)
	if w := textsafe.Width(credit); w > 0 {
		g.write(g.cols-w, g.rows-1, credit, uint8(colour.Credit))
	}
	if in.Scale {
		g.write(0, g.rows-1, scaleMark(in.View, g.cols/3), uint8(colour.Scale))
	}
	if footer := textsafe.Fit(in.Footer, g.cols); textsafe.Width(footer) > 0 && g.rows > 1 {
		g.write(0, g.rows-2, footer, uint8(colour.Credit))
	}
}

// staleMark is what a frame says when the data of an overlay on it is no
// longer current: a plain word, not a colour alone (FR-32, NFR-15).
const staleMark = "stale"

// RoundBar is the round distance a bar of at most a given width stands for,
// and how many cells long it is: the one rule the mark on the frame and the
// scale a host asks for are both drawn by, so that they never disagree.
func RoundBar(perCol float64, room int) (float64, int) {
	km := 1.0
	for _, step := range []float64{5000, 2000, 1000, 500, 200, 100, 50, 20, 10, 5, 2, 1} {
		if step/perCol <= float64(room-8) {
			km = step
			break
		}
	}
	return km, max(int(math.Round(km/perCol)), 2)
}

// scaleMark is a bar of whole cells and the round distance it spans.
func scaleMark(v project.View, room int) textsafe.Text {
	perCol, _, err := v.CellSpanKm()
	if err != nil || !(perCol > 0) || room < 8 {
		return textsafe.Text{}
	}
	km, cells := RoundBar(perCol, room)
	return textsafe.Clean("\u251C" + strings.Repeat("\u2500", cells-2) + "\u2524 " + strconv.Itoa(int(km)) + " km")
}

// colours are a cell's foreground and background: the area's colour over the
// ground; the ink's own colour where it reads on that, else black or white
// (FR-16, D-77).
func (r *Renderer) colours(c cell, in Input, groundColour colour.RGB, kind colour.GroundKind) (fg, bg colour.RGB) {
	// The background, bottom to top: the ground; water; a field or an image;
	// an alert's tint (L2 Render, steps 1 to 4).
	bg = groundColour
	tinted := colour.Token(c.area) >= colour.AlertExtremeOutline && colour.Token(c.area) <= colour.AlertUnknownTint
	if area, ok := r.painter.Colour(c.area, in.Palette, kind, in.Depth); ok {
		bg = area
	}
	if under, ok := r.painter.Colour(c.under, in.Palette, kind, in.Depth); ok && !tinted {
		bg = under
	}
	own, ok := r.painter.Colour(c.ink, in.Palette, kind, in.Depth)
	if !ok {
		own = bg
	}
	need := colour.LineContrast
	if c.strict {
		need = colour.TextContrast
	}
	// The rule is applied to what the palette will show, so that the contrast
	// on the screen is the contrast that was checked (FR-16). At 16 colours
	// that is the reference table, and the figure is indicative only.
	switch in.Depth {
	case colour.Colours256:
		_, own = colour.To256(own)
		_, bg = colour.To256(bg)
	case colour.Colours16:
		_, own = colour.ToSixteen(own)
		_, bg = colour.ToSixteen(bg)
	}
	return colour.Foreground(own, bg, need), bg
}

// sgrColour writes one colour: "38" for the foreground or "48" for the
// background, as truecolor, or at 256 colours as the nearest fixed entry.
func sgrColour(b *strings.Builder, which string, c colour.RGB, depth Depth) {
	b.WriteString("\x1b[")
	if depth == colour.Colours16 {
		index, _ := colour.ToSixteen(c)
		code := 30 + int(index) // 30 to 37, then 90 to 97 for the bright eight
		if index >= 8 {
			code = 90 + int(index) - 8
		}
		if which == "48" {
			code += 10
		}
		b.WriteString(strconv.Itoa(code))
		b.WriteByte('m')
		return
	}
	b.WriteString(which)
	if depth == colour.Colours256 {
		index, _ := colour.To256(c)
		b.WriteString(";5;")
		b.WriteString(strconv.Itoa(int(index)))
		b.WriteByte('m')
		return
	}
	b.WriteString(";2;")
	b.WriteString(strconv.Itoa(int(c.R)))
	b.WriteByte(';')
	b.WriteString(strconv.Itoa(int(c.G)))
	b.WriteByte(';')
	b.WriteString(strconv.Itoa(int(c.B)))
	b.WriteByte('m')
}

// pen is the colours last sent on the row being emitted.
type pen struct {
	fg, bg colour.RGB
	fresh  bool // nothing has been sent on this row yet
}

// ground is the ground in effect for one frame.
type ground struct {
	colour  colour.RGB
	painted bool
	kind    colour.GroundKind
}

// The two numbers of the 64-bit FNV-1a sum.
const (
	sumStart = 14695981039346656037
	sumPrime = 1099511628211
)

func mix(sum uint64, v uint64) uint64 {
	for range 8 {
		sum = (sum ^ (v & 0xff)) * sumPrime
		v >>= 8
	}
	return sum
}

// salt is a sum of everything but the cells that decides a row's bytes.
func (r *Renderer) salt(in Input, under ground) uint64 {
	sum := mix(sumStart, in.Look)
	sum = mix(sum, uint64(in.Depth)<<32|uint64(under.kind)<<24|uint64(under.colour.R)<<16|uint64(under.colour.G)<<8|uint64(under.colour.B))
	if under.painted {
		sum = mix(sum, 1)
	}
	for _, c := range r.painter.literals {
		sum = mix(sum, uint64(c.R)<<16|uint64(c.G)<<8|uint64(c.B))
	}
	return sum
}

// rowSum is a sum of one row's cells.
func (g *grid) rowSum(row int, sum uint64) uint64 {
	for _, c := range g.cells[row*g.cols : (row+1)*g.cols] {
		flags := uint64(0)
		if c.taken {
			flags |= 1
		}
		if c.strict {
			flags |= 2
		}
		sum = mix(sum, uint64(c.glyph)<<32|uint64(c.under)<<24|uint64(c.ink)<<16|uint64(c.area)<<8|flags)
		for i := range len(c.text) {
			sum = (sum ^ uint64(c.text[i])) * sumPrime
		}
	}
	return sum
}

// emit writes the rows. Each stands alone: it sets its colours afresh, sends
// a colour again only when it changes, and puts the terminal's own back at
// its end (P-06, P-07). With no colour it is glyphs alone. A row whose cells
// are what they were in the frame held is not built again.
func (r *Renderer) emit(in Input) []string {
	g := r.grid
	under := ground{kind: in.Ground.Kind(in.Palette)}
	under.colour, under.painted = in.Ground.InEffect(in.Palette)
	salt := r.salt(in, under)
	if len(r.held.Lines) != g.rows {
		r.held.Lines, r.rowSums = make([]string, g.rows), make([]uint64, g.rows)
	}
	lines := r.held.Lines
	for row := range g.rows {
		sum := g.rowSum(row, salt)
		if sum == r.rowSums[row] && lines[row] != "" {
			continue
		}
		r.rowSums[row] = sum
		r.rowsBuilt++
		r.line.Reset()
		p := pen{fresh: true}
		for col := range g.cols {
			c := g.cells[row*g.cols+col]
			if c.taken && c.text == "" {
				continue // the second half of a wide character
			}
			if !colourless(in.Depth) {
				r.colourCell(c, in, under, &p)
			}
			if c.text != "" {
				r.line.WriteString(c.text)
				continue
			}
			r.line.WriteRune(c.glyph)
		}
		if !colourless(in.Depth) {
			r.line.WriteString("\x1b[0m")
		}
		lines[row] = r.line.String()
	}
	return lines
}

// colourCell sends what of a cell's colours differs from the pen's.
func (r *Renderer) colourCell(c cell, in Input, under ground, p *pen) {
	if p == nil {
		return
	}
	fg, bg := r.colours(c, in, under.colour, under.kind)
	if c.ink == 0 && !p.fresh {
		fg = p.fg // an empty cell shows no foreground: nothing to change
	}
	if p.fresh || fg != p.fg {
		sgrColour(&r.line, "38", fg, in.Depth)
	}
	shown := under.painted || c.area != 0 || c.under != 0
	if shown && (p.fresh || bg != p.bg) {
		sgrColour(&r.line, "48", bg, in.Depth)
	}
	if !shown && !p.fresh && p.bg != under.colour {
		r.line.WriteString("\x1b[49m") // back to the terminal's own, which the host declared
	}
	p.fg, p.bg, p.fresh = fg, bg, false
}
