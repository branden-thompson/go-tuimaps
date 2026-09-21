# Integration review — template (D-60)

Up: [first-host start](../04-development/first-host-start.md) · Carries: NFR-4, NFR-5, NFR-22, D-29, D-48, D-60, D-85, D-90, D-92, D-94

| Field | Value |
|---|---|
| Phase | BUILD, task 14.19. Filled in from the integration itself, on 2026-09-20 |
| Why this exists | HUM LEAD ruled (D-60) that the remaining overlay shapes are built only after a **written** review of the first integration and HUM LEAD's GO. The plan cited this template without writing it (P2-PRD-17). |
| Who fills it in | The coordinator, from the integration itself — not from memory. Every answer names its evidence: a commit, a measurement, a line of the host's code |
| Who decides | HUM LEAD: GO, GO with changes, or stop |

## 0 - The finding that matters most to both projects

**The first host's alerts carry no geometry.** `snapshot.Alert` holds `AffectedZones []string`
and an `AreaDesc` string and no polygon; the host's own notes say no provider it uses carries
them. Its earthquakes lose their coordinates at the snapshot boundary, keeping a distance and
a compass word instead. It has no radar raster and no temperature field.

So **the overlay shapes this library was built for - the alert area above all - could not be
fed by the first host at all.** What the spike could hand in was points: fire detections and
named incidents, which worked on the first try.

This cuts two ways and both are worth stating plainly.

- **For the library**: the alert-area path, which carries the most design, the most colour
  work and the whole no-colour hatch question from the acceptance sitting, is **still
  unexercised by a real host**. Scenario 1 and 2 in the sitting are the library's own
  fixtures, not a host's data.
- **For the host**: plotting its alerts needs work upstream of the map - either carrying the
  NWS polygons it currently discards, or resolving zone identifiers to shapes. That is the
  host's decision and nothing the contract can fix.

**It also revises what D-60's "one overlay from the host's own data" proved.** It proved the
overlay contract for *points*. It did not exercise areas, lines, circles, grids or images.
A host consuming this library for its alert areas is still a thing nobody has done.

## 1 · What was integrated

| Question | Answer |
|---|---|
| Library version and commit; host version and commit | Library at `8f2bfc1` plus the resize fix below, consumed at the placeholder `v0.0.0` through the local override (D-19, CD-4). Host: watchpost `0.14.2-524-ga618bde`, branch `spike/tuimaps-first-host`, never pushed |
| Which maps, at which sizes, with which overlays, from which of the host's data sources | One map, in a modal window of the Observer dashboard on key `g`. Built at 120x40 and drawn at the modal's own box - about 85x30 on HUM LEAD's terminal, 73x26 in the tests. One overlay, `fires`, built from `snapshot.Location.Fire`: each GOES hotspot as a point roled by its confidence, each named WFIGS incident as a labelled point. The watched place itself is an `AddPlace` marker |
| How the host runs the pump: how many goroutines, how it is woken, how it stops | Two goroutines, started from the composition root beside the other pipelines. `OnPending` nudges a one-deep channel; each unit of work done calls back into the host, which does `p.Send(MapChangedMsg{})` so Bubble Tea redraws. A cancelled context stops both, and `Close` waits for them. **It went in without a fight and is the part of the contract that needed no explaining** |

## 2 · What the host needed that the contract lacked

One row for each thing the host had to work around, wanted and did not have, or got wrong at first.

