---
title: "go-tuiMaps — DISCOVERY REPORT"
date: 2026-09-19
phase: DISCOVER
report_template: discovery-report
level: LEVEL-1
sev: SEV-0
authority: HUM LEAD
directives: FULL GIT; FULL DOCS; FULL REPORTS; FULL DIAGRAMS; FULL RCC; FULL PLAN; FULL TDD
status: "APPROVED — HUM LEAD 2026-09-19 (\"APPROVED 4 PLAN;\"), with directives for PLAN recorded as D-71. DISCOVER exited; PLAN opened."
---

# go-tuiMaps — DISCOVERY REPORT

## Bottom Line Up Front

**Recommendation: proceed to PLAN.** The idea works, the scope is ruled, and what is left is design inside fixed budgets.

- **What this is.** A Go library that draws a real map in a terminal and lets a host application lay weather on it — alert areas, radar, temperature — so a person can see where a hazard is relative to their place without leaving the terminal. It is a port of TerminalMap (Rust), with overlays added, first used by Watchpost.
- **What DISCOVER settled.** 60 rulings by HUM LEAD before this report (D-11 to D-70; its approval is D-71), each put one at a time and recorded in HUM LEAD's words; nine went against the coordinator's recommendation, and one (D-55) HUM LEAD later withdrew. 66 requirements. A frozen parity matrix of 77 rows, 75 counted. A closed ledger of 17 upstream defects. 25 risks. 57 rendered specimen frames, every group seen by HUM LEAD.
- **What the evidence shows.** A weather field drawn as cell background under braille map lines is legible at both of Watchpost's map sizes. HUM LEAD has seen it and ruled on it. Nothing found in three rounds of adversarial review says the approach does not work.
- **The shape of the library.** A map library first, a weather library second (D-69): for any overlay it takes a type, its values, its colours and its data from the host, and it ships four common types ready to use — temperature, radar and precipitation, alert areas, wind — fully defined and coloured, each of which a host may override. "It just works" includes being easy for a developer.
- **What the first release is (v0.1.0).** The braille basemap; network, disk-cache and embedded tiles; three overlay shapes — features, images, scalar grids; legend, credits, valid time, scale; the view described as plain data for screen readers and speech; truecolor, 256 colours and no colour; a small app. Then it is integrated into Watchpost at its v0.17.0, a written integration review comes back to HUM LEAD, and the remaining shapes are built.
- **What it will cost.** A reviewer's estimate comes to about 9,400 hand-written lines for all of v1, of which v0.1.0 is about three-quarters. (The risk assessment says "roughly two-thirds"; the table in this report, which adds the rows up and includes what was ruled into the first release since, gives more.) That is an estimate by subsystem and cannot be verified; there is no effort estimate, and PLAN must produce one.
- **What is still risky.** Four High risks, all about PLAN getting the design right: the overlay contract, who runs background work, memory inside Watchpost's 10.7 MB margin, and Watchpost's own readiness to supply overlay data.
- **How it was checked.** Three red-team rounds. Round 1: NO-GO, 75 findings. Round 2: NO-GO on accessibility, 87 findings, security's NO-GO lifted. Round 3, narrow: ship-with-conditions from all three reviewers, no Critical, 44 items. Critical findings by round: 8, 5, 0. Every finding has a disposition; none was dropped.

## The problem, locked

> "People who follow weather and hazards from a terminal cannot see where conditions are relative to the places they care about — storms, fronts, wind and precipitation reach them only as text and numbers — and so they leave the terminal to find out whether something is coming toward them."

Locked in the [project brief](project-brief.md) (5 of 5 on the problem-statement checks). Every requirement below traces to it.

## Options considered, including doing nothing

Round 1 asked for this table and the coordinator marked it "Fix" without writing it; round 2 caught that. Here it is. Only the last row was researched in depth; the others were ruled out by a constraint already on record, which is named.

