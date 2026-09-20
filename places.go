package tuimaps

import (
	"math"
	"slices"
	"strconv"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/render"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// LonLat is a place on the world, in degrees.
type LonLat = project.LonLat

// MarkerStyle is how a place is drawn (P-58).
type MarkerStyle = render.MarkerShape

// The marker styles. A place that names none is drawn as a dot.
const (
	MarkerDot     = render.MarkerDot     // three dots by three
	MarkerCross   = render.MarkerCross   // arms of three dots each way
	MarkerDiamond = render.MarkerDiamond // radius three
	MarkerRing    = render.MarkerRing    // a circle of the radius given
	MarkerDisc    = render.MarkerDisc    // the same, filled
	MarkerGlyph   = render.MarkerGlyph   // a character of the host's own
)

// placeLimit is how many places one map holds. It is generous for "my
// places" and small enough that drawing them all is never the slow part.
const placeLimit = 4096

// Place is one of the host's own places: a name, a position, how it is
// drawn, and an id to remove it by (FR-26). A place is not an overlay: it is
// the host's "my places", not data. The least that has to be written is a
// name and a position.
type Place struct {
	ID     string
	Name   string
	At     LonLat
	Marker MarkerStyle
	Radius int    // a ring's and a filled circle's, in dots; three if not given
	Glyph  string // the character a glyph marker draws
	Blink  bool
}

func badPlace(why, todo textsafe.Text) error {
	return fault.Make(fault.InvalidCoordinates, textsafe.Const("a place was refused"), why, todo)
}

func badPlaceID(why, todo textsafe.Text) error {
	return fault.Make(fault.InvalidID, textsafe.Const("a place was refused"), why, todo)
}

// defaultID is the id a place with none of its own carries: its position
// written to six decimal places, latitude first, as upstream has it (P-61).
func defaultID(at LonLat) string {
	return strconv.FormatFloat(at.Lat, 'f', 6, 64) + "," + strconv.FormatFloat(at.Lon, 'f', 6, 64)
}

// checkPlace takes a place as the host wrote it and answers with the place
// the library keeps: its name cleaned, its id filled in.
func checkPlace(p Place) (Place, error) {
	if math.IsNaN(p.At.Lat) || math.IsNaN(p.At.Lon) || math.IsInf(p.At.Lat, 0) || math.IsInf(p.At.Lon, 0) {
		return Place{}, badPlace(textsafe.Const("its latitude or longitude is not a number"), textsafe.Const("give it a position in degrees"))
	}
	if p.At.Lat < -90 || p.At.Lat > 90 || p.At.Lon < -180 || p.At.Lon > 180 {
		return Place{}, badPlace(textsafe.Const("its position is off the world"),
			textsafe.Const("latitude runs from -90 to 90 and longitude from -180 to 180"))
	}
	if p.Marker != 0 && (p.Marker < MarkerDot || p.Marker > MarkerGlyph) {
		return Place{}, badPlaceID(textsafe.Const("it asks for a marker the library does not draw"),
			textsafe.Const("use one of the marker styles the package names, or none at all"))
	}
	if p.ID != "" && textsafe.Clean(p.ID).String() != p.ID {
		return Place{}, badPlaceID(textsafe.Const("its id holds characters that would have to be cleaned"),
			textsafe.Const("give it an id of plain text, or none, in which case its position becomes its id"))
	}
	p.Name = textsafe.Clean(p.Name).String()
	p.Glyph = textsafe.Clean(p.Glyph).String()
	if p.ID == "" {
		p.ID = defaultID(p.At)
	}
	return p, nil
}

// SetPlaces replaces the host's places with these, and answers with the id
// of each in the order given. An id left empty is the position to six
// decimal places (P-61).
func (m *Map) SetPlaces(places []Place) ([]string, error) {
	if m == nil {
		return nil, closed()
	}
	if len(places) > placeLimit {
		return nil, badPlaceID(textsafe.Const("there are more places than one map holds"),
			textsafe.Const("a map holds four thousand of them; hand in the ones in view"))
	}
	kept := make([]Place, 0, len(places))
	ids := make([]string, 0, len(places))
	for _, p := range places {
		one, err := checkPlace(p)
		if err != nil {
			return nil, err
		}
		kept = append(kept, one)
		ids = append(ids, one.ID)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return nil, closed()
	}
	m.places = kept
	m.changed++
	return ids, nil
}

// AddPlace adds one place, or replaces the one already carrying its id.
func (m *Map) AddPlace(p Place) (string, error) {
	if m == nil {
		return "", closed()
	}
	one, err := checkPlace(p)
	if err != nil {
		return "", err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return "", closed()
	}
	if len(m.places) >= placeLimit {
		return "", badPlaceID(textsafe.Const("there are more places than one map holds"),
			textsafe.Const("a map holds four thousand of them; remove one before adding another"))
	}
	for i := range m.places {
		if m.places[i].ID == one.ID {
			m.places[i] = one
			m.changed++
			return one.ID, nil
		}
	}
	m.places = append(m.places, one)
	m.changed++
	return one.ID, nil
}

// RemovePlace removes every place carrying an id, as upstream does (P-61),
// and says how many went.
func (m *Map) RemovePlace(id string) (int, error) {
	if m == nil {
		return 0, closed()
	}
	if id == "" || textsafe.Clean(id).String() != id {
		return 0, badPlaceID(textsafe.Const("that is not an id any place can carry"),
			textsafe.Const("pass the id SetPlaces or AddPlace gave back"))
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return 0, closed()
	}
	was := len(m.places)
	m.places = slices.DeleteFunc(m.places, func(p Place) bool { return p.ID == id })
	gone := was - len(m.places)
	if gone > 0 {
		m.changed++
	}
	return gone, nil
}

// Places are the host's places as the library holds them: names cleaned and
// ids filled in, in the order they were given.
func (m *Map) Places() []Place {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return slices.Clone(m.places)
}

// markers are the places as the renderer draws them.
func (m *Map) markers() []render.Marker {
	m.drawnPlaces = m.drawnPlaces[:0]
	for _, p := range m.places {
		shape := p.Marker
		if shape == 0 {
			shape = MarkerDot
		}
		m.drawnPlaces = append(m.drawnPlaces, render.Marker{At: p.At, Shape: shape, Radius: p.Radius,
			Text: p.Glyph, Label: p.Name, Ink: uint8(colour.Marker), Blink: p.Blink})
	}
	return m.drawnPlaces
}
