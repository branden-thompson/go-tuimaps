# DISCOVER — Tier 1 Synthesis

| Field | Value |
|---|---|
| Phase | DISCOVER (FULL RCC) |
| Date | 2026-09-18 |
| Inputs | [`research/AI-1`](research/AI-1-upstream-module-read.md) · [`research/AI-2 / AI-7`](research/AI-2-AI-7-mapscii-delta-licensing-tile-terms.md) · [`research/AI-3`](research/AI-3-go-ecosystem.md) · [`research/AI-4 / AI-8`](research/AI-4-AI-8-watchpost-host-model.md) · [`research/AI-5`](research/AI-5-weather-overlay-sources.md) |
| Baselines | TerminalMap `3b96072` (v0.1.0) · MAPSCII `4fe9a60` · Watchpost as checked out 2026-09-18 |
| Status | Research input, **as it stood before the rulings**. Every open question in §5 has since been ruled (D-11..D-37 in `rulings-discover.md`), and the risk statuses in §4 are superseded by `risk-assessment.md`. Kept unchanged as the record of what Tier 1 found. Two corrections: "MBTiles" as the local-file format became **PMTiles** by ruling D-18; and of AI-2's five MAPSCII additions marked worth adding, three were taken up (pointer input, D-17; headless render, D-13; local file source, D-18) and two — per-layer label margin and clustering, and loading a style by file path — have **no disposition yet** and are carried to PLAN. |

Each report's headline claims were re-checked against the primary source before it was accepted (see the Verification row in each file). One row in AI-2 was wrong and was corrected on the record (X-1).

## 1. Composition — findings that combine into something neither says alone

1. **One cell model carries the whole design.** Upstream's buffer already holds a glyph, a foreground and a *dormant* background per cell (AI-1 §8). The closest prior art draws the weather field as cell background under braille map lines (AI-5 §6). The host wants a cell grid it can place without re-measuring (AI-3 §6). Together: a cell is `{glyph, fg, bg}`; the basemap owns the foreground dots, a gridded field owns the background, and points, outlines and wind glyphs sit on top.
2. **"A lot of overlay data" is small once it reaches the screen.** A field needs one sample per cell: 1,920 values at 80×24, about 25,000 on a very large terminal (AI-5 §5, §6). Feature overlays are at most a few thousand vertices after simplification, and Watchpost's realistic load is about 400 features per view (AI-4 §4). R-4 is therefore a contract-and-simplification problem, not a throughput problem.
3. **Upstream's runtime model cannot be embedded as written.** It awaits tiles one at a time and returns no frame until all have resolved, with no timeout (AI-1 §5). Watchpost redraws every 300 ms from one clock, forbids goroutine tickers, and must never block while drawing (AI-4 §2, §9). The port needs three things upstream lacks: a synchronous render from cache, a separate cancellable fetch the host schedules, and host-driven animation time. This is a deliberate extension, not a parity row.
4. **The non-braille renderer is a first-class requirement, and upstream's is broken.** Watchpost's stated glyph floor is block and arrow characters; braille is outside it (AI-4 §6). Upstream's block mode uses dot masks that do not match its own braille layout, so the glyphs misrepresent the pixels (AI-1 §7.1, verified).
5. **Memory is the binding constraint on M4.** Watchpost has a 10.7 MB margin under its 126 MB gate and a pinned per-frame allocation budget (AI-4 §5). Upstream reallocates its canvas every frame, deep-clones parsed tiles on every cache hit, and bounds its cache by entry count, not bytes (AI-1 §7.11, P-46). A byte-budgeted cache of decoded tiles and a near-zero-allocation cached frame are requirements, not optimisations (AI-3 §5).
6. **Attribution is a new requirement on three fronts.** OpenStreetMap's guidelines do not accept attribution that needs interaction to find (AI-2 §4). OpenMapTiles and OpenFreeMap require credit (AI-2 §2, §3). Overlay sources add their own (AI-5 §7). Upstream shows none (AI-1; verified). The library needs to report the attributions in force, basemap and overlays together, so the host can place them.
7. **Hazard placement for alerts depends on work outside this repo.** 90% of live NWS alerts carry no polygon and point to zone shapes of 900 to 14,000 vertices (AI-5 §1, re-measured). Watchpost keeps no polygon at all today and reduces the ones it receives to a single vertex (AI-4 §4, verified). M1's alert-polygon scenario needs a Watchpost change plus simplification and clipping somewhere in the chain.

