# go-tuiMaps — Architecture

| Field | Value |
|---|---|
| Phase | PLAN |
| Date | 2026-09-19 |
| Status | Draft for HUM LEAD's review, revised after the PLAN red-team. Built on rulings D-11 to D-94. HUM LEAD's first read, 2026-09-19: "These look good so far" — a basic understanding, not yet a deep one; formal approval comes with the Plan of Record. |
| How to use this | These diagrams are **living references** (D-71). Point at one when asking a question; when a decision changes, the diagram changes in the same commit. Every diagram file names the rulings and requirements it carries on its first lines, so a change to one of those says which file to open; for the three diagrams in this file, the ids are in the text beside each. |

## Start here

Thirty-four diagrams is a lot to hold. **[A guided tour](architecture-tour.md)** follows one story — a flood warning and a radar picture, from the first host to the cells on the screen and the sentence it can speak — through almost every diagram once, in 29 steps, each naming the diagram to look at and the ruling behind it. It ends with the three diagrams to read if you read no others.

## The documents of PLAN, and the diagram set

Read down for more detail, up for context. Each file's first lines say which rulings and requirements it carries.

| Level | What | File |
|---|---|---|
| — | **A guided tour**: one story through almost every diagram, in 29 steps | [`architecture-tour.md`](architecture-tour.md) |
| 0 | **Context and trust boundary** — the library among its neighbours, and which side of the line each byte comes from | this file |
| 1 | **The parts** — packages, what each owns, what may import what | this file |
| 1 | **The public contract at a glance** | this file |
| 1 | **The contract** — what a host can rely on: the pump, the three-call path, the end of a borrow, the frame, which calls are safe together, shared caches, where each deferred shape lands (2 diagrams) | [`contract.md`](contract.md) |
| 1 | **Constants** — every number a test needs, and the semantic token list | [`constants.md`](constants.md) |
| 2 | Render and compositing (2 diagrams) | [`L2-render.md`](L2-render.md) |
| 2 | Tile pipeline, the network edge, the embedded tiles and their generator (3) | [`L2-tiles.md`](L2-tiles.md) |
| 2 | Overlay pipeline, shape by shape; presets (3) | [`L2-overlays.md`](L2-overlays.md) |
| 2 | Colour resolution; the ground (2) | [`L2-colour.md`](L2-colour.md) |
| 2 | Style, profiles, the schema seam, label placement (3) | [`L2-style.md`](L2-style.md) |
| 2 | The view: zoom buckets and fit-to (2) | [`L2-view.md`](L2-view.md) |
| 2 | The view described as data (2) | [`L2-describe.md`](L2-describe.md) |
| 2 | Errors and warnings (1) | [`L2-errors.md`](L2-errors.md) |
| 2 | Memory budget map (1), and the measurement behind it (no diagram) | [`L2-memory.md`](L2-memory.md) · [`memory-measurement.md`](memory-measurement.md) |
| 2 | The standalone app (1) | [`L2-app.md`](L2-app.md) |
| 2 | Tests and gates (1) | [`L2-gates.md`](L2-gates.md) |
| 3 | Sequences: a cold first frame · a pan · an idle host · a one-shot render (4) | [`L3-sequences.md`](L3-sequences.md) |
| 3 | State machines: a tile · borrowed geometry · a marker · an overlay's freshness (4) | [`L3-states.md`](L3-states.md) |
| — | The three approach notes, kept as the record of what was considered; their diagrams show the options, not the design (6) | [`approach-1`](approach-1-background-work.md) · [`approach-2`](approach-2-overlay-contract.md) · [`approach-3`](approach-3-dependencies-and-layout.md) |
| — | PLAN entry checks: the radar image source; the terminal matrix | [`plan-entry-checks.md`](plan-entry-checks.md) |
| — | The implementation plan, with the work packages' dependency diagram (1) | [`../04-development/implementation-plan.md`](../04-development/implementation-plan.md) |
| — | How the first host starts before a remote exists, and the integration-review template (D-60) | [`../04-development/first-host-start.md`](../04-development/first-host-start.md) · [`../07-readiness/integration-review-template.md`](../07-readiness/integration-review-template.md) |
| — | The parity mapping: each of this release's 62 parity rows, the work package that builds it, the test that proves it | [`../04-development/parity-mapping.md`](../04-development/parity-mapping.md) |

