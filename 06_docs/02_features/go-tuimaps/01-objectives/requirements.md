# go-tuiMaps — Requirements

| Field | Value |
|---|---|
| Phase | DISCOVER (FULL RCC) |
| Date | 2026-09-18 |
| Sources | [Project brief](../08-reports/project-brief.md) · [rulings](../02-analysis/rulings-discover.md) (D-11 onward) · [red-team round 1](../08-reports/red-team-discover.md) · [glossary](glossary.md) · [parity matrix](../02-analysis/parity-matrix.md) · [defect ledger](../02-analysis/defect-ledger.md) · [specimen findings](../02-analysis/specimens/README.md) · research AI-1..AI-9 |
| Status | For HUM LEAD approval in the Discovery Report. **Revised 2026-09-18 after red-team round 1**: rows marked ◆ were rewritten to be testable; rows marked ✚ are new. Items awaiting a HUM LEAD ruling are marked *(Q-n)* and listed in the red-team record. |

Every requirement traces to a ruling, a finding, a brief item or a red-team finding. IDs are stable and are **not** in numeric order; they are grouped by subject. Words a newcomer may not know are in the [glossary](glossary.md).

## What the brief's requirements became

| Brief | Sharpened by rulings into |
|---|---|
| R-1 parity with TerminalMap | FR-1, FR-2, FR-26, FR-27 — behavioural parity on a frozen denominator of 70 rows, 62 of them in v0.1.0 (D-11, D-37, D-49) |
| R-2 embeddable in Watchpost | FR-4, FR-24, FR-25, FR-27, NFR-1..NFR-6 — a framework-neutral library imported as a Go package (D-13) |
| R-3 overlays from sources independent of the basemap | FR-6..FR-19 — five input shapes; the host fetches (D-14, D-15) |
| R-4 high-volume host data | FR-11 — a contract and simplification problem, not throughput (S1-5, D-16) |

## Functional requirements

