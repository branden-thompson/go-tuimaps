# go-tuiMaps — Glossary

One plain sentence per term. Added after the DISCOVER red-team found that the documents' first audience — product designers — had no way to decode them.

## Map and data terms

| Term | Meaning |
|---|---|
| **Basemap** | The map itself — land, water, borders, roads, place names — before anything is drawn on top. |
| **Overlay** | Anything a host application draws on top of the basemap: an alert area, a temperature field, wind arrows, radar. |
| **Host** | The application that embeds the map. The first host is Watchpost, a terminal weather station. |
| **Cell** | One character position in a terminal. A cell holds one character, one foreground colour and one background colour. |
| **Braille renderer** | Draws with Unicode braille characters: each cell is a 2×4 grid of dots, so an 80×24 terminal becomes a 160×96-dot picture. |
| **Block renderer** | Draws with quarter-block characters (▘▝▖▗ and their combinations): 2×2 per cell. Coarser. Opt-in, and built after the first release (D-42, D-44). *Whether it is present in more fonts than braille is assumed, not shown; the glyph matrix at PLAN entry tests it.* |
| **Tile** | One square piece of the world map at one zoom level. The map is assembled from the few tiles a view needs. |
| **Vector tile** | A tile that carries shapes and names (roads as lines, lakes as outlines) rather than a picture, so it can be drawn at any size and style. |
| **Zoom level** | How far in the map is: 0 is the whole world in one tile; each level doubles the detail; 14 is the most detailed the tile source provides. |
| **OpenMapTiles** | A published naming scheme for what is inside a vector tile — which layer holds roads, what a lake is called. A *schema*. |
| **OpenFreeMap** | The free public service the standalone app fetches tiles from by default, and the one the examples show. The library itself connects to nothing until a host names a source (D-65). It serves OpenStreetMap data in the OpenMapTiles scheme. |
| **TileJSON** | A small description file a tile server publishes, saying where its tiles are and what they contain. |
| **PMTiles** | A single-file archive holding a whole region's (or the whole planet's) tiles, readable from disk or piece-by-piece over the web. |
| **Range request / "206"** | Asking a web server for just a slice of a large file. 206 is the reply code meaning "here is the slice you asked for". A reader must insist on it, or it could be handed the whole multi-gigabyte file. |
| **Embedded tiles** | A small set of low-zoom tiles compiled into the program so the map shows something with no network at all. |
| **Stand-in tile** | A coarser tile stretched to cover an area while the detailed one is still loading, so the map is never blank. |
| **Style** | The rules that say what to draw and in what colour: "motorways in light grey from zoom 5". |
| **Zoom stop** | A point in a style rule where a value changes with zoom: "this road is 1 wide until zoom 10, then 2". |
| **Style profile** | One of a small family of styles the map switches between — sparser under a wind overlay, sparser still in the block renderer. |

## Overlay terms

| Term | Meaning |
|---|---|
| **Feature overlay** | Points, lines, areas and circles given as latitude and longitude: an alert area, a storm track, an earthquake. |
| **Scalar grid** | A regular grid of single numbers laid over the map — temperature, pressure, rainfall amount. |
| **Vector grid** | A grid where each point has a direction and a strength — wind. |
| **Georeferenced image** | A picture with known map corners, such as a radar image, so it can be placed exactly. |
| **Tile-image provider** | A function the host supplies that returns a picture for any map tile asked for — radar served as tiles. |
| **Ramp** | The ordered set of colours used to show a range of values: cool blue through to hot red. |
| **Class breaks** | The values at which a ramp steps from one colour to the next: 14°, 16°, 18°… |
| **Colour-to-intensity table** | A lookup that turns a radar provider's colours back into "light, moderate, heavy", so the map can redraw them in its own or the host's colours. |
| **Legend data** | The class breaks and their colours, handed to the host so it can draw a key wherever it likes. |
| **Credits / attribution** | The "© OpenStreetMap" line map data legally requires, plus any credit an overlay's source asks for. |
| **Valid time** | The moment an overlay's data describes. A radar frame from forty minutes ago must not look current. |
| **Dot tolerance** | The smallest detail a view can show. A 14,000-point county outline is simplified to this before drawing. |
| **Meteorological "from" direction** | Wind direction is reported as where it blows *from*; arrows on the map point where it is *going*. |
| **Antimeridian** | The 180° line in the Pacific where longitude wraps from +180 to −180. |

## Rendering and colour terms

| Term | Meaning |
|---|---|
| **Colour depth** | How many colours a terminal can show: truecolor (millions), 256, 16, or none. |
| **`NO_COLOR`** | A widely honoured setting by which a user asks programs not to use colour. |
| **Contrast ratio** | A standard measure (from the web accessibility guidelines, WCAG) of how readable one colour is on another. 4.5:1 is the bar for text, 3:1 for graphics. |
| **Relative luminance** | How bright a colour appears. A ramp whose brightness rises steadily still reads in order to someone who cannot tell its hues apart. |
| **Colour-vision deficiency** | Colour blindness. The common forms confuse red with green. |
| **Contour line** | A line joining points of equal value, labelled with that value — how a printed weather map shows temperature without colour. |
| **Block shades** | The characters ░ ▒ ▓ █, used to show light-to-heavy radar with no colour. |
| **Ambiguous width** | Some characters are one column wide in most terminals but two in some East Asian settings. Braille is safe; block, box-drawing and arrow characters must be checked. |
| **Escape sequence** | Special character codes a terminal obeys rather than displays — change colour, move the cursor, even write to the clipboard. Text from outside must never be allowed to carry them into a frame. |

