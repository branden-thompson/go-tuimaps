# DISCOVER — The options as they were presented

Companion to [`rulings-discover.md`](rulings-discover.md). Many rulings there record HUM LEAD's answer as a letter. This file records **what each letter stood for when he chose it**, so a ruling can be read without the conversation it came from. Added after the DISCOVER red-team (docs finding DQ-5) found 16 of 27 rulings recorded as a bare letter with the options nowhere in the record.

Each question was presented one at a time with: what was being decided, the evidence, the options and their consequences, a recommendation, and the strongest argument against it. The option HUM LEAD chose is in **bold**. Where the recommendation was *not* what he chose, that is marked ◇.

| Ruling | Question | Options as presented | Recommended |
|---|---|---|---|
| D-11 | What "parity with TerminalMap" means | **A. Behavioural parity against the code**, with a closed defect ledger, each entry fix or replicate · B. Pixel-level parity: character-for-character, defects included · C. API-shaped parity: a Go equivalent of every public method | A |
| D-12 | Layers upstream styles but never draws (rivers, parks, airports) | **A. Draw them, as flagged extensions outside the parity count** · B. Replicate the gap · C. Rivers only | A |
| D-13 | What ships | **A. Library + a small standalone app + a one-shot headless render** · B. Library and headless render only · C. Library only | A |
| D-14 | Overlay kinds in v1 | A. Features + scalar grids + vector grids · B. A plus radar images · C. Features only · **D. All five shapes, including tile-image providers** | A ◇ |
| D-15 | Who fetches overlay data | **A. The host fetches; the library draws**, with a replaceable default basemap fetcher · B. Optional ready-made source packages · C. The library fetches weather itself | A |
| D-16 | Who simplifies and clips large shapes | **A. The library does both, cached, off the drawing path; no bypass in v1** · B. The host does both · C. The library does both, with a host bypass flag | A |
| D-17 | Pointer operations | **A. In the library, and wired to the mouse in the app** · B. In the library only; the app stays scroll-only · C. Not in v1 — *HUM LEAD added: every pointer operation needs a keyboard equivalent* | A |
| D-18 | Local map files as a tile source | A. Not in v1; design the tile source as a replaceable part · B. MBTiles in v1 as an opt-in package (pulls in a large pure-Go SQLite) · **C. PMTiles in v1: a dependency-free reader** | A ◇ |
| D-19 | Module path; when the repository becomes public | Path — **A. `github.com/branden-thompson/go-tuimaps`** · B. `…/tuimaps` · C. a custom domain. Visibility — **1. Local only; decide at SHIP** · 2. Private remote now, public at the first tag · 3. Public from the first push | A, 2 ◇ |
| D-20 | When the code-quality gate turns on | **A. Remove the language declaration now; restore it at BUILD entry** · B. Add an empty package so the gate passes · C. Carry a failing structure check across the phase exit | A |
| D-21 | The default basemap source | **A. Confirm OpenStreetMap data via OpenFreeMap**, replaceable, never bulk-downloaded · B. Something else | A |
| D-22 | Pure Go, no C toolchain | **A. A hard requirement for the library, the app and every dependency** · B. Core only; labelled exceptions for add-on packages | A |
| D-23 | Platform and terminal floor | **A. Match the first host's floor; braille and block both first-class; legible at four colour depths** · A-narrow. 16-colour and no-colour as "doesn't break, may be plain" · B. A plus a 7-bit ASCII renderer · C. Braille first-class, block best-effort — *later refined by D-42: braille is the default, block is opt-in* | A |
| D-24 | Where the map styles come from | **A. Write fresh styles against OpenMapTiles; do not port the renaming shim** · B. Keep upstream's files and add the presumed originator's notice · B-now-A-later · C. Keep them with no notice | A |
| D-25 | How attribution is shown | **A. The library reports credits and draws an optional one-line credit, on by default** · B. Reports only; hosts must display · C. Always drawn, no switch | A |
| D-26 | How a host's theme reaches the map | **A. The host supplies a plain palette; no design-system dependency** · B. Import the first host's design system · C. A, plus an optional adapter later | A |
| D-27 | Where the embedded offline tiles live | **A. A separate opt-in package; depth decided after measurement** · B. Built into the core at zoom 0–1 · C. Separate package, fixed at zoom 0–1 | A |
| D-28 | How changes reach `main` | **A. `main ← release ← feature`; local merges until a remote exists; `main` by pull request only** · B. Every merge is a hosted pull request · C. Decide at SHIP — *HUM LEAD added: move to B at the first release* | A |
| D-29 | The numbers for M4 | **A. A ceiling of 8 MB now, confirmed by measurement at PLAN exit**, with byte-capped caches and a flat heap · B. Structural rules now, the number at PLAN exit · C. A ceiling relative to the host's margin | A |
| D-30 | The numbers for M2 | **A. ≤ 1 s warm, ≤ 3 s cold, plus a never-blank rule** · B. The numbers only · C. Stricter numbers (≤ 300 ms warm, ≤ 2 s cold) | A |
| D-31 | The rendered-specimen spike | **A. All nine specimens, staged, stopping after the first if it fails** · B. Only the four that decide feasibility · C. No spike; specimens from the real build | A |
| D-33 | Depth of the embedded tiles | **A. Zoom 0–3, translations stripped (1.7 MB)** · B. Zoom 0–2, stripped (760 KB) · C. Zoom 0–1 as served (490 KB) · D. Zoom 0–3 as served (16.1 MB) | A |
| D-34 | A scalar field with no colour | A. A sparse lattice of values · B. Density glyphs · **C. Contour lines with value labels — specimen first** · D. A stated limit: fields need colour — *HUM LEAD: C with A as the fall-back* | A for v1, C as a later specimen ◇ |
| D-35 | The no-colour form, after the contour and radar specimens | **By kind of data: contours for smooth fields, block shades for patchy data** · Shades for everything · Contours for fields, and radar reported as needing colour | By kind of data |
| D-36 | How a placed image looks in colour | A. Dim to a brightness ceiling by default; the host can change it · B. Place it exactly as served · **C. Re-colour it entirely from intensity classes** | A ◇ |
| D-37 | The defect ledger | **Approve as drafted** · approve with changes · take any row separately. *L-17 (f), the antimeridian, was singled out with the option to fix it; not taken.* | Approve |
| D-42 | What "first-class" means for the block renderer | *Not put as lettered options.* HUM LEAD's words on seeing the thinned specimens settled it: acceptable as a renderer a user opts into, provided they can switch back to braille or remove roads. | — |
| D-43 | How M1 is judged | Judge — **A. HUM LEAD against a computed answer key** · B. HUM LEAD alone · C. Someone who did not build it. Live session — **1. M1 gates this project's SHIP on the library; the session in the first host is the host's metric** · 2. The session in the host gates SHIP | The seven scenarios; A; 1 |
| D-44 | What the first release contains | A. Radar-first slice: basemap, features, images · **B. A plus temperature (scalar grids, ramps, contours)** · C. Everything · D. HUM LEAD's own cut — *HUM LEAD added the sequence: then integrate into the first host at its v0.17.0, then build the other shapes* | A, with B preferred to over-building ◇ |
| D-45 | Must every placed image come with a colour table | A. Optional, with a place-as-served fallback at truecolor · B. Required, always · **C. Required, and the project ships a tested example table** | C |
| D-46 | Which tile schemas v1 reads | **A. OpenMapTiles only; the schema stays a seam** · B. Also Protomaps-schema files | A |
| D-47 | Is radar animation in scope, and when | **A. In v1, built after v0.1.0; designed into the contract now** · B. In v0.1.0 · C. Not in v1 | A |
| D-48 | What the 8 MB memory test contains | **A. A typical-day fixture, plus a separate rule for the worst case** · B. The worst case is the fixture · C. Raise the target | A |
| D-49 | Corrections to the frozen parity matrix | **A. Approve all six corrections and the two notes** · B. Approve with named exceptions · C. Leave the matrix as frozen | A |
| D-50 | Hygiene before publication | (a) A. Purge the superseded commit · B. Leave it; publish from a fresh clone · **C. Both**. (b) **A. Neutral wording in tracked documents** · B. Keep as is · C. HUM LEAD's own line. (c) **A. Keep the framework's name; replace its command names** · B. Replace the name too · C. Leave everything | a-C, b-A, c-A |
| D-51 | The reading of "0.01.0" | **A. v0.1.0** · B. v0.0.1 · C. Something else | A |
| D-52 | A plain-text answer to "where" | **A. A requirement, in v0.1.0** · B. A requirement, after v0.1.0 · C. Not the library's job | A |
| D-53 | Must ramps be safe for colour-blind users | **A. Required for the defaults; themed ramps checked and reported** · B. Required for everything; unsafe themes refused or repaired · C. A selectable safe palette, not the default — *HUM LEAD added: a diverging-ramp specimen, and that some scalars may be made non-themeable* | A |
| D-55 | Should any scalars be non-themeable | A. Everything stays themeable; the checker reports · **B. Temperature is fixed; the rest is themeable** · C. Themeable within a family, enforced | B |
| D-56 | The flashing marker, and reduce-motion | **A. Slow the flash, add reduce-motion, never hide on a timer** · B. Match upstream; add reduce-motion only · C. Drop flash entirely | A |

## Where HUM LEAD chose against the recommendation ◇

Six times: **D-14** (all five overlay shapes, not three), **D-18** (a PMTiles reader in v1, not deferred), **D-19** (no remote until SHIP, not a private one now), **D-34** (contours with a specimen first, not the value lattice), **D-36** (re-colour images, not dim them), **D-44** (the wider first slice). In five of the six he chose the wider or more demanding option. Each was his to make; the recommendation and the argument against it were both in front of him.

## Rulings not given as options

D-32, D-40, D-41, D-42 and D-54 are HUM LEAD's verdicts on specimens he viewed, in his own words. D-38 confirmed a proposed set of review personas. D-39 is his clarification of three earlier rulings.
