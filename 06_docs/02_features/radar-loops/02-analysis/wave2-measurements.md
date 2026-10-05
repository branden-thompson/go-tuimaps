---
title: "v0.2.0 Radar loops — DISCOVER wave 2 measurements"
date: 2026-09-23
phase: DISCOVER (RCC)
sev: SEV-0
status: "COMPLETE — the two measurements wave 1 left for the machine (synthesis point 12): the MRMS table and loop memory"
---

# Wave 2 — the two measurements that needed the machine

Both were run on 2026-09-23 against live data, through the library's public calls, by small programs
now filed with their inputs and raw output under `programs/` (D-32). Nothing here is read from source
alone.

## M-A — MRMS colours against the server's legend (L-2; ruled D-19)

**Inputs.** The legend from `GetLegendGraphic` for `conus_bref_qcd`: 500×30 pixels, a continuous
gradient in rows 0–9 (−20 to 70 dBZ, 0.18 dB a column), labels below. The newest frame (16:28:13Z)
at three scales, national 596×304, regional 596×358 and state 298×152, plus 15 national frames spread
across the server's whole time list (60 times, about two hours).

| | National | Regional | State |
|---|---|---|---|
| Rain pixels | 9,028 | 25,310 | 8,056 |
| Colours in the frame | 94 | 101 | 90 |
| Exactly a legend colour | 59.7 % | 59.0 % | 61.4 % |
| Within 10 (Lab) of one | 40.3 % | 40.8 % | 38.5 % |
| Further than 10 | 3 px | 43 px | 6 px |
| Near a class edge (another class within 10) | 984 px | 2,647 px | 944 px |
| **Library, 256-entry legend table, `Exact: true` — unmatched** | **7,243** | **20,397** | **6,447** |
| Library, same table, default matching — unmatched | 3 | 43 | 6 |

**Findings.**

1. **MRMS draws from a small fixed palette.** Each frame uses 88–101 colours; across 18 frames the
   total reached **111 by the twelfth frame read and did not grow after**. The frames were read in
file-name order (1, 10, 11 … 15, 2 …), not time order, so "the twelfth" is not a point in time. That fits the library's
   256-entry table limit. No partly transparent pixels were seen.
2. **The legend is a 1,232-colour gradient, and the library's table holds 256 entries, so a table
   sampled from the legend misses most frame colours.** About 60 % of rain pixels are exactly *some*
   legend colour, but with exact matching against the 256-entry sample, **80.0–80.6 %** were unmatched
   and would draw as nothing. *(Corrected 2026-09-23 after red team round 1: first written as "71–80 %",
   which nothing in the table supports, and as "no table sampled from it can match exactly", which
   the 60 % contradicts — the misses come from the 256-entry sample, not from the legend.)*
3. **The heavy end is the weak point.** The oranges and reds (from about 40 dBZ up) are 3 to 30 from
   any legend colour, and nearest-legend-colour puts several unlike colours at about 47–48 dBZ. The
   rain rates the legend gives the most dangerous colours are the least trustworthy.
4. **The palette seen is incomplete.** The day's heaviest rain was about 48 dBZ. Colours for heavier
   storms have not been observed.

**Limits (D-32).**
- One afternoon, one weather regime (2026-09-23, rain up to about 48 dBZ). Nothing here speaks for a
  severe-weather day.
- "18 frames" is 16 distinct times: the newest time appears three times, once per scale. The palette
  stopped growing across those 16 times, not across 18 independent frames.
- One server, one layer (`conus_bref_qcd`), EPSG:4326 at these sizes. Another projection or size may
  resample differently and add colours.
- Lab distances use the library's sRGB-to-Lab conversion under normal vision only; the "near a class
  edge" count is a bound on ambiguity, not a count of misclassified pixels.
- Programs, inputs and raw output: `programs/`.

## M-B — what twelve radar frames cost today

**Inputs.** Twelve real IEM `n0q` frames, the current one and `-m05m`…`-m55m`, over the continental
United States at 298×152 (5.6–5.9 KB each as PNG) and 596×304 (20.7–21.3 KB each). The library has
no loop yet, so each frame was handed in as its own image overlay on a 149×38 map, then settled and
rendered once, which prepares every frame. Heap measured after two collections.

| | Heap a map | Above an empty map |
|---|---|---|
| An empty map (baseline, three maps averaged) | 0.72 MB | — |
| 12 frames at 298×152 | 1.31 MB | **0.59 MB** |
| 12 frames at 596×304 | 2.91 MB | **2.19 MB** |

A first render with all twelve composited took about 27 ms a map at either size.

**Findings.**

1. **The measurement roughly agrees with wave 1's arithmetic** (one byte a pixel), within the
   limits below.
2. **Against NFR-3 (4 MB live):** one loop at 298×152 uses about a sixth of it, and one at 596×304
   more than half. Two sources at 596×304 on one map (5.1 MB) do not fit. Three maps, each with one
   loop at 596×304, use 8.7 MB.
3. **298×152 is exactly the braille dot grid of a 149×38 map** (2×4 dots a cell), so at the host's
   large size it gives one pixel a dot. The larger frame helps only when the host zooms in without
   fetching again.
4. **A host that names frames by index trips the near-duplicate-id warning** ("frame-1" beside
   "frame-10" and "frame-11"). It stops mattering once a loop is one overlay, but it is a trap for
   any host that builds loops out of separate overlays today.


**Limits (D-32).**
- One run a size, on one 18-core Apple M-series Mac. No spread, no repeat.
- "MB" is Go heap bytes (`HeapAlloc` after two collections) in MiB, not the process's resident size.
- **Units differ between the waves.** Wave 1's arithmetic (0.54 and 2.17 MB) is decimal megabytes;
  these measurements are MiB, in which the same pixels are 0.52 and 2.07. With the host's PNG bytes
  added (0.07 and 0.24 MiB), the expectation is 0.58 and 2.31 MiB against 0.59 and 2.19 measured:
  it closes at 298×152 and **not at 596×304**, where the heap is 0.12 MiB below it. Heap accounting
  at this grain is approximate; the per-frame pixel cost is the reliable part.
- The 27 ms first render is on this machine only, and it composites all twelve frames at once, which
  a loop never does. It is not a per-frame playback cost; that is unmeasured (R1-15).
- Twelve frames were twelve separate overlays, because no loop exists yet. A real loop's cost is
  PLAN's to measure against M4.
- Programs, inputs and raw output: `programs/`.