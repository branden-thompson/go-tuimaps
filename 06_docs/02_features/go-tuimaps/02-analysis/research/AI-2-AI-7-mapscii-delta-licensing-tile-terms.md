# AI-2 / AI-7 — MAPSCII delta, licensing obligations, tile-provider terms

| Field | Value |
|---|---|
| Phase | DISCOVER — Tier 1 research |
| Date | 2026-09-18 (web facts observed this date) |
| Scope | What MAPSCII has that TerminalMap lacks; what the licensors document as obligations; OpenFreeMap and OpenStreetMap terms; alternative tile sources (OQ-2, OQ-11, C-3, C-4, RS-6) |
| Status | Obligations are reported as documented by the licensors. This is not legal advice; items marked "Owner's judgment" are for HUM LEAD to rule on. |
| Verification | Spot-checked by the coordinator against the pinned clones: both style files are byte-identical across the two upstreams; the shared default coordinates; no user-facing attribution string anywhere in TerminalMap's src/. All held. One row corrected — see the Labels row. |

Paths: `M/` = a clone of github.com/rastapasta/mapscii at 4fe9a60a0c9da952dadc5214a9ca5c68c447fdf8; `T/` = a clone of github.com/psmux/TerminalMap at 3b960723b24b781a1a9a04c165803cd1e98e6e3c. Web facts were observed on 2026-09-18.

## 1. MAPSCII delta

