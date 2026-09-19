# Constants — every number a test needs before it can be written

Up: [architecture](architecture.md) · Carries: FR-11, FR-15, FR-16, FR-22, FR-29, NFR-3, NFR-4, NFR-5, NFR-6, NFR-10, NFR-20, D-53, D-56, D-62, D-63, D-64, D-69, D-75, D-77, D-78, D-79, D-82, D-83, D-84, D-85, D-88, D-90, D-91, D-92, P-31, P-59a

| Field | Value |
|---|---|
| Phase | PLAN |
| Date | 2026-09-19 |
| Why this exists | The requirements named these as PLAN artefacts; the PLAN red-team found them missing, so tests could not be written first (PL-CQ-12, PL-PF-2, PL-PM-1, PL-DQ-7). |
| How to read it | **Ruled** = HUM LEAD's number. **Set here** = chosen in PLAN, with its reason; host-settable where it says so. **Measured** = taken from real data in PLAN, with the sample named. **By arithmetic** = worked out, not measured; the BUILD task that measures it is named. |

## 1 · Untrusted input (NFR-10) — with what real tiles measure

Measured on real tiles across the whole range the map uses. **Zoom 0 to 4:** all 85 tiles of zoom 0 to 3 from the planet archive the embedded tiles are cut from, and four zoom-4 tiles *(added after red-team round 2 found the range started at 5 — P2-PRD-14)*. **Zoom 5 to 14:** twelve tiles chosen to be heavy — central Paris, Tokyo, New York and London at zoom 14; New York and Tokyo at 12, 10 and 8; the two fixture views. "MB" in this file is 1,000,000 bytes; "MiB" is 1,048,576.

| Limit | Default (set in DISCOVER round 2) | Largest measured | Headroom |
|---|---|---|---|
| Tile body, as received | 2 MiB | 1.10 MB — central Paris, zoom 14, served uncompressed | 1.9× |
| Decompressed | 8 MiB | **1.56 MB — world tile 2/2/1** (0.72 MB compressed). *The low zooms are the heaviest tiles there are: the first table, which stopped at zoom 5, said 1.10 MB* | 5× |
| Layers a tile | 64 | 13 | 5× |
| Features a tile | 100,000 | 16,952 — Paris | 6× |
| Geometry integers a tile | 2,000,000 | about 200,000 (96,027 vertices) — Paris | 10× |
| Attribute keys in one kept layer | 4,096 | Set in BUILD (task 03.6) | — |
| Attribute values in one kept layer | 400,000 — four times the feature limit | Set in BUILD (task 03.6). A value is not read until a kept key points at it; the table that finds it costs 8 bytes a value while a layer is decoded | — |
| Vertices in one feature | — (bounded by the above) | 20,441 | — |
| Retained after decode, one tile | 4 MiB | **0.24 MB — world tile 2/2/1.** Others: 0.19 to 0.20 MB a tile for New York at zoom 10 and the Midwest at zoom 5; 0.06 to 0.13 MB for the four city tiles at zoom 14; 0.08 MB a tile on the Gulf coast at zoom 6 | 17× |
| Tile zoom the library will address | **Set in BUILD (task 00.10): 0 to 22.** Sources in the schema the library reads stop at 14 and the view at 18; 22 is the deepest a web-map tile source goes, and keeps a column or row inside 32 bits | 14 in every source measured | — |
| **Tile extent** | **Set here: 1 to 8,192** *(was "1 to 65,536")* | 4,096 in every tile | 2× |

**What the library's own decoder keeps, measured in BUILD (task 03.12)** on the fixture's twelve tiles: the Gulf view's four tiles 326 KB together (PLAN's throwaway program said 0.33 MB), the Midwest view's four 863 KB (PLAN said 0.80 MB), the four city tiles 159 to 217 KB each. Decoding the heaviest Midwest tile — 347 KB of bytes, 297 KB kept — allocates 393 KB in all: each layer's slabs are allocated once, at their exact size, so the decoder's own peak is about 1.3 times what it keeps. **A host may lower any limit and may raise none** (`Limits.Validate`).

