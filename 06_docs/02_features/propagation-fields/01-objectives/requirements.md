---
title: "go-tuiMaps v0.3.0 — Propagation fields — REQUIREMENTS"
date: 2026-10-06
phase: DISCOVER
sev: SEV-0
authority: HUM LEAD
status: "DRAFT — approved at the DISCOVER gate; once approved it wins over the brief on any conflict, and every change is a row in 02-analysis/rulings.md."
---

# Requirements

Each row traces to the brief's L-n and to the ruling that makes it necessary. **What, never how.** PLAN
carries signatures only (watchpost 0.18.0 D-13).

Each row's **Instrument** names what will hold it at BUILD exit: a test by its planned name, a contract
row, a specimen or a metric. A planned name is a commitment PLAN may rename, never drop. Every new public
name gets a row in `contract.md` §10, which `contract_test.go` holds both ways.

**Before any change (C-5):** a specimen of a whole-globe grid and one of a host-type grid, rendered at
every depth on both grounds, so each change below is seen against what v0.2.0 draws.

## L-1 — A field over the sea, per overlay (watchpost D-32; D-1)

| # | Requirement | Instrument |
|---|---|---|
| L-1.1 | A host chooses, per field overlay, whether it continues over water; the default is unchanged (a field stops at the shore, D-32 of v0.1.0) | `TestHostCanFlipEither` (written at last, v0.1.0 plan task 09.7), `TestAFieldStopsAtTheShoreByDefault` |
| L-1.2 | A host chooses, per image, whether water masks it; the default is unchanged (an image is never masked, D-87) | `TestHostCanFlipEither`, `TestWaterNeverMasksImage` (kept) |
| L-1.3 | Contour lines follow the same choice as their field | `TestContoursFollowTheirFieldOverWater` |
| L-1.4 | Frame reuse sees the choice (a changed choice redraws) | `TestAChangedWaterChoiceRedraws` |

## L-2 — Colour for a host's own field (watchpost D-33; D-2)