| Option | What it would mean | Why not |
|---|---|---|
| Do nothing | Watchpost users keep leaving the terminal to place a hazard. Watchpost's map roadmap stays blocked. | It is the problem statement. |
| Call the Rust library from Go | Link TerminalMap through a C bridge. | Watchpost builds with the C toolchain off, and HUM LEAD made pure Go a hard rule (D-22). |
| Run an existing terminal map as a separate program and capture its output | Spawn MAPSCII or TerminalMap and paint its screen into a rectangle. | No overlays, which is the whole requirement (R-3); not embeddable in the host's own draw cycle (R-2); fails host independence (M5). |
| Show a picture using a terminal image protocol | Render a bitmap map and send it as sixel or similar. | Outside Watchpost's stated glyph floor, unsafe over SSH, and nothing in the terminal could read it — no text, no colour fallback (D-23). |
| **Port to Go, with overlays designed in** | This project. | Chosen. HUM LEAD's intent at intake; confirmed feasible by the specimens. |

## What DISCOVER produced

| Area | Document | State |
|---|---|---|
| Requirements | [`requirements.md`](../01-objectives/requirements.md) | 44 functional, 22 non-functional, 5 constraints, 6 assumptions; release tables for v0.1.0 and later |
| Objectives and metrics | [`objectives.md`](../01-objectives/objectives.md), [`m1-scenarios.md`](../01-objectives/m1-scenarios.md) | M1–M6 with ruled targets; seven M1 scenarios |
| Words | [`glossary.md`](../01-objectives/glossary.md) | Plain-language terms, and a key to every ID prefix |
| Rulings | [`rulings-discover.md`](../02-analysis/rulings-discover.md), [`rulings-options-as-presented.md`](../02-analysis/rulings-options-as-presented.md) | D-11 to D-70 verbatim; the options each letter stood for |
| Parity | [`parity-matrix.md`](../02-analysis/parity-matrix.md), [`defect-ledger.md`](../02-analysis/defect-ledger.md) | 77 rows, 75 counted; 17 defects, closed (D-37) |
| Risks | [`risk-assessment.md`](../02-analysis/risk-assessment.md) | 25 risks |
| Research | [`research/`](../02-analysis/research/) AI-1 to AI-9, [`tier1-synthesis.md`](../02-analysis/tier1-synthesis.md) | Filed as returned, each with a verification header |
| Specimens | [`specimens/`](../02-analysis/specimens/README.md) | 57 rendered frames with findings; made by throwaway code that is not in this repository |
| Review | [`red-team-discover.md`](red-team-discover.md) | Three rounds |

## Requirements

All 66 are in v1. The table says what each area commits to; the full, testable wording and each row's source are in the requirements file.

| Area | Requirements | The commitment |
|---|---|---|
| Basemap | FR-1, FR-2, FR-3, FR-19, FR-20, FR-35, FR-36 | Behavioural parity with TerminalMap at a pinned commit, defects fixed or knowingly kept; rivers, parks and airports drawn; braille by default, block opt-in; own styles; a painted ground by default (D-64) |
| Overlays | FR-6 to FR-14, FR-37 | Five input shapes — features, scalar grids, vector grids, images, tile-image providers. The host fetches, the library draws (D-15). Images are re-coloured through a required colour table (D-36, D-45). Large shapes are borrowed, not copied, and simplified off the drawing path (D-16) |
| Safety of what is shown | FR-29, FR-32, FR-33, FR-34, FR-18a | The view described as data for every place the host names (D-52); valid time and staleness; a distance reference; text made safe wherever it leaves the library |
| Colour and theme | FR-15 to FR-18 | A palette of named semantic tokens (D-63); ramps ordered, distinct, readable and colour-vision-safe by default (D-53); presets, not locks — four ready-made overlay types a host may override, temperature's being an absolute scale anchored at freezing (D-62, D-69); a safe-ramps setting for the person at the keyboard (D-63); a no-colour form chosen by kind of data (D-35) |
| Tiles and network | FR-21 to FR-23, FR-28, FR-31 and their lettered rows | Replaceable tile source; the library connects to nothing until told (D-65); disk cache off unless configured; embedded zoom 0–3 tiles (D-33) made by an in-repository generator with a minimal hardened reader (D-58); never a blank frame (D-30) |
| Embedding and control | FR-4, FR-24 to FR-27, FR-30 | Render on demand into a rectangle with no terminal attached; control by intents; time from the host; explicit, bounded, leak-free background work |
| App | FR-5 | Upstream's keys and footer, pointer operations each with a keyboard equivalent (D-17), a headless flag, a describe mode |
| Non-functional | NFR-1 to NFR-22 | Pure Go (D-22); 8 MB target (D-29, D-48); never blank, ≤ 1 s warm, ≤ 3 s cold (D-30); deterministic frames on two architectures; untrusted-input limits with numbers, every decoder fuzzed; accessibility and motion (D-56); tests first; compatibility before v1.0 (D-60) |

