---
title: "go-tuiMaps v0.2.0 — Radar loops — IMPLEMENTATION PLAN"
date: 2026-09-23
phase: PLAN
sev: SEV-0
authority: HUM LEAD
status: "DRAFT, revised after the internal plan check — for the PLAN red team, then HUM LEAD approval. No item waits on a ruling. No code: signatures, shapes, test descriptions, file paths and order only (watchpost D-13, v0.2.0 D-52)."
---

# Implementation plan

**Goal.** Build v0.2.0 as `requirements.md` states it, in the order `integration-map.md` fixes, each
task test-first (FULL TDD).

**External services** BUILD relies on: `govulncheck`'s database at tag time (L10.4), a hosted runner with the second architecture (L10.5), and the Storm Prediction Center's outlooks for OW-12's trigger (L6.8).

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
| `MaxFrames` (L-1.15), gaps included | **72** (D-68) | Two hours at MRMS's ~2-minute cadence is about 60 frames, plus forecast; a host may let its listener choose it |
| A frame's file-size cap (L-12.4) | **4 bytes a pixel plus 64 KiB** | An honest uncompressed RGBA frame fits; a padded 250,000-pixel PNG does not |
| M1's non-visual arm (D-24, D-71) | **Scored first**, on the words a listener hears (watchpost W9.7), over five recorded loops, one with several cells; the ground truth is the HUM LEAD's own reading of the loops, set before either arm is scored; **pass: four of five within one point of an 8-point compass** | Scored as M1's visual arm is |
| M6's run count (D-6) | **Five consecutive full gate runs** before SHIP with no unattributable failure, counted from `06_docs/gate-runs.md` | Enough to see a flake that fires one run in three |
| The reference machine (L-12.5) | The development Mac named in wave 2's Limits | The one wave 2 measured on |

## Where the code lives

| File | New or changed | Tasks |
|---|---|---|
| `overlays.go` | changed | L2.1, L4.9a, L5.1, L6.1 |
| `map.go` | changed | L2.6, L3.2, L3.11, L4.7, L7.1, L7.2, L9.3 |
| `playback.go` | **new** | L4.1, L4.4, L4.5, L4.7 |
| `clock.go` | changed | L4.2, L4.3, L4.9 |
| `look.go` | changed | L4.6, L10.2 |
| `report.go` | **new** | L5.2, L5.3, L5.4, L5.5 |
| `facts.go` | changed | L3.11a, L5.6 |
| `tables.go` | **new** | L6.1, L6.2, L6.3, L6.5, L6.7, L6.9 |
| `view.go` | changed | L7.1, L7.2 |
| `tiles.go` | changed | L8.1, L8.2, L8.4, L9.2, L9.4 |
| `work.go` | changed | L4.7 |
| `describe.go` | changed | L4.7, L4.9 |
| `internal/overlay/image.go`, `store.go`, `cache.go`, `classified.go`, `fresh.go` | changed | L2.1, L2.2, L2.3, L2.4, L2.5, L2.6, L2.7, L2.8, L2.9, L4.9, L4.9a, L5.1, L6.4, L6.6 |
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
| L1.3 | The five false behavioural claims are corrected (L-4.2), and a sixth: `Changed()` "moves whenever a redraw would differ" (D-66) | `contract.md`, `contract_test.go` | Text, and a behaviour test for each | **Each corrected claim is held by a test of the behaviour it describes** (D-72), not only by its name existing |
| L1.4 | The README's promises join the check (L-4.4, D-34) | `readme_test.go` | The README's user-agent, `Purge` and "nothing until you name a source" sentences are asserted against behaviour | Each promise's behaviour is exercised: break the behaviour and the test fails |
| L1.5 | A changelog section in `contract.md` for v0.2.0's breaks (D-58), including `Changed()` redefined (D-66), `Describe` removed and the other simplifications (D-70), and the borrow check with its warning kind removed (D-74) and playback made map-wide (D-67) | `contract.md` | One row per break: what changed, why, and what a host does instead | L1.2 requires the section to exist |
| L1.6 | The contract states what a host needs to know: the loop's name is the host's (L-1.10d); the time bound holds only if the host's transport honours its context, and what `CheckedDialer` does and does not cover (L-7.3, L-10.3); `Purge` does not reach a released root (L-9.3); cache-root guidance (L-9.6); dot-grid guidance and "a host that raises the budget owns the memory" (L-12.3); what carries severity inside rain cells (L-8.3) | `contract.md`, `contract_test.go` | Text | A test holds one sentence for each named row; where a sentence states behaviour, a behaviour test holds it too (D-72) |
| L1.7 | **The borrow check removed** (D-74, OW-13): the fingerprints, the re-check, `Caps.BorrowCheck`, their test and benchmark leg, and the warning kind `BorrowChanged` | `internal/overlay/cache.go`, `internal/overlay/store.go`, `internal/fault/fault.go`, `kinds.go`, `contract.md` | — | The surface snapshot records the removal; the contract test passes without the names; `go vet` finds nothing left behind |

