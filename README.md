# go-tuimaps

A map in a terminal, drawn as braille cells, with your own data over it:
alert areas, radar images, scalar fields such as temperature.

It is a library first. It starts no goroutine, opens no connection it was
not told to, reads no keys and owns no clock.

## Quick start

A map of the world in three calls, from the tiles the library carries, with
nothing fetched:

```go
package main

import (
	"context"
	"fmt"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
)

func main() {
	m, err := tuimaps.New(tuimaps.WithSize(80, 24), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		panic(err)
	}
	defer m.Close()
	if _, err := m.Settle(context.Background()); err != nil {
		panic(err)
	}
	frame, err := m.Render(tuimaps.Size{Cols: 80, Rows: 24}, time.Now())
	if err != nil {
		panic(err)
	}
	for _, line := range frame.Lines {
		fmt.Println(line)
	}
}
```

Three calls: create, settle, render. An interactive host writes a short pump
as well - see `examples/example_pump_test.go`. **A host with a pump does not
call `Settle`**: `Settle` waits for the queue, not for work another goroutine
has already taken, so a host with a pump redraws when the pump says something
changed.

## What you need to know before you start

| | |
|---|---|
| **A font with braille** | The map is drawn with braille cells. A terminal whose font has no braille shows boxes. Whether a font has it cannot be asked of the terminal, so the library cannot warn you: check your own before you judge the picture |
| **Nothing is fetched until you say so** | `Source(address)` is the only call that reaches anything. Until then the map draws from the tiles the `assets` package carries, which are the world down to zoom 3 |
| **You run the work** | `Work(ctx)` does one unit on your goroutine. `Settle(ctx)` runs it until there is none left, for one-shot renders and tests. Nothing happens on its own |
| **You own the clock** | `Render(size, now)` takes the wall clock. Markers move on it; data going out of date is judged by it. `Animate(at)` takes animation over if you want to drive it yourself |
| **Colour is yours to set** | `SetPalette` takes your own colours, `SafeRamps(true)` keeps the library's where a scale must stay readable, `ColourDepth` hints at the terminal. With no hint, a non-empty `NO_COLOR` means no colour at all |
| **Nothing is hidden by colour alone** | With no colour, alert areas are hatched and labelled, overlay lines are dashed, and a scalar field becomes contour lines carrying their values |
| **A style is bytes, never a path** | `SetStyle(body)` draws the basemap by a style of your own, in the same JSON format as the library's. The library never opens a style file: you read it, and pass what you read |

## What it sends and stores

Nothing, until you name a source. With one named:

- tiles are fetched from that source and from nowhere else, over HTTPS;
- what is sent is the tile address and a user-agent naming this library, its
  version and a token you set - nothing about your machine;
- tiles are held in memory, and on disk only if you name a directory with
  `CacheRoot`. `Purge()` empties it and `Verify()` reads it back.

## The data

The tiles the `assets` package carries are from OpenFreeMap, built from
OpenStreetMap data. The map draws their credit itself; **an application that
shows this map should also carry the credit in full in its About**, with the
links `assets.Notice()` gives.

## Compatibility

The public surface is the contract in `06_docs`. The error kinds and warning
kinds are closed lists: a later version adds to them only in a minor release,
and says so. A frame is valid until the next `Render` on the same map.

## Not built yet

Flash and pulse markers, camera tours, the block renderer, a wind overlay,
timed image sequences, 16-colour ramps. Each is designed and none is
promised for a date.

## Licence

MIT. Both upstream notices are carried: TerminalMap, which this is a port
of, and MAPSCII, which it was inspired by.
