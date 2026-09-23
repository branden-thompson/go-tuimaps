---
title: "v0.2.0 Radar loops — REQUIRED READING, every session and after every compaction"
date: 2026-09-22
phase: ALL
sev: SEV-0
authority: HUM LEAD
status: "MANDATORY.  Re-read at session start and after any context compaction, for the life of v0.2.0."
---

# Read this before touching v0.2.0

**Why this file exists.** v0.2.0 runs in parallel with watchpost 0.18.0 and ships first (watchpost
D-11). The condition for running two releases at once: **the record is kept current at every step,
and every lesson that can be a failing test becomes one.**

## Where things are

| | |
|---|---|
| Branch | `feature/radar-loops`, cut from `feature/go-tuimaps` (full history) → squash-merged `release/v0.2.0` at SHIP (D-1) |
| Phase | **DISCOVER OPEN** since 2026-09-23 (D-12), ahead of watchpost's PLAN. Watchpost's 0.18.0 DISCOVER record is reference material (`watchpost/06_docs/02_features/observer-maps/`). Update this row at every phase transition |
| Brief | `01-objectives/project-brief.md` — APPROVED (D-7), AMENDED (D-11); the body of issue #2 |
| Rulings | `02-analysis/rulings.md` — **every ruling lands here the moment it is made** |
| Host | watchpost 0.18.0 (watchpost#22) — its HR-1..HR-10 are this release's L-1..L-10 (L-1.5 is HR-9) |
| Follow-ups | `06_docs/follow-ups.md` only. F-1 root clutter, F-2 pluggable architecture — quality pass |

## Blocking right now

Nothing. The fuzz-deadline failure was fixed by construction (D-9, D-10;
`scripts/gate` uses per-target run counts) and the gate is green in about seventeen minutes.

## Machine notes

- **The docs lane (D-15):** a change that is Markdown alone runs `scripts/gate --docs`, which
  every module's tests pass through. The script refuses any other file. Everything else runs
  the full `scripts/gate`.

- `govulncheck` lives in `~/go/bin`, which must be on PATH for `scripts/gate`.
- Never run the gate while watchpost's `make verify` runs: the mutant sweep times out under the load,
  and verify's cache-clean breaks concurrent Go builds.

## The rules that cost something

1. **Rulings one at a time**, with evidence, options, a recommendation and the strongest
   counter-argument, recorded verbatim. Silence is not consent.
2. **Additive only** to the v0.1.0 contract; a breaking change is a HUM LEAD ruling (C-7).
3. **No code in PLAN** — signatures and API shape only (watchpost D-13).
4. **Never call a table exact that was derived** (L-2.2).
5. **A gate is obeyed or ruled on**, never re-run until it passes.
6. **The host chooses** (watchpost D-23, P-1..P-3) — never architect into a corner.
7. **No AI attribution** in commits, PRs or tracked files.