## WP-L2 — Loops (D-54)

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L2.1 | `LoopFrame`; `Image.Frames` | `overlays.go`, `internal/overlay/image.go` | `type LoopFrame struct{ Valid time.Time; PNG []byte; Gap, Forecast bool }` (forecast from D-67); `Image.Frames []LoopFrame`; a single picture is `PNG` with no `Frames` | `Set` of an image with three frames is accepted; with both `PNG` and `Frames` set it is refused with a clear error |
| L2.2 | Validation at hand-in (L-1.9, L-1.15) | `internal/overlay/image.go` | Every frame's header and size checked; times strictly increasing; a gap has no bytes; `const MaxFrames = 72`, exported, gaps included (D-68) | A table test: out-of-order times, a duplicate time, a gap with bytes, frame 73, a bad PNG in frame 7: each refused, naming the frame |
| L2.3 | Copy at hand-in, for the single image and every frame (L-1.14) | `internal/overlay/image.go`, `internal/overlay/store.go` | The store keeps its own copies of `PNG` and `Table`; the host's slices are never read after `Set` returns | Mutate the host's PNG and table after `Set`: the drawn picture and classes are unchanged |
| L2.4 | Decode reads the size from the copy before it allocates, and re-checks the cap and the budget (L-1.14) | `internal/overlay/image.go` (`rasterise`) | The header is checked before `png.Decode` | A header claiming 4× the cap is refused without the allocation (the allocation counter stays under a bound) |
| L2.5 | Keys: valid time plus a hash of the bytes; reuse on a re-`Set` (L-1.7) | `internal/overlay/cache.go`, `internal/overlay/classified.go` | A frame's key is its valid time and a content hash; a re-`Set` keeps decoded frames whose key it already holds | Re-`Set` with 11 frames the same and 1 new: exactly one decode (a decode counter) |
| L2.6 | The budget's surface (L-12.1): a setter; a hand-in over it is refused, saying by how much | `map.go`, `internal/overlay/store.go` | `(*Map).SetImageBudget(bytes int64) error` (D-70: no paired option); default 6 MiB (D-68, raised from D-61's 3 MiB) Lowering it below what is held keeps what is held and refuses the next hand-in that does not fit, saying so (D-72) | Over the budget: refused, and the error says by how much; lowered below use: nothing dropped, the next `Set` over it refused |
| L2.7 | What the budget counts (L-12.4, L-12.6): retained PNG bytes, classified pixels, fields and the shared classified set | `internal/overlay/store.go`, `internal/overlay/classified.go` | — | A loop's counted bytes equal the sum of its parts, checked against a hand count |
| L2.8 | A frame's file size is capped relative to its pixels (L-12.4) | `internal/overlay/image.go` | 4 bytes a pixel plus 64 KiB | A padded 250,000-pixel PNG is refused; an honest one passes |
| L2.9 | `ImageKey` hashes a value's exact bits (L-12.6) | `internal/overlay/classified.go` | The key covers each value's bit pattern | Two tables that differ only at 1e18 get different keys |
| L2.10 | The library never fetches radar (L-1.4) | `rules_test.go` | — | A rules test: no package on the image path imports `internal/fetch` or `net/http` |

## WP-L4 — Playback (D-25, D-26, D-54)

