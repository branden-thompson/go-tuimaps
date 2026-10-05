# round2 infosec — reviewer report, verbatim

Filed under v0.2.0 D-38/round 2 (H-9): the reviewer's final message exactly as returned, except that machine paths are replaced by `<workspace>` and `<scratch>`. The harness refused reviewers' own report files, so this is the only copy.

---

## InfoSec red team, round 2: go-tuiMaps v0.2.0 DISCOVER exit

I worked read-only. The scratch directory `<scratch>/rt2-infosec` was created and is still empty. I did not run the gate, fuzzing or tests, and made no network requests.

### Job 1: do the D-30 rows close F1–F9 as written?

| F | Row | Closes? |
|---|---|---|
| F1 | L-1.14 | **Gap** (S-1) |
| F2 | L-1.15, L-12.4 | Yes. Residual in S-7 |
| F3 | L-7.3 | **Gap**: the wording cannot be met by `Checked` (S-2) |
| F4 | L-7.4 | Yes, apart from in-flight jobs (Minor, S-8) |
| F5 | L-9.3, L-9.4 | **Gap** (S-3, S-4) |
| F6 | L-9.5 | Yes |
| F7 | L-10.2 | Yes as worded. It also opens S-5 |
| F8 | L-9.6 | Partly: guidance only, and the code accepts a readable root (S-6) |
| F9 | L-5.5 | **Gap**: "records" a scan, but never requires a clean result (S-9) |

**S-1 — L-1.14 can be met while the F1 exploit still works.** InfoSec. Severity: Important.
- **Evidence:**
  - `internal/overlay/store.go:437` keeps the host's `Overlay` as handed in. `overlays.go:110` (`RadarImage`) copies the struct but shares the `PNG` and `Table` slices.
  - The only check "at decode" is `image.go:248`, and it runs after `png.Decode` (`image.go:242`) has already allocated whatever the swapped IHDR asked for.
  - So the existing code already satisfies "decode re-checks the size cap", and a host mutating the bytes after hand-in still gets a process-killing allocation.
  - As worded the row covers "frames" only. The v0.1.0 single-image path, which L-1.1 keeps "additive", stays borrowed.
  - `Table` is borrowed too. `CheckTable` (`image.go:142`) validates it at hand-in, but `ImageKey` and `matcher` (`classified.go:66`, `image.go:195`) read it later, so a NaN or descending table goes in unchecked.
- **Simplify/Delete?** Y. One copy-then-validate at the hand-in boundary deletes the re-check.
- **Action:** reword L-1.14: "Every host slice an image carries (PNG, table), single image and frames alike, is copied first, and the copy is validated. Any later decode reads the header from the copy before it allocates."

**S-2 — L-7.3 asks `fetch.Checked` for a read limit and timeouts it cannot give.** InfoSec. Severity: Important. This is the one to fix first.
- **Evidence:**
  - `internal/fetch/get.go:149-171`: `Checked` calls `replacement(ctx, r)` and compares `len(data)` only after the host fetcher has returned the whole body.
  - A host fetcher that does `io.ReadAll` on a hostile tile server's 10 GB reply has already allocated it, and "wrapped by `Checked`" is still true.
  - There is no deadline: `Checked` waits for the host function to return. A fetcher that ignores `ctx` pins a worker for ever, and `remote.go:324` blocks with it.
  - `Checked` also returns the host's slice uncopied, and that slice is stored to disk after decode (`pipeline.go:640-645`).
  - `Checked` is still not wired in (`tiles.go:50-58` uses `m.fetcher` raw).
- **Simplify/Delete?** N.
- **Action:** reword L-7.3 as an outcome: "The library bounds the bytes and the time of every host fetch itself, whatever the host fetcher does." That means PLAN chooses a fetcher shape that returns a reader the library limits, and returns on a deadline without waiting for the host function. The rest the host fetcher takes on is contract.