**Why the extent changed (PL-IS-2).** Coordinates are kept as 16-bit integers (D-75), which hold −32,768 to 32,767. An extent of 65,536 does not fit. With an extent of at most 8,192, a coordinate may run a full extent outside the tile on every side — far more than any buffer — and still fit. **A cursor that leaves the 16-bit range is an error**, never a wrap.

## 2 · Memory (D-85, NFR-3, NFR-20)

| Constant | Value | Status |
|---|---|---|
| Tile cache, shared by every map | 0.5 MB — 500,000 bytes | Ruled (D-85); host-settable |
| Shape cache, shared | 0.25 MB — 250,000 bytes | Ruled (D-85); host-settable |
| Images, per map | 0.25 MB — 250,000 bytes. A 600×400 image at one byte a pixel is 240,000 | Ruled (D-85); host-settable |
| An image being replaced | The old one draws until the new one is prepared, so for that moment a map holds both. **That is peak, not live**: the cap bounds what is kept | Set here; task 14.6 measures it |
| Pending queue | 256 jobs a map; past that the oldest job for a view no map is showing is dropped first, then the oldest | Set here. A 149×38 view wants at most 9 tiles and their ancestors; 256 is an order above any honest need |
| Embedded tiles, compressed, inside the binary | **1,015,126 bytes** (85 tiles); a test holds the set under 2.5 MB | Measured in BUILD (task 04.11) on planet `20260913_164504_pt`. Decoded, the 85 tiles keep 3,535,546 bytes in all; the largest keeps 209,131 |
| What a live view is drawing — *live view* and *need* are defined in the contract, section 8 | **Never evicted (D-90).** Spare tiles and shapes are kept only in the room left under the cap; when need alone exceeds a cap, the cache holds exactly the need, a `cache-under-need` warning is raised once, and `CacheUse()` reports need and cap for each cache | Ruled (D-90). *An earlier line here — "larger than a quarter of its cache is drawn, not cached" — was the coordinator's, ratified only with the requirements as a whole, and is withdrawn* |
| A simplified shape larger than the **whole** shape cap | Not cached: drawn from the host's memory by FR-11's fallback | Ruled (D-90) |
| What one 149×38 view draws, in kept form | 0.33 MB on the Gulf coast at zoom 6; 0.80 MB in the Midwest at zoom 5 | Measured in PLAN |
| Vertices an overlay · a map | 2,000,000 · 4,000,000 | Set in DISCOVER round 3 |
| Pixels an image | 1,048,576, read from the header before decoding | Set in DISCOVER round 2 |
| Pump width the peak line is measured at | 2 `Work` calls | Ruled (D-84) |
| Warnings kept | 64, de-duplicated | Set in DISCOVER round 3 |
| The longest id of an overlay or a place | 256 bytes | Set in BUILD (task 02.8). Ids are validated and never cleaned (FR-34) |
| Untrusted text quoted in an error or a warning | 64 grapheme clusters, then three dots | FR-34; the dots set in BUILD (task 02.7) |
| Disk cache | 256 MB; prune to 90%; recency written at most hourly; layout `<root>/v1/<hash>/<z>/<x>-<y>.pbf` | Set in DISCOVER round 2; layout set in BUILD (WP-06) |

## 3 · Shapes (FR-11)

