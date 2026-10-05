# go-tuiMaps v0.2.0 — release notes

Up: [release checklist](release-checklist.md) · [requirements](../01-objectives/requirements.md)

**v0.2.0 adds to v0.1.0's contract, and breaks it only where ruled** (D-58): each break is listed
below, and in section 12 of the contract (`06_docs/02_features/go-tuimaps/03-architecture-design/contract.md`)
with what a host does instead. It is the library's first reviewed release (L-5.4).

**In plain words, for a designer or PM:**

1. Radar on the map moves: a host can play, stop, step and rewind a loop of radar frames, and the frame's time is written on the map.
2. A warning drawn over radar lets the radar show through, and each warning's severity reads as a word and a digit, not only a colour.
3. The map can tell, in words a host passes on, which way the heavier rain near a place is moving and how fast; no product voices those words yet.

## What it adds

- **Radar loops** (L-1): an image overlay may carry frames, each with its valid time; the host plays,
  stops, steps and resets them through one playback API, and the map draws the frame for its moment.
  A refreshed loop keeps drawing while its new frames are prepared (L-27).
- **Storm motion in words** (L-1.12, D-122): `Report` says which way the heavier rain near a place
  moves and how fast, measured over the whole loop; where the nearest of it is now and where it was;
  and whether it came closer, moved away or held. A loop with no motion to tell says why (D-111).
  Observation, never forecast. **No host voices it yet**: watchpost 0.18.0 does not read `Motion`
  (OW-21), so this is the library's answer, ready for a host to word. On the five loops of M1's first
  sitting its heading was 3 to 45 degrees off a human's reading, within one point of an 8-point compass
  where 45 degrees is the edge; those are the loops it was built on (D-127). `SpeedKmh` has been checked
  on synthetic loops only. Each entry carries `Stale`, true when the loop is stale by the map's rule
  (D-130).
- **Provider colour tables** (L-2, L-2.5): IEM's published table and MRMS's observed one, through one
  seam a third provider joins the same way.
- **A view bound the host sets** (L-3), **a fetcher the host may supply** (L-7), **tile-host
  confinement** (L-10), **cache age and purge** (L-9), and **an image budget** the host can change
  (L-12).
- **Meaning without colour** (L-8): an alert's severity is a word and a digit as well as a colour. Not every
  layer is free of colour yet: a quake's age, a fire's strength and a buoy against a tide station rest on
  colour, and a host carries them in the label until a library mark comes (OW-17).
- **Alerts over radar** (L-11), **detail by purpose** (L-14), **the moment** - overlays drawn only
  during their span (L-15).
- **New overlay kinds and presets**: wind with gusts on the arrows (L-16, L-24), rain as a grid with
  marks (L-17), fire (L-18), quakes as USGS draws them (L-19), waves (L-20), the sea's stations
  (L-21), UV and air quality in their official scales (L-25), rain and snow totals (L-26).
- **The map reads well under load**: the basemap always shows (L-22), place names come before data
  (L-23) and stand still as a loop plays (L-28), a render keeps an overlay's preparation (L-30), and
  a blink redraws only what blinks (L-29).
- **New warning kinds** in the closed list: `table-fallback`, `cache-root-readable`,
  `near-image-colours`.

## Changed from v0.1.0 (ruled breaks and deviations)

The breaks, from contract section 12:

- **The borrow check is removed**, with the warning kind `BorrowChanged`; the kinds after it renumber,
  so compare kinds by name (D-74).
- **`Changed()` counts inputs only**: `Render`, `Animate` and `FollowClock` no longer raise it; frame
  advances have their own counter, `FrameTicks()` (D-66).
- **Loop playback is one per map**, its position a valid time, on or off at a step the host sets
  (D-67, D-76).
- **`Describe` is removed**, with `Description`; `Report` gives every answer it gave (D-70).
- **`Purge` empties every source's tiles**, on disk and in memory, and the decoded pictures kept only to
  be used again, and returns a `PurgeReport` beside the error (L-9.3, D-70); what an overlay shows now
  stays.
- **`Fetcher` is removed**, with `Map.Fetcher`; a host's transport, user-agent and timeout go through
  `SetFetchOptions` (D-62).
- **`RoadLayer` switches the major roads only**; `MinorRoadLayer` switches the rest (D-82).
- **New names are setters only**, with no paired options; nothing to change, since the names are new
  (D-70).

Where BUILD departed from what DISCOVER and PLAN set, beyond the breaks above:

- **Severity without colour is a digit along the outline** (4, 3, 2, 1, `?`), not a dash (D-65).
- **No seek**: playback is driven by `Play`, `Stop`, `Reset` and `Step` (D-67).
- **The image budget's default is 6 MiB, not 3 MiB**, and `MaxFrames` is 72 (D-68).
- **A playback step below 400 ms lifts the flash ceiling**, by the host's choice; `Loop` says so
  (D-76).
- **A vector-tile polygon ring of zero area is a hole** of the polygon before it, not a polygon of its
  own; the change is invisible in rendering (D-106).

