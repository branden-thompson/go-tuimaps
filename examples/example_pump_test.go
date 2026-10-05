package examples_test

import (
	"context"
	"fmt"
	"sync"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
)

// Example_pump is the pump an interactive host writes. The library starts no
// goroutine, so the host runs the work on its own: two goroutines call Work,
// the library's hook wakes them when there is something to do, and each unit
// done tells the host to draw again. It all stops cleanly on quit.
//
// Note what is *not* here: Settle. Settle waits for the queue, not for work
// another goroutine has already taken (D-86), so beside a running pump it is
// the wrong call. A host with a pump redraws when the pump says something
// changed, which is what this does.
func Example_pump() {
	m, _ := tuimaps.New(tuimaps.WithSize(80, 24), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	defer m.Close()

	ctx, quit := context.WithCancel(context.Background())
	defer quit()
	redraw := make(chan struct{}, 1) // the pump has done some
	pump, err := startPump(ctx, m, redraw)
	if err != nil {
		fmt.Println(err)
		return
	}

	// The host draws, and draws again whenever the pump has done something,
	// until the frame is as good as it gets.
	frame, err := m.Render(tuimaps.Size{Cols: 80, Rows: 24}, noon)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("first frame:", frame.Status)
	for frame.Status != tuimaps.Complete {
		<-redraw
		if frame, err = m.Render(tuimaps.Size{Cols: 80, Rows: 24}, noon); err != nil {
			fmt.Println(err)
			return
		}
	}
	fmt.Println("once the pump has caught up:", frame.Status)

	quit()
	pump.Wait()
	// Output:
	// first frame: no tiles
	// once the pump has caught up: complete
}

// pumpWidth is how many goroutines the pump runs: two is what the
// contract's memory line assumes (D-84).
const pumpWidth = 2

// startPump runs the pump: pumpWidth goroutines calling Work, woken by the
// library's hook, each unit done nudging redraw. **A job that failed comes
// back as Work's error with did true, and the pump goes on** - a failed tile
// is the library's to retry and warn of. Only a map closed or a context
// ended stops it (contract, section 2).
func startPump(ctx context.Context, m *tuimaps.Map, redraw chan struct{}) (*sync.WaitGroup, error) {
	// One slot for each pump goroutine, every free slot filled at a wake, so
	// no goroutine sleeps on work another wake brought (contract, section 2).
	wake := make(chan struct{}, pumpWidth)
	nudge := func() {
		for range pumpWidth {
			select {
			case wake <- struct{}{}:
			default:
				return // every slot is full
			}
		}
	}
	if err := m.OnPending(nudge); err != nil {
		return nil, err
	}
	var pump sync.WaitGroup
	for range pumpWidth {
		pump.Add(1)
		go func() {
			defer pump.Done()
			for {
				did, err := m.Work(ctx)
				if kind, ok := tuimaps.KindOf(err); ok && (kind == tuimaps.Closed || kind == tuimaps.Cancelled) {
					return
				}
				if did {
					select {
					case redraw <- struct{}{}:
					default: // a redraw is already due
					}
					continue // there may be more
				}
				select {
				case <-ctx.Done():
					return
				case <-wake:
				}
			}
		}()
	}
	return &pump, nil
}