| Constant | Value | Reason |
|---|---|---|
| **Zoom bucket** | One bucket per whole zoom level: bucket *b* serves every zoom from *b* up to, not including, *b* + 1 | Set here. Simple to state and test; fractional zoom maps to exactly one bucket |
| Simplification | **Ramer–Douglas–Peucker**, ring by ring, keeping each ring closed — the algorithm upstream uses for its own lines (P-31), and the one PLAN's measurements were made with | Set here (FR-11 asked for it to be named — P2-PRD-15) |
| Tolerance inside a bucket | Half a braille dot **at zoom *b* + 1** — the finest zoom the bucket serves | Set here. The simplified shape is never coarser than the view needs. *What it costs was first written as "at most twice the vertices of an exact fit"; the fixture says more — its kept vertices grow about 2.3 times a zoom level (9.5 KB at zoom 6.4, 87 KB at 9, 359 KB at 12) — so a bucket can hold up to about 2.3 times what an exact fit would* |
| On a miss | The nearest prepared bucket is drawn while the right one is prepared | FR-11 |
| Dropped | A ring that simplifies to fewer than three distinct points at the bucket's tolerance — smaller than a dot | Set here. Measured: of the fixture's 1,107 rings (in 1,011 polygons, mostly islands), 62 survive at zoom 6.4 and 1,099 at zoom 12 |
| Draw-from-borrowed fallback: culling | Bounding boxes over runs of 64 vertices | Set here |
| Draw-from-borrowed fallback: **when the index is built** | **Inside `Set`, in one linear pass** (D-92), for any feature overlay with more vertices than fit the shape cache unsimplified **together with their run index**: 8 bytes a vertex plus 16 bytes for each run of 64 is 8.25 bytes a vertex, so **30,303 vertices at the default cap of 250,000 bytes**. *(First given as 31,250, which forgot the index's own bytes — red-team round 3.)* A simplified form "fits" by the same arithmetic: its vertices × 8.25. At or below the count the fallback can never be needed. Caps are fixed at creation, so the count never moves under an overlay | The pass is ruled (D-92); the count is set here |
| Run index | One 16-byte box for each run of 64 vertices; lives with the overlay while it is set; counts as need and is never evicted (D-90). 203 KB at 812,058 vertices; 500 KB at the vertex cap — twice the default shape cap, and reported as such | Set here |
| Draw-from-borrowed fallback: **bound** | **Stated as work, not as time** (PL-PF-6): at most one box test per run, plus the vertices of the runs whose box meets the view. For the synthetic worst case (812,058 vertices) that is 12,689 box tests a frame | By arithmetic. Task 14.10 asserts the count; a time figure is recorded by the benchmark, not gated |
| Line features with no colour | Dashed: five dots on, four off, two dots thick | Set here, from specimen 19c (D-77) |

## 4 · Colour (FR-15, FR-16)

| Constant | Value | Status |
|---|---|---|
| Colour-vision difference | At least 10, every pair of classes and every class against its ground; full-severity simulation in linear light; 1976 Lab difference; D65 | Ruled (D-88) |
| Contrast | Line work and glyphs 3:1; labels and legend text 4.5:1 | DISCOVER (FR-16) |
| Foreground | The line's own colour where it meets 3:1 on that cell; otherwise whichever of black and white contrasts more | D-77, specimen 20 |
| Classes in a host's own type | At most **21** | Set here. The most an ordered, colour-vision-safe scale was found to carry (specimen 21's search) |
| 256 colours | Only indices 16 to 255 | DISCOVER round 3 |
| Image colour tolerance | 0 to 25, default 10, in the same Lab difference | DISCOVER round 3. The provider PLAN tested needs 0: all 20,959 pixels matched exactly |
| Image resampling | Eight samples a cell, the heaviest class wins | Ruled (D-78) |
| Temperature preset | 17 classes; breaks every 5 °C from −30 to +45; pale break at 0 °C; °F breaks are these converted exactly. **Two sets of colours, chosen by the ground in effect:** they differ in three bands, −10 to +5 °C (D-91) **Order:** relative luminance rises to freezing and falls from it, as FR-16 defines order. In truecolor the same holds for lightness under all three simulated kinds of colour vision; at 256 colours it does not quite — three neighbouring pairs invert by 0.1 to 1.6 lightness units under protanopia or tritanopia, because that palette is coarse. Radar's ramps hold under every kind at both depths | D-62, D-91, specimen 21 |
| Alert preset | **Not yet designed in full.** The two pairs PLAN's specimens drew pass on the dark ground: outlines 21.2 apart, tints 12.6, every colour at least 25.6 from the ground, outline on its own tint at least 4.3:1. **On a light ground both outlines fail 3:1** (1.3 and 2.3), so alerts need a light-ground set as radar and temperature do; and three of the five severities have never been drawn. Designed in task 08.10, shown to HUM LEAD before reference frames are frozen | Measured here (P2-PRD-18); owed in BUILD |
| Radar preset | Six classes from 10, 20, 30, 40, 50, 60 dBZ; two ramps, chosen by the ground's luminance | D-69, specimen 22, PL-AX-1 |
| Ground counts as light when | Black gives more contrast against it than white does | Set here — the same test as the foreground rule |
| Which ground that test reads | **The ground in effect**: the `ground` token when the library paints it, the host's declared colour when it does not. "Do not paint" cannot be set without a declaration (D-64), so there is always one | Set here (P2-PRD-3) |
| Presets against a ground | Required to pass D-88 on the library's own two painted grounds — dark (16, 22, 28) and light (245, 245, 240) — and their 256-colour forms. On any other ground, painted by a host's token or declared, the result is **reported, not refused**, as for any host colour (D-53): on a mid-grey ground from 185 to 216 the light radar ramp's palest class is under 10 from it, 8.1 at worst (grey 199) | Set here (P2-PRD-3) |

