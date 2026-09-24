package tuimaps_test

// severity_render_test.go — v0.2.0 L3.9 and L3.10 (L-8.1, D-65): an alert
// carries its severity as a word in its label, and as a digit repeated along
// its outline, at every depth.

import (
	"image/color"
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

// alertMap is the five alerts on a map with no tiles, drawn once.
func alertMap(t *testing.T, cols, rows int, depth tuimaps.Depth) (*tuimaps.Map, tuimaps.Frame) {
	t.Helper()
	m, err := tuimaps.New(tuimaps.WithSize(cols, rows))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { m.Close() })
	m.ColourDepth(depth)
	must(t, m.Recentre(tuimaps.LonLat{Lon: -91, Lat: 35}))
	must(t, m.Zoom(3.3-float64(149-cols)/80))
	mustSet(t, m, fiveAlerts())
	settle(t, m)
	f, err := m.Render(tuimaps.Size{Cols: cols, Rows: rows}, noon)
	if err != nil {
		t.Fatal(err)
	}
	return m, f
}

// TestALabelThatDoesNotFitFallsBackAndIsReported is L3.11 (L-8.5, L-8.7):
// the hatch draws at sixteen colours as at none; at 69x12 a label that does
// not fit falls back to its severity word, and Frame.Dropped names each label
// the frame shortened or left out, with what stood in for it. At 149x38,
// where every label fits, nothing is dropped.
func TestALabelThatDoesNotFitFallsBackAndIsReported(t *testing.T) {
	_, wide := alertMap(t, 149, 38, tuimaps.Colours16)
	if !strings.ContainsAny(plainText(strings.Join(wide.Lines, "\n")), "╱╲╳") {
		t.Error("no hatch at sixteen colours")
	}
	if wide.Dropped != nil {
		t.Errorf("every label fits at 149x38, and the frame says %v were dropped", wide.Dropped)
	}
	_, small := alertMap(t, 69, 12, tuimaps.NoColour)
	text := strings.Join(small.Lines, "\n")
	if len(small.Dropped) == 0 {
		t.Fatal("at 69x12 nothing was dropped, so this proves nothing")
	}
	fellBack := 0 // the five-alert scene at 69x12 drops only areas out of view
	for _, d := range small.Dropped {
		if d.Kind != tuimaps.DropAlertLabel || d.Overlay != "alerts" || !strings.Contains(d.Label, " · ") {
			t.Errorf("a drop that does not name its alert: %+v", d)
		}
		if d.Shown != "" {
			fellBack++
			if !strings.HasSuffix(d.Label, d.Shown) || !strings.Contains(plainText(text), d.Shown) {
				t.Errorf("%+v: the word shown is not the label's, or is not on the frame", d)
			}
		}
	}
	if fellBack != 0 {
		t.Errorf("the areas dropped here lie outside the view, and a word was shown for one: %+v", small.Dropped)
	}
	// A label longer than the map is wide, on an area in the middle of it:
	// its word stands in.
	long := fiveAlerts()
	long.Features = long.Features[1:2]
	long.Features[0].Label = "Severe Thunderstorm Warning including the city of Little Rock and its suburbs until 9 PM"
	long.Features[0].Rings = [][]tuimaps.LonLat{{{Lon: -95, Lat: 33}, {Lon: -87, Lat: 33}, {Lon: -87, Lat: 38}, {Lon: -95, Lat: 38}, {Lon: -95, Lat: 33}}}
	m, err := tuimaps.New(tuimaps.WithSize(69, 12))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	must(t, m.Recentre(tuimaps.LonLat{Lon: -91, Lat: 35.5}))
	must(t, m.Zoom(3.3))
	mustSet(t, m, long)
	settle(t, m)
	f, err := m.Render(tuimaps.Size{Cols: 69, Rows: 12}, noon)
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Dropped) != 1 || f.Dropped[0].Shown != "SEVERE" || !strings.HasPrefix(f.Dropped[0].Label, "Severe Thunderstorm Warning") {
		t.Fatalf("a label wider than the map: %+v; want it dropped with SEVERE shown", f.Dropped)
	}
	if !strings.Contains(plainText(strings.Join(f.Lines, "\n")), "SEVERE") {
		t.Error("the word that stood in is not on the frame")
	}
}

// TestTheLegendCarriesTheDigitKeyAndTheBlend is L3.11a (D-65) and L3.8's
// legend report: an alert overlay's legend lists the five severities, each
// with its digit and word, extreme first; an image's entry says which
// severities' tints blend over it - all five for radar on the dark ground,
// none on the light.
func TestTheLegendCarriesTheDigitKeyAndTheBlend(t *testing.T) {
	m, _ := alertMap(t, 80, 24, tuimaps.Truecolor)
	mustSet(t, m, rainOver(t, -110, 20, -70, 50, color.NRGBA{R: 200, A: 255}))
	var alert, radar *tuimaps.LegendEntry
	for _, e := range m.Legend() {
		switch e.Preset {
		case "alert":
			alert = &e
		case "radar":
			radar = &e
		}
	}
	if alert == nil || radar == nil {
		t.Fatalf("the legend has no alert entry or no radar entry: %+v", m.Legend())
	}
	want := [][2]string{{"4", "extreme"}, {"3", "severe"}, {"2", "moderate"}, {"1", "minor"}, {"?", "unknown"}}
	if len(alert.Classes) != 5 {
		t.Fatalf("the alert legend has %d classes: %+v", len(alert.Classes), alert.Classes)
	}
	for i, c := range alert.Classes {
		if c.Mark != want[i][0] || c.Label != want[i][1] || !c.Drawn {
			t.Errorf("class %d: %+v; want %s %s", i, c, want[i][0], want[i][1])
		}
	}
	if len(radar.Blended) != 5 {
		t.Errorf("radar on the dark ground: blended under %v; want all five", radar.Blended)
	}
	must(t, m.Ground(tuimaps.RGB{R: 250, G: 250, B: 245}))
	for _, e := range m.Legend() {
		if e.Preset == "radar" && len(e.Blended) != 0 {
			t.Errorf("radar on the light ground: blended under %v; want none", e.Blended)
		}
	}
}
