# PLAN — Red-team record

| Field | Value |
|---|---|
| Phase | PLAN |
| Date | 2026-09-19 |
| Tree reviewed | `9d4cbd0`, frozen for the whole round — no tracked file changed while reviewers ran |
| Reviewers | Five, fresh, covering the nine lenses HUM LEAD confirmed (D-38, D-80): code quality · performance and security · accessibility and newcomer · phase lens and business · docs and hygiene |
| Subject | The architecture and its diagrams, the three approach notes, the implementation plan, the PLAN specimens (17 to 23), the memory measurement |
| Status | Round 1: **NO-GO as written**, 74 findings; all eight questions ruled (D-81 to D-88); remediation listed below. **Round 2, on the remediation only: NO-GO as written, narrowly — 60 findings, 2 Critical; all five questions ruled (D-90 to D-94); remediation listed at the foot of this file.** **Round 3, narrow: SHIP-WITH-CONDITIONS from every lens but hygiene, which is GO — no Critical, 27 findings, no new question.** Recorded at the foot of this file. |

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

Put one at a time. **All eight are ruled (D-81 to D-88)**; three went against the recommendation (D-84, D-85, D-86).

| Q | Question | From |
|---|---|---|
| PL-Q1 | The dependency allow-list, now that the fact behind D-75 is known to be wrong → **D-81: `go-runewidth` and `uax29`** | PL-CQ-1 |
| PL-Q2 | Parity row P-36, label language, against dropping translations while decoding → **D-82: one configured language kept, English by default; it joins the cache key** | PL-CQ-8, PL-BZ-1 |
| PL-Q3 | Parity row P-08, how a cell's colour is chosen → **D-83: upstream's majority vote, chosen from specimen 24; and a principle — geography and topography over road accuracy** | PL-BZ-1 |
| PL-Q4 | Whether the library limits concurrent decodes — written into D-73's record without a ruling → **D-84: no limiter; the peak line is stated two `Work` calls wide** | PL-PM-4, PL-CQ-11, PL-PF-3 |
| PL-Q5 | Default memory caps, and the line for three maps → **D-85: lean first — tiles 0.5 MB and shapes 0.25 MB shared, images 0.25 MB a map; the lines cover three maps** | PL-PF-1, PL-BZ-5 |
| PL-Q6 | How the end of a borrow is signalled → **D-86: reported, never waited for — `Set` and `Remove` return released yes-or-no; the host's own calls report the rest; an opt-in borrow check** | PL-CQ-2, PL-NC-3 |
| PL-Q7 | Whether radar is masked over water → **D-87: by kind — images never masked, scalar fields masked, a host can flip either** | PL-DQ-6 |
| PL-Q8 | The colour-vision threshold → **D-88: 10, for every pair and against the ground; method pinned** | PL-PM-1, PL-DQ-7 |

## Remediation after round 1

Every finding marked **Fix** above, and where it was done. A finding is listed here only if the change exists in the commit named.

| Commit | What | Findings |
|---|---|---|
| `e0aed45` to `e97e145` | The eight rulings, each carried into the requirements, diagrams and plan in its own commit: the allow-list (D-81); one label language (D-82); the cell-colour vote, from a side-by-side specimen (`094babd`, D-83); no decode limiter (D-84); lean caps covering three maps (D-85); the end of a borrow reported, never waited for, with the state machine redrawn (D-86); water masks fields and never images (D-87); the colour-vision threshold with its method (D-88) | PL-CQ-1, -2, -8, -11 · PL-PM-4 · PL-PF-1, -3 · PL-BZ-1, -5 · PL-NC-3 · PL-DQ-6 |
| `b43e7fd`, `d32d455` | The temperature scale retuned so every pair passes; a second, darker-is-heavier radar ramp for light grounds; specimens re-rendered, the defect kept on record as 22e. Seen by HUM LEAD (D-89) | PL-AX-1, -2 |
| `bfae89c` | **The contract document:** the groups of calls; the pump drawn, with its wake rule; the three-call path corrected; set, replace, remove; the frame, its lifetime and reuse key; which calls are safe together, with four implementation rules; closed lists of error and warning kinds; shared caches; where each deferred shape lands; places and markers; panic recovery | PL-CQ-3 · PL-PM-1, -2, -3, -6, -7 · PL-NC-1, -2, -4, -7 · PL-PF-5 · PL-IS-5 · PL-AX-4 |
| `b17049a` | **The constants table**, with per-tile maxima measured on twelve heavy real tiles; the tile extent limit cut to 8,192 because larger does not fit 16-bit coordinates; zoom buckets; the drop rule; the fallback bound stated as work; the class cap; rounding precision; the token list; targets only the library can measure, checked by arithmetic and tied to the task that measures them | PL-CQ-12 · PL-PF-2, -6 · PL-IS-2 · PL-PM-1 · PL-DQ-7 |
| `a7c763c` | Eight missing diagrams; six contradictions between diagrams removed; the contract diagram aligned; a shared `scene` package; "Carries" headers regenerated from each file; the tour's references corrected; the index completed. All 41 diagrams rendered and parse | PL-DQ-1 to -5, -9, -12, -13, -14 · PL-CQ-7 · PL-CQ-15 in part |
| `6409a24`, `842fc32` | **The plan reworked:** 210 tasks to 264; the dependency diagram and first milestone corrected; tests that could not fail, or asserted throwaway figures, rewritten; a static check in place of a goroutine count; allocation tests in a leg without the race detector; the sub-process race test; a local gate over every module; TileJSON, style, image and host-struct limits and fuzz targets; the protobuf pitfalls; a confined cache root; the generator's encoder and determinism; the pump's wake rule; public-surface and call-pair tests; accessibility switches; HUM LEAD's judging and approval sessions; the host's local-override recipe, spike and integration-review template. **The parity mapping**: owner and test for each of the 62 rows, written first | PL-CQ-4, -5, -6, -9, -10, -13, -14, -15 · PL-PF-4, -7 · PL-IS-1, -3, -4, -6 · PL-AX-3, -5, -6 · PL-NC-5, -6 · PL-PM-5, -9 · PL-BZ-2, -3, -4 · PL-DQ-10 · PL-PH-1 |
| `e2b1e49` | The pinned fixture committed: 3.3 MB of real data with sources, terms and a hash list | PL-DQ-8 · PL-PM-5 |
| `d2c84ef` | Specimen 25: isolines by classifying the field at every braille dot; assumption A-5 held | PL-PM-1 (the contour specimen) |
| `befd572` | Stale statements pointed at what overtook them; the specimens' header defects listed; D-71's options row; every figure in the memory diagram labelled | PL-DQ-11, -15 · PL-PH-2 · PL-PM-9 |
| `dbd4fc4`, `1df1de0` | HUM LEAD's second look at groups I, J and L recorded (D-89) | PL-PM-8, PL-AX-7 in part |

