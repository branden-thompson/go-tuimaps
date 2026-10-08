---
title: "go-tuiMaps v0.3.0 — Propagation fields — IMPLEMENTATION PLAN"
date: 2026-10-06
phase: PLAN
sev: SEV-0
authority: HUM LEAD
status: "APPROVED at the PLAN gate (watchpost D-139). No code: signatures, shapes, test descriptions and order only (watchpost 0.18.0 D-13)."
---

# Implementation plan

**Goal.** Build `requirements.md` (approved at watchpost D-84) as v0.3.0's release candidates, which watchpost
pins in its BUILD, and tag v0.3.0 first (watchpost D-3).

**Every task follows the library's order:**
1. Specimen or test first, watched failing.
2. Make the change.
3. Run `scripts/gate`.
4. Mutation check, anchored in `06_docs/mutants` (M4: each row anchored to a mutant the gate kills, D-7).

A new public name gets a `contract.md` §10 row, which `contract_test.go` holds both ways. A behaviour change
(L-3.2, L-6.1) gets a changelog entry naming it.

## Order

| WP | What | Rows | Needs |
|---|---|---|---|
| P0 | Specimens first: a whole-globe grid and a host-type grid at every depth on both grounds (C-5) | — | — |
| P1 | Fields over water, per overlay; `TestHostCanFlipEither` at last | L-1.1 to L-1.4 | P0 |
| P2 | Colour for a host's own field: `muf` and `fof2` presets; the generic host ramp; no silent blank | L-2.1 to L-2.5 | P0 |
| P3 | Units and labels: units on labels; north-south crossings; `Answer.ValueUnit` for every field | L-3.1 to L-3.3 | P2 |
| P4 | Day and night: tokens, helpers, the night fill, the words, a non-colour terminator | L-4.1 to L-4.4 | P2 |
| P5 | `SafeRamps` for every scale | L-6.1 | P2 |
| P6 | Owed rows: shade key, default depth, colour-only marks, checked dialer | L-7.1 to L-7.4 | P2 |
| P7 | The globe's cost (M5, target D-96: the profiled per-cell contrast, blending and allocations fixed) and gate trust (M6) | NFR-1, NFR-3 | P1 to P6 |
| P8 | M1's sitting: the script, a recorded globe field, the answer key for 10 places, both arms (watchpost D-117) | M1 | P1 to P6 |

Size (estimates, L-F14): P0 to P8, about 9 batches. The agent's estimate, not a measurement.

## Shapes (signatures only)

```go
// L-1: per overlay
type Grid struct {
    // existing fields unchanged
    OverWater bool // a field continues over the sea when true
}
type Image struct {
    // existing fields unchanged
    MaskedByWater bool
}

// L-2.1
func MUFGrid(id string, grid Grid, validAt time.Time) Overlay
func FoF2Grid(id string, grid Grid, validAt time.Time) Overlay

// L-2.2: a host's own type takes colours, or the low/middle/high defaults
type Type struct {
    // existing fields unchanged
    Colours []RGB // one per class; empty means the low/middle/high ramp
}

// L-4.2
func Terminator(at time.Time) []LonLat
func NightSide(at time.Time) [][]LonLat
```

## P0 to P7, test first

