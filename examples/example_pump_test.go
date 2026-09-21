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
	wake := make(chan struct{}, 1)   // the library has work
	redraw := make(chan struct{}, 1) // the pump has done some
	nudge := func(c chan struct{}) {
		select {
		case c <- struct{}{}:
		default: // already awake
		}
	}
	if err := m.OnPending(func() { nudge(wake) }); err != nil {
		fmt.Println(err)
		return
	}
	var pump sync.WaitGroup
	for range 2 {
		pump.Add(1)
		go func() {
			defer pump.Done()
			for {
				did, err := m.Work(ctx)
				if err != nil {
					return
				}
				if did {
					nudge(redraw)
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
