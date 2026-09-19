# Level 2 — Render and compositing

Up: [architecture](architecture.md) · Carries: FR-4, FR-12, FR-19, FR-23, FR-34, FR-36, NFR-4, NFR-6, NFR-8, D-35, D-42, D-57, D-64

## The rule this diagram protects

**Render reads only what is already on hand, and never waits** (FR-23). It is a pure function of its inputs (NFR-6): the same view, size, depth, palette, time, overlays and tiles give the same bytes on any machine. An unchanged frame costs nothing (NFR-4).

## From a call to a frame

```mermaid
flowchart TB
    CALL["Render(rectangle, now)"] --> SAME{"Same inputs as last time?<br/>view · size · depth · palette ·<br/>overlay versions · tile set · animation phase"}
    SAME -- yes --> REUSE["Return the previous frame<br/>zero allocations (NFR-4)"]
    SAME -- no --> SNAP["Take a snapshot of what is on hand<br/>tiles in a fixed order (NFR-6) · prepared overlays · markers"]
    SNAP --> NEED["Note what is missing → pending work<br/>(never fetched here; the host's Work does it — D-73)"]
    SNAP --> PROFILE["Choose the basemap profile (FR-19)<br/>by map size · by what is on top · by layers switched off (FR-36)"]
    PROFILE --> PAINT

    subgraph PAINT["Paint the cell grid — each cell is glyph + foreground + background"]
      direction TB
      L0["1 Ground<br/>painted from the ground token, or left to the terminal if the host declared it (D-64)"]
      L1["2 Water<br/>masks a scalar FIELD: temperature stops at the shore (D-32)<br/>never masks an IMAGE: rain shows over water, the coast drawn as an outline (D-87)<br/>a host can flip either, per overlay"]
      L2["3 Images and fields<br/>cell background by class → colour (L2-colour)<br/>fields skip water cells; images do not (D-87)"]
      L3["4 Area tints<br/>alert polygons' interior; hatch when there is no colour (FR-18a)"]
      L4["5 Basemap lines<br/>braille dots, 2×4 per cell: coast, borders, roads, rivers, parks"]
      L5["6 Overlay lines and outlines<br/>polygon outlines always drawn (FR-16), contours (D-35), tracks"]
      L6["7 Glyphs<br/>markers over everything beneath — never under a hatch (FR-18a); focus indicator (FR-24a)"]
      L7["8 Labels<br/>place names, value labels, edge labels — collision-checked, cleaned, by grapheme cluster (FR-34)"]
      L8["9 Furniture<br/>scale mark (FR-33) · stale mark (FR-32) · credit line (FR-14) · the no-tiles notice (FR-23)"]
      L0 --> L1 --> L2 --> L3 --> L4 --> L5 --> L6 --> L7 --> L8
    end

    PAINT --> FG["Choose each cell's foreground<br/>compute both contrasts, take the higher (FR-16)"]
    FG --> DEPTH["Map colours to the depth in use<br/>truecolor · 256 (indices 16–255 only) · 16 · none (L2-colour)"]
    DEPTH --> EMIT["Emit lines<br/>each exactly the requested width (NFR-8)<br/>colour sequences and cleaned text, nothing else (FR-34)"]
    EMIT --> STATUS["Frame + status<br/>complete · still sharpening · no tiles"]
```

## The braille canvas

```mermaid
flowchart LR
    G["Geometry in tile units<br/>small integers (D-75)"] --> P["Project to dot space<br/>2 dots across × 4 down per cell"]
    P --> C["Clip to the rectangle<br/>nothing is ever drawn outside it (FR-11)"]
    C --> RAS["Rasterise lines and fills into a dot mask"]
    RAS --> CELL["Per cell: 8 dots → one braille character<br/>U+2800 + mask; an empty cell is U+2800 (P-03a)"]
    CELL --> COLR["Per cell: one foreground colour<br/>the MAJORITY colour among its lit dots;<br/>a tie goes to the colour commoner in the 8 neighbouring cells (P-08, D-83)"]
```

**Why one colour per cell matters.** A terminal cell has one foreground and one background. Two lines of different colours crossing one cell cannot both keep their colour; within the basemap upstream's majority vote decides (P-08, D-83), and across layers the compositing order does. This is the reason the basemap thins under overlays (FR-19) rather than competing with them.

## What can change this diagram

| If this changes… | …this part moves |
|---|---|
| The compositing order (FR-12) | The numbered list in "Paint the cell grid" |
| The cell-colour rule (P-08, D-83) | The last box of "The braille canvas" |
| The block renderer is built (after v0.1.0, D-42) | A second canvas beside the braille one; steps 5 to 7 use it; its own sparser profile |
| The ground ruling (D-64) | Step 1 |
| A new kind of furniture | Step 9, and NFR-8's closed list of characters |