**Not fixed, and why.** PL-PH-3 (the process framework named in two older documents) is decided at SHIP, as D-70 rules. **Still owed by HUM LEAD before PLAN exit, and asked again:** a look at the colour specimens in a terminal, with the terminal's name and font (PL-PM-8, PL-AX-7). **Risk RS-7 stays High:** the peak line cannot be shown until the library's own benchmark exists.

## Round 2 — on the remediation only

| Field | Value |
|---|---|
| Date | 2026-09-19 |
| Tree reviewed | `a96d09c`, frozen for the whole round |
| Scope | Narrow: the remediation of round 1 and rulings D-81 to D-89 — commits `f2264f4..a96d09c` |
| Reviewers | Three, fresh: **engineering** (code quality, performance, security) · **product** (accessibility, newcomer, phase lens, business) · **docs and hygiene**. Each was told to check every round-1 "Fix" against the files, to attack what the remediation introduced, and to recount every number |
| Status | **NO-GO as written, narrowly**: 2 Critical, both a rule PLAN never stated. 60 findings. Five questions for HUM LEAD, all ruled (D-90 to D-94); remediation below |

### Verdicts, round 2

| Lens | Verdict | What drove it |
|---|---|---|
| Engineering | **NO-GO** | Nothing says what happens when the live views need more than a shared cache holds; by the documents as written, two maps on different views fetch and decode without end |
| Product | **NO-GO** | The temperature preset fails D-88's own ground rule on a light ground |
| Docs quality | SHIP-WITH-CONDITIONS | Seven diagrams or requirements still state a rule the contract or a ruling now states differently |
| Project hygiene | GO | Three Minor notes |

**Critical findings by round: 6, then 2.** Both of round 2's are gaps the coordinator left, not flaws the reviewers found in a ruling. Every headline claim below was re-checked by the coordinator before it was accepted: the colour figures were recomputed with a separate script, the image's size read from the file, the memory program re-run.

### Round-1 fixes that did not hold

| ID | Round-1 finding | What did not hold | Severity | Check | Disposition |
|---|---|---|---|---|---|
| P2-ENG-H1 | PL-PF-1 | The re-run's "3.56 MB measured" exceeds its listed parts by about 0.6 MB, unexplained; L2-memory still says RS-7 is Medium and the arithmetic "cautious by a factor of two to three" | Important | **Verified, and worse than put.** Re-run: 3.04 MB on the Gulf fixture; 3.56 was the Midwest variant, unlabelled. In both, **the view's own tiles sat outside the filled cache** — the program never decided whether tiles in use count against the cap. That is P2-ENG-1 | **Question P2-Q1**; the note and L2-memory are corrected with it |
| P2-ENG-H2 | PL-PF-2, PL-PF-6, PL-CQ-13 | The fallback's bound became a work bound only in constants and task 14.10; FR-11, task 10.10, L2-memory and L2-overlays still ask for a time | Important | Verified | **Fix** |
| P2-ENG-H3 | PL-PF-4 | Draw-time resampling exists only as task 09.23; L2-overlays still resamples inside prepare, the parts table gives `overlay` "resample", and 10.20 builds it where render cannot import it | Important | Verified | **Fix** |
| P2-ENG-H4 | PL-CQ-9 | Tasks 06.7 and 07.3 still verify with a goroutine-counting helper; FR-30 still says "the count of goroutines is unchanged" | Minor | Verified | **Fix** |
| P2-ENG-H5 | PL-CQ-10 | Only 09.19 moved off the race-detector leg; four other allocation assertions still run under it, where counts are not stable | Important | Verified | **Fix** |
| P2-ENG-H6 | PL-CQ-3 | 07.10 has the defect 07.x was cleared of: Render beside Work inside a package that cannot import render | Important | Verified | **Fix** — moved to WP-12 |
| P2-ENG-H7 | PL-CQ-7, PL-DQ-10 | Edges still missing from the work-package diagram: WP-10 needs WP-02 and WP-07; WP-03, -05, -06, -08 emit cleaned text and need WP-02; M-A lists 12.3 and 12.4, which need WP-10 and WP-11 | Important | Verified | **Fix** |
| P2-ENG-H8 | PL-PF-7 | The urban tiles arrive at 14.20 but 03.12 already tests against them | Important | Verified | **Fix** — the tiles are committed now, and 14.20 is withdrawn |
| P2-DOC-H1 | PL-DQ-3 | FR-16 still takes the higher of black and white and never keeps the line's own colour | Important | Verified | **Fix** |
| P2-DOC-H2 | PL-DQ-5 | The tile machine's two exits from Loading overlap, so a table-driven test cannot be written; L2-tiles has no "unavailable" path | Important | Verified | **Fix** |
| P2-DOC-H3 | PL-DQ-13 | L2-colour's header no longer equals the ids in its text | Minor | Verified | **Fix** — every header regenerated at the end of this remediation |
| P2-DOC-H4 | PL-DQ-14 | The index says the contract has three diagrams (two); the parity mapping is not listed | Minor | Verified | **Fix** |
| P2-DOC-H5 | PL-BZ-5 | L2-memory's header still says RS-7 is Medium | Important | Verified | **Fix** |

