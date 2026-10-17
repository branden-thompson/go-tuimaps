# Release checklist — what is done before v0.3.0 is tagged

Up: [plan](../04-development/implementation-plan.md) · [release notes](release-notes.md) · Carries: the brief's C-6, D-3, D-5

| Field | Value |
|---|---|
| Status | **Open.** v0.3.0 is paired with watchpost 0.19.0 and go-ionomaps (watchpost D-1, D-3); it ships when watchpost 0.19.0 does |
| Who runs it | The coordinator prepares each row and shows its evidence; the HUM LEAD approves the tag |
| The machine check | `scripts/gate --release v0.3.0` refuses the final tag while any row of sections 0 to 2 ends in `[ ]` (section 3 follows the tag), and runs the pinned `govulncheck`. A release candidate, `v0.3.0-rc.N`, is not held to the rows |
| How a row is ticked | Its last cell becomes `[x]`, in the commit that holds its evidence, with the evidence named in the row |

## 0 · BUILD close-out

| # | Check | Evidence | Done |
|---|---|---|---|
| 0.1 | P10 clean: `a2dh validate` 100 % | | [ ] |
| 0.2 | Every plan task built or ruled out: P1 to P9 (P7 and P8, the instruments, on their own track) | | [ ] |
| 0.3 | The Owed table holds no open row due before SHIP | | [ ] |
| 0.4 | Hosted CI green on the release commit, both architectures | | [ ] |

## 1 · The phases

| # | Check | Evidence | Done |
|---|---|---|---|
| 1.1 | BUILD exit: the blind red team, its findings remediated, the report approved | | [ ] |
| 1.2 | REVIEW: the release reviewed whole, the report approved | | [ ] |
| 1.3 | VALIDATE: every tier run, no skip path; the full suite with coverage | | [ ] |
| 1.4 | The README read for content, not only its numbers | | [ ] |
| 1.5 | The release notes say what changes in behaviour, what it adds, and what was measured | | [ ] |

## 2 · What is in the tree

| # | Check | Evidence | Done |
|---|---|---|---|
| 2.1 | Every gate is green on a fresh clone, run outside the working tree | | [ ] |
| 2.2 | The public surface is the one written down, and the contract check against v0.2.0 finds only additions plus the breaks contract section 12 lists | | [ ] |
| 2.3 | The fixture and the embedded tiles are unchanged or re-pinned by ruling | | [ ] |
| 2.4 | Licence files cover every module and the map data | | [ ] |
| 2.5 | No workspace file is tracked; no address or name that should not leave the tree | | [ ] |
| 2.6 | `scripts/gate --release v0.3.0` green on the release commit | | [ ] |
| 2.7 | The library says it is the tag: `fetch.Version` equals the tag's version | | [ ] |

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
