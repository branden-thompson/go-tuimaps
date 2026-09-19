# Level 2 — The view: zoom buckets and fit-to

Up: [architecture](architecture.md) · [constants, section 3](constants.md) · Carries: D-76

## Zoom buckets — one simplified copy of a shape per whole zoom level

```mermaid
flowchart LR
    Z["The view's zoom, say 6.4"] --> B["Bucket = the whole part: 6<br/>serves every zoom from 6 up to, not including, 7"]
    B --> T["Tolerance: half a braille dot AT ZOOM 7 —<br/>the finest zoom the bucket serves, so the shape is never too coarse"]
    T --> HAVE{"Is bucket 6 prepared for this shape?"}
    HAVE -- yes --> DRAW["Draw it"]
    HAVE -- no --> NEAR["Draw the NEAREST prepared bucket meanwhile — a shape never vanishes mid-zoom —<br/>and queue bucket 6 as pending work"]
    NEAR --> PREP["A Work call simplifies the borrowed shape for bucket 6:<br/>iterative, never recursive · rings under three distinct points dropped ·<br/>over a quarter of the shape cache: not cached, drawn from the borrowed shape instead"]
    PREP --> DRAW
```

## Fit-to — put these things in the frame (D-76)

```mermaid
flowchart TB
    IN["FitTo(places, overlay ids, margin in cells)"] --> BOX["Bounding box of everything named —<br/>places by position; overlays by the box recorded at hand-in, so no Work is needed first"]
    BOX --> ONE{"A single point, or an empty box?"}
    ONE -- yes --> KEEP["Centre on it · keep the current zoom"]
    ONE -- no --> ANTI{"Does the box cross ±180°?"}
    ANTI -- yes --> SHORT["Take the short way round: longitude is circular"]
    ANTI -- no --> SPAN
    SHORT --> SPAN["Span in web-mercator units, east-west and north-south"]
    SPAN --> FITZ["The largest zoom at which the span, plus the margin, fits the rectangle —<br/>solved directly, not stepped. A cell is twice as tall as it is wide: 2 dots across, 4 down"]
    FITZ --> CL["Clamped to the zoom range · centre on the box's centre"]
    CL --> OUT["The view changes as any intent does: the counter moves, tiles become wanted"]
```

**M1's guard** — the place and the condition share one frame — is met by calling this, and tested by it (task 14.5).

## What can change this diagram

| If this changes… | …this moves |
|---|---|
| The bucket width or its tolerance | The first diagram and the constants table |
| Fit-to gains a "keep this place centred" form | The second diagram |
