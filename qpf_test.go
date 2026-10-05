package tuimaps_test

// qpf_test.go — v0.2.0 L-26 (watchpost D-168, D-184): a period's rain and
// snow totals, liquid-equivalent, in the NWS WPC's breaks - below a trace
// nothing is drawn, as below radar's first floor.

import (
	"strings"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// TestRainTotalsAreWPCsScale is L-26.1: a totals grid keys WPC's seven
// classes from 0.25 mm (a hundredth of an inch) to 101.6 mm (four inches),
// and below the first, a trace, is keyed but never drawn.
func TestRainTotalsAreWPCsScale(t *testing.T) {
	m := reportMap(t)
	m.ColourDepth(tuimaps.Truecolor)
	mustSet(t, m, tuimaps.QPFGrid("qpf", flatField(30), noon))
	settle(t, m)
	var entry *tuimaps.LegendEntry
	for _, e := range m.Legend() {
		if e.Preset == "qpf" {
			entry = &e
		}
	}
	if entry == nil {
		t.Fatal("the legend has no qpf")
	}
	if len(entry.Classes) != 8 || entry.Classes[0].Drawn || !strings.Contains(entry.Classes[1].Label, "0.25") || !strings.Contains(entry.Classes[3].Label, "6.35") || !strings.Contains(entry.Classes[7].Label, "101.6") {
		t.Fatalf("qpf's legend is %+v; want a trace undrawn, then seven classes from 0.25 to 101.6 mm", entry.Classes)
	}
	if groundsOf(entry.Classes, momentFrame(t, m, noon)) == 0 {
		t.Error("30 mm was not drawn")
	}
	trace := reportMap(t)
	trace.ColourDepth(tuimaps.Truecolor)
	mustSet(t, trace, tuimaps.QPFGrid("qpf", flatField(0.1), noon))
	settle(t, trace)
	for _, e := range trace.Legend() {
		if e.Preset == "qpf" && groundsOf(e.Classes, momentFrame(t, trace, noon)) != 0 {
			t.Error("a trace was drawn: below the first floor is nothing")
		}
	}
}

// TestRainTotalsAreChecked is L-26.2: the preset in mm alone, or refused.
func TestRainTotalsAreChecked(t *testing.T) {
	m := reportMap(t)
	g := flatField(5)
	g.Type = tuimaps.Type{Preset: "qpf", Unit: "in"}
	if _, err := m.Set(tuimaps.Overlay{ID: "x", Valid: noon, Keeps: 3600e9, Grid: &g}); !isKind(err, fault.UnknownPreset) {
		t.Errorf("qpf in inches: %v; want refused", err)
	}
}
