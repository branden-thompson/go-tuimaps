package style

import "testing"

// ruleOf is the built-in style's rule of an id.
func ruleOf(t *testing.T, id string) *Rule {
	t.Helper()
	s := BuiltIn()
	for i := range s.rules {
		if s.rules[i].ID == id {
			return &s.rules[i]
		}
	}
	t.Fatalf("the built-in style has no rule %q", id)
	return nil
}

// TestProfileGivesUpInOrder is FR-19 and D-83: what the basemap gives up, and
// in what order. Roads go first; coast, water and borders go last.
func TestProfileGivesUpInOrder(t *testing.T) {
	order := []string{"road-minor", "road-major", "rail", "park", "border-region", "river", "border-country", "water-edge", "coast"}
	was := 0
	for _, id := range order {
		r := ruleOf(t, id)
		if r.priority <= was {
			t.Fatalf("%s ranks %d, no later than the role before it; the order given up is %v", id, r.priority, order)
		}
		was = r.priority
	}
	// A floor of n draws everything ranked n and above, and nothing below it.
	for floor := 1; floor <= was; floor++ {
		p := Profile{floor: floor}
		for _, id := range order {
			r := ruleOf(t, id)
			if got, want := p.Draws(r), r.priority >= floor; got != want {
				t.Errorf("at floor %d, %s (rank %d) drawn %v", floor, id, r.priority, got)
			}
		}
	}
}

// TestProfileThinsUnderOverlay is plan task 09.8 (FR-19): with a field or an
// image over it, the basemap drops its minor roads and draws what is left of
// its roads one dot thick, while the coast, water and borders are untouched.
func TestProfileThinsUnderOverlay(t *testing.T) {
	full := NewProfile(Bare, 149, 38, 0)
	under := NewProfile(Covered, 149, 38, 0)
	if !full.Draws(ruleOf(t, "road-minor")) {
		t.Error("a full profile draws minor roads")
	}
	if under.Draws(ruleOf(t, "road-minor")) {
		t.Error("minor roads are drawn under a field")
	}
	major := ruleOf(t, "road-major")
	if !under.Draws(major) {
		t.Error("major roads go under a field; only their weight does")
	}
	if got := under.Weight(major, 12); got != 1 {
		t.Errorf("a major road is %v dots thick under a field, want 1", got)
	}
	if got, want := full.Weight(major, 12), major.Width(12); got != want {
		t.Errorf("a major road is %v dots thick with nothing over it, want its own %v", got, want)
	}
	for _, id := range []string{"coast", "water-edge", "border-country", "border-region", "river"} {
		r := ruleOf(t, id)
		if !under.Draws(r) || under.Weight(r, 12) != r.Width(12) {
			t.Errorf("%s changed under a field; geography and topography go last (D-83)", id)
		}
	}
	if full.Labels() != 0 || under.Labels() != coveredLabels {
		t.Errorf("label budgets %d and %d; a full map has no bound, a covered one has %d", full.Labels(), under.Labels(), coveredLabels)
	}
}

// TestProfileBySize is the other half of 09.8: a map smaller than 80x20 gives
// up one more thing than its load alone asks for, and places fewer names.
func TestProfileBySize(t *testing.T) {
	cases := []struct {
		name       string
		load       Load
		cols, rows int
		floor      int
		labels     int
	}{
		{"a big bare map", Bare, 149, 38, 1, 0},
		{"a small bare map", Bare, 69, 12, 2, smallLabels},
		{"one cell short across", Bare, 79, 38, 2, smallLabels},
		{"one row short", Bare, 149, 19, 2, smallLabels},
		{"a big map under a field", Covered, 149, 38, 2, coveredLabels},
		{"a small map under a field", Covered, 69, 12, 3, coveredLabels / 2},
		{"a big sparse map", Dense, 149, 38, 3, denseLabels},
		{"a small sparse map", Dense, 69, 12, 4, denseLabels / 2},
	}
	for _, c := range cases {
		p := NewProfile(c.load, c.cols, c.rows, 0)
		if p.Floor() != c.floor || p.Labels() != c.labels {
			t.Errorf("%s: floor %d and %d labels, want %d and %d", c.name, p.Floor(), p.Labels(), c.floor, c.labels)
		}
	}
}

// TestLayerToggle is plan task 09.9 (FR-36): a layer the host has switched
// off is never drawn, whatever the profile would have drawn, and switching
// one off leaves the others where they were.
func TestLayerToggle(t *testing.T) {
	byLayer := map[Layer][]string{
		RoadLayer:      {"road-major"}, // v0.2.0 D-82: the major roads; the minor switch apart
		MinorRoadLayer: {"road-minor"},
		RailLayer:      {"rail", "runway"},
		ParkLayer:      {"park"},
		BorderLayer:    {"border-country", "border-region"},
		RiverLayer:     {"river"},
		WaterLayer:     {"water-edge", "water"}, // inland water; the sea is never switched (watchpost U1-39)
		LabelLayer:     {"label-place", "label-region", "label-water"},
	}
	for layer, ids := range byLayer {
		p := NewProfile(Bare, 149, 38, Off(layer))
		for _, id := range ids {
			if p.Draws(ruleOf(t, id)) {
				t.Errorf("%s is drawn with its layer switched off", id)
			}
		}
		for other, others := range byLayer {
			if other == layer {
				continue
			}
			for _, id := range others {
				if !p.Draws(ruleOf(t, id)) {
					t.Errorf("switching off layer %d also took %s", layer, id)
				}
			}
		}
	}
	both := NewProfile(Bare, 149, 38, Off(RoadLayer, LabelLayer))
	if both.Draws(ruleOf(t, "road-major")) || both.Draws(ruleOf(t, "label-place")) || !both.Draws(ruleOf(t, "coast")) {
		t.Error("two layers switched off at once")
	}
	if Off().Has(RoadLayer) || Off(RoadLayer).Has(Layer(0)) || Off(Layer(99)) != 0 {
		t.Error("a switch set of nothing, or of a layer that does not exist")
	}
}

// TestProfileKeepsAUsersOwnStyle: only the built-in style's roles are ranked,
// so a profile never drops a rule of a style a host wrote - but the host's
// own switches still apply to it.
func TestProfileKeepsAUsersOwnStyle(t *testing.T) {
	own := &Rule{ID: "their-roads", Layer: "transportation", Kind: Line}
	p := NewProfile(Dense, 69, 12, 0)
	if !p.Draws(own) {
		t.Error("a rule of a style the host wrote was dropped by a profile")
	}
	if NewProfile(Bare, 149, 38, Off(RoadLayer)).Draws(own) {
		t.Error("a rule of a style the host wrote outlived the host's own switch")
	}
}