| # | Requirement | Instrument |
|---|---|---|
| L-2.1 | `muf` and `fof2` presets, in MHz, with their breaks, tokens, ramps on both grounds at truecolor, 256 and 16 colours, legend words and `Describe` names | `TestTheMUFAndFoF2RampsPassCheckOnBothGrounds`, `TestEveryPresetHasLegendWordsAndAUnit`, specimens |
| L-2.2 | A host's own type gets a fill across its classes from the `low`, `middle` and `high` tokens, with library defaults, as FR-15 promised; a host may set the three colours | `TestAHostTypeIsFilledFromLowMiddleHigh`, `TestTheHostRampDefaultsWhenUnset`, `TestAHostRampIsHeldToCheckRamp` |
| L-2.3 | A host's ramp that fails `CheckRamp` is reported through `Warnings()`, never silently drawn | `TestABadHostRampIsWarned` |
| L-2.4 | A grid that would draw nothing (any depth × ground × lines on or off) is refused at hand-in or warned; never a silent blank | `TestNoGridDrawsNothingSilently` (M2's matrix) |
| L-2.5 | The legend of a host type carries colours and its unit | `TestAHostTypesLegendHasColoursAndUnit` |

## L-3 — Units and labels (watchpost D-34; D-3)

| # | Requirement | Instrument |
|---|---|---|
| L-3.1 | Contour labels carry the type's unit | `TestContourLabelsCarryTheirUnit` |
| L-3.2 | Labels are placed where a line crosses north-south as well as east-west, for every field; **behaviour change**, named in the changelog | `TestAFieldBandedEitherWayIsLabelled` (replacing contourlabel_test.go's "never labelled at all" case), specimens redrawn |
| L-3.3 | `Answer.ValueUnit` is filled for every field from its unit | `TestEveryFieldAnswerCarriesItsUnit` (M3) |

## L-4 — Day and night (watchpost D-35; D-4)

| # | Requirement | Instrument |
|---|---|---|
| L-4.1 | A `terminator` line token and a `night` area-tint token, with defaults on both grounds and in sixteen colours, passing the palette checks | `TestTheTerminatorAndNightTokensHaveDefaults`, `TestNightTintKeepsTheFieldReadable` |
| L-4.2 | Helpers that return the terminator line and the night side for a time, correct against a published solar ephemeris to within a stated tolerance | `TestTheTerminatorMatchesTheEphemeris`, `TestNightSideIsTheComplement` |
| L-4.3 | Only alert roles fill an area today (`render/basemap.go:206-208`); the night tint fills without severity words, digits or a legend entry | `TestTheNightTintCarriesNoAlertWords` |

## L-5 — Station points by value (watchpost D-35; D-4) — **OUT of v0.3.0 (watchpost D-50, this log's D-10)**; kept for the follow-up

| # | Requirement | Instrument |
|---|---|---|
| L-5.1 | A role for a value of a type, on L-2's ramp or preset | `TestRoleForFollowsTheRamp` |
| L-5.2 | The value is always in the point's label (OW-17: never colour alone) | `TestAValuedPointCarriesItsValue` |

## L-6 — `SafeRamps` for every scale (watchpost D-36; D-5)

| # | Requirement | Instrument |
|---|---|---|
| L-6.1 | Every scale token (by `ScaleClass`) honours `SafeRamps` and is reported by `Warnings()`; **behaviour change**, named in the changelog | `TestSafeRampsCoversEveryScaleToken`, `TestWarningsReportEveryScaleToken` |

## L-7 — Owed by v0.2.0, carried into v0.3.0 (watchpost D-65; D-11)

| # | Requirement | Instrument |
|---|---|---|
| L-7.1 | OW-24: a shade key on `Class`, so a reader at `NoColour` or 16 colours can map ░▒▓ to a range (MUF and foF2 legends included) | `TestEveryClassHasAShadeKey` |
| L-7.2 | OW-25: the default depth from `COLORTERM` and `TERM` when the host hints none and `NO_COLOR` is unset | `TestTheDefaultDepthFollowsTheTerminal` |
| L-7.3 | OW-17: a library mark (glyph or suffix) for quake age, fire strength and buoy against tide, so those roles do not rest on colour alone | `TestNoRoleRestsOnColourAlone` |
| L-7.4 | OW-19: `CheckedDialerFor`, a checked dialer that exempts the proxy a host's proxy function names | `TestTheCheckedDialerExemptsOnlyTheNamedProxy` |

OW-14, OW-15, OW-16, OW-18, OW-20 and OW-26 are re-targeted to v0.4.0 (watchpost D-65).

## Non-functional requirements

| # | Requirement | Instrument |
|---|---|---|
| NFR-1 | M5: a 2° whole-globe field with labels and terminator renders within a PLAN-set time and memory | a benchmark, recorded |
| NFR-2 | M4: every row above has a test that would fail without it (this release's rows, D-7) | a check over this table |
| NFR-3 | Gates green: `scripts/gate` full lane, P10 clean, mutants anchored | the gate |

## Risk register

| # | Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|---|
| RL-1 | L-2.2's interpolated colours need inks a cell cannot hold today (C-2) | medium | medium | reserved ramp tokens or literal inks, chosen in PLAN from specimens |
| RL-2 | L-3.2 moves every host's labels at once | high | low | changelog; watchpost redraws its specimens in the same batch |
| RL-3 | L-6.1 changes the colours of shipped layers under `SafeRamps` | medium | low | changelog; watchpost's themes checked |
| RL-4 | No whole-globe test exists (C-5); hidden defects surface late | medium | medium | specimens first |
| RL-5 | The terminator's accuracy is claimed without a reference | low | medium | L-4.2 tested against a published ephemeris |

## Metrics

M1 to M6 as adopted (D-7) in `project-brief.md`. Targets are set in PLAN.
