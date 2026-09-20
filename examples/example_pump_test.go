package examples_test

import (
	"context"
	"fmt"
	"sync"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
)

// Example_pump is the ten lines an interactive host writes: the library
// starts no goroutine, so the host runs the work on its own. Two goroutines
// call Work; OnPending wakes them; the whole thing stops cleanly on quit.
func Example_pump() {
	m, _ := tuimaps.New(tuimaps.WithSize(80, 24), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	defer m.Close()

	ctx, quit := context.WithCancel(context.Background())
	wake := make(chan struct{}, 1)
	if err := m.OnPending(func() {
		select {
		case wake <- struct{}{}:
		default: // already awake
		}
	}); err != nil {
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

	// The host draws whenever it likes; the pump fills the map in behind it.
	frame, err := m.Render(tuimaps.Size{Cols: 80, Rows: 24}, noon)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("first frame:", frame.Status)
	if _, err := m.Settle(ctx); err != nil {
		fmt.Println(err)
		return
	}
	frame, _ = m.Render(tuimaps.Size{Cols: 80, Rows: 24}, noon)
	fmt.Println("once the work is done:", frame.Status)

	quit()
	pump.Wait()
	// Output:
	// first frame: no tiles
	// once the work is done: complete
}
