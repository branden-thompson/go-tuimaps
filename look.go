package tuimaps

import (
	"maps"
	"os"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/render"
	"github.com/branden-thompson/go-tuimaps/internal/style"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// RGB is a colour: eight bits a channel, as a terminal takes them.
type RGB = colour.RGB

// Depth is how much colour a terminal has.
type Depth = colour.Depth

// The colour depths. A map is drawn at the depth the host hints at; with no
// hint from the host, a non-empty NO_COLOR in the environment chooses no
// colour at all, and otherwise the map draws in truecolor (NFR-15).
const (
	Truecolor  = colour.Truecolor
	Colours256 = colour.Colours256
	Colours16  = colour.Colours16
	NoColour   = colour.NoColour
)

// Layer is a basemap layer a host can switch off and on (FR-36).
type Layer = style.Layer

// The layers that can be switched.
const (
	RoadLayer   = style.RoadLayer
	RailLayer   = style.RailLayer
	ParkLayer   = style.ParkLayer
	BorderLayer = style.BorderLayer
	RiverLayer  = style.RiverLayer
	WaterLayer  = style.WaterLayer
	LabelLayer  = style.LabelLayer
)

// look is everything a host has said about how the map should be drawn. It
// is read at the next Render and re-parses no tile, except that a change of
// language re-fetches, since a language is part of a tile's cache key (D-82).
type look struct {
	palette     colour.Palette
	version     uint64 // raised whenever the palette changes, since palettes cannot be compared
	ground      colour.GroundChoice
	depth       Depth
	depthHinted bool // the host said; otherwise the environment decides (NFR-15)
	safeRamps   bool
	reduce      bool
	off         style.Switches
	language    string
}

// languageMax is the longest language code the library takes: enough for a
// tag such as "pt-BR", and short enough to be a code and not a sentence.
const languageMax = 8

func badLanguage() error {
	return fault.Make(fault.InvalidID, textsafe.Const("the label language was refused"),
		textsafe.Const("a language is a short code, such as en, de or pt-BR"),
		textsafe.Const("pass a code of at most eight letters, digits or hyphens"))
}

// SetPalette gives tokens colours of the host's own. It answers with the
// names that are no token of the library's, sorted, which are ignored; the
// rest take effect at the next Render, and nothing is decoded again for
// them (FR-15). Passing no names puts the library's own colours back.
func (m *Map) SetPalette(tokens map[string]RGB) ([]string, error) {
	if m == nil {
		return nil, closed()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return nil, closed()
	}
	palette, unknown := colour.NewPalette(maps.Clone(tokens))
	m.look.palette = palette.KeepRamps(m.look.safeRamps)
	m.look.version++
	m.changed++
	// The names come back cleaned, since everything the library hands back is
	// (FR-34). They are matched as they were written: a name with an escape in
	// it is no token, and cleaning it first would make it one by accident.
	out := make([]string, 0, len(unknown))
	for _, name := range unknown {
		out = append(out, textsafe.Clean(name).String())
	}
	return out, nil
}

// SafeRamps says whether a ramp's colours are the library's own whatever the
// host set, which is the setting to reach for when a palette has not been
// checked (D-63).
func (m *Map) SafeRamps(on bool) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut || m.look.safeRamps == on {
		return
	}
	m.look.safeRamps = on
	m.look.palette = m.look.palette.KeepRamps(on)
	m.look.version++
	m.changed++
}

// Ground says what colour is behind the map, and that the library is not to
// paint it (D-64). By default the library paints every cell's background
// from the ground token, so that the map reads the same on any terminal.
func (m *Map) Ground(behind RGB) error {
	if m == nil {
		return closed()
	}
	choice, err := colour.Declared(behind)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return closed()
	}
	m.look.ground = choice
	m.changed++
	return nil
}

// PaintGround puts the default back: the library paints the ground itself.
func (m *Map) PaintGround() {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return
	}
	m.look.ground = colour.GroundChoice{}
	m.changed++
}

// ColourDepth hints at how much colour the terminal has. With no hint the
// library chooses from the usual settings of the environment.
func (m *Map) ColourDepth(depth Depth) {
	if m == nil {
		return
	}
	if depth > NoColour {
		return // a depth that is none of the four is no hint at all
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut || (m.look.depthHinted && m.look.depth == depth) {
		return
	}
	m.look.depth, m.look.depthHinted = depth, true
	m.changed++
}

// depthInEffect is the depth this frame is drawn at: the host's hint, or
// what the environment says when the host has not hinted (NFR-15).
func (m *Map) depthInEffect() Depth {
	return colour.ChooseDepth(m.look.depth, m.look.depthHinted, os.Getenv)
}

// ReduceMotion stops the library animating: markers are drawn steadily and
// nothing is ever due on the clock (NFR-21).
func (m *Map) ReduceMotion(on bool) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut || m.look.reduce == on {
		return
	}
	m.look.reduce = on
	m.motion.Reduce(on)
	m.changed++
}

// Layers switches a basemap layer off or on (FR-36). A layer switched off is
// never drawn, whatever else the frame would have drawn.
func (m *Map) Layers(layer Layer, on bool) {
	if m == nil {
		return
	}
	if layer < RoadLayer || layer > LabelLayer {
		return // a layer the library does not draw is nothing to switch
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return
	}
	was := m.look.off
	if on {
		m.look.off &^= style.Off(layer)
	} else {
		m.look.off |= style.Off(layer)
	}
	if m.look.off != was {
		m.changed++
	}
}

// LabelLanguage sets which language names are kept in. It is the one look
// setting that reaches the tiles: a language is part of a tile's cache key,
// so a change of language wants the tiles again (D-82).
func (m *Map) LabelLanguage(code string) error {
	if m == nil {
		return closed()
	}
	if code == "" || len(code) > languageMax || textsafe.Clean(code).String() != code {
		return badLanguage()
	}
	for _, r := range code {
		if r != '-' && !(r >= 'a' && r <= 'z') && !(r >= 'A' && r <= 'Z') && !(r >= '0' && r <= '9') {
			return badLanguage()
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return closed()
	}
	if m.look.language == code {
		return nil
	}
	m.look.language = code
	m.changed++
	// A language is part of a tile's cache key, so what is on hand is another
	// language's: the view wants its tiles again (D-82).
	return m.pipe.SetLanguage(code)
}

// Changed counts the times a redraw would have differed. A host that watches
// it knows when to call Render again without comparing frames itself.
func (m *Map) Changed() uint64 {
	if m == nil {
		return 0
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.changed
}

// paint fills in everything the look decides about one frame.
func (m *Map) paint(in *render.Input) {
	if m == nil || in == nil {
		return
	}
	in.Palette = m.look.palette
	in.Look = m.look.version
	in.Ground = m.look.ground
	in.Depth = m.depthInEffect()
	in.Layers = m.look.off
}
