package render

import (
	"math"

	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// RunLength is how many vertices one box of a run index covers (D-92). It is
// the overlay store's own run length: the index it builds is read here.
const RunLength = 64

// Borrowed is an overlay drawn straight from the host's own memory, because
// even simplified it would not fit the shape cache (D-90, D-92). Its rings
// are the host's own slices, read and never copied (FR-11), and its index is
// one box for each run of 64 vertices, ring after ring, which is what makes
// the drawing cheap: one box test a run, and only the runs the view meets
// are read at all.
type Borrowed struct {
	Kind  scene.ShapeKind
	Rings [][]project.LonLat
	Index []scene.Run
	Role  uint8
	Label string
	// Mark and Word are an alert area's severity digit and word, as a cached
	// shape carries them (D-65, D-108); empty for anything else.
	Mark, Word string
	Overlay    string
}

// Reads is the work one frame did drawing from the host's memory: the bound
// the constants state is one box test a run, plus the vertices of the runs
// whose box meets the view.
type Reads struct {
	Boxes    int
	Vertices int
}

// Reads is what this frame has read from the host's memory so far.
func (p *Painter) Reads() Reads {
	if p == nil {
		return Reads{}
	}
	return p.reads
}

// seen is the view's own box in the world's fractions, grown by the clip pad
// so that a line just outside is still drawn through the rectangle.
func seen(v project.View) (scene.Run, bool) {
	w, h := float64(v.Cols*2), float64(v.Rows*4)
	topLeft, err := v.FromDot(-clipMargin, -clipMargin)
	if err != nil {
		return scene.Run{}, false
	}
	bottomRight, err := v.FromDot(w+clipMargin, h+clipMargin)
	if err != nil {
		return scene.Run{}, false
	}
	x0, y0, err := project.ToTile(project.LonLat{Lon: topLeft.Lon, Lat: topLeft.Lat}, 0)
	if err != nil {
		return scene.Run{}, false
	}
	x1, y1, err := project.ToTile(project.LonLat{Lon: bottomRight.Lon, Lat: bottomRight.Lat}, 0)
	if err != nil {
		return scene.Run{}, false
	}
	return scene.Run{MinX: whole(min(x0, x1)), MinY: whole(min(y0, y1)), MaxX: whole(max(x0, x1)), MaxY: whole(max(y0, y1))}, true
}

// whole is a fraction of the world as the index holds it.
func whole(v float64) uint32 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return math.MaxUint32
	}
	return uint32(v * (1 << 32))
}

// meets reports whether two boxes of the world overlap.
func meets(a, b scene.Run) bool {
	return a.MinX <= b.MaxX && b.MinX <= a.MaxX && a.MinY <= b.MaxY && b.MinY <= a.MaxY
}

// Borrow draws one overlay from the host's memory. It tests one box for each
// run of the index and reads only the runs whose box meets the view; with no
// index it reads the whole of the geometry, which is what a small overlay
// with no index is for.
func (p *Painter) Borrow(v project.View, b Borrowed) error {
	if p == nil {
		return badTile()
	}
	box, ok := seen(v)
	if !ok {
		return nil
	}
	// AN ALERT DRAWN FROM THE HOST'S MEMORY CARRIES ITS WORD AND DIGITS (D-108):
	// the stretches of its outline the frame reads are kept, as Shape keeps a
	// cached alert's rings, for the digit pass; its label is placed at the
	// middle of its whole box, read from the index, as a cached alert's is,
	// so a view inside the alert still places or reports it.
	p.rings, p.ring, p.starts = p.rings[:0], p.ring[:0], p.starts[:0]
	at := 0 // the run this ring begins at, counting through the whole index
	for _, ring := range b.Rings {
		runs := (len(ring) + RunLength - 1) / RunLength
		err := p.borrowRing(v, b, ring, box, at)
		if err != nil {
			return err
		}
		at += runs
	}
	if !keepsOutline(b) {
		return nil
	}
	shape := scene.Shape{Kind: b.Kind, Role: b.Role, Label: b.Label, Mark: b.Mark, Word: b.Word, Overlay: b.Overlay}
	if len(p.ring) > 0 {
		p.keptStretches()
		p.keepOutline(shape, false)
	}
	if whole, ok := indexBox(v, b.Index, box); ok {
		p.keepLabel(shape, whole, false)
	} else if len(p.ring) > 0 {
		p.keepLabel(shape, p.keptBox(), false)
	}
	return nil
}

