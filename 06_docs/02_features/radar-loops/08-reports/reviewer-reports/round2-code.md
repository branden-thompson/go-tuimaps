# round2 code — reviewer report, verbatim

Filed under v0.2.0 D-38/round 2 (H-9): the reviewer's final message exactly as returned, except that machine paths are replaced by `<workspace>` and `<scratch>`. The harness refused reviewers' own report files, so this is the only copy.

---

## Red team round 2: code quality, `ff51ccc..11d986a`

**Verdict: do not ship.** Most round-1 fixes are in place, but the tooling half of R1-16 has one fix that does not hold on this repository. The gate also has three ways to report green without having checked anything, two of them new in this range.

### Job 1: do the round-1 fixes hold?

| Item | Holds? | Evidence |
|---|---|---|
| D-31(1) wall-clock limit on fuzz legs | Mostly. The KILL fallback cannot fire (F-4) | `scripts/gate:168-196`; the `TestAStalledFuzzLegFails` test would hang without the fix |
| D-31(2) docs lane fails with nothing to check | Yes, and tested | `scripts/gate:150-153` |
| D-31(3) a module Go cannot list fails the gate | Yes, and tested. The same fix was not applied to the licence loop (F-2) | `scripts/gate:264-268` |
| D-31(4) each git command's exit status is checked | The code does it; **no test** | `scripts/gate:134-141` |
| D-31(5) unit cases in `internal/mvt` | **Yes.** I reverted the fix in scratch and both new cases fail | `internal/mvt/feature_test.go:64-73` |
| D-31(6) run log | The code does it. Only the green path is tested (F-6) | `scripts/gate:112-121` |
| R1-10 "judges the tree, not the commit" | **Dropped without a ruling** (F-5) | — |
| R1-16 / H-9: NOT RUN names the last tag | **Does not hold on this repository** (F-1) | `scripts/gate:344` |
| R1-16: atlas renamed, split, drift test | Yes, but the fix the test tells you to run fails (F-7) | `tools/atlas/main_test.go:54` |
| R1-16: count table held to fuzz targets; helper duplication removed | Yes | `gate_test.go:315-370` |

### Findings

**F-1. The NOT RUN line still says "none exists yet" on this repository, which has a `v0.1.0` tag.**
- The fix uses `git describe --tags --abbrev=0`. That only finds tags that are ancestors of HEAD.
- `v0.1.0` is not an ancestor of `11d986a` (I checked with `git merge-base --is-ancestor`), so `git describe` fails and the line falls back to "none exists yet".
- Under D-1, releases are squash-merged, so a release tag will never be an ancestor of a work branch. This fix cannot work under that branching model.
- `TestNotRunNamesTheLastTag` (`gate_test.go:376-393`) tags HEAD in its planted tree, so it passes where the real repository fails.
- Severity: **Important**. Evidence: `scripts/gate:344`, `gate_test.go:380`. Simplify/Delete? N.
- Action: use `git tag --list 'v*' --sort=-v:refname | head -1`, and have the test tag a commit that is not an ancestor of HEAD.

**F-2. The licence check checks nothing in every nested module that requires the library.**
- `GOWORK=off go list -m ... all 2>/dev/null` fails in `cmd/tuimaps`, `examples` and `tools/oracle` with "unknown revision v0.0.0". I ran it directly.
- The loop then runs zero times and the module passes.
- So `tools/oracle`'s third-party dependency (paulmach/orb) never has its licence checked.
- This predates the range. It is the same "an error read as an empty list" class that D-31(3) fixed four lines above.
- Severity: **Important**. Evidence: `scripts/gate:306-311`. Simplify/Delete? N.
- Action: capture the command's exit status and fail on error, drop `GOWORK=off` (or run it in workspace mode), and require at least one module in the list.