| # | Task | Test first |
|---|---|---|
| P0.1 | Whole-globe 2° grid specimen and host-type grid specimen, every depth, both grounds | the specimens' golden files; a host-type grid without lines draws nothing today, recorded as the defect L-2.4 closes |
| P1.1 | `Grid.OverWater` and `Image.MaskedByWater` carried to the scene; `Input` flags go | `TestHostCanFlipEither`, `TestAFieldStopsAtTheShoreByDefault`, `TestContoursFollowTheirFieldOverWater`, `TestAChangedWaterChoiceRedraws`, `TestWaterNeverMasksImage` |
| P2.1 | `muf`/`fof2` presets, breaks in MHz, tokens, ramps passing `Check` on both grounds at truecolor, 256 and 16 | `TestTheMUFAndFoF2RampsPassCheckOnBothGrounds`, `TestEveryPresetHasLegendWordsAndAUnit` |
| P2.2 | The generic host ramp (`Type.Colours`, or `low`/`middle`/`high` with defaults); the cell ink for interpolated colours chosen from P0's specimens (C-2) | `TestAHostTypeIsFilledFromLowMiddleHigh`, `TestTheHostRampDefaultsWhenUnset`, `TestAHostRampIsHeldToCheckRamp`, `TestABadHostRampIsWarned` |
| P2.3 | No grid draws nothing silently | `TestNoGridDrawsNothingSilently` (M2's matrix) |
| P2.4 | L-2.5 and L-2.6: a host type's legend carries colours and unit; the reach edge's non-colour mark, never a contour, its combination with a fill stated (watchpost D-127) | `TestAHostTypesLegendHasColoursAndUnit`, `TestTheReachIsMarkedWithoutColour` (every depth, both grounds) |
| P3.1 | Units on labels; labels on both crossings (behaviour change, changelog) | `TestContourLabelsCarryTheirUnit`, `TestAFieldBandedEitherWayIsLabelled` |
| P3.2 | `Answer.ValueUnit` for every field | `TestEveryFieldAnswerCarriesItsUnit` (M3) |
| P4.1 | `terminator` and `night` tokens; the night fill without alert words | `TestTheTerminatorAndNightTokensHaveDefaults`, `TestTheNightTintCarriesNoAlertWords` |
| P4.2 | `Terminator`, `NightSide`, within 0.5° of NOAA's solar-position algorithm (watchpost A-13) | `TestTheTerminatorMatchesTheEphemeris`, `TestNightSideIsTheComplement` |
| P4.3 | Day or night in `Report`; a non-colour terminator mark; the night tint readable under `CheckRamp`'s colour-vision rules with `SafeRamps` on | `TestReportSaysDayOrNight`, `TestTheTerminatorIsMarkedWithoutColour`, `TestNightTintKeepsTheFieldReadable` |
| P5.1 | `ScaleClass` decides `SafeRamps` and `Warnings` (behaviour change, changelog) | `TestSafeRampsCoversEveryScaleToken`, `TestWarningsReportEveryScaleToken` |
| P6.1 | OW-24, OW-25, OW-17, OW-19 | `TestEveryClassHasAShadeKey`, `TestTheDefaultDepthFollowsTheTerminal`, `TestNoRoleRestsOnColourAlone`, `TestTheCheckedDialerExemptsOnlyTheNamedProxy` |
| P7.0 | **Before P7's optimisation (A-7):** a test over P0's whole-globe specimen, with the night tint and labels, at every colour depth on both grounds, asserts every cell's contrast (`internal/colour/contrast.go`'s rule: 3:1 for lines, 4.5:1 for text); watched passing before the optimisation and kept after | `TestEveryGlobeCellKeepsItsContrast` |
| P7.1 | M5: a 2° whole-globe frame with labels, terminator, night and the reach overlay (P-7), measured at 200×56 and 400×110. watchpost's dry run measured v0.2.0 after a pan, with fill and labelled contours: 5.7 ms and 22.2 ms (fill alone 2.4 and 9.1 ms; most of the cost in per-cell colour contrast and blending, about one allocation per cell). **Target (watchpost D-96): at most 16 ms at 400×110 and 8 ms at 200×56**, after a pan, with fill, labelled contours, terminator and night | the gate asserts allocations and work per cell (D-119); the wall-clock target is a release step on the reference machine, failing when its record is missing; a linux/amd64 run recorded |
| P7.2 | M6 and the release record (watchpost D-117, D-119): 5 consecutive full gate runs with no unattributable failure, logged in `06_docs/gate-runs.md`; the release step's record of M5 holds the machine model (`sysctl -n machdep.cpu.brand_string`), the Go version, the commit and the figures, and the release check fails when it is missing or names another commit (Code N12, N14) | the release check fails on a planted record for another commit |
| P8.1 | M1: a recorded 2° MUF field (from go-ionomaps' output or the dry run's week, kept outside the tree), 10 places with the key fixed before the sitting; the picture arm and the `Report` arm; 8 of 10 each (D-117) | the key's own check: `Report` at each place names the keyed band |

## Release

- A release candidate is tagged after each WP that changes the public surface (`v0.3.0-rc.N`).
- watchpost pins the newest candidate.
- v0.3.0 is tagged after its own REVIEW and VALIDATE, before watchpost 0.19.0's SHIP.