## 2. Convergence — independent findings that agree

1. **Framework-neutral core; any Bubble Tea adapter lives apart.** Reached independently from the upstream read (AI-1 §5), the library survey (AI-3 §10) and the host read (AI-4 §9). Watchpost has no sub-model composition and one clock, so a bundled adapter would need host routing anyway and would pin every consumer to one charm release.
2. **Parity has to be behavioural.** The library survey recommends scanline fill over triangulation and an occupancy bitmap over upstream's linear scan (AI-3 §2, §3); the upstream read lists seventeen defects and quirks, several of which change pixels when fixed (AI-1 §7); MAPSCII has its own defects that must not be inherited (AI-2 §1). Pixel-level parity would freeze known bugs. The evidence supports: same behaviours, same constants where they are intentional, with a recorded defect ledger.
3. **The host fetches; the library draws.** Every overlay source maps onto a small set of plain input shapes with no weather API knowledge in the library (AI-5 §5); Watchpost routes all HTTP through its own client and replaces data wholesale as immutable values (AI-4 §9).
4. **The tile schema is a seam.** Upstream carries a shim that renames OpenMapTiles layers to a Mapbox v6 vocabulary because its styles are from that era (AI-1 P-40); OpenFreeMap states an intent to move to a different schema (AI-2 §3, §5). Styles written directly against OpenMapTiles remove the shim and the style-provenance question (X-4) at once.

## 3. Contradictions — and how each stands

| ID | Conflict | Standing |
|---|---|---|
| X-1 | AI-2 reported that upstream label collision uses an R-tree crate; AI-1 read the code and found a linear scan. | **Resolved.** The crate is declared and never referenced (grep of `src/`: 0 hits). AI-2 corrected on the record. |
| X-2 | Upstream's README claims mouse panning and an initial zoom that fits the terminal; the code has neither (AI-1 §7.14). | **Needs ruling (OQ-1).** Proposed: the code at the pinned commit is the baseline; README-only claims are not parity rows. |
| X-3 | Upstream emits 256 colours only (AI-1 P-05); the survey recommends truecolor cells with downgrade by the host (AI-3 §6); Watchpost's tokens mix 256-colour and truecolor values (AI-4 §3). | **Open for PLAN.** Not blocking DISCOVER. |
| X-4 | Both style files are byte-identical to MAPSCII's and appear derived from Mapbox's open styles (BSD for the JSON, CC BY 3.0 for the design); neither upstream carries that notice (AI-2 §2, identity verified; the derivation is an inference). | **Needs ruling (OQ-12).** |
| X-5 | A zero-dependency core means about 1,500 hand-written lines that parse untrusted network bytes (AI-3 §10), against a proven library with a deprecated transitive dependency (AI-3 §1). | **Open for PLAN**, with fuzzing and differential tests as the stated mitigation. |

## 4. Risk signals — status after Tier 1

| ID | Risk | Was | Now | Why |
|---|---|---|---|---|
| RS-1 | "Parity" is unbounded | High | **Partly mitigated** | 72 candidate rows with file:line (AI-1 §3), a 17-item defect ledger, and 4 MAPSCII additions exist. Freezing still needs the OQ-1 ruling. |
| RS-2 | Overlay contract is a new seam | High | **Partly mitigated** | Five input shapes proposed with real sources mapped to each (AI-5 §5); two injection points located in the pipeline (AI-1 §8). |
| RS-3 | Gridded legibility at braille resolution | High | **High — reframed** | A technique with a precedent exists (field as background), at one sample per cell. Unproven until rendered specimens exist (Tier 2). |
| RS-4 | Embed-first vs. app-first | Medium | **Confirmed; mitigation defined** | Composition 3. |
| RS-5 | Generality vs. YAGNI | Medium | **Active** | Watchpost can feed only points, circles and per-place vectors today (AI-4 §4). |
| RS-6 | Free third-party tile server | Medium | **Escalated within Medium** | Single maintainer, "may discontinue… without notice", planned schema change (AI-2 §3). Mitigations: local MBTiles, user-set URL, schema seam. |
| RS-7 | Embed cost in the host | Medium | **Escalated to High** | 10.7 MB margin and a pinned per-frame allocation budget (AI-4 §5). |
| RS-8 | Port fidelity | Medium | **Active** | Every parity row cites file:line; differential tests proposed (AI-3 §9). |
| RS-9 | Commit hygiene | Low | **Mitigated** | Every commit and file checked. |
| RS-10 | *new* Licence and attribution compliance | — | **Medium** | X-4, composition 6, ODbL notice for embedded tiles (AI-2 §2). |
| RS-11 | *new* M1 depends on a Watchpost change | — | **Medium** | Composition 7. |
| RS-12 | *new* Hand-written parsers on untrusted input | — | **Medium**, if the zero-dependency posture is chosen | X-5. |
| RS-13 | *new* Overlay source terms | — | **Low for the library** | Open-Meteo is non-commercial and counts each location as a call; RainViewer is personal-use only (AI-5 §7). These bind the host, not the library. |

