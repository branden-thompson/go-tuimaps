---
title: "v0.2.0 — REVIEW's scope: everything v0.1.0 shipped, and everything v0.2.0 adds"
date: 2026-10-04
phase: REVIEW (prepared at BUILD exit)
sev: SEV-0
authority: HUM LEAD
status: "OPEN — BUILD exit approved by the HUM LEAD 2026-10-04 (D-128); REVIEW reads the release commit whole."
---

# REVIEW's scope (plan task L10.11, L-5.4, D-72)

**Why the scope is the whole library.** v0.1.0 was tagged without a review (v0.1.0's release checklist,
"not started, and never run"; v0.2.0 D-18), so v0.2.0 is the library's first reviewed release, and its
REVIEW covers what v0.1.0 shipped as well as what v0.2.0 adds. BUILD exit's red team reviewed only the
change since `v0.1.0`; REVIEW reads the release commit whole.

## The code, by what v0.1.0 shipped

| Part | At `v0.1.0` | Reviewed |
|---|---|---|
| The public package | `clock.go`, `describe.go`, `doc.go`, `facts.go`, `guard.go`, `hostgeometry.go`, `kinds.go`, `look.go`, `map.go`, `overlays.go`, `places.go`, `shared.go`, `style.go`, `tiles.go`, `view.go`, `work.go` | every file at the release commit, and those v0.2.0 added: `playback.go`, `report.go` |
| Internal packages | `archive`, `colour`, `describe`, `fault`, `fetch`, `jsonsafe`, `mvt`, `overlay`, `project`, `render`, `rules`, `scene`, `style`, `testkit`, `textsafe`, `tiles`, `work` | each whole |
| Nested modules | `cmd/tuimaps`, `examples`, `tools/answer-key`, `tools/gen-assets`, `tools/oracle` (and `tools/atlas`, v0.2.0's) | each whole |
| The gate and CI | `scripts/gate`, `scripts/uat`; `.github/workflows/gate.yml` (v0.2.0) | whole |

## The requirements

- **v0.1.0's** 66 functional and non-functional rows: `06_docs/02_features/go-tuimaps/01-objectives/requirements.md`,
  and the parity matrix's 75-row denominator (`02-analysis/parity-matrix.md`).
- **v0.2.0's** L-1 to L-30, NFR-1 to NFR-3 and M1 to M6: `radar-loops/01-objectives/requirements.md`,
  NORMATIVE, every row naming the test that holds it.
- **The contract** a host builds against: `go-tuimaps/03-architecture-design/contract.md`, its section 10's
  every name and section 11's every sentence held by test, and section 12's ruled breaks.

## How it is reviewed

The full lens set, each reviewer blind and briefed from the repository's dispatch brief, every brief
naming this scope: the four axes (code quality, project hygiene, docs quality, business quality), the
REVIEW phase lens, and the personas the HUM LEAD confirms (PLAN confirmed A11y and InfoSec; BUILD exit ran
Performance under the standing clearance, for confirmation). The SHIP report shows each part above
reviewed (L10.11's test).
