# Level 3 — Sequences

Up: [architecture](architecture.md) · Carries: FR-4, FR-23, FR-24, FR-25, FR-30, FR-32, NFR-5, NFR-19, NFR-20, NFR-21, D-30, D-60, D-73, D-84

The library starts no goroutine (D-73). In every sequence below, anything slow happens inside a `Work` call made by the host.

## 1 · A cold first frame — never blank (D-30, NFR-5: ≤ 3 s to full detail)

*AS BUILT v0.2.0 (rc.8), corrected against the code: the source is named by a call after `New`; the embedded tiles are a job of their own, not a fallback after the network; overlays are queued for preparing by the first `Render`; `OnPending` wakes the pump; and a landed job raises `Changed()`, which the host watches.*

```mermaid
sequenceDiagram
    autonumber
    participant H as Host (draws on its tick)
    participant P as Host's pump (its goroutines)
    participant M as Map
    participant T as Tile sources
    H->>M: New(WithSize, Embed(embedded tiles)) · Source(named) · OnPending(wake the pump)
    H->>M: Set(alerts), Set(radar image), Set(temperature grid)
    M-->>H: created ×3 · warnings if any
    H->>M: Render(size, now)
    Note over M: Nothing decoded yet. Notes the wanted tiles, their z3 ancestor, and the three overlays to prepare.<br/>Draws ground, markers, the notice — never an empty rectangle (FR-23)
    M-->>P: OnPending hook: work went from none to some
    M-->>H: frame · status "no tiles"
    par as many as the pump likes — its width, and the memory that costs, are the host's (D-84)
        P->>M: Work(ctx)
        M->>T: the z3 ancestor, from the embedded tiles — a job of its own, the cheapest, so first
        T-->>M: bytes → gate → cache
    and
        P->>M: Work(ctx)
        Note over M: prepare the alerts: simplify, boxes
    and
        P->>M: Work(ctx)
        M->>T: a z5 tile: disk misses, the network answers (confined, secure, limits)
        T-->>M: bytes → gate → cache → disk, if its generation is unchanged
    end
    M-->>P: each Work returns · each that landed something raises Changed()
    P-->>H: "map changed" — Changed() moved
    H->>M: Render(size, now)
    M-->>H: frame: stand-in from z3 under the overlays · "still sharpening"
    Note over H,M: …the remaining tiles arrive the same way…
    H->>M: Render(size, now)
    M-->>H: frame · status "complete"
```

## 2 · A pan — the newest view wins

```mermaid
sequenceDiagram
    autonumber
    participant H as Host
    participant P as Host's pump
    participant M as Map
    H->>M: PanCells(east) — an intent, not a key (FR-24)
    H->>M: Render(size, now)
    Note over M: Tiles on hand are redrawn shifted. The new column uses stand-ins.<br/>Wanted tiles join the queue. jobs for tiles that left the view are cancelled (FR-30)
    M-->>H: frame · "still sharpening"
    H->>M: PanCells(east) again, before the first tiles arrive
    Note over M: The queue is re-ordered: newest view first. The cap drops the oldest
    P->>M: Work(ctx) …
    M-->>P: done · Changed() +1
    H->>M: Render(size, now)
    M-->>H: frame · "complete"
```

## 3 · An idle host — how it still learns that something is due (FR-25)

*AS BUILT v0.2.0 (rc.3): while a loop plays, `NextCall` also reports its next frame advance — the fourth source (L-1.8). A frame advance moves `FrameTicks()`, not `Changed()` (D-66).*

```mermaid
sequenceDiagram
    autonumber
    participant H as Host
    participant M as Map
    H->>M: Render(size, now)
    M-->>H: frame
    H->>M: NextCall(wall clock)
    Note over M: The earliest of: next visible change (a blink phase) ·<br/>while a loop plays, its next frame advance (L-1.8) ·<br/>next retry time of a failed tile (FR-23) ·<br/>next change of staleness (FR-32). On the wall clock.<br/>A host holding the animation clock (Animate) is told of no blink or advance.
    M-->>H: a deadline, or "nothing due"
    Note over H: Sleeps until the deadline, or until its pump reports a change
    H->>M: Render(size, later)
    Note over M: Marker phase and the loop frame shown are computed from the time given, never from a count of calls.<br/>With reduce-motion on, markers are steady, a loop stops where it is, and neither is reported due ·<br/>a stale moment and a retry stay due (NFR-21)
    M-->>H: frame
    Note over H,M: A host that stops calling stops retrying — the documents say so
```

## 4 · A one-shot render — three calls (NFR-19, FR-4)

```mermaid
sequenceDiagram
    autonumber
    participant C as Caller (a test, the app's headless flag, a script)
    participant M as Map
    C->>M: New(WithSize(cols, rows), Embed(embedded tiles)) — no network
    C->>M: Settle(ctx)
    Note over M: Settle first notes what a render of the map's size would want (no size: refused, 'no-size').<br/>Then the same loop as a pump, on the caller's goroutine:<br/>Work until nothing is pending, or the context ends.<br/>Work that failed and is waiting to retry is not pending — so Settle always ends (FR-30)
    M-->>C: SettleResult: jobs run · how many failed · the first failure's why · jobs in flight elsewhere — or the context's error
    C->>M: Render(size, now)
    M-->>C: frame · status "complete" — or marked incomplete, never silently (NFR-20)
```

## What can change these diagrams

| If this changes… | …these move |
|---|---|
| The background-work model (D-73) | All four |
| The integration review finds the pump awkward in the first host (D-60) | Sequences 1 and 3; the examples |
| The never-blank floor (FR-23) | The first Render in sequence 1 |
