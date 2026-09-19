# Level 3 — State machines

Up: [architecture](architecture.md) · Carries: FR-11, FR-22a, FR-23, FR-25, FR-26, FR-32, NFR-21, D-30, D-56, D-73, D-74

## 1 · A tile

```mermaid
stateDiagram-v2
    [*] --> Wanted: Render notes it is missing
    Wanted --> Queued: joins pending work (capped, de-duplicated)
    Queued --> Dropped: left the view, or the cap dropped the oldest
    Queued --> Loading: a Work call picks it up (D-73)
    Loading --> Dropped: context cancelled — it left the view
    Loading --> OnHand: passed the gate (NFR-10)
    Loading --> Waiting: refused, failed, or no source had it
    Waiting --> Queued: its not-before time has passed and it is still wanted
    Waiting --> Dropped: no longer wanted
    OnHand --> Evicted: memory cache over its byte cap
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
    end note
```

## 2 · Borrowed geometry (FR-11, D-74)

The one dangerous part of the contract: the host's memory, read by the library. The host must not change or reuse it until the library says the borrow is over.

```mermaid
stateDiagram-v2
    [*] --> Borrowed: Set(features) accepted
    Borrowed --> Borrowed: read by simplify, draw-from-borrowed, describe
    Borrowed --> Replacing: Set(same id) or Remove(id)
    Replacing --> Released: no library job still reads the old geometry
    Released --> [*]: the host may now change or reuse that memory
    note right of Replacing
      Cancellation is cooperative, so a job may
      still be reading for a moment. The old overlay
      keeps drawing until the new one is prepared.
      The result of Set and Remove carries the signal
      that the borrow has ended — designed to be
      hard to ignore (D-74).
    end note
    note left of Borrowed
      A race-detector test reuses the memory
      at once and must fail.
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
| How the end of a borrow is signalled — settled in the implementation plan | Machine 2 |
| Flash and pulse are built; camera tours arrive | Machine 3 gains the camera's own machine beside it |
| Image loops are built (FR-37) | Machine 4 gains a per-frame valid time |
