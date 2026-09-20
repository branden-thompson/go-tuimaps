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

// Depth is the colour depth a frame is emitted at.
type Depth uint8

// The depths built so far. Truecolor is the zero value.
const (
	Truecolor Depth = iota
	NoColour
)

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
	View    project.View
	Tiles   []Drawn // in any order: the renderer draws them in one
	Missing int     // tiles the view wants with nothing on hand to draw for them
	Style   *style.Style
	Palette colour.Palette
	// Look counts changes to the palette, which cannot be compared: whoever
	// changes the palette raises it, and a frame is reused only while it and
	// everything else here stay the same (contract, section 5).
	Look   uint64
	Ground colour.GroundChoice
	Depth  Depth
	Labels bool
	Scale  bool
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
	taken  bool  // text occupies the cell: a label, or the second half of a wide character
	strict bool  // the cell holds text, held to the text contrast
}

type box struct{ left, right, top, bottom int }

// grid is the frame's cells and the boxes of the labels placed so far.
type grid struct {
	cols, rows int
	cells      []cell
	boxes      []box
}

func newGrid(cols, rows int) (*grid, error) {
	if cols <= 0 || rows <= 0 || cols > maxCells || rows > maxCells || cols*rows > maxCells {
		return nil, fault.New(fault.NoSize, textsafe.Const("the map has no size to draw at"),
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
func (g *grid) label(l Label) bool {
	return g.labelAt(l, []Point{{X: l.X, Y: l.Y}})
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
	text := textsafe.Clean(l.Name)
	if l.Name == "" {
		text = textsafe.Clean(placeGlyph)
	}
	width := textsafe.Width(text)
	if width == 0 || at.X < 0 || at.Y < 0 {
		return false // above or left of the rectangle: never a negative row (P-33, L-8)
	}
	col, row := at.X/2-width/2, at.Y/4
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

// unchanged reports whether nothing a frame is a function of has changed
// since the frame held was drawn. It allocates nothing.
func (r *Renderer) unchanged(in Input) bool {
	if !r.drawn || len(in.Tiles) != len(r.lastTiles) {
		return false
	}
	l := r.last
	if in.View != l.View || in.Look != l.Look || in.Ground != l.Ground || in.Depth != l.Depth || in.Style != l.Style {
		return false
	}
	if in.Labels != l.Labels || in.Scale != l.Scale || in.Credit != l.Credit || in.Missing != l.Missing {
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
	return fault.New(fault.Internal, textsafe.Const("the frame could not be drawn"), why, textsafe.Const("this is a defect in the library; report it"))
}

// Render draws one frame. It reads only what it is given and never waits.
func (r *Renderer) Render(in Input) (Frame, error) {
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
	status := Complete
	if len(r.order) == 0 {
		return NoTiles, nil
	}
	err := r.painter.World(in.View)
	if err != nil {
		return status, err
	}
	if in.Missing > 0 {
		status = Sharpening
	}
	for _, d := range r.order {
		if d.Tile == nil {
			continue
		}
		if !d.Exact {
			status = Sharpening
		}
		err = r.painter.Tile(in.View, d.Tile, d.At, in.Style)
		if err != nil {
			return status, err
		}
	}
	return status, nil
}

// compose builds the cells: areas, line work, then names, then furniture.
func (r *Renderer) compose(in Input, status Status) {
	g := r.grid
	for row := range g.rows {
		for col := range g.cols {
			c := &g.cells[row*g.cols+col]
			c.glyph, c.ink = r.painter.Lines().Cell(col, row)
			if ink, owned := r.painter.Area(col, row); owned {
				c.area = ink
			}
		}
	}
	r.furniture(in, status)
	if !in.Labels {
		return
	}
	// Names last, most important first; the order among equals is the order
	// the tiles gave them, and the tiles were sorted (P-25).
	r.labels = append(r.labels[:0], r.painter.Labels()...)
	sort.SliceStable(r.labels, func(i, j int) bool { return r.labels[i].Rank < r.labels[j].Rank })
	for _, l := range r.labels {
		g.labelAt(l, r.painter.Points(l))
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
	credit := textsafe.Fit(in.Credit, g.cols)
	if w := textsafe.Width(credit); w > 0 {
		g.write(g.cols-w, g.rows-1, credit, uint8(colour.Credit))
	}
	if in.Scale {
		g.write(0, g.rows-1, scaleMark(in.View, g.cols/3), uint8(colour.Scale))
	}
}

// scaleMark is a bar of whole cells and the round distance it spans.
func scaleMark(v project.View, room int) textsafe.Text {
	perCol, _, err := v.CellSpanKm()
	if err != nil || !(perCol > 0) || room < 8 {
		return textsafe.Text{}
	}
	km := 1.0
	for _, step := range []float64{5000, 2000, 1000, 500, 200, 100, 50, 20, 10, 5, 2, 1} {
		if step/perCol <= float64(room-8) {
			km = step
			break
		}
	}
	cells := max(int(math.Round(km/perCol)), 2)
	return textsafe.Clean("\u251C" + strings.Repeat("\u2500", cells-2) + "\u2524 " + strconv.Itoa(int(km)) + " km")
}

// colours are a cell's foreground and background: the area's colour over the
// ground; the ink's own colour where it reads on that, else black or white
// (FR-16, D-77).
func (r *Renderer) colours(c cell, in Input, groundColour colour.RGB, kind colour.GroundKind) (fg, bg colour.RGB) {
	bg = groundColour
	if area, ok := r.painter.Colour(c.area, in.Palette, kind); ok {
		bg = area
	}
	own, ok := r.painter.Colour(c.ink, in.Palette, kind)
	if !ok {
		own = bg
	}
	need := colour.LineContrast
	if c.strict {
		need = colour.TextContrast
	}
	return colour.Foreground(own, bg, need), bg
}

func sgrColour(b *strings.Builder, lead string, c colour.RGB) {
	b.WriteString(lead)
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
		sum = mix(sum, uint64(c.glyph)<<32|uint64(c.ink)<<16|uint64(c.area)<<8|flags)
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
			if in.Depth != NoColour {
				r.colourCell(c, in, under, &p)
			}
			if c.text != "" {
				r.line.WriteString(c.text)
				continue
			}
			r.line.WriteRune(c.glyph)
		}
		if in.Depth != NoColour {
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
		sgrColour(&r.line, "\x1b[38;2;", fg)
	}
	shown := under.painted || c.area != 0
	if shown && (p.fresh || bg != p.bg) {
		sgrColour(&r.line, "\x1b[48;2;", bg)
	}
	if !shown && !p.fresh && p.bg != under.colour {
		r.line.WriteString("\x1b[49m") // back to the terminal's own, which the host declared
	}
	p.fg, p.bg, p.fresh = fg, bg, false
}
