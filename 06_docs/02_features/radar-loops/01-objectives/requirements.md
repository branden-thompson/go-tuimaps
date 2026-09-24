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
**NO INSTRUMENT YET**, which PLAN must close.

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
- **Playback is off until the host turns it on** (D-26), and **the shown frame's time, or "gap", is
  written on the map** (D-25).
- **On a light ground the blend may not be used at all**: if no blend passes there, the warning is
  drawn by outline, label and dash over the radar (D-27).

## L-1 — Loops

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-1.1 | An image overlay carries several frames, each with its own valid time; zero frames means the single image of v0.1.0 (additive). | L-1.1, C-1 | NO INSTRUMENT YET |
| L-1.2 | A frame may be a **gap**: a missing frame is stated as a gap, never replaced by a neighbour. | L-1.2, M2 | NO INSTRUMENT YET |
| L-1.3 | **The newest non-gap frame's valid time drives the `stale` word and the description**, so an old loop never looks current and neither toggles as the loop plays. Stepping changes the picture and the frame-time text (L-1.10a), not the description. | L-1.3, M2, D-39 | NO INSTRUMENT YET |
| L-1.4 | The library never fetches radar; the host hands in the data, and the library turns it into frames. | L-1.4, watchpost 0.18.0 D-35 | NO INSTRUMENT YET |
| L-1.5 | Playback is controllable, at least off / slow / normal, for every looped overlay; the control is the library's, exposed to the host and through it to the listener. | L-1.5 (HR-9) | NO INSTRUMENT YET |
| L-1.6 | **A frame advance must be seen by the renderer.** Today the renderer reuses its last frame when the overlays' version is unchanged (`internal/render/frame.go:285-302`), so a loop would freeze while reporting success. | W1-A | NO INSTRUMENT YET |
| L-1.7 | **A refresh does not re-decode the whole loop.** Today a re-`Set` drops every prepared raster (`internal/overlay/store.go:453`). | W1-A | NO INSTRUMENT YET |
| L-1.8 | `NextCall` reports the next frame change; the contract's claim that a frame-advance source exists becomes true. | L-1.1, C-2 | NO INSTRUMENT YET |
| L-1.16 | **`Frame` carries the `Changed` and `FrameTicks` values it was drawn at**, so a host can prove a stored frame is current (watchpost D-41). | D-59 | NO INSTRUMENT YET |
| L-1.9 | Every frame is validated when handed in, the loop has a total, frame times are in order, and a gap is explicit. | W1-A | NO INSTRUMENT YET |
| L-1.14 | **Every host slice an image carries — PNG and table, for the single image and every frame alike — is copied at hand-in, and the copy is what is validated.** A later decode reads the size from the copy's header before it allocates, and re-checks the cap and the budget. | D-30 (F1), D-40 (S-1) | NO INSTRUMENT YET; a test that mutates the host's slices after hand-in |
| L-1.15 | A hard maximum frame count, gaps included; the playback interval is library-owned and floored, never derived from host-supplied times. | D-30 (F2) | NO INSTRUMENT YET |
| L-1.10a | The shown frame's time, or "gap", is text on the map at every depth, beside the `stale` word. | D-25 (A2) | NO INSTRUMENT YET |
| L-1.10b | A host can step to the previous, next and newest frame and seek to any frame; the library's animation clock drives the loop. | D-25 (A3) | NO INSTRUMENT YET |
| L-1.10c | `ReduceMotion(true)` forces playback off; turning it off restores the host's setting. | D-25 (A4) | NO INSTRUMENT YET |
| L-1.10d | A state read: the playback in effect, frame index, count, the frame's time, gap, and the span the loop covers (oldest to newest). **"Playing" only while the shown frame is actually advancing**, so a host that freezes the animation clock (`Animate`) reads the loop as stopped. When playback is off for more than one reason the state names the first of: reduce motion, the host's choice, the default. A loop's name is the host's to give, and the contract says so. | D-25 (A5), D-40 (A F9, F13) | NO INSTRUMENT YET |
| L-1.10e | A frame advance is signalled apart from a data change and does not change the description's cache key. | D-25 (A6) | NO INSTRUMENT YET |
| L-1.10f | Off shows the newest non-gap frame with its age. | D-25 (A15) | NO INSTRUMENT YET |
| L-1.10g | A library-owned numeric ceiling on changes a second, **one ceiling per map**, counting every moment anything on it changes — frames of every loop, marker blink, furniture — toward WCAG 2.3.1, bounded by v0.1.0 D-56's 2.5 a second; a gap holds the last real frame with its time reading "gap", never an empty frame. When two loops play, the frame time shown is the newest loop's, and the state read gives each. Slow and normal speeds are PLAN's, within it. | D-25 (A16, P-5), D-40 (A F6) | NO INSTRUMENT YET |
| L-1.11 | **Playback defaults to off** until the host chooses; the state read says "off (default)". | D-26 (A14) | NO INSTRUMENT YET |
| L-1.13 | **One standard playback API** a host wires straight to its own controls and Settings: every control of L-1.10 (play/pause, slow/normal, step, seek, newest), the state read and a change signal, one consistent shape across every looped overlay. | D-26 | NO INSTRUMENT YET; measure: a host wires controls and a Settings row with no playback state of its own (the example, and watchpost) |
| L-1.12 | **Storm motion without the animation (D-24, D-42).** The description reports observed motion from the loop's frames: where the heavier rain was at the oldest usable frame and at the newest (distance, direction, time), whether it came closer, moved away or held, and the span the loop covers — **relative to a named place where there is one, and relative to the view (its centre and edges) where there is none**, so a listener in a view with no names still hears it. **"Heavier" is one intensity threshold, set by the colour table and held for the whole loop.** The answer is **data**: two timed positions and a relation, with no field that can express a forecast (the library words nothing; a host does). Worded as observation, never forecast, and never "you" (D-29). | D-24, D-29, D-42; red team A1, B-4, A F2 | NO INSTRUMENT YET; M1's non-visual arm, scored from the description alone |

