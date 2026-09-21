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
	at := 0 // the run this ring begins at, counting through the whole index
	for _, ring := range b.Rings {
		runs := (len(ring) + RunLength - 1) / RunLength
		err := p.borrowRing(v, b, ring, box, at)
		if err != nil {
			return err
		}
		at += runs
	}
	return nil
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
		if have && !crosses(run[i-1], at) {
			p.lines.Line(was.X, was.Y, now.X, now.Y, 1, b.Role)
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
