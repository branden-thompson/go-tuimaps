# AI-4 / AI-8 — Watchpost as a host: model shape, data in memory, budgets

| Field | Value |
|---|---|
| Phase | DISCOVER — Tier 1 research |
| Date | 2026-09-18 |
| Scope | What the first host requires of the embedding contract; which overlay kinds its data can feed today; the memory and allocation headroom (OQ-3..OQ-7, M2, M4, RS-4, RS-7) |
| Status | Section 9 is research input for PLAN. Nothing is decided until HUM LEAD rules. |
| Verification | Spot-checked by the coordinator against the Watchpost tree: national alert geometry reduced to a single point (domains/globalfeed/nws.go:245); the 10.7 MB RSS margin (06_docs/perf-measurement.md:84); CGO_ENABLED=0 release builds (Makefile:527); no mouse handling anywhere in app/, modes/, platform/, cmd/; the block-and-arrow glyph floor and 80×24 minimum (README.md:77-78). All held. |

All paths are relative to the root of the Watchpost repository (github.com/branden-thompson/watchpost) as checked out on 2026-09-18. This was read-only research; nothing was built, run or modified.

## 1. Architecture in one page

- **Layers.** Watchpost is one Go module with five roots (`architecture.md:23-42`). `app/` is the composition root. `domains/*` holds the data sources and `platform/*` the shared leaves (`snapshot`, `sched`, `httpx`, `render`, `term`, `geo`). `modes/{tty,report}` are the renderers, and `pkg/schema` is the published JSON schema.
- **Import rule.** `modes/*` and `platform/*` never import `domains/*` (`architecture.md:45`). A textual lint with a self-test enforces this: `scripts/lint-imports.sh:2-3,44-53`, `Makefile:97-98`. Domain data reaches the TUI three ways:
  - as an immutable `*snapshot.Snapshot`;
  - as typed messages sent with `p.Send` (`app/pipelines.go:178,316`);
  - as function hooks placed on the model (`app/dashboard.go:503-520`).
- **Where an external visual component is wired.** Three steps:
  1. Add a seam file in `platform/render`. The go-studs seam is the precedent (`architecture.md:162-167`). The seam maps theme tokens to the library's palette.
  2. `app/` converts domain geometry into overlay values and hands them in, following the `WithSpectrum` hook precedent (`modes/tty/dashboard.go:1509`).
  3. `modes/tty` draws the result inside a window.
- **Doc and code disagree on three points. Trust the code.**
  - The doc says `platform/render` is the only go-studs importer. `modes/tty/broadcaster.go:28` also imports it.
  - The doc pins bubbletea v2.0.9, lipgloss v2.0.6 and bubbles, with go-studs via `replace` (`architecture.md:7`). `go.mod:6-7` has v2.0.3 and v2.0.2, no bubbles and no `replace`. go-studs is an in-tree copy with no `go.mod` of its own.
  - The doc describes a view registry, a `View` interface and `SetKeyMap` (`architecture.md:197-210`). None of them exists; `app/registry.go` and `app/keymap.go` are absent.

## 2. Bubble Tea v2 model shape

- **Root model.** `Router` holds two surfaces by value, `Dashboard` and `Broadcaster` (`router.go:287-307`). There are no sub-models. Windows are method groups on `Dashboard`, selected by a single `modal` value (`view.go:78-117`).
- **Update routing.** Messages are classified by type (`router.go:474-496`):
  - Program-scoped messages go to both surfaces. These are size, background, focus, colour-profile and snapshot messages (`router.go:328-348`).
  - Console-scoped messages go to the console (`358-364`).
  - Observer-scoped replies go to `Dashboard`. The comment at `366-400` notes that a reply delivered to the wrong surface is silently dropped.
  - Everything else goes to the active surface.
