# AI-1 — Upstream module read: TerminalMap @ 3b96072 (v0.1.0)

| Field | Value |
|---|---|
| Phase | DISCOVER — Tier 1 research |
| Date | 2026-09-18 |
| Scope | Raw material for the frozen parity matrix (OQ-1, M3) |
| Verification | Six claims spot-checked against the pinned source by the coordinator: unused rstar/ratatui crates, draw order, serial tile awaits, no HTTP status check, block-glyph mask mismatch, 256-colour-only output. All six held. |

Paths are relative to a clone of github.com/psmux/TerminalMap at commit 3b960723b24b781a1a9a04c165803cd1e98e6e3c. Every src/ file was read in full. HEAD was checked and matches the pinned commit.

## 1. MODULE MAP

| File | Lines | Responsibility and key items |
|---|---|---|
| src/braille.rs | 452 | Pixel-to-cell buffer with 2x4 dots per cell. Holds the per-cell fg/bg/char state, the per-pixel colour vote, the ANSI frame encoder and 5 unit tests. `BrailleBuffer::{new, clear, set_pixel, set_pixel_forced, set_char, write_text, set_background, set_global_background, frame}` |
| src/camera.rs | 326 | Tick-driven waypoint tours. `Waypoint`, `Camera`, `ease_in_out`, `lerp_lon` |
| src/canvas.rs | 249 | Drawing primitives over the buffer. `Point`, `Canvas::{polyline, line, polygon, text, background, set_background, frame, clear}` |
| src/config.rs | 45 | `MapConfig` and its `Default` |
| src/embedded_tiles.rs | 22 | `get(z,x,y)` over 5 `include_bytes!` gz tiles |
| src/label.rs | 72 | `LabelBuffer::write_if_possible`, a linear-scan rectangle collision test |
| src/lib.rs | 14 | Declares every module `pub` |
| src/main.rs | 227 | Standalone crossterm app: event loop, keys, footer, demo markers |
| src/marker.rs | 118 | `MapMarker`, `MarkerShape`, `MarkerAnimation`, `marker_visible`, `pulse_radius` |
| src/proto.rs | 68 | Hand-written prost structs for MVT 2.1 |
| src/renderer.rs | 501 | `Renderer::{new, set_size, draw (async), frame, draw_markers}`. Covers tile selection, culling, draw order, feature drawing and label placement |
| src/styler.rs | 261 | `Styler::new`, `get_style_for`, `StyleLayer::applies_to`, the filter evaluator, `@constant` substitution, `ref` resolution |
| src/tile.rs | 518 | `load_tile`: gunzip, protobuf decode, OpenMapTiles-to-Mapbox-v6 remap, styling at parse time, geometry decode, ring classification |
| src/tile_source.rs | 166 | `TileSource::{new, get_tile (async)}`: LRU, embedded tiles, disk cache, HTTP/TileJSON |
| src/utils.rs | 131 | `base_zoom`, `tilesize_at_zoom`, `ll2tile`, `tile2ll`, `normalize`, `digits`, `hex2rgb`, `rgb_to_x256`, `population`, `simplify_points` (RDP) |
| src/widget.rs | 289 | `MapState`, the embedder API |

Style files (`styles/dark.json`; `bright.json` was checked for comparison):
- `dark.json` has the keys `name`, `constants` (75, all `#rgb`/`#rrggbb`) and `layers` (75: 46 line, 24 symbol, 4 fill, 1 background).
- The filter operators used are `==`, `all`, `in`, `!in`, `>=` and `!=`.
- The paint properties are `line-color`, `fill-color`, `text-color` and `background-color`.
- There are no `line-width` properties, no function stops, no `layout` and no `ref`.
- `minzoom` appears on 12 layers (5, 11, 12, 13–16, 15.5). `maxzoom` appears on none.
- `bright.json` adds `line-width` `{base, stops}` on 34 layers, and `text-size`, which is never read.

## 2. DATA FLOW

1. The host calls `MapState::render` (widget.rs:100). It locks the renderer, calls `set_size`, which allocates a new `Canvas` (renderer.rs:55-59), and copies `show_labels` and `ocean_background` across (widget.rs:103-104).
2. `Renderer::draw` (renderer.rs:62) clears the labels and the canvas.
3. `visible_tiles` (renderer.rs:120) selects tiles:
   - `z = base_zoom` (utils.rs:9) and `tile_size = tilesize_at_zoom` (utils.rs:15).
   - The centre comes from `ll2tile`.
   - The tile range is ±(`ceil(w/2/ts)+1`).
   - Tiles outside the grid or off-screen are dropped.
