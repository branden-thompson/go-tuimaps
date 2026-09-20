package describe

import (
	"math"
	"strings"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// kmPerMile is what a mile is, exactly.
const kmPerMile = 1.609344

// Units are the units a host asks its answers in. The zero value is
// kilometres and Celsius, and **whichever is in use is always stated**: an
// answer that says "twelve" and leaves the rest to the reader is no answer.
type Units struct {
	Miles      bool
	Fahrenheit bool
}

// Distance is a distance in the host's units, and the word for them.
func (u Units) Distance(km float64) (float64, string) {
	if u.Miles {
		return km / kmPerMile, "miles"
	}
	return km, "kilometres"
}

// Temperature is a temperature in the host's units, and the word for them.
func (u Units) Temperature(celsius float64) (float64, string) {
	if u.Fahrenheit {
		return celsius*9/5 + 32, "degrees Fahrenheit"
	}
	return celsius, "degrees Celsius"
}

// Form is the shape an overlay takes, in the word a host would say.
type Form string

// The forms. They are words, not numbers, because they are said aloud.
const (
	AreaForm   Form = "area"
	PointsForm Form = "points"
	LineForm   Form = "line"
	FieldForm  Form = "field"
	ImageForm  Form = "image"
)

// Answer is what the library says about one place and one overlay: data a
// host renders or speaks, never a sentence of the library's own (D-52).
type Answer struct {
	Place   string
	Overlay string
	Form    Form

	// Where the place is, for an area: inside or outside, with the nearest
	// edge. For points and lines, the nearest one, with its label.
	Relation Where
	Distance float64
	Unit     string
	Bearing  float64
	Compass  string
	Label    string

	// What the data says here, for a field or an image.
	Value     float64
	ValueUnit string
	Band      int
	Rises     string // the compass word the values rise towards, or empty
	Class     int
	Heavier   string // the compass word towards the nearest heavier class
	HeavierAt float64
	NoData    bool

	// What is true of the overlay itself, whatever its shape.
	Valid        time.Time
	Stale        bool
	UnderOneCell bool // the picture cannot settle this: the words must (D-67)
}

// compassOf is the word for a bearing, or empty where there is none.
func compassOf(bearing float64) string {
	word, err := project.Compass(bearing)
	if err != nil {
		return ""
	}
	return word.String()
}

// OfArea is the answer for one place against one area overlay.
func OfArea(place, overlay string, at project.LonLat, rings [][]project.LonLat, u Units) Answer {
	out := Answer{Place: place, Overlay: overlay, Form: AreaForm, Relation: InArea(at, rings)}
	edge, ok := NearestEdge(at, rings)
	if !ok {
		out.NoData = true
		return out
	}
	out.Distance, out.Unit = u.Distance(edge.Km)
	out.Bearing, out.Compass = edge.Bearing, compassOf(edge.Bearing)
	return out
}

// OfNear is the answer for one place against a set of points or lines.
func OfNear(place, overlay string, form Form, near Near, ok bool, u Units) Answer {
	out := Answer{Place: place, Overlay: overlay, Form: form}
	if !ok {
		out.NoData = true
		return out
	}
	out.Distance, out.Unit = u.Distance(near.Km)
	out.Bearing, out.Compass = near.Bearing, compassOf(near.Bearing)
	out.Label = textsafe.Clean(near.Label).String()
	return out
}

// OfField is the answer for one place against a scalar field.
func OfField(place, overlay string, reading Reading, temperature bool, u Units) Answer {
	out := Answer{Place: place, Overlay: overlay, Form: FieldForm, Band: reading.Band}
	if reading.NoData {
		out.NoData = true
		return out
	}
	out.Value, out.ValueUnit = reading.Value, ""
	if temperature {
		out.Value, out.ValueUnit = u.Temperature(reading.Value)
	}
	if reading.Rising {
		out.Rises = compassOf(reading.Rises)
	}
	return out
}

// OfImage is the answer for one place against a classified image.
func OfImage(place, overlay string, class int, here bool, heavier Heavier, u Units) Answer {
	out := Answer{Place: place, Overlay: overlay, Form: ImageForm, Class: class, NoData: !here}
	if heavier.Found {
		out.HeavierAt, out.Unit = u.Distance(heavier.Km)
		out.Heavier = compassOf(heavier.Bearing)
		out.Bearing = heavier.Bearing
	}
	return out
}

// Speakable reports whether a piece of text is one a speech engine says as
// it stands: no braille, no block or box drawing, no shades, and nothing a
// terminal would read as a command (FR-34, NFR-15). Everything the library
// hands out of this package passes it.
func Speakable(s string) bool {
	if textsafe.Clean(s).String() != s {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 0x2800 && r <= 0x28FF: // braille
			return false
		case r >= 0x2500 && r <= 0x257F: // box drawing
			return false
		case r >= 0x2580 && r <= 0x259F: // blocks and shades
			return false
		case r >= 0x25A0 && r <= 0x25FF: // geometric shapes: the marker glyphs
			return false
		case r == 0xFFFD:
			return false
		}
	}
	return true
}

// Said is every piece of text in an answer, for a host or a test that wants
// to check what it is about to say.
func (a Answer) Said() []string {
	return []string{a.Place, a.Overlay, string(a.Form), a.Unit, a.Compass, a.Label, a.ValueUnit, a.Rises, a.Heavier, a.Relation.String()}
}

// Cleaned is the answer with every piece of its text cleaned, which is how
// it leaves the library (FR-34).
func (a Answer) Cleaned() Answer {
	a.Place = textsafe.Clean(a.Place).String()
	a.Overlay = textsafe.Clean(a.Overlay).String()
	a.Label = textsafe.Clean(a.Label).String()
	return a
}

// Rounded is a number said to a sensible number of places: a distance in
// whole units once it is above ten, and one decimal place below that. A
// voice reading "twelve point three seven kilometres" has said nothing more
// than "twelve kilometres".
func Rounded(v float64) float64 {
	if math.Abs(v) >= 10 {
		return math.Round(v)
	}
	return math.Round(v*10) / 10
}

// UnderOneCell reports whether a distance is smaller than one cell of the
// view it is drawn in, which is when the picture cannot settle the question
// and the words have to (D-67, PL-AX-4).
func UnderOneCell(km, cellKm float64) bool {
	return cellKm > 0 && km < cellKm
}

// OneLine is the plainest possible rendering, for a host that wants a line
// rather than data. It is a convenience, not the contract: the data is.
func OneLine(parts ...string) string {
	kept := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			kept = append(kept, p)
		}
	}
	return strings.Join(kept, ", ")
}
