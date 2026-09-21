# AI-9 — PMTiles read path, OpenMapTiles-schema files, and measured embedded-tile sizes

| Field | Value |
|---|---|
| Phase | DISCOVER — Tier 2 research |
| Date | 2026-09-18 (all sizes measured this date) |
| Scope | Evidence for ruling D-18 (dependency-free PMTiles reader; basemap resilience) and for the pending embedded-tile depth decision under D-27 |
| Status | Section 7 is a proposal for HUM LEAD. Licence points are reported as documented by the licensors; this is not legal advice. |
| Verification | Re-checked by the coordinator on 2026-09-18: the provider's published file index lists tiles.pmtiles for the 20260913 run; a HEAD request returns 86,516,374,189 bytes with accept-ranges: bytes; a 127-byte range read of the header returned 206; upstream's five embedded zoom 0–1 tiles total 490,578 bytes, consistent with the 490,504 measured here. All held. |
| Correction recorded | Ruling D-18 carried a statement, marked not yet verified, that PMTiles basemaps do not match the OpenMapTiles styles. Section 3 refutes it for the provider's own planet file: same 16 layers, and zoom 0–1 tiles byte-identical to the tile server's. |

**Two findings change the picture.**
- OpenFreeMap now publishes a full-planet **PMTiles** file in the OpenMapTiles schema. Its README does not say so.
- The zoom 0–3 set is **16.1 MB** as served, not 2–5 MB. About 93% of those bytes are translated place names. Stripping the translations brings zoom 0–3 down to **1.7 MB**.

## 1. PMTiles v3 read path

Source: https://github.com/protomaps/PMTiles/blob/main/spec/v3/spec.md (a copy is in the scratch directory, `spec_v3.md`).

- **Header (§3), 127 bytes, little-endian.**
  - Magic `PMTiles` (7 bytes), then version `3`.
  - Eleven uint64 values from offset 8: root dir offset and length, metadata offset and length, leaf dirs offset and length, tile data offset and length, then addressed tiles, tile entries, tile contents.
  - Single bytes from offset 96: clustered, internal compression, tile compression, tile type, min zoom, max zoom.
  - Min and max position as int32 lon/lat ×1e7, then centre zoom and centre position.
- **Compression enum (§3.3):** 0 unknown, 1 none, 2 gzip, 3 brotli, 4 zstd.
  - Internal compression covers the root directory, metadata and each leaf directory, compressed individually.
  - Tile compression covers the tile blobs.
  - The standard library has gzip only. brotli and zstd must return a clear "unsupported" error.
  - Tile type must be 1 (MVT). Reject 6 (MapLibre Vector Tile).
- **Directories (§4, pseudocode A.2).**
  - Gunzip, then read varint `n`, then n delta-encoded tile IDs, n run lengths, n lengths and n offsets.
  - An offset varint of `0` with i>0 means the previous offset plus the previous length. Otherwise the offset is `value−1`.
  - Run length 0 marks a leaf pointer; its offset is relative to the leaf section. Otherwise the entry covers `run` consecutive tile IDs and its offset is relative to the tile data section.
  - Lookup is a binary search for the last entry whose ID is ≤ the target.
  - The root directory must sit inside the first 16,384 bytes (§2, §4). More than one leaf level is discouraged (§4).
