# Reference frames — the evidence M1a is judged from

Up: [readiness](../) · Plan tasks 14.3 and 14.16 · Carries: M1 (D-43, D-67), NFR-15

| Field | Value |
|---|---|
| Status | **Candidates. Not yet approved.** A frame becomes a golden when HUM LEAD has approved it (plan task 14.16, RS-15). Until then `TestReferenceFrames` keeps them from changing without anyone noticing, and nothing here is evidence of anything |
| What they are | Every M1 scenario this release draws, at both judged sizes, in colour and with none: `<scenario>-<cols>x<rows>-<colour or plain>.txt` |
| How they are made | `TestReferenceFrames` in the library, from the scenario files in [02-analysis/scenarios](../../02-analysis/scenarios), framed with `FitTo` so that the place and the hazard are in one frame (14.5), settled, and drawn on a fixed clock. Written again by running the library's tests with `TUIMAPS_WRITE_REFERENCE_FRAMES` set |
| Why a fixed clock | A frame that differed between two runs would be no evidence at all (NFR-6). The scenarios' data is valid at the same instant the frame is drawn at, so nothing here is marked stale |
| How to look at them | `cat` one in a terminal whose font has braille (D-57). The `colour` files carry the terminal's own colour sequences; the `plain` files carry none, and are the ones NFR-15's reviewer session uses |
| Scenario 5 | Wind, which arrives with the wind overlay (D-44). It has no frame here and is not counted |