4. The world bounds are computed. When the ocean option is on, every pixel outside them is set to colour 69 (renderer.rs:80-91).
5. Tiles are fetched sequentially with `TileSource::get_tile().await` per tile, and errors are dropped (renderer.rs:94-102). The lookup order is LRU, then embedded, then disk, then `fetch_http` plus `persist_tile` (tile_source.rs:64-96). `resolve_tile_url` (tile_source.rs:101) handles TileJSON.
6. `tile::load_tile` (tile.rs:51) runs `decompress_if_needed`, `proto::Tile::decode` and `load_layers`.
   - Per feature it decodes tags, runs `remap_openmaptiles`, then `Styler::get_style_for` (first match wins, otherwise the feature is dropped).
   - It converts the colour to x256 and runs `decode_geometry`.
   - Fills go through `classify_rings`. A bounding box is computed for each feature.
   - Styling is baked into the cached tile.
7. `get_tile_features` (renderer.rs:165) culls by bounding box per layer in draw order.
8. `render_tiles` (renderer.rs:206) loops layer, then tile, then feature:
   - Layers whose name contains `label` or `symbol` are deferred.
   - The rest go to `draw_feature` (renderer.rs:242).
   - `draw_feature` applies the zoom gate, then `scale_and_reduce` (renderer.rs:302), then `Canvas::polyline` or `Canvas::polygon` (earcut, then `filled_triangle`).
9. Labels are sorted by `sort`. Each one is placed if `LabelBuffer::write_if_possible` allows it, then drawn with `Canvas::text` (renderer.rs:233-239, 270-296).
10. `draw_markers` (renderer.rs:366) projects inline and draws with `set_pixel_forced`. Marker labels go through the same label buffer.
11. `BrailleBuffer::frame` (braille.rs:251) emits, per cell, the colour from `resolve_cell_color` via `term_color`, then a char, a braille glyph or a block glyph. Rows are joined with `\r\n` and the frame ends with `ESC[39;49m`.

## 3. PARITY MATRIX ROWS

