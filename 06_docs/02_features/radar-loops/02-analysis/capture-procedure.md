---
title: "The triggered MRMS capture (OW-12, L6.8)"
status: "READY — to be followed on the first severe day before SHIP"
---

# The triggered MRMS capture

MRMS publishes no colour table and keeps about two hours of history, so the colours it
draws for heavy rain can only be seen on a live severe day (RK-11). The palette the library
carries was observed on a quiet afternoon, up to about 48.5 dBZ (wave 2, M-A); above that,
`ProviderMRMS` is marked `Unverified` and colours off the table are valued along the
legend's heavy-end gradient (L6.5). This procedure is how a capture replaces that estimate
with observation. It is written to be followed cold.

## When

Start as soon as any of these is issued for any part of the United States, and while the
storms are live:

- a **Moderate** or **High** risk in the Storm Prediction Center's convective outlook;
- a **tornado watch**.

MRMS keeps about two hours: begin within that window of the storms' peak, or the heaviest
frames are gone.

## What to capture

From the MRMS WMS layer `conus_bref_qcd`, the layer wave 2 measured:

1. **The legend**, from `GetLegendGraphic` for the layer, as a PNG. Compare it with
   `programs/inputs/mrms/legend.png` byte for byte; if it differs, stop and record that the
   legend changed - every value in the table is read from it.
2. **Every time in the layer's time list** (about sixty, two hours), at the national extent
   and 596×304, as PNG with transparency. These are what the palette is read from.
3. **The newest time at three scales** - national 596×304, regional 596×358 and state 298×152,
   the region and state chosen over the heaviest storms.
4. **The time list itself**, one time a line, and the UTC time the capture began.

Save them beside wave 2's inputs, under
`06_docs/02_features/radar-loops/02-analysis/programs/inputs/mrms-severe-<YYYY-MM-DD>/`, in the
same layout: `legend.png`, `nat/<n>.png`, `national.png`, `region.png`, `state.png`,
`times.txt`, `frames-time.txt`.

## What to do with it

1. Widen `mrmsObserved` in `tables_test.go` to read the new directory as well as wave 2's, and
   run the test. It fails, naming the new colours, if the capture saw any.
2. Run it again with `TUIMAPS_WRITE_TABLES=1`. That regenerates `internal/overlay/provider_mrms.go`
   from the evidence; the committed table and the test are then one thing again.
3. Record in `wave2-measurements.md`, or a new measurement beside it, the new range: the
   heaviest value observed, how many colours were new, and how many pixels of the capture the
   old table would have sent to the fallback (the fallback-share test, L6.9, run on the new
   frames).
4. **The HUM LEAD rules** whether the capture covers the heavy end well enough to clear
   `Unverified` from MRMS's table. Only that ruling flips the flag, and the release note's line
   about the unverified range goes with it.

## If no severe day comes before SHIP

MRMS ships with `Unverified` set; the library warns whenever a colour takes the fallback
(L6.6); the release notes name the unseen range, above about 48.5 dBZ (L-2.3, L6.7). The
release checklist carries the row, and refuses the final tag without the note while the flag
is set (L10.3).
