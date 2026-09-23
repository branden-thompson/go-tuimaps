# Rendered Specimens — DISCOVER Tier 2 (AI-6)

Authorised by ruling D-31. Every overlay *rendering approach* in this project is ratified from a rendering, never from prose (SH-5).

## How these were made

A throwaway program, kept **outside this repository**, built from off-the-shelf libraries. It is not the start of the implementation and none of it is copied into the project; FULL TDD is untouched. Only the specimen files and the findings below enter the repo. **Specimens 26 onward are different**: they were drawn by the library's own public calls, and specimen 29's programs are filed under `radar-loops/02-analysis/programs/` (v0.2.0 D-32).

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
| S1-1 | **The approach works at both sizes.** The temperature gradient, the lakes, the labels and the place marker all read; "where is it warm relative to my place" is answerable at a glance, including at 69×12 cells. | `01b` at both sizes | **Coordinator's assessment: GO**, from the browser approximation. **The ruling is HUM LEAD's, from the `.ans` files in HUM LEAD's own terminal.** RS-3 moves from High to Medium on that confirmation. |
| S1-2 | **Water must own its cells.** Upstream lights every dot of a water body in the foreground. Under a field that reads as textured land, and the field bleeds across the lake. When water takes the cell *background* and the field is masked there, lakes read instantly. | `01a` against `01b` | A rendering rule for PLAN: area features that must stay recognisable under a field (water first) claim the background; the field yields. Differs from upstream P-22 / P-15 behaviour when a field is active; flagged for the matrix. |
| S1-3 | **Line work must be lighter than every ramp step, and the ramp must stay in a dark-to-mid band.** Mid-grey roads vanished on the mid-tone bands; near-white lines and white labels read on all seven. | first render against `01b` | A constraint on the D-24 styles and on the palette roles (D-26): overlay ramps get a luminance ceiling; line and label roles get a floor. Bears on the four-colour-depth promise (D-23). |
| S1-4 | **Borders and roads are indistinguishable when both are light, and the road net is too dense at regional zoom.** | `01b-149x38`: the Indiana–Ohio line cannot be told from a highway | Style work for D-24: a border / road hierarchy by colour or dash pattern, and road thinning by zoom. |
| S1-5 | **One sample per cell is enough, and the data is tiny.** 96 points interpolated smoothly across 5,662 cells. | file header; visual | Confirms the Tier 1 sizing of R-4: a scalar grid is a contract problem, not a throughput problem. |
| S1-6 | **A banded ramp needs a legend, and the legend is part of the picture.** Without the scale line the bands mean nothing. | legend row | A contract implication for scalar grids: the library exposes the class breaks and colours so a host can draw a legend wherever it likes (same pattern as credits, D-25). |
| S1-7 | **Low-zoom tiles are heavy.** The four zoom-5 tiles behind this view are 284–347 KB each as served: **1.2 MB for one cold regional view.** | tile cache sizes | Bears on M2's cold target (D-30), on decode cost and memory (D-29), and on the embedded-tile depth decision (D-27): a low-zoom set stripped to what a terminal can show may be far smaller. Handed to AI-9. |

**Ruling (D-32):** GO, judged by HUM LEAD in HUM LEAD's own terminal — "so far these look good - proceed when ready".

## Specimens 2–11

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
| 11 | `11a-radar-regional-colour-*`, `11b-radar-no-colour-shade-*`, `11c-radar-no-colour-contours-*` | Live radar over the US Southeast: in colour for reference; with no colour as block shades (both sizes); with no colour as contours |

### Findings