| ID | Behavior | Where | Semantics to match |
|---|---|---|---|
| P-01 | Pixel grid | braille.rs:56,209 | The cell index is `(x>>1)+(w>>1)*(y>>2)`. There are `w/2 × h/4` cells |
| P-02 | Dot bitmask | braille.rs:10-15 | Rows × cols: `[01,08],[02,10],[04,20],[40,80]`. The glyph is U+2800+mask |
| P-03 | Empty cell | braille.rs:299-302 | In braille mode it emits U+2800, not a space. In block mode it emits `' '` |
| P-04 | Block ("ASCII") mode | braille.rs:23-32,213-230 | Six glyphs: ▀ ▄ ■ ▌ ▐ █. The one with the highest `popcount(mask & bits)` wins, and the first one wins ties. The glyphs are Unicode blocks, not ASCII |
| P-05 | Colour depth | utils.rs:76; braille.rs:240-246 | xterm-256 only, as `38;5;n` / `48;5;n`. There is no truecolor or 16-colour path. Index 0 means "unset" |
| P-06 | SGR forms | braille.rs:239-247 | fg+bg, `49;38;5;fg`, `39;48;5;bg`, or `39;49`. Emitted only when the colour changes |
| P-07 | Row self-containment | braille.rs:258-266,310 | Colour is re-emitted at the start of each row. Rows are joined with `\r\n`. The frame ends with `ESC[39;49m` |
| P-08 | Cell colour vote | braille.rs:134-207 | The cell takes the majority colour among its lit pixels. A tie is broken by the count of that colour among lit pixels in the 8 neighbouring cells. Locked cells use the last fg |
| P-09 | Forced pixels | braille.rs:115-129 | `set_pixel_forced` locks the whole cell's colour |
| P-10 | Text cells | braille.rs:275,286-294,314 | A char takes priority over dots. It uses the cell fg and bypasses the vote |
| P-11 | Wide chars | braille.rs:288-306,432 | The cells after a CJK char are skipped. The width table is hand-rolled |
| P-12 | Text placement | braille.rs:324-336 | One char per cell, at `x+i*2` px. The centre offset is `len/2+1` (byte length) |
| P-13 | Thin line | canvas.rs:154 | Bresenham when `width<=1` |
| P-14 | Thick line | canvas.rs:103-151 | Zingl algorithm, `wd=(w+1)/2` with `w=width-1` |
| P-15 | Polygon fill | canvas.rs:57-93,208 | `earcutr::earcut` with holes. Each triangle is filled by taking the Bresenham edge points and spanning min to max x per row |
| P-16 | Degenerate rings | canvas.rs:62-69 | An outer ring with fewer than 3 points aborts the polygon. A hole with fewer than 3 points is skipped |
| P-17 | Tile zoom | utils.rs:9-12 | `clamp(floor(zoom),0,tile_range)`. `tile_range` is 14 |
| P-18 | Tile pixel size | utils.rs:15-18 | `project_size*2^(zoom-z)`. `project_size` is 256. Overzoom above 14 scales up to 4096 px at z18 |
| P-19 | Visible tiles | renderer.rs:130-150 | The range is `±(ceil(dim/2/ts)+1)`. Tiles with x or y outside the grid are dropped, so the world does not wrap |
| P-20 | Projection | utils.rs:21-35 | Standard Web Mercator `ll2tile`/`tile2ll` |
| P-21 | Normalize | utils.rs:38-49 | Longitude wraps by a single ±360. Latitude is clamped to ±85.0511 |
| P-22 | Ocean outside world | renderer.rs:80-91 | Lit pixels with hard-coded colour 69, outside the floor/ceil world bounds |
| P-23 | Draw order, z≥2 | renderer.rs:483-498 | landuse, water, marine_label, building, road, admin, then 8 label layers |
| P-24 | Draw order, z<2 | renderer.rs:480-481 | water, landuse, admin, country_label, marine_label |
| P-25 | Label deferral | renderer.rs:223-239 | Decided by layer name containing `label` or `symbol`. Labels are stable-sorted ascending by `sort`. `show_labels=false` also hides POI glyphs |
| P-26 | Sort key | tile.rs:264-268 | `localrank`, else `scalerank`, else 0. Integers only |
| P-27 | Zoom gate | renderer.rs:243-252 | Style `minzoom`/`maxzoom` are compared against the fractional map zoom. The test is inclusive |
| P-28 | Feature cull | renderer.rs:183-193 | The bounding box is tested against the viewport in tile units |
| P-29 | Scaling | renderer.rs:321-326 | `floor(pos + p/scale)`. Consecutive duplicate points are dropped |
| P-30 | Line clip pad | renderer.rs:50,314-343 | 64 px padding. Runs of points outside it are collapsed. Fills are not clipped |
| P-31 | Simplify | renderer.rs:350-352; config.rs:34 | RDP with ε=0.5. Off by default |
| P-32 | Label anchor | renderer.rs:286 | `x = px - text.len()` (byte length). Each vertex is tried in turn until one fits |
| P-33 | Label bounds | renderer.rs:278-284 | The anchor must be inside both the world bounds and the canvas |
| P-34 | Collision | label.rs:28-70 | Works in cell coordinates (`x/2`, `y/4`). The rectangle is `[x-m, x+m+chars] × [y±m/2]`, overlap is inclusive, `m=5`. The scan is linear |
| P-35 | POI glyph | renderer.rs:271; config.rs:40 | A symbol with no name draws `◉` |
| P-36 | Label language | tile.rs:249-259 | Lookup order: `name_<lang>`, `name:<lang>`, `name_en`, `name:en`, `name`, `house_num` |
| P-37 | Gzip sniff | tile.rs:58-67 | Detected by the magic bytes `1f 8b` |
| P-38 | MVT decode | tile.rs:372-438; proto.rs | MoveTo, LineTo, ClosePath (which re-pushes the first point) and zigzag. The default extent is 4096 |
| P-39 | Ring grouping | tile.rs:443-484 | Signed area ≥0 starts a new polygon. Negative area is a hole in the previous polygon. Each polygon becomes its own feature |
| P-40 | OMT remap | tile.rs:71-162 | transportation→road, with `_link` for ramps, minor→street and brunnel→structure. boundary→admin. place→country/place_label. water_name→marine/water_label. poi→poi_label. park→landuse_overlay. landcover→landuse. `name:xx`→`name_xx`. rank→scalerank or labelrank |
| P-41 | Style match | styler.rs:229-237; tile.rs:207-213 | The first matching layer, in style order, for the source-layer. The remapped name is tried first, then the raw name. No match drops the feature. One style per feature |
| P-42 | Filter ops | styler.rs:32-118 | Supported: all, any, none, ==, !=, in, !in, has, !has, >, >=, <, <=. An unknown op evaluates to true. `==` on a missing key is false. Equality is on JSON values, so integers and floats are distinct. `$type` is injected as a property (tile.rs:187) |
| P-43 | Constants and `ref` | styler.rs:132-181,240 | `@name` strings are substituted. `ref` inherits type, source-layer, zooms and filter |
| P-44 | Colour pick | tile.rs:216-242 | `line-color`, else `fill-color`, else `text-color`. For stops only the first stop is used. The fallback is `#f00`. Hex is 3 or 6 digits only (utils.rs:58) |
| P-45 | Line width | tile.rs:271-289 | A number or the first stop. The default is 1 |
| P-46 | LRU | tile_source.rs:53-55,65 | 32 parsed tiles, keyed `z-x-y`, per `MapState` |
| P-47 | Embedded tiles | embedded_tiles.rs:13-21; tile_source.rs:76 | z0 (1 tile) and z1 (4 tiles), gzipped. They are consulted before disk and HTTP, whatever `source` is |
| P-48 | Disk cache | tile_source.rs:42,152-165 | `<cache_dir>/terminalmap/<hex DefaultHasher(source)>/<z>/<x>-<y>.pbf`. Raw bytes, no expiry, no size cap |
| P-49 | URL modes | tile_source.rs:101-143 | A source ending in `/` becomes `{src}{z}/{x}/{y}.pbf`. Otherwise it is TileJSON: `tiles[0]` is used and the template is memoised |
| P-50 | HTTP | tile_source.rs:59,145-150 | Default reqwest client. No timeout, no User-Agent, no status check, no retry |
| P-51 | Fetch failure | renderer.rs:95-101 | The error is swallowed and the tile is left blank. There is no negative cache |
| P-52 [LIB] | Config defaults | config.rs:23-44 | See §4 |
| P-53 [LIB] | Size from terminal | widget.rs:92-97 | `w=(cols>>1)<<2`, `h=(rows-3)*4` |
| P-54 [LIB] | Min zoom | widget.rs:83-88 | `min(log2(w/256), log2(h/256))`. Zoom is clamped up to it |
| P-55 [LIB] | zoom_by / initial zoom | widget.rs:61,124-133 | Clamped to [min_zoom, max_zoom=18]. `None` means 0.0 |
| P-56 [LIB] | fit_world | widget.rs:166-192 | Latitude 84 to −56. `zoom=min(log2(h/span_px), log2(w/256))`. Longitude 0, latitude at the Mercator midpoint |
| P-57 [LIB] | Footer | widget.rs:195-202; utils.rs:52 | `center: lat, lon   zoom: z`, truncated with floor |
| P-58 [LIB] | Marker shapes | renderer.rs:395-440 | Dot is 3×3. Cross is ±3. Diamond has radius 3. Ring(r) uses a midpoint circle. FilledCircle(r). Char |
| P-59 [LIB] | Marker animation | marker.rs:97-118 | Blink: `(tick/8)%2==0`. Flash: `(tick/3)%2==0`. Pulse: radius 1,2,3,4,3,2 advancing every 4 ticks, for Ring only (renderer.rs:419) |
| P-60 [LIB] | Marker cull and label | renderer.rs:389,443-450 | ±20 px cull. The label sits at `px+4` and is collision-checked after the map labels. It takes the marker colour |
| P-61 [LIB] | Marker id | marker.rs:62 | The default is `"{lat:.6},{lon:.6}"`. `remove_marker` removes every marker with that id |
| P-62 [LIB] | Camera easing | camera.rs:296-302 | Cubic ease-in-out. Defaults are 60 travel ticks and 40 hold ticks (camera.rs:31-32) |
| P-63 [LIB] | Camera zoom arc | camera.rs:252-259 | The midpoint is `max(min(from,to)-0.8, 0)`, with each half eased |
| P-64 [LIB] | Longitude path | camera.rs:310-326 | Shortest way round, result kept within ±180 |
| P-65 [LIB] | Globe tour | camera.rs:98-154 | 12 cities, Paris to Berlin. Travel 60–120 ticks, hold 50, looping. The default zoom is 2.0 (widget.rs:250) |
| P-66 [LIB] | Marker tour | camera.rs:157-173 | Travel 70, hold 60. The label is the marker label or id |
| P-67 [LIB] | Tour end | camera.rs:277-283 | A non-looping tour deactivates and returns None |
| P-68 [APP] | Keys | main.rs:77-143 | q/Esc quit. a/+ zoom in. z/y/- zoom out. Arrows/hjkl pan. c braille. n labels. o ocean. w fit world. g globe tour. t marker tour at zoom 5.0. m toggles the demo markers |
| P-69 [APP] | Pan step | main.rs:87-99 | Longitude ±8/2^zoom, latitude ±6/2^zoom degrees |
| P-70 [APP] | Mouse | main.rs:147-159 | Scroll only, ±zoom_step, centred on the map centre. There is no drag |
| P-71 [APP] | Loop | main.rs:71,174-183 | 50 ms poll. tick and camera advance on every loop iteration. Animation redraws are throttled to 50 ms or more. Resize triggers a redraw |
| P-72 [APP] | Chrome | main.rs:206-226 | Help row at `rows-2`. Status row at `rows-1`, with `>> label` and `[TOUR…]` |