**F-3. Combined flags give a green run with no legs.**
- `--quick --fuzz`: `fuzzonly=1` skips vet and tests (`:278`), `quick=1` skips fuzzing (`:283`), and both flags skip everything else (`:295`, `:318`). The run prints "gate: green" and logs green under whichever flag came last.
- `--docs --fuzz` runs only the tests but logs mode `fuzz`.
- `--fuzz` on its own has a similar hole. `go test -list` errors are discarded (`:285`, `2>/dev/null`, inside `$(...)`). A package whose tests do not compile yields zero fuzz legs, and the run is green, because no test leg runs in this mode to catch it.
- Severity: **Important**. Evidence: `scripts/gate:94-101`, `:278-294`. Simplify/Delete? N.
- Action: exit 2 when more than one mode flag is given. Fail when `go test -list` fails. In `--fuzz` mode, fail if no fuzz leg ran.

**F-4. `limited`'s KILL fallback can never fire in the case it exists for.**
- When the leader process exits on TERM, the parent sends TERM to the watchdog (`:188`). The watchdog is in a foreground `sleep 5`, so its trap runs after the sleep and exits 0, skipping `kill -KILL` (`:182`).
- I reproduced this in scratch. With a 2-second limit and a child that ignores TERM, the child survived, printed its output, and the leg took 20 s instead of about 2 s. Because the survivor holds the pipe that `leg`'s `$(...)` reads from, the limit is not bounded.
- There is also a small race: if TERM reaches the watchdog before `nap` is set, the trap trips over `set -u`, and `sleep` can be orphaned holding that pipe open.
- Severity: **Minor**. Go's own processes die on TERM (go1.25.0 only handles SIGINT and SIGQUIT), so this is rare in practice.
- Evidence: `scripts/gate:174-189`. Simplify/Delete? **Y**.
- Action: either delete the fallback, or once the watchdog has fired, `wait "$dog"` without sending it TERM. Also replace the flag file (`:171`, `:179`, `:190`) with the watchdog's own exit status: that removes the `mktemp`/`rm` dance entirely.

**F-5. The run log cannot say which tree it tested, and an R1-10 item disappeared.**
- The Commit column is HEAD before the commit.
- The docs lane only works before committing (`:150-153`), so every docs-lane line is written on an uncommitted tree by design (the committed log shows 2 to 18 uncommitted files).
- M6 counts consecutive clean full runs from lines that name neither the tested tree nor the commit that followed.
- R1-10 listed "judges the tree not the commit". D-31's six items leave it out, and the ledger still marks R1-10 "all six fixed".
- Severity: **Important**. Evidence: `scripts/gate:116-118`, `06_docs/gate-runs.md:8-15`, the R1-10 row of `red-team-discover.md`. Simplify/Delete? N.
- Action: log a tree identity, e.g. `git stash create || git rev-parse HEAD^{tree}`. Put the dropped item back to the human lead for a ruling.

**F-6. `TestEveryRunIsLogged` only checks a green docs-lane run.**
- A mutant that always writes "green", or logs the wrong mode or dirty count, survives. The failing-gate tests have no log file planted.
- The Result column is the one M6 counts.
- Severity: **Minor**. Evidence: `gate_test.go:286-313`. Simplify/Delete? N.
- Action: add a failing run to that test and assert the line says `FAILED`.

**F-7. The drift test tells you to run a command that fails.**
- The message says to run `go run ./tools/atlas` from the repository root. `tools/atlas` is its own module, and outside the gate's temporary workspace the command fails with "main module does not contain package" (verified).
- Running it from inside `tools/atlas` writes to a `06_docs/` that does not exist there.
- The same instruction appears in `tools/atlas/main.go:14` and `template.html`.
- Severity: **Minor**. Evidence: `tools/atlas/main_test.go:43,54`. Simplify/Delete? N.
- Action: document a command that works (`cd tools/atlas && go run . -root ../..`, with a root flag), or document `GOWORK`.

**F-8. D-31(4) was ruled "test-first" but has no test.** Removing either `if !` check survives the test suite. Severity: **Minor**. Evidence: `scripts/gate:134-141`. Action: add a test (for example, point `GIT_DIR` at a corrupt index), or record that no test exists.

