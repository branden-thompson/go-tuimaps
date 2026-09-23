---
title: "go-tuiMaps v0.2.0 — Radar loops — IMPLEMENTATION PLAN"
date: 2026-09-23
phase: PLAN
sev: SEV-0
authority: HUM LEAD
status: "DRAFT — for internal review, then the PLAN red team, then HUM LEAD approval. No code: signatures, shapes, test descriptions, file paths and order only (watchpost D-13, v0.2.0 D-52)."
---

# Implementation plan

**Goal.** Build v0.2.0 as `requirements.md` states it, in the order `integration-map.md` fixes, each
task test-first (FULL TDD).

**Architecture.** The four approaches as ruled: the declarative loop and one playback API (D-54), the
host transport (D-55), fetch-time age and a cache generation (D-56), and the structured `Report`
(D-57). Within v0.1.0's shape: the library starts no goroutine, the host drives `Work`, and `Set`
replaces (D-74, D-86).

**Branch.** `feature/radar-loops`, squash-merged into `release/v0.2.0` at SHIP (D-1).

**Every task follows the same order.** (1) Write the named test and watch it fail for the named reason.
(2) Make the change. (3) Run the package's tests, then the lane the commit needs: the full gate for any
code. (4) Mutation check: revert the change, watch the test fail, and restore. A task is not done until
its row in `requirements.md` names the test as its instrument.

## Work packages and order

| WP | What | Requirements | Needs |
|---|---|---|---|
| L1 | Contract skeleton and a surface test that records signatures | L-4 | — |
| L2 | Loops: frames, copy, validation, keys, budget | L-1.1–L-1.4, L-1.7, L-1.9, L-1.14, L-1.15, L-12 | L1 |
| L3 | Renderer: frame identity, blend, furniture, place labels, severity word and dash | L-1.6, L-1.16, L-8, L-11, L-12.5 | L2 |
| L4 | Playback API and clock | L-1.5, L-1.8, L-1.10a–g, L-1.11, L-1.13 | L3 |
| L5 | `Report` | L-1.12, L-13.5–L-13.10 | L4 |
| L6 | MRMS table and the provider-table seam | L-2 | L1 |
| L7 | View bound | L-3 | L1 |
| L8 | Fetch options and confinement | L-7, L-10 | L1 |
| L9 | Cache age and purge | L-9 | L8 |
| L10 | Defects, and the first release's close-out | L-5, L-6, L-13.1–L-13.4, L-13.8 | L5–L9 |

L2→L3→L4→L5 is the critical path. L6, L7 and L8→L9 run beside it once L1 lands.

---

## WP-L1 — Contract skeleton and signature-level surface test (L-4)

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L1.1 | The surface snapshot records signatures and struct fields, not names only (L-4.3) | `surface_test.go`, `06_docs/02_features/go-tuimaps/07-readiness/public-surface.txt` | The snapshot line becomes `func (*Map) Render(Size, time.Time) (Frame, error)`, and `type Frame struct{ Lines []string; Status Status; … }` | Change one exported signature in a scratch copy: today's test passes; the new one must fail |
| L1.2 | A test holds `contract.md` to the surface both ways (L-4.1) | `contract_test.go` (new), `contract.md` | Every exported name in the snapshot appears in `contract.md`, and every code span in `contract.md` naming a Go identifier exists | Today it fails, listing `SetSize`, `Pan`, `ZoomAround`, `Focused`, `BorrowCheck`, `Frame.Line`, `WriteTo` (W1-B) |
| L1.3 | The five false behavioural claims are corrected (L-4.2) | `contract.md` | Text only | L1.2 green; a checklist row for each claim in the PLAN report |
| L1.4 | The README's promises join the check (L-4.4, D-34) | `readme_test.go` | The README's user-agent, `Purge` and "nothing until you name a source" sentences are asserted against behaviour | Change one sentence: the test fails |
| L1.5 | A changelog section in `contract.md` for v0.2.0's breaks (D-58) | `contract.md` | One row per break: what changed, why, and what a host does instead | L1.2 requires the section to exist |

