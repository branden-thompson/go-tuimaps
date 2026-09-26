package style

import "strconv"

// Load is what is drawn over the basemap, which decides how much of itself
// the basemap gives up (FR-19). This is not a navigation tool (D-83): roads
// go first, and coast, water and borders go last.
type Load uint8

// The loads a basemap draws under.
const (
	Bare    Load = iota // nothing on top, or feature overlays only
	Covered             // a scalar field or an image covers the map
	Dense               // wind, later: the basemap gives up its roads
)

// Layer is a basemap layer a host can switch off and on (FR-36).
type Layer uint8

// The layers that can be switched. WaterLayer is the lakes and inland water:
// the sea and its coast are never switched, since a map without them has no
// shore at all (watchpost UAT-1 U1-39).
const (
	RoadLayer Layer = iota + 1
	RailLayer
	ParkLayer
	BorderLayer
	RiverLayer
	WaterLayer
	LabelLayer
	// MinorRoadLayer is the minor roads, switched apart from the major ones
	// (v0.2.0 D-82): RoadLayer is the motorways, trunks and primaries a
	// weather map keeps for bearings.
	MinorRoadLayer
)

// Switches are the layers a host has switched off.
type Switches uint16

// Off is the set of layers named.
func Off(layers ...Layer) Switches {
	var s Switches
	for _, l := range layers {
		if l >= RoadLayer && l <= MinorRoadLayer {
			s |= 1 << l
		}
	}
	return s
}

// Has reports whether a layer is switched off.
func (s Switches) Has(l Layer) bool {
	return l >= RoadLayer && l <= MinorRoadLayer && s&(1<<l) != 0
}

// A map smaller than this gives up one more thing than its load alone asks
// for: there is less room for everything.
const (
	smallCols = 80
	smallRows = 20
)

// The label budgets: no bound on a full map, fewer under an overlay, fewer
// again when the basemap is sparse or the map is small.
const (
	coveredLabels = 12
	denseLabels   = 8
	smallLabels   = 16
)

// Profile is what a frame draws of the basemap, and how heavily: a floor on
// the priorities drawn, whether the heaviest roads left are drawn thin, a
// label budget, and the host's own switches, which are applied last.
type Profile struct {
	floor  int
	thin   bool
	labels int
	off    Switches
	detail Detail // the host's level; zero reads as Full
}

// NewProfile chooses the profile for one frame: by what is drawn on top, by
// the map's size, and by the layers the host has switched off (FR-19, FR-36).
func NewProfile(load Load, cols, rows int, off Switches) Profile {
	p := Profile{off: off, floor: 1}
	switch load {
	case Covered:
		p.floor, p.thin, p.labels = 2, true, coveredLabels
	case Dense:
		p.floor, p.labels = 3, denseLabels
	}
	if cols < smallCols || rows < smallRows {
		p.floor++
		if p.labels == 0 {
			p.labels = smallLabels
		} else {
			p.labels /= 2
		}
	}
	return p
}

// Group is the layer a rule belongs to, which is what a host switches: a
// rule that draws names belongs to the labels whatever it reads, and every
// other rule belongs to the layer of the tile it reads.
func (r *Rule) Group() Layer {
	if r == nil || r.always {
		return 0 // no switch takes it
	}
	if r.Kind == Symbol {
		return LabelLayer
	}
	switch r.Layer {
	case "transportation":
		switch r.ID {
		case "rail":
			return RailLayer
		case "road-minor":
			return MinorRoadLayer
		}
		return RoadLayer
	case "aeroway":
		return RailLayer
	case "park", "landcover":
		return ParkLayer
	case "boundary":
		return BorderLayer
	case "waterway":
		return RiverLayer
	case "water":
		return WaterLayer
	}
	return 0
}

// Draws reports whether a rule is drawn under this profile. A rule of a
// user's own style has no priority and is always drawn, since only the
// built-in style's roles are ranked; a host's switches apply to both.
func (p Profile) Draws(r *Rule) bool {
	if r == nil {
		return false
	}
	if p.off.Has(r.Group()) {
		return false
	}
	if r.detail > p.Detail() {
		return false // the host's level does not reach this rule (L-14)
	}
	return r.priority == 0 || r.priority >= p.floor
}

// Weight is a rule's line width under this profile: one dot for the roads
// that are left when something is drawn over them, and the rule's own width
// otherwise.
func (p Profile) Weight(r *Rule, zoom float64) float64 {
	if r == nil {
		return 0
	}
	if g := r.Group(); p.thin && (g == RoadLayer || g == MinorRoadLayer) {
		return 1
	}
	return r.Width(zoom)
}

// Labels is how many names a frame may place, or 0 for no bound.
func (p Profile) Labels() int {
	return p.labels
}

// Floor is the lowest priority drawn: what the basemap has given up.
func (p Profile) Floor() int {
	return p.floor
}

// Detail is how much of the basemap a host asks for, by purpose (v0.2.0
// D-82, L-14): each level draws what the one below does and more. Only the
// built-in style's rules are ranked; a host's own style draws whole.
type Detail uint8

// The levels.
const (
	DetailEssential Detail = iota + 1 // coast, water, borders
	DetailWeather                     // and rivers, place names, the major roads
	DetailStandard                    // and rail, parks
	DetailFull                        // and the minor roads, runways: the picture as it was
)

// Valid reports whether a detail is one of the four.
func (d Detail) Valid() bool { return d >= DetailEssential && d <= DetailFull }

// String names the level.
func (d Detail) String() string {
	switch d {
	case DetailEssential:
		return "essential"
	case DetailWeather:
		return "weather"
	case DetailStandard:
		return "standard"
	case DetailFull:
		return "full"
	}
	return "detail(" + strconv.Itoa(int(d)) + ")"
}

// WithDetail is the profile at a host's level.
func (p Profile) WithDetail(d Detail) Profile {
	p.detail = d
	return p
}

// Detail is the profile's level: Full when none was set.
func (p Profile) Detail() Detail {
	if !p.detail.Valid() {
		return DetailFull
	}
	return p.detail
}

// ruleDetail is the least level each built-in rule is drawn at (D-82). A rule
// not named here - a host's own - is drawn at every level.
var ruleDetail = map[string]Detail{
	"coast": DetailEssential, "sea": DetailEssential, "water-edge": DetailEssential, "water": DetailEssential,
	"border-country": DetailEssential, "border-region": DetailEssential,
	"river": DetailWeather, "label-region": DetailWeather, "label-place": DetailWeather,
	"label-water": DetailWeather, "road-major": DetailWeather,
	"park": DetailStandard, "rail": DetailStandard,
	"road-minor": DetailFull, "runway": DetailFull,
}
