# Project Brief — go-tuiMaps

| Field | Value |
|---|---|
| Report | project-brief v1.0.0 |
| Phase | pre-DISCOVER (collect-brief handoff) |
| Date | 2026-09-18 |
| Author of record | Branden Thompson (HUM LEAD) |
| Branch | `feature/go-tuimaps` |
| Status | APPROVED — HUM LEAD, 2026-09-18 |

---

## Summary & Intent

go-tuiMaps is a terminal map renderer written in Go: a live, zoomable, real-world map that another terminal application can mount inside its own layout, with the host's own data drawn on top. It is the third iteration of a proven concept. MAPSCII came first and TerminalMap (Rust) second. go-tuiMaps carries forward what TerminalMap does today and adds overlays, which neither predecessor has.

**Why now / why it matters.** Watchpost tells a person what the weather and hazards are at their places, but not where. An alert, a front, a wind shift or a precipitation band is a geographic fact, and today it arrives as text and numbers. The next Watchpost capability, weather maps, cannot exist without an embeddable map that accepts weather data as layers.

**Who benefits.** Watchpost users first, especially those in severe-weather and wildfire regions who need to know whether something is headed toward them. After them, anyone building a Go terminal application that needs geographic context.

**If we do nothing.** Watchpost users keep leaving the terminal to locate a hazard relative to themselves. Watchpost's weather-map roadmap stays blocked. The only embeddable terminal map remains a Rust library that a Go host cannot mount in-process.

## Locked Problem Statement

> **"People who follow weather and hazards from a terminal cannot see where conditions are relative to the places they care about — storms, fronts, wind and precipitation reach them only as text and numbers — and so they leave the terminal to find out whether something is coming toward them."**

| # | Criterion | Score | Evidence |
|---|---|---|---|
| 1 | Bad Outcome | ✓ | "cannot see where conditions are… leave the terminal to find out" |
| 2 | Affected Humans | ✓ | People who follow weather and hazards from a terminal |
| 3 | Tech Agnostic | ✓ | No language, library, tile format or product is named. "Terminal" describes where these people work, the same reading Watchpost's locked statement uses. |
| 4 | Non-prescriptive | ✓ | It does not say map, port, Go or overlay. Several solutions could address it. |
| 5 | Verifiable | ✓ | Give a Watchpost user an active alert and ask where it sits relative to them. Observe whether they leave the terminal to answer. |

**Score: 5/5 — LOCKED.** Ratified by HUM LEAD 2026-09-18 (D-10, G-1).

**Refinement trace.**
- The raw input was a solution: "I want to make a golang port of TerminalMap". It scored 0/5.
- Candidate A, the end-user anchor, was approved by HUM LEAD (D-5).
- At lock, Candidate A's two sentences were folded into one. Its closing "never on a map", which named a solution, was replaced with the observable consequence: they leave the terminal.
- That rewording is the only change from the approved candidate, and it was ratified at G-1.
- The Go port, embedding and overlays remain as requirements and constraints.

## Metrics of Success

Each metric was checked for ways to hit the number without solving the problem. Ratified by HUM LEAD 2026-09-18 (D-10, G-2). Targets marked *TBD* are set at DISCOVER exit.

| # | Name | Symbol | Type | Definition (direction) | Measured in |
|---|---|---|---|---|---|
| M1 | Hazard Placement | HP | Primary | Share of a fixed scenario set for which a viewer of the embedded map can say where the condition sits relative to their place (inside or outside, direction, rough distance) without leaving the terminal. **Higher is better; target 100% of v1 overlay kinds.** Guard: the place marker and the condition must share one frame at 80×24. A map with no conditions on it scores 0. | Scripted-terminal specimen renders, plus a live HUM LEAD acceptance session inside Watchpost |
| M2 | Time to Placed View | TPV | Primary | Seconds from the host requesting a place's map to a frame showing basemap, place and at least one overlay. **Lower is better; targets *TBD*** (proposed ≤ 1 s warm, ≤ 3 s cold). Guard: timed to the frame that carries the overlay, not to the first blank or basemap-only frame. | Go benchmarks and timing instrumentation |
| M3 | Parity Coverage | PAR | Secondary | Percentage of parity-matrix rows with a passing Go test or an accepted rendered specimen. The matrix is frozen at DISCOVER exit against a pinned TerminalMap commit. **Higher is better; target 100%.** Guard: the denominator is frozen, and a row can be excluded only by a recorded HUM LEAD ruling. | Parity matrix in the docs tree, plus `go test` |
| M4 | Embed Cost | EMB | Primary | Memory and CPU added to the host with one map mounted, in steady state. **Lower is better; budgets *TBD*** against Watchpost's own memory and CPU targets. Guard: measured on a fixed set of views with labels and overlays on, so detail cannot be dropped to pass. Heap must stay flat over 1 hour. | `pprof`, soak test in VALIDATE |
| M5 | Host Independence | HI | Secondary | The library never touches the terminal, stdin, stdout or process-global state. Two instances render independently in one process, and a headless render works with no TUI framework. **Pass/fail; target pass.** Guard against a library that only works inside its own app. | Test harness: dual-instance and headless render tests |
| M6 | Correction Count | CC | Maintenance | HUM LEAD corrections per phase. **Lower is better.** Intake: 2. | REFLECT reports |

