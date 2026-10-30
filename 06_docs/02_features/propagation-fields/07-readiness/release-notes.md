# go-tuiMaps v0.3.0 — release notes

Up: [requirements](../01-objectives/requirements.md) · [contract, section 12](../../go-tuimaps/03-architecture-design/contract.md)

Written as v0.3.0 is built, and completed at its release step.

## What changes in behaviour

Ruled before each landed (the brief's C-6), and listed with what a host does in the contract's section 12.

- **Contour labels carry their unit, and every field is labelled** (L-3.1, L-3.2): `10°C`, `14MHz`; a field whose lines run east-west is labelled as one whose lines run north-south always was. Every host's labels move at once (RL-2); the specimens and the reference frames are redrawn (D-30).
- **A field's answer carries its unit** (L-3.3): `Answer.ValueUnit` is the unit the host declared, for every field.
- **A host's own field type is filled** (L-2.2, L-2.4): at truecolor and 256 colours a host type's bands are drawn, where before only its contours were, and only when asked; its legend keys each class in its colour (L-2.5). The host-type specimens are redrawn. A host type drawn as its edge - a reach area, asked for with `Type.Edge` (below) - has a broken edge and, over a field filled before it, no fill of its own (L-2.6).
- **Safe ramps keep every preset's classes** (L-6.1, D-29): a host's colour for a wind, wave, UV, AQI, QPF, MUF or foF2 class is passed over with safe ramps on, as radar's, temperature's and the alerts' always were.
- **`Warnings` reports every preset's classes, the host ramp's anchors with them** (L-6.1): a host's colour for `low`, `middle` or `high` that breaks the checker, spread as a host type would be at the most classes the defaults pass, is reported up front as `RampRuleBroken` ("host ramp"), not only once a host type is drawn.
- **The default depth follows the terminal** (L-7.2): with no hint and `NO_COLOR` unset, `COLORTERM` and `TERM` choose truecolor, 256, sixteen or no colour, where it was always truecolor; a host that wants one depth hints it.
- **Roles no longer rest on colour alone** (L-7.3): a weaker hotspot, a tide station and a quake's age each draw a dot shape of their own.
- **The edge mark is asked for** (watchpost D-216, C-18): a host type of two classes is drawn as its edge alone only with `Type.Edge`; before, every two-class host type was.
- **A wrong colour count is a malformed ramp** (watchpost D-216, C-17): refused as `MalformedRamp`, where it was `UnsortedBreaks`.
- **One ramp warning an overlay** (C-26): a host type set again with the same broken colours is one warning, replaced on each `Set` and gone once a `Set` passes.
- **A tile reply's headers are capped at 64 KiB** (watchpost R2I-3): a reply whose headers run past it is refused, where Go's own bound was about ten megabytes.
- **A cell on the antimeridian is drawn with the antimeridian's value** (watchpost D-343): a field's cell whose dots lie either side of 180° was sampled at the mean of their longitudes, about 0°, so a world field drew one column with Greenwich's value, and a region's field one column of nothing; an image's dot beside it covered half the world. Both are now taken the short way round. The globe and host-type specimens are redrawn: in each, the two cells on 180° alone change, from Greenwich's class to their row's.
- **Every look setter checks a host ramp again** (watchpost RC-15, R2C-6): `ColourDepth`, `SetPalette`, `SafeRamps`, `Ground` and `PaintGround` each warn of, or clear, a host type's broken ramp at the look they set.

## What it adds

- **Fields over water, per overlay** (L-1): `Grid.OverWater`, `Image.MaskedByWater`.
- **The `muf` and `fof2` presets** (L-2.1), in MHz, broken at the amateur band edges.
- **A host type's colours** (L-2.2, L-2.3): `Type.Colours`, one a class; without them, the `low`, `middle` and `high` tokens spread across the classes, with defaults that pass the checker to 9 classes at truecolor and 7 at 256 colours. Colours that break the checker are drawn and reported as `RampRuleBroken`.
- **Day and night** (L-4): `Terminator(at)` and `NightSide(at)`, the terminator within 0.5° of NOAA's, held by its test over 78 moments of 2026, the equinoxes and solstices among them (a one-off check measured the sun within 0.011° of NOAA's solar position over 2020 to 2035, 126 moments, A-7; no test holds that figure); a `NightTint` polygon shades the ground and sea beneath it, never a field's colours, with no words or legend entry; a `TerminatorLine` is dotted at every depth; `PlaceReport.Daylight` says day or night and when it next turns.
- **A shade key** (L-7.1, D-35): `Class.Shade`, the block shade an image's or rain's class is drawn in at sixteen colours and none.
- **`CheckedDialerFor`** (L-7.4): the checked dialer for a host behind a proxy at a private address, with the proxy function to set beside it.
- **The error kind `BadEdge`** (watchpost D-229): an edge mark asked for on a preset, or on a host type of other than two classes, is refused as it, its words saying how to ask for one.
- **`NightOverlay(id, at)`** (watchpost D-216): the night side and its terminator, each in its role, as one overlay in one call.
- **`DaylightAt(place, moment)`** (watchpost D-263): a place's day or night and its next turn at any moment, the night `NightOverlay` draws for it, so a host showing an hour ahead says what its picture shows; `PlaceReport.Daylight` is the same at the map's clock. A place off the globe, or no number, is not known (watchpost R2I-5).
- **`Type.Edge`** (watchpost D-216, L-2.6): a host type of two classes drawn as the edge between them alone, a broken line at every depth.
- **`socks5h` proxies** take port 1080 when none is given, as `socks5` does (I-7).
- **`TerminalDepth`** (watchpost D-207, D-216): `ColourDepth(TerminalDepth)` takes a depth hint back, so a host offering "follow the terminal" beside the depths can return to it; the terminal's settings choose again (L-7.2).
- **A host's zone** (L-8, watchpost D-195, D-196): a `Zone` Polygon or Circle is outlined in red and its inside tinted, beneath any alert's tint, with no severity or legend entry; over a field its tint cannot blend with (the MUF and foF2 among them) the field is drawn exact and the zone stippled; at sixteen colours and none, the outline alone. `Report` now answers a circle given in kilometres.

