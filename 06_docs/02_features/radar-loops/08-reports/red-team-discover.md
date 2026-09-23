---
title: "v0.2.0 Radar loops — DISCOVER exit red team"
date: 2026-09-23
phase: DISCOVER (RCC) — exit
sev: SEV-0
authority: HUM LEAD
status: "ROUND 2 RECEIVED — consolidated below; dispositions follow one at a time"
---

# DISCOVER exit red team

## Round 1 — dispatch

**Five reviewers, none with prior context**, each briefed from watchpost's
`06_docs/red-team-brief.md` template (this repository has no brief of its own yet), li-A2DH's
DISCOVER phase lens, and its InfoSec and accessibility personas, verbatim. Each worked read-only in
its own scratch directory and was told not to run the gate, fuzzing or a full test sweep. Scope:
`feature/go-tuimaps..feature/radar-loops` (`650f267..ff51ccc`) and this release record.

| Reviewer | Verdict |
|---|---|
| Code quality | **Do not ship** — the gate has success paths without evidence |
| Project hygiene + docs quality | **Do not proceed** — the brief is wrong or incomplete; owed work has no tracker; the record is on one disk |
| Business quality + DISCOVER lens | **Do not proceed** — no requirement set, no risk register |
| InfoSec persona | **Proceed conditionally** — amend L-1, L-7, L-9 and L-10 first |
| Accessibility persona | **Do not proceed** — motion has no text alternative; the no-colour defect is wider than recorded |

The harness refused every reviewer's report file, so each report came back as its final message.
**Every report is filed verbatim in `reviewer-reports/`** (added in round 2, H-9), machine paths redacted.

## Verified by the coordinator before consolidation

| Claim | How checked | Result |
|---|---|---|
| Radar erases furniture at 16 colours too (A7) | `rampless` is NoColour **and** Colours16 (`internal/render/field.go:103-105`); specimen 29 re-rendered at `Colours16` | **Confirmed, and wider:** label, marker and credit gone at both sizes, scale gone at 69×12, and **every rain cell drawn as `░`**, so heavy rain looks like light rain at 16 colours (cause not yet found) |
| The decoder fix has no unit test (code F1) | Fix reverted; `go test ./internal/mvt` | **Confirmed:** passes with the fix reverted |
| `FUZZ_TIMEOUT` is inert (code Q10) | go1.25.0 `testing.go:2335-2340` (alarm stopped before `runFuzzing`); `cmd/go/internal/test/test.go:807` (no kill timeout when fuzzing) | **Confirmed:** a fuzz-engine stall now hangs the gate without limit |
| Images are borrowed, not copied (InfoSec F1) | `internal/overlay/store.go:437`; decode at `image.go:242-247` re-checks only `maxPixels` | **Confirmed** |
| `Fetcher` takes effect only at the next `Source` (InfoSec F4) | `tiles.go:67-83` | **Confirmed** |
| Wave 2's "71–80 %" (docs D-10) | 7,243/9,028; 20,397/25,310; 6,447/8,056 | **The coordinator's error:** 80.0–80.6 %. D-19 repeats it |
| README promises a user-agent token a host sets (docs D-3) | `README.md:68-69`; no exported call sets `fetch.Options.Token` | **Confirmed false** |
| The record exists on one disk (hygiene H-1) | `git branch -vv` | **Confirmed:** neither feature branch has an upstream |

## Round 1 — consolidated findings

Grouped by what disposes of them. Reviewer ids: C = code, H/D = hygiene/docs, B/P = business/DISCOVER
lens, F = InfoSec, A = accessibility.