## Requirements

Verbatim from HUM LEAD intake, numbered for traceability.

**Functional (R-n)**

| ID | Requirement |
|---|---|
| R-1 | Feature / functionality parity with Rust's TerminalMap. |
| R-2 | Must be able to be embedded into Watchpost (a future version will allow for weather maps). |
| R-3 | In addition to OpenMaps and the same feature parity, must also allow for overlays — particularly weather overlays on the maps — which will likely be supplied via another API than OpenMaps for things like wind / temp / precipitation. |
| R-4 *(derived)* | The overlay path must accept a large volume of additional host-supplied data. *Derived from Other Considerations ("take a lot of additional data for overlays"); listed so it is tracked.* |

**Technical (T-n)**

| ID | Requirement |
|---|---|
| T-A | Written in Go. |
| T-B | Mountable by a Bubble Tea v2 / Lipgloss v2 host (Watchpost: Go 1.25). |
| T-C | Overlay data comes from sources independent of the basemap tile source. |
| T-D | Publishes via the `branden-thompson` GitHub account. Verified: commits author under the personal address. |

**Sharpening notes**

- **SH-1 (R-1 / M3):** "Parity" becomes testable through a pinned upstream commit and a feature-by-feature matrix. Baseline inventory from a light read of upstream:
  - MVT/protobuf parsing
  - 2×8 braille and ASCII-block rendering
  - Mapbox GL style JSON (`bright`, `dark`)
  - label collision detection
  - polygon triangulation and fill
  - memory LRU plus disk tile cache
  - embedded offline tiles, zoom 0–1
  - markers with shapes and blink / pulse / flash
  - scripted fly-to camera and tours
  - multiple independent instances
  - an API surface that does not capture input

  Still open: the parity surface (behavioral, API-shaped or pixel-level) and whether the standalone app is included (OQ-1, OQ-5).
- **SH-2 (R-2 / M5):** Proposed testable form: a host can mount one or more map instances in an arbitrary rectangle, drive them from its own event loop, and the library never owns the terminal, input or output. Whether it must also work with no TUI framework at all is part of OQ-5.
- **SH-3 (R-3, R-4):** "Overlays" covers four different rendering problems:
  - points
  - vector shapes (alert polygons, storm tracks)
  - gridded fields (temperature, precipitation, radar)
  - vector fields (wind)

  The v1 set and who fetches the data are open (OQ-3, OQ-4). The draft recommendation is that the library defines the overlay contract and the host owns the APIs.
- **SH-4 (R-3):** Upstream's default tile source is OpenFreeMap (`tiles.openfreemap.org/planet`), which serves OpenStreetMap data. "OpenMaps" is read as that (OQ-2).
- **SH-5 (R-3 / M1):** Gridded data at braille resolution (2×8 sub-pixels per cell) may not be legible. Any overlay rendering approach is approved from an actual rendered specimen at representative terminal widths, never from a prose description.

## Technical Constraints

HUM LEAD stated "none known". These standing project rules apply:

| ID | Constraint |
|---|---|
| C-1 | Go; binaries build to `./dist/`; SemVer with tagged releases; tests first for all Go code (FULL TDD). |
| C-2 | No AI attribution or watermarks in commits, PRs, code or shipped artifacts. The development harness stays untracked. |
| C-3 | Upstream TerminalMap and MAPSCII are MIT licensed. Derived work preserves the required notices and credits both. |
| C-4 | OpenStreetMap and tile-provider attribution and usage terms are respected by design (cache, backoff, visible attribution in the host). |
| C-5 | Personal-account git identity only; no push without explicit HUM LEAD instruction. |

## Other Considerations

- **Lineage:** MAPSCII (rastapasta), then TerminalMap (psmux; Rust, MIT, app plus SDK, about 15 source files), then go-tuiMaps. Two prior implementations mean the basemap problem is well understood. The overlay contract is new.
- **First consumer:** Watchpost, a terminal weather station using Bubble Tea v2, Go 1.25, a domain-first layout and SEV-0 process. Its alert, wind, precipitation, fire and seismic data are the first overlay sources.
- **Design system:** Watchpost renders through go-studs. Whether go-tuiMaps takes theme tokens from go-studs or stays independent of any design system is open (OQ-7).
- **Process reference:** Watchpost's docs tree, brief format, no-watermark calibration and PR protocol are the working examples for this project.
- **Timeline:** none stated.

