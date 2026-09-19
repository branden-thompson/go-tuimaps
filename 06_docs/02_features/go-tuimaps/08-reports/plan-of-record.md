---
title: "go-tuiMaps — PLAN OF RECORD"
date: 2026-09-19
phase: PLAN
report_template: por-report
level: LEVEL-1
sev: SEV-0
authority: HUM LEAD
directives: FULL GIT; FULL DOCS; FULL REPORTS; FULL DIAGRAMS; FULL RCC; FULL PLAN; FULL TDD
status: "APPROVED — HUM LEAD 2026-09-19 (\"APPROVED; Go 4 BUILD\"), recorded as D-95. The six matters listed as the coordinator's were accepted; none was struck. PLAN exited; BUILD opened."
---

# go-tuiMaps — PLAN OF RECORD

## Bottom Line Up Front

**Recommendation: approve the plan and open BUILD.** Three rounds of adversarial review ended with no Critical finding open; every question they raised is ruled; six matters that are the coordinator's and not rulings are listed below for HUM LEAD to accept or reject with this plan.

- **What PLAN decided.** Three approach questions, each put to HUM LEAD with its alternatives: the host runs all background work and the library never starts a goroutine (D-73); overlays are plain structs set by id, with the four ready-made types as values a host can override (D-74); the library parses tiles itself and depends on two small text-measuring modules and nothing else (D-75, amended by D-81). A fourth ruling added a "fit these things in the frame" call to the first release (D-76).
- **What PLAN produced.** A written contract a host can rely on. Every number a test needs, in one file. 34 architecture diagrams at four levels of detail, with a guided tour, plus 7 in the approach notes and the plan — all 41 checked to parse. Eight new groups of rendered specimens (M to T on the review page), each seen by HUM LEAD. A pinned 6.6 MB test fixture of real data. An implementation plan of 15 work packages and 284 tasks, every one a failing test first, with no code in it (D-71). A mapping from each of the 62 parity rows in this release to the work package that owns it and the test that proves it.
- **What the evidence added.** The first release has a demonstrated way to get radar in — one public-domain image per bounding box, every one of 20,959 sampled pixels matched to its table exactly. A 17-class absolute temperature scale and a six-class radar ramp, each with colours for a dark ground and for a light one, pass a colour-vision check across every pair of classes and against their ground (D-88, D-91). Painting the ground takes the share of map cells under 3:1 contrast on a white terminal from 85% to none. One view costs 1.0 to 1.5 MB of lasting memory.
- **What is still unproven.** Peak memory. Three maps over the fixture region come to about 3.0 MB of lasting memory by the measured parts, against lean caps HUM LEAD chose (D-85); what the maps are drawing is never evicted, so three maps on three different dense views run to about 4.9 MB — over the 4 MB line, reported by the library, not prevented (D-90). The peak while two tiles decode at once can only be measured with the library. **RS-7 stays High** until BUILD's benchmark (task 14.6) says otherwise.
- **What is not designed yet.** The alert preset's colours (risk RS-26). PLAN's specimens drew two of five severities, on a dark ground; checked in round 2, those pass there and **fail 3:1 on a light ground**. Task 08.10 designs all five for both grounds, and HUM LEAD looks at them inside milestone M-B (task 08.23).
- **How it was checked.** Three red-team rounds in PLAN, each on a frozen tree with fresh reviewers. Round 1: NO-GO, 74 findings, 6 Critical. Round 2, on the remediation: NO-GO narrowly, 60 findings, 2 Critical — both rules the coordinator had never written down. Round 3, narrow: ship-with-conditions, 27 findings, no Critical. Every finding has a disposition; every round found fixes from the round before that had not held (13, then 8), which is why there were three.
- **Rulings.** 24 in PLAN before this report (D-71 to D-94; its approval is D-95), one at a time, recorded in HUM LEAD's words. Five went against the coordinator's recommendation (D-73, D-84, D-85, D-86, D-92); in D-86 HUM LEAD's instinct was the better design, and the record says so.

## Context

