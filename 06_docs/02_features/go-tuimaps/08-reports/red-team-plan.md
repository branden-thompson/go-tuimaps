# PLAN — Red-team record

| Field | Value |
|---|---|
| Phase | PLAN |
| Date | 2026-09-19 |
| Tree reviewed | `9d4cbd0`, frozen for the whole round — no tracked file changed while reviewers ran |
| Reviewers | Five, fresh, covering the nine lenses HUM LEAD confirmed (D-38, D-80): code quality · performance and security · accessibility and newcomer · phase lens and business · docs and hygiene |
| Subject | The architecture and its diagrams, the three approach notes, the implementation plan, the PLAN specimens (17 to 23), the memory measurement |
| Status | Round 1 recorded. **Round verdict: NO-GO as written.** Remediation and rulings follow; a second round is required. |

## Verdicts

| Lens | Verdict | What drove it |
|---|---|---|
| Code quality | **NO-GO** | The dependency allow-list cannot be met as ruled; the end of a borrow is undefined, and Render was not counted as a reader |
| Performance | **NO-GO** | The memory measurement was not NFR-3's condition; numbers owed at PLAN exit are absent |
| Security | SHIP-WITH-CONDITIONS | Untrusted inputs with no limit test or fuzz target; a tile extent that does not fit the coordinate type |
| Accessibility | **NO-GO** | The radar ramp fails on a light ground: its heaviest class is invisible |
| Newcomer | SHIP-WITH-CONDITIONS | The pump is never shown; "hard to ignore" is not delivered |
| Phase lens | **NO-GO** | Artefacts DISCOVER named as "not a loose promise" were neither produced nor deferred by ruling |
| Business | SHIP-WITH-CONDITIONS | Two frozen parity rows contradicted by the design; parity tested last |
| Docs | SHIP-WITH-CONDITIONS | Diagrams that disagree with each other; major parts with no diagram |
| Hygiene | **GO** | Clean |

No finding says the design is wrong. Every NO-GO is a gap in PLAN's own work: a number not produced, a rule not stated, a ruling relied on without checking its facts.

## How findings are handled

As in DISCOVER. **Verified** = re-checked by the coordinator; **accepted** = the reviewer's evidence read and found sound. **Fix** = to be done in the remediation that follows this record, and listed there when done — this record does not call anything fixed that is not. **Question** = put to HUM LEAD, one at a time.

## Findings and dispositions

### Code quality

| ID | Finding | Severity | Check | Disposition |
|---|---|---|---|---|
| PL-CQ-1 | The two-module allow-list of D-75 cannot be met: the first host pins a width library whose releases since v0.0.19 require a different segmentation module, not the one D-75 names | Critical | **Verified** from the host's module file and the local module cache | **Question PL-Q1.** *The coordinator told HUM LEAD the width library imports the named module; that was true of v0.0.16 and is not true of the version the host uses* |
| PL-CQ-2 | The end-of-borrow signal has no type, no stated firer, and the state machine counts jobs but not Render, which also reads borrowed memory; `Remove` returns nothing; nothing fires if the host stops pumping | Critical | Verified | **Question PL-Q6**, then Fix |
| PL-CQ-3 | No concurrency contract anywhere, though NFR-20 requires one; the lifetime of a reused frame is unstated; task 09.20 cannot fail where it sits | Important | Verified | Fix: a table of which calls are safe together, in the contract document |
| PL-CQ-4 | A test that "must fail under the race detector" cannot live inside a green suite | Important | accepted | Fix: run as a guarded sub-process; add the opposite test |
| PL-CQ-5 | Separate modules: an untracked workspace file, no `replace`, no remote — a fresh clone cannot build them; nested tags and per-module tests unplanned | Important | accepted | Fix |
| PL-CQ-6 | No remote until SHIP, yet the plan asks for hosted runners and a nightly job | Important | Verified | Fix: a local gate script; say what is untested until SHIP |
| PL-CQ-7 | The dependency diagram and milestone M-A are wrong; no package owns the types that tiles and overlays produce and render and describe consume | Important | Verified | Fix: an early `scene` package; corrected edges and milestones |
| PL-CQ-8 | Parity row P-36 (label-language lookup, Match) contradicts dropping every place-name translation | Important | **Verified** | **Question PL-Q2** |
| PL-CQ-9 | A goroutine-count helper gives false failures when the network is used and cannot see timers | Important | accepted | Fix: prove D-73 by a static check — no `go` statement or timer in library packages — plus a filtered leak check |
| PL-CQ-10 | Allocation assertions as written will not work under the race detector | Important | accepted | Fix: separate files and a gate leg without it |
| PL-CQ-11 | The decode limiter's position and number are unspecified; how `Settle` treats work in flight elsewhere is undefined; a shared fetch has no goroutine to live on under D-73 | Important | accepted | With PL-Q4; then Fix |
| PL-CQ-12 | Numbers owed "in PLAN" are missing, so tests cannot be written first | Important | Verified | Fix: a constants table before BUILD |
| PL-CQ-13 | Figures from throwaway code used as specifications (03.12, 10.7, 04.11, 11.1) | Important | Verified | Fix: one-sided bounds from the requirement; define the rule, then derive the count |
| PL-CQ-14 | Of 25 requirements sampled, 13 groups have no task — TileJSON, style limits, image fuzz, built-in styles, the generator's encoder, how the assets package registers, and others | Important | accepted | Fix: add the tasks |
| PL-CQ-15 | Smaller test and scope defects, including a task that builds a deferred feature and a wrong distance in a diagram | Minor | Verified in part | Fix |

