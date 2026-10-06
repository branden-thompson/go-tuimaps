---
title: "go-tuiMaps v0.3.0 (propagation fields) — HUM LEAD rulings"
date: 2026-10-06
phase: DISCOVER
sev: SEV-0
authority: HUM LEAD
status: "LIVE — every ruling is written here the moment it is made."
---

# v0.3.0 — rulings

Recorded verbatim. A correction to a ruling is a new row, never an edit of an old one.

v0.3.0 is paired with watchpost 0.19.0 and go-giro-data (watchpost D-3). A decision that binds more than
one project is written in watchpost's log
(`watchpost/06_docs/02_features/propagation-overlays/02-analysis/rulings.md`, cited "watchpost D-n") and
restated here; this log holds the library's own. v0.2.0's log is `../radar-loops/02-analysis/rulings.md`.

| # | Date | Question | HUM LEAD verbatim | Ruling |
|---|---|---|---|---|
| D-0 | 2026-10-06 | Restates watchpost D-3 | see watchpost D-3 | go-tuiMaps v0.3.0 is a paired release, run as v0.2.0 was: its own branch, rulings and gates; watchpost pins its release candidates in BUILD and ships on its final tag. Branch `feature/propagation-fields`, cut from `main` (`d1c4d5e`) with v0.2.0's REFLECT and release checklist carried from `feature/radar-loops` (watchpost A-7). |
| D-1 | 2026-10-06 | Restates watchpost D-32 | see watchpost D-32 | HR-1: a field continues over water per overlay (a field on `Grid`, the mask on `Image`); FR-12/D-87's promise kept; `TestHostCanFlipEither` written. |
| D-2 | 2026-10-06 | Restates watchpost D-33 | see watchpost D-33 | HR-2: `muf` and `fof2` presets, and FR-15's generic host ramp from `low`/`middle`/`high`, held to `CheckRamp`; a host type that would draw nothing is refused or warned. |
| D-3 | 2026-10-06 | Restates watchpost D-34 | see watchpost D-34 | HR-3: contour labels carry units; labels on north-south crossings too, for every field; `Answer.ValueUnit` for every field. |
| D-4 | 2026-10-06 | Restates watchpost D-35 | see watchpost D-35 | HR-4: `terminator` and `night` tokens with `Terminator(at)` and `NightSide(at)`. HR-5: `RoleFor(type, value)`, the value always in the point's label (OW-17). |
| D-5 | 2026-10-06 | Restates watchpost D-36 | see watchpost D-36 | `SafeRamps` decided by `ScaleClass`, with a test over every scale token. |
| D-6 | 2026-10-06 | Lock v0.3.0's problem statement as drafted | "Approve as written" | **LOCKED:** "A developer embedding go-tuiMaps cannot draw a global field of their own kind so that it can be read: the field stops at every coast, has no colours or legend of its own and can draw nothing at all, its labels carry no unit and are missing where its lines run east to west, and nothing shows where day ends, so the person reading the map sees a broken or empty picture, or numbers without meaning." |
| D-7 | 2026-10-06 | Adopt the metrics M1 to M6 as drafted | "Adopt; M4 for v0.3.0's rows only" | **ADOPTED:** M1 field read, M2 no silent blank, M3 units everywhere, M4 promise truth (this release's rows only), M5 globe cost, M6 gate trust; targets set in PLAN. |
| D-8 | 2026-10-06 | Approve the brief as presented | "Approve" | **APPROVED.** Intake closed for v0.3.0; its `requirements.md` follows in DISCOVER. No issue of record opened (not asked for). |
| D-9 | 2026-10-06 | Restates watchpost D-49 | see watchpost D-49 | M5's target is set from PLAN's dry run by its own ruling, before PLAN exits. |
| D-10 | 2026-10-06 | Restates watchpost D-50 | see watchpost D-50 | **L-5 (station points by value) leaves v0.3.0**, amending D-4's HR-5 half; carried as a watchpost follow-up with the station dots. |
| D-11 | 2026-10-06 | Restates watchpost D-65 | see watchpost D-65 | OW-17, OW-19, OW-24 and OW-25 are v0.3.0 requirements (L-7); OW-14, OW-15, OW-16, OW-18, OW-20 and OW-26 are re-targeted to v0.4.0. |
| D-12 | 2026-10-06 | Restates watchpost D-66 | see watchpost D-66 | `feature/radar-loops` and `feature/go-tuimaps` bundled to `~/` and deleted. |