## WP-L2 — Loops (D-54)

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L2.1 | `LoopFrame`; `Image.Frames` | `overlays.go`, `internal/overlay/image.go` | `type LoopFrame struct{ Valid time.Time; PNG []byte; Gap bool }`; `Image.Frames []LoopFrame`; a single picture is `PNG` with no `Frames` | `Set` of an image with three frames is accepted; with both `PNG` and `Frames` set it is refused with a clear error |
| L2.2 | Validation at hand-in (L-1.9) | `internal/overlay/image.go` | Every frame's header and size checked; times strictly increasing; a gap has no bytes; at most `MaxFrames` frames, gaps included (L-1.15) | A table test: out-of-order times, a duplicate time, a gap with bytes, one frame too many, a bad PNG in frame 7: each refused, naming the frame |
| L2.3 | Copy at hand-in, for the single image and every frame (L-1.14) | `internal/overlay/image.go`, `store.go` | The store keeps its own copies of `PNG` and `Table`; the host's slices are never read after `Set` returns | Mutate the host's PNG and table after `Set`: the drawn picture and classes are unchanged |
| L2.4 | Decode reads the size from the copy before it allocates, and re-checks the cap and the budget (L-1.14) | `internal/overlay/image.go` (`rasterise`) | The header is checked before `png.Decode` | A header claiming 4× the cap is refused without the allocation (the allocation counter stays under a bound) |
| L2.5 | Keys: valid time plus a hash of the bytes; reuse on a re-`Set` (L-1.7) | `internal/overlay/cache.go`, `classified.go` | `frameKey{valid, sha256[:16]}`; a re-`Set` keeps decoded frames whose key it already holds | Re-`Set` with 11 frames the same and 1 new: exactly one decode (a decode counter) |
| L2.6 | The total image budget per map (L-12.1–L-12.4) | `internal/overlay/store.go`, `map.go` | `WithImageBudget(bytes int64) Option` and `(*Map).SetImageBudget(int64) error`; default 3 MiB (D-61); counts retained PNG bytes and classified pixels, fields and the shared set (L-12.6); a frame's file size is capped relative to its pixel count | Over the budget: refused, and the error says by how much; a padded 250k-pixel PNG is refused |
| L2.7 | `ImageKey` hashes a value's exact bits (L-12.6) | `internal/overlay/classified.go` | `math.Float64bits` | Two tables that differ only at 1e18 get different keys |

## WP-L3 — Renderer (D-14, D-17, D-28, D-45, D-59, D-60)

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L3.1 | The frame-reuse test sees the shown frame (L-1.6) | `internal/render/frame.go` (`sameOverlays`), `raster.go` | `scene.Raster` gains the shown frame's key; reuse compares it | Advance the shown frame with nothing else changed: the frame is redrawn, not reused |
| L3.2 | `Frame` carries `Changed` and `FrameTicks` (L-1.16) | `map.go` | `type Frame struct{ Lines []string; Status Status; Changed, Ticks uint64 }` | A render after a `Set` carries the new `Changed`; after a tick, the new `Ticks` |
| L3.3 | Furniture is never erased by an image or field, at NoColour and Colours16 (L-8.3) | `internal/render/field.go`, `frame.go` | Shade cells yield to outline, hatch, label, marker, `stale` word, notice, frame time, scale and credit | Re-render specimen 29's scenario in a test at both depths: every furniture item present (today they vanish) |
| L3.4 | The tint blends over the image (L-11.1) | `internal/render/frame.go` (`colours`) | The bare tint only where there is no image; otherwise the image's class colour shifted toward the tint in linear light | Pixel test over a blended cell; mutant: the tint wins |
| L3.5 | The checker holds the blend: class separation ≥ 10 and visibility ≥ 5, on each ground where the blend is used (L-11.2, L-11.5) | `internal/colour/check.go` | `CheckBlend(ramp, tints, ground, strength) []Finding` | Today's 35 % fails; 20 % passes on dark; the light ground selects L-11.4's fallback |
| L3.6 | The light-ground search, or its named fallback (L-11.3, L-11.4, D-27) | `internal/colour/preset.go` | Search the light ramp, the alert colours and per-class tints; if nothing passes, radar over tint on light | The checker picks the fallback when no candidate passes, and says so |
| L3.7 | The severity word in the label; five distinct dashes; the hatch at Colours16 (L-8.1, L-8.5–L-8.7) | `internal/render/hatch.go`, `frame.go` | A dash table of five patterns distinct from the line-overlay dash; a label that doesn't fit falls back to the word alone, then reports it | Golden frames of the D-17 specimen (OW-2) at both sizes and depths, drawn over radar |
| L3.8 | Contrast of outline and label over the blend (L-8.8) | `internal/colour/check.go` | 3:1 for outlines and 4.5:1 for labels, truecolor and 256 | Checker test at each severity on each ground |
| L3.9 | **The named place's marker and name are preserved** (L-8.9, D-60; OW-3) | `internal/render/marker.go`, `frame.go` | Places outrank alert labels, basemap labels and furniture where they collide; a shorter form, or the marker alone, when the name cannot fit, and the frame reports it | Specimens 29, 30 and 31 as tests: the place's name is drawn at 149×38 in each, and at 69×12 either it or its reported short form |
| L3.10 | A frame advance costs ≤ 15 ms at 149×38 (L-12.5, D-61) | `internal/render/profile_test.go` | Benchmark | `BenchmarkFrameAdvance` over a 12-frame loop, gated at 15 ms on the reference machine |

