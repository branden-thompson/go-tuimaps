---
title: "v0.2.0 — the gate's fuzz legs failing at their deadline (L-6.4, D-3, D-4)"
date: 2026-09-22
phase: intake → DISCOVER
sev: SEV-0
status: "FIXED BY AVOIDANCE (D-9, D-10) — cause never reproduced on demand; M6 watches it"
---

# The gate's fuzz legs fail at their deadline

**What is established, and what is not.** Kept as a running record so the next session does not
repeat a run or re-assert a hypothesis that failed.

## Observed

| Run | Leg | Result |
|---|---|---|
| Gate 1 | `FuzzAgree` (`tools/oracle`) | froze at 0 execs/s from 18 s; failed at 60 s, "context deadline exceeded" |
| Gate 2 | `FuzzDirectory` (`internal/archive`) | healthy (~170–260 k/s) to 60 s; failed at 61 s, "context deadline exceeded", just after a new input |
| Gate 2 | `FuzzHandIn` (`internal/overlay`) | healthy (~240 k/s) to 60 s; same failure, same moment |
| Every other leg, both runs | — | passed, including the vulnerability scan once `govulncheck` was installed |

No failing input was saved to `testdata/fuzz` in any run — consistent with an engine-side error,
not a crasher.

## Ruled out

- **A slow or hanging input** in `FuzzAgree`: 221 cached inputs timed through both decoders (slowest
  58 µs); a one-second watchdog on both decoders over ~9 minutes and >27 M executions never fired.

## Hypothesis 1 — a race at the time budget (from the Go 1.25.0 source)

With `-fuzztime <duration>` the coordinator wakes on the parent context's done channel and calls
`stop(ctx.Err())`, which suppresses the error only if it equals `fuzzCtx.Err()`
(`internal/fuzz/fuzz.go:129`). A parent context closes its done channel before cancelling its children
(`context/context.go:565`, then `:569`), so a coordinator on another core could see the deadline while
`fuzzCtx.Err()` is still nil.

**Neither confirmed nor refuted** — see the attempts below; none reproduced it.

## Reproduction attempts — all clean

| Attempt | Runs | Deadline failures |
|---|---|---|
| Trivial branching target, 3 s budget | 20 | 0 |
| `FuzzHandIn`, 10 s budget | 10 | 0 |
| `FuzzHandIn` + `FuzzDirectory`, 60 s (gate-shaped) | 6 | 0 |
| `FuzzHandIn`, 10 s, **cold cache** (fresh GOCACHE each run) | 8 | 0 |
| **The full gate** (D-8) | 3 | 0 |

44 isolated fuzz runs and three full gate runs, none failing — while **both** gate runs earlier the
same day failed. What the failing runs had in common was a nearly cold input cache and a high rate of
new inputs right up to the deadline (the first gate run started with 11 cached inputs for `FuzzAgree`;
the second found 94 new in `FuzzHandIn`'s last 60 s). Today's runs start warm and find few. **That
matters for hosted CI (L-6.1): a runner is always cold.**

## The fix taken (D-9, D-10)

No red run exists, and D-8 ruled the fallback: fix on the source evidence, verify by M6. The fuzz
legs now take a **count of runs** (`-fuzztime Nx`), which exits through the execution limit and
creates no deadline at all, so this failure cannot occur however cold the cache is.

**The departure is recorded, not hidden:** this is a fix without a reproduction. Its verification is
M6 — a stated number of consecutive clean full gate runs — and the honest claim is "the failure mode
is removed by construction", never "the bug was found and fixed".

**Calibration (D-10).** One flat count was wrong in both directions, so each target carries its own,
set from the leg's measured duration. A short sampling run is not a valid instrument here: it misses
the baseline-coverage gathering every leg does over the whole cached corpus, which is why the first
table ran 70% long and `FuzzTable` fifteen times short. Every leg now prints its duration, so drift
is visible in every run, and an uncalibrated target is named and run at a default.
