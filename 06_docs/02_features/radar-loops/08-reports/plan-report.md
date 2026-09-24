---
title: "go-tuiMaps v0.2.0 — Radar loops — PLAN REPORT"
date: 2026-09-23
phase: PLAN — exit
level: LEVEL-1
sev: SEV-0
authority: HUM LEAD
directives: FULL GIT; FULL DOCS; FULL REPORTS; FULL DIAGRAMS; FULL RCC; FULL PLAN; FULL TDD; FULL INST
branch: feature/radar-loops
issue: "branden-thompson/go-tuimaps#2"
status: "FOR THE HUM LEAD'S APPROVAL — PLAN's exit artefact"
---

# go-tuiMaps v0.2.0 — Radar loops — PLAN REPORT

## Bottom line up front

**PLAN is complete and recommends proceeding to BUILD.** v0.2.0 is planned as ten work packages
(WP-L1 … WP-L10) and 88 test-first tasks, each naming its files, its signatures and the test it writes
first, with no implementation code (D-52). Every requirement, metric and owed item is traced to a task
(`04-development/implementation-plan.md`, "The trace"). It is built in parallel with watchpost 0.18.0;
`03-architecture-design/integration-map.md` is where the two meet.

What a reader should know before the detail:

1. **The API follows the listener, not the other way round.** Playback is the control flow the HUM
   LEAD set out (D-67): the map opens on "right now"; play runs from the oldest frame through the
   present and on into forecast frames; stop holds; reset returns to now; stepping scrubs. A position
   is a moment in time, so a refresh never moves the viewer.
2. **The one signal the library and its host share was broken, and is fixed by design** (D-66).
   `Changed()` was raised only inside `Render`, so a host could not learn that new data had arrived
   without drawing. It now counts inputs; `NextCall` carries time.
3. **Severity without colour is a digit on the outline** (D-65), chosen from two drawn specimens: the
   first approach (dashes) failed at the smallest map size; the second (a mark along the outline)
   carried severity at every size.
4. **Simpler than first drafted.** Seven API simplifications under D-58 (D-70): one `Purge`, setters
   without duplicate options, the provider seam internal, `Describe` replaced by `Report`.
5. **Watchpost can integrate early.** Release candidates are tagged as packages land (D-69), so
   integration defects surface while v0.2.0 is still open. The final tag still needs every package.

## What BUILD will make

| Package | What | Tasks |
|---|---|---|
| WP-L1 | The contract matched to the code, both ways, by a test; its false claims corrected and each held by a behaviour test; a changelog of every break | 6 |
| WP-L2 | Loops: frames handed in declaratively, copied, validated, keyed for reuse, inside a host-set image budget (6 MiB default) | 10 |
| WP-L4 | Playback: one per map; play, stop, reset, step; position by time; forecast frames; the change ceiling; the redefined counters | 12 |
| WP-L3 | The renderer: frame identity, the radar-over-alert blend, furniture never erased, place names kept, the severity word and digit | 15 |
| WP-L5 | `Report`: the alerts shown, each place's answer per alert (inside, outside, nearby), motion, the legend's approximate flag; `Describe` removed | 8 |
| WP-L6 | MRMS's colour table and the internal provider seam; the heavy end valued better, warned on, and marked unverified until a severe day is captured | 9 |
| WP-L7 | A view bound, set by the host and held on every path, across the antimeridian | 3 |
| WP-L8 | Fetch options: the host's transport, user-agent and timeout; one allow-list; a checked dialer; the real version | 7 |
| WP-L9 | Cache age from fetch time, a maximum age, a purge of everything, a generation so no write lands after a purge | 6 |
| WP-L10 | Defects and the first release's close-out: CI, the vulnerability check, the checklist refusal, M1/M4/M6's evidence, the REVIEW scope | 12 |

Order: WP-L1, then WP-L2 → WP-L4 → WP-L3 → WP-L5 on the critical path, with WP-L6, WP-L7 and WP-L8 → WP-L9
beside it; WP-L10 closes.

## What PLAN decided