## WP-L4 — Playback (D-25, D-26, D-54)

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L4.1 | Playback state per looped overlay; off by default (L-1.5, L-1.11) | `playback.go` (new), `internal/render/motion.go` | `type Playback uint8` (Off, Slow, Normal); `(*Map).SetPlayback(id string, p Playback) error` | A new loop reads `Off` with `OffByDefault` |
| L4.2 | The loop advances on the animation clock at 1 and 2 frames a second, holds the newest 2 s, one ceiling per map (L-1.10g, watchpost D-43) | `clock.go`, `internal/render/motion.go` | The frame shown is a function of the animation time, as the blink phase is | Table over animation times: the index shown at each; two loops together never exceed 2.5 changes a second, blink included |
| L4.3 | `NextCall` includes the next frame change (L-1.8) | `clock.go` | A fourth source | With a loop playing, `NextCall` returns the next advance time |
| L4.4 | Stepping and seeking (L-1.10b) | `playback.go` | `Step(id string, by int) error`, `Seek(id string, index int) error`, `ShowNewest(id string) error` | Step at each end, seek out of range, step with playback on |
| L4.5 | The state read (L-1.10d) | `playback.go` | `type LoopState struct{…}` as approach 1; `Loop(id string) (LoopState, bool)`; `Advancing` false when the clock is frozen; the precedence of off reasons | Freeze with `Animate`: `Advancing` false; reduce motion over a host "normal": off because of reduce motion |
| L4.6 | `ReduceMotion` forces off and restores (L-1.10c) | `look.go` | Existing call, new effect | Normal → reduce on → off → reduce off → normal |
| L4.7 | `FrameTicks`, apart from `Changed` (L-1.10e) | `playback.go`, `describe.go` | `FrameTicks() uint64`; a tick changes neither `Changed()` nor the description's key | Advance ten frames: `Changed()` unchanged, `FrameTicks()` +10, the description cached |
| L4.8 | Gaps: hold the last real frame, the time reads "gap"; the shown frame's time on the map (L-1.2, L-1.10a, L-1.10f) | `internal/render/frame.go` (furniture) | Frame-time text beside `stale` | Golden: a gap shows the previous frame's picture with "gap 14:10" |
| L4.9 | The newest non-gap frame drives `stale` and the description (L-1.3, D-39) | `fresh.go`, `describe.go` | — | Stepping to an old frame does not raise `stale`; an old newest frame does |

