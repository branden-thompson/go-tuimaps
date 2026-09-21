# Parity Matrix — FROZEN

| Field | Value |
|---|---|
| Baseline | TerminalMap `3b960723b24b781a1a9a04c165803cd1e98e6e3c` (v0.1.0) — the **code**, not the README (D-11) |
| Surface | Behavioural parity: same features, same controls, same intentional constants. Byte-for-byte frame equality is not required (D-11). |
| Source of rows | [`research/AI-1`](research/AI-1-upstream-module-read.md) §3. The first four columns of every P-row are copied from that file by script and verified identical by diff; nothing was retyped — except the ten half-rows of D-61, which quote parts of five rows kept whole at the foot of this file. |
| Defects | [`defect-ledger.md`](defect-ledger.md), approved and closed (D-37) |
| Frozen | 2026-09-18, on approval of the defect ledger. **Corrected 2026-09-18 under its own change control by ruling D-49** — and again by D-56, for the flash rate in P-59 — after the DISCOVER red-team found four Match rows written in the vocabulary of the superseded renaming shim, unrecorded exceptions to P-08, and an "Open for PLAN" disposition that left the denominator unfrozen. The first four columns of every P-row remain the verified copy of the source read; every correction is in the Disposition, Release and Note columns — **except the ten half-rows made by ruling D-61**, which quote parts of five rows whose unaltered text is kept at the foot of this file. The corrections were approved by D-49; the matrix as a whole is ratified with the Discovery Report. |
| Change control | **The denominator is frozen.** A row changes disposition, or leaves the denominator, only by a recorded HUM LEAD ruling (the anti-solution guard on metric M3). |

## Dispositions

| Disposition | Meaning | Rows |
|---|---|---|
| **Match** | go-tuiMaps behaves as upstream does. | 48 |
| **Fix** | Upstream's behaviour is a ledgered defect — or, for P-59b alone, a behaviour HUM LEAD ruled unsafe (D-56: flash faster than three a second) — and go-tuiMaps does the intended thing. | 13 |
| **Replicate** | A ledgered convention kept deliberately. | 2 |
| **Extended** | Upstream's behaviour is kept and added to by a ruling. | 12 |
| **Superseded** | Replaced by design under a ruling; **outside the M3 denominator**. | 2 |

**M3 denominator: 75 rows** (Match + Fix + Replicate + Extended) — *70 until ruling D-61 split five rows that v0.1.0 contains only in part (P-03, P-05, P-59, P-68, P-72) into an early and a later half each; the rows as read from the source are kept unaltered at the foot of this file.* **Excluded by named ruling: P-40** (the layer-renaming shim — D-24) **and P-46** (a tile cache bounded by a count of 32 — D-29, confirmed as an exclusion by D-49). **M3 is stated per release (D-44): v0.1.0's denominator is the 62 rows marked v0.1.0** below (62 of 75; 13 later); the rest are demonstrated when their scope is built. Each must be shown by a passing Go test or an accepted rendered specimen. `[LIB]` rows belong to the library, `[APP]` rows to the standalone app; untagged rows are rendering internals shared by both.

## Upstream behaviours (P-rows)

