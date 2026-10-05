# round1 hygiene docs — reviewer report, verbatim

Filed under v0.2.0 D-38/round 2 (H-9): the reviewer's final message exactly as returned, except that machine paths are replaced by `<workspace>` and `<scratch>`. The harness refused reviewers' own report files, so this is the only copy.

---

I'd **not proceed to PLAN yet**. The analysis is careful and the rulings are well kept, but the brief that PLAN designs against, and issue #2 (identical to it), is now wrong or incomplete. The fix is a short amendment, not new work.

`report.md` was not written: the harness blocks subagents from writing report files. The full report is below. My scratch dir holds only a throwaway clone.

What I ran: the gate, fuzzing and full test sweeps were not run, as instructed. From a scratch clone, `go build ./...` at the root and `go test ./tools/atlas` both pass, and regenerating the atlas leaves `06_docs/architecture-atlas.html` unchanged. `gh issue view 2` worked, and the issue body matches the brief word for word, so every brief finding also applies to issue #2.

## Section 1 — Project Hygiene (staff engineer, Monday, no handover)

**Q1. Anything in the tree that should not be?**

- **H-1. The whole v0.2.0 record exists on one disk only.** `origin` has only `main` and `release/v0.1.0`; `feature/radar-loops` is 241 commits ahead of `origin/main`, with no upstream. The public issue #2 cites commit `532806b`, which nobody else can resolve. The record's own premise is durable context.
  - Severity: Important
  - Evidence: `rulings.md:30`, `00-REQUIRED-READING.md:12-14,20`
  - Simplify/Delete? N
  - Action: push both `feature/*` branches, or rule on why they stay local.
- **H-2. Two local branches are stale and misleading.** Local `main` is at `4185587` ("initialize repository") while `origin/main` is `bbc039a`. Local `release/v0.1.0` has diverged from its remote (126 ahead, 1 behind).
  - Severity: Minor
  - Evidence: `00-REQUIRED-READING.md:20` (the branch model these refs break)
  - Simplify/Delete? Y
  - Action: reset local `main` to origin; delete or rename the diverged release branch.
- **H-3. A 14 MB test binary `go-tuimaps.test` sits at the repository root.** It is ignored and untracked, but it is in the working tree.
  - Severity: Minor
  - Evidence: `.gitignore:17`
  - Simplify/Delete? Y
  - Action: delete it.

No secrets found.

**Q2. Does the build work from a clean clone, with the documented commands?** It builds (see above); the gate was not run.

- **H-4. The only documented build and verify commands are in a feature's required-reading file.** `README.md` never mentions `scripts/gate`, `govulncheck`, the go1.25.0 toolchain download or the emulated-amd64 leg.
  - Severity: Minor
  - Evidence: `README.md:1-100`, `00-REQUIRED-READING.md:34-40`
  - Simplify/Delete? N
  - Action: add a develop/contributing section that points to the gate and what it needs.

**Q3. Are the gates actually run, on every path that claims them?**

- **H-5. No gate run leaves a trace, so "full gate before commit" can't be checked and M6 can't be counted.** D-15 says the specimen commit "took the full gate", D-21 says "full gate runs before commit", and the required reading says "green in about seventeen minutes". No log, count or commit trailer records any of it. M6 is "consecutive clean full runs" and has nothing to count from.
  - Severity: Important
  - Evidence: `rulings.md:32,38`, `00-REQUIRED-READING.md:29-30`, `project-brief.md:194`
  - Simplify/Delete? N
  - Action: have the gate append its result and HEAD to a tracked run ledger or a commit trailer.
- **H-6. The docs lane can report success on the wrong change.** It compares the working tree with HEAD, so run after a commit it prints "nothing has changed" and exits 0. Nothing ties a commit to the lane it took, so the rule still depends on memory, which D-15 set out to remove.
  - Severity: Minor
  - Evidence: `scripts/gate:120-123`, `rulings.md:32`
  - Simplify/Delete? N
  - Action: exit non-zero on an empty change, or check `HEAD~1..HEAD`.
- **H-7. What the lane runs differs from D-15.** D-15 names six test files that read documents. The script runs every test, without vet and without race. That covers more tests, but it is not what was ruled.
  - Severity: Minor
  - Evidence: `rulings.md:32` vs `scripts/gate:207-210`
  - Simplify/Delete? N
  - Action: add a clerical ruling row recording the implemented scope.
