---
title: "go-tuiMaps v0.2.0 — Radar loops — IMPLEMENTATION PLAN"
date: 2026-09-23
phase: PLAN
sev: SEV-0
authority: HUM LEAD
status: "DRAFT, revised after the internal plan check — for the PLAN red team, then HUM LEAD approval. Two items wait on rulings, marked PENDING. No code: signatures, shapes, test descriptions, file paths and order only (watchpost D-13, v0.2.0 D-52)."
---

# Implementation plan

**Goal.** Build v0.2.0 as `requirements.md` states it, in the order `integration-map.md` fixes, each
task test-first (FULL TDD).

**Tech stack.** Go at the module's floor toolchain; **no new module**. The library's third-party
graph stays go-runewidth and uax29 (v0.1.0 D-81). Hashing, HTTP and PNG handling are standard
library. `govulncheck` (L10.4) runs pinned through `go run …@vX` in the gate and never enters
`go.mod`.

**Architecture.** The four approaches as ruled: the declarative loop and one playback API (D-54), the
host transport (D-55), fetch-time age and a cache generation (D-56), and the structured `Report`
(D-57). Within v0.1.0's shape: the library starts no goroutine, the host drives `Work`, and `Set`
replaces (v0.1.0 D-74, D-86).

**The contract.** v0.2.0 edits the library's one contract in place:
`06_docs/02_features/go-tuimaps/03-architecture-design/contract.md` (called `contract.md` below).

**Branch.** `feature/radar-loops`, squash-merged into `release/v0.2.0` at SHIP (D-1).