| ID | The library / app must… | From |
|---|---|---|
| **Basemap** | | |
| FR-1 ◆ | Meet every row in the M3 denominator of the frozen parity matrix according to its disposition: reproduce the *Match* and *Replicate* rows; do the intended thing for the *Fix* rows; keep and add to the *Extended* rows. | R-1, D-11, D-37, BQ-5 |
| FR-2 | Draw waterways, parks, airport runways and airport labels. | D-12, S8-1 |
| FR-3 ◆ | Provide two renderers, each with reference-frame tests and approved specimens. **Braille is the default.** The block renderer is one a user opts into, with its own sparser profile (no roads unless asked for); the renderer can be switched at runtime in both directions. | D-23, D-42, S2-1 |
| FR-20 | Ship its own `dark` and `bright` styles written against OpenMapTiles; accept a user's style in the same JSON format, honouring every zoom stop. | D-24, L-9 |
| FR-19 | *(ratified by D-41, D-42, D-49)* Thin the basemap by renderer, map size and what is drawn on top; let water own its cells when a field or image is active. | S1-2, S2-1, S3-4, S5-1 |
| **Overlays** | | |
| FR-6 | Accept feature overlays: points, lines, polygons and circles in longitude/latitude, each with style, optional label, id and credit. | D-14 |
| FR-7 | Accept scalar grids: a regular longitude/latitude grid of values with class breaks and a ramp. | D-14, S1-5 |
| FR-8 | Accept vector grids: two scalar grids, with a helper for speed and meteorological "from" direction. | D-14, S5-1 |
| FR-9 | Accept georeferenced images, each with a **required** colour-to-intensity table, and re-colour them with the library's ramp; a provider's own colours are never shown. Matching is nearest-colour within a stated tolerance; unmatched pixels become no-data and are counted and reported, so a changed palette is visible rather than silently wrong. | D-14, D-36, D-39, D-45 |
| FR-10 | Accept tile-image providers — a host-supplied function per tile — under the same table rule. | D-14, D-36 |
| FR-11 ◆ | Simplify host shapes to the dot tolerance of the view. Shapes outside the view cost nothing to draw, and nothing is ever drawn outside the given rectangle. The simplified, *unclipped* result is cached per zoom bucket (buckets defined in PLAN; fractional zoom maps to one); on a miss, the nearest cached bucket is drawn while the right one is prepared, so a shape never vanishes mid-zoom. Simplification is iterative, never recursive, with a worst case no worse than n·log n. Host geometry is held by reference, not copied; raw and simplified bytes both count against the shape cap; input over the cap is refused with an error (NFR-20). Shapes crossing ±180° longitude are split there; the longitude convention is −180..180 and is stated. | D-16, CQ-7, CQ-10, P-4, S-3 |
| FR-12 | *(ratified by D-41, D-42, D-49)* Composite by a stated order: areas own the cell background (water, then images and fields, then tints); lines and glyphs own the foreground; labels sit on top. | specimens, cross-cutting 2 |
| FR-13 | Expose legend data — class breaks and colours — for every overlay. | D-36, S1-6 |
| FR-14 | Report the credits in force, basemap and overlays together; draw an optional one-line credit on the map, on by default. | D-25, S9-1 |
| **Colour and theme** | | |
| FR-15 | Take a host palette covering basemap roles and overlay ramps; apply changes at runtime without re-parsing tiles; expose a change counter. | D-26, D-36 |
| FR-16 ◆ | Keep every ramp — default or supplied by a theme — **ordered** (relative luminance changes in one direction along the ramp, or on each side of a labelled midpoint for a diverging ramp), **distinct** (each step maps to a different palette entry at the colour depth in use) and **readable** (line work and glyphs at least 3:1, labels and legend text at least 4.5:1, against every cell background they are drawn on, by the WCAG formula, with the foreground chosen per cell). A checker reports each violation; a host can run it in its own tests. *Colour-vision safety is proposed as part of "distinct" (Q12).* | D-36, S1-3, S10-2, A-3, A-4, CQ-14 |
| FR-17 | Render legibly at truecolor, 256 colours, 16 colours and no colour, with a ramp per depth; take the depth as a hint from the host. | D-23, S10-1 |
| FR-18 | With no colour, draw smooth fields as contours with value labels and patchy data as block shades; default by input shape; let the host set it. | D-34, D-35 |
| **Tiles** | | |
| FR-21 | Treat the tile source as a replaceable part, with network (TileJSON or URL prefix), embedded, disk-cache (no expiry, stable documented layout) and PMTiles (file or range request) sources. | D-18, D-21, L-12 |
| FR-22 | Ship a default fetcher with a timeout, an identifying User-Agent and status checking — and, for range reads, insistence on a 206 with a matching range; accept a host's replacement. | D-15, L-2, AI-9 §8 |
| FR-23 ◆ | Render performs no input or output of any kind (tested with a transport that blocks forever). With any ancestor tile on hand the frame is non-empty: a stand-in is drawn and sharpens as tiles arrive. **With no tile from any source**, the frame still shows overlays, markers, a plain ground, and a one-line notice naming the assets package — never an empty rectangle. Fetches run in parallel, are de-duplicated, and a tile that failed is retried with back-off (30 s doubling to 10 min), never on every render. | D-30, S7-1, N-2, P-2, P-5 |
| FR-28 | Provide an opt-in assets package with zoom 0–3 tiles, place-name translations stripped, generated by a committed script from a pinned planet file, with its own data notice; never let embedded tiles override a chosen source. | D-27, D-33, L-13 |
| **Embedding and control** | | |
| FR-4 | Render a complete frame on demand into a given rectangle with no terminal attached and no TUI framework loaded. | D-13, M5 |
| FR-24 | Expose control as intents — pan, pan by cells, zoom, zoom around a point, re-centre, fit world — and never read keys or the mouse itself. | D-17, AI-4 §7 |
| FR-25 ◆ | Take animation time from the host: advance by the clock, run no timers of its own, report **when the next visible change is due** (a deadline, not a yes/no — a 150 ms flash sampled by a 300 ms clock would otherwise never be seen to flash), and report whether the visible state changed so the change counter moves. | L-15, AI-4 §9, CQ-5 |
| FR-26 | Reproduce upstream's markers (shapes, blink, flash, pulse) and camera (easing, zoom arc, tours), with durations converted from upstream's ticks. | P-58..P-67 |
| FR-27 | Support several independent map instances in one process. | upstream, M5 |
| **Standalone app** | | |
| FR-5 | Match upstream's key bindings and footer; add drag-to-pan and zoom-toward-the-pointer with a keyboard equivalent for each; offer a one-shot headless flag; always restore the terminal, including on error. | D-13, D-17, L-16 |

## Non-functional requirements

