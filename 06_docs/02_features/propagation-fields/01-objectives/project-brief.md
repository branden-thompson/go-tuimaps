---
title: "go-tuiMaps v0.3.0 — Propagation fields — PROJECT BRIEF"
date: 2026-10-06
phase: DISCOVER
report_template: project-brief v1.1.0
level: LEVEL-1
sev: SEV-0
authority: HUM LEAD
directives: FULL GIT; FULL DOCS; FULL REPORTS; FULL DIAGRAMS; FULL RCC; FULL PLAN; FULL TDD; FULL INST
branch: feature/propagation-fields
paired_release: "watchpost 0.19.0 — Propagation overlays (branden-thompson/watchpost#25), with go-ionomaps; this release's candidates are pinned in watchpost's BUILD and it ships first (watchpost D-3)"
status: "APPROVED by the HUM LEAD 2026-10-06 (D-8). Problem statement LOCKED (D-6). Metrics ADOPTED (D-7, M4 narrowed to this release's rows). Host requirements ruled (watchpost D-32 to D-36; this log's D-1 to D-5)."
---

# Library Release | `go-tuiMaps v0.3.0 — Propagation fields`

**LEVEL-1; SEV-0; FULL GIT; FULL DOCS; FULL REPORTS; FULL DIAGRAMS; FULL RCC; FULL PLAN; FULL TDD; FULL INST**

## Summary & Intent

watchpost 0.19.0 draws two global fields of a new kind: the maximum usable frequency, MUF(3000), and the F2
layer's critical frequency, foF2, both in MHz, over the whole Earth, with a day/night terminator (watchpost
D-21, D-24). go-tuiMaps v0.2.0 can hold a whole-globe grid, but it cannot draw one a reader can read.

**What v0.3.0 delivers**, as ruled by its host:
- a field that continues over the sea, chosen per overlay (L-1, HR-1);
- colour and a legend for a field of the host's own kind: named `muf` and `fof2` presets, and the generic host ramp the library promised in FR-15 (L-2, HR-2);
- contour labels and words that carry their unit, with labels where lines run east-west (L-3, HR-3);
- a terminator and a night tint (L-4, HR-4);
- station points coloured by value (L-5, HR-5);
- `SafeRamps` honoured for every scale (L-6).

**Why now.** The host's DISCOVER found every gap at `file:line`
(`watchpost/06_docs/02_features/propagation-overlays/02-analysis/wave1-findings.md`). Two of them are
promises this library wrote down and never built: FR-12/D-87's "a host can flip either, per overlay", and
FR-15's host-type ramp.

**What happens if it is not built.** watchpost's Propagation mode would show land only, in no colour or
nothing at all, with unlabelled numbers and no sense of where the sun is.

## Problem Statement — LOCKED (D-6)

> **"A developer embedding go-tuiMaps cannot draw a global field of their own kind so that it can be
> read: the field stops at every coast, has no colours or legend of its own and can draw nothing at
> all, its labels carry no unit and are missing where its lines run east to west, and nothing shows
> where day ends, so the person reading the map sees a broken or empty picture, or numbers without
> meaning."**

| Criterion | |
|---|---|
| Bad outcome | a broken or empty picture, or numbers without meaning |
| Affected humans | the embedding developer, and the person reading the map (as v0.2.0's) |
| Tech agnostic | names the library, as v0.2.0's statement did; no mechanism named |
| Non-prescriptive | presets, a generic ramp or other inks could each answer the colour clause |
| Verifiable | each clause is a test: over water, a drawn fill, units, labels on both crossings, a terminator |

## Requirements — from the host

Watchpost's HR-1 to HR-5 (`watchpost/.../propagation-overlays/01-objectives/project-brief.md`, ruled at
watchpost D-32 to D-36) map to L-1 to L-5. L-6 is the defect wave 1 found. `requirements.md` is written
from these in DISCOVER and wins on any conflict once approved.

### L-1 — A field over the sea, per overlay (HR-1; watchpost D-32)
- **L-1.1** A host chooses, per field overlay, whether it continues over water. The same choice is offered per image for masking by water (FR-12, D-87).
- **L-1.2** `TestHostCanFlipEither`, named by v0.1.0's plan (task 09.7) and never written, is written and holds both choices.

### L-2 — Colour for a host's own field (HR-2; watchpost D-33)
- **L-2.1** Named `muf` and `fof2` presets in MHz, hand-tuned per depth on both grounds, with their own legend words, in the pattern of UV, AQI and QPF.
- **L-2.2** The generic host ramp FR-15 promised: a host's own type gets a fill across its classes from `low`, `middle` and `high`, with library defaults, held to `CheckRamp`.
- **L-2.3** A host type that would draw nothing (`internal/render/field.go:146-163`) is refused or warned, never silent.

### L-3 — Units and labels (HR-3; watchpost D-34)
- **L-3.1** Contour labels carry the type's unit.
- **L-3.2** Labels are placed on north-south crossings as well as east-west, for every field. **A behaviour change:** existing fields' labels move.
- **L-3.3** `Answer.ValueUnit` is filled for every field from its unit, not only temperature (`internal/describe/answer.go:138-141`).

