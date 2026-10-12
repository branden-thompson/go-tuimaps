# go-tuiMaps v0.3.0 — release notes

Up: [requirements](../01-objectives/requirements.md) · [contract, section 12](../../go-tuimaps/03-architecture-design/contract.md)

Written as v0.3.0 is built, and completed at its release step.

## What changes in behaviour

Ruled before each landed (the brief's C-6), and listed with what a host does in the contract's section 12.

- **Contour labels carry their unit, and every field is labelled** (L-3.1, L-3.2): `10°C`, `14MHz`; a field whose lines run east-west is labelled as one whose lines run north-south always was. Every host's labels move at once (RL-2); the specimens and the reference frames are redrawn (D-30).
- **A field's answer carries its unit** (L-3.3): `Answer.ValueUnit` is the unit the host declared, for every field.
- **Safe ramps keep every preset's classes** (L-6.1, D-29): a host's colour for a wind, wave, UV, AQI, QPF, MUF or foF2 class is passed over with safe ramps on, as radar's, temperature's and the alerts' always were.

## What it adds

- **Fields over water, per overlay** (L-1): `Grid.OverWater`, `Image.MaskedByWater`.
- **The `muf` and `fof2` presets** (L-2.1), in MHz, broken at the amateur band edges.
