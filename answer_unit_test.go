package tuimaps

import (
	"context"
	"testing"
)

// TestEveryFieldAnswerCarriesItsUnit is L-3.3 (M3): a place's answer against
// a field says its value's unit - the unit the host declared, temperature's
// in words.
func TestEveryFieldAnswerCarriesItsUnit(t *testing.T) {
	g := coolNorth()
	for _, c := range []struct {
		o    Overlay
		want string
	}{
		{MUFGrid("f", g, drawnAt), "MHz"},
		{FoF2Grid("f", g, drawnAt), "MHz"},
		{UVGrid("f", g, drawnAt), "index"},
		{QPFGrid("f", g, drawnAt), "mm"},
		{TemperatureGrid("f", g, Celsius, drawnAt), "degrees Celsius"},
	} {
		m, err := New(WithSize(80, 24))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := m.Set(c.o); err != nil {
			t.Fatal(err)
		}
		if _, err := m.Settle(context.Background()); err != nil {
			t.Fatal(err)
		}
		r, err := m.Report([]Place{{Name: "Wichita", At: LonLat{Lon: -97.3, Lat: 37.7}}})
		if err != nil || len(r.Places) != 1 || len(r.Places[0].Answers) != 1 {
			t.Fatalf("%s: %+v %v", c.o.Grid.Type.Preset, r, err)
		}
		if got := r.Places[0].Answers[0].ValueUnit; got != c.want {
			t.Errorf("%s: the answer's unit is %q; want %q", c.o.Grid.Type.Preset, got, c.want)
		}
		m.Close()
	}
}
