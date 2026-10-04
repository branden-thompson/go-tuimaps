---
title: "go-tuiMaps v0.2.0 — Radar loops — REQUIREMENTS (normative)"
date: 2026-09-23
phase: BUILD (exit in progress)
sev: SEV-0
authority: HUM LEAD
status: "NORMATIVE (D-23). On any conflict with the brief or another document, this file wins. Every change is a row in 02-analysis/rulings.md."
---

# v0.2.0 requirements

**How to read this.** The brief (`project-brief.md`, and issue #2) says what the release is and
why. **This file is the requirement set the release is built and reviewed against** (D-23). Requirements keep the brief's
L-numbers and extend them. Each row gives its source and what holds it: a test, a gate check, a
checklist row or a document, and where a part is still owed, the owed item that closes it.

Sources: **L-** the brief; **D-** this release's rulings; **S29-** specimen 29
(`../go-tuimaps/02-analysis/specimens/README.md`); **W1-/W2-** waves 1 and 2; **v0.1.0 FR-/NFR-** the
first release's record.

## What changes on screen

For designers and PMs before engineers, the five visible decisions:

- **Radar moves** (L-1): a loop of frames, with playback off or on at a step the host sets (D-76).
- **An alert over radar lets the radar show through** (D-14, specimen 29d–f): the warning's tint is
  blended into the radar colours, and its outline and label stay on top.
- **Severity reads without colour** (D-17, D-65): a severity word in the label, and a severity digit
  repeated along each alert's outline (Extreme 4, Severe 3, Moderate 2, Minor 1, Unknown `?`), at
  every colour depth (specimens 32, 33).
- **Playback is off until the host turns it on** (D-26), and **the shown frame's time, or "gap", is
  written on the map** (D-25).
- **On a light ground the blend may not be used at all**: if no blend passes there, the warning is
  drawn by outline, label and digit over the radar (D-27, D-65).

## L-1 — Loops

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-1.1 | An image overlay carries several frames, each with its own valid time; zero frames means the single image of v0.1.0 (additive). | L-1.1, C-1 | `TestAnImageCarriesFrames` |
| L-1.2 | A frame may be a **gap**: a missing frame is stated as a gap, never replaced by a neighbour. | L-1.2, M2 | `TestTheFrameTimeIsOnTheMap` (a gap holds the last real frame, its time reading "gap"); `TestEveryFrameIsCheckedAtHandIn` (a gap is explicit) |
| L-1.3 | **The newest non-gap frame's valid time drives the `stale` word and the description**, so an old loop never looks current and neither toggles as the loop plays. Stepping changes the picture and the frame-time text (L-1.10a), not the description. | L-1.3, M2, D-39 | `TestTheNewestObservedFrameDrivesStale`; `TestAFrameAdvanceIsATickNotAChange` (the description stays cached as the loop steps) |
| L-1.4 | The library never fetches radar; the host hands in the data, and the library turns it into frames. | L-1.4, watchpost 0.18.0 D-35 | `TestTheImagePathNeverFetches` |
| L-1.5 | Playback is controllable, off or on at a frame step from 200 ms to 1000 ms, 500 ms by default (D-76, replacing off / slow / normal), for every looped overlay; the control is the library's, exposed to the host and through it to the listener. | L-1.5 (HR-9) | `TestPlaybackIsTheMapsAndPersists` |
| L-1.6 | **A frame advance must be seen by the renderer.** The renderer reuses its last frame only while the overlays' version is unchanged, and the map raises that version whenever the moment shown changes, so a loop never freezes while reporting success. | W1-A | `TestAStepIsRedrawnNotReused`, `TestEveryAdvanceIsDrawn`, `TestThePathThroughThePublicMap` |
| L-1.7 | **A refresh does not re-decode the whole loop.** A frame's key is its valid time and a hash of its bytes, and a re-`Set` keeps every decoded frame whose key it already holds. | W1-A | `TestARefreshDecodesOnlyWhatIsNew` |
| L-1.8 | `NextCall` reports the next frame change; the contract's claim that a frame-advance source exists becomes true. | L-1.1, C-2 | `TestNextCallIncludesTheNextAdvance` |
| L-1.16 | **`Frame` carries the `Changed` and `FrameTicks` values it was drawn at**, so a host can prove a stored frame is current (watchpost D-41). | D-59 | `TestTheFrameCarriesItsCounters` |
| L-1.9 | Every frame is validated when handed in, the loop has a total, frame times are in order, and a gap is explicit. | W1-A | `TestEveryFrameIsCheckedAtHandIn` |
| L-1.14 | **Every host slice an image carries — PNG and table, for the single image and every frame alike — is copied at hand-in, and the copy is what is validated.** A later decode reads the size from the copy's header before it allocates, and re-checks the cap and the budget. | D-30 (F1), D-40 (S-1) | `TestTheStoreKeepsItsOwnCopy` (the host's slices mutated after hand-in), `TestDecodeReadsTheHeaderBeforeItAllocates` |
| L-1.15 | A hard maximum frame count, gaps included; the playback interval is library-owned and floored, never derived from host-supplied times. | D-30 (F2) | `TestEveryFrameIsCheckedAtHandIn` (frame 73 refused); `TestPlaybackIsTheMapsAndPersists` (a step outside 200–1000 ms refused) |
| L-1.10a | The shown frame's time, or "gap", is text on the map at every depth, beside the `stale` word. | D-25 (A2) | `TestTheFrameTimeIsOnTheMap`; `TestAnImageNeverErasesFurniture` (the frame time survives an image at NoColour and Colours16) |
| L-1.10b | A host moves the loop with `Play()` (from the oldest held frame through "right now" and on through forecast frames), `Stop()` (holds the shown frame), `Reset()` (back to "right now", the newest observed frame) and `Step(by int)` (by frames, stopping playback). Position is a valid time, never an index, and there is no seek (D-67). The library's animation clock drives the loop. | D-25 (A3), D-67 | `TestTheListenersControls` |
| L-1.10c | `ReduceMotion(true)` forces playback off; turning it off restores the host's setting. | D-25 (A4) | `TestReduceMotionForcesOffAndRestores` |
| L-1.10d | A state read: the playback in effect, frame index, count, the frame's time, gap, and the span the loop covers (oldest to newest). **"Playing" only while the shown frame is actually advancing**, so a host that freezes the animation clock (`Animate`) reads the loop as stopped. When playback is off for more than one reason the state names the first of: reduce motion, the host's choice, the default. A loop's name is the host's to give, and the contract says so. | D-25 (A5), D-40 (A F9, F13) | `TestTheStateRead`; the loop's name is the host's: contract section 11, held by `TestTheContractSaysWhatAHostNeeds` |
| L-1.10e | A frame advance is signalled apart from a data change and does not change the description's cache key. | D-25 (A6) | `TestAFrameAdvanceIsATickNotAChange` |
| L-1.10f | Off shows the newest non-gap frame with its age. | D-25 (A15) | `TestTheFrameTimeIsOnTheMap` |
| L-1.10g | A library-owned numeric ceiling on changes a second, **one ceiling per map**, counting every moment anything on it changes — frames of every loop, marker blink, furniture — toward WCAG 2.3.1, bounded by v0.1.0 D-56's 2.5 a second; a gap holds the last real frame with its time reading "gap", never an empty frame. When two loops play, the frame time shown is the newest loop's, and the state read gives each. **D-76:** a step from 400 ms up keeps within it, the blink moving onto the loop's grid while a loop plays; a step below 400 ms lifts it, as the host's explicit choice, and the state read says so. | D-25 (A16, P-5), D-40 (A F6) | `TestOneChangeCeilingPerMap`, `TestEveryAdvanceIsDrawn`, `TestTheFrameTimeIsOnTheMap` |
| L-1.11 | **Playback defaults to off** until the host chooses; the state read says "off (default)". | D-26 (A14) | `TestPlaybackIsTheMapsAndPersists` |
| L-1.13 | **One standard playback API** a host wires straight to its own controls and Settings: `SetPlayback` (off or on) and `SetPlaybackStep` (D-76), `Play`, `Stop`, `Reset` and `Step(by)` (D-67), the state read (`Loop`) and the change signals (`Changed`, `NextCall`), one playback per map across every looped overlay. | D-26, D-67, D-76 | `Example_playback`, `TestThePlaybackHostKeepsNoState` (the example host keeps no playback state of its own); watchpost's controls and Settings row |
| L-1.12 | **Storm motion without the animation (D-24, D-42).** The description reports observed motion from the loop's frames: where the heavier rain was at the oldest usable frame and at the newest (distance, direction, time), whether it came closer, moved away or held, and the span the loop covers — **relative to a named place where there is one, and relative to the view (its centre and edges) where there is none**, so a listener in a view with no names still hears it. **"Heavier" is one intensity threshold, set by the colour table and held for the whole loop.** The answer is **data**: two timed positions and a relation, with no field that can express a forecast (the library words nothing; a host does). Worded as observation, never forecast, and never "you" (D-29). | D-24, D-29, D-42; red team A1, B-4, A F2 | `TestMotionTowardAPlace`, `TestEachPlaceGetsItsNearestCell`, `TestMotionWithNoPlaceIsRelativeToTheView`, `TestAGapAtTheOldestFrame`, `TestAForecastIsNeverASighting`, `TestAHostTypesHeavierRainIsItsTopThird`; M1's non-visual arm, scored from the description alone, is owed (L10.9, the release checklist) |

