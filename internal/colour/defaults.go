package colour

// darkDefault is the library's colour for a token on a dark ground: the
// dark style (FR-20). The ramps - alerts, radar, temperature, a host's own
// type - are not here: a preset gives those.
func darkDefault(t Token) (RGB, bool) {
	if t > Runway {
		return darkFurniture(t)
	}
	switch t {
	case Ground:
		return RGB{16, 22, 28}, true
	case WaterFill:
		return RGB{24, 44, 72}, true
	case WaterLine:
		return RGB{120, 160, 200}, true
	case Coast:
		return RGB{150, 190, 225}, true
	case BorderCountry:
		return RGB{255, 255, 255}, true
	case BorderRegion, RoadMinor:
		return RGB{190, 190, 190}, true
	case River:
		return RGB{170, 215, 255}, true
	case RoadMajor:
		return RGB{225, 225, 225}, true
	case Rail:
		return RGB{200, 180, 160}, true
	case Park:
		return RGB{90, 140, 90}, true
	case Runway:
		return RGB{210, 210, 230}, true
	}
	return RGB{}, false
}

// darkFurniture is the dark style's text, furniture, places and tracks.
func darkFurniture(t Token) (RGB, bool) {
	switch t {
	case LabelPlace, Focus, Marker, MarkerLabel, TrackLabel:
		return RGB{255, 255, 255}, true
	case LabelWater:
		return RGB{170, 215, 255}, true
	case LabelRegion, Scale:
		return RGB{200, 200, 200}, true
	case Credit:
		return RGB{160, 160, 160}, true
	case Notice, Track:
		return RGB{255, 214, 90}, true
	case Stale:
		return RGB{255, 170, 80}, true
	}
	return RGB{}, false
}

// lightDefault is the same for a light ground: the bright style.
func lightDefault(t Token) (RGB, bool) {
	if t > Runway {
		return lightFurniture(t)
	}
	switch t {
	case Ground:
		return RGB{245, 245, 240}, true
	case WaterFill:
		return RGB{196, 220, 240}, true
	case WaterLine:
		return RGB{60, 110, 170}, true
	case Coast:
		return RGB{30, 80, 140}, true
	case BorderCountry:
		return RGB{20, 20, 20}, true
	case BorderRegion, RoadMajor:
		return RGB{95, 95, 95}, true
	case RoadMinor:
		return RGB{140, 140, 140}, true
	case River:
		return RGB{40, 100, 170}, true
	case Rail:
		return RGB{110, 90, 70}, true
	case Park:
		return RGB{70, 120, 70}, true
	case Runway:
		return RGB{80, 80, 100}, true
	}
	return RGB{}, false
}

// lightFurniture is the bright style's text, furniture, places and tracks.
func lightFurniture(t Token) (RGB, bool) {
	switch t {
	case LabelPlace, Focus, Marker, MarkerLabel, TrackLabel:
		return RGB{}, true
	case LabelWater:
		return RGB{30, 80, 140}, true
	case LabelRegion:
		return RGB{70, 70, 70}, true
	case Scale:
		return RGB{60, 60, 60}, true
	case Credit:
		return RGB{90, 90, 90}, true
	case Notice:
		return RGB{120, 80, 0}, true
	case Stale:
		return RGB{150, 70, 0}, true
	case Track:
		return RGB{120, 60, 160}, true
	}
	return RGB{}, false
}
