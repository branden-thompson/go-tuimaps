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
	if t < Ground || t > TrackLabel {
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
	}
	return "track.label"
}

// Tokens lists every token, in the documented order.
func Tokens() []Token {
	out := make([]Token, 0, TrackLabel)
	for t := Ground; t <= TrackLabel; t++ {
		out = append(out, t)
	}
	return out
}

// ParseToken finds a token by its name.
func ParseToken(name string) (Token, bool) {
	if name == "" || len(name) > 32 {
		return 0, false
	}
	for t := Ground; t <= TrackLabel; t++ {
		if t.Name() == name {
			return t, true
		}
	}
	return 0, false
}
