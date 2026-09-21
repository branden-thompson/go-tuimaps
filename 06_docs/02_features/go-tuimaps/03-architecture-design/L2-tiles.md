# Level 2 — The tile pipeline

Up: [architecture](architecture.md) · Carries: FR-21a, FR-21b, FR-22a, FR-22b, FR-23, FR-25, FR-28a, FR-31, FR-35, NFR-10, D-18, D-30, D-33, D-58, D-65, D-73, D-75, D-82, D-90, L-13

## Where a tile can come from

A wanted tile is asked of the source the host chose: the disk cache first, then the named network source, inside one job. The embedded set is asked in a job of its own, which is the cheapest there is and so sorts first. Embedded tiles never override a source the host chose (L-13): with a network source named, an embedded tile is only ever a stand-in — the network's tile wins when it arrives, and the embedded one is what is drawn until then. With no source named, the embedded set is the source for the zooms it holds.

**What is drawn for a tile is the first tile on hand of a chain**: the chosen source's tile, then the embedded one at the same zoom, then the same pair for each ancestor in turn. The memory cache is told the chains of every live view, and the first tile on hand of each chain is that view's need (D-90). *(BUILD, WP-06: the first design had one job try all three sources in order. That could not draw the embedded tile while a slow network fetch was still running, which is the behaviour this page has always promised.)*

```mermaid
flowchart LR
    NEED["Render — or Settle — notes: tile z/x/y is wanted<br/>AND its nearest ancestors the sources can supply, so there is a stand-in to draw"] --> MEM{"In the memory cache?<br/>keyed by source identity + label language + z/x/y<br/>(never by style — FR-31, D-82)"}
    MEM -- yes --> USE["On hand → drawn next frame"]
    MEM -- no --> ANC["Meanwhile: draw the nearest ancestor ON HAND as a stand-in (D-30).<br/>None is on hand until a Work call has decoded one — before that the frame shows<br/>ground, places, overlays and the notice (FR-23). With embedded tiles passed, the z0–3 ancestor<br/>is wanted first and is the cheapest job, so it arrives first"]
    MEM -- no --> Q[("Pending work<br/>capped · de-duplicated · newest view wins")]
    Q -- "host calls Work (D-73)" --> ORDER

    subgraph ORDER["Two kinds of job, each keyed by its source"]
      direction TB
      S3["The embedded tile, z0–3 — only if the host passes them as an option.<br/>The tile itself when no source is named; otherwise only a stand-in (L-13)"]
      S1["The chosen source's tile: 1 the disk cache<br/>only if the host set a root (FR-21b)"]
      S2["2 the host's named network source<br/>only if one was named (D-65)"]
      S1 -- miss --> S2
    end

    ORDER --> BYTES["Bytes"] --> GATE
    subgraph GATE["Untrusted-input gate (NFR-10) — every limit checked before allocating"]
      direction TB
      G1["Body ≤ 2 MiB · gzip or none · decompressed ≤ 8 MiB"]
      G2["Own decoder (D-75): layers ≤ 64 · features ≤ 100,000 · geometry integers ≤ 2,000,000"]
      G2R["Refused, never read past (D-75, D-126): a field of a layer or a feature carrying the wrong kind of value —<br/>an id or a type that is not a number, tags or geometry that are not packed · a point that runs a line or closes a ring ·<br/>a count of zero, a count beyond the integers left, a pen that leaves the 16-bit range"]
      G3["Drop while decoding: layers the schema mapping does not use (FR-35) · every place-name language but the configured one (D-82)"]
      G4["Count first, then allocate once: each kept layer's geometry is counted without decoding it,<br/>and its slabs are made at their exact size · coordinates as 16-bit integers ·<br/>keys ≤ 4,096 · values ≤ 400,000 · retained ≤ 4 MiB · a host may lower any limit, never raise one"]
      G1 --> G2 --> G2R --> G3 --> G4
    end
    GATE -- "ok" --> STORE["Memory cache (byte-capped)<br/>what a live view draws is never evicted · spares only in the room left (D-90)"]
    GATE -- "ok, and it came from the network" --> WRITE["Write to disk cache<br/>only after a complete successful decode (FR-22a)<br/>temporary file, then rename"]
    GATE -- "refused or failed" --> FAIL["Count the failure; the owner call that follows stamps a not-before time<br/>from the host's clock: 30 s doubling to 10 min (FR-23)<br/>a bad cached file is deleted and the network asked"]
    STORE --> BUMP["Change counter +1 · call-me-by = now"]
    FAIL --> DL["Call-me-by includes the earliest not-before time (FR-25)"]
```

## The network edge

Everything that touches the network goes through one package, so the rules are enforced in one place.

