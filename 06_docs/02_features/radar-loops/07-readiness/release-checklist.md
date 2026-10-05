# Release checklist — what is done before v0.2.0 is tagged

Up: [plan](../04-development/implementation-plan.md) · Plan tasks L10.3, L10.4, L10.11 · Carries: L-5.3, L-5.4,
L-5.5, D-18, D-69, D-102

| Field | Value |
|---|---|
| Status | **Open.** v0.2.0 is the library's first reviewed release (L-5.4): v0.1.0's checklist was never run, and this one covers everything v0.1.0 shipped as well as what v0.2.0 adds |
| Who runs it | The coordinator prepares each row and shows its evidence; the HUM LEAD approves the tag |
| The machine check | `scripts/gate --release v0.2.0` refuses the final tag while any row of sections 0 to 2 ends in `[ ]` (section 3 follows the tag), and runs the pinned `govulncheck` (L-5.5). A release candidate, `v0.2.0-rc.N`, is not held to the rows (D-69) |
| How a row is ticked | Its last cell becomes `[x]`, in the commit that holds its evidence, with the evidence named in the row |

## 0 · BUILD close-out

| # | Check | Evidence | Done |
|---|---|---|---|
| 0.1 | P10 clean: `a2dh validate` 100 % | L10.13, D-103: 0 live findings | [x] |
| 0.2 | F-3: an image is answered in its own projection | L10.14: `internal/describe/projection_test.go`, `projection_test.go` | [x] |
| 0.3 | OW-10: a near colour is counted and warned of | L10.15, D-104: `internal/overlay/near_test.go`, `near_test.go` | [x] |
| 0.4 | L10.12: the plans and the integration map are kept in step by a test | L10.12, D-72: `plan_map_test.go`; the plan records the map commit it was reconciled against | [x] |
| 0.5 | L10.5: hosted CI green on the release commit, both architectures | The hosted full gate green on linux/amd64 and linux/arm64, run 37222838055 at `33dc8d7` (`GATE_FUZZ_SCALE` 8, D-129); the quick gate green on every push since. Dispatched again on the release commit at SHIP | [x] |
| 0.6 | L10.7 / OW-4: `FuzzAgree` run ten times under the limiter, each recorded | `gate-fuzz-deadline.md`, L10.7: ten green after D-106; the freeze explained | [x] |
| 0.7 | L10.8 / M4: the hour soak run once and recorded | `scripts/gate --soak`, 2026-10-04 at VALIDATE: 597 KB for the loop; over 1h0m0s, refreshed 60 times, the heap held within ±5 % (1590 KB at the start, 1633 KB at most) | [x] |
| 0.8 | L10.10 / M6: five consecutive clean full gate runs, counted from `06_docs/gate-runs.md` | Five consecutive clean full runs after D-140 reset the count (2026-10-04), `m6_test.go` reading `06_docs/gate-runs.md` | [x] |
| 0.9 | L10.9 / M1: the HUM LEAD's scores for both arms recorded | `m1-sitting.md`, sitting 1 (2026-10-04), recorded: the visual arm passed 5/5; **the non-visual arm failed, 3/5**. D-122 changed the motion; `TestTheMotionOfSittingOneIsWithinAPoint` is a regression pin on the same loops, not evidence (D-123, D-127; no second sitting). RK-8 stays open, transferred to OW-21 | [x] |
| 0.10 | OW-12: the triggered MRMS capture taken, or MRMS shipped marked unverified with its release-note line | No severe day before SHIP: L-2.3's planned outcome - `Unverified` stays set, `table-fallback` warns, and `release-notes.md` names the range above about 48.5 dBZ, and what an unmatched colour on a severe day means (D-113); `TestTheMRMSTableIsTheObservedPalette` holds the flag | [x] |
| 0.11 | The Owed table holds no open row due before SHIP | `requirements.md`, Owed: none due before SHIP - OW-6 at the quality pass, OW-12 on trigger, OW-14 before v0.3.0, OW-21 watchpost, OW-23 v1, the rest v0.3.0; OW-22 done at VALIDATE | [x] |

## 1 · The phases