| Field | Value |
|---|---|
| Project | go-tuiMaps — a Go library that draws a map in a terminal and lets a host lay its own data over it; first used by Watchpost |
| Phase | PLAN, at exit |
| Level · severity | LEVEL-1 · SEV-0 — HUM LEAD approves every decision |
| Collaboration | HUM LEAD |
| Dates | PLAN opened and closed 2026-09-19 |
| Branch | `feature/go-tuimaps`, to merge into `release/v0.1.0`; no remote exists and nothing is pushed (D-19, D-28) |
| Baselines | TerminalMap `3b960723`, MAPSCII `4fe9a60a` — unchanged since DISCOVER |
| Module | `github.com/branden-thompson/go-tuimaps`, Go 1.25 |

## Selected approach

PLAN had three questions where more than one design could meet the requirements. Each was written up with its alternatives before HUM LEAD ruled.

| Question | Ruled | In one sentence | Why this one |
|---|---|---|---|
| Who runs background work | **The host** (D-73) | The library never starts a goroutine; the host calls `Work` from its own, asks `Pending` whether there is any, or calls `Settle` to run it all at once. | One rule with no exceptions: nothing happens that the host did not start. It fits a host that already owns its clock and its goroutines, it makes every test deterministic, and it keeps a library inside someone else's program honest about what it costs. HUM LEAD chose it over the coordinator's recommendation. |
| The overlay contract | **Plain structs, set by id** (D-74) | A host fills in a struct — shape, data, a type — and calls `Set`; the four ready-made types are values it can copy and change. | The least to learn, nothing hidden, and a host's own type is the same struct as a preset (D-69). |
| Dependencies and layout | **Own tile decoder; a two-module allow-list** (D-75, D-81) | The library parses tiles itself, with the field-tested decoder kept outside the shipped module as a test oracle; it imports `go-runewidth` and `uax29` for text width, which the first host already carries. | Memory: data must be dropped *while* decoding and limits enforced *before* allocating, which a borrowed decoder cannot do. |

Rulings that shaped the design after the approaches:

| Ruling | What it set |
|---|---|
| D-76 | `FitTo(places, overlays, margin)` is in the first release |
| D-78 | A cell shows the heaviest rain inside it — "better to overstate than under" |
| D-82 | One configured label language, English by default; it joins the tile cache key |
| D-83 | A cell's colour is decided by majority vote, as upstream does — "this is not a navigation tool so road accuracy is less important the greography and topogrphy" |
| D-84 | No limiter on decoding; the peak is the host's to govern, and is stated at two `Work` calls wide |
| D-85 | Lean caps: tile cache 0.5 MB and shape cache 0.25 MB, shared by every map; images 0.25 MB a map |
| D-86 | `Set` and `Remove` never block; the end of a borrow is reported to the host, never waited for |
| D-87 | Water masks scalar fields and never images; a host can flip either |
| D-88 | Colour-vision difference of at least 10 for every pair of classes and against the ground |
| D-90 | Need first, cap second: what a live view is drawing is never evicted; a cache over its cap because of need says so |
| D-91 | The temperature preset has a light-ground set of colours, as radar has |
| D-92 | `Set` itself indexes a shape large enough to be drawn from the host's memory, so a replaced shape never vanishes and a one-goroutine host is always told "released" |
| D-94 | The app gains `--style PATH`; per-layer label margin and clustering are not taken up for v1 |

## Alternatives considered

### Background work — Model A: the library owns a small pool
**Description:** the library starts a bounded set of goroutines on first use and stops them on `Close`.
**Pros:** a map in three calls with no pump to write; the usual shape for a Go library.
**Cons:** goroutines a host did not start, inside a host with published budgets; harder to test deterministically; a `Close` that must be called.
**Why not selected:** HUM LEAD ruled B. The three-call path survives through `Settle`.

### Background work — Model C: B as the engine, A as a thin default
**Description:** the host-run engine, plus an optional built-in pump.
**Pros:** both audiences served. **Cons:** two modes to document, test and keep in agreement.
**Why not selected:** the coordinator's recommendation; HUM LEAD preferred one rule. The example pump (task 12.13) is the documentation C would have been.