## 4. PUBLIC API SURFACE

**MapState** (widget.rs)
- Public fields: `center_lat`, `center_lon`, `zoom: f64`, `config`, `width`, `height: usize`, `tick: u64`.
- **async**: `new(MapConfig) -> Result<Self>`. It contains no awaits internally.
- **async**: `render(&self) -> Result<String>`.
- Sync, setup and view:
  - `set_size(usize,usize)`
  - `set_size_from_terminal(u16,u16)`
  - `zoom_by(f64)`
  - `move_by(dlat,dlon)`
  - `set_center(lat,lon)`
  - `fit_world()`
  - `footer()->String`
- Sync, toggles: `toggle_braille()`, `toggle_labels()`, `toggle_ocean_background()`.
- Sync, markers and animation:
  - `add_marker(MapMarker)`
  - `remove_marker(&str)`
  - `clear_markers()`
  - `markers()->&[MapMarker]`
  - `advance_tick()`
  - `has_animated_markers()->bool`
  - `needs_animation_redraw()->bool`
- Sync, camera:
  - `camera()->&Camera`
  - `camera_mut()->&mut Camera`
  - `start_globe_tour()`
  - `start_globe_tour_at(f64)`
  - `start_marker_tour(f64)`
  - `toggle_camera()->bool`
  - `update_camera()->bool`

