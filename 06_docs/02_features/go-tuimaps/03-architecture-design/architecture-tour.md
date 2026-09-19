# A guided tour of the architecture

Up: [architecture](architecture.md)

The diagram set is wide — twenty-four diagrams. This page is the other way in: **one story, followed end to end**, that passes through almost every diagram once. Read it with the diagrams open beside it. Each step says which diagram you are standing in and which ruling put that step there.

**The story:** it is a stormy evening. Watchpost is showing Fort Wayne. A flood warning is in force and rain is moving in from the south-west. Follow the warning and the radar picture from Watchpost to the cells on your screen — and to the sentence Watchpost can speak.

## Part 1 — Before anything is drawn

| # | What happens | Where to look | Why it is this way |
|---|---|---|---|
| 1 | Watchpost creates a map and tells it where tiles may come from. Until it does, the library will not touch the network. | L0 Context — the line from *Map* to *Tile service* is labelled "only if the host names one" | D-65 |
| 2 | Watchpost hands over its theme as a palette of named colour jobs — "heavy rain", "severe alert outline", "ground". | L2 Colour — step 2, *Semantic tokens* | D-63 |
| 3 | Watchpost — not the library — fetches tonight's weather: the warning's outline, one radar picture for the area, a temperature grid. | L0 Context — *Host's weather fetchers* | D-15 |
| 4 | It hands each one in as a plain struct: `Set(alerts)`, `Set(radar)`, `Set(temperature)`. It names a preset for each, so it supplies no colours or breaks at all. | L1 Public contract — *What is on it* · L2 Overlays — *Presets and host-defined types* | D-74, D-69 |
| 5 | Each hand-in is checked at the door. A missing colour table for the radar picture would be refused here, with a message saying what to do. | L2 Overlays — *One path for every overlay*, the first diamond | NFR-20, D-45 |
| 6 | The warning's outline is **borrowed**: the library reads Watchpost's copy and never makes its own. Watchpost must leave it alone until told the borrow is over. | L3 States — *Borrowed geometry* · L2 Memory — *The host's memory* | FR-11, D-48 |

## Part 2 — The first frame, which is never blank

| # | What happens | Where to look | Why |
|---|---|---|---|
| 7 | Watchpost's 300 ms clock ticks and it calls `Render`. Nothing has been decoded yet, so the frame shows the ground, the place marker and a one-line notice — not an empty box. | L3 Sequences — *A cold first frame*, steps 4–6 | D-30, FR-23 |
| 8 | `Render` never waits and never fetches. It only **notes what is missing** — tiles, and the three overlays still to be prepared — as pending work. | L2 Render — *From a call to a frame*, the "Note what is missing" box | FR-23 |
| 9 | Nothing happens to that pending work until **Watchpost's own goroutines** call `Work`. The library has none. | L3 Sequences — the *Host's pump* lane · L0 Context — *Host's pump* | D-73 |

## Part 3 — The slow work, done on Watchpost's time

| # | What happens | Where to look | Why |
|---|---|---|---|
| 10 | A `Work` call picks up a tile. It tries the disk cache, then the named service, then the tiles shipped inside the program. | L2 Tiles — *Where a tile can come from* | FR-21, D-27 |
| 11 | The network request leaves through one guarded door: secure transport, at most three redirects, never into a private address, and no error message ever prints the address. | L2 Tiles — *The network edge* | FR-22b |
| 12 | The bytes that come back are treated as hostile. Sizes and counts are checked **before** memory is set aside; layers the map does not draw and the place-name translations are thrown away during decoding. | L2 Tiles — *Untrusted-input gate* | NFR-10, D-75 |
| 13 | If the tile fails, it is given a "not before" time — 30 seconds, doubling. No timer is set; the time simply becomes part of the answer to "when should I call you next?" | L3 States — *A tile* · L3 Sequences — *An idle host* | FR-23, FR-25 |
| 14 | Another `Work` call prepares the radar picture: each pixel's colour is looked up in the provider's table and kept as **one byte — how heavy** — and the provider's own colours are thrown away. | L2 Overlays — *Image*, steps I3–I4 | D-36, D-45 |
| 15 | Another prepares the warning: the borrowed outline is simplified to what a braille dot can show at this zoom. | L2 Overlays — *Features*, step F2 | D-16 |
| 16 | Each finished job bumps a counter. Watchpost's pump tells its interface "the map changed". | L3 Sequences — *A cold first frame*, steps 12–13 | FR-25 |