| Need | What the host did instead | Contract section | Proposed change | Breaking? (NFR-22) |
|---|---|---|---|---|
| **The view had to survive a change of size.** A terminal host cannot know its size when it builds the map: the size arrives with the first frame and changes with the window. Every size change replaced the view with a whole-world view, so the host's chosen place was lost on the **first** frame, silently | Nothing could be done host-side; the size is not knowable at `New` and changes again on every resize. **Fixed in the library during the spike, on HUM LEAD's instruction** | Render | Taken: a map keeps the place the host chose across a resize, and one nobody has pointed anywhere is still refitted to the world | No - it restores what a host already expected |
| ~~A name for the type of `Feature.Role`~~ | **Withdrawn. The coordinator was wrong, twice, and it is recorded because of how the error was made.** `tuimaps.Token` is exported and documented directly above the role constants. It was called missing because the package's types were listed with the output truncated one line before `Token`, and the wiring agent reached the same conclusion independently - **two agreeing reports, neither of which had checked**. The host wrote an inline `switch` to work around a gap that did not exist; it has been replaced with the ordinary helper | - | None | - |
| ~~A structured reader for centre and zoom~~ | **Withdrawn, and recorded because the mistake is instructive.** This was written up as a missing call on the strength of `Footer()` returning a human string. `Map.Centre() (LonLat, float64)` already answers it exactly, and was found by checking before implementing rather than by trusting the report. The host's wrapper simply never exposed it, and its window titled itself from the host's own snapshot instead. **A host mistake, not a contract gap** | - | None | - |
| **The deepest zoom a map can actually draw at.** `Zoom` accepts anything from -8 to 18, so `Zoom(6)` succeeds against embedded tiles that stop at 3 | Read `assets.MaxZoom`, which is a property of the assets package and not of the map that was built | View | A reader for the deepest drawable zoom, or a clamp that says it clamped | No - an addition |
| **A way to drive the map from a key press.** `Recentre` and `Zoom` mutate, so they cannot be called from a Bubble Tea `View()`; every interaction has to be routed back through the host as a command | Not attempted in the spike: the window draws and does not pan or zoom. **HUM LEAD has named this as needed for the real integration** | View | No API change needed, but the contract should *say* this and show the shape, because the same mutation assumption is what caused the resize defect above | No - documentation |
| **A frame's status, carried out of the wrapper.** `Frame.Status` distinguishes "still loading" from "finished", which this host draws everywhere else | The wrapper returned `[]string` and dropped it; the window cannot show a loading state | Render | None - the library offers it and the host's first wrapper threw it away. **Recorded as a host mistake, not a contract gap** | - |

## 3 · What the contract has that the host did not use

Unused surface is a cost. One row each; say whether it should stay.

| Call or field | Why unused | Keep? |
|---|---|---|
| `FitTo` | The host centres on one watched place, not on a set. It would be the natural call for "show me all my places at once", which this console will want later | Keep |
| `Describe` and the whole description path | The console already speaks its own alerts and has its own voice. **This is worth noticing: the library's strongest feature in the acceptance sitting was the one the first host had no use for**, because the host had already solved that problem its own way | Keep, but do not assume a host wants it |
| `SetStyle`, `Layers`, `SafeRamps`, `ColourDepth`, `Units`, `LabelLanguage` | The spike took the defaults | Keep - untested by this integration rather than unwanted |
| `Circle` and `Grid` features, images | The host has no radar raster and no field data; its alerts carry no geometry at all | Keep |
| `Legend`, `Scale`, `ShowFooter`, `Warnings`, `Places`, `Verify`, `Purge` | Not reached in a spike of this size | Unknown - a longer integration is needed before any of these is judged |

## 4 · Measured in the real host

| Measure | Target | Measured | How |
|---|---|---|---|
| Live memory added, three maps | 4 MB (D-48, D-85) | | |
| Peak memory added | 8 MB (D-29) | | |
| `CacheUse`: need against cap, in the host's real views (D-90) | need ≤ cap in ordinary use | | |
| Cold start to full detail | 3 s (NFR-5) | | |
| A changed frame | the host's own window cost (NFR-4) | | |
| `Set` on the largest overlay the host really sends (D-92) | a few milliseconds | | |

## 5 · The first-hour mistakes

Which of these did the integration actually make, and did the library catch it? A discarded `Set` error · a mistyped id · a wrong unit · a pump that never ran · a borrowed buffer reused too early.