**S-3 — L-9.4's "maximum age" is measured by a timestamp that every read rewrites.** InfoSec. Severity: Important.
- **Evidence:**
  - The file's mtime is the only time it has, and it is the recency clock. `Load` (`disk.go:217-218`) and `Flush` (`disk.go:321-322`) rewrite it on read. `Store` stamps it at write (`disk.go:253`).
  - A tile viewed at least hourly therefore never reaches any maximum age. Built as worded, retention of the location record is unbounded for exactly the places the user looks at most.
- **Simplify/Delete?** N.
- **Action:** L-9.4 says age runs from when the tile was fetched, recorded apart from recency (for example in the file name or a sidecar), and never refreshed by a read.

**S-4 — L-9.3 names Purge only; other paths still write or keep data.** InfoSec. Severity: Important.
- **Evidence:**
  - Jobs snapshot `o.Disk` (`pipeline.go:569`) and `Store` into it afterwards (`pipeline.go:644-645`). So `CacheRoot("")`, where a host turns the cache off for privacy, and `Close` both still get in-flight writes.
  - Replacing the root (`tiles.go:107-111`) never calls `Release` on the old `Disk` and never purges it. `Close` releases only the pipeline (`map.go:453`), so the root file descriptor leaks for each call.
  - `Purge` then empties only the current root, so "everything" misses earlier roots.
- **Simplify/Delete?** N.
- **Action:** extend L-9.3: "No write lands after Purge, CacheRoot change/off, or Close. Replaced roots are released, and the contract says Purge does not reach a root the map no longer holds."

### Job 2: what is new

**S-5 — the private-address check is skipped whenever a proxy is in the environment, and L-10.2 opens a direct-dial bypass.** InfoSec. Severity: Important.
- **Evidence:**
  - `fetch.go:182-184` drops the dial `Control` when `ctx` carries `viaProxy`.
  - The mark is set once per request (`get.go:90-92`) and inherited by redirects.
  - Once L-10.2 exposes an allow-list, a redirect to an allowed host listed in `NO_PROXY` dials directly with no check. That host is then DNS-rebound to `169.254.169.254`.
  - `isPublic` (`fetch.go:144-156`) also misses `0.0.0.0/8`, `64:ff9b::/96` (NAT64), `2002::/16`, `198.18.0.0/15` and `240.0.0.0/4`.
- **Simplify/Delete?** N.
- **Action:** add to L-10.2: decide per connection whether it goes through the proxy, not per request, and finish the reserved-range list. Put both in the contract.

**S-6 — the cache root may be readable by other users; L-9.6 is guidance only.** InfoSec. Severity: Minor.
- **Evidence:** `disk.go:81` refuses group/other *write* only. An existing `0755` root is accepted, and its `v1/<hash>/<z>/<x>-<y>.pbf` names plus host-clock mtimes are a dated record of where the user looked.
- **Simplify/Delete?** N.
- **Action:** refuse or warn on `0o077` bits, and give L-9.6 an instrument.

**S-7 — two SEC-03 gaps.** InfoSec. Severity: Minor.
- **Evidence:**
  - `ReadBack` does `root.ReadFile` before its size check (`disk.go:364-365`), unlike `Load` (`disk.go:212`).
  - The image budget (L-12.1) says "image". Fields (`store.go` `s.fields`) and the shared `Classified` set sit outside the per-map count.
  - `ImageKey` converts `int64(v*1e6)` (`classified.go:58,63,67`). Table values that are valid but huge saturate, so two different tables collide on one key, and one map is served another map's classification.
- **Simplify/Delete?** N.
- **Action:** read through the limit; bring fields into scope in L-12.1; hash `math.Float64bits`.

