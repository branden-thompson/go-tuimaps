package tuimaps_test

// moment_test.go — v0.2.0 L11.10 and L11.11 (watchpost D-94 to D-98): an
// overlay drawn only while the map's moment meets its span, the moment the
// host says while no loop is held, and a field that shares the map with an
// image drawn as its lines.

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// spanAlert is a warning west of the loops' box, drawn during a span.
func spanAlert(during tuimaps.Span) tuimaps.Overlay {
	return tuimaps.Overlay{ID: "alert", Valid: noon, Keeps: 72 * time.Hour, During: during,
		Features: []tuimaps.Feature{boxAlert("warn", tuimaps.AlertSevere, "Severe Thunderstorm Warning", -92, 35, -91, 36)}}
}

func momentFrame(t *testing.T, m *tuimaps.Map, now time.Time) string {
	t.Helper()
	f, err := m.Render(tuimaps.Size{Cols: 80, Rows: 24}, now)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Join(f.Lines, "\n")
}

// TestAnOverlayIsDrawnOnlyWhileTheLoopsFrameMeetsIt is L-15.1: an alert that
// began five minutes ago is not drawn on the loop's frame of ten minutes ago,
// and is on the frame of now - a radar loop never shows a warning before it
// was issued (watchpost D-98).
func TestAnOverlayIsDrawnOnlyWhileTheLoopsFrameMeetsIt(t *testing.T) {
	m := reportMap(t)
	must(t, m.SetPlayback(tuimaps.PlaybackOn))
	mustSet(t, m, loopAt(t, "radar", []int{-10, -5, 0}, nil, nil))
	settle(t, m)
	must(t, m.Reset())
	bareNow := momentFrame(t, m, noon)
	must(t, m.Step(-2))
	bareBefore := momentFrame(t, m, noon)

	mustSet(t, m, spanAlert(tuimaps.Span{From: at(-5)}))
	settle(t, m)
	if got := m.Loop().At; !got.Equal(at(-10)) {
		t.Fatalf("the moment is %v; want the frame of ten minutes ago still held", got)
	}
	if momentFrame(t, m, noon) != bareBefore {
		t.Error("the alert was drawn on a frame from before it began")
	}
	must(t, m.Reset())
	if momentFrame(t, m, noon) == bareNow {
		t.Error("the alert was not drawn on the frame of now, which its span meets")
	}
}

// TestTheHostSaysTheMomentWhileNoLoopIsHeld is L-15.2: with no loop the
// moment is the frame's clock, or the span the host says - a forecast day,
// met by an alert that overlaps it anywhere; zero is the clock again.
func TestTheHostSaysTheMomentWhileNoLoopIsHeld(t *testing.T) {
	m := reportMap(t)
	bare := momentFrame(t, m, noon)
	tomorrow := tuimaps.Span{From: noon.Add(24 * time.Hour), Until: noon.Add(48 * time.Hour)}
	mustSet(t, m, spanAlert(tomorrow))
	settle(t, m)
	if momentFrame(t, m, noon) != bare {
		t.Error("an alert for tomorrow was drawn on today's clock")
	}
	for _, c := range []struct {
		name     string
		from, to time.Time
		drawn    bool
	}{
		{"inside it", noon.Add(30 * time.Hour), noon.Add(30 * time.Hour), true},
		{"a day that overlaps its end", noon.Add(47 * time.Hour), noon.Add(60 * time.Hour), true},
		{"a day after it", noon.Add(50 * time.Hour), noon.Add(60 * time.Hour), false},
		{"the clock again", time.Time{}, time.Time{}, false},
	} {
		before := m.Changed()
		must(t, m.ShowMoment(c.from, c.to))
		if m.Changed() == before {
			t.Errorf("%s: Changed did not move; the host's own moment is an input, and a host redraws on it", c.name)
		}
		if got := momentFrame(t, m, noon) != bare; got != c.drawn {
			t.Errorf("%s: drawn %v; want %v", c.name, got, c.drawn)
		}
	}
	// A loop held takes the moment back: the host's waits.
	must(t, m.ShowMoment(noon.Add(30*time.Hour), noon.Add(30*time.Hour)))
	must(t, m.SetPlayback(tuimaps.PlaybackOn))
	mustSet(t, m, loopAt(t, "radar", []int{-5, 0}, nil, nil))
	settle(t, m)
	withLoop := momentFrame(t, m, noon)
	if _, err := m.Remove("alert"); err != nil {
		t.Fatal(err)
	}
	if momentFrame(t, m, noon) != withLoop {
		t.Error("with a loop held the host's moment was used: the alert for tomorrow was drawn on the loop's frame of now")
	}
}

// TestTheClockCrossingASpanRedraws is L-15.1 with no loop and no host moment:
// the frame's clock is the moment, so a later frame meets a span an earlier
// one did not - and is drawn anew, never the earlier frame reused.
func TestTheClockCrossingASpanRedraws(t *testing.T) {
	m := reportMap(t)
	bare := momentFrame(t, m, noon)
	mustSet(t, m, spanAlert(tuimaps.Span{From: noon.Add(time.Hour)}))
	settle(t, m)
	if momentFrame(t, m, noon) != bare {
		t.Fatal("an alert beginning in an hour was drawn now")
	}
	if momentFrame(t, m, noon.Add(2*time.Hour)) == bare {
		t.Error("two hours on the clock meets the alert's span, and the frame did not draw it")
	}
}