## Discovery Handoff Package

### Areas to Investigate

| ID | Area | Concrete artifact / question |
|---|---|---|
| AI-1 | Upstream module read | For `renderer.rs`, `tile.rs`, `braille.rs`, `canvas.rs`, `styler.rs`, `label.rs`, `camera.rs`, `marker.rs`, `widget.rs`, `tile_source.rs` and `embedded_tiles.rs`: responsibilities, data flow, and what each one's parity row is. |
| AI-2 | MAPSCII delta | What MAPSCII does that TerminalMap dropped or changed, to decide whether any of it belongs in the matrix. |
| AI-3 | Go ecosystem | MVT/protobuf decoding, polygon triangulation, label collision (R-tree), and Mapbox GL style filter evaluation. For each, choose between an existing library and a hand port, and check whether CGO is needed. |
| AI-4 | Host model | The Bubble Tea v2 component and render model, where Watchpost would mount a map, how async tile fetches become messages, and the resize path. |
| AI-5 | Overlay source formats | NWS alert polygons (GeoJSON/CAP), gridded forecast fields, radar and precipitation tiles, wind fields; projections, and which sources can be licensed for use. |
| AI-6 | Rendering legibility | Rendered specimens of gridded and vector-field data at braille resolution; options for color, density and glyphs. |
| AI-7 | Tile provider terms | OpenFreeMap and OSM attribution, rate expectations, caching rules, alternatives and self-hosting. |
| AI-8 | Budgets | Watchpost's current memory and CPU headroom, to set the M2 and M4 targets. |

### Stakeholders to Consider

| Who | Role | Mode |
|---|---|---|
| Branden Thompson | HUM LEAD; product owner; Watchpost maintainer; approver of every gate | HUMAN LEAD (SEV-0) |
| Watchpost users | People who will see the maps | Represented via personas in red-team |
| Future Go TUI embedders | Consumers of the library API | Represented by M5 |
| Upstream authors (psmux, rastapasta) | MIT licensors | Credited; informed at SHIP if desired |
| Tile and weather data providers | Terms and rate-limit constraints | Constraint source (AI-5, AI-7) |

### Risk Signals

| ID | Risk | Severity | Why it is a risk |
|---|---|---|---|
| RS-1 | "Parity" is unbounded | High | Without a pinned commit and a frozen matrix, R-1 can never be declared done. |
| RS-2 | The overlay contract is a new design seam | High | Neither predecessor has one, so it carries the most architectural risk and is the reason the project exists. |
| RS-3 | Gridded data legibility at braille resolution | High | If temperature or precipitation cannot be read, M1 fails regardless of engineering quality. |
| RS-4 | Embed-first vs. app-first tension | Medium | Upstream is async on a Rust runtime, and the host here is message-driven. A straight port of the structure may fight the host. |
| RS-5 | Generality vs. YAGNI | Medium | Designing for hypothetical embedders beyond Watchpost could bloat the API. |
| RS-6 | Dependence on a free third-party tile server | Medium | Availability and terms are outside our control. Offline tiles cover only zoom 0–1. |
| RS-7 | Embed cost inside a host with published budgets | Medium | Tile decoding, triangulation and label layout on every pan could breach Watchpost's memory and CPU targets. |
| RS-8 | Port fidelity | Medium | A re-typed port drifts silently from upstream behavior. Parity rows need tests or specimens, not recollection. |
| RS-9 | No-watermark enforcement | Low | The harness injects attribution trailers by default. One slipped at intake and was amended before anything was pushed. |

### Open Questions (for DISCOVER, FULL RCC)

| ID | Question |
|---|---|
| OQ-1 | Parity baseline: which upstream commit or tag, and which surface (behavioral, API-shaped or pixel-level)? |
| OQ-2 | "OpenMaps" means OpenStreetMap data via OpenFreeMap vector tiles. Confirm. |
| OQ-3 | Which overlay kinds are in v1: points, vector shapes, gridded fields, vector fields? |
| OQ-4 | Does the library fetch overlay data, or only render what the host hands it? |
| OQ-5 | Is the standalone app in scope, or only the library? Must it run headless with no TUI framework? |
| OQ-6 | Platform and terminal floor: Windows, macOS and Linux; SSH; 256-color versus truecolor; minimum size? |
| OQ-7 | Relationship to go-studs: themed through it, or independent of any design system? |
| OQ-8 | Module path (`github.com/branden-thompson/go-tuimaps` proposed) and repo visibility. |
| OQ-9 | Same ops protocol as Watchpost: `main` only via PR using the A2DH PR template? |
| OQ-10 | Is a pure-Go, CGO-free build a requirement, given Watchpost's install story? |
| OQ-11 | Embedded offline tiles add roughly 0.5 MB to the host binary at zoom 0–1. Is that acceptable, optional, or a candidate for a deeper zoom? |