The product reviewer's seven rows under this heading point at eight of its findings below — P2-PRD-1, -2, -4, -5, -6, -7, -9 and -10 (one row named two); the docs reviewer's other three are P2-DOC-1, -4 and -8.

### New findings

#### Engineering

| ID | Finding | Severity | Check | Disposition |
|---|---|---|---|---|
| P2-ENG-1 | **No rule for live views that need more than a shared cache holds.** The tile machine goes on-hand → evicted → wanted; nothing keeps a tile a view is drawing. Two maps on different views each evict what the other needs, and every render re-wants it: fetching and decoding without end, against a public tile server, the status never "complete", the pump never idle. Shapes the same. D-85 accepted "panning back re-decodes", not this | **Critical** | **Verified**, with new measurements: a view's tiles in compact form are 0.26 to 0.80 MB (Gulf 0.33, Midwest 0.80; single dense city tiles 0.06 to 0.19 MB each) against a 0.5 MB cache | **Question P2-Q1** |
| P2-ENG-2 | **"Larger than a quarter of its cache is drawn, not cached" against lean caps.** The threshold is 0.125 MB a tile: every Midwest tile, the zoom-3 ancestor and several city tiles are over it, so they are held somewhere the cap does not count, and the 3.0 MB sum leaves them out. For shapes it is 62.5 KB: the fixture's own alert overlay is 87 KB at zoom 9, so the "far beyond this" fallback is the ordinary path from zoom 8, and memory-measurement's sentence saying otherwise is false | Important | Verified | **Question P2-Q1** — one rule answers both |
| P2-ENG-3 | **A lost wake-up.** A cancelled shared job returns to the queue inside a pump call; the other map's pump was already told there was nothing to do, and `OnPending` fires only inside owner calls. Also unstated: whether `Pending` counts work in flight | Important | Verified | **Fix** |
| P2-ENG-4 | **Holes in "which calls are safe together".** A sequence has the pump call `Pending`, listed as an owner call; seven calls are in no cell; `Settle` returns no released ids; `Close` during `Work` is undefined; Render as "last reader" cannot happen under the table's own rule | Important | Verified | **Fix**; the last part with P2-DOC-3 |
| P2-ENG-5 | Task 07.13's "calling the map from inside the hook is detected" cannot be built for pump calls — there is no goroutine identity to tell a re-entrant `Work` from a legal one | Important | Verified | **Fix** — detection is limited to owner calls |
| P2-ENG-6 | No error kind for a context that ends, and the no-foreign-errors rule makes `errors.Is(err, context.Canceled)` fail | Important | Verified | **Fix** — a `cancelled` kind that answers `errors.Is` for both context errors and carries none of their text |
| P2-ENG-7 | Constants tests need are still missing: the pending-queue cap, the fetch timeout, the near-duplicate-id and implausible-unit rules, the caps in bytes (MB and MiB are mixed), the embedded-asset bound, and whether an old image is kept while its replacement is prepared | Important | Verified | **Fix** |
| P2-ENG-8 | The static check's scope: "no `go` statement" would fail legitimate tests and the test kit; the fetch timeout is a standard-library timer; the typed error has no package to live in | Important | Verified | **Fix** |
| P2-ENG-9 | Numbers: cold start one wide is 2.55 s, not 2.45; "at most twice the vertices" is contradicted by the document's own data (about 2.3× a level); 14.6 reads the heap after `Work` returns and so misses the peak inside it; the reuse key omits places and reduce-motion | Minor | Verified by re-adding | **Fix** |
| P2-ENG-10 | amd64 under emulation is not amd64: two math functions the projection would use take a different path by processor feature, and the gate's guards could skip silently | Minor | Accepted; the feature-flag claim is the reviewer's, **unverified** here | **Fix** — the projection is written with functions that have no assembly on either architecture, and the gate fails rather than skips |
| P2-ENG-11 | Security wording: "no exported constructor" is impossible as written; TileJSON addresses have no stated rules; the confined root follows symlinks inside it; the dial hook is convention only; 04.8 uses an encoder 04.13 builds | Minor | Verified | **Fix** |

#### Product