**MapConfig defaults** (config.rs:23-44)

| Field | Default |
|---|---|
| `language` | `"en"` |
| `source` | `"https://tiles.openfreemap.org/planet"` |
| `style_data` | None (embedded dark.json) |
| `initial_zoom` | None |
| `max_zoom` | 18 |
| `zoom_step` | 0.2 |
| `initial_lat` | 52.51298 |
| `initial_lon` | 13.42012 |
| `simplify_polylines` | false |
| `use_braille` | true |
| `persist_downloaded_tiles` | true |
| `tile_range` | 14 |
| `project_size` | 256 |
| `label_margin` | 5 |
| `poi_marker` | `◉` |
| `show_labels` | true |
| `ocean_background` | true |

**MapMarker**
- Fields: `lat`, `lon`, `label`, `color: u8`, `shape`, `animation`, `id`.
- Constructors and builders:
  - `dot(lat,lon,u8)`
  - `dot_rgb(lat,lon,r,g,b)`
  - `with_label`, `with_animation`, `with_shape`, `with_id`

**Camera**
- `new()`, `globe_tour(f64)`, `from_markers(&[MapMarker],f64)`
- `add_waypoint`, `set_zoom`
- `start(lat,lon,zoom)`, `stop`, `toggle->bool`
- `is_active`, `current_label->Option<&str>`
- `tick->Option<(lat,lon,zoom)>`
- The public field `looping`

**Waypoint**
- `new(lat,lon,zoom)`, `with_travel`, `with_hold`, `with_label`
- All fields are public.