**Counted:** 3 in this file, 2 in the contract, 21 at level 2, 8 at level 3 — **34**; with the approach notes' 6 and the plan's 1, **41**.

---

## Level 0 — Context and trust boundary

Who the library talks to, and who it trusts. **The library trusts nothing that arrives as bytes**, including what its own host hands in: the host's data is validated (NFR-20), everyone else's is limited, decoded defensively and fuzzed (NFR-10). Nothing crosses back out to a terminal or a host uncleaned (FR-34).

```mermaid
flowchart LR
    subgraph PEOPLE[" "]
      U(["Person at a terminal"])
      DEV(["Host developer"])
    end

    subgraph HOSTBOX["Host application — for example the first host, or the project's own app"]
      direction TB
      HUI["Host's interface and clock<br/>draws every tick · owns keys and pointer"]
      HPUMP["Host's pump<br/>its goroutines call Work (D-73)"]
      HFETCH["Host's weather fetchers<br/>alerts · radar image · temperature grid (D-15)"]
      HTHEME["Host's theme → palette tokens (D-63)"]
    end

    subgraph LIBBOX["go-tuiMaps library — starts no goroutine, touches no terminal (M5)"]
      direction TB
      GATE{{"Hand-in gate<br/>validate · typed errors · warnings (NFR-20)"}}
      CORE["Map<br/>view · overlays · tiles on hand · pending work"]
      OUTGATE{{"Way out<br/>every string cleaned (FR-34)"}}
    end

    subgraph UNTRUSTED["Untrusted — limited before allocating, fuzzed (NFR-10)"]
      direction TB
      TS[("Tile service<br/>only if the host names one (D-65)")]
      DISK[("Disk cache<br/>off unless configured (FR-21b)")]
      EMB[("Embedded tiles, zoom 0–3<br/>opt-in package, hash-listed (FR-28a)")]
      USRSTYLE[/"A user's style file"/]
    end

    WX[("Weather services")] --> HFETCH
    U <--> HUI
    DEV -. "writes" .-> HOSTBOX
    HFETCH -- "overlay structs (D-74)" --> GATE
    HTHEME -- "palette" --> GATE
    HUI -- "intents: pan · zoom · focus (FR-24)<br/>time (FR-25) · rectangle · colour depth" --> GATE
    GATE --> CORE
    HPUMP -- "Work(ctx)" --> CORE
    CORE -- "secure fetch, limits, no address in errors (FR-22b)" --> TS
    CORE <--> DISK
    EMB --> CORE
    USRSTYLE --> GATE
    CORE --> OUTGATE
    OUTGATE -- "frame: cells of glyph + colours<br/>legend · credits · description · warnings<br/>changed? · call-me-by deadline" --> HUI
```

**What this diagram settles**

| Question | Answer | From |
|---|---|---|
| Who fetches weather? | The host. The library never calls a weather service. | D-15 |
| Does the library reach the network by itself? | Never, unless the host names a tile source. | D-65 |
| Whose goroutines? | The host's. The library starts none. | D-73 |
| Who reads keys and the mouse? | The host; the library takes intents. | D-17, FR-24 |
| Who owns time? | The host supplies it; the library answers "call me by…". | FR-25 |
| What is trusted? | Nothing that arrives as bytes. Host input is validated; the rest is limited, decoded defensively and fuzzed. | NFR-10, NFR-20 |
| What leaves? | Cells, and plain data; every string cleaned on the way out. | FR-34, D-52 |

---

## Level 1 — The parts

One public package holds the whole contract (D-74). Everything under `internal/` can change freely without breaking a host (D-60). The app, examples, generator and test oracle are **separate modules**, so nothing they import appears in a host's dependency graph (D-75).

