# Integration review — template (D-60)

Up: [first-host start](../04-development/first-host-start.md) · Carries: NFR-4, NFR-5, NFR-22, D-29, D-48, D-60, D-85, D-90, D-92, D-94

| Field | Value |
|---|---|
| Phase | Written in PLAN; filled in after the first host has integrated v0.1.0 |
| Why this exists | HUM LEAD ruled (D-60) that the remaining overlay shapes are built only after a **written** review of the first integration and HUM LEAD's GO. The plan cited this template without writing it (P2-PRD-17). |
| Who fills it in | The coordinator, from the integration itself — not from memory. Every answer names its evidence: a commit, a measurement, a line of the host's code |
| Who decides | HUM LEAD: GO, GO with changes, or stop |

## 1 · What was integrated

| Question | Answer |
|---|---|
| Library version and commit; host version and commit | |
| Which maps, at which sizes, with which overlays, from which of the host's data sources | |
| How the host runs the pump: how many goroutines, how it is woken, how it stops | |

## 2 · What the host needed that the contract lacked

One row for each thing the host had to work around, wanted and did not have, or got wrong at first.

| Need | What the host did instead | Contract section | Proposed change | Breaking? (NFR-22) |
|---|---|---|---|---|
| | | | | |

## 3 · What the contract has that the host did not use

Unused surface is a cost. One row each; say whether it should stay.

| Call or field | Why unused | Keep? |
|---|---|---|
| | | |

## 4 · Measured in the real host

| Measure | Target | Measured | How |
|---|---|---|---|
| Live memory added, three maps | 4 MB (D-48, D-85) | | |
| Peak memory added | 8 MB (D-29) | | |
| `CacheUse`: need against cap, in the host's real views (D-90) | need ≤ cap in ordinary use | | |
| Cold start to full detail | 3 s (NFR-5) | | |
| A changed frame | the host's own window cost (NFR-4) | | |
| `Set` on the largest overlay the host really sends (D-92) | a few milliseconds | | |

## 5 · The first-hour mistakes

Which of these did the integration actually make, and did the library catch it? A discarded `Set` error · a mistyped id · a wrong unit · a pump that never ran · a borrowed buffer reused too early.

| Mistake | Made? | Caught by |
|---|---|---|
| | | |

## 6 · M1, in the host

For each scenario the host can now show: does the frame answer "where is it, relative to my place", and does the description match the key.

| Scenario | Frame | Description | Note |
|---|---|---|---|
| | | | |

## 7 · Does the plan for the rest still hold

| Deferred part (contract, section 9) | Still wanted? | Still additive? | Changed by what was learned |
|---|---|---|---|
| Wind · image loops · tile-image provider · block renderer · PMTiles source · pointer operations · flash, pulse, tours | | | |
| Reopened by this review: per-layer label margin and clustering (D-94) | | | |

## 8 · Recommendation

GO, GO with the changes listed in section 2, or stop — with the strongest argument against the recommendation.