## What was measured

| Metric | Target | Measured |
|---|---|---|
| M1, field read | 10 places on a recorded MUF field, 8 of 10 from the picture and 8 of 10 from `Report`'s words, graded by the HUM LEAD | **Not measured by a sitting** (watchpost D-348). Its ground is held by tests: every field answer carries its unit, every contour its label and unit |
| M2, no silent blank | no grid that draws nothing without a refusal or a warning, at any depth × ground × lines | `TestNoGridDrawsNothingSilently`, the whole matrix |
| M3, units everywhere | every field label and answer carries its unit, every preset and a host type | `TestEveryFieldAnswerCarriesItsUnit`, `TestContourLabelsCarryTheirUnit` |
| M4, promise truth | each of this release's requirements held by a test a mutant proves can fail | every mutant in `06_docs/mutants/mutants.json` killed by its test in every full gate |
| M5, globe cost | a whole-globe 2° field after a pan, filled, with labelled contours, the night and its terminator and a reach's edge: at most 16 ms at 400 × 110 and 8 ms at 200 × 56 (Apple M5 Pro) | 7.4 to 8.0 ms and 2.0 to 2.2 ms, five runs each, on an Apple M5 Pro, go1.27.3 (2026-10-30) |
| M6, gate trust | five consecutive clean full runs of the gate | counted from `06_docs/gate-runs.md` by `m6_test.go` before the tag |

The contract breaks nothing of v0.2.2's public surface that section 12 does not list
(`TestNothingOfTheLastReleaseBreaksUnlisted`, held to the last release by `TestTheContractIsHeldToTheLastRelease`).

## Known limits

- **A whole-globe frame at 800 × 220** takes 27 to 29 ms after a pan on the reference machine, past the 16 ms
  M5 holds at 400 × 110 (watchpost D-261, D-40): recorded, not a target.
