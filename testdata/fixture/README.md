# The pinned fixture

Test data for the memory and timing requirements (NFR-3, NFR-5; rulings D-29, D-48, D-85). Pinned in PLAN on 2026-09-19 and described in `06_docs/02_features/go-tuimaps/03-architecture-design/memory-measurement.md`. `HASHES` lists every file with its SHA-256; a test (plan task 00.9) fails if any file changes. 25 files, 6.6 MB (6,595,652 bytes). The map tiles are unmodified map data and hold street and place names from all over the world; a scan of this repository for words that should not appear must leave the `.pbf` files out.

| Part | What | Source, fetched 2026-09-19 | Terms |
|---|---|---|---|
| `zones/` | Ten county zones of Florida's Gulf coast — 56,827 vertices in 1,107 rings, which make 1,011 polygons | The United States National Weather Service's public zone shapes | A work of the United States government; public domain |
| `tiles-gulf-z6/` | The four zoom-6 vector tiles of the fixture view (Tampa Bay, 149×38 cells, zoom 6.4) | OpenFreeMap, at the version-pinned address `planet/20260913_164504_pt` | © OpenMapTiles · © OpenStreetMap contributors — see below |
| `tiles-midwest-z5/` | The four zoom-5 tiles named in NFR-5 — the heavier variant | The same | The same |
| `radar/gulf-2026-09-19.png` | **The fixture's radar image:** one image for one bounding box around the fixture view, −85.4,25.6 to −79.4,29.6, 600×400 — 240,000 bytes at one byte a pixel, inside the default image cap | The same provider as the next row | The same |
| `radar/indiana-2026-09-19.png` | A second radar image, −91,37.5 to −80,44.5, 1100×700. **Not part of the memory fixture** — at one byte a pixel it is 770,000 bytes, over the default image cap, which is what the refusal test wants (plan task 10.19); it is also the image PLAN checked the provider's colour table against, pixel by pixel | Iowa Environmental Mesonet, Iowa State University | "In the public domain and may be used freely by anyone for any lawful purpose"; attribution appreciated, and given |
| `tiles-urban-z14/` | Four heavy city tiles — central New York, Tokyo, London and Paris at zoom 14 — for the decoder's size tests (plan task 03.12) | OpenFreeMap, the same version-pinned address | The same as the other tiles |
| *(no file)* the temperature grid | 64×48 values, **built by the benchmark from a stated rule** — 20 + 10 · sin(row ÷ 8) · cos(column ÷ 8), in °C — because only its size matters to memory | — | — |
| `radar/provider-colour-table.json` | That provider's published table: 256 entries, colour to reflectivity | The same | The same |

## Data notice

The vector tiles contain map data **© OpenStreetMap contributors**, available under the Open Database Licence (ODbL), in the **OpenMapTiles** schema (© OpenMapTiles, CC-BY 4.0), served by OpenFreeMap. They are included unmodified, for testing only.
