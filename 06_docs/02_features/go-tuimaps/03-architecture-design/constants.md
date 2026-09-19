# Constants — every number a test needs before it can be written

Up: [architecture](architecture.md) · Carries: NFR-3, NFR-4, NFR-5, NFR-6, NFR-10, NFR-20, FR-9, FR-11, FR-15, FR-16, FR-18a, FR-23, D-63, D-84, D-85, D-88

| Field | Value |
|---|---|
| Phase | PLAN |
| Date | 2026-09-19 |
| Why this exists | The requirements named these as PLAN artefacts; the PLAN red-team found them missing, so tests could not be written first (PL-CQ-12, PL-PF-2, PL-PM-1, PL-DQ-7). |
| How to read it | **Ruled** = HUM LEAD's number. **Set here** = chosen in PLAN, with its reason; host-settable where it says so. **Measured** = taken from real data in PLAN, with the sample named. **By arithmetic** = worked out, not measured; the BUILD task that measures it is named. |

## 1 · Untrusted input (NFR-10) — with what real tiles measure

Measured on twelve real tiles from zoom 5 to zoom 14, chosen to be heavy: central Paris, Tokyo, New York and London at zoom 14; New York and Tokyo at 12, 10 and 8; the two fixture views.

| Limit | Default (set in DISCOVER round 2) | Largest measured | Headroom |
|---|---|---|---|
| Tile body | 2 MiB | 1.10 MB — central Paris, zoom 14 | 1.9× |
| Decompressed | 8 MiB | 1.10 MB | 7× |
| Layers a tile | 64 | 13 | 5× |
| Features a tile | 100,000 | 16,952 — Paris | 6× |
| Geometry integers a tile | 2,000,000 | about 200,000 (96,027 vertices) — Paris | 10× |
| Vertices in one feature | — (bounded by the above) | 20,441 | — |
| Retained after decode | 4 MiB | 0.80 MB for four zoom-5 tiles together | — |
| **Tile extent** | **Set here: 1 to 8,192** *(was "1 to 65,536")* | 4,096 in every tile | 2× |

**Why the extent changed (PL-IS-2).** Coordinates are kept as 16-bit integers (D-75), which hold −32,768 to 32,767. An extent of 65,536 does not fit. With an extent of at most 8,192, a coordinate may run a full extent outside the tile on every side — far more than any buffer — and still fit. **A cursor that leaves the 16-bit range is an error**, never a wrap.

## 2 · Memory (D-85, NFR-3, NFR-20)

| Constant | Value | Status |
|---|---|---|
| Tile cache, shared by every map | 0.5 MB | Ruled (D-85); host-settable |
| Shape cache, shared | 0.25 MB | Ruled (D-85); host-settable |
| Images, per map | 0.25 MB | Ruled (D-85); host-settable |
| A tile or simplified shape larger than a quarter of its cache | Drawn, not cached | Set here |
| Vertices an overlay · a map | 2,000,000 · 4,000,000 | Set in DISCOVER round 3 |
| Pixels an image | 1,048,576, read from the header before decoding | Set in DISCOVER round 2 |
| Pump width the peak line is measured at | 2 `Work` calls | Ruled (D-84) |
| Warnings kept | 64, de-duplicated | Set in DISCOVER round 3 |
| Disk cache | 256 MB; prune to 90%; recency written at most hourly | Set in DISCOVER round 2 |

## 3 · Shapes (FR-11)

