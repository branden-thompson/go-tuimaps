// Package style decides what of a tile is drawn, and in which role: the
// library's own style, written against OpenMapTiles (D-24), and a user's
// style in the same JSON format with legacy filters (FR-20). It resolves no
// colour of its own for the built-in style - a rule names a token, and the
// palette and the ground decide the colour at draw time (D-26).
package style

import (
	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// Attrs is what a filter can ask about a feature: what the decoder keeps.
type Attrs struct {
	Kind       scene.GeomKind
	Class      string
	Name       string
	Rank       int32
	AdminLevel uint8
	Maritime   bool
}

// AttrsOf is a kept feature's attributes.
func AttrsOf(f scene.Feature) Attrs {
	return Attrs{Kind: f.Kind, Class: f.Class, Name: f.Name, Rank: f.Rank, AdminLevel: f.AdminLevel, Maritime: f.Maritime}
}

// Kind is how a rule draws.
type Kind uint8

// The kinds of rule. Layers of any other type in a user's style are passed
// over: the ground is the ground token's, not a background layer's.
const (
	Line Kind = iota + 1
	Fill
	Symbol
)

type colourStop struct {
	zoom   float64
	colour colour.RGB
}

type widthStop struct {
	zoom, width float64
}

// Rule is one style layer: which features it takes, and how they are drawn.
type Rule struct {
	ID      string
	Layer   string // the tile layer it reads
	Kind    Kind
	MinZoom float64 // the rule applies from this zoom; zero means no lower bound...
	MaxZoom float64 // ...up to and including this one, as upstream has it (P-27); zero means no upper bound
	// Token is the role a built-in rule draws in. A user's rule has none: it
	// gives literal colours, read with Colour.
	Token    colour.Token
	priority int
	filter   []step
	colours  []colourStop
	widths   []widthStop
}

// Style is an ordered list of rules.
type Style struct {
	rules []Rule
}

func (r *Rule) inZoom(zoom float64) bool {
	if r.MinZoom != 0 && zoom < r.MinZoom {
		return false // a rule from zoom 0 has no lower bound: a small world map is below it
	}
	return r.MaxZoom == 0 || zoom <= r.MaxZoom
}

// accepts reports whether the rule's filter takes the feature.
func (r *Rule) accepts(a Attrs) bool {
	if r == nil {
		return false
	}
	return run(r.filter, a)
}

// Match is the rule a feature is drawn by: the first in style order, for its
// tile layer, whose zooms and filter take it. No match drops the feature;
// there is one rule a feature (P-41).
func (s *Style) Match(layer string, a Attrs, zoom float64) (*Rule, bool) {
	if s == nil || layer == "" {
		return nil, false
	}
	for i := range s.rules {
		r := &s.rules[i]
		if r.Layer == layer && r.inZoom(zoom) && r.accepts(a) {
			return r, true
		}
	}
	return nil, false
}

// Fill is the rule an area is filled by: the first fill rule that takes it.
// The built-in style draws water's edge by one rule and fills it by another.
func (s *Style) Fill(layer string, a Attrs, zoom float64) (*Rule, bool) {
	if s == nil || layer == "" {
		return nil, false
	}
	for i := range s.rules {
		r := &s.rules[i]
		if r.Kind == Fill && r.Layer == layer && r.inZoom(zoom) && r.accepts(a) {
			return r, true
		}
	}
	return nil, false
}

// Priority is how long a role outlasts the others when the basemap must give
// something up: the higher, the later it goes (D-83). It is false for a
// token no rule draws in.
func (s *Style) Priority(t colour.Token) (int, bool) {
	if s == nil || t == 0 {
		return 0, false
	}
	for i := range s.rules {
		if s.rules[i].Token == t && s.rules[i].Kind != Fill {
			return s.rules[i].priority, true
		}
	}
	return 0, false
}

// Colour is a user's rule's literal colour at a zoom: the last stop at or
// below it. It is false for a built-in rule, whose colour is its token's.
func (r *Rule) Colour(zoom float64) (colour.RGB, bool) {
	if r == nil || len(r.colours) == 0 {
		return colour.RGB{}, false
	}
	got := r.colours[0].colour
	for _, stop := range r.colours {
		if stop.zoom <= zoom {
			got = stop.colour
		}
	}
	return got, true
}

// Width is the rule's line width at a zoom: every stop honoured, a straight
// line between them, the ends held; 1 if the style gives none (P-45).
func (r *Rule) Width(zoom float64) float64 {
	if r == nil || len(r.widths) == 0 {
		return 1
	}
	if zoom <= r.widths[0].zoom {
		return r.widths[0].width
	}
	for i := 1; i < len(r.widths); i++ {
		lo, hi := r.widths[i-1], r.widths[i]
		if zoom <= hi.zoom && hi.zoom > lo.zoom {
			// The product is converted before it is added, so that no
			// machine fuses the two into one rounding: this number is
			// rounded to a whole number of dots, and a width that rounds
			// to 2 on one machine and 1 on another draws a different
			// picture from the same data (NFR-6, constants section 6).
			return lo.width + float64((hi.width-lo.width)*(zoom-lo.zoom))/(hi.zoom-lo.zoom)
		}
	}
	return r.widths[len(r.widths)-1].width
}