| Capability | MAPSCII | TerminalMap | Add to parity matrix? |
|---|---|---|---|
| Default tile source | `http://mapscii.me/`, osm2vectortiles data from 2016 (M/src/config.js:4-6) | OpenFreeMap TileJSON (T/src/config.rs:27; T/src/tile_source.rs:101-143) | **No.** `mapscii.me/0/0/0.pbf` now returns `text/html`, not a tile. Keep TerminalMap's direct-prefix mode (tile_source.rs:103). |
| Local MBTiles | Optional `@mapbox/mbtiles` (M/src/TileSource.js:19-25, 52-58, 125-134) | None: "HTTP tile servers and TileJSON" (tile_source.rs:15) | **Yes.** It is the cheapest real offline option, and OpenFreeMap publishes weekly MBTiles. |
| Style format | Mapbox GL with `constants`, loaded from a path (M/main.js:60-65; M/src/Mapscii.js:83) | The same JSON files, byte-identical to MAPSCII's (compared with `cmp`), passed as the `style_data` string (config.rs:6) | **Yes, trivial.** Accept both a path and bytes. |
| Mouse | Drag to pan (Mapscii.js:170-203), click to centre (:123-136), scroll zooms toward the pointer (:138-168) | Scroll only, zooming toward the centre (T/README.md:113; T/src/main.rs:149-153) | **Yes.** Expose `DragBy` and `ZoomAt(col,row)` methods so the host keeps ownership of input. This is the largest UX gain available. |
| Telnet | Described in the README only (M/README.md:9-11). The repo has no server code; the hook is injectable `input`/`output` (config.js:51-52). | None | **No.** Rendering to an `io.Writer` would allow it later. |
| Zoom step | 0.2 (config.js:14); `minZoom` is derived from the width (Mapscii.js:99) | 0.2 (config.rs:9), plus `fit_world` (README:400) | No difference. |
| Labels | rbush `LabelBuffer` (M/src/LabelBuffer.js:27-39). Per-layer `margin` and `cluster`, falling back to the POI marker (config.js:35-49; M/src/Renderer.js:223-236). Uses `string-width` for wide glyphs (LabelBuffer.js:53). | `rstar` is declared (T/Cargo.toml:36) but unused; collision is a linear rectangle scan (T/src/label.rs:28-70). One global `label_margin` (config.rs:17); no per-layer settings in `MapConfig`. *(Corrected 2026-09-18 against AI-1's full read of src/.)* | **Yes.** Add per-layer margin/cluster and rune-width-aware boxes. |
| POI hover | `featuresAt` is a stub that never returns a value (LabelBuffer.js:41-43); listed as a TODO (README:108-110) | Not present | **No.** |
| Label language | `name_<lang>`, then `name_en`, then `name`, then `house_num` (M/src/Tile.js:89) | The same, plus `name:xx` normalisation (T/src/tile.rs:75-78, 249-255) | No difference. |
| Headless and CLI | `--headless`, `--width`, `--height`, lat/lon/zoom, source and style flags (main.js:14-65; Mapscii.js:44-47, 96) | No argument parsing found in `main.rs` (checked by grep only); the library's `render()` returns a String (README:141) | **Yes, cheap.** A one-shot render gives a golden-file test harness. |
| Host callbacks | `keyCallback`, `mouseCallback`, `quitCallback`, `onUpdate` (Mapscii.js:174, 210, 217, 280) | The host owns input (README:196) | **No.** TerminalMap's model fits Bubble Tea. |
| In-memory cache | Eviction is broken (TileSource.js:90-95) | LRU of 32 entries, plus a disk cache keyed by source hash (tile_source.rs:42, 54) | **No.** |

Do not port these MAPSCII defects:
- The label sort reads a mistyped property `sorty` (Renderer.js:161).
- The `all` and `none` filter operators appear inverted (M/src/Styler.js:84-86, 100-102).
- Zoom stops are ignored (Renderer.js:195-196; Tile.js:80-83).

## 2. Licensing obligations (as documented by the licensors; not legal advice)

**Certain**
- MAPSCII: `"license": "MIT"` (M/package.json:31). "Copyright (c) 2017 Michael Straßburger / Copyright (c) 2019 The MapSCII authors" (M/LICENSE:3-4). Authors are listed in M/AUTHORS.
- TerminalMap: `license = "MIT"` (T/Cargo.toml:7). "Copyright (c) 2026 psmux" (T/LICENSE:3).
- Both licences carry the same condition: "The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software."
- TerminalMap credits MAPSCII only as "Inspired by" (README:617) and does not reproduce MapSCII's notice. It nonetheless ships MAPSCII's style files byte-identical and uses the same default coordinates (config.js:20-21 = config.rs:32-33). go-tuiMaps should carry both upstream notices itself rather than rely on TerminalMap's chain.
- Recommended layout:
  - `LICENSE`: the owner's MIT text.
  - `THIRD_PARTY_NOTICES`: both upstream MIT texts with their copyright lines, plus the data notices below.
  - README Credits section.

**Styles**
- The style JSONs contain no licence metadata; their only keys are `name`, `constants` and `layers`.
- 64 of the 66 layer ids in `bright.json` match Mapbox Open Styles `bright-v8.json`. Derivation from Mapbox is an inference; the file history is UNVERIFIED because the clone is shallow.
- The Mapbox Open Styles licence says: "copyright (c) 2014, Mapbox". The JSON is "licensed under the BSD license" and the design is under CC BY 3.0: "Attribution need not be provided on map images, but should be reasonably accessable" (https://github.com/mapbox/mapbox-gl-styles/blob/master/LICENSE.md).
- Neither upstream carries this notice.
- **Owner's judgment:** either add the Mapbox BSD notice, or write fresh styles directly against OpenMapTiles layers. The second option also removes the remap shim at tile.rs:71-132.

**Embedded tiles**
- They are described as "OpenMapTiles-schema PBF tiles from OpenFreeMap" (T/src/embedded_tiles.rs:2). The layers observed are boundary, landcover, place, water and water_name, with no attribution strings inside the tiles.
- OpenMapTiles: "Products or services using maps derived from OpenMapTiles schema need to visibly credit 'OpenMapTiles.org' or reference 'OpenMapTiles'" under CC-BY 4.0 for the design (https://github.com/openmaptiles/openmaptiles/blob/master/LICENSE.md).
- OSMF, for databases: "You must include attribution to OpenStreetMap and either the text of the ODbL or a link to it… in a location (such as a relevant directory)… such as a readme file" (https://osmfoundation.org/wiki/Licence/Attribution_Guidelines).
- **Owner's judgment:** whether five z0–1 tiles count as a Derivative Database, a Produced Work or an insubstantial extract. A NOTICE file in the tiles directory costs little and covers all three readings.

## 3. OpenFreeMap terms

Sources: https://openfreemap.org/, https://openfreemap.org/tos/, https://github.com/hyperknot/openfreemap (README).

- **Key and limits:** "there are no limits on the number of map views or requests. There's no registration, no user database, no API keys, and no cookies."
- **Commercial use:** "Is commercial usage allowed? Yes."
- **Attribution:**
  - "Attribution is required… If you are using alternative clients… you must add the following attribution: OpenFreeMap © OpenMapTiles Data from OpenStreetMap."
  - "You do not need to display the OpenFreeMap part, but it is nice if you do."
- **SLA and availability:**
  - "I don't offer SLA guarantees or personalized support."
  - The ToS says: "We aim to maintain the Site's availability but may discontinue it at any time without notice."
  - The service is provided "AS-IS".
- **Automated collection:** the ToS forbids "Attempt to collect data from the service in automated ways without permission."
- **Bulk data:**
  - "weekly full planet downloads both in Btrfs and MBTiles formats"
  - URL pattern: `https://btrfs.openfreemap.com/areas/planet/{version}/tiles.btrfs.gz (and .mbtiles)`
- **Caching:** tiles are served with `cache-control: public, max-age=315360000` (observed header).
- **Self-hosting:** "You can either self-host or use our public instance. Everything is open-source." The repo is MIT.
- **Schema:** "The map schema is unmodified OpenMapTiles."
- **TileJSON** (observed at `https://tiles.openfreemap.org/planet`):
  - `tilejson 3.0.0`
  - `tiles: [".../planet/20260913_164504_pt/{z}/{x}/{y}.pbf"]`
  - `minzoom 0`, `maxzoom 14`, `version 3.16.0`
  - 16 `vector_layers`
  - an HTML `attribution` string
  - "`/planet/latest`… always points to the latest deployed TileJSON."
- **Update cadence:** "Planet is generated weekly, every Wednesday. Set-latest runs every Saturday."
- **Roadmap:** "Future: Migrate to Shortbread schema and possibly VersaTiles."

## 4. OSM attribution

Source: https://osmfoundation.org/wiki/Licence/Attribution_Guidelines (adopted 2021-06-25).

- "Attribution must be to 'OpenStreetMap'"
- "'© OpenStreetMap contributors' or '© OpenStreetMap' are acceptable."
- "The attribution format should not require individuals to interact with the map… to see the attribution."
- For interactive maps:
  - "the credit should typically appear in a corner of the map… Alternatively, the attribution may be placed adjacent to the map or on a splash screen or pop-up shown when a user starts the app."
  - It may collapse "automatically on map interaction… automatically after five seconds".
  - After that, "the user must still be able to find the licence information… for example… an 'About' option in a menu."
- "If attribution is presented to the user upon application startup, it does not need to be presented… every time."
- There is no carve-out for small displays. The only one is "Small thumbnails/icons do not require attribution."
- osm.org/copyright says: "In media where links are not possible… include the full URL".

Compliant approach for a terminal:
- Showing attribution only on an About screen is outside the safe harbour, because it requires the user to interact.
- By default, put `© OpenMapTiles © OpenStreetMap` (30 columns, fits a 40-column map) on the map's bottom row or on an adjacent status line. It may fade after 5 seconds or on the first pan.
- Add an About screen with the full text, `https://www.openstreetmap.org/copyright` and the ODbL link. Use OSC-8 hyperlinks where the terminal supports them.
- For the library, expose `Attributions() []string`, seeded from the TileJSON, so that Watchpost can place them and weather-data credits can be appended.

## 5. Alternative sources

| Source | Key? | Free tier | Works with TerminalMap styles/remap? | Offline / self-host |
|---|---|---|---|---|
| Self-hosted OpenFreeMap, or any server over OpenFreeMap's MBTiles | No | Your own infrastructure | Exact match (OpenMapTiles) | Weekly planet MBTiles. `tileserver-gl` and `martin` specifics are UNVERIFIED. |
| Stadia | Yes off localhost: "you can get started without any API keys" applies only to localhost | 200,000 credits/month; "Commercial use not allowed" (https://stadiamaps.com/pricing/) | "compatible with OpenMapTiles"; TileJSON at `tiles.stadiamaps.com/data/openmaptiles.json`; z0–14 (https://docs.stadiamaps.com/vector/) | No |
| MapTiler | Registration required; key is UNVERIFIED but expected | 100,000 requests/month; non-commercial; "MapTiler logo on the map" (https://www.maptiler.com/cloud/pricing/) | OpenMapTiles, since MapTiler authors the schema | Paid data |
| Protomaps | Hosted API: "requires an API key… free for non-commercial use" (https://protomaps.com/api) | — | **No.** Its layers (earth, roads, pois…) are derived from Tilezen (https://docs.protomaps.com/basemaps/layers). | Best single-file option. Daily PMTiles builds, z0–15, about 120 GB, with `pmtiles extract`. "hotlinking… discouraged" (https://docs.protomaps.com/basemaps/downloads). |
| VersaTiles | Public server terms UNVERIFIED | — | **No.** It uses the Shortbread schema. | A 62 GB planet in `.versatiles`, `.pmtiles` and `.mbtiles`, with bbox extracts (https://download.versatiles.org/) |

**Recommendation**
- **Default:** the OpenFreeMap public instance. It is the only option that is key-free, OpenMapTiles-schema and allows commercial use.
- **Fallback:** a local MBTiles file plus a user-set URL pointing at self-hosted OpenFreeMap or any other OpenMapTiles server. This is not a second public service, because no other no-key OpenMapTiles service was found.
- **Strongest counter-argument:** the default is a donation-funded service run by one maintainer that "may discontinue… without notice" and says it intends to move to Shortbread. Every shipped binary could break at once.
- **Mitigation:** make schema remapping an interface keyed on the TileJSON `vector_layers`, and consider first-class Shortbread support early.

## 6. Risk notes for the owner

1. **Style provenance.** See section 2. Decide between adding the Mapbox BSD/CC-BY notice and writing new styles.
2. **Embedded tiles.**
   - No OpenFreeMap term was found that forbids redistribution, and OpenFreeMap itself offers the planet for download.
   - The ToS bar on "automated" collection does apply to the public instance.
   - Source the z0–1 tiles from the weekly MBTiles, record the version, and ship the ODbL and OpenMapTiles notices.
   - A feature that pre-fetches a region for offline use against the public instance would need OpenFreeMap's permission. Restrict it to self-hosted or MBTiles sources.
3. **Parity is not compliance.** Grep finds no attribution string anywhere in `T/src/*.rs`. go-tuiMaps needs on-screen attribution even though upstream has none.
4. **OpenFreeMap schema migration.** See section 5.
5. **Cache staleness.**
   - The disk cache never expires and is keyed by source only (tile_source.rs:33-42, 152-165), while OpenFreeMap tiles are versioned weekly.
   - Embedded tiles always take priority over the network (tile_source.rs:76).
   - Decide on a TTL or version-keyed cache.
6. **`mapscii.me` is dead as a tile source.** Do not offer it as a preset.
