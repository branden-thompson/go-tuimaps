# Level 2 — The tile pipeline

Up: [architecture](architecture.md) · Carries: FR-21a, FR-21b, FR-22a, FR-22b, FR-23, FR-25, FR-28a, FR-31, FR-35, NFR-10, D-18, D-30, D-33, D-58, D-65, D-73, D-75, D-82, D-90, L-13

## Where a tile can come from

Each wanted tile is one job, and each job asks the sources in one fixed order: disk cache, the named network source, the embedded set. Embedded tiles never override a source the host chose (L-13): for zoom 0 to 3 with a network source named, the network's tile wins when it arrives, and the embedded one is what is drawn until then.

```mermaid
flowchart LR
    NEED["Render — or Settle — notes: tile z/x/y is wanted<br/>AND its nearest ancestors the sources can supply, so there is a stand-in to draw"] --> MEM{"In the memory cache?<br/>keyed by source identity + label language + z/x/y<br/>(never by style — FR-31, D-82)"}
    MEM -- yes --> USE["On hand → drawn next frame"]
    MEM -- no --> ANC["Meanwhile: draw the nearest ancestor ON HAND as a stand-in (D-30).<br/>None is on hand until a Work call has decoded one — before that the frame shows<br/>ground, places, overlays and the notice (FR-23). With the assets imported, the z0–3 ancestor<br/>is wanted first and is the cheapest job, so it arrives first"]
    MEM -- no --> Q[("Pending work<br/>capped · de-duplicated · newest view wins")]
    Q -- "host calls Work (D-73)" --> ORDER

    subgraph ORDER["One job: try sources in order"]
      direction TB
      S1["1 Disk cache<br/>only if the host set a root (FR-21b)"]
      S2["2 The host's named network source<br/>only if one was named (D-65)"]
      S3["3 Embedded tiles z0–3<br/>only if the host passes them as an option"]
      S1 -- miss --> S2 -- "fail or none" --> S3
    end

    ORDER --> BYTES["Bytes"] --> GATE
    subgraph GATE["Untrusted-input gate (NFR-10) — every limit checked before allocating"]
      direction TB
      G1["Body ≤ 2 MiB · gzip or none · decompressed ≤ 8 MiB"]
      G2["Own decoder (D-75): layers ≤ 64 · features ≤ 100,000 · geometry integers ≤ 2,000,000"]
      G3["Drop while decoding: layers the schema mapping does not use (FR-35) · every place-name language but the configured one (D-82)"]
      G4["Count first, then allocate once: each kept layer's geometry is counted without decoding it,<br/>and its slabs are made at their exact size · coordinates as 16-bit integers ·<br/>keys ≤ 4,096 · values ≤ 400,000 · retained ≤ 4 MiB · a host may lower any limit, never raise one"]
      G1 --> G2 --> G3 --> G4
    end
    GATE -- "ok" --> STORE["Memory cache (byte-capped)<br/>what a live view draws is never evicted · spares only in the room left (D-90)"]
    GATE -- "ok, and it came from the network" --> WRITE["Write to disk cache<br/>only after a complete successful decode (FR-22a)<br/>temporary file, then rename"]
    GATE -- "refused or failed" --> FAIL["Record a not-before time for this tile<br/>30 s doubling to 10 min (FR-23)<br/>a bad cached file is deleted"]
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
