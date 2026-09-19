# Level 2 — Style, profiles, the schema seam, and labels

Up: [architecture](architecture.md) · Carries: FR-19, FR-34, FR-35, FR-36, NFR-8, D-46, D-64, D-75, D-82, D-83, P-34, P-36, L-9

## From a tile's layers to what is drawn

```mermaid
flowchart LR
    SRC["Tile source declares its layers<br/>(TileJSON, or the embedded set's manifest)"] --> MAP{"Schema mapping (FR-35)<br/>does a known mapping cover these layers?"}
    MAP -- no --> UNS["Error: unsupported-schema (D-46)"]
    MAP -- "yes: OpenMapTiles" --> ROLES["Layer + class — for a boundary, its administrative level and whether it is at sea — → a ROLE<br/>coast · water · river · border.country · border.region · road.major · road.minor · rail · park · runway · place"]
    ROLES --> DEC["The decoder keeps only layers and attributes some role needs —<br/>everything else is dropped while decoding (D-75, D-82)"]
    ROLES --> STYLE{"Style"}
    STYLE -- "built in" --> BI["dark or bright, chosen by the ground's luminance (D-64)"]
    STYLE -- "a user's file" --> UF["Same JSON format · every zoom stop honoured (L-9) · legacy filters only in v0.1.0<br/>parsed under the untrusted-input limits"]
    BI --> PROF
    UF --> PROF
    PROF["Profile: which roles are drawn, and how heavily (FR-19)"] --> OUT["Roles to draw, each with a colour token and a priority"]
```

## Choosing the profile — what the basemap gives up, and in what order

**This is not a navigation tool (D-83): roads go first; coast, water, borders and terrain go last.**

```mermaid
flowchart TB
    Q1{"Is anything drawn on top?"} -- "a field or an image" --> P2["Overlay profile:<br/>minor roads off · major roads thin · labels fewer"]
    Q1 -- "wind (later)" --> P3["Sparse profile: roads off"]
    Q1 -- "nothing, or features only" --> P1["Full profile"]
    P1 --> Q2
    P2 --> Q2
    P3 --> Q2
    Q2{"Map smaller than 80×20?"} -- yes --> S1["Small-map step: drop the next thing in the order below"]
    Q2 -- no --> S0["As chosen"]
    S1 --> L
    S0 --> L
    L["Host's layer switches applied last (FR-36): a layer the host turned off is never drawn"]
```

| Given up first → last | Minor roads · major roads · rail · parks · region borders · place labels beyond the largest · rivers · country borders · water · coast |
|---|---|

## Placing labels

```mermaid
flowchart LR
    C["Candidates: places on hand, largest rank first<br/>name by upstream's order with one language kept (P-36, D-82): the configured language — English unless the host says otherwise — then the local name, then the house number. No other language survives decoding"] --> CL["Cleaned and measured by grapheme cluster (FR-34, NFR-8)<br/>a wide character takes two cells"]
    CL --> FIT{"Fits inside the rectangle?"}
    FIT -- no --> SKIP["Skipped whole — never cut mid-cluster"]
    FIT -- yes --> COL{"Collides with a label already placed?<br/>upstream's rule: a padded box in cell units (P-34)"}
    COL -- yes --> SKIP
    COL -- no --> PUT["Placed · its cells locked against lines"]
    PUT --> CAP{"Profile's label budget reached?"}
    CAP -- no --> C
    CAP -- yes --> DONE["Done"]
    OVL["Overlay labels and markers are placed BEFORE basemap labels,<br/>so a place name never hides a warning's word"] -.-> COL
```

## What can change this diagram

| If this changes… | …this moves |
|---|---|
| The provider changes its schema | Only the mapping box and the role table |
| Style expressions arrive (after v0.1.0) | The user's-file branch. Until then a style that uses them is refused as `malformed-style`, saying so (D-102) |
| What a filter can ask about | `$type`, `class`, `name`, `rank`, `admin_level`, `maritime` — what the decoder keeps. `==` on any other key is false, as upstream has it for a key a feature lacks (P-42) |
| The block renderer is built | A second, sparser set of profiles |