| Mistake | Made? | Caught by |
|---|---|---|
| Ran the toolchain's tidy step before any host code imported the library | **Yes.** Tidy then deleted the requirement again and left only the override, which reads as the recipe not working | Noticing the requirement had gone. **The recipe has been corrected** to say tidy comes after the first import |
| Drew the map without running the pump, and believed the frame | **Yes.** The frame said `no map tiles: name a source, or pass the assets package's tiles` **although the embedded tiles had been passed**. The state was right - nothing had been decoded yet - but the advice names the two things the host had already done and never mentions `Settle` or the pump | A probe that printed the frame. **This wording should change**: it is the one place in the spike where the library told the host to fix something that was not wrong |
| Zoomed deeper than the embedded tiles reach | **Yes**, `Zoom(6)` against tiles that stop at 3. It succeeded and drew from stand-ins | Only by reading `assets.MaxZoom`. See section 2 |
| Threw away `Frame.Status` in the host's own wrapper | **Yes** | Writing this review, not by anything failing |
| Guessed a size at construction | **Yes**, 120x40, invented. It is what exposed the resize defect | The defect |

## 6 · M1, in the host

For each scenario the host can now show: does the frame answer "where is it, relative to my place", and does the description match the key.

| Scenario | Frame | Description | Note |
|---|---|---|---|
| | | | |

## 7 · Does the plan for the rest still hold

| Deferred part (contract, section 9) | Still wanted? | Still additive? | Changed by what was learned |
|---|---|---|---|
| Wind · image loops · tile-image provider · block renderer · PMTiles source · pointer operations · flash, pulse, tours | | | |
| Reopened by this review: per-layer label margin and clustering (D-94) | | | |

## 8 · Recommendation

GO, GO with the changes listed in section 2, or stop — with the strongest argument against the recommendation.

## 7 - Three of the gaps this review first recorded did not exist

The spike's findings were written up from two sources: the coordinator's own
reading, and a report from an agent that did the host wiring. **Three of them
were wrong, and all three had the same shape** - "the library has no call for
this" - reached without looking at the whole surface.

| Claimed missing | Actually |
|---|---|
| A name for the type of `Feature.Role` | `tuimaps.Token`, exported and documented above the role constants. Missed because a listing of the package's types was read with the output cut off one line short |
| A structured reader for centre and zoom | `Map.Centre() (LonLat, float64)`. The report reasoned from `Footer()` returning a human string and stopped there |
| Anything for clock-driven change, so a memo would replay a stale frame | `Map.NextCall(wall)` gives "the earliest of the next marker phase, a failed tile's retry time and an overlay going stale (FR-25)", and `Stale` is compared in the renderer's own memo, so `Changed` moves too |

**The two wrong ones the coordinator carried were each believed the more firmly
because a second report agreed** - and that second report had not checked
either. Two agreeing accounts of an absence are not evidence of one.

What the pattern cost: a host wrapper written with an inline `switch` around a
type that has a name, a window that titled itself from the host's own snapshot
instead of asking the map where it was, and a memo that would have been keyed
on the wrong thing. All three were found by checking before implementing, which
is the only reason they are corrections here rather than defects in the library.

**The gaps that survived checking** are the ones acted on in step 1: the view
lost on a resize, a field's contours carrying no values, no way to list the
palette's token names, no way to ask the deepest drawable zoom, a notice that
named the wrong cause, the scale mark and credit touching on a narrow map, and
nothing in the documentation saying which calls move the map and which only
read it.

## 8 - What the oracle's fuzzer found after the first host, and how it was found

Two defects in this decoder, both of which it used to read past rather than refuse (D-126):
a field of a *feature* carrying the wrong kind of value - the check its *layer* has always
had - and a feature that says it is a point and then runs a line or closes a ring. The
proven decoder read a malformed id's bytes as though they were geometry; this decoder grew
one point into fifteen positions. Both are now refused (D-75), and 2.3 million fuzz inputs
have since agreed.

