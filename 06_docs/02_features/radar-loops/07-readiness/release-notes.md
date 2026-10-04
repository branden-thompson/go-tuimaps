# go-tuiMaps v0.2.0 — release notes

Up: [release checklist](release-checklist.md) · [requirements](../01-objectives/requirements.md)

**v0.2.0 adds to v0.1.0's contract, and breaks it only where ruled** (D-58): each break is listed
below, and in section 12 of the contract (`06_docs/02_features/go-tuimaps/03-architecture-design/contract.md`)
with what a host does instead. It is the library's first reviewed release (L-5.4).

## What it adds

- **Radar loops** (L-1): an image overlay may carry frames, each with its valid time; the host plays,
  stops, steps and resets them through one playback API, and the map draws the frame for its moment.
  A refreshed loop keeps drawing while its new frames are prepared (L-27).
- **Storm motion in words** (L-1.12): `Report` says where the heavier rain was at the oldest usable
  frame and where it is at the newest - distance, direction, time - and whether it came closer, moved
  away or held. Observation, never forecast.
- **Provider colour tables** (L-2, L-2.5): IEM's published table and MRMS's observed one, through one
  seam a third provider joins the same way.
- **A view bound the host sets** (L-3), **a fetcher the host may supply** (L-7), **tile-host
  confinement** (L-10), **cache age and purge** (L-9), and **an image budget** the host can change
  (L-12).
- **Meaning without colour** (L-8): patterns and digits so no layer depends on colour alone.
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

Where BUILD departed from what DISCOVER and PLAN set:

- **Severity without colour is a digit along the outline** (4, 3, 2, 1, `?`), not a dash (D-65).
- **No seek**: playback is driven by `Play`, `Stop`, `Reset` and `Step` (D-67).
- **The image budget's default is 6 MiB, not 3 MiB**, and `MaxFrames` is 72 (D-68).
- **`Describe` is removed** in favour of `Report` (D-70).
- **The borrow check is dropped** (D-74).
- **A playback step below 400 ms lifts the flash ceiling**, by the host's choice; `Loop` says so
  (D-76).
- **A vector-tile polygon ring of zero area is a hole** of the polygon before it, not a polygon of its
  own; the change is invisible in rendering (D-106).

## What a host should know

- **Loops do not play until the host says so.** Playback is off by default: the map shows the newest
  observed frame. A host turns it on with `SetPlayback(PlaybackOn)` and starts it with `Play`.
- **A refreshed loop keeps drawing.** A loop handed in again draws from its old frames until its new
  ones are decoded, then from the new ones (L-27).
- **MRMS's heavy end is unverified.** MRMS publishes no colour table; its table was built from the
  colours MRMS was seen to use, on days that reached about 48.5 dBZ. Above about 48.5 dBZ a colour is
  valued along the legend's gradient, and every pixel valued that way is counted and warned of as
  `table-fallback`. MRMS's table is marked `Unverified` until a severe day's capture is ruled to cover
  that range (L-2.3, OW-12).
- **A palette that moves is said.** A pixel matched only within the tolerance of a published or host
  table is warned of as `near-image-colours` (L-2.2). MRMS's legend table expects near colours and is
  exempt.
- **v0.1.0 defects fixed** (L-13), and an image is now answered in its own projection: a web Mercator
  image's answers and motion sightings were off in latitude (F-3).

## Measured

- M4: a 12-frame loop at dot resolution adds about 0.6 MiB of heap over an empty map (the line is
  1 MiB), and holds within ±5 % while it plays; 24 region-sized frames add about 5.8 MiB (D-68).
- The vulnerability scan (`govulncheck` v1.8.0, pinned) finds nothing reachable.