**Brief → requirement trace.** R-1 parity → FR-1, FR-2, FR-26, FR-27. R-2 embeddable → FR-4, FR-24, FR-25, FR-27, NFR-1 to NFR-6. R-3 overlays from other sources → FR-6 to FR-19. R-4 high-volume host data → FR-11.

## Metrics of success

| Metric | Target as ruled | Where it stands |
|---|---|---|
| M1 Hazard Placement | All seven scenarios, judged by HUM LEAD against an answer key computed by an independent script; gates SHIP on the library (D-43). **Two parts, both must pass (D-67):** the frame alone, judged to its own resolution — under one cell from an edge, "on the edge" is the right reading; and the text description, exact. Scenario 4 accepts "no visible change" on a flat day (D-68). In v0.1.0: the scenarios whose overlay shape is in the slice | Scenarios 1, 3, 4 and 5 have renderings; **2, 6 and 7 have none yet** — owed before PLAN exit |
| M2 Time to Placed View | ≤ 1 s warm, ≤ 3 s cold to full detail, never blank (D-30) | Test and link model pinned (NFR-5); unmeasured |
| M3 Parity Coverage | 100% of 75 rows, stated per release: 62 in v0.1.0 (D-49, D-61) | Matrix frozen; two rows excluded by named ruling (P-40, P-46) |
| M4 Embed Cost | 8 MB target; live heap ≤ 4 MB and peak ≤ 8 MB on a pinned typical-day fixture; worst case tested separately (D-29, D-48) | A reviewer's arithmetic puts the fixture at about 3.6 MB, with 10% slack, *only if* unused tile data is dropped at decode and coordinates are stored compactly. Measured at PLAN exit |
| M5 Host Independence | Pass | Requirements FR-4, FR-27, FR-30 |
| M6 Correction Count | Lower is better. Intake: 2 | **DISCOVER: 6**, in four messages — see below |

**M6 tally for DISCOVER.** (1) "Present questions requiring my ruling 1 by 1" — the coordinator had been batching decisions. (2–4) D-39: three rulings whose recorded implication went beyond HUM LEAD's words (D-14, D-34, D-36). (5) D-40: "I have not seen any other than the one you showed me" — the record had claimed HUM LEAD had seen specimens HUM LEAD had not. (6) D-69: "we might be making this too hard" — asked to confirm a detail of a locked temperature scale, HUM LEAD replaced the lock with presets, a simpler design the coordinator had not offered. Separately, reviewers — not HUM LEAD — found 8 coordinator errors in round 1, 6 in round 2 and 6 in round 3, each listed in the red-team record. The repeating pattern is a claim written without looking: a legibility claim, a marker said to be readable in a frame that had none, counts in commit messages, a fix said to be made that was not. The second pattern is recording more than was said: consent taken from silence (twice), a metric's acceptance loosened without a ruling.

## Constraints and dependencies