| Constant | Value | Reason |
|---|---|---|
| **Zoom bucket** | One bucket per whole zoom level: bucket *b* serves every zoom from *b* up to, not including, *b* + 1 | Set here. Simple to state and test; fractional zoom maps to exactly one bucket |
| Tolerance inside a bucket | Half a braille dot **at zoom *b* + 1** — the finest zoom the bucket serves | Set here. The simplified shape is never coarser than the view needs; it costs at most twice the vertices of an exact fit |
| On a miss | The nearest prepared bucket is drawn while the right one is prepared | FR-11 |
| Dropped | A ring that simplifies to fewer than three distinct points at the bucket's tolerance — smaller than a dot | Set here. Measured: of 1,107 real island polygons, 62 survive at zoom 6.4 and 1,099 at zoom 12 |
| Draw-from-borrowed fallback: culling | Bounding boxes over runs of 64 vertices | Set here |
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
| Temperature preset | 17 classes; breaks every 5 °C from −30 to +45; pale break at 0 °C; °F breaks are these converted exactly | D-62, specimen 21 |
| Radar preset | Six classes from 10, 20, 30, 40, 50, 60 dBZ; two ramps, chosen by the ground's luminance | D-69, specimen 22, PL-AX-1 |
| Ground counts as light when | Black gives more contrast against it than white does | Set here — the same test as the foreground rule |

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
| Radar and precipitation | `radar.1` to `radar.6` — named by class floor: 10, 20, 30, 40, 50, 60 dBZ |
| Temperature | `temperature.1` to `temperature.17` — named by band, coldest first |
| A host's own type | `low` · `middle` · `high` — interpolated across its classes |
| Line features | `track` · `track.label` |

The sixteen-colour depth has its own small set of values for the basemap tokens, chosen from the sixteen by hand (specimen 23, D-79).

## 5 · Time

| Constant | Value | Status |
|---|---|---|
| Retry after a failed tile | 30 s, doubling to 10 min | DISCOVER round 1 |
| Blink | Upstream's 800 ms period | Parity P-59a |
| Flash, when built | At most 2.5 a second | Ruled (D-56) |
| Overlay currency | More than 0, at most 7 days; a valid time over 5 minutes ahead is uncertain and treated as stale | DISCOVER round 2 |
| "No `Work` has been called" warning | After 20 consecutive renders that found work pending | Set here — about 6 s at the first host's tick |

## 6 · Determinism (NFR-6)

| Constant | Value | Reason |
|---|---|---|
| Projection results | Rounded to 1/256 of a dot before they are used to raster | Set here. Finer than anything visible; coarse enough to absorb last-bit differences between architectures in the standard library's transcendental functions. The two-architecture reference frames decide whether it is enough |
| Tile order | By zoom, then x, then y | NFR-6 |
| Ties in the cell-colour vote after neighbours | The colour with the larger value, read as one number | D-83; any fixed rule would do |

## 7 · Targets that only the library can measure

PLAN was to "validate or revise" these. Without the library they can only be checked by arithmetic, and that is what is recorded — with the BUILD task that measures each.

| Target | By arithmetic | Measured by |
|---|---|---|
| NFR-5 cold start ≤ 3 s, on the stated link, **two `Work` calls wide** (D-84) | Connection setup 0.30 s + two rounds of requests 0.30 s + 1.245 MB at 8 Mbit/s 1.25 s + decode 0.10 s + one 0.30 s tick = **about 2.25 s**. One wide: about 2.45 s. The target holds, with less room than it looked | Task 14.9, on a virtual-clock link so it is deterministic; a real-time run is recorded but not gating |
| NFR-5 warm ≤ 1 s | Nothing fetched; four decodes and three overlay prepares, tens of milliseconds | Task 14.9 |
| NFR-4 changed frame, marker phase ≤ 16 KB | Only the marker's row is rebuilt: 149 cells at about 41 bytes is 6 KB | Task 14.7 |
| NFR-4 one-cell pan | Every row changes: about 232 KB of row bytes, **reused, not allocated** — the frame's buffers persist between renders (contract, section 5). The target of "no more than the first host's own window cost" stands | Task 14.7 |
| FR-29 description: fixture with 60 places ≤ 50 ms | 60 places × about 1,200 simplified-or-boxed segment tests is well inside it; the exact unsimplified test runs only for shapes whose box contains the place | A new benchmark task in WP-11 |
| NFR-3 live and peak | See [memory-measurement.md](memory-measurement.md): 3.0 MB live by the measured parts; the peak sits at the line by arithmetic | Task 14.6. **Risk RS-7 stays High until then** |
