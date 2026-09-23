---
title: "go-tuiMaps v0.2.0 — Radar loops — REQUIREMENTS (normative)"
date: 2026-09-23
phase: DISCOVER (RCC)
sev: SEV-0
authority: HUM LEAD
status: "NORMATIVE (D-23). On any conflict with the brief or another document, this file wins. Every change is a row in 02-analysis/rulings.md."
---

# v0.2.0 requirements

**How to read this.** The brief (`project-brief.md`, and issue #2) says what the release is and
why. **This file is the requirement set PLAN designs against** (D-23). Requirements keep the brief's
L-numbers and extend them. Each row gives its source, and either the instrument that will hold it or
**NO INSTRUMENT YET**, which PLAN must close. **OPEN — R1-n** marks a row whose final wording waits
on a ruling from the DISCOVER exit red team (`08-reports/red-team-discover.md`).

Sources: **L-** the brief; **D-** this release's rulings; **S29-** specimen 29
(`../go-tuimaps/02-analysis/specimens/README.md`); **W1-/W2-** waves 1 and 2; **v0.1.0 FR-/NFR-** the
first release's record.

## What changes on screen

For designers and PMs before engineers, the three visible decisions:

- **Radar moves** (L-1): a loop of frames, with playback off, slow or normal.
- **An alert over radar lets the radar show through** (D-14, specimen 29d–f): the warning's tint is
  blended into the radar colours, and its outline and label stay on top.
- **Severity reads without colour** (D-17): a severity word in the label, and a distinct outline
  dash for each level, at every colour depth. A specimen is owed before PLAN commits to it.

## L-1 — Loops

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-1.1 | An image overlay carries several frames, each with its own valid time; zero frames means the single image of v0.1.0 (additive). | L-1.1, C-1 | NO INSTRUMENT YET |
| L-1.2 | A frame may be a **gap**: a missing frame is stated as a gap, never replaced by a neighbour. | L-1.2, M2 | NO INSTRUMENT YET |
| L-1.3 | Each frame's own valid time drives staleness and the description, so an old loop never looks current. | L-1.3, M2 | NO INSTRUMENT YET |
| L-1.4 | The library never fetches radar; the host hands in the data, and the library turns it into frames. | L-1.4, watchpost 0.18.0 D-35 | NO INSTRUMENT YET |
| L-1.5 | Playback is controllable, at least off / slow / normal, for every looped overlay; the control is the library's, exposed to the host and through it to the listener. | L-1.5 (HR-9) | NO INSTRUMENT YET |
| L-1.6 | **A frame advance must be seen by the renderer.** Today the renderer reuses its last frame when the overlays' version is unchanged (`internal/render/frame.go:285-302`), so a loop would freeze while reporting success. | W1-A | NO INSTRUMENT YET |
| L-1.7 | **A refresh does not re-decode the whole loop.** Today a re-`Set` drops every prepared raster (`internal/overlay/store.go:453`). | W1-A | NO INSTRUMENT YET |
| L-1.8 | `NextCall` reports the next frame change; the contract's claim that a frame-advance source exists becomes true. | L-1.1, C-2 | NO INSTRUMENT YET |
| L-1.9 | Every frame is validated when handed in, the loop has a total, frame times are in order, and a gap is explicit. | W1-A | NO INSTRUMENT YET |
| L-1.14 | **Frames are copied at hand-in, never borrowed;** decode re-checks the size cap and the budget. | D-30 (F1) | NO INSTRUMENT YET |
| L-1.15 | A hard maximum frame count, gaps included; the playback interval is library-owned and floored, never derived from host-supplied times. | D-30 (F2) | NO INSTRUMENT YET |
| L-1.10a | The shown frame's time, or "gap", is text on the map at every depth, beside the `stale` word. | D-25 (A2) | NO INSTRUMENT YET |
| L-1.10b | A host can step to the previous, next and newest frame and seek to any frame; the library's animation clock drives the loop. | D-25 (A3) | NO INSTRUMENT YET |
| L-1.10c | `ReduceMotion(true)` forces playback off; turning it off restores the host's setting. | D-25 (A4) | NO INSTRUMENT YET |
| L-1.10d | A state read: the playback in effect (including "off because reduce motion is on"), frame index, count, the frame's time, gap. | D-25 (A5) | NO INSTRUMENT YET |
| L-1.10e | A frame advance is signalled apart from a data change and does not change the description's cache key. | D-25 (A6) | NO INSTRUMENT YET |
| L-1.10f | Off shows the newest non-gap frame with its age. | D-25 (A15) | NO INSTRUMENT YET |
| L-1.10g | A library-owned numeric ceiling on frame changes a second, every change counted toward WCAG 2.3.1, bounded by D-56's 2.5 a second; a gap holds the last real frame with its time reading "gap", never an empty frame. Slow and normal speeds are PLAN's, within it. | D-25 (A16, P-5) | NO INSTRUMENT YET |
| L-1.11 | **Playback defaults to off** until the host chooses; the state read says "off (default)". | D-26 (A14) | NO INSTRUMENT YET |
| L-1.13 | **One standard playback API** a host wires straight to its own controls and Settings: every control of L-1.10 (play/pause, slow/normal, step, seek, newest), the state read and a change signal, one consistent shape across every looped overlay. | D-26 | NO INSTRUMENT YET; measure: a host wires controls and a Settings row with no playback state of its own (the example, and watchpost) |
| L-1.12 | **Storm motion without the animation (D-24).** The description reports observed motion from the frames: for each described place, where the heavier rain was at the oldest usable frame and at the newest (distance, direction, time), whether it came closer, moved away or held, and the span the loop covers. Worded as observation, never forecast, and relative to a named place, never "you" (D-29). | D-24, D-29; red team A1, B-4 | NO INSTRUMENT YET; M1's non-visual arm, scored from the description alone |

## L-2 — The MRMS table

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-2.1 | A table for NOAA/NCEP MRMS `conus_bref_qcd` ships, built from **the colours MRMS actually uses** (111 seen across 18 frames), each valued from its position on the legend, labelled **approximate**. | D-19, W2 M-A | NO INSTRUMENT YET |
| L-2.2 | Matching is exact first; the tolerance fallback is kept for colours not yet seen, and such pixels are counted and reported. | D-19 | Existing unmatched-colour warning (`kinds.go:50`) |
| L-2.3 | **PLAN pre-plans the heavy end:** colour affordances for heavy-rain colours not yet observed, and a better way than the nearest legend colour to value the oranges and reds. | D-19 | NO INSTRUMENT YET; a test that fails when the fallback carries a meaningful share of pixels is the candidate |
| L-2.4 | A table derived from the legend is never called exact. | L-2.2 | — |

## L-3 — The view bound

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-3.1 | A host sets a minimum zoom and a bounding box once, and the view stays inside them on every path: resize, fit, pan, zoom and the fall-back to the whole world. | L-3.1, W1-C | NO INSTRUMENT YET (M3) |
| L-3.2 | A host that sets nothing gets v0.1.0's behaviour. | L-3.2 | NO INSTRUMENT YET |

## L-4 — The contract matches the code

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-4.1 | `contract.md` names nothing the public surface lacks, and the surface nothing the contract lacks. | L-4.1 | NO INSTRUMENT YET (M5) |
| L-4.2 | The five false behavioural claims are corrected: `Work`/`Settle` released ids; a panicking `Render`'s "failed" frame; `NextCall`'s frame-advance source; `CacheRoot`'s signature; FR-22b's "the library checks what comes back". | W1-B | NO INSTRUMENT YET |
| L-4.3 | The surface test records signatures, not names only. | W1-B | NO INSTRUMENT YET |
| L-4.4 | The README's user-agent claim ("a token you set") is made true or corrected. | red team D-3 | **OPEN — R1-13** |

## L-5 — v0.1.0's close-out (supersedes the brief's L-5.1)

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-5.1 | ~~The nineteen-item triage is written.~~ **Superseded by D-18:** the triage never existed and is recorded as not recoverable. | D-18 | — |
| L-5.2 | One row in v0.1.0's record says its close-out did not run, with pointers to the evidence that does exist. | D-18 | Owed (OW-1) |
| L-5.3 | A gate test refuses a release tag while its checklist is unfinished. | D-18 | NO INSTRUMENT YET |
| L-5.4 | v0.2.0 is the library's first release: its REVIEW and close-out cover everything v0.1.0 shipped. The `v0.1.0` tag stays. | D-18 | — |
| L-5.5 | The release checklist records a pinned `govulncheck` at the tag commit. | D-30 (F9) | The release checklist |

## L-6 — Found while preparing the brief

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-6.1 | Hosted CI reproduces the gate, including the second-architecture leg a Linux runner cannot emulate as this machine does. | L-6.1, W1-B | NO INSTRUMENT YET |
| L-6.2 | (Became L-7.) | L-6.2 | — |
| L-6.3 | The default memory tile cache is at least documented against one large view (909 KB needed, 500 KB default). | L-6.3 | NO INSTRUMENT YET |
| L-6.4 | ~~The fuzz legs fail by accident; every commit is blocked.~~ **Fixed by construction** (D-9, D-10): run-count budgets. **Not reproduced** — the cause is probable, not proven. The 40-second freeze seen once in `FuzzAgree` stays open (OW-4). | D-9, D-10 | M6 |

## L-7 — A host can write a fetcher

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-7.1 | A host can supply a fetcher without reflection: the request type and its options are exported. | L-7.1 | NO INSTRUMENT YET |
| L-7.2 | Tiles can go out under the host's own user-agent. | L-7.2 | NO INSTRUMENT YET |
| L-7.3 | **Every host fetcher is wrapped by `fetch.Checked`** (read limit, timeouts); the contract states what a host fetcher takes on (TLS, redirects, private addresses). | D-30 (F3), W1-B | NO INSTRUMENT YET |
| L-7.4 | Setting a fetcher takes effect at once. | D-30 (F4) | NO INSTRUMENT YET |

## L-8 — Meaning without colour

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-8.1 | An alert area's label carries its **severity as a word**, and each of the five levels has **its own outline dash**, at every colour depth; the hatch stays as a secondary cue. | D-17 | Specimen owed before PLAN (OW-2) |
| L-8.2 | The hatch strokes cannot rank five levels: Extreme and Severe share `╳`, **Minor and Unknown share `╱`**. | L-8.2 corrected, W1-C | — (context for L-8.1) |
| L-8.3 | **An image or field never erases furniture.** Outline, label, marker, `stale` word, the no-tiles notice, frame time, scale and credit survive an image or field **at every depth**. Today they are erased at NoColour **and Colours16**. | S29-2, red team A7 | NO INSTRUMENT YET; tests at NoColour and Colours16 |
| L-8.4 | **At 16 colours, rain classes stay distinguishable.** Today every rain cell draws as `░`. | round 1 verification | NO INSTRUMENT YET (cause not yet found) |
| L-8.5 | A label that does not fit falls back to the severity word alone; if that does not fit, the description carries it and the frame reports the dropped label to the host. | D-28 (A8) | NO INSTRUMENT YET; the D-17 specimen (OW-2) |
| L-8.6 | The five outline dashes differ from each other, from the line-overlay dash (5 on, 4 off) and from every basemap stroke. | D-28 (A8) | NO INSTRUMENT YET; the D-17 specimen (OW-2) |
| L-8.7 | The hatch draws at Colours16 as well as NoColour. | D-28 (A8) | NO INSTRUMENT YET |
| L-8.8 | Outlines hold 3:1 (WCAG 1.4.11) and labels 4.5:1 against the blended inside and the plain radar outside, on both grounds, at truecolor and 256 colours, and at 16 against its reference table. | D-28 (A9) | The checker, extended |

## L-9 — Cache retention and purge

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-9.1 | A host-settable **maximum age** for the disk cache. | L-9.1 | NO INSTRUMENT YET |
| L-9.2 | ~~A purge call.~~ **`Map.Purge` exists** (`tiles.go:145-165`); the gap is its **scope**: it empties only the current source, not memory, shared caches or decoded pictures. | L-9.2 corrected, W1-C | NO INSTRUMENT YET |
| L-9.3 | **Purge empties everything**: every source, the memory caches, decoded pictures; in-flight writes cannot land after it. | D-30 (F5) | NO INSTRUMENT YET |
| L-9.4 | Maximum age is enforced at `CacheRoot` and in every job; future-dated files count as expired. | D-30 (F5) | NO INSTRUMENT YET |
| L-9.5 | `cache-write-failed` is raised; only removals that succeeded are counted. | D-30 (F6) | NO INSTRUMENT YET |
| L-9.6 | Contract guidance: the cache root under the OS user cache directory, owner-only; purge is not secure erasure. | D-30 (F8) | — |

## L-10 — Tile-host confinement

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-10.1 | A source's tiles come only from that source's host; this is contract, stated and kept. | L-10.1 | `TestTileJSONAddressesObeyFetchRules`, `FuzzTileJSON`; a root-package test through `Map.Source` owed |
| L-10.2 | **One allow-list** for fetching and TileJSON, covering scheme, host and port under any fetcher; plain http refused unless the host allows it. | D-30 (F7) | NO INSTRUMENT YET; a root-package test through `Map.Source` with both fetchers |

## L-11 — Alerts over radar

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-11.1 | Inside an alert area, **the tint blends over the radar**: each radar cell keeps its class colour shifted toward the tint; outline and label stay on top. Replaces v0.1.0 FR-12 step 4 for images and fields. | D-14 | NO INSTRUMENT YET |
| L-11.2 | The blend must not wash out: every tinted class stays at least 10 (D-88's measure) from every other class, inside and outside the area, on both grounds. | D-14 | The ramp checker, extended |
| L-11.3 | The light ground is solved by adjusting the radar ramps or the alert colours. | D-14 | — |
| L-11.4 | **If no blend passes on a light ground, the light ground uses 29b's order**: radar over the tint, the warning carried by its outline, label and L-8.1's severity word and dash. The dark ground keeps the blend. | D-27 (A10, B-6) | The checker decides which applies |

## L-12 — Memory

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-12.1 | A **total image budget per map**, settable by the host; a hand-in over it is refused with an error that says so. | D-20 | NO INSTRUMENT YET |
| L-12.2 | The default fits v0.1.0 NFR-3 (4 MB live): about 3 MB of images — one 12-frame loop at 596×304, or about five at dot resolution (298×152). | D-20, W2 M-B | M4 |
| L-12.3 | Guidance steers hosts to frames at the dot grid of the view (298×152 at 149×38). A host that raises the budget owns the memory it asks for, and the contract says so. | D-20 | — |
| L-12.4 | The budget counts retained PNG bytes as well as pixels; a frame's file size is capped relative to its pixel count. | D-30 (F2) | NO INSTRUMENT YET |

## L-13 — v0.1.0 defects fixed in this release

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-13.1 | The library identifies itself with its real version, not `0.1.0-dev` (`fetch.go:24`). | D-18, W1-B | NO INSTRUMENT YET |
| L-13.2 | `fetch.Checked` is wired, or removed. | W1-B | NO INSTRUMENT YET |
| L-13.3 | The gate's NOT RUN line is current. | W1-B | — |
| L-13.4 | The drifted cache doc comments and `ReduceMotion`'s overstated comment are corrected. | W1-B, W1-A | `make lint`-class review |
| L-13.5 | **The description lists the alerts being shown**, each with its name, severity word and valid time or stale mark, independent of any place. | D-29 (A11) | NO INSTRUMENT YET |
| L-13.6 | **The description never says "you."** For a discrete place the host names, it says whether that place is inside, outside or nearby, as v0.1.0 does. | D-29 | NO INSTRUMENT YET |
| L-13.7 | *Deferred, "necessity unproven yet" (D-29):* intensity in words for image answers; a summary of what is in view; new place-to-weather statements beyond inside / outside / nearby. | D-29 (A12, A13) | — |

## Non-functional

| # | Requirement | Source |
|---|---|---|
| NFR-1 | Additive only: a breaking change to the v0.1.0 contract is a HUM LEAD ruling. | C-7 |
| NFR-2 | Memory within v0.1.0 NFR-3 at the default budget, flat over an hour with a 12-frame loop. | M4, D-20 |
| NFR-3 | Motion: no flash above D-56's ceiling (2.5 a second); `ReduceMotion` stops all animation. | v0.1.0 D-56, NFR-21 |
| NFR-4 | The gate is green before every commit that touches anything but Markdown; a Markdown-only change takes the docs lane, run before committing. Every run is logged in `06_docs/gate-runs.md`. | D-15, D-31 |

## Risk register

Likelihood and severity are **H / M / L**, judged from the evidence cited.

| # | Risk | L | S | Mitigation |
|---|---|---|---|---|
| RK-1 | **A loop that freezes or re-decodes while every unit test passes** — the three changes (frames, renderer, store) land apart | M | H | L-1.6, L-1.7 each held by a test that sees a frame advance |
| RK-2 | **The MRMS heavy end is wrong on a severe-weather day** — the palette was seen on a quiet day (≤ 48 dBZ) | H | H | L-2.3; the fallback-share test |
| RK-3 | **No blend passes on a light ground** (every strength tried failed) | M | L | L-11.3; the named fallback L-11.4 (D-27) |
| RK-4 | **Watchpost's radar slips** because this release does | M | M | watchpost's ship-without-radar rule (its RK-4); a scope split is open (R1-15) |
| RK-5 | **A provider changes its palette or endpoint** (IEM has no SLA; MRMS no published table) | M | M | The unmatched-colour count is the detector (L-2.2) |
| RK-6 | **The contract stays wrong** — a names-only test passes with false behavioural claims | H | M | L-4.2, L-4.3 |
| RK-7 | **The gate hangs on a fuzz stall** — the count budget has no wall clock | L | M | Each fuzz leg runs under a wall-clock limit and fails as STALLED (D-31) |
| RK-8 | **A listener who cannot watch the animation is excluded** from the release's main point | M | H | L-1.12 (D-24); M1's non-visual arm. Residual: two-frame motion can mislead when cells grow or decay, so it is worded as observation |
| RK-9 | **Frames change after hand-in** — the library borrows host image bytes | L | H | L-1.14 (D-30) |
| RK-10 | **Evidence cannot be re-run** — wave 2 and specimen 29's blends came from throwaway programs | L | L | Filed with inputs and output in `02-analysis/programs/` (D-32); they may go stale as the API changes |

## Owed

| # | What | Due | Source |
|---|---|---|---|
| OW-1 | The row in v0.1.0's record saying its close-out did not run | before DISCOVER exit | D-18 |
| OW-2 | The D-17 specimen: severity word and five dashes, at 69×12 and 149×38, NoColour and Colours16 | before PLAN commits to L-8.1 | D-17 |
| OW-3 | S29-3: the Hamlin marker missing at 69×12 even without radar — diagnose | PLAN | S29-3 |
| OW-4 | The 40-second `FuzzAgree` freeze | PLAN | L-6.4 |
| OW-5 | ~~Correct wave 2's "71–80 %" and D-19's copy of it~~ **Done 2026-09-23**: 80.0–80.6 %, and wave 2's finding 2 restated | — | round 1 verification |
| OW-6 | F-2 (a pluggable architecture): its trigger — a second radar source — has fired | PLAN entry | red team H-10; OPEN R1-14 |
| OW-7 | Every **OPEN — R1-n** row above | before DISCOVER exit | round 1 |