**Every task follows the same order.**
1. Write the named test and watch it fail for the named reason.
2. Make the change.
3. Verify: the package's tests, then the full gate (`scripts/gate`) for any code; `scripts/gate
   --docs` for a Markdown-only change.
4. Mutation check: delete or disable what the test protects, watch it fail, and restore.

**The trace.** The table at the end maps every requirement, metric and owed item to its task.
`requirements.md`'s Instrument column is brought up to date once, at BUILD exit, under one ruling row.

## Numbers this plan proposes

Set here because D-49 item 3 gives them to PLAN; each goes to the HUM LEAD in the PLAN report.

| Number | Proposed | Why |
|---|---|---|
| `MaxFrames` (L-1.15), gaps included | **36** | Three hours at a five-minute cadence; 36 dot-resolution frames fit the 3 MiB budget (0.59 MiB for 12) |
| A frame's file-size cap (L-12.4) | **4 bytes a pixel plus 64 KiB** | An honest uncompressed RGBA frame fits; a padded 250,000-pixel PNG does not |
| M1's non-visual arm (D-24) | The HUM LEAD reads only the `Report`'s motion over five recorded loops and states each cell's direction; **pass: four of five within one compass point** | Scored as M1's visual arm is |
| M6's run count (D-6) | **Five consecutive full gate runs** before SHIP with no unattributable failure, counted from `06_docs/gate-runs.md` | Enough to see a flake that fires one run in three |
| The reference machine (L-12.5) | The development Mac named in wave 2's Limits | The one wave 2 measured on |

## Where the code lives

| File | New or changed | Tasks |
|---|---|---|
| `overlays.go` | changed | L2.1, L5.1, L6.1 |
| `map.go` | changed | L2.6, L3.2, L3.11, L7.1, L7.2, L9.3 |
| `playback.go` | **new** | L4.1, L4.4, L4.5, L4.7 |
| `clock.go` | changed | L4.2, L4.3, L4.9 |
| `look.go` | changed | L4.6, L10.2 |
| `report.go` | **new** | L5.2, L5.3, L5.4, L5.5 |
| `facts.go` | changed | L5.6 |
| `tables.go` | **new** | L6.1, L6.2, L6.3, L6.5, L6.7, L6.9 |
| `view.go` | changed | L7.1, L7.2 |
| `tiles.go` | changed | L8.1, L8.2, L8.4, L9.2, L9.4 |
| `describe.go` | changed | L4.7, L4.9 |
| `internal/overlay/image.go`, `store.go`, `cache.go`, `classified.go`, `fresh.go` | changed | L2.1, L2.2, L2.3, L2.4, L2.5, L2.6, L2.7, L2.8, L2.9, L4.9, L5.1, L6.4, L6.6 |
| `internal/render/frame.go`, `raster.go`, `field.go`, `hatch.go`, `marker.go`, `motion.go` | changed | L3.1, L3.3, L3.6, L3.9, L3.10, L3.11, L3.13, L4.1, L4.2, L4.8 |
| `internal/render/profile_test.go` | changed | L3.14 |
| `internal/colour/check.go`, `preset.go` | changed | L3.7, L3.8, L3.12 |
| `internal/describe/area.go`, `near.go` | changed | L5.4 |
| `internal/fetch/fetch.go`, `get.go` | changed | L8.1, L8.2, L8.3, L8.4, L8.5, L8.6, L8.7 |
| `internal/tiles/disk.go`, `pipeline.go`, `remote.go` | changed | L8.5, L9.1, L9.2, L9.3, L9.5, L9.6, L10.2 |
| `surface_test.go`, `readme_test.go`, `rules_test.go`, `memory_test.go`, `view_test.go`, `gate_test.go` | changed | L1.1, L1.4, L2.10, L7.3, L10.1, L10.3, L10.8 |
| `contract_test.go` | **new** | L1.2, L1.6, L10.6 |
| `playback_test.go` | **new** | L4.10 |
| `examples/example_playback_test.go` | **new** | L4.11 |
| `examples/example_map_test.go` | changed | L5.7 |
| `contract.md`, `README.md` | changed | L1.2, L1.3, L1.5, L1.6, L5.7, L10.6 |
| `06_docs/02_features/go-tuimaps/07-readiness/public-surface.txt`, `release-checklist.md` | changed | L1.1, L6.7, L10.4 |
| `scripts/gate` | changed | L10.3, L10.4, L10.5, L10.8 |
| `.github/workflows/gate.yml` | **new** | L10.5 |
| `testdata/mrms/`, `testdata/loops/` | **new** | L6.3, L6.9, L10.9 |
| `06_docs/02_features/radar-loops/02-analysis/capture-procedure.md` | **new** | L6.8 |

## Work packages and order

| WP | What | Requirements | Needs |
|---|---|---|---|
| L1 | Contract skeleton, a surface test that records signatures, the contract's new text | L-4, contract text for L-1.10d, L-7.3, L-8.3, L-9.3, L-9.6, L-12.3 | — |
| L2 | Loops: frames, copy, validation, keys, budget | L-1.1–L-1.4, L-1.7, L-1.9, L-1.14, L-1.15, L-12 | L1 |
| L3 | Renderer: frame identity, blend, furniture, place labels, severity word and dash | L-1.6, L-1.16, L-8, L-11, L-12.5 | L2, L4.1–L4.4 for L3.1 and L3.2 |
| L4 | Playback API and clock | L-1.5, L-1.8, L-1.10a–g, L-1.11, L-1.13 | L2 |
| L5 | `Report` | L-1.12, L-13.5–L-13.7, L-13.9, L-13.10 | L3, L4, L6 for L5.6 |
| L6 | MRMS table, the provider-table seam, the heavy end | L-2 | L1 |
| L7 | View bound | L-3 | L1 |
| L8 | Fetch options and confinement | L-7, L-10 | L1 |
| L9 | Cache age and purge | L-9 | L8 |
| L10 | Defects, the first release's close-out, the gate's own evidence | L-5, L-6, L-13.1–L-13.4, L-13.8, M4, M6, OW-4 | L5–L9 |

L2 → L4 → L3 → L5 is the critical path (L3's frame-identity tasks need a way to advance the shown
frame, which L4 gives). L6, L7 and L8 → L9 run beside it once L1 lands.

---

## WP-L1 — Contract skeleton and signature-level surface test (L-4)

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L1.1 | The surface snapshot records signatures and struct fields, not names only (L-4.3) | `surface_test.go`, `public-surface.txt` | The snapshot line becomes `func (*Map) Render(Size, time.Time) (Frame, error)`, and `type Frame struct{ Lines []string; Status Status; … }` | Change one exported signature in a scratch copy: today's test passes; the new one must fail |
| L1.2 | A test holds `contract.md` to the surface both ways (L-4.1) | `contract_test.go`, `contract.md` | Every exported name in the snapshot appears in `contract.md`, and every code span in `contract.md` naming a Go identifier exists | Today it fails, listing `SetSize`, `Pan`, `ZoomAround`, `Focused`, `BorrowCheck`, `Frame.Line`, `WriteTo` (W1-B) |
| L1.3 | The five false behavioural claims are corrected (L-4.2) | `contract.md` | Text only | L1.2 green; a checklist row for each claim in the PLAN report |
| L1.4 | The README's promises join the check (L-4.4, D-34) | `readme_test.go` | The README's user-agent, `Purge` and "nothing until you name a source" sentences are asserted against behaviour | Change one sentence: the test fails |
| L1.5 | A changelog section in `contract.md` for v0.2.0's breaks (D-58), including `CacheRoot` gaining options (L9.2) | `contract.md` | One row per break: what changed, why, and what a host does instead | L1.2 requires the section to exist |
| L1.6 | The contract states what a host needs to know: the loop's name is the host's (L-1.10d); the time bound holds only if the host's transport honours its context, and what `CheckedDialer` does and does not cover (L-7.3, L-10.3); `Purge` does not reach a released root (L-9.3); cache-root guidance (L-9.6); dot-grid guidance and "a host that raises the budget owns the memory" (L-12.3); what carries severity inside rain cells (L-8.3) | `contract.md`, `contract_test.go` | Text | A test holds one sentence for each named row, as L1.4 does for the README |

## WP-L2 — Loops (D-54)

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L2.1 | `LoopFrame`; `Image.Frames` | `overlays.go`, `internal/overlay/image.go` | `type LoopFrame struct{ Valid time.Time; PNG []byte; Gap bool }`; `Image.Frames []LoopFrame`; a single picture is `PNG` with no `Frames` | `Set` of an image with three frames is accepted; with both `PNG` and `Frames` set it is refused with a clear error |
| L2.2 | Validation at hand-in (L-1.9, L-1.15) | `internal/overlay/image.go` | Every frame's header and size checked; times strictly increasing; a gap has no bytes; `const MaxFrames = 36`, exported, gaps included | A table test: out-of-order times, a duplicate time, a gap with bytes, frame 37, a bad PNG in frame 7: each refused, naming the frame |
| L2.3 | Copy at hand-in, for the single image and every frame (L-1.14) | `internal/overlay/image.go`, `internal/overlay/store.go` | The store keeps its own copies of `PNG` and `Table`; the host's slices are never read after `Set` returns | Mutate the host's PNG and table after `Set`: the drawn picture and classes are unchanged |
| L2.4 | Decode reads the size from the copy before it allocates, and re-checks the cap and the budget (L-1.14) | `internal/overlay/image.go` (`rasterise`) | The header is checked before `png.Decode` | A header claiming 4× the cap is refused without the allocation (the allocation counter stays under a bound) |
| L2.5 | Keys: valid time plus a hash of the bytes; reuse on a re-`Set` (L-1.7) | `internal/overlay/cache.go`, `internal/overlay/classified.go` | A frame's key is its valid time and a content hash; a re-`Set` keeps decoded frames whose key it already holds | Re-`Set` with 11 frames the same and 1 new: exactly one decode (a decode counter) |
| L2.6 | The budget's surface (L-12.1): an option and a setter; a hand-in over it is refused, saying by how much | `map.go`, `internal/overlay/store.go` | `WithImageBudget(bytes int64) Option`; `(*Map).SetImageBudget(int64) error`; default 3 MiB (D-61) | Over the budget: refused, and the error says by how much |
| L2.7 | What the budget counts (L-12.4, L-12.6): retained PNG bytes, classified pixels, fields and the shared classified set | `internal/overlay/store.go`, `internal/overlay/classified.go` | — | A loop's counted bytes equal the sum of its parts, checked against a hand count |
| L2.8 | A frame's file size is capped relative to its pixels (L-12.4) | `internal/overlay/image.go` | 4 bytes a pixel plus 64 KiB | A padded 250,000-pixel PNG is refused; an honest one passes |
| L2.9 | `ImageKey` hashes a value's exact bits (L-12.6) | `internal/overlay/classified.go` | The key covers each value's bit pattern | Two tables that differ only at 1e18 get different keys |
| L2.10 | The library never fetches radar (L-1.4) | `rules_test.go` | — | A rules test: no package on the image path imports `internal/fetch` or `net/http` |

## WP-L4 — Playback (D-25, D-26, D-54)

L4 comes before L3 in the build order, because L3.1 and L3.2 need to advance the shown frame.

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L4.1 | Playback state per looped overlay; off by default (L-1.5, L-1.11) | `playback.go`, `internal/render/motion.go` | `type Playback uint8` (`PlaybackOff`, `PlaybackSlow`, `PlaybackNormal`); `(*Map).SetPlayback(id string, p Playback) error` | A new loop reads `PlaybackOff` |
| L4.2 | The loop advances on the animation clock at 1 and 2 frames a second and holds the newest 2 s; **one change ceiling per map** (L-1.10g, watchpost D-43). Every loop's advances and the blink share one tick grid per map, so advances that fall together count as one change | `clock.go`, `internal/render/motion.go` | The frame shown is a function of the animation time, as the blink phase is | A table over animation times: the index shown at each; two loops at normal plus blink never exceed 2.5 changes a second |
| L4.3 | `NextCall` includes the next frame change (L-1.8) | `clock.go` | A fourth source | With a loop playing, `NextCall` returns the next advance time |
| L4.4 | Stepping and seeking (L-1.10b) | `playback.go` | `Step(id string, by int) error`, `Seek(id string, index int) error`, `ShowNewest(id string) error` | Step at each end, seek out of range, step with playback on |
| L4.5 | The state read (L-1.10d) | `playback.go` | `type LoopState struct{…}` as approach 1; `Loop(id string) (LoopState, bool)`; `Advancing` false when the clock is frozen; the precedence of off reasons | Freeze with `Animate`: `Advancing` false; reduce motion over a host "normal": off because of reduce motion |
| L4.6 | `ReduceMotion` forces off and restores (L-1.10c) | `look.go` | Existing call, new effect | Normal → reduce on → off → reduce off → normal |
| L4.7 | `FrameTicks`, apart from `Changed` (L-1.10e) | `playback.go`, `describe.go` | `FrameTicks() uint64`; a tick changes neither `Changed()` nor the description's key | Advance ten frames: `Changed()` unchanged, `FrameTicks()` +10, the description cached |
| L4.8 | Gaps: hold the last real frame, the time reads "gap"; the shown frame's time on the map; with two loops, the newest loop's time (L-1.2, L-1.10a, L-1.10f, L-1.10g) | `internal/render/frame.go` (furniture) | Frame-time text beside `stale` | Golden: a gap shows the previous frame's picture with "gap 14:10"; two loops show the newer one's time |
| L4.9 | The newest non-gap frame drives `stale` and the description (L-1.3, D-39) | `internal/overlay/fresh.go`, `clock.go` (`stale`), `describe.go` | — | Stepping to an old frame does not raise `stale`; an old newest frame does |
| L4.10 | **Correct parts, correctly connected** (RK-1): the whole path through the public `Map` | `playback_test.go` | — | `Set` a loop, `SetPlayback`, `Animate` forward, `Render`: the lines differ from the first render and `FrameTicks` moved |
| L4.11 | One standard playback API a host wires with no state of its own (L-1.13) | `examples/example_playback_test.go` | An example host: controls and a Settings row driven only by `SetPlayback`, `Step`, `Seek`, `ShowNewest` and `Loop`, woken by `FrameTicks` and `NextCall` | The example's output; a rules test that the example declares no playback state |

## WP-L3 — Renderer (D-14, D-17, D-28, D-45, D-59, D-60)

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L3.1 | The frame-reuse test sees the shown frame (L-1.6) | `internal/render/frame.go` (`sameOverlays`), `internal/render/raster.go` | `scene.Raster` gains the shown frame's key; reuse compares it | `Step` the shown frame with nothing else changed: the frame is redrawn, not reused |
| L3.2 | `Frame` carries `Changed` and `Ticks` (L-1.16) | `map.go` | `type Frame struct{ Lines []string; Status Status; Changed, Ticks uint64; Dropped []Drop }` | A render after a `Set` carries the new `Changed`; after a frame advance, the new `Ticks` |
| L3.3 | Furniture is never erased by an image or field at NoColour: outline, hatch, label, marker (L-8.3) | `internal/render/field.go`, `internal/render/frame.go` | Shade cells yield to them | Specimen 29's scene at NoColour: each present (today they vanish) |
| L3.4 | The same at NoColour for the `stale` word, the notice, the frame time, the scale and the credit (L-8.3) | same | — | Specimen 29's scene: each present |
| L3.5 | L3.3 and L3.4 at Colours16 (L-8.3) | same | — | The same scene at Colours16 |
| L3.6 | The tint blends over the image (L-11.1) | `internal/render/frame.go` (`colours`) | The bare tint only where there is no image; otherwise the image's class colour shifted toward the tint in linear light | Pixel test over a blended cell; mutant: the tint wins |
| L3.7 | The checker holds the blend: class separation ≥ 10 and visibility ≥ 5, on each ground where the blend is used (L-11.2, L-11.5) | `internal/colour/check.go` | `CheckBlend(ramp, tints, ground, strength) []Finding` | Today's 35 % fails; 20 % passes on dark; the light ground selects L-11.4's fallback |
| L3.8 | The light-ground search, or its named fallback (L-11.3, L-11.4, D-27) | `internal/colour/preset.go` | Search the light ramp, the alert colours and per-class tints; if nothing passes, radar over tint on light | The checker picks the fallback when no candidate passes, and says so |
| L3.9 | **PENDING the D-17 specimen (OW-2).** The severity word in the label (L-8.1) | `internal/render/frame.go` | The word derived from the feature's role until L5.1 makes severity data | Golden frames of the D-17 specimen at both sizes and depths, drawn over radar |
| L3.10 | **PENDING OW-2.** Five distinct outline dashes, one a severity, distinct from the line-overlay dash (L-8.6) | `internal/render/hatch.go` | A dash table of five patterns | The same goldens; a test that the five and the line dash differ |
| L3.11 | **PENDING OW-2.** The hatch at Colours16 (L-8.7); a label that doesn't fit falls back to the word, and a dropped label is reported (L-8.5) | `internal/render/hatch.go`, `map.go` | `type Drop struct{ Kind DropKind; Overlay, Label, Shown string }`; kinds `DropAlertLabel`, `DropPlaceName` | At 69×12 the fallback is drawn and `Frame.Dropped` names it |
| L3.12 | Contrast of outline and label over the blend (L-8.8) | `internal/colour/check.go` | 3:1 for outlines and 4.5:1 for labels, truecolor and 256 | Checker test at each severity on each ground |
| L3.13 | **The named place's marker and name are preserved** (L-8.9, D-60) | `internal/render/marker.go`, `internal/render/frame.go` | Places outrank alert labels, basemap labels and furniture where they collide; a shorter form, or the marker alone, and a `DropPlaceName` in `Frame.Dropped` | Specimens 29, 30 and 31 as tests: the place's name is drawn at 149×38 in each, and at 69×12 either it or its reported short form |
| L3.14 | A frame advance costs ≤ 15 ms at 149×38 (L-12.5, D-61) | `internal/render/profile_test.go` | `BenchmarkFrameAdvance` over a 12-frame loop. **Not a gated timing** (timing in the gate is flaky): the gate holds an allocation count per advance, and the time is measured on the reference machine and recorded in the SHIP report | The allocation pin; the recorded time ≤ 15 ms |

## WP-L5 — `Report` (D-42, D-43, D-57)

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L5.1 | Severity and times as data on a feature (L-13.9) | `overlays.go`, `internal/overlay/store.go` | `Feature.Severity Severity; Feature.Valid, Expires time.Time` (zero: derived as today); constants `SeverityUnknown`, `SeverityMinor`, `SeverityModerate`, `SeveritySevere`, `SeverityExtreme` | A feature with no severity reports the one its role implies |
| L5.2 | `Report`'s shape | `report.go` | `type Report struct{ Alerts []AlertShown; Places []PlaceReport; Motion []MotionReport }`; `AlertShown{Overlay, Feature, Label string; Severity Severity; Valid, Expires time.Time; Stale bool}`; `PlaceReport{Place string; Alerts []PlaceAlert}`; `PlaceAlert{Overlay, Feature string; Where Where; DistanceKm float64; Bearing string}`; `(*Map).Report(places []Place) (Report, error)` | An empty map returns an empty `Report`, never nil sections |
| L5.3 | The alerts shown, no place needed (L-13.5) | `report.go` | — | Three alerts in view → three entries in draw order |
| L5.4 | Per place, each alert on its own; "nearby" (L-13.6, D-43) | `report.go`, `internal/describe/area.go`, `internal/describe/near.go` | `Where` gains `Nearby`; `SetNearby(km float64) error`, default 10 km | Specimen 31: Fort Davis is "outside" the partial watch, and with its zone in, "inside"; a place 6 km from an edge is "nearby", 14 km is "outside"; `SetNearby(20)` moves it |
| L5.5 | Motion, relative to a place or to the view (L-1.12, D-42) | `report.go`, `internal/describe/` | `MotionReport{Overlay, Place string; Threshold int; From, To Sighting; Relation string; Span time.Duration}`; an empty `Place` means relative to the view's centre | A synthetic two-frame loop with a cell moving 12 km east: the relation and positions; with no place, relative to the view |
| L5.6 | The legend says when a table is approximate (L-13.10) | `facts.go` | `LegendEntry.Approximate bool`, read from the image's provider table (L6.1) | MRMS's legend entry is approximate; IEM's is not |
| L5.7 | The contract and the example show a host wording intensity from `Legend()` (L-13.10) | `contract.md`, `examples/example_map_test.go` | Example only | The example's output names a class range |

## WP-L6 — MRMS, the table seam, the heavy end (D-19, D-35, D-44)

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L6.1 | The provider-table seam (L-2.5), and how an image uses it | `tables.go`, `overlays.go`, `internal/colour/` | `type ProviderTable struct{ Name string; Entries []TableEntry; Approximate bool }`; `Tables() []ProviderTable`; one registration per provider; `Image.Provider string` names one, used when `Image.Table` is empty, and carries its `Approximate` | A third table is added without editing the other two (an architectural test, as watchpost FR-9.3); an image naming a provider draws with its table |
| L6.2 | IEM's published table as the first entry | `tables.go` | From `02-analysis/programs/inputs/spec29/n0q-table-raw.json` | Round trip against those values |
| L6.3 | MRMS's observed palette, approximate (L-2.1, L-2.4) | `tables.go`, `testdata/mrms/` | The 111 colours from `02-analysis/programs/output/mrms-palette.txt`, valued from the legend | Every observed colour matches exactly; the entry is approximate |
| L6.4 | Fallback matches counted apart from unmatched ones (L-2.3, OW-10) | `internal/overlay/image.go` | — | A frame with known fallback pixels reports their count |
| L6.5 | **The heavy end valued better than the nearest legend colour** (L-2.3): a colour off the table is valued by where it projects onto the legend's heavy-end gradient, not by the nearest single colour | `internal/colour/`, `tables.go` | — | Colours 3–30 from any legend colour (wave 2 M-A) are valued in order along the gradient |
| L6.6 | A warning whenever a colour takes the fallback (L-2.3) | `internal/overlay/image.go` | Warning kind `table-fallback`, with the count | A frame with one fallback pixel raises it |
| L6.7 | The shipped state if no severe day comes before SHIP (L-2.3, RK-2, RK-11): MRMS's heavy end marked unverified in the legend, and a release-note line | `tables.go`, `release-checklist.md` | `ProviderTable` gains `Unverified string`, the range not yet seen | The legend says so; the checklist refuses a tag without the note while it is set |
| L6.8 | The triggered capture procedure (OW-12) | `02-analysis/capture-procedure.md` | When a Moderate or High risk or a tornado watch is issued, capture the two-hour window and extend the palette | A reviewer can follow it cold; a checklist row points to it |
| L6.9 | The fallback-share test (L-2.3) | `tables.go` tests, `testdata/mrms/` | — | A frame whose fallback share exceeds the threshold fails; the OW-11 archive frames pass |

## WP-L7 — View bound (L-3)

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L7.1 | A bound set once | `view.go`, `map.go` | `type Bound struct{ MinZoom float64; W, S, E, N float64 }`; `WithBound(Bound) Option`; `(*Map).SetBound(Bound) error`. **A box that crosses the antimeridian (W > E) is accepted**: Alaska's and the Pacific territories' regions need it | A zero-area box is refused; an Aleutians box is held |
| L7.2 | Held on every path | `view.go` (`moveLocked`), `map.go` (`viewAt`) | Both enforce it (W1-C) | M3: resize, fit, pan, zoom and fall-back each stay inside, including across the antimeridian |
| L7.3 | A host that sets none gets v0.1.0's behaviour (L-3.2) | `view_test.go` | — | The existing view tests pass unchanged, with a named test that sets no bound |

## WP-L8 — Fetch options and confinement (D-55)

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L8.1 | `FetchOptions`, taking effect at once (L-7.1, L-7.2, L-7.4) | `tiles.go`, `internal/fetch/fetch.go` | `type FetchOptions struct{ Transport http.RoundTripper; UserAgent string; Timeout time.Duration; AllowHTTP []string }`; `(*Map).SetFetchOptions(FetchOptions) error` | A transport set after `Source` is the one used for the next fetch |
| L8.2 | **PENDING a ruling.** `Fetcher` (`tiles.go:13`, `:70`) removed, and `fetch.Checked` with it | `tiles.go`, `internal/fetch/get.go` | — | The surface test records the removal; the changelog row exists |
| L8.3 | Bytes bounded; late answers dropped; time bounded while the context is honoured (L-7.3) | `internal/fetch/get.go` | The library reads the body through its limit; checks the deadline after the transport returns | An endless-body transport is cut at the limit; a transport that ignores its context blocks only its own `Work` call, and its late answer is discarded |
| L8.4 | `CheckedDialer`, exported at the root (L-10.3) | `tiles.go` (wrapper), `internal/fetch/fetch.go` | `func CheckedDialer() func(context.Context, string, string) (net.Conn, error)` | A host transport using it refuses a private address |
| L8.5 | One allow-list; scheme, host, port; plain http refused unless allowed (L-10.2) | `internal/fetch/fetch.go`, `internal/tiles/remote.go` | One list feeds both | Through `Map.Source`, with the library's and a host's transport: another host, another port and http each refused |
| L8.6 | The proxy decision per connection; the reserved ranges completed (L-10.3) | `internal/fetch/fetch.go` | — | A redirect to a host the proxy does not carry still meets the private-address check; each new range refused |
| L8.7 | The real version in the user-agent (L-13.1) | `internal/fetch/fetch.go` | `Version` stays a constant; the release checklist bumps it, and L10.3's `--release` check refuses a tag it does not match | At the tag, the user-agent names the tag; `-dev` at a tag is refused |

## WP-L9 — Cache age and purge (D-56)

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L9.1 | The file time is the fetch time; reads never touch it (L-9.4) | `internal/tiles/disk.go` (`Load`, `Flush`) | The file time is set only at store | Load a tile hourly for a day: its file time never changes |
| L9.2 | `MaxAge`, enforced at `CacheRoot` and in every job; oldest-fetched evicted first, tiles in view protected; a future time counts as expired (L-9.4) | `tiles.go`, `internal/tiles/disk.go`, `internal/tiles/pipeline.go` | `func MaxAge(time.Duration) CacheOption`; `CacheRoot(dir string, capBytes int64, opts ...CacheOption) error` | A tile past its age is fetched again; a future-dated one too; `CacheRoot` over an aged root drops them; a tile in view survives eviction over the cap |
| L9.3 | The cache generation (L-9.3) | `internal/tiles/pipeline.go`, `map.go` | A counter raised by `Purge`, a `CacheRoot` change or off, and `Close`; a store lands only if unchanged | A fetch in flight across each of the three writes nothing |
| L9.4 | `Purge` of everything; replaced roots released (L-9.2, L-9.3); counts reported (L-9.5) | `tiles.go` | `Purge() error`; `type PurgeReport struct{ Removed, Failed int }`; `PurgeWithReport() (PurgeReport, error)` (approach 3) | After `Purge`, nothing under the root, memory empty; the old root's descriptor is closed; a failed remove shows in `Failed` |
| L9.5 | Failures raised and counted honestly (L-9.5) | `internal/tiles/pipeline.go`, `internal/tiles/disk.go` | `cache-write-failed` raised; only successful removals counted | A read-only root raises the warning; a failed remove is not counted |
| L9.6 | A root others can read is warned of (L-9.6); `ReadBack` reads through its limit (L-12.6) | `internal/tiles/disk.go` | — | A 0755 root warns; an oversized file is not read whole |

## WP-L10 — Defects, close-out, and the gate's own evidence

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L10.1 | A test lists the library's environment reads (L-13.8) | `rules_test.go` | Allowed: `look.go`'s colour-depth read | A new `os.Getenv` in a library file fails |
| L10.2 | Doc comments corrected (L-13.4) | `internal/tiles/disk.go`, `look.go` | Text | The two comments' claims each asserted by a test of the behaviour they describe |
| L10.3 | A gate test refuses a tag while its checklist is unfinished (L-5.3) | `gate_test.go`, `scripts/gate` | A `--release` check | An unticked row refuses; all ticked passes |
| L10.4 | A pinned `govulncheck` at the tag, requiring no reachable finding (L-5.5) | `scripts/gate`, `release-checklist.md` | The version pinned | An injected vulnerable module fails the release check |
| L10.5 | Hosted CI (L-6.1) | `.github/workflows/gate.yml` | The gate on Linux; the second architecture on a runner that has it | A run on the branch, green |
| L10.6 | The default tile memory cache documented against a large view (L-6.3) | `contract.md`, `contract_test.go` | Text | L1.6's sentence test covers it |
| L10.7 | **OW-4:** the 40-second `FuzzAgree` freeze examined | `06_docs/gate-runs.md` | `FuzzAgree` under the limiter, ten runs, each recorded | Ten runs recorded; a freeze, if seen, gets a cause or an open row with its evidence |
| L10.8 | **M4:** a 12-frame loop at dot resolution adds ≤ 1 MiB of heap over an empty map, within ±5 % over an hour (D-61) | `memory_test.go`, `scripts/gate` | A five-minute form in the full gate; the hour form in `scripts/gate --soak`, run once before SHIP and recorded | The five-minute form's bound; the hour run in the SHIP report |
| L10.9 | **M1:** the visual arm's recorded loops, and the non-visual arm's grader (D-24) | `testdata/loops/`, the PLAN report | Five recorded loops from real radar, the OW-11 outbreak among them | **Human-graded:** done when the HUM LEAD's scores for both arms are recorded |
| L10.10 | **M6:** five consecutive full runs with no unattributable failure before SHIP | `06_docs/gate-runs.md` | Counted from the log | A test reads the log and reports the current run of green |

## Dependencies within packages

L2.1 → L2.2 → L2.3 → L2.4 → L2.5 → L2.6 → L2.7 → L2.8 · L4.1 → L4.2 → L4.3; L4.4 and L4.5 need L4.1;
L4.10 needs L4.4 · L3.1 and L3.2 need L4.4 and L4.7 · L3.3 → L3.4 → L3.5; L3.6 → L3.7 → L3.8 ·
L3.9–L3.11 need OW-2 ruled and L3.5 · L5.1 → L5.2 → L5.3–L5.5; L5.6 needs L6.1 · L6.1 → L6.2, L6.3 →
L6.4 → L6.5, L6.6 → L6.9 · L8.1 → L8.2, L8.3; L8.5 → L8.6 · L9.3 needs L8.1; L9.1 → L9.2.

## The trace

| Requirement | Task | | Requirement | Task |
|---|---|---|---|---|
| L-1.1 | L2.1 | | L-7.1, L-7.2, L-7.4, L-13.2 | L8.1, L8.2 (pending) |
| L-1.2, L-1.10a, L-1.10f | L4.8 | | L-7.3 | L8.3, L1.6 |
| L-1.3 | L4.9 | | L-8.1, L-8.5–L-8.7 | L3.9–L3.11 (pending OW-2) |
| L-1.4 | L2.10 | | L-8.3 | L3.3–L3.5, L1.6 |
| L-1.5, L-1.11 | L4.1 | | L-8.8 | L3.12 |
| L-1.6 | L3.1 | | L-8.9 | L3.13 |
| L-1.7 | L2.5 | | L-9.2, L-9.3 | L9.3, L9.4, L1.6 |
| L-1.8 | L4.3 | | L-9.1, L-9.4 | L9.1, L9.2 |
| L-1.9, L-1.15 | L2.2 | | L-9.5 | L9.4, L9.5 |
| L-1.10b | L4.4 | | L-9.6 | L9.6, L1.6 |
| L-1.10c | L4.6 | | L-10.2 | L8.5 |
| L-1.10d | L4.5, L1.6 | | L-10.3 | L8.4, L8.6 |
| L-1.10e | L4.7 | | L-11.1 | L3.6 |
| L-1.10g | L4.2, L4.8 | | L-11.2, L-11.5 | L3.7 |
| L-1.12 | L5.5 | | L-11.3, L-11.4 | L3.8 |
| L-1.13 | L4.11 | | L-12.1 | L2.6 |
| L-1.14 | L2.3, L2.4 | | L-12.2, M4 | L2.6, L10.8 |
| L-1.16 | L3.2 | | L-12.3 | L1.6 |
| L-2.1, L-2.4 | L6.3 | | L-12.4 | L2.7, L2.8 |
| L-2.3, RK-2, RK-11, OW-12 | L6.4–L6.9 | | L-12.5 | L3.14 |
| L-2.5 | L6.1, L6.2 | | L-12.6 | L2.7, L2.9, L9.6 |
| L-3.1, M3 | L7.1, L7.2 | | L-13.1 | L8.7 |
| L-3.2 | L7.3 | | L-13.4 | L10.2 |
| L-4.1, L-4.2, M5 | L1.2, L1.3 | | L-13.5 | L5.3 |
| L-4.3 | L1.1 | | L-13.6 | L5.4 |
| L-4.4 | L1.4 | | L-13.9 | L5.1, L5.2 |
| L-5.3 | L10.3 | | L-13.8 | L10.1 |
| L-5.5 | L10.4 | | L-13.10 | L5.6, L5.7 |
| L-6.1 | L10.5 | | M1 | L10.9 |
| L-6.3 | L10.6 | | M2 | L4.8, L4.9 |
| L-6.4, OW-4 | L10.7 | | M6 | L10.10 |
| RK-1 | L4.10 | | NFR-1 (breaks listed) | L1.5 |
| L-10.1 | existing: `TestTileJSONAddressesObeyFetchRules`, `FuzzTileJSON` | | L-13.3 | done (D-38) |
| L-13.7 | deferred (D-29), no task | | L-8.2 | context for L-8.1, no task |

## Deviations from DISCOVER, recorded

- **L-7.3's time promise** is what a library that starts no goroutine can keep (D-55).
- **L-8.4 withdrawn** (D-50): the 16-colour shade defect was a miscount.
- **L-7.1's wording** ("supply a fetcher … request type exported") is met by `SetFetchOptions` if the
  `Fetcher` removal is ruled (L8.2); the row is reworded by that ruling.
