# AI-3 — Go ecosystem survey

| Field | Value |
|---|---|
| Phase | DISCOVER — Tier 1 research |
| Date | 2026-09-18 |
| Scope | Library-versus-hand-written choices for decode, fill, labels, styles, cache, output contract, geodesy, assets, testing (OQ-10; feeds PLAN) |
| Status | Recommendations are research input. Nothing here is decided until HUM LEAD rules in PLAN. |

All facts were checked on 2026-09-18. Anything marked UNVERIFIED was not confirmed against a source.

## 1. MVT / protobuf decoding
- **`github.com/paulmach/orb/encoding/mvt`** is MIT and pure Go. v0.13.0 was released 2026-03-30, with the latest commits the same day. About 1.1k stars, 14 open issues, 637 importers (https://pkg.go.dev/github.com/paulmach/orb, https://github.com/paulmach/orb).
  - Decoding uses `github.com/paulmach/protoscan` with a reusable decoder struct. protoscan is MIT with about 10 stars (https://github.com/paulmach/protoscan).
  - The package also imports `github.com/gogo/protobuf/proto`, whose README is headed "[Deprecated]" (https://github.com/gogo/protobuf). It imports `encoding/json` and `reflect` as well (https://pkg.go.dev/github.com/paulmach/orb/encoding/mvt?tab=imports).
  - orb's go.mod requires `go.mongodb.org/mongo-driver/v2` (https://raw.githubusercontent.com/paulmach/orb/master/go.mod). Whether module pruning keeps it out of a consumer's go.sum is UNVERIFIED.
  - Output is `geojson.Feature` with a `map[string]any` of properties and float64 geometries, so each feature costs a map plus slices.
  - Verdict: correct and proven, but the wrong shape for a hot path.
- **`github.com/go-spatial/geom/encoding/mvt`** is MIT. v0.1.0 dates from 2024-07-29 and the last commits from 2026-02-23. 188 stars, 20 open issues, and it does decode (https://pkg.go.dev/github.com/go-spatial/geom/encoding/mvt).
  - The module go.mod lists `golang/protobuf v1.3.2` and `mattn/go-sqlite3`, which needs CGO (https://raw.githubusercontent.com/go-spatial/geom/master/go.mod).
  - The mvt package has 15 imports.
  - Verdict: reject.
- Other candidates found:
  - `github.com/wisborg/osmbase/mvt` is Apache-2.0, v0.8.0, published today, with 0 importers, so it is too new (https://pkg.go.dev/search?q=mvt+vector+tile+decode).
  - `github.com/akhenakh/mvtgo` has no detected licence, so it is unusable (https://pkg.go.dev/github.com/akhenakh/mvtgo).
- `google.golang.org/protobuf` generated code was not examined (licence and size UNVERIFIED). It would allocate a struct per feature and add a large module for three message types.
- Hand-rolling needs only varint, length-delimited, fixed32 and fixed64 wire types, roughly 300 lines. It minimises allocations in four ways:
  - skip whole layers the style never references;
  - test the style filter before decoding a feature's geometry;
  - decode zig-zag commands straight into reusable int32 arenas;
  - intern keys and resolve values lazily, with no maps.

**Recommendation:** Hand-roll the decoder and use orb only as a test oracle (see §9). Hand-rolling is justified here because the format is frozen (MVT 2.1), tiny and read-only, it sits on the hot path, and every library's output type forces allocation. It would not be justified if encoding or general geometry operations were needed.

## 2. Polygon fill
- Earcut ports found:
  - `github.com/tchayen/triangolatte` is MIT. It has no go.mod, the pseudo-version is from 2021-08-04, and it has 37 stars. It self-reports triangulating "99.76%" of buildings, so it fails about 0.24% of the time (https://pkg.go.dev/github.com/tchayen/triangolatte).
  - `github.com/rclancey/go-earcut` is ISC, last published 2018-04-11, with 6 stars (https://github.com/rclancey/go-earcut).
  - `github.com/mmp/earcut` is ISC and pure Go, published 2026-08-25, with 0 stars (https://github.com/mmp/earcut).
  - `github.com/oliverbestmann/earcut-go` is v1.0.0 from 2025-11-24, ISC (https://pkg.go.dev/search?q=earcut). Whether it is pure Go is UNVERIFIED.
- Triangulation is a GPU idiom. On a canvas of about 160×96 sub-pixels it adds a degenerate-input failure class, seams between triangles, and an extra pass.
- Scanline even-odd fill over all rings of one polygon handles holes for free:
  - no winding analysis is needed, and the MVT spec already requires valid, non-overlapping rings;
  - it needs an integer active-edge table and cannot fail on bad input;
  - it is about 80 lines, and setting a bit twice is harmless where tile buffers overlap.

**Recommendation:** Do not triangulate. Hand-write a scanline even-odd fill, clipped to the canvas Y-range, plus Bresenham line drawing.

## 3. Label collision
- `github.com/tidwall/rtree` is MIT, v1.11.1 (2026-08-11), has no dependencies beyond the standard library, 349 stars and 2 issues, and offers a generic `RTreeG[T]` (https://pkg.go.dev/github.com/tidwall/rtree). It is good, but unnecessary here.
- `github.com/dhconnelly/rtreego` is BSD-3-Clause, v1.2.0 (2023-02-01), with 4 imports. It is N-dimensional and interface-based (https://pkg.go.dev/github.com/dhconnelly/rtreego). It is slower and stale.
- An 80×24 frame is 1,920 cells, and labels are horizontal runs on a single row. A per-frame bitset of 30 `uint64` values answers "is this span plus margin free?" with a mask test. It needs no allocation and is about 60 lines.

**Recommendation:** Use an occupancy bitmap. Revisit `tidwall/rtree` only if cursor hit-testing over thousands of features is added later.

## 4. Mapbox GL style JSON
No viable library exists.
- `github.com/go-spatial/tegola/mapbox/style` contains struct types only. It is v0.21.2 (2025-01-07), MIT (https://pkg.go.dev/github.com/go-spatial/tegola/mapbox/style).
- `github.com/flywave/go-mapbox/style` has no licence (https://pkg.go.dev/github.com/flywave/go-mapbox/style).
- `github.com/akhenakh/maprender` is MIT, published 2026-08-24, with 1 star, and it does evaluate filters and expressions.
  - The evaluator lives inside the root package rather than a separate importable one.
  - That package is tied to `tdewolff/canvas` and `peterstace/simplefeatures`, and has 27 imports (https://pkg.go.dev/github.com/akhenakh/maprender).
  - It is useful as reference reading only.
- Searches of pkg.go.dev and GitHub for MapLibre or Mapbox expression evaluators returned nothing else.

Minimum subset for a hand-written evaluator:
- **Legacy filters:** `==`, `!=`, `<`, `<=`, `>`, `>=`, `in`, `!in`, `has`, `!has`, `all`, `any`, `none`, plus the special keys `$type` and `$id`.
- **Expressions:** `get`, `has`, `!`, `all`, `any`, the comparison operators, `in`, `match`, `case`, `coalesce`, `literal`, `zoom`, `geometry-type`, `id`, `step`, `interpolate` (`linear` and `exponential`), `to-string`, `to-number`, `concat`.
- **Zoom functions:** the legacy `{base, stops}` form, exponential for numbers and stepped for colours and enums. Property functions are not needed.
- **Values:** number, string, bool, and colour. Colour means hex, `rgb()`, `rgba()`, `hsl()` and a small table of named colours.
- Compile each layer once into closures that read a property-accessor interface, with no maps. Tell legacy filters from expressions using MapLibre's `isExpressionFilter` rule.

**Recommendation:** Hand-write a compiled evaluator for a documented subset. An unknown operator should produce a warning at load time, never a panic.

## 5. HTTP + caching
- Embedded KV options:
  - `go.etcd.io/bbolt` is v1.5.0 (2026-06-03), MIT, pure Go, memory-mapped, with 19 imports (https://pkg.go.dev/go.etcd.io/bbolt). It is single-writer, and my understanding (UNVERIFIED) is that it holds an exclusive file lock, so a second Watchpost process would block.
  - `github.com/cockroachdb/pebble/v2` is v2.1.7, BSD-3-Clause, with 75 imports (https://pkg.go.dev/github.com/cockroachdb/pebble/v2).
  - `github.com/dgraph-io/badger/v4` is v4.9.6, Apache-2.0, with 40 imports and an LSM plus value-log design (https://pkg.go.dev/github.com/dgraph-io/badger/v4).
  - All three are over-engineered for immutable blobs keyed by z/x/y, and they work against a 100 MB RSS budget.
- Plain files would work as follows:
  - path `UserCacheDir/go-tuimaps/<source-hash>/z/x/y.pbf.gz`;
  - bytes stored as received;
  - written with temp file plus rename;
  - refreshed by mtime or ETag;
  - kept under an occasional size-capped prune.
  - This is safe across processes and can be inspected by hand.
- `github.com/hashicorp/golang-lru/v2` is v2.0.7 (2023-09-21), MPL-2.0, with 5.1k stars and 34 issues (https://pkg.go.dev/github.com/hashicorp/golang-lru/v2).
  - It bounds the number of entries, not bytes.
  - MPL-2.0 adds a licence notice to an MIT project.

**Recommendation:** Use plain files. For memory, hand-write a byte-budgeted LRU of *decoded* tiles using `container/list` plus a map, about 60 lines, with a 16–32 MB default. Add a hand-written in-flight request de-duplicator of about 20 lines.

## 6. Output contract
- `charm.land/lipgloss/v2` v2.0.6 (2026-08-11, MIT) measures width through x/ansi, and `lipgloss.Color` returns `image/color.Color`. Its `Canvas.SetCell` takes a `*uv.Cell` (https://pkg.go.dev/charm.land/lipgloss/v2).
- `github.com/charmbracelet/x/ansi` v0.11.8 (2026-08-11) has ANSI-aware `StringWidth`, `Truncate` and `Cut`, each with a `Wc` variant.
  - It depends on `clipperhouse/displaywidth`, `uax29` and `go-runewidth` (https://raw.githubusercontent.com/charmbracelet/x/main/ansi/go.mod).
  - displaywidth defaults to `EastAsianWidth: false`, so ambiguous-width characters count as 1 (https://pkg.go.dev/github.com/clipperhouse/displaywidth).
- Braille U+2800..28FF is East Asian Width **N** (neutral) in Unicode (https://www.unicode.org/Public/UCD/latest/ucd/EastAsianWidth.txt).
  - go-runewidth has no table entry for it (https://raw.githubusercontent.com/mattn/go-runewidth/master/runewidth_table.go).
  - Braille is therefore width 1 regardless of ambiguous-width settings.
  - Box-drawing 2500..254B and block elements 2580..258F are **A** (ambiguous), so keep them out of map frames.
- `charm.land/bubbletea/v2` v2.0.9 (2026-08-19) returns a `tea.View` set with `SetContent(string)`. It has a cell-based renderer and built-in colour downsampling (https://pkg.go.dev/charm.land/bubbletea/v2).
- `github.com/charmbracelet/colorprofile` v0.4.3 (2026-03-09) provides a `Writer` that downgrades SGR sequences in the byte stream, plus `Profile.Convert` (https://pkg.go.dev/github.com/charmbracelet/colorprofile).
- `github.com/charmbracelet/ultraviolet` has only pseudo-versions and states "The API may change" (https://pkg.go.dev/github.com/charmbracelet/ultraviolet).
- The tiletea README reports that Bubble Tea v2.0.8 and later drop Kitty graphics sequences (https://github.com/akhenakh/tiletea). I read that as evidence the renderer re-parses content into cells, so only plain SGR should be emitted.

**Recommendation:** Use a two-level contract.
- The core returns a `Frame` of W×H cells, each `Cell{Rune, FG, BG}`, using only standard-library colour types.
- `Frame.String()` emits:
  - exactly H lines of exactly W width-1 cells;
  - run-length truecolor SGR sequences;
  - a reset at the end of every line;
  - no cursor movement and no trailing newline.
- The host performs colour downgrade. The standalone app wraps stdout in `colorprofile.Writer`.
- Add an optional palette or profile hint, so the library can quantise colours when the style is compiled.
- The strongest counter-argument is that the host re-parses the string on every frame, and that downsampling truecolor after the fact can muddy a 16-colour map. A direct `uv.Drawable` would avoid both, but it would couple the library to an unstable API.

## 7. Geodesy
- Hand-written Web Mercator is about 40 lines, including the ±85.0511° clamp, and nothing more is needed.
- orb is pre-1.0 (v0.13). Putting `orb.Geometry` in public signatures would tie our semver to theirs.

**Recommendation:** Define our own `LngLat` type and overlay structs, plus a parser over `encoding/json` of about 150 lines covering Point, LineString, Polygon, the Multi* types, Feature and FeatureCollection. Since `orb.Point` is `[2]float64`, show a conversion example in the docs and do not depend on orb.

## 8. Embedding assets
- `embed` does not compress, so keep tiles as `.pbf.gz`.
- Embed patterns skip files starting with `.` or `_` unless the `all:` prefix is used. Symlinks and files outside the module are not allowed.
- Embedded files are part of the module zip, so every consumer downloads them even when build tags exclude them.
- Build tags apply to the whole build, so they put the burden on whoever builds the final binary.
- Whether the linker drops an unreferenced `embed.FS` is UNVERIFIED.
- For sizing, z0–z3 is 85 tiles and z0–z4 is 341.
- OSM and OpenMapTiles attribution obligations travel with the embedded data.

**Recommendation:** Put the assets in a separate opt-in package, for example `tuimaps/assets`, exposing an `fs.FS`, and have the core accept any `fs.FS`. Consumers exclude the assets by not importing the package. Cap the embedded data at about 2–5 MB (z0–z3).

## 9. Testing
- `github.com/charmbracelet/x/exp/golden` is MIT, has only pseudo-versions, and has 8 imports. It provides `RequireEqual` and an `-update` flag (https://pkg.go.dev/github.com/charmbracelet/x/exp/golden).
- `github.com/charmbracelet/x/exp/teatest/v2` has a pseudo-version from 2026-09-13 and imports `charm.land/bubbletea/v2` (https://pkg.go.dev/github.com/charmbracelet/x/exp/teatest/v2).
- Native `go test -fuzz` with real tiles as the seed corpus should check three invariants:
  - the decoder never panics;
  - declared lengths never exceed the bytes that remain;
  - allocation stays bounded.

**Recommendation:**
- For the core:
  - a hand-written golden helper of about 30 lines, comparing plain-text frames and SGR frames separately, with `-update`;
  - native fuzzing;
  - `testing.AllocsPerRun` and `-benchmem` as budget gates;
  - a differential test against orb in a nested test-only module.
- Use teatest/v2 only in the adapter and app module.

## 10. Dependency posture
- **Core module:** zero direct third-party dependencies. It uses only the standard library: `net/http`, `compress/gzip`, `encoding/json`, `embed`, `container/list`, `image/color`.
- **Adapter and app module:** `charm.land/bubbletea/v2`, `charm.land/lipgloss/v2` and `github.com/charmbracelet/colorprofile`.
- **Test-only, in nested modules:** orb and teatest.
- **Hand-written**, about 1,500 lines in total:

| Component | Approx. lines |
|---|---|
| MVT decoder | 300 |
| Style evaluator | 450 |
| Scanline fill and line drawing | 150 |
| Braille canvas | 100 |
| Label bitmap | 60 |
| Mercator | 40 |
| Disk cache | 150 |
| LRU and in-flight de-duplicator | 80 |
| GeoJSON reader | 150 |

- The strongest counter-argument is that each hand-written parser consuming untrusted network bytes is security-relevant code that no one else has audited and that one person maintains.
  - orb has years of production use.
  - The style spec is large, and users will bring arbitrary styles.
  - The mitigations are fuzzing, differential tests, and a published supported subset with strict warnings.

**Recommendation:** Zero third-party dependencies in the core, three charm modules in the adapter, and orb for tests only.
