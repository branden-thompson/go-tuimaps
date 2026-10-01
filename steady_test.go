package tuimaps_test

// steady_test.go — v0.2.0 L11.33 (L-28, watchpost D-200, UAT-2 U2-46): as a
// loop played, an alert drawn by its time (L-15.1) took its word's room on
// the frames it met and gave it back on the others, and the place names
// around it came and went with it. The names are placed as if every alert
// in view were present.

import (
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// placeWords are a frame's words of four letters or more: but the alert's own
// and the furniture's capitals, the place names.
var placeWords = regexp.MustCompile(`[A-Za-z]{4,}`)

func wordsOn(t *testing.T, m *tuimaps.Map) []string {
	t.Helper()
	f, err := m.Render(tuimaps.Size{Cols: 120, Rows: 40}, noon)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, w := range placeWords.FindAllString(colours.ReplaceAllString(strings.Join(f.Lines, "\n"), ""), -1) {
		if w != "Tornado" && w != "Warning" && w != strings.ToUpper(w) && !slices.Contains(out, w) {
			out = append(out, w)
		}
	}
	slices.Sort(out)
	return out
}

// TestNamesHoldStillAsALoopPlays is L-28: a warning issued five minutes ago
// over the plains, a loop of ten minutes ago, five and now. The frame of ten
// minutes ago does not draw the warning, and places the same names as the
// frames that do.
func TestNamesHoldStillAsALoopPlays(t *testing.T) {
	m := world(t, 120, 40)
	m.ColourDepth(tuimaps.Truecolor)
	must(t, m.Recentre(tuimaps.LonLat{Lon: -95, Lat: 38}))
	must(t, m.Zoom(3.5))
	bare := wordsOn(t, m)
	must(t, m.SetPlayback(tuimaps.PlaybackOn))
	mustSet(t, m, loopAt(t, "radar", []int{-10, -5, 0}, nil, nil))
	mustSet(t, m, tuimaps.Overlay{ID: "alert", Valid: noon, Keeps: 72 * time.Hour, During: tuimaps.Span{From: at(-5)},
		Features: []tuimaps.Feature{boxAlert("warn", tuimaps.AlertSevere, "Tornado Warning", -99, 36, -91, 40)}})
	settle(t, m)
	must(t, m.Reset())
	now := wordsOn(t, m)
	f, err := m.Render(tuimaps.Size{Cols: 120, Rows: 40}, noon)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(colours.ReplaceAllString(strings.Join(f.Lines, "\n"), ""), "Tornado Warning") {
		t.Fatal("the frame of now does not draw the warning's word")
	}
	if slices.Equal(bare, now) {
		t.Fatalf("the warning's word displaces no name here (%v): the test proves nothing", bare)
	}
	must(t, m.Step(-2))
	if got := m.Loop().At; !got.Equal(at(-10)) {
		t.Fatalf("the moment is %v; want the frame of ten minutes ago", got)
	}
	if before := wordsOn(t, m); !slices.Equal(before, now) {
		t.Errorf("the frame before the warning places %v; the frame of now %v - the names move as the loop plays", before, now)
	}
}

// TestAFieldOutsideTheMomentHoldsTheNames is L-28 for a day's field: on a day
// it does not cover, the names are those under it, not a bare map's more.
func TestAFieldOutsideTheMomentHoldsTheNames(t *testing.T) {
	m := world(t, 120, 40)
	must(t, m.Recentre(tuimaps.LonLat{Lon: -95, Lat: 38}))
	must(t, m.Zoom(3.5))
	bare := wordsOn(t, m)
	day := func(k int) tuimaps.Span {
		return tuimaps.Span{From: noon.Add(time.Duration(24*k) * time.Hour), Until: noon.Add(time.Duration(24*k+23) * time.Hour)}
	}
	g := tuimaps.TemperatureGrid("day1", warmField(), tuimaps.Celsius, noon)
	g.During = day(1)
	mustSet(t, m, g)
	settle(t, m)
	must(t, m.ShowMoment(day(1).From, day(1).Until))
	under := wordsOn(t, m)
	if slices.Equal(bare, under) {
		t.Fatalf("the field changes no name here (%v): the test proves nothing", bare)
	}
	must(t, m.ShowMoment(day(2).From, day(2).Until))
	if got := wordsOn(t, m); !slices.Equal(got, under) {
		t.Errorf("the day without the field places %v; the day with it %v", got, under)
	}
}