## WP-L5 — `Report` (D-42, D-43, D-57)

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L5.1 | Severity and times as data on a feature (L-13.9) | `overlays.go`, `internal/overlay/store.go` | `Feature.Severity Severity; Feature.Valid, Expires time.Time` (zero: derived as today) | A feature with no severity reports the one its role implies |
| L5.2 | `Report`'s shape | `report.go` (new) | `type Report struct{ Alerts []AlertShown; Places []PlaceReport; Motion []MotionReport }`; `(*Map).Report(places []Place) (Report, error)` | An empty map returns an empty `Report`, never nil sections |
| L5.3 | The alerts shown, no place needed (L-13.5) | `report.go` | Each alert's label, severity, valid and expiry times, and stale mark | Three alerts in view → three entries in draw order |
| L5.4 | Per place, each alert on its own; "nearby" (L-13.6, D-43) | `report.go`, `internal/describe/area.go` | `Where` gains `Nearby`; `SetNearby(km float64) error`, default 10 km | Specimen 31's scenario: Fort Davis is "outside" the partial watch; with its zone in, "inside" |
| L5.5 | Motion, relative to a place or to the view (L-1.12, D-42) | `report.go`, `internal/describe/` | `type Motion struct{ Threshold int; From, To Sighting; Relation string; Span time.Duration }` | A synthetic two-frame loop with a cell moving 12 km east: the relation and positions; with no place, relative to the view's centre |
| L5.6 | The legend says when a table is approximate (L-13.10) | `facts.go` | `LegendEntry.Approximate bool` | MRMS's legend entry is approximate; IEM's is not |
| L5.7 | The contract and the example show a host wording intensity from `Legend()` (L-13.10) | `contract.md`, `examples/example_map_test.go` | Example only | The example's output names a class range |

## WP-L6 — MRMS and the table seam (D-19, D-35, D-44)

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L6.1 | The provider-table seam (L-2.5) | `tables.go` (new), `internal/colour/` | `type ProviderTable struct{ Name string; Entries []TableEntry; Approximate bool }`; `Tables() []ProviderTable`; one registration per provider | A third table is added without editing the other two (an architectural test, as watchpost FR-9.3) |
| L6.2 | IEM's published table as the first entry | `tables.go` | — | Round trip against IEM's published values |
| L6.3 | MRMS's observed palette, approximate (L-2.1, L-2.4) | `tables.go`, `testdata/mrms/` | The 111 colours from `02-analysis/programs/output/mrms-palette.txt`, valued from the legend | Every observed colour matches exactly; the entry is approximate |
| L6.4 | The heavy end and the fallback-share test (L-2.3) | `tables.go`, `internal/overlay/image.go` | Fallback matches are counted apart from unmatched ones (OW-10) | A frame whose fallback share exceeds a threshold fails the test |

## WP-L7 — View bound (L-3)

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L7.1 | A bound set once | `view.go`, `map.go` | `type Bound struct{ MinZoom float64; W, S, E, N float64 }`; `WithBound(Bound) Option`; `(*Map).SetBound(Bound) error` | A zero-area or antimeridian box is refused |
| L7.2 | Held on every path | `view.go` (`moveLocked`), `map.go` (`viewAt`) | Both enforce it (W1-C) | M3: resize, fit, pan, zoom and fall-back each stay inside |
| L7.3 | A host that sets none gets v0.1.0's behaviour | — | — | The existing view tests pass unchanged |

## WP-L8 — Fetch options and confinement (D-55)

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L8.1 | `FetchOptions`, taking effect at once (L-7.1, L-7.2, L-7.4) | `tiles.go`, `internal/fetch/` | `type FetchOptions struct{ Transport http.RoundTripper; UserAgent string; Timeout time.Duration; AllowHTTP []string }`; `(*Map).SetFetchOptions(FetchOptions) error`; `Fetcher` and `fetch.Checked` removed (D-58) | A transport set after `Source` is the one used for the next fetch |
| L8.2 | Bytes bounded; late answers dropped; time bounded while the context is honoured (L-7.3) | `internal/fetch/get.go` | The library reads the body through its limit; checks the deadline after the transport returns | An endless-body transport is cut at the limit; a transport that ignores its context blocks only its own `Work` call, and its late answer is discarded |
| L8.3 | `CheckedDialer` (L-10.3) | `internal/fetch/fetch.go` | `func CheckedDialer() func(context.Context, string, string) (net.Conn, error)` | A host transport using it refuses a private address |
| L8.4 | One allow-list; scheme, host, port; plain http refused unless allowed (L-10.2) | `internal/fetch/fetch.go`, `internal/tiles/remote.go` | One list feeds both | Through `Map.Source`, with the library's and a host's transport: another host, another port and http each refused |
| L8.5 | The proxy decision per connection; the reserved ranges completed (L-10.3) | `internal/fetch/fetch.go` | — | A redirect to a host the proxy does not carry still meets the private-address check; each new range refused |
| L8.6 | The real version in the user-agent (L-13.1) | `internal/fetch/fetch.go` | `Version` from the release, not `-dev` | The user-agent at a tagged build names the tag |

