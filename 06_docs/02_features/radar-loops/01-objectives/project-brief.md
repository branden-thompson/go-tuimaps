---
title: "go-tuiMaps v0.2.0 — Radar loops — PROJECT BRIEF"
date: 2026-09-22
phase: DISCOVER (RCC)
report_template: project-brief v1.1.0
level: LEVEL-1
sev: SEV-0
authority: HUM LEAD
directives: FULL GIT; FULL DOCS; FULL REPORTS; FULL DIAGRAMS; FULL RCC; FULL PLAN; FULL TDD; FULL INST
branch: feature/radar-loops
issue: "branden-thompson/go-tuimaps#2 — this brief is its body (D-7)"
paired_release: "watchpost 0.18.0 — Observer maps (branden-thompson/watchpost#22); this release ships first (watchpost D-11)"
status: "APPROVED by the HUM LEAD 2026-09-22 (D-7); AMENDED 2026-09-23 (D-11) for the five host requirements watchpost's red-team rounds added — L-1.5, L-7..L-10; CORRECTED 2026-09-23 (D-23): the requirement set is requirements.md, which wins on any conflict.  Problem statement LOCKED (D-5)."
---

# Library Release | `go-tuiMaps v0.2.0 — Radar loops`

**LEVEL-1; SEV-0; FULL GIT; FULL DOCS; FULL REPORTS; FULL DIAGRAMS; FULL RCC; FULL PLAN; FULL TDD; FULL INST**

## Summary & Intent

go-tuiMaps v0.1.0 draws radar as **one picture**. That was deliberate: a single georeferenced image
proved the plumbing — projection, resampling, the colour-to-intensity table — without the weight of
time. Nobody reads radar that way. A person looking at precipitation wants to see it **move**.

v0.2.0 makes radar a **loop**, and closes the gaps the first real host found in the contract.

**Why now.** Watchpost 0.18.0 draws the library's first real Observer maps, and its HUM LEAD named
radar the most common view (watchpost D-7) and ruled it must loop (watchpost D-10). The paired-release
ruling (watchpost D-11) puts this release first: watchpost's alert drawing proceeds on v0.1.0, and its radar waits for
this tag.

**Who benefits.** Watchpost's listener: watchpost is the only host, and no other was consulted, so
benefit to other terminal applications is hypothetical (corrected, D-37).

**Scope (D-37).** Three parts, of which the locked problem below names the first: **loops**, with
playback and their accessibility; **the contract repaired**, so a host's reading of it is true; and
**the library's first reviewed release** (D-18), carrying the accessibility and security work the
DISCOVER red team added. Nothing is split out to a later release (D-36); there is no target date.

**What happens if this is not built.** Watchpost ships radar as a still frame the HUM LEAD has
already ruled is "not the product", or ships no radar at all in the release whose most common view
is radar.

## Problem Statement — LOCKED

> **"A developer embedding go-tuiMaps can show where precipitation is, but not where it is going:
> the library draws one radar frame, so the person reading the map cannot see a storm's motion,
> and the host has no way to give the library the frames that would show it."**

**LOCKED by the HUM LEAD, 2026-09-22** (*"A approved"*, D-5). It names two sufferers — the embedding developer and the
person reading the map — because a library's problem always reaches the second through the first.

## Requirements — from the host, and from the record

> **The requirement set PLAN designs against is `requirements.md` (D-23).** It extends the L-numbers
> below with everything ruled since (D-14 onward, specimen 29), holds the risk register and the list
> of owed work, and wins on any conflict. The rows below are the brief as approved, with its stale
> rows corrected.