- **View.** `View` returns a `tea.View` carrying a `Content` string, `AltScreen` and `BackgroundColor` (`view.go:19-46`). Windows are composited into that string with the lipgloss Compositor/Layer API (`platform/render/panel.go:274-297`). No Bubble Tea layers are used.
- **Resize.** `WindowSizeMsg` only stores `width` and `height` (`dashboard.go:925-927`). Geometry is recomputed once per frame (`layout.go:35`) and passed down as `render.Opts{Width…}` (`view.go:362-365`). The doc's rule is "widths always passed explicitly" (`architecture.md:173`).
- **Ticks.**
  - `tickEvery` is the only constructor of a clock (`dashboard.go:816-820`).
  - The 300 ms tick is armed after every `Update`, and only while `tickNeeded()` is true (`829-850`, `861-877`, `891-899`).
  - A 50 ms tick re-arms itself behind a `vizTicking` guard (`radio_panel.go:295-312`).
  - Goroutine tickers are forbidden (`architecture.md:172-173`).
- **Async I/O.** Three patterns are in use:
  - a closure `tea.Cmd` that returns a typed message (`modal_location.go:121-130`, `locate.go:133-143`);
  - a sequence-numbered debounce, where the tick always fires and a stale sequence is discarded (`locate.go:58-60,111-124`);
  - pushes from goroutines through `p.Send`, coalesced to 5 s (`pipelines.go:59`; `app/radio.go:466-471`).
- **Memoisation.** The frame body is cached on a complete key that includes `ThemeGeneration`. The slot is a pointer because `View` is a value receiver (`memo.go:8-11,34-56`).
- **Closest existing pattern: the radio visualiser.**
  - The feed is a hook, `Spectrum func() []float64` (`dashboard.go:133,1509`), wired by `app/dashboard.go:503`.
  - Its lifecycle is `vizActive`, `armViz` and `vizFrame` (`radio_panel.go:276-312`), driven by `vizTickMsg` (`dashboard.go:879-882,953-954`).
  - It is a pull model. The host's tick pulls the latest state; the producer never pushes frames.

## 3. Render pipeline and theming

- **Colour values.** Every colour is a semantic `Token` resolved by `render.Tok()` (`platform/render/theme.go:9-15,404`). Values are raw SGR parameter strings such as `"38;5;250"` or `"48;2;29;40;48"`. Window and gradient tokens are `#RRGGBB` (`theme.go:4-6`; `themes.go:20-22`).
- **Themes.** Thirteen themes swap at runtime, and `ThemeGeneration()` invalidates the memo (`themes.go:12-34`). WCAG AA contrast is tested per token (`docs/extending.md:129-135`).
- **Colour gate.** One gate applies: go-studs `ColorsEnabled()`, which checks `NO_COLOR` and whether stdout is a TTY (`sgr.go:27-40`; `platform/term/term.go:6,123`).
- **Colour profile.** `tea.ColorProfileMsg` is fanned out but handled nowhere (`router.go:323-331`). Truecolor is emitted raw, and downgrade is left to Bubble Tea's writer.
- **Light and dark.**
  - `Dashboard` sends `tea.RequestBackgroundColor` at `Init`.
  - The reply is stored as `darkBG` (`dashboard.go:887,936-938`).
  - Dark themes paint their own ground regardless (`theme.go:383-387`).
- **Width measurement.** Width is `go-runewidth` applied to ANSI-stripped text (`text.go:215,393-394`).
- **What a map must accept from the host.**
  - A plain palette struct with roles such as land, water, road, label and overlay severities. Each role is a foreground and background, either RGB or an ANSI-256 index.
  - A colour on/off flag.
  - An ASCII flag (`view.go:364`).
  - A dark/light hint.
  - The library should return lines of exactly the requested width. It should not define its own theme interface.
- **go-studs.** It is an in-tree copy at `third_party/go-studs`, under the import path `github.com/branden-thompson/watchpost/third_party/go-studs/...`. It is MIT and by the same author (`README.md:412-414`). Upstream commit `3e85e77` is patched through `patches/` and tracked in `LOCAL_CHANGES.md`.