### Performance and security

| ID | Finding | Severity | Check | Disposition |
|---|---|---|---|---|
| PL-PF-1 | The measurement was one view, not NFR-3's filled caches and three instances; no default cap exists anywhere; by the documents' own rule three full-size instances exceed the live line | Critical | **Verified by a re-run**: a 1.25 MB tile cache filled, three instances — **3.56 MB live**. The peak line **cannot be shown** by a program that uses a general decoder; at the default collector setting peak runs to about twice live, which puts it at the line | **Question PL-Q5** (default caps; the line for three maps). RS-7 returns to High |
| PL-PF-2 | Numbers owed at PLAN exit are absent: allocation, timing, the fallback's time bound, per-tile maxima, FR-29's targets | Critical | Verified | Fix where PLAN can produce them; the rest named, with owner, in the constants table |
| PL-PF-3 | Under D-73 a one-wide pump makes the 3 s cold target tight (2.45 s by arithmetic); the decode limit has no number; no pump width is stated for the measurement | Important | accepted | Fix, with PL-Q4 |
| PL-PF-4 | Resampling is per view inside `Work`, so every pan invalidates it and images have no stand-in | Important | accepted | Fix: resample at draw time into a reused buffer |
| PL-PF-5 | The unchanged-frame key omits staleness, status, focus, layers and safe ramps; the frame type is not fixed; string lines would cost 232 KB a frame against a 16 KB budget | Important | accepted | Fix |
| PL-PF-6 | Three benchmark tasks cannot run deterministically | Important | accepted | Fix |
| PL-PF-7 | The fixture's tiles are light; no dense urban tile | Minor | accepted | Fix |
| PL-IS-1 | Untrusted inputs with no limit-before-allocation test and no fuzz target: style file, TileJSON, PNG, cached files, host structs, archive metadata | Important | accepted | Fix |
| PL-IS-2 | A tile extent of 65,536 does not fit a 16-bit coordinate; protobuf pitfalls unnamed; the oracle sees only good tiles | Important | Verified (arithmetic) | Fix |
| PL-IS-3 | Separate modules escape the allow-list, the scan and the loopback rule | Important | accepted | Fix |
| PL-IS-4 | Check-then-use on the cache root; no symlink test | Important | accepted | Fix: open the root once as a confined handle |
| PL-IS-5 | "Every string cleaned on the way out" is a convention, not enforced; foreign errors reachable by unwrapping | Important | accepted | Fix: a type that only the cleaning package can construct |
| PL-IS-6 | No test that the generator's output is byte-identical on a re-run | Minor | accepted | Fix |

### Accessibility and newcomer