- **Technical.** Go no newer than Watchpost's (1.25). No C toolchain. macOS, Linux, Windows. From a 69×12-cell rectangle up. Nothing that depends on querying the terminal.
- **The first host must change (CD-1, RS-11 — High).** Today Watchpost holds points, circles and one wind reading per place, and reduces alert polygons to one vertex. To feed the v0.1.0 shapes it must keep alert geometry and fetch zone shapes, add a public radar source with a colour table, and add a gridded temperature source. HUM LEAD has confirmed it can add these (D-39). This is a named companion backlog for Watchpost, outside this repository. D-57 adds one line to it: Watchpost's README states braille as a need for maps.
- **Services.** OpenFreeMap is free, without guarantee, and has said its schema may change; the tile source and the schema are both replaceable parts (FR-21, FR-35, D-46).
- **Git.** `main ← release/v0.1.0 ← feature/*`; local merges until the first release, pull requests after (D-28). No remote until SHIP (D-19). Sole-author commits; no tool-generated trailers or watermarks.
- **Timeline.** None has been set. There is no effort estimate; PLAN produces the first.

## Cost

One figure exists, and it is a reviewer's estimate by subsystem, unverifiable, given to HUM LEAD with D-44 in three rounded forms that this table reconciles.

| Part | Rough hand-written lines | In v0.1.0? |
|---|---|---|
| Basemap core: tile decoding, styles, drawing, labels, view | 2,450 | Yes |
| Tile sources: network, cache, embedded, fetcher safety | 900 | Yes |
| Overlay model, features, simplification, compositing | 1,200 | Yes |
| Images re-coloured through a table | 350 | Yes |
| Legend, credits, valid time, scale, text safety, validation | 450 | Yes |
| Scalar grids, ramps, contouring | 650 | Yes (D-44 option B) |
| Ramp validation, colour-vision checks, per-depth ramps | 300 | Mostly, taken as 250 |
| Markers and camera | 500 | Markers only, taken as 200 |
| Minimal archive reader for the generator (D-58) | 250–300 | Yes — added in round 2 |
| Vector grids (wind) | 150 | No |
| Tile-image providers | 200 | No |
| The rest of the PMTiles source | about 150–200 | No |
| Block renderer, 16-colour ramps | part of 600 | No |
| Standalone app | 1,200 | A minimal one, taken as half |
| **All of v1** | **about 9,400** (the rows above, mid-points) | **v0.1.0: about 6,700 without the app, about 7,300 with a minimal one — roughly three-quarters** |

Not in the estimate, because they were ruled after it: the description as data (D-52, "a few hundred lines"), the semantic-token contract and safe-ramps setting (D-63), the painted ground (D-64), the contract check (D-60).

## Risk assessment

25 risks: **High 4 · Medium 13 · Low 7 · Closed 1.** Full table in the risk assessment; the High ones and the three added in rounds 2 and 3:

| # | Risk | Now | Status | Mitigation |
|---|---|---|---|---|
| RS-2 | The overlay contract is designed before any real use, for five shapes, and Watchpost will import it | High | ACTIVE | PLAN designs for all five and builds three; compatibility promise (D-60); a written integration review before the rest is built (D-60) |
| RS-4 | Who runs background work — fetching, decoding, simplifying — inside a host that redraws from one clock | High | ACTIVE | FR-30 states the constraints; PLAN compares at least two models in a named design note |
| RS-7 | Memory: 8 MB inside a 10.7 MB margin | High | ACTIVE | Pinned fixture and metric (NFR-3); measured at PLAN exit; the target is revised only by ruling |
| RS-11 | Watchpost cannot yet supply four of five overlay shapes | High | ACTIVE | Companion backlog; M1 judged on host-shaped test data (D-43) |
| RS-23 | The first release has no demonstrated way to get radar in: every radar specimen fetched tiles, a deferred shape | Medium | ACTIVE | Confirm a keyless bounding-box image source, with its terms, at PLAN entry; if none exists, the slice returns to HUM LEAD |
| RS-24 | Light-background terminals | Medium | MITIGATED | Painted ground by default (D-64); specimens on light and dark grounds in PLAN |
| RS-25 | A font without braille gives no map and no warning: the first release is braille-only (D-57) and the terminal cannot be asked | Medium | ACTIVE | Braille coverage joins the glyph matrix at PLAN entry; the app's help and the README state the need and point to the text description |
| RS-3 | A field under braille may not be legible | Low (braille) | MITIGATED | HUM LEAD has seen every specimen group — in a browser. **A look at the specimen files in HUM LEAD's own terminal is asked for before PLAN exit** |

