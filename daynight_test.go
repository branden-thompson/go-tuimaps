package tuimaps

import (
	"context"
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
