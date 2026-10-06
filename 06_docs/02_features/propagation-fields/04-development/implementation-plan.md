---
title: "go-tuiMaps v0.3.0 — Propagation fields — IMPLEMENTATION PLAN"
date: 2026-10-06
phase: PLAN
sev: SEV-0
authority: HUM LEAD
status: "DRAFT for the PLAN gate. No code: signatures, shapes, test descriptions and order only (watchpost 0.18.0 D-13)."
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
| P7 | The globe's cost (M5) and gate trust (M6) | NFR-1, NFR-3 | P1 to P6 |

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
| P1.1 | `Grid.OverWater` and `Image.MaskedByWater` carried to the scene; `Input` flags go | `TestHostCanFlipEither`, `TestAFieldStopsAtTheShoreByDefault`, `TestContoursFollowTheirFieldOverWater`, `TestAChangedWaterChoiceRedraws` |
| P2.1 | `muf`/`fof2` presets, breaks in MHz, tokens, ramps passing `Check` on both grounds at truecolor, 256 and 16 | `TestTheMUFAndFoF2RampsPassCheckOnBothGrounds`, `TestEveryPresetHasLegendWordsAndAUnit` |
| P2.2 | The generic host ramp (`Type.Colours`, or `low`/`middle`/`high` with defaults); the cell ink for interpolated colours chosen from P0's specimens (C-2) | `TestAHostTypeIsFilledFromLowMiddleHigh`, `TestTheHostRampDefaultsWhenUnset`, `TestAHostRampIsHeldToCheckRamp`, `TestABadHostRampIsWarned` |
| P2.3 | No grid draws nothing silently | `TestNoGridDrawsNothingSilently` (M2's matrix) |
| P3.1 | Units on labels; labels on both crossings (behaviour change, changelog) | `TestContourLabelsCarryTheirUnit`, `TestAFieldBandedEitherWayIsLabelled` |
| P3.2 | `Answer.ValueUnit` for every field | `TestEveryFieldAnswerCarriesItsUnit` (M3) |
| P4.1 | `terminator` and `night` tokens; the night fill without alert words | `TestTheTerminatorAndNightTokensHaveDefaults`, `TestTheNightTintCarriesNoAlertWords` |
| P4.2 | `Terminator`, `NightSide`, within 0.5° of NOAA's solar-position algorithm (watchpost A-13) | `TestTheTerminatorMatchesTheEphemeris`, `TestNightSideIsTheComplement` |
| P4.3 | Day or night in `Report`; a non-colour terminator mark; the night tint readable under `CheckRamp`'s colour-vision rules with `SafeRamps` on | `TestReportSaysDayOrNight`, `TestTheTerminatorIsMarkedWithoutColour`, `TestNightTintKeepsTheFieldReadable` |
| P5.1 | `ScaleClass` decides `SafeRamps` and `Warnings` (behaviour change, changelog) | `TestSafeRampsCoversEveryScaleToken`, `TestWarningsReportEveryScaleToken` |
| P6.1 | OW-24, OW-25, OW-17, OW-19 | `TestEveryClassHasAShadeKey`, `TestTheDefaultDepthFollowsTheTerminal`, `TestNoRoleRestsOnColourAlone`, `TestTheCheckedDialerExemptsOnlyTheNamedProxy` |
| P7.1 | M5: a 2° whole-globe frame with labels and terminator, measured at 200×56 and 400×110 (the round-1 Performance reviewer measured 3.9 ms and 14.8 ms with contours on today's code); target set from the dry run (watchpost D-49) | a benchmark, recorded |

## Release

- A release candidate is tagged after each WP that changes the public surface (`v0.3.0-rc.N`).
- watchpost pins the newest candidate.
- v0.3.0 is tagged after its own REVIEW and VALIDATE, before watchpost 0.19.0's SHIP.