```mermaid
flowchart TB
    subgraph LIB["Library module"]
      direction TB
      PUB["<b>tuimaps</b> — the public contract<br/>Map · overlay structs · presets · palette tokens<br/>Work · Pending · Settle · Close<br/>Frame · Legend · Credits · Description · Warnings"]
      ASSETS["<b>tuimaps/assets</b><br/>opt-in embedded tiles z0–3 (D-27, D-33)"]

      subgraph INT["internal/"]
        direction TB
        subgraph DATA["getting data in"]
          direction LR
          TILES["<b>tiles</b><br/>sources · cache · stand-ins · retry times"]
          MVT["<b>mvt</b><br/>own decoder · limits · drop-at-decode (D-75)"]
          ARC["<b>archive</b><br/>minimal single-file reader (D-58)"]
          FETCH["<b>fetch</b><br/>secure transport · redirects · limits (FR-22b)"]
          OVR["<b>overlay</b><br/>validate · simplify · classify · index — resampling is render's, at draw time"]
        end
        subgraph DRAW["turning it into cells"]
          direction LR
          PROJ["<b>project</b><br/>web-mercator · view ↔ cell · distances"]
          STYL["<b>style</b><br/>dark and bright · user styles · profiles (FR-19, FR-20)"]
          REN["<b>render</b><br/>braille canvas · compositing order · labels · markers"]
          COL["<b>colour</b><br/>tokens · presets · ramps per depth · checker"]
        end
        SCENE["<b>scene</b><br/>the prepared types everything shares: decoded tile, prepared overlay, job — imports nothing here"]
        subgraph OUTP["what leaves"]
          direction LR
          DESC["<b>describe</b><br/>description as data (D-52)"]
          TXT["<b>textsafe</b><br/>cleaning · clusters · width (FR-34, NFR-8)"]
          WORKQ["<b>work</b><br/>capped queue · newest view wins · no limiter: the pump's width is the host's (D-73, D-84)"]
          FAULT["<b>fault</b><br/>the typed error · both closed lists of kinds"]
          KIT["<b>testkit</b><br/>imported by test files only — a static check says so"]
        end
      end

      PUB --> WORKQ
      PUB --> TILES
      PUB --> FAULT
      FAULT --> TXT
      TILES --> FAULT
      OVR --> FAULT
      FETCH --> FAULT
      MVT --> FAULT
      STYL --> FAULT
      ARC --> FAULT
      WORKQ --> FAULT
      REN --> FAULT
      COL --> FAULT
      DESC --> FAULT
      PROJ --> FAULT
      TILES --> SCENE
      OVR --> SCENE
      REN --> SCENE
      DESC --> SCENE
      WORKQ --> SCENE
      PUB --> REN
      PUB --> OVR
      PUB --> DESC
      PUB --> COL
      ASSETS -. "handed to New as an option — never registers itself" .-> PUB
      TILES --> FETCH
      TILES --> MVT
      TILES --> ARC
      REN --> PROJ
      REN --> STYL
      REN --> COL
      REN --> TXT
      OVR --> PROJ
      DESC --> PROJ
      DESC --> TXT
      TXT --> DEP["go-runewidth · uax29<br/>the only third-party imports (D-75, D-81)"]
    end

    subgraph OUT["Separate modules, same repository"]
      direction LR
      APP["<b>cmd/tuimaps</b><br/>keys · pointer · describe mode · headless flag"]
      EX["<b>examples/</b><br/>one per shape · a pump · a pump in the first host's idiom · the radar table"]
      GEN["<b>tools/gen-assets</b><br/>builds the embedded tiles (FR-28a)"]
      ORA["<b>tools/oracle</b><br/>a proven decoder, tests only"]
      AK["<b>tools/answer-key</b><br/>the independent M1 key · imports nothing from the library (D-67)"]
    end
    APP --> PUB
    EX --> PUB
    GEN --> ARC
    GEN --> MVT
    ORA -. "differential tests" .-> MVT
```

**Rules of the layout**