L4 comes before L3 in the build order, because L3.1 and L3.2 need to advance the shown frame.

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L4.1 | Playback is one per map, and persists across `Set`; off by default (L-1.5, L-1.11, D-67) | `playback.go`, `internal/render/motion.go` | `type Playback uint8` (`PlaybackOff`, `PlaybackSlow`, `PlaybackNormal`); `(*Map).SetPlayback(p Playback) error` | A new map reads `PlaybackOff`; a refresh `Set` leaves the setting as it was |
| L4.2 | The loop advances on the animation clock at 1 and 2 frames a second, holds the last frame 2 s and repeats; **one change ceiling per map**: every loop's advances and the blink share one tick grid, so advances that fall together count as one change (L-1.10g, watchpost D-43) | `clock.go`, `internal/render/motion.go` | The frame shown is a function of the animation time since `Play`, as the blink phase is | A table over animation times: the frame shown at each; two loops at normal plus blink never exceed 2.5 changes a second |
| L4.3 | `NextCall` includes the next frame change (L-1.8, D-66) | `clock.go` | A fourth source; the signal that time has something due | With a loop playing, `NextCall` returns the next advance time |
| L4.4 | **The listener's controls** (L-1.10b, D-67): `Play()` from the oldest held frame through "right now" and on through forecast frames; `Stop()` holds; `Reset()` returns to "right now", the newest observed frame; `Step(by int)` moves by frames and stops playback; **position is a valid time** | `playback.go` | `(*Map).Play() error`, `Stop() error`, `Reset() error`, `Step(by int) error` | Open → "right now"; play → oldest first, then in time order, then forecast; stop → held; step at each end; reset; **a refresh `Set` that drops the oldest frame keeps the same moment on screen** |
| L4.5 | The state read (L-1.10d) | `playback.go` | `type LoopState struct{ At, Now time.Time; Playing, Advancing, Forecast bool; Playback Playback; Off OffReason }`, the zero `OffReason` meaning not off (D-70); `(*Map).Loop() LoopState`; `Advancing` false when the clock is frozen; the precedence of off reasons | Freeze with `Animate`: `Advancing` false; reduce motion over a host "normal": off because of reduce motion |
| L4.6 | `ReduceMotion` forces off and restores (L-1.10c) | `look.go` | Existing call, new effect | Normal → reduce on → off → reduce off → normal |
| L4.7 | **The counters, one job each** (L-1.10e, D-66): `Changed()` counts inputs, raised by `Work` when it lands something visible and never by `Render`; `FrameTicks()` counts frame advances as of the last time given | `map.go`, `work.go`, `playback.go`, `describe.go` | `FrameTicks() uint64`; a tick changes neither `Changed()` nor the description's key | A tile landed by `Work` raises `Changed()` with no `Render`; a `Render` of unchanged inputs leaves it alone; advance ten frames: `Changed()` unchanged, `FrameTicks()` +10, the description cached |
| L4.8 | Gaps hold the last real frame, the time reads "gap"; the shown frame's time is on the map, marked when it is a forecast; with two loops, the newest loop's time (L-1.2, L-1.10a, L-1.10f, L-1.10g, D-67) | `internal/render/frame.go` (furniture) | Frame-time text beside `stale` | Golden: a gap shows the previous frame's picture with "gap 14:10"; a forecast frame's time reads as a forecast; two loops show the newer one's time |
| L4.9 | The newest observed non-gap frame drives `stale` and the description (L-1.3, D-39, D-67) | `internal/overlay/fresh.go`, `clock.go` (`stale`), `describe.go` | — | Stepping to an old frame does not raise `stale`; an old newest frame does; a forecast frame never counts as newest |
| L4.9a | Forecast frames (D-67): `LoopFrame.Forecast`; only a forecast frame may carry a future valid time | `overlays.go`, `internal/overlay/image.go` | `LoopFrame{Valid time.Time; PNG []byte; Gap, Forecast bool}` | An observed frame dated past now plus a small skew is refused; a forecast frame so dated is accepted and drawn after "right now" |
| L4.10 | **Correct parts, correctly connected** (RK-1): the whole path through the public `Map` | `playback_test.go` | — | `Set` a loop, `SetPlayback`, `Play`, `Animate` forward, `Render`: the lines differ from the first render and `FrameTicks` moved |
| L4.11 | One standard playback API a host wires with no state of its own (L-1.13) | `examples/example_playback_test.go` | An example host: controls and a Settings row driven only by `SetPlayback`, `Play`, `Stop`, `Reset`, `Step` and `Loop`, woken by `Changed` and `NextCall` | The example's output; a rules test that the example declares no playback state |