- **H-8. Nothing checks that the generated atlas page matches the documents.** A Markdown change to a mermaid block passes the docs lane and leaves `architecture-atlas.html` stale; commit `7feb23a` edited both by hand. It is in sync today.
  - Severity: Minor
  - Evidence: `architecture-atlas.html:177`, `L2-gates.md:17`
  - Simplify/Delete? N
  - Action: a test that regenerates the page to a temporary file and compares it.

**Q4. Is anything owed that has quietly stopped being true?**

- **H-9. The gate's NOT RUN line still says no tag exists, but `v0.1.0` exists.** Wave 1 flagged it; it is still there and in no tracker.
  - Severity: Minor
  - Evidence: `scripts/gate:271`, `wave1-findings.md:122-123,196-198`
  - Simplify/Delete? N
  - Action: fix it, or list it as a tracked item.
- **H-10. F-2's trigger has fired, but the row still says "quality pass".** F-2 is due "before v0.2.0 adds a second radar source and its colour table", and D-19 has now ruled that the MRMS table ships.
  - Severity: Important
  - Evidence: `follow-ups.md:13`, `rulings.md:36`, `project-brief.md:198-201`
  - Simplify/Delete? N
  - Action: bring F-2 to the HUM LEAD as a PLAN-entry question now.
- **H-11. Owed work has no home.** `follow-ups.md` says it "is the record" (lines 3-5) but holds only F-1 and F-2. These live only in rulings text or finding tables:
  - D-17's specimen, owed before PLAN;
  - D-18's "one row in v0.1.0's record", which is not written (nothing in `go-tuimaps/07-readiness/` mentions v0.2.0), and D-18's tag-refusing gate test;
  - S29-3, recorded but not diagnosed;
  - the six v0.1.0 defects in wave 1 synthesis point 11;
  - D-19's candidate test for the fallback share.

  Details:
  - Severity: Important
  - Evidence: `rulings.md:34-36`, `specimens/README.md:318`, `wave1-findings.md:196-198`
  - Simplify/Delete? N
  - Action: an owed-items list in the release record, one row per item, each with the phase it is due.
- **H-12. The brief still says every commit is blocked and still promises a reproduction.** Both stopped being true at D-8 and D-9.
  - Severity: Minor
  - Evidence: `project-brief.md:108-116,171-172`
  - Simplify/Delete? Y
  - Action: amend L-6.4 and C-6.
- **H-13. The ratified P10 exemptions and their evidence do not survive a clone.** The D-21 and D-22 entries sit in the ignored `.a2dh-p10-exemptions.yml`, and the four-OS P10 output is filed nowhere. The ruling rows are the only durable copy.
  - Severity: Minor
  - Evidence: `.gitignore:28`, `rulings.md:38-39`
  - Simplify/Delete? N
  - Action (a recommendation to the HUM LEAD, not a disposition): file the run output in the record, or rule that the rows are enough.

**Q5. Does the record say what happened, or what was intended?**

- **H-14. The gate diagram promises an hour-long fuzz run before a release; no such mode exists.** Commit `7feb23a` rewrote that label and kept "(an hour before a release)".
  - Severity: Minor
  - Evidence: `L2-gates.md:17`, `architecture-atlas.html:177`, `scripts/gate:64-85`
  - Simplify/Delete? N
  - Action: mark it NOT BUILT, or build the mode.
- **H-15. D-18 reads as done but is intent.** It says "Recorded as not recoverable: one row…", and that row does not exist.
  - Severity: Minor
  - Evidence: `rulings.md:35`
  - Simplify/Delete? N
  - Action: write the row, or reword the ruling cell as owed.
- **H-16. The first record commit landed before the gate fix that "blocks every commit".** `493f60f` (docs only) precedes `7feb23a`.
  - Severity: Minor
  - Evidence: `project-brief.md:115`, `rulings.md:21`
  - Simplify/Delete? N
  - Action: add a clerical note saying whether D-4 covered documentation.

## Section 2 — Docs Quality (the reader who has not read the code)

**Q1. Does each document answer its title's question?**

