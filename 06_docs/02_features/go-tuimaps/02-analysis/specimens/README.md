# Rendered Specimens — DISCOVER Tier 2 (AI-6)

Authorised by ruling D-31. Every overlay *rendering approach* in this project is ratified from a rendering, never from prose (SH-5).

## How these were made

A throwaway program, kept **outside this repository**, built from off-the-shelf libraries. It is not the start of the implementation and none of it is copied into the project; FULL TDD is untouched. Only the specimen files and the findings below enter the repo.

| File type | What it is | How to judge it |
|---|---|---|
| `*.ans` | The specimen with 24-bit colour codes | **This is the real test.** In a truecolor terminal at least as wide as the specimen: `cat <file>.ans` |
| `*.txt` | The same frame with no colour | Shape only |
| `*.browser.png` | A browser approximation on a forced cell grid | Convenience only. Browser fonts draw braille differently from a terminal; the terminal is the truth. |

Data: real OpenMapTiles vector tiles from the default network source, one view's worth, fetched once and cached (© OpenMapTiles © OpenStreetMap). Field data as labelled in each file's header.

## Specimen 1 — a field as cell background under braille map lines (RS-3, go / no-go)

View: US Midwest centred on Fort Wayne, Indiana. Field: live 2 m temperature, a 12×8 lattice (96 points) from one keyless request, interpolated to **one sample per cell** and drawn as a 7-step banded ramp.

| File | Variant |
|---|---|
| `01a-field-water-filled-149x38` | Water drawn as filled dots, as upstream does |
| `01b-field-water-owns-cells-149x38` | Water owns its cells — the field is masked over water; lighter line work |
| `01b-field-water-owns-cells-69x12` | The same at the first host's smallest map window |
| `01c-control-no-field-149x38` | Control: the same view with no field |

### Findings

| ID | Finding | Evidence | Consequence |
|---|---|---|---|
| S1-1 | **The approach works at both sizes.** The temperature gradient, the lakes, the labels and the place marker all read; "where is it warm relative to my place" is answerable at a glance, including at 69×12 cells. | `01b` at both sizes | **Agent assessment: GO**, from the browser approximation. **The ruling is HUM LEAD's, from the `.ans` files in his own terminal.** RS-3 moves from High to Medium on that confirmation. |
| S1-2 | **Water must own its cells.** Upstream lights every dot of a water body in the foreground. Under a field that reads as textured land, and the field bleeds across the lake. When water takes the cell *background* and the field is masked there, lakes read instantly. | `01a` against `01b` | A rendering rule for PLAN: area features that must stay recognisable under a field (water first) claim the background; the field yields. Differs from upstream P-22 / P-15 behaviour when a field is active; flagged for the matrix. |
| S1-3 | **Line work must be lighter than every ramp step, and the ramp must stay in a dark-to-mid band.** Mid-grey roads vanished on the mid-tone bands; near-white lines and white labels read on all seven. | first render against `01b` | A constraint on the D-24 styles and on the palette roles (D-26): overlay ramps get a luminance ceiling; line and label roles get a floor. Bears on the four-colour-depth promise (D-23). |
| S1-4 | **Borders and roads are indistinguishable when both are light, and the road net is too dense at regional zoom.** | `01b-149x38`: the Indiana–Ohio line cannot be told from a highway | Style work for D-24: a border / road hierarchy by colour or dash pattern, and road thinning by zoom. |
| S1-5 | **One sample per cell is enough, and the data is tiny.** 96 points interpolated smoothly across 5,662 cells. | file header; visual | Confirms the Tier 1 sizing of R-4: a scalar grid is a contract problem, not a throughput problem. |
| S1-6 | **A banded ramp needs a legend, and the legend is part of the picture.** Without the scale line the bands mean nothing. | legend row | A contract implication for scalar grids: the library exposes the class breaks and colours so a host can draw a legend wherever it likes (same pattern as credits, D-25). |
| S1-7 | **Low-zoom tiles are heavy.** The four zoom-5 tiles behind this view are 284–347 KB each as served: **1.2 MB for one cold regional view.** | tile cache sizes | Bears on M2's cold target (D-30), on decode cost and memory (D-29), and on the embedded-tile depth decision (D-27): a low-zoom set stripped to what a terminal can show may be far smaller. Handed to AI-9. |

### Not yet shown

Specimens 2–9 (block renderer, no-colour field, alert polygon, wind lattice, radar image, stand-in basemap, rivers and parks, attribution line), and the same views at 256-colour and 16-colour depth.
