package colour

import "testing"

// TestAlertPreset is plan task 08.10 (FR-16, D-88, RS-26): five severities,
// an outline and a tint each, two sets by ground. PLAN drew two of the five,
// on a dark ground; the rest are designed here, and are HUM LEAD's to look at
// before any reference frame is frozen (task 08.23).
func TestAlertPreset(t *testing.T) {
	var none Palette
	severities := []string{"extreme", "severe", "moderate", "minor", "unknown"}
	for _, kind := range []GroundKind{Dark, Light} {
		for _, depth := range []Depth{Truecolor, Colours256} {
			show := func(c RGB) RGB {
				if depth == Colours256 {
					_, c = To256(c)
				}
				return c
			}
			ground, _ := none.Resolve(Ground, kind)
			water, _ := none.Resolve(WaterFill, kind)
			ground, water = show(ground), show(water)
			var outlines, tints []RGB
			for i := range severities {
				o, okO := none.ResolveAt(AlertExtremeOutline+Token(2*i), kind, depth)
				tint, okT := none.ResolveAt(AlertExtremeTint+Token(2*i), kind, depth)
				if !okO || !okT {
					t.Fatalf("%v: %s has no default", kind, severities[i])
				}
				outlines, tints = append(outlines, show(o)), append(tints, show(tint))
			}
			for i, name := range severities {
				if got := Contrast(outlines[i], ground); got < LineContrast {
					t.Errorf("%v %v: the %s outline is %.2f:1 on the ground", kind, depth, name, got)
				}
				if got := Contrast(outlines[i], tints[i]); got < LineContrast {
					t.Errorf("%v %v: the %s outline is %.2f:1 on its own tint; a tint is never the only edge", kind, depth, name, got)
				}
				if text := Foreground(tints[i], tints[i], TextContrast); Contrast(text, tints[i]) < TextContrast {
					t.Errorf("%v %v: no text reads on the %s tint", kind, depth, name)
				}
				if d := Difference(tints[i], water); d < threshold {
					t.Errorf("%v %v: the %s tint is %.1f from water; an alert over the sea must show", kind, depth, name, d)
				}
			}
			for name, ramp := range map[string][]RGB{"outlines": outlines, "tints": tints} {
				for _, f := range Check(ramp, RampCheck{Ground: ground, Depth: depth, Midpoint: -1}) {
					if f.Rule == VisionSafe || f.Rule == Distinct {
						t.Errorf("%v %v: %s %+v", kind, depth, name, f)
					}
				}
			}
		}
	}
	// The two pairs PLAN's specimens drew, which HUM LEAD has seen, are kept.
	if o, _ := none.Resolve(AlertSevereOutline, Dark); o != (RGB{255, 128, 80}) {
		t.Errorf("the severe outline on a dark ground: %v", o)
	}
	if tint, _ := none.Resolve(AlertModerateTint, Dark); tint != (RGB{104, 78, 18}) {
		t.Errorf("the moderate tint on a dark ground: %v", tint)
	}
}