| # | Check | Evidence | Done |
|---|---|---|---|
| 1.1 | BUILD exit: the blind red team on all four axes and the phase lens, its findings remediated, the report approved | `08-reports/red-team-build.md`, `build-report.md`: two rounds remediated; approved D-128 (corrected by D-129) | [x] |
| 1.2 | REVIEW covers everything v0.1.0 shipped (L-5.4, L10.11): its scope written as v0.1.0's packages and requirements, and each reviewed | `08-reports/review-scope.md`, `review-report.md`: eight blind reviewers over v0.1.0 and v0.2.0 whole; approved D-139 | [x] |
| 1.3 | VALIDATE: every tier run, no skip path; the full suite with coverage | `08-reports/validate-report.md`: every tier run - the full gate, the hosted full gate on both architectures, the hour soak, the dense measure; the suite with coverage, 75-100 % a package | [x] |
| 1.4 | The README read for content, not only its numbers (the framework's pre-SHIP audit) | README read at VALIDATE: the colour-alone claim corrected, `Report` named as the text alternative, the pump rule stated | [x] |
| 1.5 | The release notes say what the release adds, what a host should know, and what was measured | `release-notes.md`: what it adds, what a host should know, what was measured, M1 to M6 | [x] |

## 2 · What is in the tree

| # | Check | Evidence | Done |
|---|---|---|---|
| 2.1 | Every gate is green on a fresh clone, run outside the working tree | `scripts/gate` (full) green on a fresh clone of `feature/radar-loops` at `32a3fc9`, from GitHub, outside the working tree (2026-10-05); the P10 leg not run there, by design, with no harness in a clone | [x] |
| 2.2 | The public surface is the one written down (`TestPublicSurfaceSnapshot`), and the contract check against v0.1.0 finds only additions plus the breaks contract section 12 lists (D-58) | `TestPublicSurfaceSnapshot`; `TestNothingOfTheLastReleaseBreaksUnlisted` (D-137) against v0.1.0 | [x] |
| 2.3 | The fixture and the embedded tiles are unchanged or re-pinned by ruling | Unchanged since v0.1.0: `git diff v0.1.0 -- assets testdata/fixture` is empty | [x] |
| 2.4 | Licence files cover every module and the map data | The gate's licence leg, every module the packages and tests are built from; the map data's notice in `assets` | [x] |
| 2.5 | No workspace file is tracked; no address or name that should not leave the tree (the exposure scan) | No workspace file tracked; `TestNoPathOrAddressLeavesTheTree` scans every tracked file for a home directory or an email address | [x] |
| 2.6 | `scripts/gate --release v0.2.0` green on the release commit: the pinned `govulncheck` finds nothing reachable, or each finding has a ruled exception row below | `scripts/gate --release v0.2.0` on the release commit of `release/v0.2.0`, which carries this tick: the check refuses the tag if it fails; its line is in that branch's `06_docs/gate-runs.md` | [x] |
| 2.7 | The library says it is the tag: `fetch.Version` equals the tag's version; `scripts/gate --release` refuses otherwise | `internal/fetch/fetch.go`, `TestTheTagIsTheReleaseTheLibrarySays` | [x] |

## 3 · The tags, in order (after the release check)

These rows follow the tag, so `scripts/gate --release` lists them and does not hold them; each is ticked
as it is done, at SHIP.

| # | Check | Evidence | Done |
|---|---|---|---|
| 3.1 | `release/v0.2.0` squash-merged to `main` from `feature/radar-loops` (D-1) | PR #3 squash-merged to `main` as `3cacdfc` on a fully green run - the hosted full gate on both architectures (run 37259603864) and the quick gates - the release check green on that commit; issue #2 closed | [x] |
| 3.2 | `v0.2.0` at the root, then the nested modules' tags against it | — | [ ] |
| 3.3 | Every nested module is tidy against the tagged library: `GOWORK=off go mod tidy -diff` clean in `cmd/tuimaps`, `examples` and `tools/*`, so `go install .../cmd/tuimaps@v0.2.0` works | `cmd/tuimaps`, `examples`, `tools/gen-assets` and `tools/oracle` require the published `v0.2.0`; every nested module `GOWORK=off go mod tidy -diff` clean (2026-10-05) | [x] |
| 3.4 | Watchpost 0.18.0 pins `v0.2.0` (its W8.1) and its gate is green on it | — | [ ] |

## Ruled exceptions to the vulnerability scan

None. A row here names a GO- id in its first cell and the ruling that accepts it, and the release check
reads it.
