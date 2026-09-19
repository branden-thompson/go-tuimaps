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

**Ruling (D-32):** GO, judged by HUM LEAD in his own terminal — "so far these look good - proceed when ready".

## Specimens 2–10

`contact-*.browser.png` are browser approximations of the pairs listed; the `.ans` files in a terminal are the real test.

| # | Files | What it shows |
|---|---|---|
| 2 | `02-block-renderer-*` | The specimen-1 view in the block renderer: the 16 quadrant glyphs, 2×2 per cell |
| 3 | `03a-no-colour-shade-*`, `03b-no-colour-digits-*`, `03c-no-colour-contours-*`, `03d-no-colour-contours-with-roads-*` | The field with no colour: density glyphs; a sparse lattice of values; contour lines with value labels, without and with the road net (`.txt` only — there is no colour to show) |
| 4 | `04-alert-polygons-*` | Four live NWS flood-warning polygons around Fort Wayne: outline plus tint, place marker |
| 5 | `05a-wind-by-speed-*`, `05b-wind-over-field-*` | A wind-arrow lattice coloured by speed; white arrows stacked over the temperature field |
| 6 | `06a-radar-as-is-*`, `06b-radar-dimmed-*` | A live national radar mosaic under the map: colours as served; at 60% brightness |
| 7 | `07a`…`07f` | The local view drawn from zoom-1, -2, -3 and -5 tiles, and at full detail; the regional view drawn from zoom-3 |
| 8 | `08a-upstream-like`, `08b-rivers`, `08c-rivers-and-parks-*` | Fort Wayne without and with the D-12 layers |
| 9 | `09-attribution-inmap-69x12` | The credit line inside a 12-row map |
| 10 | `10a`, `10b` naive; `10c`, `10d` purpose-built | Specimen 1 at 256 and 16 colours |

### Findings

