package colour

// SixteenReference is the table that checks at 16 colours are made against:
// the colours a common terminal gives the sixteen entries. At this depth the
// terminal owns the colours and cannot be asked, so every figure against
// this table is indicative only (FR-20, S23-3).
func SixteenReference() [16]RGB {
	return [16]RGB{{0, 0, 0}, {205, 0, 0}, {0, 205, 0}, {205, 205, 0}, {0, 0, 238}, {205, 0, 205}, {0, 205, 205}, {229, 229, 229},
		{127, 127, 127}, {255, 0, 0}, {0, 255, 0}, {255, 255, 0}, {92, 92, 255}, {255, 0, 255}, {0, 255, 255}, {255, 255, 255}}
}

// ToSixteen is the entry of the reference table nearest a colour, and the
// colour the table gives it: how a host's own colour is drawn at this depth.
func ToSixteen(c RGB) (uint8, RGB) {
	table := SixteenReference()
	best := 0
	for i := range table {
		if squared(c, table[i]) < squared(c, table[best]) {
			best = i
		}
	}
	return uint8(best), table[best]
}

// Sixteen is a token's entry at 16 colours: the basemap's own small palette,
// chosen from the sixteen by hand, because converting the style's colours
// turns every pale blue and grey into white (S23-1, D-79). A token that
// belongs to a ramp has none: what needs a ramp is drawn in its no-colour
// form at this depth (D-59). Roads and borders differ in hue, not only in
// brightness (PL-AX-6).
func Sixteen(t Token, ground GroundKind) (uint8, bool) {
	if t < Ground || t > TrackLabel {
		return 0, false
	}
	if t >= AlertExtremeOutline && t <= High {
		return 0, false
	}
	if ground == Light {
		return sixteenLight(t), true
	}
	return sixteenDark(t), true
}

func sixteenDark(t Token) uint8 {
	switch t {
	case Ground:
		return 0
	case WaterFill:
		return 4
	case WaterLine:
		return 6
	case Coast, River, LabelWater:
		return 14
	case RoadMajor:
		return 3
	case RoadMinor, Credit:
		return 8
	case Rail:
		return 5
	case Park:
		return 2
	case BorderRegion, Runway, LabelRegion, Scale:
		return 7
	case Notice, Marker:
		return 11
	case Stale:
		return 9
	case Track:
		return 13
	}
	return 15 // country borders, place names, the focus, a marker's label, a track's
}

func sixteenLight(t Token) uint8 {
	switch t {
	case Ground:
		return 15
	case WaterFill:
		return 14
	case WaterLine, Coast, River, LabelWater:
		return 4
	case RoadMajor, Track:
		return 5
	case Rail, Notice, Stale, Marker:
		return 1
	case BorderRegion, RoadMinor, Park, Runway, LabelRegion, Credit:
		return 8
	}
	return 0 // country borders, place names, the scale, the focus, labels
}

// ChooseDepth is the depth a map draws at. A hint from the host wins: the
// host knows where its output goes. With no hint, a non-empty NO_COLOR in
// the environment selects no colour (NFR-15); otherwise truecolor. A hint
// that is no depth gives no colour, the form that is safe anywhere.
func ChooseDepth(hint Depth, hinted bool, getenv func(string) string) Depth {
	if hinted && hint > NoColour {
		return NoColour
	}
	if hinted {
		return hint
	}
	if getenv != nil && getenv("NO_COLOR") != "" {
		return NoColour
	}
	return Truecolor
}
