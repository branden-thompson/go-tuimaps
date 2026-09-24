---
title: "v0.2.0 ⇄ watchpost 0.18.0 — PLAN red team"
date: 2026-09-23
phase: PLAN
sev: SEV-0
authority: HUM LEAD
status: "ROUND 1 RECEIVED — dispositions pending, one ruling at a time. One round only (the HUM LEAD's lean PLAN close)."
---

# PLAN red team — both plans, one round

**Why one record for two repositories.** The round reviewed both implementation plans and the
integration map together, because the plans meet there. Watchpost's
`observer-maps/08-reports/red-team-plan.md` points here and records watchpost's own rulings.

## Dispatch

Three reviewers, none with prior context, each briefed from watchpost's `06_docs/red-team-brief.md`
template, with li-A2DH's PLAN phase lens (`02_skills/critical-analysis/red-team/phases/plan.md`) and
its A11y and InfoSec personas. Scope: both plans at go-tuiMaps `c3438c6` and watchpost `74274ba`, and
the integration map. The D-17 specimen (OW-2) was being drawn at the same time (D-63) and L3.9–L3.11
were marked PENDING.

| Reviewer | Lenses | Verdict |
|---|---|---|
| Architecture + code | Code quality axis, PLAN lens | **Not ready for BUILD**; fix first: `Changed()`'s meaning (P-1) |
| A11y + InfoSec | Staff Accessibility Advocate, Principal InfoSec Engineer | A11y **not ready**; InfoSec **ready once three tasks are added** |
| Business + PLAN | Business axis, PLAN lens | go-tuiMaps **approve with conditions**; watchpost **not yet**; fix first: W3.9's single-flight zone fetch (P-2) |

Before the round, an internal plan check (A2DH's plan-document reviewer prompt, one reviewer a plan)
found and fixed a first set of gaps (go-tuiMaps `6b5e989`, watchpost `74274ba`).

The three reports are filed verbatim in `reviewer-reports/plan-*.md`, machine paths redacted.

## Findings, consolidated

**Checked against the code before listing.** P-1 was reproduced: in v0.1.0 `Changed()` is raised only
inside `Render` (`map.go:395-397`), nowhere else. P-2 was confirmed: `zones.go:37` already bounds zone
fetches at six, and NFR-4 asks for "sequential or bounded".

### For rulings, most severe first

| # | Finding | Reviewers | Severity | Proposed |
|---|---|---|---|---|
| P-1 | **The shared redraw signal doesn't work.** `Changed()` moves only inside `Render`, and `Work` never raises it, so the contract's "moves whenever a redraw would differ" is false (a sixth false claim). Watchpost's design redraws when `Changed()` or `FrameTicks()` move, so a tile landing in `Work` never reaches the screen. L4.7's "an advance leaves `Changed()` alone" can't hold while `Render` bumps it. Guard 4 as planned either can't fire (one owner) or refuses every real redraw (on v0.1.0) | Arch F3, F7, F8, F14 | Critical | Ruling |
| P-2 | **M5's 2.0 s can't be met on the cold path W3.2 creates.** Zone seeding moves to the first `g`; a cold 41-zone resolve was ~2 s at six in parallel before tiles; my W3.9 made it one at a time (7–15 s) with no requirement asking | Business Q1, Q10 | Critical | W3.9 is my error and is reverted in the batch; M5's definition needs a ruling |
| P-3 | **Every radar request sends the exact on-screen rectangle, centred on the station**, usually a home, timestamped, to two third parties; errors carry the box too | InfoSec S3-1, S5-1 | Important | Ruling |
| P-4 | **Playback semantics are unstated and per-overlay.** Phase origin, `Step` while playing, whether a held position survives a refresh `Set`, Off vs step/seek; playback per overlay id pushes state onto the host, which has one Setting | Arch F1, F6 | Important | Ruling |
| P-5 | **Nothing surfaces a failed or unverified MRMS heavy end to the listener**, and `MaxFrames = 36` refuses a full two-hour MRMS loop (about 60 frames) | Business Q2, Q10; Arch F16 | Important | Ruling (with the numbers) |
| P-6 | **P1-b waits on the whole v0.2.0 tag**, including hosted CI, the soak, M6's runs, OW-4 and M1 grading; integration defects like P-1 would surface only after release | Arch F22; Business Q6 | Important | Ruling |
| P-7 | **Simplification under D-58**: `PurgeWithReport` beside `Purge`; option-plus-setter pairs; a variadic `CacheOption` for one option; public `Tables()`/`ProviderTable`; `Describe` kept beside `Report`; watchpost's self-registering registry beside a closed table; about 45 new exported names for one consumer | Arch F9, F18; Business Q4 | Important | Ruling (a break list) |
| P-8 | **Watchpost's accessibility in P1-a**: the description built on `Describe` merges alerts (needs one overlay per alert and its own alert list); no playback keys in the window; the motion default is unstated; the description can't scroll and is read after the braille; the "existing voice can speak it" claim has no task; the severity word is not put in P1-a's labels although watchpost writes them | A11y A1-1, A2-1, A3-1, A3-2, A4-1, A5-1 | Important | Ruling on the voice claim and the motion default; the rest in the batch |
| P-9 | **Metrics that the plans' own tests could game or skip**: M1b declared "green" on `Report` without a human re-score; M1's non-visual arm has no ground truth and grades data, not words; watchpost's M6 narrowed to schedule drift (which can't fail) and gated on wall-clock time; timing asserted inside `make verify` | Business Q9; A11y A6-1, A6-2; Arch F17, F19 | Important | Ruling (instruments batch) |