### Overlay contract — Style B: layer handles
**Description:** `AddLayer` returns a handle; data is set on the handle.
**Pros:** typed per shape. **Cons:** two calls for every overlay, handles to keep, a lifetime to get wrong.
**Why not selected:** more to learn for no capability gained.

### Overlay contract — Style C: functional options
**Description:** `AddOverlay(shape, WithRamp(...), WithUnit(...))`.
**Pros:** short call sites. **Cons:** hides what an overlay *is*; a host's own type becomes a different thing from a preset.
**Why not selected:** conflicts with D-69 — presets are values, not locks.

### Dependencies — Option A: the standard library only
**Pros:** the cleanest module graph. **Cons:** thousands of generated lines of Unicode tables, regenerated each Unicode release, to arrive at best where the first host already is.
**Why not selected:** cost with no benefit to the first host.

### Dependencies — Option C: a proven tile decoder too
**Pros:** the least parsing code to write. **Cons:** cannot drop data at decode or check limits before allocating; six to eight times memory expansion; a deprecated module in every consumer's graph.
**Why not selected:** the memory arithmetic fails.

Full write-ups: [approach 1](../03-architecture-design/approach-1-background-work.md) · [approach 2](../03-architecture-design/approach-2-overlay-contract.md) · [approach 3](../03-architecture-design/approach-3-dependencies-and-layout.md).

## Architecture and design

**Components.** One public package and a set of internal ones that meet only through shared types in `internal/scene`: `project` (projection and the braille canvas), `textsafe` (cleaning every string on the way out), `mvt` (the tile decoder), `archive` and the generator (embedded tiles), `fetch` (the one guarded door to the network), `tiles` (sources, cache, states), `work` (the job queue the host drains), `style` and `colour`, `render`, `overlay`, `describe`, `fault` (the typed error and both closed lists of kinds), and a small app. `render` draws prepared types and never imports the packages that prepare them; `work` knows jobs only by `scene`'s job type. That is what keeps the import graph free of cycles.

**Interfaces.** The [contract](../03-architecture-design/contract.md) is the reference. In brief: `New` with options; `SetSize` and the view intents, `FitTo` among them; places by id; `Set` and `Remove` for overlays, which never block and say whether the old geometry was released; `Render`, which never waits, never fetches and is never blank; `Work(ctx)`, `Pending`, `OnPending(wake)` and `Settle` for the host's pump; `Changed` and `NextCall` for an idle host; `Legend`, `Credits`, `Scale`, `Footer` and `Describe` for the same facts as plain data; `CacheUse`; `Warnings`; `Close`. Closed lists of 24 error kinds and 13 warning kinds. Calls in three classes — owner, pump, any goroutine — with four rules the implementation must keep.

