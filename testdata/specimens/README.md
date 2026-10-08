# The specimens (v0.3.0 C-5)

What the library draws for a whole-globe grid and for a host-type grid, at every colour depth on both grounds. Each change v0.3.0 makes to fields is seen against these frames. `TestTheSpecimensAreWhatTheLibraryDraws` (`specimen_test.go`) fails when the library draws a different frame. A change that is meant writes them again with `go test -run TestTheSpecimensAreWhatTheLibraryDraws -update-specimens .`, and the changelog says so.

Each file is one frame, 120 × 36 cells, of the whole world (`FitWorld`), as the terminal receives it. `cat` one in a terminal to see it.

Files are named `<specimen>-<depth>-<ground>.ans`:

- **Depths:** `truecolor`, `256` and `16`.
- **Grounds:** `dark` (RGB 16, 16, 16) and `light` (RGB 250, 250, 245), each declared with `Ground`.

| Specimen | What is set |
|---|---|
| `globe-temperature` | A whole-globe 2° grid (180 × 91 points, 180°W to 180°E, 90°N to 90°S) with the `temperature` preset |
| `host-type` | The same grid in the host's own type: MHz, broken at 8, 12, 16, 20, 24 and 28, lines off |
| `host-type-lines` | The same, lines on |

The grid's values are a smooth rule, highest near the equator on the day side, as a MUF map is: `scale · (10 + 12 · cos(lat) · (1 + 0.6 · cos(lon − 30°)) + 2 · sin(lon · 4))`, with a scale of 1 for the host type and 1.4 for temperature.

## What they show (v0.2.0's drawing)

- **A host-type grid without lines draws nothing at truecolor and at 256 colours, on either ground.** Its frames are the bare world's, byte for byte. This is the silent blank that L-2.4 closes.
- **At 16 colours, a host-type grid draws its labelled contours with lines on or off.** The two frames are identical.
- The `globe-temperature` frames draw at every depth on both grounds.

## Data notice

The frames draw the library's embedded map tiles: map data **© OpenStreetMap contributors** (ODbL), in the **OpenMapTiles** schema (© OpenMapTiles, CC-BY 4.0), served by OpenFreeMap. Each frame carries that attribution on its last line.
