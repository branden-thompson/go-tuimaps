package tuimaps

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/render"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/style"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
	"github.com/branden-thompson/go-tuimaps/internal/tiles"
	"github.com/branden-thompson/go-tuimaps/internal/work"
)

// queueLimit is how many jobs may wait for one map (constants, section 2).
const queueLimit = 256

// Size is a map's size in terminal cells.
type Size struct {
	Cols, Rows int
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
	if s < Complete || s > NoTiles {
		return "unknown"
	}
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

// Frame is a drawn map: one string a row, each exactly the map's width in
// cells. It is valid until the next Render.
type Frame struct {
	Lines  []string
	Status Status
}

// SettleResult is what one Settle did.
type SettleResult struct {
	Ran      int   // jobs run
	Failed   int   // of those, how many failed
	InFlight int   // jobs still inside a Work call elsewhere; Settle never waits for them
	Why      error // the first failure, so a result with nothing fetched says why
}

// Option is one choice made when a map is created.
type Option func(*config) error

type config struct {
	size     Size
	embedded tiles.EmbeddedFunc
	maxZoom  uint8
}

func noSize() error {
	return fault.Make(fault.NoSize,
		textsafe.Const("the map has no size"),
		textsafe.Const("its width and height in cells must each be at least 1, and together no more than a million cells"),
		textsafe.Const("create the map with WithSize, or give Render a size, before asking it to settle"))
}

func closed() error {
	return fault.Make(fault.Closed,
		textsafe.Const("the map is closed"),
		textsafe.Const("Close was called on it"),
		textsafe.Const("create another map"))
}

// WithSize sets the map's size in cells. The size is state the map has before
// the first Settle or Render, which is what lets Settle find work to do.
func WithSize(cols, rows int) Option {
	return func(c *config) error {
		if c == nil {
			return noSize()
		}
		if _, err := project.FitWorld(cols, rows); err != nil {
			return noSize()
		}
		if _, err := render.NewCanvas(cols, rows); err != nil {
			return noSize()
		}
		c.size = Size{Cols: cols, Rows: rows}
		return nil
	}
}

// Embed passes the embedded tiles to the map - tuimaps.Embed(assets.Tile,
// assets.MaxZoom) - so that it can draw with no network at all. Importing the assets package does nothing by itself. With
// no source named they are the map's tiles for the zooms they hold; with one
// named they only ever stand in until its tiles arrive.
func Embed(tile func(z uint8, x, y uint32) ([]byte, bool), maxZoom uint8) Option {
	return func(c *config) error {
		if c == nil || tile == nil {
			return fault.Make(fault.Internal, textsafe.Const("the embedded tiles were refused"),
				textsafe.Const("no function was passed to read them with"), textsafe.Const("pass assets.Tile and assets.MaxZoom"))
		}
		c.embedded, c.maxZoom = tile, maxZoom
		return nil
	}
}

// Map is one map. Its calls are the host's to make from one goroutine, except
// those the documentation says may be made from any.
type Map struct {
	mu       sync.Mutex
	shut     bool
	inside   atomic.Int32
	sized    bool
	view     project.View
	noted    project.View // the view whose tiles were last asked for
	pipe     *tiles.Pipeline
	member   *work.Member
	renderer *render.Renderer
	look     *style.Style
	drawn    []render.Drawn
}

// New creates a map. It starts nothing: no goroutine, no connection, no file.
func New(options ...Option) (*Map, error) {
	var c config
	for _, o := range options {
		if o == nil {
			continue
		}
		err := o(&c)
		if err != nil {
			return nil, err
		}
	}
	cache, err := tiles.NewCache(tiles.DefaultCacheBytes)
	if err != nil {
		return nil, err
	}
	pipe, err := tiles.NewPipeline(tiles.Options{Cache: cache, Embedded: c.embedded, EmbeddedMaxZoom: c.maxZoom})
	if err != nil {
		return nil, err
	}
	member, err := work.NewQueue(queueLimit).Join()
	if err != nil {
		return nil, err
	}
	m := &Map{pipe: pipe, member: member, look: style.BuiltIn()}
	if c.size != (Size{}) {
		err = m.resize(c.size)
	}
	return m, err
}

// resize gives the map a size. Until the view can be moved, a map shows the
// whole world at the size it has (P-56).
func (m *Map) resize(s Size) error {
	if s.Cols <= 0 || s.Rows <= 0 {
		return noSize()
	}
	if m.sized && m.view.Cols == s.Cols && m.view.Rows == s.Rows {
		return nil
	}
	v, err := project.FitWorld(s.Cols, s.Rows)
	if err != nil {
		return noSize()
	}
	r, err := render.NewRenderer(s.Cols, s.Rows)
	if err != nil {
		return noSize()
	}
	m.view, m.renderer, m.sized = v, r, true
	return nil
}

// note asks for what the view needs: the tiles to fetch now, those waiting
// out a failure, and nothing the view has left.
func (m *Map) note(now time.Time) error {
	if !m.sized {
		return noSize()
	}
	wanted, err := m.view.Tiles()
	if err != nil {
		return err
	}
	plan := m.pipe.Plan(now, wanted)
	if m.noted != m.view {
		m.member.NewView()
		m.noted = m.view
	}
	err = m.member.Keep(m.pipe.StillWanted)
	if err != nil {
		return err
	}
	for _, job := range plan.Jobs {
		err = m.member.Add(job)
		if err != nil {
			return err
		}
	}
	for _, later := range plan.Later {
		err = m.member.Defer(later.Job, later.NotBefore)
		if err != nil {
			return err
		}
	}
	return nil
}

// Pending is how many jobs wait for a Work or Settle call. It may be called
// from any goroutine.
func (m *Map) Pending() int {
	if m == nil {
		return 0
	}
	return m.member.Backlog()
}

// Settle notes what the map's view and size need, then works, on the caller's
// goroutine, until nothing is pending or ctx ends. It is for one-shot renders
// and tests; an interactive host calls Work from goroutines of its own.
func (m *Map) Settle(ctx context.Context) (SettleResult, error) {
	if m == nil {
		return SettleResult{}, closed()
	}
	if ctx == nil {
		return SettleResult{}, fault.Make(fault.Internal, textsafe.Const("the map could not settle"), textsafe.Const("it was given no context"), textsafe.Const("pass a context; context.Background() if there is no other"))
	}
	m.inside.Add(1)
	defer m.inside.Add(-1)
	m.mu.Lock()
	if m.shut {
		m.mu.Unlock()
		return SettleResult{}, closed()
	}
	if !m.sized {
		m.mu.Unlock()
		return SettleResult{}, noSize()
	}
	err := m.note(time.Time{})
	m.mu.Unlock()
	if err != nil {
		return SettleResult{}, err
	}
	res, err := m.member.Drain(ctx)
	return SettleResult{Ran: res.Ran, Failed: res.Failed, InFlight: res.InFlight, Why: res.Why}, err
}

// Render draws the map at a size, which becomes the map's size. now is the
// host's clock. It reads only what is on hand and never waits: what is
// missing is noted as wanted, and the frame's status says so.
func (m *Map) Render(size Size, now time.Time) (Frame, error) {
	if m == nil {
		return Frame{}, closed()
	}
	if size.Cols <= 0 || size.Rows <= 0 {
		return Frame{}, noSize()
	}
	m.inside.Add(1)
	defer m.inside.Add(-1)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return Frame{}, closed()
	}
	err := m.resize(size)
	if err != nil {
		return Frame{}, err
	}
	err = m.note(now)
	if err != nil {
		return Frame{}, err
	}
	in := render.Input{View: m.view, Style: m.look, Labels: true,
		Credit: textsafe.Const("OpenFreeMap (c) OpenMapTiles Data from OpenStreetMap")}
	in.Tiles, in.Missing = m.onHand()
	frame, err := m.renderer.Draw(in)
	if err != nil {
		return Frame{}, err
	}
	return Frame{Lines: frame.Lines, Status: Status(frame.Status)}, nil
}

// onHand is what can be drawn for the view now: each wanted tile, or what
// stands in for it, once each; and how many have nothing at all.
func (m *Map) onHand() ([]render.Drawn, int) {
	wanted, err := m.view.Tiles()
	if err != nil {
		return nil, 0
	}
	m.drawn = m.drawn[:0]
	missing := 0
	for _, id := range wanted {
		tile, at, exact, ok := m.pipe.Draw(id)
		if !ok {
			missing++
			continue
		}
		if !seen(m.drawn, at) {
			m.drawn = append(m.drawn, render.Drawn{Tile: tile, At: at, Exact: exact})
		}
	}
	return m.drawn, missing
}

func seen(drawn []render.Drawn, at scene.TileID) bool {
	for _, d := range drawn {
		if d.At == at {
			return true
		}
	}
	return false
}

// Close closes the map at once, and reports how many of the host's calls are
// still inside it.
func (m *Map) Close() int {
	if m == nil {
		return 0
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.shut {
		m.shut = true
		m.pipe.Release()
		m.member.Leave()
	}
	return int(m.inside.Load())
}
