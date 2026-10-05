---
title: "v0.2.0 Radar loops — DISCOVER wave 1 findings"
date: 2026-09-23
phase: DISCOVER (RCC)
sev: SEV-0
status: "COMPLETE — W1-A loops/playback/memory, W1-B contract/surface/triage/fetcher/CI, W1-C host-facing gaps; synthesis at the end"
---

# Wave 1 — what the library actually does, measured against what its record says

Three read-only surveys of this tree at `532806b`, dispatched with non-overlapping scope while
another project's gate held the machine, so **nothing below was run** — every finding is read from
source and cited to `file:line`. The headline claims were then re-checked by hand against the code
before this document was written. Watchpost's 0.18.0 DISCOVER record is cited as *watchpost …*
(D-12 makes it admissible) and was re-verified here wherever it made a claim about this library.

## W1-A — Loops, playback, memory

**Two ways a loop would fail silently, both confirmed in code:**

1. **The picture would freeze.** The renderer reuses its last frame when the overlays' version and
   slice lengths are unchanged (`internal/render/frame.go:285-302`). A frame advance changes
   neither, so the loop stops moving while everything reports success. A frame selector has to reach
   that comparison, the way `MarkerPhase` already does.
2. **Every refresh would re-decode the whole loop.** A re-`Set` drops every prepared raster for that
   overlay (`internal/overlay/store.go:453`, `cache.go:262-271`). The shared `Classified` set that
   could save them is 1 MiB (`classified.go:14`) and evicts oldest-first, so at 596×304 it holds five
   of twelve frames and cycling through them hits almost nothing. That undoes watchpost's measured
   economy: a refresh should cost one request, not twelve decodes.

**What else a frames field touches:** validation (every frame, a loop total, times in order, an
explicit gap marker — a missing PNG is refused today, `image.go:114-118`), the store's per-id
`rasters`/`reports`/warnings (`store.go:121-123, 337-350`), the job key (`cache.go:328-333`),
staleness (`Valid`/`Keeps` exist once per overlay, `fresh.go:50-62`), and `Describe`
(`describe.go:166, 245-256`).

**Contract §9's frame-advance claim is false.** `NextCall` has three sources — marker blink, overlay
staleness, tile retry (`clock.go:93-112`) — and none is a frame. Blink is gated on blinking places and
fixed at 400 ms.

**Playback today.** `ReduceMotion` sets markers steady and makes `Motion.Next` return zero
(`look.go:206-221`, `motion.go:28-36`); staleness and retry deadlines still fire, so its comment
("nothing is ever due") overstates it. The clock machinery that exists (`Animate`, `FollowClock`,
phase computed from time rather than call count) is the right model for choosing a frame. The flash
ceiling is `BlinkPeriod` 800 ms, under D-56's 2.5 per second.

**Proposed shape (signatures only — the loop's API is PLAN's):**

```go
type LoopFrame struct { PNG []byte; Valid time.Time; Gap bool }   // element type; not "Frame", which exists
type Playback uint8                                               // PlaybackOff, PlaybackSlow, PlaybackNormal
func (m *Map) Playback(id string, p Playback) error               // on Map, not Overlay: a re-Set drops rasters
func (m *Map) Shown(id string) (index int, valid time.Time, gap bool, ok bool)
```

`ReduceMotion(true)` implies Off, as NFR-21 ("stops all animation") requires. **Contract §9's
`Frames []Frame` collides with the exported `Frame`** (`map.go:58`) and must be renamed.

**Memory, at one byte a pixel (`image.go:260`):**

| Frame | Pixels | 12 frames | vs the 0.25 MB image budget | Three unshared maps |
|---|---|---|---|---|
| 298×152 | 45,296 | 0.54 MB | 2.2× | 1.63 MB |
| 596×304 | 181,184 | 2.17 MB | 8.7× | 6.52 MB |
| at the per-image cap | 250,000 | 3.00 MB | 12× | 9.0 MB |

NFR-3 is ≤ 4 MB live and ≤ 8 MB peak. The smaller frame fits if budgets are rebalanced; the larger
takes more than half the live line for one overlay, and two sources at that size do not fit. There
is **no store-wide image cap at all** (`store.go:20` caps vertices only). Host PNGs are extra and
uncounted by `OwnedBytes` (`store.go:507-530`).

**Describe over a loop.** Recommended: it answers for the **newest non-gap frame**, with that
frame's valid time, whatever is on screen — so a description depends only on the data, and watchpost's
"freeze the description with the picture" works for free.

## W1-B — Contract, surface, triage, fetcher, hosted CI

