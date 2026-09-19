# Red-Team Review — DISCOVER exit

> red-team: **NO-GO (round 1)** · multi-reviewer · scope: DISCOVER exit, committed work at `8830c11` · personas: [accessibility, newcomer, performance, infosec] (confirmed by HUM LEAD, D-38)

| Field | Value |
|---|---|
| Phase | DISCOVER exit (SEV-0: red-team mandatory) |
| Date | 2026-09-18 |
| Scope reviewed | Everything under `06_docs/02_features/go-tuimaps/`, plus `go.mod`, `.gitignore` and the git history, at commit `8830c11` |
| Lens set | Four fixed axes (Code Quality solo; Project Hygiene; Business Quality and Docs Quality sectioned, independent verdicts) · the DISCOVER phase lens · four confirmed personas in two sectioned pairs. Safety-critical not activated: it triggers on compiled code under the P10 gate, declared at BUILD entry (D-20). |
| Tree state | All six reviewers recorded a clean tree at `8830c11`. One uncommitted edit by the coordinator (the D-38 row in `rulings-discover.md`) was made after four reviewers had started, against the project's own freeze rule; one reviewer saw it, attributed it correctly, and did not report it as a defect. No other mutation during the round. |
| Status | Round 1 recorded and remediated; its thirteen questions ruled (D-40 to D-56). **Round 2 recorded below**: four lenses SHIP-WITH-CONDITIONS, one NO-GO (accessibility), since remediated; ten questions for HUM LEAD. |

## Verdicts

| Lens | Verdict | Driving findings |
|---|---|---|
| DISCOVER phase lens | SHIP-WITH-CONDITIONS | RS-3 over-claimed; M1 has no scenario set; host dependency under-stated |
| Code Quality | SHIP-WITH-CONDITIONS | Async ownership undefined; NFR-3 unmeasurable and fails its arithmetic; untrusted-input limits and escape injection missing |
| Project Hygiene | SHIP-WITH-CONDITIONS | A superseded commit kept only in the local undo history; identity positional, not repo-local; wording of commit-hygiene rules in public docs |
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
| C-8 | **The colour-to-intensity table as a hard requirement is a usability cliff, and D-36 was ratified from prose against the project's own specimen rule.** | Phase lens 5 · Newcomer B-1 · Docs B4 | Accepted. HUM LEAD has since clarified D-36 (D-39); the table was then ruled required (D-45). |
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
| PH-1 | A superseded commit, reachable only through the local undo history, was not in keeping with the sole-author rule | Important | **Verified** | **Fixed (D-50):** shown to HUM LEAD, then purged; branch tips, all commits and the integrity check verified unchanged. Publish only from a fresh clone. |
| PH-2 | Identity was positional, not repo-local | Important | **Verified** | **Fixed:** repo-local identity set. **PLAN:** identity check in the phase-exit gate. |
| PH-3 | How much the public docs say about commit-hygiene rules and their enforcement | Important (escalated) | Accepted; the first host's public docs carry the same narrative, so there is precedent either way | **Fixed (D-50):** neutral wording in tracked documents. |
| PH-4 | An employer-prefixed token inside a verbatim quote (brief, D-1); the framework's name in tracked text | Important | **Verified** | **Fix:** the quotation shortened with an ellipsis mark. Framework name kept, command names replaced with plain descriptions (D-50). |
| PH-5 | Process vocabulary pervasive; undecodable by a public reader | Minor | Accepted | **Fix:** glossary. |
| PH-6 | `.gitignore` names vendor products | Minor | **Declined:** the tracked ignore block is the guard that keeps local tooling files out; the first host does the same. Offered to HUM LEAD in Q9. | — |
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
| DQ-B4 | Widest gaps between HUM LEAD's words and the recorded implication: D-36, D-14, D-34 | — | **Verified** | **Resolved by HUM LEAD (D-39)**; the one item then open (Q4) was ruled in D-45. |

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

75 findings across nine lenses. Reviewer-labelled Critical: 14 *(this line said 13 until round 2 recounted the rows)*, collapsing to 8 distinct after convergence. After the coordinator's re-assessment: **Critical 8** (C-1, C-2, C-3, C-4, C-5, A-1, A-2, and S-2, which its own row rates "Critical — accepted" and this line had left out) · the remainder Important or Minor. **Declined: 2** (PH-6; PM-10 in part), each with rationale. **None dropped.**

