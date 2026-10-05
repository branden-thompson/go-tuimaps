# round2 hygiene docs — reviewer report, verbatim

Filed under v0.2.0 D-38/round 2 (H-9): the reviewer's final message exactly as returned, except that machine paths are replaced by `<workspace>` and `<scratch>`. The harness refused reviewers' own report files, so this is the only copy.

---

## Verdict

**Do not proceed to PLAN yet.** Most of round 1's rulings did land in `requirements.md`: D-24 through D-30 and D-35 through D-37 all have rows, and the issue #2 body matches the brief plus `requirements.md` exactly, apart from the frontmatter. But one fix marked "Done" is false on this branch. Three rulings reach the record weaker than ruled or contradicted by it. And D-33 undercuts D-32. All of these are about an hour of record edits. There are no Critical findings.

**Fix first:** L-13.3's "Done" (H-3 below). The gate on this branch still prints "none exists yet", and the test that "holds" the fix cannot see the problem.

I worked read-only. I cloned into my own scratch directory and built the root module there (it built). I did not run the gate.

---

## Axis: Project Hygiene

### 1. Anything in the tree that should not be?
There are no secrets and no machine paths: `git grep` for `/Users/`, `/home/` and the user name at 11d986a found nothing. About 1 MB of PNG inputs sits under `programs/inputs/`, which D-32 deliberately allows. The local `feature/go-tuimaps` stays, as D-33 rules, and the 14 MB test binary is gone.

- **H-1 · `02-analysis/wave2-measurements.md:11`, `programs/README.md:6`, `project-brief.md:161`, `requirements.md:212`** · Important · Simplify: N
  - **Evidence:** The record's reproducibility rests on work-branch commits (`ca24461`, `7288375`, `650f267`, `0665ae7`). `git branch -r --contains` finds each one only on `origin/feature/radar-loops`, which D-33 deletes after release. D-33's own consequence says the record must cite rulings and documents, not work-branch hashes, "wherever it must outlive the branch". Once the branch is deleted, "written against the library at `ca24461`–`7288375`" names an API nobody can check out, so D-32's "every number can be re-run" dies at release. rulings.md holds 13 more hashes and the ledger 4 more.
  - **Action:** Anchor the programs to something that survives: a tag, or "the API as of D-32, recorded in `public-surface.txt`". Sweep the hashes that must outlive the branch.

### 2. Does the build work from a clean clone?
A fresh clone builds the root module with `go build ./...`. `cmd/tuimaps` and `examples` build only inside the workspace file the gate writes.

- **H-2 · `README.md:94-107`; `scripts/gate:172`** · Minor · Simplify: N
  - **Evidence:** The "Working on the library" section does not say that the gate needs `perl` (the STALLED watchdog uses it), or that the nested modules only build through the gate's `go.work`.
  - **Action:** Add both, one line each.

### 3. Are the claimed gates actually run on every path?
- **H-3 · `requirements.md:169`; `scripts/gate:344`; `gate_test.go:369-386`** · Important · Simplify: N
  - **Evidence:** L-13.3 says "Done: it names the last tag (D-38), held by `TestNotRunNamesTheLastTag`". But `v0.1.0` sits on the squash-merged release commit, which is not an ancestor of this branch. `git describe --tags --abbrev=0 11d986a` fails with "No tags can describe", so the gate on this branch still prints "(none exists yet)". The test tags HEAD inside a planted repository, so it cannot catch the real topology. D-1 and D-33 make that topology permanent: every work branch will hit this.
  - **Action:** Resolve the last tag with `git tag --sort=-v:refname`, not `describe`. The test should tag a commit that is not an ancestor of HEAD. Reopen L-13.3.
- **H-4 · `08-reports/red-team-discover.md:60` vs `rulings.md:48`; `scripts/gate:124-126`** · Important · Simplify: N
  - **Evidence:** R1-10 lists "the docs lane … judges the tree not the commit (C-F10, H-6)". D-31's six fixes do not include it, yet the ledger reads "all six fixed". The lane still judges the whole working tree, untracked files included. So a partial commit can pass because of a Markdown file it does not carry.
  - **Action:** Either put it before the HUM LEAD as its own ruling, or have the lane judge the index. Right now the finding was dropped without anyone saying so.
