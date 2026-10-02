package tuimaps_test

import (
	"context"
	"sync"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// A RENDER LEAVES AN OVERLAY'S PREPARATION RUNNING (watchpost UAT-2 U2-59):
// a render plans the view's work and withdraws what the view no longer wants,
// and the preparation of an overlay the map still holds, at the bucket in
// view, is wanted. Withdrawn, a loop whose preparation outlasts the gap
// between two renders is cancelled at every one of them and never drawn.
func TestARenderKeepsAHeldOverlaysPreparation(t *testing.T) {
	m := world(t, 149, 38)
	if _, err := m.Set(radarLoop(t, "radar", 12, 298, 152)); err != nil {
		t.Fatal(err)
	}
	if !tuimaps.KeptAtRender(m, "radar", 0) {
		t.Error("a held loop's preparation at the bucket in view is withdrawn by a render")
	}
	if tuimaps.KeptAtRender(m, "radar", 1) {
		t.Error("a held loop's preparation at a bucket not in view is kept")
	}
	if tuimaps.KeptAtRender(m, "never-set", 0) {
		t.Error("the preparation of an overlay the map does not hold is kept")
	}
	if _, err := m.Remove("radar"); err != nil {
		t.Fatal(err)
	}
	if tuimaps.KeptAtRender(m, "radar", 0) {
		t.Error("a removed loop's preparation is kept")
	}
}

// And through the public calls: a loop prepared while the host renders as
// fast as it can is prepared once - every unit of work the first render
// planned runs once and lands, none cancelled and asked again.
func TestALoopIsPreparedOnceWhileTheHostRenders(t *testing.T) {
	m := world(t, 149, 38)
	size := tuimaps.Size{Cols: 149, Rows: 38}
	settle(t, m) // the basemap first: what is left is the loop's
	if _, err := m.Set(radarLoop(t, "radar", 48, 298, 152)); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Render(size, noon); err != nil {
		t.Fatal(err)
	}
	planned := m.Pending()
	if planned == 0 {
		t.Fatal("the render planned no work; this test measures nothing")
	}
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				_, _ = m.Render(size, noon)
			}
		}
	}()
	ran := 0
	for m.Pending() > 0 && ran < 100 {
		did, err := m.Work(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if did {
			ran++
		}
	}
	close(stop)
	wg.Wait()
	if ran != planned {
		t.Errorf("%d units of work ran for the %d planned: the loop's preparation was cancelled and asked again", ran, planned)
	}
}