| ID | Requirement | From |
|---|---|---|
| NFR-1 | Pure Go: the library, the app and every dependency build with CGO off, enforced by a cross-compile for macOS (arm64, amd64), Linux (amd64, arm64) and Windows (amd64). | D-22 |
| NFR-2 | Declare a Go version no newer than the first host's (1.25). | T-B, D-19 |
| NFR-3 ◆ | **Target (D-29): 8 MB.** Measured reproducibly: against a **pinned fixture** (named tiles, a named feature overlay with its vertex count, a named image or field), the growth in live heap after a forced collection, read from the runtime's own metrics, and the peak heap, at the default collector setting. The host's resident-memory protocol is confirmatory only, because its own run-to-run spread (8.5 MB) exceeds the figure being measured. Every cache — memory **and disk** — is capped by bytes with host-set caps. **The fixture (D-48):** one regional view; an alert overlay of 10 zone shapes, about 50,000 vertices; one single-frame radar image; one temperature grid; committed to the repository. **Pass:** live heap added ≤ 4 MB, peak ≤ 8 MB. **Worst case, tested separately:** a 58-zone, 812,058-vertex alert is accepted and drawn without the library copying the source geometry, its simplified form fits the shape cache's cap, and growth is bounded. Validated or revised at PLAN exit by measurement. | D-29, D-48, CQ-2, P-1 |
| NFR-4 ◆ | An unchanged frame makes **zero** allocations. Over a one-hour soak the live heap after collection grows by no more than 1 KB a minute (least-squares slope) and the count of goroutines at idle does not change. **Target, to validate at PLAN exit:** a changed frame at 149×38 with tiles warm allocates a bounded amount that does not grow with the number of features. | D-29, CQ-14, P-3 |
| NFR-5 ◆ | **Target (D-30):** the placed view within 1 s warm; within 3 s cold **to full detail**; a non-blank frame from the first render call in both cases. Clocked from the host handing in its data. Measured against a local test server shaped to a **stated link** — proposed 150 ms per request and 8 Mbit/s in aggregate, no more than 6 connections — serving a pinned fixture (the four measured zoom-5 tiles, 1.2 MB). Validated or revised at PLAN exit. | D-30, BQ-5, P-5 |
| NFR-6 ◆ | Render is a pure function of the view, size, colour depth, palette, the time supplied, the overlays and the **ordered set of tiles on hand**. Tiles are visited in a fixed order; nothing depends on map iteration order or on which fetch finished first. Reference frames are byte-identical on amd64 and arm64: every float-to-integer step is an explicit conversion, and non-finite values never reach one. | AI-4 §9, CQ-11 |
| NFR-7 | macOS, Linux and Windows; UTF-8; usable from a 69×12-cell rectangle; safe over SSH, with nothing that depends on querying the terminal. | D-23 |
| NFR-8 ◆ | Every line of a frame is exactly the requested width under one pinned width table — the same one the first host measures with. Every character the renderer itself emits comes from a closed list checked against the Unicode width data: braille is neutral width and safe; block, box-drawing, shade and arrow characters are *ambiguous* width and are verified against a matrix of supported terminals **at PLAN entry**, before any design depends on them. Wide characters in place names occupy two cells. | AI-3 §6, S2-2, S3-3, CQ-6, A-7 |
| NFR-9 ◆ | Dependencies are an explicit allow-list; each carries a licence file. A vulnerability scan of the module **and of the standard library at the toolchain in use** gates every change and runs weekly; releases are built with the latest patch toolchain, scanned as binaries, and ship with SHA-256 checksums. No `replace` directive in the module file. *The allow-list's contents are a PLAN decision.* | D-13, X-5, S-7 |
| NFR-10 ◆ | **Everything that arrives from outside is untrusted**: compressed bodies, vector tiles, TileJSON, PMTiles headers, directories and metadata, style files, images, and feature input. Limits are enforced **before allocating** (defaults host-settable; derived from measured maxima of 722 KB compressed and 1.56 MB decompressed): tile body ≤ 2 MiB, decompressed ≤ 8 MiB; PMTiles directory ≤ 1 MiB decompressed, entry count no greater than the bytes remaining, leaf depth ≤ 3 with a cycle guard, every offset and length checked for overflow and against the file size; metadata, TileJSON and style ≤ 1 MiB, nesting ≤ 64, reference and constant cycles detected; any count ≤ the bytes remaining; tile extent 1..65,536; compression other than none or gzip, and any non-vector tile type, returns a clear "unsupported" error; image dimensions read before decoding and refused over a pixel cap. Every decoder is fuzzed — at least 60 s per target on every change and an hour before a release — never panics, and crashers become permanent test seeds. | RS-12, S-2, CQ-3 |
| NFR-11 | Automated tests never contact the public tile server. | D-30 |
| NFR-12 | The library never bulk-downloads or prefetches from the public tile server. | D-21 |
| NFR-13 | MIT licence; both upstream notices carried; OpenMapTiles and OpenStreetMap credited; the embedded data carries its own notice. | D-13, D-25, D-33 |
| NFR-14 | Sole-author commits: no tool-generated trailers or watermarks in commits, PRs, code or shipped artifacts. Local development tooling stays untracked. Checked on every commit and at every phase exit. | C-2, D-7 |
| NFR-15 ◆ | Accessibility: nothing reachable by pointer only, shown by a scripted keys-only session that reaches every state a pointer session reaches; legible with no colour, shown by a reviewer answering the M1 questions from the no-colour reference frames alone; with no depth hint from the host, a non-empty `NO_COLOR` selects the no-colour depth. A non-visual description of the view is FR-29 (D-52). *Ramp safety and motion are Q12–Q13.* | D-17, D-23, D-35, A-6, A-7 |
| NFR-16 ◆ | Tests first for all Go code, run with the race detector; the code-quality gate is declared and green from BUILD entry; builds use the floor toolchain (Go 1.25) with read-only modules, so an API newer than the first host's cannot slip in. | FULL TDD, D-20, CQ-12 |

