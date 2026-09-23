# round1 code — reviewer report, verbatim

Filed under v0.2.0 D-38/round 2 (H-9): the reviewer's final message exactly as returned, except that machine paths are replaced by `<workspace>` and `<scratch>`. The harness refused reviewers' own report files, so this is the only copy.

---

# RT1 — Code Quality — `feature/go-tuimaps..feature/radar-loops` (650f267..ff51ccc)

(The harness would not let me write `rt1-code/report.md`, so the full report is below. My scratch directory holds only a `probe/` folder from one `go list` check.)

**What I read:** the full diff of `scripts/gate`, `gate_test.go`, `area.go`, `geometry.go` and `tools/atlas/*`; both saved fuzz inputs; rulings D-0 to D-22; `gate-fuzz-deadline.md`; the go1.25.0 toolchain source for `testing` and `cmd/go/internal/test`; and orb v0.13.0 `unmarshal.go`.

**What I ran:** `go test -count=1 -run FuzzArea ./internal/describe` (ok) and `-run 'Feature|Close' ./internal/mvt` (ok). I did not run the gate or any fuzzing.

## 1. Understandability

**F1. The only test that fails without the decoder fix is a 456 KB fuzz input that was never minimized, and it lives in another module.**
- The change at `internal/mvt/geometry.go:161-164` is tested only by `tools/oracle/testdata/fuzz/FuzzAgree/fb1ffe4164f68657` (455,982 bytes). No reader can tell what that file proves.
- `internal/mvt` has no case where a ring is already closed. The rings in `feature_test.go:56`, `:71-73` and `real_test.go:131` are all open.
- So `go test ./internal/mvt` would still pass with the fix reverted. I did not revert it to confirm this.
- The one-point ring's behaviour also changed (two points before, one now). This matches orb's `Ring.Closed`, but no test covers it.
- Severity: Important. Evidence: `internal/mvt/geometry.go:161`, `feature_test.go:55-60`. Simplify/Delete? N.
- Action: add two small cases to `feature_test.go`: a closed ring A,B,C,A followed by ClosePath, and a one-point ring followed by ClosePath. Keep the corpus file.

**F2. The comment says "the proven decoder" without saying which one.**
- It means paulmach/orb, which a reader only learns from `tools/oracle`.
- The comment at `feature_test.go:55` still says ClosePath always re-pushes the first point, which is no longer true.
- Severity: Minor. Evidence: `geometry.go:150-153`, `feature_test.go:55`. Simplify/Delete? N.
- Action: name orb's `decodePolygon` (`unmarshal.go:363`) in the comment, and fix the test comment.

**F3. In the subtest table, an empty string means "delete a file", and only the `switch` explains that.**
- Severity: Minor. Evidence: `gate_test.go:204`, `:209-219`. Simplify/Delete? Y.
- Action: use a struct per case: `{name, want string; mutate func(root string)}`.

## 2. Maintainability

**F4. Nothing checks the fuzz count table against the real fuzz targets.**
- Today its 16 rows match the 16 targets (I checked).
- If a target is renamed, its old row stays unused, and the new name falls to the default with only a `note` line. The note is ratified at D-10.
- Severity: Minor. Evidence: `scripts/gate:63-83`. Simplify/Delete? N.
- Action: recommend to the lead a `gate_test` that every row names a target `go test -list` still reports.

**F5. The P10 exemptions ruled at D-21 and D-22 are in a file git ignores.**
- `.a2dh-p10-exemptions.yml` is matched by `.gitignore:28` (`.a2dh*`) and is not in the branch.
- A fresh clone therefore has none of the ratified exemptions.
- Severity: Minor. Evidence: `.gitignore:28`, `.a2dh-p10-exemptions.yml:213`. Simplify/Delete? N.
- Action: this is the lead's decision — track the file, or accept that the exemptions exist only on this machine.

## 3. Complexity

**F6. Two test helpers duplicate existing code.**
- `runDocsLane` is `runGate` with one flag changed.
- `writeFile` repeats the loop body inside `plantTree`.
- Severity: Minor. Evidence: `gate_test.go:125-137` vs `:47-58`; `:160-169` vs `:37-44`. Simplify/Delete? Y.
- Action: give `runGate` a flag parameter, and have `plantTree` call `writeFile`.

## 4. Necessity (primary lens)

These can be deleted:

- **`FUZZ_TIMEOUT` and the `-timeout "$FUZZ_TIMEOUT"` flag** (`scripts/gate:85`, `:220`).
  - 10m is already `go test`'s default (`testflag.go:68`).
  - The test's timeout alarm is stopped before fuzzing starts (`testing/testing.go:2335-2340`, then `runFuzzing` at `:2363`).
  - cmd/go sets no kill timeout while fuzzing (`test.go:808`).
  - So the flag does nothing, yet it looks like a limit on the fuzz leg. See Q10.
- **The one-element loop `for _, root := range []string{"06_docs"}`** (`tools/atlas/main.go:83`). It was left behind when D-21 dropped `docs/` from the tool.
- **The check `execs == FUZZ_EXECS_DEFAULT`** (`scripts/gate:219`). It decides "uncalibrated" by comparing values, so a calibrated row that happened to equal the default would be reported as uncalibrated. Return nothing for an unknown target and apply the default at the call site.
- **The incident history in the gate comment** (`scripts/gate:36-41`, `:46-48`, `:56-58`). See F9.

Not a deletion candidate:
- `area.go:93-95` is ratified by D-16 (option A). It runs only on input the public calls refuse, and exists so `InArea` and `NearestEdge` agree.

## 5. Name semanticism

