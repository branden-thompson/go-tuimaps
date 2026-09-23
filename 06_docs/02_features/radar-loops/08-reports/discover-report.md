---
title: "go-tuiMaps v0.2.0 — Radar loops — DISCOVERY REPORT"
date: 2026-09-23
phase: DISCOVER (RCC) — exit
level: LEVEL-1
sev: SEV-0
authority: HUM LEAD
directives: FULL GIT; FULL DOCS; FULL REPORTS; FULL DIAGRAMS; FULL RCC; FULL PLAN; FULL TDD; FULL INST
branch: feature/radar-loops
issue: "branden-thompson/go-tuimaps#2"
status: "APPROVED by the HUM LEAD 2026-09-23 (D-49) as presented; the DISCOVER phase artefact"
---

# go-tuiMaps v0.2.0 — Radar loops — DISCOVERY REPORT

## Bottom line up front

**DISCOVER is complete and recommends proceeding to PLAN.** v0.2.0 turns the library's single radar
picture into a **loop** a listener can see move, or hear described, and repairs the contract its only
host reads. It is also **the library's first reviewed release** (D-18). The requirement set is written
down, normative and traceable: `01-objectives/requirements.md`.

What a reader should know before the detail:

1. **The picture a listener sees was settled by drawing it.** Specimen 29 rendered live radar under a
   live Flash Flood Warning. From it: the warning's tint blends over the radar (D-14). And a real defect:
   with no colour, and at 16 colours, radar erases a warning's outline, label, marker and `stale` word.
2. **Every number was measured.** The MRMS table and loop memory came from live data through the
   library's own calls. The programs, their inputs and raw output are filed so anyone can re-run them
   (D-32).
3. **Two adversarial rounds found most of their defects in the work of the round before**, several in
   the coordinator's own code and wording. They were caught by reading the tree and by mutation tests,
   not by re-reading the account.
4. **The gate was hardened along the way.** It used to report green without having checked; it now
   refuses to. A fuzz stall fails at a wall-clock limit instead of hanging. Every run is logged with
   the tree it tested (D-31, D-40, D-46).

## The problem, locked

> **"A developer embedding go-tuiMaps can show where precipitation is, but not where it is going: the
> library draws one radar frame, so the person reading the map cannot see a storm's motion, and the
> host has no way to give the library the frames that would show it."**

Locked at D-5. The release covers more than the lock does (D-37): loops; the contract repaired; and
the first reviewed release, with the accessibility and security work the red team added. Nothing is
split out (D-36), and there is no target date.

## What changes on screen

- **Radar moves**: playback off, slow or normal. **It is off until the host turns it on** (D-26), and
  one standard playback API lets a host wire controls and Settings straight to it.
- **An alert over radar lets the radar show through**: the tint blends in, the outline and label stay
  on top, and the warning stays visible inside the blend (D-14, D-45). On a light ground, if no blend
  passes, the warning is drawn by outline, label and dash over the radar (D-27).
- **Severity reads without colour**: a severity word in the label and a distinct outline dash per level,
  at every depth (D-17, D-28). A specimen over radar is owed before PLAN commits to it.
- **The shown frame's time, or "gap", is written on the map** (D-25). The newest frame drives the
  `stale` word (D-39).

## What a listener hears

- **Motion, observed, never forecast**: where the heavier rain was at the oldest usable frame and at
  the newest, and whether it came closer. It is relative to a named place, or **relative to the view
  when there is none** (D-24, D-42).
- **The alerts shown**: each by name, severity and valid time. For a named place, whether it is inside,
  outside or nearby each one (D-29, D-43). **Never "you"**: a selected place is not the listener's
  position.

## What DISCOVER produced

| Artefact | What it holds |
|---|---|
| `00-REQUIRED-READING.md` | What a cold session must know; what gates exit |
| `01-objectives/project-brief.md` | The approved brief, corrected (D-23); with `requirements.md`, the body of issue #2 |
| `01-objectives/requirements.md` | **Normative.** 83 requirement rows, each traced to its source, with its instrument or NO INSTRUMENT YET; M1–M6; 11 risks with evidence and status; the owed list |
| `02-analysis/rulings.md` | D-0..D-48, the HUM LEAD's words verbatim |
| `02-analysis/wave1-findings.md`, `wave2-measurements.md` | Three surveys read from the tree; two measurements run on the machine, each with its limits |
| `02-analysis/programs/` | The measurement programs, their inputs and raw output |
| `08-reports/red-team-discover.md` | Two rounds, every finding fixed or ruled; `reviewer-reports/` holds all ten reports verbatim |
| Specimen 29 (`../go-tuimaps/02-analysis/specimens/`) | Radar under a warning: as built, radar on top, three blend strengths, and a control |

