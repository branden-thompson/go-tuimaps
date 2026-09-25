# Level 2 — The view described as data

Up: [architecture](architecture.md) · Carries: FR-5, FR-11, FR-29, FR-34, D-43, D-52, D-67, D-68, D-73

## Why it exists

A screen reader reads braille map characters as noise, and no picture can show a distance smaller than one cell (in specimen 13a at 69×12 a column is 1.6 km and a row 3.2 km). The description is the exact answer to the project's question — *where is it, relative to my place* — computed from the real shapes and values, never from the drawn cells. It is **M1b**: it must match the answer key exactly (D-67).

## How it is computed

*AS BUILT v0.2.0 (rc.5): `Describe` and `Description` are removed; `Report(places)` replaces them and carries every answer (D-57, L5.2–L5.8). It is worked out inside the call, from the host's own geometry, and needs no `Work` — except that an image's answer and observed motion read pictures a `Work` call has decoded.*

```mermaid
flowchart TB
    ASK["Report(places) — the places given, or the map's own"] --> MEMO{"Anything changed since last time?<br/>overlays · places · places asked · units · staleness · nearby distance · view · work landed"}
    MEMO -- no --> SAME["Return the same report — no allocation (FR-29)"]
    MEMO -- yes --> NOW["Worked out inside the call, under the map's lock — never inside Render, never queued as work"]
    NOW --> SHOWN
    NOW --> PER
    NOW --> MOT

    SHOWN["<b>Alerts</b>: every alert in view — overlay, feature id, label, severity, valid time, expiry, stale or not (L-13.5)"]

    subgraph PER["<b>Places</b>: for each place"]
      direction TB
      A["<b>Each alert on its own</b> (L-13.6)<br/>inside, nearby (10 km by default, SetNearby) or outside — even-odd within a polygon, a feature is the union of its polygons (FR-11)<br/>distance and compass bearing to the nearest edge · longitude is circular, a ±180° seam is never an edge · under one cell or not (D-67)"]
      B["<b>Points and lines</b><br/>distance and bearing to the nearest, with its label"]
      C["<b>Fields</b><br/>value with its unit · its band · which way it rises — from the data, exact even on a flat day (D-68)"]
      D["<b>Images</b><br/>the class here — its words are the legend's, for the host to use (L5.7) · the nearest heavier class, with distance and bearing ·<br/>read from the newest observed frame; 'no data' until a Work call has decoded it"]
      E["<b>Every answer</b><br/>valid time · stale or not · the form it is drawn in · 'no data here', said plainly"]
    end

    MOT["<b>Motion</b>, for each loop (L-1.12, D-42)<br/>where the heavier rain was at the oldest decoded observed frame and where it is at the newest,<br/>from each place asked — or the view's centre — under one threshold held for the loop:<br/>closer, away or held, over the span · observation only, never a forecast"]

    SHOWN --> GEO
    PER --> GEO
    MOT --> GEO
    GEO["Great-circle distances · units follow Units(miles, fahrenheit) and are always stated"]
    GEO --> CLEAN["Every string cleaned on the way out (FR-34)<br/>no braille, block or box characters · no abbreviations a speech engine would spell out"]
    CLEAN --> DATA["Plain data, not sentences — the host words it or speaks it, and never says 'you' of a place (D-52, D-29)"]
    DATA --> HOST["Host: a line beside the map · its own speech"]
    DATA --> APP["App: describe mode words it as text, with no map (FR-5)"]
```

## Picture and description cannot disagree

```mermaid
flowchart LR
    SHAPE[("Borrowed geometry<br/>one copy, the host's")] --> SIMP["Simplified for the view"] --> PIC["The picture<br/>(right to within a cell)"]
    SHAPE --> EXACT["Used unsimplified"] --> TXT["Report<br/>(exact)"]
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