## 4. Geographic data already in memory

| Data | Go type and location | Geometry | Cadence (priority / recent) | Volume |
|---|---|---|---|---|
| Places | `snapshot.Location` `types.go:33-39`; `LocationRef` `:412-430` | point | per commit | 10 watchlist + 50 recent (`modal_location.go:113-114`), plus the station pool |
| Alerts per place | `snapshot.Alert` `types.go:178-198` | **none**: zone codes and `AreaDesc` only. Fetched by zone (`nws/alerts.go:74`) | 20 s / 2 min (`pipelines.go:127,220`) | a handful |
| National NWS alerts | `globalfeed.Event` `event.go:34-65` | GeoJSON polygon **reduced to its first vertex** (`globalfeed/nws.go:117-147,245-253`). `HasPoint=false` when the alert is zone-only | 2 min (`app/ticker.go:91`) | about 30 per lane; 60 rows in the bench (`bench_test.go:123,216`) |
| Tropical storms | `TropicalDetail` `detail.go:31-46`; `nhc.go:46-49,97-98` | point, plus motion direction/speed and wind in knots. No track or cone | 2 min | 0–5 |
| Earthquakes (global) | `Event` with `QuakeDetail`; `usgs.go:137-138` | point with magnitude | 2 min | tens |
| Earthquakes (per place) | `snapshot.Quake` `types.go:264-277` | **no lat/lon**: distance and bearing only. The provider drops them (`seismic/usgs/usgs.go:312-318`) | 5 / 15 min | tens |
| Fire detections | `snapshot.Hotspot` `types.go:231-239` | point with fire radiative power (FRP) | 10 / 15 min | up to 300 per place (`:205`); upstream about 25k (`hms.go:5`) |
| Wildfire incidents | `snapshot.Incident` `types.go:242-251` | point with acres. No perimeter (`wfigs.go:154-157`) | 10 / 15 min | a few |
| Wind | `Conditions.Wind/WindDirDeg/WindGust` `types.go:133-135`; `Hourly` `:160-161` | one vector per place | 90 s | 60 |
| Temperature and precipitation | `Conditions`, `Hourly` | one scalar per place. No grid | 90 s / 30 min | 60 |
| Radar | none. Only "Radar Indicated" text in `domains/severe/detection.go` | none | — | — |
| Buoys and tide stations | `snapshot.Marine` `types.go:84-85,92-93` | station id and distance only. Coordinates stay private in the providers (`ndbc.go:53-56`, `coops.go:61-64`) | 10 min | 1–3 per place |
| Weather-radio (NWR) transmitters | `stream.Transmitter` `stream/table.go:28-37` | point, with county codes | static, embedded | 1,035 rows |
| Station area | `StationAreaMsg` `broadcaster.go:100-102`; fire and alert radii `app/dashboard.go:358,384-385` | circle | on change | 1–3 |

Points with severity styling, circles, per-point vectors and labels are needed first. **No polygon, track or grid is in memory today.** Polygon overlays require a change in Watchpost to retain the alert geometry. A realistic overlay for one map view is at most about 400 features.

## 5. Performance budgets and gates

- **Frame allocations, cached frame.** 365–386 allocations × 1.05, which the code calls "THE PIN THAT MATTERS" (`modes/tty/bench_test.go:45-59`).
- **Frame allocations, uncached frame.** 2,188 at 80×24 up to 14,477 at 200×60, × 1.05 (`:38-43`).
- **Severe window.** Cached 1,740 × 1.05; uncached 5,248 × 1.05 (`:208`).
- **Open window.** Compositing costs 1,042 µs, 765 KB and 1,346 allocations per frame (`docs/accepted-costs.md:22-25`). A frame with no window is therefore about 137 µs.
- **Always drawing.** A frame is drawn every 300 ms for the life of the process. That accounts for 13.6 of the 23.6 MB/min allocated (`accepted-costs.md:59-61`; `dashboard.go:823-829`).
- **Memory.**
  - The resident-memory (RSS) gate is ≤ 126 MB (`multi-voice-support/07-readiness/perf-protocol.md:89-90`).
  - Measured RSS is 106.8–115.3 MB, leaving a **10.7 MB margin** (`perf-measurement.md:84`).
  - Physical footprint is 93.7–100.2 MB.
