# Release checklist — what is done before v0.1.0 is tagged

Up: [readiness](.) · Plan task 14.14 · Carries: D-19, D-13, D-50, NFR-1, NFR-9, NFR-22

| Field | Value |
|---|---|
| Status | **Not started.** Nothing here is done until SHIP, and SHIP is HUM LEAD's to open |
| Who runs it | The coordinator prepares each line and shows the result; HUM LEAD approves the tag |
| The rule that shapes it | **No remote exists until SHIP (D-19).** Every line below is run against a *fresh clone* of the repository, never against the working tree, so that what is released is what is committed and nothing else |

## 0 · Before anything

| # | Check | How |
|---|---|---|
| 0.1 | The sitting is done and recorded | [`uat-record.md`](uat-record.md) is filled in: M1a judged, the frames approved, the description heard, the no-colour reading taken |
| 0.2 | Every gate is green on a fresh clone | clone this repository to a directory outside it, and run `scripts/gate` there |
| 0.3 | The structure check reads 18 of 18 | the framework's code-quality check, then its structure check |
| 0.4 | The build log's last row is this release | no task left with "owed" against it |
| 0.5 | The first host's spike and its written review are in | plan task 14.19, [the template](integration-review-template.md) |

## 1 · What is in the tree

| # | Check | How |
|---|---|---|
| 1.1 | The fixture is unchanged | `go test ./internal/testkit -run TestFixturePinned` |
| 1.2 | The embedded tiles match their hash list, and the notice says what it must | `go test ./assets` |
| 1.3 | Licence files are present and cover everything | `LICENSE` at the root covers every module of the repository; `assets/NOTICE` covers the map data, which the code licence does not |
| 1.4 | No workspace file is tracked | `git ls-files go.work go.work.sum` prints nothing |
| 1.5 | The dependency graph names only what was ruled | `go test . -run 'TestAllowList\|TestHostIndependence'` — `go-runewidth` and `uax29/v2`, and no terminal library (D-75, D-81, M5) |
| 1.6 | The public surface is the one written down | `go test . -run TestPublicSurfaceSnapshot` |
| 1.7 | Nothing in the tree carries an address or a name that should not leave it | the exposure scan over the whole tree, `.pbf` files exempted |

## 2 · What is built

| # | Check | How |
|---|---|---|
| 2.1 | Every target builds with the C toolchain off (NFR-1) | the gate's cross-compile leg: darwin/arm64, darwin/amd64, linux/amd64, linux/arm64, windows/amd64 |
| 2.2 | The app is built to `dist/` and nowhere else | `scripts/uat` writes `dist/tuimaps`; `dist/` is not tracked |
| 2.3 | **Checksums** for every artefact | `shasum -a 256 dist/* > dist/SHA256SUMS`, recorded in the release notes |
| 2.4 | **The binary is scanned** before it is offered | `strings dist/tuimaps` carries no path from this machine, no personal name, no token, and no address but the tile source's; `go version -m dist/tuimaps` names only the ruled modules |
| 2.5 | The vulnerability scan is clean on the newest toolchain | the gate's own leg, which names the toolchain it ran on |

## 3 · The tags, in order

The library is one module and the rest are nested modules of the same repository, so **the library is tagged first**: a nested module's own tag resolves against a library version that must already exist.

| # | Tag | Why this order |
|---|---|---|
| 3.1 | `v0.1.0` | the library itself, at the repository root |
| 3.2 | `cmd/tuimaps/v0.1.0` | the app, which requires the library at the version just tagged |
| 3.3 | `examples/v0.1.0` | the examples, likewise |
| 3.4 | `tools/answer-key/v0.1.0`, `tools/gen-assets/v0.1.0`, `tools/oracle/v0.1.0` | the tools, which are test and generation aids and are tagged so that a reader can fetch the exact ones |

Before 3.2 and after 3.1: **the `replace` directives come out and the required version goes in.** Tracked module files carry no `replace` at any point (plan task 00.2); what changes at SHIP is the version each nested module requires, from the placeholder `v0.0.0` to `v0.1.0`.

## 4 · What is said with it

| # | Check | How |
|---|---|---|
| 4.1 | The README says what a host needs before it starts | the braille font, that nothing is fetched until a source is named, who owns the work and the clock, what is sent and stored, the credit the data asks for, and what is not built yet |
| 4.2 | Compatibility is stated (NFR-22) | the error kinds and warning kinds are closed lists; a later version adds to them only in a minor release, and says so |
| 4.3 | The release notes carry the measured figures, not promises | M1, M2 (cold 2.64 s, warm 900 ms), M3 (62 of 62 parity rows), NFR-3 (live 2.16 MB, peak 6.20 MB), NFR-4 (a marker turn at 48 allocations whatever is on the map) |
| 4.4 | What is **not** run anywhere is said plainly | Linux and Windows test runs, real amd64 hardware, the nightly soak — the gate prints them as NOT RUN and the notes repeat it |
| 4.5 | The credit the map data asks for is in the notes as well as on the map | `assets/NOTICE`, in full, with its links |

## 5 · The push

| # | Check | How |
|---|---|---|
| 5.1 | **Nothing is pushed without HUM LEAD saying so** (D-19, D-28, D-50) | the coordinator never pushes; the remote is named at SHIP and not before |
| 5.2 | It is published from a **fresh clone**, not from this working tree | clone to a directory outside this one, and push from there |
| 5.3 | The identity on every commit is the personal one | `git log --format='%an <%ae>' | sort -u` shows one author, the personal address |
| 5.4 | Sole-author commits; no tool-generated trailers or watermarks | read every commit message of the release and confirm that none carries a trailer naming anything but its author |