## Part 4 — Painting the cells

| # | What happens | Where to look | Why |
|---|---|---|---|
| 17 | `Render` runs again and paints in a fixed order: ground, water, the radar, the warning's tint, the map's braille lines, the warning's outline, the marker, labels, then the scale mark and credit. | L2 Render — *Paint the cell grid*, 1 to 9 | FR-12 |
| 18 | Each radar cell's "how heavy" byte becomes a colour: preset → named job → Watchpost's theme if it set one → checked → fitted to the terminal's colour depth. With "safe ramps" on, the theme is skipped. | L2 Colour — *How a value becomes a cell colour* | D-69, D-63, D-53 |
| 19 | For every cell the map line is drawn black or white — whichever contrasts more with what is behind it. | L2 Colour — step 7 | FR-16 |
| 20 | Roads thin out so the weather can be read. | L2 Render — *Choose the basemap profile* | FR-19 |
| 21 | If the terminal had no colour, the radar would become ░▒▓ shades, the temperature contour lines, the warning a hatched area with the word FLOOD. | L2 Colour — step 5, *No colour* | D-35, FR-18a |

## Part 5 — The same facts, as words

| # | What happens | Where to look | Why |
|---|---|---|---|
| 22 | Watchpost asks `Describe(Fort Wayne)`. The answer is worked out from the **real outline**, not the simplified one and not the drawn cells: "inside the Flood Warning; nearest edge 3 km east; heavy rain 38 km south-west." | L2 Describe — *How it is computed* | D-52 |
| 23 | That is why the picture and the words cannot disagree about which side of a line you are on — one outline, one rule for "inside", two uses. | L2 Describe — *Picture and description cannot disagree* | FR-11, FR-29 |
| 24 | At the smallest map size a column is 1.6 km wide. If Fort Wayne sits closer to the edge than that, the honest reading of the picture is "on the edge" — and the words give the exact answer. Both are tested. | L2 Describe — the *Answer key* arrows | D-67 |
| 25 | Every string is cleaned on its way out, so a hostile place name in a tile can never send control codes to your terminal or your speech engine. | L0 Context — *Way out* | FR-34 |

## Part 6 — Afterwards

| # | What happens | Where to look | Why |
|---|---|---|---|
| 26 | Watchpost asks "when should I call you next?" The answer is the soonest of: the marker's next blink, a failed tile's retry time, the moment the radar goes stale. | L3 Sequences — *An idle host* | FR-25 |
| 27 | Twenty minutes later the radar has not been refreshed. It is marked stale on the map, in the legend and in the description. | L3 States — *An overlay's freshness* | FR-32 |
| 28 | Watchpost sets a new radar picture under the same id. The old one keeps drawing until the new one is ready. | L2 Overlays — *Same id already set?* | FR-11, D-74 |
| 29 | All of this had to fit in about 4 MB of lasting memory. The map of where those bytes go — and the one place the numbers are tight — is its own diagram. | L2 Memory — *Where the bytes live* | D-29, D-48 |

## If you only look at three diagrams

1. **L0 Context** — who does what, and what is trusted.
2. **L3 Sequences, "A cold first frame"** — how a frame comes to exist when the library has no goroutines of its own.
3. **L2 Colour, "How a value becomes a cell colour"** — where presets, themes, safe ramps and colour depth meet.
