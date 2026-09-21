package overlay

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

func isKind(err error, k fault.Kind) bool {
	var f *fault.Error
	return errors.As(err, &f) && f.Kind() == k
}

var noon = time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)

// TestFreshnessStates is plan task 10.21 (FR-32).
func TestFreshnessStates(t *testing.T) {
	hour := time.Hour
	cases := []struct {
		name  string
		valid time.Time
		keeps time.Duration
		want  Freshness
	}{
		{"valid ten minutes ago, current for an hour", noon.Add(-10 * time.Minute), hour, Current},
		{"valid exactly an hour ago", noon.Add(-hour), hour, Current},
		{"valid an hour and a second ago", noon.Add(-hour - time.Second), hour, Stale},
		{"valid four minutes ahead: clocks differ", noon.Add(4 * time.Minute), hour, Current},
		{"valid six minutes ahead: a bad timestamp", noon.Add(6 * time.Minute), hour, Uncertain},
		{"valid a year ahead", noon.AddDate(1, 0, 0), hour, Uncertain},
	}
	for _, c := range cases {
		if got := FreshnessAt(c.valid, c.keeps, noon); got != c.want {
			t.Errorf("%s: %v, want %v", c.name, got, c.want)
		}
	}
	if !Uncertain.DrawnStale() || !Stale.DrawnStale() || Current.DrawnStale() {
		t.Error("uncertain is treated as stale on the frame, so a bad timestamp cannot keep data looking fresh")
	}
	for name, keeps := range map[string]time.Duration{"zero": 0, "negative": -time.Second, "over seven days": 7*24*hour + time.Second} {
		if err := CheckCurrency(noon, keeps); !isKind(err, fault.BadCurrency) {
			t.Errorf("currency %s: %v", name, err)
		}
	}
	if err := CheckCurrency(noon, 7*24*hour); err != nil {
		t.Errorf("seven days exactly: %v", err)
	}
	if err := CheckCurrency(time.Time{}, hour); !isKind(err, fault.BadCurrency) {
		t.Errorf("no valid time at all: %v; every overlay carries one", err)
	}
	// When the next change comes, for the host's NextCall.
	if at, ok := NextChange(noon.Add(-10*time.Minute), hour, noon); !ok || !at.Equal(noon.Add(50*time.Minute).Add(time.Nanosecond)) {
		t.Errorf("a current overlay goes stale at %v, %v", at, ok)
	}
	if _, ok := NextChange(noon.Add(-2*hour), hour, noon); ok {
		t.Error("a stale overlay has no next change")
	}
}

// TestGridClassify is plan task 10.12: a value's class is the number of
// breaks at or below it; a value that is no number is no data.
func TestGridClassify(t *testing.T) {
	breaks := []float64{0, 10, 20}
	for value, want := range map[float64]int{-5: 0, -0.001: 0, 0: 1, 9.99: 1, 10: 2, 19: 2, 20: 3, 1e9: 3, math.Inf(1): NoData, math.Inf(-1): NoData} {
		if got := Classify(value, breaks); got != want {
			t.Errorf("%v is class %d, want %d", value, got, want)
		}
	}
	if Classify(math.NaN(), breaks) != NoData {
		t.Error("a value that is no number has a class")
	}
	celsius := colour.TemperatureBreaks(colour.Celsius)
	if got := Classify(-0.5, celsius); got != 6 {
		t.Errorf("-0.5 C is class %d, want 6: the pale class just below freezing", got)
	}
	if got := Classify(0, celsius); got != 7 {
		t.Errorf("0 C is class %d, want 7", got)
	}
}

// Plan task 10.13 (D-69).
func TestGridPresetSuppliesBreaks(t *testing.T) {
	kind, err := ResolveType(Type{Preset: "temperature", Unit: "C"})
	if err != nil || len(kind.Breaks) != 16 || kind.Breaks[6] != 0 || kind.Preset != colour.Temperature {
		t.Fatalf("%+v, %v", kind, err)
	}
	fahrenheit, err := ResolveType(Type{Preset: "temperature", Unit: "F"})
	if err != nil || fahrenheit.Breaks[6] != 32 {
		t.Errorf("%+v, %v", fahrenheit, err)
	}
	radar, err := ResolveType(Type{Preset: "radar", Unit: "dBZ"})
	if err != nil || len(radar.Breaks) != 6 || radar.Breaks[0] != 10 {
		t.Errorf("%+v, %v", radar, err)
	}
	// A preset may be overridden: the host's breaks stand, and it is still the preset's colours.
	own, err := ResolveType(Type{Preset: "temperature", Unit: "C", Breaks: []float64{-10, 0, 10}})
	if err != nil || len(own.Breaks) != 3 {
		t.Errorf("%+v, %v", own, err)
	}
	if _, err := ResolveType(Type{Preset: "rainbow", Unit: "C"}); !isKind(err, fault.UnknownPreset) {
		t.Errorf("an unknown preset: %v", err)
	}
	if _, err := ResolveType(Type{Preset: "temperature", Unit: "K"}); !isKind(err, fault.UnknownPreset) {
		t.Errorf("a unit the preset does not have: %v", err)
	}
}

func TestOwnTypeNeedsBreaks(t *testing.T) {
	if _, err := ResolveType(Type{Unit: "mm"}); !isKind(err, fault.UnsortedBreaks) {
		t.Errorf("a host's own type with no breaks: %v", err)
	}
	for name, breaks := range map[string][]float64{"out of order": {0, 10, 5}, "a repeat": {0, 10, 10}, "not a number": {0, math.NaN(), 10}, "infinite": {0, math.Inf(1)}} {
		if _, err := ResolveType(Type{Unit: "mm", Breaks: breaks}); !isKind(err, fault.UnsortedBreaks) {
			t.Errorf("breaks %s: %v", name, err)
		}
	}
	kind, err := ResolveType(Type{Unit: "mm", Breaks: []float64{1, 5, 25}})
	if err != nil || kind.Preset != 0 || len(kind.Breaks) != 3 {
		t.Errorf("%+v, %v", kind, err)
	}
}

func TestClassCap(t *testing.T) {
	breaks := make([]float64, 21)
	for i := range breaks {
		breaks[i] = float64(i)
	}
	if _, err := ResolveType(Type{Unit: "mm", Breaks: breaks}); !isKind(err, fault.UnsortedBreaks) {
		t.Errorf("21 breaks make 22 classes: %v; 21 classes is the most a safe scale was found to carry", err)
	}
	if _, err := ResolveType(Type{Unit: "mm", Breaks: breaks[:20]}); err != nil {
		t.Errorf("20 breaks make 21 classes: %v", err)
	}
}
