# go-tuiMaps — Objectives

Source of truth: [`../08-reports/project-brief.md`](../08-reports/project-brief.md) (APPROVED 2026-09-18).

## Problem Statement (locked, 5/5)

> People who follow weather and hazards from a terminal cannot see where conditions are relative to the places they care about — storms, fronts, wind and precipitation reach them only as text and numbers — and so they leave the terminal to find out whether something is coming toward them.

## Metrics of Success (ratified)

| ID | Metric | Type | Target |
|---|---|---|---|
| M1 | Hazard Placement | Primary | 100% of v1 overlay kinds placeable at 80×24 |
| M2 | Time to Placed View | Primary | **Target** ≤ 1 s warm / ≤ 3 s cold to full detail, never blank (D-30) — validated or revised at PLAN exit |
| M3 | Parity Coverage | Secondary | 100% of the frozen matrix's denominator (68 rows; see `parity-matrix.md`) |
| M4 | Embed Cost | Primary | **Target** ≤ 8 MB against a pinned fixture (D-29; fixture is Q7) — validated or revised at PLAN exit; flat heap over 1 h |
| M5 | Host Independence | Secondary | pass (dual-instance + headless render) |
| M6 | Correction Count | Maintenance | lower is better |

## Requirements (from the brief)

R-1 parity with TerminalMap · R-2 embeddable in Watchpost · R-3 overlays from sources independent of the basemap · R-4 *(derived)* high-volume host-supplied overlay data.
T-A Go · T-B Bubble Tea v2 / Lipgloss v2 host · T-C overlay sources independent of the tile source · T-D published under the `branden-thompson` account.

## Upstream baseline

| Project | Pin | Notes |
|---|---|---|
| TerminalMap (psmux, MIT) | `3b960723b24b781a1a9a04c165803cd1e98e6e3c` — v0.1.0, 2026-07-18 | 16 Rust source files, 3,459 lines. The parity baseline (D-11). |
| MAPSCII (rastapasta, MIT) | `4fe9a60a0c9da952dadc5214a9ca5c68c447fdf8` — 2023-01-10 | Reference for the delta review (AI-2). |

## Scope of v1

Set by rulings D-11 onward; gathered in [`requirements.md`](requirements.md). A first-release slice (v0.1.0) is a question before HUM LEAD (red-team Q3). Words used in these documents are in the [glossary](glossary.md).

## Directives

LEVEL-1 · SEV-0 (HUMAN LEAD) · FULL GIT · FULL DOCS · FULL REPORTS · FULL DIAGRAMS · FULL RCC · FULL PLAN · FULL TDD · Theme: BRTOPS