## 5. Open questions — status after Tier 1

| ID | Question | Status | What the evidence supports |
|---|---|---|---|
| OQ-1 | Parity baseline and surface | Partly answered | Behavioural parity against the **code** at `3b96072`; README-only claims excluded; defects handled by a ledger (fix / replicate, per item). |
| OQ-2 | "OpenMaps" | Answered, needs confirmation | OpenStreetMap data via OpenFreeMap: OpenMapTiles schema, zoom 0–14, keyless, commercial use allowed, attribution required. |
| OQ-3 | Overlay kinds in v1 | Partly answered | Features and scalar grids are essential; vector grids are cheap once grids exist; a georeferenced image is needed only if radar is a launch feature. |
| OQ-4 | Who fetches overlay data | Answered, needs confirmation | The host. The library takes an injectable fetcher for basemap tiles only. |
| OQ-5 | Standalone app; headless | Open | A one-shot headless render is cheap and doubles as the golden-test harness. Upstream's app is about 230 lines. |
| OQ-6 | Platform and terminal floor | Partly answered | Watchpost: 80×24, UTF-8, block and arrow glyphs, `NO_COLOR`, no mouse, macOS / Linux / Windows. |
| OQ-7 | Relationship to go-studs | Partly answered | A plain host-supplied palette; no design-system dependency. go-studs is an in-tree copy inside Watchpost with no module of its own. |
| OQ-8 | Module path and visibility | **Open — blocks the P10 gate** | — |
| OQ-9 | `main` only via PR | Open | — |
| OQ-10 | CGO-free | Answered by evidence | Watchpost's release matrix builds with `CGO_ENABLED=0`. Hard requirement. |
| OQ-11 | Embedded offline tiles | Partly answered | 490 KB at zoom 0–1. A separate opt-in assets package lets a consumer leave them out. |
| OQ-12 | *new* Style provenance | Open | Carry the Mapbox notice, or write fresh styles against OpenMapTiles. |
| OQ-13 | *new* Attribution display | Open | Library reports attributions; whether it also draws a default on-map line. |
| OQ-14 | *new* Polygon simplification and clipping | Open | Library or host. Zone shapes make it unavoidable somewhere. |
| OQ-15 | *new* Radar at launch | Open | Decides whether the georeferenced-image shape is v1. |
| OQ-16 | *new* Layers upstream styles but never draws | Open | Rivers, parks and airports: replicate the gap or draw them. |
| OQ-17 | *new* Pointer intents in the API | Open | `ZoomAt` / `DragBy` for hosts with a mouse; Watchpost has none. |
| OQ-18 | *new* Local MBTiles source in v1 | Open | The cheapest real offline story (AI-2 §1). |

## 6. Implications for Tier 2 and for PLAN

1. **Rendered specimens come next (AI-6).** RS-3 is the risk that decides whether M1 is reachable. Specimens at Watchpost's real rectangles — 69×12 and 149×38 cells (AI-4 §8) — of: a field as background under braille lines; the same view in the block fallback; a wind-arrow lattice; a polygon outline with tint. Nothing about overlay rendering is ratified from prose.
2. **The parity matrix can be frozen once OQ-1 and OQ-16 are ruled.** Source: AI-1 §3 plus the four MAPSCII additions.
3. **M2 and M4 can now be given numbers.** Proposed for ruling: M4 ≤ 8 MB added resident memory with one map mounted, tile cache hard-capped by bytes, cached-frame allocations within the host's existing pin; M2 as drafted (≤ 1 s warm, ≤ 3 s cold).
4. **Requirements to add at DISCOVER exit**, all from evidence above: render never blocks; host-driven time; injectable fetcher; byte-bounded caches; deterministic output for golden tests; non-braille renderer; attribution reporting; host-supplied palette; input as intents; pure Go.
5. **One dependency on the host is now explicit** (RS-11) and belongs in Watchpost's backlog, not this repo's.