| ID | Finding | Evidence | Consequence |
|---|---|---|---|
| S2-1 | **The block renderer works with a field behind it, but line work dominates.** Each quadrant is four times the area of a braille dot, so the same road net becomes thick white bars. At 149×38 it is a usable map; at 69×12 the roads swamp the field. | `02-*` at both sizes | The block renderer needs its own thinning rules, far sparser than braille's — borders, coast and the largest roads only at small sizes. It is a second style profile, not a glyph swap (bears on D-23, D-24). |
| S2-2 | The specimen uses all 16 quadrant glyphs; upstream uses 6. Block-element characters are East Asian *ambiguous* width (AI-3 §6). | — | To verify in PLAN: quadrant glyphs render one column wide in the supported terminals. **Not yet verified.** |
| S3-1 | **Density glyphs are a poor no-colour field.** The gradient is perceivable at 149×38, but the light glyphs (`·` `:`) are indistinguishable from braille map dots, and at 69×12 the frame is mush. | `03a-*` | Not recommended. |
| S3-2 | *[Note 2026-09-18: now the fall-back, not the recommendation — rulings D-34, D-35. Rendered at 149×38 only.]* **A sparse lattice of values reads clearly with no colour** — like a station plot: honest, precise, and it leaves the map lines alone. | `03b-*` | Recommended no-colour form for scalar grids. Closes the choice left open in D-23, subject to HUM LEAD's ruling. |
| S3-3 | **Contour lines with value labels work with no colour, at both sizes** (added under D-34). Isotherms every 2 °C drawn in box-drawing strokes (`─ │ /`) with the value set into the line read as a monochrome weather map: the warm south-west and cool north-east are visible as a *pattern*, which the value lattice cannot show. Box-drawing strokes are visibly different from braille map lines and from solid-braille lakes. | `03c-*` at 149×38 and 69×12 | Contours are viable as the no-colour form for scalar grids. The specimen traces cell boundaries, so its lines are stair-stepped; a proper contouring pass with diagonal strokes would smooth them — cosmetic, not a feasibility question. Box-drawing characters are East Asian *ambiguous* width (as S2-2); **not yet verified** in the supported terminals. |
| S3-4 | **Contours need the basemap thinned.** With the full road net left in, the roads shred the contours into fragments. With roads removed — borders, water and labels kept — they read cleanly. | `03d` against `03c` | The no-colour field profile drops roads. A fourth instance of the cross-cutting rule that the basemap yields to what is on top. |
| S4-1 | **Alert polygons are the strongest result.** The warnings follow the rivers; the place marker is visibly *inside* the northern polygon; "am I in it?" is answered at a glance at both sizes. Event labels fit at 149×38 only. | `04-*` | M1's core scenario is demonstrated. Tint carries the shape; the outline adds little once tinted. Labels need a size threshold. |
| S5-1 | **The wind lattice is legible and stacks over a field.** The flow pattern and the stronger winds over both lakes read correctly. Single-cell arrows compete with a dense road net. | `05a`, `05b` | A rule for PLAN: the basemap thins when a vector overlay is active. Colour-by-speed works on a plain basemap; over a field the arrows go monochrome. |
| S6-1 | *[Superseded 2026-09-18: corrected by S6-4 below, then overtaken by ruling D-36 — the library re-colours placed images; see D-39, and D-45, which settled what D-39 left open.]* **Radar placed as-is is immediately recognisable**, nationally, from four small tiles. | `06a` | "Place the pre-coloured image" is good enough at full colour; palette-to-intensity conversion is not needed for v1 (answers the D-14 question carried to PLAN, subject to ruling). |
| S6-2 | **As served, radar greens drown white map lines; at 60% brightness both read** and the intense cores still stand out. | `06a` against `06b` | The same luminance ceiling as S1-3 applies to images: the library dims a placed image rather than trusting its brightness. |
| S6-3 | Taking the strongest sample per cell produces isolated speckle. | `06a` | The resampling rule (strongest, majority, mean) is a contract choice for georeferenced images; settle in PLAN from a further specimen. |
| S6-4 | **Correction to S6-1.** Reading intensity back from the image's colours *is* needed — not for truecolor display, but for every other case: the no-colour form (S11-2) and the 16- and 256-colour forms (S10-1 applies to images exactly as it does to ramps). | `11b`; S10-1 | "Place the pre-coloured image as-is" holds at full colour only. A placed image needs a colour-to-class table supplied with it *("optional" as first written; D-45 made it required)*, so the library can re-ramp it per colour depth and shade it with no colour. Revises the answer carried to PLAN under D-14. |
| S7-1 | **A stand-in from up to three zoom levels above is a usable skeleton; from five or more it is nearly empty.** The local view from zoom-5 tiles shows highways, the state line and a river; from zoom-1, -2 and -3 it shows one state line and the marker. Never misleading, but close to blank. | `07a`, `07d`, `07e` against `07b` | The never-blank rule (D-30) is safe and cheap. Its value at city scale depends on the cache, not on embedded tiles. |
| S7-2 | **The regional view drawn from zoom-3 tiles alone is fully serviceable**: both lakes, state borders, major cities. | `07f` | **Embedded zoom 0–3 makes national and regional weather maps work with no network at all** — the scales at which fields, wind and radar are shown. Zoom 0–1 supports a world view only. Primary evidence for the D-27 depth decision, with AI-9's measured sizes. |
| S8-1 | **Rivers are legible and earn their place**: the three rivers meet at the marker, and they are the rivers the flood warnings follow. Parks as a sparse dot pattern stay unobtrusive. | `08b`, `08c` | Supports D-12. Parks-as-pattern is a workable default; revisit under a field, where parks compete with the background. |
| S9-1 | The credit line fits in the bottom row of a 12-row map, right-aligned, using 30 of 69 columns. It covers map content in that row. | `09-*` | Workable default for D-25. Fading it after a few seconds remains an option the guidelines allow. |
| S10-1 | **Automatic colour downgrade fails for fields.** At 16 colours, six of the seven muted ramp steps collapse to one grey. At 256 colours the coolest step lands on the same dark grey as water, so cold air looks like a lake. | `10a`, `10b` | **The library must know the colour depth and own a ramp per depth.** "Emit truecolor and let the host downgrade" (AI-3 §6) does not hold for overlays. Decides part of X-3 for PLAN. |
| S10-2 | **Purpose-built ramps work at both depths.** Exact palette colours, with line work chosen black or white per cell by background brightness: all seven bands distinct, every label readable. The 16-colour result is garish but fully legible; exact hues follow the user's terminal theme. | `10c`, `10d` | D-23's promise of legibility at four colour depths is achievable. It adds: a depth hint from the host; per-depth ramps; per-cell contrast selection. S1-3's "lines lighter than the ramp" becomes "lines contrast with the cell". |

### Precipitation and radar — the most common view (added at HUM LEAD's direction)

> "show precipitation specimen - that w/radar will be the most common use-case for maps"

| ID | Finding | Evidence | Consequence |
|---|---|---|---|
| S11-1 | **A lattice of forecast points is not a usable precipitation source.** A 10×10 lattice across the Southeast returned 7 non-zero points of 100, one of them an isolated 34 mm cell, while radar over the same area at the same time showed a rain shield several states wide. Convective rain is far patchier than a ~100 km lattice can sample. | live probe, 2026-09-18; `11a` | **For precipitation the radar image is the data.** The scalar-grid shape still serves smooth fields (temperature, pressure, forecast accumulations from a dense model grid), but the everyday precipitation view is a placed image, not a grid. Raises the importance of the image shape (D-14 shape 4) and of S6-4. |
| S11-2 | **Radar reads well with no colour as block shades.** Intensity read back from the image by hue, drawn as `▒ ▓ █` for light, moderate and heavy-or-worse, with roads removed and water as outline only: the rain shield, the heavy cores and the place marker inside the rain are all legible at 149×38 and at 69×12. Block shades are visibly different from braille border lines. | `11b-*` at both sizes | Radar has a workable no-colour form. **This withdraws the limit recorded under D-34** that placed images would be reported as needing colour. |
| S11-3 | **Contours fail for patchy data.** Contouring the same radar produces a litter of fragments. | `11c` | Contours suit smooth fields; shades suit patchy ones — as in printed cartography. The no-colour form is chosen by the *kind* of data, not once for all overlays. |
| S11-4 | **Why shades failed for temperature (S3-1) and work for radar.** Temperature covers every cell, was drawn over a full road net, and used dot-like light glyphs (`·` `:`). Radar is absent over much of the map, was drawn over a thinned basemap, and uses only block shades. | `03a` against `11b` | Refines S3-1: the failure was the glyph choice and the clutter, not shading as such. |
| S11-5 | The very lightest returns (blue and cyan) were dropped as clutter; light rain is drawn `▒`. A lighter `░` for light rain would make the heavy cores stand out more. | `11b` | A tuning choice for PLAN, from a further specimen. |

### Specimens 12–15 — owed after red-team round 1

Produced because the round found claims resting on renderings that did not exist. **Reviewed by HUM LEAD on the review page (D-42):** 13, 14 and 15 seen and fine; 12 — "Better; but the motorways only still looks a bit too much", acceptable as a renderer the user opts into, provided they can switch back to braille or remove roads.

| # | Files | What it shows | From |
|---|---|---|---|
| 12 | `12a-block-motorways-only-*`, `12b-block-no-roads-*` | The block renderer with the road net thinned: motorways only; no roads at all. Both map sizes. | HUM LEAD's concern on group D (D-41); RS-22 |
| 13 | `13a-alerts-no-colour-*`, `13b-wind-no-colour-149x38`, `03b-no-colour-digits-69x12` | Feature overlays with no colour: alert areas hatched (`╱`) and labelled; wind as arrow plus speed in km/h. The value lattice at the small size, which a ruling had claimed and nobody had rendered. | Accessibility A-2; phase lens PM-1 |
| 14 | `14-safe-ramp-*` | Temperature on a brightness-ordered, colour-vision-safe ramp (the viridis family), with line work chosen black or white per cell. | Accessibility A-3, A-4 |
| 15 | `15a-radar-recoloured-*`, `15b-radar-themed-amber-149x38`, `15c-…256-colour…`, `15d-…16-colour…` | Radar re-coloured by the library from intensity classes (ruling D-36): a brightness-ordered default; a single-hue amber ramp as a host theme might supply; the same at 256 and 16 colours. | Phase lens PM-5; ruling D-36 had been ratified from prose |

