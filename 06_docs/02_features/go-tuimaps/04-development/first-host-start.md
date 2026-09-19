# How the first host starts before a remote exists

Up: [implementation plan](implementation-plan.md) · Carries: D-19, D-81

| Field | Value |
|---|---|
| Phase | PLAN |
| Date | 2026-09-19 |
| Why this exists | No remote exists until SHIP (D-19), so the first host cannot fetch the library the usual way. The plan cited "the local-override recipe" in tasks 12.24 and 14.19 without writing it (red-team rounds 1 and 2: PL-BZ-4, P2-PRD-17). |

## The recipe

Both repositories sit on the same machine. Nothing here is committed to either.

| # | In | Do | Why |
|---|---|---|---|
| 1 | The host's repository | Require the library's module path at the placeholder version `v0.0.0`, and add one `replace` line that points that path at the library's working tree on disk | The toolchain then builds the host against the local tree; no network is touched |
| 2 | The host's repository | Keep both lines **out of every commit** — make the change in a scratch branch, or in a workspace file the host's ignore rules already cover | A `replace` that reaches the host's main branch breaks every other checkout |
| 3 | The host's repository | Run the host's own gate | This is where the two-module allow-list meets reality: the host already carries both modules (D-81), so its module graph must gain nothing but the library |
| 4 | The library's repository | Run task 12.24's test | It builds a throw-away host module by exactly these steps, so the recipe cannot rot unnoticed |

At SHIP the `replace` line is deleted and the placeholder version becomes the first tag. Nothing else changes.

## What the spike before the tag must show (task 14.19)

| Shown | Risk it answers |
|---|---|
| One map, drawn inside the host's own layout, from embedded tiles | — |
| One overlay, handed in from the host's own data | RS-2 — the overlay contract, in a real host |
| The pump, written in the host's idiom: `OnPending` wakes it, two goroutines, a message to the interface when the map changes, a clean stop on quit | RS-4 — background work run by the host |
| The host's memory figure before and after, by the host's own protocol | RS-7 — confirmatory only |

What the spike finds goes into the integration review ([template](../07-readiness/integration-review-template.md)).
