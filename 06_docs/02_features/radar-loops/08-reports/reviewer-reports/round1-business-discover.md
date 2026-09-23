# round1 business discover — reviewer report, verbatim

Filed under v0.2.0 D-38/round 2 (H-9): the reviewer's final message exactly as returned, except that machine paths are replaced by `<workspace>` and `<scratch>`. The harness refused reviewers' own report files, so this is the only copy.

---

Note: the harness refused my write of `report.md` to `.../scratchpad/rt1-business-discover/`. That report is not on disk. The full report is below.

# RT1: Business Quality and DISCOVER lens: go-tuiMaps v0.2.0 "Radar loops"

**What I read:**
- everything under `radar-loops/`;
- specimen 29 (README §29 and the `29*` files);
- the diff `feature/go-tuimaps..feature/radar-loops` (11 commits);
- watchpost's `observer-maps/01-objectives/requirements.md`, for comparison.

Paths are relative to `go-tuimaps/06_docs/02_features/` unless marked `wp:` (watchpost `observer-maps/`).

## Section 1: Business Quality

### Q1. Does the work meet the requirement that was written down, or the one that was remembered?

**B-1. There is no requirements document, and that is a gap.** The brief is the only written requirement set. It was last amended at D-11, and eleven rulings since then changed scope without touching it.

- **Cancelled but still written.** L-5.1 still says the triage "is written" (`radar-loops/01-objectives/project-brief.md:91-94`). D-18 ruled it not recoverable (`radar-loops/02-analysis/rulings.md:35`).
- **Already exists but still asked for.** L-9.2 still asks for a purge call (`project-brief.md:137`). Wave 1 found `Map.Purge` exists and said "the brief needs correcting before it misdirects PLAN" (`wave1-findings.md:158-160, 180-181`). It was not corrected.
- **New obligations the brief does not carry:**
  - S29-2, the no-colour defect ("becomes a v0.2.0 requirement", `rulings.md:31`).
  - D-17: a severity word plus a dash (`rulings.md:34`).
  - D-14: the blend replaces FR-12 step 4.
  - D-20: a per-map image budget (`rulings.md:37`).
  - D-18: the tag-gate test, the `-dev` string fix, and a REVIEW covering all of v0.1.0.
  - D-19: the heavy-end colour affordances (`rulings.md:36`).
  - The "fix regardless" list (`wave1-findings.md:88-89, 196-198`). It includes `fetch.Checked`, a security check nothing calls.
- **Stale text.** The brief still says "HR-1..HR-5" (`project-brief.md:51`).

A PLAN written from the brief would build one cancelled requirement and one that already exists, and miss at least eight. Watchpost keeps a traced `requirements.md` (`wp:01-objectives/requirements.md:44-96`); this release has nothing equivalent.
- Severity: **Critical**
- Evidence: as cited
- Simplify/Delete? N
- Action: before PLAN, write `01-objectives/requirements.md`. Trace every item to an L-, D- or S29- id, give it an instrument or "NO INSTRUMENT YET", and mark L-5 and L-9.2 as superseded.

**B-2. D-18's action is not done and not tracked.** D-18 rules that "one row in v0.1.0's record says the close-out did not run" (`rulings.md:35`). No v0.1.0 file contains it: grepping for "close-out" and "did not run" hits only v0.2.0 files. `go-tuimaps/07-readiness/release-checklist.md:7` still reads "Not started".
- Severity: Important
- Simplify/Delete? N
- Action: write the row, or list it as an open DISCOVER exit item.

**B-3. The record is not quite verbatim or consistent.**
- The D-0 quote is truncated with "..." (`rulings.md:17`).
- The docs lane runs "every module's tests" (`scripts/gate:125,208`), but D-15 lists specific tests (`rulings.md:32`). That is broader than ruled, and nobody ratified it.
- Gate time is given three ways: "seventeen minutes" (`00-REQUIRED-READING.md:29`), 18 min (`rulings.md:27`) and 20–40 min (`rulings.md:32`).
- Severity: Minor
- Simplify/Delete? N
- Action: quote D-0 in full, send the lane widening to the lead for ratification, and settle on one timing.

### Q2. What does the user lose if this ships as it stands?

**B-4. A reader who can't watch the animation gets no motion at all.** The locked problem is that the reader "cannot see a storm's motion" (`project-brief.md:42-44`). The record answers it only with animation:
- motion-off keeps the newest frame (`wp:requirements.md:95`);
- Describe answers for the newest frame only (`wave1-findings.md:72-74`);
- M1 tests a sighted viewer (`project-brief.md:189`).