| ID | Finding | Severity | Check | Disposition |
|---|---|---|---|---|
| PL-AX-1 | On the painted light ground the radar ramp's heaviest class is 1.01:1 and differs from the ground by 3.1; the checker has no rule about the ground | Critical | **Verified** — the coordinator's checker gives 3.1 | Fix: a darker-is-heavier ramp for light grounds (found: worst pair 16.0, 12.6 in the 256 palette, every class at least 18 from the ground); a ground rule in the checker; a specimen for HUM LEAD |
| PL-AX-2 | Two non-neighbouring warm temperature classes differ by 7.5 under deuteranopia; the checker tested neighbours only — the flaw used to fail the broadcast-style scale | Important | **Verified** — 7.5, and 8.4 in the 256 version | Fix: an all-pairs rule; the scale retuned by one small move in each version (all pairs now at least 12.2 and 11.1) |
| PL-AX-3 | "Speakable" is tested by string rules only | Important | Verified | Fix: a manual pass through a speech engine, on the release checklist |
| PL-AX-4 | "On the edge" exists only in the answer key; a sighted user gets no cue that the picture is under-resolved | Minor | Verified | Fix: the description data carries "closer than one cell" |
| PL-AX-5 | Reduce-motion, safe ramps and no colour are not all reachable by key and flag, nor documented in help | Important | Verified | Fix |
| PL-AX-6 | The 16-colour basemap tells kinds apart by colour alone; no reference-table task | Important | accepted | Fix |
| PL-AX-7 | Specimen claims with no task; looks still owed | Minor | Verified | Fix: an exit checklist |
| PL-NC-1 | The pump is never shown; how it learns that new work exists is unspecified; FR-30's callback is in no diagram or task | Important | Verified | Fix |
| PL-NC-2 | Calling `Work` on the interface's goroutine blocks it; never calling it leaves the map unsharpened — neither is diagnosed | Important | accepted | Fix: a warning after renders with pending work and no `Work` |
| PL-NC-3 | "Hard to ignore" is not delivered: the result of `Set` can be discarded, and the plan's own snippets discard it | Important | Verified | With PL-Q6 |
| PL-NC-4 | First-hour mistakes pass silently: a discarded error, a mistyped id on refresh, °F values declared as °C | Important | accepted | Fix |
| PL-NC-5 | Examples cannot be fetched or built as planned | Important | accepted | Fix, with PL-CQ-5 |
| PL-NC-6 | Sixteen items of the contract diagram have no public-surface task | Important | Verified | Fix |
| PL-NC-7 | `Settle` with no source reports "settled" and says nothing about why the map is empty | Minor | accepted | Fix |

### Phase lens and business

| ID | Finding | Severity | Check | Disposition |
|---|---|---|---|---|
| PL-PM-1 | Named PLAN artefacts absent and not deferred by ruling: the contract document, the token list, error kinds, zoom buckets, the colour-vision threshold, per-tile maxima, the contour specimen, the fallback's time figure, allocation and timing | Critical | Verified | **Fix: produce them.** HUM LEAD has ruled that the work takes the time it needs (D-80). The colour-vision threshold is **Question PL-Q8** |
| PL-PM-2 | The contract view omits markers and places, shared caches, `Describe` while pending, callbacks | Important | Verified | Fix |
| PL-PM-3 | The three-call path contradicts itself: tiles become wanted only when Render notes them, but `Settle` runs first | Important | Verified | Fix: the rectangle reaches the map before `Settle` |
| PL-PM-4 | D-73 not carried into NFR-5, FR-23 or the glossary; **the D-73 record adds a library-side decode limiter — more than "B", and against option B as presented — without labelling it as not ruled** | Important | **Verified** | **Question PL-Q4.** *The coordinator labelled its additions under D-74 and did not here* |
| PL-PM-5 | Unowned dependencies: runners, workspace file, a shaped secure test server, the fixture outside the repository, the asset not yet generated | Important | Verified | Fix; the fixture is committed |
| PL-PM-6 | Deferred shapes that do not yet fit: a provider function inside "plain data"; image loops with no frame-advance source and a byte cap that ten frames nearly fill; no cell-to-coordinate call for the pointer | Important | accepted | Fix: a sequence each, in the contract document |
| PL-PM-7 | Only a panicking job is tested; other public calls' panic path is unstated | Important | accepted | Fix |
| PL-PM-8 | Looks owed before PLAN exit are unrecorded: corrected groups I, J, L; colour files in a terminal | Important | Verified | Asked of HUM LEAD again, plainly |
| PL-PM-9 | The generator needs an encoder that appears nowhere; stale "awaits" text; a wrong risk number | Minor | Verified | Fix |
| PL-BZ-1 | Two frozen v0.1.0 rows contradicted by the design: P-36, and **P-08 — the matrix says a cell takes the majority colour of its lit dots; the diagram says the highest-priority line wins** | Important | **Verified** | **Questions PL-Q2, PL-Q3.** *The coordinator drew what the throwaway renderer does and did not check the matrix* |
| PL-BZ-2 | Parity is tested last, by one catch-all task; seven of twelve sampled rows have no task | Important | accepted | Fix: the mapping first; a task per row |
| PL-BZ-3 | M1a is HUM LEAD's reading in the app: no judging task, no app task loads scenario overlays, golden frames approved by nobody | Important | Verified | Fix |
| PL-BZ-4 | No local-override recipe for the host; the integration review of D-60 is one sentence; the host spike named in two risk rows has no task | Important | Verified | Fix |
| PL-BZ-5 | RS-7 lowered to Medium without justification | Important | Verified, and confirmed by the re-run | Fix: RS-7 returns to High |

### Docs and hygiene

