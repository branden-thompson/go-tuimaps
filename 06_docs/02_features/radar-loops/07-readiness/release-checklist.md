# Release checklist — what is done before v0.2.0 is tagged

Up: [plan](../04-development/implementation-plan.md) · Plan tasks L10.3, L10.4, L10.11 · Carries: L-5.3, L-5.4,
L-5.5, D-18, D-69, D-102

| Field | Value |
|---|---|
| Status | **Open.** v0.2.0 is the library's first reviewed release (L-5.4): v0.1.0's checklist was never run, and this one covers everything v0.1.0 shipped as well as what v0.2.0 adds |
| Who runs it | The coordinator prepares each row and shows its evidence; the HUM LEAD approves the tag |
| The machine check | `scripts/gate --release v0.2.0` refuses the final tag while any row below ends in `[ ]`, and runs the pinned `govulncheck` (L-5.5). A release candidate, `v0.2.0-rc.N`, is not held to the rows (D-69) |
| How a row is ticked | Its last cell becomes `[x]`, in the commit that holds its evidence, with the evidence named in the row |

## 0 · BUILD close-out

| # | Check | Evidence | Done |
|---|---|---|---|
| 0.1 | P10 clean: `a2dh validate` 100 % | L10.13, D-103: 0 live findings | [x] |
| 0.2 | F-3: an image is answered in its own projection | L10.14: `internal/describe/projection_test.go`, `projection_test.go` | [x] |
| 0.3 | OW-10: a near colour is counted and warned of | L10.15, D-104: `internal/overlay/near_test.go`, `near_test.go` | [x] |
| 0.4 | L10.12: the plans and the integration map are kept in step by a test | L10.12, D-72: `plan_map_test.go`; the plan records the map commit it was reconciled against | [x] |
| 0.5 | L10.5: hosted CI green on the release commit, both architectures | — | [ ] |
| 0.6 | L10.7 / OW-4: `FuzzAgree` run ten times under the limiter, each recorded | `gate-fuzz-deadline.md`, L10.7: ten green after D-106; the freeze explained | [x] |
| 0.7 | L10.8 / M4: the hour soak run once and recorded | `scripts/gate --soak`, 2026-10-04: 597 KB for the loop; over 1h0m0s the heap held within ±5 % (1588 KB at the start, 1631 KB at most) | [x] |
| 0.8 | L10.10 / M6: five consecutive clean full gate runs, counted from `06_docs/gate-runs.md` | — | [ ] |
| 0.9 | L10.9 / M1: the HUM LEAD's scores for both arms recorded | `m1-sitting.md`, sitting 1 (2026-10-04): visual 5/5 pass, non-visual 3/5 fail; remedied by D-122 and held by `TestTheMotionOfSittingOneIsWithinAPoint` (D-123, no second sitting) | [x] |
| 0.10 | OW-12: the triggered MRMS capture taken, or MRMS shipped marked unverified with its release-note line | No severe day before SHIP: L-2.3's planned outcome - `Unverified` stays set, `table-fallback` warns, and `release-notes.md` names the range above about 48.5 dBZ. Re-checked at the tag | [x] |
| 0.11 | The Owed table holds no open row due before SHIP | `requirements.md`, Owed (OW-6 is due at the quality pass) | [ ] |

## 1 · The phases

| # | Check | Evidence | Done |
|---|---|---|---|
| 1.1 | BUILD exit: the blind red team on all four axes and the phase lens, its findings remediated, the report approved | `08-reports/` | [ ] |
| 1.2 | REVIEW covers everything v0.1.0 shipped (L-5.4, L10.11): its scope written as v0.1.0's packages and requirements, and each reviewed | `08-reports/` | [ ] |
| 1.3 | VALIDATE: every tier run, no skip path; the full suite with coverage | `08-reports/` | [ ] |
| 1.4 | The README read for content, not only its numbers (the framework's pre-SHIP audit) | — | [ ] |
| 1.5 | The release notes say what the release adds, what a host should know, and what was measured | `release-notes.md` | [ ] |

## 2 · What is in the tree

| # | Check | Evidence | Done |
|---|---|---|---|
| 2.1 | Every gate is green on a fresh clone, run outside the working tree | — | [ ] |
| 2.2 | The public surface is the one written down (`TestPublicSurfaceSnapshot`), and the contract check against v0.1.0 finds only additions plus the breaks contract section 12 lists (D-58) | — | [ ] |
| 2.3 | The fixture and the embedded tiles are unchanged or re-pinned by ruling | — | [ ] |
| 2.4 | Licence files cover every module and the map data | — | [ ] |
| 2.5 | No workspace file is tracked; no address or name that should not leave the tree (the exposure scan) | — | [ ] |
| 2.6 | `scripts/gate --release v0.2.0` green on the release commit: the pinned `govulncheck` finds nothing reachable, or each finding has a ruled exception row below | — | [ ] |
| 2.7 | The library says it is the tag: `fetch.Version` equals the tag's version; `scripts/gate --release` refuses otherwise | `internal/fetch/fetch.go`, `TestTheTagIsTheReleaseTheLibrarySays` | [x] |

## 3 · The tags, in order

| # | Check | Evidence | Done |
|---|---|---|---|
| 3.1 | `release/v0.2.0` squash-merged to `main` from `feature/radar-loops` (D-1) | — | [ ] |
| 3.2 | `v0.2.0` at the root, then the nested modules' tags against it | — | [ ] |
| 3.3 | Every nested module is tidy against the tagged library: `GOWORK=off go mod tidy -diff` clean in `cmd/tuimaps`, `examples` and `tools/*`, so `go install .../cmd/tuimaps@v0.2.0` works | — | [ ] |
| 3.4 | Watchpost 0.18.0 pins `v0.2.0` (its W8.1) and its gate is green on it | — | [ ] |

## Ruled exceptions to the vulnerability scan

None. A row here names a GO- id in its first cell and the ruling that accepts it, and the release check
reads it.
