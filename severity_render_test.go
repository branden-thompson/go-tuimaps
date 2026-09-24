package tuimaps_test

// severity_render_test.go — v0.2.0 L3.9 and L3.10 (L-8.1, D-65): an alert
// carries its severity as a word in its label, and as a digit repeated along
// its outline, at every depth.

import (
	"strings"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// fiveAlerts is specimen 33's arrangement: one area a severity, side by side
// and each a step further north.
func fiveAlerts() tuimaps.Overlay {
	roles := []tuimaps.Token{tuimaps.AlertExtreme, tuimaps.AlertSevere, tuimaps.AlertModerate, tuimaps.AlertMinor, tuimaps.AlertUnknown}
	names := []string{"Tornado Warning", "Thunderstorm Warning", "Flood Watch", "Flood Advisory", "Outlook"}
	o := tuimaps.Overlay{ID: "alerts", Valid: noon, Keeps: time.Hour, Credit: "National Weather Service"}
	for i, role := range roles {
		west, south := -100+4*float64(i), 26+3*float64(i) // a step down the map each, so no two labels share a row
		o.Features = append(o.Features, tuimaps.Feature{Kind: tuimaps.Polygon, Role: role, Label: names[i],
			Rings: [][]tuimaps.LonLat{{{Lon: west, Lat: south}, {Lon: west + 3, Lat: south}, {Lon: west + 3, Lat: south + 6}, {Lon: west, Lat: south + 6}, {Lon: west, Lat: south}}}})
	}
	return o
}

// alertFrame draws alerts on a map with no tiles, so that no basemap name
// can put a digit on it; its rows as plain text.
func alertFrame(t *testing.T, cols, rows int, depth tuimaps.Depth, alerts *tuimaps.Overlay) []string {
	t.Helper()
	m, err := tuimaps.New(tuimaps.WithSize(cols, rows))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	m.ColourDepth(depth)
	must(t, m.Recentre(tuimaps.LonLat{Lon: -91, Lat: 35}))
	must(t, m.Zoom(3.3-float64(149-cols)/80)) // the small map further out, so all five areas are in view
	if alerts != nil {
		mustSet(t, m, *alerts)
	}
	settle(t, m)
	f, err := m.Render(tuimaps.Size{Cols: cols, Rows: rows}, noon)
	if err != nil {
		t.Fatal(err)
	}
	out := make([]string, len(f.Lines))
	for i, l := range f.Lines {
		out[i] = strings.ReplaceAll(plainText(l), "⠀", " ")
	}
	return out
}

// TestTheLabelCarriesTheSeverityWord is L3.9 (L-8.1): every label that fits
// reads "<label> · <SEVERITY>", at both depths.
func TestTheLabelCarriesTheSeverityWord(t *testing.T) {
	alerts := fiveAlerts()
	for _, depth := range []tuimaps.Depth{tuimaps.Truecolor, tuimaps.NoColour} {
		frame := strings.Join(alertFrame(t, 149, 38, depth, &alerts), "\n")
		for _, want := range []string{"Tornado Warning · EXTREME", "Thunderstorm Warning · SEVERE", "Flood Watch · MODERATE", "Flood Advisory · MINOR", "Outlook · UNKNOWN"} {
			if !strings.Contains(frame, want) {
				t.Errorf("depth %v: no label reads %q", depth, want)
			}
		}
	}
}

// TestTheOutlineCarriesTheSeverityDigit is L3.10 (L-8.1, D-65): every area
// carries its digit - extreme 4, severe 3, moderate 2, minor 1, unknown ? -
// at 69x12 and 149x38 and at every depth; none is on a furniture row; and an
// area of a few cells still gets one.
func TestTheOutlineCarriesTheSeverityDigit(t *testing.T) {
	alerts := fiveAlerts()
	for _, size := range [][2]int{{69, 12}, {149, 38}} {
		for _, depth := range []tuimaps.Depth{tuimaps.Truecolor, tuimaps.Colours16, tuimaps.NoColour} {
			with := alertFrame(t, size[0], size[1], depth, &alerts)
			without := alertFrame(t, size[0], size[1], depth, nil)
			middle := strings.Join(with[1:len(with)-1], "\n")
			for _, digit := range []string{"4", "3", "2", "1", "?"} {
				if !strings.Contains(middle, digit) {
					t.Errorf("%dx%d depth %v: no area carries %q", size[0], size[1], depth, digit)
				}
			}
			for _, row := range []int{0, len(with) - 1} {
				a, b := []rune(with[row]), []rune(without[row])
				for col := range a {
					if col < len(b) && a[col] != b[col] && strings.ContainsRune("4321?", a[col]) {
						t.Errorf("%dx%d depth %v: a digit on furniture row %d, col %d: %q", size[0], size[1], depth, row, col, with[row])
						break
					}
				}
			}
		}
	}
	// Each severity alone carries its own digit and no other's.
	for i, digit := range []string{"4", "3", "2", "1", "?"} {
		one := fiveAlerts()
		one.Features = one.Features[i : i+1]
		rows := alertFrame(t, 149, 38, tuimaps.NoColour, &one)
		middle := strings.Join(rows[1:len(rows)-1], "\n")
		for _, other := range []string{"4", "3", "2", "1", "?"} {
			if has := strings.Contains(middle, other); has != (other == digit) {
				t.Errorf("the %s area alone: carries %q is %v", digit, other, has)
			}
		}
	}
	tiny := tuimaps.Overlay{ID: "tiny", Valid: noon, Keeps: time.Hour, Features: []tuimaps.Feature{{Kind: tuimaps.Polygon, Role: tuimaps.AlertMinor,
		Rings: [][]tuimaps.LonLat{{{Lon: -93, Lat: 39}, {Lon: -91.8, Lat: 39}, {Lon: -91.8, Lat: 40.5}, {Lon: -93, Lat: 40.5}, {Lon: -93, Lat: 39}}}}}} // off the notice's row
	rows := alertFrame(t, 69, 12, tuimaps.NoColour, &tiny)
	cells := 0
	for _, r := range rows[1 : len(rows)-1] {
		cells += len([]rune(strings.TrimSpace(strings.Split(r, "no map tiles")[0]))) - strings.Count(strings.TrimSpace(r), " ")
	}
	if cells == 0 || cells > 8 {
		t.Fatalf("the small area draws %d cells, not a few:\n%s", cells, strings.Join(rows, "\n"))
	}
	if !strings.Contains(strings.Join(rows[1:len(rows)-1], "\n"), "1") {
		t.Errorf("an area a few cells across carries no digit:\n%s", strings.Join(rows, "\n"))
	}
}
