# The pinned fixture

Test data for the memory and timing requirements (NFR-3, NFR-5; rulings D-29, D-48, D-85). Pinned in PLAN on 2026-09-19 and described in `06_docs/02_features/go-tuimaps/03-architecture-design/memory-measurement.md`. `HASHES` lists every file with its SHA-256; a test (plan task 00.9) fails if any file changes.

| Part | What | Source, fetched 2026-09-19 | Terms |
|---|---|---|---|
| `zones/` | Ten county zones of Florida's Gulf coast — 56,827 vertices in 1,107 polygons | The United States National Weather Service's public zone shapes | A work of the United States government; public domain |
| `tiles-gulf-z6/` | The four zoom-6 vector tiles of the fixture view (Tampa Bay, 149×38 cells, zoom 6.4) | OpenFreeMap, at the version-pinned address `planet/20260913_164504_pt` | © OpenMapTiles · © OpenStreetMap contributors — see below |
| `tiles-midwest-z5/` | The four zoom-5 tiles named in NFR-5 — the heavier variant | The same | The same |
| `radar/indiana-2026-09-19.png` | One radar image for one bounding box, −91,37.5 to −80,44.5, 1100×700 | Iowa Environmental Mesonet, Iowa State University | "In the public domain and may be used freely by anyone for any lawful purpose"; attribution appreciated, and given |
| `radar/provider-colour-table.json` | That provider's published table: 256 entries, colour to reflectivity | The same | The same |

## Data notice

The vector tiles contain map data **© OpenStreetMap contributors**, available under the Open Database Licence (ODbL), in the **OpenMapTiles** schema (© OpenMapTiles, CC-BY 4.0), served by OpenFreeMap. They are included unmodified, for testing only.
