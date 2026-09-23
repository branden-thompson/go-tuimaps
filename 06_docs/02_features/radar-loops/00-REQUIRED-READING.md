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
| Brief | `01-objectives/project-brief.md` — APPROVED (D-7), AMENDED (D-11), CORRECTED (D-23); with `requirements.md`, the body of issue #2 |
| Rulings | `02-analysis/rulings.md` — **every ruling lands here the moment it is made** |
| Requirements | `01-objectives/requirements.md` — **NORMATIVE** (D-23); wins over the brief on any conflict |
| Host | watchpost 0.18.0 (watchpost#22). Its HR-1..HR-5 are L-1..L-5, HR-6 is L-7, HR-7 L-8, HR-8 L-9, HR-9 L-1.5, HR-10 L-10; L-6 is this release's own |
| Follow-ups | Two lists, each with its own job: `06_docs/follow-ups.md` for work beyond this release (F-1 root clutter, F-2 pluggable architecture — quality pass), and `requirements.md`'s **Owed** table for work this release still owes |

## Blocking right now

**DISCOVER exit is gated by:** the DISCOVER exit red team (`08-reports/red-team-discover.md`) — round 1
dispositioned, further rounds until reviewers converge; then the DISCOVER report and HUM LEAD
approval. The D-17 specimen (OW-2) is due before PLAN commits to L-8.1, not before exit. The full gate takes about eighteen minutes; every run's time is in
`06_docs/gate-runs.md`.

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