### L-4 — Day and night (HR-4; watchpost D-35)
- **L-4.1** `terminator` (a line) and `night` (an area tint) tokens, with defaults on both grounds and in sixteen colours.
- **L-4.2** Helpers that return the terminator and the night side for a time.

### L-5 — Station points by value (HR-5; watchpost D-35) — OUT of v0.3.0 (watchpost D-50)
- **L-5.1** A generic role for a value of a type, on L-2's ramp, with the value always in the point's label (OW-17: never colour alone).

### L-6 — `SafeRamps` for every scale (watchpost D-36)
- **L-6.1** "Is a ramp token" is decided by `ScaleClass`, not by position (`internal/colour/palette.go:72`, `sixteen.go:35`). A test covers every scale token. **A behaviour change:** wind, wave, UV, AQI and QPF now honour `SafeRamps`, and `Warnings()` reports them.

## Technical Constraints

Measured at `origin/main` `d1c4d5e` by the host's wave 1, and spot-checked.

- **C-1 — One place builds the render input.** `Map.Render` (`map.go:431`) through `paint` and `draw`. Neither sets `FieldsOverWater` or `ImagesMaskedByWater` (`internal/render/frame.go:92`, `:96`).
- **C-2 — A cell's background holds a token ink** (`frame.go:159-160`). L-2.2's interpolated colours need reserved ramp tokens or literal inks; the Painter already has literal inks (`basemap.go:471-500`).
- **C-3 — Only alert roles fill an area** (`render/basemap.go:206-208`). L-4.1's night tint needs a second fill rule.
- **C-4 — `FitWorld` frames 84°N to 56°S** (`internal/project/fit.go:158-162`). Whether a propagation view needs the far south is the host's question; the library states the frame.
- **C-5 — No test renders a whole-globe grid or a host-type grid.** "Works today" rests on reading the code. Both need specimens before any change.
- **C-6 — Two changes are not additive** (L-3.2, L-6.1). Each is ruled (watchpost D-34, D-36) and gets a contract row and a changelog entry naming it.
- **C-7 — Owed work touching the same code.** OW-24 (a shade key on `Class`) shares L-2's legend `Class`; OW-17 is L-5's constraint.

## Metrics of Success — ADOPTED (D-7)

| # | Metric | Type | Definition | The anti-solution it closes |
|---|---|---|---|---|
| M1 | Field read | Primary | From the map alone, a reader states the MUF band at N places and where it rises, by scripted UAT over recorded fields, **plus a non-visual arm from `Report`'s words alone** | a pretty map with wrong or missing classes; words without units |
| M2 | No silent blank | Primary | Zero host-type or preset grids that draw nothing without a refusal or warning, across every depth × ground × lines on/off, by a test matrix | refusing everything: a grid that can draw must draw |
| M3 | Units everywhere | Primary | Every field label and every field `Answer` carries its unit, for every preset and a host type, by a test | a unit on MUF only |
| M4 | Promise truth | Primary | Every v0.3.0 requirement the release's documents say is delivered has a test that would fail without it, each anchored to a mutant the gate kills (D-7: this release's rows only; FR-12 and FR-15 are covered through L-1 and L-2) | the v0.1.0 shape: a promise and a build-log claim with no test |
| M5 | Globe cost | Secondary | Render time and memory for a 2° whole-globe field with labels and terminator, at a PLAN-set target, measured | a coarse grid: measured at 2° |
| M6 | Gate trust | Secondary | Zero unattributable gate failures across a stated number of consecutive full runs (carried from v0.2.0) | dropping legs |

## Other Considerations

- **Ships first.** watchpost pins this release's candidates in its BUILD and ships on its final tag (watchpost D-3), as with v0.2.0.
- **Standing principles** (watchpost 0.18.0 D-23): P-1, P-2 (the host chooses), P-3. L-2.2 is P-3's reason for building the generic ramp beside the presets.
- **No issue of record.** D-8 approved the brief without one; none is opened.
- **Minor rulings** follow watchpost D-13 (A-n rows, batched for veto).

## Completeness Check

```
PROJECT BRIEF — COMPLETENESS CHECK
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  [✓] Header              — library release; paired with watchpost 0.19.0 and go-ionomaps
  [✓] Directives          — inherited (watchpost D-1, D-3)
  [✓] Summary / Intent    — what, why now, cost of not building
  [✓] Problem statement   — LOCKED (D-6)
  [✓] Requirements        — L-1..L-6 from HR-1..HR-5 and D-36, ruled
  [✓] Metrics of Success  — ADOPTED (D-7)
  [✓] Tech Constraints    — C-1..C-7, measured at d1c4d5e
  [✓] Considerations      — ships first, principles, no issue of record (D-8), minor rulings
  [✓] Rulings             — D-0..D-8 (D-0..D-5 restate watchpost D-3, D-32..D-36)
  [✓] Approval            — APPROVED as presented (D-8).  INTAKE CLOSED.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```