### Batch — corrections that enforce what is already ruled

| # | Finding | Reviewers | Fix |
|---|---|---|---|
| B-1 | `Report`'s shapes clash with v0.1.0's (`DistanceKm` ignores `Units()`; `Bearing` a word where v0.1.0 means degrees; `Relation` a string where `Where` exists); `Feature` has no ID to join back to; `PlaceReport` lost image answers | Arch F10, F11 | Reuse `Distance`, `Unit`, `Bearing`, `Compass`, `Where`; add `Feature.ID`; image answers in `PlaceReport`; `Frame.FrameTicks` to match the method |
| B-2 | Radar's HTTP client refuses no private address and doesn't pin https, while RK-11 says it does; no memory-only option (radar can reach disk) | InfoSec S4-1; Arch F20 | Radar dials through `CheckedDialer` or an equivalent in `httpx`, https pinned; an explicit memory-only request option |
| B-3 | "Clear map data" misses zone geometry and races in-flight fetches; W9.5 overstates | InfoSec S6-1 | Zones get an explicit TTL and join the clear; the map closes before the P1-a delete; a test with a fetch in flight; `PurgeHost` kept |
| B-4 | Frame valid times aren't bounded against the clock; the advertised time list isn't trimmed to the cap; `Report` strings reach the screen uncleaned; CI actions unpinned; proxy variables unlisted | InfoSec S1-1, S1-2, S2-1, S4-2, S7-1 | Each as a task line |
| B-5 | Missing tasks: L-5.4 (v0.2.0's REVIEW covers v0.1.0); the MRMS unverified state on the notes line; missing zones named from the area description, not codes; the partial-area line inside the degradation order | Business Q1, Q2; A11y A3-3, A5-1 | Add them |
| B-6 | Under-specified: L5.5's sighting rule (how "the heavier rain" is found, with a multi-cell fixture); lowering the budget below what is held; when the light-ground fallback is decided (re-run on palette or ground change) | Business Q8; Arch F4, F24 | Specify in the plan |
| B-7 | Dependency edges: `Drop` defined in pending L3.11 but used by L3.2 and L3.13; L5.5 needs L6; L3.9 needs L5.1; W9.1 needs `SetBound` on region change; W8 needs W5; W9.8/W9.9 depend on OW-2; external services (vuln DB, arm64 runner, SPC trigger) | Arch F23; Business Q8 | Add the edges; move `Drop` into L3.2 |
| B-8 | Record: RK-4's cost list is incomplete (labels, `Describe`, the `-dev` agent, retention); RK-5's basemap fallback never ruled; what ships if OW-2 fails; the integration map's node ids shifted; a check that the map and both plans agree | Business Q2, Q3; Arch F2, F5 | Correct the record; the check as a docs-lane test |
| B-9 | Shapes that aren't Go: `const` struct; untyped parameters; W2.2's goroutine test names no mechanism; L1.3/L1.4/L1.6 test that sentences exist, not that they are true; L10.5 can pass without its architecture leg | Arch F12, F13, F15, F17 | Correct each |
| B-10 | Watchpost's M5/W3.9 conflict: restore the bound of six | Business Q1 | Revert W3.9 to "bounded at six, tested" |

## Dispositions

Each ruling lands in the owning repository's rulings log the moment it is made; watchpost's are listed
in its `observer-maps/08-reports/red-team-plan.md`.

| # | Ruling | Outcome |
|---|---|---|
| P-1 | **D-66** here, watchpost **D-45** | `Changed()` counts inputs and is raised by `Work`, never by `Render`; `NextCall` is the time signal; `Frame` carries its counters. Watchpost renders on every event it owns; a freshness property replaces the memo key and guard 4 |
| P-2 | watchpost **D-46** | M5 keeps "complete" as every alert's area, ≤ 3.5 s p90; W3.9 reverted to the bound of six |
| P-3 | watchpost **D-47** | Radar requested per fixed state-regional region, never per view |
| P-5 | **D-68** here, watchpost **D-50** | Two hours of loop, 24 frames at 5 minutes, image budget default 6 MiB; memory revisited once it works; MRMS at 5 minutes by default, source and step host Settings; `MaxFrames` 72 |
| P-4 | **D-67** here, watchpost **D-48** (and D-49) | The listener's control flow is the API: one playback per map; play from the oldest through observed to forecast; stop; reset to "right now"; step; position by valid time; forecast frames. Pan and zoom came into 0.18.0 |