```mermaid
flowchart TB
    REQ["A fetch for the named source"] --> SAME{"Same scheme and host as the source,<br/>or a host the options allow?"}
    SAME -- no --> X0["Refused before anything is sent"]
    SAME -- yes --> T1{"Secure transport?"}
    T1 -- "no, and not a LITERAL loopback address (a name can resolve anywhere), and not explicitly allowed" --> X1["Refused"]
    T1 -- yes --> DIAL["Dial — check the address actually connected to:<br/>not loopback, link-local or private unless the source already is (FR-22b)"]
    DIAL --> SEND["Send: identifying User-Agent (library, version, host's token)<br/>no Referer · proxy settings honoured"]
    SEND --> RESP{"Response"}
    RESP -- "redirect" --> RD{"≤ 3 · never secure→plain ·<br/>never to another host unless allowed"}
    RD -- ok --> DIAL
    RD -- no --> X2["Refused"]
    RESP -- "range asked, but not 206 with exactly that range" --> X3["Closed unread (AI-9 §8)"]
    RESP -- "success" --> LIM["Body read through a limit"]
    LIM --> OUT["Bytes to the gate"]
    X0 & X1 & X2 & X3 --> ERR["Error names scheme and host only<br/>never the address — it may hold a key<br/>never wraps the transport error (FR-22b)"]
```

## The disk cache

| Rule | From |
|---|---|
| Off unless the host sets a root — its file names are a record of where the user looked | FR-21b |
| Paths built only from a hash of the source identity and integers the library formats; all access confined to the root | FR-21b |
| Private to the user; a root writable by others is refused where the platform can tell | FR-21b |
| No expiry. Byte cap (default 256 MB); evict least-recently-read; recency = modification time set on read, at most hourly; total kept in memory, seeded by one walk at open; prune to 90% during `Work`, never while drawing | D-18, FR-21a |
| `Purge` and `Verify` exposed, and offered by the app | FR-22a |
| **Layout, stable and documented:** `<root>/v1/<hash>/<z>/<x>-<y>.pbf`, where `<hash>` is the first 32 hex digits of the SHA-256 of the source's identity. Raw bytes as the source sent them, so a file serves every label language | L-12, P-48; set in BUILD |
| The root is opened once as a confined handle, and the permission check is made on that handle. Every entry on the way to a file is looked at without following links before the file is opened, and what was opened is compared with what was looked at | PL-IS-4 |
| A tile still being drawn from memory is still being read: the plan notes it in memory, and the next job writes its recency down. A plan does no input or output | FR-21a, FR-23 |

## The named source's two modes (P-49)

| Rule | From |
|---|---|
| A source ending in `/` is a prefix: `{source}{z}/{x}/{y}.pbf`. Any other is a TileJSON document, whose first tile address is used | P-49 |
| The TileJSON is read once, inside the first tile job that needs it — never when the source is named. Until then the source is taken to hold zoom 0 to 14; afterwards, the zooms it declares. A refusal is kept and not asked for again; a failure to fetch is not kept | D-65, NFR-5 |
| At most 1 MiB, nested at most 64 deep, checked before it is parsed | NFR-10 |
| Its tile address: a secure scheme, no user name, the TileJSON's own host unless the application allowed another, only `{z}`, `{x}` and `{y}` filled and all three present, no token in the host, **no fragment** — a fragment is never sent, so every tile would be the same request *(found by the fuzz target in BUILD)* | FR-22b |
| A source that declares layers of which the map knows none is refused as an unsupported schema | FR-35 |

## The embedded tiles and their generator

```mermaid
flowchart LR
    PLANET[("Pinned planet archive<br/>version · length · entity tag")] -- "range requests, 206 only" --> GEN["tools/gen-assets (separate module)<br/>uses the minimal archive reader (D-58)"]
    GEN --> STRIP["Decode with the library's own decoder<br/>strip every translation but English (D-33, D-82)"]
    STRIP --> OUT1["85 tiles, zoom 0–3: 1.02 MB measured,<br/>held under 2.5 MB by a test"]
    STRIP --> HASH["SHA-256 list: every source tile, every output tile"]
    OUT1 --> PKG["tuimaps/assets (opt-in import)"]
    HASH --> CI["Continuous integration:<br/>re-hash the asset · decode every tile through the gate"]
    PIN["Pin · hash list · asset<br/>change only together (FR-28a)"] -.-> GEN
```

## What can change this diagram

| If this changes… | …this part moves |
|---|---|
| The public PMTiles source is built (after v0.1.0) | A fourth source in "try sources in order"; it wraps the same archive reader |
| Per-tile maxima, measured across zoom 0 to 14 ([constants](constants.md), section 1), change | The numbers in the gate |
| The provider changes its schema | Only the schema mapping (FR-35), and "drop while decoding" follows it |