**Also public**, because lib.rs exports every module:
- `Renderer::{new, set_size, draw (async), frame, draw_markers}`
- `TileSource::{new, get_tile (async)}`
- `Canvas`, `BrailleBuffer`, `Styler`, `utils::*`

## 5. CONCURRENCY & RUNTIME MODEL

- `render` holds a `tokio::sync::Mutex<Renderer>` for the whole frame (widget.rs:101).
- Tiles are awaited one at a time, in series (renderer.rs:94-102). Nothing fetches in parallel, prefetches, cancels or times out.
- Rendering is not progressive. `render` returns only after every visible tile has resolved or failed. A cold view at high zoom blocks for N sequential round trips.
- The standalone app awaits `draw_map` inside its event loop (main.rs:168,180), so input stalls while tiles are being fetched.
- Failed tiles are not negatively cached. The next frame pays for the failure again.
- Disk I/O uses blocking `std::fs` inside async code (tile_source.rs:155,163).
- `draw` never returns `Err` (renderer.rs:95). `render` fails only if locking or allocation fails.
- Animation is counted in ticks, not wall-clock time. The host has to call `advance_tick` and `update_camera`.

For the Go port this means parity only requires "render(view) returns a string". The message-driven host needs a split the upstream does not have: a synchronous render from cache, plus asynchronous "tile arrived" messages that trigger a re-render. Treat that as a deliberate extension, not a parity row.

## 6. THIRD-PARTY CRATES

| Crate | Use | Capability the port needs |
|---|---|---|
| crossterm | main.rs only: raw mode, alt screen, events | The host framework provides this |
| ratatui | **Unused**. No reference anywhere in `src/` | None |
| tokio | Runtime, `Mutex` | Native concurrency |
| reqwest (rustls) | HTTP GET | An HTTP client. Timeout and User-Agent need adding |
| prost | MVT protobuf decode | Protobuf or MVT decoding |
| bytes | **Unused** directly | None |
| flate2 | gunzip | gzip |
| serde / serde_json | Style, TileJSON and property values. `serde` derive is unused | JSON with distinct int and float numbers |
| rstar | **Unused**. Labels use a linear scan | None, or optionally an R-tree |
| earcutr | Triangulation | earcut with holes |
| ansi_colours | RGB to xterm-256 | A nearest-256 mapping that gives the same result. It must map `#5f87ff` to 69 |
| lru | Tile cache | LRU |
| dirs | OS cache directory | User cache directory |
| anyhow | Errors | None |

## 7. DEFECTS, TODOs, AND QUIRKS

1. **Block-mode masks do not match the braille bit layout** (braille.rs:25-29 vs 10-15). ▀ is given bits 1,2,16,32, which are left rows 0–1 plus right rows 1–2. The glyphs therefore misrepresent the pixels. Fix it. Counter-argument: block-mode golden frames would then diverge from upstream.
2. **Disk cache poisoning**. The response body is persisted with no status check (tile_source.rs:83,147-149), so a 404 or HTML body gets cached and fails to decode on every later render. Fix it. Counter-argument: none worth keeping the behaviour for.
3. **The clip re-entry bug**. `last_x`/`last_y` are overwritten before they are used (renderer.rs:327-339). The current point is pushed twice and the line jumps from the first outside point to the re-entry point, which can draw a spurious chord across the view. Fix it by pushing the previous outside point. Counter-argument: pixel parity.
4. **Styled layers that are never drawn**. waterway, aeroway, landuse_overlay (which `park` remaps to) and airport_label are all absent from the draw order (renderer.rs:479-500), so rivers and parks never appear. OMT `housenumber` is never remapped. Replicate this for the frozen matrix and add the layers later as a flagged extension. Counter-argument: users will see missing rivers and judge the port by it.
5. **Background is never applied**. The style's `background` layer has no source-layer, so it is dropped (styler.rs:213). `Canvas::set_background` and `Canvas::background` have no callers. `term_color` combines the cell bg and the global bg with a bitwise OR (braille.rs:234). Do not replicate the OR. Keep the bg buffer (see §8).
6. **Ocean colour is hard-coded to 69** (renderer.rs:81) and ignores a custom style. Derive it from the style's water colour. Counter-argument: with dark.json the result is identical, so parity is safe either way.
7. **Byte length is used for label centring** (renderer.rs:286; braille.rs:326), so non-ASCII labels are shifted left. Fix it by using the rune or cell width. Counter-argument: label positions would differ from upstream for non-Latin labels.
8. **Negative y cast to `usize`** (renderer.rs:290,438,445). The label reserves collision space but is never drawn. Guard against negative y.
9. **Only the first zoom stop is honoured** (tile.rs:226-232,278-284). Replicate it. dark.json has no stops, so nothing changes for the default style.
10. **Config drift**. The renderer holds a clone of the config. Only two fields are synced on each render (widget.rs:103-104), so later edits to `label_margin` or `poi_marker` are ignored. Fix it.
11. **Per-frame waste**. The canvas is reallocated and the block-glyph table is rebuilt on every render (renderer.rs:58; braille.rs:69). `ParsedTile` is deep-cloned on every cache hit (tile_source.rs:71), and features are cloned again afterwards (renderer.rs:194). Fix it. The output is unchanged.
12. **The hash used for the cache directory is not stable.** `DefaultHasher` (tile_source.rs:33-37) cannot be reproduced portably. Use the port's own directory scheme.
13. **Embedded tiles shadow custom sources at z0–1** (tile_source.rs:76). Gate them on the default source.
14. **Docs and comments disagree with the code**:
   - The README says mouse "pan". Mouse pan is not implemented (main.rs:147-159).
   - The README says `initial_zoom` None "fits the terminal". The code uses 0.0 (widget.rs:61).
   - A comment says `Dot` is a "single braille dot". It is 3×3 (marker.rs:19).
   - A comment on `fit_world` says 72°N. The code uses 84 (widget.rs:164-168).
   - A comment says "~200ms refresh". The code uses 50 ms (main.rs:174-178).
   - label.rs describes itself as "grid-based". It is a linear scan.
