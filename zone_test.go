package tuimaps_test

// zone_test.go — v0.3.0 L-8 (watchpost D-195, D-196): a host's zone, its
// outline always drawn, its inside tinted where nothing is under it or the
// tint blends, stippled over a field it cannot tint, beneath any alert's
// tint, and answered by Report.

import (
	"fmt"
	"image/color"
	"regexp"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/colour"
)

// zoneCentre is a station's: the zone's middle in every test here.
var zoneCentre = tuimaps.LonLat{Lon: -117.23, Lat: 33.28}

// zoneOf is a host's zone: a circle of a radius in kilometres.
func zoneOf(id string, km float64) tuimaps.Overlay {
	return tuimaps.Overlay{ID: id, Valid: noon, Keeps: time.Hour, Features: []tuimaps.Feature{{Kind: tuimaps.Circle, Centre: zoneCentre, RadiusKm: km, Role: tuimaps.Zone}}}
}

// zoneFrame is a 60x20 truecolor frame over the zone's region with the
// overlays given, set in order; its raw text.
func zoneFrame(t *testing.T, overlays ...tuimaps.Overlay) string {
	t.Helper()
	return zoneFrameAt(t, 4, overlays...)
}

// zoneFrameAt is zoneFrame at a zoom: at 6 the zone fills the view.
func zoneFrameAt(t *testing.T, zoom float64, overlays ...tuimaps.Overlay) string {
	t.Helper()
	m := world(t, 60, 20)
	m.ColourDepth(tuimaps.Truecolor)
	must(t, m.Recentre(zoneCentre))
	must(t, m.Zoom(zoom))
	for _, o := range overlays {
		mustSet(t, m, o)
	}
	settle(t, m)
	f, err := m.Render(tuimaps.Size{Cols: 60, Rows: 20}, noon)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Join(f.Lines, "\n")
}

func fgCode(c colour.RGB) string { return fmt.Sprintf("38;2;%d;%d;%d", c.R, c.G, c.B) }

// sgr is one colour sequence.
var sgr = regexp.MustCompile("^\x1b\\[([0-9;]*)m")

// cellsWithBackground counts a frame's cells whose background is a colour: the emitter
// sends a colour only when it changes, so a run of cells is one sequence and
// counting sequences would count runs.
func cellsWithBackground(frame string, bg colour.RGB) int {
	want := fmt.Sprintf("%d;%d;%d", bg.R, bg.G, bg.B)
	n := 0
	for _, line := range strings.Split(frame, "\n") {
		current := ""
		for len(line) > 0 {
			if m := sgr.FindStringSubmatch(line); m != nil {
				params := strings.Split(m[1], ";")
				for i := 0; i < len(params); i++ {
					switch {
					case params[i] == "" || params[i] == "0" || params[i] == "49":
						current = ""
					case params[i] == "48" && i+4 < len(params) && params[i+1] == "2":
						current = strings.Join(params[i+2:i+5], ";")
						i += 4
					case (params[i] == "38" || params[i] == "48") && i+1 < len(params) && params[i+1] == "5":
						i += 2
					case params[i] == "38" && i+4 < len(params) && params[i+1] == "2":
						i += 4
					}
				}
				line = line[len(m[0]):]
				continue
			}
			_, size := utf8.DecodeRuneInString(line)
			if current == want {
				n++
			}
			line = line[size:]
		}
	}
	return n
}

// zoneColours are the zone's outline and tint on the dark ground.
func zoneColours() (line, tint colour.RGB) {
	var none colour.Palette
	line, _ = none.ResolveAt(colour.Zone, colour.Dark, colour.Truecolor)
	tint, _ = none.ResolveAt(colour.ZoneTint, colour.Dark, colour.Truecolor)
	return line, tint
}

// TestAZoneIsOutlinedAndFilledWithoutAlertWords is L-8.1: a zone is drawn
// in its outline's ink and its inside in its tint, with no severity word or
// digit and no entry in the legend.
func TestAZoneIsOutlinedAndFilledWithoutAlertWords(t *testing.T) {
	line, tint := zoneColours()
	plain := zoneFrame(t)
	frame := zoneFrame(t, zoneOf("zone", 160))
	if strings.Count(frame, fgCode(line)) == 0 {
		t.Error("the zone's outline is not drawn in its ink")
	}
	if strings.Count(frame, bgCode(tint)) == 0 {
		t.Error("the zone's inside is not its tint")
	}
	if strings.Count(plain, fgCode(line))+strings.Count(plain, bgCode(tint)) != 0 {
		t.Fatal("the plain map already has the zone's colours, so this proves nothing")
	}
	text := colours.ReplaceAllString(frame, "")
	for _, w := range []string{"Extreme", "Severe", "Moderate", "Minor", "Unknown"} {
		if strings.Contains(text, w) {
			t.Errorf("the zone is drawn with an alert's word %q", w)
		}
	}
	if regexp.MustCompile(`[1-5]`).MatchString(colours.ReplaceAllString(plain, "")) {
		return // the basemap writes a digit of its own here: the next check would prove nothing
	}
	if regexp.MustCompile(`[1-5]`).MatchString(text) {
		t.Error("the zone is drawn with a severity digit")
	}
	m := world(t, 60, 20)
	mustSet(t, m, zoneOf("zone", 160))
	settle(t, m)
	if n := len(m.Legend()); n != 0 {
		t.Errorf("the zone has %d legend entries; a host names its zone in its own words", n)
	}
}