Watchpost's host requirements HR-1..HR-10 (`watchpost/06_docs/02_features/observer-maps/
01-objectives/project-brief.md`) map to L-1 (HR-1, with L-1.5 from HR-9), L-2 (HR-2), L-3 (HR-3),
L-4 (HR-4), L-5 (HR-5), L-7 (HR-6), L-8 (HR-7), L-9 (HR-8) and L-10 (HR-10). L-6 is this brief's own.

### L-1 — Loops (HR-1)
- **L-1.1** An image overlay carries **several frames, each with its own valid time**; zero frames
  means today's single image (contract §9: additive). `NextCall` must report the next frame change —
  **the contract says it "already has a frame-advance source"; the code does not** (C-2).
- **L-1.2** Frames may have **gaps** — sources skip and expire (W1-A: both sources return empty
  frames with a success code). A missing frame is a stated gap, never a duplicated neighbour.
- **L-1.3** Each frame carries its own valid time into staleness and description, so an old loop
  never looks current.
- **L-1.4** The library never fetches radar. The host hands in finished frames (v0.1.0 D-65: nothing is
  fetched until the host says so; tile-image providers stay deferred unless ruled in).
- **L-1.5 (HR-9)** **Playback of an overlay's loop is controllable — at least off / slow / normal** —
  and the control is the library's, exposed to the host and through it to the listener. The host
  supplies the data; this library turns it into frames and plays them, so the rate belongs where the
  frames are, and a host must not have to fake a slower loop by withholding data. Applies to every
  looped overlay, not only radar. A **stated maximum rate** rides with it: `ReduceMotion` is a
  boolean about markers and the clock, and nothing yet says what it means over a twelve-frame loop.
  *(Watchpost's motion Settings row and its FR-5.9 ceiling have no mechanism without this.)*

### L-2 — The MRMS table (HR-2)
- **L-2.1** A colour-to-intensity table for NOAA/NCEP MRMS `conus_bref_qcd`, owned here (watchpost
  D-9).
- **L-2.2** **MRMS publishes no exact table** (W1-A §5: no `ColorMapEntry`; a 500×30 picture legend,
  −20…70 dBZ). An exact table like IEM's needs the source colour map; a table derived from the legend
  is approximate. *Ruled at D-19:* the palette MRMS actually uses, valued from the legend, labelled
  approximate — the Designed-for vs Validated-on calibration forbids calling it exact.

### L-3 — The view bound (HR-3)
- **L-3.1** A host can set a **minimum zoom and a bounding box once**, and the library holds the view
  inside them across every path — resize, fit, pan, zoom, and the fall-back to the whole world
  (`map.go:259-265`), which today ignores any host intent.
- **L-3.2** Additive: a host that sets nothing gets v0.1.0's behaviour.

### L-4 — The contract matches the code (HR-4)
- **L-4.1** `contract.md` names `Frame.Line`, `WriteTo`, `SetSize`, `Focused`, `BorrowCheck`; the
  public surface has none of them. The document is corrected, and a test holds it to the public
  surface so it cannot drift again ("the rule an agent has to remember…", watchpost D-11).

### L-5 — The after-tag triage (HR-5)
- **L-5.1** *Superseded by D-18.* v0.1.0 D-123 named a triage of nineteen open items; wave 1 found it never
  existed and v0.1.0's close-out never ran. D-18 records that as not recoverable, adds a gate test
  that refuses a tag while its checklist is unfinished, and makes v0.2.0 the first release.

### L-6 — Found while preparing this brief
- **L-6.1** **Hosted CI.** `L2-gates.md:42` says SHIP turns the local gate into the hosted workflow.
  v0.1.0 shipped; `.github/` has no workflow. The gate still runs only on one machine.
- **L-6.2** **A host cannot write a `Fetcher`.** `Fetcher = fetch.Func` takes the internal
  `fetch.Request`; the watchpost basemap survey needed reflection to supply one.
- **L-6.3** **The default memory tile cache is smaller than one large view.** A state view at 149×38
  needs ~909 KB against a 500 KB default (`internal/tiles/cache.go:17`); the cap is reported, not
  enforced. At minimum documented for hosts; possibly a better default.
- **L-6.4** **The gate's fuzz legs fail by accident on this machine.** Three targets in three packages
  (`FuzzAgree`, `FuzzDirectory`, `FuzzHandIn`), across two gate runs, ran healthily and then failed with
  "context deadline exceeded" at their time budget. No input is at fault: a one-second watchdog on both
  of `FuzzAgree`'s decoders never fired across ~27 M executions (D-3). **Probable root cause, from the
  Go 1.25.0 source — to be proven by D-4's reproduction:** with a time budget the coordinator wakes on
  the parent context's done channel and calls `stop(ctx.Err())`, which suppresses the error only if it
  equals `fuzzCtx.Err()` (`internal/fuzz/fuzz.go:129`); a parent context closes its done channel
  *before* cancelling its children (`context/context.go:565`, then `:569`), so on another core the
  deadline can be reported as a failure — a race, far likelier with 18 cores. **Candidate fix:** a
  run-count budget (`-fuzztime Nx`) stops through the execution limit and creates no deadline. A
  single 40-second mid-run freeze seen once in `FuzzAgree` is a separate symptom, not explained by
  this, and stays open. *Corrected (D-23):* **fixed by construction at D-9 and D-10** (run-count
  budgets), **not reproduced**, so the cause above stays probable; commits are no longer blocked.

### L-7 — A host can write a fetcher (HR-6)
- **L-7.1** `Fetcher` is exported, but its type alias resolves to an **internal** request type that
  cannot be named outside the module, so a host cannot supply one without reflection. Export the
  request type and its options — a user-agent token, allowed hosts, a timeout, roots.
- **L-7.2** Until it lands, tiles go out as **this library's** user-agent, naming neither the host
  nor a contact, against sources whose terms the host is asked to honour. That is stated in
  watchpost's NFR-4 as an exemption; it should not need one.

### L-8 — An alert pattern that does not depend on colour depth (HR-7)
- **L-8.1** The hatch runs only at `NoColour` depth, so a host's colour-on monochrome or light theme
  draws a tint with **no second channel** — which breaks the "meaning never by colour alone" promise
  the library makes and the host repeats.
- **L-8.2** `AlertExtremeTint` and `AlertSevereTint` share the stroke `╳`, differing only by spacing
  2 versus 3, and **Minor and Unknown share `╱`** (corrected, D-23). On a small area that is not a
  distinction a reader can use. Ruled at D-17: a severity word and a distinct dash per level.

### L-9 — Cache retention and a purge call (HR-8)
- **L-9.1** `CacheRoot(dir, capBytes)` takes bytes only; the disk cache keeps tiles **without
  expiry**, and its file names are a record of where the reader has looked. A host that promises its
  listener a retention cannot keep that promise today.
- **L-9.2** A host-settable **maximum age**. *Corrected (D-23):* **`Map.Purge` already exists**
  (`tiles.go:145-165`); what is missing is its **scope** — it empties only the current source, not
  memory, shared caches or decoded pictures.

### L-10 — The confinement of a source's tiles (HR-10)
- **L-10.1** The effective tile host comes from the TileJSON document, not from the address a host
  configures. The library confines it today — a source's tiles must come from that source's host —
  and that confinement is **contract**, stated and kept, because a host's closed source list is only
  as good as it.

## Technical Constraints

**Measured against `feature/radar-loops` at `650f267`, not assumed.**

### C-1 — An image is one picture, and the shape is ready for more
`overlay.Image` holds one `PNG`, its bounds, projection and colour table (`internal/overlay/image.go:53-61`).
Frames are an added field; nothing existing has to change shape (contract §9).

### C-2 — The contract promises a frame-advance source the code does not have
`NextCall` consults the markers' motion and each overlay's valid/keeps change (`clock.go:93-110`).
There is no frame source to reuse — it is new work, and the contract's sentence is corrected with it.

### C-3 — The image cap is per image, and a host cannot raise it
250,000 pixels at one byte each (`internal/overlay/store.go:76`); a cap may only be *lowered*
(`store.go:135`), and `New` passes an empty `Caps{}` (`map.go:204`). A 12-frame loop at the cap is ~3 MB
of classified pixels — the number M4 is measured against, and the first question PLAN answers.

### C-4 — Nothing holds `contract.md` to the code
`TestPublicSurfaceSnapshot` (`surface_test.go:22`) pins `public-surface.txt`; no test reads the
contract. That is how it drifted (L-4).

### C-5 — The view can widen without the host
Watchpost's wave 1 (W1-B) counted five paths; the four this record traces are: unplaced map → whole world (`map.go:214-215,
259-260`); resize keeps zoom so ground grows (`:262-263`); invalid view → whole world (`:264-265`);
`MinZoom = −8` a constant (`view.go:16`).

### C-6 — The gate failed by accident on this machine
L-6.4. Fixed by construction at D-9 and D-10; not reproduced.

### C-7 — Every change is additive
v0.1.0 is published and resolvable, and watchpost will depend on it. A breaking change is a HUM LEAD
ruling (contract §9).

### C-8 — Five of these requirements came from a host's review, not from this record
L-1.5, L-7, L-8, L-9 and L-10 were found by watchpost's accessibility and security reviewers during
its 0.18.0 DISCOVER (its rounds 2 and 3), not by this library's own rounds. Two consequences: this
brief's requirement set grew by half after it was approved, and the library's own DISCOVER should
expect its blind spots to be where a host looks and it does not — the host is the only reader that
uses the contract rather than writing it.

## Metrics of Success — ADOPTED (D-6)

| # | Metric | Type | Definition | The anti-solution it closes |
|---|---|---|---|---|
| M1 | Motion seen | Primary | From a loop alone, a reader states a precipitation cell's direction of motion — scripted UAT over loops recorded from real radar; **plus a non-visual arm from the description alone (D-24; `requirements.md` holds the normative metrics)** | Blinking one frame; the loops carry real motion |
| M2 | Loop honesty | Primary | Zero frames drawn under the wrong valid time; every gap stated; every frame older than the host's stated cadence marked stale | Dropping gaps and old frames — a missing frame is **shown as a gap**, not skipped |
| M3 | Bound holds | Primary | With a host bound set, zero frames outside it across resize, fit, pan, zoom and fall-back | Refusing to draw — a blank frame with data on hand fails M4 |
| M4 | Embed cost with a loop | Primary | Memory with a 12-frame loop within a PLAN-set target, flat over an hour | A tiny loop — measured at 12 frames, the host's largest box |
| M5 | Contract truth | Secondary | Zero names in `contract.md` missing from the public surface, and none the other way — a test | Deleting the document — the test demands coverage both ways |
| M6 | Gate trust | Secondary | Zero unattributable gate failures across a stated number of consecutive full runs on this machine | Dropping the fuzz legs — every v0.1.0 leg still runs |

## Other Considerations

- **Quality pass, F-1 and F-2** (`06_docs/follow-ups.md`): root clutter, and a pluggable
  architecture like watchpost's. HUM LEAD: *"No action at this time"*. Recorded here because v0.2.0
  adds a second radar source's table — the first new source since the seam question was raised.
  *Ruled at D-35:* v0.2.0 builds one narrow seam, for provider colour tables; the broad
  restructure stays with the quality pass.
- **Standing principles** (watchpost D-23): P-1 performance and structure; P-2 host and user choice
  by default; P-3 never architect into a corner. For a library, P-2 reads as *the host chooses*.
- **Additive only.** Every change is additive to the v0.1.0 contract (contract §9); a breaking
  change is a HUM LEAD ruling, not a PLAN detail.
- **Issue of record.** Per the standing convention, the approved brief becomes the body of a GitHub
  issue labelled `enhancement` on `branden-thompson/go-tuimaps`, cross-linked to watchpost#22.

## Completeness Check

```
PROJECT BRIEF — COMPLETENESS CHECK
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  [✓] Header              — Scope, type, name, paired release
  [✓] Directives          — LEVEL-1, SEV-0, full set (D-2)
  [✓] Summary / Intent    — What, why now, who benefits, cost of not building
  [✓] Problem statement   — LOCKED (D-5), 4.5/5
  [✓] Requirements        — L-1..L-10, from watchpost's HR-1..HR-10 and four findings of this brief's own
  [✓] Metrics of Success  — M1–M4 primary, M5–M6 secondary (D-6)
  [✓] Tech Constraints    — C-1..C-8, measured at 650f267
  [✓] Considerations      — quality pass, principles, additivity, issue of record

  [✓] Approval            — APPROVED as presented (D-7)
  [✓] Outstanding         — none.  INTAKE CLOSED.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```