Coordinator's own errors found by this round: the RS-3 over-claim; the D-34 "both sizes" claim; the risk miscount; "2×8"; an incomplete check of a quotation; recording silence as assent in D-34; certifying HUM LEAD's own condition in D-14; rating RS-4 settled. Each is corrected on the record rather than quietly.

## Questions for HUM LEAD arising from this round — all ruled

Presented one at a time, 2026-09-18. What each lettered option stood for is in [`rulings-options-as-presented.md`](../02-analysis/rulings-options-as-presented.md).

| Q | Question | From | Ruling |
|---|---|---|---|
| Q1 | Which specimens HUM LEAD had viewed | C-1 | **D-40** — specimen 1 only; a review page of every specimen was built at his request. **D-41, D-42, D-54** — he then viewed them all: fine, except the block renderer, which is acceptable thinned and as an opt-in renderer |
| Q2 | How M1 is judged | C-5 | **D-43** — seven scenarios; judged against a computed answer key; the session in the first host is the host's metric |
| Q3 | The first release | C-6 | **D-44** — three overlay shapes; then integration into the first host; then the rest |
| Q4 | Whether the colour table is mandatory | C-8, D-39 | **D-45** — required; a tested example table ships with the project |
| Q5 | The D-18 schema item | PM-5 | **D-46** — OpenMapTiles only in v1; the schema stays a seam |
| Q6 | Radar animation | PM-6 | **D-47** — in v1, built after v0.1.0, designed into the contract now |
| Q7 | What the 8 MB test contains | C-3 | **D-48** — a typical-day fixture; a separate rule for the worst case |
| Q8 | Parity-matrix changes | CQ-4, BQ-4, P-4 | **D-49** — six corrections; denominator 70, 62 in v0.1.0 |
| Q9 | Pre-publication hygiene | PH-1, PH-3, PH-4 | **D-50** — purged and fresh-clone rule; neutral wording; command names replaced |
| Q10 | The reading of "0.01.0" | DQ-5 | **D-51** — v0.1.0 |
| Q11 | A text answer to "where" | C-7 | **D-52** — the view described as data, in v0.1.0; renderable and speakable by a host |
| Q12 | Colour-vision-safe ramps | A-3, A-4 | **D-53** — required for defaults, reported for themes; **D-54** — the diverging ramp is fine; **D-55** — temperature's ramp is fixed |
| Q13 | The flash rate; reduce-motion | A-5 | **D-56** — flash at no more than 2.5 per second; reduce-motion; nothing hidden by a timer |

## Remediation — what was done after round 1