- **D-1. The brief no longer states the requirement set PLAN designs against.** None of these rulings' requirements is in the brief or in issue #2:
  - D-14: the tint blends over the radar, and the no-colour defect becomes a requirement;
  - D-17: severity as a word and a dash;
  - D-18: the version string and the tag test;
  - D-19: the observed-palette MRMS table;
  - D-20: the per-map image budget.

  A PLAN reader has to rebuild the requirement set from a 23-row rulings table.
  - Severity: Important
  - Evidence: `project-brief.md:49-143,219`, `rulings.md:31-37`
  - Simplify/Delete? N
  - Action: amend the brief and update issue #2.
- **D-2. "Blocking right now: Nothing" hides what gates DISCOVER exit.** D-17's owed specimen and D-14's "pass on both grounds" both stand before PLAN commits, and no light-ground specimen exists.
  - Severity: Minor
  - Evidence: `00-REQUIRED-READING.md:27-30`, `rulings.md:34`
  - Simplify/Delete? N
  - Action: list what gates DISCOVER exit.

The other documents answer their titles.

**Q2. Does anything contradict the code or another document?**

- **D-3. The README promises a user-agent token the host cannot set.** It says the user-agent carries "a token you set", but the code always passes `fetch.Options{}` and no exported call sets `Token`. The brief says the opposite, correctly. Wave 1's list of "five false claims" missed this, so M5's scope is wider than recorded.
  - Severity: Important
  - Evidence: `README.md:68-69` vs `tiles.go:52`, `internal/fetch/fetch.go:44-45`, `project-brief.md:122-124`
  - Simplify/Delete? N
  - Action: correct the README now; add the claim to L-4/L-7.
- **D-4. The brief says a purge call is missing; `Map.Purge` exists.** Wave 1 said the brief needed correcting "before it misdirects PLAN". It wasn't, and issue #2 repeats it.
  - Severity: Important
  - Evidence: `project-brief.md:137` vs `tiles.go:145`, `wave1-findings.md:180-181`
  - Simplify/Delete? N
  - Action: restate L-9.2 as Purge's scope.
- **D-5. The brief under-states the hatch collision.** L-8.2 names only Extreme/Severe sharing `╳`; Minor and Unknown also share `╱`.
  - Severity: Minor
  - Evidence: `project-brief.md:130-131` vs `internal/render/hatch.go:27-36`
  - Simplify/Delete? N
  - Action: amend L-8.2.
- **D-6. The gate docs say a count budget "creates no deadline at all", but the script sets one.** Every fuzz leg runs with `-timeout 10m`. A leg calibrated to about 60 s here fails on a runner about ten times slower — the same class of accidental failure, and hosted CI (L-6.1) is in scope.
  - Severity: Important
  - Evidence: `scripts/gate:41-42,85,220`, `gate-fuzz-deadline.md:60-61`
  - Simplify/Delete? N
  - Action: state the remaining deadline, and give PLAN a calibration rule for hosted CI.
- **D-7. The HR-to-L mapping is wrong as written.** "HR-1..HR-10 are L-1..L-10", but L-6 is no HR, and L-7..L-10 map to HR-6, 7, 8 and 10.
  - Severity: Minor
  - Evidence: `00-REQUIRED-READING.md:24`
  - Simplify/Delete? Y
  - Action: state the actual mapping.
- **D-8. Small contradictions, fixable in one clerical pass.** Severity: Minor. Simplify/Delete? Y.
  - C-5 says "five paths" and lists four (`project-brief.md:166-169`).
  - C-8 comes before C-7 (`:174-181`).
  - The frontmatter phase still says "pre-DISCOVER" (`:4`).
  - S29-4 cites W1-B for the image cap, which is W1-A/C-3 (`specimens/README.md:319`).
  - The specimens README header says specimens come from a program that is "not the library"; specimen 29 was drawn by the library's public calls (`README.md:7` vs `:306`).
  - v0.1.0 rulings are cited without the "v0.1.0" prefix the rulings file's own convention requires (`rulings.md:12-13` vs `:31`).

**Q3. A claim nothing verifies, in a document that reads as verified?**

- **D-9. An unconfirmed cause is stated as fact.** The investigation says Hypothesis 1 is "neither confirmed nor refuted". The gate diagram and the script comment both state it as fact.
  - Severity: Important
  - Evidence: `gate-fuzz-deadline.md:39` vs `L2-gates.md:17`, `scripts/gate:33-42`
  - Simplify/Delete? N
  - Action: reword both to "probable, not reproduced".