- **Enforcement.**
  - `make alloc-budget` (`Makefile:489-490`) is part of `verify-gates` (`Makefile:299`) and is listed in `required-gates.txt:33`.
  - Wall-clock benchmarks are recorded but never gated (`Makefile:492-495`).
  - Soak runs use `soak.sh` with `tools/slope` (`perf-measurement.md:18-25`).
- **Headroom.** There is about 10 MB of RSS and effectively no per-tick allocation budget.
  - The map raster must be cached, and the tile cache must be hard-bounded.
  - A background writer or a second render path would reopen the tooling decision (`perf-measurement.md:72-73`).

## 6. Conventions a dependency must not break

- **CGO.** The release matrix builds with `CGO_ENABLED=0` for darwin arm64/amd64, linux amd64/arm64 and windows/amd64 (`Makefile:6,527`; `README.md:408`). Linux needs glibc (`README.md:79`).
- **Dependency admission.** I found no written admission checklist.
  - The effective gates are `tidy` and `vuln` via govulncheck (`Makefile:93-94`).
  - The written rule is to patch narrowly, send the change upstream and never fork a dependency for performance (`docs/extending.md:140-146`).
- **Licences.**
  - `THIRD_PARTY_LICENSES.md` is generated by `scripts/third-party-licenses.sh` from `go list -m all`.
  - A module without a LICENSE file is flagged in the output (script lines 1-5, 27-40).
  - The file ships in releases (`release.yml:38`).
- **Exposure scans.**
  - An identity scan looks for home paths, scratchpad paths and e-mail addresses across tracked files (`cmd/watchpost/identity_test.go:37-51`). The licences file is exempt (`:69`).
  - A watermark lint covers files and commit messages (`scripts/lint-watermark.sh:5`).
  - Builds use `-trimpath` (`Makefile:11-17`).
  - The Go tree may not contain shell programs (`code-standards.md:113-118`).
  - These scans matter only if library code or fixtures are copied in-tree.
- **Terminal floor.**
  - A UTF-8 locale and a font with block and arrow glyphs are assumed.
  - The minimum size is 80×24 (`README.md:77-78`).
  - SSH use is claimed (`README.md:38`).
  - Braille is outside the stated glyph floor, so the `--ascii` path (`README.md:251-254`) needs an equivalent in the library.
  - `NO_COLOR` must still read in text (`README.md:257`).
- **Build output.** Binaries go to `./dist` (`Makefile:1,5,19-21`).
- **Goldens.** Golden files are keyed by size, for example `modes/tty/testdata/frame-133x44-{ascii,colour,plain}.golden`. They are re-recorded with `-update-golden` (`golden_test.go:48-51`). Library output must therefore be byte-deterministic.

## 7. Keybinding and input model

- **Keymap.** One flat, merged `KeyMap`, not layers.
  - `term.Merge` rejects duplicate keys and locks `?` to help (`platform/term/term.go:152-198`).
  - User `[keys]` overrides are applied at start-up only (`app/dashboard.go:66,87`).
  - There is no runtime swap.
- **Mouse.** Not enabled. No mouse option or mouse message appears anywhere in the tree, and `View` sets only `AltScreen` and `BackgroundColor` (`view.go:42-45`).
- **Focus.** A window that owns the keyboard takes raw keys before the keymap lookup (`dashboard.go:1090-1099`). The router sends keys to the window on top (`router.go:517-540`).
- **Conflicts.** The arrow keys are bound to navigation and alert stepping, and `+`, `-` and `=` are volume (`dashboard.go:438-445`). Two consequences follow:
  - The map needs a window that owns the keyboard.
  - The library should expose intent methods such as `Pan`, `Zoom` and `Recenter`. It should not read keys itself. Watchpost then declares `map-*` actions, so help and rebinding stay accurate.