A listener who uses ReduceMotion, turns motion off, or uses a screen reader keeps the original problem. Watchpost's RK-12 notes three exclusions that were "found only by a blind reviewer" (`wp:requirements.md:158`).
- Severity: **Important**
- Simplify/Delete? N
- Action: this is a UX fork for the lead. Either express motion without animation (in the description, or as a still trail form), or record the exclusion as accepted.

**B-5. The MRMS table is weakest on the days that matter most.** It is built from one day's palette. The day's peak was about 48 dBZ, and the heavy-end colours are the least trustworthy (`wave2-measurements.md:39-43`). D-19 is "for now" and defers the heavy end to PLAN (`rulings.md:36`). Watchpost calls an outbreak "this product's core moment" (`wp:requirements.md:154`).
- Severity: Important
- Simplify/Delete? N
- Action: make "heavy end valued and tested before SHIP" a written requirement with an instrument, or state the exposure in the release notes.

**B-6. The light-ground blend has no solution and no fallback.** S29-6 says it fails "at every strength tried" and "may have no solution" (`go-tuimaps/02-analysis/specimens/README.md:326`). D-14 rules it "solvable" (`rulings.md:31`), which is confidence, not evidence. Nothing says what light-theme users get if PLAN cannot solve it.
- Severity: Important
- Simplify/Delete? N
- Action: ask the lead for a fallback now, for example the 29b order on a light ground. This is a UX fork; the lead decides.

### Q3. What was cut, and is the cut recorded where the next reader will find it?

**B-7. Several deferrals are not in `follow-ups.md`.** The record calls that file the only place for follow-ups (`00-REQUIRED-READING.md:25`; `06_docs/follow-ups.md:3-5`). Missing from it:
- **The 40-second freeze** in `FuzzAgree`. The brief says it "stays open" (`project-brief.md:113-114`), but it is absent from `gate-fuzz-deadline.md`.
- **S29-3.** The Hamlin marker is missing at 69×12 even in the control, "recorded, not diagnosed" (`specimens/README.md:318`). That is a user-facing defect.
- **Hosted CI's second-architecture leg.** It relies on emulation that a Linux runner lacks (`wave1-findings.md:121-122`).
- **The stale NOT RUN line** (`scripts/gate:271`).
- Severity: Important
- Simplify/Delete? N
- Action: add each one to follow-ups or to B-1's requirements.

**B-8. The public v0.1.0 tag stays unreviewed, with no signal to anyone who depends on it.** D-18's options never considered a `retract` directive in v0.2.0's `go.mod` (`rulings.md:35`). It is one line, and it is Go's own way of saying "don't use v0.1.0".
- Severity: Minor
- Simplify/Delete? N
- Action: offer it to the lead as an option.

### Q4. Is there a cheaper way to get the same outcome for the user?

**B-9. Two cheaper paths were never evaluated.**
- **A loop the host builds itself.** Wave 2 already ran a working 12-frame loop on v0.1.0, with one overlay per frame: 27 ms render, 0.59 MB (`wave2-measurements.md:47-58`). The record never weighs this as an interim step or as "do nothing in the library". Watchpost's D-35 may rule it out, but this record doesn't show it was considered.
- **A narrower release.** Watchpost can wait on some requirements:
  - it already has workarounds for L-7 (`wp:requirements.md:142`) and L-9 (`wp:requirements.md:71`);
  - it has a ship-without-radar criterion (`wp:requirements.md:153`).

  Meanwhile D-18 added a REVIEW of everything v0.1.0 shipped. Nothing separates what must land before radar from what can follow in a v0.2.x.
- Severity: Important
- Simplify/Delete? **Y**
- Action: put a before-radar / can-follow split to the lead as one ruling.

## Section 2: DISCOVER lens

### 1. Problem Framing

**P-1. The problem statement covers about half the release.**
- D-5 concedes that L-3..L-6 are "not [justified] by the statement" (`rulings.md:22`).
- D-18 made v0.2.0 the library's first reviewed release.
- The statement also prescribes a mechanism, "give the library the frames" (`project-brief.md:43-44`). That rules out answers to "where it is going" that aren't frames, which is how B-4 happened.
- Severity: Important
- Simplify/Delete? N
- Action: add a scope statement naming the three parts of the release: loops, contract repair, and the first release review. Keep the lock unless the lead chooses to reopen it.

### 2. Stakeholder Coverage

**P-2. Only the lead and one host's reviewers were heard.** C-8 admits the library's own rounds missed a third of the requirements (`project-brief.md:174-179`). Also:
- no library-side review round has been held;
- no second host was consulted, although one is claimed as a beneficiary (`project-brief.md:33-34`);
- no non-visual reader takes part in M1 (`project-brief.md:189`);
- no data provider was contacted.
- Severity: Important
- Simplify/Delete? N
- Action: run the library's own accessibility and security rounds before exit, and either drop or substantiate the "other hosts" claim.

