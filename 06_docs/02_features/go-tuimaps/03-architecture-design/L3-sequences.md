# Level 3 — Sequences

Up: [architecture](architecture.md) · Carries: D-73, D-30, D-65, FR-4, FR-23, FR-25, FR-30, NFR-5, NFR-19

The library starts no goroutine (D-73). In every sequence below, anything slow happens inside a `Work` call made by the host.

## 1 · A cold first frame — never blank (D-30, NFR-5: ≤ 3 s to full detail)

```mermaid
sequenceDiagram
    autonumber
    participant H as Host (draws on its tick)
    participant P as Host's pump (its goroutines)
    participant M as Map
    participant T as Tile sources
    H->>M: New(Source(named), assets imported)
    H->>M: Set(alerts), Set(radar image), Set(temperature grid)
    M-->>H: created ×3 · warnings if any
    H->>M: Render(rect, now)
    Note over M: Nothing decoded yet.<br/>Draws ground, markers, the notice — never an empty rectangle (FR-23)
    M-->>H: frame · status "still sharpening" · NextCall = now
    P->>M: Pending()?
    M-->>P: embedded z0–3 tiles · 3 overlays to prepare · 4 network tiles
    par as many as the pump likes — its width, and the memory that costs, are the host's (D-84)
        P->>M: Work(ctx)
        M->>T: embedded tile
        T-->>M: bytes → gate → cache
    and
        P->>M: Work(ctx)
        Note over M: prepare the alerts: simplify, boxes
    and
        P->>M: Work(ctx)
        M->>T: network tile z5 (secure, limits)
        T-->>M: bytes → gate → cache → disk
    end
    M-->>P: each done · change counter +1
    P-->>H: "map changed"
    H->>M: Render(rect, now)
    M-->>H: frame: stand-in from z3 under the overlays
    Note over H,M: …the remaining tiles arrive the same way…
    H->>M: Render(rect, now)
    M-->>H: frame · status "complete"
```

## 2 · A pan — the newest view wins

```mermaid
sequenceDiagram
    autonumber
    participant H as Host
    participant P as Host's pump
    participant M as Map
    H->>M: Pan(east) — an intent, not a key (FR-24)
    H->>M: Render(rect, now)
    Note over M: Tiles on hand are redrawn shifted. The new column uses stand-ins.<br/>Wanted tiles join the queue. jobs for tiles that left the view are cancelled (FR-30)
    M-->>H: frame · "still sharpening"
    H->>M: Pan(east) again, before the first tiles arrive
    Note over M: The queue is re-ordered: newest view first. The cap drops the oldest
    P->>M: Work(ctx) …
    M-->>P: done · counter +1
    H->>M: Render(rect, now)
    M-->>H: frame · "complete"
```

## 3 · An idle host — how it still learns that something is due (FR-25)

```mermaid
sequenceDiagram
    autonumber
    participant H as Host
    participant M as Map
    H->>M: Render(rect, now)
    M-->>H: frame
    H->>M: NextCall()
    Note over M: The earliest of: next visible change (a blink phase) ·<br/>next retry time of a failed tile (FR-23) ·<br/>next change of staleness (FR-32). On the wall clock.
    M-->>H: a deadline, or "nothing due"
    Note over H: Sleeps until the deadline, or until its pump reports a change
    H->>M: Render(rect, later)
    Note over M: Marker phase computed from the time given, never from a count of calls.<br/>With reduce-motion on, markers are steady and no blink deadline is reported (NFR-21)
    M-->>H: frame
    Note over H,M: A host that stops calling stops retrying — the documents say so
```

## 4 · A one-shot render — three calls (NFR-19, FR-4)

```mermaid
sequenceDiagram
    autonumber
    participant C as Caller (a test, the app's headless flag, a script)
    participant M as Map
    C->>M: New(assets imported, no network)
    C->>M: Settle(ctx)
    Note over M: The same loop as a pump, on the caller's goroutine:<br/>Work until nothing is pending, or the context ends.<br/>Work that failed and is waiting to retry is not pending — so Settle always ends (FR-30)
    M-->>C: settled · or context ended · with a count of work that failed
    C->>M: Render(rect, now)
    M-->>C: frame · status "complete" — or marked incomplete, never silently (NFR-20)
```

## What can change these diagrams

| If this changes… | …these move |
|---|---|
| The background-work model (D-73) | All four |
| The integration review finds the pump awkward in the first host (D-60) | Sequences 1 and 3; the examples |
| The never-blank floor (FR-23) | The first Render in sequence 1 |
