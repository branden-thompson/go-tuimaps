package colour

import "testing"

// TestTheZoneTokensHaveDefaults is L-8.1: a host's zone is an outline and a
// tint with the library's own colours on both grounds and at sixteen; the
// outline reads at 3:1 on its ground and on its own tint; neither is any
// preset's class, nor an alert's, and each stays at least 10 apart from every
// alert outline and tint under every kind of vision (D-88's rule), at
// truecolor and at 256 colours.
func TestTheZoneTokensHaveDefaults(t *testing.T) {
	for _, c := range []struct {
		tok  Token
		name string
	}{{Zone, "zone"}, {ZoneTint, "zone.tint"}} {
		if c.tok.Name() != c.name {
			t.Errorf("%d is named %q; want %q", c.tok, c.tok.Name(), c.name)
		}
		if back, ok := ParseToken(c.name); !ok || back != c.tok {
			t.Errorf("%q does not parse back to its token", c.name)
		}
		if ScaleClass(c.tok) || (c.tok >= AlertExtremeOutline && c.tok <= AlertUnknownTint) {
			t.Errorf("%s is a preset's class or an alert's", c.name)
		}
		for _, g := range []GroundKind{Dark, Light} {
			if _, ok := (Palette{}).Resolve(c.tok, g); !ok {
				t.Errorf("%s has no default on the %v ground", c.name, g)
			}
			if _, ok := Sixteen(c.tok, g); !ok {
				t.Errorf("%s has no sixteen-colour entry on the %v ground", c.name, g)
			}
		}
	}
	var none Palette
	for _, g := range []GroundKind{Dark, Light} {
		for _, d := range []Depth{Truecolor, Colours256} {
			ground := groundOf(g, d)
			water, _ := none.ResolveAt(WaterFill, g, d)
			line, _ := none.ResolveAt(Zone, g, d)
			tint, _ := none.ResolveAt(ZoneTint, g, d)
			line, tint, water = shownAt(line, d), shownAt(tint, d), shownAt(water, d)
			if r := Contrast(line, ground); r < LineContrast {
				t.Errorf("%v %v: the zone's outline reads at %.1f:1 on its ground; line work needs %.1f", g, d, r, LineContrast)
			}
			if r := Contrast(line, tint); r < LineContrast {
				t.Errorf("%v %v: the zone's outline reads at %.1f:1 on its tint; line work needs %.1f", g, d, r, LineContrast)
			}
			outs, tints := alertColours(g, d)
			for _, own := range []RGB{line, tint} {
				for _, other := range append([]RGB{ground, water}, append(outs[:], tints[:]...)...) {
					if diff := Difference(own, other); diff < threshold {
						t.Errorf("%v %v: the zone's %v is %.1f from %v; at least %.0f apart under every kind of vision", g, d, own, diff, other, threshold)
					}
				}
			}
		}
	}
}

// TestTheZoneTintBlendsWhereTheSearchAllows is L-8.2: the search gives the
// zone's tint, over each preset, the strongest strength that keeps every
// class separated and shows the tint - the checker agreeing - or none; and
// over the MUF and foF2 ramps none, as measured (watchpost D-196), which is
// why the zone is stippled there.
func TestTheZoneTintBlendsWhereTheSearchAllows(t *testing.T) {
	var none Palette
	for _, g := range []GroundKind{Dark, Light} {
		for _, d := range []Depth{Truecolor, Colours256} {
			gc := groundOf(g, d)
			b := SearchBlends(none, g, gc, d)
			tint, _ := none.ResolveAt(ZoneTint, g, d)
			blended := 0
			for p := Temperature; p <= FoF2; p++ {
				ramp := rampOf(none, p, g, d)
				want := 0.0
				for _, s := range blendStrengths {
					separate := true
					for _, f := range CheckBlend(ramp, tint, gc, s, d) {
						separate = separate && f.Rule != BlendSeparate
					}
					if !separate {
						continue
					}
					for i := range ramp {
						if Difference(shownAt(Blend(ramp[i], tint, s), d), shownAt(ramp[i], d)) >= visibleFloor {
							want = s
							break
						}
					}
					break
				}
				got, ok := b.ZoneStrength(p)
				if got != want || ok != (want > 0) {
					t.Errorf("%v %v preset %d: the zone blends at %.2f (%v); the checker's strongest is %.2f", g, d, p, got, ok, want)
				}
				if ok {
					blended++
				}
			}
			for _, p := range []Preset{MUF, FoF2} {
				if _, ok := b.ZoneStrength(p); ok {
					t.Errorf("%v %v: the zone's tint blends over preset %d, where none was measured to (watchpost D-196)", g, d, p)
				}
			}
			if g == Dark && d == Truecolor && blended == 0 {
				t.Error("the zone's tint blends over no preset on the dark ground, so this proves nothing")
			}
		}
	}
	if _, ok := (Blends{}).ZoneStrength(0); ok {
		t.Error("a preset that is none blends")
	}
}

// TestAlertBlendsAreUnchanged is L-8.2's last clause: the zone's search
// leaves every alert's strength as it was before it (measured at
// propagation-fields e97d5ab): temperature, then radar; extreme to unknown.
func TestAlertBlendsAreUnchanged(t *testing.T) {
	var none Palette
	want := map[[2]int][2][5]float64{
		{int(Dark), int(Truecolor)}:   {{0, 0, 0, 0, 0}, {0.2, 0.3, 0.35, 0.25, 0.25}},
		{int(Dark), int(Colours256)}:  {{0, 0, 0.1, 0, 0}, {0.25, 0.25, 0.2, 0.15, 0.2}},
		{int(Light), int(Truecolor)}:  {{0, 0, 0, 0, 0}, {0, 0, 0, 0, 0}},
		{int(Light), int(Colours256)}: {{0, 0, 0, 0, 0}, {0, 0, 0, 0, 0.35}},
	}
	for key, strengths := range want {
		g, d := GroundKind(key[0]), Depth(key[1])
		b := SearchBlends(none, g, groundOf(g, d), d)
		for i, p := range []Preset{Temperature, Radar} {
			for tint := range 5 {
				got, _ := b.Strength(p, tint)
				if got != strengths[i][tint] {
					t.Errorf("%v %v preset %d tint %d: %.2f; before the zone %.2f", g, d, p, tint, got, strengths[i][tint])
				}
			}
		}
	}
}