## What a host should know

- **Loops do not play until the host says so.** Playback is off by default: the map shows the newest
  observed frame. A host turns it on with `SetPlayback(PlaybackOn)` and starts it with `Play`.
- **Say "playing" from `Advancing`, not `Playing`.** `Loop()`'s `Playing` is play pressed with playback
  on; `Advancing` is the frames moving, false while the animation clock is held still (D-109).
- **Every loop has a motion entry.** One with no motion to tell says why in `Missing`: no heavier rain,
  too few frames, or frames still being read (D-111).
- **A pump keeps going past a failed job.** `Work` returns a failed job's error with `did` true; only
  `closed` and `cancelled` end a pump (contract section 2).
- **Large alerts keep their cues.** An alert drawn straight from the host's memory carries its severity
  word and digits like any other (D-108).
- **A blend no class would show is not drawn.** Where every class would stay within 5 of itself, the
  image is drawn over the alert's tint (D-114).
- **A refreshed loop keeps drawing.** A loop handed in again draws from its old frames until its new
  ones are decoded, then from the new ones (L-27).
- **MRMS's heavy end is unverified.** MRMS publishes no colour table; its table was built from the
  colours MRMS was seen to use, on days that reached about 48.5 dBZ. Above about 48.5 dBZ a colour is
  valued along the legend's gradient, and every pixel valued that way is counted and warned of as
  `table-fallback`. A colour farther than 30 from that heavy-end gradient draws nothing and is warned of
  as `unmatched-image-colours`: **on MRMS, a host treats that warning as possible severe rain**, since the
  heaviest cores may be what drew blank (D-113). MRMS's table is marked `Unverified` until a severe day's
  capture is ruled to cover that range (L-2.3, OW-12).
- **A palette that moves is said.** A pixel matched only within the tolerance of a published or host
  table is warned of as `near-image-colours` (L-2.2). MRMS's legend table expects near colours and is
  exempt.
- **`Report`'s stale marks need a clock.** They are judged by the wall clock the last `Render` passed, and
  with none given nothing is judged stale; `MotionReport.Stale` follows the same rule (D-130). A host that
  reports before it renders renders once first.
- **`no-work-called` is sent.** A host whose renders keep finding work pending with no `Work` run is warned
  once (D-131).
- **A cache root the running user does not own is refused**, and on macOS one that carries an access
  control list, with a message naming the fix (D-134).
- **A place's id is held to the overlay id rule**: at most 256 bytes of plain text, or the place is
  refused as `invalid-id`. Names are cleaned, not capped.
- **The module floor is go 1.25.13** (D-133): at go 1.25.0 the code reaches standard-library
  vulnerabilities that 1.25.13 fixes.
- **The pump rule is unchanged**; the example pumps (`examples/example_pump_test.go` and the demo,
  `cmd/tuimaps`) keep going past a failed job, as the rule says.
- **v0.1.0 defects fixed** (L-13), and an image is answered in its own projection: a web Mercator
  image's answers and motion sightings were off in latitude in v0.1.0 (F-3).

## Measured

- **M1, motion seen:** sitting 1 (2026-10-04) - visual arm 5/5, pass; non-visual arm 3/5, **fail**.
  D-122 changed the motion, and D-123 and D-127 pin it on the same five loops by test; that pin is not a
  second sitting, and a blind sitting on fresh loops is owed when a host voices motion (OW-21).
- **M2, loop honesty:** held by tests - the frame time and "gap" on the map, staleness by the newest
  observed frame (`TestTheFrameTimeIsOnTheMap`, `TestTheNewestObservedFrameDrivesStale`).
- **M3, the bound holds:** held by tests on every path (`TestTheBoundIsHeldOnEveryPath`).
- **M4, embed cost with a loop:** a 12-frame loop at dot resolution adds about 0.6 MiB of heap over an
  empty map (the line is 1 MiB), and holds within ±5 % while it plays and is refreshed once a minute
  (the hour soak, 2026-10-04: 60 refreshes, the heap from 1590 KB to at most 1633 KB, one machine); 24
  region-sized frames add about 5.8 MiB (D-68).
- **L-12.5, a frame advance:** 4.4-4.6 ms on a near-empty map and 7.7-8.2 ms on a dense one (a region at
  zoom 6, a noisy twelve-frame loop, 1,000 alert areas, 20 places), against 15 ms; the reference machine,
  five runs each.
- **M5, contract truth:** held by tests both ways, the contract against the package and the README's
  promises against the code.
- **M6, gate trust:** five consecutive clean full runs, counted from `06_docs/gate-runs.md` (`m6_test.go`);
  a failed run resets the count, as one at VALIDATE did (D-140). The hosted full gate is green on linux/amd64 and linux/arm64
  (run 37222838055, its fuzz budgets divided by 8).
- The vulnerability scan (`govulncheck` v1.8.0, pinned) finds nothing reachable, checked at the release
  commit at the module's floor toolchain and at the local one.