## WP-L9 — Cache age and purge (D-56)

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L9.1 | The file time is the fetch time; reads never touch it (L-9.4) | `internal/tiles/disk.go` (`Load`, `Flush`) | `Chtimes` only at store | Load a tile hourly for a day: its file time never changes |
| L9.2 | `MaxAge`; oldest-fetched evicted first; a future time counts as expired | `tiles.go`, `internal/tiles/disk.go` | `func MaxAge(time.Duration) CacheOption`; `CacheRoot(dir string, capBytes int64, opts ...CacheOption) error` | A tile past its age is fetched again; a future-dated one too |
| L9.3 | The cache generation (L-9.3) | `internal/tiles/pipeline.go`, `map.go` | A counter raised by `Purge`, a `CacheRoot` change or off, and `Close`; a store lands only if unchanged | A fetch in flight across each of the three writes nothing |
| L9.4 | `Purge` of everything; replaced roots released (L-9.2, L-9.3) | `tiles.go` | Every source, the memory caches, decoded pictures | After `Purge`, nothing under the root, memory empty; the old root's descriptor is closed |
| L9.5 | Failures raised and counted honestly (L-9.5) | `internal/tiles/pipeline.go`, `disk.go` | `cache-write-failed` raised; only successful removals counted | A read-only root raises the warning; a failed remove is not counted |
| L9.6 | A root others can read is warned of (L-9.6); `ReadBack` reads through its limit (L-12.6) | `internal/tiles/disk.go` | — | A 0755 root warns; an oversized file is not read whole |

## WP-L10 — Defects and the first release's close-out

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L10.1 | A test lists the library's environment reads (L-13.8) | `rules_test.go` | Allowed: `look.go`'s colour-depth read | A new `os.Getenv` in a library file fails |
| L10.2 | Doc comments corrected (L-13.4) | `internal/tiles/disk.go`, `look.go` | Text | Review |
| L10.3 | A gate test refuses a tag while its checklist is unfinished (L-5.3) | `gate_test.go`, `scripts/gate` | A `--release` check | An unticked row refuses; all ticked passes |
| L10.4 | A pinned `govulncheck` at the tag, requiring no reachable finding (L-5.5) | `scripts/gate`, `release-checklist.md` | The version pinned | An injected vulnerable module fails the release check |
| L10.5 | Hosted CI (L-6.1) | `.github/workflows/gate.yml` | The gate on Linux; the second architecture on a runner that has it | A run on the branch, green |
| L10.6 | The default tile memory cache documented against a large view (L-6.3) | `contract.md` | Text | Review |

## Dependencies within packages

L2.1 → L2.2 → L2.3 → L2.4 → L2.5 → L2.6 · L3.1 → L3.2; L3.3 and L3.4 → L3.5 → L3.6; L3.7 needs L3.3 ·
L4.1 → L4.2 → L4.3; L4.4 and L4.5 need L4.1 · L5.1 → L5.2 → L5.3–L5.5 · L8.1 → L8.2; L8.4 → L8.5 ·
L9.3 needs L8.1; L9.1 → L9.2.

## Deviations from DISCOVER, recorded

- **L-7.3's time promise** is what a library that starts no goroutine can keep (D-55).
- **`Fetcher` and `fetch.Checked` are removed**, not kept beside `SetFetchOptions`: a break under
  D-58, listed in the contract's changelog.
- **L-8.4 withdrawn** (D-50): the 16-colour shade defect was a miscount.