### The semantic tokens (D-63)

Stable names; adding one is a minor version. A host sets any subset; the rest keep the library's defaults.

| Group | Tokens |
|---|---|
| Ground and water | `ground` · `water.fill` · `water.line` |
| Basemap lines | `coast` · `border.country` · `border.region` · `river` · `road.major` · `road.minor` · `rail` · `park` · `runway` |
| Basemap text | `label.place` · `label.water` · `label.region` |
| Furniture | `scale` · `credit` · `notice` · `stale` · `focus` |
| Places | `marker` · `marker.label` |
| Alert areas | `alert.extreme.outline` · `alert.extreme.tint` · and the same pair for `severe`, `moderate`, `minor`, `unknown` |
| Radar and precipitation | `radar.1` to `radar.6` — numbered by position, lightest rain first; the floors are 10, 20, 30, 40, 50, 60 dBZ. One set of names: the library's defaults for them differ by ground, and a host that sets them has set them for its own ground |
| Temperature | `temperature.1` to `temperature.17` — numbered by position, coldest first |
| A host's own type | `low` · `middle` · `high` — interpolated across its classes |
| Line features | `track` · `track.label` |

The sixteen-colour depth has its own small set of values for the basemap tokens, chosen from the sixteen by hand (specimen 23, D-79).

### Two warnings that need a rule to be testable (P2-ENG-7)

| Warning | Fires when |
|---|---|
| `near-duplicate-id` | A `Set` **creates** an overlay whose id is not already set, and that id differs from an existing id by letter case only, by surrounding or repeated white space only, or by exactly one character added, dropped or changed |
| `implausible-unit` | Temperature declared °C with any value above 60 or below −90; declared °F with any value above 140 or below −130; radar declared dBZ with any value above 95. Values are accepted either way |

### The closed glyph list (NFR-8) — fixed in BUILD, task 02.10

372 characters the renderer may emit of its own accord, each one cell wide under the pinned table and each seen on the terminal test card: the space and printable ASCII (95) · braille U+2800 to U+28FF (256) · markers U+00B7, U+2022, U+25C6, U+25C9, U+25CB, U+25CF · box-drawing U+2500, U+2502, U+250C, U+2510, U+2514, U+2518, U+251C, U+2524 · hatch U+2571, U+2572, U+2573 · shades U+2591 to U+2593 · U+FFFD. Arrows and block quadrants join it with wind and the block renderer. Outside text is not held to this list; it is cleaned and measured.

## 5 · Time