**The method is worth more than either fix.** Three hypotheses were reached by reading the
tile's bytes by eye and all three were wrong. HUM LEAD stopped it and said to consult the
diagrams. **L2-tiles named the decoder's two passes** - "count first, then allocate once" -
which led in minutes to the single line that drops a feature silently, and instrumentation
written with the decoder's *own* framing then showed both causes without guesswork.

The diagrams were updated in the same commit, and the rebuild of the architecture atlas
showed something worth keeping in mind: **the atlas is built from the mermaid blocks, so a
change written only in the prose beside a diagram never reaches it.** The label-ordering
change was moved into the diagram itself for that reason.

## 9 - What the second look at the first host found, and the one thing deferred

The first review of this integration read the library. This round read **both projects against
each other**, at HUM LEAD's direction, and found three things the first pass could not have:
two defects in this library, and one design question about its shape.

### 9.1 Two areas of one hazard cancelled each other in the spoken answer

`featureAnswer` gathered the rings of every feature into one list and `InArea` counted crossings
over all of them by the even-odd rule. **Two overlapping areas then cancel**, so a house inside
both was reported *outside*. The National Weather Service produces this constantly: one alert
naming a marine zone and the coastal land zone beside it, or two counties sharing a border.

The fill was always right - each feature is its own fill, so the overlap is drawn (FR-11). Only
the words were wrong, which is the worse half: **a sighted reader saw the house shaded and a
screen reader was told it was outside.** The 24-screen acceptance sitting could not have caught
it, because every scenario in it had one area.

Fixed by `InAnyArea`: each area is asked in turn, and even-odd still applies *within* an area so
holes still read as holes. Mutating the fix back to a flat list fails the test; single-area
answers are unchanged, so the sitting's approved frames still stand.

### 9.2 The description memo answered about the wrong place

`describeKey` recorded how MANY places a call asked about, not which. `Describe` takes its places
as an argument precisely so a host can ask about places the map does not store - **which is what a
station watching several locations does** - and `placesVersion` tracks only the map's own. So a
second call about a different place returned the first place's answer, name and all: asked about
Denver, answered about Fort Wayne, and reported a thousand kilometres away as *inside*.

The comment above the key said it "changes whenever the answer would (FR-29)". It did not. It now
fingerprints each asked place's name and position.

**Both defects are the same mistake**: a value that stood in for identity - all rings treated as
one area's, a count treated as a set. Neither was reachable by reading this library alone; both
took a second project asking it real questions.

### 9.3 Deferred: the host's own point type, and why generics were NOT the answer

A host keeps geometry in its own type, so handing it over costs a copy - 7.7 MB measured on the
largest alert in the country, once per overlay rebuild. `Feature` claimed its rings were
"borrowed and never copied", which **no host but a hypothetical one could achieve**. The claim is
corrected and `Rings`/`Ring` now do the conversion in one call.

Making the library generic so the copy disappears was proposed, approved, and then **withdrawn on
measurement** - recorded here because the measurements are the useful part:

| Approach | Cost over 2,000,000 positions | Why not |
|---|---|---|
| Interface accessor | **2.41 ms, +28%** | Pays every frame to save a per-update copy |
| Struct-layout constraint | free | **Refuses a tagged struct** - the first host's point carries JSON tags |
| Type parameters through the API | **1.61 ms, -14%** | Needs `Map[P]`; one point type per map, while `Place.At` stays `LonLat` |
| Convert once at the seam (**chosen**) | one copy per rebuild | - |

**What makes it due:** a measured workload where the per-rebuild copy matters - a host rebuilding
overlays at frame rate, or geometry an order of magnitude past a national alert set. Until then
the copy is a few hundred kilobytes every few minutes and the API stays one type simpler.

`Positioned` asks a point to *say* where it is rather than inspecting how it is laid out, so field
order, field names and struct tags stay the host's business - and a host's data model satisfies it
without importing this library at all, which is what the first host's layering required.