## L-2 — The MRMS table

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-2.1 | A table for NOAA/NCEP MRMS `conus_bref_qcd` ships, built from **the colours MRMS actually uses** (111 seen across 16 distinct times on one quiet afternoon, rain up to about 48 dBZ; see wave 2's Limits), each valued from its position on the legend, labelled **approximate**. | D-19, W2 M-A, D-40 (H D-13) | `TestTheMRMSTableIsTheObservedPalette` |
| L-2.2 | Matching is exact first; the tolerance fallback is kept for colours not yet seen, and such pixels are counted and reported. | D-19 | The unmatched-colour warning (`kinds.go:50`); **pixels matched only within the tolerance of a table that is not approximate are counted and warned of as `near-image-colours` (OW-10, D-104)** |
| L-2.3 | **PLAN pre-plans the heavy end:** colour affordances for heavy-rain colours not yet observed, and a better way than the nearest legend colour to value the oranges and reds. **Required and tested before SHIP**, against the live capture of OW-12 when it comes; **if no severe day comes before SHIP**, MRMS ships with its heavy end marked unverified, the library warns whenever a colour takes the fallback, and the release notes say so. | D-19, D-37, D-44 | `TestTheHeavyEndIsValuedAlongTheGradient`, `TestAFallbackIsCountedAndWarned`, `TestTheFallbackShare`, `TestTheLegendSaysWhenATableIsApproximate` (MRMS's heavy end unverified); the release checklist row 0.10 (the release-note line); the live oracle is owed (OW-12) |
| L-2.4 | A table derived from the legend is never called exact. | L-2.2 | — |
| L-2.5 | **One additive seam for provider colour tables**, with IEM and MRMS as its first entries, so a third provider is added the same way. | D-35 (F-2's trigger) | `TestAProviderIsOneFileOfItsOwn`, `TestAnImageNamingAProviderDrawsWithItsTable`, `TestTheIEMTableIsThePublishedOne` |

## L-3 — The view bound

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-3.1 | A host sets a minimum zoom and a bounding box once, and the view stays inside them on every path: resize, fit, pan, zoom and the fall-back to the whole world. | L-3.1, W1-C | M3: `TestTheBoundIsHeldOnEveryPath`, `TestABoundIsRefusedOrHeld` |
| L-3.2 | A host that sets nothing gets v0.1.0's behaviour. | L-3.2 | `TestNoBoundIsV010` |

## L-4 — The contract matches the code

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-4.1 | `contract.md` names nothing the public surface lacks, and the surface nothing the contract lacks. | L-4.1 | M5: `TestTheContractNamesOnlyWhatExists`, `TestTheContractNamesEverythingExported` |
| L-4.2 | The five false behavioural claims are corrected: `Work`/`Settle` released ids; a panicking `Render`'s "failed" frame; `NextCall`'s frame-advance source; `CacheRoot`'s signature; FR-22b's "the library checks what comes back". | W1-B | `TestSetAndRemoveReportTheRelease` (released ids); `TestAPanickingRenderReturnsAnEmptyFrameAndAnInternalError`; `TestNextCallIncludesTheNextAdvance`; `TestTheSurfaceRecordsSignatures` (`CacheRoot`'s signature); `TestAHostTransport` (what comes back is read through the library's limit); and D-66's sixth, `TestChangedCountsInputsAndWhatWorkLands` |
| L-4.3 | The surface test records signatures, not names only. | W1-B | `TestTheSurfaceRecordsSignatures`, `TestPublicSurfaceSnapshot` |
| L-4.4 | **The README's promises match the code, and are held by the same check as the contract.** Its user-agent and `Purge` sentences state what the code does (D-34, L-7.2, L-9.3). | D-34 (D-3) | M5, extended: `TestReadmeWhatItSendsIsHeldByBehaviour`, `TestReadmePurgeEmptiesEverySource` |

## L-5 — v0.1.0's close-out (supersedes the brief's L-5.1)

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-5.1 | ~~The nineteen-item triage is written.~~ **Superseded by D-18:** the triage never existed and is recorded as not recoverable. | D-18 | — |
| L-5.2 | One row in v0.1.0's record says its close-out did not run, with pointers to the evidence that does exist. | D-18 | Done: `release-checklist.md`'s status row (D-38) |
| L-5.3 | A gate test refuses a release tag while its checklist is unfinished. | D-18 | `scripts/gate --release`, held by `release_gate_test.go` (L10.3) |
| L-5.4 | v0.2.0 is the library's first release: its REVIEW and close-out cover everything v0.1.0 shipped. The `v0.1.0` tag stays. | D-18 | — |
| L-5.5 | The release checklist records a pinned `govulncheck` at the tag commit, and **requires no reachable finding, or a ruled exception for each**. | D-30 (F9), D-40 (S-9) | The release checklist (row 2.6) and `scripts/gate --release`'s pinned scan (L10.4) |

## L-6 — Found while preparing the brief

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-6.1 | Hosted CI reproduces the gate, including the second-architecture leg a Linux runner cannot emulate as this machine does. | L-6.1, W1-B | `.github/workflows/gate.yml` (L10.5, D-105): each architecture native, the workflow failing unless both ran green |
| L-6.2 | (Became L-7.) | L-6.2 | — |
| L-6.3 | The default memory tile cache is at least documented against one large view (909 KB needed, 500 KB default). | L-6.3 | `TestTheDefaultTileCacheAgainstALargeView` (contract section 11) |
| L-6.4 | ~~The fuzz legs fail by accident; every commit is blocked.~~ **Fixed by construction** (D-9, D-10): run-count budgets. **Not reproduced** — the cause is probable, not proven. The 40-second freeze seen once in `FuzzAgree` is the engine minimising a new input (OW-4, done: `02-analysis/gate-fuzz-deadline.md`). | D-9, D-10 | M6 |

## L-7 — A host can write a fetcher

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-7.1 | A host supplies its own transport, user-agent and timeout through `SetFetchOptions`, with no reflection; `Fetcher` and `fetch.Checked` are removed (D-62). | L-7.1, D-55, D-62 | `TestFetchOptionsTakeEffectAtOnce`, `TestAHostTransport`; the removal: `TestPublicSurfaceSnapshot`, `TestTheChangelogListsEveryBreak` |
| L-7.2 | Tiles can go out under the host's own user-agent. | L-7.2 | `TestFetchOptionsTakeEffectAtOnce` (the host's name reaches the user-agent) |
| L-7.3 | **Host fetches are bounded as far as a library that starts no goroutine can bound them** (D-55, a recorded deviation from D-40): **bytes** are bounded whatever the host code does — the library reads the body itself, through its limit; **a late answer is never used**; **time** is bounded while the host's code honours its context, and the contract says so. The host supplies a transport (`SetFetchOptions`), and the library keeps its client: request shape, redirect confinement, headers, the body limit. | D-30 (F3), D-40 (S-2), D-55 | `TestAHostTransport` (a body that never ends is cut at the limit; a transport that ignores its context; a late answer dropped), `TestALateAnswerIsNeverUsed`; the contract sentence, `TestTheContractSaysWhatAHostNeeds` |
| L-7.4 | Setting a fetcher takes effect at once. | D-30 (F4) | `TestFetchOptionsTakeEffectAtOnce` |

## L-8 — Meaning without colour

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-8.1 | An alert area's label carries its **severity as a word**, and **a severity digit is repeated along its outline** — Extreme 4, Severe 3, Moderate 2, Minor 1, Unknown `?` — at every colour depth, never on the furniture rows, at least one on every area; the outline is solid; the hatch stays as a secondary cue; `Legend()` carries the digit key. | D-17, D-65 | `TestTheLabelCarriesTheSeverityWord`, `TestTheOutlineCarriesTheSeverityDigit`, `TestTheLegendCarriesTheDigitKeyAndTheBlend`; specimens 32, 33 |
| L-8.2 | The hatch strokes cannot rank five levels: Extreme and Severe share `╳`, **Minor and Unknown share `╱`**. | L-8.2 corrected, W1-C | — (context for L-8.1) |
| L-8.3 | **An image or field never erases furniture.** Outline, **hatch**, label, marker, `stale` word, the no-tiles notice, frame time, scale and credit survive an image or field **at every depth**, NoColour **and Colours16** included: an image's shade takes only a cell nothing else claimed (D-77). Inside rain cells, where a shade glyph takes the whole cell and the hatch cannot draw, what carries severity is stated. | S29-2, D-14, red team A7, D-40 (H D-3, A F7) | `TestAnImageNeverErasesFurniture` (NoColour and Colours16); what carries severity inside rain: contract section 11, `TestTheContractSaysWhatAHostNeeds` |
| L-8.4 | ~~At 16 colours, rain classes stay distinguishable; every rain cell draws as `░`.~~ **Withdrawn (D-50):** a miscount; 16 colours draws the same three shades as no colour. | D-50 | — |
| L-8.9 | **The named place's marker and name are preserved**: the host's places outrank alert labels, basemap labels and furniture where they collide; where the name still cannot fit, a shorter form or the marker alone is drawn, and the frame reports it. | D-60; S29-3, S30-3, S31-2 | `TestTheNamedPlaceOutranksAlertLabelsAndNames`, `TestAPlaceNameNeverCoversFurniture`, `TestAOneWordNameThatDoesNotFitIsReported` |
| L-8.5 | A label that does not fit falls back to the severity word alone; if that does not fit, the description carries it and the frame reports the dropped label to the host. The outline's digit carries severity either way (D-65). | D-28 (A8), D-65 | `TestALabelThatDoesNotFitFallsBackAndIsReported`; specimens 32, 33 |
| L-8.6 | ~~The five outline dashes differ from each other …~~ **Superseded by D-65**: the digit replaces the dash; an alert outline is solid and stays distinct from the line-overlay dash (5 on, 4 off). | D-28 (A8), D-65 | — |
| L-8.7 | The hatch draws at Colours16 as well as NoColour. | D-28 (A8) | `TestALabelThatDoesNotFitFallsBackAndIsReported` |
| L-8.8 | Outlines hold 3:1 (WCAG 1.4.11) and labels 4.5:1 against the plain radar and, **on each ground where the blend is used**, the blended inside — at truecolor and 256 colours. **At 16 colours there is no ramp and no blend to check** (rain draws as shades), so severity there rests on the word and the digit (D-65). | D-28 (A9), D-40 (A F8, B N-3) | The checker, extended |

## L-9 — Cache retention and purge

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-9.1 | A host-settable **maximum age** for the disk cache. | L-9.1 | `TestCacheMaxAge` |
| L-9.2 | ~~A purge call.~~ **`Map.Purge` exists** (`tiles.go:145-165`); the gap is its **scope**: with a source named it empties only that source's tiles (with none named, the whole disk cache), and never memory, shared caches or decoded pictures. | L-9.2 corrected, W1-C, D-40 (H D-9) | `TestPurgeEmptiesEverything`, `TestReadmePurgeEmptiesEverySource` |
| L-9.3 | **Purge empties everything**: every source, the memory caches, decoded pictures. **No write lands after Purge, after `CacheRoot` changes or turns the cache off, or after `Close`.** A replaced root is released, and the contract says Purge does not reach a root the map no longer holds. | D-30 (F5), D-40 (S-4) | `TestNoWriteLandsAfterPurgeRootChangeOrClose` (a fetch in flight across each of the three), `TestPurgeEmptiesEverything`, `TestAReplacedRootIsLetGo`, `TestPurgeDoesNotReachAReleasedRoot`, `TestAStoreFromAnEarlierGenerationWritesNothing` |
| L-9.4 | **Age runs from when the tile was fetched: the tile file's time is its fetch time, set once, never touched by a read** (D-56). Maximum age (`MaxAge`) is enforced at `CacheRoot` and in every job; the cap evicts oldest-fetched first, tiles in view protected; a future-dated time counts as expired. | D-30 (F5), D-40 (S-3), D-56 | `TestReadsNeverTouchTheFileTime`, `TestATilePastItsAgeIsFetchedAgain` (a future-dated tile expired too), `TestEveryJobDropsWhatHasAged`, `TestEvictOldestFetchedFirst`, `TestTheViewsTilesSurviveThePipelinesPruning`, `TestCacheMaxAge` |
| L-9.5 | `cache-write-failed` is raised; only removals that succeeded are counted. | D-30 (F6) | `TestAFailedWriteIsWarnedOf`, `TestPurgeCountsWhatItRemoved`, `TestPurgeCountsWhatItCouldNotRemove` |
| L-9.6 | The cache root is refused, or warned of, when others can read it, as it already is when they can write it; contract guidance: the root under the OS user cache directory, owner-only; purge is not secure erasure. | D-30 (F8), D-40 (S-6) | `TestARootOthersCanReadIsWarnedOf`; the guidance, contract section 11, `TestTheContractSaysWhatAHostNeeds` |

## L-10 — Tile-host confinement

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-10.1 | A source's tiles come only from that source's host; this is contract, stated and kept. | L-10.1 | `TestTileJSONAddressesObeyFetchRules`, `FuzzTileJSON`; through `Map.Source`, `TestOneAllowListThroughSource` |
| L-10.2 | **One allow-list** for fetching and TileJSON, covering scheme, host and port under any fetcher; plain http refused unless the host allows it. | D-30 (F7) | `TestOneAllowListThroughSource` (through `Map.Source`, with the library's transport and a host's), `TestPlainHTTPRefused` |
| L-10.3 | **The private-address check holds through a proxy's exceptions**: whether a connection goes through the proxy is decided per connection, not per request, so a redirect to a host the proxy does not carry is still checked; the refused ranges include `0.0.0.0/8`, `64:ff9b::/96`, `2002::/16`, `198.18.0.0/15` and `240.0.0.0/4`, and - after the BUILD-exit red team (InfoSec) - local-use NAT64 `64:ff9b:1::/48`, site-local `fec0::/10`, Teredo `2001::/32`, and the documentation and protocol ranges `2001:db8::/32`, `192.0.0.0/24`, `192.0.2.0/24`, `198.51.100.0/24` and `203.0.113.0/24`. | D-40 (S-5) | `TestTheProxyDecisionIsPerConnection`, `TestReservedRangesRefused`, `TestTheLesserReservedRangesAreRefused`, `TestCheckedDialerRefusesAPrivateAddress` |

## L-11 — Alerts over radar

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-11.1 | Inside an alert area, **the tint blends over the radar**: each radar cell keeps its class colour shifted toward the tint; outline and label stay on top. Replaces v0.1.0 FR-12 step 4 for images and fields. | D-14 | `TestTheTintBlendsOverTheRadar`, `TestBlendMixesInLinearLight` |
| L-11.2 | **On each ground where the blend is used**, it must not wash out: every tinted class stays at least 10 (v0.1.0 D-88's measure) from every other class, inside and outside the area. | D-14, D-40 (B N-3) | The ramp checker, extended |
| L-11.3 | **PLAN searches for a light-ground blend that passes**: the light radar ramp, the alert colours, or a per-class tint; L-11.4 says what ships if none does. | D-14, D-27 | The checker |
| L-11.5 | **The warning stays visible inside the blend**: each rain class inside an alert area differs from the same class outside by at least 5 (v0.1.0 D-88's measure), on each ground where the blend is used, tuned together with L-11.2. **Where both cannot hold, the outline, label and L-8.1's word and digit carry the warning there**, and the record says where. The 5 is a proposal PLAN measures. | D-45 (B N-4) | The checker, extended |
| L-11.4 | **If no blend passes on a light ground, the light ground uses 29b's order**: radar over the tint, the warning carried by its outline, label and L-8.1's severity word and digit. The dark ground keeps the blend. | D-27 (A10, B-6) | The checker decides which applies |

## L-12 — Memory

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-12.1 | A **total image budget per map**, settable by the host; a hand-in over it is refused with an error that says so. The **per-image** cap (250,000 pixels, C-3) is unchanged in v0.2.0: a host can lower it, not raise it. | D-20, D-40 (B N-9) | `TestTheImageBudgetIsTheHosts`, `TestTheBudgetRefusesAndSaysByHowMuch` |
| L-12.2 | The default is **6 MiB** of images (D-68, raising D-61's 3 MiB): two hours of loop, 24 frames at 5-minute steps of region images up to about 200,000 pixels, counted as L-12.4 counts. It is above v0.1.0 NFR-3's 4 MB live; D-68 revisits memory once the feature works (OW-14). | D-20, D-68, W2 M-B, D-40 (H D-12) | M4: `TestALoopsHeapCostAndHold`; `TestTwentyFourRegionFramesCost` (24 region frames, about 5.8 MiB); `TestTheImageBudgetIsTheHosts` (the default holds a 12-frame loop at the dot grid) |
| L-12.3 | Guidance steers hosts to frames at the dot grid of the view (298×152 at 149×38). A host that raises the budget owns the memory it asks for, and the contract says so. | D-20 | — |
| L-12.5 | A frame advance costs **≤ 15 ms** at 149×38 on the reference machine (an 18-core Apple M-series Mac, the gate's reference machine), the blend included (D-61). | D-37 (P-6), D-61 | `BenchmarkFrameAdvance` |
| L-12.6 | `ReadBack` reads through its size limit, as `Load` does; fields and the shared classified set count inside the budget; the classification cache key hashes a table value's exact bits, so two tables never share a key. | D-40 (S-7) | `TestACachedFileIsReadThroughItsLimit`, `TestTheBudgetCountsEveryPart`, `TestTheKeyHoldsAValuesExactBits` |
| L-12.4 | The budget counts retained PNG bytes as well as pixels; a frame's file size is capped relative to its pixel count. | D-30 (F2) | `TestTheBudgetCountsEveryPart`, `TestAFramesFileIsCappedByItsPixels` |

## L-13 — v0.1.0 defects fixed in this release

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-13.1 | The library identifies itself with its real version, not `0.1.0-dev` (`fetch.go:24`). | D-18, W1-B | `TestTheTagIsTheReleaseTheLibrarySays` (`scripts/gate --release`) |
| L-13.2 | `fetch.Checked` is wired, or removed. | W1-B | Removed (D-62): `TestPublicSurfaceSnapshot` holds `Fetcher` gone and `TestTheChangelogListsEveryBreak` its changelog row; `fetch.Checked`, internal, went with it and nothing calls it |
| L-13.3 | The gate's NOT RUN line is current. | W1-B | Done: it names the last tag (D-38), held by `TestNotRunNamesTheLastTag` |
| L-13.4 | The drifted cache doc comments and `ReduceMotion`'s overstated comment are corrected. | W1-B, W1-A | `make lint`-class review |
| L-13.9 | **Severity and valid time are data on each alert**, not read from its colour role. | D-43 (A F5) | `TestAFeatureCarriesItsSeverity` |
| L-13.10 | **The contract states that a host words an image answer's class from `Legend()`**, with the example showing how; **a legend entry says when its table is approximate** (D-19). | D-47 (A F12) | `TestTheLegendSaysWhenATableIsApproximate`, `Example_intensityFromTheLegend` |
| L-13.8 | A test lists every environment read the library makes (`os.Getenv` today only in `look.go`), so a new one cannot arrive unnoticed. | D-40 (S-10) | `TestEveryEnvironmentReadIsListed` |
| L-13.5 | **A list of the alerts shown, needing no place**: each alert's name, severity word and valid time or stale mark — in the new `Report` (D-57). | D-29 (A11), D-43, D-57 | `TestTheAlertsShown`, `TestAnEmptyReportHasEmptySections` |
| L-13.6 | **For a discrete place the host names, each alert is answered on its own** — inside, outside or **nearby** — with its name and severity word, instead of one answer merged across an overlay. **"Nearby" is new in v0.2.0** (v0.1.0 has only inside and outside): within a distance the host sets, default proposed 10 km, PLAN confirms. The description never says "you"; the library returns data, so this binds the contract's, the example's and watchpost's wording. | D-29, D-43 (A F4) | `TestEachAlertIsAnsweredOnItsOwn` (inside, outside, nearby; `SetNearby`) |
| L-13.7 | *Deferred, "necessity unproven yet" (D-29):* intensity in words for image answers (a host words it from `Legend()`, L-13.10); a general summary of what is in view; new place-to-weather statements beyond inside / outside / nearby — **except L-1.12's motion statement, which D-29 keeps and D-42 widens to views with no place, and the answers v0.1.0 already gives** (`Heavier`, `HeavierAt`). | D-29, D-40, D-42, D-47 | — |

## L-14 — Detail by purpose (D-82, D-83: watchpost UAT-1 U1-10, U1-16)

Added during watchpost 0.18.0's UAT-1. **A host chooses how much of the basemap is drawn for its
purpose, and the picture it does not ask to thin is the picture as it was.**

| # | Requirement | Source | Instrument |
|---|---|---|---|
| L-14.1 | `SetDetail` draws the library's own basemap at a level: Essential (coast, water, borders), Weather (and rivers, place names, the major roads), Standard (and rail, parks), Full (and the minor roads, runways). Full is the default. A level and a switched-off layer both apply; a host's own style draws whole. | D-82 | `TestEachDetailLevelDrawsItsRules`, `TestSetDetailThinsThePictureAndFullIsTheDefault`, `TestALevelChangedOnADrawnMapRedraws` |
| L-14.2 | The major roads (motorway, trunk, primary) and the minor roads switch apart. | D-82 | `TestMajorAndMinorRoadsSwitchApart`, `TestMajorAndMinorRoadsSwitchApartOnTheMap` |
| L-14.4 | A replaced overlay is drawn as it was until its replacement is prepared (contract section 4): an overlay handed in again never drops out of the frame. | D-83 (watchpost U1-28) | `TestAReplacedOverlayIsDrawnUntilItsReplacementIsReady` |
| L-14.3 | The footer's scale mark and credit never touch; the credit keeps its width and the scale mark takes what is left, or is not drawn. | D-83 (watchpost U1-1) | `TestTheScaleMarkAndTheCreditNeverTouch` |
| L-14.5 | `WaterLayer` is the lakes and inland water; the sea and its coast are never switched, so the land keeps its edge. | D-85 (watchpost U1-39) | `TestTheSeaOutlastsTheWaterSwitch` |
| L-14.6 | The world repeats across the antimeridian: a view past 180° draws the other side there, and every overlay, marker and name in each copy of the world the view reaches, each whole. | D-86 (watchpost U2-7) | `TestTheWorldRepeatsAcrossTheAntimeridian`, `TestAnOverlayAcrossTheSeamIsDrawnWhole` |
| L-14.7 | `ShowStamp(false)` leaves the top row's stamp - the loop's moment and the stale word - to the host, which must then show both; on by default. | D-87 (watchpost D-92) | `TestAHostCanTakeTheStampOver` |

## L-15 — The moment (D-88: watchpost D-94 to D-98)

| # | Requirement | Source | Test |
|---|---|---|---|
| L-15.1 | An overlay may carry a span, `During`; it is drawn only while the map's moment meets it, both ends included, and a zero end is open. It stays prepared while it is not drawn, so the frame that meets it draws it at once. A span that ends before it begins is refused. | D-88 (watchpost D-96, D-98) | `TestAnOverlayIsDrawnOnlyWhileTheLoopsFrameMeetsIt`, `TestTheClockCrossingASpanRedraws`, `TestASpanOrMomentBackwardsIsRefused` |
| L-15.2 | The moment is the loop's frame while a loop is held; else the span the host says with `ShowMoment(from, to)`; else the frame's clock. A moment the host moves is an input. | D-88 (watchpost D-94, D-97) | `TestTheHostSaysTheMomentWhileNoLoopIsHeld`, `TestAStepSwapsOneDaysFieldForAnother` |
| L-15.4 | A grid marked `Lines` takes L-15.3's look with or without an image on the map, and the legend gives its bands as drawn, faint. | D-89 (watchpost D-102) | `TestAFieldCanAskForItsLinesAlone` |
| L-15.3 | A field that shares the map with an image - drawn in the frame, or on a gap of its loop - is its labelled contours over its bands made faint; the image keeps its own colours. | D-88 (watchpost D-95) | `TestAFieldSharingTheMapWithAnImageIsItsLines` |

## L-16 — Wind (D-90: watchpost D-108 to D-110; go-tuiMaps FR-8)

| # | Requirement | Source | Test |
|---|---|---|---|
| L-16.1 | A vector grid - speeds, and `From`, the direction each blows from - is drawn as braille arrows on an even spacing, pointing downwind, length and colour by class, every other labelled with its speed; a cell with no speed or no direction draws nothing; there is no fill. | D-90 | `TestAnArrowPointsWhereTheWindBlows`, `TestNoWindIsNoArrow`, `TestAStrongerWindIsALongerArrow`, `TestEveryOtherArrowCarriesItsSpeed` |
| L-16.2 | A place's answer is the speed in the grid's unit and the compass word it blows from. | D-90 | `TestAWindGridIsItsArrowsAndItsWords` |
| L-16.3 | The wind preset: six classes in mph, km/h, m/s or knots, calmest first, its colours `wind.1` to `wind.6`, passing the ramp checker as line work on both grounds. | D-90 | `TestTheWindLegendIsItsSixClasses`, `TestTokensMatchTheDocumentedList` |
| L-16.4 | `From` has one direction a value, each 0 to 360 or NaN; else the grid is refused. | D-90 | `TestAWindGridsDirectionsAreChecked` |

## L-17 — Rain as a grid, and a grid's marks (D-91: watchpost D-115 to D-118)

| # | Requirement | Source | Test |
|---|---|---|---|
| L-17.1 | A grid in radar's scale is drawn as rain: its own colours at full strength, over the sea, over any field, never lined; a field beside it takes its lines. | D-91 | `TestARadarGridIsDrawnAsRain` |
| L-17.2 | `Grid.Marks`, the host's text a value, is written at the arrows' spacing on the points an arrow leaves unlabelled, a cell from its neighbours, before any value a field writes. | D-91 | `TestAGridsMarksAreWrittenOnTheMap` |
| L-17.3 | Marks are one a value, each at most eight cells wide; else the grid is refused. | D-91 | `TestAGridsMarksAreChecked` |
| L-17.4 | The legend keys radar's classes in the colours drawn: the class below the first floor not drawn, a grid of rain never faint. | D-91 | `TestTheRainLegendKeysTheColoursDrawn` |

## L-18 — Fire (D-92: watchpost D-121)

| # | Requirement | Source | Test |
|---|---|---|---|
| L-18.1 | Fire is a role of its own: `Fire` for a perimeter, an incident and a strong hotspot, `FireFaint` for a weaker hotspot, tokens `fire` and `fire.faint`; a feature in them is never an alert. | D-92 | `TestFireIsARoleOfItsOwn` |

## L-19 — Quakes as USGS draws them (D-93: watchpost D-122, D-123)

| # | Requirement | Source | Test |
|---|---|---|---|
| L-19.1 | A circle with `RadiusDots` is a braille ring that size on the screen at every zoom, its label beside it. | D-93 | `TestAScreenRingKeepsItsSizeAtEveryZoom` |
| L-19.2 | A radius in km or in dots, never both; in dots 1 to 48. | D-93 | `TestAScreenRingIsChecked` |
| L-19.3 | Three quake roles by age, `quake.hour`, `quake.day`, `quake.older`, apart from fire's and the track's; never an alert. | D-93 | `TestTheQuakesColoursAreTokens`, `TestTheQuakesHaveTheirColoursOnBothGrounds` |

## L-20 — Waves (D-94: watchpost D-125, D-126)

| # | Requirement | Source | Test |
|---|---|---|---|
| L-20.1 | `WaveGrid`, feet or metres, is drawn over the sea alone. | D-94 | `TestWavesAreDrawnOverTheSeaAlone` |
| L-20.2 | Six classes, calmest first, round in each unit, tokens `wave.1` to `wave.6`, passing the checker on each ground's water. | D-94 | `TestTheWaveLegendIsItsSixClasses`, `TestTheWaveScalePassesOnTheSea` |
| L-20.3 | Feet and metres alone. | D-94 | `TestTheWavePresetIsCheckedAndNamed` |

## L-21 — The sea's stations (D-95: watchpost D-127, D-128)

| # | Requirement | Source | Test |
|---|---|---|---|
| L-21.1 | `Buoy` and `Tide`, tokens `buoy` and `tide`, apart from every other feature role and the waves' scale, readable on each ground's water; never an alert. | D-95 | `TestBuoysAndTidesAreRolesOfTheirOwn`, `TestBuoysAndTidesHaveTheirColoursOnBothGrounds` |

## L-22 — The basemap always shows (D-96: watchpost U2-34, D-124)

| # | Requirement | Source | Test |
|---|---|---|---|
| L-22.1 | The basemap is never starved by the overlays: past the queue's cap, a job of a view no map shows goes first, then the oldest that is no tile, a tile last; within a view the tiles run first. | D-96 | `TestTheBasemapDrawsUnderAFloodOfOverlays`, `TestTheBasemapIsNeverStarved`, `TestAViewNoLongerShownGoesFirst`, `TestNewestViewFirst` |
| L-22.2 | While the tiles load, the map says "Loading the map…" and names no call; the host learns it from the `NoTiles` status. | D-96 | `TestTheNoticeWhileTheMapLoadsSpeaksToAPerson`, `TestTheNoTilesNoticeNamesTheRealCause` |

## L-23 — The place names before the data (D-97: watchpost U2-38, D-135)

| # | Requirement | Source | Test |
|---|---|---|---|
| L-23.1 | A frame's words claim cells in this order: the host's places; an alert's label and a marker-role overlay's; the alerts' digits; the basemap's names within their budget; every other overlay's label; the contours' values. | D-97 | `TestThePlaceNamesComeBeforeTheData`, `TestAFieldsContoursCarryTheirValues`, `TestOverlayLabelBeforeBasemap` |

## L-24 — Gusts on the arrows (D-98: watchpost D-136)

| # | Requirement | Source | Test |
|---|---|---|---|
| L-24.1 | A wind grid's `Gusts`, one a value, NaN where none: an arrow's label reads "15G30" where one is given, the speed alone where not. | D-98 | `TestAGustIsSaidBesideItsSpeed` |
| L-24.2 | Gusts only with `From`, one a value, none below zero, or the grid is refused. | D-98 | `TestAWindGridsGustsAreChecked` |

## L-25 — UV and air quality (D-99: watchpost D-137 to D-140)

| # | Requirement | Source | Test |
|---|---|---|---|
| L-25.1 | `uv` and `aqi` presets, their categories' breaks, tokens `uv.1`-`uv.5`, `aqi.1`-`aqi.6`, passing the checker on both grounds at truecolor and 256. | D-99 | `TestUVAndAirQualityAreTheirScales`, `TestTheUVAndAirQualityScalesPass` |
| L-25.2 | Each preset in its one unit, or refused. | D-99 | `TestUVAndAirQualityAreChecked` |
| L-25.3 | `AirQualityRole`: a reading's category as a feature's role; its words in the markers' ink. | D-99 | `TestAMonitorIsDrawnInItsCategory` |

## L-26 — Rain and snow totals (D-100: watchpost D-168, D-184)

| # | Requirement | Source | Test |
|---|---|---|---|
| L-26.1 | A `qpf` preset in mm: WPC's seven classes from 0.25 to 101.6 mm, tokens `qpf.1`-`qpf.7`, passing the checker on both grounds at truecolor and 256; below 0.25, a trace, nothing drawn. | D-100 | `TestRainTotalsAreWPCsScale`, `TestTheRainTotalsScalePasses` |
| L-26.2 | The preset in mm alone, or refused. | D-100 | `TestRainTotalsAreChecked` |

## L-27 — A refreshed loop keeps drawing (watchpost U2-47, its C-10 and D-199)

| # | Requirement | Source | Test |
|---|---|---|---|
| L-27.1 | A loop handed in again draws its old pictures, at the moment shown, until its new ones are decoded - as a replaced shape draws its old form (L11.5) - and keeps none once they land or the loop is removed. | watchpost D-199 | `TestARefreshedLoopKeepsDrawing`, `TestARefreshedLoopDrawsTheOldFramesUntilTheNewLand` |

## L-28 — The names stand still as a loop plays (watchpost UAT-2 U2-46, D-200)

| # | Requirement | Source | Test |
|---|---|---|---|
| L-28.1 | An overlay outside the moment (L-15.1) is not drawn, but the basemap's names are placed as if it were: an alert's word and severity digits hold their room, after every drawn alert's, and a field or image - or a loop on a gap - leaves the name budget the one under it. A name never comes and goes as a loop plays past an alert's hours. Two alerts' words that collide may still trade places frame to frame; a name moves only where it sat in that room. | watchpost D-200 | `TestNamesHoldStillAsALoopPlays`, `TestAReservedAlertHoldsItsWordsRoom`, `TestAReservedOutlineHoldsItsDigits`, `TestTheNameBudgetHoldsAcrossALoop`, `TestReservingIsAChangeOfFrame` |

## L-30 — A render keeps an overlay's preparation (watchpost UAT-2 U2-59)

| # | Requirement | Source | Test |
|---|---|---|---|
| L-30.1 | A render withdraws only the work its view no longer wants: a tile it needs is kept, and so is the preparation, at the bucket in view, of an overlay the map holds - so a loop whose preparation outlasts the gap between two renders is prepared once and drawn. A removed overlay's preparation, or another bucket's, is withdrawn. | watchpost U2-59 | `TestARenderKeepsAHeldOverlaysPreparation` |
| L-30.2 | A loop prepared while the host renders as fast as it can is prepared once: no unit of the work a render planned is cancelled and asked again. | watchpost U2-59 | `TestALoopIsPreparedOnceWhileTheHostRenders` |


## L-29 — A blink phase redraws only what blinks (watchpost W14 P-10)

| # | Requirement | Source | Test |
|---|---|---|---|
| L-29.1 | A frame is redrawn for the blink phase only when a marker on it blinks. With every marker steady the frame is the same in either half of the blink, so the phase flipping - twice a blink period, whatever is drawn - is no change and draws nothing. | watchpost W14 P-10 | `TestAPhaseFlipRedrawsOnlyWhatBlinks` |
## Metrics of success

The brief's M1–M6 (D-6), with round 1's changes. **This table is the normative one.**

| # | Metric | Type | Definition |
|---|---|---|---|
| M1 | Motion seen | Primary | From a loop alone, a reader states a precipitation cell's direction of motion (scripted UAT over loops recorded from real radar); **and a non-visual arm (D-24): from the description alone**, scored by a grader PLAN defines |
| M2 | Loop honesty | Primary | Zero frames drawn under the wrong valid time; every gap stated; every frame older than the host's stated cadence marked stale |
| M3 | Bound holds | Primary | With a host bound set, zero frames outside it across resize, fit, pan, zoom and fall-back |
| M4 | Embed cost with a loop | Primary | A 12-frame loop at dot resolution adds **≤ 1 MiB** of heap over an empty map, within ±5 % over an hour (D-61). Measured +0.59 MiB once, on one machine (an 18-core Apple M-series Mac, the gate's reference machine), under D-61's budget, before D-68. Now 0.6 MiB for 12 frames at dot resolution (298×152), held within ±5 % while playing (`loop_memory_test.go`, `TestALoopsHeapCostAndHold`); 24 region frames cost about 5.8 MiB (`TestTwentyFourRegionFramesCost`) |
| M5 | Contract truth | Secondary | Zero names in `contract.md` — and promises in the README (D-34) — missing from the code, and none the other way; a test |
| M6 | Gate trust | Secondary | Zero unattributable gate failures across a stated number of consecutive full runs, counted from `06_docs/gate-runs.md` (D-31, D-40) |

## Non-functional

| # | Requirement | Source |
|---|---|---|
| NFR-1 | **For v0.2.0, a new feature's better long-term API may break v0.1.0's shape** (D-58): v0.1.0 has no consumer but watchpost, and each break is listed in the contract's changelog. **An earlier decision is revisited only for significant upside**, weighing simplification and developer ergonomics highest, within the shared structure (plugins, discrete domains). Rows that change behaviour a v0.1.0 host can observe: L-7.3 and L-7.4, L-9.3, L-9.4, L-11.1, and the `Report` call, which replaces `Describe` (D-57, D-70). | C-7, D-40, D-58 |
| NFR-2 | Memory at the default budget is D-68's 6 MiB of images, above v0.1.0 NFR-3, until D-68's revisit (OW-14); flat over an hour with a 12-frame loop (M4). | M4, D-20, D-68 |
| NFR-3 | Motion: at a frame step from 400 ms up, nothing on the map changes more than v0.1.0 D-56's 2.5 times a second. A step below 400 ms lifts that ceiling as the host's explicit choice, and the state read says so; it stays under WCAG 2.3.1's three flashes a second (200 ms makes 2.5 flash pairs a second). `ReduceMotion` stops all animation. | v0.1.0 D-56, NFR-21, D-76 |
| NFR-4 | The gate is green before every commit that touches anything but Markdown; a Markdown-only change takes the docs lane, run before committing. Every run is logged in `06_docs/gate-runs.md`. | D-15, D-31 |

## Risk register

Likelihood and severity are **H / M / L**. **Evidence** says what each rating rests on; **Status** says
where it stands.

| # | Risk | L | S | Evidence | Mitigation | Status |
|---|---|---|---|---|---|---|
| RK-1 | **A loop that freezes or re-decodes while every unit test passes** — the three changes (frames, renderer, store) land apart | M | H | W1-A: `internal/render/frame.go:285-302` reuses the last frame; `internal/overlay/store.go:453` drops rasters on re-Set; this codebase has shipped "correct parts, wrongly connected" before (watchpost RK-1) | L-1.6, L-1.7 each held by a test that sees a frame advance (`TestAStepIsRedrawnNotReused`, `TestARefreshDecodesOnlyWhatIsNew`), and the whole path by `TestThePathThroughThePublicMap` | Mitigated |
| RK-2 | **The MRMS heavy end is wrong on a severe-weather day** — the palette was seen on a quiet day (≤ 48 dBZ) | H | H | W2 M-A: palette seen on one quiet afternoon (≤ 48 dBZ); heavy-end colours 3–30 from any legend colour | L-2.3; the fallback-share test; the triggered capture (OW-12); the archive specimen (OW-11) | Open; depends on live weather (RK-11) |
| RK-3 | **No blend passes on a light ground** (every strength tried failed) | H | L | S29-7: 0 of 15 light-ground configurations (3 strengths × 5 severities) clear 10 | L-11.3; the named fallback L-11.4 (D-27) | Mitigated by a named fallback |
| RK-4 | **Watchpost's radar slips** because this release does | H | M | D-36 (no split) and D-37 (no date) accept it; watchpost's own RK-4 | Accepted; watchpost's fallback is to ship on v0.1.0 **without radar, the view bound (HR-3) or the colour-independent pattern (HR-7)**; v0.1.0 is not retracted (D-48) | Accepted (D-36, D-37, D-48) |
| RK-5 | **A provider changes its palette or endpoint** (IEM has no SLA; MRMS no published table) | M | M | W1 watchpost: IEM has no SLA, MRMS no published table; the unmatched count sees only colours more than 10 from any table colour | The unmatched-colour count is the detector (L-2.2). A shift under 10 on IEM's table or a host's is counted and warned of (`near-image-colours`, OW-10, D-104); real IEM frames match exactly (204,000 pixels of the two fixtures, none near). Residual: MRMS's table is read off its legend, so a near match there is expected and not counted | Mitigated |
| RK-6 | **The contract stays wrong** — a names-only test passes with false behavioural claims | H | M | W1-B: five false behavioural claims pass `TestPublicSurfaceSnapshot`, which records names only | L-4.2, L-4.3: each corrected claim held by a behaviour test (`contract_claims_test.go`), the surface by `TestTheSurfaceRecordsSignatures` | Mitigated |
| RK-7 | **The gate hangs on a fuzz stall** — the count budget has no wall clock | L | M | D-3, L-6.4: the engine stalled before; D-31/D-40 limiter tested, mutation-proven | Each fuzz leg runs under a wall-clock limit, TERM then KILL, and fails as TIME LIMIT (D-31, D-40) | Open |
| RK-8 | **A listener who cannot watch the animation is excluded** from the release's main point | M | H | A1, B-4; D-24, D-42 | L-1.12, with or without a named place; M1's non-visual arm. Residual: two-frame motion can mislead when cells grow or decay, so it is worded as observation | Open until M1's non-visual arm is scored (L10.9) |
| RK-9 | **Frames change after hand-in** — the library borrows host image bytes | L | H | InfoSec F1, S-1: `store.go:437` keeps host slices; decode allocates before its check | L-1.14 (D-30): `TestTheStoreKeepsItsOwnCopy`, `TestDecodeReadsTheHeaderBeforeItAllocates` | Mitigated |
| RK-10 | **Evidence cannot be re-run** — wave 2 and specimen 29's blends came from throwaway programs | H | L | D-32 accepts that filed programs go stale unnoticed; BUILD rewrites the loop API they call | Filed with inputs and output in `02-analysis/programs/` (D-32); they may go stale as the API changes | Accepted (D-32) |

| RK-11 | **The heavy-end evidence depends on the weather** — MRMS keeps about two hours of history and publishes no table, so its heavy colours can be seen only on a live severe day | M | H | D-44; W2 M-A | The triggered capture (OW-12); L-2.3's stated fallback if no day comes before SHIP | Open |

## Owed

| # | What | Due | Source |
|---|---|---|---|
| OW-1 | ~~The row in v0.1.0's record saying its close-out did not run~~ **Done (D-38)**: `go-tuimaps/07-readiness/release-checklist.md` | — | D-18 |
| OW-2 | ~~The D-17 specimen: severity word and five dashes, at 69×12 and 149×38, NoColour and Colours16, drawn over specimen 29's radar~~ **Done: specimens 32 and 33; ruled D-64, D-65 (a digit on the outline)** | — | D-17, D-40 (A F7, B) |
| OW-3 | ~~Diagnose S29-3~~ **Folded into L-8.9 (D-60)** | — | D-60 |
| OW-4 | ~~The 40-second `FuzzAgree` freeze~~ **Done (L10.7)**: the engine minimising a new input, for up to `-fuzzminimizetime`; confirmed by a control run (`gate-fuzz-deadline.md`) | PLAN | L-6.4 |
| OW-5 | ~~Correct wave 2's "71–80 %" and D-19's copy of it~~ **Done 2026-09-23**: 80.0–80.6 %, and wave 2's finding 2 restated | — | round 1 verification |
| OW-6 | F-2 (a pluggable architecture): its trigger fired; the narrow seam is L-2.5 (D-35); the broad restructure stays with the quality pass | quality pass | D-35 |
| OW-9 | ~~Diagnose L-8.4~~ **Closed (D-50): no defect** — the round 1 count was wrong | — | D-50 |
| OW-11 | ~~A specimen of a past outbreak~~ **Done at PLAN entry: specimen 30** (2011-04-27, west-central Alabama, 20 warnings over IEM archive radar). Findings S30-1..S30-4 | — | D-44 |
| OW-12 | **A triggered MRMS capture**: when a Storm Prediction Center Moderate or High risk, or a tornado watch, is issued over US radar coverage, capture MRMS frames across the whole two-hour window and extend the palette | when triggered, before SHIP | D-44 |
| OW-13 | ~~The borrow check has no public switch~~ **Ruled D-74: dropped** (L1.7): the check, `Caps.BorrowCheck` and the warning kind `BorrowChanged` are removed | — | L1.2 |
| OW-10 | ~~RK-5's detector: count tolerance-fallback matches, not only unmatched pixels, so a palette shift under 10 is seen~~ **Done (D-104, L10.15)**: `near-image-colours` | PLAN | D-40 (B N-8) |
| OW-8 | ~~The gate never runs `gofmt`~~ **Closed by D-46**: the gate has a formatting leg | — | D-46 |
| OW-14 | **The memory revisit D-68 promised**: the 6 MiB image budget and the region frames' cost (about 5.8 MiB for 24, `TestTwentyFourRegionFramesCost`) weighed against v0.1.0 NFR-3, "once the feature works", against measurements | before v0.3.0, or as the HUM LEAD rules | D-68 |