| ID | Finding | Severity | Check | Disposition |
|---|---|---|---|---|
| PL-DQ-1 | Tile source order and stand-ins disagree between diagrams; nobody "wants" ancestor tiles | Important | Verified | Fix |
| PL-DQ-2 | In the overlay diagram a warning is a dead end; `Remove` has no path | Important | Verified | Fix |
| PL-DQ-3 | The foreground rule is stated three ways | Important | Verified | Fix: one rule — keep the line's colour where it passes on the cell, otherwise the higher-contrast of black and white |
| PL-DQ-4 | A wall clock separate from animation time, and shared caches, are required but in no contract diagram | Important | Verified | Fix |
| PL-DQ-5 | No tile state for "drawn, not cached"; an offline tile retries for ever | Important | Verified | Fix |
| PL-DQ-6 | Water is painted over images, which for radar hides rain over water — against "better to overstate", and un-ruled | Important | accepted | **Question PL-Q7** |
| PL-DQ-7 | Named artefacts moved into BUILD tasks with no pointer | Important | Verified | With PL-PM-1 |
| PL-DQ-8 | The fixture is pinned by description only; its shapes were fetched live | Important | Verified | Fix: the data is committed, with its sources and a hash list |
| PL-DQ-9 | Major parts with no diagram: errors and warnings, style and profile selection with the schema mapping, label placement, the app, tests and gates, zoom buckets, fit-to, shared caches | Important | Verified | Fix: the diagrams (D-71) |
| PL-DQ-10 | Milestone M-A omits packages it needs | Important | Verified | Fix |
| PL-DQ-11 | Stale text with no pointer, in requirements, risks and findings | Minor | Verified | Fix |
| PL-DQ-12 | The tour's step references are off in four places; a wrong risk number | Minor | Verified | Fix |
| PL-DQ-13 | Every "Carries" header disagrees with what its diagram shows | Minor | accepted | Fix: generated from the diagram text |
| PL-DQ-14 | The index omits the approach notes, the entry checks and the plan's diagram | Minor | Verified | Fix |
| PL-DQ-15 | Specimen headers that overstate; D-71 missing from the options file; dash lengths promised and not given | Minor | Verified | Fix |
| PL-PH-1 | One plan task narrates a hygiene step | Minor | Verified | Fix |
| PL-PH-2 | A commit message overstates: four boxes of the memory diagram carry figures with no label, one of them wrong | Minor | Verified | Fix the diagram; the message is noted here |
| PL-PH-3 | The process framework's name in two older documents | Minor | Verified | Decided at SHIP (D-70) |

## Tally

**74 findings** across nine lenses. Reviewer-labelled Critical: **6** (PL-CQ-1, PL-CQ-2, PL-PF-1, PL-PF-2, PL-AX-1, PL-PM-1). Questions for HUM LEAD: **8**. None dropped.

**Coordinator's own errors found by this round:** a fact given to HUM LEAD with a ruling and never checked (D-75: which module the width library imports); two frozen parity rows designed against without reading them (P-08, P-36); an addition written into a ruling's record without saying it was not ruled (D-73's decode limiter); a risk lowered on a measurement that was not the requirement's condition (RS-7); a checker that tested neighbours only, after using the same flaw to fail another scale; a ramp shown to HUM LEAD that works on one ground only; artefacts DISCOVER called "not a loose promise" left as promises; a work-package diagram whose edges were never checked against its own tasks.

## Questions for HUM LEAD

Put one at a time. Rulings are recorded in `rulings-discover.md` from D-81.

| Q | Question | From |
|---|---|---|
| PL-Q1 | The dependency allow-list, now that the fact behind D-75 is known to be wrong → **D-81: `go-runewidth` and `uax29`** | PL-CQ-1 |
| PL-Q2 | Parity row P-36, label language, against dropping translations while decoding → **D-82: one configured language kept, English by default; it joins the cache key** | PL-CQ-8, PL-BZ-1 |
| PL-Q3 | Parity row P-08, how a cell's colour is chosen → **D-83: upstream's majority vote, chosen from specimen 24; and a principle — geography and topography over road accuracy** | PL-BZ-1 |
| PL-Q4 | Whether the library limits concurrent decodes — written into D-73's record without a ruling → **D-84: no limiter; the peak line is stated two `Work` calls wide** | PL-PM-4, PL-CQ-11, PL-PF-3 |
| PL-Q5 | Default memory caps, and the line for three maps | PL-PF-1, PL-BZ-5 |
| PL-Q6 | How the end of a borrow is signalled | PL-CQ-2, PL-NC-3 |
| PL-Q7 | Whether radar is masked over water | PL-DQ-6 |
| PL-Q8 | The colour-vision threshold | PL-PM-1, PL-DQ-7 |