## WP-L3 — Renderer (D-14, D-17, D-28, D-45, D-59, D-60)

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L3.1 | The frame-reuse test sees the shown frame (L-1.6) | `internal/render/frame.go` (`sameOverlays`), `internal/render/raster.go` | `scene.Raster` gains the shown frame's key; reuse compares it | `Step` the shown frame with nothing else changed: the frame is redrawn, not reused |
| L3.2 | `Frame` carries `Changed` and `FrameTicks` (L-1.16) | `map.go` | `type Frame struct{ Lines []string; Status Status; Changed, FrameTicks uint64; Dropped []Drop }` (named as the methods, D-72); `type Drop struct{ Kind DropKind; Overlay, Label, Shown string }`, kinds `DropAlertLabel`, `DropPlaceName` | A render after a `Set` carries the new `Changed`; after a frame advance, the new `FrameTicks` |
| L3.3 | Furniture is never erased by an image or field at NoColour: outline, hatch, label, marker (L-8.3) | `internal/render/field.go`, `internal/render/frame.go` | Shade cells yield to them | Specimen 29's scene at NoColour: each present (today they vanish) |
| L3.4 | The same at NoColour for the `stale` word, the notice, the frame time, the scale and the credit (L-8.3) | same | — | Specimen 29's scene: each present |
| L3.5 | L3.3 and L3.4 at Colours16 (L-8.3) | same | — | The same scene at Colours16 |
| L3.6 | The tint blends over the image (L-11.1) | `internal/render/frame.go` (`colours`) | The bare tint only where there is no image; otherwise the image's class colour shifted toward the tint in linear light | Pixel test over a blended cell; mutant: the tint wins |
| L3.7 | The checker holds the blend: class separation ≥ 10 and visibility ≥ 5, on each ground where the blend is used (L-11.2, L-11.5) | `internal/colour/check.go` | `CheckBlend(ramp, tints, ground, strength) []Finding` | Today's 35 % fails; 20 % passes on dark; the light ground selects L-11.4's fallback |
| L3.8 | The light-ground search, or its named fallback (L-11.3, L-11.4, D-27) | `internal/colour/preset.go` | Search the light ramp, the alert colours and per-class tints; if nothing passes, radar over tint on light. **Re-run whenever the palette or the ground changes** (a host palette can break the blend), and the result reported in `Legend()` and `Warnings()` (D-72) | The checker picks the fallback when no candidate passes, and says so; a host palette set after `New` re-runs it |
| L3.9 | The severity word in the label (L-8.1) | `internal/render/frame.go` | The word from the feature's severity (L5.1) | Specimen 33's scene as goldens at both sizes and depths: every label that fits carries its word |
| L3.10 | **A severity digit repeated along each alert outline** (L-8.1, D-65): Extreme 4, Severe 3, Moderate 2, Minor 1, Unknown `?`; the outline solid; never on the furniture rows; at least one digit on every area | `internal/render/frame.go`, `internal/render/basemap.go` | A digit every few outline cells, staggered by row, placed after overlay labels and before basemap names | Specimen 33's scenes: every area carries its digit at 69×12 and 149×38; no digit in a furniture row; a four-cell area still gets one |
| L3.11 | The hatch at Colours16 (L-8.7); a label that doesn't fit falls back to the word, and a dropped label is reported (L-8.5) | `internal/render/hatch.go`, `map.go` | Uses L3.2's `Frame.Dropped` | At 69×12 the fallback is drawn and `Frame.Dropped` names it |
| L3.11a | `Legend()` carries the severity digit key (D-65) | `facts.go` | A legend section for alert severity: digit and word | The legend of a frame with alerts lists the five |
| L3.12 | Contrast of outline and label over the blend (L-8.8) | `internal/colour/check.go` | 3:1 for outlines and 4.5:1 for labels, truecolor and 256 | Checker test at each severity on each ground |
| L3.13 | **The named place's marker and name are preserved** (L-8.9, D-60) | `internal/render/marker.go`, `internal/render/frame.go` | Places outrank alert labels, basemap labels and furniture where they collide; a shorter form, or the marker alone, and a `DropPlaceName` in `Frame.Dropped` | Specimens 29, 30 and 31 as tests: the place's name is drawn at 149×38 in each, and at 69×12 either it or its reported short form |
| L3.14 | A frame advance costs ≤ 15 ms at 149×38 (L-12.5, D-61) | `internal/render/profile_test.go` | `BenchmarkFrameAdvance` over a 12-frame loop. **Not a gated timing** (timing in the gate is flaky): the gate holds an allocation count per advance, and the time is measured on the reference machine and recorded in the SHIP report | The allocation pin; the recorded time ≤ 15 ms |

