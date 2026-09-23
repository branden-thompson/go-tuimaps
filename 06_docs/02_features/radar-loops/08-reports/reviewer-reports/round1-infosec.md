# round1 infosec — reviewer report, verbatim

Filed under v0.2.0 D-38/round 2 (H-9): the reviewer's final message exactly as returned, except that machine paths are replaced by `<workspace>` and `<scratch>`. The harness refused reviewers' own report files, so this is the only copy.

---

# go-tuiMaps v0.2.0 Radar loops: DISCOVER exit, InfoSec red team

Reviewer: Principal InfoSec Engineer, blind. Read-only. The harness refused a write to `report.md`, so the report is here and nowhere else. The only thing I ran was a scratch `url.Parse` probe with no network access. It confirmed that Go refuses percent-encoded and `\`-userinfo hosts, so the L-10 host comparison has no parser gap on the library's own path.

## 1. Is external input validated and bounded? (SEC-03)

**F1: Image bytes are checked when the host hands them in, then decoded later from the host's own memory, and the decode only enforces the absolute pixel ceiling.** InfoSec
- Severity: **Important**
- Evidence:
  - `internal/overlay/store.go:437` keeps the host's `Overlay` as it is, including `Image.PNG`, without copying it.
  - `internal/overlay/image.go:117` checks the header against the map cap when the overlay is set.
  - `image.go:242-247` decodes the same slice later, inside `Work`, and re-checks only `maxPixels` (1,048,576), not the cap.
  - `internal/overlay/cache.go:118-131`: `BorrowCheck` fingerprints feature rings only, never images.
- Exploit: a host that reuses its fetch buffer (common with `bytes.Buffer`), or hostile bytes that arrive through it, can replace a frame after `Set`:
  - The frame then decodes at up to 4× the cap. D-20's "refused at hand-in" budget is bypassed, because it is judged against bytes the library does not own.
  - The wrong picture draws under the frame's valid time, which breaks M2 (loop honesty).
  - Twelve frames multiply both effects.
- Simplify/Delete? **Y**. Frames are 5–21 KB (wave2 M-B), so copying them costs nothing and removes the borrow.
- Action: add a requirement that loop frames are **copied at hand-in**, or hashed at hand-in and verified before decode (`ImageKey` already hashes the bytes, `classified.go:50-56`). Make the decode re-check the cap and the D-20 budget.

**F2: D-20's budget counts the wrong bytes, and nothing bounds the frame count.** InfoSec
- Severity: **Important**
- Evidence:
  - `image.go:21` lets a PNG be up to 8 MiB, while `image.go:106` limits pixels to 250,000.
  - `store.go:524-527` (`OwnedBytes`) counts one byte per classified pixel only. The retained PNG bytes are not counted, which wave1 already noted at `wave1-findings.md:69-70`.
- Exploit:
  - A 250k-pixel PNG padded with ancillary chunks to 8 MiB passes the checks. Twelve of them hold about 96 MiB against a "3 MB" budget.
  - Gap frames carry no PNG, so their byte cost is zero. The only bound on their number is the budget, and it cannot see them.
  - The brief's L-1 (`project-brief.md:55-70`) requires none of the validation wave1 lists at `wave1-findings.md:31-35`: every frame checked, a loop total, times in order, a gap marker.
- Simplify/Delete? N
- Action:
  - The budget counts retained PNG bytes plus classified pixels.
  - Cap each frame's file size relative to its pixel count.
  - Set a hard maximum frame count, gaps included.
  - Put W1-A's per-frame validation into L-1 as a requirement.
  - State that the playback interval (L-1.5) is library-owned and floored, never derived from host-supplied valid times. Duplicate or zero-interval times must not drive `NextCall` into a busy loop or past D-56's flash ceiling.

**F3: A host fetcher's responses are unbounded at the fetch layer, and L-7's text does not close this.** InfoSec
- Severity: **Important**
- Evidence:
  - `internal/fetch/get.go:149`: `Checked` exists, and no production code calls it.
  - `tiles.go:50-57` uses `m.fetcher` raw.
  - A host fetcher silently loses TLS ≥ 1.2, redirect confinement, the private-address refusal, the timeouts and the read limit (`fetch.go:99-119, 177-219`).
  - `internal/tiles/remote.go:282-285` passes through any `*fault.Error` the host's fetcher returns.
- The TileJSON size check (`remote.go:165`) and the decoder caps limit how much gets parsed. They do not limit what the host's fetcher has already pulled into memory.
- Wiring `Checked` appears only as a v0.1.0 defect (`wave1-findings.md:196-197`). L-7.1 (`project-brief.md:119-121`) does not require it.
- Simplify/Delete? N
- Action: L-7 requires every host fetcher to be wrapped in `Checked`. The contract states the split: what the library still enforces, and what a host fetcher takes on.

(TileJSON documents themselves are otherwise well bounded: 1 MiB and depth 64 before parsing (`remote.go:165`), a 2,048-byte template with only `{z}{x}{y}` tokens (`remote.go:63-89`), zooms range-checked, attribution cleaned.)

## 2. Authorization on state-changing paths; errors on destructive operations (SEC-02)

The library has no principal inside the process: holding `*Map` is the authority, which is right for a library. The findings are about calls whose effect is not what the host believes.

**F4: `Fetcher()` takes effect only at the next `Source`, and says nothing.** InfoSec
- Severity: **Important**
- Evidence: `tiles.go:67-83`. It returns no error, and the running `remote` keeps the library's own fetcher.
- Exploit: a host switches to a privacy fetcher (a proxy, Tor, its own user-agent) after `Source`. The library keeps opening direct connections that the host believes it has replaced.
- Simplify/Delete? N
- Action: the new L-7 surface either applies immediately (rebuild the remote) or refuses while a source is set, and a test holds whichever is chosen.

**F5: Purge and retention (L-9) cannot keep a promise made to the listener.** InfoSec
- Severity: **Important**
- Evidence:
  - (a) A job in flight writes after `Purge` or `CacheRoot("")`. `load` receives a snapshot of the options (`internal/tiles/pipeline.go:610`) and stores unconditionally (`:644-645`). `Disk.Empty` (`internal/tiles/disk.go:329-345`) invalidates nothing.
  - (b) Scope. `Purge` empties only the **current** source (`tiles.go:160-164`). Tiles from any earlier source survive, and the memory caches and decoded pictures are untouched.
  - (c) Age. "Recency" is an mtime written from the host's clock, at most hourly (`disk.go:217-218, 253`). A future-dated mtime never ages out. Enforcement that runs only in jobs leaves stale files on disk while the application is closed; the cache is only walked at open (`disk.go:86`).
  - (d) The brief was never corrected. L-9.2 (`project-brief.md:137`) still asks for a purge call "missing", which wave1 synthesis item 5 (`wave1-findings.md:180-181`) said would misdirect PLAN.
- Simplify/Delete? N
- Action: L-9 requires:
  - a purge of everything, regardless of source;
  - a cache generation that in-flight stores must match;
  - maximum age enforced at `CacheRoot` and in every job, with future-dated mtimes treated as expired;
  - a correction row for L-9.2.

**F6: Errors from cache writes and removals are swallowed.** InfoSec
- Severity: **Minor**
- Evidence:
  - `pipeline.go:645` discards the error from `Disk.Store`.
  - The contract's `cache-write-failed` warning exists (`internal/fault/fault.go:80`, `kinds.go:60`) and nothing ever raises it.
  - `disk.go:183` ignores a `Remove` that fails.
  - `ReadBack` counts a file as removed whether or not it was (`disk.go:370-371`).
- Consequence: pruning or age expiry can fail silently while `Verify` reports success. That is fatal to a retention promise.
- Simplify/Delete? N
- Action: raise `cache-write-failed`; count only removals that succeeded; have purge and expiry return what they could not delete. (`Empty` does return its error, `disk.go:341-343`, which is good.)

## 3. Secrets out of logs, errors and source; vetted crypto

**Clean.**
- Errors never repeat an address: `fetch.go:5-7`, `get.go:24-41`, `remote.go:54`.
- The host's error is not relayed (`get.go:161-162`).
- Source identities, which may hold keys, are hashed before reaching a path (`disk.go:137-139`).
- The user-agent token is filtered for header-breaking characters (`fetch.go:126-134`).
- TLS is standard `crypto/tls` with a 1.2 minimum and no `InsecureSkipVerify` (`fetch.go:110`). SHA-256 comes from the standard library.
- The `0.1.0-dev` user-agent (`fetch.go:24`) is already recorded.

## 4. Injection and traversal (SEC-01)

**Clean, with one contract gap.**
- Cache paths use only a hex hash and library-formatted integers, under `os.Root`. Symlinks are checked with `Lstat` at every step and the file is re-verified with `SameFile` (`disk.go:142-145, 151-171, 206-211`).
- Tile coordinates are validated before any path or URL is built (`disk.go:192`, `remote.go:317`).
- The `Range` header is built from integers (`get.go:84`).
- Every parser has a limit: JSON, MVT and PNG.

**F7: Tile-host confinement (L-10) depends on which fetcher is in use, and on two allow-lists.** InfoSec
- Severity: **Important**
- Evidence:
  - The fetcher's `AllowHosts` (`fetch.go:41-43`, `:190-203`) and the TileJSON check's `allowHosts` (`remote.go:96, 224`) are separate parameters. `Map.Source` passes empty values for both (`tiles.go:52, 58`), which is why confinement is strict today.
  - L-7.1 proposes exporting "allowed hosts" (`project-brief.md:121`). If only one of the lists is wired, the two paths diverge.
  - Under a host fetcher, `NewRemote` accepts plain `http` to any host (`remote.go:235`), and the tile template may then be `http` (`remote.go:104`). The library's own fetcher refuses that (`fetch.go:99`).
- Simplify/Delete? **Y**: one list.
- Action: L-10's contract promise covers host, port and scheme ("https unless the host allowed plain http"), holds for any fetcher, and comes from one allow-list. Test it through `Map.Source` with both fetchers.

## 5. Do errors leak structure, paths or stack traces?

**Clean.** Panics become a fixed message that names only the call (`guard.go:53-57`). Refusals are fixed text plus counts (`image.go:106-108`). No file paths appear: cache errors are fixed text (`disk.go:53-56`). HTTP status codes are cleaned before they are shown (`get.go:124`).

## 6. Blast radius, least privilege, what the disk reveals

The privilege posture is good:
- Nothing reaches the network until `Source` (`tiles.go:30-33`).
- Nothing touches disk until `CacheRoot`, and then only inside an `os.Root`. Directories are made 0700 and files 0600 (`disk.go:68, 237, 243`), and a root others can write to is refused (`disk.go:81`).
- A decoder compromise is still in-process compromise of the host. That is inherent to the design, and the fuzzed decoders are the mitigation.

**F8: The disk cache is a timestamped record of where the listener has looked.** InfoSec
- Severity: **Minor**
- Evidence:
  - `<hash>/<z>/<x>-<y>.pbf` (`disk.go:38-40, 142-145`) plus mtimes gives the viewed areas, to the hour.
  - The hash is unsalted (`disk.go:138`), so against a closed source list (watchpost FR-3.8) each hash maps to its source trivially.
  - The root is checked for others writing to it, not reading it (`disk.go:81`). A pre-existing 0755 root exposes only `v1/`, because the directories below it are 0700.
  - Purge unlinks files; APFS snapshots and backups keep them.
- Simplify/Delete? N
- Action: contract guidance: put the root under `os.UserCacheDir`, which macOS excludes from Time Machine; purge is not secure erasure; the root should be 0700. Optionally refuse a root others can read, as is done for one they can write.

## 7. Dependencies scanned before shipping? (SEC-04)

**F9: The vulnerability scan exists, but nothing records it at a tag, and it runs on one machine.** InfoSec
- Severity: **Minor**
- Evidence:
  - `scripts/gate:223-229` runs `govulncheck`, unpinned `@latest`, on the local toolchain only.
  - The docs lane skips it (`scripts/gate:93`).
  - There is no hosted workflow: `.github/` holds only `copilot-instructions.md` (L-6.1).
  - v0.1.0 was tagged with no close-out (D-18), so no scan at the tag is on record.
  - The toolchain floor is `go 1.25.0` (`go.mod:3`), and a host's build uses its own standard library.
- The direct dependencies are small (two, `go.mod:5-8`).
- Simplify/Delete? N
- Action: the release checklist records `govulncheck` output **at the tag commit**, with the scanner's version pinned. The contract tells hosts that standard-library fixes come from their toolchain.

## Verdict

**Proceed to PLAN, conditionally.** The v0.1.0 base is unusually disciplined: confinement, `os.Root`, errors that never repeat an address, parsers bounded before use. What falls short is the requirements text:
- F1, F2, F3, F5 and F7 are threats wave 1 partly saw that never reached L-1, L-7, L-9 or L-10.
- F4 was not recorded at all.

Amend those requirement rows (a HUM LEAD ruling, not mine) before PLAN designs signatures against them.

**Fix first: F1.** D-20's budget, L-1's per-frame validation and M2's loop honesty all rest on the library owning the frame bytes it validated. Today it does not.