- **H-5 · `06_docs/gate-runs.md:4,8-16`; `scripts/gate:116`** · Minor · Simplify: N
  - **Evidence:** The "Commit" column is HEAD while the gate runs, which is the parent of the commit being gated. For example, the full run logged at `aabfcea` with 18 uncommitted files is the run for `11d986a`. The header says "the commit it ran at", and M6 counts from this table. Commits `a8aa384`..`7288375` have no gate evidence at all.
  - **Action:** Name the column "HEAD before commit", or log a hash of the tree. Say in the header that the log starts at D-31.

### 4. Owed items that have quietly stopped being true
- **H-6 · `requirements.md:15,211` (OW-7)** · Minor · Simplify: **Y**
  - **Evidence:** No row carries "OPEN — R1-n", yet OW-7 still says those rows are due before DISCOVER exit.
  - **Action:** Delete OW-7 and the sentence on line 15 that defines the marker.
- **H-7 · `requirements.md:212` (OW-8)** · Important · Simplify: N
  - **Evidence:** A deferral that sets its own due date ("PLAN (a gate leg, test-first)"), sourced to "D-38 batch". D-38's ruling text (`rulings.md:55`) says nothing about gofmt. D-23 (`rulings.md:40`) requires every change to `requirements.md` to be a row in the rulings log. The standing rule is that a reason is ratified, never self-issued.
  - **Action:** Get a HUM LEAD ruling on it, or add the gofmt leg now, under D-31's class.
- **H-8 · `00-REQUIRED-READING.md:26` vs `requirements.md:201-212`** · Minor · Simplify: Y
  - **Evidence:** Required reading says "Follow-ups: `06_docs/follow-ups.md` **only**", but `requirements.md` now keeps a second list of owed work.
  - **Action:** Name both, and say which list holds what.

### 5. Does the record say what happened, or what was intended?
- **H-9 · `08-reports/red-team-discover.md:28-29`** · Important · Simplify: N
  - **Evidence:** "Summaries are kept with the session's scratch output." The round-1 reports exist only in a temporary scratch directory. Every reviewer id in the ledger (A1, C-F10, H-6 …) points at text nobody can read, so a disposition cannot be checked against its source. H-4 was found only because the ledger's own paraphrase gives it away.
  - **Action:** File the five reviewer reports verbatim under `08-reports/`.

---

## Axis: Docs Quality

### 1. Does each document answer its title's question?
- **D-1 · `requirements.md:24-30`** · Minor · Simplify: N
  - **Evidence:** "What changes on screen" leaves out two visible decisions: playback is **off by default** (D-26), and the frame's time or "gap" is written on the map (L-1.10a). It also leaves out that the light ground may not blend at all (D-27).
  - **Action:** Add three bullets. The layout is the HUM LEAD's; this is about accuracy.

### 2. Does anything contradict the code or another document?
These are the rows where a ruling reached the record weaker than ruled, or contradicted.

