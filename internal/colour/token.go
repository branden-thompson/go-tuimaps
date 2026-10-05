// Package colour is how a role becomes a cell's colour: the semantic tokens
// (D-63), a host's palette over the library's defaults, the ground in effect
// (D-64), the contrast formula, and the foreground rule (FR-16, D-77). It
// draws nothing and knows nothing of tiles.
package colour

import "strconv"

// RGB is a colour in 8-bit sRGB.
type RGB struct{ R, G, B uint8 }

// Token is a named role: a place in the map's look, never a colour. The
// names are stable; adding one is a minor version.
type Token uint8

// The tokens, in the order of the constants document, section 4.
const (
	Ground Token = iota + 1
	WaterFill
	WaterLine
	Coast
	BorderCountry
	BorderRegion
	River
	RoadMajor
	RoadMinor
	Rail
	Park
	Runway
	LabelPlace
	LabelWater
	LabelRegion
	Scale
	Credit
	Notice
	Stale
	Focus
	Marker
	MarkerLabel
	AlertExtremeOutline
	AlertExtremeTint
	AlertSevereOutline
	AlertSevereTint
	AlertModerateOutline
	AlertModerateTint
	AlertMinorOutline
	AlertMinorTint
	AlertUnknownOutline
	AlertUnknownTint
	Radar1 // radar.1 to radar.6 follow in order
)

// The tokens after the first radar class.
const (
	Radar6        = Radar1 + 5
	Temperature1  = Radar6 + 1 // temperature.1 to temperature.17 follow in order
	Temperature17 = Temperature1 + 16
	Low           = Temperature17 + 1
	Middle        = Low + 1
	High          = Middle + 1
	Track         = High + 1
	TrackLabel    = Track + 1
	// Wind1 to Wind6 are the wind preset's classes, calmest first (FR-8,
	// L-16.3; from watchpost D-109): after the rest, so no token already named moves.
	Wind1 = TrackLabel + 1
	Wind6 = Wind1 + 5
	// Fire and FireFaint are a fire's (L-18.1; from watchpost D-121): its perimeter, its
	// incident and a strong hotspot; a weaker hotspot. Never an alert.
	Fire      = Wind6 + 1
	FireFaint = Fire + 1
	// QuakeHour, QuakeDay and QuakeOlder are a quake's ring by its age, as
	// USGS colours it (L-19.3; from watchpost D-123): the past hour, the past day, older.
	QuakeHour  = FireFaint + 1
	QuakeDay   = QuakeHour + 1
	QuakeOlder = QuakeDay + 1
	// Wave1 to Wave6 are the wave preset's classes, calmest first (L-20).
	Wave1 = QuakeOlder + 1
	Wave6 = Wave1 + 5
	// Buoy and Tide are the sea's stations (L-21, watchpost D-127, D-128): a
	// buoy's marker and words, a tide station's. Never an alert.
	Buoy = Wave6 + 1
	Tide = Buoy + 1
	// UV1 to UV5 are the UV preset's classes, AQI1 to AQI6 the US AQI's,
	// lowest first (L-25, watchpost D-137 to D-140).
	UV1  = Tide + 1
	UV5  = UV1 + 4
	AQI1 = UV5 + 1
	AQI6 = AQI1 + 5
	// QPF1 to QPF7 are the rain totals' classes, lightest first (L-26,
	// watchpost D-184).
	QPF1 = AQI6 + 1
	QPF7 = QPF1 + 6
	// lastToken is the last token there is.
	lastToken = QPF7
)

// fixedNames are the names of the tokens before the numbered ramps.
func fixedNames() []string {
	return []string{"ground", "water.fill", "water.line", "coast", "border.country", "border.region", "river",
		"road.major", "road.minor", "rail", "park", "runway", "label.place", "label.water", "label.region",
		"scale", "credit", "notice", "stale", "focus", "marker", "marker.label",
		"alert.extreme.outline", "alert.extreme.tint", "alert.severe.outline", "alert.severe.tint",
		"alert.moderate.outline", "alert.moderate.tint", "alert.minor.outline", "alert.minor.tint",
		"alert.unknown.outline", "alert.unknown.tint"}
}

// Name is the token's stable name, or nothing for a value that is no token.
func (t Token) Name() string {
	if t < Ground || t > lastToken {
		return ""
	}
	switch {
	case t < Radar1:
		return fixedNames()[t-1]
	case t <= Radar6:
		return "radar." + strconv.Itoa(int(t-Radar1)+1)
	case t <= Temperature17:
		return "temperature." + strconv.Itoa(int(t-Temperature1)+1)
	case t == Low:
		return "low"
	case t == Middle:
		return "middle"
	case t == High:
		return "high"
	case t == Track:
		return "track"
	case t == TrackLabel:
		return "track.label"
	case t == Fire:
		return "fire"
	case t == FireFaint:
		return "fire.faint"
	case t == QuakeHour:
		return "quake.hour"
	case t == QuakeDay:
		return "quake.day"
	case t == QuakeOlder:
		return "quake.older"
	case t >= Wave1 && t <= Wave6:
		return "wave." + strconv.Itoa(int(t-Wave1)+1)
	case t == Buoy:
		return "buoy"
	case t == Tide:
		return "tide"
	case t >= UV1 && t <= UV5:
		return "uv." + strconv.Itoa(int(t-UV1)+1)
	case t >= AQI1 && t <= AQI6:
		return "aqi." + strconv.Itoa(int(t-AQI1)+1)
	case t >= QPF1 && t <= QPF7:
		return "qpf." + strconv.Itoa(int(t-QPF1)+1)
	}
	return "wind." + strconv.Itoa(int(t-Wind1)+1)
}

// ScaleClass reports whether a token is one of a preset's classes: drawn in
// the scale's colour, never a word's (L-25).
func ScaleClass(t Token) bool {
	return (t >= Radar1 && t <= High) || (t >= Wind1 && t <= Wind6) || (t >= Wave1 && t <= Wave6) || (t >= UV1 && t <= QPF7)
}

// Tokens lists every token, in the documented order.
func Tokens() []Token {
	out := make([]Token, 0, lastToken)
	for t := Ground; t <= lastToken; t++ {
		out = append(out, t)
	}
	return out
}

// ParseToken finds a token by its name.
func ParseToken(name string) (Token, bool) {
	if name == "" || len(name) > 32 {
		return 0, false
	}
	for t := Ground; t <= lastToken; t++ {
		if t.Name() == name {
			return t, true
		}
	}
	return 0, false
}