// TestAStepSwapsOneDaysFieldForAnother is L-15.2 as a forecast steps: two
// days' fields, each drawn during its own day, and the host's moment moved
// from one to the other - as many fields drawn either way, so only the moment
// tells the frames apart, and the second is never the first reused.
func TestAStepSwapsOneDaysFieldForAnother(t *testing.T) {
	m := reportMap(t)
	m.ColourDepth(tuimaps.Truecolor)
	cold, warm := warmField(), warmField()
	for i := range cold.Values {
		cold.Values[i] -= 20
	}
	day := func(k int) tuimaps.Span {
		return tuimaps.Span{From: noon.Add(time.Duration(24*k) * time.Hour), Until: noon.Add(time.Duration(24*k+23) * time.Hour)}
	}
	a, b := tuimaps.TemperatureGrid("day1", cold, tuimaps.Celsius, noon), tuimaps.TemperatureGrid("day2", warm, tuimaps.Celsius, noon)
	a.During, b.During = day(1), day(2)
	mustSet(t, m, a)
	mustSet(t, m, b)
	settle(t, m)
	must(t, m.ShowMoment(day(1).From, day(1).Until))
	first := momentFrame(t, m, noon)
	must(t, m.ShowMoment(day(2).From, day(2).Until))
	if momentFrame(t, m, noon) == first {
		t.Error("the second day drew the first day's frame: the swap was not seen")
	}
}

// TestASpanOrMomentBackwardsIsRefused is L-15.1 and L-15.2's refusals.
func TestASpanOrMomentBackwardsIsRefused(t *testing.T) {
	m := reportMap(t)
	if _, err := m.Set(spanAlert(tuimaps.Span{From: noon, Until: noon.Add(-time.Minute)})); !isKind(err, fault.BadCurrency) {
		t.Errorf("a span ending before it begins: %v; want it refused", err)
	}
	if err := m.ShowMoment(noon, noon.Add(-time.Minute)); !isKind(err, fault.BadCurrency) {
		t.Errorf("a moment ending before it begins: %v; want it refused", err)
	}
}

// aValue is a contour's value written on the map.
var aValue = regexp.MustCompile(`-?\d+`)

// warmField is a temperature field over the lower 48, warming eastward, so
// its bands run down the frame.
func warmField() tuimaps.Grid {
	const cols, rows = 24, 24
	g := tuimaps.Grid{West: -125, South: 24, East: -66, North: 50, Cols: cols, Rows: rows}
	for range rows {
		for c := range cols {
			g.Values = append(g.Values, -20.0+float64(c)*2.5)
		}
	}
	return g
}

// bandGrounds counts the frame's cells painted in one of the field's own band
// colours, as backgrounds.
func bandGrounds(m *tuimaps.Map, raw string) int {
	n := 0
	for _, e := range m.Legend() {
		if e.Preset != "temperature" {
			continue
		}
		for _, c := range e.Classes {
			n += strings.Count(raw, "48;2;"+strconv.Itoa(int(c.Colour.R))+";"+strconv.Itoa(int(c.Colour.G))+";"+strconv.Itoa(int(c.Colour.B))+"m")
		}
	}
	return n
}

// TestAFieldSharingTheMapWithAnImageIsItsLines is L-15.3 (watchpost D-95):
// alone, a field fills its bands and draws no line; with an image on the map
// it draws its labelled contours, and its bands only faintly - no cell keeps
// a band's own colour, which would read as an echo.
func TestAFieldSharingTheMapWithAnImageIsItsLines(t *testing.T) {
	m := world(t, 120, 40)
	m.ColourDepth(tuimaps.Truecolor)
	must(t, m.Recentre(tuimaps.LonLat{Lon: -96, Lat: 37}))
	must(t, m.Zoom(3))
	mustSet(t, m, tuimaps.TemperatureGrid("temp", warmField(), tuimaps.Celsius, noon))
	settle(t, m)
	f, err := m.Render(tuimaps.Size{Cols: 120, Rows: 40}, noon)
	if err != nil {
		t.Fatal(err)
	}
	alone := strings.Join(f.Lines, "\n")
	aloneBody := colours.ReplaceAllString(strings.Join(f.Lines[:len(f.Lines)-1], "\n"), "")
	if bandGrounds(m, alone) == 0 {
		t.Fatal("the field alone painted no cell in a band's colour: nothing to compare")
	}
	if aValue.MatchString(aloneBody) {
		t.Fatalf("the field alone wrote a value on the map; its bands say it:\n%s", aloneBody)
	}

	must(t, m.SetPlayback(tuimaps.PlaybackOn))
	m.ShowStamp(false) // the loop's stamp writes its time on the frame: only a contour may carry a number here
	mustSet(t, m, loopAt(t, "radar", []int{0}, nil, nil))
	settle(t, m)
	f, err = m.Render(tuimaps.Size{Cols: 120, Rows: 40}, noon)
	if err != nil {
		t.Fatal(err)
	}
	shared := strings.Join(f.Lines, "\n")
	body := colours.ReplaceAllString(strings.Join(f.Lines[:len(f.Lines)-1], "\n"), "")
	if !aValue.MatchString(body) {
		t.Errorf("with an image on the map the field's contours carry no value:\n%s", body)
	}
	if n := bandGrounds(m, shared); n != 0 {
		t.Errorf("with an image on the map %d cells keep a band's own colour; the bands are faint there", n)
	}
}
