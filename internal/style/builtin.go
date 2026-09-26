package style

import "github.com/branden-thompson/go-tuimaps/internal/colour"

// classIn is the filter "its class is one of these".
func classIn(texts ...string) step {
	n := &node{op: "in", key: "class"}
	for _, t := range texts {
		n.values = append(n.values, value{text: t, isText: true})
	}
	return step{leaf: n}
}

// level is the filter "a boundary of this administrative level, or the one
// below it, on land".
func level(lo, hi float64) []step {
	return []step{
		{leaf: &node{op: ">=", key: "admin_level", values: []value{{number: lo, isWhole: true}}}},
		{leaf: &node{op: "<=", key: "admin_level", values: []value{{number: hi, isWhole: true}}}},
		{leaf: &node{op: "!=", key: "maritime", values: []value{{number: 1, isWhole: true}}}},
		{group: "all", count: 3},
	}
}

// named is the filter "it has a name", and with classes "and is one of these".
func named(classes ...string) []step {
	has := step{leaf: &node{op: "has", key: "name"}}
	if len(classes) == 0 {
		return []step{has}
	}
	return []step{has, classIn(classes...), {group: "all", count: 2}}
}

// BuiltIn is the library's own style, written against OpenMapTiles (D-24).
// Dark and bright are this one set of roles: a rule names a token, and the
// ground in effect picks between the token's two defaults (FR-20, D-64).
//
// The finer roles start at the zoom where they can be read: at world scale
// a region's border, a river or a road is clutter. The starting zooms are a
// first setting, made from the first frames drawn, and are HUM LEAD's to tune.
//
// Priority is what the basemap gives up, first to last, when it must thin
// itself (D-83): this is not a navigation tool, so roads go first, and
// coast, water and borders go last.
func BuiltIn() *Style {
	line := func(id, layer string, token colour.Token, priority int, from float64, filter ...step) Rule {
		return Rule{ID: id, Layer: layer, Kind: Line, Token: token, priority: priority, MinZoom: from, filter: filter}
	}
	return levelled(&Style{rules: []Rule{
		// THE SEA AND ITS COAST ARE NEVER SWITCHED (watchpost UAT-1 U1-39): a
		// map with its water off keeps the land's edge. Rules match in order,
		// so the ocean is the coast's and the sea's, and every other water
		// area falls through to the edge and fill the Water switch takes.
		{ID: "coast", Layer: "water", Kind: Line, Token: colour.Coast, priority: 10, filter: []step{classIn("ocean")}, always: true},
		{ID: "sea", Layer: "water", Kind: Fill, Token: colour.WaterFill, priority: 9, filter: []step{classIn("ocean")}, always: true},
		line("water-edge", "water", colour.WaterLine, 9, 0),
		{ID: "water", Layer: "water", Kind: Fill, Token: colour.WaterFill, priority: 9},
		line("border-country", "boundary", colour.BorderCountry, 8, 0, level(1, 2)...),
		line("river", "waterway", colour.River, 7, 3, classIn("river", "canal")),
		line("border-region", "boundary", colour.BorderRegion, 5, 3, level(3, 4)...),
		line("park", "park", colour.Park, 4, 6),
		line("rail", "transportation", colour.Rail, 3, 8, classIn("rail")),
		line("road-major", "transportation", colour.RoadMajor, 2, 5, classIn("motorway", "trunk", "primary")),
		line("road-minor", "transportation", colour.RoadMinor, 1, 10, classIn("secondary", "tertiary", "minor")),
		line("runway", "aeroway", colour.Runway, 3, 10, classIn("runway", "taxiway")),
		{ID: "label-region", Layer: "place", Kind: Symbol, Token: colour.LabelRegion, priority: 6, filter: named("country", "state", "province", "continent")},
		{ID: "label-place", Layer: "place", Kind: Symbol, Token: colour.LabelPlace, priority: 6, filter: named()},
		{ID: "label-water", Layer: "water_name", Kind: Symbol, Token: colour.LabelWater, priority: 6, filter: named()},
	}})
}

// levelled gives each built-in rule its level (D-82).
func levelled(s *Style) *Style {
	for i := range s.rules {
		s.rules[i].detail = ruleDetail[s.rules[i].ID]
	}
	return s
}