## Open questions

All eleven of the brief's open questions are ANSWERED by rulings D-11 to D-37. All thirteen questions from red-team round 1 are ANSWERED (D-40 to D-56), all ten from round 2 (D-57 to D-66) and all three from round 3 (D-67 to D-70). What remains is design, carried to PLAN with an owner:

| # | Question | Status | Next step |
|---|---|---|---|
| PQ-1 | Which model runs background work | STILL OPEN | PLAN: a design note comparing at least two against FR-30 |
| PQ-2 | Are block, box-drawing, shade and arrow characters one column wide in the supported terminals; do fonts carry braille | STILL OPEN | A terminal matrix at PLAN **entry** (NFR-8, D-57) |
| PQ-3 | Is there a keyless source that returns one radar image for one bounding box, with terms that allow an example table | STILL OPEN | PLAN entry (RS-23) |
| PQ-4 | Does 8 MB hold | STILL OPEN | Measured at PLAN exit against the pinned fixture |
| PQ-5 | Which established temperature scale the temperature preset uses, passing the colour-vision rule at every depth | STILL OPEN | A specimen with checker results in PLAN; if none passes, back to HUM LEAD (D-62) |
| PQ-6 | A passing radar ramp at 256 colours; a 16-colour hint specimen; a painted-ground specimen; clean contours; the radar resampling rule | STILL OPEN | Specimens in PLAN (S15-3, D-59, D-64, A-5) |
| PQ-7 | The keyboard mechanism for zoom-around-a-target; the dependency allow-list; the module layout | STILL OPEN, by design | PLAN (D-17, NFR-9) |

## Critical analysis

| Round | Reviewers · lenses | Verdict | Findings | Outcome |
|---|---|---|---|---|
| 1 | 9 lenses, personas confirmed by HUM LEAD (D-38) | **NO-GO** (security; no control on terminal escape bytes) | 75; 8 Critical after re-assessment; 12 convergences | Requirements rewritten to be testable; 13 questions ruled (D-40 to D-56) |
| 2 | 5 reviewers · 9 lenses, tree frozen | **NO-GO** on accessibility; security's NO-GO lifted; the rest ship-with-conditions | 87; 5 Critical, all fixed; 1 declined in part | 10 questions ruled (D-57 to D-66) |
| 3 | 3 reviewers, narrow — round 2's fixes and rulings only, tree frozen | Ship-with-conditions from all three; **no Critical** | 44: 11 round-2 fixes that held only in part, 33 new | 3 questions ruled (D-67 to D-70); a fourth round not proposed — the find-rate has converged and two reviewers judged it would review its own fixes |

No finding in any round says the idea does not work. Most round-2 findings were introduced by round-1 fixes — the cost of a wide, fast remediation. Full record, with every disposition: [`red-team-discover.md`](red-team-discover.md).

## Decisions for ratification with this report

Approving this report also ratifies these, each already visible in the documents:

1. **The brief's amendments** D-10a, D-10b and D-10c (v1.0.1 to v1.0.3): two factual corrections, neutral wording, in-place notes where rulings replaced a metric's wording.
2. **The parity matrix as frozen**: 77 rows, 75 counted, with **P-40 and P-46 excluded by named ruling** (D-24; D-29 and D-49).
3. **One ruling withdrawn:** D-55 (temperature not themeable), by D-69, marked in place.
4. **Extension rows** for requirements added after the matrix was frozen, added on approval as E-18 to E-28: the view described as data (D-52); valid time (FR-32); distance reference (FR-33); layers switched by intent (D-42); image loops (D-47); reduce-motion (D-56); focus intents (FR-24a); the ramp checker and safe-ramps setting (D-53, D-63); the semantic-token palette (D-63); overlay presets, the temperature preset's absolute scale among them (D-62, D-69); the painted ground (D-64).
5. **A post-SHIP check**, which round 1 asked for and the coordinator failed to write: the written integration review of D-60 doubles as it, and reports M1 judged live in the host, M4 measured in the real host, and the correction count.

