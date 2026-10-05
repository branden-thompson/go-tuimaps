package examples_test

import (
	"context"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
)

// TestThePumpGoesOnPastAFailedJob (REVIEW, docs 2): the example's pump,
// over tiles deeper than the embedded ones with no source - every one a
// failed job - keeps working until nothing is pending, as contract section 2
// says a pump does; one that stopped at the first error would leave the
// rest queued for ever.
func TestThePumpGoesOnPastAFailedJob(t *testing.T) {
	m, err := tuimaps.New(tuimaps.WithSize(80, 24), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if err := m.Zoom(8); err != nil {
		t.Fatal(err)
	}
	ctx, quit := context.WithCancel(context.Background())
	redraw := make(chan struct{}, 1)
	pump, err := startPump(ctx, m, redraw)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { quit(); pump.Wait() }()
	if _, err := m.Render(tuimaps.Size{Cols: 80, Rows: 24}, noon); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(10 * time.Second)
	for m.Pending() > 0 || m.InFlight() > 0 {
		if time.Now().After(deadline) {
			t.Fatalf("%d jobs still pending: the pump stopped at a failed job", m.Pending())
		}
		time.Sleep(10 * time.Millisecond)
	}
}
