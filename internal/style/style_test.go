package style

import (
	"errors"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
	"os"
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

func TestMain(m *testing.M) { os.Exit(testkit.Main(m)) }

func isKind(err error, k fault.Kind) bool {
	var f *fault.Error
	return errors.As(err, &f) && f.Kind() == k
}

func rgb(r, g, b uint8) colour.RGB { return colour.RGB{R: r, G: g, B: b} }

func line(class string) Attrs { return Attrs{Kind: scene.GeomLine, Class: class} }

func mustParse(t *testing.T, body string) *Style {
	t.Helper()
	s, err := Parse([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// TestBuiltInStylesCoverEveryRole is plan task 08.21 (FR-20, D-83). Dark and
// bright are one set of roles; the ground in effect picks the colours.
func TestBuiltInStylesCoverEveryRole(t *testing.T) {
	s := BuiltIn()
	cases := []struct {
		layer string
		attrs Attrs
		want  colour.Token
	}{
		{"water", Attrs{Kind: scene.GeomPolygon, Class: "ocean"}, colour.Coast},
		{"water", Attrs{Kind: scene.GeomPolygon, Class: "lake"}, colour.WaterLine},
		{"waterway", line("river"), colour.River},
		{"boundary", Attrs{Kind: scene.GeomLine, AdminLevel: 2}, colour.BorderCountry},
		{"boundary", Attrs{Kind: scene.GeomLine, AdminLevel: 4}, colour.BorderRegion},
		{"transportation", line("motorway"), colour.RoadMajor},
		{"transportation", line("trunk"), colour.RoadMajor},
		{"transportation", line("primary"), colour.RoadMajor},
		{"transportation", line("secondary"), colour.RoadMinor},
		{"transportation", line("minor"), colour.RoadMinor},
		{"transportation", line("rail"), colour.Rail},
		{"park", Attrs{Kind: scene.GeomPolygon, Class: "national_park"}, colour.Park},
		{"aeroway", line("runway"), colour.Runway},
		{"place", Attrs{Kind: scene.GeomPoint, Class: "city", Name: "Chicago", Rank: 3}, colour.LabelPlace},
		{"place", Attrs{Kind: scene.GeomPoint, Class: "country", Name: "Cuba", Rank: 1}, colour.LabelRegion},
		{"place", Attrs{Kind: scene.GeomPoint, Class: "state", Name: "Ohio", Rank: 2}, colour.LabelRegion},
		{"water_name", Attrs{Kind: scene.GeomPoint, Class: "ocean", Name: "Atlantic Ocean"}, colour.LabelWater},
	}
	for _, c := range cases {
		rule, ok := s.Match(c.layer, c.attrs, 12)
		if !ok || rule.Token != c.want {
			t.Errorf("%s %+v: %+v, %v; want token %s", c.layer, c.attrs, rule, ok, c.want.Name())
		}
	}
	// A line across the sea is not drawn; nor is what no role covers.
	for _, c := range []struct {
		layer string
		attrs Attrs
	}{
		{"boundary", Attrs{Kind: scene.GeomLine, AdminLevel: 2, Maritime: true}},
		{"boundary", Attrs{Kind: scene.GeomLine, AdminLevel: 8}},
		{"transportation", line("motorway_construction")},
		{"transportation", line("ferry")},
		{"poi", Attrs{Kind: scene.GeomPoint, Class: "shop"}},
		{"place", Attrs{Kind: scene.GeomPoint, Class: "city"}}, // a place with no name has nothing to say
	} {
		if rule, ok := s.Match(c.layer, c.attrs, 8); ok {
			t.Errorf("%s %+v matched %s", c.layer, c.attrs, rule.ID)
		}
	}

	// A small map of the whole world is below zoom 0, and is still drawn: a
	// rule that starts at zoom 0 has no lower bound.
	for _, zoom := range []float64{-3, -0.6, 0} {
		if _, ok := s.Match("water", Attrs{Kind: scene.GeomPolygon, Class: "ocean"}, zoom); !ok {
			t.Errorf("the coast is not drawn at zoom %v", zoom)
		}
		if _, ok := s.Match("place", Attrs{Kind: scene.GeomPoint, Class: "continent", Name: "Africa"}, zoom); !ok {
			t.Errorf("a continent's name is not drawn at zoom %v", zoom)
		}
	}

	// The finer roles wait for a zoom at which they can be read: at world
	// scale a region's border and a river are clutter, a country's is not.
	for _, c := range []struct {
		layer     string
		attrs     Attrs
		from      float64
		worldWide bool
	}{
		{"boundary", Attrs{Kind: scene.GeomLine, AdminLevel: 2}, 0, true},
		{"water", Attrs{Kind: scene.GeomPolygon, Class: "ocean"}, 0, true},
		{"boundary", Attrs{Kind: scene.GeomLine, AdminLevel: 4}, 3, false},
		{"waterway", line("river"), 3, false},
		{"transportation", line("motorway"), 5, false},
		{"transportation", line("rail"), 8, false},
		{"transportation", line("minor"), 10, false},
		{"park", Attrs{Kind: scene.GeomPolygon, Class: "national_park"}, 6, false},
	} {
		if _, ok := s.Match(c.layer, c.attrs, 1.2); ok != c.worldWide {
			t.Errorf("%s %+v at zoom 1.2: drawn=%v", c.layer, c.attrs, ok)
		}
		if _, ok := s.Match(c.layer, c.attrs, c.from); !ok {
			t.Errorf("%s %+v is not drawn from zoom %v", c.layer, c.attrs, c.from)
		}
	}

	// What the basemap gives up, first to last (D-83): roads first; coast,
	// water and borders last.
	order := []colour.Token{colour.RoadMinor, colour.RoadMajor, colour.Rail, colour.Park, colour.BorderRegion, colour.River, colour.BorderCountry, colour.WaterLine, colour.Coast}
	last := -1
	for _, tok := range order {
		p, ok := s.Priority(tok)
		if !ok || p <= last {
			t.Errorf("%s has priority %d (%v); each role in the order outlasts the one before", tok.Name(), p, ok)
		}
		last = p
	}
	if _, ok := s.Match("water", Attrs{Kind: scene.GeomPolygon, Class: "ocean"}, 8); !ok {
		t.Fatal("no fill rule")
	}
	if fill, ok := s.Fill("water", Attrs{Kind: scene.GeomPolygon, Class: "ocean"}, 8); !ok || fill.Token != colour.WaterFill {
		t.Errorf("water's fill: %+v, %v", fill, ok)
	}
}

const userStyle = `{
 "constants": {"@water": "#5f87ff", "@road": "#fff"},
 "layers": [
  {"id": "bg", "type": "background", "paint": {"background-color": "#000"}},
  {"id": "roads-major", "type": "line", "source-layer": "transportation",
   "filter": ["all", ["==", "$type", "LineString"], ["in", "class", "motorway", "trunk"]],
   "minzoom": 4, "maxzoom": 12,
   "paint": {"line-color": {"stops": [[4, "#888"], [8, "@road"], [10, "#ff0"]]}, "line-width": {"stops": [[4, 1], [8, 3]]}}},
  {"id": "roads-major-casing", "ref": "roads-major", "paint": {"line-color": "#123456"}},
  {"id": "roads-other", "type": "line", "source-layer": "transportation", "paint": {"fill-color": "@water"}},
  {"id": "lakes", "type": "fill", "source-layer": "water", "filter": ["!=", "class", "ocean"], "paint": {"fill-color": "@water"}},
  {"id": "names", "type": "symbol", "source-layer": "place", "filter": ["has", "name"], "paint": {"text-color": "#0f0"}},
  {"id": "plain", "type": "line", "source-layer": "waterway", "paint": {}}
 ]}`

// TestUserStyleLegacyFilters is plan task 08.17 (L-9): every zoom stop is
// honoured, and a style that needs expressions says so.
func TestUserStyleLegacyFilters(t *testing.T) {
	s := mustParse(t, userStyle)
	for zoom, want := range map[float64]colour.RGB{4: rgb(0x88, 0x88, 0x88), 7.9: rgb(0x88, 0x88, 0x88), 8: rgb(255, 255, 255), 9: rgb(255, 255, 255), 10: rgb(255, 255, 0), 11.5: rgb(255, 255, 0)} {
		rule, ok := s.Match("transportation", line("motorway"), zoom)
		if !ok || rule.ID != "roads-major" {
			t.Fatalf("zoom %v: %+v, %v", zoom, rule, ok)
		}
		if got, literal := rule.Colour(zoom); !literal || got != want {
			t.Errorf("zoom %v: colour %v, want %v; upstream read the first stop only", zoom, got, want)
		}
	}
	for _, zoom := range []float64{3.9, 12.1, 14} {
		if rule, _ := s.Match("transportation", line("motorway"), zoom); rule.ID != "roads-other" {
			t.Errorf("zoom %v matched %s; the layer's own zooms bound it", zoom, rule.ID)
		}
	}
}

// TestParityP27_ZoomGate: a layer's zooms are compared with the map's
// fractional zoom, and both ends are included.
func TestParityP27_ZoomGate(t *testing.T) {
	s := mustParse(t, userStyle)
	for zoom, want := range map[float64]string{3.99: "roads-other", 4: "roads-major", 8.5: "roads-major", 12: "roads-major", 12.01: "roads-other"} {
		if rule, _ := s.Match("transportation", line("motorway"), zoom); rule.ID != want {
			t.Errorf("zoom %v matched %s, want %s", zoom, rule.ID, want)
		}
	}
	for name, body := range map[string]string{
		"an expression filter": `{"layers":[{"id":"a","type":"line","source-layer":"water","filter":["match",["get","class"],"ocean",true,false]}]}`,
		"an expression colour": `{"layers":[{"id":"a","type":"line","source-layer":"water","paint":{"line-color":["case",["has","x"],"#fff","#000"]}}]}`,
		"not JSON":             `layers:`,
		"layers not a list":    `{"layers":{}}`,
		"a ref to nothing":     `{"layers":[{"id":"a","ref":"b"}]}`,
		"a ref to itself":      `{"layers":[{"id":"a","ref":"a"}]}`,
		"a ref in a ring":      `{"layers":[{"id":"a","ref":"b"},{"id":"b","ref":"a"}]}`,
		"a filter not a list":  `{"layers":[{"id":"a","type":"line","source-layer":"water","filter":"class"}]}`,
		"a constant in a ring": `{"constants":{"@a":"@b","@b":"@a"},"layers":[{"id":"a","type":"line","source-layer":"water","paint":{"line-color":"@a"}}]}`,
	} {
		_, err := Parse([]byte(body))
		if !isKind(err, fault.MalformedStyle) {
			t.Errorf("%s: %v; want the malformed-style kind", name, err)
		}
		if name == "an expression filter" && !strings.Contains(err.Error(), "expression") {
			t.Errorf("%v; the error must say that expressions are what is not supported", err)
		}
	}
	big := `{"layers":[],"pad":"` + strings.Repeat("x", 1<<20) + `"}`
	if _, err := Parse([]byte(big)); !isKind(err, fault.OverLimit) {
		t.Errorf("over 1 MiB: %v", err)
	}
	if _, err := Parse(nil); !isKind(err, fault.MalformedStyle) {
		t.Errorf("no style at all: %v", err)
	}
}

// TestParityP41_StyleMatch: the first matching layer, in style order, for
// the tile's layer; no match drops the feature; one style a feature.
func TestParityP41_StyleMatch(t *testing.T) {
	s := mustParse(t, userStyle)
	if rule, ok := s.Match("transportation", line("motorway"), 8); !ok || rule.ID != "roads-major" {
		t.Errorf("%+v, %v; the first match in style order wins, not the casing after it", rule, ok)
	}
	if rule, ok := s.Match("transportation", line("path"), 8); !ok || rule.ID != "roads-other" {
		t.Errorf("%+v, %v", rule, ok)
	}
	if _, ok := s.Match("building", Attrs{Kind: scene.GeomPolygon}, 8); ok {
		t.Error("a layer the style does not name matched")
	}
	if _, ok := s.Match("water", Attrs{Kind: scene.GeomPolygon, Class: "ocean"}, 8); ok {
		t.Error("a feature no filter accepts matched; it is dropped")
	}
}

// TestParityP42_FilterOps: upstream's operators, and its three oddities.
func TestParityP42_FilterOps(t *testing.T) {
	a := Attrs{Kind: scene.GeomLine, Class: "river", Name: "Ohio", Rank: 3, AdminLevel: 4}
	cases := map[string]bool{
		`["==","class","river"]`:                    true,
		`["==","class","stream"]`:                   false,
		`["!=","class","stream"]`:                   true,
		`["==","$type","LineString"]`:               true,
		`["==","$type","Polygon"]`:                  false,
		`["in","class","canal","river"]`:            true,
		`["!in","class","canal","river"]`:           false,
		`["has","name"]`:                            true,
		`["!has","name"]`:                           false,
		`["has","brunnel"]`:                         false,
		`["!has","brunnel"]`:                        true,
		`[">","rank",2]`:                            true,
		`[">=","rank",3]`:                           true,
		`["<","rank",3]`:                            false,
		`["<=","admin_level",4]`:                    true,
		`["all",["has","name"],[">","rank",5]]`:     false,
		`["any",["has","brunnel"],[">","rank",2]]`:  true,
		`["none",["has","brunnel"],[">","rank",5]]`: true,
		`["all"]`:                   true,
		`["any"]`:                   false,
		`["==","brunnel","tunnel"]`: false, // == on a missing key is false
		`["!=","brunnel","tunnel"]`: true,
		`["==","rank",3]`:           true,
		`["==","rank",3.0]`:         false, // an integer and a float are distinct
		`["==","rank","3"]`:         false,
		`["within","class","x"]`:    true, // an operator upstream does not know evaluates to true
	}
	for filter, want := range cases {
		body := `{"layers":[{"id":"a","type":"line","source-layer":"w","filter":` + filter + `}]}`
		s := mustParse(t, body)
		if _, got := s.Match("w", a, 8); got != want {
			t.Errorf("%s: %v, want %v", filter, got, want)
		}
	}
}

// TestParityP43_ConstantsAndRef: @name strings are substituted; a ref
// inherits type, source layer, zooms and filter, and keeps its own paint.
func TestParityP43_ConstantsAndRef(t *testing.T) {
	s := mustParse(t, userStyle)
	lakes, ok := s.Match("water", Attrs{Kind: scene.GeomPolygon, Class: "lake"}, 8)
	if got, _ := lakes.Colour(8); !ok || got != rgb(0x5f, 0x87, 0xff) {
		t.Errorf("a constant: %v, %v", got, ok)
	}
	var casing *Rule
	for i := range s.rules {
		if s.rules[i].ID == "roads-major-casing" {
			casing = &s.rules[i]
		}
	}
	if casing == nil || casing.Layer != "transportation" || casing.Kind != Line || casing.MinZoom != 4 || casing.MaxZoom != 12 {
		t.Fatalf("the ref did not inherit: %+v", casing)
	}
	if !casing.accepts(line("motorway")) || casing.accepts(line("path")) {
		t.Error("the ref did not inherit the filter")
	}
	if got, _ := casing.Colour(8); got != rgb(0x12, 0x34, 0x56) {
		t.Errorf("the ref's own paint: %v", got)
	}
}

// TestParityP44_ColourPick: line colour, else fill, else text; upstream's
// red when there is none or it cannot be read; hex of 3 or 6 digits only.
func TestParityP44_ColourPick(t *testing.T) {
	s := mustParse(t, userStyle)
	other, _ := s.Match("transportation", line("path"), 8)
	if got, _ := other.Colour(8); got != rgb(0x5f, 0x87, 0xff) {
		t.Errorf("no line colour: %v, want the fill colour", got)
	}
	names, _ := s.Match("place", Attrs{Kind: scene.GeomPoint, Name: "Ohio"}, 8)
	if got, _ := names.Colour(8); got != rgb(0, 255, 0) {
		t.Errorf("no line or fill colour: %v, want the text colour", got)
	}
	plain, _ := s.Match("waterway", line("river"), 8)
	if got, literal := plain.Colour(8); !literal || got != rgb(255, 0, 0) {
		t.Errorf("no colour at all: %v, want upstream's red", got)
	}
	for value, want := range map[string]colour.RGB{`"#abc"`: rgb(0xaa, 0xbb, 0xcc), `"#A1B2C3"`: rgb(0xa1, 0xb2, 0xc3), `"#abcd"`: rgb(255, 0, 0), `"rgba(1,2,3,1)"`: rgb(255, 0, 0), `"blue"`: rgb(255, 0, 0), `7`: rgb(255, 0, 0)} {
		one := mustParse(t, `{"layers":[{"id":"a","type":"line","source-layer":"w","paint":{"line-color":`+value+`}}]}`)
		rule, _ := one.Match("w", line(""), 8)
		if got, _ := rule.Colour(8); got != want {
			t.Errorf("%s: %v, want %v", value, got, want)
		}
	}
}

// TestParityP45_LineWidth: a number or its stops; default 1.
func TestParityP45_LineWidth(t *testing.T) {
	s := mustParse(t, userStyle)
	major, _ := s.Match("transportation", line("motorway"), 8)
	for zoom, want := range map[float64]float64{4: 1, 6: 2, 8: 3, 11: 3} {
		if got := major.Width(zoom); got != want {
			t.Errorf("zoom %v: width %v, want %v; every stop is honoured, with a straight line between them", zoom, got, want)
		}
	}
	plain, _ := s.Match("waterway", line("river"), 8)
	if plain.Width(8) != 1 {
		t.Errorf("no width: %v, want 1", plain.Width(8))
	}
	fixed := mustParse(t, `{"layers":[{"id":"a","type":"line","source-layer":"w","paint":{"line-width":2.5}}]}`)
	rule, _ := fixed.Match("w", line(""), 3)
	if rule.Width(3) != 2.5 {
		t.Errorf("a plain number: %v", rule.Width(3))
	}
	if BuiltIn().rules[0].Width(8) != 1 {
		t.Error("a built-in rule is one dot wide")
	}
}

func TestAttrsOf(t *testing.T) {
	a := AttrsOf(scene.Feature{Kind: scene.GeomLine, Class: "river", Name: textsafe.Clean("Ohio"), Rank: 3, AdminLevel: 2, Maritime: true})
	if a != (Attrs{Kind: scene.GeomLine, Class: "river", Name: "Ohio", Rank: 3, AdminLevel: 2, Maritime: true}) {
		t.Errorf("%+v", a)
	}
}
