# Level 2 — The view described as data

Up: [architecture](architecture.md) · Carries: FR-29, FR-5, FR-32, FR-33, FR-34, D-52, D-67, D-68, D-73

## Why it exists

A screen reader reads braille map characters as noise, and no picture can show a distance smaller than one cell (at 69×12, a column is 1.6 km). The description is the exact answer to the project's question — *where is it, relative to my place* — computed from the real shapes and values, never from the drawn cells. It is **M1b**: it must match the answer key exactly (D-67).

## How it is computed

```mermaid
flowchart TB
    ASK["Describe(places)"] --> MEMO{"Anything changed since last time?<br/>overlays · places · staleness · colour depth"}
    MEMO -- no --> SAME["Return the same data — no allocation"]
    MEMO -- yes --> Q[("Pending work")]
    Q -- "host calls Work (D-73) — never inside Render" --> PER

    subgraph PER["For each place × each overlay"]
      direction TB
      A["<b>Areas</b> (polygons, circles)<br/>inside or outside — even-odd within a polygon, a feature is the union of its polygons (FR-11)<br/>distance and compass bearing to the nearest edge · longitude is circular, a ±180° seam is never an edge"]
      B["<b>Points and lines</b><br/>distance and bearing to the nearest, with its label"]
      C["<b>Fields</b><br/>value with its unit · its band · which way it rises — from the data, exact even on a flat day (D-68)"]
      D["<b>Images</b><br/>the class here · the nearest heavier class, with distance and bearing"]
      E["<b>Every overlay</b><br/>valid time · stale or not · the form it is drawn in · 'no data here', said plainly"]
    end

    PER --> GEO["Great-circle distances · units follow the host's setting and are always stated"]
    GEO --> CLEAN["Every string cleaned on the way out (FR-34)<br/>no braille, block or box characters · no abbreviations a speech engine would spell out"]
    CLEAN --> DATA["Plain data, not sentences — the host renders or speaks it (D-52)"]
    DATA --> HOST["Host: a line beside the map · its own speech"]
    DATA --> APP["App: describe mode prints it as text, with no map (FR-5)"]
```

## Picture and description cannot disagree

```mermaid
flowchart LR
    SHAPE[("Borrowed geometry<br/>one copy, the host's")] --> SIMP["Simplified for the view"] --> PIC["The picture<br/>(right to within a cell)"]
    SHAPE --> EXACT["Used unsimplified"] --> TXT["The description<br/>(exact)"]
    RULE["One fill rule, one longitude convention,<br/>one projection package"] -.-> SIMP
    RULE -.-> EXACT
    KEY["Answer key — a separate script over the same data (D-43, D-67)"] -. "M1a: the frame must not contradict it,<br/>to the frame's resolution ('on the edge' under one cell)" .-> PIC
    KEY -. "M1b: must match exactly" .-> TXT
```

## Cost rules (FR-29)

| Rule | Provisional target, unverified |
|---|---|
| Never computed inside Render | — |
| Recomputed only when overlays, places, staleness or depth change | An unchanged repeat allocates nothing |
| Cancellable; any index it builds counts against the shape cache's byte cap | The fixture with 60 places ≤ 50 ms; the worst case ≤ 500 ms |
