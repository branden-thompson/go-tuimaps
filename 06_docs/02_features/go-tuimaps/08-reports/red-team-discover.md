# Red-Team Review — DISCOVER exit

> red-team: **NO-GO (round 1)** · multi-reviewer · scope: DISCOVER exit, committed work at `8830c11` · personas: [accessibility, newcomer, performance, infosec] (confirmed by HUM LEAD, D-38)

| Field | Value |
|---|---|
| Phase | DISCOVER exit (SEV-0: red-team mandatory) |
| Date | 2026-09-18 |
| Scope reviewed | Everything under `06_docs/02_features/go-tuimaps/`, plus `go.mod`, `.gitignore` and the git history, at commit `8830c11` |
| Lens set | Four fixed axes (Code Quality solo; Project Hygiene; Business Quality and Docs Quality sectioned, independent verdicts) · the DISCOVER phase lens · four confirmed personas in two sectioned pairs. Safety-critical not activated: it triggers on compiled code under the P10 gate, declared at BUILD entry (D-20). |
| Tree state | All six reviewers recorded a clean tree at `8830c11`. One uncommitted edit by the coordinator (the D-38 row in `rulings-discover.md`) was made after four reviewers had started, against the project's own freeze rule; one reviewer saw it, attributed it correctly, and did not report it as a defect. No other mutation during the round. |
| Status | Round 1 recorded. Remediation in progress. A second round is required (foundational scope; material find-rate). |

## Verdicts

| Lens | Verdict | Driving findings |
|---|---|---|
| DISCOVER phase lens | SHIP-WITH-CONDITIONS | RS-3 over-claimed; M1 has no scenario set; host dependency under-stated |
| Code Quality | SHIP-WITH-CONDITIONS | Async ownership undefined; NFR-3 unmeasurable and fails its arithmetic; untrusted-input limits and escape injection missing |
| Project Hygiene | SHIP-WITH-CONDITIONS | Reflog-only commit with a trailer; identity positional, not repo-local; public docs narrate the attribution rule |
| Business Quality | SHIP-WITH-CONDITIONS | No written v0.1.0 slice; M1 unanchored; host work under-counted |
| Docs Quality | SHIP-WITH-CONDITIONS | RS-3 claim unsupported by the record; stale statements unflagged |
| Accessibility | SHIP-WITH-CONDITIONS | No non-visual answer to "where"; feature overlays meaningless with no colour; ramps not colour-vision-safe |
| Newcomer | SHIP-WITH-CONDITIONS | The required colour table is a cliff; "just works" is prose, not a requirement |
| Performance | SHIP-WITH-CONDITIONS | Goroutine ownership unwritten; both headline targets unfalsifiable |
| **InfoSec** | **NO-GO** | **No control on terminal escape sequences reaching a frame; security posture is two sentences without numbers** |

**Round verdict: NO-GO** until the Critical findings are remediated. Every Critical is a requirements-text, record-accuracy or evidence gap; none says the idea does not work.

## How findings were handled

