package style

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

// The layers that can be switched. Water and the coast are one: a map whose
// water is off is a map with no shore at all.
const (
	RoadLayer Layer = iota + 1
	RailLayer
	ParkLayer
	BorderLayer
	RiverLayer
	WaterLayer
	LabelLayer
)

// Switches are the layers a host has switched off.
type Switches uint16

// Off is the set of layers named.
func Off(layers ...Layer) Switches {
	var s Switches
	for _, l := range layers {
		if l >= RoadLayer && l <= LabelLayer {
			s |= 1 << l
		}
	}
	return s
}

// Has reports whether a layer is switched off.
func (s Switches) Has(l Layer) bool {
	return l >= RoadLayer && l <= LabelLayer && s&(1<<l) != 0
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
	if r == nil {
		return 0
	}
	if r.Kind == Symbol {
		return LabelLayer
	}
	switch r.Layer {
	case "transportation":
		if r.ID == "rail" {
			return RailLayer
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
	return r.priority == 0 || r.priority >= p.floor
}

// Weight is a rule's line width under this profile: one dot for the roads
// that are left when something is drawn over them, and the rule's own width
// otherwise.
func (p Profile) Weight(r *Rule, zoom float64) float64 {
	if r == nil {
		return 0
	}
	if p.thin && r.Group() == RoadLayer {
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