## L-2 — The MRMS table

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-2.1 | A table for NOAA/NCEP MRMS `conus_bref_qcd` ships, built from **the colours MRMS actually uses** (111 seen across 16 distinct times on one quiet afternoon, rain up to about 48 dBZ; see wave 2's Limits), each valued from its position on the legend, labelled **approximate**. | D-19, W2 M-A, D-40 (H D-13) | NO INSTRUMENT YET |
| L-2.2 | Matching is exact first; the tolerance fallback is kept for colours not yet seen, and such pixels are counted and reported. | D-19 | Existing unmatched-colour warning (`kinds.go:50`) |
| L-2.3 | **PLAN pre-plans the heavy end:** colour affordances for heavy-rain colours not yet observed, and a better way than the nearest legend colour to value the oranges and reds. **Required and tested before SHIP**, against the live capture of OW-12 when it comes; **if no severe day comes before SHIP**, MRMS ships with its heavy end marked unverified, the library warns whenever a colour takes the fallback, and the release notes say so. | D-19, D-37, D-44 | NO INSTRUMENT YET; the fallback-share test |
| L-2.4 | A table derived from the legend is never called exact. | L-2.2 | — |
| L-2.5 | **One additive seam for provider colour tables**, with IEM and MRMS as its first entries, so a third provider is added the same way. | D-35 (F-2's trigger) | NO INSTRUMENT YET |

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
| L-4.4 | **The README's promises match the code, and are held by the same check as the contract.** Its user-agent and `Purge` sentences were corrected under D-34; L-7.2 and L-9.3 change them again when they land. | D-34 (D-3) | NO INSTRUMENT YET (M5, extended) |

## L-5 — v0.1.0's close-out (supersedes the brief's L-5.1)

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-5.1 | ~~The nineteen-item triage is written.~~ **Superseded by D-18:** the triage never existed and is recorded as not recoverable. | D-18 | — |
| L-5.2 | One row in v0.1.0's record says its close-out did not run, with pointers to the evidence that does exist. | D-18 | Done: `release-checklist.md`'s status row (D-38) |
| L-5.3 | A gate test refuses a release tag while its checklist is unfinished. | D-18 | NO INSTRUMENT YET |
| L-5.4 | v0.2.0 is the library's first release: its REVIEW and close-out cover everything v0.1.0 shipped. The `v0.1.0` tag stays. | D-18 | — |
| L-5.5 | The release checklist records a pinned `govulncheck` at the tag commit, and **requires no reachable finding, or a ruled exception for each**. | D-30 (F9), D-40 (S-9) | The release checklist |

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
| L-7.1 | A host supplies its own transport, user-agent and timeout through `SetFetchOptions`, with no reflection; `Fetcher` and `fetch.Checked` are removed (D-62). | L-7.1, D-55, D-62 | NO INSTRUMENT YET |
| L-7.2 | Tiles can go out under the host's own user-agent. | L-7.2 | NO INSTRUMENT YET |
| L-7.3 | **Host fetches are bounded as far as a library that starts no goroutine can bound them** (D-55, a recorded deviation from D-40): **bytes** are bounded whatever the host code does — the library reads the body itself, through its limit; **a late answer is never used**; **time** is bounded while the host's code honours its context, and the contract says so. The host supplies a transport (`SetFetchOptions`), and the library keeps its client: request shape, redirect confinement, headers, the body limit. | D-30 (F3), D-40 (S-2), D-55 | NO INSTRUMENT YET; a test with a transport that streams without end, and one that ignores its context |
| L-7.4 | Setting a fetcher takes effect at once. | D-30 (F4) | NO INSTRUMENT YET |

## L-8 — Meaning without colour

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-8.1 | An alert area's label carries its **severity as a word**, and **a severity digit is repeated along its outline** — Extreme 4, Severe 3, Moderate 2, Minor 1, Unknown `?` — at every colour depth, never on the furniture rows, at least one on every area; the outline is solid; the hatch stays as a secondary cue; `Legend()` carries the digit key. | D-17, D-65 | NO INSTRUMENT YET; specimens 32, 33 |
| L-8.2 | The hatch strokes cannot rank five levels: Extreme and Severe share `╳`, **Minor and Unknown share `╱`**. | L-8.2 corrected, W1-C | — (context for L-8.1) |
| L-8.3 | **An image or field never erases furniture.** Outline, **hatch**, label, marker, `stale` word, the no-tiles notice, frame time, scale and credit survive an image or field **at every depth**. Today they are erased at NoColour **and Colours16**. Inside rain cells, where a shade glyph takes the whole cell and the hatch cannot draw, what carries severity is stated. | S29-2, D-14, red team A7, D-40 (H D-3, A F7) | NO INSTRUMENT YET; tests at NoColour and Colours16 |
| L-8.4 | ~~At 16 colours, rain classes stay distinguishable; every rain cell draws as `░`.~~ **Withdrawn (D-50):** a miscount; 16 colours draws the same three shades as no colour. | D-50 | — |
| L-8.9 | **The named place's marker and name are preserved**: the host's places outrank alert labels, basemap labels and furniture where they collide; where the name still cannot fit, a shorter form or the marker alone is drawn, and the frame reports it. | D-60; S29-3, S30-3, S31-2 | NO INSTRUMENT YET; specimens 29, 30 and 31 as tests |
| L-8.5 | A label that does not fit falls back to the severity word alone; if that does not fit, the description carries it and the frame reports the dropped label to the host. The outline's digit carries severity either way (D-65). | D-28 (A8), D-65 | NO INSTRUMENT YET; specimens 32, 33 |
| L-8.6 | ~~The five outline dashes differ from each other …~~ **Superseded by D-65**: the digit replaces the dash; an alert outline is solid and stays distinct from the line-overlay dash (5 on, 4 off). | D-28 (A8), D-65 | — |
| L-8.7 | The hatch draws at Colours16 as well as NoColour. | D-28 (A8) | NO INSTRUMENT YET |
| L-8.8 | Outlines hold 3:1 (WCAG 1.4.11) and labels 4.5:1 against the plain radar and, **on each ground where the blend is used**, the blended inside — at truecolor and 256 colours. **At 16 colours there is no ramp and no blend to check** (rain draws as shades), so severity there rests on the word and the dash. | D-28 (A9), D-40 (A F8, B N-3) | The checker, extended |

## L-9 — Cache retention and purge

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-9.1 | A host-settable **maximum age** for the disk cache. | L-9.1 | NO INSTRUMENT YET |
| L-9.2 | ~~A purge call.~~ **`Map.Purge` exists** (`tiles.go:145-165`); the gap is its **scope**: with a source named it empties only that source's tiles (with none named, the whole disk cache), and never memory, shared caches or decoded pictures. | L-9.2 corrected, W1-C, D-40 (H D-9) | NO INSTRUMENT YET |
| L-9.3 | **Purge empties everything**: every source, the memory caches, decoded pictures. **No write lands after Purge, after `CacheRoot` changes or turns the cache off, or after `Close`.** A replaced root is released, and the contract says Purge does not reach a root the map no longer holds. | D-30 (F5), D-40 (S-4) | NO INSTRUMENT YET; a test with a fetch in flight across each of the three |
| L-9.4 | **Age runs from when the tile was fetched: the tile file's time is its fetch time, set once, never touched by a read** (D-56). Maximum age (`MaxAge`) is enforced at `CacheRoot` and in every job; the cap evicts oldest-fetched first, tiles in view protected; a future-dated time counts as expired. | D-30 (F5), D-40 (S-3), D-56 | NO INSTRUMENT YET; a test that reads a tile hourly past its maximum age |
| L-9.5 | `cache-write-failed` is raised; only removals that succeeded are counted. | D-30 (F6) | NO INSTRUMENT YET |
| L-9.6 | The cache root is refused, or warned of, when others can read it, as it already is when they can write it; contract guidance: the root under the OS user cache directory, owner-only; purge is not secure erasure. | D-30 (F8), D-40 (S-6) | NO INSTRUMENT YET |

## L-10 — Tile-host confinement

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-10.1 | A source's tiles come only from that source's host; this is contract, stated and kept. | L-10.1 | `TestTileJSONAddressesObeyFetchRules`, `FuzzTileJSON`; a root-package test through `Map.Source` owed |
| L-10.2 | **One allow-list** for fetching and TileJSON, covering scheme, host and port under any fetcher; plain http refused unless the host allows it. | D-30 (F7) | NO INSTRUMENT YET; a root-package test through `Map.Source` with both fetchers |
| L-10.3 | **The private-address check holds through a proxy's exceptions**: whether a connection goes through the proxy is decided per connection, not per request, so a redirect to a host the proxy does not carry is still checked; the refused ranges include `0.0.0.0/8`, `64:ff9b::/96`, `2002::/16`, `198.18.0.0/15` and `240.0.0.0/4`. | D-40 (S-5) | NO INSTRUMENT YET |

## L-11 — Alerts over radar

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-11.1 | Inside an alert area, **the tint blends over the radar**: each radar cell keeps its class colour shifted toward the tint; outline and label stay on top. Replaces v0.1.0 FR-12 step 4 for images and fields. | D-14 | NO INSTRUMENT YET |
| L-11.2 | **On each ground where the blend is used**, it must not wash out: every tinted class stays at least 10 (v0.1.0 D-88's measure) from every other class, inside and outside the area. | D-14, D-40 (B N-3) | The ramp checker, extended |
| L-11.3 | **PLAN searches for a light-ground blend that passes**: the light radar ramp, the alert colours, or a per-class tint; L-11.4 says what ships if none does. | D-14, D-27 | The checker |
| L-11.5 | **The warning stays visible inside the blend**: each rain class inside an alert area differs from the same class outside by at least 5 (v0.1.0 D-88's measure), on each ground where the blend is used, tuned together with L-11.2. **Where both cannot hold, the outline, label and L-8.1's word and dash carry the warning there**, and the record says where. The 5 is a proposal PLAN measures. | D-45 (B N-4) | The checker, extended |
| L-11.4 | **If no blend passes on a light ground, the light ground uses 29b's order**: radar over the tint, the warning carried by its outline, label and L-8.1's severity word and dash. The dark ground keeps the blend. | D-27 (A10, B-6) | The checker decides which applies |

## L-12 — Memory

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-12.1 | A **total image budget per map**, settable by the host; a hand-in over it is refused with an error that says so. The **per-image** cap (250,000 pixels, C-3) is unchanged in v0.2.0: a host can lower it, not raise it. | D-20, D-40 (B N-9) | NO INSTRUMENT YET |
| L-12.2 | The default fits v0.1.0 NFR-3 (4 MB live): about **3 MiB** of images — one 12-frame loop at 596×304 (about 2.3 MiB with its PNG bytes), or about five at dot resolution (298×152, about 0.59 MiB each with its PNG bytes, counted as L-12.4 counts). One run on one machine; see wave 2's Limits. | D-20, W2 M-B, D-40 (H D-12) | M4 |
| L-12.3 | Guidance steers hosts to frames at the dot grid of the view (298×152 at 149×38). A host that raises the budget owns the memory it asks for, and the contract says so. | D-20 | — |
| L-12.5 | A frame advance costs **≤ 15 ms** at 149×38 on the reference machine, the blend included (D-61). | D-37 (P-6), D-61 | `BenchmarkFrameAdvance` |
| L-12.6 | `ReadBack` reads through its size limit, as `Load` does; fields and the shared classified set count inside the budget; the classification cache key hashes a table value's exact bits, so two tables never share a key. | D-40 (S-7) | NO INSTRUMENT YET |
| L-12.4 | The budget counts retained PNG bytes as well as pixels; a frame's file size is capped relative to its pixel count. | D-30 (F2) | NO INSTRUMENT YET |

## L-13 — v0.1.0 defects fixed in this release

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-13.1 | The library identifies itself with its real version, not `0.1.0-dev` (`fetch.go:24`). | D-18, W1-B | NO INSTRUMENT YET |
| L-13.2 | `fetch.Checked` is wired, or removed. | W1-B | NO INSTRUMENT YET |
| L-13.3 | The gate's NOT RUN line is current. | W1-B | Done: it names the last tag (D-38), held by `TestNotRunNamesTheLastTag` |
| L-13.4 | The drifted cache doc comments and `ReduceMotion`'s overstated comment are corrected. | W1-B, W1-A | `make lint`-class review |
| L-13.9 | **Severity and valid time are data on each alert**, not read from its colour role. | D-43 (A F5) | NO INSTRUMENT YET |
| L-13.10 | **The contract states that a host words an image answer's class from `Legend()`**, with the example showing how; **a legend entry says when its table is approximate** (D-19). | D-47 (A F12) | NO INSTRUMENT YET |
| L-13.8 | A test lists every environment read the library makes (`os.Getenv` today only in `look.go`), so a new one cannot arrive unnoticed. | D-40 (S-10) | NO INSTRUMENT YET |
| L-13.5 | **A list of the alerts shown, needing no place**: each alert's name, severity word and valid time or stale mark — in the new `Report` (D-57). | D-29 (A11), D-43, D-57 | NO INSTRUMENT YET |
| L-13.6 | **For a discrete place the host names, each alert is answered on its own** — inside, outside or **nearby** — with its name and severity word, instead of one answer merged across an overlay. **"Nearby" is new in v0.2.0** (v0.1.0 has only inside and outside): within a distance the host sets, default proposed 10 km, PLAN confirms. The description never says "you"; the library returns data, so this binds the contract's, the example's and watchpost's wording. | D-29, D-43 (A F4) | NO INSTRUMENT YET |
| L-13.7 | *Deferred, "necessity unproven yet" (D-29):* intensity in words for image answers (a host words it from `Legend()`, L-13.10); a general summary of what is in view; new place-to-weather statements beyond inside / outside / nearby — **except L-1.12's motion statement, which D-29 keeps and D-42 widens to views with no place, and the answers v0.1.0 already gives** (`Heavier`, `HeavierAt`). | D-29, D-40, D-42, D-47 | — |

## Metrics of success

The brief's M1–M6 (D-6), with round 1's changes. **This table is the normative one.**

| # | Metric | Type | Definition |
|---|---|---|---|
| M1 | Motion seen | Primary | From a loop alone, a reader states a precipitation cell's direction of motion (scripted UAT over loops recorded from real radar); **and a non-visual arm (D-24): from the description alone**, scored by a grader PLAN defines |
| M2 | Loop honesty | Primary | Zero frames drawn under the wrong valid time; every gap stated; every frame older than the host's stated cadence marked stale |
| M3 | Bound holds | Primary | With a host bound set, zero frames outside it across resize, fit, pan, zoom and fall-back |
| M4 | Embed cost with a loop | Primary | A 12-frame loop at dot resolution adds **≤ 1 MiB** of heap over an empty map, within ±5 % over an hour (D-61; measured +0.59 MiB) |
| M5 | Contract truth | Secondary | Zero names in `contract.md` — and promises in the README (D-34) — missing from the code, and none the other way; a test |
| M6 | Gate trust | Secondary | Zero unattributable gate failures across a stated number of consecutive full runs, counted from `06_docs/gate-runs.md` (D-31, D-40) |

## Non-functional

| # | Requirement | Source |
|---|---|---|
| NFR-1 | **For v0.2.0, a new feature's better long-term API may break v0.1.0's shape** (D-58): v0.1.0 has no consumer but watchpost, and each break is listed in the contract's changelog. **An earlier decision is revisited only for significant upside**, weighing simplification and developer ergonomics highest, within the shared structure (plugins, discrete domains). Rows that change behaviour a v0.1.0 host can observe: L-7.3 and L-7.4, L-9.3, L-9.4, L-11.1, and the `Report` call beside `Describe` (D-57). | C-7, D-40, D-58 |
| NFR-2 | Memory within v0.1.0 NFR-3 at the default budget, flat over an hour with a 12-frame loop. | M4, D-20 |
| NFR-3 | Motion: no flash above v0.1.0 D-56's ceiling (2.5 a second); `ReduceMotion` stops all animation. | v0.1.0 D-56, NFR-21 |
| NFR-4 | The gate is green before every commit that touches anything but Markdown; a Markdown-only change takes the docs lane, run before committing. Every run is logged in `06_docs/gate-runs.md`. | D-15, D-31 |

## Risk register

Likelihood and severity are **H / M / L**. **Evidence** says what each rating rests on; **Status** says
where it stands.

| # | Risk | L | S | Evidence | Mitigation | Status |
|---|---|---|---|---|---|---|
| RK-1 | **A loop that freezes or re-decodes while every unit test passes** — the three changes (frames, renderer, store) land apart | M | H | W1-A: `internal/render/frame.go:285-302` reuses the last frame; `internal/overlay/store.go:453` drops rasters on re-Set; this codebase has shipped "correct parts, wrongly connected" before (watchpost RK-1) | L-1.6, L-1.7 each held by a test that sees a frame advance | Open |
| RK-2 | **The MRMS heavy end is wrong on a severe-weather day** — the palette was seen on a quiet day (≤ 48 dBZ) | H | H | W2 M-A: palette seen on one quiet afternoon (≤ 48 dBZ); heavy-end colours 3–30 from any legend colour | L-2.3; the fallback-share test; the triggered capture (OW-12); the archive specimen (OW-11) | Open; depends on live weather (RK-11) |
| RK-3 | **No blend passes on a light ground** (every strength tried failed) | H | L | S29-7: 0 of 15 light-ground configurations (3 strengths × 5 severities) clear 10 | L-11.3; the named fallback L-11.4 (D-27) | Mitigated by a named fallback |
| RK-4 | **Watchpost's radar slips** because this release does | H | M | D-36 (no split) and D-37 (no date) accept it; watchpost's own RK-4 | Accepted; watchpost's fallback is to ship on v0.1.0 **without radar, the view bound (HR-3) or the colour-independent pattern (HR-7)**; v0.1.0 is not retracted (D-48) | Accepted (D-36, D-37, D-48) |
| RK-5 | **A provider changes its palette or endpoint** (IEM has no SLA; MRMS no published table) | M | M | W1 watchpost: IEM has no SLA, MRMS no published table; the unmatched count sees only colours more than 10 from any table colour | The unmatched-colour count is the detector (L-2.2). Residual: a shift under 10 is matched to the wrong value silently; counting fallback matches as drift is owed (OW-10) | Open |
| RK-6 | **The contract stays wrong** — a names-only test passes with false behavioural claims | H | M | W1-B: five false behavioural claims pass `TestPublicSurfaceSnapshot`, which records names only | L-4.2, L-4.3 | Open |
| RK-7 | **The gate hangs on a fuzz stall** — the count budget has no wall clock | L | M | D-3, L-6.4: the engine stalled before; D-31/D-40 limiter tested, mutation-proven | Each fuzz leg runs under a wall-clock limit, TERM then KILL, and fails as TIME LIMIT (D-31, D-40) | Open |
| RK-8 | **A listener who cannot watch the animation is excluded** from the release's main point | M | H | A1, B-4; D-24, D-42 | L-1.12, with or without a named place; M1's non-visual arm. Residual: two-frame motion can mislead when cells grow or decay, so it is worded as observation | Mitigated |
| RK-9 | **Frames change after hand-in** — the library borrows host image bytes | L | H | InfoSec F1, S-1: `store.go:437` keeps host slices; decode allocates before its check | L-1.14 (D-30) | Open |
| RK-10 | **Evidence cannot be re-run** — wave 2 and specimen 29's blends came from throwaway programs | H | L | D-32 accepts that filed programs go stale unnoticed; BUILD rewrites the loop API they call | Filed with inputs and output in `02-analysis/programs/` (D-32); they may go stale as the API changes | Accepted (D-32) |

| RK-11 | **The heavy-end evidence depends on the weather** — MRMS keeps about two hours of history and publishes no table, so its heavy colours can be seen only on a live severe day | M | H | D-44; W2 M-A | The triggered capture (OW-12); L-2.3's stated fallback if no day comes before SHIP | Open |

## Owed

| # | What | Due | Source |
|---|---|---|---|
| OW-1 | ~~The row in v0.1.0's record saying its close-out did not run~~ **Done (D-38)**: `go-tuimaps/07-readiness/release-checklist.md` | — | D-18 |
| OW-2 | The D-17 specimen: severity word and five dashes, at 69×12 and 149×38, NoColour and Colours16, **drawn over specimen 29's radar**, with the frame time and the `stale` word on the same bottom row, so the crowding at 69×12 is seen | before PLAN commits to L-8.1 | D-17, D-40 (A F7, B) **Done: specimens 32 and 33; ruled D-64, D-65 (a digit on the outline).** |
| OW-3 | ~~Diagnose S29-3~~ **Folded into L-8.9 (D-60)** | — | D-60 |
| OW-4 | The 40-second `FuzzAgree` freeze | PLAN | L-6.4 |
| OW-5 | ~~Correct wave 2's "71–80 %" and D-19's copy of it~~ **Done 2026-09-23**: 80.0–80.6 %, and wave 2's finding 2 restated | — | round 1 verification |
| OW-6 | F-2 (a pluggable architecture): its trigger fired; the narrow seam is L-2.5 (D-35); the broad restructure stays with the quality pass | quality pass | D-35 |
| OW-9 | ~~Diagnose L-8.4~~ **Closed (D-50): no defect** — the round 1 count was wrong | — | D-50 |
| OW-11 | ~~A specimen of a past outbreak~~ **Done at PLAN entry: specimen 30** (2011-04-27, west-central Alabama, 20 warnings over IEM archive radar). Findings S30-1..S30-4 | — | D-44 |
| OW-12 | **A triggered MRMS capture**: when a Storm Prediction Center Moderate or High risk, or a tornado watch, is issued over US radar coverage, capture MRMS frames across the whole two-hour window and extend the palette | when triggered, before SHIP | D-44 |
| OW-10 | RK-5's detector: count tolerance-fallback matches, not only unmatched pixels, so a palette shift under 10 is seen | PLAN | D-40 (B N-8) |
| OW-8 | ~~The gate never runs `gofmt`~~ **Closed by D-46**: the gate has a formatting leg | — | D-46 |
