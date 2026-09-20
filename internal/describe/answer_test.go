package describe

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/project"
)

// TestUnitsFollowHost is plan task 11.9: kilometres or miles, Celsius or
// Fahrenheit, and **always stated**.
func TestUnitsFollowHost(t *testing.T) {
	km, word := Units{}.Distance(100)
	if km != 100 || word != "kilometres" {
		t.Errorf("by default a distance is %v %q", km, word)
	}
	miles, word := Units{Miles: true}.Distance(kmPerMile * 50)
	if math.Abs(miles-50) > 0.001 || word != "miles" {
		t.Errorf("in miles it is %v %q", miles, word)
	}
	c, word := Units{}.Temperature(20)
	if c != 20 || word != "degrees Celsius" {
		t.Errorf("by default a temperature is %v %q", c, word)
	}
	f, word := Units{Fahrenheit: true}.Temperature(20)
	if math.Abs(f-68) > 0.001 || word != "degrees Fahrenheit" {
		t.Errorf("in Fahrenheit it is %v %q", f, word)
	}
	// Every answer that carries a distance carries the word for its units.
	area := OfArea("home", "alerts", project.LonLat{Lon: 0, Lat: 0}, [][]project.LonLat{box(5, -5, 15, 5)}, Units{Miles: true})
	if area.Unit != "miles" {
		t.Errorf("an answer in miles says its units are %q", area.Unit)
	}
	if area.Distance <= 0 {
		t.Errorf("it is %v miles away", area.Distance)
	}
}

// TestSpeakable is plan task 11.10: nothing the description hands out is a
// character a speech engine reads as noise, and the compass words are in
// full rather than letters.
func TestSpeakable(t *testing.T) {
	for _, bad := range []string{"⠀", "a ░ shade", "─ a rule", "◉", "a\x1b[0m"} {
		if Speakable(bad) {
			t.Errorf("%q is said to be speakable", bad)
		}
	}
	for _, good := range []string{"north-east", "kilometres", "Tornado Warning", "Mérida", "12 degrees Celsius"} {
		if !Speakable(good) {
			t.Errorf("%q is said not to be speakable", good)
		}
	}
	// Every string of a real answer passes, and the compass is a word.
	answer := OfArea("home", "alerts", project.LonLat{Lon: 0, Lat: 0}, [][]project.LonLat{box(5, -5, 15, 5)}, Units{})
	for _, said := range answer.Said() {
		if !Speakable(said) {
			t.Errorf("an answer carries %q, which is not speakable", said)
		}
	}
	if answer.Compass != "east" {
		t.Errorf("the compass word is %q, want a word and not a letter", answer.Compass)
	}
	// A label with an escape in it is cleaned on the way out, not passed on.
	near := Near{Km: 5, Bearing: 45, Label: "a track\x1b[31m"}
	got := OfNear("home", "tracks", LineForm, near, true, Units{})
	if strings.Contains(got.Label, "\x1b") {
		t.Errorf("the label came out as %q", got.Label)
	}
	if got.Compass != "north-east" {
		t.Errorf("45 degrees is %q", got.Compass)
	}
}

// TestNoDataSaidPlainly is part of plan task 11.8: an overlay with nothing
// to say here says so, rather than answering zero.
func TestNoDataSaidPlainly(t *testing.T) {
	// An area with no rings at all.
	if got := OfArea("home", "alerts", project.LonLat{}, nil, Units{}); !got.NoData {
		t.Errorf("an area with no rings answered %+v", got)
	}
	// No point anywhere near.
	if got := OfNear("home", "lightning", PointsForm, Near{}, false, Units{}); !got.NoData {
		t.Errorf("no points at all answered %+v", got)
	}
	// A field that has no value here.
	if got := OfField("home", "temperature", Reading{NoData: true}, true, Units{}); !got.NoData {
		t.Errorf("a field with no data answered %+v", got)
	}
	// An image with no data here still says where the rain is.
	heavier := Heavier{Class: 3, Km: 40, Bearing: 180, Found: true}
	got := OfImage("home", "radar", 0, false, heavier, Units{})
	if !got.NoData {
		t.Error("an image with no data here did not say so")
	}
	if got.Heavier != "south" || got.HeavierAt <= 0 {
		t.Errorf("it says the nearest heavier is %q at %v", got.Heavier, got.HeavierAt)
	}
}

// TestFieldAnswerCarriesItsUnits is the rest of 11.9: a temperature field
// is converted and stated; a field of something else keeps its own numbers
// and says nothing about units it does not know.
func TestFieldAnswerCarriesItsUnits(t *testing.T) {
	reading := Reading{Value: 20, Band: 2, Rises: 90, Rising: true}
	warm := OfField("home", "temperature", reading, true, Units{Fahrenheit: true})
	if math.Abs(warm.Value-68) > 0.001 || warm.ValueUnit != "degrees Fahrenheit" {
		t.Errorf("%v %q", warm.Value, warm.ValueUnit)
	}
	if warm.Rises != "east" {
		t.Errorf("it rises %q", warm.Rises)
	}
	other := OfField("home", "pressure", reading, false, Units{Fahrenheit: true})
	if other.Value != 20 || other.ValueUnit != "" {
		t.Errorf("a field the library knows no units for answered %v %q", other.Value, other.ValueUnit)
	}
}

// TestRoundedAndUnderOneCell: numbers are said to a length a voice can
// carry, and the description says when the picture cannot settle a question.
func TestRoundedAndUnderOneCell(t *testing.T) {
	for _, c := range []struct{ in, want float64 }{{12.37, 12}, {12.6, 13}, {1.234, 1.2}, {0.04, 0}, {-3.55, -3.6}} {
		if got := Rounded(c.in); math.Abs(got-c.want) > 0.001 {
			t.Errorf("%v rounds to %v, want %v", c.in, got, c.want)
		}
	}
	if !UnderOneCell(1.2, 1.6) {
		t.Error("1.2 km is not under a cell of 1.6 km")
	}
	if UnderOneCell(3, 1.6) {
		t.Error("3 km is under a cell of 1.6 km")
	}
	if UnderOneCell(1, 0) {
		t.Error("a view with no cell size settles anything")
	}
	_ = time.Now
}