// TestAnAlertsTintIsOverAZones is L-8.1: where an alert's area and a zone
// meet, the alert's tint is shown, whichever overlay is set first or sorts
// first.
func TestAnAlertsTintIsOverAZones(t *testing.T) {
	_, severe := radarColours(colour.Dark)
	alert := warningAt("m-alert", -117.6, 33.0, -116.9, 33.6)
	want := cellsWithBackground(zoneFrameAt(t, 6, alert), severe)
	if want < 20 {
		t.Fatalf("the alert's tint covers %d cells alone, so this proves little", want)
	}
	for _, order := range [][]tuimaps.Overlay{{zoneOf("a-zone", 160), alert}, {alert, zoneOf("z-zone", 160)}} {
		if got := cellsWithBackground(zoneFrameAt(t, 6, order...), severe); got != want {
			t.Errorf("with a zone (%s) the alert's tint covers %d cells; alone %d", order[0].ID, got, want)
		}
	}
}

// TestAZoneOverAFieldItCannotTintIsStippled is L-8.2 (watchpost D-196): over
// a MUF field, which no tint can blend with, every cell keeps the field's own
// colour and the zone's inside carries dots in the zone's ink.
func TestAZoneOverAFieldItCannotTintIsStippled(t *testing.T) {
	line, tint := zoneColours()
	g := tuimaps.Grid{West: -130, South: 20, East: -100, North: 45, Cols: 30, Rows: 25, Values: make([]float64, 30*25), OverWater: true}
	for i := range g.Values {
		g.Values[i] = 8 + float64(i%30)/2
	}
	field := tuimaps.MUFGrid("muf", g, noon)
	alone := zoneFrame(t, field)
	frame := zoneFrame(t, field, zoneOf("zone", 160))
	backgrounds := func(s string) map[string]bool {
		out := map[string]bool{}
		for _, m := range regexp.MustCompile(`48;2;\d+;\d+;\d+`).FindAllString(s, -1) {
			out[m] = true
		}
		return out
	}
	with, without := backgrounds(frame), backgrounds(alone)
	for bg := range with {
		if !without[bg] {
			t.Errorf("a cell's background is %s inside the zone, a colour the field alone never has: the field is not exact", bg)
		}
	}
	if with[bgCode(tint)] {
		t.Error("the zone's tint is drawn over the field")
	}
	// The dots are counted, not their ink: over a field's middle classes no
	// red reaches 3:1, and the foreground rule draws the dot in black or
	// white there (FR-16), as it does an alert's outline.
	dots := func(s string) int { return strings.Count(colours.ReplaceAllString(s, ""), "⠂") }
	if n, before := dots(frame), dots(alone); n < before+10 {
		t.Errorf("%d single dots with the zone, %d without; the zone over the field is not stippled", n, before)
	}
	if strings.Count(zoneFrame(t, zoneOf("zone", 160)), fgCode(line)) == 0 {
		t.Error("the zone's outline is not in its ink on the bare ground, so its ink is not the zone's")
	}
}

// TestReportAnswersAKilometreCircle is L-8.3: Report answers whether a place
// is inside a circle given by its radius in kilometres, as for a polygon.
func TestReportAnswersAKilometreCircle(t *testing.T) {
	m := world(t, 60, 20)
	mustSet(t, m, zoneOf("zone", 160))
	settle(t, m)
	r, err := m.Report([]tuimaps.Place{{Name: "tower", At: zoneCentre}, {Name: "far", At: tuimaps.LonLat{Lon: -100, Lat: 40}}})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]tuimaps.Where{"tower": tuimaps.Inside, "far": tuimaps.Outside}
	for _, p := range r.Places {
		found := false
		for _, a := range p.Answers {
			if a.Overlay == "zone" {
				found = true
				if a.Relation != want[p.Place] {
					t.Errorf("%s is %v the zone; want %v", p.Place, a.Relation, want[p.Place])
				}
			}
		}
		if !found {
			t.Errorf("%s has no answer for the zone", p.Place)
		}
	}
}