| ID | Behaviour | Where | Semantics to match | Disposition | Release | Note / governing ruling |
|---|---|---|---|---|---|---|
| P-01 | Pixel grid | braille.rs:56,209 | The cell index is `(x>>1)+(w>>1)*(y>>2)`. There are `w/2 × h/4` cells | **Match** | v0.1.0 |  |
| P-02 | Dot bitmask | braille.rs:10-15 | Rows × cols: `[01,08],[02,10],[04,20],[40,80]`. The glyph is U+2800+mask | **Match** | v0.1.0 |  |
| P-03a | Empty cell | braille.rs:299-302 | In braille mode it emits U+2800, not a space | **Match** | v0.1.0 | Braille half of P-03 (split by D-61). |
| P-03b | Empty cell | braille.rs:299-302 | In block mode it emits `' '` | **Match** | later | Block half of P-03 (split by D-61); the block renderer is built after the integration (D-42, D-44). |
| P-04 | Block ("ASCII") mode | braille.rs:23-32,213-230 | Six glyphs: ▀ ▄ ■ ▌ ▐ █. The one with the highest `popcount(mask & bits)` wins, and the first one wins ties. The glyphs are Unicode blocks, not ASCII | **Fix** | later | L-1. Built as a first-class quadrant renderer (D-23, S2-1); upstream's six-glyph table is not ported. |
| P-05a | Colour depth | utils.rs:76; braille.rs:240-246 | xterm-256 only, as `38;5;n` / `48;5;n`. There is no truecolor or 16-colour path. Index 0 means "unset" | **Extended** | v0.1.0 | Truecolor, 256 colours and no colour, each with its own ramp (D-23, S10-1, S10-2). 256-colour output remains available. Split by D-61. |
| P-05b | Colour depth | utils.rs:76; braille.rs:240-246 | As P-05a | **Extended** | later | The 16-colour ramps. Until they are built a 16-colour hint is handled as D-59 rules. Split by D-61. |
| P-06 | SGR forms | braille.rs:239-247 | fg+bg, `49;38;5;fg`, `39;48;5;bg`, or `39;49`. Emitted only when the colour changes | **Extended** | v0.1.0 | **Form pending (D-49).** In the denominator: whatever output form PLAN chooses — a colour-coded string, a cell grid, or both — must be demonstrated. Supersedes the earlier "Open for PLAN" disposition, which left the denominator unfrozen. |
| P-07 | Row self-containment | braille.rs:258-266,310 | Colour is re-emitted at the start of each row. Rows are joined with `\r\n`. The frame ends with `ESC[39;49m` | **Extended** | v0.1.0 | As P-06 (D-49). |
| P-08 | Cell colour vote | braille.rs:134-207 | The cell takes the majority colour among its lit pixels. A tie is broken by the count of that colour among lit pixels in the 8 neighbouring cells. Locked cells use the last fg | **Extended** | v0.1.0 | The cell-colour vote is matched, with three recorded exceptions (D-49): under a field or image the foreground is chosen for contrast with the cell (S10-2); water does not vote when it owns the cell background (S1-2); the vote is defined for the braille renderer, and the block renderer's equivalent is settled in PLAN. Colour is resolved at draw time (D-26). *Confirmed from a side-by-side specimen (D-83, specimen 24).* |
| P-09 | Forced pixels | braille.rs:115-129 | `set_pixel_forced` locks the whole cell's colour | **Match** | v0.1.0 |  |
| P-10 | Text cells | braille.rs:275,286-294,314 | A char takes priority over dots. It uses the cell fg and bypasses the vote | **Match** | v0.1.0 |  |
| P-11 | Wide chars | braille.rs:288-306,432 | The cells after a CJK char are skipped. The width table is hand-rolled | **Match** | v0.1.0 | Behaviour matched — a wide character occupies two cells. The *hand-rolled width table* is not: one pinned table, the same the first host measures with (NFR-8, D-49). |
| P-12 | Text placement | braille.rs:324-336 | One char per cell, at `x+i*2` px. The centre offset is `len/2+1` (byte length) | **Fix** | v0.1.0 | L-7: display width, not byte length. |
| P-13 | Thin line | canvas.rs:154 | Bresenham when `width<=1` | **Match** | v0.1.0 |  |
| P-14 | Thick line | canvas.rs:103-151 | Zingl algorithm, `wd=(w+1)/2` with `w=width-1` | **Match** | v0.1.0 |  |
| P-15 | Polygon fill | canvas.rs:57-93,208 | `earcutr::earcut` with holes. Each triangle is filled by taking the Bresenham edge points and spanning min to max x per row | **Match** | v0.1.0 | Behaviour matched — filled polygons with holes; the fill method is free (D-11). Under an active field, water owns the cell background instead (S1-2). |
| P-16 | Degenerate rings | canvas.rs:62-69 | An outer ring with fewer than 3 points aborts the polygon. A hole with fewer than 3 points is skipped | **Match** | v0.1.0 |  |
| P-17 | Tile zoom | utils.rs:9-12 | `clamp(floor(zoom),0,tile_range)`. `tile_range` is 14 | **Match** | v0.1.0 |  |
| P-18 | Tile pixel size | utils.rs:15-18 | `project_size*2^(zoom-z)`. `project_size` is 256. Overzoom above 14 scales up to 4096 px at z18 | **Match** | v0.1.0 |  |
| P-19 | Visible tiles | renderer.rs:130-150 | The range is `±(ceil(dim/2/ts)+1)`. Tiles with x or y outside the grid are dropped, so the world does not wrap | **Replicate** | v0.1.0 | L-17 (f): no repeat across the antimeridian. Flagged for a later look (D-37). |
| P-20 | Projection | utils.rs:21-35 | Standard Web Mercator `ll2tile`/`tile2ll` | **Match** | v0.1.0 |  |
| P-21 | Normalize | utils.rs:38-49 | Longitude wraps by a single ±360. Latitude is clamped to ±85.0511 | **Replicate** | v0.1.0 | L-17 (c), (e). |
| P-22 | Ocean outside world | renderer.rs:80-91 | Lit pixels with hard-coded colour 69, outside the floor/ceil world bounds | **Fix** | v0.1.0 | L-6: ocean colour from the style and palette. Under an active field, water owns the background (S1-2). |
| P-23 | Draw order, z≥2 | renderer.rs:483-498 | landuse, water, marine_label, building, road, admin, then 8 label layers | **Extended** | v0.1.0 | **Restated for go-tuiMaps (D-49):** the draw order is written in OpenMapTiles layer names, since the renaming shim is not ported (D-24). D-12: waterways, parks, airports and airport labels join it (rows E-01..E-04). |
| P-24 | Draw order, z<2 | renderer.rs:480-481 | water, landuse, admin, country_label, marine_label | **Extended** | v0.1.0 | As P-23. |
| P-25 | Label deferral | renderer.rs:223-239 | Decided by layer name containing `label` or `symbol`. Labels are stable-sorted ascending by `sort`. `show_labels=false` also hides POI glyphs | **Match** | v0.1.0 | **Restated for go-tuiMaps (D-49):** a layer is deferred to the label pass when its *style* layer type is `symbol` — upstream's test on the layer *name* cannot work with OpenMapTiles names, none of which contains `label` or `symbol`. Same behaviour: labels drawn last, stable-sorted by rank; hiding labels hides place glyphs too. |
| P-26 | Sort key | tile.rs:264-268 | `localrank`, else `scalerank`, else 0. Integers only | **Match** | v0.1.0 |  |
| P-27 | Zoom gate | renderer.rs:243-252 | Style `minzoom`/`maxzoom` are compared against the fractional map zoom. The test is inclusive | **Match** | v0.1.0 |  |
| P-28 | Feature cull | renderer.rs:183-193 | The bounding box is tested against the viewport in tile units | **Match** | v0.1.0 |  |
| P-29 | Scaling | renderer.rs:321-326 | `floor(pos + p/scale)`. Consecutive duplicate points are dropped | **Match** | v0.1.0 |  |
| P-30 | Line clip pad | renderer.rs:50,314-343 | 64 px padding. Runs of points outside it are collapsed. Fills are not clipped | **Fix** | v0.1.0 | L-3. Polygon fills are clipped too (D-16). |
| P-31 | Simplify | renderer.rs:350-352; config.rs:34 | RDP with ε=0.5. Off by default | **Match** | v0.1.0 |  |
| P-32 | Label anchor | renderer.rs:286 | `x = px - text.len()` (byte length). Each vertex is tried in turn until one fits | **Fix** | v0.1.0 | L-7. |
| P-33 | Label bounds | renderer.rs:278-284 | The anchor must be inside both the world bounds and the canvas | **Fix** | v0.1.0 | L-8: negative rows guarded. |
| P-34 | Collision | label.rs:28-70 | Works in cell coordinates (`x/2`, `y/4`). The rectangle is `[x-m, x+m+chars] × [y±m/2]`, overlap is inclusive, `m=5`. The scan is linear | **Match** | v0.1.0 | **The rectangle semantics are matched (D-49)** — cell coordinates, margin 5, inclusive overlap. The linear scan is not: how the search is done is free (D-11). |
| P-35 | POI glyph | renderer.rs:271; config.rs:40 | A symbol with no name draws `◉` | **Match** | v0.1.0 |  |
| P-36 | Label language | tile.rs:249-259 | Lookup order: `name_<lang>`, `name:<lang>`, `name_en`, `name:en`, `name`, `house_num` | **Match** | v0.1.0 | Matched for one configured language, English by default; every other language is dropped while decoding, and the embedded tiles keep English only (D-82). The language is part of the tile cache key (FR-31). |
| P-37 | Gzip sniff | tile.rs:58-67 | Detected by the magic bytes `1f 8b` | **Match** | v0.1.0 |  |
| P-38 | MVT decode | tile.rs:372-438; proto.rs | MoveTo, LineTo, ClosePath (which re-pushes the first point) and zigzag. The default extent is 4096 | **Match** | v0.1.0 |  |
| P-39 | Ring grouping | tile.rs:443-484 | Signed area ≥0 starts a new polygon. Negative area is a hole in the previous polygon. Each polygon becomes its own feature | **Match** | v0.1.0 |  |
| P-40 | OMT remap | tile.rs:71-162 | transportation→road, with `_link` for ramps, minor→street and brunnel→structure. boundary→admin. place→country/place_label. water_name→marine/water_label. poi→poi_label. park→landuse_overlay. landcover→landuse. `name:xx`→`name_xx`. rank→scalerank or labelrank | **Superseded** | — | D-24: fresh styles are written against OpenMapTiles; the renaming shim is not ported. |
| P-41 | Style match | styler.rs:229-237; tile.rs:207-213 | The first matching layer, in style order, for the source-layer. The remapped name is tried first, then the raw name. No match drops the feature. One style per feature | **Match** | v0.1.0 | **Restated for go-tuiMaps (D-49):** the first matching style layer, in style order, for the tile's real (OpenMapTiles) layer name; no match drops the feature; one style per feature. Colour resolution moves to draw time (D-26). |
| P-42 | Filter ops | styler.rs:32-118 | Supported: all, any, none, ==, !=, in, !in, has, !has, >, >=, <, <=. An unknown op evaluates to true. `==` on a missing key is false. Equality is on JSON values, so integers and floats are distinct. `$type` is injected as a property (tile.rs:187) | **Match** | v0.1.0 |  |
| P-43 | Constants and `ref` | styler.rs:132-181,240 | `@name` strings are substituted. `ref` inherits type, source-layer, zooms and filter | **Match** | v0.1.0 |  |
| P-44 | Colour pick | tile.rs:216-242 | `line-color`, else `fill-color`, else `text-color`. For stops only the first stop is used. The fallback is `#f00`. Hex is 3 or 6 digits only (utils.rs:58) | **Fix** | v0.1.0 | L-9: every zoom stop honoured. Colour resolved at draw time (D-26). Upstream's `#f00` fallback is matched. |
| P-45 | Line width | tile.rs:271-289 | A number or the first stop. The default is 1 | **Fix** | v0.1.0 | L-9. |
| P-46 | LRU | tile_source.rs:53-55,65 | 32 parsed tiles, keyed `z-x-y`, per `MapState` | **Superseded** | — | D-29: caches are capped by bytes with host-set caps, not by a count of 32. |
| P-47 | Embedded tiles | embedded_tiles.rs:13-21; tile_source.rs:76 | z0 (1 tile) and z1 (4 tiles), gzipped. They are consulted before disk and HTTP, whatever `source` is | **Extended** | v0.1.0 | D-27, D-33: an opt-in assets package, zoom 0–3, names stripped. L-13: never overrides a chosen source. |
| P-48 | Disk cache | tile_source.rs:42,152-165 | `<cache_dir>/terminalmap/<hex DefaultHasher(source)>/<z>/<x>-<y>.pbf`. Raw bytes, no expiry, no size cap | **Fix** | v0.1.0 | L-12: go-tuiMaps' own stable layout. Upstream's no-expiry behaviour is kept deliberately (D-18). |
| P-49 | URL modes | tile_source.rs:101-143 | A source ending in `/` becomes `{src}{z}/{x}/{y}.pbf`. Otherwise it is TileJSON: `tiles[0]` is used and the template is memoised | **Extended** | v0.1.0 | Both URL modes matched; a PMTiles source is added (D-18, row E-07) — after v0.1.0 (D-44). |
| P-50 | HTTP | tile_source.rs:59,145-150 | Default reqwest client. No timeout, no User-Agent, no status check, no retry | **Fix** | v0.1.0 | L-2 and D-15: timeout, User-Agent, status check; host-replaceable fetcher. |
| P-51 | Fetch failure | renderer.rs:95-101 | The error is swallowed and the tile is left blank. There is no negative cache | **Extended** | v0.1.0 | D-30: stand-in tiles and parallel fetching; render never blocks on the network. Failure handling is a PLAN detail. |
| P-52 [LIB] | Config defaults | config.rs:23-44 | See §4 | **Match** | v0.1.0 |  |
| P-53 [LIB] | Size from terminal | widget.rs:92-97 | `w=(cols>>1)<<2`, `h=(rows-3)*4` | **Fix** | v0.1.0 | L-17 (g): the library uses the exact rectangle it is given; the three-row reservation moves to the app. |
| P-54 [LIB] | Min zoom | widget.rs:83-88 | `min(log2(w/256), log2(h/256))`. Zoom is clamped up to it | **Match** | v0.1.0 |  |
| P-55 [LIB] | zoom_by / initial zoom | widget.rs:61,124-133 | Clamped to [min_zoom, max_zoom=18]. `None` means 0.0 | **Match** | v0.1.0 |  |
| P-56 [LIB] | fit_world | widget.rs:166-192 | Latitude 84 to −56. `zoom=min(log2(h/span_px), log2(w/256))`. Longitude 0, latitude at the Mercator midpoint | **Match** | v0.1.0 |  |
| P-57 [LIB] | Footer | widget.rs:195-202; utils.rs:52 | `center: lat, lon   zoom: z`, truncated with floor | **Match** | v0.1.0 |  |
| P-58 [LIB] | Marker shapes | renderer.rs:395-440 | Dot is 3×3. Cross is ±3. Diamond has radius 3. Ring(r) uses a midpoint circle. FilledCircle(r). Char | **Match** | v0.1.0 |  |
| P-59a [LIB] | Marker animation | marker.rs:97-118 | Blink: `(tick/8)%2==0` | **Match** | v0.1.0 | Blink keeps upstream's rate. Durations matched, converted from ticks at upstream's 50 ms loop; time is supplied by the host's clock (L-15). Reduce-motion and "never hidden by a frozen clock" apply (NFR-21). Split by D-61; *Match* was this row's disposition before D-56. |
| P-59b [LIB] | Marker animation | marker.rs:97-118 | Flash: `(tick/3)%2==0`. Pulse: radius 1,2,3,4,3,2 advancing every 4 ticks, for Ring only (renderer.rs:419) | **Fix** | later | **Flash: Fix (D-56)** — no more than 2.5 per second; upstream's 3.33 per second is above the three-per-second accessibility threshold. Pulse keeps upstream's rate. Durations as P-59a. Reduce-motion applies (NFR-21). Split by D-61. |
| P-60 [LIB] | Marker cull and label | renderer.rs:389,443-450 | ±20 px cull. The label sits at `px+4` and is collision-checked after the map labels. It takes the marker colour | **Match** | v0.1.0 |  |
| P-61 [LIB] | Marker id | marker.rs:62 | The default is `"{lat:.6},{lon:.6}"`. `remove_marker` removes every marker with that id | **Match** | v0.1.0 |  |
| P-62 [LIB] | Camera easing | camera.rs:296-302 | Cubic ease-in-out. Defaults are 60 travel ticks and 40 hold ticks (camera.rs:31-32) | **Match** | later | As P-59a. |
| P-63 [LIB] | Camera zoom arc | camera.rs:252-259 | The midpoint is `max(min(from,to)-0.8, 0)`, with each half eased | **Match** | later |  |
| P-64 [LIB] | Longitude path | camera.rs:310-326 | Shortest way round, result kept within ±180 | **Match** | later |  |
| P-65 [LIB] | Globe tour | camera.rs:98-154 | 12 cities, Paris to Berlin. Travel 60–120 ticks, hold 50, looping. The default zoom is 2.0 (widget.rs:250) | **Match** | later | As P-59a. |
| P-66 [LIB] | Marker tour | camera.rs:157-173 | Travel 70, hold 60. The label is the marker label or id | **Match** | later | As P-59a. |
| P-67 [LIB] | Tour end | camera.rs:277-283 | A non-looping tour deactivates and returns None | **Match** | later |  |
| P-68a [APP] | Keys | main.rs:77-143 | q/Esc quit. a/+ zoom in. z/y/- zoom out. Arrows/hjkl pan. n labels. o ocean. w fit world. m toggles the demo markers | **Match** | v0.1.0 | Pan, zoom and toggles, matched. Split by D-61. |
| P-68b [APP] | Keys | main.rs:77-143 | c braille. g globe tour. t marker tour at zoom 5.0 | **Extended** | later | The renderer switch and the tour keys, with keyboard equivalents added for every pointer operation (D-17, D-44). Split by D-61. |
| P-69 [APP] | Pan step | main.rs:87-99 | Longitude ±8/2^zoom, latitude ±6/2^zoom degrees | **Match** | v0.1.0 |  |
| P-70 [APP] | Mouse | main.rs:147-159 | Scroll only, ±zoom_step, centred on the map centre. There is no drag | **Extended** | later | D-17: drag to pan and zoom toward the pointer (rows E-05, E-06). |
| P-71 [APP] | Loop | main.rs:71,174-183 | 50 ms poll. tick and camera advance on every loop iteration. Animation redraws are throttled to 50 ms or more. Resize triggers a redraw | **Fix** | v0.1.0 | L-15: clock-driven animation. |
| P-72a [APP] | Chrome | main.rs:206-226 | Help row at `rows-2`. Status row at `rows-1`, with `>> label` | **Match** | v0.1.0 | Split by D-61. |
| P-72b [APP] | Chrome | main.rs:206-226 | `[TOUR…]` in the status row | **Match** | later | Arrives with camera tours (D-44). Split by D-61. |