**Data model.** A frame is rows of cells — glyph, foreground, background — whose buffers persist between renders. A tile is decoded straight into a compact form with 16-bit coordinates (extent at most 8,192), keeping only the layers the map draws and one label language. An overlay is a struct: an id, a shape (features, image, or scalar grid in this release), borrowed geometry or owned pixels, a type (a preset or the host's own: classes, breaks, colours, unit), a valid time. Borrowed geometry is never copied; the host is told when the library has stopped reading it.

**Numbers.** Every limit, cap, tolerance and target a test needs is in [constants.md](../03-architecture-design/constants.md), each marked as ruled, set in PLAN with its reason, measured, or worked out by arithmetic with the BUILD task that measures it.

## Architecture diagrams

**Tool:** Mermaid, in the repository beside the text it explains (D-72). **Status:** Active — all 41 parse, checked in a browser harness. They are living references: a design change updates its diagram in the same commit (D-71). Start with the [guided tour](../03-architecture-design/architecture-tour.md) — one story, 29 steps, through almost every diagram.

| Level | File | Diagrams |
|---|---|---|
| 0 · Context | [architecture.md](../03-architecture-design/architecture.md) | Context and trust boundary |
| 1 · The parts | [architecture.md](../03-architecture-design/architecture.md) | The parts · The public contract at a glance |
| 1 · The contract | [contract.md](../03-architecture-design/contract.md) | The pump, drawn · Shared caches |
| 2 · Render | [L2-render.md](../03-architecture-design/L2-render.md) | From a call to a frame · The braille canvas |
| 2 · Tiles | [L2-tiles.md](../03-architecture-design/L2-tiles.md) | Where a tile can come from · The network edge · The embedded tiles and their generator |
| 2 · Overlays | [L2-overlays.md](../03-architecture-design/L2-overlays.md) | One path for every overlay · What "prepare" means for each shape · Presets and host-defined types |
| 2 · Colour | [L2-colour.md](../03-architecture-design/L2-colour.md) | How a value becomes a cell colour · The ground |
| 2 · Style | [L2-style.md](../03-architecture-design/L2-style.md) | From a tile's layers to what is drawn · Choosing the profile · Placing labels |
| 2 · View | [L2-view.md](../03-architecture-design/L2-view.md) | Zoom buckets · Fit-to |
| 2 · Describe | [L2-describe.md](../03-architecture-design/L2-describe.md) | How it is computed · Picture and description cannot disagree |
| 2 · Errors | [L2-errors.md](../03-architecture-design/L2-errors.md) | Errors and warnings |
| 2 · Memory | [L2-memory.md](../03-architecture-design/L2-memory.md) | Where the bytes live |
| 2 · App | [L2-app.md](../03-architecture-design/L2-app.md) | The standalone app |
| 2 · Gates | [L2-gates.md](../03-architecture-design/L2-gates.md) | Tests and gates |
| 3 · Sequences | [L3-sequences.md](../03-architecture-design/L3-sequences.md) | A cold first frame · A pan · An idle host · A one-shot render |
| 3 · States | [L3-states.md](../03-architecture-design/L3-states.md) | A tile · Borrowed geometry · A marker's animation · An overlay's freshness |

That is 34. The other 7: three, two and one in the approach notes, and the work-package order in the plan.

## Implementation plan

The [plan](../04-development/implementation-plan.md) holds the tasks; this is its shape. Every task is one cycle — a failing test, the code to pass it, a refactor — and names its requirement and its diagram. There is no code in the plan (D-71). Each package that owns parity rows closes with a task listing the rows no other task names, so all 62 mapped tests have a home.

| # | Work package | Depends on | Tasks | Builds |
|---|---|---|---|---|
| WP-00 | Scaffold, gates, scene types, parity mapping | — | 13 | L1 "The parts" |
| WP-01 | project | 00, 02 | 13 | The braille canvas; projection |
| WP-02 | textsafe | 00 | 11 | L0 "Way out" |
| WP-03 | mvt decoder | 00, 02 | 19 | Untrusted-input gate |
| WP-04 | archive · generator · assets | 03, 05 | 16 | The embedded tiles and their generator |
| WP-05 | fetch | 00, 02 | 12 | The network edge |
| WP-06 | tiles | 02, 03, 04, 05 | 21 | Where a tile can come from; a tile's states |
| WP-07 | work | 00, 02 | 15 | All four sequences |
| WP-08 | style and colour | 01, 02 | 24 | Both colour diagrams; style |
| WP-09 | render | 01, 02, 08 | 32 | Both render diagrams; a marker |
| WP-10 | overlay | 01, 02, 07, 08 | 26 | All three overlay diagrams; borrowed geometry; freshness |
| WP-11 | describe · answer key | 01, 02, 10 | 17 | Both describe diagrams |
| WP-12 | public package · examples | 06, 07, 09, 10, 11 | 29 | The public contract |
| WP-13 | app | 12 | 16 | The standalone app |
| WP-14 | parity · reference frames · benchmarks · judging | 12, 13 | 20 | The evidence for M1 to M5 |
| | **Total** | | **284** | |

| Milestone | Reached when | Shows |
|---|---|---|
| M-A A map from embedded tiles | WP-00 to WP-07, the basemap half of WP-08 and WP-09, tasks 12.1, 12.2 and 12.4 | The three-call world map; never blank; no goroutines; deterministic frames |
| M-B Overlays | The rest of WP-08 and WP-09; WP-10 | Alerts, radar by the image path, temperature; presets; safe ramps; no-colour forms |
| M-C The same facts as words | WP-11 | The description exact for every scenario in the slice, against the independent key |
| M-D Release candidate | WP-12 to WP-14 | The app; examples; 62 parity rows; M1 to M5; the benchmark against the pinned fixture |

**Quality checkpoint: after M-A**, the first working map is shown to HUM LEAD before overlays start.

**On time.** HUM LEAD, D-80: "estimates mean nothing - I want it done right, so I'm willing to wait / use the time that's needed". The plan keeps a first estimate as a record — 33 to 41 working sessions, 43 to 53 with a 30% allowance — and nothing in BUILD is to be cut, hurried or re-ordered to meet it.

**After this release.** Integration into Watchpost at its v0.17.0; a written integration review; HUM LEAD's GO before the remaining overlay shapes are built (D-60). The API may break only at a minor version, guarded by a contract check.

## Risk mitigations

26 risks: 3 High, 14 Medium, 7 Low, 2 Closed. The [risk assessment](../02-analysis/risk-assessment.md) has all of them; these are the ones the plan has to carry.

| Risk | Level | Mitigation in this plan | What remains |
|---|---|---|---|
| RS-2 The overlay contract is designed before real use | High | Designed to D-74 and D-69; one struct for presets and a host's own types; built in WP-10 and WP-12 | It stays High until the first host has used it. The integration review (D-60) is the check, and the compatibility promise (NFR-22) prices a change |
| RS-7 Memory inside the host's margin | High | Own decoder that drops data while decoding (D-75); lean shared caps (D-85); need first, cap second, with `CacheUse` to read it (D-90); borrowed geometry never copied; measured in PLAN at 1.0 to 1.5 MB a view and about 3.0 MB lasting for three maps over the fixture region | **The peak is unproven.** Task 14.6 measures it on the pinned fixture at two `Work` calls wide. If it fails, the caps and the pump width are the levers, and both are the host's to set (D-84) |
| RS-11 M1 depends on changes in the first host | High | The host fetches (D-15); PLAN proved a radar source for the image shape and pinned real alert, radar and tile data as the fixture | Watchpost's own producers are Watchpost's work, at its v0.17.0 |
| RS-4 Background work | Medium | Ruled (D-73); contract sections 2 and 6; WP-07; an example pump | A host that never calls `Work` sees a warning after 20 renders, and a map that is never blank |
| RS-12 Own parser of hostile bytes | Medium | Limits before allocation, all with measured headroom; fuzzing; the field-tested decoder as a test-only oracle (WP-03) | Fuzzing finds what it finds |
| RS-19 A misleading picture | Medium | One outline, one fill rule, two uses; the description is exact and tested against an independent key (WP-11); "on the edge" under one cell (D-67) | The 14,000-vertex zone is a BUILD test (14.11) |
| RS-24 Light-background terminals | Medium | The ground is painted by default (D-64); radar and temperature each have colours for a light ground, checked to D-88 (D-91, D-93) | **The alert preset is not designed for a light ground yet** (task 08.10). A host that declares its own ground takes on the contrast |
| RS-25 A font without braille | Medium | The host states its needs; the description is the fallback (D-57); the app's test card (13.10) | Cannot be detected from inside a terminal |
| RS-26 The alert preset's colours are not designed | Medium | Task 08.10 designs five severities for two grounds, test first; HUM LEAD's look is task 08.23, inside M-B | If no set passes, outlines fall back to black or white and severity is carried by hatch and label, as with no colour |
| RS-14 Themeable ramps that must still make sense | Medium | Presets pass the check; a host's colours are checked and reported; safe-ramps skips the theme (D-63) | A host may ignore the report |

## Decisions

Rulings in PLAN, each put alone with its evidence, options, a recommendation and the strongest counter-argument. Verbatim text and the options as presented: [rulings](../02-analysis/rulings-discover.md) · [options](../02-analysis/rulings-options-as-presented.md). ◇ marks a ruling against the coordinator's recommendation.

| # | Subject | HUM LEAD's words |
|---|---|---|
| D-71 | Discovery Report; directives for PLAN | "APPROVED 4 PLAN" — no large code blocks in the plan; full diagrams for every major part |
| D-72 | PLAN kickoff, Mermaid-first | "Approved" |
| D-73 ◇ | Who runs background work | "B" |
| D-74 | The overlay contract | "A" |
| D-75 | Dependencies and layout | "B" |
| D-76 | Fit-to | "A" |
| D-77 | Specimen groups M, N, O | "O - looks good / N - looks good / M - looks good" |
| D-78 | Radar resampling (group P) | "looks good - actually like it. Agree with your logic approach (better to overstate than under)" |
| D-79 | Sixteen colours (group Q) | "Q - looks good" |
| D-80 | Milestones, estimates, reviewers | "Milestones good / estimates mean nothing - I want it done right […] Red team approved" |
| D-81 | The allow-list, corrected | "A" |
| D-82 | Label language | "A" |
| D-83 | The cell-colour vote | "24b looks better for general purpose - this is not a navigation tool so road accuracy is less important the greography and topogrphy" |
| D-84 ◇ | No decode limiter | "B" |
| D-85 ◇ | Lean memory caps | "B - this makes sense for now - we can adjust if we find we have more of a memory requirement, but in this area lean first makes more sense to me." |
| D-86 ◇ | The end of a borrow | "B as shaped above" |
| D-87 | Water and overlays | "A" |
| D-88 | Colour-vision threshold | "A" |
| D-89 | Group S; a second look at I, J, L | "S looks fine - 22d as well / I looks fine / J looks fine" · "L looks fine as well" |
| D-90 | When live views need more than a cache's cap | "A" |
| D-91 | The temperature preset beside a light ground | "A" |
| D-92 ◇ | A very large borrowed shape, replaced | "C" |
| D-93 | The two light-ground corrections (group T) | "22f is fine / 21d is fine" |
| D-94 | The two MAPSCII additions carried since DISCOVER | "A" |

## Critical analysis (red-team)

> red-team: round 1 NO-GO · round 2 NO-GO, narrowly · round 3 SHIP-WITH-CONDITIONS, remediated · multi-agent · scope: PLAN's architecture, contract, constants, diagrams, specimens, plan; rounds 2 and 3 on the remediation only · personas: code quality, performance, security, accessibility, newcomer, phase lens, business, docs quality, project hygiene (confirmed by HUM LEAD, D-38, D-80)

Full record, every finding with its check and disposition: [red-team-plan.md](red-team-plan.md).

| Round | Tree | Reviewers | Verdict | Findings | Critical | Fixes from the round before that did not hold | Questions ruled |
|---|---|---|---|---|---|---|---|
| 1 | `9d4cbd0` | 5 | NO-GO | 74 | 6 | — | 8 (D-81 to D-88) |
| 2 | `a96d09c` | 3 | NO-GO, narrowly | 60 | 2 | 13 | 5 (D-90 to D-94) |
| 3 | `5572fb5` | 2 | Ship with conditions | 27 | 0 | 8 | 0 |

The Critical findings, and what became of each:

| # | Round | Finding | Persona | Outcome |
|---|---|---|---|---|
| PL-CQ-1 | 1 | The dependency allow-list as ruled could not be met | Code quality | D-81 |
| PL-CQ-2 | 1 | The end of a borrow was undefined | Code quality | D-86, then D-92 |
| PL-PF-1 | 1 | The memory measurement was not the requirement's condition | Performance | Re-run; D-85; RS-7 held at High |
| PL-PF-2 | 1 | Numbers tests need were absent | Performance | constants.md |
| PL-AX-1 | 1 | The radar ramp vanished on a light ground | Accessibility | A second ramp; D-88; corrected again in round 2 (22f, D-93) |
| PL-PM-1 | 1 | Artefacts DISCOVER had promised were missing | Phase lens | The contract, constants, fixture, specimens 24 and 25 |
| P2-ENG-1 | 2 | Nothing said what a cache may evict; two maps would fetch without end | Performance | D-90 |
| P2-PRD-1 | 2 | The temperature preset fails D-88 on a light ground | Accessibility | D-91, D-93 |

**What the rounds say about the coordinator's work**, recorded in full in the red-team file: a fact given with a ruling and never checked (D-75); frozen parity rows designed against without reading them, twice; additions written into rulings' records without saying they were not ruled, three times; a colour checker that tested neighbours only, then differences and never order; a measurement reported without saying which fixture it was; task rows that test things their package cannot see, three rounds running. Each was found by a reviewer, verified, fixed and recorded. None changed a ruling's substance; two changed what HUM LEAD had been told, and the records say so.

**No fourth round is proposed.** HUM LEAD may order one.

## What approving this plan accepts

Six matters that are the coordinator's, not rulings. Approving the plan accepts them; HUM LEAD can strike any one and the plan changes to match.

| # | Matter | Where |
|---|---|---|
| 1 | Only `Work` and `Settle` report the ids whose borrow they ended. D-86's record says "`Work` or `Render`"; since owner calls are one at a time, a `Render` can never be the last reader | Contract, section 4 |
| 2 | `Set` builds the run index for overlays over **30,303 vertices** at the default cap, in runs of **64** | Constants, section 3 |
| 3 | The memory peak is stated at a pump **two** `Work` calls wide. D-84's option said "a named pump width" | NFR-3, constants |
| 4 | One temperature colour was moved by an amount below what the eye can tell, after the scale was approved (D-89, D-93), so that truecolor is ordered under every simulated kind of colour vision. At 256 colours three neighbouring pairs still invert slightly; recorded, not fixed | Specimen 21's ramp file, S21-8 |
| 5 | The alert preset's colours are designed in BUILD, with HUM LEAD's look inside M-B | RS-26; tasks 08.10, 08.23 |
| 6 | Twenty-two requirement rows that the Discovery Report ratified were revised in PLAN, each by a ruling or a red-team finding, each change stated inside its row: FR-7, -9, -11, -12, -15, -16, -17, -19, -20, -23, -24, -25, -27, -29, -30, -31, -34; NFR-3, -10, -19, -20, -22 | requirements.md, status line |

## Still owed by HUM LEAD

None blocks this approval; all are wanted before reference frames are frozen in BUILD (task 14.16).

1. A look at the colour specimens — the `.ans` files — in HUM LEAD's own terminal. The review page shows them as a browser draws them.
2. The name of that terminal and its font, for the record of what M1 was judged on.
3. In BUILD: a look at the alert preset's colours, all five severities on both grounds (task 08.23, inside M-B).

## Source documents

| Document | Location | Status |
|---|---|---|
| Requirements | `01-objectives/requirements.md` | found — revised rows marked |
| Objectives · M1 scenarios · glossary | `01-objectives/` | found |
| Rulings · options as presented | `02-analysis/rulings-discover.md`, `rulings-options-as-presented.md` | found — D-11 to D-94 |
| Parity matrix · defect ledger | `02-analysis/` | found — frozen |
| Risk assessment | `02-analysis/risk-assessment.md` | found |
| Specimens and findings | `02-analysis/specimens/` | found — 17 to 25 added in PLAN, with 21d, 21e and 22f after round 2 |
| Entry checks · approach notes | `03-architecture-design/plan-entry-checks.md`, `approach-1` to `-3` | found |
| Architecture · tour · contract · constants | `03-architecture-design/` | found |
| Level 2 and 3 diagrams | `03-architecture-design/L2-*.md`, `L3-*.md` | found — 13 files |
| Memory measurement | `03-architecture-design/memory-measurement.md` | found |
| Implementation plan · parity mapping · first-host start | `04-development/` | found |
| Integration-review template (D-60) | `07-readiness/integration-review-template.md` | found — to be filled in after the integration |
| Test fixture | `testdata/fixture/` | found — hashes pinned |
| PLAN red-team record | `08-reports/red-team-plan.md` | found |
| Discovery Report | `08-reports/discovery-report.md` | found — approved |

## Next steps

1. HUM LEAD approves this plan, or sends it back.
2. On approval: commit; merge `feature/go-tuimaps` into `release/v0.1.0` without fast-forward; nothing is pushed.
3. **PHASE TRANSITION PLAN → BUILD.** At BUILD's door: Go is restored to the project's language list (approved in D-20), and the first code-quality check runs on the scaffold.
4. BUILD starts at WP-00, test-first, and stops at M-A for HUM LEAD's look at the first working map.
