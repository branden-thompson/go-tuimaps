package style

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// TestStyleLimits is plan task 08.20 (PL-IS-1): a user's style is untrusted.
func TestStyleLimits(t *testing.T) {
	layer := `{"id":"a","type":"line","source-layer":"w","filter":%s}`
	deep := strings.Repeat(`["all",`, 70) + `["has","name"]` + strings.Repeat(`]`, 70)
	if _, err := Parse([]byte(`{"layers":[` + strings.Replace(layer, "%s", deep, 1) + `]}`)); !isKind(err, fault.OverLimit) {
		t.Errorf("a filter nested 70 deep: %v", err)
	}
	nested := strings.Repeat(`["all",`, 30) + `["has","name"]` + strings.Repeat(`]`, 30)
	s, err := Parse([]byte(`{"layers":[` + strings.Replace(layer, "%s", nested, 1) + `]}`))
	if err != nil {
		t.Fatalf("a filter nested 30 deep: %v", err)
	}
	if _, ok := s.Match("w", Attrs{Kind: scene.GeomLine, Name: "Ohio"}, 5); !ok {
		t.Error("a deep filter that holds did not match")
	}
	if _, ok := s.Match("w", Attrs{Kind: scene.GeomLine}, 5); ok {
		t.Error("a deep filter that does not hold matched")
	}
	// A group wider than the evaluator's fixed room still works out.
	var wide []string
	for range 100 {
		wide = append(wide, `["has","brunnel"]`)
	}
	any100 := `["any",` + strings.Join(wide, ",") + `,["has","name"]]`
	s, err = Parse([]byte(`{"layers":[` + strings.Replace(layer, "%s", any100, 1) + `]}`))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := s.Match("w", Attrs{Kind: scene.GeomLine, Name: "Ohio"}, 5); !ok {
		t.Error("a group of 101 parts, the last of which holds, did not match")
	}
	// A chain of refs longer than any real style's is a ring as far as the
	// parser will follow it.
	var chain []string
	for i := range 12 {
		chain = append(chain, `{"id":"l`+string(rune('a'+i))+`","ref":"l`+string(rune('a'+i+1))+`"}`)
	}
	chain = append(chain, `{"id":"lm","type":"line","source-layer":"w"}`)
	if _, err := Parse([]byte(`{"layers":[` + strings.Join(chain, ",") + `]}`)); !isKind(err, fault.MalformedStyle) {
		t.Errorf("a chain of twelve refs: %v", err)
	}
	// Zooms and widths outside any real map are read as none.
	odd := mustParse(t, `{"layers":[{"id":"a","type":"line","source-layer":"w","minzoom":-4,"maxzoom":1e9,"paint":{"line-width":{"stops":[[1,1e308],[2,-3],[3,"x"],[4,2]]}}}]}`)
	rule, ok := odd.Match("w", Attrs{Kind: scene.GeomLine}, 25)
	if !ok || rule.MinZoom != 0 || rule.MaxZoom != 0 {
		t.Errorf("%+v, %v", rule, ok)
	}
	if w := rule.Width(9); w != 2 {
		t.Errorf("width %v; only the stop with a real width is kept", w)
	}
}

func FuzzStyle(f *testing.F) {
	f.Add([]byte(userStyle))
	f.Add([]byte(`{"layers":[{"id":"a","ref":"a"}]}`))
	f.Add([]byte(`{"constants":{"@a":"@a"},"layers":[{"id":"a","type":"fill","source-layer":"w","filter":["none",["in","class","@a",1,2.5,true,null]],"paint":{"fill-color":{"stops":[[1,"@a"],[2]]}}}]}`))
	f.Add([]byte(`{"layers":[{"id":"a","type":"line","source-layer":"w","filter":["all",["any"],["none",[">","rank","x"]],[]]}]}`))
	attrs := []Attrs{
		{},
		{Kind: scene.GeomLine, Class: "motorway"},
		{Kind: scene.GeomPolygon, Class: "lake", Name: "Erie", Rank: 2},
		{Kind: scene.GeomLine, AdminLevel: 2, Maritime: true},
		{Kind: scene.GeomPoint, Class: "city", Name: "Chicago", Rank: 3},
	}
	f.Fuzz(func(t *testing.T, body []byte) {
		s, err := Parse(body)
		if err != nil {
			var own *fault.Error
			if !errors.As(err, &own) {
				t.Fatalf("a foreign error: %v", err)
			}
			if k := own.Kind(); k != fault.MalformedStyle && k != fault.OverLimit {
				t.Fatalf("kind %v", k)
			}
			return
		}
		for i := range s.rules {
			r := &s.rules[i]
			if r.Kind < Line || r.Kind > Symbol {
				t.Fatalf("rule %q has kind %d", r.ID, r.Kind)
			}
			if !(r.MinZoom >= 0 && r.MinZoom <= 30 && r.MaxZoom >= 0 && r.MaxZoom <= 30) {
				t.Fatalf("rule %q has zooms %v to %v", r.ID, r.MinZoom, r.MaxZoom)
			}
			for _, zoom := range []float64{-5, 0, 7.5, 14, 99} {
				if w := r.Width(zoom); math.IsNaN(w) || w <= 0 || w > 30 {
					t.Fatalf("rule %q is %v wide at zoom %v", r.ID, w, zoom)
				}
				if _, literal := r.Colour(zoom); !literal {
					t.Fatalf("a user's rule %q has no colour of its own", r.ID)
				}
			}
			for _, a := range attrs {
				r.accepts(a)
				s.Match(r.Layer, a, 8)
				s.Fill(r.Layer, a, 8)
			}
		}
	})
}
