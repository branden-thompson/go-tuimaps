# Measurement programs — reference copies (D-32)

The programs behind wave 2 (`../wave2-measurements.md`) and specimen 29's radar and blend findings,
filed with the exact inputs they ran on and their raw output, **so every number can be re-run**.
They are **reference files, not built by the gate**: the `.go.txt` suffix keeps them out of the
build. They were written against the library at `ca24461`–`7288375`. **If the API has changed since,
they may no longer compile, and nothing will say so** (the accepted cost of D-32). Update a program
when you re-run it.

| Program | What it measures | Inputs | Output |
|---|---|---|---|
| `spec29-render.go.txt` | Specimen 29's frames through the public calls: 69×12 and 149×38, truecolor and no colour; `NORADAR=1` gives the 29c control | `inputs/spec29/` | the `29*` specimen files |
| `spec29-radar-crop.go.txt` | Crops the county radar image under the 250,000-pixel cap | `inputs/spec29/radar-county.png` | `inputs/spec29/radar-crop.png` |
| `spec29b-radar-over.patch` | 29b: the one-line renderer change that draws the image over the tint | — | — |
| `spec29d-f-blend.patch` | 29d–f: the blend of the tint over the image, strength from `SPECIMEN_ALPHA` | — | — |
| `blend-measure_test.go.txt` | Class separation under the blend, at 20/35/50 %, all five severities, both grounds, both pair kinds | the library's own ramps | `output/blend-measure.txt` |
| `mrms-legend.go.txt` | The MRMS legend's pixel layout | `inputs/mrms/legend.png` | — |
| `mrms-measure.go.txt` | Wave 2 M-A: frames against a legend-derived table, independently and through the library | `inputs/mrms/` | `output/mrms-measure.txt` |
| `mrms-palette.go.txt` | Wave 2 M-A: the palette MRMS uses, across 18 frames | `inputs/mrms/` | `output/mrms-palette.txt` |
| `loopmem.go.txt` | Wave 2 M-B: heap cost of twelve frames on a map | `inputs/loopmem/`, `inputs/spec29/n0q-table-raw.json` | `output/loopmem.txt` |

## Re-running one

1. Copy the program into an empty directory outside the repository as `main.go`
   (`blend-measure_test.go.txt` goes into a copy of the repository instead, as
   `internal/colour/zz_blend_specimen_test.go`, because it uses unexported code).
2. Add a `go.mod` with `require github.com/branden-thompson/go-tuimaps v0.0.0` and
   `replace github.com/branden-thompson/go-tuimaps => <path to your checkout>`, then run
   `GOWORK=off go mod tidy`.
3. Lay the inputs out where the program looks, which is relative to where it runs:
   - `mrms-measure` and `mrms-palette` run in a subdirectory of a copy of `inputs/mrms/`.
   - `loopmem` runs in a directory holding `frames/`, a copy of `inputs/loopmem/`, and has
     `../spec29/n0q-table-raw.json` beside it.
   - `spec29-render` takes the inputs directory as its first argument.
4. `GOWORK=off go run .`, then compare with `output/`.

Specimen 29's frames also need the network: they draw OpenFreeMap tiles for the basemap.

**Machine:** every output here came from one 18-core Apple M-series Mac on 2026-09-23. The memory
figures in `output/loopmem.txt` are Go heap bytes (`HeapAlloc` after two collections) in MiB.