## Added after red-team round 1 ✚

| ID | Requirement | From |
|---|---|---|
| **Asynchronous work** | | |
| FR-30 ✚ | **Who runs background work is explicit, bounded and leak-free.** Whatever model PLAN chooses (PLAN proposes at least two: host-scheduled blocking calls; a small library-owned pool), it meets all of these: every goroutine has a stated owner and none outlives the map instance (leak-tested); in-flight fetches and the simplification backlog are capped, newest view wins; every job takes a context and is cancelled when its tile or shape leaves the view; one render may run while one fetch runs, shown under the race detector; completion moves the change counter and keeps the next-change deadline (FR-25) current so an idle host still redraws; an optional callback lets a host be told rather than poll; a *settle* call waits for pending work, which is how the headless render (FR-4) produces a complete frame without breaking FR-23. No panic escapes a public call or a library goroutine. | CQ-1, P-2, S-3 |
| FR-31 ✚ | The tile cache never serves a tile parsed for a different style profile or label language: either the key includes them, or tiles are decoded without reference to style. | CQ-8 |
| **Safety of what is shown** | | |
| FR-37 ✚ | An image overlay may carry a **sequence of timed frames**, played on the host's clock (FR-25), with the frame count held inside the image cache's byte cap. *In v1; designed with the contract in PLAN; built after v0.1.0 (D-47). v0.1.0 shows the latest frame.* | D-47, PM-6 |
| FR-29 ✚ | **Describe the view, as data, for each place the host names**: inside or outside each area feature, with distance and compass bearing to the nearest edge; the field value, its band, and which way it rises; the image intensity class here and the nearest heavier intensity with distance and bearing; each overlay's valid time. Computed from the unsimplified shapes and the data, never from the drawn cells. Deterministic; no braille, block or box-drawing characters; speakable — values with units, compass words. The host renders or speaks it; the app prints it with a flag. Acceptance: with the specimen-4 data the description says the place is inside the flood warning; every M1 scenario's description matches the answer key (D-43). | D-52, A-1, PM-7 |
| FR-32 ✚ | **Every overlay carries a valid time, and the map can show when data is old.** The host states when each overlay's data was valid and how long it stays current; the library exposes that beside the legend data (FR-13) and can mark an overlay stale on the frame. Weather freshness is the one kind that matters (D-18). | PM-6 |
| FR-33 ✚ | **A distance reference.** The library exposes the ground distance a cell spans at the view's centre, and can draw a scale mark, so "roughly how far" (M1) has an answer. | PM-2 |
| FR-34 ✚ | Text that reaches a frame is made safe **at the point it is written to a cell**, whatever its source — tile names, source metadata, style text, host labels, credits, legend and error text: invalid UTF-8 becomes U+FFFD; control characters (C0, DEL, C1), line and paragraph separators, bidirectional controls and zero-width characters are dropped; one printable character per cell. The library emits colour sequences and nothing else. Fuzzed so that no input produces an escape byte outside the library's own colour sequences. | S-1, CQ-3 |
| FR-18a ✚ | With no colour, **feature overlays** remain distinguishable from the basemap and from each other by non-colour means: a distinct stroke family plus interior hatch or an edge label for polygons; a glyph or label difference for severity and for wind-speed class. Acceptance: from the 69×12 no-colour reference frame alone, a reviewer says whether the marker is inside the polygon and ranks two wind speeds. *Specimen owed.* | A-2 |
| FR-36 ✚ | Basemap layers — roads at least, and labels as upstream already allows — can be switched off and on by intent, in either renderer. | D-42 |
| FR-24a ✚ | Intents include focus-next, focus-previous and zoom-around-the-focused-target. The focused target carries a non-colour indicator, and its label and id are exposed as data. *The mechanism remains a PLAN decision (D-17).* | A-6 |
| **Tiles and the network** | | |
| FR-21a ✚ | The disk cache has a host-set byte cap (proposed default 256 MB) and evicts by least-recently-read, never by age — consistent with "no expiry" (D-18). Pruning runs off the drawing path. | P-6, CQ-3 |
| FR-21b ✚ | Cache paths are built only from a hash of the source identity and integers the library formats itself; all cache file access is confined to the cache root; directories are private to the user, files likewise, written by temporary file and rename; a cache root writable by others is refused. | S-5 |
| FR-22a ✚ | A tile is written to the cache **only after** a success status, a length within limits and a complete successful decode. A cached file that fails to decode is deleted and fetched again. The cache key is the full source identity (for a remote archive: its entity tag and length). `Purge` and `Verify` are exposed, and offered as flags in the app. | S-4, L-2 |
| FR-22b ✚ | Secure transport by default; plain `http://` only by explicit option or on the loopback address. At most 3 redirects, never from secure to plain, never into loopback, link-local or private address space unless the configured source is already there; the same for tile addresses read from TileJSON. Every response body is read through a limit. A range reply that is not a 206 with exactly the requested range is closed unread. The replacement fetcher's contract carries the range and the maximum length, and the library checks what comes back. **No connection is ever made unless the host configured a network source.** The User-Agent names the library and version and a host-supplied token, never the operating system, host name or user. Proxy settings from the environment are honoured. | S-6, AI-9 §8 |
| FR-28a ✚ | The embedded tiles pin the planet file by version, length and entity tag; a SHA-256 list of every source tile and every output tile is committed; continuous integration re-hashes the asset and decodes every embedded tile through the hardened decoder. The generator is written in Go and lives in the repository. | S-7, BQ-6 |
| FR-35 ✚ | **The tile schema is a seam.** Layer and attribute names a style depends on are resolved through one replaceable mapping keyed on the source's declared layers, so a change of schema by the provider is a new mapping, not a new renderer. | PM-5, AI-2 §5 |
| **Newcomers and host mistakes** | | |
| NFR-17 ✚ | A runnable, continuously tested example for each input shape lives under `examples/`, outside the library's contract — including radar from one keyless source **with its complete, tested colour-to-intensity table**, which a newcomer can copy. | N-1, D-15, D-36, D-45 |
| NFR-18 ✚ | The README quick-start, pasted unchanged into an empty module with the network disabled, builds and renders a world map. Continuous integration runs exactly that. | N-2, D-27 |
| NFR-19 ✚ | A map in a rectangle takes no more than 3 calls, and a first overlay one more, with everything else defaulted — including a ramp per colour depth. The example test compiles with exactly those calls. | N-4 |
| NFR-20 ✚ | Every value a host hands in is validated on hand-in: non-finite or out-of-range coordinates, grid sizes that do not match their data, unsorted class breaks, a missing table, a ramp that fails FR-16, an overlay over its size cap (proposed 2,000,000 vertices, host-settable). A failure returns a typed error saying what happened, why, and what to do. Non-fatal problems are retrievable as a list and through an optional callback. The library never writes to standard output or standard error. | N-5, S-3 |