| # | Finding | From | Severity | Disposition |
|---|---|---|---|---|
| R1-1 | **No normative requirement set.** The brief (and issue #2) is stale since D-11: L-5.1's triage was cancelled (D-18), L-9.2 asks for a Purge that exists, L-8.2 under-states the hatch collision, L-6.4/C-6 still say every commit is blocked; D-14, D-17, D-18, D-19, D-20, the specimen 29 defect and wave 1's "fix regardless" list are in no requirement. No risk register. No owed-items tracker. | B-1, D-1, D-4, D-5, H-11, H-12, P-8, B-7 | Critical | **FIXED — D-23**: `requirements.md` written (normative, traced, risk register, owed list); brief corrected; issue #2 carries both (`6b6b462`) |
| R1-2 | **Storm motion has no non-visual form.** Describe answers for the newest frame; M1 is scored by sight; the locked problem is "cannot see a storm's motion". | A1, B-4, P-1 | Critical | **RULED — D-24**: the description reports observed motion; M1 gains a non-visual arm (L-1.12) |
| R1-3 | **The specimen 29 defect is wider:** at Colours16 too, and it erases the `stale` word and the no-tiles notice (a stale loop can look current); at 16 colours every class draws as `░`. | A7 + coordinator | Critical | **RECORDED as requirements L-8.3 and L-8.4** (D-23): a defect against FR-18a, not a ruling; fixed in BUILD, with tests at NoColour and Colours16 |
| R1-4 | **Loop accessibility requirements missing:** frame time or gap as text (A2), step/seek (A3), ReduceMotion vs playback (A4), a state read (A5), a tick distinguishable from a data change (A6), the still form of Off (A15), a numeric rate ceiling counting every frame change and no flashed gaps (A16, P-5). | A2–A6, A15, A16, P-5 | Important | **RULED — D-25**: seven requirements, L-1.10a–g |
| R1-5 | **Default playback state** (WCAG 2.2.2). | A14 | Important | **RULED — D-26**: default off, plus one standard playback API (L-1.11, L-1.13) |
| R1-6 | **Light-ground blend has no named fallback.** | A10, B-6 | Important | **RULED — D-27**: search as D-14 directs; 29b's order is the named light-ground fallback (L-11.4) |
| R1-7 | **D-17 can fail silently:** a label that does not fit is dropped whole; dashes must differ from the line-overlay dash; no second cue at 16 colours. D-14's check covers class distance only, not outline/label contrast, and truecolor only. | A8, A9 | Important | **RULED — D-28**: four requirements, L-8.5–L-8.8 |
| R1-8 | **Description gaps:** area answers lack label and severity; image answers lack value, unit and "approximate"; nothing when no place is registered. | A11, A12, A13 | Important | **RULED — D-29**: list the alerts shown with name and severity; never "you"; inside / outside / nearby for named places; the rest deferred as necessity unproven |
| R1-9 | **Security rows missing from L-1/L-7/L-9/L-10:** frames borrowed not copied (F1); the budget ignores retained PNG bytes, no frame-count bound, playback interval must be library-owned (F2); host fetcher unbounded, `Checked` unwired (F3); `Fetcher` silently deferred (F4); Purge scope, in-flight writes, future mtimes (F5); two allow-lists, http under a host fetcher (F7). Minor: swallowed cache errors (F6), the cache as a viewing record (F8), no scan recorded at a tag (F9). | F1–F9 | Important | **RULED — D-30**: nine requirements (L-1.14, L-1.15, L-5.5, L-7.3, L-7.4, L-9.3–L-9.6, L-10.2, L-12.4) |
| R1-10 | **Gate defects:** `FUZZ_TIMEOUT` inert, so a stall hangs (C-Q10, D-6); the docs lane exits 0 when nothing changed and judges the tree not the commit (C-F10, H-6); a module `go list` cannot load reads as "empty" (C-F11); `git diff`'s exit status lost; the decoder fix has no unit test (C-F1); gate runs leave no trace, so M6 cannot be counted (H-5). | C, H-5, H-6, D-6 | Important | **RULED — D-31**: all six fixed now, test-first |
| R1-11 | **Measurement soundness:** the "71 %" error (D-10); no blind-spot sections (D-12); the two blend results measure different pairs (D-13); D-20 under-states 3 MB (holds about five dot-resolution loops, D-14); the programs are not filed, so nothing re-runs (P-9, D-11). | D-10–D-14, P-9 | Important | **RULED — D-32**: Limits sections in wave 2; the blend re-measured on both pair kinds and both grounds (S29-7; conclusions stand); D-20 corrected; programs, inputs and raw output filed in `02-analysis/programs/` |
| R1-12 | **The record is on one disk:** neither feature branch is pushed; issue #2 cites an unresolvable commit. | H-1 | Important | **RULED — D-33**: `feature/radar-loops` pushed and kept pushed, for portability; not a permanent record; deleted after the release merges |
| R1-13 | **README's user-agent token claim is false**, widening M5. | D-3 | Important | **RULED — D-34**: README corrected; L-4 extends to the README |
| R1-14 | **F-2's trigger has fired** (a second radar source is ruled). | H-10 | Important | **RULED — D-35**: a narrow seam for provider colour tables in PLAN (L-2.5); F-2 and F-1 stay with the quality pass |
| R1-15 | **Scope and value questions:** a before-radar / can-follow split (B-9); the problem statement covers about half the release (P-1); stakeholders (P-2); MRMS heavy end as a written requirement (B-5); provider drift (P-7); per-advance render budget (P-6); `retract v0.1.0` (B-8); timeline (P-4). | B, P | Important/Minor | **RULED — D-36** (no split) **and D-37** (scope line, benefit claim, heavy end before SHIP, per-advance render cost, retract v0.1.0, no date) |
| R1-16 | **Clerical batch:** D-18's v0.1.0 row not written (B-2, H-15); D-0 quote truncated, three gate timings (B-3); lane scope wider than D-15's text (H-7, B-3); unconfirmed fuzz cause stated as fact (D-9); HR↔L mapping (D-7); small contradictions (D-8); required-reading "Blocking: nothing" (D-2); NOT RUN line (H-9); hour-long fuzz mode not built (H-14); first record commit before the gate fix (H-16); the atlas titled "Watchpost" (C-F7); gate history comments (C-F9); code comments (C-F2, C-F8); test helpers (C-F3, C-F6); count-table check (C-F4); stale local branches and a 14 MB test binary (H-2, H-3); README has no develop section (H-4); atlas page drift (H-8); exemptions file untracked (C-F5, H-13); audience order (Q4 fork). | many | Minor | **RULED — D-38**: done. Record corrections (D-0 in full, one gate timing, the HR↔L mapping, the brief's C-5 and the order of C-7/C-8, "Blocking" replaced by what gates exit, the diagram's hour-long fuzz marked NOT BUILT and its cause marked probable, D-15's lane scope noted, v0.1.0 ruling citations prefixed); OW-1 written; tooling (NOT RUN names the last tag; the atlas titled go-tuiMaps, split into testable parts, held to the documents by a test; the count table held to the fuzz targets by a test; helper duplication removed; two comments corrected); a README "Working on the library" section; exemptions: the rows are the record and the four-OS P10 output is filed; local cleanup done. **H-16, recorded rather than ruled:** `493f60f` (documents only) landed before the gate fix while the brief said every commit was blocked. The audience-order fork (Q4) is answered by `requirements.md`'s "What changes on screen" section. |

## Round 2 — dispatch

**Five fresh reviewers, the same lenses**, each given round 1's findings and dispositions and asked
two things in order: do round 1's fixes hold in the tree, and what is new. Scope: `ff51ccc..11d986a`
and the whole record. Reports: `reviewer-reports/round2-*.md`.

| Reviewer | Verdict |
|---|---|
| Code quality | **Do not ship** — three ways the gate reports green without checking; one "fixed" item false |
| Project hygiene + docs quality | **Do not proceed** — one "Done" false, three rulings weaker than ruled; about an hour of edits |
| Business quality + DISCOVER lens | **Proceed with conditions** — severe weather never examined; rows contradict each other |
| InfoSec persona | **Do not proceed** until four security rows are reworded — each can be met while its exploit works |
| Accessibility persona | **Do not proceed** — three contradictions need short rulings |

**The pattern of watchpost's DISCOVER repeated: most of round 2's findings are in round 1's
remediation**, several in the coordinator's own code and wording.

## Verified by the coordinator

| Claim | How checked | Result |
|---|---|---|
| NOT RUN still says "none exists yet" (code F-1, hygiene H-3) | `git describe --tags --abbrev=0` at `11d986a` | **Confirmed**: "No tags can describe"; `v0.1.0` is not an ancestor under squash-merge, and the test tags HEAD |
| The licence loop checks nothing in nested modules (code F-2) | `GOWORK=off go list -m all` in `tools/oracle` | **Confirmed**: "unknown revision v0.0.0"; the loop runs zero times |
| `fetch.Checked` bounds nothing (InfoSec S-2) | `internal/fetch/get.go:149-171` | **Confirmed**: the length is checked after the host fetcher returned the whole body; no deadline |
| Maximum age by mtime never expires a viewed tile (S-3) | `internal/tiles/disk.go:217-218` | **Confirmed**: a read rewrites the mtime at most hourly |
| The limiter's KILL fallback cannot fire (code F-4, S-8) | `scripts/gate:174-189`; the code reviewer reproduced it | **Confirmed** by reading; a TERM-ignoring child ran 20 s against a 2 s limit |
| v0.1.0 has no "nearby" (a11y F4) | `internal/describe/area.go:21-22` | **Confirmed — the coordinator's error in writing up D-29**: only `Inside` and `Outside` exist |
| OW-8 was self-issued (hygiene H-7) | D-23 requires a rulings row for every change | **Confirmed — the coordinator's error** |

## Round 2 — consolidated findings

| # | Finding | From | Severity | Disposition |
|---|---|---|---|---|
| R2-1 | **The gate still reports green without checking**: combined mode flags run nothing (C F-3); `--fuzz` with `go test -list` errors gives zero legs; the licence loop checks nothing in nested modules (C F-2); NOT RUN's tag lookup fails under squash-merge (C F-1, H-3); the KILL fallback cannot fire, and the flag file races (C F-4, S-8); the run log cannot name the tree it tested, records no override (`GATE_FUZZ_LIMIT`, `GATE_ROOT`) and tests only the green path (C F-5, F-6, F-10, H-5, S-9); D-31(4) has no test (C F-8); "STALLED" misnames a time limit (C F-9); the atlas rebuild instruction fails (C F-7); README omits perl and the workspace (H-2); `compose` sorts its caller's slice | C, H, S | Important | pending |
| R2-2 | **"The docs lane judges the tree, not the commit" was dropped from R1-10 without a ruling** | C F-5, H-4 | Important | pending — HUM LEAD fork |
| R2-3 | **Four security rows can be met while their exploit works**: L-1.14 (S-1: copy every host slice first — PNG and table, single image too — and validate the copy); L-7.3 (S-2: the library bounds bytes and time of every host fetch itself); L-9.3 (S-4: no write after Purge, CacheRoot change/off, or Close; replaced roots released); L-9.4 (S-3: age from fetch time, never refreshed by a read). New: a proxy disables the private-address check, and the reserved-range list is incomplete (S-5). Minor: a readable cache root (S-6); `ReadBack` reads before its limit, fields outside the budget, `ImageKey` collisions (S-7); L-5.5 does not require a clean scan (S-9); a patch adds `os.Getenv` to the renderer (S-10) | S-1–S-10 | Important | pending |
| R2-4 | **Which frame drives the `stale` word and the description** — the shown frame makes `stale` toggle every cycle and breaks L-1.10e; the newest frame contradicts L-1.3 as worded | A F1 | Critical | **RULED — D-39**: the newest non-gap frame drives both; L-1.3 reworded |
| R2-5 | **Motion without a place, and "heavier"** — no place, no motion; "heavier" is relative to the place's own class, which moves between frames | A F2, N-2, H D-14 | Important | pending — HUM LEAD fork |
| R2-6 | **The description rows cannot be built as written** — v0.1.0 has no "nearby" (the coordinator's error in D-29's write-up); area answers carry no label and merge every feature of an overlay; L-13.5's "independent of any place" has no surface; valid time is per overlay; severity exists only as a colour role | A F4, F5 | Important | pending |
| R2-7 | **Severe weather was never examined** — MRMS ≤ 48 dBZ, one warning over moderate rain, no overlapping warnings over heavy radar; the heavy-end evidence waits on live severe weather (the server keeps about two hours); L-2.3's test has no oracle | B N-1 | Critical | pending — HUM LEAD |
| R2-8 | **Nothing requires the warning to stay visible inside the blend** — at 20 % the tint is faint; the checker guards only against wash-out | B N-4 | Important | pending — HUM LEAD fork |
| R2-9 | **OW-8 (gofmt) was self-issued** — D-23 requires a ruling for every change | H-7 | Important | pending — HUM LEAD |
| R2-10 | **Intensity in words** (deferred by D-29) against WCAG 1.1.1 — let M1's non-visual grader decide | A F12 | Important | pending — HUM LEAD fork |
| R2-11 | **L-5.6's retraction against watchpost's fallback** — the ship-without-radar path ships watchpost on the version L-5.6 retracts, and also loses HR-3 and HR-7 | B N-5 | Important | pending — HUM LEAD |
| R2-12 | **Record corrections** — rows that contradict each other (L-11.2 vs L-11.4, L-11.3 missing D-27's per-class tint, L-8.8 under the fallback; L-13.7 must except L-1.12); M1's non-visual arm missing from the metrics; L-8.3 omits the hatch; the D-17 specimen must be drawn over radar; the flash ceiling's scope; the state read can say "playing" while the clock is frozen; the 16-colour arm of L-8.8 is trivial; wording rules as data shape; name, span and off-precedence in the state read; README says "its version" but the code sends `0.1.0-dev`; the brief's stale lines; the register lacks evidence and status columns and misrates RK-3, RK-10, RK-5, RK-4; the per-image cap; the additive-only check on L-7.3, L-7.4, L-9.3, L-11.1; units in L-12.2; caveats in L-2.1; palette order is file-name order; the programs cite work-branch hashes that die at release (H-1); OW-7 dead; OW-2's two due dates; required reading's "only" follow-ups list; the diagram misses `tools/atlas`; L-9.2's Purge wording; on-screen section misses default-off, frame time and the light-ground fallback; A17 uncited (covered by L-13.4) | H, B, A, S | Important/Minor | pending |