| Rule | Why |
|---|---|
| A host imports `tuimaps`, and `tuimaps/assets` if it wants offline tiles. Nothing else is importable. | D-74; internal parts stay free to change under D-60's promise |
| `render` never imports `tiles`, `fetch` or `overlay`'s slow paths. It reads what is already on hand. | FR-23: Render does no input or output, ever |
| Only `work` runs slow jobs, and only when the host calls `Work`. It knows jobs only through `scene`'s job type — `tiles`, `overlay` and `describe` supply jobs; `work` imports none of them. | D-73; this is what keeps the import graph free of cycles (PL-CQ-7) |
| Only `fetch` opens a connection; only `textsafe` lets a string out. Every package that reports a problem does it through `fault`, whose only import is `textsafe`. | One place to enforce FR-22b; one place to enforce FR-34 |
| The public package imports `tiles` to configure sources and caches and to hand its jobs to `work`. | Without this edge nothing reaches `tiles`, and through it `fetch`, `mvt` and `archive` (P2-DOC-2) |
| Only `textsafe` imports third-party code. | D-75 |
| The framework the first host uses appears only under `examples/`. | D-13, D-75 |

---

## Level 1 — The public contract at a glance

Everything a host can call or hand in, grouped by what it is for. Names are illustrative (D-71). **The full statement — what each call promises, which calls are safe together, the pump, the end of a borrow — is [the contract](contract.md).**

```mermaid
flowchart LR
    subgraph IN["Host → Map"]
      direction TB
      A1["<b>Life</b><br/>New(options, WithSize) · Close() — closes at once and reports calls still inside"]
      A2["<b>Size and view</b><br/>SetSize(cols, rows) — state, set BEFORE Settle or Render<br/>intents: Pan · PanCells · Zoom · ZoomAround · Recentre · FitWorld · FitTo(places, overlays, margin) (FR-24, D-76)"]
      A2b["<b>Places and markers</b><br/>SetPlaces(places) · AddPlace(place) · RemovePlace(id) — ids as upstream (P-61)<br/>what Describe answers for, what FitTo can fit, what markers draw (FR-26)"]
      A3["<b>Overlays</b> (D-74, D-86)<br/>Set(overlay) → created or replaced · old geometry released yes/no<br/>Remove(id) → found or not · released yes/no · InUse(id) · BorrowCheck(on)<br/>structs: Features · ScalarGrid · Image (· VectorGrid · TileImages later)<br/>presets: Temperature · Radar · Alerts (· Wind later) (D-69)"]
      A4["<b>Look</b><br/>SetPalette(tokens) (D-63) · SafeRamps(on) · Ground(painted or declared) (D-64)<br/>ColourDepth(hint) · ReduceMotion(on) (NFR-21) · Layers(on/off) (FR-36) · LabelLanguage(code) (D-82)"]
      A5["<b>Tiles</b><br/>Source(named network source) (D-65) · CacheRoot(path) · Fetcher(replacement)<br/>SharedCaches(handle) (FR-27, D-85) · CacheUse() (D-90) · Purge() · Verify() (FR-22a)"]
      A6["<b>Running the work</b> (D-73, D-84)<br/>Pending() · Work(ctx) · Settle(ctx) · OnPending(wake)<br/>Work and Settle report the ids whose borrow they ended (D-86)"]
    end
    subgraph OUTB["Map → Host"]
      direction TB
      B1["<b>The picture</b><br/>Render(size, now) → Frame: exactly-sized rows (NFR-8), valid until the next Render<br/>status: complete · still sharpening · no tiles · failed"]
      B2["<b>When to call again</b><br/>Changed() counter · NextCall(wallClock): the earliest of a marker phase, a retry time, an overlay going stale (FR-25, FR-32)"]
      B3["<b>The same facts as data</b><br/>Legend() · Credits() · Scale() (FR-13, FR-14, FR-33)<br/>Footer(): centre and zoom in upstream's wording (P-57)<br/>Describe(places): each part ready or pending (D-52) · Focused()"]
      B4["<b>What went wrong</b><br/>errors of a closed list of kinds · Warnings() ≤ 64 (NFR-20)<br/>CheckRamp(ramp, ground) for a host's own tests (D-53, D-88)"]
    end
    IN --> M((Map)) --> OUTB
```