| Constant | Value | Status |
|---|---|---|
| Retry after a failed tile | 30 s, doubling to 10 min | DISCOVER round 1 |
| Blink | Upstream's 800 ms period | Parity P-59a |
| Flash, when built | At most 2.5 a second | Ruled (D-56) |
| Overlay currency | More than 0, at most 7 days; a valid time over 5 minutes ahead is uncertain and treated as stale | DISCOVER round 2 |
| Fetch: whole request · waiting for the first byte | 20 s · 10 s. Enforced by the context `Work` was given plus the standard library's own client timeout — the one timer the static check allows, because it lives and dies inside the host's `Work` call | Set here (FR-22); task 05.10 |
| "No `Work` has been called" warning | After 20 consecutive renders that found work pending | Set here — about 6 s at the first host's tick |

## 6 · Determinism (NFR-6)

| Constant | Value | Reason |
|---|---|---|
| Projection: which functions | Forward: *y* = atanh(sin φ). Inverse: φ = asin(tanh *y*). Written with the standard library's sine, arcsine, hyperbolic tangent and its inverse — **none of which has an assembly version on either gated architecture**, unlike the exponential and logarithm the textbook forms use, which on one architecture also branch on a processor feature | Set here (P2-ENG-10; the processor-feature claim is the reviewer's, not re-checked) |
| Projection results | Rounded to 1/256 of a dot before they are used to raster | Set here. Finer than anything visible; coarse enough to absorb last-bit differences between architectures in the standard library's transcendental functions. The two-architecture reference frames decide whether it is enough |
| Tile order | By zoom, then x, then y | NFR-6 |
| Distances | Great-circle, on a sphere of the Earth's mean radius, 6371.0088 km: within about half a percent of the ellipsoid, and finer than a cell can show | Set in BUILD (task 01.5) |
| Compass words | Eight, each 45 degrees wide and centred on its direction: 342 degrees is "north" | Set in BUILD (task 01.6), from scenario 2's answer key |
| Ties in the cell-colour vote after neighbours | The colour with the larger value, read as one number | D-83; any fixed rule would do |

## 7 · Targets that only the library can measure

PLAN was to "validate or revise" these. Without the library they can only be checked by arithmetic, and that is what is recorded — with the BUILD task that measures each.

| Target | By arithmetic | Measured by |
|---|---|---|
| NFR-5 cold start ≤ 3 s, on the stated link, **two `Work` calls wide** (D-84) | Connection setup 0.30 s + two rounds of requests 0.30 s + 1.245 MB at 8 Mbit/s 1.25 s + decode 0.10 s + one 0.30 s tick = **about 2.25 s**. One wide: about 2.55 s (three rounds of requests, not two). Not counted in either: a TileJSON round trip, about 0.15 s, paid once for a source named by TileJSON. The target holds, with less room than it looked | Task 14.9, on a virtual-clock link so it is deterministic; a real-time run is recorded but not gating |
| NFR-5 warm ≤ 1 s | Nothing fetched; four decodes and three overlay prepares, tens of milliseconds | Task 14.9 |
| NFR-4 changed frame, marker phase ≤ 16 KB | Only the marker's row is rebuilt: 149 cells at about 41 bytes is 6 KB | Task 14.7 |
| NFR-4 one-cell pan | Every row changes: about 232 KB of row bytes, **reused, not allocated** — the frame's buffers persist between renders (contract, section 5). The target of "no more than the first host's own window cost" stands | Task 14.7 |
| FR-29 description: fixture with 60 places ≤ 50 ms | 60 places × about 1,200 simplified-or-boxed segment tests is well inside it; the exact unsimplified test runs only for shapes whose box contains the place | A new benchmark task in WP-11 |
| NFR-3 live and peak | See [memory-measurement.md](memory-measurement.md): about 3.0 MB live for three maps over the fixture region, by the measured parts; the peak sits at the line by arithmetic. Three maps on three different dense views run to about 4.9 MB live — over the line, reported, recorded by the benchmark and not gated (D-90) | Task 14.6. **Risk RS-7 stays High until then** |