- **Tile ID (§4.1).** The ID is Σ4^i for i<z, plus the Hilbert index. The spec test vector 12/3423/1763→19078479 passed in the measurement script.
- **Metadata (§5).** A JSON object. `vector_layers` is mandatory for MVT. `attribution` should be shown to the user.
- **Range-request pattern.**
  - The spec fixes only the 16 KiB prefix rule.
  - The docs say the format is read "with at most two cacheable intermediate requests" (https://github.com/protomaps/docs/blob/main/pmtiles/index.md).
  - A cold read is therefore **3 requests**: bytes 0–16383 (header and root), one leaf directory, then the tile.
  - Cache the header and root permanently, and keep leaf directories in an LRU.
- **Measured on OpenFreeMap's planet file.**
  - The root directory is 16,213 bytes and holds 3,501 leaf pointers.
  - One leaf directory is 29,024 bytes and holds 14,729 entries. Decoded, that is about 350 KB at 24 bytes per entry (a calculation, not a measurement).
  - All 85 zoom 0–3 tiles came back in **4 range requests** totalling 16.15 MB (`pm_read.py` in the scratch directory).
- **Estimated Go size: 400–500 lines plus tests.** That covers header parsing, directory decoding, Hilbert conversion, `io.ReaderAt` and HTTP range backends, and the LRU.
  - Standard-library packages: `encoding/binary`, `compress/gzip`, `bytes`, `io`, `os`, `net/http`, `encoding/json`, `math/bits`, `sort`, `sync`, `context`, `errors`, `fmt`.
  - For reference, go-pmtiles' `directory.go` (546 lines, includes writer code) and `tile_id.go` (56 lines) import only the standard library.

## 2. Existing Go implementations

- **github.com/protomaps/go-pmtiles**
  - Licence: **BSD-3-Clause** (GitHub API and its LICENSE file).
  - Latest release v1.31.2, published 2026-07-22.
  - Direct requires in go.mod: cloud.google.com/go/storage, Azure azcore and azblob, RoaringBitmap/roaring, alecthomas/kong, aws-sdk-go-v2 (plus service/s3 and smithy-go), **caddyserver/caddy/v2**, cespare/xxhash, dustin/go-humanize, paulmach/orb, prometheus/client_golang, rs/cors, schollz/progressbar, stretchr/testify, go.uber.org/zap, gocloud.dev, golang.org/x/sync, google.golang.org/api, zombiezen.com/go/sqlite.
  - That is 21 direct and 176 indirect requires. The release binary is 55 MB (darwin/arm64).
  - SQLite comes through zombiezen, which uses modernc.org/sqlite, a CGO-free implementation. I found no `import "C"` in the files I inspected. **A `CGO_ENABLED=0` build was not run.**
  - The reader files are stdlib-only, but they live in a package whose sibling files import the cloud SDKs. Importing the package pulls in the entire dependency graph.
  - **It is suitable as a test oracle in a nested test-only module, or through CLI-generated golden fixtures. It is unsuitable as a shipped dependency.**
- **Other Go readers** (GitHub search, 2026-09-18):
  - jzs/libpmtiles: no licence, last pushed 2023-11-04.
  - akhenakh/kvtiles: Apache-2.0, last pushed 2024-01-24.
  - Neither is a credible dependency.

## 3. PMTiles in the OpenMapTiles schema

**(a) Does OpenFreeMap publish PMTiles?**
- The README still says "weekly full planet downloads both in Btrfs and MBTiles formats", and explains "I would have loved to use PMTiles… on Cloudflare, range requests in 90 GB files have terrible latency".
- Its published index (https://btrfs.openfreemap.com/files.txt) nevertheless lists `areas/planet/{version}/tiles.pmtiles` for every run since `20260526_232801_pt`.
- A HEAD request on the 20260913 file returns 200, 86,516,374,189 bytes and `accept-ranges: bytes`. Its hash appears in that run's SHA256SUMS.
- The repo's `docs/tilegen_release_cadence.md` has the steps "Convert tiles.mbtiles to tiles.pmtiles… Verify… Upload".
- `tilegen/tilegen_lib/pmtiles.py` runs `pmtiles convert tiles.mbtiles tiles.pmtiles` with go-pmtiles 1.31.2.
- **The earlier statement to the owner is refuted for this file.**
  - Its metadata lists the same 16 OpenMapTiles layers as the TileJSON.
  - The five zoom 0–1 tiles, once decompressed, are SHA-256 **identical** to the tile server's.

**(b) `pmtiles convert INPUT.mbtiles OUTPUT.pmtiles`** (https://github.com/protomaps/docs/blob/main/pmtiles/cli.md)
- It copies `tile_data` blobs and the metadata table, and de-duplicates. `--no-deduplication` and `--tmpdir` are its documented options.
- The docs never use the word "lossless". The byte identity measured above is the evidence.

**(c) `pmtiles extract`**
- Docs: "The source archive may be local or remote. The source archive must be clustered."
- Documented flags: `--bbox`, `--region`, `--minzoom`, `--maxzoom`, `--dry-run`, `--download-threads`, `--overfetch`, `--bucket`.
- It accepts PMTiles input only. It refuses unclustered sources and has no MBTiles input path.
- I ran it against the OpenFreeMap planet URL and it works.

**(d) Third parties.** MapTiler sells OpenMapTiles data. Whether it offers PMTiles, and on what terms, is UNVERIFIED.

**What a user who wants an offline region would do:**
1. Install the `pmtiles` CLI. This is an external tool.
2. Run `pmtiles extract https://btrfs.openfreemap.com/areas/planet/<version>/tiles.pmtiles region.pmtiles --bbox=…`.
3. Point go-tuiMaps at `region.pmtiles`.

Only step 3 happens inside go-tuiMaps.

## 4. Planet and regional sizes

Planet file sizes come from HEAD requests. Regional sizes come from `pmtiles extract --dry-run` against the 20260913 planet file, using bounding boxes.

| File or extract | Size |
|---|---|
| Planet MBTiles | 102,484,172,800 B |
| Planet PMTiles | 86,516,374,189 B |
| Planet Btrfs, gzipped | 98,125,039,706 B |
| Planet Btrfs, uncompressed | 164,092,379,136 B |
| Metro (Denver, 0.7°×0.5°), zoom 0–14 | 50 MB (1,424 tiles) |
| State (Colorado), zoom 0–14 | 322 MB |
| State (Colorado), zoom 0–10 | 14 MB |
| Contiguous US, zoom 0–14 | 12 GB |
| Contiguous US, zoom 0–10 | 407 MB |
| World, zoom 0–6 | 111 MB |

## 5. Embedded low-zoom sets (measured)

**Method.** The tiles were read by range request from the published planet PMTiles, not from the tile server. Sizes are the gzipped blobs as stored. For zoom 0–1, the decompressed tiles match the tile server byte for byte (checked 2026-09-18).

| Zoom | Tiles | Bytes (gzip) | min / median / max | Cumulative |
|---|---|---|---|---|
| 0 | 1 | 47,985 | 47,985 / 47,985 / 47,985 | 47,985 |
| 1 | 4 | 442,519 | 103,131 / 106,878 / 125,631 | **490,504** |
| 2 | 16 | 6,655,870 | 81,618 / 512,151 / 721,894 | **7,146,374** |
| 3 | 64 | 8,957,190 | 102 / 67,382 / 654,547 | **16,103,564** |

- The upstream clone's five embedded tiles total 490,578 B. That is a different data date and is consistent with the 490,504 measured here.
- The largest decompressed tile is 1,557,618 B (2/2/1).

**Byte split by layer, zoom 0–3, each layer gzipped alone:**

| Layer | Bytes (gzip) | Share |
|---|---|---|
| place | 14,994,830 | 93.2% |
| water_name | 635,443 | 3.9% |
| boundary | 200,819 | 1.2% |
| water | 180,028 | 1.1% |
| landcover | 76,083 | 0.5% |
| waterway | 3,672 | <0.1% |

- There are no roads or buildings at these zooms.
- The bulk is in the 89 distinct attribute keys, almost all `name:xx` translations, carried on 34,004 place features.
- Dropping whole layers barely helps. Removing water_name, landcover and waterway still leaves 95.7% of the bytes.

**Rewriting the label layers** (`strip_names.py` and `rank_filter.py` in the scratch directory). Feature counts are unchanged where no rank filter applies (36,549 before and after). The rows are cumulative gzip sizes.

| Variant | zoom 0–1 | zoom 0–2 | zoom 0–3 |
|---|---|---|---|
| As served | 490,504 | 7,146,374 | 16,103,564 |
| Translations stripped (keeps name, name:latin, name_en, name:en, class, rank, capital, iso_a2) | 89,246 | 759,090 | **1,697,545** |
| Stripped, plus place rank ≤ 4 | 89,246 | 601,039 | 1,382,154 |
| Stripped, plus place rank ≤ 2 | 89,246 | 544,151 | 1,172,721 |
| Geometry only, no labels (estimate) | — | — | ≈460,000 |

- With translations stripped, the largest decompressed tile falls from 1.56 MB to 159 KB. That matters for the 10 MB resident-memory budget.

**Tooling.**
- A roughly 100-line stdlib-only script is enough; I wrote one for this measurement. An MVT tile is a sequence of layer messages, so rewriting the key and value tables is simple.
- tippecanoe's tile-join can exclude layers and attributes. I did not test it.

**Licence.**
- ODbL defines a "Derivative Database" as "any translation, adaptation, arrangement, modification, or any other alteration of the Database or of a Substantial part of the Contents".
- Redistributed tiles, stripped or not, fall under §4.2: convey "only under the terms of this License", include the licence or its URI, and keep the notices.
- §4.6(b) is satisfied by publishing "the method of making the alterations… (such as an algorithm)", which here means committing the script.
- The OpenMapTiles LICENSE says "Products or services using maps derived from OpenMapTiles schema need to visibly credit 'OpenMapTiles.org'".
- **Stripping changes none of these obligations.** But the embedded data is ODbL, not MIT, so the embed package needs its own notice.

## 6. Sourcing the embedded tiles legitimately

- This is solved. `pmtiles extract <planet tiles.pmtiles URL> out.pmtiles --maxzoom=3` against the published download took **5 requests, 16 MB and 1.7 s**. `pmtiles verify` passed, and the result is in the scratch directory.
  - The docs say: "Extracting a full sub-pyramid from 0 to maxzoom is always an efficient operation."
  - Tooling: the CLI, or the project's own reader, followed by the strip script.
- Range reads against the MBTiles file would need a SQLite HTTP virtual file system. That is impractical, and now moot. Btrfs images would have to be mounted on Linux, so they are no better.
- I found nobody publishing a low-zoom-only file.
- One quirk: the extract's JSON metadata still says `maxzoom 14`, while its header says 3. Trust the header.

## 7. First-run proposal

**Lookup order:** memory LRU → disk cache → local PMTiles file, if configured → network → overzoomed ancestor → embedded tiles.

- **Embedded floor.**
  - An opt-in package holding the name-stripped set, generated by a committed script from a pinned planet version.
  - I recommend zoom 0–3 at 1.7 MB, with zoom 0–2 at 0.76 MB if binary size wins.
  - The standalone app always imports it. A host opts in with one import.
- **Disk cache.**
  - No expiry, keyed by z/x/y. OpenFreeMap itself serves tiles with `cache-control: max-age=315360000`.
  - Fill it only with tiles the user actually viewed. Never prefetch.
- **Stand-in tiles.** Draw the overzoomed parent immediately and replace it when detail arrives.
- **Default network source.**
  - The OpenFreeMap tile server via TileJSON.
  - Do not make the planet PMTiles URL a default live source. The operator's README calls that latency "terrible", and the bucket exists for downloads.
- **Standalone app.**
  - `--pmtiles path|URL` and the default cache location come from the app.
- **Library host.**
  - Sources are injected. A host that configures no network source gets no network access.
  - The host supplies the cache directory.
- **Very first run with no network.**
  - The world map renders instantly from the embedded tiles.
  - Deeper zooms show overzoomed zoom-3 geometry with a one-line "offline — low-detail basemap" status.
  - There is never a blank screen or an error dialog.
- **Strongest counter-argument.**
  - A stripped set adds a build pipeline, derivative-database duties and Latin/English-only world labels.
  - Upstream's as-served zoom 0–1 (490 KB) needs no tooling.
  - Rebuttal: the stripped zoom 0–1 set is 89 KB, and the as-served zoom-2 tiles decode to 1.5 MB each.

## 8. Risks and unknowns

1. **A range request was ignored once.**
   - The first range request, made from Python, returned **HTTP 200** and offered the whole 86 GB. I aborted it before reading the body.
   - Three immediate retries and every later request returned 206. I could not reproduce it.
   - The reader MUST require a 206 with a matching Content-Range, and must otherwise close the connection without reading the body.
   - Go's `net/http` was not tested.
2. **The planet PMTiles is undocumented in the README.** It could be withdrawn, so pin a version and its checksum.
3. **OpenFreeMap's terms forbid users to "Attempt to collect data from the service in automated ways without permission".**
   - A no-expiry cache of viewed tiles seems to comply.
   - For the owner to rule: whether any prefetch is allowed.
4. **Putting ODbL data inside an MIT repository.**
   - For the owner to rule: the wording of the separate notice, and how attribution is shown in a text interface.
   - This is not legal advice.
5. **Leaf directory memory.** Roughly 350 KB decoded per planet leaf, and the planet has 3,501 of them. The LRU size needs a budget. Regional extracts have far fewer.
6. **Missing tiles.** The planet addresses 275,963,617 tiles, against 357,913,941 in a full zoom 0–14 pyramid. Missing tiles must be handled. Which tiles are absent is UNVERIFIED.
7. **Unverified or untested:**
   - MapTiler PMTiles availability and terms.
   - go-pmtiles building with `CGO_ENABLED=0`.
   - tile-join for stripping.
   - Regional sizes are bounding-box dry-runs, not administrative boundaries.
8. **Requests made for this research.** Six to the tile server (the TileJSON and five tiles). About 400 range requests to the download bucket, mostly directory reads, plus two 16 MB tile pulls.
