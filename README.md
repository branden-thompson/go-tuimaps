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
| **Loops do not play until you say so** | An image overlay can carry a loop of frames, and its playback is off by default: the map shows the newest observed frame and holds it. Turn playback on with `SetPlayback(PlaybackOn)`, then call `Play` |
| **A fast step lifts the flash ceiling** | The project keeps everything on the map to at most 2.5 changes a second. `SetPlaybackStep` takes 200 ms to 1000 ms, 500 ms by default; a step below 400 ms lifts that ceiling, and `Loop` says so. Say so beside the setting where you offer it (D-76) |

## What it sends and stores

Nothing, until you name a source. With one named:

- tiles are fetched from that source and from nowhere else, over HTTPS;
- what is sent is the tile address and a user-agent naming this library and
  its version, `go-tuimaps/0.2.0` - nothing about your machine. The version
  is the release's: `scripts/gate --release` refuses a tag it does not
  match. `SetFetchOptions` adds a name of your own to it, and takes a
  transport of your own for your proxy or trust roots;
- tiles are held in memory, and on disk only if you name a directory with
  `CacheRoot`. A cached file's time is when its tile was fetched: reading a
  tile writes nothing. `SetCacheMaxAge` sets how long a tile is kept.
  `Purge()` empties every source's tiles from disk and memory, and
  `Verify()` reads the disk cache back.

## The data

The tiles the `assets` package carries are from OpenFreeMap, built from
OpenStreetMap data. The map draws their credit itself; **an application that
shows this map should also carry the credit in full in its About**, with the
links `assets.Notice()` gives.

## Compatibility

The public surface is the contract in `06_docs`. The error kinds and warning
kinds are closed lists: a later version adds to them only in a minor release,
and says so. A frame is valid until the next `Render` on the same map.

v0.2.0 breaks some of v0.1.0's shape, by ruling D-58: names removed or
redefined, and the warning kinds renumbered. Each break, and what a host does
instead, is listed in section 12 of
`06_docs/02_features/go-tuimaps/03-architecture-design/contract.md`, which is
the migration note from v0.1.0. Compare kinds by name, never by number.

## Not built yet

Flash and pulse markers (blinking markers are built), camera tours, the block
renderer, a tile-image provider, a PMTiles source, pointer operations (a cell
to a coordinate, or to what lies under it), and styles written with
expressions (a style's filters are read in the legacy format). Each is
designed and none is promised for a date.

## Working on the library

Every change is checked by `scripts/gate`, which is what stands before a merge:

- `scripts/gate` runs every leg, about thirty minutes on an 18-core Apple M-series Mac; every run's
  seconds are in `06_docs/gate-runs.md`. It needs the go1.25.0 toolchain (downloaded on first use),
  the pinned `govulncheck` on `PATH` (`go install golang.org/x/vuln/cmd/govulncheck@v1.8.0`), and a
  machine that can run the other architecture's tests (an Apple silicon Mac runs amd64 under
  emulation).
- `scripts/gate --quick` skips fuzzing, the vulnerability scan, the cross-compiles, the five-minute
  loop-memory leg and the second architecture. Formatting, vet, the tests and the licence check
  still run.
- `scripts/gate --docs` is for a staged change of Markdown and nothing else: stage, run it, commit.
  It refuses any other staged file.
- `scripts/gate --fuzz` runs the fuzz legs alone.
- `scripts/gate --soak` runs the loop-memory check for an hour: a 12-frame loop played while the heap
  is watched. It is run once before a release and recorded; the full gate runs the same test for
  five minutes.
- `scripts/gate --release TAG` is the release check. A final tag (`vX.Y.Z`) is refused while a row of
  the release checklist is unticked; a release candidate (`vX.Y.Z-rc.N`) is not. Either way the
  library's version must be the tag's, and the pinned `govulncheck` must find nothing reachable that
  the checklist does not carry as a ruled exception.

Give one mode at most. Every run leaves a line in `06_docs/gate-runs.md`. The gate also needs `perl`
(present on macOS and most Linux systems) for the fuzz legs' time limit, and it writes a throw-away
`go.work` so the nested modules (`cmd/tuimaps`, `examples`, `tools/*`) resolve the library from this
tree; outside the gate, build those from their own directory with a workspace of your own.

## Licence

MIT. Both upstream notices are carried: TerminalMap, which this is a port
of, and MAPSCII, which it was inspired by.
