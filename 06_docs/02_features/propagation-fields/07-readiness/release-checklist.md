# Release checklist — what is done before v0.3.0 is tagged

Up: [plan](../04-development/implementation-plan.md) · [release notes](release-notes.md) · Carries: the brief's C-6, D-3, D-5

| Field | Value |
|---|---|
| Status | **Open.** v0.3.0 is paired with watchpost 0.19.0 and go-ionomaps (watchpost D-1, D-3); it ships when watchpost 0.19.0 does. Its newest release candidate is `v0.3.0-rc.11`, at `d01ca8e`, for the antimeridian's column (watchpost D-343). Sections 0 to 2 are ticked with their evidence; the tag waits for the HUM LEAD (watchpost D-344) |
| Who runs it | The coordinator prepares each row and shows its evidence; the HUM LEAD approves the tag |
| The machine check | `scripts/gate --release v0.3.0` refuses the final tag while any row of sections 0 to 2 ends in `[ ]` (section 3 follows the tag), and runs the pinned `govulncheck`. A release candidate, `v0.3.0-rc.N`, is not held to the rows |
| How a row is ticked | Its last cell becomes `[x]`, in the commit that holds its evidence, with the evidence named in the row |

## 0 · BUILD close-out

| # | Check | Evidence | Done |
|---|---|---|---|
| 0.1 | P10 clean: `a2dh validate` 100 % | `a2dh validate` 100 % (16 of 16) and `a2dh p10 check` 0 live findings in changed code at `2185bee`; its 7 ledger exemptions each RATIFIED by the HUM LEAD | [x] |
| 0.2 | Every plan task built or ruled out: P1 to P9 (P7 and P8, the instruments, on their own track) | P0 to P9 built, each task's test in the plan's Test-first column: P7.0 (`TestEveryGlobeCellKeepsItsContrast`) and P7.1 (M5's allocations, record and release step; watchpost D-350, D-49) built at SHIP, P7.2 as built, P9.4 for the antimeridian (D-47); P8.1, M1's sitting, ruled out (watchpost D-348, D-48) | [x] |
| 0.3 | The Owed table holds no open row due before SHIP | `requirements.md`, L-7: OW-24, OW-25, OW-17 and OW-19 built in P6.1 (`TestEveryClassHasAShadeKey`, `TestTheDefaultDepthFollowsTheTerminal`, `TestNoRoleRestsOnColourAlone`, `TestTheCheckedDialerExemptsOnlyTheNamedProxy`); OW-14, OW-15, OW-16, OW-18, OW-20 and OW-26 re-targeted to v0.4.0 (watchpost D-65); none due before SHIP | [x] |
| 0.4 | Hosted CI green on the release commit, both architectures | The hosted full gate green on linux/amd64 and linux/arm64, run 41144472299 at `2185bee` (`GATE_FUZZ_SCALE` 8); the quick gate green on every push. Dispatched again on the release commit at SHIP | [x] |

## 1 · The phases

| # | Check | Evidence | Done |
|---|---|---|---|
| 1.1 | BUILD exit: the blind red team, its findings remediated, the report approved | watchpost D-233: BUILD exited on `e889e72a`, the report approved (watchpost `08-reports/build-report.md`) | [x] |
| 1.2 | REVIEW: the release reviewed whole, the report approved | watchpost D-322: REVIEW exited on `v0.3.0-rc.10` after ten red-team rounds, the report approved (watchpost `08-reports/review-report.md`) | [x] |
| 1.3 | VALIDATE: every tier run, no skip path; the full suite with coverage | watchpost D-341: VALIDATE exited, the report signed off (watchpost `08-reports/validate-report.md`) - the full gate and the hosted full gate on both architectures; the suite with coverage at `2185bee`: 75 to 100 % a package (the root 89.6 %, `internal/fault` the lowest) | [x] |
| 1.4 | The README read for content, not only its numbers | Read for content against the code at VALIDATE (`57a9c29`): v0.3.0's fields and terminator named, the gate's time; its gate section brought to `--cost` and M5's record (`2185bee`) | [x] |
| 1.5 | The release notes say what changes in behaviour, what it adds, and what was measured | `release-notes.md`: what changes in behaviour, what it adds, what was measured - M1 to M6, M5 from [its record](m5-record.md) and a hosted linux/amd64 run - and its known limits | [x] |

## 2 · What is in the tree

| # | Check | Evidence | Done |
|---|---|---|---|
| 2.1 | Every gate is green on a fresh clone, run outside the working tree | `scripts/gate` (full) green on a fresh clone of `feature/propagation-fields` at `2185bee` from GitHub, outside the working tree, with the harness's two local files copied in so P10 runs (2026-10-31, 2,186 s; the tree `8db343770f2e`, the working tree's) | [x] |
| 2.2 | The public surface is the one written down, and the contract check against the last release - v0.2.2, its surface v0.2.0's - finds only additions plus the breaks contract section 12 lists | `TestPublicSurfaceSnapshot`; `TestNothingOfTheLastReleaseBreaksUnlisted` against `v0.2.2`, its surface v0.2.0's, held to the newest final tag below `fetch.Version` by `TestTheContractIsHeldToTheLastRelease` (m099) | [x] |
| 2.3 | The fixture and the embedded tiles are unchanged or re-pinned by ruling | Unchanged: `git diff v0.2.2 -- assets testdata/fixture` and `git diff v0.2.0 -- assets testdata/fixture` are empty | [x] |
| 2.4 | Licence files cover every module and the map data | The gate's licence leg, every module the packages and tests are built from, green in every full gate; the map data's notice in `assets/NOTICE` | [x] |
| 2.5 | No workspace file is tracked; no address or name that should not leave the tree | `git ls-files go.work go.work.sum` prints nothing; `TestNoPathOrAddressLeavesTheTree` scans every tracked file (1,179 at `2185bee`) for a home directory or an email address | [x] |
| 2.6 | `scripts/gate --release v0.3.0` green on the release commit | `scripts/gate --release v0.3.0` on the release commit of `release/v0.3.0`, which carries this tick: the check refuses the tag if it fails, M5's record included; its line is in that branch's `06_docs/gate-runs.md` | [x] |
| 2.7 | The library says it is the tag: `fetch.Version` equals the tag's version | `internal/fetch/fetch.go`: `Version = "0.3.0"`; `scripts/gate --release` refuses a tag the library does not say it is (`TestTheTagIsTheReleaseTheLibrarySays`) | [x] |
| 2.8 | M5, the frame (NFR-1, watchpost D-209): `BenchmarkGlobeFieldPan`, its contours drawn, on the reference machine at the release commit, at most 16 ms at 400 × 110 and 8 ms at 200 × 56; the figures and the machine named here (watchpost round 2 CN2-10, CN2-17) | [M5's record](m5-record.md) at the code's tree `e5cb59079964`, `2185bee`'s: 8.016 ms at 400 × 110 and 2.147 ms at 200 × 56, on an Apple M5 Pro, go1.26.10 (2026-10-30), held by `scripts/gate --release` (watchpost D-350); on a hosted linux/amd64 runner 17.8 ms and 4.8 ms, recorded, not held | [x] |

## 3 · The tags, in order (after the release check)

These rows follow the tag, so `scripts/gate --release` lists them and does not hold them; each is ticked
as it is done, at SHIP.

| # | Check | Evidence | Done |
|---|---|---|---|
| 3.1 | `release/v0.3.0` squash-merged to `main` from `feature/propagation-fields` | | [ ] |
| 3.2 | `v0.3.0` at the root, then the nested modules' tags against it | | [ ] |
| 3.3 | Every nested module is tidy against the tagged library | | [ ] |
| 3.4 | watchpost 0.19.0 pins `v0.3.0` and its gate is green on it | | [ ] |

## Ruled exceptions to the vulnerability scan

None. A row here names a GO- id in its first cell and the ruling that accepts it, and the release check
reads it.