| ID | Finding (coordinator's assessment; the specimens were seen by HUM LEAD — D-42, D-54 — the wording of these findings was not separately reviewed) | Evidence |
|---|---|---|
| S13-1 | *[Corrected 2026-09-18 after red-team round 2. This finding previously said "the place marker's position inside or outside is readable". **The specimen it cited contained no place marker at all**: the hatch was written first and the marker was refused the cell. The claim was made without looking. Both files were redrawn, the marker now over the hatch.]* With no colour, a hatched interior plus a label makes an alert area distinct from braille roads and rivers at both sizes. **Whether the place is inside is not always readable from the frame**: at 69×12 a column is 1.6 km and the marker sits at the tip of the warning area, so the picture cannot settle it — which is what the description-as-data requirement (FR-29, D-52) is for. The short label was `[FW]`, which in the weather service's own codes means *fire weather*; it is now a plain word. The corrected files also carry the first scale mark any specimen has shown (FR-33). | `13a-*` (regenerated) |
| S13-2 | Arrow plus speed (`←14`, `↙29`) carries wind strength with no colour, like a station plot. It needs the road net removed to stay readable. | `13b` |
| S15-1 | Re-coloured radar reads well in the one browser look taken: a teal light-rain shield, green moderate, amber heavy cores with line work switched to black; white map lines read cleanly on the teal. Intensity was read back **by hue**, a rough method; no exact colour table for this provider has been built. | `15a-149x38` |
| S15-2 | The note line in the 15-series headers said the image was "placed as-is" when it was re-coloured. Corrected in `15a-149x38`; `15b`–`15d` and `15a-69x12` still carry the old note. **Every radar specimen fetched radar as tiles — an input shape deferred past v0.1.0 (D-44).** The image shape v0.1.0 does have is not yet demonstrated with radar (NFR-17). | `15*` headers |
| S15-3 | **Found by red-team round 2:** the re-coloured radar ramp fails the colour-vision test at 256 colours — "light" and "moderate" nearly tie in brightness (relative luminance 0.191 against 0.198) and under protanopia "moderate" and "heavy" differ by 5.0 against a proposed threshold of 10. 256 colours is in v0.1.0. A passing ramp per depth is owed in PLAN (FR-16). | reviewer's measurement; not re-derived by the coordinator |
| S12-1 | Thinning helps the block renderer but does not make it braille's equal. Braille is the default; the block renderer is opt-in with no roads by default (D-42). | `12a`, `12b` |

### Specimen 16 — a diverging temperature ramp (approved by ruling D-53)

Made to see whether the weather convention — cold is blue, hot is red — can be kept safely. Files: `16-diverging-ramp-149x38`, `16-diverging-ramp-69x12`, `16-diverging-ramp-256-colour-149x38`. The ramp is the published seven-class blue–pale–red scheme from the ColorBrewer family, with line work chosen black or white per cell. **Reviewed by HUM LEAD on the review page (D-54): "Seen, Looks fine".**

Measured with the same test the accessibility review used (relative luminance; colour difference between steps under simulated protanopia, deuteranopia and tritanopia):

| Measure | Result |
|---|---|
| Brightness | 0.128 · 0.358 · 0.759 · **0.930** · 0.757 · 0.374 · 0.103 — rises to the pale midpoint, then falls: ordered on each side, as FR-16 allows for a diverging ramp |
| Cold arm against its hot twin (steps 1~7, 2~6, 3~5) | protanopia 60.8 / 55.4 / 19.4 · deuteranopia 83.9 / 69.8 / 23.7 · tritanopia 105.5 / 91.0 / 26.3 — blue and red are never confused |
| Closest adjacent pair | protanopia **9.4** · deuteranopia 11.5 · tritanopia 13.0 — one marginal case against a proposed threshold of 10 |
| Contrast with per-cell line work | 5.9 · 8.2 · 16.2 · 19.6 · 16.1 · 8.5 · 6.9 — every step clears 4.5:1 |

| ID | Finding (coordinator's assessment; the specimens were seen by HUM LEAD — D-54 — the wording of these findings was not separately reviewed) | Consequence |
|---|---|---|
| S16-1 | A diverging blue–pale–red ramp keeps the weather convention and passes the colour-vision test on every measure but one marginal adjacent pair, which a tuned ramp would fix. | A safe default for temperature need not give up "cold is blue, hot is red". |
| S16-2 | It needs a **meaningful midpoint**. In this specimen the midpoint is simply the middle of the data's range (about 21 °C); a real default needs a stated one — freezing, a seasonal normal, or a host-supplied value — and a legend that labels it. | A contract question for scalar grids: the midpoint is data the host supplies. *[Overtaken by D-62: for temperature the midpoint is freezing and the breaks are the library's; the host supplies neither.]* |
| S16-3 | The pale middle bands are the brightest cells on a dark terminal, and line work there is black, not white. HUM LEAD saw group L and found it fine (D-54). *The legends in specimens 14 and 16 as first filed showed the previous ramp's colours, not the ramp drawn (red-team round 2); regenerated, with legend text chosen black or white per swatch. HUM LEAD's D-54 look was at the uncorrected legends.* | Re-review of the corrected files is asked for in round 2. |

### Specimen 17 — M1 scenario 2: an alert defined by a zone shape, the place close to the edge (PLAN, 2026-09-19)

Made in PLAN, because the scenario HUM LEAD called "the one to watch" had never been rendered (red-team round 2, R2-BZ-A3). A live Heat Advisory from the saved feed carried no polygon of its own, as about nine alerts in ten do not; its shape is its forecast zone — Grayson County, Kentucky, 527 vertices, much of it a river boundary — fetched from the weather service. The place was put 0.9 km inside the wiggliest stretch of that edge. Files: `17a-*` (truecolor, both sizes), `17b-*` (no colour, both sizes), `17-scenario2-answer-key.json`.

**The answer key**, computed by a separate script that shares no code with the renderer (D-43, D-67): **inside**; nearest edge **0.9 km to the north** (bearing 342°). Distance from the edge in cells: **0.56 at 149×38, 0.26 at 69×12** — so by D-67 the correct reading of the frame is **"on the edge" at both sizes**, and the description (M1b) must say "inside, 0.9 km, north".

| ID | Finding (coordinator's assessment; **seen by HUM LEAD on the review page and found good — D-77**) | Consequence |
|---|---|---|
| S17-1 | In the no-colour frames the zone draws as a hatched area with a plain-word label, its river edge follows the basemap's river, and the marker sits on the area's top edge at both sizes — which is what the key says a frame can show here. | M1a for this scenario is "on the edge", and the frame does not contradict the key. |
| S17-2 | **A cell is twice as tall as it is wide, so an edge to the north or south is resolved only half as well as one to the east or west**: 1.6 km a row against 0.8 km a column at 149×38. A place 0.9 km from a northern edge is under one cell away even at the large size. Nobody had stated this before; D-67's "under one cell" must name which dimension, and the key does — by the bearing to the edge. | The description (M1b) matters at the large size too, not only at 69×12. Carried into the implementation plan's definition of the key. |
| S17-3 | Of fourteen real zones sampled for this specimen, the largest had 527 vertices and most had 50 to 400. The scenario table's "900 to 14,000" describes the hard cases DISCOVER measured, not the usual one. | The 14,000-vertex case — where simplifying could flip "inside" (risk RS-19) — is **still unrendered**; it belongs to the worst-case test (NFR-3), and a specimen of it is still owed. |

### Specimens 18 and 19 — M1 scenarios 6 and 7: point hazards, and a track passing the place (PLAN, 2026-09-19)

The last two unrendered scenarios. The place is Great Falls, Montana. **Scenario 6:** a real earthquake (magnitude 3.6, from the public earthquake feed, fetched 2026-09-19) and a **synthetic** fire detection. **Scenario 7:** a **synthetic** storm track. Both map sizes, with colour and without; the headers say what is real and what is not. Files: `18a-*`, `18b-*`, `19a-*`, `19b-*`, `18-19-scenarios6-7-answer-key.json`.

**The answer key** (separate script): earthquake **6.0 km north** (343°) — 3.4 cells away at 149×38, 1.6 at 69×12; fire **43.3 km south-east**; the track's closest approach **31.9 km, passing to the north-west**. The cell figures use the mean of a cell's two dimensions and are approximate.

| ID | Finding (coordinator's assessment; **seen by HUM LEAD on the review page and found good — D-77**) | Consequence |
|---|---|---|
| S18-1 | With no colour, the three kinds of point read apart by glyph alone — ◉ place, ◆ earthquake, ● fire — and the label carries the magnitude. At 69×12 the earthquake sits one row up and one column left of the place, which matches "very close, to the north". | Glyph differences are enough for points (FR-18a). All three glyphs passed the terminal card. |
| S18-2 | **The first placement of the fire point fell off the frame**: 36 km south is seven rows, and a 12-row map centred on the place shows only six below it. M1's guard — the place and the condition must share one frame — is therefore a demand on **the view the host chooses**, and the contract has no intent for "fit these things in" (FR-24 has only "fit the world"). | Put to HUM LEAD and ruled: **D-76** adds an intent that fits the view to named places and overlays, in v0.1.0. |
| S19-1 | **With no colour, a track is a braille line exactly like a river.** Drawn two dots thick and labelled, it is still only the word "track" that tells it from the river a few cells away. FR-18a already asks for "a distinct stroke family"; this shows why. | Owed: a stroke for line features that no basemap line uses — dashed, or beaded — and a specimen of it. |
| S19-3 | **A dashed stroke settles S19-1.** Redrawn with five dots on and four off, a little heavier than a road, the track reads as a different kind of line from the unbroken river beside it at both sizes, before the label is read (`19c-*`, no colour). No basemap line is dashed, so the family is free. | The no-colour stroke for line features is a dashed one (FR-18a); the dash lengths — five dots on, four off, two dots thick — are fixed in the constants document. Seen by HUM LEAD and found good (D-77). |
| S19-2 | Which side the track passes, and roughly how far, reads from the frame at both sizes (about 12 columns west of the place at 69×12; the scale mark gives 20 km for 8). | Scenario 7 is answerable from the frame (M1a) once S19-1 is dealt with. |

### Specimen 20 — a painted ground, light and dark (PLAN, 2026-09-19)

Owed by ruling D-64: HUM LEAD had never seen a painted ground. The same view as specimen 4 (Fort Wayne, live alert areas), four ways. Every drawn cell's contrast was **measured** from the colour files with the WCAG formula, not judged by eye.

| File | Ground | Drawn cells under 3:1 |
|---|---|---|
| `20c-unpainted-ground-149x38` | Not painted — as every earlier specimen | **84.7% on a white terminal** · 0.3% on a black one |
| `20b-painted-ground-dark-149x38` | Painted dark (16·22·28), dark style | 0.3% (four park cells on an alert tint) |
| `20a-painted-ground-light-*` | Painted light (245·245·240), a first bright style with coloured lines | 16.2% — white labels left over from the dark style, water lines on water, coloured lines crossing an alert tint |
| `20d-painted-ground-light-best-contrast-149x38` | Painted light, every line and label drawn black or white by the higher contrast (FR-16) | **0%** |

| ID | Finding (coordinator's assessment, measured from the colour files; **seen by HUM LEAD on the review page and found good — D-77**) | Consequence |
|---|---|---|
| S20-1 | **The number behind D-64:** drawn as every earlier specimen was, 85 cells in 100 fall below 3:1 on a white terminal. Painted, the same map holds on either ground. | The painted default is justified by measurement, not only by argument. |
| S20-2 | A bright style with **coloured** lines is not automatically safe: one cell in six failed, mostly where a coloured line crosses a tint or sits on water. Applying FR-16's rule — compute both, take the higher — to every line and label brought it to zero, at the price of all line work being black or white. | The bright style needs the per-cell rule everywhere, or coloured lines checked against every background they can cross. A choice for the style work in the implementation plan; how it looks is for HUM LEAD. |
| S20-3 | At 256 and 16 colours the painted ground cannot match a host's truecolor background exactly, so the map may sit as a visible panel. Not rendered here. | Still owed with the 16-colour specimen (D-59). |

### Specimen 21 — a candidate for the temperature preset: an absolute scale (PLAN, 2026-09-19)

Owed by ruling D-62, which asked for a longer ramp than specimen 16's seven steps and said this had been solved many times by other weather maps. **The coordinator had doubted that a long scale could stay colour-vision-safe, and said so when D-62 was put. That doubt was wrong, and this is the evidence.** Files: `21a-*` (regional, both sizes), `21b-*` (the same, using only colours from the 256-colour palette), `21c-*` (continental), `21-temperature-scale-candidate.json` (the colours, the breaks and the checker results).

**The scale:** 17 classes — below −30 °C, fifteen classes 5 °C wide from −30 to +45, and +45 and above — with the pale break at freezing: seven cold classes, ten warm. Built by starting from an established cartographic scale (ColorBrewer's red–yellow–blue, which itself passes at its full 11 classes) and searching near it for the colours that keep every adjacent pair furthest apart under all three kinds of colour blindness.

| Measured (worst adjacent pair; proposed threshold 10) | Normal | Protanopia | Deuteranopia | Tritanopia | Ordered about the break | Text contrast at least |
|---|---|---|---|---|---|---|
| The 17-class scale, truecolor | 16.0 | 13.7 | 13.9 | 14.3 | yes | 4.6 |
| The same, 256-colour palette only | 11.1 (the worst of all four) | | | | yes | 5.3 |
| For comparison: ColorBrewer red–blue, 11 classes | 11.5 | **9.4** | 11.5 | 13.0 | yes | 5.6 |
| For comparison: simply stretching that to 15 classes | 8.2 | **6.8** | 8.2 | 9.2 | yes | 4.6 |
| For comparison: a many-hued scale in the style of broadcast maps, 15 classes (approximate colours) | 15.9 | **6.3** | 10.2 | 11.5 | **no** | 4.7 |

A search for the ceiling found ordered, safe scales of up to **21 classes** in truecolor (worst pair 12.2).

| ID | Finding (coordinator's assessment, measured from the colour files; **seen by HUM LEAD on the review page and found good — D-77**) | Consequence |
|---|---|---|
| S21-1 | **HUM LEAD was right and the coordinator was too cautious:** a long, absolute, colour-vision-safe temperature scale exists. Stretching a short scale fails; choosing the colours for distinctness passes with room to spare. | D-62 and D-53 do not conflict. 5 °C bands across −30 to +45 °C are available. |
| S21-2 | The many-hued scale common on broadcast maps fails on three counts: it is not ordered by brightness, its closest adjacent pair is 6.3 under protanopia, and two steps that are **not** neighbours (yellow-green and orange) differ by only 4.4 under deuteranopia. | "Established" is not enough; the checker decides. The preset draws on the cartographic family, not the broadcast one. |
| S21-3 | On the evening sampled, the regional view shows three or four bands and the continental view five or six — the flat day HUM LEAD accepted in D-68, now seen. | As ruled. The description carries the gradient. |
| S21-4 | The 256-colour version passes narrowly (11.1) and its hues zig-zag in two places, because that palette is coarse. | Seen by HUM LEAD and found fine (D-77, and the retuned version D-89). If it reads badly, the 256-colour preset can use fewer, wider classes — a preset may differ by depth (FR-17). |
| S21-5 | *(Written when the checker compared neighbouring bands only; since D-88 it compares every pair — see S21-6.)* This measures whether bands can be told apart on the map. Matching one band to its swatch in a 17-entry legend is a harder task that this measure does not test. | The legend states values, and the description gives the value at the place; the map is not the only way to read a number. |
| S21-7 | **Found by red-team round 2 (P2-PRD-1):** the scale was never checked against a ground. On the painted light ground its two bands at freezing are 5.0 and 5.5 from the ground (threshold 10) — `21e-*` shows it. | **Ruled D-91:** a light-ground variant. Three bands change (−10 to +5 °C), hue kept on its side of freezing; every pair still 12.2 apart (11.1 at 256), every class at least 13.8 from the ground. `21d-*`, both sizes. **`21d` and `21e` are colour tests, not forecasts:** the evening's real temperatures are drawn 20 °C colder than they were, so that the freezing bands appear; their headers say so. The variant also passes on the dark ground (12.1). Seen by HUM LEAD and found fine (D-93). **Their legends are 20 °C off too** — the band coloured for −5 to 0 °C sits under "15.0" — because the throwaway renderer's legend follows the shifted scale; D-93 judged the colours, not the legend. |
| S21-8 | **Found by red-team round 3 (P3-B-3):** "ordered" had been checked as relative luminance only. Under protanopia the truecolor scale got lighter from class 14 to 15. | One colour moved by 1.9 — below what the eye can tell — fixes truecolor under every kind; the ramp file carries it and says the rendered specimens use the earlier colour. At 256 colours three small inversions remain and are recorded. Listed for HUM LEAD in the Plan of Record. |
| S21-6 | **Found by the PLAN red-team (PL-AX-2):** the scale as first filed passed for neighbouring classes only. Two warm classes that are not neighbours — +5 to 10 °C and +15 to 20 °C — were 7.5 apart under deuteranopia (8.4 in the 256-colour version): the same flaw used above to fail the broadcast-style scale. One class was moved in each version; **every pair** now scores at least 12.2 (11.1 at 256), against the ruled threshold of 10 (D-88). The files are re-rendered. | The checker tests every pair, not only neighbours. HUM LEAD's D-77 look was at the first version; the retuned version was seen and found fine (D-89). |

### Specimen 22 — radar by the first release's own path (PLAN, 2026-09-19)

Every earlier radar specimen fetched radar as tiles, a shape deferred past v0.1.0 (S15-2, risk RS-23). This one uses the shape v0.1.0 has: **one image for one bounding box**, fetched from the Iowa Environmental Mesonet with one keyless request, read through the provider's **published 256-entry colour table** into six classes by reflectivity (10, 20, 30, 40, 50, 60 dBZ), and re-coloured by a new six-class ramp. The weather is real: a rain shield over Indiana on 2026-09-19 with Fort Wayne at its edge. Files: `22a-*` (truecolor on a painted dark ground, both sizes), `22b-*` (256-colour palette only), `22c-*` (no colour), `22-radar-ramp-candidate.json`.

| Measured | Result |
|---|---|
| Pixels of a national image checked against the published table | 20,959 visible, **0 unmatched** |
| Samples unmatched while drawing this view | 0 |
| The new ramp, worst pair under any kind of colour blindness (threshold 10) | **21.5** in truecolor · **23.8** within the 256-colour palette · ordered by brightness at both |
| Drawn cells under 3:1 | 0% in `22a` and in `22b` |

| ID | Finding (coordinator's assessment, measured from the files; **seen by HUM LEAD — "actually like it" — and the heaviest-in-cell rule agreed, D-78**) | Consequence |
|---|---|---|
| S22-1 | **The first release can get radar in, and exactly.** One request, one image, the provider's own table, no guessing by hue. | Risk RS-23 is closed. The example (NFR-17) is this path. |
| S22-2 | The ramp that failed at 256 colours (S15-3) is replaced by one that passes at both depths with a wide margin. | The radar preset's colours are settled as a candidate; liked by HUM LEAD (D-78). |
| S22-3 | Each cell takes the **heaviest** of eight samples inside it, so a small heavy core is not averaged away at coarse sizes. With no colour, the rain shield, its heavier bands (▓) and the dry slot around Fort Wayne all read from block shades alone. | This is a first answer to the "radar resampling rule" carried from DISCOVER: **the heaviest in the cell, not the mean** — a safety choice, since under-stating rain is the worse error. Stated in the implementation plan. |
| S22-5 | **Found by red-team round 2 (P2-PRD-2):** the light-ground ramp of `22d`, which HUM LEAD had seen and found fine (D-89), got **lighter** from class 2 to class 3 under every kind of colour vision — "darker means heavier" broke in the middle of the scale. The coordinator's checker tested differences and never the order. | Corrected as `22f-*`: an olive green, a burnt orange, a crimson. Darker at every step under all four kinds of vision; every pair at least 17.7 apart (was 16.0), 14.4 at 256 (was 12.6); every class at least 18.7 from the ground. `22d` is kept as the record of what was approved. Seen by HUM LEAD and found fine (D-93). |
| S22-4 | **Found by the PLAN red-team (PL-AX-1):** the ramp above works on a dark ground only. On the painted light ground of specimen 20 its **heaviest** class is 3.1 from the ground and 1.01:1 — a 60 dBZ core would read as a dry hole (`22e`, kept as the record of the defect). The checker had no rule about the ground. A second ramp, **darker means heavier**, is chosen when the ground is light (`22d`): every pair at least 16.0 apart (12.6 in the 256 palette), every class at least 18.7 from the ground. | The radar preset is two ramps, chosen by the ground's luminance as the style is; the checker gains a ground rule (D-88). `22d` seen by HUM LEAD and found fine (D-89). *The light-ground figures in this row are the first version's; S22-5 supersedes them.* |

### Specimen 23 — a 16-colour hint (PLAN, 2026-09-19)

Owed by ruling D-59, which HUM LEAD made without ever having seen a coloured basemap beside a colourless overlay. The radar view of specimen 22 with a 16-colour hint: line work, labels, the marker and a painted black ground use the 16-colour palette; the radar, which needs a ramp, is drawn in its no-colour form — block shades. Files: `23a-*`, both sizes.

| ID | Finding (coordinator's assessment; **seen by HUM LEAD and found good — D-79**) | Consequence |
|---|---|---|
| S23-1 | **Converting the style's colours to 16 does to the basemap what it does to a ramp** (S10-1): the first attempt came out in white and grey only, because the dark style's pale blues and greys all land on "white". Choosing the basemap's colours *from* the sixteen — bright cyan rivers, bright blue water, green parks, white roads, bright white borders and labels, a bright yellow marker — restored the distinctions. | The 16-colour depth needs its own small basemap palette, picked by hand, as the ramps do. It is a handful of tokens (D-63). |
| S23-2 | With that palette the frame carries seven distinct colour codes and the radar's grey shades sit among coloured map lines without borrowing their colours. Whether the mix reads well, or whether a plain no-colour frame would be calmer, is a judgement of looks. | Seen and found good (D-79); the fall-back is not needed. |
| S23-3 | At this depth the terminal decides what "bright cyan" really looks like, and it cannot be asked. Any contrast figure here would be against an assumed palette, so none is given. | As FR-20 now says: checks at 16 colours are indicative only. |

### Specimen 24 — how a cell's colour is chosen: line priority against upstream's majority vote (PLAN, 2026-09-19)

Made because HUM LEAD, asked to rule between the two (PLAN red-team PL-Q3), answered "C" — decide from a specimen. Parity row P-08 says a cell takes the majority colour among its lit dots, ties broken by the eight neighbouring cells; every earlier specimen was drawn by a different rule — the highest-priority line in the cell wins — which the coordinator had used without reading the row. The same two views of Fort Wayne, a regional one and a city one, drawn both ways on a painted dark ground. Files: `24a-priority-*`, `24b-vote-*`.

| Measured | Regional view | City view |
|---|---|---|
| Braille cells drawn | 1,290 | 1,220 |
| Cells whose colour differs under the vote | 41 (3.2%) | 62 (5.1%) |

| ID | Finding (coordinator's assessment; the difference was counted; **seen by HUM LEAD, who chose 24b, the majority vote — D-83**) | Consequence |
|---|---|---|
| S24-1 | The two rules agree on about 95 to 97 cells in 100. They differ only where two kinds of line share a cell — a river under a road, a border along a highway. | Whichever is ruled, the map's overall look barely moves; the choice is about those crossings. |
| S24-2 | Overlays are untouched by either rule: they sit above the basemap in the compositing order (FR-12). | The ruling concerns basemap lines only. |

### Specimen 25 — temperature with no colour: isolines at the preset's absolute breaks (PLAN, 2026-09-19)

Owed since DISCOVER as assumption A-5 — "a proper contouring pass yields clean lines at terminal resolution" — and named by the PLAN red-team as an artefact not produced (PL-PM-1). A first attempt with the throwaway renderer's old contour mode gave broken staircases of box-drawing characters and was discarded. This one classifies the field **at every braille dot** and lights a dot wherever the class changes against the dot to its right or below: continuous braille isolines, every 5 °C — the temperature preset's own breaks — with a value label along each. Files: `25a-*` (continental, both sizes), `25b-*` (regional).

| ID | Finding (coordinator's assessment, from the text frames; not yet seen by HUM LEAD) | Consequence |
|---|---|---|
| S25-1 | **Assumption A-5 holds.** Classifying per dot yields unbroken isolines that bend smoothly at terminal resolution, with no special cases; the 10 and 15 °C lines cross the continental view as single curves and carry their values. | The no-colour form of a scalar field (D-35) is per-dot classification at the type's breaks; it needs no box-drawing characters, so NFR-8's closed list loses none and gains none. |
| S25-2 | At the regional view the evening's range sits inside one or two bands, so there is one isoline or none — the flat day of D-68, as it looks with no colour. | As ruled: the description carries the gradient. |
| S25-3 | **An isoline is a plain braille line, and so is a state border.** With no colour the two are told apart only by the isoline's value labels and its curve. | The same question specimen 19 raised for tracks. A lighter dotted stroke for isolines is the obvious answer; it is not drawn here. Owed in BUILD's style work, for HUM LEAD's eyes before the golden frames are approved (plan task 14.16). |

### Known defects in the PLAN specimens' own headers

Found by the PLAN red-team (PL-DQ-15). The files are evidence and are left as generated; read them with these corrections. **"Re-coloured by the library"** in the headers of 22a, 22c and 23a means "by the throwaway renderer, in the way the library is designed to" — no library exists yet; and in 22c and 23a the radar has no colour at all: it is drawn as block shades. **23a** says the ground is "PAINTED 0,0,0"; at 16 colours that is palette entry 0, whose real colour is the terminal's. **17a and 17b:** the place is not a real town; it is a point chosen 0.9 km inside the zone's edge. **Specimens 21** as first committed used a scale that failed for non-neighbouring classes; the files now in the tree are the retuned ones.

### Cross-cutting

1. **One rendering core carried every specimen.** A cell is a glyph, a foreground and a background; features, grids, vector grids and images all reduce to writing those three. The five input shapes of D-14 do not need five renderers.
2. **The background belongs to areas, the foreground to lines and glyphs, and there is a priority order among areas**: water, then placed images and fields, then tints. Stating that order is most of the overlay compositing design.
3. **Legibility is a contrast budget.** S1-3, S6-2 and S10-2 are the same rule seen three times: whatever owns the background is held to a brightness band, and whatever is drawn on it is chosen to contrast.
4. **The basemap is not fixed; it yields.** It thins under vector overlays (S5-1), under the block renderer (S2-1) and at small sizes; water gives up its dots under a field (S1-2). The D-24 styles are therefore a small family of profiles, selected by renderer, size and what is on top.
5. *[Corrected 2026-09-18 after red-team round 1. This item previously read "RS-3 can close… legible at both of the first host's map sizes, in both renderers, at every colour depth". That overstated the evidence.]* **What the specimens actually cover:** under the braille renderer in truecolor — a field, alert polygons, wind and radar, at both map sizes; under the block renderer — a field only, in truecolor, and at 69×12 the roads swamp it (S2-1); at 256 and 16 colours — one view, one size, braille, temperature only. HUM LEAD has ruled on specimen 1 (D-32). ~~RS-3 stays Medium.~~ *[2026-09-18: after HUM LEAD reviewed every group on the review page (D-40, D-41, D-42, D-54), RS-3 is **Low for the braille renderer** — see the risk assessment, which is the authority. HUM LEAD's looks were in a browser, not HUM LEAD's terminal; a terminal look at the `.ans` files is still owed before PLAN exit.]*

### Limits of this evidence

**Known defects in the specimen files themselves** (found by red-team rounds 1 and 2; the files regenerated in round 2 — `13a`, `14`, `16`, `15a-149x38` — are corrected, the rest are not; **specimen 14 was corrected a second time after round 3**, which found its text colour still chosen by a brightness threshold — white on the teal band at 3.82:1 where black gives 5.49:1. Both contrasts are now compared and the higher taken; measured over every background and text pair in both files, the lowest is 5.49:1): header lines read "256-bit colour" and "16-bit colour" where they mean 256 colours and 16 colours, and the no-colour `.txt` files say "24-bit colour". The default temperature ramp used throughout is **not colour-vision-safe**: its brightness rises to the amber step and then falls, so under red-green colour blindness the 20 °C and 27 °C bands nearly coincide (red-team A-3). Upstream's marker flash rate (3.33 Hz, above the three-per-second accessibility threshold) was not exercised.

Throwaway code; nothing here measures speed or memory. Browser approximations draw braille differently from a terminal. One region, one evening's weather, a dark terminal background; a light-background terminal was not tried. Line simplification and polygon clipping (D-16) were not exercised — the live alert polygons had 5 to 21 vertices. Radar resampling and the quadrant-glyph width question are open.

## Specimen 26 — the first map the library drew (BUILD, milestone M-A)

Unlike every specimen above, these were not drawn by the throwaway program of PLAN. They are what the library's own three public calls return - create, settle, render - from the embedded tiles alone, with no network.

| File | What it is |
|---|---|
| `26-first-map-149x38` | The whole world at the first host's large size. `.ans` carries the colours; `.txt` is the same frame with the colour sequences taken out |
| `26-first-map-69x12` | The same at the first host's small size: a map wider than the world, with ocean beyond the world's edge |

| # | Finding | Consequence |
|---|---|---|
| S26-1 | **A zoom-0 tile carries every place three times, a world apart, in its buffer.** With upstream's rule of trying each vertex in turn, the first copy - a world to the east - took each name, and on the small map "Asia" was drawn west of the Americas. | Upstream's other rule, that an anchor must be inside the world as well as the canvas (P-33), had been built by halves. Now whole, with a test through the public calls |
| S26-2 | A fitted world on a small rectangle sits **below zoom 0** (about -1.8 at 69 by 12). The style's rules "from zoom 0" drew nothing there, and the first frame through the public calls was blank but for the credit line. | A rule that starts at zoom 0 has no lower bound. The never-blank test now runs through the public calls as well |
| S26-3 | Region borders, rivers and roads at world scale were clutter. | The built-in style's finer roles start at the zoom where they can be read (L2 Style). A first setting: HUM LEAD's to tune |
| S26-4 | At the world's western edge a short vertical line is drawn near Antarctica: the polygon's own edge along the antimeridian, which lies just inside the tile and so is not taken for the tile's border. | **Fixed (D-107):** the edge starts one unit inside the tile, so an edge now counts as the border when both ends are within 1/2048 of the extent of the side. The frames here were drawn again after the fix |

## Specimen 27 — the alert colours, as swatches (BUILD, task 08.10)

`27-alert-colours-swatch.ans`: for each ground, dark and light, and each depth, truecolor and 256 colours, the five severities - each an outline drawn on the ground around its tint, with its name in the text colour the foreground rule picks - beside a patch of water. Show it with `cat`. It is a swatch, not a map: the look the plan asks of HUM LEAD (task 08.23) is at alert areas drawn on a map, once overlays draw.

| # | Finding | Consequence |
|---|---|---|
| S27-1 | Pale tints on a light ground cannot be told apart under simulated colour vision: the first search ended below the threshold. | On a light ground the scheme turns over: dark outlines, mid-tone tints, black text |
| S27-2 | The truecolor sets converted to the 256-colour palette fail on both grounds - one tint 4.6 from water, two tints 6.6 apart. | The preset has its own 256-colour sets, searched within that palette, as the ramps do (S10-1) |

## Specimen 28 — alert areas on a map (BUILD, task 08.23)

Four frames, 149 by 38, drawn by the library's own packages from the embedded tiles: five alert areas over the central United States, one of each severity - extreme, severe, moderate, minor, unknown - each its tint, its outline and its label. `28-alert-areas-dark-truecolor`, `-light-truecolor`, `-dark-256`, `-light-256`; the `.txt` beside the first is the same frame without colour. Show them with `cat`. **This is the look task 08.23 asks of HUM LEAD before any reference frame is frozen (RS-26).**


## Specimen 29 — radar under a warning (v0.2.0 DISCOVER)

No earlier specimen drew radar and an alert together: 22 is radar alone, 28 is alerts alone. This one uses real weather captured 2026-09-23 14:28Z: a Flash Flood Warning (Severe) over Lincoln and Putnam counties, West Virginia, with heavy rain inside it, and IEM NEXRAD n0q reflectivity read through the provider's published 256-entry table. Hamlin is the selected place. The frames come from the library's public calls, with OpenFreeMap tiles, at 69 by 12 and 149 by 38; `.ans` is truecolor on a painted dark ground and `.txt` has no colour.

| File | What it is |
|---|---|
| `29a-radar-under-warning-as-built-*` | The library as built (FR-12): the tint is drawn over the image, so the radar inside the warning is hidden |
| `29b-radar-under-warning-radar-wins-*` | **A one-line patch to a scratch copy, not the library:** the image is drawn over the tint. It shows the alternative and does not propose an implementation |
| `29c-warning-without-radar-control-*` | The same warning with no radar: the control |

| # | Finding | Consequence |
|---|---|---|
| S29-1 | In colour, 29a draws the warning as a solid block, and the heaviest rain on the map, which falls inside that block, cannot be seen. In 29b the rain shows through; the tint survives only on the label's cells and in dry cells, so on a wet day the outline alone carries the warning. | A HUM LEAD ruling for v0.2.0 |
| S29-2 | **With no colour, the radar's block shades replace everything else in their cells**: the warning's outline, hatch and label, the place marker, the scale mark and the credit line. 29c shows all of these drawn when the radar is absent. The same happens in both orders. | A defect against FR-18a (outlines go over everything beneath them), not a ruling. It becomes a v0.2.0 requirement |
| S29-3 | At 69 by 12 the Hamlin marker and name are not drawn even in the control, although the fit includes Hamlin. | Recorded, not diagnosed |
| S29-4 | The county picture (596 by 546) was refused at the fixed 250,000-byte image cap and had to be cropped to 447 by 348. | Evidence for C-3 (wave 1, W1-A): a host cannot raise the cap |

**Added at HUM LEAD's request, 2026-09-23: a blend of the two orders.** HUM LEAD preferred 29b, and asked to see the warning's colour laid over the radar as a partial tint. `29d-*`, `29e-*` and `29f-*` blend it in at 20%, 35% and 50%, composited in linear light. They are made by a small scratch function, not by the library, and are `.ans` only: with no colour they are identical to 29b.

| # | Finding | Consequence |
|---|---|---|
| S29-5 | Measured with the library's own checker (D-88: the least Lab distance under four kinds of vision; floor 10). On a dark ground under the Severe tint, 20% keeps every tinted class at least 14.5 from every other plain class. At 35%, tinted heavy rain (class 5) is 9.3 from plain moderate rain (class 3); at 50% it is 1.8. Across all five severities, 20% is the only strength that clears 10 (lowest 11.7). | A fixed blend strength holds only on a dark ground, and only faintly |
| S29-6 | On a light ground the blend fails at every strength tried: tinted classes 3 and 4 come within 3.9 to 9.6 of each other even at 20%. | A per-class tint chosen by search, as the ramps were, is the only form that might hold, and on a light ground it may have no solution. An option for PLAN, not shown here |
| S29-7 | **Re-measured with both kinds of pair on both grounds (D-32)**, because S29-5 reported tinted-against-plain on the dark ground and S29-6 tinted-against-tinted on the light. Lowest distance across all five severities: **dark ground** — 20 %: 17.0 tinted/tinted, 11.7 tinted/plain (**both clear 10**); 35 %: 13.9 and 6.2; 50 %: 9.7 and 1.8. **Light ground** — 20 %: 3.9 and 1.9; 35 %: 2.6 and 5.0; 50 %: 1.6 and 4.0 (**nothing clears 10**). Raw output: `radar-loops/02-analysis/programs/output/blend-measure.txt`. | S29-5 and S29-6's conclusions stand: a fixed blend holds only on a dark ground at about 20 %. The light ground is D-27's search, with 29b's order as its fallback |

**Ruled (v0.2.0 D-14, `radar-loops/02-analysis/rulings.md`):** the tint blends over the radar, with the outline and label on top. Blend strength and colours are tuned in PLAN, held by the checker, and must pass on both grounds. HUM LEAD rules the light ground solvable by adjusting the radar ramps or the alert colours.

## Specimen 30 — an outbreak: overlapping warnings over heavy radar (v0.2.0 PLAN entry, OW-11, D-44)

A past severe day, taken from the Iowa Environmental Mesonet's archive: **27 April 2011, 21:00 UTC,
west-central Alabama**. That is IEM `n0q` archive radar (500×430 for one box) under the 20 NWS
warnings valid then and in view (17 tornado, 3 flash flood), with Tuscaloosa as the named place. Tornado
warnings are drawn as Extreme, flash flood as Severe. Drawn by the library's public calls with
OpenFreeMap tiles, at 69×12 and 149×38; `.ans` is truecolor on a dark ground, `.txt` has no colour.
The program and its inputs are in `radar-loops/02-analysis/programs/` (`ow11-render.go.txt`,
`inputs/ow11/`).

| File | What it is |
|---|---|
| `30a-outbreak-as-built-*` | The library as built: the tint over the radar |
| `30b-outbreak-blend-20pct-*` | The D-14 blend at 20 %, the only strength that passed on a dark ground (S29-7), by the scratch patch `spec29d-f-blend.patch` |

| # | Finding | Consequence |
|---|---|---|
| S30-1 | **As built, tornado tint covers 1,234 cells at 149×38 and hides the radar in all of them.** With the blend, the radar's classes show through inside the warnings (19 blended colours); bare tint is left only in the warnings' dry cells. | Confirms D-14 on the picture that matters most. |
| S30-2 | **Labels crowd out in an outbreak:** at 149×38 only 7 of the 17 tornado warnings and none of the 3 flash-flood warnings get a label. | Most warnings on a severe day carry no word. Severity must come from D-17's dash, and L-8.5's fallback and the description (L-13.5) carry the rest. PLAN should weigh the labelling rule for overlapping areas. |
| S30-3 | **The named place's name, "Tuscaloosa", is drawn at neither size, even in colour.** Whether its ring is drawn cannot be told from the text. | This widens S29-3 (OW-3). On a severe day the one place the listener chose can vanish from the picture. |
| S30-4 | With no colour, the radar's shades again replace the warnings' outlines and hatch across the rain. | L-8.3 at an outbreak's scale. |

## Specimen 31 — a partial alert area (watchpost 0.18.0 PLAN, FR-4.4, M4)

Watchpost's DISCOVER asked for "the partial-area ruling, made with a drawing on screen". This is that
drawing. A live NWS **Flood Watch (Severe) over five West Texas zones**, 2026-09-23, with the
**Davis Mountains** zone's shape withheld to make it partial, the way a failed or capped zone fetch
does. **Fort Davis**, the named place, lies inside the missing zone. At 69×12 and 149×38, with and
without colour. Program and inputs: `radar-loops/02-analysis/programs/partial-render.go.txt`,
`inputs/partial/`.

| File | What it is |
|---|---|
| `31a-whole-*` | All five zones: the truth, for comparison |
| `31b-found-labelled-*` | Partial, drawn as found; the label says "4 of 5 zones" |
| `31c-found-labelled-noted-*` | As 31b, plus a line below the map, in watchpost's chrome, naming the missing zone and saying the place is in it |
| `31d-withheld-*` | Partial, withheld: nothing drawn until every zone is in; the line says why |

| # | Finding | Consequence |
|---|---|---|
| S31-1 | **Drawn as found, the picture puts Fort Davis outside the watch it is inside.** The label's "4 of 5" says the area is incomplete, but not where the gap is, and not that the place is in it. | Watchpost knows the selected place's zone and the alert's zone list, so it can say "the place is in the missing zone" — a fact the picture cannot show. |
| S31-2 | **The named place's name is drawn at neither size**, in any variant. | The third specimen in a row (S29-3, S30-3); OW-3's diagnosis matters. |
