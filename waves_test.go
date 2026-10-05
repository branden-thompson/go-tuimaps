package tuimaps_test

// waves_test.go — v0.2.0 L-20 (watchpost D-125, D-126): wave height, a
// field of its own scale drawn over the sea alone.

import (
	"strings"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// waveField is 5 feet over a box.
func waveField(w, s, e, n float64) tuimaps.Grid {
	g := tuimaps.Grid{West: w, South: s, East: e, North: n, Cols: 8, Rows: 8}
	for range 64 {
		g.Values = append(g.Values, 5)
	}
	return g
}

// waveGrounds counts the frame's cells painted in one of the wave scale's
// colours.
func waveGrounds(m *tuimaps.Map, raw string) int {
	for _, e := range m.Legend() {
		if e.Preset == "waves" {
			return groundsOf(e.Classes, raw)
		}
	}
	return 0
}

// TestWavesAreDrawnOverTheSeaAlone is L-20.1: a wave grid over the Gulf
// coast is drawn over its sea; the same over Arkansas draws nothing - the
// inverse of a temperature, which stops at the shore.
func TestWavesAreDrawnOverTheSeaAlone(t *testing.T) {
	m := rainMap(t)
	mustSet(t, m, tuimaps.WaveGrid("waves", waveField(-94, 26, -84, 31), tuimaps.Feet, noon)) // the Gulf coast: its sea and its land
	settle(t, m)
	if waveGrounds(m, momentFrame(t, m, noon)) == 0 {
		t.Error("waves over the Gulf were not drawn")
	}
	mustSet(t, m, tuimaps.WaveGrid("waves", waveField(-94, 33, -89, 35), tuimaps.Feet, noon))
	settle(t, m)
	if n := waveGrounds(m, momentFrame(t, m, noon)); n != 0 {
		t.Errorf("waves over Arkansas painted %d runs of cells; the sea alone carries them", n)
	}
}

// TestTheWaveLegendIsItsSixClasses is L-20.2: six classes, calm first, in
// feet or metres, each set round in its unit.
func TestTheWaveLegendIsItsSixClasses(t *testing.T) {
	for unit, want := range map[tuimaps.WaveUnit][2]string{tuimaps.Feet: {"ft", "under 2"}, tuimaps.Metres: {"m", "under 0.5"}} {
		m := reportMap(t)
		mustSet(t, m, tuimaps.WaveGrid("waves", waveField(-94, 24, -86, 28.5), unit, noon))
		found := false
		for _, e := range m.Legend() {
			if e.Preset == "waves" {
				found = true
				if len(e.Classes) != 6 || e.Unit != want[0] || e.Classes[0].Label != want[1] {
					t.Errorf("the %s wave legend is %+v", want[0], e)
				}
			}
		}
		if !found {
			t.Errorf("the legend has no waves in %s", want[0])
		}
	}
}

// TestTheWavePresetIsCheckedAndNamed is L-20.3: the preset in feet or
// metres alone; its tokens wave.1 to wave.6.
func TestTheWavePresetIsCheckedAndNamed(t *testing.T) {
	m := reportMap(t)
	g := waveField(-94, 24, -86, 28.5)
	g.Type = tuimaps.Type{Preset: "waves", Unit: "fathoms"}
	if _, err := m.Set(tuimaps.Overlay{ID: "w", Valid: noon, Keeps: 3600e9, Grid: &g}); !isKind(err, fault.UnknownPreset) {
		t.Errorf("waves in fathoms: %v; want refused", err)
	}
	names := " " + strings.Join(tuimaps.TokenNames(), " ") + " "
	for _, n := range []string{"wave.1", "wave.6"} {
		if !strings.Contains(names, " "+n+" ") {
			t.Errorf("the tokens lack %s", n)
		}
	}
}