15. **The app's tick rate depends on input** (main.rs:71,175). Events shorten the poll, so animations speed up while keys are being pressed. Use wall-clock ticks in the Go host.
16. **There is no terminal cleanup on error** (main.rs:52-58,187-192).
17. **Hard-coded assumptions**:
   - 2×4 dots per cell with no cell-aspect correction.
   - 256 px base tile size, extent 4096, latitude clamp 85.0511.
   - Tiles above zoom 14 are overzoomed.
   - Longitude wraps only once (utils.rs:41-46).
   - The map does not repeat across the antimeridian.
   - 3 rows are reserved for the footer in the library method `set_size_from_terminal` (widget.rs:95).

   Replicate all of these except the 3-row reservation. An embedded widget should be given its exact rectangle.

## 8. OVERLAY SEAMS

**Seam A, area layers under labels.** Insert in `render_tiles` between the end of the geometry loop and `labels.sort_by` (renderer.rs:230-233). Host polygons and gridded fields drawn here cover the base geometry but leave place names readable.

**Seam B, point and vector glyphs on top.** Insert after `renderer.draw(...)` and beside `draw_markers` (widget.rs:107-118). This is the existing marker seam.

The projection has to be factored out first. It currently sits inline at renderer.rs:374-386 (`ll2tile`, then `w/2+(m-c)*tile_size`). Expose one `Project(lon,lat)->(px,py)` function and its inverse, and have the tiles, the markers and every overlay use it.

Colour-model constraints:
- Each cell has one 8-bit fg, one 8-bit bg and one optional char. 0 means "none".
- Per-pixel colour exists only as input to the majority vote (braille.rs:46,134). Two colours cannot coexist inside one cell.
- There is no alpha channel and no way to clear a pixel.
- Water and ocean cells are fully lit (renderer.rs:87). Over sea, the dot pattern therefore carries no information, and an overlay can only change colour there.
- Sparse overlay dots lose the majority vote to fills.
- `set_pixel_forced` wins, but it recolours every base-map dot in that cell.

Recommendation: put gridded fields (radar, temperature) into the dormant per-cell **background** channel (braille.rs:42,86; emission is already supported at 240-244). Keep the fg dots for coastlines and borders. This is the only way to show a field and the map together in one cell, and its resolution is one value per cell. Draw polygon outlines and wind glyphs as forced fg pixels or as chars at seam B. Widen the colour type to 24-bit at the buffer boundary, and keep an x256 quantiser for parity mode. Strongest counter-argument: background fills look heavy, vary with the terminal theme, and clash with the ocean, which is lit as fg colour 69. In practice, turning `ocean_background` off (or dimming it) whenever a field overlay is active is probably necessary.
