package tuimaps_test

import (
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
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
		{"SetStyle", func(m *tuimaps.Map) error { return m.SetStyle([]byte(userStyle)) }},
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
		{"FitTo", func(m *tuimaps.Map) error { return m.FitTo([]tuimaps.LonLat{{Lon: -84, Lat: 33}}, nil, 1) }},
		{"SetBound", func(m *tuimaps.Map) error { return m.SetBound(tuimaps.Bound{}) }},
		{"SetDetail", func(m *tuimaps.Map) error { return m.SetDetail(tuimaps.DetailFull) }},
		{"SetImageBudget", func(m *tuimaps.Map) error { return m.SetImageBudget(0) }},
		{"SetNearby", func(m *tuimaps.Map) error { return m.SetNearby(0) }},
		{"Report", func(m *tuimaps.Map) error { _, err := m.Report(nil); return err }},
		{"Source", func(m *tuimaps.Map) error { return m.Source("") }},
		{"SetFetchOptions", func(m *tuimaps.Map) error { return m.SetFetchOptions(tuimaps.FetchOptions{}) }},
		{"CacheRoot", func(m *tuimaps.Map) error { return m.CacheRoot(t.TempDir(), 0) }},
		{"SetCacheMaxAge", func(m *tuimaps.Map) error { return m.SetCacheMaxAge(0) }},
		{"Purge", func(m *tuimaps.Map) error { _, err := m.Purge(); return err }},
		{"Verify", func(m *tuimaps.Map) error { _, _, err := m.Verify(); return err }},
		{"SetPlayback", func(m *tuimaps.Map) error { return m.SetPlayback(tuimaps.PlaybackOn) }},
		{"SetPlaybackStep", func(m *tuimaps.Map) error { return m.SetPlaybackStep(0) }},
		{"Play", func(m *tuimaps.Map) error { return m.Play() }},
		{"Stop", func(m *tuimaps.Map) error { return m.Stop() }},
		{"Reset", func(m *tuimaps.Map) error { return m.Reset() }},
		{"Step", func(m *tuimaps.Map) error { return m.Step(1) }},
		{"ShowMoment", func(m *tuimaps.Map) error { return m.ShowMoment(noon, noon.Add(time.Hour)) }},
		// A panic inside Settle's locked region, where the map's lock is held:
		// the lock is released with it, and the map is still usable.
		{"Settle.locked", func(m *tuimaps.Map) error { _, err := m.Settle(context.Background()); return err }},
	}
	for _, c := range calls {
		m := world(t, 40, 12)
		tuimaps.PlantPanic(m, c.name)
		err := c.run(m)
		tuimaps.ClearPanic(m)
		if !isKind(err, fault.Internal) {
			t.Errorf("a panic inside %s gave %v; want an error of the internal kind", c.name, err)
		}
		// And the map still works: a lock left held would hang the next call.
		if err := renderWithin(m, 5*time.Second); err != nil {
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
		{"Changed", func(m *tuimaps.Map) { m.Changed() }},
		{"FrameTicks", func(m *tuimaps.Map) { m.FrameTicks() }},
		{"Loop", func(m *tuimaps.Map) { m.Loop() }},
		{"Units", func(m *tuimaps.Map) { m.Units(true, true) }},
		{"ShowStamp", func(m *tuimaps.Map) { m.ShowStamp(false) }},
		{"PaintGround", func(m *tuimaps.Map) { m.PaintGround() }},
		{"CacheUse", func(m *tuimaps.Map) { m.CacheUse() }},
		{"SourceCredit", func(m *tuimaps.Map) { m.SourceCredit() }},
		{"DeepestZoom", func(m *tuimaps.Map) { m.DeepestZoom() }},
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
}

// panicProven is every public call a panic test plants inside, with the
// calls that cannot be planted for a reason of their own: Warnings is how a
// planted panic in a quiet call is read back, so it cannot be its own
// witness; Close is planted in TestAPanicInsideCloseStaysInside.
var panicProven = map[string]bool{"Warnings": true, "Close": true}

// TestEveryPublicCallIsPanicTested (REVIEW, code quality 2): the panic
// tests name every exported method of Map, so a call added later cannot be
// left out of rule 4's proof.
func TestEveryPublicCallIsPanicTested(t *testing.T) {
	src, err := os.ReadFile("panic_test.go")
	if err != nil {
		t.Fatal(err)
	}
	mt := reflect.TypeOf(&tuimaps.Map{})
	for i := range mt.NumMethod() {
		name := mt.Method(i).Name
		if panicProven[name] {
			continue
		}
		if !strings.Contains(string(src), "{\""+name+"\", func(m *tuimaps.Map)") {
			t.Errorf("Map.%s is in no panic test; add it to TestPanicRecoveredAtEveryPublicCall or TestPanicInACallThatAnswersNothing", name)
		}
	}
}

// TestAPanicInsideCloseStaysInside: a panic planted in Close does not escape
// and is noted, and the map closes on the next Close.
func TestAPanicInsideCloseStaysInside(t *testing.T) {
	m, err := tuimaps.New(tuimaps.WithSize(40, 12))
	if err != nil {
		t.Fatal(err)
	}
	tuimaps.PlantPanic(m, "Close")
	m.Close() // must not panic out of the library
	tuimaps.ClearPanic(m)
	if len(m.Warnings()) == 0 {
		t.Error("a panic inside Close left no warning behind")
	}
	m.Close()
	if _, err := m.Render(tuimaps.Size{Cols: 40, Rows: 12}, noon); !isKind(err, fault.Closed) {
		t.Errorf("after Close a Render gave %v; want the closed kind", err)
	}
}

// renderWithin renders, or fails if the map does not answer in time - a lock
// a panic left held hangs the call rather than failing it.
func renderWithin(m *tuimaps.Map, limit time.Duration) error {
	done := make(chan error, 1)
	go func() {
		_, err := m.Render(tuimaps.Size{Cols: 40, Rows: 12}, noon)
		done <- err
	}()
	select {
	case err := <-done:
		return err
	case <-time.After(limit):
		return errors.New("Render did not return: the map's lock is still held")
	}
}