## Gates

| Gate | Result |
|---|---|
| Framework structure check | 18 of 18; 2 checks skipped with a declared reason (no code yet; the language declaration returns at BUILD entry, D-20) |
| Counts recomputed from the rows | 44 + 22 requirements · 60 rulings, 54 with options rows and 6 recorded as not put as options · 25 risks · parity 48 · 13 · 2 · 12 · 2, 75 counted, 62 in v0.1.0 — also recounted independently by a round-3 reviewer |
| Tables | Every row in every tracked document has its header's cell count |
| Fidelity | The five parity rows kept whole are character-identical to the research report (round-3 reviewer) |
| Commits | 62 before this report. Sole-author commits; no tool-generated trailers or watermarks |
| Tracked content | 19 documents, 44 + 57 specimen files, 11 images, the module file and two dot-files. No code |
| Repository | No remote, no stash, clean tree, nothing unreachable, integrity check clean; `main` and `release/v0.1.0` at the root commit, awaiting the first merge |
| Tree freeze | Held through rounds 2 and 3. Broken once in round 1, recorded there |

## Recommendation

**Proceed to PLAN.**

**PLAN should focus on**

1. The overlay contract for all five shapes, image loops, presets and the description as data — the thing Watchpost imports and D-60 now promises to keep stable. **Designed for the developer's ease first (D-69):** a preset in one call, an override in one more.
2. The background-work model, compared in a design note against FR-30.
3. At entry: the terminal glyph matrix; a confirmed radar image source.
4. The pinned fixture, then the memory and allocation measurement at exit.
5. Ramps: an established temperature scale that passes; a radar ramp that passes at 256 colours; the semantic-token list.
6. The owed specimens: scenarios 2, 6 and 7; a 16-colour hint; painted light and dark grounds; contours.
7. An effort estimate, which does not exist.
8. At BUILD entry: restore the language declaration in the project configuration (D-20) and run the first code-quality check.

**PLAN should NOT revisit** — these are ruled: what parity means (D-11) and the defect ledger (D-37); the three deliverables and the MIT licence (D-13); five overlay shapes in v1 (D-14); the host fetches (D-15); pure Go (D-22); braille by default, block opt-in (D-42); the first-release slice and the sequence after it (D-44, D-57); re-coloured images with a required table (D-36, D-45); the no-colour form by kind of data (D-35); presets, not locks, with the temperature preset's absolute scale (D-62, D-69); M1 in two parts (D-67) and flat days accepted (D-68); the description as data in v0.1.0 (D-52); the painted ground (D-64); no network unless told (D-65); the compatibility promise and the integration review (D-60).

## Next steps

1. HUM LEAD approves this report, or sends it back. **Approved 2026-09-19 (D-71)**, with two directives for PLAN: no large embedded code blocks — signatures and short examples only; and a diagram for every major part of the architecture, complete, comprehensive, at more than one level of detail, kept as living references.
2. On approval: the extension rows are added; this report is committed; `feature/go-tuimaps` is merged into `release/v0.1.0` with a merge commit — the first merge in the repository; the phase moves to PLAN.
3. Asked of HUM LEAD before PLAN exit, not blocking this approval: a second look at corrected specimen groups I, J and L; a look at the specimen files in HUM LEAD's own terminal.
4. Upstream: offer the defect findings to TerminalMap's author. Two MAPSCII additions were never given a disposition and are carried to PLAN.

## Source documents

Listed in "What DISCOVER produced". Baselines: TerminalMap `3b960723b24b781a1a9a04c165803cd1e98e6e3c` (v0.1.0); MAPSCII `4fe9a60a0c9da952dadc5214a9ca5c68c447fdf8`.
