package overlay

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/project"
)

func square(west, south, size float64) []project.LonLat {
	return []project.LonLat{{Lon: west, Lat: south}, {Lon: west + size, Lat: south}, {Lon: west + size, Lat: south + size}, {Lon: west, Lat: south + size}, {Lon: west, Lat: south}}
}

func alert(id string, rings ...[]project.LonLat) Overlay {
	return Overlay{ID: id, Valid: noon, Keeps: time.Hour, Credit: "National Weather Service",
		Features: []Feature{{Kind: Polygon, Rings: rings, Role: colour.AlertSevereOutline, Label: "Tornado Warning"}}}
}

func store(t *testing.T) *Store {
	t.Helper()
	s, err := NewStore(Caps{})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// Plan task 10.1 (D-74, D-86).
func TestSetReportsCreatedOrReplaced(t *testing.T) {
	s := store(t)
	res, err := s.HandIn(alert("warnings", square(-95, 38, 1)))
	if err != nil || !res.Created || !res.Released {
		t.Errorf("a first set: %+v, %v; want created, and nothing old to hold", res, err)
	}
	res, err = s.HandIn(alert("warnings", square(-96, 38, 1)))
	if err != nil || res.Created || !res.Released {
		t.Errorf("the same id again: %+v, %v; want replaced, the old geometry released at once", res, err)
	}
	if got := s.IDs(); len(got) != 1 || got[0] != "warnings" {
		t.Errorf("ids %v", got)
	}
}

func TestRemoveReportsFound(t *testing.T) {
	s := store(t)
	s.HandIn(alert("warnings", square(-95, 38, 1)))
	if res, err := s.Drop("warnings"); err != nil || !res.Found || !res.Released {
		t.Errorf("%+v, %v", res, err)
	}
	if res, err := s.Drop("warnings"); err != nil || res.Found {
		t.Errorf("removing what is not there: %+v, %v; want not found, and no error", res, err)
	}
	if len(s.IDs()) != 0 || s.Reading("warnings") {
		t.Error("the overlay outlived its removal")
	}
}

// TestRemoveUnknownReportsNotFound, with the rest of 10.25 below.
func TestRemoveUnknownReportsNotFound(t *testing.T) {
	s := store(t)
	if res, err := s.Drop("never-set"); err != nil || res.Found {
		t.Errorf("%+v, %v", res, err)
	}
	if _, err := s.Drop("bad\x1bid"); !isKind(err, fault.InvalidID) {
		t.Errorf("an id that would need cleaning: %v", err)
	}
}

// TestDrainingReportedByWorkReturn is the heart of 10.5 (D-86): with a reader
// still inside the old geometry, a replace says "not yet released", InUse is
// true, and the reader's own return is what reports the release.
func TestDrainingReportedByWorkReturn(t *testing.T) {
	s := store(t)
	s.HandIn(alert("warnings", square(-95, 38, 1)))
	reader, ok := s.Read("warnings")
	if !ok {
		t.Fatal("nothing to read")
	}
	res, err := s.HandIn(alert("warnings", square(-96, 38, 1)))
	if err != nil || res.Created || res.Released {
		t.Fatalf("a replace under a reader: %+v, %v; want replaced and not yet released", res, err)
	}
	if !s.Reading("warnings") {
		t.Error("InUse is false while a call is still reading the old geometry")
	}
	if released := reader.Done(); len(released) != 1 || released[0] != "warnings" {
		t.Errorf("the reader's return reported %v; it was the last to read the old geometry", released)
	}
	if s.Reading("warnings") {
		t.Error("InUse is still true after the last reader left")
	}
	if again := reader.Done(); len(again) != 0 {
		t.Errorf("a reader reported its release twice: %v", again)
	}
	// A reader of the current geometry reports nothing: nothing was replaced.
	current, _ := s.Read("warnings")
	if released := current.Done(); len(released) != 0 {
		t.Errorf("%v released with nothing replaced", released)
	}
	// Removal under a reader is the same.
	held, _ := s.Read("warnings")
	if res, _ := s.Drop("warnings"); !res.Found || res.Released {
		t.Errorf("a remove under a reader: %+v", res)
	}
	if released := held.Done(); len(released) != 1 {
		t.Errorf("released %v", released)
	}
}

// TestBorrowNotCopied is plan task 10.4 (FR-11): the store reads the host's
// slices where they are.
func TestBorrowNotCopied(t *testing.T) {
	s := store(t)
	ring := square(-95, 38, 1)
	s.HandIn(alert("warnings", ring))
	reader, _ := s.Read("warnings")
	defer reader.Done()
	got := reader.Overlay().Features[0].Rings[0]
	if &got[0] != &ring[0] {
		t.Error("the ring the store reads is not the host's own memory")
	}
	if s.OwnedBytes() > 4096 {
		t.Errorf("the store owns %d bytes for one small overlay; borrowed geometry is not counted because it is not held", s.OwnedBytes())
	}
}

// TestHandInMistakes is plan task 10.2 (NFR-20): each likely mistake gives
// its own kind, and a message that says what to do.
func TestHandInMistakes(t *testing.T) {
	good := alert("warnings", square(-95, 38, 1))
	cases := []struct {
		name   string
		break_ func(o *Overlay)
		kind   fault.Kind
		says   string
	}{
		{"an empty id", func(o *Overlay) { o.ID = "" }, fault.InvalidID, "id"},
		{"an id that would need cleaning", func(o *Overlay) { o.ID = "warn\x1b[2Jings" }, fault.InvalidID, "id"},
		{"no valid time", func(o *Overlay) { o.Valid = time.Time{} }, fault.BadCurrency, "valid"},
		{"a currency of eight days", func(o *Overlay) { o.Keeps = 8 * 24 * time.Hour }, fault.BadCurrency, "seven days"},
		{"latitude and longitude the wrong way round", func(o *Overlay) {
			o.Features[0].Rings = [][]project.LonLat{{{Lon: 38, Lat: -95}, {Lon: 39, Lat: -95}, {Lon: 39, Lat: -94}, {Lon: 38, Lat: -95}}}
		}, fault.InvalidCoordinates, "longitude"},
		{"a coordinate that is no number", func(o *Overlay) { o.Features[0].Rings[0][1].Lon = math.NaN() }, fault.InvalidCoordinates, "number"},
		{"a ring of two positions", func(o *Overlay) { o.Features[0].Rings = [][]project.LonLat{{{Lon: 1, Lat: 1}, {Lon: 2, Lat: 2}}} }, fault.RingTooShort, "three"},
		{"a polygon with no ring", func(o *Overlay) { o.Features[0].Rings = nil }, fault.RingTooShort, "ring"},
		{"a circle with no radius", func(o *Overlay) {
			o.Features[0] = Feature{Kind: Circle, Centre: project.LonLat{Lon: -95, Lat: 38}, Role: colour.Marker}
		}, fault.InvalidCoordinates, "radius"},
		{"a circle that also has rings", func(o *Overlay) {
			o.Features[0] = Feature{Kind: Circle, Centre: project.LonLat{Lon: -95, Lat: 38}, RadiusKm: 10, Role: colour.Marker, Rings: [][]project.LonLat{{{Lon: math.NaN()}}}}
		}, fault.InvalidCoordinates, "no rings"},
		{"a feature of no kind", func(o *Overlay) { o.Features[0].Kind = 0 }, fault.InvalidCoordinates, "kind"},
		{"an overlay with nothing in it", func(o *Overlay) { o.Features = nil }, fault.SizeMismatch, "nothing"},
		{"a role that is no token", func(o *Overlay) { o.Features[0].Role = colour.Token(250) }, fault.UnknownPreset, "token"},
	}
	for _, c := range cases {
		s := store(t)
		o := good
		o.Features = []Feature{good.Features[0]}
		o.Features[0].Rings = [][]project.LonLat{square(-95, 38, 1)}
		c.break_(&o)
		_, err := s.HandIn(o)
		if !isKind(err, c.kind) {
			t.Errorf("%s: %v; want the %v kind", c.name, err, c.kind)
			continue
		}
		if !strings.Contains(err.Error(), c.says) {
			t.Errorf("%s: the message does not mention %q: %v", c.name, c.says, err)
		}
		// TestRefusedSetLeavesWarning (10.25): a discarded error still shows.
		w := s.TakeWarnings()
		if len(w) != 1 || w[0].Kind != fault.SetRefused {
			t.Errorf("%s: warnings %+v; a refused Set leaves one", c.name, w)
		}
		if len(s.IDs()) != 0 {
			t.Errorf("%s: a refused overlay was kept", c.name)
		}
	}
}

// TestVertexCaps is plan task 10.11.
func TestVertexCaps(t *testing.T) {
	s, err := NewStore(Caps{OverlayVertices: 1000, StoreVertices: 1500})
	if err != nil {
		t.Fatal(err)
	}
	big := circle(project.LonLat{Lon: -95, Lat: 38}, 2, 1200)
	if _, err := s.HandIn(alert("huge", big)); !isKind(err, fault.OverVertexCap) || !strings.Contains(err.Error(), "1,000") {
		t.Errorf("1,201 vertices against a cap of 1,000: %v", err)
	}
	ok := circle(project.LonLat{Lon: -95, Lat: 38}, 2, 899)
	if _, err := s.HandIn(alert("first", ok)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.HandIn(alert("second", ok)); !isKind(err, fault.OverVertexCap) {
		t.Errorf("900 and 900 against a store cap of 1,500: %v", err)
	}
	if _, err := s.HandIn(alert("first", ok)); err != nil {
		t.Errorf("replacing an overlay counts the new one, not both: %v", err)
	}
	defaults := store(t)
	if defaults.caps.OverlayVertices != 2_000_000 || defaults.caps.StoreVertices != 4_000_000 {
		t.Errorf("default caps %+v", defaults.caps)
	}
	if _, err := NewStore(Caps{OverlayVertices: 3_000_000}); err == nil {
		t.Error("a cap may be lowered, never raised")
	}
}

// TestWarningsCappedAndDeduplicated is plan task 10.3.
func TestWarningsCappedAndDeduplicated(t *testing.T) {
	s := store(t)
	for range 10 {
		s.HandIn(Overlay{ID: "same"})
	}
	w := s.TakeWarnings()
	if len(w) != 1 || w[0].Count != 10 {
		t.Errorf("ten refusals of one id: %+v; want one warning counting ten", w)
	}
	for i := range 200 {
		s.HandIn(Overlay{ID: "id-" + strings.Repeat("x", i%100) + string(rune('a'+i%26))})
	}
	if w := s.TakeWarnings(); len(w) > 64 {
		t.Errorf("%d warnings; at most 64 are kept", len(w))
	}
	if w := s.TakeWarnings(); len(w) != 0 {
		t.Errorf("warnings are taken once: %+v", w)
	}
}

// TestNearDuplicateIDWarned is part of 10.25: a mistyped id on a refresh
// makes a second overlay, and the result and a warning both show it.
func TestNearDuplicateIDWarned(t *testing.T) {
	s := store(t)
	s.HandIn(alert("nws-warnings", square(-95, 38, 1)))
	for _, typo := range []string{"nws-warning", "NWS-warnings", "nws_warnings", "nws-warnings "} {
		res, err := s.HandIn(alert(typo, square(-95, 38, 1)))
		if err != nil || !res.Created {
			t.Fatalf("%q: %+v, %v", typo, res, err)
		}
		w := s.TakeWarnings()
		if len(w) != 1 || w[0].Kind != fault.NearDuplicateID {
			t.Errorf("%q beside \"nws-warnings\": warnings %+v", typo, w)
		}
		s.Drop(typo)
	}
	s.HandIn(alert("radar", square(-95, 38, 1)))
	if w := s.TakeWarnings(); len(w) != 0 {
		t.Errorf("an id nothing like the other: %+v", w)
	}
}

func TestSetAndRemoveNeverBlock(t *testing.T) {
	s := store(t)
	s.HandIn(alert("warnings", square(-95, 38, 1)))
	reader, _ := s.Read("warnings")
	done := make(chan struct{})
	go func() {
		s.HandIn(alert("warnings", square(-96, 38, 1)))
		s.Drop("warnings")
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Set or Remove waited on a reader")
	}
	reader.Done()
}
