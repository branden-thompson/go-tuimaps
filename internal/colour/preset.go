package colour

// Preset is one of the library's own types of overlay colouring (D-69).
type Preset uint8

// The presets built so far.
const (
	Temperature Preset = iota + 1
	Radar
)

// Unit is a unit of temperature.
type Unit uint8

// The units. The breaks are defined once, in Celsius (D-62).
const (
	Celsius Unit = iota
	Fahrenheit
)

// TemperatureBreaks are the temperature preset's sixteen breaks: every 5 C
// from -30 to +45, with the pale classes either side of freezing. In
// Fahrenheit they are these converted exactly, which multiples of five are.
func TemperatureBreaks(u Unit) []float64 {
	out := make([]float64, 0, 16)
	for c := -30.0; c <= 45; c += 5 {
		if u == Fahrenheit {
			out = append(out, c*9/5+32)
			continue
		}
		out = append(out, c)
	}
	return out
}

// RadarFloors are the radar preset's six class floors, in dBZ.
func RadarFloors() []float64 {
	return []float64{10, 20, 30, 40, 50, 60}
}

// Midpoint is the class a preset's ramp is lightest at, or -1 for a ramp that
// runs one way. The temperature scale's two pale classes lie either side of
// freezing, and which of them is the lighter differs: the one just below, in
// truecolor on a dark ground; the one just above, on a light ground (D-91
// changed those bands) and in the 256-colour palette.
func Midpoint(p Preset, ground GroundKind, depth Depth) int {
	if p != Temperature {
		return -1
	}
	if ground == Light || depth == Colours256 {
		return 7
	}
	return 6
}

// Ramp is a preset's colours for a kind of ground and a depth. Each depth
// has the library's own ramp: the 256-colour one is chosen from that palette
// and checked there, never a conversion of the truecolor one (S10-1). With
// no colour, or sixteen, there is none: the no-colour form is drawn (D-59).
func Ramp(p Preset, ground GroundKind, depth Depth) ([]RGB, bool) {
	if depth != Truecolor && depth != Colours256 {
		return nil, false
	}
	switch p {
	case Temperature:
		return temperatureRamp(ground, depth), true
	case Radar:
		return radarRamp(ground, depth), true
	}
	return nil, false
}

// temperatureRamp is specimen 21's scale: seventeen classes, coldest first.
func temperatureRamp(ground GroundKind, depth Depth) []RGB {
	if depth == Colours256 {
		r := []RGB{{0, 0, 95}, {0, 95, 175}, {95, 95, 135}, {0, 135, 215}, {135, 135, 175}, {135, 215, 255}, {215, 255, 255}, {255, 255, 215}, {215, 215, 135}, {255, 175, 95}, {215, 175, 135}, {255, 135, 0}, {255, 95, 95}, {255, 0, 0}, {175, 0, 0}, {135, 0, 0}, {95, 0, 0}}
		if ground == Light {
			r[5], r[6], r[7] = RGB{95, 175, 215}, RGB{175, 215, 255}, RGB{215, 255, 175}
		}
		return r
	}
	r := []RGB{{34, 0, 68}, {0, 34, 119}, {34, 85, 153}, {51, 136, 187}, {136, 170, 204}, {153, 221, 238}, {221, 255, 255}, {238, 255, 187}, {240, 232, 144}, {255, 204, 102}, {255, 170, 119}, {255, 136, 68}, {255, 102, 85}, {221, 51, 34}, {170, 0, 17}, {148, 29, 46}, {85, 0, 17}}
	if ground == Light {
		r[5], r[6], r[7] = RGB{120, 220, 250}, RGB{180, 220, 230}, RGB{210, 250, 180}
	}
	return r
}

// radarRamp is specimen 22's: six classes, lightest rain first. Lighter means
// heavier on a dark ground; darker means heavier on a light one.
func radarRamp(ground GroundKind, depth Depth) []RGB {
	switch {
	case ground == Light && depth == Colours256:
		return []RGB{{175, 215, 175}, {135, 175, 175}, {0, 175, 95}, {215, 95, 0}, {175, 0, 0}, {95, 0, 95}}
	case ground == Light:
		return []RGB{{150, 205, 215}, {60, 160, 175}, {110, 150, 60}, {200, 70, 0}, {150, 10, 30}, {80, 0, 70}}
	case depth == Colours256:
		return []RGB{{95, 95, 175}, {0, 135, 135}, {135, 175, 95}, {255, 175, 0}, {255, 215, 135}, {255, 255, 255}}
	}
	return []RGB{{0, 51, 102}, {34, 119, 136}, {119, 170, 85}, {221, 170, 34}, {255, 221, 119}, {255, 255, 238}}
}

// alertColours are the alert preset's outlines and tints, most severe first:
// extreme, severe, moderate, minor, unknown. On a dark ground the outlines
// are bright and the tints dark; on a light ground it is the other way
// round, because both of the dark set's first outlines fail 3:1 there
// (RS-26). Every outline reads on its ground and on its own tint, and every
// pair of outlines, every pair of tints, and each against its ground and
// against water is at least 10 apart under every kind of vision (D-88).
//
// The 256-colour sets are the library's own, chosen from that palette and
// checked there: the truecolor sets converted entry by entry fail, on both
// grounds (S10-1 again).
func alertColours(ground GroundKind, depth Depth) (outlines, tints [5]RGB) {
	switch {
	case ground == Light && depth == Colours256:
		return [5]RGB{{95, 0, 95}, {175, 0, 0}, {175, 95, 0}, {0, 95, 95}, {48, 48, 48}},
			[5]RGB{{255, 95, 215}, {255, 175, 175}, {255, 215, 135}, {135, 175, 255}, {158, 158, 158}}
	case ground == Light:
		return [5]RGB{{107, 10, 109}, {168, 36, 8}, {121, 91, 66}, {30, 91, 94}, {49, 49, 50}},
			[5]RGB{{224, 137, 202}, {255, 190, 165}, {255, 228, 124}, {142, 205, 255}, {184, 187, 184}}
	case depth == Colours256:
		return [5]RGB{{255, 95, 215}, {255, 135, 95}, {255, 215, 95}, {95, 255, 175}, {228, 228, 228}},
			[5]RGB{{95, 0, 95}, {95, 0, 0}, {135, 95, 0}, {0, 95, 175}, {98, 98, 98}}
	}
	return [5]RGB{{255, 110, 200}, {255, 128, 80}, {255, 214, 90}, {87, 239, 176}, {226, 230, 239}},
		[5]RGB{{109, 31, 112}, {112, 36, 28}, {104, 78, 18}, {33, 86, 97}, {107, 105, 101}}
}

// rampDefault is the library's colour for a ramp token.
func rampDefault(t Token, ground GroundKind, depth Depth) (RGB, bool) {
	switch {
	case t >= AlertExtremeOutline && t <= AlertUnknownTint:
		outlines, tints := alertColours(ground, depth)
		i := int(t - AlertExtremeOutline)
		if i%2 == 0 {
			return outlines[i/2], true
		}
		return tints[i/2], true
	case t >= Radar1 && t <= Radar6:
		ramp, ok := Ramp(Radar, ground, depth)
		return pick(ramp, int(t-Radar1), ok)
	case t >= Temperature1 && t <= Temperature17:
		ramp, ok := Ramp(Temperature, ground, depth)
		return pick(ramp, int(t-Temperature1), ok)
	}
	return RGB{}, false
}

func pick(ramp []RGB, i int, ok bool) (RGB, bool) {
	if !ok || i < 0 || i >= len(ramp) {
		return RGB{}, false
	}
	return ramp[i], true
}
