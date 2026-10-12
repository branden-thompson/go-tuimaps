package colour

// NightShade is how far the night side's ground and sea are shaded toward
// the night's colour, on each kind of ground (L-4.4): the most that still
// leaves every preset's ramp passing its checks against the shaded ground,
// searched over the ramps (watchpost D-174's UAT-2 judges how it reads).
// A field's own colours are never shaded: it stays readable at night.
func NightShade(g GroundKind) float64 {
	if g == Light {
		return 0.6
	}
	return 0.8
}