| ID | Finding | Severity | Check | Disposition |
|---|---|---|---|---|
| P2-PRD-1 | **The temperature preset fails D-88 on a light ground.** The two bands either side of freezing are 5.0 and 5.5 from the light ground (threshold 10), at both depths; they pass on the dark ground. No task checks temperature against a ground | **Critical** | **Verified** — recomputed: 5.0 and 5.5 truecolor, 5.0 and 5.6 at 256 colours | **Question P2-Q2** |
| P2-PRD-2 | The light-ground radar ramp is not "darker is heavier": it gets lighter from class 2 to class 3 under every kind of vision | Important | **Verified** — lightness 52.4 then 57.7 | **Fix** — retuned; **HUM LEAD looks again (P2-Q4)**, because D-89 approved the old colours |
| P2-PRD-3 | The ramp switch is not in L2-colour's diagram; nothing says what a declared ground (D-64) does, or which ground applies when none is painted; one token set serves two ramps; on mid-grey grounds the light ramp's palest class falls to 9.9 | Important | Verified | **Fix** |
| P2-PRD-4 | The pump as drawn: the wake channel must be buffered; one wake reaches one goroutine, so "two wide" never happens; the re-queue (P2-ENG-3); a retry coming due is covered only by inference; the map's state after a `Close` that found a call inside; calls missing from the table | Important | Verified | **Fix** |
| P2-PRD-5 | Sequence 4 still has `New` with no size; `Settle` with no size is undefined; the app's one-shot mode has no size flag | Important | Verified | **Fix** |
| P2-PRD-6 | Parity owners assigned by upstream's source file, not by who builds the behaviour: P-44, P-45, P-51 mis-owned; P-31, P-52, P-54, P-55, P-58, P-60 with no task; test names that differ between the mapping and the plan | Important | Verified on eleven rows | **Fix** |
| P2-PRD-7 | P-61 (markers by id, remove by id — Match) cannot be met by a wholesale `SetPlaces`; P-57 (the library's footer — Match) has no call | Important | Verified against the frozen rows | **Fix** — the contract gains the calls. *The coordinator designed against frozen rows without reading them, as in round 1* |
| P2-PRD-8 | M1 cannot be judged as planned: no way to set the two sizes; no task writes scenarios 1, 3 and 4 or their keys; no protocol for the sitting | Important | Verified | **Fix** |
| P2-PRD-9 | The committed radar image is 1100×700 over Indiana; NFR-3's fixture is 600×400 over the Gulf view. At one byte a pixel it is 0.77 MB and the default cap refuses it | Important | **Verified** from the file | **Fix** |
| P2-PRD-10 | FR-19 still has water owning its cells under an image, against D-87 | Important | Verified | **Fix** |
| P2-PRD-11 | Ruling records that say more than the ruling: D-83's "borders" and "terrain last"; D-84's "two"; the options file's D-89 row still says group L is owed | Minor | Verified. D-84's option said "a named pump width"; *two* is the coordinator's figure | **Fix** — each gloss labelled as the coordinator's |
| P2-PRD-12 | The contract lets a very large borrowed shape disappear between being replaced and its replacement being ready. FR-11 says "never vanishes"; D-86 does not mention it; for a hazard it cuts against "better to overstate" | Minor | Verified. *The coordinator wrote the exception to keep D-86's "always yes for a one-goroutine host" literally true, and did not label it as unruled* | **Question P2-Q3** |
| P2-PRD-13 | Neighbours-only residue and stale text: L2-colour's "worst adjacent pair"; specimen 22's 256-colour figure is the adjacent one (23.8; all pairs 18.2); approach 2's "hard to ignore"; FR-7's "longest default ramp"; token names by band in one file and by position in another | Minor | Verified | **Fix** |
| P2-PRD-14 | Per-tile maxima were measured for zoom 5 to 14, not 0 to 14 | Minor *(graded by the coordinator; the reviewer gave none for -14 to -20)* | Verified | **Fix** — zoom 0 to 4 measured |
| P2-PRD-15 | The simplification algorithm is never named, though FR-11 requires it | Minor | Verified | **Fix** |
| P2-PRD-16 | The allocation "measurement" is arithmetic | Minor | True, and labelled so in constants section 7 | **Accepted** — only the library can measure it (task 14.7) |
| P2-PRD-17 | The local-override recipe and the integration-review template are cited and not written | Minor | Verified | **Fix** — both written |
| P2-PRD-18 | The alert preset's colours have no checker results | Minor | Verified | **Partly.** The two pairs the specimens drew were checked: they pass on a dark ground, and **both outlines fail 3:1 on a light ground**. Five severities on two grounds are designed in BUILD (task 08.10) and shown to HUM LEAD before reference frames are frozen. *A new gap, found by doing the check* |
| P2-PRD-19 | The keyboard equivalent of zoom-toward-the-pointer is only a call name | Minor | Verified | **Fix** |
| P2-PRD-20 | Two MAPSCII additions — per-layer label margin and clustering; loading a style by file path — still have no disposition, carried since DISCOVER round 1 | Minor | Verified | **Question P2-Q5** |

#### Docs and hygiene

| ID | Finding | Severity | Check | Disposition |
|---|---|---|---|---|
| P2-DOC-1 | As P2-PRD-9; also the 64×48 temperature grid is neither committed nor said to be generated; the image does not overlap the fixture's view | Important | Verified | **Fix** |
| P2-DOC-2 | L1 "The parts": after the `work` edge was removed, nothing in the library imports `tiles`, so `fetch`, `mvt` and `archive` are unreachable in the diagram | Important | Verified | **Fix** |
| P2-DOC-3 | Render as the reporter of a borrow's end contradicts the table that makes `Render`, `Set` and `Remove` one-at-a-time; the large-shape exception is in the contract only | Important | Verified. D-86's record names "`Work` or `Render`"; under the contract's own rule only `Work` can be last | **Fix**, labelled as the coordinator's reading of D-86 and listed for HUM LEAD in the Plan of Record; the exception itself is **P2-Q3** |
| P2-DOC-4 | The frame-reuse key is stated two ways; L2-render's status box lacks "failed" | Important | Verified | **Fix** |
| P2-DOC-5 | L2-memory contradicts D-85 and the measurement note | Important | Verified | **Fix**, with P2-Q1 |
| P2-DOC-6 | Label language: L2-style has a three-step fallback the decoder cannot serve (D-82 keeps one language); `house_num` omitted; tour step 12 cites the wrong ruling | Important | Verified | **Fix** |
| P2-DOC-7 | "Twenty-four diagrams" in two places; there are 34 in the architecture set, 41 in all | Important | Verified — recounted | **Fix** |
| P2-DOC-8 | Tour step 4 names a box that was renamed | Minor | Verified | **Fix** |
| P2-DOC-9 | Sequence 4: no size; `Settle`'s result incomplete | Minor | Verified | **Fix** (with P2-PRD-5) |
| P2-DOC-10 | L2-errors has Render's panic raise an "internal" *warning*; that kind is an error | Minor | Verified | **Fix** |
| P2-DOC-11 | Stale "owed" text in four diagram files, the options file, FR-30 and task 00.9 | Minor | Verified | **Fix** |
| P2-DOC-12 | "1,107 polygons" is 1,107 *rings* in 1,011 polygons | Minor | Verified | **Fix** |
| P2-DOC-13 | architecture.md has no "Carries" header; `SharedCaches` grouped two ways; `BorrowCheck` missing from the contract's first table; the test kit absent from L1 | Minor | Verified | **Fix** |
| P2-HYG-1 | One commit body (`d32d455`) says two rows changed where one did, and tells the story of an authoring slip | Minor | Verified | **Accepted** — history is not rewritten (D-70); recorded here |
| P2-HYG-2 | "3.3 MB" for the fixture is MiB; every other figure is decimal | Minor | Verified: 3,450,876 bytes | **Fix** — and the fixture has grown; the new figure is stated in decimal |
| P2-HYG-3 | The rulings file marks "against the recommendation" only from D-61 on; six earlier ones are marked only in the options file | Minor | Verified; predates this range | **Fix** — a note at the head of the rulings file says where the mark lives |

**Found by the coordinator while verifying:** the record of D-86 refers to HUM LEAD with a pronoun, against the rule adopted this session. Fixed.

### Tally, round 2

| | Critical | Important | Minor | Total |
|---|---|---|---|---|
| Fixes that did not hold | 0 | 10 | 3 | 13 |
| Engineering | 1 | 7 | 3 | 11 |
| Product | 1 | 9 | 10 | 20 |
| Docs and hygiene | 0 | 7 | 9 | 16 |
| **Total** | **2** | **33** | **25** | **60** |

**What held.** The reviewers' own recounts: 264 tasks; 62 parity rows, every one in the matrix with the same disposition; 25 risks; D-11 to D-89 contiguous; the two kinds lists identical; all 22 commits sole-author with clean messages; no exposure in tracked content; hashes 20 of 20; every temperature and radar figure the documents claim, recomputed independently to the first decimal, except the one in P2-PRD-13. Sound by the engineering reviewer's analysis: the 16-bit coordinate range; "last to read" without a lock held across the read; Render's snapshot rule beside drawing from borrowed memory; the confined root on the Go 1.25 floor; the race detector without the C toolchain on this platform.

### Questions for HUM LEAD, round 2

Put one at a time; each ruling is recorded in HUM LEAD's words in the rulings file.

| # | Question | From | Ruling |
|---|---|---|---|
| P2-Q1 | When the live views need more tiles or shapes than a shared cache's cap, what gives | P2-ENG-1, -2, -H1, P2-DOC-5 | **D-90: "A"** — need first, cap second |
| P2-Q2 | The temperature preset beside a light ground | P2-PRD-1 | **D-91: "A"** — a light-ground variant; three bands change, no margin lost (specimen 21d) |
| P2-Q3 | A very large borrowed shape, replaced: keep drawing the old, or stop until the new is ready | P2-PRD-12, P2-DOC-3 | **D-92: "C"** ◇ — `Set` builds the index; both rulings hold as written |
| P2-Q4 | A look at the retuned light-ground radar ramp | P2-PRD-2 | **D-93: "22f is fine / 21d is fine"** |
| P2-Q5 | The two MAPSCII additions carried since DISCOVER | P2-PRD-20 | **D-94: "A"** — the app gains `--style PATH`; label margin and clustering not taken up for v1 |

### Remediation after round 2

| Commit | What it did | Findings answered |
|---|---|---|
| `76e69cc` | Round 2 recorded as returned, every disposition written before any fix; a pronoun for HUM LEAD removed from D-86's record | — |
| `83cbc49` | **Fixture:** the 600×400 radar image over the fixture's own view; the four city tiles, each matching its pinned hash; rings and polygons counted separately; the temperature grid stated as built from a rule | P2-PRD-9, P2-DOC-1, P2-DOC-12, P2-ENG-H8, P2-HYG-2 |
| `dc33c12` | **The fixes that needed no ruling.** Contract: how `wake` is written, the re-queued job's wake, `Pending` defined, `Settle`'s result, `Close`, the three classes of call, places by id, the footer, four new kinds, the error type's home. Constants: maxima across zoom 0 to 4, caps in bytes, the queue cap, timeouts, two warning rules, the algorithm, the projection's functions, corrected sums. Diagrams: L1's edges, resampling at draw time, the tile machine's exits, only `Work` ends a borrow, sequence 4, the reuse key, label language, the count of diagrams. Requirements: FR-7, FR-11, FR-16, FR-19, FR-30. Plan: new tasks, edges, the assets-order test moved, the encoder before its user, allocation assertions off the race leg. Parity mapping: four rows re-owned | P2-ENG-3 to -11, -H2 to -H7; P2-PRD-3 to -8, -10, -14, -15, -19; P2-DOC-2, -3, -4, -6 to -11, -13, -H1 to -H4 |
| `503fbdc` | **D-90:** need first, cap second; the quarter rule withdrawn everywhere; the memory note says which fixture each figure came from; L2-memory corrected | P2-ENG-1, -2, -H1; P2-DOC-5, -H5 |
| `6c1484f` | **D-91:** the temperature preset's light-ground variant; the light radar ramp corrected as 22f; specimens 21d, 21e, 22f; both ramp files updated with the old colours kept and the defects named | P2-PRD-1, -2, -13 (the 256-colour figure) |
| `510b22b` | **D-92:** `Set` indexes a very large shape itself; the unruled vanishing-shape text replaced; task 10.26 | P2-PRD-12, P2-DOC-3 |
| `4c7112c` | **D-93:** both light-ground corrections seen and found fine; the options file's D-89 row corrected | P2-PRD-2, -11 (part) |
| The commit that carries this section | **D-94:** the app's style-path flag; label margin and clustering not taken up. Glosses in the records of D-83 and D-84 labelled as the coordinator's; where the against-the-recommendation mark lives; the rulings summary extended; the local-override recipe and the integration-review template written; the alert colours checked; every "Carries" header regenerated from its file's text; all 41 diagrams parse | P2-PRD-11, -13, -17, -18, -20; P2-HYG-3; P2-DOC-H3 |

**Counts after remediation, recounted from the rows:** 272 plan tasks (13, 12, 11, 18, 15, 12, 20, 15, 22, 30, 26, 17, 26, 15, 20); 62 parity rows by owner 4, 5, 1, 4, 6, 31, 7, 4; rulings D-11 to D-94, fourteen against the recommendation; error kinds 24, warning kinds 13, identical in the contract and L2-errors; 34 diagrams in the architecture set, 41 in all; fixture 25 files, 6.6 MB.

**Not fixed, and why.** P2-PRD-16: the allocation figures are arithmetic and labelled so; only the library can measure them (task 14.7). P2-HYG-1: a commit body's inaccuracy is recorded, not rewritten (D-70). P2-PRD-18: **partly** — and doing the check found a new gap: the alert outlines PLAN's specimens used fail 3:1 on a light ground, and three of five severities have never been drawn. The alert preset's colours are designed in BUILD task 08.10 and shown to HUM LEAD before any reference frame is frozen. P2-ENG-10's claim about a processor-feature branch was not re-checked; the design no longer depends on it.

**Found by the coordinator while remediating, not by a reviewer.** The heaviest tile is a zoom-2 world tile, not the city tile first recorded (1.56 MB decompressed against 1.10). The "quarter of the cache" rule had been written into FR-11 and NFR-3 in DISCOVER; the coordinator first told HUM LEAD it had never been written down as ruled, and corrected that in D-90's record. A first retune of the temperature variant put a warm colour on the below-freezing band; caught before it was shown. The question P2-Q2 said two bands would change with a thin margin; three changed with none lost, and D-91's record says so. The vertex count in D-92 (31,250) is the coordinator's, labelled as not ruled.

**The coordinator's errors this round exposed:** two rules never stated (what a cache may evict; a preset against its ground); a checker that tested differences and never order, after round 1 had already caught it testing neighbours only; a measurement reported without saying which fixture it was; an unruled exception written into the contract to keep a ruling literally true; parity owners assigned by upstream's file layout and never checked against who builds the behaviour; frozen parity rows designed against without reading them — for the second round running.

**A third round follows**, narrow: whether round 2's fixes hold, and rulings D-90 to D-94. Both earlier rounds found fixes that did not hold, so this remediation gets the same check.

## Round 3 — narrow: do round 2's fixes hold

| Field | Value |
|---|---|
| Date | 2026-09-19 |
| Tree reviewed | `5572fb5`, frozen for the whole round |
| Scope | Commits `a96d09c..5572fb5`: round 2's remediation and rulings D-90 to D-94 |
| Reviewers | Two, fresh: **A** — code quality, performance, security, accessibility, newcomer, business · **B** — docs quality, the phase lens, project hygiene |
| Status | **No Critical. 27 findings, counting once what both reviewers raised: 15 Important, 12 Minor.** Eight are round-2 fixes that did not hold. No new question for HUM LEAD; five matters are listed in the Plan of Record for HUM LEAD to accept or reject with it |

| Lens | Verdict |
|---|---|
| Reviewer A, all six lenses | SHIP-WITH-CONDITIONS |
| Docs quality | SHIP-WITH-CONDITIONS |
| Phase lens | SHIP-WITH-CONDITIONS |
| Project hygiene | GO |

**Critical findings by round: 6, 2, 0.** Reviewer A's summary line gave 10 Important and 4 Minor; its rows give 12 and 5, and the rows are what is counted here. The coordinator re-checked the claims that carry numbers before accepting them: the lightness inversions of P3-B-3, the mid-grey figure of P3-B-6 and the 47 unnamed parity tests of P3-A-H3 all reproduce.

### Round-2 fixes that did not hold

| ID | Round-2 item | What did not hold | Severity | Check | Disposition |
|---|---|---|---|---|---|
| P3-A-H1 | D-90, "the quarter rule withdrawn everywhere" | Task 06.6 still tests a held-for-view state "over a quarter of the cache"; the tile machine has no such state. *Found by both reviewers* | Important | Verified | **Fix** |
| P3-A-H2 | P2-ENG-H7, edges | WP-07 and WP-01 report problems through `fault` and have no edge from WP-02; L1 has no `work` → `fault` edge | Important | Verified | **Fix** |
| P3-A-H3 | P2-PRD-6, rows with no task | Only the six rows round 2 named were given tasks. 47 of the 62 mapping tests are named in no task, and several of their behaviours — forced pixels, normalising, ocean outside the world, the POI glyph — are built by no task either | Important | Verified: 47, by script | **Fix** — every owning work package gains a closing task that lists its rows |
| P3-A-H4 | P2-ENG-H4 | Task 14.8 still gates on a goroutine count | Minor | Verified | **Fix** |
| P3-A-H5 | P2-ENG-7 | L2-memory, L2-tiles and NFR-3 still give the embedded tiles as 1.7 MB; constants and 04.11 say at most 2.5 MB | Minor | Verified | **Fix** |
| P3-B-H1 | P2-DOC-3, "only `Work` ends a borrow" | FR-11 still says "`Work` or `Render`" | Important | Verified | **Fix**, with the contract's label |
| P3-B-H2 | P2-DOC-2, -13 | L1 draws the test kit among the separate modules; five packages that report problems have no edge to `fault` | Minor | Verified | **Fix** |
| P3-B-H3 | P2-DOC-H1 | Task 08.4 is still named and described as the higher-contrast rule alone | Minor | Verified | **Fix** |

*(P3-A-H1 is counted once; reviewer B's first row is the same finding.)*

### New findings

| ID | Finding | Severity | Check | Disposition |
|---|---|---|---|---|
| P3-A-1 | **"Live view" is never defined.** State or last render? Does a stand-in ancestor count? Tiles wanted and not arrived? An idle map? A closed one? And need has no stated bound | Important | Verified | **Fix** — defined in the contract, constants and glossary, with the bound |
| P3-A-2 | `CacheUse` is classed as a short read under "the" lock, but need spans every map on a shared handle; the handle's lock and the lock order are unstated | Important | Verified | **Fix** |
| P3-A-3 | Tasks 06.3, 10.10 and 10.26 assert things their packages cannot see — maps, views, frames; no render task draws through the run index. *The same class of defect as P2-ENG-H6, a third time* | Important | Verified | **Fix** — package tasks restated in what the package owns; a render task and two public-package tasks added |
| P3-A-4 | The re-queue exception breaks the re-entrancy rule: a legal owner call arriving while another map's hook fires inside a `Work` would be refused. And a `Work` abandoned by `Close` is not said to re-queue its shared job | Important | Verified | **Fix** |
| P3-A-5 | Task 08.9's "further from the ground at every step" fails for both ruled ramps if measured as colour difference; only lightness is monotone | Important | Verified | **Fix** — the task says lightness |
| P3-A-6 | D-92's run index has no stated size, no stated home, and no rule for whether it is kept; at the vertex cap it is twice the shape cache; two bullets of the contract conflict for the fixture's own alert | Important | Verified | **Fix** |
| P3-A-7, P3-B-4 | "Below 31,250 the fallback can never be needed" ignores the index's own bytes; the count is tied to a cap that might change | Minor | Verified: with the index counted it is 30,303 | **Fix** — recomputed; caps are fixed at creation |
| P3-A-8 | The borrow check is a second linear pass in `Set`, and "re-checked at every read" re-reads a whole shape on every frame that draws from memory | Important | Verified | **Fix** — fingerprints by run; benchmarked on and off |
| P3-A-9, P3-B-2 | **The alert-preset gap is not carried visibly**: no risk row, no plan risk row, not in L2-colour's owed table; HUM LEAD's look is a clause inside a task, after M-B | Important | Verified | **Fix** — a risk row (RS-26), both tables, and the look as its own task before M-B |
| P3-A-10 | The local-override recipe omits the host's toolchain line, the tidy step and the checksum file, and misstates what a workspace file may hold; its test needs the network off and a warm module cache | Important | Accepted | **Fix** |
| P3-A-11 | `Close` "drops every reference" against `InUse` answering yes afterwards; what a `Work` returning `closed` carries | Minor | Verified | **Fix** |
| P3-A-12, P3-B-8 (part) | `CacheUse` missing from the L1 contract box and task 12.18; `render-failed` has no test; L2-colour's defaults box omits alerts | Minor | Verified | **Fix** |
| P3-B-1 | **D-90's record and its options row disagree.** The record states the withdrawal of the quarter rule and the whole-cap threshold as ruled; the options row mentions neither | Important | **Checked against the question as put:** option A's text did say that a simplified shape goes to the host's memory "only when it is bigger than the whole cap, not a quarter of it". The options row shortened it away | **Fix** — the options row restored. The "accepted with the ruling" sentences in D-90 and D-94 are the strongest counter-argument as put with each question; the options file records options only, and now says so |
| P3-B-3 | **Order under every kind of colour vision is asserted for radar only.** The temperature scale gets lighter from class 14 to 15 under protanopia in truecolor, and inverts in three places at 256 colours; its file says "ordered" with no scope | Important | **Verified** — 25.7 then 27.4; at 256: 43.3/42.4, 43.7/42.1, 58.0/57.9 | **Fix, in part.** FR-16 defines order as relative luminance, which holds. In truecolor one colour moved by 1.9 — below what the eye can tell — makes the scale ordered under all four kinds with every pair still 12.2 apart; done, and labelled as the coordinator's. At 256 colours the palette is too coarse; the scope is now stated and the three small inversions recorded. **Listed for HUM LEAD in the Plan of Record** |
| P3-B-5 | Stale status headers in the plan; requirement rows edited in PLAN with no note in the file's status | Minor | Verified | **Fix** — and the edited rows are listed in the Plan of Record |
| P3-B-6 | The mid-grey figure is 8.1, not 9.9; the temperature file's 19.8 is against black; S22-4 still gives the first light ramp's figures as current | Minor | Verified: 8.1 at grey 199, under 10 from 185 to 216 | **Fix** |
| P3-B-7 | Specimens 21d and 21e carry a legend 20 °C off; the header's caveat covers the data, not the legend | Minor | Verified | **Fix** — a README line; D-93 judged colours |
| P3-B-8 | "Seven rows" that name eight; D-89 in two summary rows; six verdict rulings missing from the list of rulings not given as options | Minor | Verified | **Fix** |
| P3-B-9 | A vendor-name scan matches street names inside a Paris map tile | Minor | Verified: unmodified map data | **Accepted** — SHIP's scan exempts the tile files; noted in the fixture's README |

### Tally, round 3

| | Critical | Important | Minor | Total |
|---|---|---|---|---|
| Fixes that did not hold | 0 | 4 | 4 | 8 |
| New — reviewer A | 0 | 9 | 3 | 12 |
| New — reviewer B alone | 0 | 2 | 5 | 7 |
| **Total** | **0** | **15** | **12** | **27** |

*Counted from the rows above. Three rows carry two ids each, because both reviewers raised the same thing; each is counted once, under reviewer A.*

**What held,** by the reviewers' own recounts: 272 tasks; 62 parity rows with owners 4, 5, 1, 4, 6, 31, 7, 4; D-11 to D-94 contiguous, fourteen ◇ marks; 24 and 13 kinds, identical in both places; 34 and 41 diagrams; every "Carries" header equal to its file's text; 48 files with no broken link or ragged table; fixture hashes 25 of 25; all eight commits sole-author and clean; no pronoun for HUM LEAD anywhere. Every colour figure recomputed independently to the first decimal. The pump walked through with two goroutines and five jobs: no wake lost, nothing left spinning. The memory sums re-added: 3.04, 3.34, 4.94 MB.

### Remediation after round 3

One commit carries all of it, with this section.

| Area | What changed | Findings answered |
|---|---|---|
| Contract | "Live view" and "need" defined, with the bound on need; the shared-caches handle's own lock and the lock order; the re-entrancy guard armed only by hooks an owner call fired; a `Work` that gives up a shared job for **any** reason — cancelled, or its map closed — re-queues it and wakes the other map; `Close` reworded, and what a `Work` returning `closed` carries; the run index's size, home and life, and which bullet governs a large shape's replacement; the borrow check by run; caps fixed at creation | P3-A-1, -2, -4, -6, -8, -11 |
| Constants, glossary | The vertex count recomputed with the index's own bytes — 30,303, not 31,250; the run index's row; the mid-grey figure corrected to 8.1; the scope of "ordered" stated; three glossary terms | P3-A-7, P3-B-4, -6, -3 |
| Requirements | FR-11's "`Work` or `Render`"; the embedded-tiles figure; the status line now lists every row PLAN revised | P3-B-H1, P3-A-H5, P3-B-5 |
| Diagrams | L1: the test kit inside the library's module, six more edges into `fault`, `CacheUse` on the contract; L2-colour: alerts in the defaults box, the alert colours and the scope of "ordered" in the owed table; L2-memory and L2-tiles: the embedded-tiles figure; L2-overlays: the vertex count. All 41 parse | P3-B-H2, P3-A-12, P3-B-2 |
| Plan | 284 tasks (was 272). 06.6 rewritten to the diagram's transitions; 06.3, 10.10 and 10.26 restated in what their packages own, with 09.31 (drawing through the run index), 12.27 and 12.28 (the two end-to-end cases) added; 08.4 renamed; 08.8 and 08.9 say lightness and state the scope; 08.23 is HUM LEAD's look at the alert colours, inside M-B; 14.8 no longer counts goroutines; edges from WP-02 to WP-01 and WP-07; **each of the eight owning packages closes with a task that lists its parity rows no other task names — 47 rows** | P3-A-H1, -H2, -H3, -H4, P3-A-3, -5, -9, P3-B-H3 |
| Risks | RS-26, the alert preset's colours, Medium, with its fallback; 26 risks: High 3, Medium 14, Low 7, Closed 2 | P3-A-9, P3-B-2 |
| Rulings and options | D-90's options row restored to what option A said; D-92's record corrected to 30,303 and the run length labelled; the summary's two rows no longer share D-89; the list of rulings not given as options completed; a note on what "accepted with the ruling" is | P3-B-1, -8 |
| Specimens | The temperature ramp file: class 15 moved by 1.9 so that truecolor is ordered under every kind of vision, the 256-colour inversions recorded, the dark-ground figure's ground named; README: S21-8, the legends of 21d and 21e, S22-4's superseded figures | P3-B-3, -6, -7 |
| First-host start; fixture README | The toolchain line, the tidy step, what a workspace file may hold, the test's conditions; the scan exemption for map tiles | P3-A-10, P3-B-9 |

**For HUM LEAD, with the Plan of Record — not ruled, and not hidden:**

1. Only `Work` and `Settle` report released ids; `Render` never can (the coordinator's reading of D-86's "`Work` or `Render`").
2. The vertex count above which `Set` builds the run index — 30,303 at the default cap — and the run length of 64.
3. "Two" as the pump width the peak line is stated at (D-84's option said "a named pump width").
4. One temperature colour moved by an amount below what the eye can tell, after D-89 and D-93 approved the scale; and the 256-colour scale's three small inversions under simulated colour blindness, recorded and not fixed.
5. The alert preset's colours are designed in BUILD, with HUM LEAD's look inside M-B.
6. Twenty-two requirement rows that the Discovery Report ratified were revised in PLAN; the requirements file's status line lists them.

**No fourth round.** Critical findings went 6, 2, 0. What round 3 found was one task row and one requirement sentence that had kept a withdrawn rule, tasks phrased in terms their package cannot see, and definitions a first test would have demanded on its first day. Both reviewers said the same in their counter-arguments. BUILD is test-first from these documents and stops at M-A for HUM LEAD's look. This is the coordinator's judgement; HUM LEAD may order another round before approving.

