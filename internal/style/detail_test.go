package style

// detail_test.go — v0.2.0 WP-L11, L11.1 and L11.2 (D-82): a host chooses how
// much of the basemap is drawn for its purpose, and switches major and minor
// roads apart (L-14).

import (
	"slices"
	"sort"
	"testing"
)

// drawnAt is the built-in rules a large, bare map draws at a level.
func drawnAt(d Detail, off Switches) []string {
	p := NewProfile(Bare, 200, 60, off).WithDetail(d)
	var ids []string
	for i := range BuiltIn().rules {
		r := &BuiltIn().rules[i]
		if p.Draws(r) {
			ids = append(ids, r.ID)
		}
	}
	sort.Strings(ids)
	return ids
}

func sorted(ids ...string) []string { sort.Strings(ids); return ids }

// TestEachDetailLevelDrawsItsRules: each level draws exactly its rules, each
// level everything the one below draws, and Full everything - the picture as
// it was for a host that never asks.
func TestEachDetailLevelDrawsItsRules(t *testing.T) {
	essential := []string{"coast", "sea", "water-edge", "water", "border-country", "border-region"}
	weather := append(slices.Clone(essential), "river", "label-region", "label-place", "label-water", "road-major")
	standard := append(slices.Clone(weather), "park", "rail")
	full := append(slices.Clone(standard), "road-minor", "runway")
	for _, c := range []struct {
		level Detail
		want  []string
	}{{DetailEssential, essential}, {DetailWeather, weather}, {DetailStandard, standard}, {DetailFull, full}} {
		if got := drawnAt(c.level, 0); !slices.Equal(got, sorted(c.want...)) {
			t.Errorf("%v draws %v, want %v", c.level, got, sorted(c.want...))
		}
	}
	var all []string
	for _, r := range BuiltIn().rules {
		all = append(all, r.ID)
	}
	if got := drawnAt(DetailFull, 0); !slices.Equal(got, sorted(all...)) {
		t.Errorf("Full draws %v, not every rule %v", got, sorted(all...))
	}
	if got := NewProfile(Bare, 200, 60, 0); got.Detail() != DetailFull {
		t.Errorf("a profile nobody set a level on is at %v, not Full", got.Detail())
	}
}

// TestAHostsOwnRulesDrawAtEveryLevel: only the built-in style is ranked; a
// host's own rule is drawn at every level.
func TestAHostsOwnRulesDrawAtEveryLevel(t *testing.T) {
	own := &Rule{ID: "theirs", Layer: "transportation", Kind: Line}
	if !NewProfile(Bare, 200, 60, 0).WithDetail(DetailEssential).Draws(own) {
		t.Error("a host's own rule was thinned by the level")
	}
}

// TestMajorAndMinorRoadsSwitchApart is L11.2: the minor roads off leaves the
// major; the major off leaves the minor.
func TestMajorAndMinorRoadsSwitchApart(t *testing.T) {
	if got := drawnAt(DetailFull, Off(MinorRoadLayer)); slices.Contains(got, "road-minor") || !slices.Contains(got, "road-major") {
		t.Errorf("minor roads off draws %v", got)
	}
	if got := drawnAt(DetailFull, Off(RoadLayer)); slices.Contains(got, "road-major") || !slices.Contains(got, "road-minor") {
		t.Errorf("major roads off draws %v", got)
	}
}

// TestADetailOutsideTheFourIsNotALevel: the levels are the four.
func TestADetailOutsideTheFourIsNotALevel(t *testing.T) {
	for _, d := range []Detail{DetailEssential, DetailWeather, DetailStandard, DetailFull} {
		if !d.Valid() {
			t.Errorf("%v is not valid", d)
		}
	}
	if Detail(0).Valid() || Detail(9).Valid() {
		t.Error("a value outside the four is valid")
	}
}