| ID | Finding | Evidence | Consequence |
|---|---|---|---|
| S2-1 | **The block renderer works with a field behind it, but line work dominates.** Each quadrant is four times the area of a braille dot, so the same road net becomes thick white bars. At 149×38 it is a usable map; at 69×12 the roads swamp the field. | `02-*` at both sizes | The block renderer needs its own thinning rules, far sparser than braille's — borders, coast and the largest roads only at small sizes. It is a second style profile, not a glyph swap (bears on D-23, D-24). |
| S2-2 | The specimen uses all 16 quadrant glyphs; upstream uses 6. Block-element characters are East Asian *ambiguous* width (AI-3 §6). | — | To verify in PLAN: quadrant glyphs render one column wide in the supported terminals. **Not yet verified.** |
| S3-1 | **Density glyphs are a poor no-colour field.** The gradient is perceivable at 149×38, but the light glyphs (`·` `:`) are indistinguishable from braille map dots, and at 69×12 the frame is mush. | `03a-*` | Not recommended. |
| S3-2 | **A sparse lattice of values reads clearly with no colour** — like a station plot: honest, precise, and it leaves the map lines alone. | `03b-*` | Recommended no-colour form for scalar grids. Closes the choice left open in D-23, subject to HUM LEAD's ruling. |
| S3-3 | **Contour lines with value labels work with no colour, at both sizes** (added under D-34). Isotherms every 2 °C drawn in box-drawing strokes (`─ │ /`) with the value set into the line read as a monochrome weather map: the warm south-west and cool north-east are visible as a *pattern*, which the value lattice cannot show. Box-drawing strokes are visibly different from braille map lines and from solid-braille lakes. | `03c-*` at 149×38 and 69×12 | Contours are viable as the no-colour form for scalar grids. The specimen traces cell boundaries, so its lines are stair-stepped; a proper contouring pass with diagonal strokes would smooth them — cosmetic, not a feasibility question. Box-drawing characters are East Asian *ambiguous* width (as S2-2); **not yet verified** in the supported terminals. |
| S3-4 | **Contours need the basemap thinned.** With the full road net left in, the roads shred the contours into fragments. With roads removed — borders, water and labels kept — they read cleanly. | `03d` against `03c` | The no-colour field profile drops roads. A fourth instance of the cross-cutting rule that the basemap yields to what is on top. |
| S4-1 | **Alert polygons are the strongest result.** The warnings follow the rivers; the place marker is visibly *inside* the northern polygon; "am I in it?" is answered at a glance at both sizes. Event labels fit at 149×38 only. | `04-*` | M1's core scenario is demonstrated. Tint carries the shape; the outline adds little once tinted. Labels need a size threshold. |
| S5-1 | **The wind lattice is legible and stacks over a field.** The flow pattern and the stronger winds over both lakes read correctly. Single-cell arrows compete with a dense road net. | `05a`, `05b` | A rule for PLAN: the basemap thins when a vector overlay is active. Colour-by-speed works on a plain basemap; over a field the arrows go monochrome. |
| S6-1 | **Radar placed as-is is immediately recognisable**, nationally, from four small tiles. | `06a` | "Place the pre-coloured image" is good enough at full colour; palette-to-intensity conversion is not needed for v1 (answers the D-14 question carried to PLAN, subject to ruling). |
| S6-2 | **As served, radar greens drown white map lines; at 60% brightness both read** and the intense cores still stand out. | `06a` against `06b` | The same luminance ceiling as S1-3 applies to images: the library dims a placed image rather than trusting its brightness. |
| S6-3 | Taking the strongest sample per cell produces isolated speckle. | `06a` | The resampling rule (strongest, majority, mean) is a contract choice for georeferenced images; settle in PLAN from a further specimen. |
| S7-1 | **A stand-in from up to three zoom levels above is a usable skeleton; from five or more it is nearly empty.** The local view from zoom-5 tiles shows highways, the state line and a river; from zoom-1, -2 and -3 it shows one state line and the marker. Never misleading, but close to blank. | `07a`, `07d`, `07e` against `07b` | The never-blank rule (D-30) is safe and cheap. Its value at city scale depends on the cache, not on embedded tiles. |
| S7-2 | **The regional view drawn from zoom-3 tiles alone is fully serviceable**: both lakes, state borders, major cities. | `07f` | **Embedded zoom 0–3 makes national and regional weather maps work with no network at all** — the scales at which fields, wind and radar are shown. Zoom 0–1 supports a world view only. Primary evidence for the D-27 depth decision, with AI-9's measured sizes. |
| S8-1 | **Rivers are legible and earn their place**: the three rivers meet at the marker, and they are the rivers the flood warnings follow. Parks as a sparse dot pattern stay unobtrusive. | `08b`, `08c` | Supports D-12. Parks-as-pattern is a workable default; revisit under a field, where parks compete with the background. |
| S9-1 | The credit line fits in the bottom row of a 12-row map, right-aligned, using 30 of 69 columns. It covers map content in that row. | `09-*` | Workable default for D-25. Fading it after a few seconds remains an option the guidelines allow. |
| S10-1 | **Automatic colour downgrade fails for fields.** At 16 colours, six of the seven muted ramp steps collapse to one grey. At 256 colours the coolest step lands on the same dark grey as water, so cold air looks like a lake. | `10a`, `10b` | **The library must know the colour depth and own a ramp per depth.** "Emit truecolor and let the host downgrade" (AI-3 §6) does not hold for overlays. Decides part of X-3 for PLAN. |
| S10-2 | **Purpose-built ramps work at both depths.** Exact palette colours, with line work chosen black or white per cell by background brightness: all seven bands distinct, every label readable. The 16-colour result is garish but fully legible; exact hues follow the user's terminal theme. | `10c`, `10d` | D-23's promise of legibility at four colour depths is achievable. It adds: a depth hint from the host; per-depth ramps; per-cell contrast selection. S1-3's "lines lighter than the ramp" becomes "lines contrast with the cell". |

### Cross-cutting

1. **One rendering core carried every specimen.** A cell is a glyph, a foreground and a background; features, grids, vector grids and images all reduce to writing those three. The five input shapes of D-14 do not need five renderers.
2. **The background belongs to areas, the foreground to lines and glyphs, and there is a priority order among areas**: water, then placed images and fields, then tints. Stating that order is most of the overlay compositing design.
3. **Legibility is a contrast budget.** S1-3, S6-2 and S10-2 are the same rule seen three times: whatever owns the background is held to a brightness band, and whatever is drawn on it is chosen to contrast.
4. **The basemap is not fixed; it yields.** It thins under vector overlays (S5-1), under the block renderer (S2-1) and at small sizes; water gives up its dots under a field (S1-2). The D-24 styles are therefore a small family of profiles, selected by renderer, size and what is on top.
5. **RS-3 can close.** Field, radar, polygons and wind are all legible at both of the first host's map sizes, in both renderers, at every colour depth, given the rules above. What remains is design work, not feasibility.

### Limits of this evidence

Throwaway code; nothing here measures speed or memory. Browser approximations draw braille differently from a terminal. One region, one evening's weather, a dark terminal background; a light-background terminal was not tried. Line simplification and polygon clipping (D-16) were not exercised — the live alert polygons had 5 to 21 vertices. Radar resampling and the quadrant-glyph width question are open.