## Extensions beyond upstream (E-rows) — outside the M3 denominator

Tracked so they are tested and reported, never counted as parity.

| ID | Extension | Governing ruling / evidence |
|---|---|---|
| E-01 | Waterways drawn | D-12, S8-1 |
| E-02 | Parks drawn | D-12, S8-1 |
| E-03 | Airport runways and taxiways drawn | D-12 |
| E-04 | Airport labels drawn | D-12 |
| E-05 | Pan by cells; drag to pan in the app | D-17 |
| E-06 | Zoom around a chosen point; zoom toward the pointer in the app; a keyboard equivalent for each | D-17 |
| E-07 | Dependency-free PMTiles tile source, from disk or by range request | D-18, AI-9 |
| E-08 | One-shot headless render, as a function and an app flag | D-13 |
| E-09 | Credits reported by the library; optional on-map credit line, on by default | D-25, S9-1 |
| E-10 | Host-supplied palette, changeable at runtime; themeable overlay ramps that stay ordered, distinct and readable | D-26, D-36 |
| E-11 | Overlays: features, scalar grids, vector grids, georeferenced images, tile-image providers | D-14, D-15 |
| E-12 | Simplification of host shapes, cached per zoom bucket, off the drawing path; host geometry borrowed, not copied; nothing drawn outside the rectangle | D-16, FR-11 |
| E-13 | Never-blank: stand-in tiles while detail loads | D-30, S7-1 |
| E-14 | No-colour forms by kind of data: contours for smooth fields, block shades for patchy data; basemap thinned | D-34, D-35 |
| E-15 | Placed images re-coloured through a required colour-to-intensity table; legend data exposed for every overlay | D-36, S1-6 |
| E-16 | Legibility at four colour depths with a ramp per depth | D-23, S10 |
| E-17 | Style profiles: the basemap thins by renderer, size and what is drawn on top; water owns its cells under a field | S1-2, S2-1, S5-1, S3-4 |
| E-18 | The view described as data for every place the host names — inside or outside, distance and bearing, band, intensity, valid time, staleness; speakable | D-52, D-67, FR-29 |
| E-19 | A valid time on every overlay, with a stale mark | FR-32 |
| E-20 | A distance reference: ground distance per column and per row, and a scale mark | FR-33 |
| E-21 | Basemap layers switched off and on by intent | D-42, FR-36 |
| E-22 | Image overlays as a sequence of timed frames, built after v0.1.0 | D-47, FR-37 |
| E-23 | Reduce-motion; nothing hidden by a frozen clock | D-56, NFR-21 |
| E-24 | Focus intents: next, previous, zoom around the focused target | FR-24a |
| E-25 | A ramp checker a host can run, and a safe-ramps setting that returns every overlay to its preset | D-53, D-63, FR-15, FR-16 |
| E-26 | The palette as a documented contract of semantic tokens | D-63, FR-15 |
| E-27 | Overlay presets — temperature, radar and precipitation, alert areas, wind — fully defined and overridable; the temperature preset an absolute scale anchored at freezing | D-62, D-68, D-69, FR-16 |
| E-28 | A painted ground by default, with an opt-out in which the host declares it | D-64, FR-20 |
| E-29 | An intent that fits the view to named places and overlays, with a margin in cells | D-76, FR-24 |