Each finding was checked against the documents or the source before being accepted (the project's verify-before-accept rule). **Verified** = re-checked by the coordinator; **accepted** = the reviewer's evidence was read and found sound without an independent re-derivation. Severity is re-assessed independently of the reviewer's label. Dispositions: **Fix** (done in this remediation; commit cited in the round-2 record), **Ruling** (needs HUM LEAD; question number), **Specimen** (needs a rendering first), **PLAN** (tracked to the next phase), **Declined** (with rationale).

## Convergence — findings several reviewers reached independently

| # | Convergent finding | Reached by | Coordinator's assessment |
|---|---|---|---|
| C-1 | **RS-3 was cut to Low on a claim the record does not contain.** The risk assessment said HUM LEAD had seen fields, radar, polygons and wind legible at both sizes, in both renderers, at four colour depths, citing D-32. D-32 covers specimen 1 only. The grid was never rendered: reduced-colour specimens exist for one view, one size, braille, temperature; the block renderer exists in truecolor only, and its own finding says roads swamp the field at 69×12. | Phase lens 1 · Docs B-1 | **Verified. The coordinator's error.** An overstated guarantee, the exact failure a standing calibration warns against. Critical. |
| C-2 | **Where asynchronous work runs is written nowhere**, while RS-4 was recorded as settled. | Code 1 · Perf P2 | Verified (no mention of goroutines, concurrency or cancellation in the requirements). Critical. |
| C-3 | **The 8 MB target is unsized, unmeasurable as worded, and fails arithmetic for the ruled worst case** (58 × 14,001 × 16 B = 12.4 MiB of source geometry alone). | Code 2 · Perf P1 | Arithmetic verified. Critical as a requirements defect; the figure itself was always a target to be measured (D-29). |
| C-4 | **Terminal escape injection has no control anywhere.** | Code 3 · InfoSec S1 | Verified by search. Critical. |
| C-5 | **M1 has no scenario set, and depends on first-host work for four of five overlay shapes.** | Phase lens 2, 3 · Business A-2, A-3 | Verified. Critical. |
| C-6 | **No written slice of v1; scope mitigation is a deferral; aggregate cost never shown.** | Phase lens 4 · Business A-1, A-6 · Code 13 | Accepted. Important, and a HUM LEAD decision. |
| C-7 | **A plain-text answer to "where" is missing, and may be the purest and cheapest form of M1.** | Phase lens 7, 8 · Accessibility A-1 | Accepted. Critical for accessibility; a HUM LEAD decision on scope. |
| C-8 | **The colour-to-intensity table as a hard requirement is a usability cliff, and D-36 was ratified from prose against the project's own specimen rule.** | Phase lens 5 · Newcomer B-1 · Docs B4 | Accepted. HUM LEAD has since clarified D-36 (D-39); whether the table is mandatory remains open. |
| C-9 | **Glyph width of block and box-drawing characters is carried only as an assumption, with no risk entry.** | Phase lens 9 · Code 6 · Accessibility A-7 | Verified. Important. |
| C-10 | **"Cached per zoom level" contradicts clip-to-view and is undefined at fractional zoom.** | Code 7 · Perf P4 | Accepted. Important; requirement reworded. |
| C-11 | **The disk cache has no size bound.** | Code 3 · Perf P6 | Verified. Important. |
| C-12 | **Poisoning of a no-expiry cache is permanent; a status check is not enough.** | Code 3 · InfoSec S4 | Accepted. Important. |

## Findings and dispositions

### DISCOVER phase lens (Distinguished PM)

| # | Finding | Reviewer | Coordinator | Disposition |
|---|---|---|---|---|
| PM-1 | RS-3 over-claimed (C-1); D-34 also claims the value lattice was shown at both sizes when only 149×38 exists | Critical | **Critical — verified** | **Fix:** RS-3 restored to Medium with the untested cells listed; D-34 claim corrected. **Ruling Q1:** viewing rulings per specimen group. **Specimen:** block renderer at 69×12 with thinning. |
| PM-2 | M1 not measurable: no scenario set; denominator widened by D-14 without change control; judge is the author; only the 10% polygon case rendered; "rough distance" has no scale reference | Critical | **Critical — verified** | **Fix:** scenario set drafted; scale-reference requirement added. **Ruling Q2:** ratify the set and the judge. |
| PM-3 | Host dependency hand-waved: CD-1 names polygons only; M1's live session quietly rewritten; circular with D-19 | Critical | **Important — verified** (the dependency is real; "circular" overstates a sequencing fact D-19 records openly) | **Fix:** CD-1 expanded per shape (with D-39); RS-11 raised to High. **Ruling Q2:** M1 live session as a SHIP gate or a post-SHIP host metric. |
| PM-4 | RS-5 mitigation is a deferral; no effort estimate; research recommended narrower | Important | Important — accepted | **Ruling Q3:** a ruled v0.1.0 cut line, with rough sizes shown. |
| PM-5 | D-36 ratified from prose; D-18 schema item never ruled; schema seam deleted with the shim; D-34 evidence gap | Important | Important — verified | **Specimen:** re-coloured and themed radar. **Ruling Q4:** table mandatory or optional. **Ruling Q5:** the D-18 schema item. **Fix:** schema-seam requirement; RS-6 to Medium. |
| PM-6 | No valid time or stale state on any overlay; no misleading-render risk; animation assumed but never required | Important | **Important — verified** | **Fix:** valid-time requirement and misleading-render risk added. **Ruling Q6:** is radar animation in v1. |
| PM-7 | Do-nothing and cheaper options never evaluated | Important | Important — accepted | **Fix:** options table added to the Discovery Report. |
| PM-8 | Missing parties: non-US users, screen-reader users, colour-vision deficiency, upstream | Important | Important — accepted | Covered by A-1, A-3 below. **Fix:** "US-first evidence" stated plainly. **PLAN:** offer the defect findings upstream. |
| PM-9 | Glyph width has no risk entry (C-9) | Important | Important — verified | **Fix:** risk entry added; terminal matrix check made a PLAN-entry task. |
| PM-10 | Pointer work and globe tours do not trace to the problem statement | Minor | **Declined in part:** pointer operations were ruled by HUM LEAD (D-17). Tours at a 100% parity target go to **Ruling Q3**. | — |
| PM-11 | Two stale records (objectives; rulings summary) | Minor | Verified | **Fix.** |

### Code Quality

| # | Finding | Reviewer | Coordinator | Disposition |
|---|---|---|---|---|
| CQ-1 | Async ownership undefined (C-2) | Critical | **Critical — verified** | **Fix:** requirement added stating the constraints any model must meet (ownership explicit, bounded, cancellable, leak-free, a completion signal, a settle path for headless). The model itself is a PLAN choice between at least two approaches. RS-4 reopened. |
| CQ-2 | NFR-3 not reproducible; arithmetic fails (C-3) | Critical | **Critical — verified** | **Fix:** measurement protocol pinned (live heap via runtime metrics against a pinned fixture; resident memory confirmatory). **Ruling Q7:** what the 8 MB scenario contains. |
| CQ-3 | Security text too thin; escape injection; disk cache cap; persist-after-decode | Critical | **Critical — verified** | **Fix:** see InfoSec S1–S7. |
| CQ-4 | Match rows written in the superseded shim's vocabulary (P-23, P-24, P-25, P-41); P-08 has three unrecorded exceptions | Important | **Important — verified** | **Ruling Q8:** matrix changes need a recorded ruling under its own change control. |
| CQ-5 | A boolean tick signal cannot show a 150 ms flash to a 300 ms clock | Important | **Verified** (300 ÷ 150 = 2, always the same phase) | **Fix:** the time requirement returns the next deadline and reports a visible phase change. |
| CQ-6 | L-7 changes P-34, P-60, P-11; NFR-8 contradicts wide characters; a hand-rolled width table would disagree with the host's | Important | Accepted | **Fix:** NFR-8 rewritten. **PLAN:** buy the width table the host already uses. |
| CQ-7 | FR-11 over-specifies: a clamped fill needs no clipper; per-zoom cache undefined (C-10); recursive simplification breaks the P10 rules | Important | Accepted | **Fix:** FR-11 reworded to outcomes; iterative simplifier with a stated bound. |
| CQ-8 | "Output unchanged" for L-11 breaks once style profiles switch: a tile cache keyed by z/x/y alone serves wrong tiles | Important | Accepted | **Fix:** requirement added on the cache key or style-free decoding. |
| CQ-9 | L-9 under-specified: interpolation of colours and widths; legacy filters or expressions | Important | Accepted (reviewer's spec recollection flagged as unverified) | **PLAN:** state the function semantics; **Ruling Q3** proposes legacy filters only in the first slice. |
| CQ-10 | L-17 (f) silent on overlays crossing the antimeridian | Important | Accepted | **Fix:** longitude convention and splitting at ±180° required. |
| CQ-11 | NFR-6 incomplete: tile order, map iteration, fused multiply-add across architectures | Important | Accepted | **Fix:** NFR-6 rewritten. |
| CQ-12 | go.mod traps: newer local toolchain; module layout for the app; line endings on Windows | Important | Accepted | **PLAN / BUILD entry:** build with the floor toolchain and read-only modules; decide module layout before the first tag; add `.gitattributes`. |
| CQ-13 | The hand-written estimate of ~1,500 lines is low by about six times (~8,000–10,000) | Important | Accepted (the estimate predated most scope rulings) | **Fix:** recorded in the risk assessment; shown to HUM LEAD in **Ruling Q3**. |
| CQ-14 | Other untestable wording: NFR-4, NFR-5, FR-16, FR-23, NFR-7, NFR-9, NFR-15, NFR-16 | Important | Accepted | **Fix:** each given a pass criterion. |

### Project Hygiene

| # | Finding | Reviewer | Coordinator | Disposition |
|---|---|---|---|---|
| PH-1 | A reflog-only commit holds an attribution trailer and the harness files; not pushed by `git push`, but travels with any copy of `.git` | Important | **Verified** | **Ruling Q9:** destructive step; needs HUM LEAD's confirmation. |
| PH-2 | Identity was positional, not repo-local | Important | **Verified** | **Fixed:** repo-local identity set. **PLAN:** identity check in the phase-exit gate. |
| PH-3 | Public docs narrate the attribution rule and its enforcement | Important (escalated) | Accepted; the first host's public docs carry the same narrative, so there is precedent either way | **Ruling Q9.** |
| PH-4 | An employer-prefixed token inside a verbatim quote (brief, D-1); the framework's name in tracked text | Important | **Verified** — the coordinator's scan pattern had a hyphen where the quote has a space | **Fix:** token removed with an ellipsis mark; scan pattern corrected. Framework name: **Ruling Q9**. |
| PH-5 | Process vocabulary pervasive; undecodable by a public reader | Minor | Accepted | **Fix:** glossary. |
| PH-6 | `.gitignore` names vendor products | Minor | **Declined:** the tracked ignore block is the guard that keeps harness files out; the first host does the same. Offered to HUM LEAD in Q9. | — |
| PH-7 | `.gitignore` gaps and one redundant line | Minor | Accepted | **BUILD entry.** |
| PH-8 | Commit-unit drift in three commits | Minor | Accepted for DISCOVER | One unit per commit from here. |
| PH-9 | No DISCOVER exit report yet; folders 03–07 absent | Minor | Phase-appropriate | The Discovery Report follows this review. |
| PH-10 | No adversarial-review record tracked | Minor | — | **Fixed** by this file. |

### Business Quality

| # | Finding | Reviewer | Coordinator | Disposition |
|---|---|---|---|---|
| BQ-1 | No path to a first usable release (C-6) | Important | Accepted | **Ruling Q3.** |
| BQ-2 | M1 unanchored (C-5) | Important | Verified | As PM-2. |
| BQ-3 | Host dependency under-counted (C-5) | Important | Verified | As PM-3. |
| BQ-4 | M3 exclusions: only P-40 is cleanly ruled; "Open for PLAN" leaves the denominator unfrozen and is a loophole | Important | **Verified** | **Ruling Q8:** P-06 and P-07 return to the denominator (70) as "form pending"; every exclusion listed by row for ratification. |
| BQ-5 | M3 counts departures as parity; NFR-5 dropped "to full detail"; M6 has no tally | Important / Minor | Verified | **Fix:** FR-1 restated; NFR-5 restored; M6 tally recorded. |
| BQ-6 | Unpriced, unowned cost; FR-12 and FR-19 trace to findings nobody ruled; the tile-strip script is outside the repository and is not Go | Important | Accepted | **Ruling Q8** for FR-12 and FR-19. **BUILD:** the generator is written in Go and committed (FR-28); the 1.7 MB figure is re-measured by the coordinator (1,697,545 bytes, 85 tiles). |
| BQ-7 | Problem evidence is first-hand only; no post-SHIP check | Minor | Accepted | **Fix:** stated plainly; post-SHIP check added. |

### Docs Quality

| # | Finding | Reviewer | Coordinator | Disposition |
|---|---|---|---|---|
| DQ-1 | RS-3 claim unsupported (C-1) | Critical | Critical — verified | As PM-1. |
| DQ-2 | Risk summary miscount (said Medium 5, Low 8; rows give 4 and 9), repeated in a commit message | Important | **Verified** | **Fix.** The commit message stays as it is; this record is the correction. |
| DQ-3 | Stale statements unflagged where a reader meets them | Important | Verified | **Fix:** dated supersession notes in place; banners on older documents. History is not rewritten. |
| DQ-4 | Same fact stated two ways ("2×8" braille; file counts; "10 MB"; MAPSCII additions; citations) | Important | Verified ("2×8" was copied from upstream's README; the code is 2×4) | **Fix.** The two MAPSCII additions with no disposition are recorded. |
| DQ-5 | 16 of 27 rulings record a bare letter; the options as presented are not in the record; D-28's reading never confirmed | Important | **Verified** | **Fix:** options appended as presented. D-28's reading: **Ruling Q10.** |
| DQ-6 | "Verified" is thinner than it reads | Important | Accepted | **Fix:** wording tightened to "copied, not re-verified"; RS-8 held at Medium. |
| DQ-7 | Requirement IDs out of order; hard-coded ruling range; no glossary | Minor | Accepted | **Fix.** |
| DQ-B4 | Widest gaps between HUM LEAD's words and the recorded implication: D-36, D-14, D-34 | — | **Verified** | **Resolved by HUM LEAD (D-39)**, with one item still open (Q4). |

### Accessibility persona

| # | Finding | Reviewer | Coordinator | Disposition |
|---|---|---|---|---|
| A-1 | No non-visual alternative; braille glyphs are noise to a screen reader (C-7) | Critical | **Critical — accepted** | **Ruling Q11:** a text description of the view as a first-class deliverable. |
| A-2 | Feature overlays, wind speed and severity are colour-only; meaningless with no colour | Critical | **Critical — verified** in the 69×12 text specimens | **Fix:** requirement added. **Specimen:** polygons and wind classes with no colour. |
| A-3 | Ramps are not colour-vision-safe; luminance rises then falls; FR-16 would pass a red-green ramp | Important | **Verified** (the ramp was the coordinator's; luminance is non-monotonic by construction) | **Ruling Q12:** luminance-monotonic, colour-vision-safe ramps as a requirement. **Specimen** to follow. |
| A-4 | No contrast number binds the ramps | Important | Accepted | **Ruling Q12.** |
| A-5 | Upstream's flash is 3.33 Hz, above the three-per-second threshold; no reduce-motion; the credit may fade on a timer | Important | **Verified** (50 ms × 3 × 2 = 300 ms) | **Ruling Q13:** a recorded deviation from P-59 and a reduce-motion option. |
| A-6 | Keyboard zoom-around-a-point has no mechanism, focus intent or acceptance test | Important | Accepted | **Fix:** focus intents and a non-colour focus indicator required; the mechanism stays a PLAN decision (D-17). |
| A-7 | The no-colour path relies on ambiguous-width glyphs (C-9); who reads `NO_COLOR` is unspecified | Minor | Accepted | **Fix.** |

### Newcomer persona

| # | Finding | Reviewer | Coordinator | Disposition |
|---|---|---|---|---|
| N-1 | The required table is the cliff (C-8) | Critical | Important — accepted | **Ruling Q4.** **Fix:** tested examples become a numbered requirement. |
| N-2 | "Following the README just works" is prose; behaviour with no tile at all is undefined | Important | Verified | **Fix:** both made requirements with acceptance tests. |
| N-3 | No glossary | Important | Verified | **Fix.** |
| N-4 | ~20 concepts before a first overlay; no minimal embedding | Important | Accepted | **Fix:** a minimal-embedding requirement with everything else defaulted. |
| N-5 | No requirement on host mistakes; no warning channel | Important | Verified | **Fix:** validation and error requirement added. |
| N-6 | Specimen headers read "256-bit colour", "16-bit colour", and "24-bit" on no-colour files | Minor | Verified | Noted in the specimen record; corrected when specimens are regenerated. |

### Performance persona

| # | Finding | Reviewer | Coordinator | Disposition |
|---|---|---|---|---|
| P-1 | NFR-3 scenario unsized; documents disagree (8, "10", "16–32" MB); raw host geometry unbudgeted (C-3) | Critical | Critical — verified | As CQ-2. |
| P-2 | Goroutine ownership; no completion signal; failed tiles re-fetched every render (13 requests a second at a free server) | Critical | Critical — verified | As CQ-1; back-off for failed tiles required. |
| P-3 | No bound on a changed frame | Important | Accepted; the reviewer's numbers are unverified | **Fix:** added as a target to be validated at PLAN exit. |
| P-4 | Clip cache contradiction (C-10); the linear label scan marked Match | Important | Accepted | **Fix:** FR-11 reworded; P-34 matches the rectangle semantics, not the scan (**Ruling Q8**). |
| P-5 | NFR-5's delay and bandwidth unspecified, so the cold target is unfalsifiable | Important | Verified | **Fix:** link numbers pinned as targets. |
| P-6 | Disk cache unbounded (C-11) | Important | Verified | **Fix.** |

### InfoSec persona

| # | Finding | Reviewer | Coordinator | Disposition |
|---|---|---|---|---|
| S-1 | **Escape injection: no requirement exists** (C-4) | Critical | **Critical — verified** | **Fix:** sanitisation at the cell write, fuzzed. |
| S-2 | NFR-10 has no numbers and omits gzip, TileJSON, PMTiles header and metadata, images, GeoJSON-like input | Critical | Critical — accepted | **Fix:** limits added, taken from measured maxima. |
| S-3 | Host data unbounded; quadratic or recursive simplification; non-finite values; a panic in a library goroutine kills the host | Important | Accepted | **Fix.** |
| S-4 | Cache poisoning is now permanent (C-12) | Important | Accepted | **Fix:** persist only after a complete successful decode; purge and verify exposed. |
| S-5 | Cache filesystem safety unspecified | Important | Accepted | **Fix.** |
| S-6 | No TLS, redirect or body-size rule; the replacement fetcher's signature cannot express a range or a limit | Important | Accepted | **Fix.** |
| S-7 | A checksum of an 86 GB file cannot be verified by a range-reading script; vulnerability scanning of a zero-dependency core is vacuous; no release integrity | Important | Accepted | **Fix.** |

## Tally

75 findings across nine lenses. Reviewer-labelled Critical: 13 (collapsing to 8 distinct after convergence). After the coordinator's re-assessment: **Critical 7** (C-1, C-2, C-3, C-4, C-5, A-1, A-2) · the remainder Important or Minor. **Declined: 2** (PH-6; PM-10 in part), each with rationale. **None dropped.**

Coordinator's own errors found by this round: the RS-3 over-claim; the D-34 "both sizes" claim; the risk miscount; "2×8"; a scan pattern that missed a token; recording silence as assent in D-34; certifying HUM LEAD's own condition in D-14; rating RS-4 settled. Each is corrected on the record rather than quietly.

## Questions for HUM LEAD arising from this round

Presented one at a time.

| Q | Question | From |
|---|---|---|
| Q1 | Viewing rulings for specimens 2–11, by group | C-1 |
| Q2 | The M1 scenario set, who judges it, and whether the live session in the first host gates SHIP | C-5 |
| Q3 | A v0.1.0 cut line, with rough sizes shown | C-6 |
| Q4 | Whether the colour-to-intensity table is mandatory | C-8, D-39 |
| Q5 | The D-18 schema item, now that AI-9 has reported | PM-5 |
| Q6 | Whether radar animation is in v1 | PM-6 |
| Q7 | What the 8 MB scenario contains | C-3 |
| Q8 | Parity-matrix changes under its change control | CQ-4, BQ-4, P-4 |
| Q9 | Pre-publication hygiene: the reflog, the attribution narrative, the framework's name | PH-1, PH-3, PH-4 |
| Q10 | Confirm the reading of "0.01.0" as v0.1.0 | DQ-5 |
| Q11 | A text description of the view as a first-class deliverable | C-7 |
| Q12 | Colour-vision-safe, luminance-ordered ramps with contrast numbers | A-3, A-4 |
| Q13 | The flash rate, and a reduce-motion option | A-5 |
