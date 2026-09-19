# PLAN — Entry checks

| Field | Value |
|---|---|
| Phase | PLAN (entry) |
| Date | 2026-09-19 |
| Why this exists | The Discovery Report named two checks that gate the rest of PLAN: a confirmed way to get radar into the first release (risk RS-23), and a terminal matrix for glyph width and braille coverage (risks RS-25 and RS-13, requirement NFR-8, ruling D-57). |

## 1. A radar image for one bounding box, from a keyless source — CONFIRMED

The first release accepts radar as **one image for one bounding box** (FR-9); fetching radar as tiles is deferred (D-44). Every specimen so far had used tiles, so this path was undemonstrated.

| | Iowa Environmental Mesonet (Iowa State University) | NOAA / NCEP |
|---|---|---|
| Request | A standard web-map-service "GetMap" request: layer `nexrad-n0q-900913`, a bounding box, a width and height, PNG | The same, layer `conus_bref_qcd` |
| Key or account | None | None |
| Tried 2026-09-19 | Bounding box −90,31 to −83,36, 512×366, in longitude/latitude **and** in web-mercator: both returned a PNG in about 0.3 s | The same box: a PNG in about 0.3 s. *One attempt on 2026-09-18 met "service unavailable"; it was transient.* |
| What came back | 8-bit RGBA; **no partly transparent pixels** — every pixel is fully clear or fully opaque; 84 distinct visible colours in a mostly dry scene | 8-bit RGBA; no partly transparent pixels; 44 distinct visible colours |
| **Colour table** | **Published and exact:** 256 entries, index 0 = missing, then −32 to 95 dBZ in steps of 0.5. Checked against the fetched image: its commonest colour, 96·180·212, is index 95 = 15 dBZ in the published table | Not found published; would have to be derived from the legend |
| Terms, as the provider states them | "The materials found on this website are in the public domain and may be used freely by anyone for any lawful purpose." "Attributing the Iowa Environmental Mesonet of Iowa State University would be appreciated." No rate limit is stated | United States government work |
| Coverage | United States | United States |

**Conclusions**

1. **RS-23 is retired for the first release.** The example (NFR-17) uses the Iowa service with its published table: one request, one image, an exact table a newcomer can copy. Credit is given although only "appreciated" (FR-14).
2. **Hard edges.** Neither service returns partly transparent pixels, so the nearest-colour matching of FR-9 will meet exact colours, and the tolerance can default low.
3. **One requirement was wrong by one.** FR-9 capped a colour table at 255 entries; this real table has 256. Corrected to 256.
4. **Still true:** both sources are US-only (NFR-17 says so), and no rate limit is published — the example fetches on demand, never in a loop, and tests use a recorded image (NFR-11).
5. **Carried into the design:** the image arrives in either projection the host asks for. The contract must let the host say which; resampling to the view is the library's job (the "radar resampling rule" carried from DISCOVER).

## 2. Terminal matrix — ONE TERMINAL PASSES

[`terminal-matrix/test-card.txt`](terminal-matrix/test-card.txt) is a plain-text card. Shown with `cat` in a terminal, each row is a bar, twenty copies of one character, and a closing bar that must land under a marked column. It tests, for every character the first release's renderer can emit (NFR-8's closed list) and those the later renderers need: **width** — does the bar land in column 22 — and **coverage** — is it the character, or an empty box.

| Terminal and font | Braille | Quadrants | Shades | Box | Hatch | Arrows | Markers | U+FFFD | Picture reads | Colour specimens read |
|---|---|---|---|---|---|---|---|---|---|---|
| HUM LEAD's terminal on macOS, 2026-09-19 — application and font not yet stated | pass | pass | pass | pass | pass | pass | pass | pass | pass | not yet reported |

**As reported:** "passed on my terminal" — HUM LEAD ran the card and pasted its output. The result is HUM LEAD's reading of the screen; a paste of text cannot show widths or missing glyphs, so the coordinator has not verified it independently. In the pasted text three sections (box drawing, hatch, arrows) showed some lines joined by long runs of spaces; with the card reported as passing, that is taken to be an artefact of copying from the terminal, and is noted here in case it recurs. **What this does and does not settle:** the closed list of characters (NFR-8) is one column wide and present in one real terminal and font, including braille (risk RS-25) — one data point, not a survey.

This look also serves as the terminal look at the specimens that DISCOVER still owes (risk RS-3). Terminals the first host supports but HUM LEAD does not use are listed in PLAN as untested, not assumed.