**F7. The page the atlas tool generates for go-tuiMaps is titled "Watchpost Architecture Atlas".**
- The template predates this range, but the tool is in scope under D-21.
- Severity: Minor. Evidence: `tools/atlas/template.html:1`, `:58`. Simplify/Delete? N.
- Action: rename it.

Otherwise clean: `check`, `safeID`, `fuzz_execs` and `onGlobe` all mean what they say.

## 6. Code documentation

**F8. The comment calls a skipped edge "not an edge", but skipping it leaves the ring open.**
- Once an edge is skipped, the ring is an open polyline, so "inside" by the even-odd rule means nothing definite for it.
- The code is consistent with `NearestEdge`; only the comment is incomplete.
- Severity: Minor. Evidence: `internal/describe/area.go:90-95`. Simplify/Delete? N.
- Action: say in the comment that the result is defined only for that consistency, and that public calls refuse such input.

**F9. The gate comment tells incident history rather than current state (AP-HIST-01).**
- Examples: "Three targets…failed", "A single five-million count…forty-one", "seventy per cent high".
- The explanation of the mechanism and the recalibration recipe are good and should stay.
- Severity: Minor. Evidence: `scripts/gate:36-41`, `:46-48`, `:56-58`. Simplify/Delete? Y.
- Action: replace the history with a pointer to D-9 and D-10.

## 7. Testability

- **Docs lane:** tested for a Markdown-only change passing, a rejected document failing, and code, untracked and deleted files being refused.
- **Docs lane gaps:** no test for the "nothing changed" branch (`scripts/gate:121-123`), the no-HEAD branch (`:109-112`), a staged-only change, or a rename to `.md`.
- The green test does not check that vet, race and fuzz *did not* run (`gate_test.go:185`), and not running them is the lane's whole point. Minor.
- **Atlas:** `check` is tested case by case, including a set that should pass (`main_test.go:10-36`). That is sound.
- **Decoder fix:** no unit test (F1).

## 8. P10 conformance

- **P10-02 (bounded loops):** clean. Every changed loop runs over a finite slice or count.
- **P10-05:** `check` is exempt under D-22, which is ratified. `closePath` and `crossingsOf` keep their guards.
- **P10-07:** the new `gate_test.go` code checks every error. `tools/atlas/main.go:133,134,140,157` ignore the returns of `fmt.Fprintf` and `fmt.Printf`. That code predates the range and is conventionally exempt, so this is noted only.
- **P10-04:** `run()` is 79 lines (`tools/atlas/main.go:81-159`), but D-21 lists P10-05 as the only atlas finding. Either the meter's limit is above 79 or the meter missed it. Minor. Recommend the lead confirm which.

## 9. Evidence soundness (verifiers fail closed)

**F10. The docs lane passes when there is nothing to check.**
- At `scripts/gate:121-123`, a clean working tree makes `--docs` print "nothing has changed since the last commit" and exit 0, with zero tests run.
- So if someone commits first and then runs `--docs`, an unchecked commit gets exit 0.
- The lane compares the working tree against `HEAD` (`:113`). D-15, though, speaks of "a commit whose changes are all `.md`".
- Severity: Important. Evidence: `scripts/gate:113`, `:121-123`. Simplify/Delete? N.
- Action: exit non-zero with "nothing to check". Put the larger choice to the lead: should the lane judge the commits since the merge base, or the uncommitted change?

**F11. A module that `go list` cannot load is reported as "empty" and passes.**
- `scripts/gate:203` discards `go list`'s exit status.
- Confirmed in scratch: with a malformed `go.mod`, `go list ./... 2>/dev/null` prints nothing and exits 1. The gate would then print "empty no Go packages yet" and `exit 0`.
- Today this is masked because `go work init` (`:176`) fails loudly on the same file. Masked is not closed.
- Severity: Minor. Evidence: `scripts/gate:203-206`. Simplify/Delete? N.
- Action: fail when `go list` exits non-zero, and treat a module as empty only when `go list` succeeds with no output.

## 10. Evasion lens (`scripts/gate`, including `--docs`)

- **Exit status lost:** the `$( { git diff …; git ls-files …; } | sort -u)` at `scripts/gate:113` is never checked. If `git diff HEAD` fails, only untracked files are judged. Minor. Action: capture and check each command separately.
- **Check skipped on a branch:** F10 and F11 above.
- **A limit that looks real but does nothing:** `FUZZ_TIMEOUT` (`:85`, `:220`), read with "creates no deadline at all" (`:42-44`), suggests each fuzz leg is capped at 10 minutes. It is not capped.
  - D-3 recorded the fuzz engine stalling ("0 execs/sec from 18 s").
  - A budget counted in runs never runs out during a stall, so the gate now hangs with no bound instead of failing. It never goes falsely green, but a failure becomes a hang behind a constant that says otherwise.
  - Severity: Important.
  - Action: delete it. If a limit is wanted, wrap each fuzz leg in an external wall-clock kill that fails the leg, and document it.
- **Checks that depend on the environment, or scope narrowed after seeing a count:** none found. The counts are measured and ratified at D-10. The lane refuses every file that is not `.md`. Paths git quotes (non-ASCII) no longer end in `.md`, so they are refused too. The docs lane runs every module's tests, which is more than D-15 requires — the safe direction.

## Verdict

**Do not ship yet.** Both library fixes are correct: the decoder now matches orb, and `InArea` agrees with `NearestEdge`. The atlas checks are sound. But the gate still has paths that report success, or appear to impose a limit, without evidence behind them. Each takes minutes to fix.

**Fix first:** the inert `FUZZ_TIMEOUT` (`scripts/gate:85`, `:220`), together with F10. It looks like a real guard on the exact failure path (L-6.4) this release set out to fix. As things stand, the next fuzz-engine stall hangs the gate indefinitely.