## Engineering terms

| Term | Meaning |
|---|---|
| **Headless render** | Producing a map frame as text with no terminal attached — for tests, scripts and one-shot output. |
| **Intent** | A request in the map's own terms — "pan", "zoom around this point" — that the host calls when *it* decides a key or mouse event means that. The map never reads the keyboard itself. |
| **Change counter** | A number that goes up whenever the map's picture would change, so the host knows when to redraw. |
| **Goroutine** | Go's lightweight unit of background work. Who starts them, and who stops them, must be stated. |
| **CGO** | Go code that links C code. Forbidden here: it breaks building for every platform from one machine. |
| **Fuzzing** | Feeding a decoder enormous numbers of mangled inputs to find any that crash it. |
| **Reference frame / golden file** | A saved, approved rendering that tests compare new output against, character for character. |
| **Resident memory** | How much memory the operating system says a program is using. Noisy; the requirements measure live heap instead. |
| **Parity** | Behaving as the original Rust program, TerminalMap, does. |
| **M3 denominator** | The count of TerminalMap behaviours go-tuiMaps must demonstrate. Frozen so it cannot be shrunk to flatter the result. |

## Process terms used in these documents

| Term | Meaning |
|---|---|
| **HUM LEAD** | The human lead: the project's owner, who approves every decision at this project's risk level. |
| **SEV-0** | The highest rigor level: every decision is the human lead's. |
| **DISCOVER (RCC) · PLAN · BUILD · REVIEW · VALIDATE · SHIP · REFLECT** | The project's phases, in order: requirements; design; implementation; quality review; checking it works where it will run; release; lessons learned. |
| **Ruling (D-n)** | A recorded decision by the human lead, with his words quoted. |
| **Red-team** | A deliberate adversarial review at the end of each phase, by reviewers told to find what is wrong. |
| **Specimen** | A real rendering produced to judge a visual idea. Nothing visual is approved from a description. |
| **Defect ledger (L-n)** | The closed list of known bugs in the original program, each marked *fix* or *replicate*. |
| **Upstream** | The project being ported: TerminalMap, and behind it MAPSCII. |
| **Palette · theme** | A palette is the set of colours a host hands the map; a theme is the host's name for one such set. |
| **Fixture** | A fixed, committed set of test data, so a measurement can be repeated. |
| **Seam** | A place where one part can be swapped for another without touching the rest. |
| **Zoom bucket** | A band of zoom levels that share one simplified copy of a shape. |
| **Diverging ramp** | A colour ramp with a meaningful middle — pale at the midpoint, one hue below it and another above. |
| **Protanopia · deuteranopia · tritanopia** | The three common kinds of colour blindness: weak red, weak green, weak blue. |
| **Reduce-motion** | A setting for people whom animation harms or distracts: nothing blinks, nothing glides. |
| **Hatch** | Diagonal strokes filling an area, used where colour is not available. |
| **Borrowed geometry** | Shapes the host hands in and keeps owning. The library reads them and does not copy them; the host must not change them while they are on the map. |
| **Settle** | A call that waits until background work for the current view is finished, so a one-shot render is complete. |
| **Back-off** | Waiting longer after each failure before trying again. |
| **Grapheme cluster** | What a reader sees as one character, even when it is stored as several — a letter with its accent, a flag. |
| **Answer key** | For each M1 scenario, the correct answers — inside or outside, bearing, distance — computed by a separate script. |
| **Planet file** | One very large file holding every tile for the whole world. |
| **Race detector** | A Go tool that finds two parts of a program touching the same data at once. |
| **Discovery Report** | The document that closes this phase, for HUM LEAD's approval. |
| **Kind (of a grid)** | What a grid of numbers measures, from a short list the library owns: *generic*, or *temperature* with its unit. The kind decides who chooses the colours. |
| **Absolute scale** | A colour scale pinned to fixed values, so one colour means one temperature on every map, every day (D-62). |
| **Semantic token** | A stable name for a colour's job — "heavy rain", "severe alert outline", "ground" — so a host can re-theme by meaning without knowing how the map is drawn (D-63). |
| **Ground** | The colour painted behind the whole map, so it reads the same on a dark or a light terminal (D-64). |
| **Safe ramps** | A setting that swaps a theme's overlay colours for the library's own colour-blind-safe ones (D-63). |
| **ID prefixes** | R- brief requirement · FR- / NFR- functional / non-functional requirement · CD- constraint or dependency · A- assumption (A-1..A-6, at the foot of the requirements) · RS- risk · OQ- open question · AI- research report · S*n*-*n* specimen finding · P-01..P-72 parity row, five of them split into *a* and *b* halves (D-61) · E- extension row · L- ledger entry · X- contradiction between research reports · M- metric · D- ruling · Q- question for the human lead (round 1) · R2-Q- the same (round 2). **Red-team finding codes** reuse some letters: in a requirement's "From" column and in the red-team record, bare A-, P-, S-, CQ-, PM-, BQ-, N-, DQ-, PH- and C- codes are round-1 findings (accessibility, performance, security, code quality, phase lens, business, newcomer, docs, hygiene, convergence); R2- codes are round 2's. The requirements file says which is which at its head. |