- **D-2 · `project-brief.md:203` (and issue #2)** · Important · Simplify: N
  - **Evidence:** D-24 (`rulings.md:41`) rules that "M1 gains a non-visual arm". The adopted metric table still defines M1 by sight alone: "a reader states a precipitation cell's direction of motion". The arm exists only in L-1.12's Instrument cell, and `requirements.md` has no metrics section that would override the brief.
  - **Action:** Amend M1 in the brief, citing D-24, and refresh issue #2.
- **D-3 · `requirements.md:119` (L-8.3)** · Important · Simplify: N
  - **Evidence:** D-14 (`rulings.md:31`) and S29-2 list the **hatch** among what the radar erases, and D-14 makes that defect a requirement. L-8.3's list of what must survive leaves the hatch out. That leaves L-8.7 (hatch at Colours16) and L-8.1's "secondary cue" empty over radar.
  - **Action:** Add the hatch to L-8.3.
- **D-4 · `requirements.md:149-151`; `go-tuimaps/02-analysis/specimens/README.md:329`** · Important · Simplify: N
  - **Evidence:** L-11.2 requires the blend to pass "on both grounds", with no condition. L-11.4 (D-27) lets the light ground fall back to no blend. PLAN cannot satisfy both rows as written. L-11.3 also drops D-27's third search route ("or a per-class tint", `rulings.md:44`). The specimens README still says the blend "must pass on both grounds".
  - **Action:** Make L-11.2 conditional on L-11.4, add the per-class tint to L-11.3, and add a note to the specimens README citing D-27.
- **D-5 · `README.md:68-70`; `internal/fetch/fetch.go:24`** · Important · Simplify: N
  - **Evidence:** D-34 corrected the README to say the user-agent names "this library and its version". The code sends the hard-coded string `0.1.0-dev` (L-13.1). The fix meant to make the README true (L-4.4) states something new the code does not do.
  - **Action:** Say "a fixed development version string until L-13.1".
- **D-6 · `project-brief.md:57,215`** · Minor · Simplify: Y
  - **Evidence:** "Everything ruled since (D-14 to D-22)" is stale; the rulings now run to D-38. Line 215, "Whether v0.2.0 builds on a seam … is a PLAN question", contradicts D-35.
  - **Action:** Correct both.
- **D-7 · `wave2-measurements.md:11-12`** · Minor · Simplify: Y
  - **Evidence:** "by throwaway programs kept outside the repository" contradicts D-32 and the document's own lines 58 and 102.
  - **Action:** Correct it.
- **D-8 · `00-REQUIRED-READING.md:30-31` vs `requirements.md:206`** · Minor · Simplify: N
  - **Evidence:** Required reading says DISCOVER exit is gated by the D-17 specimen (OW-2). OW-2 is due "before PLAN commits to L-8.1", and D-17 agrees. Row 22 also still reads "APPROVED (D-7), AMENDED (D-11)", without D-23's CORRECTED.
  - **Action:** Pick one exit criterion and use it in both files.
- **D-9 · `requirements.md:131` (L-9.2); `tiles.go:141-142`** · Minor · Simplify: N
  - **Evidence:** "It empties only the current source." The code, and its own doc comment, empty everything when no source is named.
  - **Action:** Qualify the sentence.
- **D-10 · `go-tuimaps/03-architecture-design/L2-gates.md:11`** · Minor · Simplify: N
  - **Evidence:** The gate diagram names six modules. `tools/atlas/go.mod` is a seventh, and the gate checks it.
  - **Action:** Add it.

### 3. Unverified claims in a document that reads as verified
- **H-3's "Done" belongs here too.** It is the only unverified claim that matters.
- **D-11 · `02-analysis/programs/mrms-palette.go.txt:51`; `wave2-measurements.md:35`** · Minor · Simplify: N
  - **Evidence:** "111 by the twelfth frame and did not grow after" depends on the order the program read the frames. It globs `../nat/*.png` in name order (1, 10, 11 … 15, 2 …), not time order, so "twelfth" is not a point in time.
  - **Action:** State that the order is by file name, or re-run in time order.

### 4. Is the audience order right?
Yes. `requirements.md` opens with a designer/PM section, and the brief leads with intent. The gaps in that section are D-1.

### 5. Numbers without their blind spots (INST-5)
- **D-12 · `requirements.md:158` (L-12.2)** · Minor · Simplify: N
  - **Evidence:** "About 3 MB … or about five at dot resolution" gives no unit, and wave 2's Limits say the waves differ in unit. Counted as L-12.4 requires (pixels plus retained PNG), a dot-resolution loop is about 0.61 MB in decimal megabytes, so five come to 3.07 MB, over 3 MB. Only in MiB do five fit. The figure also comes from one run on one machine, and the row does not say so.
  - **Action:** State the unit and put the Limits pointer beside the number.
- **D-13 · `requirements.md:62` (L-2.1)** · Minor · Simplify: N
  - **Evidence:** "111 seen across 18 frames" repeats a figure that wave 2's Limits correct to 16 distinct times, on one quiet day (at most about 48 dBZ).
  - **Action:** Carry the caveat into the row.
- **D-14 · `requirements.md:197` (RK-8)** · Minor · Simplify: N
  - **Evidence:** The residual risk leaves out that, under D-29, motion is described only relative to a named place. A host that names no place gives a listener who cannot see the animation no motion at all.
  - **Action:** Add it to the residual; the ruling itself stands.

---

**Round 1 checked by recomputation, and they hold:** S29-7's twelve blend numbers match `output/blend-measure.txt`; the 80.0–80.6 % figures (7,243/9,028, 20,397/25,310, 6,447/8,056); D-20's sic note and the M-B arithmetic (5.1 MB, 8.7 MB); D-30's nine items mapped to eleven rows; D-25's seven items to L-1.10a–g; D-37's six items; the atlas title and its staleness test; the unit tests for `internal/mvt`; `feature/radar-loops` pushed and matching 11d986a; and every commit took the gate lane its file types require.
