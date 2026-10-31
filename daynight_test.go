package tuimaps

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/assets"
	"github.com/branden-thompson/go-tuimaps/internal/sun"
)

// TestReportSaysDayOrNight is L-4.4: a place's report says whether it is
// day or night at the map's clock and when the terminator next passes it;
// before a frame gives the map a clock, it says it does not know.
func TestReportSaysDayOrNight(t *testing.T) {
	m, err := New(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	places := []Place{{Name: "Wichita", At: LonLat{Lon: -97.3, Lat: 37.7}}, {Name: "Sydney", At: LonLat{Lon: 151.2, Lat: -33.9}}}
	r, err := m.Report(places)
	if err != nil || r.Places[0].Daylight.Known {
		t.Fatalf("before any frame the day is known: %+v %v", r.Places[0].Daylight, err)
	}
	at := time.Date(2026, 10, 11, 18, 0, 0, 0, time.UTC) // noon in Wichita, 5 in the morning in Sydney
	if _, err := m.Render(Size{Cols: 80, Rows: 24}, at); err != nil {
		t.Fatal(err)
	}
	r, err = m.Report(places)
	if err != nil {
		t.Fatal(err)
	}
	wichita, sydney := r.Places[0].Daylight, r.Places[1].Daylight
	if !wichita.Known || !wichita.Day || sydney.Day {
		t.Errorf("at 18:00 UTC Wichita is %+v and Sydney %+v; want day, then night", wichita, sydney)
	}
	if want := sun.Turn(-33.9, 151.2, at); !sydney.Turns.Equal(want) || !sydney.Turns.After(at) {
		t.Errorf("Sydney's sun turns at %v; want its next rising, %v", sydney.Turns, want)
	}
	if _, err := m.Render(Size{Cols: 80, Rows: 24}, sydney.Turns.Add(time.Minute)); err != nil { // past its sunrise: the kept report is worked out again
		t.Fatal(err)
	}
	if r, err = m.Report(places); err != nil || !r.Places[1].Daylight.Day || !r.Places[1].Daylight.Turns.After(sydney.Turns) {
		t.Errorf("a minute past Sydney's sunrise its day is %+v (%v)", r.Places[1].Daylight, err)
	}
	// A whole night between two asks (watchpost RC-14): asked about Wichita
	// alone, day at both, so the report's key is the same either side; its
	// turn held, the sunset, is past, and the report says the next.
	wichitaOnly := places[:1]
	noon, err := m.Report(wichitaOnly)
	if err != nil || !noon.Places[0].Daylight.Day {
		t.Fatalf("Wichita alone, a minute past Sydney's sunrise: %+v %v", noon.Places[0].Daylight, err)
	}
	next := sydney.Turns.Add(time.Minute + 24*time.Hour)
	if _, err := m.Render(Size{Cols: 80, Rows: 24}, next); err != nil {
		t.Fatal(err)
	}
	r, err = m.Report(wichitaOnly)
	if err != nil || !r.Places[0].Daylight.Day || !r.Places[0].Daylight.Turns.After(next) {
		t.Errorf("Wichita a day on: %+v, its turn %v of a clock at %v (held %v); want a turn still to come", r.Places[0].Daylight, r.Places[0].Daylight.Turns, next, noon.Places[0].Daylight.Turns)
	}
}

// TestTheNightIsShadedAndUnlisted is L-4.3 on the map: an overlay of the
// night side changes the frame - its ground is shaded - and adds no legend
// entry and none of its words.
func TestTheNightIsShadedAndUnlisted(t *testing.T) {
	at := time.Date(2026, 10, 11, 18, 0, 0, 0, time.UTC)
	frame := func(withNight bool) (string, int) {
		m, err := New(WithSize(120, 40), Embed(assets.Tile, assets.MaxZoom))
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		if withNight {
			o := Overlay{ID: "night", Valid: at, Keeps: time.Hour, Features: []Feature{{Kind: Polygon, Rings: NightSide(at), Role: NightTint, Label: "NIGHTWORD"}}} // the shade alone: no line to change the frame
			if _, err := m.Set(o); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := m.Settle(context.Background()); err != nil {
			t.Fatal(err)
		}
		f, err := m.Render(Size{Cols: 120, Rows: 40}, at)
		if err != nil {
			t.Fatal(err)
		}
		return strings.Join(f.Lines, "\n"), len(m.Legend())
	}
	day, _ := frame(false)
	night, entries := frame(true)
	if night == day {
		t.Error("the night side changed nothing in the frame")
	}
	if strings.Contains(night, "NIGHTWORD") || entries != 0 {
		t.Errorf("the night was drawn with its words (%v) or listed in the legend (%d entries)", strings.Contains(night, "NIGHTWORD"), entries)
	}
}

// TestDaylightAtIsTheNightDrawn is watchpost D-263: a place's day or night
// at a moment is the night NightOverlay draws for that moment - a host
// showing an hour ahead says what its picture shows, not the frame's clock;
// at the map's clock it is Report's, and the zero time is not known.
func TestDaylightAtIsTheNightDrawn(t *testing.T) {
	now := time.Date(2026, 10, 11, 18, 0, 0, 0, time.UTC) // noon in Wichita
	for _, at := range []time.Time{now, now.Add(9 * time.Hour), time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC), time.Date(2026, 12, 21, 0, 0, 0, 0, time.UTC)} {
		line := Terminator(at)
		pole := 90.0
		if NightSide(at)[0][len(line)].Lat < 0 { // the ring closes through the night's pole
			pole = -90
		}
		checked := 0
		for _, p := range line {
			for lat := -80.0; lat <= 80; lat += 10 {
				if math.Abs(lat-p.Lat) < 2 {
					continue // within the line's own tolerance of the edge
				}
				drawnNight := (pole > 0) == (lat > p.Lat)
				if d := DaylightAt(LonLat{Lon: p.Lon, Lat: lat}, at); !d.Known || d.Day == drawnNight {
					t.Fatalf("at %v, %.0f°, %.0f°: DaylightAt says %+v; the night drawn says night=%v", at, lat, p.Lon, d, drawnNight)
				}
				checked++
			}
		}
		if checked < 3000 {
			t.Fatalf("at %v only %d places checked: this measures little", at, checked)
		}
	}
	wichita := LonLat{Lon: -97.3, Lat: 37.7}
	if DaylightAt(wichita, now.Add(9*time.Hour)).Day {
		t.Error("Wichita at 03:00 UTC is said to be in day")
	}
	m, err := New(WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if _, err := m.Render(Size{Cols: 80, Rows: 24}, now); err != nil {
		t.Fatal(err)
	}
	r, err := m.Report([]Place{{Name: "Wichita", At: wichita}})
	if err != nil || r.Places[0].Daylight != DaylightAt(wichita, now) {
		t.Errorf("at the map's clock Report says %+v (%v); DaylightAt %+v", r.Places[0].Daylight, err, DaylightAt(wichita, now))
	}
	if d := DaylightAt(wichita, time.Time{}); d.Known {
		t.Errorf("the zero time is known: %+v", d)
	}
	for _, off := range []LonLat{{Lon: 0, Lat: math.NaN()}, {Lon: 0, Lat: 1000}, {Lon: 1e300, Lat: 0}, {Lon: math.Inf(1), Lat: 0}} {
		if d := DaylightAt(off, now); d.Known {
			t.Errorf("a place off the globe, %+v, is known: %+v", off, d)
		}
	}
}