**Contract drift is wider than the brief recorded.** Names the contract states that do not exist:
`SetSize`, `Pan`, `ZoomAround`, `BorrowCheck(on)`, `Focused()`, `Frame.Line`/`WriteTo`/`String`;
`SharedCaches` is an `Option`, not a call. Some twenty exported names are never mentioned. **Five
behavioural claims are false**, three on paths v0.2.0 touches:

- `Work` and `Settle` return released ids — they do not (`work.go:17`, `map.go:64-69`);
- a panicking `Render` returns a "failed" frame with the last good rows — it returns an empty frame
  and an `Internal` error, and `Status` has no failed value (`guard.go:22-29`, `map.go:34-38`);
- `NextCall` has a frame-advance source — it does not;
- `CacheRoot(path)` is really `CacheRoot(dir, capBytes)`;
- requirement FR-22b says "the library checks what comes back" — `fetch.Checked` exists
  (`internal/fetch/get.go:149`) and **nothing outside tests calls it**.

**The surface test compares names only.** `TestPublicSurfaceSnapshot` records `func N`, `type N` and
so on (`surface_test.go:22-128`); its comment says "a function with its signature" (`:65`), but no
signature is recorded, so signature changes and fields behind type aliases pass unseen.

**The nineteen-item triage never existed.** "Nineteen" appears only in D-123 and in this brief
quoting it; no list is in either tree, any branch, or any commit body. And **D-123's step 1 was never
closed**: the integration review's measurement, M1 and recommendation sections are blank
(`integration-review.md:70-110`), the UAT findings table is empty (`uat-record.md:84`), Finding 4's
tag decision "has not been taken" (`:254, :303`), `release-checklist.md:7` still reads "Not started",
and no ruling authorised the tag. **v0.1.0 is published and works; the record of why it was ready
does not exist.**

**v0.1.0 defects found on the way:** the tagged library identifies itself as `go-tuimaps/0.1.0-dev`
(`fetch.go:24`, confirmed at the tag) — every tile request from a released host carries a dev version.

**The fetcher (L-7) is about one line.** `Fetcher` is `func(context.Context, fetch.Request)` and
`fetch.Request` sits under `internal/`, so a host cannot spell it. Minimal surface:
`type FetchRequest = fetch.Request`, `type FetchOptions = fetch.Options`,
`func (m *Map) FetchOptions(o FetchOptions) error` — and wire `fetch.Checked` around a host fetcher.
**A host fetcher bypasses** TLS ≥ 1.2, redirect confinement, private-address refusal, timeouts and the
body limit (`fetch.go:95-219`); what still holds is the TileJSON limits, the decoder's body cap, and
tile-host confinement. The contract should state that split.

**Tile-host confinement (L-10) is real and unit-tested only.** `checkAddress`
(`internal/tiles/remote.go` ~95-131) refuses another host, port, user info, tokens and downgrades;
`TestTileJSONAddressesObeyFetchRules` and `FuzzTileJSON` cover it. It applies to the TileJSON
document, not the transport, so it holds even with a host fetcher. Making it contract needs one
sentence and one root-package test through `Map.Source`.

**Hosted CI (L-6.1)** must reproduce the floor toolchain, the throw-away workspace, both test legs,
sixteen fuzz targets at their counts, `govulncheck`, the licence scan and five cross-compiles. **The
second-architecture leg relies on emulation this Mac has and a Linux runner does not.** The gate's
NOT RUN line is hard-coded and stale: it still says no tag exists.

## W1-C — The view bound, colour, patterns, caches

**The MRMS table (L-2): no exact, citable table exists.** Twenty consumers of `conus_bref_qcd` were
found; those with a table say they sampled the server's legend. The legend is 500 px over 90 dB
(0.18 dB a pixel), so resolution is not the limit — whether map pixels match legend pixels is.
The matcher is exact-first, then nearest within ΔE 10 (`image.go:195-205`); an unmatched pixel draws
nothing and is counted and reported. **The measurement that settles "exact or approximate" is cheap:**
decode real frames with `Exact: true` against a legend-derived table and read `Report.Unmatched`.

**The view bound (L-3).** Every view change passes through `moveLocked` (`view.go:207-220`) or
`viewAt` (`map.go:245-265`) — **enforcing the bound in those two places covers every path.** The
resize fall-back to `WholeWorld` at `map.go:264-265` looks dead (every size that reaches it fails
the same validation first), matching watchpost's observation. Proposed: `Bound{MinZoom, W, S, E, N}`,
`WithBound(b)`, `Map.Bound()`; zoom never below the minimum and the centre never outside the box;
**in a window wider than the box at the minimum zoom, the box is centred and the frame says so**
rather than silently zooming in. A zero-area or antimeridian-crossing box is refused when set.