## 8. Where a map would go

**How the rectangles are computed.**

- Usable width is W−7 (`view.go:362-364`).
- Window body height is max(5, H−12) (`:121`).
- Inner content width is the window width minus 4 (`:216-221`).
- The Severe window's 130 columns is fixed by `view.go:153`.

| Placement | 80×24 | 160×50 |
|---|---|---|
| 1. A dedicated map window at full usable width, opened from a focused place. It shows the place, its rings, hotspots, incidents and quakes. | 69×12 cells (138×48 braille dots) | 149×38 cells |
| 2. A map pane in the Severe window's detail record. National events already carry points. | 69×12 cells, so make it a toggle rather than a split | 126×38 cells; split about 60×30 beside the text |
| 3. The Broadcaster console's station-area panel: transmitter, service-radius circle and pool. | skip at this size | about 50×20 cells |

Start with placement 1. At 80×24 an inline map section inside Location Details would consume the entire 12-row body, so a separate window is the better fit.

## 9. Implications for the library contract

1. **Pure, synchronous render on demand.** The call is `Render(cols, rows) []string`, and every line must be exactly `cols` cells wide. The host passes widths explicitly (`architecture.md:173`), and the window wrapper re-wraps any line that is too long (`view.go:204-221`).
2. **Never block or fetch inside render.** The render path runs every 300 ms (`accepted-costs.md:59-61`). Expose tile fetching as a separate function taking a context, which the host wraps in a `tea.Cmd` (`locate.go:130-143`).
3. **No internal goroutine tickers.** Animation advances through a host-driven `Step(now)`, and a `NeedsTick()` method feeds the host's tick predicate (`dashboard.go:816-850`).
4. **Injectable fetcher.** Accept a `func(ctx, url) ([]byte, error)`. All HTTP in Watchpost goes through `httpx`, which handles the user agent, pacing, caching and status statistics (`httpx.go:222,369`; `app/app.go:33`).
5. **Caching.**
   - Output must be deterministic.
   - A cheap generation or dirty counter lets the host key its memo (`memo.go:34-56`).
   - The cached path must allocate almost nothing (`bench_test.go:45-59`).
6. **Bounded memory.** Tile and geometry caches need a hard byte cap of a few MB, given the 10.7 MB RSS margin (`perf-measurement.md:84`).
7. **Host-supplied palette.** Take a palette, a colour on/off flag, a dark hint, and an ASCII or block fallback renderer (`theme.go:4-6`; `sgr.go:27-40`; `view.go:364`). Braille must not be the only renderer (`README.md:77-78`).
8. **Input as intents.** Expose intent methods, with no key or mouse handling inside the library (`dashboard.go:1090-1099,438-445`).
9. **Overlays as plain values.** Overlays are points, circles, vectors, labels and polygons. They are replaced wholesale in the immutable-snapshot style (`architecture.md:135`).
10. **Dependencies.** Pure Go with no CGO. Every transitive module needs a LICENSE file and a clean govulncheck result (`Makefile:527,93-94`; `third-party-licenses.sh`).

**Strongest counter-argument to shipping a Bubble Tea adapter inside the library.** I would keep the core framework-neutral, for three reasons:

- **Watchpost has no sub-model composition.** The Router delivers messages by type, and an unclassified foreign message reaches only the active surface (`router.go:328-400,474-496`). The adapter's private messages and ticks would need host routing anyway.
- **Adapter-owned ticks would break the single-clock rule** (`dashboard.go:813-820`).
- **A bundled adapter pins the library to one `charm.land` v2 release and pulls bubbletea and lipgloss into every consumer's dependency graph and licence file.** The Watchpost doc's own version pins have already drifted from `go.mod`.

If an adapter ships at all, it should be a separate nested module.