## WP-L5 — `Report` (D-42, D-43, D-57)

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L5.1 | Severity and times as data on a feature (L-13.9) | `overlays.go`, `internal/overlay/store.go` | `Feature.Severity Severity; Feature.Valid, Expires time.Time` (zero: derived as today); constants `SeverityUnknown`, `SeverityMinor`, `SeverityModerate`, `SeveritySevere`, `SeverityExtreme` | A feature with no severity reports the one its role implies |
| L5.2 | `Report`'s shape | `report.go` | `type Report struct{ Alerts []AlertShown; Places []PlaceReport; Motion []MotionReport }`; `AlertShown{Overlay, Feature, Label string; Severity Severity; Valid, Expires time.Time; Stale bool}`; `PlaceReport{Place string; Alerts []PlaceAlert; Images []ImageAnswer}`; `PlaceAlert{Overlay, Feature string; Where Where; Distance float64; Unit string; Bearing float64; Compass string}` — v0.1.0's names and units, honouring `Units()`; `Feature.ID string` so an entry joins back to the host's alert; **every string cleaned on the way out** (FR-34); `(*Map).Report(places []Place) (Report, error)` (D-72) | An empty map returns an empty `Report`, never nil sections; a hostile label comes back cleaned; distances follow `Units()` |
| L5.3 | The alerts shown, no place needed (L-13.5) | `report.go` | — | Three alerts in view → three entries in draw order |
| L5.4 | Per place, each alert on its own; "nearby" (L-13.6, D-43) | `report.go`, `internal/describe/area.go`, `internal/describe/near.go` | `Where` gains `Nearby`; `SetNearby(km float64) error`, default 10 km | Specimen 31: Fort Davis is "outside" the partial watch, and with its zone in, "inside"; a place 6 km from an edge is "nearby", 14 km is "outside"; `SetNearby(20)` moves it |
| L5.5 | Motion, relative to a place or to the view (L-1.12, D-42) | `report.go`, `internal/describe/` | `MotionReport{Overlay, Place string; Threshold int; From, To Sighting; Relation Where; Span time.Duration}`; an empty `Place` means relative to the view's centre. **The sighting rule** (D-72): the heavier rain is the largest connected area at or above the threshold class (taken from the provider table's legend), its position the area's centroid; a cell is "held" from one frame to the next when the nearest such area lies within a stated distance; with several cells, each place reports the nearest | A two-frame loop with a cell moving 12 km east; a two-cell scene where each place gets its nearest; a gap at the oldest frame; with no place, relative to the view |
| L5.6 | The legend says when a table is approximate (L-13.10) | `facts.go` | `LegendEntry.Approximate bool`, read from the image's provider table (L6.1) | MRMS's legend entry is approximate; IEM's is not |
| L5.7 | The contract and the example show a host wording intensity from `Legend()` (L-13.10) | `contract.md`, `examples/example_map_test.go` | Example only | The example's output names a class range |
| L5.8 | **`Describe` is removed** (D-70); every answer it gave is in `Report` (`PlaceReport` keeps the image answers, B-1) | `describe.go`, `report.go`, `examples/` | — | The surface test records the removal; the changelog row exists; each `Describe` test is ported to `Report` and passes |

