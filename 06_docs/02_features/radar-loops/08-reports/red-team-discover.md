---
title: "v0.2.0 Radar loops — DISCOVER exit red team"
date: 2026-09-23
phase: DISCOVER (RCC) — exit
sev: SEV-0
authority: HUM LEAD
status: "ROUND 1 RECEIVED — findings consolidated below; dispositions are made one at a time with the HUM LEAD and recorded here and in rulings.md"
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
Summaries are kept with the session's scratch output.

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
| R1-7 | **D-17 can fail silently:** a label that does not fit is dropped whole; dashes must differ from the line-overlay dash; no second cue at 16 colours. D-14's check covers class distance only, not outline/label contrast, and truecolor only. | A8, A9 | Important | pending |
| R1-8 | **Description gaps:** area answers lack label and severity; image answers lack value, unit and "approximate"; nothing when no place is registered. | A11, A12, A13 | Important | pending |
| R1-9 | **Security rows missing from L-1/L-7/L-9/L-10:** frames borrowed not copied (F1); the budget ignores retained PNG bytes, no frame-count bound, playback interval must be library-owned (F2); host fetcher unbounded, `Checked` unwired (F3); `Fetcher` silently deferred (F4); Purge scope, in-flight writes, future mtimes (F5); two allow-lists, http under a host fetcher (F7). Minor: swallowed cache errors (F6), the cache as a viewing record (F8), no scan recorded at a tag (F9). | F1–F9 | Important | pending |
| R1-10 | **Gate defects:** `FUZZ_TIMEOUT` inert, so a stall hangs (C-Q10, D-6); the docs lane exits 0 when nothing changed and judges the tree not the commit (C-F10, H-6); a module `go list` cannot load reads as "empty" (C-F11); `git diff`'s exit status lost; the decoder fix has no unit test (C-F1); gate runs leave no trace, so M6 cannot be counted (H-5). | C, H-5, H-6, D-6 | Important | pending |
| R1-11 | **Measurement soundness:** the "71 %" error (D-10); no blind-spot sections (D-12); the two blend results measure different pairs (D-13); D-20 under-states 3 MB (holds about five dot-resolution loops, D-14); the programs are not filed, so nothing re-runs (P-9, D-11). | D-10–D-14, P-9 | Important | pending |
| R1-12 | **The record is on one disk:** neither feature branch is pushed; issue #2 cites an unresolvable commit. | H-1 | Important | pending — HUM LEAD (outward-facing) |
| R1-13 | **README's user-agent token claim is false**, widening M5. | D-3 | Important | pending |
| R1-14 | **F-2's trigger has fired** (a second radar source is ruled). | H-10 | Important | pending |
| R1-15 | **Scope and value questions:** a before-radar / can-follow split (B-9); the problem statement covers about half the release (P-1); stakeholders (P-2); MRMS heavy end as a written requirement (B-5); provider drift (P-7); per-advance render budget (P-6); `retract v0.1.0` (B-8); timeline (P-4). | B, P | Important/Minor | pending |
| R1-16 | **Clerical batch:** D-18's v0.1.0 row not written (B-2, H-15); D-0 quote truncated, three gate timings (B-3); lane scope wider than D-15's text (H-7, B-3); unconfirmed fuzz cause stated as fact (D-9); HR↔L mapping (D-7); small contradictions (D-8); required-reading "Blocking: nothing" (D-2); NOT RUN line (H-9); hour-long fuzz mode not built (H-14); first record commit before the gate fix (H-16); the atlas titled "Watchpost" (C-F7); gate history comments (C-F9); code comments (C-F2, C-F8); test helpers (C-F3, C-F6); count-table check (C-F4); stale local branches and a 14 MB test binary (H-2, H-3); README has no develop section (H-4); atlas page drift (H-8); exemptions file untracked (C-F5, H-13); audience order (Q4 fork). | many | Minor | pending |