## Rows as read, before the D-61 split

The five rows below are the verified copy of the source read, unaltered, kept so that the copy can still be checked against research report AI-1 §3. Their halves above quote parts of the same text.

| ID | Behaviour | Where | Semantics to match |
|---|---|---|---|
| P-03 | Empty cell | braille.rs:299-302 | In braille mode it emits U+2800, not a space. In block mode it emits `' '` |
| P-05 | Colour depth | utils.rs:76; braille.rs:240-246 | xterm-256 only, as `38;5;n` / `48;5;n`. There is no truecolor or 16-colour path. Index 0 means "unset" |
| P-59 [LIB] | Marker animation | marker.rs:97-118 | Blink: `(tick/8)%2==0`. Flash: `(tick/3)%2==0`. Pulse: radius 1,2,3,4,3,2 advancing every 4 ticks, for Ring only (renderer.rs:419) |
| P-68 [APP] | Keys | main.rs:77-143 | q/Esc quit. a/+ zoom in. z/y/- zoom out. Arrows/hjkl pan. c braille. n labels. o ocean. w fit world. g globe tour. t marker tour at zoom 5.0. m toggles the demo markers |
| P-72 [APP] | Chrome | main.rs:206-226 | Help row at `rows-2`. Status row at `rows-1`, with `>> label` and `[TOUR…]` |