## WP-L6 — MRMS, the table seam, the heavy end (D-19, D-35, D-44)

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L6.1 | The provider-table seam (L-2.5), and how an image uses it | `tables.go`, `overlays.go`, `internal/colour/` | An internal table registry, one registration per provider (D-70); `type Provider uint8` with `ProviderIEM`, `ProviderMRMS`; `Image.Provider Provider` names one, used when `Image.Table` is empty, and carries its approximate and unverified marks | A third provider is added inside the library without editing the other two (an architectural test); an image naming a provider draws with its table |
| L6.2 | IEM's published table as the first entry | `tables.go` | From `02-analysis/programs/inputs/spec29/n0q-table-raw.json` | Round trip against those values |
| L6.3 | MRMS's observed palette, approximate (L-2.1, L-2.4) | `tables.go`, `testdata/mrms/` | The 111 colours from `02-analysis/programs/output/mrms-palette.txt`, valued from the legend | Every observed colour matches exactly; the entry is approximate |
| L6.4 | Fallback matches counted apart from unmatched ones (L-2.3, OW-10) | `internal/overlay/image.go` | — | A frame with known fallback pixels reports their count |
| L6.5 | **The heavy end valued better than the nearest legend colour** (L-2.3): a colour off the table is valued by where it projects onto the legend's heavy-end gradient, not by the nearest single colour | `internal/colour/`, `tables.go` | — | Colours 3–30 from any legend colour (wave 2 M-A) are valued in order along the gradient |
| L6.6 | A warning whenever a colour takes the fallback (L-2.3) | `internal/overlay/image.go` | Warning kind `table-fallback`, with the count | A frame with one fallback pixel raises it |
| L6.7 | The shipped state if no severe day comes before SHIP (L-2.3, RK-2, RK-11): MRMS's heavy end marked unverified in the legend, and a release-note line | `tables.go`, `facts.go`, `release-checklist.md` | `LegendEntry.Unverified bool` (D-70); the unseen range named in the contract and the release note | The legend says so; the checklist refuses a tag without the note while it is set |
| L6.8 | The triggered capture procedure (OW-12) | `02-analysis/capture-procedure.md` | When a Moderate or High risk or a tornado watch is issued, capture the two-hour window and extend the palette | A reviewer can follow it cold; a checklist row points to it |
| L6.9 | The fallback-share test (L-2.3) | `tables.go` tests, `testdata/mrms/` | — | A frame whose fallback share exceeds the threshold (the PLAN report's number) fails; the IEM archive frames (OW-11) pass as a control; **MRMS's heavy end has no oracle before OW-12**, and the record says so (D-72) |

## WP-L7 — View bound (L-3)

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L7.1 | A bound set once | `view.go`, `map.go` | `type Bound struct{ MinZoom float64; W, S, E, N float64 }`; `(*Map).SetBound(Bound) error`, called again whenever the host's region changes (D-70: no paired option). **A box that crosses the antimeridian (W > E) is accepted**: Alaska's and the Pacific territories' regions need it | A zero-area box is refused; an Aleutians box is held |
| L7.2 | Held on every path | `view.go` (`moveLocked`), `map.go` (`viewAt`) | Both enforce it (W1-C) | M3: resize, fit, pan, zoom and fall-back each stay inside, including across the antimeridian |
| L7.3 | A host that sets none gets v0.1.0's behaviour (L-3.2) | `view_test.go` | — | The existing view tests pass unchanged, with a named test that sets no bound |

## WP-L8 — Fetch options and confinement (D-55)

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L8.1 | `FetchOptions`, taking effect at once (L-7.1, L-7.2, L-7.4) | `tiles.go`, `internal/fetch/fetch.go` | `type FetchOptions struct{ Transport http.RoundTripper; UserAgent string; Timeout time.Duration; AllowHTTP []string }`; `(*Map).SetFetchOptions(FetchOptions) error` | A transport set after `Source` is the one used for the next fetch |
| L8.2 | `Fetcher` (`tiles.go:13`, `:70`) removed, and `fetch.Checked` with it (D-62) | `tiles.go`, `internal/fetch/get.go` | — | The surface test records the removal; the changelog row exists |
| L8.3 | Bytes bounded; late answers dropped; time bounded while the context is honoured (L-7.3) | `internal/fetch/get.go` | The library reads the body through its limit; checks the deadline after the transport returns | An endless-body transport is cut at the limit; a transport that ignores its context blocks only its own `Work` call, and its late answer is discarded |
| L8.4 | `CheckedDialer`, exported at the root (L-10.3) | `tiles.go` (wrapper), `internal/fetch/fetch.go` | `func CheckedDialer() func(context.Context, string, string) (net.Conn, error)` | A host transport using it refuses a private address |
| L8.5 | One allow-list; scheme, host, port; plain http refused unless allowed (L-10.2) | `internal/fetch/fetch.go`, `internal/tiles/remote.go` | One list feeds both | Through `Map.Source`, with the library's and a host's transport: another host, another port and http each refused |
| L8.6 | The proxy decision per connection; the reserved ranges completed (L-10.3) | `internal/fetch/fetch.go` | — | A redirect to a host the proxy does not carry still meets the private-address check; each new range refused |
| L8.7 | The real version in the user-agent (L-13.1) | `internal/fetch/fetch.go` | `Version` stays a constant; the release checklist bumps it, and L10.3's `--release` check refuses a tag it does not match | At the tag, the user-agent names the tag; `-dev` at a tag is refused |

## WP-L9 — Cache age and purge (D-56)

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L9.1 | The file time is the fetch time; reads never touch it (L-9.4) | `internal/tiles/disk.go` (`Load`, `Flush`) | The file time is set only at store | Load a tile hourly for a day: its file time never changes |
| L9.2 | A maximum age, enforced at `CacheRoot` and in every job; oldest-fetched evicted first, tiles in view protected; a future time counts as expired (L-9.4) | `tiles.go`, `internal/tiles/disk.go`, `internal/tiles/pipeline.go` | `(*Map).SetCacheMaxAge(time.Duration) error` (D-70); `CacheRoot`'s signature unchanged | A tile past its age is fetched again; a future-dated one too; `CacheRoot` over an aged root drops them; a tile in view survives eviction over the cap |
| L9.3 | The cache generation (L-9.3) | `internal/tiles/pipeline.go`, `map.go` | A counter raised by `Purge`, a `CacheRoot` change or off, and `Close`; a store lands only if unchanged | A fetch in flight across each of the three writes nothing |
| L9.4 | `Purge` of everything; replaced roots released (L-9.2, L-9.3); counts reported (L-9.5) | `tiles.go` | `Purge() (PurgeReport, error)`; `type PurgeReport struct{ Removed, Failed int }` (D-70: one call) | After `Purge`, nothing under the root, memory empty; the old root's descriptor is closed; a failed remove shows in `Failed` |
| L9.5 | Failures raised and counted honestly (L-9.5) | `internal/tiles/pipeline.go`, `internal/tiles/disk.go` | `cache-write-failed` raised; only successful removals counted | A read-only root raises the warning; a failed remove is not counted |
| L9.6 | A root others can read is warned of (L-9.6); `ReadBack` reads through its limit (L-12.6) | `internal/tiles/disk.go` | — | A 0755 root warns; an oversized file is not read whole |

## WP-L10 — Defects, close-out, and the gate's own evidence

| # | Task | Files | Shape | Test first (RED) |
|---|---|---|---|---|
| L10.1 | A test lists the library's environment reads (L-13.8) | `rules_test.go`, `contract.md` | Allowed: `look.go`'s colour-depth read; the proxy variables `net/http` reads are named in the contract (D-72) | A new `os.Getenv` in a library file fails |
| L10.2 | Doc comments corrected (L-13.4) | `internal/tiles/disk.go`, `look.go` | Text | The two comments' claims each asserted by a test of the behaviour they describe |
| L10.3 | A gate test refuses the **final** tag while its checklist is unfinished (L-5.3); release candidates `v0.2.0-rc.N` are tagged as packages land, each naming the packages it holds (D-69) | `gate_test.go`, `scripts/gate`, `release-checklist.md` | A `--release` check; an rc row in the checklist | An unticked row refuses the final tag; all ticked passes; an rc tag is not refused |
| L10.4 | A pinned `govulncheck` at the tag, requiring no reachable finding (L-5.5) | `scripts/gate`, `release-checklist.md` | The version pinned | An injected vulnerable module fails the release check |
| L10.5 | Hosted CI (L-6.1) | `.github/workflows/gate.yml` | The gate on Linux; the second architecture on a runner that has it; actions pinned by SHA, `permissions: contents: read` (D-72) | A run on the branch, green; **the workflow fails when the architecture leg did not run** |
| L10.6 | The default tile memory cache documented against a large view (L-6.3) | `contract.md`, `contract_test.go` | Text | L1.6's sentence test covers it |
| L10.7 | **OW-4:** the 40-second `FuzzAgree` freeze examined | `06_docs/gate-runs.md` | `FuzzAgree` under the limiter, ten runs, each recorded | Ten runs recorded; a freeze, if seen, gets a cause or an open row with its evidence |
| L10.8 | **M4:** a 12-frame loop at dot resolution adds ≤ 1 MiB of heap over an empty map, within ±5 % over an hour (D-61); **re-measured on 24 region frames and brought back as a number** (D-68) | `memory_test.go`, `scripts/gate` | A five-minute form in the full gate; the hour form in `scripts/gate --soak`, run once before SHIP and recorded | The five-minute form's bound; the hour run in the SHIP report |
| L10.9 | **M1:** the ground truth first, then the non-visual arm, then the visual arm (D-24, D-71) | `testdata/loops/`, the SHIP report | Five recorded loops from real radar, the OW-11 outbreak and a several-cell scene among them; the non-visual arm read from watchpost's worded description (W9.7) | **Human-graded:** done when the HUM LEAD's scores for both arms are recorded |
| L10.10 | **M6:** five consecutive full runs with no unattributable failure before SHIP | `06_docs/gate-runs.md` | Counted from the log | A test reads the log and reports the current run of green |
| L10.11 | **v0.2.0's REVIEW covers everything v0.1.0 shipped** (L-5.4, D-72): the review's scope written as a list of v0.1.0's packages and requirements, and the red team briefed on it | `release-checklist.md`, `08-reports/` | A scope statement | The checklist row names the scope; the SHIP report shows each item reviewed |
| L10.12 | The integration map and both plans kept in step (D-72): each plan records the map commit it was reconciled against, and a docs-lane test refuses a plan that names a work package the map does not | `scripts/gate`, `gate_test.go` | — | A plan citing a package missing from the map fails the docs lane |

## Dependencies within packages

L2.1 → L2.2 → L2.3 → L2.4 → L2.5 → L2.6 → L2.7 → L2.8 · L4.1 → L4.2 → L4.3; L4.4 and L4.5 need L4.1;
L4.10 needs L4.4 · L3.1 and L3.2 need L4.4 and L4.7 · L3.3 → L3.4 → L3.5; L3.6 → L3.7 → L3.8 ·
L3.9–L3.11a need L3.5 and L5.1 · L5.1 → L5.2 → L5.3–L5.5; L5.5 and L5.6 need L6.1 · L6.1 → L6.2, L6.3 →
L6.4 → L6.5, L6.6 → L6.9 · L8.1 → L8.2, L8.3; L8.5 → L8.6 · L9.3 needs L8.1; L9.1 → L9.2.

## The trace

| Requirement | Task | | Requirement | Task |
|---|---|---|---|---|
| L-1.1 | L2.1 | | L-7.1, L-7.2, L-7.4, L-13.2 | L8.1, L8.2 |
| L-1.2, L-1.10a, L-1.10f | L4.8 | | L-7.3 | L8.3, L1.6 |
| L-1.3 | L4.9 | | L-8.1, L-8.5, L-8.7 | L3.9–L3.11a (D-65) |
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
| L-5.4 | L10.11 | | the plans in step | L10.12 |
| RK-1 | L4.10 | | NFR-1 (breaks listed) | L1.5 |
| L-10.1 | existing: `TestTileJSONAddressesObeyFetchRules`, `FuzzTileJSON` | | L-13.3 | done (D-38) |
| L-13.7 | deferred (D-29), no task | | L-8.2 | context for L-8.1, no task |

## Deviations from DISCOVER, recorded

- **L-7.3's time promise** is what a library that starts no goroutine can keep (D-55).
- **L-8.4 withdrawn** (D-50): the 16-colour shade defect was a miscount.
- **`Fetcher` and `fetch.Checked` are removed** and L-7.1 reworded (D-62).
