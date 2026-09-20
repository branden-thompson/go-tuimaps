package tuimaps_test

import (
	"context"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// TestPanicRecoveredAtEveryPublicCall is plan task 12.23 (contract, section
// 6, rule 4): a panic planted inside any public call becomes an error of the
// internal kind, or a warning for a call that returns none, and the map is
// still usable afterwards.
func TestPanicRecoveredAtEveryPublicCall(t *testing.T) {
	// Every public call, with a way to make it and to read what it answered.
	calls := []struct {
		name string
		run  func(m *tuimaps.Map) error
	}{
		{"Settle", func(m *tuimaps.Map) error { _, err := m.Settle(context.Background()); return err }},
		{"Render", func(m *tuimaps.Map) error {
			_, err := m.Render(tuimaps.Size{Cols: 40, Rows: 12}, noon)
			return err
		}},
		{"Work", func(m *tuimaps.Map) error { _, err := m.Work(context.Background()); return err }},
		{"OnPending", func(m *tuimaps.Map) error { return m.OnPending(func() {}) }},
		{"SetPalette", func(m *tuimaps.Map) error { _, err := m.SetPalette(nil); return err }},
		{"Ground", func(m *tuimaps.Map) error { return m.Ground(tuimaps.RGB{}) }},
		{"LabelLanguage", func(m *tuimaps.Map) error { return m.LabelLanguage("en") }},
		{"SetPlaces", func(m *tuimaps.Map) error { _, err := m.SetPlaces(nil); return err }},
		{"AddPlace", func(m *tuimaps.Map) error { _, err := m.AddPlace(miami()); return err }},
		{"RemovePlace", func(m *tuimaps.Map) error { _, err := m.RemovePlace("x"); return err }},
		{"Set", func(m *tuimaps.Map) error { _, err := m.Set(warning("alerts")); return err }},
		{"Remove", func(m *tuimaps.Map) error { _, err := m.Remove("alerts"); return err }},
		{"Recentre", func(m *tuimaps.Map) error { return m.Recentre(tuimaps.LonLat{}) }},
		{"Zoom", func(m *tuimaps.Map) error { return m.Zoom(3) }},
		{"ZoomBy", func(m *tuimaps.Map) error { return m.ZoomBy(1) }},
		{"PanCells", func(m *tuimaps.Map) error { return m.PanCells(1, 1) }},
		{"FitWorld", func(m *tuimaps.Map) error { return m.FitWorld() }},
	}
	for _, c := range calls {
		m := world(t, 40, 12)
		tuimaps.PlantPanic(m, c.name)
		err := c.run(m)
		tuimaps.ClearPanic(m)
		if !isKind(err, fault.Internal) {
			t.Errorf("a panic inside %s gave %v; want an error of the internal kind", c.name, err)
		}
		// And the map still works.
		if _, err := m.Render(tuimaps.Size{Cols: 40, Rows: 12}, noon); err != nil {
			t.Errorf("the map is unusable after a panic inside %s: %v", c.name, err)
		}
	}
}

// TestPanicInACallThatAnswersNothing: the calls that return no error stop the
// panic too, and say so in the warnings a host can ask for.
func TestPanicInACallThatAnswersNothing(t *testing.T) {
	quiet := []struct {
		name string
		run  func(m *tuimaps.Map)
	}{
		{"SafeRamps", func(m *tuimaps.Map) { m.SafeRamps(true) }},
		{"ColourDepth", func(m *tuimaps.Map) { m.ColourDepth(tuimaps.NoColour) }},
		{"ReduceMotion", func(m *tuimaps.Map) { m.ReduceMotion(true) }},
		{"Layers", func(m *tuimaps.Map) { m.Layers(tuimaps.RoadLayer, false) }},
		{"Places", func(m *tuimaps.Map) { m.Places() }},
		{"Legend", func(m *tuimaps.Map) { m.Legend() }},
		{"Credits", func(m *tuimaps.Map) { m.Credits() }},
		{"Scale", func(m *tuimaps.Map) { m.Scale() }},
		{"Footer", func(m *tuimaps.Map) { m.Footer() }},
		{"ShowFooter", func(m *tuimaps.Map) { m.ShowFooter(true) }},
		{"Animate", func(m *tuimaps.Map) { m.Animate(noon) }},
		{"FollowClock", func(m *tuimaps.Map) { m.FollowClock() }},
		{"NextCall", func(m *tuimaps.Map) { m.NextCall(noon) }},
		{"Pending", func(m *tuimaps.Map) { m.Pending() }},
		{"InFlight", func(m *tuimaps.Map) { m.InFlight() }},
		{"InUse", func(m *tuimaps.Map) { m.InUse("alerts") }},
		{"Overlays", func(m *tuimaps.Map) { m.Overlays() }},
		{"Centre", func(m *tuimaps.Map) { m.Centre() }},
	}
	for _, c := range quiet {
		m := world(t, 40, 12)
		tuimaps.PlantPanic(m, c.name)
		c.run(m) // must not panic out of the library
		tuimaps.ClearPanic(m)
		if len(m.Warnings()) == 0 {
			t.Errorf("a panic inside %s left no warning behind", c.name)
		}
		if _, err := m.Render(tuimaps.Size{Cols: 40, Rows: 12}, noon); err != nil {
			t.Errorf("the map is unusable after a panic inside %s: %v", c.name, err)
		}
	}
	_ = time.Now
}
