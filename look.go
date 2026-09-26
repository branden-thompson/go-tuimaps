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
	// MinorRoadLayer is the minor roads, switched apart from RoadLayer's
	// motorways, trunks and primaries (v0.2.0 D-82).
	MinorRoadLayer = style.MinorRoadLayer
)

// Detail is how much of the basemap is drawn, by purpose (v0.2.0 D-82,
// L-14): each level draws what the one below does and more. Only the
// library's own style is ranked; a host's own style (SetStyle) draws whole.
type Detail = style.Detail

// The levels. DetailFull is the picture a host that never asks gets.
const (
	DetailEssential = style.DetailEssential // coast, water, borders
	DetailWeather   = style.DetailWeather   // and rivers, place names, the major roads
	DetailStandard  = style.DetailStandard  // and rail, parks
	DetailFull      = style.DetailFull      // and the minor roads, runways
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
	detail      style.Detail // the host's level; zero reads as Full (D-82)
	language    string
	blends      colour.Blends // how each alert tint blends over each image ramp (L-11.3)
	blendsFor   blendKey      // the look the search was run for
}

// blendKey is everything the blend search reads: it runs again only when one
// of them changes, never on a frame (L3.8).
type blendKey struct {
	searched bool
	palette  uint64
	ground   colour.GroundChoice
	depth    Depth
}

// languageMax is the longest language code the library takes: enough for a
// tag such as "pt-BR", and short enough to be a code and not a sentence.
const languageMax = 8

func badLanguage() error {
	return fault.Make(fault.InvalidID, textsafe.Const("the label language was refused"),
		textsafe.Const("a language is a short code, such as en, de or pt-BR"),
		textsafe.Const("pass a code of at most eight letters, digits or hyphens"))
}

// TokenNames are the names SetPalette accepts, in the order the constants
// document lists them. **A host theming the map to its own design system needs
// to know what there is to theme**, and without this the only way to find out
// was to guess a name and read it back from SetPalette's list of refusals
// (task 14.19, the first host).
func TokenNames() []string {
	all := colour.Tokens()
	out := make([]string, 0, len(all))
	for _, t := range all {
		if name := t.Name(); name != "" {
			out = append(out, name)
		}
	}
	return out
}

// SetPalette gives tokens colours of the host's own. It answers with the
// names that are no token of the library's, sorted, which are ignored; the
// rest take effect at the next Render, and nothing is decoded again for
// them (FR-15). Passing no names puts the library's own colours back.
func (m *Map) SetPalette(tokens map[string]RGB) (names []string, err error) {
	defer guard("SetPalette", &err)
	m.plant("SetPalette")

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
	defer m.guardQuiet("SafeRamps")
	m.plant("SafeRamps")

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
func (m *Map) Ground(behind RGB) (err error) {
	defer guard("Ground", &err)
	m.plant("Ground")

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
	defer m.guardQuiet("PaintGround")
	m.plant("PaintGround")

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
	defer m.guardQuiet("ColourDepth")
	m.plant("ColourDepth")

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

// ReduceMotion stops the library animating (NFR-21): markers are drawn
// steadily, and a loop playing stops where it is and cannot be played until
// it is turned off again. The clock still has work that is not motion: a
// stale overlay's moment and a failed tile's retry stay due.
func (m *Map) ReduceMotion(on bool) {
	defer m.guardQuiet("ReduceMotion")
	m.plant("ReduceMotion")

	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut || m.look.reduce == on {
		return
	}
	if on {
		m.holdLocked() // a loop playing stops where it is (L-1.10c)
	}
	m.look.reduce = on
	m.motion.Reduce(on)
	m.changed++
}

// Layers switches a basemap layer off or on (FR-36). A layer switched off is
// never drawn, whatever else the frame would have drawn.
func (m *Map) Layers(layer Layer, on bool) {
	defer m.guardQuiet("Layers")
	m.plant("Layers")

	if m == nil {
		return
	}
	if layer < RoadLayer || layer > MinorRoadLayer {
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
func (m *Map) LabelLanguage(code string) (err error) {
	defer guard("LabelLanguage", &err)
	m.plant("LabelLanguage")

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
	in.Detail = m.look.detail
	m.searchBlendsLocked(in.Depth)
	in.Blends = m.look.blends
}

// searchBlendsLocked runs the blend search when the palette, the ground or
// the depth has changed since it last ran (L3.8). A tint that a host's
// palette leaves unable to blend is reported: the image is drawn over it,
// and the warning is carried by the outline, label and severity digit.
func (m *Map) searchBlendsLocked(depth Depth) {
	key := blendKey{searched: true, palette: m.look.version, ground: m.look.ground, depth: depth}
	if m.look.blendsFor == key {
		return
	}
	colourOf, _ := m.look.ground.InEffect(m.look.palette)
	m.look.blends = colour.SearchBlends(m.look.palette, m.look.ground.Kind(m.look.palette), colourOf, depth)
	m.look.blendsFor = key
	if !m.look.palette.Own() || (depth != colour.Truecolor && depth != colour.Colours256) {
		return // the library's own colours: where they fall back is documented, not a warning
	}
	for _, f := range m.look.blends.Fallbacks() {
		if len(m.own) < 64 {
			m.own = append(m.own, fault.Warning{Kind: fault.RampRuleBroken, Count: 1,
				Subject: textsafe.Clean(presetNames[colour.Preset(f[0])] + " under the " + alertNames[f[1]] + " tint: drawn over it, not blended")})
		}
	}
}

// presetNames are the image presets by name.
var presetNames = map[colour.Preset]string{colour.Temperature: "temperature", colour.Radar: "radar"}

// alertNames are the alert severities, extreme first, as the tints are ordered.
var alertNames = [5]string{"extreme", "severe", "moderate", "minor", "unknown"}

// SetDetail sets how much of the basemap is drawn (v0.2.0 D-82, L-14): a
// level and a switched-off layer both apply. A value outside the four levels
// is refused and the map keeps the level it had.
func (m *Map) SetDetail(d Detail) (err error) {
	defer guard("SetDetail", &err)
	m.plant("SetDetail")

	if m == nil {
		return closed()
	}
	if !d.Valid() {
		return badView(textsafe.Const("that is not a detail the map has"),
			textsafe.Const("pass DetailEssential, DetailWeather, DetailStandard or DetailFull"))
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return closed()
	}
	if m.look.detail != d {
		m.look.detail = d
		m.changed++
	}
	return nil
}