**F-9. STALLED is the wrong word.** The limit is a wall-clock cap, not a check for progress. A leg that is just slow (a grown corpus, or a slower machine) also gets reported as "STALLED". Severity: **Minor**. Evidence: `scripts/gate:44-45,192`. Simplify/Delete? N. Action: say "exceeded its time limit".

**F-10. The environment can change the gate without leaving a trace.**
- `GATE_FUZZ_LIMIT` is not validated and not logged. A very large value switches the limit off.
- `GATE_ROOT` is not logged either.
- Severity: **Minor**. Evidence: `scripts/gate:87`, `:118`. Action: log both whenever they are set.

### The axis

1. **Understandability.** Comments are good. Problems a reader will hit: the log's "Commit" column (F-5), the STALLED name (F-9), the wrong rebuild instructions (F-7), and `scripts/gate:59` is a single unwrapped line.
2. **Maintainability.**
   - The tag lookup breaks under the branching model already in use (F-1).
   - The atlas drift test runs in the docs lane, and the regenerated `.html` is not Markdown. So every diagram edit now needs the full gate (`tools/atlas/main_test.go:44`, `scripts/gate:143`). That is a lane-policy decision for the human lead, not something I'm ruling on.
3. **Complexity.** `limited` uses perl, a flag file and a watchdog in a subshell. Using the watchdog's exit status, and possibly `set -m` for process groups, would be simpler (F-4).
4. **Necessity.** These could be removed:
   - The flag file (`scripts/gate:171,179,190`).
   - The dead KILL fallback (`:181-182`), unless it is made to work.
   - The default-count branch and its note (`:287-290`). `TestFuzzCountTableMatchesTheTargets` now requires every target to have a row, so the gate could fail on an uncalibrated target instead of quietly running the default.
   - `build`'s returned count (`tools/atlas/main.go:100`), which is the same number as `len(found)` already written into the page.
5. **Name semanticism.** STALLED (F-9). "Commit" in `gate-runs.md` means the parent commit (F-5).
6. **Code documentation.** The comments describe current state, and the history paragraphs were cut. The comment for `limited` claims "stops the whole group", which F-4 shows is false. The rebuild instructions are wrong (F-7).
7. **Testability.** Adequate. The gaps are F-1 (the test's setup hides the bug), F-6 and F-8.
8. **P10.**
   - `compose` sorts its caller's slice in place (`tools/atlas/main.go:138`), a hidden side effect.
   - `collect` and `compose` contain no assertions. D-21 already flagged P10-05 (assertion density) on this file, so the split probably pushed that number back down. I did not measure it.
   - `fmt.Printf`'s result is ignored (`main.go:93`, P10-07).
9. **Evidence soundness.**
   - F-1 is a false statement in the output.
   - F-2 skips a check and still passes.
   - F-3 prints green with nothing run.
   - In `--fuzz` and `--quick` modes, "gate: green" is followed by the full-gate NOT RUN line (`scripts/gate:343-345`). That line does not mention the vet, tests, vulnerability scan or cross-compiles those modes skipped.
10. **Evasion lens.**
    - **Full:** the licence loop (F-2).
    - **Fuzz:** a package whose tests do not compile gives zero legs (F-3); `GATE_FUZZ_LIMIT` is unlogged (F-10).
    - **Combined flags:** zero legs and green (F-3).
    - **Docs:** sound, apart from tree versus commit (F-5).
    - **`trap`:** `rc` is kept correctly. I tested that a trap on EXIT is not inherited into `( )` or `$(...)`, so the module subshells do not log duplicate lines.
    - **`limited`:** a nonzero exit status is always kept, and a stall returns 124, so a failing leg always fails. The only weakness is F-4's unbounded limit.

**Ship / do not ship:** do not ship. **Fix first: F-3.** The gate must refuse combined mode flags and fail when a fuzz run finds no targets or `go test -list` fails. It is the only defect that prints and logs "green" with nothing checked. F-1 and F-2 are close behind: one is a ledger claim the tree does not have, the other a check that never ran.

Everything I made is in `<scratch>/rt2-code`: the F-4 reproduction (`exp1.sh`) and the subshell trap check (`exp2.sh`), plus the `repo/` copy used for the mvt mutation.
