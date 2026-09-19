# Level 2 — The memory budget map

Up: [architecture](architecture.md) · Carries: FR-9, FR-11, FR-37, NFR-3, NFR-4, D-29, D-36, D-48, D-73, D-75, D-84, D-85, D-90 · Risk RS-7 (**High** until the library's own benchmark)

## The target, as ruled

**8 MB** added to the host (D-29). Against a pinned typical-day fixture: **live heap ≤ 4 MB, peak ≤ 8 MB** (D-48). The first host's whole margin is 10.7 MB. *The ≈ figures in the diagram are a reviewer's arithmetic. **PLAN has since measured** a throwaway build of these structures against the pinned fixture — [memory-measurement.md](memory-measurement.md): for **one view**, 1.0 to 1.5 MB live; for **three maps over the fixture region with the ruled caps, about 3.0 MB live by the measured parts**, with the peak at the 8 MB line by arithmetic, not under it. Three maps on three different dense views run to about 4.9 MB — over the line, reported, not prevented (D-90). The ruled lines change only by ruling.*

## Where the bytes live

```mermaid
flowchart TB
    subgraph HOSTMEM["The host's memory — not counted against the library"]
      HG["Borrowed geometry (FR-11)<br/>synthetic upper bound: 58 zones · 812,058 vertices · 12.4 MiB<br/>measured: ten Gulf-coast counties · 56,827 vertices · 0.9 MB<br/>the library reads it, never copies it"]
      HP["The PNG the host fetched — may be released once the image is prepared"]
    end

    subgraph LIVE["Library, live after a collection — the 4 MB line"]
      direction TB
      FIX["<b>Fixed buffers</b> — measured 0.44 MB<br/>cell grid × 2 frames · dot mask · output lines (149×38)"]
      TC["<b>Tile cache</b> — default cap 0.5 MB, shared · one view draws 0.33 to 0.80 MB<br/>what a live view draws is never evicted; spares only in the room left; over the cap is reported (D-90) · with data dropped while decoding and coordinates as 16-bit integers (D-75)"]
      SC["<b>Shape cache</b> — byte-capped · measured 0.01 MB for 56,827 borrowed vertices<br/>simplified forms and bounding boxes · same rule; a form larger than the whole cap is drawn from the host's memory (D-90)"]
      IC["<b>Image cache</b> — byte-capped · measured 0.25 MB for a 600×400 image<br/>one byte per pixel (D-36) · a loop's frames must fit inside this cap (FR-37)"]
      GC2["<b>Grids, contours, ramps, legend</b> ≈ 0.05 MB (an estimate; NOT measured)"]
      ST["<b>Styles, tokens, pools</b> ≈ 0.6 MB (an estimate; NOT measured)"]
      CAPS["Default caps (D-85, lean first): tiles 0.5 MB and shapes 0.25 MB, shared by every map;<br/>images 0.25 MB a map · all host-settable · the lines cover THREE maps sharing caches<br/>Rule: default caps + fixed buffers ≤ 3 MB, leaving 1 MB for everything else (NFR-3)"]
    end

    subgraph TRANS["Transient — what pushes the peak toward 8 MB"]
      direction TB
      DEC["One tile in decode: its bytes + what is kept<br/>measured tiles run 0.15 to 1.1 MB each (the four Midwest tiles TOGETHER are 1.2 MB) → multiplied by the pump's width, which is the host's (D-84);<br/>the peak line is measured two wide"]
      PNGD["One image in decode: up to 4 MiB for the largest allowed PNG (FR-9) — by arithmetic"]
      GARB["Garbage between collections: at the default collector setting,<br/>heap objects run to about twice live — by arithmetic, not measured"]
    end

    subgraph NOTHEAP["Not heap at all"]
      EMB["Embedded tiles: at most 2.5 MB of program data —<br/>invisible to heap metrics, present in the host's resident memory"]
    end

    LIVE --> SUM["One view, MEASURED: 1.0 to 1.5 MB live<br/>Three maps over the fixture region, ruled caps, by the measured parts: about 3.0 MB live"]
    TRANS --> PEAK["Peak ≈ live + one decode + garbage"]
    SUM --> T1["<b>Tension (NFR-3):</b> at live = 4 MB the 8 MB peak is reached by garbage alone.<br/>Live must sit near 3 MB — and for three maps it does, with no room to spare. RS-7 stays High."]
    PEAK --> T1
```

## How it is measured (NFR-3, NFR-4)

| What | How |
|---|---|
| Live | The change in the runtime's live-heap metric after a forced collection, **after a scripted tour that fills every cache to its cap** — not after a first view |
| Peak | The maximum change in the runtime's heap-objects metric, sampled every 10 ms in a standalone harness, default collector setting |
| Instances | **Three sharing caches — two at 149×38, one at 69×12 — over the fixture region: this is what the lines cover and what gates (D-85, D-90)**; one map; and three maps on three different views, recorded and not gated |
| Unchanged frame | Zero allocations |
| Changed frame — the steady state, since a blinking marker changes most frames | Provisional, unverified: a marker-phase change ≤ 16 KB and ≤ 64 allocations; a one-cell pan no more than the first host's own window cost; identical at 100 and 1,000 features |
| One hour | Live heap grows ≤ 1 KB a minute. That the library starts no goroutine is proved statically, not counted (D-73) |
| Worst case, separately | The 58-zone alert at the fixture view and at zoom 12: accepted, drawn without copying, library-owned bytes within the shape cap, the draw-from-borrowed path inside its stated bound on work (constants, section 3) |
| Extreme input | Not held to 8 MB: bounded by the stated formula — per tile in decode, body cap + decompressed cap + retained cap, times the number of `Work` calls the host runs at once (D-84) |
| The host's resident-memory protocol | Confirmatory only; its run-to-run spread (8.5 MB) exceeds what is being measured |

## What can change this diagram

| If this changes… | …this part moves |
|---|---|
| The library's own benchmark in BUILD against the pinned fixture | Every figure; possibly the ruled lines, by ruling only |
| The pinned fixture changes (pinned 2026-09-19, see the measurement) | The tile and shape figures |
| Image loops are built (FR-37) | The image cache's cap and what fits in it |
