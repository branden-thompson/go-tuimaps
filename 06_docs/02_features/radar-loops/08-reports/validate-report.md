---
title: "v0.2.0 — VALIDATE report"
date: 2026-10-04
phase: VALIDATE
sev: SEV-0
authority: HUM LEAD
status: "APPROVED by the HUM LEAD 2026-10-05 (D-141); SHIP proceeds."
---

# VALIDATE report — go-tuiMaps v0.2.0

VALIDATE runs every tier the release claims, on the code REVIEW approved (D-139), with no skip path; it
also closes what REVIEW left to it (OW-22) and makes the release check able to pass a final tag.

## The tiers

| Tier | Result |
|---|---|
| The full gate, local (the reference machine) | six consecutive clean full runs; then one failed: `FuzzAgree` found a three-point ring the decoders close differently, fixed by D-140, and a scanner check that read the vulnerability database's reachability as the scanner's version, fixed to read the binary's build record. The failure reset M6 (any failed full run does); five consecutive clean full runs followed, and M6 is met |
| The hosted full gate | green on linux/amd64 and linux/arm64 for the first time: run 37222838055 at `33dc8d7`, its fuzz budgets divided by 8 for the 4-core runners (D-129). Dispatched again on the release commit at SHIP |
| The hosted quick gate | green on every push since REVIEW's flake fix |
| The suite with coverage | 75-100 % a package: the root package 89.9 %; `fault` 75.0 %, the lowest, an error type and two closed lists |
| The hour soak (M4) | green: a 12-frame loop adds 597 KB; over an hour, refreshed 60 times, sharing eleven frames each time, the heap held within ±5 % (1590 KB at the start, 1633 KB at most) |
| L-12.5, a frame advance against 15 ms | 4.4-4.6 ms on the near-empty map (`BenchmarkFrameAdvance`); 7.7-8.2 ms on a dense one (`BenchmarkFrameAdvanceDense`: a region at zoom 6, a noisy twelve-frame web map loop in six classes, 1,000 alert areas, 20 places), five runs each, the reference machine. OW-22 done |
| P10 | clean, the gate's own leg (D-136) |

**Blind spots, beside the numbers.** The soak and the measures are one machine's. The dense advance
allocates about 12,000 times an advance against about 560 on the near-empty map: within L-12.5's time,
unpinned (OW-26, v0.3.0). The hosted full gate runs a scaled fuzz budget; the full budget runs on the
reference machine alone. Windows and Linux are cross-compiled and not tested here; the hosted gate tests
Linux on both architectures.

## What VALIDATE changed

| Change | Why |
|---|---|
| `BenchmarkFrameAdvanceDense`, `TestADenseAdvanceDraws` | OW-22: L-12.5 measured on a dense map |
| `TestNoPathOrAddressLeavesTheTree` | checklist row 2.5's exposure scan existed only by hand: every tracked file is scanned for a home directory or an email address outside reserved and listed public ones |
| D-140: a ring is closed only at four points or more, as the proven decoder judges it (`TestARingOfThreePointsIsClosedAgain`) | `FuzzAgree` found it; invisible in rendering |
| The scanner's pin read from the binary's build record | `govulncheck -version` asks the vulnerability database too, and an unreachable one failed every module's scan |
| The release check lists a section headed "after the release check" without holding it (`TestTheTagsAreDoneAfterTheCheck`) | the tags, and watchpost pinning them, follow the check: held, no final tag could ever pass |
| The README read for content | the colour-alone claim corrected; `Report` named as the text alternative to the braille; the pump's one rule |
| The release notes' Measured section, the checklist's evidence | the numbers above, M1 to M6, each row's evidence |

## What SHIP has left

Row 2.1 (every gate green on a fresh clone, outside the working tree), row 2.6 (`scripts/gate --release
v0.2.0` on the release commit), and section 3: the squash to `main` through `release/v0.2.0`, the tags,
the nested modules tidy against the tag, and watchpost pinning v0.2.0 with its gate green (where OW-21's
`Advancing` fix lands).