- **D-10. Wave 2's "71–80 %" doesn't match its own table.**
  - The table works out to 80.0–80.6 % (7,243/9,028; 20,397/25,310; 6,447/8,056); nothing shown gives 71.
  - Finding 2 says "no table sampled from the legend can match exactly", but the same table shows 59–61 % of pixels are exactly a legend colour. The misses come from the 256-entry sample, not the legend.

  Details:
  - Severity: Minor
  - Evidence: `wave2-measurements.md:25,29,37-38`, `rulings.md:36`
  - Simplify/Delete? N
  - Action: correct the range and restate finding 2.
- **D-11. None of the measurements can be re-run.** Wave 2 and S29-5/S29-6 came from throwaway programs kept outside the repository, and no raw output is filed. The documents still read as measured. D-22's "mutation-proven" has no record either.
  - Severity: Minor
  - Evidence: `wave2-measurements.md:11-12`, `specimens/README.md:321`, `rulings.md:39`
  - Simplify/Delete? N
  - Action: file the program or its output beside the numbers.

**Q4. Is the audience order right: designers, then PMs, then engineers?** This is a fork for the HUM LEAD, not a ruling.
- The release's two visual decisions (D-14, D-17) can be reached only through the rulings table and a specimens README filed under v0.1.0's feature directory (`specimens/README.md:304`).
- The brief goes from a PM summary straight to requirements full of `file:line` references (`project-brief.md:20-143`).
- The required reading opens with the branch (`00-REQUIRED-READING.md:18-25`).
- Wave 1's one designer-facing finding, that the tint hides the radar, sits at `wave1-findings.md:148`.

The options:
- (a) Keep the engineering-first order.
- (b) Add a short "what changes on screen" section to the brief, linking specimens 22, 28 and 29.

**Q5. A number published without its blind spots (INST-5)?**

- **D-12. Wave 2 has no blind-spot section.**
  - M-A covers one afternoon, one weather regime, and rain up to about 48 dBZ (that last one is stated). Its "18 frames" count the same time three times at three scales, which flatters "the palette levels off".
  - M-B is one run per size, on one 18-core machine. It says "MB" without saying heap or RSS, MB or MiB. "PNG bytes account for the rest" doesn't close at 596×304: 2.17 MB plus about 0.25 MB of PNGs is more than the 2.19 MB measured.
  - The 27 ms render names no machine.

  Details:
  - Severity: Important
  - Evidence: `wave2-measurements.md:17-19,34-35,49-58,62-63`
  - Simplify/Delete? N
  - Action: add a limits section to each measurement.
- **D-13. The two blend results measure different pairs.** On the dark ground, S29-5 measures tinted classes against plain ones; on the light ground, S29-6 measures tinted against tinted. D-14 requires "every other class". So "20 % clears 10" on the dark ground is unshown for tinted-to-tinted pairs.
  - Severity: Important
  - Evidence: `specimens/README.md:325-326`, `rulings.md:31`
  - Simplify/Delete? N
  - Action: re-measure both grounds on the same pair set before PLAN tunes the blend.
- **D-14. D-20 under-states the 3 MB default.** It calls it enough for "two [loops] at dot resolution". At about 0.54 MB a loop, 3 MB holds about five.
  - Severity: Minor
  - Evidence: `rulings.md:37`, `wave1-findings.md:63`
  - Simplify/Delete? N
  - Action: state the capacity correctly.
- **D-15. "About 60 s a leg" and "about seventeen minutes" are calibrated on one machine, with no tolerance.** Nothing fails when durations drift.
  - Severity: Minor
  - Evidence: `scripts/gate:46-50`, `00-REQUIRED-READING.md:30`
  - Simplify/Delete? N
  - Action: name the machine and set a drift threshold.

## Verdict

**Do not proceed to PLAN yet.** Three things stand in the way, all cheap:
- **The brief and issue #2 are wrong or incomplete.** Purge, L-6.4 and L-8.2 are out of date, and five rulings' requirements are missing (D-1, D-4).
- **Owed work has no tracker** (H-11).
- **The v0.2.0 record exists on one disk only** (H-1).

**Fix first: D-1.** Amend the brief and issue #2 to carry D-14..D-20's requirements and wave 1's corrections. It is the one document PLAN reads, and right now it points PLAN at a purge call that already exists and leaves out five ruled requirements.
