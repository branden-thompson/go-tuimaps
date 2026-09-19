# Level 3 — State machines

Up: [architecture](architecture.md) · Carries: FR-11, FR-23, FR-25, FR-26, FR-32, FR-37, NFR-10, NFR-21, D-30, D-56, D-73, D-74, D-86, D-90, D-92, P-59a

## 1 · A tile

```mermaid
stateDiagram-v2
    [*] --> Wanted: Render notes it is missing
    Wanted --> Queued: joins pending work (capped, de-duplicated)
    Queued --> Dropped: left the view, or the cap dropped the oldest
    Queued --> Loading: a Work call picks it up (D-73)
    Loading --> Dropped: context cancelled — it left the view
    Loading --> OnHand: passed the gate (NFR-10)
    Loading --> Unavailable: NO source is named, and the embedded tiles do not hold it — nothing was tried that could fail
    Loading --> Waiting: a NAMED source refused, failed, or answered that it has no such tile
    Waiting --> Queued: its not-before time has passed and it is still wanted
    Waiting --> Dropped: no longer wanted
    Unavailable --> Wanted: a source is named, or the assets are registered
    OnHand --> Evicted: NO live view of any map draws it, and the cache is over its cap (D-90)
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

The one dangerous part of the contract: the host's memory, read by the library. **Every reader runs inside a call the host itself made** — `Work` (simplifying, describing) or `Render` (drawing a very large shape directly) — so the end of a borrow is **reported to the host, never waited for** (D-86). `Render`, `Set` and `Remove` are owner calls, one at a time (contract, section 6), so a `Render` is never still reading when a `Set` arrives: only a `Work` can be the last reader.

```mermaid
stateDiagram-v2
    [*] --> Borrowed: Set(features) accepted
    Borrowed --> Borrowed: read inside the host's own Work and Render calls
    Borrowed --> ReleasedAtOnce: Set(same id) or Remove(id) with no reader in flight
    Borrowed --> Draining: Set(same id) or Remove(id) while a Work on another goroutine is reading
    Draining --> Released: that Work call — or the Settle running it — returns and says so
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
      Opt-in borrow check: fingerprint at hand-in,
      re-checked at every read. A change is a warning
      that names the overlay.
    end note
```

## 3 · A marker's animation (FR-26, D-56, NFR-21)

```mermaid
stateDiagram-v2
    [*] --> Steady: static marker, or reduce-motion on
    [*] --> Animated: blink (v0.1.0) · flash and pulse (later)
    Animated --> Steady: reduce-motion switched on
    Steady --> Animated: reduce-motion switched off
    state Animated {
        [*] --> On
        On --> Off: phase computed from the time the host supplies
        Off --> On: never from a count of calls (FR-25)
    }
    note right of Animated
      Blink: upstream's 800 ms period (P-59a).
      Flash: no faster than 2.5 a second (D-56).
      A frozen clock never leaves a marker hidden:
      if time does not advance, the marker is drawn On.
    end note
```

## 4 · An overlay's freshness (FR-32)

```mermaid
stateDiagram-v2
    [*] --> Current: valid time given · current-for between 0 and 7 days
    [*] --> Uncertain: valid time more than 5 minutes in the future
    Current --> Stale: wall clock passes valid time + current-for
    Uncertain --> Stale: treated as stale at once
    Stale --> Current: the host sets fresher data
    note right of Stale
      Marked on the frame, stated in the legend data
      and in the description. Uses the wall clock,
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
