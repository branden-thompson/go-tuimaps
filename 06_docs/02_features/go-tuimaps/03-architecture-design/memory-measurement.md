# PLAN — The memory measurement (D-29, D-48)

Up: [architecture](architecture.md) · [memory budget map](L2-memory.md)

| Field | Value |
|---|---|
| Phase | PLAN |
| Date | 2026-09-19 |
| Why this exists | HUM LEAD ruled an 8 MB target "confirmed by measurement at PLAN exit" (D-29), against a typical-day fixture with the worst case tested separately (D-48). Risk RS-7. |
| What was measured | A **throwaway program**, kept outside this repository, that builds the structures the design calls for and reads the Go runtime's own metrics. **It is not the library.** Tiles were decoded with a general-purpose decoder and then converted to the compact form of ruling D-75, so the *live* figures describe the design and the *peak* figures are an upper bound. |

## The fixture, now pinned

| Part | Contents | Source |
|---|---|---|
| View | 149×38 cells, zoom 6.4, centred on Tampa Bay (27.6, −82.4) | — |
| Alert overlay | **Ten real county zones of Florida's Gulf coast** — Hernando, Pasco, Pinellas, Hillsborough, Manatee, Sarasota, Charlotte, Lee, Collier, Dixie: **56,827 vertices in 1,107 polygons** (mostly small islands), 0.91 MB as the host holds them | The weather service's public zone shapes, fetched 2026-09-19 |
| Basemap | The four zoom-6 tiles of that view, 540 KB of vector-tile data, from a version-pinned address | The tile service |
| Radar | One image, 600×400, held as one byte a pixel (D-36) | sized for the view |
| Temperature | One grid, 64×48 | — |
| A heavier variant | The same, with the four zoom-5 Midwest tiles named in NFR-5 — 1,245 KB of vector-tile data | — |

## Results

Three runs each; the spread between runs was under 0.02 MB live and 0.2 MB peak.

| | Gulf-coast fixture | Heavier Midwest tiles | Line (D-48) |
|---|---|---|---|
| Fixed buffers: two cell grids, dot mask, output | 0.44 MB | 0.44 MB | |
| Tiles, compact form — layers the map does not draw and every place-name translation dropped, coordinates as 16-bit integers | 0.33 MB (1,079 features) | 0.80 MB (2,371 features) | |
| Shapes simplified for the view: 56,827 → 1,192 vertices; 62 of 1,107 polygons survive, the rest are smaller than a dot | 0.01 MB | 0.01 MB | |
| Radar at one byte a pixel, and the grid | 0.25 MB | 0.25 MB | |
| **Live, added** | **1.03 MB** | **1.50 MB** | **4 MB** |
| **Peak heap objects, above the baseline** — an upper bound | **2.8 to 2.9 MB** | **4.75 MB** | **8 MB** |

**The close-zoom question (FR-11).** The simplified, unclipped form of all ten zones: 9.5 KB at zoom 6.4 · 87 KB at zoom 9 · **359 KB at zoom 12** (44,892 vertices). It never approaches a byte cap measured in megabytes, so the draw-from-borrowed fallback is for inputs far beyond this.

## What this does and does not show

- **It shows** that the design's live memory for a typical day is about a quarter to a third of the 4 MB line, and the peak — even paying for a general decoder's waste — is inside 8 MB. The reviewer's arithmetic (about 3.6 MB) was cautious by a factor of two to three.
- **Not included:** styles and colour tokens, label placement, the description's index, the pending-work queue, and everything else a real library carries — estimated by the reviewer at 0.6 MB, unmeasured. Adding it leaves the Gulf-coast fixture near 1.7 MB and the heavier one near 2.1 MB.
- **Not a filled cache.** NFR-3 measures after a tour that fills every cache to its cap. This is one view. What it tells the cache design: a view's tiles are 0.3 to 0.8 MB, so a tile cache of about 1.5 MB holds two to four views, and default caps plus fixed buffers can sit under 3 MB as NFR-3 now requires.
- **The tension NFR-3 names is eased, not removed.** At 1.0 to 1.5 MB live there is room between live and the 8 MB peak line; at 4 MB live there would not be.
- **The real confirmation is the library's own benchmark in BUILD**, against this same fixture (NFR-3). Until then the 8 MB target stands as ruled, with this as the evidence that it is reachable.

## Re-run after the PLAN red-team, and the ruling that followed

The red-team (PL-PF-1, PL-BZ-5) was right that the first run was not NFR-3's condition: it was one view, with no cache filled and one map. Re-run with a tile cache filled to 1.25 MB (thirteen real tiles) and three maps — two at 149×38 and one at 69×12, each with a radar image: **3.56 MB live**, under the 4 MB line.

**The peak line cannot be shown by this program.** It uses a general decoder, whose waste dominates the peak it sees (11.8 MB in that run, which says nothing about the library). By arithmetic — the heap roughly doubles between collections at the default setting, and under D-84 two concurrent `Work` calls each hold about 1 MB while decoding — peak is about twice live plus 2 MB, so live must sit near 3 MB for the peak to stay under 8.

**Ruled (D-85): lean first.** The lines cover three maps sharing caches. Default caps: tiles 0.5 MB and shapes 0.25 MB, shared; images 0.25 MB a map. By the measured parts above, three maps then hold about 0.94 (fixed buffers) + 0.5 + 0.25 + 0.75 + 0.6 (unmeasured) ≈ **3.0 MB live**, and about 8 MB peak by the same arithmetic — at the line, not under it with room. **Risk RS-7 stays High** until the library's own benchmark (plan task 14.6) measures it.

## A correction to DISCOVER's worst case

The "58-zone, 812,058-vertex alert" in NFR-3 was never observed. It is the largest zone in a three-zone sample (14,001 vertices — Citrus County, Florida, a coast of islands) multiplied by the largest zone count seen in any alert (58). Measured in PLAN:

| Sample | Vertices a zone |
|---|---|
| 62 inland forecast zones, from live multi-zone alerts | 47 to 527; the ten largest of a real 58-zone alert total about 1,600 |
| 12 Gulf-coast county zones | 1,836 to 14,001; twelve total 82,251 |

So the ordinary alert is thousands of vertices, a coastal one tens of thousands, and 812,058 is a **synthetic upper bound** — still a fair stress test, and now labelled as one.