## What the research measured

- **MRMS draws from a small palette.** 111 colours were seen across 16 distinct times, while its legend
  is a 1,232-colour gradient. A table sampled from the legend leaves 80 % of rain pixels unmatched under
  exact matching. The table ships as the observed palette, labelled approximate (D-19).
- **Loop memory**: 12 frames cost +0.59 MiB a map at 298×152 (the braille dot grid of a 149×38 map),
  and +2.19 MiB at 596×304. A host-settable image budget per map defaults inside NFR-3 (D-20).
- **The tint blend holds only on a dark ground, at about 20 %.** No blend strength passes on a light
  ground, hence the named fallback (S29-7, D-27).

## Risks

Eleven, each with its evidence and status. The ones that shape PLAN:

- **RK-2 and RK-11: the MRMS heavy end on a severe-weather day.** The evidence depends on the weather.
  A past outbreak from IEM's archive is owed at PLAN entry, and a live MRMS capture has a named
  trigger. If no severe day comes before SHIP, MRMS ships with its heavy end marked unverified, and
  the library warns (D-44).
- **RK-1: a loop that freezes or re-decodes while every unit test passes.** The renderer, the store
  and the frames must land together.
- **RK-4: watchpost's radar waits for all of v0.2.0.** Accepted: no split, no date. Watchpost's
  fallback loses radar, the view bound and the colour-independent pattern (D-36, D-48).

## Critical analysis

| | Reviewers | Verdicts | What it found |
|---|---|---|---|
| Round 1 | 5 (4 axes + DISCOVER lens + InfoSec + accessibility) | 4 do not proceed, 1 conditional | No current requirement set; motion with no non-visual form; the no-colour defect wider than recorded; security threats never carried into requirements; a gate that could hang or pass with nothing checked |
| Round 2 | 5, given round 1's dispositions | 3 do not proceed, 2 conditional | **Mostly defects in round 1's remediation**: a "Done" that was false, security rows that could be met while the exploit worked, rows that contradicted each other, and severe weather never examined |

Every finding is fixed or ruled, D-23..D-48. **The HUM LEAD closed the reviews at round 2** rather than
running a third, matching process depth to the release's blast radius (D-48). PLAN's own exit review
takes up the scrutiny from there.

## Gates

```
QUALITY GATE REPORT | go-tuiMaps v0.2.0 | SEV-0 | DISCOVER exit
-------------------------------------------------------------
  [PASS] requirements_documented     : requirements.md, normative; 83 rows traced; 63 marked NO INSTRUMENT YET for PLAN
  [PASS] constraints_identified      : C-1..C-8 in the brief, measured
  [PASS] risks_assessed              : 11, each with evidence, mitigation and status
  [PASS] critical_analysis_complete  : 2 rounds, every finding dispositioned (D-23..D-48)
  [PASS] a2dh validate               : 100 % (16/16)
  [PASS] scripts/gate                : green — see 06_docs/gate-runs.md
  [PASS] p10                         : 0 live findings
  [PASS] human_approval              : D-49, 2026-09-23 — "approved; GO 4 PLAN"
  OVERALL: ALL PASS — DISCOVER closed
-------------------------------------------------------------
```

## What PLAN inherits

1. **63 requirements marked NO INSTRUMENT YET.** PLAN names the test or check for each, signatures
   and API shape only (no code in PLAN).
2. **The loop API as one design**: frames, the renderer's frame advance, the store keeping frames
   across refreshes, and one standard playback API (L-1, L-1.13).
3. **Numbers PLAN sets from measurement**: the image budget's default, the "nearby" default, the
   blend strength and the visibility floor, the frame-rate ceiling, the per-advance render cost and
   M4's target.
4. **Owed at or before PLAN**: the archive outbreak specimen (OW-11), the D-17 specimen over radar
   (OW-2), and diagnosing why 16 colours draws every rain class as `░` (OW-9).
5. **The contract rewrite** (L-4), which every other requirement lands in.

## Recommendation

**Proceed to PLAN.** The record answers what is being built, for whom, at what cost, and how each
promise will be held. Where it cannot answer yet, the requirement says so.

## Source documents

`00-REQUIRED-READING.md` · `01-objectives/project-brief.md` · `01-objectives/requirements.md` ·
`02-analysis/rulings.md` · `02-analysis/wave1-findings.md` · `02-analysis/wave2-measurements.md` ·
`02-analysis/programs/` · `08-reports/red-team-discover.md` · `08-reports/reviewer-reports/` ·
specimen 29 · `06_docs/gate-runs.md`