// TestAZoneBlendsWhereItCanAndIsNotStippledThere is L-8.2: over radar on the
// dark ground, where the search finds a strength for the zone's tint, the
// rain is blended toward the tint, and no cell is stippled.
func TestAZoneBlendsWhereItCanAndIsNotStippledThere(t *testing.T) {
	_, tint := zoneColours()
	ramp, _ := radarColours(colour.Dark)
	var none colour.Palette
	groundColour, _ := none.ResolveAt(colour.Ground, colour.Dark, colour.Truecolor)
	s, ok := colour.SearchBlends(none, colour.Dark, groundColour, colour.Truecolor).ZoneStrength(colour.Radar)
	if !ok {
		t.Fatal("the zone's tint does not blend over radar on the dark ground, so this proves nothing")
	}
	rain := rainOver(t, -130, 20, -100, 45, color.NRGBA{R: 200, A: 255})
	alone := zoneFrame(t, rain)
	frame := zoneFrame(t, rain, zoneOf("zone", 160))
	blends := 0
	for _, c := range ramp {
		blends += strings.Count(frame, bgCode(colour.Blend(c, tint, s)))
	}
	if blends == 0 {
		t.Errorf("no cell is radar blended toward the zone's tint at %.2f", s)
	}
	dots := func(s string) int { return strings.Count(colours.ReplaceAllString(s, ""), "⠂") }
	if n, before := dots(frame), dots(alone); n > before+3 {
		t.Errorf("%d single dots with the zone, %d without: the zone is stippled where its tint blends", n, before)
	}
}

// TestAZoneAtSixteenColoursIsItsOutlineAlone is L-8.1: at sixteen colours
// the zone's inside is not painted - no background the map did not already
// have - and its outline is still drawn.
func TestAZoneAtSixteenColoursIsItsOutlineAlone(t *testing.T) {
	frame := func(overlays ...tuimaps.Overlay) string {
		m := world(t, 60, 20)
		m.ColourDepth(tuimaps.Colours16)
		must(t, m.Recentre(zoneCentre))
		must(t, m.Zoom(4))
		for _, o := range overlays {
			mustSet(t, m, o)
		}
		settle(t, m)
		f, err := m.Render(tuimaps.Size{Cols: 60, Rows: 20}, noon)
		if err != nil {
			t.Fatal(err)
		}
		return strings.Join(f.Lines, "\n")
	}
	backgrounds := func(s string) map[string]bool {
		out := map[string]bool{}
		for _, seq := range regexp.MustCompile(`\x1b\[([0-9;]*)m`).FindAllStringSubmatch(s, -1) {
			for _, p := range strings.Split(seq[1], ";") {
				if (len(p) == 2 && p[0] == '4') || (len(p) == 3 && strings.HasPrefix(p, "10")) {
					out[p] = true
				}
			}
		}
		return out
	}
	plain, zoned := frame(), frame(zoneOf("zone", 160))
	for bg := range backgrounds(zoned) {
		if !backgrounds(plain)[bg] {
			t.Errorf("at sixteen colours the zone paints background %s", bg)
		}
	}
	if colours.ReplaceAllString(zoned, "") == colours.ReplaceAllString(plain, "") {
		t.Error("at sixteen colours the zone's outline is not drawn")
	}
}

// TestAZoneSetLaterIsDrawn: a zone set on a map that has drawn a frame is
// tinted in the next, every row inside it - none kept from the frame before.
func TestAZoneSetLaterIsDrawn(t *testing.T) {
	_, tint := zoneColours()
	m := world(t, 60, 20)
	m.ColourDepth(tuimaps.Truecolor)
	must(t, m.Recentre(zoneCentre))
	must(t, m.Zoom(6)) // the zone fills the view: rows with no outline in them to be kept
	draw := func() string {
		f, err := m.Render(tuimaps.Size{Cols: 60, Rows: 20}, noon)
		if err != nil {
			t.Fatal(err)
		}
		return strings.Join(f.Lines, "\n")
	}
	before := draw()
	mustSet(t, m, zoneOf("zone", 160))
	settle(t, m)
	after := draw()
	fresh := zoneFrameAt(t, 6, zoneOf("zone", 160))
	if got, want := cellsWithBackground(after, tint), cellsWithBackground(fresh, tint); got != want || want == 0 || cellsWithBackground(before, tint) != 0 {
		t.Errorf("a zone set after a frame is tinted in %d cells; on a fresh map %d", got, want)
	}
}
