---
title: "go-tuiMaps v0.3.0 (propagation fields) — HUM LEAD rulings"
date: 2026-10-06
phase: DISCOVER
sev: SEV-0
authority: HUM LEAD
status: "LIVE — every ruling is written here the moment it is made."
---

# v0.3.0 — rulings

Recorded verbatim. A correction to a ruling is a new row, never an edit of an old one.

v0.3.0 is paired with watchpost 0.19.0 and go-ionomaps, named go-giro-data until watchpost D-56 (watchpost D-3). A decision that binds more than
one project is written in watchpost's log
(`watchpost/06_docs/02_features/propagation-overlays/02-analysis/rulings.md`, cited "watchpost D-n") and
restated here; this log holds the library's own. v0.2.0's log is `../radar-loops/02-analysis/rulings.md`.

| # | Date | Question | HUM LEAD verbatim | Ruling |
|---|---|---|---|---|
| D-0 | 2026-10-06 | Restates watchpost D-3 | see watchpost D-3 | go-tuiMaps v0.3.0 is a paired release, run as v0.2.0 was: its own branch, rulings and gates; watchpost pins its release candidates in BUILD and ships on its final tag. Branch `feature/propagation-fields`, cut from `main` (`d1c4d5e`) with v0.2.0's REFLECT and release checklist carried from `feature/radar-loops` (watchpost A-7). |
| D-1 | 2026-10-06 | Restates watchpost D-32 | see watchpost D-32 | HR-1: a field continues over water per overlay (a field on `Grid`, the mask on `Image`); FR-12/D-87's promise kept; `TestHostCanFlipEither` written. |
| D-2 | 2026-10-06 | Restates watchpost D-33 | see watchpost D-33 | HR-2: `muf` and `fof2` presets, and FR-15's generic host ramp from `low`/`middle`/`high`, held to `CheckRamp`; a host type that would draw nothing is refused or warned. |
| D-3 | 2026-10-06 | Restates watchpost D-34 | see watchpost D-34 | HR-3: contour labels carry units; labels on north-south crossings too, for every field; `Answer.ValueUnit` for every field. |
| D-4 | 2026-10-06 | Restates watchpost D-35 | see watchpost D-35 | HR-4: `terminator` and `night` tokens with `Terminator(at)` and `NightSide(at)`. HR-5: `RoleFor(type, value)`, the value always in the point's label (OW-17). |
| D-5 | 2026-10-06 | Restates watchpost D-36 | see watchpost D-36 | `SafeRamps` decided by `ScaleClass`, with a test over every scale token. |
| D-6 | 2026-10-06 | Lock v0.3.0's problem statement as drafted | "Approve as written" | **LOCKED:** "A developer embedding go-tuiMaps cannot draw a global field of their own kind so that it can be read: the field stops at every coast, has no colours or legend of its own and can draw nothing at all, its labels carry no unit and are missing where its lines run east to west, and nothing shows where day ends, so the person reading the map sees a broken or empty picture, or numbers without meaning." |
| D-7 | 2026-10-06 | Adopt the metrics M1 to M6 as drafted | "Adopt; M4 for v0.3.0's rows only" | **ADOPTED:** M1 field read, M2 no silent blank, M3 units everywhere, M4 promise truth (this release's rows only), M5 globe cost, M6 gate trust; targets set in PLAN. |
| D-8 | 2026-10-06 | Approve the brief as presented | "Approve" | **APPROVED.** Intake closed for v0.3.0; its `requirements.md` follows in DISCOVER. No issue of record opened (not asked for). |
| D-9 | 2026-10-06 | Restates watchpost D-49 | see watchpost D-49 | M5's target is set from PLAN's dry run by its own ruling, before PLAN exits. |
| D-10 | 2026-10-06 | Restates watchpost D-50 | see watchpost D-50 | **L-5 (station points by value) leaves v0.3.0**, amending D-4's HR-5 half; carried as a watchpost follow-up with the station dots. |
| D-11 | 2026-10-06 | Restates watchpost D-65 | see watchpost D-65 | OW-17, OW-19, OW-24 and OW-25 are v0.3.0 requirements (L-7); OW-14, OW-15, OW-16, OW-18, OW-20 and OW-26 are re-targeted to v0.4.0. |
| D-12 | 2026-10-06 | Restates watchpost D-66 | see watchpost D-66 | `feature/radar-loops` and `feature/go-tuimaps` bundled to `~/` and deleted. |
| D-13 | 2026-10-06 | Restates watchpost D-84 | see watchpost D-84 | DISCOVER exits; v0.3.0's requirements are APPROVED; PLAN opens. |
| D-14 | 2026-10-07 | Restates watchpost D-96 | see watchpost D-96 | M5's target: after a pan, a whole-globe 2° field with fill, labelled contours, terminator and night in at most 16 ms at 400 × 110 and 8 ms at 200 × 56 (Apple M5 Pro); the profiled hotspots (per-cell contrast with `math.Pow`, blending, one allocation per cell) are fixed in P7. |
| D-15 | 2026-10-07 | Restates watchpost D-117 | see watchpost D-117 | M1: 10 places, 8 of 10 from the picture and from `Report`'s words, graded by the HUM LEAD (P8); M6: 5 consecutive clean full runs. |
| D-16 | 2026-10-07 | Restates watchpost D-119 | see watchpost D-119 | Gates assert machine-independent proxies; wall-clock targets are a release step on the reference machine, failing when the record is missing. |
| D-17 | 2026-10-07 | Restates watchpost D-127 | see watchpost D-127 | L-2.6: a two-class host field's edge (the reach area) carries a non-colour mark, is never a contour, and its combination with a fill beneath is stated. |
| D-18 | 2026-10-07 | Restates watchpost D-137 | see watchpost D-137 | M5's targets apply to the full frame including the reach overlay and its non-colour edge (L-2.6). |
| D-19 | 2026-10-07 | Restates watchpost D-139 | see watchpost D-139 | PLAN exit approved for v0.3.0 with watchpost 0.19.0 and go-ionomaps; BUILD opens. |
| D-20 | 2026-10-08 | Restates watchpost D-144 | see watchpost D-144 | L-2.1: the `muf` preset breaks at the amateur band edges, 3.5, 5.3, 7, 10.1, 14, 18.068, 21, 24.89 and 28 MHz; ten classes. |
| D-21 | 2026-10-08 | Restates watchpost D-145 | see watchpost D-145 | L-2.1: the `fof2` preset breaks at the band edges below 15 MHz, 1.8, 3.5, 5.3, 7, 10.1 and 14 MHz; seven classes. |
| D-22 | 2026-10-08 | Restates watchpost D-146 | see watchpost D-146 | L-2.1: the `muf` and `fof2` ramps stand as shown in the specimens: dark violet to pale, higher frequency lighter, one ramp on both grounds. |
| D-23 | 2026-10-08 | Restates watchpost D-147 | see watchpost D-147 | D-22 withdrawn for MUF (its evidence drew nothing, A-5): the MUF ramp is searched again, one smooth curve with even steps on the violet-to-pale path, and shown again; foF2 shown beside it. |
| D-24 | 2026-10-08 | Restates watchpost D-148 | see watchpost D-148 | The smooth MUF ramp stands; the foF2 ramp is redone from the MUF curve, frequency for frequency, and shown again. |
| D-25 | 2026-10-08 | Restates watchpost D-149 | see watchpost D-149 | foF2 is floors first: below 1.8 MHz nothing is drawn; its six drawn classes take the MUF colours of the same frequencies (tokens `fof2.1`-`fof2.6`). |
| D-26 | 2026-10-08 | Restates watchpost D-150 | see watchpost D-150 | A-5 is hotfixed: go-tuiMaps v0.2.1 (the fix alone, from `main`) and watchpost 0.18.1 pinning it, each through its full gate and CI, with a GitHub issue. |
| D-27 | 2026-10-08 | Restates watchpost D-151 | see watchpost D-151 | UAT-1, the early slice, needs this library's P0, P1 and P2.1 and `v0.3.0-rc.1`. |
| D-28 | 2026-10-08 | Restates watchpost D-152 | see watchpost D-152 | GO-2026-6617: every module's `go` directive is 1.26.9; the gate, CI and builds run on go1.27.2. Supersedes v0.2.0 D-133's floor of 1.25.13. |

## Agent decisions (A-n, watchpost D-13)

Decided by the agent under watchpost D-13 and reported in a batch for veto. A vetoed row stays, and its veto is a D-row.

| # | Date | Decision | Why | Where |
|---|---|---|---|---|
| A-1 | 2026-10-08 | **A rain grid is drawn over the sea whatever `Grid.OverWater` says**, as a wave grid stays on the sea whatever it says; `Image.MaskedByWater` is an image's alone | L-1.1 keeps v0.2.0's defaults. A rain grid is drawn over the sea today (D-87, L-17.1), and `OverWater`'s zero value is "stops at the shore", so honouring it for rain would change every rain grid's default | P1 |
| A-2 | 2026-10-08 | **NFR-2's mutants:** a table, `06_docs/mutants/mutants.json` (id, row, what it breaks, file, the text and its change, package, the killing test). `TestEveryMutantFindsItsLine` runs on every test run; `TestEveryMutantIsKilled` (tag `mutants`) applies each mutant to a copy of the tree, never the tree itself, requires the copy to vet, and requires the test to fail; `scripts/gate`'s full lane runs it ("NFR-2, every mutant killed by its test") and fails when no mutant ran | watchpost's harness edits the working tree in place, and its own comments record two losses of uncommitted work to restores; a copy cannot touch the tree. A mutant that does not compile is no evidence, so vet comes first | P1 |
| A-3 | 2026-10-08 | **A polygon ring left open - by the next ring's MoveTo, or by the end of the geometry - or run by a second LineTo is a damaged stream, refused (D-75).** The oracle stays exact, with no exemption (D-118 made the last one); the failing input stays as `FuzzAgree`'s seed | P1's full gate: `FuzzAgree` found a tile on which the two decoders disagree. The proven decoder (orb `decodePolygon`) reads one command after every ring as its ClosePath, so it took the next ring's MoveTo for one and read that MoveTo's numbers as commands; ours kept the open ring. The next full run found the same misreading after a second LineTo (a ring of MoveTo, LineTo, LineTo, ClosePath). MVT 2.1 (4.3.4.4) ends every ring with ClosePath, so both read a stream neither should accept. The precedents are v0.1.0's fuzz fixes, which refused damaged streams (D-75, D-126), and v0.2.0 D-140, which kept the oracle exact. A well-formed tile draws as before: the fixture tiles agree exactly. Only a malformed tile is refused where it was drawn | P1 |
| A-4 | 2026-10-08 | **A rank is a whole number from zero: a negative one is no rank** (0) | The second full run's `FuzzAgree` input ranked a feature -27 from a zig-zag (sint64) value; the proven decoder's reading has none, and D-117 holds that a number this decoder invents where they read none is a failure. A negative int64 was already passed over (its encoding is past 32 bits), so the two encodings of a negative number were treated differently. OpenMapTiles' ranks are from zero; the admin level (1 to 11) and the maritime flag (1) were already held to their ranges | P1 |
| A-5 | 2026-10-08 | **The literal inks start at 160, not 96, and a test holds every token below them** (`TestEveryTokenIsAnInkBelowTheLiterals`, `TestEveryClassOfEveryPresetIsDrawn`) | A cell's ink below `firstLiteral` is a token's number, at or above it a style's literal colour. v0.2.0's tokens ran to 97, so `qpf.6` and `qpf.7` (96, 97) were read as literals and drawn in no colour: **v0.2.0 draws nothing for rain totals of 50.8 mm and more** (measured: 10 and 30 mm drew 1,920 cells each, 60 and 120 mm none). The legend keyed both classes, so no test saw it. The new MUF and foF2 tokens fell past it too and drew nothing, which is how it was found. 160 leaves room for tokens to come and 96 literal colours for a style. Shipped in watchpost 0.18.0 (go-tuiMaps v0.2.0); its handling there is the HUM LEAD's | P2.1 |