| Ruling | Decision |
|---|---|
| D-53 | The docs lane takes the generated atlas |
| D-54 | Loops are declarative: the host hands in frames; the library keeps and reuses them |
| D-55 | The host supplies the transport; time is bounded only if the host's transport honours its context (a recorded deviation) |
| D-56 | A cached tile's age is its fetch time; reads never write |
| D-57, D-70 | A structured `Report`, which replaces `Describe` |
| D-58 | v0.2.0 may break v0.1.0 for a better long-term shape, each break listed |
| D-59, D-66 | Each frame carries its counters; `Changed()` counts inputs |
| D-60 | The named place's label is kept |
| D-61, D-68 | Numbers: image budget 6 MiB, "nearby" 10 km, ≤ 15 ms a frame advance, `MaxFrames` 72 |
| D-62 | `Fetcher` removed |
| D-63, D-64, D-65 | Severity without colour: a digit on the outline (Extreme 4 … Minor 1, Unknown `?`), the key in the host's legend |
| D-67 | Playback follows the listener's control flow; forecast frames |
| D-69 | Release candidates per landed package |
| D-71 | M1's non-visual arm scored first, on words, against a ground truth set beforehand |
| D-72 | The red team's batch of corrections |

## Numbers for approval

| Number | Proposed | Why |
|---|---|---|
| A frame's file-size cap | 4 bytes a pixel plus 64 KiB | An honest uncompressed frame fits; a padded one does not |
| The fallback-share threshold (L6.9) | **1 % of rain pixels** | Default matching left under 0.2 % unmatched on live MRMS (wave 2); five times that flags a changed palette without firing on a normal day |
| M6's run count | Five consecutive full runs with no unattributable failure before SHIP | Cheap; it catches a flake that fires one run in three about 87 % of the time, and says so |
| The reference machine | The development Mac named in wave 2's Limits | Where wave 2 measured |

## Critical analysis

**Internal plan check** (A2DH's plan reviewer, one reviewer a plan) found missing tasks, wrong paths,
order errors and shapes left open; fixed before the red team (`6b5e989`).

**PLAN red team**, one round, three reviewers across both plans and the map
(`08-reports/red-team-plan.md`, reports verbatim in `reviewer-reports/plan-*.md`):

| Reviewer | Verdict | Fix first |
|---|---|---|
| Architecture + code | Not ready | The shared redraw signal (P-1) |
| A11y + InfoSec | A11y not ready; InfoSec ready with three tasks | One overlay per alert; radar by region |
| Business + PLAN | Approve with conditions | Watchpost's cold path (P-2) |

Every finding is ruled or applied: P-1 … P-9 and the batch B-1 … B-10 (D-66 … D-72 here; watchpost
D-44 … D-55).

## Risks carried into BUILD

| Risk | Where it stands |
|---|---|
| MRMS's heavy end on a severe day (RK-2, RK-11) | No oracle until a triggered capture (OW-12); if none before SHIP, MRMS ships marked unverified and the host says so |
| The blend on a light ground (L-11.4) | Searched in BUILD; the named fallback (radar over tint) ships if nothing passes |
| Memory at region scale | 6 MiB by ruling; re-measured on region frames in BUILD (M4), revisited once it works |
| Integration surprises | Release candidates per package (D-69) put watchpost on the code early |

## Gates

```
QUALITY GATE REPORT | go-tuiMaps v0.2.0 | SEV-0 | PLAN exit
-------------------------------------------------------------
  [PASS] approach_selected          : four approaches ruled (D-54 … D-57), refined by D-66, D-67, D-70
  [PASS] design_documented          : approach docs, integration map, atlas (PLAN diagrams labelled)
  [PASS] implementation_plan        : 10 packages, 88 tasks, every requirement traced; no code (D-52)
  [PASS] specimens                  : OW-2 drawn twice (32, 33) and ruled (D-65); OW-11 (30)
  [PASS] critical_analysis_complete : internal plan check + 1 red-team round, every finding dispositioned
  [PASS] scripts/gate               : green — see 06_docs/gate-runs.md
  [PEND] human_approval             : this report
-------------------------------------------------------------
```

## Recommendation

**Proceed to BUILD**, starting with WP-L1 (the contract skeleton, which every other package lands
in), and approve the four numbers above.

## Source documents

`01-objectives/requirements.md` · `02-analysis/rulings.md` · `03-architecture-design/approach-*.md` ·
`03-architecture-design/integration-map.md` · `04-development/implementation-plan.md` ·
`08-reports/red-team-plan.md` · `08-reports/reviewer-reports/plan-*.md` · specimens 29–33 ·
`06_docs/gate-runs.md` · the atlas