| Commit | What |
|---|---|
| `8119003` | Requirements: 15 rows rewritten to be testable, 17 added *(this row, and that commit's message, said "13" and "20"; round 2 counted the marked rows at that commit)*; risk assessment corrected in the open; glossary; dated supersession notes; brief amended to v1.0.1 |
| `0479e5a`, `8d24075` *(and, before round 1, `dde7df9` and `05ce211`, wrongly listed here as remediation)* | Owed specimens: thinned block renderer; feature overlays and wind with no colour; contours; radar with no colour, re-coloured and themed; a colour-vision-safe ramp; a diverging ramp |
| `80cc327` … `82921ee` | Rulings D-40 to D-56. *This row said "each with its requirement, risk and matrix changes". Round 2 showed that several ruling commits updated their own row and little else; the sweep after round 2 caught up.* |
| `3b47b9f` | Parity matrix corrected under its own change control |
| `fdab035` | Pre-publication hygiene: neutral wording; a superseded commit purged after being shown to HUM LEAD |
| (this commit) | The options as presented for every ruling (docs finding DQ-5) |

**Still open after remediation, carried to PLAN with an owner:** the asynchronous-work model (FR-30 states the constraints; PLAN chooses between at least two models); the terminal glyph-width matrix (PLAN entry); a proper contouring pass; the radar resampling rule; an exact colour table for one radar provider; the memory measurement against the D-48 fixture (PLAN exit); module layout before the first tag; the two MAPSCII additions with no disposition; offering the defect findings upstream.

**A second round follows**, with fresh reviewers given this record and the fix commits, told to re-verify the fixes, hunt for what round 1 missed, and attack what the fixes introduced.

---

# Round 2

| Field | Value |
|---|---|
| Date | 2026-09-18 |
| Tree reviewed | `1e889b1`, frozen for the whole round — no tracked file changed while reviewers ran |
| Reviewers | Five fresh reviewers, nine lenses, the personas HUM LEAD confirmed for round 1 (D-38). Each was given this record and the fix commits and told to re-verify the fixes, hunt for what round 1 missed, and attack what the fixes introduced. |
| Finding codes | R2-AX accessibility · R2-NC newcomer · R2-CQ code quality · R2-BZ-A phase lens · R2-BZ-B business · R2-DQ docs · R2-PH hygiene · R2-PF performance · R2-IS security |

## Verdicts

| Lens | Verdict | What drove it |
|---|---|---|
| Accessibility | **NO-GO** | Two Criticals: the specimen that closed round-1 A-2 had no place marker; FR-29 could not describe points or lines |
| Newcomer | SHIP-WITH-CONDITIONS | NFR-19 could not be met as worded; the only demonstrated radar route uses a deferred shape |
| Code quality | SHIP-WITH-CONDITIONS | Two contradictions inside the first slice, both introduced by round-1 fixes |
| Phase lens · Business | SHIP-WITH-CONDITIONS | A braille-only first release against the host's glyph floor; M1's key and the thing it checks were the same code; "Fix" dispositions never done |
| Docs · Hygiene | SHIP-WITH-CONDITIONS | Ruling commits updated their own row and little else; the neutral-wording rule was breached by the documents recording it |
| Performance | SHIP-WITH-CONDITIONS | One fix made the worst-case rule impossible to pass; the steady-state cost had no number |
| Security | SHIP-WITH-CONDITIONS — **round 1's NO-GO lifted** | Text cleaning stopped at the cell; whether the library reaches the network by default needs a ruling |

**Round verdict: NO-GO as reviewed; every Critical is remediated in the commits listed below.** As in round 1, no finding says the idea does not work.

## Round-1 fixes, re-verified

Held: the RS-3 correction, the risk recount, the 2×4 correction, the options file, the specimen-rule note, the parity counts (44 · 13 · 2 · 11 · 2; denominator 70), the defect ledger, identity on all commits, no unreachable objects, no trailers or tool names in any message or tracked file. **Did not hold, or held in part:** neutral wording (R2-PH-1); stale text (R2-DQ-3, -4, -5); the memory test (R2-PF-A1, -A2); the allocation target (R2-PF-A3); text safety (R2-IS-B1); the no-colour alert specimen (R2-AX-1).

## Findings and dispositions

Dispositions as in round 1. **Fixed** means done in the commits listed below — not promised. **Owed** names where it will be done. **Question** goes to HUM LEAD. "Verified" = re-checked by the coordinator; "accepted" = the reviewer's evidence read and found sound.

### Accessibility

| ID | Finding | Severity | Check | Disposition |
|---|---|---|---|---|
| R2-AX-1 | Specimen 13a has no place marker at either size, yet S13-1 said its position was readable | Critical | **Verified** — a count of the marker glyph in both files was zero | **Fixed.** Redrawn; S13-1 corrected with the earlier wording quoted. At 69×12 the frame still cannot always settle inside-or-outside, and the finding now says so |
| R2-AX-2 | FR-29 covered areas, fields and images only; no points, lines, "no data" or staleness; the app gave a screen-reader user no path | Critical | Verified against the row | **Fixed** (FR-29; describe mode in FR-5) |
| R2-AX-3 | The re-coloured radar ramp fails the colour-vision check at 256 colours | Important | accepted — the reviewer reproduced the coordinator's other figures exactly | **Owed in PLAN:** a passing ramp per depth (FR-16); recorded as S15-3 |
| R2-AX-4 | A brightness threshold picks the worse foreground on four steps | Important | accepted | **Fixed** (FR-16: compute both, take the higher) |
| R2-AX-5 | Legends in specimens 14 and 16 showed the previous ramp | Important | **Verified** | **Fixed**; HUM LEAD's D-54 look was at the uncorrected files, so a second look is asked for |
| R2-AX-6 | NFR-21 outside the first-release range | Important | Verified | **Fixed** |
| R2-AX-7 | "Reported, not refused" reaches the host's developer, not a colour-blind user of a themed host | Important | accepted | **Question R2-Q7** |
| R2-AX-8 | Round 1 found the alert tint at 1.96:1 and nothing was done | Important | Verified — no requirement mentioned it | **Fixed** (FR-16: a tint is never the only edge; the outline meets 3:1 and is not optional) |
| R2-AX-9 | No specimen had a scale mark | Important | Verified | **Fixed** in the redrawn specimens; FR-33 sharpened |
| R2-AX-10 | The short label `[FW]` means fire weather in the weather service's own codes | Important | accepted | **Fixed** (plain words; FR-18a) |
| R2-AX-11 | On a light terminal the basemap is about 1.3:1 | Important | accepted | **Question R2-Q8**; risk RS-24 |
| R2-AX-12 | "Speakable, with units" had no owner for units | Important | Verified | **Fixed** (FR-7, FR-13, FR-29) |
| R2-AX-13 | No minimum size or zoom for users who cannot resolve braille dots | Minor | accepted | **Owed in PLAN** (contract note) |

### Newcomer

| ID | Finding | Severity | Check | Disposition |
|---|---|---|---|---|
| R2-NC-1 | NFR-19's call counts cannot be met for an image or a grid | Important | Verified | **Fixed** (rewritten with counts that can be met) |
| R2-NC-2 | Every radar specimen fetched tiles, a shape deferred past the first release | Important | **Verified** from the specimen headers | **Owed at PLAN entry:** confirm a keyless bounding-box image source (NFR-17); risk RS-23. One request was tried in this round and the service was unavailable |
| R2-NC-3 | The glossary says tiles are fetched "by default"; FR-22b says never unless configured | Important | Verified | **Question R2-Q9** |
| R2-NC-4 | One US-only provider; its terms unchecked | Important | accepted | **Fixed** (NFR-17) |
| R2-NC-5 | NFR-20 needs a closed list of error kinds and a table-driven test | Important | accepted | **Fixed** |
| R2-NC-6 | No versioning promise; no README list of what is deferred | Important | accepted | README list **fixed** (NFR-20); the promise is **Question R2-Q4** |
| R2-NC-7 | Glossary inaccuracies | Minor | Verified | **Fixed** |
| R2-NC-8 | Stale markers read literally | Minor | Verified | **Fixed** |

### Code quality

| ID | Finding | Severity | Check | Disposition |
|---|---|---|---|---|
| R2-CQ-1 | Nothing wakes a back-off when the library runs no timers | Important | accepted | **Fixed** (FR-23: a not-before time checked on the next call) |
| R2-CQ-2 | Settle cannot end | Important | accepted | **Fixed** (FR-30) |
| R2-CQ-3 | "One character per cell" contradicts NFR-8 and two parity rows | Critical | Verified against the rows | **Fixed** (FR-34, NFR-8: grapheme clusters) |
| R2-CQ-4 | By-reference with no ownership rule is a data race in the host | Critical | Verified | **Fixed** (FR-11: borrowed geometry) |
| R2-CQ-5 | Blink aliases at the host's 300 ms tick | Important | accepted | **Fixed** (FR-25) |
| R2-CQ-6 | NFR-6's mechanism was wrong; cross-compiling proves nothing | Important | accepted | **Fixed** |
| R2-CQ-7 | FR-29's acceptance cited a deferred scenario; its cost had no home | Important | Verified | **Fixed** |
| R2-CQ-8 | No distance model | Important | accepted | **Fixed** (FR-29, FR-33) |
| R2-CQ-9 | "n·log n" silently picks an algorithm, against a Match row | Important | accepted | **Fixed** (a target, not a choice) |
| R2-CQ-10 | "Kind" is undefined; no midpoint or default breaks | Important | accepted | **Question R2-Q6** |
| R2-CQ-11 | The pinned width table reads the locale at start-up | Important | accepted | **Fixed** (NFR-8: the fixed narrow condition) |
| R2-CQ-12 | NFR-10's entry-count bound was four times too loose | Important | accepted; agrees with R2-IS-B4 | **Fixed** |
| R2-CQ-13 | Some first-release checks pass with nothing to check | Important | Verified | **Fixed** (the slice table says such checks are not counted as passed) |
| R2-CQ-14 | "Least-recently-read" has no portable mechanism | Important | accepted; agrees with R2-PF-A7 | **Fixed** (FR-21a, FR-21b) |
| R2-CQ-15 | No attributes file; specimen files treated as text | Minor | Verified | **Fixed** |
| R2-CQ-16 | "Designed in PLAN" names no artefact | Minor | Verified | **Fixed** (the named list in the requirements) |

### Phase lens and business

| ID | Finding | Severity | Check | Disposition |
|---|---|---|---|---|
| R2-BZ-A1 | The first release is braille-only; the host's stated glyph floor is block characters | Important | Verified against the host research | **Question R2-Q1** |
| R2-BZ-A2 | M1's answer key and FR-29 were the same code | Important | Verified — D-52's own presentation said so | **Fixed** (the key comes from an independent script) |
| R2-BZ-A3 | Scenarios 2, 6 and 7 have never been rendered | Important | Verified | **Owed before PLAN exit** |
| R2-BZ-A4 | The options file shortened D-44, D-48 and D-52 too far | Important | Verified against the conversation record | **Fixed**; the change of memory measure was stated with D-48, and the file now shows it |
| R2-BZ-A5 · R2-BZ-B5 | Four round-1 items labelled "Fix" were never done: an options-and-do-nothing table, a US-first statement, the M6 tally, a post-SHIP check | Important | Verified | US-first **fixed** (NFR-17). The other three are **owed in the Discovery Report**. *Round 1's label was wrong; this is the second time a disposition said "Fix" for something deferred.* |
| R2-BZ-A6 | No compatibility policy; no gate between the host integration and the remaining shapes | Important | accepted | **Question R2-Q4** |
| R2-BZ-A7 | A host can declare a generic scalar and theme temperature anyway | Minor | accepted | With R2-Q6 |
| R2-BZ-A8 | RS-3 "Low" rests on looks taken in a browser | Important | Verified | Wording reconciled; **a look at the specimen files in HUM LEAD's own terminal is asked for before PLAN exit** |
| R2-BZ-B1 | Five rows marked for the first release cannot be fully shown in it | Important | accepted | **Question R2-Q5** |
| R2-BZ-B2 | The tile generator needs the deferred archive reader | Important | Verified; agrees with R2-IS-B9 | **Question R2-Q2** |
| R2-BZ-B3 | A 16-colour hint has no defined behaviour in the first release | Important | Verified | **Question R2-Q3** |
| R2-BZ-B4 | The cost is a line count in three versions; "two-thirds" has no breakdown | Important | accepted | **Owed in the Discovery Report** |
| R2-BZ-B6 · B7 | Stale metric rows in the brief; NFR-21 unplaced | Minor | Verified | **Fixed** (brief v1.0.3) |

### Docs and hygiene

| ID | Finding | Severity | Check | Disposition |
|---|---|---|---|---|
| R2-DQ-1 | "13 rows rewritten, 20 added" was wrong, here and in a commit message | Important | **Verified** by counting the marked rows at that commit | **Fixed** here; the message cannot be changed without rewriting history and is noted instead |
| R2-DQ-2 | NFR-21 in neither release table | Important | Verified | **Fixed** |
| R2-DQ-3 · 4 · 5 | Answered questions shown as open; the brief's metric rows; the rulings summary out of step | Important | Verified | **Fixed** |
| R2-DQ-6 | P-59 marked Fix without a ledger entry | Important | Verified | **Fixed** (the definition of Fix; the ledger's not-covered list, dated) |
| R2-DQ-7 | The extension rows were never extended for the new requirements | Minor | Verified | **Owed:** proposed rows put to HUM LEAD with the Discovery Report, because the matrix is under change control |
| R2-DQ-8 | Finding codes collide with requirement prefixes | Important | Verified | **Fixed** (keys in the requirements and the glossary; round 2 uses R2- codes) |
| R2-DQ-9 | D-46 cited an error NFR-20 did not contain | Minor | Verified | **Fixed** |
| R2-DQ-10 | This record was stale and miscounted | Minor | **Verified** — 14 rows carry a reviewer's Critical | **Fixed** above, in place |
| R2-DQ-11 | The specimen findings were out of date with later rulings | Minor | Verified | **Fixed** in part; the specimens 2–11 section still predates D-40 and D-41 and says so by its date |
| R2-PH-1 | Tracked text said more about commit hygiene than the ruled sentence | Important | Verified | **Fixed** where it was commentary. **Declined in part:** the D-50 ruling row and its options record what was ruled and what was offered; a ruling cannot be recorded without saying what it was |
| R2-PH-2 | A framework command name survived in a quotation | Important | Verified | **Fixed** |
| R2-PH-3 | "Agent" in an assessor's position | Important | Verified | **Fixed** |
| R2-PH-4 | One commit message narrates the hygiene work | Important | Verified | **Question R2-Q10** |
| R2-PH-5 | Commit discipline slipped: several units in one commit | Minor | Verified | Accepted; round-2 remediation is one unit per commit |
| R2-PH-6 | "Two superseded commits" against "a superseded commit" | Minor | Verified | **Fixed** by the rewording |
| R2-PH-7 | The root commit's two dates differ by under two minutes | Minor | accepted | Noted; harmless |
| R2-PH-8 | Contact-sheet images are half the tracked bytes | Minor | accepted | **Owed at SHIP:** keep or drop before publication |

### Performance and security

| ID | Finding | Severity | Check | Disposition |
|---|---|---|---|---|
| R2-PF-A1 | The worst case could not pass: raw bytes counted against a cap the 12.4 MiB input exceeds | Critical | **Verified** — 812,058 × 16 bytes | **Fixed** (FR-11, NFR-3) |
| R2-PF-A2 | The test measured a first view and an undefined "peak" | Important | accepted | **Fixed** |
| R2-PF-A3 | A changed frame is the steady state and had no number | Important | accepted | **Fixed**, with provisional figures marked unverified |
| R2-PF-A4 | FR-29 had no cost rule | Important | accepted | **Fixed** |
| R2-PF-A5 | The host plans three maps; the test measures one | Important | accepted | **Fixed** (FR-27: shared caches, shared caps) |
| R2-PF-A6 | The cold-start link model was undefined | Important | accepted | **Fixed** |
| R2-PF-A7 | No recency source for the disk cache | Important | accepted | **Fixed** |
| R2-IS-B1 | Text cleaning stopped at the cell; strings returned as data or printed by the app were uncovered | Important; blocked exit | Verified against the row | **Fixed** (FR-34, FR-5) |
| R2-IS-B2 | Is the library's own default source "configured by the host"? Every tile request discloses where the user is looking | Important | Verified — FR-22b and D-21 disagree | **Question R2-Q9** |
| R2-IS-B3 | Tile addresses can hold keys and would appear in errors | Important | accepted | **Fixed** (FR-22b) |
| R2-IS-B4 | Decode limits bounded bytes, not objects | Important | accepted | **Fixed** (NFR-10) |
| R2-IS-B5 | Colour table and image unbounded | Important | accepted | **Fixed** (FR-9) |
| R2-IS-B6 | Picture and description could disagree: no fill rule, unclosed rings, the ±180° seam | Important | accepted | **Fixed** (FR-11, FR-29) |
| R2-IS-B7 | A future timestamp never goes stale; a frozen clock hides old data | Important | accepted | **Fixed** (FR-32, FR-25) |
| R2-IS-B8 | The disk cache is a location history with no off switch | Important | accepted | **Fixed** (FR-21b: off unless configured) |
| R2-IS-B9 | The generator and continuous integration escaped the network rules | Important | accepted | **Fixed** in part (NFR-11, FR-28a); the reader is **Question R2-Q2** |

## Convergence in round 2

| Reached independently by | Finding |
|---|---|
| Accessibility · Code quality · Phase lens | FR-29 was incomplete and its acceptance could not be met |
| Business · Security | The tile generator needs a reader the slice defers |
| Newcomer · Security | The library's default network source contradicts "no connection unless configured" |
| Code quality · Performance | The disk cache's recency rule had no mechanism |
| Code quality · Security | The directory entry-count bound |
| Accessibility · Business · Docs | NFR-21 outside the first-release table |
| Newcomer · Phase lens | No compatibility promise |

## Tally

87 findings across nine lenses. Reviewer-labelled Critical: 5 (R2-AX-1, R2-AX-2, R2-CQ-3, R2-CQ-4, R2-PF-A1), all fixed. **Declined: 1, in part** (R2-PH-1). **Questions for HUM LEAD: 10, all since ruled (D-57 to D-66).** None dropped. Most Important findings carry "introduced by a fix": round 1's remediation was wide and fast, and this is its cost.

**Coordinator's own errors found by this round:** S13-1 claimed a marker was readable in a specimen that had none — the round-1 over-claim pattern, repeated; legends that did not match their ramps; a miscount in a commit message, again; four dispositions labelled "Fix" that were deferrals, again; ruling commits that did not carry their consequences through the other documents; a security number left as "proposed" where tests must come first.

## Questions for HUM LEAD from round 2

Put one at a time, as in round 1. **All ten are ruled (D-57 to D-66)**, recorded in `rulings-discover.md`; two went against the recommendation (D-61, D-62) and one added to it (D-63).

| Q | Question | From |
|---|---|---|
| R2-Q1 | A braille-only first release against the first host's block-character glyph floor → **D-57: stands; the host states the need; the text description is the fallback** | R2-BZ-A1 |
| R2-Q2 | How the tile generator reads the planet file when the archive reader is deferred → **D-58: a minimal internal reader in v0.1.0, hardened and fuzzed** | R2-BZ-B2, R2-IS-B9 |
| R2-Q3 | What the first release does with a 16-colour hint → **D-59: basemap and features in 16 colours, ramps in their no-colour form; specimen owed** | R2-BZ-B3 |
| R2-Q4 | The compatibility promise before v1, and whether integration findings return to HUM LEAD before the remaining shapes are built → **D-60: breaks only at a minor version, listed and guarded; a written integration review and his GO** | R2-BZ-A6, R2-NC-6 |
| R2-Q5 | The first release's parity denominator, given five split rows → **D-61: the five rows are split; 62 of 75** | R2-BZ-B1 |
| R2-Q6 | The closed set of scalar kinds, temperature's midpoint per unit, and default breaks → **D-62: a fully absolute, library-owned temperature scale; a closed set of kinds; a convention, not a control** | R2-CQ-10, R2-BZ-A7 |
| R2-Q7 | A user-level "safe ramps" switch → **D-63: a switch in v0.1.0, and the palette as a documented semantic-token contract** | R2-AX-7 |
| R2-Q8 | Who owns the ground colour on a light terminal → **D-64: painted by default from a `ground` token; a host may opt out by declaring the ground** | R2-AX-11 |
| R2-Q9 | Whether the library reaches the network without being told to → **D-65: never; the app enables the default source, says so, and offers an offline flag** | R2-IS-B2, R2-NC-3 |
| R2-Q10 | One commit message that narrates hygiene work → **D-66: it stands; history is not rewritten** | R2-PH-4 |

Also asked of HUM LEAD, not as rulings: a second look at the corrected specimens (13a, 14, 16); a look at the specimen files in his own terminal before PLAN exit.

## Remediation after round 2

| Commit | What |
|---|---|
| `d3e00c8` | Corrected specimens 13a, 14, 15a, 16 and their findings |
| `35b46cd` | Requirements revised: 31 rows |
| `fe4e7eb` | Specimen files marked byte-exact |
| `ae17b89` | Brief amended to v1.0.3 |
| `2d97f91` | Sweep: answered questions, forward pointers, options in full, two risks |
| (this commit) | This record; glossary terms; units and error kinds in the requirements |

**Carried to PLAN, added by round 2:** a passing radar ramp per depth; a confirmed bounding-box image source, at entry; renderings of scenarios 2, 6 and 7; per-tile maxima across all zoom levels; the pinned fixture's exact contents; a minimum-size note for braille. **Owed in the Discovery Report:** the options-and-do-nothing table, the M6 tally, the post-SHIP check, the cost breakdown, proposed extension rows.

**A third round:** the find-rate has not converged — round 2 found five Criticals, most in round 1's fixes. A narrow third round follows the round-2 rulings: fresh reviewers, the round-2 fix commits only.
