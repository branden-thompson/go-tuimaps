package tuimaps_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// TestWorkThroughThePublicPackage is plan task 12.19 (FR-30): the host does
// the work, one unit a call, on its own goroutines; the library starts none.
func TestWorkThroughThePublicPackage(t *testing.T) {
	m, err := tuimaps.New(tuimaps.WithSize(80, 24), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if _, err := m.Render(tuimaps.Size{Cols: 80, Rows: 24}, noon); err != nil {
		t.Fatal(err)
	}
	if m.Pending() == 0 {
		t.Fatal("a map that has drawn once wants nothing")
	}
	done := 0
	for range 100 {
		did, err := m.Work(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if !did {
			break
		}
		done++
	}
	if done == 0 {
		t.Error("Work did nothing at all")
	}
	if m.Pending() != 0 {
		t.Errorf("%d units still pending after the work was done", m.Pending())
	}
	if did, err := m.Work(context.Background()); did || err != nil {
		t.Errorf("a call with nothing to do: %v, %v", did, err)
	}
	if m.InFlight() != 0 {
		t.Errorf("%d units in flight with no other goroutine working", m.InFlight())
	}
	// A context already over is answered, not worked.
	over, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := m.Work(over); err != nil && !isKind(err, fault.Cancelled) {
		t.Errorf("a cancelled context: %v", err)
	}
	var none context.Context // a host that passes nothing at all
	if _, err := m.Work(none); !isKind(err, fault.Internal) {
		t.Errorf("no context at all: %v", err)
	}
}

// TestOnPendingWakesTheHost is the other half of 12.19: a host is told when
// there is work, rather than polling for it.
func TestOnPendingWakesTheHost(t *testing.T) {
	m, err := tuimaps.New(tuimaps.WithSize(80, 24), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	var woken atomic.Int32
	if err := m.OnPending(func() { woken.Add(1) }); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Render(tuimaps.Size{Cols: 80, Rows: 24}, noon); err != nil {
		t.Fatal(err)
	}
	if woken.Load() == 0 {
		t.Error("the hook was not called when the first frame found work to do")
	}
}

// TestCallsSafeTogether is plan task 12.22 (FR-30): the owner's calls, Render
// among them, are safe beside a pump running Work on other goroutines. Run
// under the race detector, which is where it earns its keep.
func TestCallsSafeTogether(t *testing.T) {
	m, err := tuimaps.New(tuimaps.WithSize(80, 24), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	ctx, stop := context.WithCancel(context.Background())
	var pumps sync.WaitGroup
	for range 2 {
		pumps.Add(1)
		go func() {
			defer pumps.Done()
			for ctx.Err() == nil {
				if _, err := m.Work(ctx); err != nil {
					return
				}
			}
		}()
	}
	// The owner's calls, one after another on this goroutine, while the pump
	// runs: none of them may race with it.
	for i := range 40 {
		if _, err := m.Render(tuimaps.Size{Cols: 80, Rows: 24}, noon.Add(time.Duration(i)*time.Second)); err != nil {
			t.Fatal(err)
		}
		if _, err := m.Set(warning("alerts")); err != nil {
			t.Fatal(err)
		}
		if _, err := m.SetPlaces([]tuimaps.Place{{Name: "Beacon", At: tuimaps.LonLat{Lon: 1, Lat: 1}}}); err != nil {
			t.Fatal(err)
		}
		m.SafeRamps(i%2 == 0)
		m.Layers(tuimaps.RoadLayer, i%2 == 0)
		_ = m.Legend()
		_ = m.Credits()
		_, _ = m.Scale()
		_ = m.Footer()
		_, _ = m.NextCall(noon)
		_ = m.Warnings()
		_ = m.Pending()
		_ = m.Changed()
		_ = m.InUse("alerts")
		if err := m.ZoomBy(0.25); err != nil {
			t.Fatal(err)
		}
		if _, err := m.Remove("alerts"); err != nil {
			t.Fatal(err)
		}
	}
	stop()
	pumps.Wait()
}

// TestCloseWhileWorking: a map closes at once while a pump is inside a Work
// call, and says how many calls are still inside it.
func TestCloseWhileWorking(t *testing.T) {
	m, err := tuimaps.New(tuimaps.WithSize(149, 38), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Render(tuimaps.Size{Cols: 149, Rows: 38}, noon); err != nil {
		t.Fatal(err)
	}
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	var pumps sync.WaitGroup
	for range 2 {
		pumps.Add(1)
		go func() {
			defer pumps.Done()
			for ctx.Err() == nil {
				if _, err := m.Work(ctx); err != nil {
					return
				}
			}
		}()
	}
	inside := m.Close()
	if inside < 0 {
		t.Errorf("Close says %d calls are inside the map", inside)
	}
	stop()
	pumps.Wait()
	// Everything answers the closed kind afterwards, and nothing panics.
	if _, err := m.Work(context.Background()); !isKind(err, fault.Closed) {
		t.Errorf("Work on a closed map: %v", err)
	}
	if _, err := m.Render(tuimaps.Size{Cols: 149, Rows: 38}, noon); !isKind(err, fault.Closed) {
		t.Errorf("Render on a closed map: %v", err)
	}
	if _, err := m.Set(warning("alerts")); !isKind(err, fault.Closed) {
		t.Errorf("Set on a closed map: %v", err)
	}
	if again := m.Close(); again < 0 {
		t.Errorf("closing twice: %d", again)
	}
}
