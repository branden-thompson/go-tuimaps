# go-tuiMaps v0.3.0 — release notes

Up: [requirements](../01-objectives/requirements.md) · [contract, section 12](../../go-tuimaps/03-architecture-design/contract.md)

Written as v0.3.0 is built, and completed at its release step.

## What changes in behaviour

Ruled before each landed (the brief's C-6), and listed with what a host does in the contract's section 12.

- **Contour labels carry their unit, and every field is labelled** (L-3.1, L-3.2): `10°C`, `14MHz`; a field whose lines run east-west is labelled as one whose lines run north-south always was. Every host's labels move at once (RL-2); the specimens and the reference frames are redrawn (D-30).
- **A field's answer carries its unit** (L-3.3): `Answer.ValueUnit` is the unit the host declared, for every field.
- **A host's own field type is filled** (L-2.2, L-2.4): at truecolor and 256 colours a host type's bands are drawn, where before only its contours were, and only when asked; its legend keys each class in its colour (L-2.5). The host-type specimens are redrawn. A two-class host type - a reach area - has a broken edge and, over a field filled before it, no fill of its own (L-2.6).
- **Safe ramps keep every preset's classes** (L-6.1, D-29): a host's colour for a wind, wave, UV, AQI, QPF, MUF or foF2 class is passed over with safe ramps on, as radar's, temperature's and the alerts' always were.
- **`Warnings` reports every preset's classes, the host ramp's anchors with them** (L-6.1): a host's colour for `low`, `middle` or `high` that breaks the checker, spread as a host type would be at the most classes the defaults pass, is reported up front as `RampRuleBroken` ("host ramp"), not only once a host type is drawn.
- **The default depth follows the terminal** (L-7.2): with no hint and `NO_COLOR` unset, `COLORTERM` and `TERM` choose truecolor, 256, sixteen or no colour, where it was always truecolor; a host that wants one depth hints it.
- **Roles no longer rest on colour alone** (L-7.3): a weaker hotspot, a tide station and a quake's age each draw a dot shape of their own.

## What it adds

- **Fields over water, per overlay** (L-1): `Grid.OverWater`, `Image.MaskedByWater`.
- **The `muf` and `fof2` presets** (L-2.1), in MHz, broken at the amateur band edges.
- **A host type's colours** (L-2.2, L-2.3): `Type.Colours`, one a class; without them, the `low`, `middle` and `high` tokens spread across the classes, with defaults that pass the checker to 9 classes at truecolor and 7 at 256 colours. Colours that break the checker are drawn and reported as `RampRuleBroken`.
- **Day and night** (L-4): `Terminator(at)` and `NightSide(at)`, within 0.011° of NOAA's solar position; a `NightTint` polygon shades the ground and sea beneath it, never a field's colours, with no words or legend entry; a `TerminatorLine` is dotted at every depth; `PlaceReport.Daylight` says day or night and when it next turns.
- **A shade key** (L-7.1, D-35): `Class.Shade`, the block shade an image's or rain's class is drawn in at sixteen colours and none.
- **`CheckedDialerFor`** (L-7.4): the checked dialer for a host behind a proxy at a private address, with the proxy function to set beside it.
- **A host's zone** (L-8, watchpost D-195, D-196): a `Zone` Polygon or Circle is outlined in red and its inside tinted, beneath any alert's tint, with no severity or legend entry; over a field its tint cannot blend with (the MUF and foF2 among them) the field is drawn exact and the zone stippled; at sixteen colours and none, the outline alone. `Report` now answers a circle given in kilometres.
