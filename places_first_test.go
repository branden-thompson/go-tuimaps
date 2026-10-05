package tuimaps_test

// places_first_test.go — v0.2.0 L3.13 (L-8.9, D-60, D-80): the host's named
// place keeps its marker and its name against alert labels and basemap
// names; the frame's furniture keeps its rows; a name that cannot fit falls
// back to a shorter form or the marker alone, and the frame says so.

import (
	"strings"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

func placeScene(t *testing.T, cols, rows int, place tuimaps.Place, withAlert bool) tuimaps.Frame {
	t.Helper()
	m := world(t, cols, rows)
	must(t, m.Recentre(tuimaps.LonLat{Lon: -92.3, Lat: 34.7}))
	must(t, m.Zoom(5))
	if withAlert {
		mustSet(t, m, tuimaps.Overlay{ID: "alerts", Valid: noon, Keeps: time.Hour, Features: []tuimaps.Feature{{Kind: tuimaps.Polygon, Role: tuimaps.AlertSevere,
			Label: "Severe Thunderstorm Warning", Rings: [][]tuimaps.LonLat{{{Lon: -95, Lat: 33}, {Lon: -89.6, Lat: 33}, {Lon: -89.6, Lat: 36.4}, {Lon: -95, Lat: 36.4}, {Lon: -95, Lat: 33}}}}}})
	}
	if _, err := m.AddPlace(place); err != nil {
		t.Fatal(err)
	}
	settle(t, m)
	f, err := m.Render(tuimaps.Size{Cols: cols, Rows: rows}, noon)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

// TestTheNamedPlaceOutranksAlertLabelsAndNames is L3.13: a place in the
// middle of an alert, where the alert's label wants the same cells, keeps its
// name at both sizes; at 149x38 the alert's label still finds room or is
// reported.
func TestTheNamedPlaceOutranksAlertLabelsAndNames(t *testing.T) {
	home := tuimaps.Place{ID: "home", Name: "Little Rock", At: tuimaps.LonLat{Lon: -92.3, Lat: 34.7}}
	for _, size := range [][2]int{{69, 12}, {149, 38}} {
		f := placeScene(t, size[0], size[1], home, true)
		if !strings.Contains(plainText(strings.Join(f.Lines, "\n")), "Little Rock") {
			t.Errorf("%dx%d: the place's name is lost under the alert", size[0], size[1])
		}
		for _, d := range f.Dropped {
			if d.Kind == tuimaps.DropPlaceName {
				t.Errorf("%dx%d: the place was reported dropped with room for it: %+v", size[0], size[1], d)
			}
		}
	}
}

// TestAPlaceNameNeverCoversFurniture is D-80: a place just above the bottom
// row, whose name would run into the scale and credit, keeps the furniture
// whole; its name falls back to a shorter form or the marker alone, and the
// frame reports it.
func TestAPlaceNameNeverCoversFurniture(t *testing.T) {
	far := tuimaps.Place{ID: "far", Name: "Pine Bluff Arsenal Area", At: tuimaps.LonLat{Lon: -93.9, Lat: 33.62}}
	base := placeScene(t, 69, 12, tuimaps.Place{ID: "far", At: far.At}, false)
	f := placeScene(t, 69, 12, far, false)
	last := len(f.Lines) - 1
	if plainText(f.Lines[last]) != plainText(base.Lines[last]) || plainText(f.Lines[0]) != plainText(base.Lines[0]) {
		t.Errorf("a place's name changed a furniture row:\n%q\n%q", plainText(f.Lines[0]), plainText(f.Lines[last]))
	}
	var drop *tuimaps.Drop
	for i := range f.Dropped {
		if f.Dropped[i].Kind == tuimaps.DropPlaceName {
			drop = &f.Dropped[i]
		}
	}
	if drop == nil {
		t.Fatalf("the name could not fit whole and nothing was reported: %+v\n%s", f.Dropped, plainText(strings.Join(f.Lines, "\n")))
	}
	if drop.Overlay != "far" || drop.Label != far.Name {
		t.Errorf("the report does not name the place: %+v", drop)
	}
	if drop.Shown != "" && !strings.Contains(plainText(strings.Join(f.Lines, "\n")), drop.Shown) {
		t.Errorf("the shorter form %q is not on the frame", drop.Shown)
	}
}

// TestAOneWordNameThatDoesNotFitIsReported: a name with no shorter form that
// cannot fit leaves the marker alone, and the frame still says so.
func TestAOneWordNameThatDoesNotFitIsReported(t *testing.T) {
	one := tuimaps.Place{ID: "far", Name: "Pinebluffarsenalwestgateyard", At: tuimaps.LonLat{Lon: -93.9, Lat: 33.62}}
	f := placeScene(t, 69, 12, one, false)
	for _, d := range f.Dropped {
		if d.Kind == tuimaps.DropPlaceName && d.Overlay == "far" && d.Shown == "" {
			return
		}
	}
	t.Errorf("a one-word name that cannot fit: %+v; want a DropPlaceName with nothing shown", f.Dropped)
}
