package tuimaps_test

import (
	"strings"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// miami is a place with a name and a position, and nothing else: the least a
// host has to write (D-69).
func miami() tuimaps.Place {
	return tuimaps.Place{Name: "Miami", At: tuimaps.LonLat{Lon: -80.19, Lat: 25.77}}
}

// TestSetPlaces is plan task 12.26 (FR-26): the host's own places are drawn
// as markers, with their names, and setting them again replaces the lot.
func TestSetPlaces(t *testing.T) {
	const cols, rows = 149, 38
	m := world(t, cols, rows)
	_, without := drawn(t, m, cols, rows)
	ids, err := m.SetPlaces([]tuimaps.Place{miami(), {Name: "Reykjavik", At: tuimaps.LonLat{Lon: -21.94, Lat: 64.15}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Fatalf("setting two places gave %d ids", len(ids))
	}
	_, with := drawn(t, m, cols, rows)
	if with == without {
		t.Error("the frame is unchanged with two places set")
	}
	if strings.Count(with, "Reykjavik") == 0 {
		t.Errorf("the place's name is not on the frame:\n%s", with)
	}
	// Setting again replaces: the first two are gone.
	if _, err := m.SetPlaces(nil); err != nil {
		t.Fatal(err)
	}
	_, empty := drawn(t, m, cols, rows)
	if strings.Contains(empty, "Reykjavik") {
		t.Error("a place outlived the set that replaced it")
	}
	if len(m.Places()) != 0 {
		t.Errorf("%d places after they were all replaced by none", len(m.Places()))
	}
}

// TestParityP61_MarkerId is the parity row of the same name: an id left
// empty is the position written to six decimal places, and RemovePlace
// removes every place carrying that id.
func TestParityP61_MarkerId(t *testing.T) {
	m := world(t, 80, 24)
	ids, err := m.SetPlaces([]tuimaps.Place{miami()})
	if err != nil {
		t.Fatal(err)
	}
	if want := "25.770000,-80.190000"; ids[0] != want {
		t.Errorf("an id left empty is %q, want %q: the position to six decimal places", ids[0], want)
	}
	// Three places, two of them sharing an id.
	_, err = m.SetPlaces([]tuimaps.Place{
		{ID: "base", Name: "One", At: tuimaps.LonLat{Lon: 0, Lat: 0}},
		{ID: "base", Name: "Two", At: tuimaps.LonLat{Lon: 1, Lat: 1}},
		{ID: "other", Name: "Three", At: tuimaps.LonLat{Lon: 2, Lat: 2}},
	})
	if err != nil {
		t.Fatal(err)
	}
	gone, err := m.RemovePlace("base")
	if err != nil {
		t.Fatal(err)
	}
	if gone != 2 {
		t.Errorf("RemovePlace took %d places, want both that carried the id", gone)
	}
	left := m.Places()
	if len(left) != 1 || left[0].Name != "Three" {
		t.Errorf("what is left is %+v", left)
	}
	if gone, err := m.RemovePlace("base"); err != nil || gone != 0 {
		t.Errorf("removing an id that is not there: %d, %v", gone, err)
	}
}

// TestAddPlace: one place at a time, and an id already set is replaced
// rather than doubled.
func TestAddPlace(t *testing.T) {
	m := world(t, 80, 24)
	id, err := m.AddPlace(tuimaps.Place{ID: "home", Name: "Home", At: tuimaps.LonLat{Lon: 5, Lat: 5}})
	if err != nil || id != "home" {
		t.Fatalf("%q, %v", id, err)
	}
	if _, err := m.AddPlace(tuimaps.Place{ID: "home", Name: "Home again", At: tuimaps.LonLat{Lon: 6, Lat: 6}}); err != nil {
		t.Fatal(err)
	}
	places := m.Places()
	if len(places) != 1 || places[0].Name != "Home again" {
		t.Errorf("adding the same id twice gave %+v", places)
	}
}

// TestPlacesRefused: the mistakes a host can make with a place each give a
// kind from the closed list, and a message that says what to do.
func TestPlacesRefused(t *testing.T) {
	m := world(t, 80, 24)
	cases := []struct {
		name  string
		place tuimaps.Place
		kind  fault.Kind
	}{
		{"a latitude off the world", tuimaps.Place{Name: "x", At: tuimaps.LonLat{Lon: 0, Lat: 100}}, fault.InvalidCoordinates},
		{"a longitude off the world", tuimaps.Place{Name: "x", At: tuimaps.LonLat{Lon: 200, Lat: 0}}, fault.InvalidCoordinates},
		{"a latitude that is not a number", tuimaps.Place{Name: "x", At: tuimaps.LonLat{Lon: 0, Lat: notANumber()}}, fault.InvalidCoordinates},
		{"an id that would need cleaning", tuimaps.Place{ID: "a\x1bb", Name: "x", At: tuimaps.LonLat{}}, fault.InvalidID},
		{"a marker that is no marker", tuimaps.Place{Name: "x", Marker: 99, At: tuimaps.LonLat{}}, fault.InvalidID},
	}
	for _, c := range cases {
		if _, err := m.AddPlace(c.place); !isKind(err, c.kind) {
			t.Errorf("%s: %v", c.name, err)
		}
	}
	if len(m.Places()) != 0 {
		t.Errorf("a refused place was kept: %+v", m.Places())
	}
	// A name is cleaned rather than refused: it is text, not an identifier.
	id, err := m.AddPlace(tuimaps.Place{Name: "Bad\x1b[0mName", At: tuimaps.LonLat{}})
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Places()[0].Name; strings.Contains(got, "\x1b") {
		t.Errorf("the name kept an escape: %q (id %q)", got, id)
	}
}

// TestBlinkingPlaceFollowsTheClock: a place that blinks is drawn in one half
// of the blink and not in the other, on the clock the host passes to Render,
// and reduce-motion draws it steadily (NFR-21).
func TestBlinkingPlaceFollowsTheClock(t *testing.T) {
	const cols, rows = 149, 38
	m := world(t, cols, rows)
	if _, err := m.SetPlaces([]tuimaps.Place{{Name: "Beacon", At: tuimaps.LonLat{Lon: -80.19, Lat: 25.77}, Marker: tuimaps.MarkerDot, Blink: true}}); err != nil {
		t.Fatal(err)
	}
	on := frameAt(t, m, cols, rows, 0)
	off := frameAt(t, m, cols, rows, 500)
	if on == off {
		t.Error("a blinking place is drawn the same in both halves of its blink")
	}
	m.ReduceMotion(true)
	steady := frameAt(t, m, cols, rows, 0)
	again := frameAt(t, m, cols, rows, 500)
	if steady != again {
		t.Error("a place still blinks with reduce-motion on")
	}
}