### 3. Requirements Depth

**P-3. Some requirements are solutions, or not yet backed by evidence.**
- L-1.5 imports the host's Settings design (`project-brief.md:64`).
- L-6.4 is gate tooling, not a product requirement (`project-brief.md:104`).
- D-17 commits to its answer while its own specimen is still owed "before PLAN" (`rulings.md:34`).
- Severity: Minor, except the D-17 specimen, which blocks DISCOVER exit
- Simplify/Delete? N
- Action: produce the D-17 specimen at 69×12 and 149×38, and move L-6.4 into the constraints.

### 4. Inaction Analysis

**P-4. The "not built" case is overstated, and there is no schedule.** The brief says watchpost would ship a still frame or no radar (`project-brief.md:36-38`). Watchpost's own RK-4 says alerts plus the description "satisfy the locked problem on their own", with radar in 0.18.1 (`wp:requirements.md:153`). The urgency is real but soft. No date appears anywhere, so D-18's scope growth can't be weighed against the timeline.
- Severity: Minor
- Simplify/Delete? N
- Action: cite RK-4 and state a target window.

### 5. Implicit Requirements

**P-5. Nobody owns the maximum loop rate.**
- The library says "a stated maximum rate rides with it", with no number (`project-brief.md:68-69`).
- Watchpost's FR-5.9 waits on the library, "NO INSTRUMENT YET" (`wp:requirements.md:96`).
- NFR-21 is a flash ceiling, not a rate (`wave1-findings.md:45`).

A frame change can swap large colour areas, so the photosensitivity question is real, and each record defers it to the other.
- Severity: Important
- Simplify/Delete? N
- Action: make the rate and a flash analysis a library requirement with an owner.

**P-6. There is no budget for per-frame playback cost.** M4 covers memory only (`project-brief.md:192`). The CPU cost of each frame advance, including D-14's blend, is unmeasured, and watchpost's RK-7 depends on it (`wp:requirements.md:160`).
- Severity: Minor
- Simplify/Delete? N
- Action: add a render-per-advance budget alongside M4.

### 6. External Context

**P-7. The table depends on NOAA's palette, and nothing covers provider drift.** L-1.4 correctly keeps fetching in the host (`project-brief.md:62-63`). But D-19's table is built from colours observed on one day (`rulings.md:36`). Nothing covers a palette change, an endpoint moving, or IEM's terms. Watchpost records "IEM has no SLA, MRMS has no published table" (`wp:requirements.md:155`); this record has no equivalent.
- Severity: Minor
- Simplify/Delete? N
- Action: add a provider-drift risk. The existing unmatched count could serve as its detector.

### 7. Risk Assessment Quality

**P-8. There is no risk register.** The only risk lines are two "risk update" notes, for M4 and M5, with no likelihood or severity (`wave1-findings.md:185-188`). Watchpost has RK-1..RK-12, each with likelihood, severity and mitigation (`wp:requirements.md:150-162`). Missing here:
- the heavy end (B-5);
- the light ground (B-6);
- the loop rate (P-5);
- D-18's scope against the timeline (P-4);
- provider drift (P-7);
- a gate fix that was never reproduced (`gate-fuzz-deadline.md:63-65`);
- the freeze (B-7).
- Severity: **Important**
- Simplify/Delete? N
- Action: build the register inside B-1's requirements document.

**P-9. The evidence behind three rulings can't be re-run.**
- Wave 2 used "throwaway programs kept outside the repository" (`wave2-measurements.md:11-12`).
- Specimens 29b and 29d–f came from a "scratch copy" or "scratch function" (`specimens/README.md:311, 321`).
- D-19 and D-20 hand PLAN numbers taken "from these measurements" (`rulings.md:37`), and D-14 can't be re-measured after tuning.
- Severity: Important
- Simplify/Delete? N
- Action: commit the measurement programs so the numbers can be reproduced.

## Verdict

**Do not proceed to PLAN yet.** The investigation itself is strong: the measurements are real, contradictions are stated, and every ruling traces to evidence. What's missing is what PLAN designs against, and each gap is hours of work, not a redesign:
- a current, traced requirement set;
- a risk register;
- the D-17 specimen, which D-17 itself requires before PLAN;
- D-18's undone row;
- two forks for the lead: B-4 (non-animated motion) and P-5 (who owns the loop rate).

**Fix first: B-1.** Write `requirements.md` from the brief plus D-11..D-22 and the S29 findings, mark L-5 and L-9.2 as superseded, and include the risk register. Without it, PLAN builds a triage the lead cancelled and a purge call that already exists, and misses what D-14, D-17, D-18 and D-20 added.
