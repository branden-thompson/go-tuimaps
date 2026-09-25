# Level 3 — State machines

Up: [architecture](architecture.md) · Carries: FR-11, FR-23, FR-25, FR-26, FR-32, FR-37, NFR-10, NFR-21, D-30, D-56, D-73, D-74, D-86, D-90, D-92, P-59a

## 1 · A tile

*AS BUILT v0.2.0 (rc.8): `Purge` also takes a fetched tile out of memory, and a disk file past its maximum age is fetched again rather than served (D-56).*

```mermaid
stateDiagram-v2
    [*] --> Wanted: Render or Settle notes it is missing
    Wanted --> Queued: joins pending work (capped, de-duplicated)
    Queued --> Dropped: left the view, or the cap dropped the oldest
    Queued --> Loading: a Work call picks it up (D-73)
    Loading --> Dropped: context cancelled — it left the view
    Loading --> OnHand: passed the gate (NFR-10)
    Loading --> Unavailable: NO source is named, and the embedded tiles do not hold it — nothing was tried that could fail
    Loading --> Waiting: a NAMED source refused, failed, or answered that it has no such tile
    Waiting --> Queued: its not-before time has passed and it is still wanted
    Waiting --> Dropped: no longer wanted
    Unavailable --> Wanted: a source is named, or embedded tiles are passed
    OnHand --> Evicted: NO live view of any map draws it, and the cache is over its cap (D-90)
    OnHand --> Evicted: Purge() — every fetched tile, needed or not · embedded tiles stay (L-9.3)
    Evicted --> Wanted: wanted again
    Dropped --> [*]
    note right of Waiting
      30 s, doubling to 10 min (FR-23).
      No timer: the time is part of the
      call-me-by deadline (FR-25).
    end note
    note right of OnHand
      While not OnHand, the nearest ancestor
      on hand is drawn as a stand-in (D-30).
      A tile a live view draws is never evicted,
      even over the cap (D-90). Unavailable: not retried on a
      timer, so an offline map reports nothing due.
    end note
```

## 2 · Borrowed geometry (FR-11, D-74, D-86)

The one dangerous part of the contract: the host's memory, read by the library. **Every reader runs inside a call the host itself made** — `Work` (simplifying, describing) or `Render` (drawing a very large shape directly) — so the end of a borrow is **reported to the host, never waited for** (D-86). *AS BUILT (v0.1.0, unchanged in v0.2.0), corrected against the code: `Work` and `Settle` return no released ids; after "not yet", `InUse(id)` is how the host learns the borrow is over; and the opt-in borrow check was never built.* `Render`, `Set` and `Remove` are owner calls, one at a time (contract, section 6), so a `Render` is never still reading when a `Set` arrives: only a `Work` can be the last reader.

```mermaid
stateDiagram-v2
    [*] --> Borrowed: Set(features) accepted
    Borrowed --> Borrowed: read inside the host's own Work and Render calls
    Borrowed --> ReleasedAtOnce: Set(same id) or Remove(id) with no reader in flight
    Borrowed --> Draining: Set(same id) or Remove(id) while a Work on another goroutine is reading
    Draining --> Released: that Work call — or the Settle running it — returns, and InUse(id) now says no
    ReleasedAtOnce --> [*]: the host may reuse the memory
    Released --> [*]: the host may reuse the memory
    note right of ReleasedAtOnce
      Set and Remove never block. They return
      "released: yes". With one goroutine making
      every call this is the only path there is.
    end note
    note right of Draining
      Set and Remove return "released: not yet".
      Queued jobs for the old geometry are dropped
      at once. The job in flight is cancelled and
      stops at its next check. A query answers
      "still in use?" at any time. Nothing fires
      later on its own.
    end note
    note left of Borrowed
      The old shape keeps drawing from the library's
      OWN simplified copy until the new one is ready.
      A very large shape has no such copy: Set indexes
      it in one pass, so the new one draws next frame
      and the old is released at once (D-92).
      Close lets go of every borrow at once, and what a
      call still inside reads goes when it returns.
      NOT BUILT: the opt-in borrow check PLAN named
      (a fingerprint re-checked at every read).
    end note
```

## 3 · A marker's animation (FR-26, D-56, NFR-21)

*AS BUILT v0.2.0 (rc.3): while a loop plays, the blink's half-period is put on the loop's step, the smallest multiple of it at least half the blink's own period, so every blink change falls on a frame advance (D-76). Flash and pulse are not built.*

```mermaid
stateDiagram-v2
    [*] --> Steady: static marker, or reduce-motion on
    [*] --> Animated: blink · flash and pulse not built
    Animated --> Steady: reduce-motion switched on
    Steady --> Animated: reduce-motion switched off
    state Animated {
        [*] --> On
        On --> Off: phase computed from the time the host supplies
        Off --> On: never from a count of calls (FR-25)
    }
    note right of Animated
      Blink: upstream's 800 ms period (P-59a),
      on the loop's step while a loop plays (D-76).
      A host holding the clock (Animate) gets no
      blink deadline from NextCall.
      A frozen clock never leaves a marker hidden:
      if time does not advance, the marker is drawn On.
    end note
```

## 4 · An overlay's freshness (FR-32)

*AS BUILT v0.2.0 (rc.3): a loop's freshness is its newest observed frame's — never the frame shown, a gap or a forecast (L-1.3, D-67); an observed frame dated more than five minutes past the map's clock is refused at `Set`, since only a forecast may be in the future.*

```mermaid
stateDiagram-v2
    [*] --> Current: valid time given · current-for between 0 and 7 days · a loop: its newest observed frame's time
    [*] --> Uncertain: valid time more than 5 minutes in the future · a loop's observed frame that far ahead is refused at Set instead
    Current --> Stale: wall clock passes valid time + current-for
    Uncertain --> Stale: treated as stale at once
    Stale --> Current: the host sets fresher data
    note right of Stale
      Marked on the frame and stated in Report,
      on each answer and alert (the legend data
      carries no stale mark). Uses the wall clock,
      supplied separately from animation time, so a
      frozen animation clock cannot hide old data.
      The change to Stale is part of the
      call-me-by deadline (FR-25).
    end note
```

## What can change these diagrams

| If this changes… | …this moves |
|---|---|
| The retry times (FR-23) | Machine 1's note |
| How the end of a borrow is reported (D-86) | Machine 2 |
| Flash and pulse are built; camera tours arrive | Machine 3 gains the camera's own machine beside it |
| Image loops are built (FR-37) | Machine 4 gains a per-frame valid time |