**S-8 — the gate's kill escalation is unreachable, and its flag file has a race.** InfoSec (gate availability). Severity: Minor.
- **Evidence:**
  - `scripts/gate:180-182, 188`: when the group leader dies on TERM, the parent TERMs the watchdog during its `sleep 5`. The trap then exits before `kill -KILL`, so any group member that survived TERM keeps the `$(...)` pipe (`scripts/gate:207`) open and the gate hangs, which is the stall D-31 closed.
  - `scripts/gate:171` creates a temp file with `mktemp`, deletes it, and reuses the name. On the Linux runner L-6.1 introduces, a shared `/tmp` lets another user plant a symlink that `: >` truncates, or force a false STALLED.
  - `perl ... exec @ARGV` (`:172`) goes through a shell if it is ever given one argument. Use `exec { $ARGV[0] } @ARGV`.
- **Simplify/Delete?** Y. Signal the stall through the watchdog's exit status, not a file.
- **Action:** KILL the group whenever the flag is set; drop the file.

**S-9 — the run log and the scan cannot be verified.** InfoSec. Severity: Minor.
- **Evidence:**
  - `scripts/gate:118` appends plain text to a tracked, hand-editable file. It does not record `GATE_FUZZ_LIMIT` or `GATE_ROOT` (`:87`, `:103`), so a run with a limit of 999999, or a run of another tree, logs as "full green". M6 counts from this file.
  - The gate asks for `govulncheck@latest` (`:302`), unpinned.
  - L-5.5 "records" a scan at the tag, but never requires zero reachable findings or a ruled exception.
- **Simplify/Delete?** N.
- **Action:** log any override that is in effect and whether the scan ran; L-5.5 requires a clean result or a ruling.

**S-10 — a reference program patch adds `os.Getenv` to the renderer.** InfoSec. Severity: Minor.
- **Evidence:** `02-analysis/programs/spec29d-f-blend.patch:35` reads `SPECIMEN_ALPHA` inside `internal/render/frame.go`. Nothing in the tree stops environment reads in library code; `look.go:201` is today's only one. A patch applied in-tree for a re-run can ship environment-controlled rendering.
- **Simplify/Delete?** N.
- **Action:** add a test listing the library's permitted `os.Getenv` calls. Separately, the programs README's step 3 says `spec29-render` takes one argument, but the code reads three (`spec29-render.go.txt:41`); that is a docs error, not security.

### The seven questions

1. **Every external input validated and bounded before use? (SEC-03):** No. Host fetch bodies (S-2), borrowed image bytes and tables (S-1), `ReadBack` (S-7).
2. **Authorization and destructive-operation errors (SEC-02):** There is no authorization layer; the host is trusted. Destructive errors are swallowed at `pipeline.go:645` and miscounted at `disk.go:370-371` (both covered by L-9.5), and writes land after cache-off or Close (S-4).
3. **Secrets out of logs, errors and source; vetted crypto?** Clean. Errors name only scheme and host (`get.go:27-41`), the source address is hashed in paths and job keys, TLS is at least 1.2, SHA-256.
4. **Input reaching a query, shell, path or deserializer unescaped? (SEC-01):** Library clean: paths are built from integers under `os.Root`, and the template allows only `{z}`, `{x}`, `{y}`. The gate's `perl` exec is a latent risk (S-8).
5. **Errors leaking structure, paths or stack traces?** No finding.
6. **Blast radius and what the disk cache reveals:** Tile coordinates and mtimes record where and when the user looked. A readable root is accepted (S-6), age is unbounded (S-3), old roots and file descriptors are retained (S-4), and a proxy disables the SSRF check (S-5).
7. **Dependencies scanned before shipping? (SEC-04):** Scanned in the full gate only, with an unpinned tool and no requirement for a clean result (S-9).

### Verdict

**Do not proceed to PLAN until L-1.14, L-7.3, L-9.3 and L-9.4 are reworded** (S-1 to S-4). Each is a text change to a row, not new scope, but as written each one can be met while its exploit still works. S-5 should be added to L-10.2 in the same pass; the rest can go to PLAN.

**Fix S-2 first.** It is the only path a remote party controls: a hostile tile server behind a host's own fetcher. As worded, L-7.3 promises a limit that the `Checked` design in `get.go:149-171` cannot enforce.