### Context carried forward

- **Problem statement:** refined and locked (5/5) above.
- **Sharpening observations:** SH-1 to SH-5.
- **Backlog review:** no `06_docs/02_features/*/backlog.yml` exists (new project), so it was skipped.

## Brief Metadata

| Field | Value |
|---|---|
| Header | **NEW MAJOR PROJECT \| 'go-tuiMaps'** |
| Scope adjective | Major |
| Project type | Project (Go library + terminal application) |
| Project name / branch | `go-tuimaps` / `feature/go-tuimaps` |
| LEVEL / SEV | LEVEL-1 / SEV-0 (HUMAN LEAD) |
| Phase instructions | FULL GIT; FULL DOCS; FULL REPORTS; FULL DIAGRAMS; FULL TDD; FULL RCC (DISCOVER); FULL PLAN |
| Theme | BRTOPS |

**Completeness scorecard**

```
PROJECT BRIEF — COMPLETENESS CHECK
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  [✓] Header              — Major Project | 'go-tuiMaps'
  [✓] Directives          — LEVEL-1; SEV-0; 7 FULL directives
  [✓] Summary / Intent    — what/why/who/if-not; statement 5/5
  [✓] Requirements        — 3 functional + 1 derived; 4 technical
  [✓] Metrics of Success  — 6, anti-solution checked, ratified
  [✓] Tech Constraints    — 5 standing rules
  [✓] Considerations      — lineage, first consumer, references
  Required sections: 5/5 complete
  Overall: READY FOR DISCOVER — APPROVED
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

BRIEF SUFFICIENCY CHECK
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  [✓] WHAT to investigate        — AI-1..AI-8
  [✓] WHO to consider            — 5 stakeholder groups
  [✓] WHERE to look              — upstream source, MAPSCII,
                                   Watchpost, provider docs
  [✓] Targeted questions         — OQ-1..OQ-11
  [✓] Risks                      — RS-1..RS-9
  Sufficiency: READY FOR DISCOVER
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

## Decision Log

| # | Date | Decision | By | Rationale (verbatim where given) |
|---|---|---|---|---|
| D-1 | 2026-09-18 | Thin A2DH install | HUM LEAD | "li A2DH STARTUP + THIN INSTALL" |
| D-2 | 2026-09-18 | LEVEL-1, SEV-0, five FULL directives | HUM LEAD | "LEVEL-1; SEV-0; FULL GIT; FULL DOCS; FULL REPORTS; FULL DIAGRAMS; FULL TDD" |
| D-3 | 2026-09-18 | FULL RCC and FULL PLAN added | HUM LEAD | "FULL RCC; FULL PLAN approved" |
| D-4 | 2026-09-18 | FULL TDD bound to the framework's test-driven-development skill | HUM LEAD | "FULL TDD should have some skills in the skillfamily" |
| D-5 | 2026-09-18 | Problem statement anchor: end user | HUM LEAD | "CANDIDATE A approved" |
| D-6 | 2026-09-18 | Git initialized; `main` plus feature branch; no remote, no push | HUM LEAD | "Git init + recommended steps approved" |
| D-7 | 2026-09-18 | No AI attribution or watermarks | HUM LEAD | "NO AI ATTRIBUTION / WATERMARKS rule is in effect". A trailer on the first commit was amended out before any push. |
| D-8 | 2026-09-18 | Personal git identity | HUM LEAD | "this is a personal project so git identity must be branden-thompson". Verified: the personal address. |
| D-9 | 2026-09-18 | Watchpost is the working example; harness untracked; root commit rewritten to a `.gitignore` only | Agent under D-7/D-9; ratified by HUM LEAD at G-7 (D-10) | "use watchpost as an example if needed" |
| D-10 | 2026-09-18 | G-1..G-7 approved: problem statement ratified; metrics M1–M6 ratified; brief approved; branch renamed `feature/go-tuimaps`; project config set (SEV-0, Go, BRTOPS); Watchpost's no-watermark calibration copied verbatim into the local harness (diff-verified identical); D-9 ratified. GO for DISCOVER. | HUM LEAD | "G-1 thru G-7 Approved; Approved; GO 4 DISCOVER" |

## Next Steps

The DISCOVER entry gate runs (config loaded, startup complete), followed by DISCOVER under FULL RCC, starting with AI-1 and the OQ rulings.