## Release slices (D-44)

Every requirement in this file is in **v1**. This table says which are in the **first release, v0.1.0**, after which go-tuiMaps is integrated into the first host at its v0.17.0 as the first in-app use and test, and work then returns here for the rest.

| In v0.1.0 | Requirements |
|---|---|
| Basemap, braille renderer | FR-1 (the parity rows this scope touches), FR-2 (waterways and parks; airports later), FR-3 (braille; the block renderer later), FR-19, FR-20 (own styles; a user's style with legacy filters — expressions later), FR-35, FR-36 |
| Overlays — **three shapes** | FR-6 features, FR-7 scalar grids, FR-9 images; FR-11, FR-12, FR-13, FR-14, **FR-29 (the view described as data)**, FR-32, FR-33, FR-34 |
| Colour and theme | FR-15, FR-16, FR-17 (truecolor, 256 colours, no colour — 16 colours later), FR-18, FR-18a |
| Tiles | FR-21 (network, embedded, disk cache — PMTiles later), FR-21a, FR-21b, FR-22, FR-22a, FR-22b, FR-23, FR-28, FR-28a, FR-31 |
| Embedding and control | FR-4, FR-24 (pan, zoom, re-centre, fit world), FR-25, FR-26 (static and blinking markers), FR-27, FR-30 |
| App | FR-5 (keys for pan, zoom and toggles; headless flag; terminal always restored) |
| Non-functional | **All of NFR-1 to NFR-20.** No security, safety or accessibility requirement is deferred. |

| Deferred past v0.1.0 — still v1 | Requirements |
|---|---|
| Wind | FR-8 |
| Radar and other image loops | FR-37 |
| Tile-image providers | FR-10 |
| Local and remote single-file archives | the PMTiles part of FR-21 |
| Block renderer | the block part of FR-3 (opt-in, D-42) |
| 16-colour depth | that part of FR-17 |
| Camera tours; flash and pulse | those parts of FR-26 |
| Pointer operations and focus-zoom | those parts of FR-24 and FR-5; FR-24a |
| Airport layers; style expressions | those parts of FR-2 and FR-20 |

**Safeguard for PLAN:** the contract is designed for all five overlay shapes and built for three, so the deferred shapes constrain the design.

## Constraints and dependencies

| ID | Constraint or dependency | Note |
|---|---|---|
| CD-1 ◆ | **The first host must change before it can feed four of the five overlay shapes.** Today it holds points, circles and one wind reading per place. *Features:* it must keep alert geometry, and for the 90% of alerts with none, fetch zone shapes. *Images (radar — the most common view, D-35):* it must **add** a public radar source, which HUM LEAD has confirmed it can (D-39), and a colour table if one is required (Q4). *Scalar and vector grids:* it must add a gridded source; a lattice of forecast points is not usable for precipitation (S11-1), and model grids arrive in a format with no mature pure-Go decoder. *Tile-image providers:* as images. | Outside this repository; a named companion backlog for the host (RS-11, now High). |
| CD-2 | The default network source is a free service with no guarantee and a stated intent to change schema. | Mitigated by FR-21; tracked in PLAN (D-21). |
| CD-3 | The embedded tiles are generated from a published planet file that the provider's README does not document. | Pin the version and checksum; regenerate deliberately (AI-9 §8). |
| CD-4 | No remote exists until SHIP; the first host can consume the library only through a local override until then. | D-19. |
| CD-5 | Git: `main ← release/v0.1.0 ← feature/*`; `main` by pull request only; every merge by pull request from the first release. | D-28. |

## Assumptions to validate

| ID | Assumption | Validated by |
|---|---|---|
| A-1 | 8 MB is achievable for the fixture HUM LEAD rules in Q7 — a **single-frame** image in v0.1.0 (D-47); a loop's frames must fit the image cache's byte cap when built. | A measured prototype at PLAN exit (D-29). |
| A-2 | Every non-braille character the renderer emits — quadrant blocks, box-drawing, shades, arrows — is one column wide in the supported terminals. | A terminal matrix check **at PLAN entry** (NFR-8). |
| A-3 | The rendering rules hold on a light-background terminal. | A specimen in PLAN; untested so far. |
| A-4 | The standard HTTP client returns 206 reliably for range reads of a very large file. | A test in BUILD; one anomalous 200 was seen from another client (AI-9 §8). |
| A-5 | A proper contouring pass yields clean lines at terminal resolution. | A specimen in PLAN. |
| A-6 | Simplifying tens of large polygons off the drawing path fits the host's budgets. | Measurement in BUILD. |
