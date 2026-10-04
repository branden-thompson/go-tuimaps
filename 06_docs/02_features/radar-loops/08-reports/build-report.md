---
title: "v0.2.0 — BUILD report"
date: 2026-10-04
phase: BUILD (exit)
sev: SEV-0
authority: HUM LEAD
status: "PRESENTED — BUILD's work is in; the BUILD-exit red team's round 1 is remediated; the decisions below are the HUM LEAD's before BUILD exits."
---

# BUILD report — go-tuiMaps v0.2.0

## Where BUILD stands

- **Every plan task is built** but the human-graded one (L10.9, M1), and every WP-L11 task watchpost's UAT
  asked for (L11.1 - L11.35). The plan's rows carry their as-built notes; the requirements file names a
  test for every row (`01-objectives/requirements.md`).
- **`a2dh validate` 100 %**: P10 clean with no exemption added in v0.2.0's close-out (D-103).
- **The release check** (`scripts/gate --release`) holds the final tag to the checklist, the version and a
  pinned `govulncheck` (nothing reachable). **Hosted CI** runs the gate on linux/amd64 and linux/arm64
  (D-105).
- **Measured:** M4 - a 12-frame loop adds about 0.6 MiB over an empty map and holds within ±5 % while it
  plays (the hour soak, `scripts/gate --soak`, 2026-10-04: 597 KB, the heap 1588 KB at the start and 1631 KB at most over the hour); 24 region frames
  about 5.8 MiB. `FuzzAgree` ten runs green after D-106 (L10.7). The worst-case picture 0.21 s a frame.
- **M6** counts from `06_docs/gate-runs.md` (`m6_test.go`); five consecutive clean full runs are needed
  before SHIP.
- **The red team** (`red-team-build.md`): eight blind reviewers; one Critical found and fixed (a refreshed
  loop never decoded its new frames); every other finding fixed, recorded with its reason, or listed below.

## Decisions for the HUM LEAD before BUILD exits


1. A11y 4.1 — A large alert drawn straight from the host's memory (the borrowed-geometry path) loses every
   non-colour cue: no "· SEVERE" word, no outline digit, no hatch. At NoColour it is an anonymous line.
   Rec: draw the word and digits for borrowed alert rings (one pass over the runs already read), and report
   a drop in Frame.Dropped. Counter: changes what watchpost draws for large alerts, late in the release.
2. A11y 3.1 — `LoopState.Playing` stays true while the host freezes the clock (a separate `Advancing`
   exists); L-1.10d says Playing is true only while frames advance. Rec: amend L-1.10d to the code (Playing =
   the listener pressed play; Advancing = frames moving) and make the example read Advancing. Counter: the
   requirement as written is the more honest state for a non-visual user.
3. A11y 3.3 / BUILD-lens 5 — motion relative to the view is the centre only (no edges, no in-view flag), and
   L-1.10g's per-loop state read does not exist (D-67 made playback map-wide). Rec: amend L-1.12 and L-1.10g
   to what was built, under D-57/D-67, and record the view-edge motion as a v0.3.0 follow-up. Counter: these
   are accessibility promises written into the normative file; amending them is narrowing the release.
4. A11y 3.4 — no Motion entry at all when there is no heavier rain, too few frames, or frames still
   decoding: the listener cannot tell silence from "clear". Rec: v0.3.0 follow-up (an additive reason
   field). Counter: cheap to add now as a new field.
5. A11y 4.2 — quake age and fire strength are colour-only roles. Rec: v0.3.0 follow-up, and the contract
   tells hosts to put age/strength in the label now (watchpost already labels quakes "M3.1 2:14 PM").
6. Business F7 — MRMS pixels more than 30 from the heavy-end gradient draw as nothing; on a severe day the
   heaviest cores could vanish. Rec: draw them as the top class until OW-12's capture rules otherwise.
   Counter: inventing a class for an unknown colour; D-44 governs.
7. Business F13 — blends that leave no class visible could be "drawn over" instead (less work, same
   picture). Rec: v0.3.0 follow-up.
8. Perf F6 / Code 2.2 / BUILD-lens 7 — the image budget does not count a refreshed loop's stand-in or spare
   pictures, so memory runs to about twice the budget during a refresh, against "errs high, never low".
   Rec: count them in ImageUse (reporting) and state the transient in the contract; do not refuse on them.
   Counter: counting may push a host near its budget into refusals.
9. Perf F8 — the shared decoded-picture set is a fixed 1 MiB FIFO, too small for a loop, so two maps barely
   share. Rec: v0.3.0 follow-up (scale it to the budget). Counter: watchpost opens one map at a time.
10. InfoSec S-4 / BUILD-lens 12 — CheckedDialer refuses a host's own proxy at a private address. Rec: state
    the limit in the contract now; a CheckedDialerFor(proxy) is v0.3.0. Counter: the safe option is broken
    in the common enterprise case.
11. Code C-8 — the P10 exemptions ledger is gitignored (.a2dh*), so "P10 clean" cannot be reproduced from
    the repo. Rec: commit `.a2dh-p10-exemptions.yml` (no edits to it). Counter: it is harness configuration.
12. BUILD-lens 2 / Perf F11 — Purge, Verify and CacheRoot walk the whole disk cache under the map lock;
    Render stalls for the walk. Rec: state it in the contract now; move the walks off the lock in v0.3.0.
13. BUILD-lens 13 — no recorded before-you-write-code READY verdict at BUILD start; PLAN reviewers had
    said "not ready". Rec: rule on the omission (the blockers were closed by D-66..D-73 before L1).
14. M1 (L10.9) — human-graded: five loops are recorded in testdata/loops and the demo app plays them
    (`tuimaps --loop testdata/loops/<name>`); set the ground truth, then score the non-visual arm, then the
    visual arm (D-71).