**Patterns (L-8).** The hatch runs only at `NoColour` because FR-18a is scoped to "with no colour"
(`frame.go:513, 555`). Drawing it at colour-on depth is cheap — it fills only blank cells, and `╱╲╳`
are already on the closed glyph list. But **three strokes cannot rank five levels**: Extreme and
Severe share `╳`, **Minor and Unknown share `╱`** (`hatch.go:27-36`). Better channels: a severity
word as an edge label, and a distinct outline dash per level, with spacing kept as a secondary cue.

**An alert tint hides the radar underneath it — by design.** The background order is ground, water,
image or field, then alert tint on top, and an image under a tinted cell is explicitly skipped
(`frame.go:653-663`, `!tinted`). At colour-on depth, **radar inside a warning polygon is replaced by
the warning's colour** — the one place a reader most needs to see the storm. It follows v0.1.0's
FR-12 composite order, so it is a product decision to revisit, not a bug.

**Caches (L-6.3, L-9).** The memory cap is a target, not a limit: tiles a live view needs are never
evicted, so a view needing 909 KB holds 909 KB and keeps no spares, raising `CacheUnderNeed`
(`cache.go:40-43, 232-280`). The disk cache has no expiry; its "recency" is a file mtime rewritten at
most hourly, so a max age would mean *time since last read, to within an hour*. File paths encode
tile coordinates. **`Map.Purge` already exists** (`tiles.go:145-165`) — the brief said it was
missing; what is missing is its **scope** (it empties only the current source, and not memory,
shared caches or decoded pictures). Several internal doc comments have drifted onto the wrong
functions (`disk.go:100, 328, 347`).

## Cross-cutting synthesis

1. **Composition — the loop is three changes, not one field.** A frames field (W1-A) is useless
   without a renderer change test that sees a frame advance and a store that keeps frames across a
   refresh. Any of the three alone ships a loop that freezes or re-decodes.
2. **Composition — the loop's cost decides its API.** Whether frames live per map or in a shared set
   depends on the byte budget, and the budget depends on which frame size the host uses. **A byte
   budget for a loop is a ruling before any signature.**
3. **Convergence — the contract describes an API that does not exist.** W1-A (frame-advance, the
   image cap a host "raises"), W1-B (five false claims, phantom calls) and W1-C (a purge the brief
   said was missing) all converge: the document a host reads is materially wrong, and the only test
   holding it checks names. **L-4 is the foundation for the rest**, because every other requirement
   will be written into that document.
4. **Convergence — v0.1.0 shipped with its readiness record unfinished.** W1-B's blank review
   sections, empty findings table and missing tag ruling, plus the `0.1.0-dev` version string and an
   unwired safety check, point the same way: the release was tagged before its own close-out ran.
   That is a finding about the process, and it is recorded as one.
5. **Contradiction — the brief is wrong about Purge.** L-9.2 asks for a purge call; one exists. The
   gap is its scope. The brief needs correcting before it misdirects PLAN.
6. **Contradiction — the host's primary view is undermined by the library's composite order.**
   Watchpost ruled radar its most common view; the library draws a warning's tint over the radar
   inside the warning. Neither record noticed.
7. **Risk update — M4 (embed cost with a loop).** Newly measurable, and at risk at the larger frame
   size.
8. **Risk update — M5 (contract truth).** Worse than the brief assumed: a name-level test would pass
   today with five false behavioural claims standing.
9. **Open questions answered.** Is there an exact MRMS table? No — and the measurement to settle
   approximate-versus-exact is named. Does a frame-advance source exist? No. Where must a view bound
   live? In two functions. Can a host write a fetcher? Not without one exported alias.
10. **Open questions raised for the HUM LEAD:** the loop byte budget and frame size; whether the
    alert tint should cover radar; how severity is distinguished without colour (a word, a dash, or
    both); whether the MRMS table ships "approximate" if the measurement says so; and how v0.1.0's
    unfinished close-out is handled — reconstructed, or recorded as not recoverable.
11. **v0.1.0 defects to fix in this release regardless:** the `-dev` version string; `fetch.Checked`
    unwired; the stale NOT RUN line; the drifted cache doc comments; `ReduceMotion`'s overstated
    comment; the contract's false claims.
12. **Implications for the next wave:** the MRMS `Unmatched` measurement and a loop memory
    measurement at both frame sizes both need the machine and both need to be run, not read — they
    wait for the gate queue.