// keepsOutline reports whether a borrowed shape's outline is kept for its
// word and digits: an alert's area (D-108).
func keepsOutline(b Borrowed) bool {
	return alertRole(b.Role) && b.Kind == scene.ShapeArea
}

// indexBox is the box, in dots, of every run of an index, and whether it
// meets the view's own box: the whole shape's box, with no vertex read.
func indexBox(v project.View, index []scene.Run, view scene.Run) ([4]int, bool) {
	if len(index) == 0 {
		return [4]int{}, false
	}
	whole := index[0]
	for _, r := range index[1:] {
		whole = scene.Run{MinX: min(whole.MinX, r.MinX), MinY: min(whole.MinY, r.MinY), MaxX: max(whole.MaxX, r.MaxX), MaxY: max(whole.MaxY, r.MaxY)}
	}
	if !meets(whole, view) {
		return [4]int{}, false
	}
	x, y, side, err := v.TilePlace(scene.TileID{})
	if err != nil {
		return [4]int{}, false
	}
	dot := func(f uint32, from float64) int { return toDot(from + float64(f)/(1<<32)*side) }
	return [4]int{dot(whole.MinX, x), dot(whole.MinY, y), dot(whole.MaxX, x), dot(whole.MaxY, y)}, true
}

// keptBox is the box, in dots, of the outline stretches kept.
func (p *Painter) keptBox() [4]int {
	box := [4]int{math.MaxInt, math.MaxInt, math.MinInt, math.MinInt}
	for _, pt := range p.ring {
		box = [4]int{min(box[0], pt.X), min(box[1], pt.Y), max(box[2], pt.X), max(box[3], pt.Y)}
	}
	return box
}

// keepDot adds a dot to the outline stretch being kept, or begins a new one.
func (p *Painter) keepDot(pt Point, fresh bool) {
	if !fresh && len(p.ring) > 0 && p.ring[len(p.ring)-1] == pt {
		return
	}
	if fresh {
		p.starts = append(p.starts, len(p.ring))
	}
	p.ring = append(p.ring, pt)
}

// keptStretches cuts the dots kept into their stretches, as rings.
func (p *Painter) keptStretches() {
	p.rings = p.rings[:0]
	for i, start := range p.starts {
		end := len(p.ring)
		if i+1 < len(p.starts) {
			end = p.starts[i+1]
		}
		p.rings = append(p.rings, p.ring[start:end:end])
	}
}

// borrowRing draws the runs of one ring whose box meets the view.
func (p *Painter) borrowRing(v project.View, b Borrowed, ring []project.LonLat, box scene.Run, at int) error {
	for start := 0; start < len(ring); start += RunLength {
		end := min(start+RunLength, len(ring)-1)
		if run := at + start/RunLength; run < len(b.Index) {
			p.reads.Boxes++
			if !meets(b.Index[run], box) {
				continue // nothing of this run is in the view: its vertices are never read
			}
		}
		err := p.borrowRun(v, b, ring[start:end+1])
		if err != nil {
			return err
		}
	}
	return nil
}

// borrowRun draws one run of vertices, splitting the line where it crosses
// the antimeridian rather than drawing a line the width of the world.
func (p *Painter) borrowRun(v project.View, b Borrowed, run []project.LonLat) error {
	p.reads.Vertices += len(run)
	p.lines.Forcing(true)
	defer p.lines.Forcing(false)
	was := Point{}
	have := false
	for i, at := range run {
		x, y, err := v.ToDot(at)
		if err != nil {
			return err
		}
		now := Point{X: toDot(x), Y: toDot(y)}
		if b.Kind == scene.ShapePoint {
			p.markBlock(now, b.Role)
			continue
		}
		joined := have && !crosses(run[i-1], at)
		if joined {
			p.lines.Line(was.X, was.Y, now.X, now.Y, 1, b.Role)
		}
		if keepsOutline(b) {
			p.keepDot(now, !joined) // a run, or a crossing of the antimeridian, begins a new stretch
		}
		was, have = now, true
	}
	return nil
}

// crosses reports whether a step between two places goes the short way round
// the world's edge: more than half a turn of longitude is the antimeridian,
// not a line across the map.
func crosses(a, b project.LonLat) bool {
	return math.Abs(b.Lon-a.Lon) > 180
}
